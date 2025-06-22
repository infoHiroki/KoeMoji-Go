package testdata

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/hirokitakamura/koemoji-go/internal/config"
	"github.com/hirokitakamura/koemoji-go/internal/logger"
)

// CreateTestConfig creates a test configuration
func CreateTestConfig(t *testing.T) *config.Config {
	t.Helper()
	
	tempDir := t.TempDir()
	
	return &config.Config{
		WhisperModel:        "base",
		Language:            "ja",
		UILanguage:          "ja",
		ScanIntervalMinutes: 1,
		MaxCpuPercent:       80,
		ComputeType:         "int8",
		UseColors:           false,
		OutputFormat:        "txt",
		InputDir:            filepath.Join(tempDir, "input"),
		OutputDir:           filepath.Join(tempDir, "output"),
		ArchiveDir:          filepath.Join(tempDir, "archive"),
	}
}

// CreateTestLogger creates a test logger with buffer
func CreateTestLogger() (*log.Logger, *[]logger.LogEntry, *sync.RWMutex) {
	testLogger := log.New(os.Stdout, "", log.LstdFlags)
	logBuffer := &[]logger.LogEntry{}
	logMutex := &sync.RWMutex{}
	return testLogger, logBuffer, logMutex
}

// CreateTestAudioFile creates a test audio file
func CreateTestAudioFile(t *testing.T, inputDir string, filename string) string {
	t.Helper()
	
	err := os.MkdirAll(inputDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create input directory: %v", err)
	}
	
	audioFile := filepath.Join(inputDir, filename)
	
	// Create a dummy audio file with WAVE header
	waveHeader := []byte{
		0x52, 0x49, 0x46, 0x46, // "RIFF"
		0x24, 0x00, 0x00, 0x00, // File size - 8
		0x57, 0x41, 0x56, 0x45, // "WAVE"
		0x66, 0x6d, 0x74, 0x20, // "fmt "
		0x10, 0x00, 0x00, 0x00, // Subchunk1Size
		0x01, 0x00,             // AudioFormat (PCM)
		0x01, 0x00,             // NumChannels (Mono)
		0x44, 0xAC, 0x00, 0x00, // SampleRate (44100)
		0x88, 0x58, 0x01, 0x00, // ByteRate
		0x02, 0x00,             // BlockAlign
		0x10, 0x00,             // BitsPerSample
		0x64, 0x61, 0x74, 0x61, // "data"
		0x00, 0x00, 0x00, 0x00, // Subchunk2Size
	}
	
	err = os.WriteFile(audioFile, waveHeader, 0644)
	if err != nil {
		t.Fatalf("Failed to create test audio file: %v", err)
	}
	
	return audioFile
}

// AssertFileExists checks if a file exists
func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist, but it doesn't", path)
	}
}

// AssertFileNotExists checks if a file does not exist
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()
	
	if _, err := os.Stat(path); err == nil {
		t.Errorf("Expected file %s to not exist, but it does", path)
	}
}

// CreateDirectories creates test directories
func CreateDirectories(t *testing.T, dirs ...string) {
	t.Helper()
	
	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
}

// CleanupTestFiles removes test files and directories
func CleanupTestFiles(t *testing.T, paths ...string) {
	t.Helper()
	
	for _, path := range paths {
		if err := os.RemoveAll(path); err != nil {
			t.Logf("Warning: Failed to cleanup %s: %v", path, err)
		}
	}
}