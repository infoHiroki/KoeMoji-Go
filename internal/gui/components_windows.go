//go:build windows

package gui

import (
	"time"

	"github.com/infoHiroki/KoeMoji-Go/internal/logger"
	"github.com/infoHiroki/KoeMoji-Go/internal/recorder"
)

// initializeRecorder initializes the appropriate recorder based on platform and settings
func (app *GUIApp) initializeRecorder() error {
	var err error

	// Check if dual recording is enabled (Windows only)
	if app.Config.DualRecordingEnabled {
		// Use DualRecorder for system audio + microphone
		var dr *recorder.DualRecorder
		// Check for actual device name (not UI placeholder strings)
		if app.Config.RecordingDeviceName != "" &&
			app.Config.RecordingDeviceName != "デフォルトデバイス" {
			dr, err = recorder.NewDualRecorderWithDevices(app.Config.RecordingDeviceName, app.logger, &app.logBuffer, &app.logMutex, app.debugMode)
		} else {
			dr, err = recorder.NewDualRecorder()
		}

		if err != nil {
			logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "デュアル録音の初期化に失敗: %v", err)
			return err
		}

		// Set volume levels
		dr.SetVolumes(app.Config.SystemAudioVolume, app.Config.MicrophoneVolume)
		app.recorder = dr

		// Convert internal values to relative scale for display
		systemLabel := volumeToRelativeLabel(app.Config.SystemAudioVolume, true)
		micLabel := volumeToRelativeLabel(app.Config.MicrophoneVolume, false)

		if app.debugMode {
			logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex,
				"[DEBUG] デュアル録音設定: システム音声 %s (%.2f) + マイク %s (%.2f)",
				systemLabel, app.Config.SystemAudioVolume, micLabel, app.Config.MicrophoneVolume)
		} else {
			logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex,
				"デュアル録音モード: システム音声(%s) + マイク(%s)", systemLabel, micLabel)
		}
	} else {
		// Use standard Recorder for single device
		if app.Config.RecordingDeviceName != "" {
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
