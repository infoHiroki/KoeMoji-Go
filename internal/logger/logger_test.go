package logger

import (
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLogEntry_Creation(t *testing.T) {
	entry := LogEntry{
		Level:     "INFO",
		Message:   "test message",
		Timestamp: time.Now(),
	}

	assert.Equal(t, "INFO", entry.Level)
	assert.Equal(t, "test message", entry.Message)
	assert.False(t, entry.Timestamp.IsZero())
}

func TestLogInfo(t *testing.T) {
	var logBuffer []LogEntry
	var logMutex sync.RWMutex
	logger := log.New(os.Stdout, "", log.LstdFlags)

	LogInfo(logger, &logBuffer, &logMutex, "test info message")

	logMutex.RLock()
	defer logMutex.RUnlock()

	assert.Len(t, logBuffer, 1)
	assert.Equal(t, "INFO", logBuffer[0].Level)
	assert.Equal(t, "test info message", logBuffer[0].Message)
}

func TestLogError(t *testing.T) {
	var logBuffer []LogEntry
	var logMutex sync.RWMutex
	logger := log.New(os.Stdout, "", log.LstdFlags)

	LogError(logger, &logBuffer, &logMutex, "test error: %s", "failure")

	logMutex.RLock()
	defer logMutex.RUnlock()

	assert.Len(t, logBuffer, 1)
	assert.Equal(t, "ERROR", logBuffer[0].Level)
	assert.Equal(t, "test error: failure", logBuffer[0].Message)
}

func TestLogDebug_EnabledAndDisabled(t *testing.T) {
	var logBuffer []LogEntry
	var logMutex sync.RWMutex
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Debug disabled
	LogDebug(logger, &logBuffer, &logMutex, false, "debug message")
	assert.Len(t, logBuffer, 0)

	// Debug enabled
	LogDebug(logger, &logBuffer, &logMutex, true, "debug message")

	logMutex.RLock()
	defer logMutex.RUnlock()

	assert.Len(t, logBuffer, 1)
	assert.Equal(t, "DEBUG", logBuffer[0].Level)
	assert.Equal(t, "debug message", logBuffer[0].Message)
}

func TestLogProc(t *testing.T) {
	var logBuffer []LogEntry
	var logMutex sync.RWMutex
	logger := log.New(os.Stdout, "", log.LstdFlags)

	LogProc(logger, &logBuffer, &logMutex, "processing %s", "file.wav")

	logMutex.RLock()
	defer logMutex.RUnlock()

	assert.Len(t, logBuffer, 1)
	assert.Equal(t, "PROC", logBuffer[0].Level)
	assert.Equal(t, "processing file.wav", logBuffer[0].Message)
}

func TestLogDone(t *testing.T) {
	var logBuffer []LogEntry
	var logMutex sync.RWMutex
	logger := log.New(os.Stdout, "", log.LstdFlags)

	LogDone(logger, &logBuffer, &logMutex, "completed %s", "task")

	logMutex.RLock()
	defer logMutex.RUnlock()

	assert.Len(t, logBuffer, 1)
	assert.Equal(t, "DONE", logBuffer[0].Level)
	assert.Equal(t, "completed task", logBuffer[0].Message)
}

func TestCircularBuffer_12EntryLimit(t *testing.T) {
	var logBuffer []LogEntry
	var logMutex sync.RWMutex
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Add 15 entries (exceeds 12 limit)
	for i := 0; i < 15; i++ {
		LogInfo(logger, &logBuffer, &logMutex, "message %d", i)
	}

	logMutex.RLock()
	defer logMutex.RUnlock()

	// Should only keep latest 12 entries
	assert.Len(t, logBuffer, 12)

	// First entry should be message 3 (0, 1, 2 were removed)
	assert.Equal(t, "message 3", logBuffer[0].Message)

	// Last entry should be message 14
	assert.Equal(t, "message 14", logBuffer[11].Message)
}

func TestLogBuffer_NilLogger(t *testing.T) {
	var logBuffer []LogEntry
	var logMutex sync.RWMutex

	// Should not panic with nil logger
	LogInfo(nil, &logBuffer, &logMutex, "test message")

	logMutex.RLock()
	defer logMutex.RUnlock()

	assert.Len(t, logBuffer, 1)
	assert.Equal(t, "INFO", logBuffer[0].Level)
	assert.Equal(t, "test message", logBuffer[0].Message)
}

func TestLogBuffer_ConcurrentAccess(t *testing.T) {
	var logBuffer []LogEntry
	var logMutex sync.RWMutex
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Test concurrent logging
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			LogInfo(logger, &logBuffer, &logMutex, "concurrent message %d", id)
		}(i)
	}

	wg.Wait()

	logMutex.RLock()
	defer logMutex.RUnlock()

	// All messages should be logged
	assert.Len(t, logBuffer, 10)

	// All entries should have INFO level
	for _, entry := range logBuffer {
		assert.Equal(t, "INFO", entry.Level)
		assert.Contains(t, entry.Message, "concurrent message")
	}
}

func TestAddToLogBuffer_Timestamp(t *testing.T) {
	var logBuffer []LogEntry
	var logMutex sync.RWMutex

	before := time.Now()
	addToLogBuffer(&logBuffer, &logMutex, "TEST", "timestamp test")
	after := time.Now()

	logMutex.RLock()
	defer logMutex.RUnlock()

	assert.Len(t, logBuffer, 1)

	timestamp := logBuffer[0].Timestamp
	assert.True(t, timestamp.After(before) || timestamp.Equal(before))
	assert.True(t, timestamp.Before(after) || timestamp.Equal(after))
}
