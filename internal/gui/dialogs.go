package gui

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/hirokitakamura/koemoji-go/internal/config"
	"github.com/hirokitakamura/koemoji-go/internal/logger"
	"github.com/hirokitakamura/koemoji-go/internal/recorder"
)

// showConfigDialog displays the configuration dialog with tabbed interface
func (app *GUIApp) showConfigDialog() {
	// Create form entries for basic settings
	whisperModelEntry := widget.NewEntry()
	whisperModelEntry.SetText(app.Config.WhisperModel)

	languageEntry := widget.NewEntry()
	languageEntry.SetText(app.Config.Language)

	uiLanguageSelect := widget.NewSelect([]string{"en", "ja"}, func(value string) {
		// Handle UI language selection
	})
	uiLanguageSelect.SetSelected(app.Config.UILanguage)

	scanIntervalEntry := widget.NewEntry()
	scanIntervalEntry.SetText(strconv.Itoa(app.Config.ScanIntervalMinutes))

	colorsCheck := widget.NewCheck("", nil)
	colorsCheck.SetChecked(app.Config.UseColors)

	// Basic settings form
	basicForm := widget.NewForm(
		widget.NewFormItem("Whisper Model", whisperModelEntry),
		widget.NewFormItem("Language", languageEntry),
		widget.NewFormItem("UI Language", uiLanguageSelect),
		widget.NewFormItem("Scan Interval (min)", scanIntervalEntry),
		widget.NewFormItem("Use Colors", colorsCheck),
	)

	// Directory settings
	inputDirEntry := widget.NewEntry()
	inputDirEntry.SetText(app.Config.InputDir)

	outputDirEntry := widget.NewEntry()
	outputDirEntry.SetText(app.Config.OutputDir)

	archiveDirEntry := widget.NewEntry()
	archiveDirEntry.SetText(app.Config.ArchiveDir)

	dirForm := widget.NewForm(
		widget.NewFormItem("Input Directory", inputDirEntry),
		widget.NewFormItem("Output Directory", outputDirEntry),
		widget.NewFormItem("Archive Directory", archiveDirEntry),
	)

	// LLM settings
	llmEnabledCheck := widget.NewCheck("", nil)
	llmEnabledCheck.SetChecked(app.Config.LLMSummaryEnabled)

	llmAPIKeyEntry := widget.NewPasswordEntry()
	llmAPIKeyEntry.SetText(app.Config.LLMAPIKey)

	llmModelSelect := widget.NewSelect([]string{"gpt-4o", "gpt-4-turbo", "gpt-3.5-turbo"}, nil)
	llmModelSelect.SetSelected(app.Config.LLMModel)

	llmForm := widget.NewForm(
		widget.NewFormItem("Enable LLM Summary", llmEnabledCheck),
		widget.NewFormItem("API Key", llmAPIKeyEntry),
		widget.NewFormItem("Model", llmModelSelect),
	)

	// Recording settings
	recordingForm := app.createRecordingForm()

	// Create tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Basic", basicForm),
		container.NewTabItem("Directories", dirForm),
		container.NewTabItem("LLM", llmForm),
		container.NewTabItem("Recording", recordingForm),
	)

	// Create dialog content
	content := container.NewVBox(
		widget.NewLabel("KoeMoji-Go Configuration"),
		tabs,
	)

	// Create dialog with Save/Cancel buttons
	configDialog := dialog.NewCustomConfirm("Settings", "Save", "Cancel", content,
		func(save bool) {
			if save {
				// Save configuration changes only when Save is clicked
				app.saveConfigFromDialog(whisperModelEntry, languageEntry, uiLanguageSelect,
					scanIntervalEntry, colorsCheck, inputDirEntry, outputDirEntry,
					archiveDirEntry, llmEnabledCheck, llmAPIKeyEntry, llmModelSelect)
			}
			// If Cancel is clicked, changes are discarded automatically
		}, app.window)
	configDialog.Resize(fyne.NewSize(600, 450))

	configDialog.Show()
}

// createRecordingForm creates the recording settings form
func (app *GUIApp) createRecordingForm() *widget.Form {
	// Get available recording devices
	devices, err := recorder.ListDevices()
	if err != nil {
		logger.LogError(nil, &app.logBuffer, &app.logMutex, "Failed to list recording devices: %v", err)
		return widget.NewForm(
			widget.NewFormItem("Error", widget.NewLabel("Failed to load recording devices")),
		)
	}

	// Create device options
	var deviceNames []string
	var deviceMap = make(map[string]int)
	var selectedDevice string

	deviceNames = append(deviceNames, "Default Device")
	deviceMap["Default Device"] = -1
	selectedDevice = "Default Device"

	for _, device := range devices {
		deviceNames = append(deviceNames, device.Name)
		deviceMap[device.Name] = device.ID
		if device.ID == app.Config.RecordingDeviceID {
			selectedDevice = device.Name
		}
	}

	// If current device ID is -1, keep "Default Device" selected
	if app.Config.RecordingDeviceID == -1 {
		selectedDevice = "Default Device"
	}

	// Create device selection widget
	deviceSelect := widget.NewSelect(deviceNames, nil)
	deviceSelect.SetSelected(selectedDevice)

	// Store reference for saving
	app.recordingDeviceSelect = deviceSelect
	app.recordingDeviceMap = deviceMap

	return widget.NewForm(
		widget.NewFormItem("Recording Device", deviceSelect),
	)
}

// saveConfigFromDialog saves the configuration from dialog form entries
func (app *GUIApp) saveConfigFromDialog(whisperModel, language *widget.Entry,
	uiLanguage *widget.Select, scanInterval *widget.Entry, useColors *widget.Check,
	inputDir, outputDir, archiveDir *widget.Entry, llmEnabled *widget.Check,
	llmAPIKey *widget.Entry, llmModel *widget.Select) {

	// Update configuration
	app.Config.WhisperModel = whisperModel.Text
	app.Config.Language = language.Text
	app.Config.UILanguage = uiLanguage.Selected

	if interval, err := strconv.Atoi(scanInterval.Text); err == nil {
		app.Config.ScanIntervalMinutes = interval
	}

	app.Config.UseColors = useColors.Checked
	app.Config.InputDir = inputDir.Text
	app.Config.OutputDir = outputDir.Text
	app.Config.ArchiveDir = archiveDir.Text
	app.Config.LLMSummaryEnabled = llmEnabled.Checked
	app.Config.LLMAPIKey = llmAPIKey.Text
	app.Config.LLMModel = llmModel.Selected

	// Update recording configuration
	if app.recordingDeviceSelect != nil && app.recordingDeviceMap != nil {
		selectedDevice := app.recordingDeviceSelect.Selected
		if deviceID, exists := app.recordingDeviceMap[selectedDevice]; exists {
			app.Config.RecordingDeviceID = deviceID
			app.Config.RecordingDeviceName = selectedDevice
		}
	}

	// Save to file
	if err := config.SaveConfig(app.Config, app.configPath); err != nil {
		logger.LogError(nil, &app.logBuffer, &app.logMutex, "Failed to save config: %v", err)
		dialog.ShowError(err, app.window)
	} else {
		logger.LogInfo(nil, &app.logBuffer, &app.logMutex, "Configuration saved successfully")
		dialog.ShowInformation("Success", "Configuration saved successfully!", app.window)
	}
}

// showRecordingExitWarning shows a warning dialog when user tries to exit while recording
func (app *GUIApp) showRecordingExitWarning() {
	// KISS Design: Get duration directly from recorder
	elapsed := app.getRecordingDuration()

	// Create warning message with elapsed time
	warningMessage := fmt.Sprintf("録音中です（%s経過）\n録音データが失われますが終了しますか？",
		formatRecordingDuration(elapsed))

	// Create warning dialog
	confirmDialog := dialog.NewConfirm(
		"録音中",
		warningMessage,
		func(confirmed bool) {
			if confirmed {
				// User confirmed exit - force quit
				app.forceQuit()
			}
			// If not confirmed, dialog just closes and continues
		},
		app.window)

	confirmDialog.Show()
}

// forceQuit performs immediate application exit with cleanup
func (app *GUIApp) forceQuit() {
	// Perform cleanup
	app.ForceCleanup()

	// Immediate exit - OS will handle whisper process termination
	app.fyneApp.Quit()
}

// formatRecordingDuration formats a duration for display in recording dialog
func formatRecordingDuration(d time.Duration) string {
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
