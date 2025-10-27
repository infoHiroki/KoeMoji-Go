//go:build darwin

package gui

import (
	"time"

	"github.com/infoHiroki/KoeMoji-Go/internal/logger"
	"github.com/infoHiroki/KoeMoji-Go/internal/recorder"
)

// initializeRecorder initializes the appropriate recorder based on platform and settings
// macOS: Supports both single device and dual recording (system audio + microphone)
func (app *GUIApp) initializeRecorder() error {
	var err error

	// Check if dual recording is enabled (macOS 13+)
	if app.Config.DualRecordingEnabled {
		// Use DualRecorder for system audio + microphone
		var dr *recorder.DualRecorder
		// Check for actual device name (not UI placeholder strings)
		if app.Config.RecordingDeviceName != "" &&
			app.Config.RecordingDeviceName != "デフォルトデバイス" {
			dr, err = recorder.NewDualRecorderWithDevices(app.Config.RecordingDeviceName)
		} else {
			dr, err = recorder.NewDualRecorder()
		}

		if err != nil {
			logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "デュアル録音の初期化に失敗: %v", err)
			return err
		}

		app.recorder = dr

		if app.debugMode {
			logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex,
				"[DEBUG] デュアル録音モード: システム音声(48kHz Stereo) + マイク(44.1kHz Mono)")
		} else {
			logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex,
				"デュアル録音モード: システム音声 + マイク")
		}
	} else {
		// Use standard Recorder for single device
		if app.Config.RecordingDeviceName != "" &&
		   app.Config.RecordingDeviceName != "デフォルトデバイス" {
			app.recorder, err = recorder.NewRecorderWithDeviceName(app.Config.RecordingDeviceName)
		} else {
			app.recorder, err = recorder.NewRecorder()
		}

		if err != nil {
			logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "録音の初期化に失敗: %v", err)
			return err
		}
	}

	// Set recording limits
	var maxDuration time.Duration
	var maxFileSize int64

	if app.Config.RecordingMaxHours > 0 {
		maxDuration = time.Duration(app.Config.RecordingMaxHours) * time.Hour
	}

	if app.Config.RecordingMaxFileMB > 0 {
		maxFileSize = int64(app.Config.RecordingMaxFileMB) * 1024 * 1024 // Convert MB to bytes
	}

	app.recorder.SetLimits(maxDuration, maxFileSize)
	return nil
}
