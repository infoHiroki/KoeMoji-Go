package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/hirokitakamura/koemoji-go/internal/config"
	"github.com/hirokitakamura/koemoji-go/internal/logger"
)

// Color constants
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m" // ERROR
	ColorGreen  = "\033[32m" // DONE
	ColorYellow = "\033[33m" // PROC
	ColorBlue   = "\033[34m" // INFO
	ColorGray   = "\033[37m" // DEBUG
)

// UI functions
func RefreshDisplay(config *config.Config, startTime, lastScanTime time.Time, logBuffer *[]logger.LogEntry,
	logMutex *sync.RWMutex, inputCount, outputCount, archiveCount int, queuedFiles *[]string,
	processingFile string, isProcessing bool, mu *sync.Mutex,
	isRecording bool, recordingStartTime time.Time) {

	if config == nil {
		return
	}

	// Clear screen and move cursor to top
	fmt.Print("\033[2J\033[H")

	displayHeader(config, startTime, lastScanTime, inputCount, outputCount, archiveCount,
		queuedFiles, processingFile, isProcessing, mu, isRecording, recordingStartTime)
	displayRealtimeLogs(config, logBuffer, logMutex)
	displayCommands(config)
}

func displayHeader(config *config.Config, startTime, lastScanTime time.Time, inputCount, outputCount, archiveCount int,
	queuedFiles *[]string, processingFile string, isProcessing bool, mu *sync.Mutex,
	isRecording bool, recordingStartTime time.Time) {

	updateFileCounts(config, &inputCount, &outputCount, &archiveCount)
	msg := GetMessages(config)

	status := "🟢 " + msg.Active
	if isProcessing {
		status = "🟡 " + msg.Processing
	}

	uptime := time.Since(startTime)

	fmt.Println("=== KoeMoji-Go ===")

	mu.Lock()
	queueCount := len(*queuedFiles)
	processingDisplay := msg.None
	if processingFile != "" {
		processingDisplay = processingFile
	}
	mu.Unlock()

	fmt.Printf("%s | %s: %d | %s: %s\n",
		status, msg.Queue, queueCount, msg.Processing, processingDisplay)
	fmt.Printf("📁 %s: %d → %s: %d → %s: %d\n",
		msg.Input, inputCount, msg.Output, outputCount, msg.Archive, archiveCount)

	// Recording status
	if isRecording {
		elapsed := time.Since(recordingStartTime)
		fmt.Printf("🔴 %s - %s\n", msg.Recording, formatDuration(elapsed))
	}

	lastScanStr := msg.Never
	nextScanStr := msg.Soon
	if !lastScanTime.IsZero() {
		lastScanStr = lastScanTime.Format("15:04:05")
		nextScan := lastScanTime.Add(time.Duration(config.ScanIntervalMinutes) * time.Minute)
		nextScanStr = nextScan.Format("15:04:05")
	}

	fmt.Printf("⏰ %s: %s | %s: %s | %s: %s\n",
		msg.Last, lastScanStr, msg.Next, nextScanStr, msg.Uptime, formatDuration(uptime))
	fmt.Println()
}

func displayRealtimeLogs(config *config.Config, logBuffer *[]logger.LogEntry, logMutex *sync.RWMutex) {
	logMutex.RLock()
	defer logMutex.RUnlock()
	msg := GetMessages(config)

	for _, entry := range *logBuffer {
		color := getLogColor(config, entry.Level)
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
			fmt.Printf("[%s] %s %s\n", localizedLevel, timestamp, entry.Message)
		}
	}

	// Fill remaining lines to maintain 12-line display
	for i := len(*logBuffer); i < 12; i++ {
		fmt.Println()
	}
}

func displayCommands(config *config.Config) {
	msg := GetMessages(config)
	fmt.Printf("c=%s l=%s s=%s i=%s o=%s r=%s q=%s\n", msg.ConfigCmd, msg.LogsCmd, msg.ScanCmd, msg.InputDirCmd, msg.OutputDirCmd, msg.RecordCmd, msg.QuitCmd)
	fmt.Print("> ")
}

func DisplayLogs(config *config.Config) {
	msg := GetMessages(config)

	if _, err := os.Stat("koemoji.log"); os.IsNotExist(err) {
		fmt.Println(msg.FileNotFound)
		return
	}

	switch runtime.GOOS {
	case "windows":
		// Try to open with PowerShell to jump to end, fallback to regular notepad
		powershellCmd := createCommand("powershell", "-Command",
			`notepad koemoji.log; Start-Sleep -Milliseconds 500; Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.SendKeys]::SendWait("^{END}")`)
		if err := powershellCmd.Run(); err != nil {
			// Fallback to regular notepad if PowerShell fails
			fallbackCmd := createCommand("notepad", "koemoji.log")
			if fallbackErr := fallbackCmd.Run(); fallbackErr != nil {
				fmt.Printf(msg.LogFileError, fallbackErr)
			}
		}
	case "darwin":
		// Get absolute path for AppleScript
		absPath, err := filepath.Abs("koemoji.log")
		if err != nil {
			absPath = "koemoji.log"
		}

		// Try to open with AppleScript to jump to end, fallback to regular open
		appleScriptCmd := createCommand("osascript", "-e",
			fmt.Sprintf(`tell application "TextEdit" to open POSIX file "%s"`, absPath),
			"-e", `tell application "TextEdit" to goto paragraph -1`)
		if err := appleScriptCmd.Run(); err != nil {
			// Fallback to regular open if AppleScript fails
			fallbackCmd := createCommand("open", "koemoji.log")
			if fallbackErr := fallbackCmd.Run(); fallbackErr != nil {
				fmt.Printf(msg.LogFileError, fallbackErr)
			}
		}
	default:
		fmt.Println(msg.UnsupportedOS)
		return
	}
}

func OpenDirectory(dirPath string) error {
	// Convert to absolute path for Windows explorer compatibility
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		absPath = dirPath // Fallback to original path
	}
	
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// Explorer doesn't work with HideWindow flag, use regular exec.Command
		cmd = exec.Command("explorer", absPath)
	case "darwin":
		cmd = createCommand("open", absPath)
	default:
		return fmt.Errorf("opening directories not supported on this platform")
	}

	return cmd.Start()
}

// Utility functions
func supportsColor(config *config.Config) bool {
	if config == nil || !config.UseColors {
		return false
	}

	if runtime.GOOS == "windows" {
		return true // Windows 10以降は強制有効
	}

	term := os.Getenv("TERM")
	return term != "" && term != "dumb"
}

func getLogColor(config *config.Config, level string) string {
	if !supportsColor(config) {
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

func formatDuration(d time.Duration) string {
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

func updateFileCounts(config *config.Config, inputCount, outputCount, archiveCount *int) {
	// Count files in each directory
	if entries, err := os.ReadDir(config.InputDir); err == nil {
		*inputCount = 0
		for _, entry := range entries {
			if !entry.IsDir() && IsAudioFile(entry.Name()) {
				(*inputCount)++
			}
		}
	}

	if entries, err := os.ReadDir(config.OutputDir); err == nil {
		*outputCount = len(entries)
	}

	if entries, err := os.ReadDir(config.ArchiveDir); err == nil {
		*archiveCount = len(entries)
	}
}

func IsAudioFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	audioExts := []string{".mp3", ".wav", ".m4a", ".flac", ".ogg", ".aac", ".mp4", ".mov", ".avi"}
	for _, audioExt := range audioExts {
		if ext == audioExt {
			return true
		}
	}
	return false
}
