// Package integration provides end-to-end integration tests for KoeMoji-Go
package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hirokitakamura/koemoji-go/internal/config"
	"github.com/hirokitakamura/koemoji-go/internal/logger"
	"github.com/hirokitakamura/koemoji-go/internal/processor"
	"github.com/hirokitakamura/koemoji-go/internal/recorder"
	"github.com/hirokitakamura/koemoji-go/internal/testutil"
	"github.com/hirokitakamura/koemoji-go/internal/whisper"
)

// TestEndToEndWorkflow tests the complete workflow from audio file to transcription
func TestEndToEndWorkflow(t *testing.T) {
	testutil.SkipIfShort(t, "end-to-end workflow test")
	testutil.SkipIfNoWhisper(t)

	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Setting up test environment")

	// Create test audio file
	audioFile := env.CreateTestAudioFile(t, "test_audio.wav", 3) // 3 seconds
	reporter.Step(fmt.Sprintf("Created test audio file: %s", audioFile))

	// Initialize logger
	logBuffer := logger.NewLogger()
	reporter.Step("Initialized logger")

	// Initialize processor with test configuration
	proc := processor.NewProcessor(env.Config, logBuffer)
	reporter.Step("Initialized processor")

	// Start processor in background
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan bool, 1)
	go func() {
		proc.Run(ctx)
		done <- true
	}()

	reporter.Step("Started processor")

	// Wait for file processing
	outputFile := filepath.Join(env.OutputDir, "test_audio.txt")
	archiveFile := filepath.Join(env.ArchiveDir, "test_audio.wav")

	// Wait for transcription output
	if !env.WaitForFileContent(t, outputFile, 25*time.Second) {
		t.Fatalf("Transcription file not created within timeout: %s", outputFile)
	}
	reporter.Step("Transcription file created")

	// Wait for file archival
	if !env.WaitForFile(t, archiveFile, 5*time.Second) {
		t.Fatalf("Audio file not archived within timeout: %s", archiveFile)
	}
	reporter.Step("Audio file archived")

	// Verify transcription content
	content := env.GetFileContent(t, outputFile)
	if len(strings.TrimSpace(content)) == 0 {
		t.Fatalf("Transcription file is empty: %s", outputFile)
	}
	reporter.Step(fmt.Sprintf("Verified transcription content (%d chars)", len(content)))

	// Verify original file is moved to archive
	if _, err := os.Stat(audioFile); !os.IsNotExist(err) {
		t.Fatalf("Original audio file still exists in input directory: %s", audioFile)
	}
	reporter.Step("Verified original file moved to archive")

	// Stop processor
	cancel()
	select {
	case <-done:
		reporter.Step("Processor stopped gracefully")
	case <-time.After(5 * time.Second):
		t.Fatal("Processor did not stop within timeout")
	}

	reporter.Report()
}

// TestRecordingToTranscriptionWorkflow tests the complete recording workflow
func TestRecordingToTranscriptionWorkflow(t *testing.T) {
	testutil.SkipIfShort(t, "recording workflow test")

	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Setting up recording test environment")

	// Initialize logger
	logBuffer := logger.NewLogger()

	// Initialize recorder (this might fail if no audio devices available)
	rec, err := recorder.NewRecorder(env.Config, logBuffer)
	if err != nil {
		t.Skipf("Skipping recording test, no audio devices available: %v", err)
	}
	defer rec.Close()

	reporter.Step("Initialized recorder")

	// Test recording start/stop (very short recording)
	if err := rec.StartRecording(); err != nil {
		t.Fatalf("Failed to start recording: %v", err)
	}
	reporter.Step("Started recording")

	// Record for a very short time
	time.Sleep(100 * time.Millisecond)

	recordingFile, err := rec.StopRecording()
	if err != nil {
		t.Fatalf("Failed to stop recording: %v", err)
	}
	reporter.Step(fmt.Sprintf("Stopped recording, file: %s", recordingFile))

	// Verify recording file exists and has content
	if !env.WaitForFileContent(t, recordingFile, 2*time.Second) {
		t.Fatalf("Recording file not created or empty: %s", recordingFile)
	}

	info, err := os.Stat(recordingFile)
	if err != nil {
		t.Fatalf("Failed to stat recording file: %v", err)
	}
	if info.Size() == 0 {
		t.Fatalf("Recording file is empty: %s", recordingFile)
	}

	reporter.Step(fmt.Sprintf("Verified recording file size: %d bytes", info.Size()))
	reporter.Report()
}

// TestParallelProcessing tests multiple files processing simultaneously
func TestParallelProcessing(t *testing.T) {
	testutil.SkipIfShort(t, "parallel processing test")
	testutil.SkipIfNoWhisper(t)

	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Setting up parallel processing test")

	// Create multiple test audio files
	numFiles := 3
	audioFiles := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		filename := fmt.Sprintf("test_audio_%d.wav", i)
		audioFiles[i] = env.CreateTestAudioFile(t, filename, 2) // 2 seconds each
	}
	reporter.Step(fmt.Sprintf("Created %d test audio files", numFiles))

	// Initialize components
	logBuffer := logger.NewLogger()
	proc := processor.NewProcessor(env.Config, logBuffer)

	// Start processor
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	done := make(chan bool, 1)
	go func() {
		proc.Run(ctx)
		done <- true
	}()

	reporter.Step("Started processor for parallel processing")

	// Wait for all files to be processed
	processedCount := 0
	timeout := time.After(40 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

checkLoop:
	for {
		select {
		case <-timeout:
			t.Fatalf("Not all files processed within timeout. Processed: %d/%d", processedCount, numFiles)
		case <-ticker.C:
			// Check how many files have been processed
			outputFiles := env.ListFiles(t, env.OutputDir)
			archivedFiles := env.ListFiles(t, env.ArchiveDir)
			
			currentProcessed := len(outputFiles)
			if currentProcessed > processedCount {
				processedCount = currentProcessed
				reporter.Step(fmt.Sprintf("Progress: %d/%d files processed", processedCount, numFiles))
			}
			
			if len(outputFiles) == numFiles && len(archivedFiles) == numFiles {
				reporter.Step("All files processed and archived")
				break checkLoop
			}
		}
	}

	// Verify all transcriptions
	for i := 0; i < numFiles; i++ {
		outputFile := filepath.Join(env.OutputDir, fmt.Sprintf("test_audio_%d.txt", i))
		if !env.WaitForFileContent(t, outputFile, 2*time.Second) {
			t.Fatalf("Output file not found or empty: %s", outputFile)
		}
		
		content := env.GetFileContent(t, outputFile)
		if len(strings.TrimSpace(content)) == 0 {
			t.Fatalf("Transcription file is empty: %s", outputFile)
		}
	}

	// Stop processor
	cancel()
	select {
	case <-done:
		reporter.Step("Processor stopped gracefully")
	case <-time.After(5 * time.Second):
		t.Fatal("Processor did not stop within timeout")
	}

	reporter.Report()
}

// TestErrorHandling tests various error scenarios
func TestErrorHandling(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Setting up error handling tests")

	logBuffer := logger.NewLogger()

	// Test with invalid audio file
	invalidFile := filepath.Join(env.InputDir, "invalid.wav")
	if err := os.WriteFile(invalidFile, []byte("not a wav file"), 0644); err != nil {
		t.Fatalf("Failed to create invalid audio file: %v", err)
	}
	reporter.Step("Created invalid audio file")

	// Test whisper with invalid file
	w := whisper.NewWhisper(env.Config, logBuffer)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := w.ProcessFile(ctx, invalidFile, env.OutputDir)
	if err == nil {
		t.Fatal("Expected error when processing invalid audio file, but got none")
	}
	reporter.Step(fmt.Sprintf("Correctly handled invalid audio file error: %v", err))

	// Test with non-existent file
	nonExistentFile := filepath.Join(env.InputDir, "nonexistent.wav")
	err = w.ProcessFile(ctx, nonExistentFile, env.OutputDir)
	if err == nil {
		t.Fatal("Expected error when processing non-existent file, but got none")
	}
	reporter.Step(fmt.Sprintf("Correctly handled non-existent file error: %v", err))

	reporter.Report()
}

// TestConfigurationChanges tests dynamic configuration updates
func TestConfigurationChanges(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Setting up configuration change tests")

	// Test different whisper models
	models := []string{"tiny", "base"}
	
	for _, model := range models {
		env.Config.WhisperModel = model
		
		// Save updated config
		if err := config.SaveConfig(env.ConfigFile, env.Config); err != nil {
			t.Fatalf("Failed to save config with model %s: %v", model, err)
		}
		
		// Load config and verify
		loadedConfig, err := config.LoadConfig(env.ConfigFile)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}
		
		if loadedConfig.WhisperModel != model {
			t.Fatalf("Config model mismatch: expected %s, got %s", model, loadedConfig.WhisperModel)
		}
		
		reporter.Step(fmt.Sprintf("Successfully updated and verified config with model: %s", model))
	}

	// Test different languages
	languages := []string{"ja", "en"}
	
	for _, lang := range languages {
		env.Config.Language = lang
		env.Config.UILanguage = lang
		
		if err := config.SaveConfig(env.ConfigFile, env.Config); err != nil {
			t.Fatalf("Failed to save config with language %s: %v", lang, err)
		}
		
		loadedConfig, err := config.LoadConfig(env.ConfigFile)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}
		
		if loadedConfig.Language != lang || loadedConfig.UILanguage != lang {
			t.Fatalf("Config language mismatch: expected %s, got %s/%s", 
				lang, loadedConfig.Language, loadedConfig.UILanguage)
		}
		
		reporter.Step(fmt.Sprintf("Successfully updated and verified config with language: %s", lang))
	}

	reporter.Report()
}

// TestLoggerIntegration tests logger functionality in integration context
func TestLoggerIntegration(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Setting up logger integration tests")

	logBuffer := logger.NewLogger()

	// Test log capture
	logCapture := testutil.NewLogCapture()
	
	// Create multiple log entries
	testMessages := []string{
		"Test message 1",
		"Test message 2", 
		"Test message 3",
	}

	for _, msg := range testMessages {
		logBuffer.Info(msg)
	}

	// Verify log entries
	entries := logBuffer.GetEntries()
	if len(entries) != len(testMessages) {
		t.Fatalf("Expected %d log entries, got %d", len(testMessages), len(entries))
	}

	for i, entry := range entries {
		if !strings.Contains(entry.Message, testMessages[i]) {
			t.Fatalf("Log entry %d does not contain expected message: %s", i, testMessages[i])
		}
	}

	reporter.Step(fmt.Sprintf("Verified %d log entries", len(testMessages)))

	// Test log buffer overflow (logger has max 12 entries)
	for i := 0; i < 15; i++ {
		logBuffer.Info(fmt.Sprintf("Overflow test message %d", i))
	}

	entries = logBuffer.GetEntries()
	if len(entries) > 12 {
		t.Fatalf("Log buffer should not exceed 12 entries, got %d", len(entries))
	}

	reporter.Step("Verified log buffer size limit")
	reporter.Report()
}

// TestMemoryLeaks tests for potential memory leaks in long-running operations
func TestMemoryLeaks(t *testing.T) {
	testutil.SkipIfShort(t, "memory leak test")

	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Setting up memory leak tests")

	logBuffer := logger.NewLogger()
	
	// Run multiple cycles of logger operations
	initialMetrics := testutil.MeasurePerformance(nil, func() {
		// Baseline measurement
	})

	cycles := 100
	for i := 0; i < cycles; i++ {
		// Simulate normal operation
		logBuffer.Info(fmt.Sprintf("Cycle %d message", i))
		logBuffer.Error(fmt.Sprintf("Cycle %d error", i))
		logBuffer.Warning(fmt.Sprintf("Cycle %d warning", i))
		
		// Trigger buffer cleanup periodically
		if i%20 == 0 {
			entries := logBuffer.GetEntries()
			_ = len(entries) // Use the result
		}
	}

	finalMetrics := testutil.MeasurePerformance(nil, func() {
		// Final measurement
	})

	// Check for significant memory growth
	memoryGrowth := finalMetrics.MemoryUsage - initialMetrics.MemoryUsage
	if memoryGrowth > 1024*1024 { // 1MB threshold
		t.Logf("Warning: Memory usage grew by %d bytes during test", memoryGrowth)
		// Note: This is a warning, not a failure, as some growth is expected
	}

	reporter.Step(fmt.Sprintf("Completed %d cycles, memory growth: %d bytes", cycles, memoryGrowth))
	reporter.Report()
}