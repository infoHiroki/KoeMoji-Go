package logger

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type LogEntry struct {
	Level     string
	Message   string
	Timestamp time.Time
}

// Logging functions
func addToLogBuffer(logBuffer *[]LogEntry, logMutex *sync.RWMutex, level, message string) {
	logMutex.Lock()
	defer logMutex.Unlock()

	entry := LogEntry{
		Level:     level,
		Message:   message,
		Timestamp: time.Now(),
	}

	*logBuffer = append(*logBuffer, entry)
	if len(*logBuffer) > 12 {
		*logBuffer = (*logBuffer)[1:]
	}
}

func LogInfo(logger *log.Logger, logBuffer *[]LogEntry, logMutex *sync.RWMutex, format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	logger.Printf("[INFO] %s", message)
	addToLogBuffer(logBuffer, logMutex, "INFO", message)
}

func LogError(logger *log.Logger, logBuffer *[]LogEntry, logMutex *sync.RWMutex, format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	logger.Printf("[ERROR] %s", message)
	addToLogBuffer(logBuffer, logMutex, "ERROR", message)
}

func LogDebug(logger *log.Logger, logBuffer *[]LogEntry, logMutex *sync.RWMutex, debugMode bool, format string, v ...any) {
	if debugMode {
		message := fmt.Sprintf(format, v...)
		logger.Printf("[DEBUG] %s", message)
		addToLogBuffer(logBuffer, logMutex, "DEBUG", message)
	}
}

func LogProc(logger *log.Logger, logBuffer *[]LogEntry, logMutex *sync.RWMutex, format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	logger.Printf("[PROC] %s", message)
	addToLogBuffer(logBuffer, logMutex, "PROC", message)
}

func LogDone(logger *log.Logger, logBuffer *[]LogEntry, logMutex *sync.RWMutex, format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	logger.Printf("[DONE] %s", message)
	addToLogBuffer(logBuffer, logMutex, "DONE", message)
}