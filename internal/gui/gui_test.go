package gui

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"github.com/infoHiroki/KoeMoji-Go/internal/config"
	"github.com/infoHiroki/KoeMoji-Go/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestGUIApp creates a test GUI application instance
func createTestGUIApp(t *testing.T) *GUIApp {
	tempDir := t.TempDir()
	
	cfg := config.GetDefaultConfig()
	cfg.InputDir = filepath.Join(tempDir, "input")
	cfg.OutputDir = filepath.Join(tempDir, "output")
	cfg.ArchiveDir = filepath.Join(tempDir, "archive")
	
	// Create directories
	for _, dir := range []string{cfg.InputDir, cfg.OutputDir, cfg.ArchiveDir} {
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	
	app := &GUIApp{
		Config:            cfg,
		configPath:        filepath.Join(tempDir, "config.json"),
		debugMode:         false,
		processedFiles:    make(map[string]bool),
		mu:                sync.Mutex{},
		logger:            log.New(os.Stdout, "", log.LstdFlags),
		startTime:         time.Now(),
		logBuffer:         make([]logger.LogEntry, 0, 12),
		logMutex:          sync.RWMutex{},
		queuedFiles:       make([]string, 0),
		ctx:               ctx,
		cancelFunc:        cancel,
		recordingDeviceMap: make(map[string]int),
	}
	
	return app
}

func TestGUIApp_Creation(t *testing.T) {
	app := createTestGUIApp(t)
	
	assert.NotNil(t, app.Config)
	assert.NotNil(t, app.logger)
	assert.NotNil(t, app.ctx)
	assert.NotEmpty(t, app.configPath)
	assert.False(t, app.debugMode)
	assert.False(t, app.isProcessing)
	assert.Empty(t, app.queuedFiles)
	assert.Empty(t, app.processingFile)
}

func TestGUIApp_FyneAppInitialization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping GUI test in short mode")
	}
	
	app := createTestGUIApp(t)
	
	// Initialize the Fyne app (using test app to avoid GUI display)
	testApp := test.NewApp()
	app.fyneApp = testApp
	
	assert.NotNil(t, app.fyneApp)
}

func TestGUIApp_LoadConfig(t *testing.T) {
	app := createTestGUIApp(t)
	
	// Create a test config file
	err := config.SaveConfig(app.Config, app.configPath)
	require.NoError(t, err)
	
	// Test loadConfig functionality (simulate the private method)
	loadedConfig, err := config.LoadConfig(app.configPath, app.logger)
	require.NoError(t, err)
	
	assert.Equal(t, app.Config.WhisperModel, loadedConfig.WhisperModel)
	assert.Equal(t, app.Config.Language, loadedConfig.Language)
}

func TestGUIApp_StatusUpdate(t *testing.T) {
	app := createTestGUIApp(t)
	
	// Test status counting functionality
	app.updateFileCounts()
	
	// Initially should be 0
	assert.Equal(t, 0, app.inputCount)
	assert.Equal(t, 0, app.outputCount)
	assert.Equal(t, 0, app.archiveCount)
	
	// Create test files
	testInputFile := filepath.Join(app.Config.InputDir, "test.wav")
	err := os.WriteFile(testInputFile, []byte("test audio data"), 0644)
	require.NoError(t, err)
	
	testOutputFile := filepath.Join(app.Config.OutputDir, "test.txt")
	err = os.WriteFile(testOutputFile, []byte("test transcription"), 0644)
	require.NoError(t, err)
	
	testArchiveFile := filepath.Join(app.Config.ArchiveDir, "archived.wav")
	err = os.WriteFile(testArchiveFile, []byte("archived audio"), 0644)
	require.NoError(t, err)
	
	// Update counts
	app.updateFileCounts()
	
	assert.Equal(t, 1, app.inputCount)
	assert.Equal(t, 1, app.outputCount)
	assert.Equal(t, 1, app.archiveCount)
}

func TestGUIApp_LogBuffer(t *testing.T) {
	app := createTestGUIApp(t)
	
	// Test log buffer functionality
	logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "Test info message")
	logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "Test error message")
	
	app.logMutex.RLock()
	defer app.logMutex.RUnlock()
	
	assert.Len(t, app.logBuffer, 2)
	assert.Equal(t, "INFO", app.logBuffer[0].Level)
	assert.Equal(t, "Test info message", app.logBuffer[0].Message)
	assert.Equal(t, "ERROR", app.logBuffer[1].Level)
	assert.Equal(t, "Test error message", app.logBuffer[1].Message)
}

func TestGUIApp_ProcessingState(t *testing.T) {
	app := createTestGUIApp(t)
	
	// Test processing state management
	assert.False(t, app.isProcessing)
	assert.Empty(t, app.processingFile)
	assert.Empty(t, app.queuedFiles)
	
	// Simulate adding files to queue
	app.mu.Lock()
	app.queuedFiles = append(app.queuedFiles, "test1.wav", "test2.wav")
	app.mu.Unlock()
	
	app.mu.Lock()
	queueLength := len(app.queuedFiles)
	app.mu.Unlock()
	
	assert.Equal(t, 2, queueLength)
	
	// Simulate processing
	app.mu.Lock()
	app.isProcessing = true
	app.processingFile = "test1.wav"
	app.mu.Unlock()
	
	app.mu.Lock()
	processing := app.isProcessing
	currentFile := app.processingFile
	app.mu.Unlock()
	
	assert.True(t, processing)
	assert.Equal(t, "test1.wav", currentFile)
}

func TestGUIApp_ConfigurationAccess(t *testing.T) {
	app := createTestGUIApp(t)
	
	// Test configuration access
	assert.NotNil(t, app.Config)
	assert.Equal(t, "large-v3", app.Config.WhisperModel)
	assert.Equal(t, "ja", app.Config.Language)
	assert.NotEmpty(t, app.Config.InputDir)
	assert.NotEmpty(t, app.Config.OutputDir)
	assert.NotEmpty(t, app.Config.ArchiveDir)
}

func TestGUIApp_DirectoryOperations(t *testing.T) {
	app := createTestGUIApp(t)
	
	// Test directory existence
	for _, dir := range []string{app.Config.InputDir, app.Config.OutputDir, app.Config.ArchiveDir} {
		_, err := os.Stat(dir)
		assert.NoError(t, err, "Directory should exist: %s", dir)
	}
	
	// Test directory permissions (write test)
	testFile := filepath.Join(app.Config.InputDir, "permission_test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	assert.NoError(t, err)
	
	// Clean up
	os.Remove(testFile)
}

func TestGUIApp_ConcurrentAccess(t *testing.T) {
	app := createTestGUIApp(t)
	
	// Test concurrent access to shared resources
	var wg sync.WaitGroup
	numGoroutines := 10
	
	// Test concurrent log writing
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "Concurrent message %d", id)
		}(i)
	}
	
	// Test concurrent processing state access
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			app.mu.Lock()
			app.queuedFiles = append(app.queuedFiles, fmt.Sprintf("file_%d.wav", id))
			app.mu.Unlock()
		}(i)
	}
	
	wg.Wait()
	
	// Verify results
	app.logMutex.RLock()
	logCount := len(app.logBuffer)
	app.logMutex.RUnlock()
	
	app.mu.Lock()
	queueCount := len(app.queuedFiles)
	app.mu.Unlock()
	
	assert.Equal(t, numGoroutines, logCount)
	assert.Equal(t, numGoroutines, queueCount)
}

func TestGUIApp_Cleanup(t *testing.T) {
	app := createTestGUIApp(t)
	
	// Test cleanup functionality
	assert.NotNil(t, app.ctx)
	assert.NotNil(t, app.cancelFunc)
	
	// Add some data
	logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "Test message")
	app.mu.Lock()
	app.queuedFiles = append(app.queuedFiles, "test.wav")
	app.mu.Unlock()
	
	// Cancel context (simulating app shutdown)
	app.cancelFunc()
	
	// Context should be cancelled
	select {
	case <-app.ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Context was not cancelled")
	}
}

// Benchmark tests for performance
func BenchmarkGUIApp_LogWrite(b *testing.B) {
	app := createTestGUIApp(&testing.T{})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "Benchmark message %d", i)
	}
}

func BenchmarkGUIApp_QueueOperation(b *testing.B) {
	app := createTestGUIApp(&testing.T{})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.mu.Lock()
		app.queuedFiles = append(app.queuedFiles, fmt.Sprintf("bench_%d.wav", i))
		app.mu.Unlock()
	}
}