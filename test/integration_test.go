// Package integration_test provides end-to-end integration tests for KoeMoji-Go
package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hirokitakamura/koemoji-go/internal/config"
	"github.com/hirokitakamura/koemoji-go/internal/recorder"
	testutil "github.com/hirokitakamura/koemoji-go/test/shared"
)

// TestBasicPackageImports tests that all packages can be imported without conflicts
func TestBasicPackageImports(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Setting up basic integration test environment")

	// Test that basic configuration loading works
	if env.Config == nil {
		t.Fatal("Failed to create test configuration")
	}
	reporter.Step("Configuration loaded successfully")

	// Test that directories were created
	for _, dir := range []string{env.InputDir, env.OutputDir, env.ArchiveDir} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Fatalf("Directory not created: %s", dir)
		}
	}
	reporter.Step("Test directories created successfully")

	// Create a test file to verify file operations work
	testFile := filepath.Join(env.InputDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if !env.WaitForFile(t, testFile, 1*time.Second) {
		t.Fatal("Test file not found after creation")
	}
	reporter.Step("Basic file operations work")

	reporter.Report()
}

// TestRecorderBasicFunctionality tests basic recorder functionality
func TestRecorderBasicFunctionality(t *testing.T) {
	testutil.SkipIfShort(t, "recorder basic functionality test")

	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Setting up recorder test environment")

	// Initialize recorder (this might fail if no audio devices available)
	rec, err := recorder.NewRecorder()
	if err != nil {
		t.Skipf("Skipping recording test, no audio devices available: %v", err)
	}
	defer rec.Close()

	reporter.Step("Initialized recorder successfully")

	// Test basic recording state
	if rec.IsRecording() {
		t.Fatal("Recorder should not be recording initially")
	}
	reporter.Step("Verified initial recording state")

	// Test device listing
	devices, err := recorder.ListDevices()
	if err != nil {
		t.Logf("Warning: Could not list devices: %v", err)
	} else {
		reporter.Step("Successfully listed audio devices")
		t.Logf("Found %d audio devices", len(devices))
	}

	reporter.Report()
}

// TestMultipleFileHandling tests basic file handling functionality
func TestMultipleFileHandling(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Setting up multiple file handling test")

	// Create multiple test audio files
	numFiles := 3
	audioFiles := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		filename := fmt.Sprintf("test_audio_%d.wav", i)
		audioFiles[i] = env.CreateTestAudioFile(t, filename, 2) // 2 seconds each
	}
	reporter.Step("Created multiple test audio files")

	// Verify all files were created
	for i, audioFile := range audioFiles {
		if !env.WaitForFile(t, audioFile, 1*time.Second) {
			t.Fatalf("Audio file %d not created: %s", i, audioFile)
		}

		info, err := os.Stat(audioFile)
		if err != nil {
			t.Fatalf("Failed to stat audio file %d: %v", i, err)
		}
		if info.Size() == 0 {
			t.Fatalf("Audio file %d is empty: %s", i, audioFile)
		}
	}
	reporter.Step("Verified all test files exist and have content")

	// Test file listing functionality
	files := env.ListFiles(t, env.InputDir)
	if len(files) != numFiles {
		t.Fatalf("Expected %d files, found %d", numFiles, len(files))
	}
	reporter.Step("File listing functionality works correctly")

	reporter.Report()
}

// TestErrorHandling tests various error scenarios
func TestErrorHandling(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Setting up error handling tests")

	// Test with invalid directory operations
	invalidDir := "/nonexistent/directory"
	_, err := os.ReadDir(invalidDir)
	if err == nil {
		t.Fatal("Expected error when reading non-existent directory, but got none")
	}
	reporter.Step("Correctly handled directory read error")

	// Test with invalid file operations
	invalidFile := filepath.Join(env.InputDir, "nonexistent.wav")
	_, err = os.Stat(invalidFile)
	if err == nil {
		t.Fatal("Expected error when stating non-existent file, but got none")
	}
	reporter.Step("Correctly handled file stat error")

	// Test with invalid file creation in protected directory
	if os.Geteuid() != 0 { // Skip if running as root
		protectedFile := "/root/test.txt"
		err = os.WriteFile(protectedFile, []byte("test"), 0644)
		if err == nil {
			t.Fatal("Expected error when writing to protected directory, but got none")
		}
		reporter.Step("Correctly handled protected directory write error")
	}

	reporter.Report()
}

// TestConfigurationChanges tests dynamic configuration updates
func TestConfigurationChanges(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Setting up configuration change tests")

	// Test default configuration values
	if env.Config.WhisperModel == "" {
		t.Fatal("WhisperModel should not be empty")
	}
	reporter.Step("Verified default configuration values")

	// Test different whisper models
	models := []string{"tiny", "base"}

	for _, model := range models {
		env.Config.WhisperModel = model

		// Save updated config
		if err := config.SaveConfig(env.Config, env.ConfigFile); err != nil {
			t.Fatalf("Failed to save config with model %s: %v", model, err)
		}

		// Load config and verify
		loadedConfig, err := config.LoadConfig(env.ConfigFile, nil)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if loadedConfig.WhisperModel != model {
			t.Fatalf("Config model mismatch: expected %s, got %s", model, loadedConfig.WhisperModel)
		}

		reporter.Step("Successfully updated and verified config with model: " + model)
	}

	// Test different languages
	languages := []string{"ja", "en"}

	for _, lang := range languages {
		env.Config.Language = lang
		env.Config.UILanguage = lang

		if err := config.SaveConfig(env.Config, env.ConfigFile); err != nil {
			t.Fatalf("Failed to save config with language %s: %v", lang, err)
		}

		loadedConfig, err := config.LoadConfig(env.ConfigFile, nil)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if loadedConfig.Language != lang || loadedConfig.UILanguage != lang {
			t.Fatalf("Config language mismatch: expected %s, got %s/%s",
				lang, loadedConfig.Language, loadedConfig.UILanguage)
		}

		reporter.Step("Successfully updated and verified config with language: " + lang)
	}

	reporter.Report()
}

// TestLoggerIntegration tests logger functionality in integration context
func TestLoggerIntegration(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Setting up logger integration tests")

	// Test log capture functionality
	logCapture := testutil.NewLogCapture()

	// Test writing to log capture
	testMessage := "Test log message"
	_, err := logCapture.Write([]byte(testMessage))
	if err != nil {
		t.Fatalf("Failed to write to log capture: %v", err)
	}

	// Verify log entries
	entries := logCapture.GetEntries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(entries))
	}

	if entries[0] != testMessage {
		t.Fatalf("Log entry mismatch: expected %s, got %s", testMessage, entries[0])
	}

	reporter.Step("Verified log capture functionality")

	// Test clearing log entries
	logCapture.Clear()
	entries = logCapture.GetEntries()
	if len(entries) != 0 {
		t.Fatalf("Expected 0 log entries after clear, got %d", len(entries))
	}

	reporter.Step("Verified log clearing functionality")
	reporter.Report()
}

// TestPerformanceBasics tests basic performance measurement functionality
func TestPerformanceBasics(t *testing.T) {
	testutil.SkipIfShort(t, "performance test")

	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	reporter := testutil.NewTestReporter(t)
	reporter.Step("Setting up performance tests")

	// Test performance measurement
	metrics := testutil.MeasurePerformance(nil, func() {
		// Simulate some work
		time.Sleep(10 * time.Millisecond)
	})

	if metrics.Duration < 10*time.Millisecond {
		t.Fatalf("Expected duration >= 10ms, got %v", metrics.Duration)
	}

	reporter.Step("Verified performance measurement works")

	// Test file size calculation
	testFile := filepath.Join(env.InputDir, "test.dat")
	testData := make([]byte, 1024) // 1KB
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat test file: %v", err)
	}

	if info.Size() != 1024 {
		t.Fatalf("Expected file size 1024, got %d", info.Size())
	}

	reporter.Step("Verified file size calculations")
	reporter.Report()
}
