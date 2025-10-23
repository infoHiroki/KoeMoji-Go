package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hirokitakamura/koemoji-go/internal/logger"
	"github.com/hirokitakamura/koemoji-go/internal/processor"
	"github.com/hirokitakamura/koemoji-go/internal/recorder"
	"github.com/hirokitakamura/koemoji-go/internal/ui"
	"github.com/hirokitakamura/koemoji-go/internal/whisper"
)

// startPeriodicUpdate starts the 5-second periodic UI update
func (app *GUIApp) startPeriodicUpdate() {
	msg := ui.GetMessages(app.Config)
	
	// Initialize dependencies once
	if err := processor.EnsureDirectories(app.Config, app.logger); err != nil {
		logger.LogError(app.logger, &app.logBuffer, &app.logMutex, msg.DirCreateError, "", err)
	}

	if err := whisper.EnsureDependencies(app.Config, app.logger, &app.logBuffer, &app.logMutex, app.debugMode); err != nil {
		logger.LogError(app.logger, &app.logBuffer, &app.logMutex, msg.WhisperNotFound+": %v", err)
		logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "音声認識機能を除く機能で続行します")
		
		// Show dependency error dialog in GUI mode
		go func() {
			// Wait for UI to be ready before showing dialog (max 5 seconds)
			for i := 0; i < 50 && !app.isUIReady(); i++ {
				time.Sleep(100 * time.Millisecond)
			}
			if app.isUIReady() {
				app.showDependencyErrorDialog(err)
			}
		}()
	}

	// Phase 2: Start file processing with context
	go processor.StartProcessing(app.ctx, app.Config, nil, &app.logBuffer, &app.logMutex,
		&app.lastScanTime, &app.queuedFiles, &app.processingFile, &app.isProcessing,
		&app.processedFiles, &app.mu, nil, app.debugMode)

	// Start periodic updates in a goroutine with context cancellation
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-app.ctx.Done():
				logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "GUI定期更新を停止しました")
				return
			case <-ticker.C:
				// Use fyne.Do to safely update UI from goroutine
				fyne.Do(func() {
					app.updateUI()
				})
			}
		}
	}()
}

// KISS Design: Helper methods for state management
// These provide a simple, consistent interface to recording state

// isRecording returns the current recording state from the single source of truth
func (app *GUIApp) isRecording() bool {
	return app.recorder != nil && app.recorder.IsRecording()
}

// getRecordingDuration returns the current recording duration
func (app *GUIApp) getRecordingDuration() time.Duration {
	if !app.isRecording() {
		return 0
	}
	return app.recorder.GetElapsedTime()
}

// updateUI updates all UI components with current data
func (app *GUIApp) updateUI() {
	// Check if UI is ready for updates
	if !app.isUIReady() {
		return
	}

	// KISS Design: Direct query, no synchronization needed
	isCurrentlyRecording := app.isRecording()

	msg := ui.GetMessages(app.Config)

	// Update file counts
	app.updateFileCounts()

	// Update status label and icon
	status := msg.Active
	if app.isProcessing {
		status = msg.Processing
		app.statusIcon.SetResource(theme.WarningIcon())  // ⚠ 処理中
	} else {
		app.statusIcon.SetResource(theme.ConfirmIcon())  // ✓ 稼働中
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
	filesText := fmt.Sprintf("%s: %d → %s: %d → %s: %d",
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

	timingText := fmt.Sprintf("%s: %s | %s: %s | %s: %s",
		msg.Last, lastScanStr, msg.Next, nextScanStr, msg.Uptime, formatDuration(uptime))

	// Add recording status if recording and update icon
	if isCurrentlyRecording {
		elapsed := app.getRecordingDuration()
		timingText += fmt.Sprintf(" | %s: %s", msg.Recording, formatDuration(elapsed))
		app.timingIcon.SetResource(theme.MediaRecordIcon())  // ⏺ 録音中
	} else {
		app.timingIcon.SetResource(theme.SearchIcon())  // 🔍 スキャン
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
		app.logText.ParseMarkdown("**ログエントリを待機中...**")
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
	logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "設定ダイアログを開きました")
}

// onLogsPressed handles the logs button press
func (app *GUIApp) onLogsPressed() {
	// Open log file using existing UI function
	ui.DisplayLogs(app.Config)

	// Log the action
	logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "ログファイルを開きました")
}

// onScanPressed handles the scan button press
func (app *GUIApp) onScanPressed() {
	logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "手動スキャンを実行しました")

	// Use existing sync.WaitGroup reference if available, or create minimal scan
	processor.ScanAndProcess(app.Config, nil, &app.logBuffer, &app.logMutex,
		&app.lastScanTime, &app.queuedFiles, &app.processingFile, &app.isProcessing,
		&app.processedFiles, &app.mu, nil, app.debugMode)
}

// onInputDirPressed handles the input directory button press
func (app *GUIApp) onInputDirPressed() {
	if err := ui.OpenDirectory(app.Config.InputDir); err != nil {
		logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "入力ディレクトリを開けませんでした: %v", err)
	}
}

// onOutputDirPressed handles the output directory button press
func (app *GUIApp) onOutputDirPressed() {
	if err := ui.OpenDirectory(app.Config.OutputDir); err != nil {
		logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "出力ディレクトリを開けませんでした: %v", err)
	}
}

// onRecordPressed handles the record button press
func (app *GUIApp) onRecordPressed() {
	// KISS Design: Simple toggle logic with single source of truth
	if app.isRecording() {
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
		// Use device name if specified, otherwise use default device
		if app.Config.RecordingDeviceName != "" {
			app.recorder, err = recorder.NewRecorderWithDeviceName(app.Config.RecordingDeviceName)
		} else {
			app.recorder, err = recorder.NewRecorder()
		}

		if err != nil {
			logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "録音の初期化に失敗: %v", err)
			return
		}

		// Phase 1: Set recording limits
		var maxDuration time.Duration
		var maxFileSize int64

		if app.Config.RecordingMaxHours > 0 {
			maxDuration = time.Duration(app.Config.RecordingMaxHours) * time.Hour
		}

		if app.Config.RecordingMaxFileMB > 0 {
			maxFileSize = int64(app.Config.RecordingMaxFileMB) * 1024 * 1024 // Convert MB to bytes
		}

		app.recorder.SetLimits(maxDuration, maxFileSize)
	}

	// Start recording
	err := app.recorder.Start()
	if err != nil {
		logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "録音の開始に失敗: %v", err)
		return
	}

	// KISS Design: No state sync needed, query directly
	if app.isRecording() {
		logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "録音を開始しました")
	}

	// Update button appearance
	app.updateRecordingUI()
}

// stopRecording stops audio recording
func (app *GUIApp) stopRecording() {
	if app.recorder == nil {
		logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "録音が初期化されていません")
		app.updateRecordingUI()
		return
	}

	// Stop recording
	err := app.recorder.Stop()
	if err != nil {
		logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "録音の停止に失敗: %v", err)
		app.updateRecordingUI()
		return
	}

	// Generate filename with current timestamp
	now := time.Now()
	filename := fmt.Sprintf("recording_%s.wav", now.Format("20060102_1504"))

	// Save to input directory with normalization
	outputPath := filepath.Join(app.Config.InputDir, filename)
	err = app.recorder.SaveToFileWithNormalization(outputPath, app.Config.AudioNormalizationEnabled)
	if err != nil {
		logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "録音ファイルの保存に失敗: %v", err)
		app.updateRecordingUI()
		return
	}

	// KISS Design: Get duration directly from recorder
	duration := app.getRecordingDuration()
	logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "録音を停止しました: %s (録音時間: %s)", filename, duration.Round(time.Second))

	// Update button appearance
	app.updateRecordingUI()
}

// updateRecordingUI updates the recording-related UI elements
func (app *GUIApp) updateRecordingUI() {
	// Check if UI is ready and record button exists
	if !app.isUIReady() {
		return
	}

	msg := ui.GetMessages(app.Config)
	// KISS Design: Direct query for current state
	isCurrentlyRecording := app.isRecording()

	// Use fyne.Do to safely update UI
	fyne.Do(func() {
		if isCurrentlyRecording {
			app.recordButton.SetText("録音停止")
			app.recordButton.Importance = widget.DangerImportance
		} else {
			app.recordButton.SetText(msg.RecordCmd)
			app.recordButton.Importance = widget.WarningImportance
		}
		app.recordButton.Refresh()
	})
}

// onQuitPressed handles the quit button press
func (app *GUIApp) onQuitPressed() {
	// KISS Design: Simple, consistent state check
	if app.isRecording() {
		// Show warning dialog if recording is in progress
		app.showRecordingExitWarning()
		return
	}
	// Immediate exit if not recording
	app.forceQuit()
}
