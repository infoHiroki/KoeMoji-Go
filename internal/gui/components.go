package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"github.com/hirokitakamura/koemoji-go/internal/logger"
	"github.com/hirokitakamura/koemoji-go/internal/processor"
	"github.com/hirokitakamura/koemoji-go/internal/recorder"
	"github.com/hirokitakamura/koemoji-go/internal/ui"
	"github.com/hirokitakamura/koemoji-go/internal/whisper"
)

// startPeriodicUpdate starts the 5-second periodic UI update
func (app *GUIApp) startPeriodicUpdate() {
	// Initialize dependencies once
	processor.EnsureDirectories(app.Config, nil)
	whisper.EnsureDependencies(app.Config, nil, &app.logBuffer, &app.logMutex, app.debugMode)

	// Start file processing
	go processor.StartProcessing(app.Config, nil, &app.logBuffer, &app.logMutex,
		&app.lastScanTime, &app.queuedFiles, &app.processingFile, &app.isProcessing,
		&app.processedFiles, &app.mu, nil, app.debugMode)

	// Start periodic updates in a goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// Use fyne.Do to safely update UI from goroutine
			fyne.Do(func() {
				app.updateUI()
			})
		}
	}()
}

// updateUI updates all UI components with current data
func (app *GUIApp) updateUI() {
	if app.statusLabel == nil || app.filesLabel == nil || app.timingLabel == nil || app.logText == nil {
		return
	}

	msg := ui.GetMessages(app.Config)

	// Update file counts
	app.updateFileCounts()

	// Update status label
	status := "🟢 " + msg.Active
	if app.isProcessing {
		status = "🟡 " + msg.Processing
	}

	app.mu.Lock()
	queueCount := len(app.queuedFiles)
	processingDisplay := msg.None
	if app.processingFile != "" {
		processingDisplay = app.processingFile
	}
	app.mu.Unlock()

	statusText := fmt.Sprintf("%s | %s: %d | %s: %s",
		status, msg.Queue, queueCount, msg.Processing, processingDisplay)

	// Update files label
	filesText := fmt.Sprintf("📁 %s: %d → %s: %d → %s: %d",
		msg.Input, app.inputCount, msg.Output, app.outputCount, msg.Archive, app.archiveCount)

	// Update timing label with recording status
	uptime := time.Since(app.startTime)
	lastScanStr := msg.Never
	nextScanStr := msg.Soon
	if !app.lastScanTime.IsZero() {
		lastScanStr = app.lastScanTime.Format("15:04:05")
		nextScan := app.lastScanTime.Add(time.Duration(app.Config.ScanIntervalMinutes) * time.Minute)
		nextScanStr = nextScan.Format("15:04:05")
	}

	timingText := fmt.Sprintf("⏰ %s: %s | %s: %s | %s: %s",
		msg.Last, lastScanStr, msg.Next, nextScanStr, msg.Uptime, formatDuration(uptime))

	// Add recording status if recording
	if app.isRecording {
		elapsed := time.Since(app.recordingStartTime)
		timingText += fmt.Sprintf(" | 🔴 %s: %s", msg.Recording, formatDuration(elapsed))
	}

	// Update UI elements on main thread
	app.statusLabel.SetText(statusText)
	app.filesLabel.SetText(filesText)
	app.timingLabel.SetText(timingText)

	// Update log display
	app.updateLogDisplay()
}

// updateFileCounts updates the file count fields
func (app *GUIApp) updateFileCounts() {
	// Count files in each directory (similar to ui/ui.go)
	if entries, err := os.ReadDir(app.Config.InputDir); err == nil {
		app.inputCount = 0
		for _, entry := range entries {
			if !entry.IsDir() && ui.IsAudioFile(entry.Name()) {
				app.inputCount++
			}
		}
	}

	if entries, err := os.ReadDir(app.Config.OutputDir); err == nil {
		app.outputCount = len(entries)
	}

	if entries, err := os.ReadDir(app.Config.ArchiveDir); err == nil {
		app.archiveCount = len(entries)
	}
}

// updateLogDisplay updates the log viewer with recent entries
func (app *GUIApp) updateLogDisplay() {
	app.logMutex.RLock()
	defer app.logMutex.RUnlock()

	if len(app.logBuffer) == 0 {
		app.logText.ParseMarkdown("**Waiting for log entries...**")
		return
	}

	// Build log text from buffer
	var logText string
	for _, entry := range app.logBuffer {
		timestamp := entry.Timestamp.Format("15:04:05")
		// Format: [LEVEL] timestamp message
		logText += fmt.Sprintf("**[%s]** %s %s\n\n", entry.Level, timestamp, entry.Message)
	}

	app.logText.ParseMarkdown(logText)
}

// formatDuration formats a duration for display (copied from ui/ui.go)
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

// Button action handlers

// onConfigPressed handles the config button press
func (app *GUIApp) onConfigPressed() {
	// Show the configuration dialog
	app.showConfigDialog()

	// Log the action
	logger.LogInfo(nil, &app.logBuffer, &app.logMutex, "Configuration dialog opened")
}

// onLogsPressed handles the logs button press
func (app *GUIApp) onLogsPressed() {
	// Open log file using existing UI function
	ui.DisplayLogs(app.Config)

	// Log the action
	logger.LogInfo(nil, &app.logBuffer, &app.logMutex, "Log file opened")
}

// onScanPressed handles the scan button press
func (app *GUIApp) onScanPressed() {
	logger.LogInfo(nil, &app.logBuffer, &app.logMutex, "Manual scan triggered")

	// Use existing sync.WaitGroup reference if available, or create minimal scan
	processor.ScanAndProcess(app.Config, nil, &app.logBuffer, &app.logMutex,
		&app.lastScanTime, &app.queuedFiles, &app.processingFile, &app.isProcessing,
		&app.processedFiles, &app.mu, nil, app.debugMode)
}

// onInputDirPressed handles the input directory button press
func (app *GUIApp) onInputDirPressed() {
	if err := ui.OpenDirectory(app.Config.InputDir); err != nil {
		logger.LogError(nil, &app.logBuffer, &app.logMutex, "Failed to open input directory: %v", err)
	}
}

// onOutputDirPressed handles the output directory button press
func (app *GUIApp) onOutputDirPressed() {
	if err := ui.OpenDirectory(app.Config.OutputDir); err != nil {
		logger.LogError(nil, &app.logBuffer, &app.logMutex, "Failed to open output directory: %v", err)
	}
}


// onRecordPressed handles the record button press
func (app *GUIApp) onRecordPressed() {
	if app.isRecording {
		// Stop recording
		app.stopRecording()
	} else {
		// Start recording
		app.startRecording()
	}
}

// startRecording starts audio recording
func (app *GUIApp) startRecording() {
	// Initialize recorder if not already done
	if app.recorder == nil {
		var err error
		if app.Config.RecordingDeviceID == -1 {
			app.recorder, err = recorder.NewRecorder()
		} else {
			app.recorder, err = recorder.NewRecorderWithDevice(app.Config.RecordingDeviceID)
		}

		if err != nil {
			logger.LogError(nil, &app.logBuffer, &app.logMutex, "録音の初期化に失敗: %v", err)
			return
		}
	}

	// Start recording
	err := app.recorder.Start()
	if err != nil {
		logger.LogError(nil, &app.logBuffer, &app.logMutex, "録音の開始に失敗: %v", err)
		return
	}

	app.isRecording = true
	app.recordingStartTime = time.Now()
	logger.LogInfo(nil, &app.logBuffer, &app.logMutex, "録音を開始しました")

	// Update button appearance
	app.updateRecordingUI()
}

// stopRecording stops audio recording
func (app *GUIApp) stopRecording() {
	if app.recorder == nil {
		logger.LogError(nil, &app.logBuffer, &app.logMutex, "録音が初期化されていません")
		return
	}

	// Stop recording
	err := app.recorder.Stop()
	if err != nil {
		logger.LogError(nil, &app.logBuffer, &app.logMutex, "録音の停止に失敗: %v", err)
		return
	}

	// Generate filename with current timestamp
	now := time.Now()
	filename := fmt.Sprintf("recording_%s.wav", now.Format("20060102_1504"))

	// Save to input directory
	outputPath := filepath.Join(app.Config.InputDir, filename)
	err = app.recorder.SaveToFile(outputPath)
	if err != nil {
		logger.LogError(nil, &app.logBuffer, &app.logMutex, "録音ファイルの保存に失敗: %v", err)
		return
	}

	app.isRecording = false
	duration := time.Since(app.recordingStartTime)
	logger.LogInfo(nil, &app.logBuffer, &app.logMutex, "録音を停止しました: %s (時間: %s)", filename, duration.Round(time.Second))

	// Update button appearance
	app.updateRecordingUI()
}

// updateRecordingUI updates the recording-related UI elements
func (app *GUIApp) updateRecordingUI() {
	if app.recordButton == nil {
		return
	}

	msg := ui.GetMessages(app.Config)
	if app.isRecording {
		app.recordButton.SetText("🔴 停止")
		app.recordButton.Importance = widget.DangerImportance
	} else {
		app.recordButton.SetText("🎤 " + msg.RecordCmd)
		app.recordButton.Importance = widget.WarningImportance
	}
	app.recordButton.Refresh()
}

// onQuitPressed handles the quit button press
func (app *GUIApp) onQuitPressed() {
	// Stop recording if in progress before quitting
	if app.isRecording {
		app.stopRecording()
	}
	// Immediate exit as per design document
	app.fyneApp.Quit()
}
