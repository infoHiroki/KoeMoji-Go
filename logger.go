package main

import (
	"fmt"
	"time"
)

// Logging functions
func (app *App) addToLogBuffer(level, message string) {
	app.logMutex.Lock()
	defer app.logMutex.Unlock()

	entry := LogEntry{
		Level:     level,
		Message:   message,
		Timestamp: time.Now(),
	}

	app.logBuffer = append(app.logBuffer, entry)
	if len(app.logBuffer) > 12 {
		app.logBuffer = app.logBuffer[1:]
	}
}

func (app *App) logInfo(format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	app.logger.Printf("[INFO] %s", message)
	app.addToLogBuffer("INFO", message)
	app.refreshDisplay()
}

func (app *App) logError(format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	app.logger.Printf("[ERROR] %s", message)
	app.addToLogBuffer("ERROR", message)
	app.refreshDisplay()
}

func (app *App) logDebug(format string, v ...any) {
	if app.debugMode {
		message := fmt.Sprintf(format, v...)
		app.logger.Printf("[DEBUG] %s", message)
		app.addToLogBuffer("DEBUG", message)
		app.refreshDisplay()
	}
}

func (app *App) logProc(format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	app.logger.Printf("[PROC] %s", message)
	app.addToLogBuffer("PROC", message)
	app.refreshDisplay()
}

func (app *App) logDone(format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	app.logger.Printf("[DONE] %s", message)
	app.addToLogBuffer("DONE", message)
	app.refreshDisplay()
}
