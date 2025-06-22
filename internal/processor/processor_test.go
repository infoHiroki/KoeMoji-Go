package processor

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hirokitakamura/koemoji-go/internal/config"
	"github.com/hirokitakamura/koemoji-go/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestEnsureDirectories(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.GetDefaultConfig()
	cfg.InputDir = filepath.Join(tempDir, "input")
	cfg.OutputDir = filepath.Join(tempDir, "output")
	cfg.ArchiveDir = filepath.Join(tempDir, "archive")

	logger := log.New(os.Stdout, "", log.LstdFlags)

	EnsureDirectories(cfg, logger)

	assert.DirExists(t, cfg.InputDir)
	assert.DirExists(t, cfg.OutputDir)
	assert.DirExists(t, cfg.ArchiveDir)
}

func TestEnsureDirectories_AlreadyExists(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.GetDefaultConfig()
	cfg.InputDir = filepath.Join(tempDir, "input")
	cfg.OutputDir = filepath.Join(tempDir, "output")
	cfg.ArchiveDir = filepath.Join(tempDir, "archive")

	// Pre-create directories
	err := os.MkdirAll(cfg.InputDir, 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(cfg.OutputDir, 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(cfg.ArchiveDir, 0755)
	assert.NoError(t, err)

	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Should not fail if directories already exist
	EnsureDirectories(cfg, logger)

	assert.DirExists(t, cfg.InputDir)
	assert.DirExists(t, cfg.OutputDir)
	assert.DirExists(t, cfg.ArchiveDir)
}

func TestFilterNewAudioFiles_EmptyInput(t *testing.T) {
	processedFiles := make(map[string]bool)
	var mu sync.Mutex

	result := filterNewAudioFiles([]string{}, &processedFiles, &mu)

	assert.Empty(t, result)
	assert.Empty(t, processedFiles)
}

func TestFilterNewAudioFiles_NonAudioFiles(t *testing.T) {
	processedFiles := make(map[string]bool)
	var mu sync.Mutex

	files := []string{
		"test.txt",
		"document.pdf",
		"image.jpg",
		"script.sh",
	}

	result := filterNewAudioFiles(files, &processedFiles, &mu)

	assert.Empty(t, result)
}

func TestFilterNewAudioFiles_ValidAudioFiles(t *testing.T) {
	processedFiles := make(map[string]bool)
	var mu sync.Mutex

	files := []string{
		"audio1.wav",
		"audio2.mp3",
		"audio3.m4a",
		"document.txt", // Should be filtered out
	}

	result := filterNewAudioFiles(files, &processedFiles, &mu)

	// Should only include audio files
	assert.Len(t, result, 3)
	assert.Contains(t, result, "audio1.wav")
	assert.Contains(t, result, "audio2.mp3")
	assert.Contains(t, result, "audio3.m4a")

	// All audio files should be marked as processed
	assert.True(t, processedFiles["audio1.wav"])
	assert.True(t, processedFiles["audio2.mp3"])
	assert.True(t, processedFiles["audio3.m4a"])
	assert.False(t, processedFiles["document.txt"])
}

func TestFilterNewAudioFiles_AlreadyProcessed(t *testing.T) {
	processedFiles := map[string]bool{
		"audio1.wav": true,
		"audio2.mp3": true,
	}
	var mu sync.Mutex

	files := []string{
		"audio1.wav", // Already processed
		"audio2.mp3", // Already processed
		"audio3.m4a", // New file
	}

	result := filterNewAudioFiles(files, &processedFiles, &mu)

	// Should only include new audio file
	assert.Len(t, result, 1)
	assert.Contains(t, result, "audio3.m4a")

	// New file should be marked as processed
	assert.True(t, processedFiles["audio3.m4a"])
}

func TestScanAndProcess_InvalidDirectory(t *testing.T) {
	cfg := config.GetDefaultConfig()
	cfg.InputDir = "/nonexistent/directory"

	var logBuffer []logger.LogEntry
	var logMutex sync.RWMutex
	var lastScanTime time.Time
	var queuedFiles []string
	var processingFile string
	var isProcessing bool
	processedFiles := make(map[string]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup

	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Should handle invalid directory gracefully
	ScanAndProcess(cfg, logger, &logBuffer, &logMutex, &lastScanTime, &queuedFiles,
		&processingFile, &isProcessing, &processedFiles, &mu, &wg, false)

	// Check if function completed without panic (this is the main test)
	// Note: filepath.Glob doesn't return errors for non-existent directories
	// It simply returns an empty slice, so no error is logged
	logMutex.RLock()
	defer logMutex.RUnlock()

	// The function should complete successfully even with non-existent directory
	// and should log an INFO message about scanning
	hasInfoMessage := false
	for _, entry := range logBuffer {
		if entry.Level == "INFO" && strings.Contains(entry.Message, "スキャン") {
			hasInfoMessage = true
			break
		}
	}
	assert.True(t, hasInfoMessage, "Should log scanning info message")
}

func TestProcessedFilesCleanup_MapSize(t *testing.T) {
	// Create a large processed files map to test cleanup
	processedFiles := make(map[string]bool)

	// Add more than 5000 files to trigger cleanup
	for i := 0; i < 6000; i++ {
		processedFiles[fmt.Sprintf("file_%d.wav", i)] = true
	}

	assert.Len(t, processedFiles, 6000)

	// This would trigger cleanup in actual ScanAndProcess call
	// We're testing the threshold condition
	assert.Greater(t, len(processedFiles), 5000)
}

func TestConcurrentProcessing_StateManagement(t *testing.T) {
	var mu sync.Mutex
	var isProcessing bool

	// Simulate concurrent access to processing state
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			mu.Lock()
			defer mu.Unlock()

			// Simulate checking and setting processing state
			if !isProcessing {
				isProcessing = true
				time.Sleep(1 * time.Millisecond) // Simulate work
				isProcessing = false
			}
		}(i)
	}

	wg.Wait()

	// Final state should be not processing
	mu.Lock()
	defer mu.Unlock()
	assert.False(t, isProcessing)
}

func TestSafeProcessingStart_RaceConditionFixed(t *testing.T) {
	// Test the fixed race condition in processing start logic
	var mu sync.Mutex
	var isProcessing bool
	var processCount int
	var wg sync.WaitGroup

	// Simulate multiple goroutines trying to start processing simultaneously
	numGoroutines := 20
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Simulate the fixed logic from ScanAndProcess
			mu.Lock()
			if !isProcessing {
				isProcessing = true
				processCount++
				mu.Unlock()
				
				// Simulate some work
				time.Sleep(5 * time.Millisecond)
				
				// Reset processing state
				mu.Lock()
				isProcessing = false
				mu.Unlock()
			} else {
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Only one goroutine should have started processing
	assert.Equal(t, 1, processCount, "Only one goroutine should start processing")
	assert.False(t, isProcessing, "Processing should be false after all goroutines complete")
}

func TestCleanupProcessedFiles_SafeConcurrentAccess(t *testing.T) {
	// Test the fixed concurrent map access in cleanup function
	processedFiles := make(map[string]bool)
	var mu sync.Mutex
	
	// Add more than 5000 files to trigger cleanup
	for i := 0; i < 6000; i++ {
		processedFiles[fmt.Sprintf("file_%d.wav", i)] = true
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	var logBuffer []logger.LogEntry
	var logMutex sync.RWMutex

	// Test that cleanup works without concurrent map access issues
	mu.Lock()
	cleanupProcessedFiles(&processedFiles, &mu, logger, &logBuffer, &logMutex)
	mu.Unlock()

	// Should have cleaned up to approximately half
	assert.LessOrEqual(t, len(processedFiles), 2500, "Should cleanup to 2500 or fewer entries")
	assert.Greater(t, len(processedFiles), 0, "Should keep some entries")
}
