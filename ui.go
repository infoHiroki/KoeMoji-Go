package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
	msg := app.getMessages()

	status := "🟢 " + msg.Active
	if app.isProcessing {
		status = "🟡 " + msg.Processing
	}

	uptime := time.Since(app.startTime)

	fmt.Println("=== KoeMoji-Go v" + version + " ===")

	app.mu.Lock()
	queueCount := len(app.queuedFiles)
	processingDisplay := msg.None
	if app.processingFile != "" {
		processingDisplay = app.processingFile
	}
	app.mu.Unlock()

	fmt.Printf("%s | %s: %d | %s: %s\n",
		status, msg.Queue, queueCount, msg.Processing, processingDisplay)
	fmt.Printf("📁 %s: %d → %s: %d → %s: %d\n",
		msg.Input, app.inputCount, msg.Output, app.outputCount, msg.Archive, app.archiveCount)

	lastScanStr := msg.Never
	nextScanStr := msg.Soon
	if !app.lastScanTime.IsZero() {
		lastScanStr = app.lastScanTime.Format("15:04:05")
		nextScan := app.lastScanTime.Add(time.Duration(app.config.ScanIntervalMinutes) * time.Minute)
		nextScanStr = nextScan.Format("15:04:05")
	}

	fmt.Printf("⏰ %s: %s | %s: %s | %s: %s\n",
		msg.Last, lastScanStr, msg.Next, nextScanStr, msg.Uptime, app.formatDuration(uptime))
	fmt.Println()
}

func (app *App) displayRealtimeLogs() {
	app.logMutex.RLock()
	defer app.logMutex.RUnlock()
	msg := app.getMessages()

	for _, entry := range app.logBuffer {
		color := app.getLogColor(entry.Level)
		timestamp := entry.Timestamp.Format("15:04:05")
		
		// Convert log level to localized version
		localizedLevel := entry.Level
		switch entry.Level {
		case "INFO":
			localizedLevel = msg.LogInfo
		case "PROC":
			localizedLevel = msg.LogProc
		case "DONE":
			localizedLevel = msg.LogDone
		case "ERROR":
			localizedLevel = msg.LogError
		case "DEBUG":
			localizedLevel = msg.LogDebug
		}

		if color != "" {
			fmt.Printf("%s%-5s%s %s %s\n", color, localizedLevel, ColorReset, timestamp, entry.Message)
		} else {
			fmt.Printf("[%-5s] %s %s\n", localizedLevel, timestamp, entry.Message)
		}
	}

	// Fill remaining lines to maintain 12-line display
	for i := len(app.logBuffer); i < 12; i++ {
		fmt.Println()
	}
}

func (app *App) displayCommands() {
	msg := app.getMessages()
	fmt.Printf("c=%s l=%s s=%s i=%s o=%s q=%s\n", msg.ConfigCmd, msg.LogsCmd, msg.ScanCmd, msg.InputDirCmd, msg.OutputDirCmd, msg.QuitCmd)
	fmt.Print("> ")
}

func (app *App) displayLogs() {
	msg := app.getMessages()
	
	if _, err := os.Stat("koemoji.log"); os.IsNotExist(err) {
		fmt.Println(msg.FileNotFound)
		return
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "koemoji.log")
	case "darwin":
		cmd = exec.Command("open", "koemoji.log")
	default:
		fmt.Println(msg.UnsupportedOS)
		return
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf(msg.LogFileError, err)
	}
}

func (app *App) openDirectory(dirPath string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", dirPath)
	case "darwin":
		cmd = exec.Command("open", dirPath)
	default:
		return fmt.Errorf("opening directories not supported on this platform")
	}

	return cmd.Start()
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

func (app *App) formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	
	if hours > 0 {
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	} else {
		return fmt.Sprintf("%ds", seconds)
	}
}

func (app *App) updateFileCounts() {
	// Count files in each directory
	if entries, err := os.ReadDir(app.config.InputDir); err == nil {
		app.inputCount = 0
		for _, entry := range entries {
			if !entry.IsDir() && isAudioFile(entry.Name()) {
				app.inputCount++
			}
		}
	}
	
	if entries, err := os.ReadDir(app.config.OutputDir); err == nil {
		app.outputCount = len(entries)
	}
	
	if entries, err := os.ReadDir(app.config.ArchiveDir); err == nil {
		app.archiveCount = len(entries)
	}
}

func isAudioFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	audioExts := []string{".mp3", ".wav", ".m4a", ".flac", ".ogg", ".aac", ".mp4", ".mov", ".avi"}
	for _, audioExt := range audioExts {
		if ext == audioExt {
			return true
		}
	}
	return false
}
