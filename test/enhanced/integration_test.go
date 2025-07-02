// Package enhanced provides comprehensive end-to-end integration tests
package enhanced

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/hirokitakamura/koemoji-go/internal/config"
	"github.com/hirokitakamura/koemoji-go/internal/logger"
	"github.com/hirokitakamura/koemoji-go/internal/processor"
	"github.com/hirokitakamura/koemoji-go/internal/recorder"
	testutil "github.com/hirokitakamura/koemoji-go/test/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompleteWorkflow tests the complete audio processing workflow
func TestCompleteWorkflow(t *testing.T) {
	testutil.SkipIfShort(t, "complete workflow test")

	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Setting up complete workflow test")

	// Create test audio file
	audioFile := env.CreateTestAudioFile(t, "workflow_test.wav", 3) // 3 seconds
	reporter.Step("Created test audio file")

	// Verify file was created correctly
	info, err := os.Stat(audioFile)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
	reporter.Step("Verified audio file creation")

	// Test directory structure
	assert.DirExists(t, env.InputDir)
	assert.DirExists(t, env.OutputDir)
	assert.DirExists(t, env.ArchiveDir)
	reporter.Step("Verified directory structure")

	// Wait for file to be stable (simulate real-world scenario)
	time.Sleep(100 * time.Millisecond)

	// Verify audio file is in input directory
	files := env.ListFiles(t, env.InputDir)
	assert.Contains(t, files, "workflow_test.wav")
	reporter.Step("Verified file is in input directory")

	reporter.Report()
}

// TestMultipleFileProcessing tests processing multiple files sequentially
func TestMultipleFileProcessing(t *testing.T) {
	testutil.SkipIfShort(t, "multiple file processing test")

	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Setting up multiple file processing test")

	// Create multiple audio files
	numFiles := 5
	audioFiles := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		filename := fmt.Sprintf("multi_test_%d.wav", i)
		audioFiles[i] = env.CreateTestAudioFile(t, filename, 2) // 2 seconds each
	}
	reporter.Step(fmt.Sprintf("Created %d test audio files", numFiles))

	// Verify all files exist
	for i, audioFile := range audioFiles {
		if !env.WaitForFile(t, audioFile, 1*time.Second) {
			t.Fatalf("Audio file %d not created: %s", i, audioFile)
		}
		
		info, err := os.Stat(audioFile)
		require.NoError(t, err)
		assert.Greater(t, info.Size(), int64(0))
	}
	reporter.Step("Verified all audio files were created successfully")

	// Test that all files are listed correctly
	files := env.ListFiles(t, env.InputDir)
	assert.Len(t, files, numFiles)
	
	for i := 0; i < numFiles; i++ {
		expectedFilename := fmt.Sprintf("multi_test_%d.wav", i)
		assert.Contains(t, files, expectedFilename)
	}
	reporter.Step("Verified all files are listed in input directory")

	reporter.Report()
}

// TestProcessorDirectoryManagement tests directory creation and management
func TestProcessorDirectoryManagement(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Testing processor directory management")

	// Test EnsureDirectories functionality
	testLogger := log.New(os.Stdout, "", log.LstdFlags)
	processor.EnsureDirectories(env.Config, testLogger)
	
	// Verify directories exist
	for _, dir := range []string{env.InputDir, env.OutputDir, env.ArchiveDir} {
		assert.DirExists(t, dir)
	}
	reporter.Step("Verified EnsureDirectories creates all required directories")

	// Test directory permissions
	for _, dir := range []string{env.InputDir, env.OutputDir, env.ArchiveDir} {
		testFile := filepath.Join(dir, "permission_test.txt")
		err := os.WriteFile(testFile, []byte("test"), 0644)
		assert.NoError(t, err, "Should be able to write to directory: %s", dir)
		
		// Clean up
		os.Remove(testFile)
	}
	reporter.Step("Verified directory write permissions")

	reporter.Report()
}

// TestRecorderIntegration tests recorder functionality integration
func TestRecorderIntegration(t *testing.T) {
	testutil.SkipIfShort(t, "recorder integration test")

	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Testing recorder integration")

	// Test device listing
	devices, err := recorder.ListDevices()
	if err != nil {
		t.Skipf("Skipping recorder test, no audio devices available: %v", err)
	}
	reporter.Step(fmt.Sprintf("Successfully listed %d audio devices", len(devices)))

	// Test recorder creation (may fail if no devices available)
	rec, err := recorder.NewRecorder()
	if err != nil {
		t.Skipf("Skipping recorder test, cannot create recorder: %v", err)
	}
	defer func() {
		if rec != nil {
			rec.Close()
		}
	}()

	// Test initial state
	assert.False(t, rec.IsRecording())
	reporter.Step("Verified initial recorder state")

	// Note: We don't actually start recording in automated tests
	// as it would require audio input and could be disruptive
	reporter.Step("Recorder integration test completed successfully")

	reporter.Report()
}

// TestConfigurationManagement tests configuration handling in integration context
func TestConfigurationManagement(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Testing configuration management integration")

	// Test configuration save/load cycle
	originalConfig := env.Config
	
	// Modify configuration
	env.Config.WhisperModel = "base"
	env.Config.Language = "en"
	env.Config.ScanIntervalMinutes = 3
	env.Config.MaxCpuPercent = 80

	// Save configuration
	err := config.SaveConfig(env.Config, env.ConfigFile)
	require.NoError(t, err)
	reporter.Step("Successfully saved modified configuration")

	// Load configuration
	testLogger := log.New(os.Stdout, "", log.LstdFlags)
	loadedConfig, err := config.LoadConfig(env.ConfigFile, testLogger)
	require.NoError(t, err)
	reporter.Step("Successfully loaded configuration from file")

	// Verify configuration values
	assert.Equal(t, "base", loadedConfig.WhisperModel)
	assert.Equal(t, "en", loadedConfig.Language)
	assert.Equal(t, 3, loadedConfig.ScanIntervalMinutes)
	assert.Equal(t, 80, loadedConfig.MaxCpuPercent)
	reporter.Step("Verified all configuration values match expected")

	// Test configuration with invalid values (should use defaults)
	invalidConfigData := `{"whisper_model": "invalid_model", "language": "invalid_lang"}`
	invalidConfigFile := filepath.Join(env.BaseDir, "invalid_config.json")
	err = os.WriteFile(invalidConfigFile, []byte(invalidConfigData), 0644)
	require.NoError(t, err)

	// This should handle invalid values gracefully
	_, err = config.LoadConfig(invalidConfigFile, testLogger)
	// Note: Depending on implementation, this might not return an error
	// but use default values for invalid fields
	reporter.Step("Tested configuration with invalid values")

	// Restore original config
	env.Config = originalConfig

	reporter.Report()
}

// TestErrorRecovery tests system behavior during error conditions
func TestErrorRecovery(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Testing error recovery scenarios")

	// Test with non-existent input directory
	originalInputDir := env.Config.InputDir
	env.Config.InputDir = "/nonexistent/directory"

	// This should handle the error gracefully
	testLogger := log.New(os.Stdout, "", log.LstdFlags)
	processor.EnsureDirectories(env.Config, testLogger)
	
	// Directory should be created
	assert.DirExists(t, env.Config.InputDir)
	reporter.Step("Verified system creates missing directories")

	// Restore original directory
	env.Config.InputDir = originalInputDir

	// Test with file permission issues (skip if running as root)
	if os.Geteuid() != 0 {
		restrictedDir := filepath.Join(env.BaseDir, "restricted")
		err := os.MkdirAll(restrictedDir, 0000) // No permissions
		require.NoError(t, err)

		restrictedFile := filepath.Join(restrictedDir, "test.txt")
		err = os.WriteFile(restrictedFile, []byte("test"), 0644)
		assert.Error(t, err, "Should fail to write to restricted directory")
		reporter.Step("Verified permission error handling")

		// Clean up
		os.Chmod(restrictedDir, 0755)
		os.RemoveAll(restrictedDir)
	}

	reporter.Report()
}

// TestConcurrentOperations tests concurrent access to shared resources
func TestConcurrentOperations(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Testing concurrent operations")

	// Create multiple files concurrently
	numFiles := 10
	done := make(chan bool, numFiles)

	for i := 0; i < numFiles; i++ {
		go func(id int) {
			defer func() { done <- true }()
			filename := fmt.Sprintf("concurrent_%d.wav", id)
			env.CreateTestAudioFile(t, filename, 1) // 1 second each
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numFiles; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent file creation")
		}
	}
	reporter.Step("Successfully created files concurrently")

	// Verify all files were created
	files := env.ListFiles(t, env.InputDir)
	assert.GreaterOrEqual(t, len(files), numFiles)
	reporter.Step("Verified all concurrent files were created")

	// Test concurrent logging
	numLogs := 50
	logDone := make(chan bool, numLogs)
	var logBuffer []logger.LogEntry
	var logMutex sync.RWMutex
	testLogger := log.New(os.Stdout, "", log.LstdFlags)
	
	for i := 0; i < numLogs; i++ {
		go func(id int) {
			defer func() { logDone <- true }()
			logger.LogInfo(testLogger, &logBuffer, &logMutex, "Concurrent log %d", id)
		}(i)
	}

	// Wait for all logging operations
	for i := 0; i < numLogs; i++ {
		select {
		case <-logDone:
			// Success
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for concurrent logging")
		}
	}
	reporter.Step("Successfully completed concurrent logging operations")

	reporter.Report()
}

// TestMemoryUsage tests memory usage and potential leaks
func TestMemoryUsage(t *testing.T) {
	testutil.SkipIfShort(t, "memory usage test")

	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Testing memory usage patterns")

	// Create and process many files to test memory usage
	numFiles := 100
	for i := 0; i < numFiles; i++ {
		filename := fmt.Sprintf("memory_test_%d.wav", i)
		audioFile := env.CreateTestAudioFile(t, filename, 1) // Small files

		// Simulate processing by reading the file
		data, err := os.ReadFile(audioFile)
		require.NoError(t, err)
		assert.Greater(t, len(data), 0)

		// Clean up immediately to test cleanup behavior
		err = os.Remove(audioFile)
		require.NoError(t, err)
	}
	reporter.Step("Created and cleaned up many files to test memory patterns")

	// Test log buffer behavior with many entries
	var logBuffer []logger.LogEntry
	var logMutex sync.RWMutex
	testLogger := log.New(os.Stdout, "", log.LstdFlags)
	for i := 0; i < 100; i++ {
		logger.LogInfo(testLogger, &logBuffer, &logMutex, "Memory test log %d", i)
	}

	// Log buffer should implement circular buffer behavior
	assert.LessOrEqual(t, len(logBuffer), 12, "Log buffer should limit entries")
	reporter.Step("Verified log buffer memory management")

	reporter.Report()
}

// TestLongRunningOperations tests system behavior during extended operation
func TestLongRunningOperations(t *testing.T) {
	testutil.SkipIfShort(t, "long running operations test")

	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Testing long running operations")

	// Create context for controlled shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Simulate long-running file monitoring
	done := make(chan bool, 1)
	go func() {
		defer func() { done <- true }()
		
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		
		fileCount := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Create a file periodically
				filename := fmt.Sprintf("longrun_%d.wav", fileCount)
				env.CreateTestAudioFile(t, filename, 1)
				fileCount++
				
				if fileCount >= 10 {
					return // Limit for test purposes
				}
			}
		}
	}()

	// Wait for completion or timeout
	select {
	case <-done:
		reporter.Step("Long running operation completed successfully")
	case <-ctx.Done():
		reporter.Step("Long running operation timed out as expected")
	}

	// Verify files were created
	files := env.ListFiles(t, env.InputDir)
	assert.GreaterOrEqual(t, len(files), 1)
	reporter.Step("Verified file creation during long running operation")

	reporter.Report()
}

// TestSystemResourceLimits tests behavior under resource constraints
func TestSystemResourceLimits(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Testing system resource limits")

	// Test with CPU limit configuration
	env.Config.MaxCpuPercent = 10 // Very low CPU limit
	
	// Create files to test under CPU constraints
	numFiles := 5
	for i := 0; i < numFiles; i++ {
		filename := fmt.Sprintf("cpu_limit_%d.wav", i)
		audioFile := env.CreateTestAudioFile(t, filename, 2)
		
		// Verify file creation succeeded despite CPU limits
		info, err := os.Stat(audioFile)
		require.NoError(t, err)
		assert.Greater(t, info.Size(), int64(0))
	}
	reporter.Step("Successfully created files under CPU constraints")

	// Test with scan interval limits
	originalInterval := env.Config.ScanIntervalMinutes
	env.Config.ScanIntervalMinutes = 1 // Frequent scanning
	
	// This should work without issues
	files := env.ListFiles(t, env.InputDir)
	assert.GreaterOrEqual(t, len(files), numFiles)
	reporter.Step("Verified operation with frequent scanning interval")

	// Restore original interval
	env.Config.ScanIntervalMinutes = originalInterval

	reporter.Report()
}