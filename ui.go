package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// UI functions
func (app *App) refreshDisplay() {
	if app.config == nil || app.config.UIMode != "enhanced" {
		return
	}

	// Clear screen and move cursor to top
	fmt.Print("\033[2J\033[H")

	app.displayHeader()
	app.displayRealtimeLogs()
	app.displayCommands()
}

func (app *App) displayHeader() {
	app.updateFileCounts()

	status := "🟢 Active"
	if app.isProcessing {
		status = "🟡 Processing"
	}

	uptime := time.Since(app.startTime)

	fmt.Println("=== KoeMoji-Go v" + version + " ===")

	app.mu.Lock()
	queueCount := len(app.queuedFiles)
	processingDisplay := "None"
	if app.processingFile != "" {
		processingDisplay = app.processingFile
	}
	app.mu.Unlock()

	fmt.Printf("%s | Queue: %d | Processing: %s\n",
		status, queueCount, processingDisplay)
	fmt.Printf("📁 Input: %d → Output: %d → Archive: %d\n",
		app.inputCount, app.outputCount, app.archiveCount)

	lastScanStr := "Never"
	nextScanStr := "Soon"
	if !app.lastScanTime.IsZero() {
		lastScanStr = app.lastScanTime.Format("15:04:05")
		nextScan := app.lastScanTime.Add(time.Duration(app.config.ScanIntervalMinutes) * time.Minute)
		nextScanStr = nextScan.Format("15:04:05")
	}

	fmt.Printf("⏰ Last: %s | Next: %s | Uptime: %s\n",
		lastScanStr, nextScanStr, app.formatDuration(uptime))
	fmt.Println()
}

func (app *App) displayRealtimeLogs() {
	app.logMutex.RLock()
	defer app.logMutex.RUnlock()

	for _, entry := range app.logBuffer {
		color := app.getLogColor(entry.Level)
		timestamp := entry.Timestamp.Format("15:04:05")

		if color != "" {
			fmt.Printf("%s%-5s%s %s %s\n", color, entry.Level, ColorReset, timestamp, entry.Message)
		} else {
			fmt.Printf("[%-5s] %s %s\n", entry.Level, timestamp, entry.Message)
		}
	}

	// Fill remaining lines to maintain 12-line display
	for i := len(app.logBuffer); i < 12; i++ {
		fmt.Println()
	}
}

func (app *App) displayCommands() {
	fmt.Println("c=config l=logs s=scan q=quit")
	fmt.Print("> ")
}

func (app *App) displayLogs() {
	if _, err := os.Stat("koemoji.log"); os.IsNotExist(err) {
		fmt.Println("Log file not found")
		return
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "koemoji.log")
	case "darwin":
		cmd = exec.Command("open", "koemoji.log")
	default:
		fmt.Println("Log viewing not supported on this platform")
		return
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
	}
}

// Utility functions
func (app *App) supportsColor() bool {
	if app.config == nil || !app.config.UseColors {
		return false
	}

	if runtime.GOOS == "windows" {
		return os.Getenv("WT_SESSION") != "" || os.Getenv("ConEmuPID") != ""
	}

	term := os.Getenv("TERM")
	return term != "" && term != "dumb"
}

func (app *App) getLogColor(level string) string {
	if !app.supportsColor() {
		return ""
	}

	switch level {
	case "INFO":
		return ColorBlue
	case "PROC":
		return ColorYellow
	case "DONE":
		return ColorGreen
	case "ERROR":
		return ColorRed
	case "DEBUG":
		return ColorGray
	default:
		return ""
	}
}
