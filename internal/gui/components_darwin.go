//go:build darwin

package gui

import (
	"time"

	"github.com/infoHiroki/KoeMoji-Go/internal/logger"
	"github.com/infoHiroki/KoeMoji-Go/internal/recorder"
)

// initializeRecorder initializes the appropriate recorder based on platform and settings
// macOS: Only single device recording is supported
func (app *GUIApp) initializeRecorder() error {
	var err error

	// macOS does not support DualRecorder, always use standard Recorder
	if app.Config.RecordingDeviceName != "" {
		app.recorder, err = recorder.NewRecorderWithDeviceName(app.Config.RecordingDeviceName)
	} else {
		app.recorder, err = recorder.NewRecorder()
	}

	if err != nil {
		logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "録音の初期化に失敗: %v", err)
		return err
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
