package ui

import (
	"sync"
	"testing"
	"time"

	"github.com/infoHiroki/KoeMoji-Go/internal/config"
	"github.com/infoHiroki/KoeMoji-Go/internal/logger"
	"github.com/stretchr/testify/assert"
)

// createTestTUI creates a test TUI instance
func createTestTUI(t *testing.T) *TUI {
	cfg := config.GetDefaultConfig()
	logBuffer := make([]logger.LogEntry, 0, 12)
	logMutex := sync.RWMutex{}
	queuedFiles := make([]string, 0)
	mu := sync.Mutex{}

	// Note: We don't call NewTUI here because it initializes tview.Application
	// which requires a terminal. Instead, we create a minimal TUI for testing.
	tui := &TUI{
		config:      cfg,
		startTime:   time.Now(),
		logBuffer:   &logBuffer,
		logMutex:    &logMutex,
		queuedFiles: &queuedFiles,
		mu:          &mu,
	}

	return tui
}

func TestTUI_Creation(t *testing.T) {
	tui := createTestTUI(t)

	assert.NotNil(t, tui.config)
	assert.NotNil(t, tui.logBuffer)
	assert.NotNil(t, tui.logMutex)
	assert.NotNil(t, tui.queuedFiles)
	assert.NotNil(t, tui.mu)
	assert.False(t, tui.startTime.IsZero())
}

func TestTUI_SetLastScanTime(t *testing.T) {
	tui := createTestTUI(t)

	now := time.Now()
	tui.SetLastScanTime(now)

	assert.Equal(t, now, tui.lastScanTime)
}

func TestTUI_LogColorCode(t *testing.T) {
	tui := createTestTUI(t)

	tests := []struct {
		level    string
		expected string
	}{
		{"INFO", "blue"},
		{"PROC", "yellow"},
		{"DONE", "green"},
		{"ERROR", "red"},
		{"DEBUG", "gray"},
		{"UNKNOWN", "white"},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			color := tui.getLogColorCode(tt.level)
			assert.Equal(t, tt.expected, color)
		})
	}
}

func TestTUI_StatusTracking(t *testing.T) {
	tui := createTestTUI(t)

	// Test status tracking fields
	assert.Equal(t, 0, tui.inputCount)
	assert.Equal(t, 0, tui.outputCount)
	assert.Equal(t, 0, tui.archiveCount)
	assert.Empty(t, tui.processingFile)
	assert.False(t, tui.isProcessing)
	assert.False(t, tui.isRecording)
}

func TestTUI_CallbacksNil(t *testing.T) {
	tui := createTestTUI(t)

	// Initially, callbacks should be nil
	assert.Nil(t, tui.onScan)
	assert.Nil(t, tui.onRecord)
	assert.Nil(t, tui.onConfig)
	assert.Nil(t, tui.onLogs)
	assert.Nil(t, tui.onInput)
	assert.Nil(t, tui.onOutput)
	assert.Nil(t, tui.onQuit)
}

func TestTUI_SetCallbacks(t *testing.T) {
	tui := createTestTUI(t)

	// Create test callbacks
	scanCalled := false
	recordCalled := false
	configCalled := false
	logsCalled := false
	inputCalled := false
	outputCalled := false
	quitCalled := false

	tui.SetCallbacks(
		func() { scanCalled = true },
		func() { recordCalled = true },
		func() { configCalled = true },
		func() { logsCalled = true },
		func() { inputCalled = true },
		func() { outputCalled = true },
		func() { quitCalled = true },
	)

	// Test that callbacks are set
	assert.NotNil(t, tui.onScan)
	assert.NotNil(t, tui.onRecord)
	assert.NotNil(t, tui.onConfig)
	assert.NotNil(t, tui.onLogs)
	assert.NotNil(t, tui.onInput)
	assert.NotNil(t, tui.onOutput)
	assert.NotNil(t, tui.onQuit)

	// Test that callbacks work
	tui.onScan()
	assert.True(t, scanCalled)

	tui.onRecord()
	assert.True(t, recordCalled)

	tui.onConfig()
	assert.True(t, configCalled)

	tui.onLogs()
	assert.True(t, logsCalled)

	tui.onInput()
	assert.True(t, inputCalled)

	tui.onOutput()
	assert.True(t, outputCalled)

	tui.onQuit()
	assert.True(t, quitCalled)
}

func TestTUI_LogBuffer(t *testing.T) {
	tui := createTestTUI(t)

	// Add log entries
	tui.logMutex.Lock()
	*tui.logBuffer = append(*tui.logBuffer, logger.LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Message:   "Test info message",
	})
	*tui.logBuffer = append(*tui.logBuffer, logger.LogEntry{
		Timestamp: time.Now(),
		Level:     "ERROR",
		Message:   "Test error message",
	})
	tui.logMutex.Unlock()

	// Verify log buffer
	tui.logMutex.RLock()
	defer tui.logMutex.RUnlock()

	assert.Len(t, *tui.logBuffer, 2)
	assert.Equal(t, "INFO", (*tui.logBuffer)[0].Level)
	assert.Equal(t, "Test info message", (*tui.logBuffer)[0].Message)
	assert.Equal(t, "ERROR", (*tui.logBuffer)[1].Level)
	assert.Equal(t, "Test error message", (*tui.logBuffer)[1].Message)
}

func TestTUI_QueuedFiles(t *testing.T) {
	tui := createTestTUI(t)

	// Add files to queue
	tui.mu.Lock()
	*tui.queuedFiles = append(*tui.queuedFiles, "test1.wav", "test2.wav")
	tui.mu.Unlock()

	// Verify queue
	tui.mu.Lock()
	defer tui.mu.Unlock()

	assert.Len(t, *tui.queuedFiles, 2)
	assert.Equal(t, "test1.wav", (*tui.queuedFiles)[0])
	assert.Equal(t, "test2.wav", (*tui.queuedFiles)[1])
}

func TestTUI_ConcurrentAccess(t *testing.T) {
	tui := createTestTUI(t)

	var wg sync.WaitGroup
	numGoroutines := 10

	// Test concurrent log writing
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			tui.logMutex.Lock()
			*tui.logBuffer = append(*tui.logBuffer, logger.LogEntry{
				Timestamp: time.Now(),
				Level:     "INFO",
				Message:   "Concurrent message",
			})
			tui.logMutex.Unlock()
		}(i)
	}

	// Test concurrent queue operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			tui.mu.Lock()
			*tui.queuedFiles = append(*tui.queuedFiles, "file.wav")
			tui.mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify results
	tui.logMutex.RLock()
	logCount := len(*tui.logBuffer)
	tui.logMutex.RUnlock()

	tui.mu.Lock()
	queueCount := len(*tui.queuedFiles)
	tui.mu.Unlock()

	assert.Equal(t, numGoroutines, logCount)
	assert.Equal(t, numGoroutines, queueCount)
}

// Benchmark tests
func BenchmarkTUI_LogColorCode(b *testing.B) {
	tui := createTestTUI(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tui.getLogColorCode("INFO")
	}
}

func BenchmarkTUI_SetLastScanTime(b *testing.B) {
	tui := createTestTUI(&testing.T{})
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tui.SetLastScanTime(now)
	}
}
