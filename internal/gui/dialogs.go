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
	"github.com/hirokitakamura/koemoji-go/internal/ui"
)

// showConfigDialog displays the configuration dialog with tabbed interface
func (app *GUIApp) showConfigDialog() {
	// Get messages for the current language
	msg := ui.GetMessages(app.Config)

	// Create form entries for basic settings
	// UI Language first - most important setting
	// UI Language options with display names
	uiLanguageOptions := []string{"English", "日本語"}
	uiCodeToDisplayMap := map[string]string{
		"en": "English",
		"ja": "日本語",
	}

	uiLanguageSelect := widget.NewSelect(uiLanguageOptions, func(value string) {
		// Handle UI language selection
	})

	// Set current selection based on config
	if displayName, exists := uiCodeToDisplayMap[app.Config.UILanguage]; exists {
		uiLanguageSelect.SetSelected(displayName)
	} else {
		uiLanguageSelect.SetSelected("日本語") // Default fallback
	}

	// Whisper model selection (dropdown)
	whisperModels := []string{
		"tiny", "tiny.en", "base", "base.en",
		"small", "small.en", "medium", "medium.en",
		"large", "large-v1", "large-v2", "large-v3",
	}
	whisperModelSelect := widget.NewSelect(whisperModels, nil)
	whisperModelSelect.SetSelected(app.Config.WhisperModel)

	// Language selection (dropdown) with display names
	languageOptions := []string{
		"Auto（自動検出）",
		"日本語",
		"English",
		"中文（简体）",
		"한국어",
		"Español",
		"Français",
		"Deutsch",
		"Русский",
		"العربية",
		"हिन्दी",
		"Italiano",
		"Português",
	}

	// Reverse map for setting current selection
	codeToDisplayMap := map[string]string{
		"auto": "Auto（自動検出）",
		"ja":   "日本語",
		"en":   "English",
		"zh":   "中文（简体）",
		"ko":   "한국어",
		"es":   "Español",
		"fr":   "Français",
		"de":   "Deutsch",
		"ru":   "Русский",
		"ar":   "العربية",
		"hi":   "हिन्दी",
		"it":   "Italiano",
		"pt":   "Português",
	}

	languageSelect := widget.NewSelect(languageOptions, nil)
	if displayName, exists := codeToDisplayMap[app.Config.Language]; exists {
		languageSelect.SetSelected(displayName)
	} else {
		languageSelect.SetSelected("日本語") // Default fallback
	}

	scanIntervalEntry := widget.NewEntry()
	scanIntervalEntry.SetText(strconv.Itoa(app.Config.ScanIntervalMinutes))

	// Basic settings form
	basicForm := widget.NewForm(
		widget.NewFormItem(msg.LanguageLabel, uiLanguageSelect),
		widget.NewFormItem(msg.WhisperModelLabel, whisperModelSelect),
		widget.NewFormItem(msg.SpeechLanguageLabel, languageSelect),
		widget.NewFormItem(msg.ScanIntervalLabel, scanIntervalEntry),
	)

	// Directory settings
	inputDirEntry := widget.NewEntry()
	inputDirEntry.SetText(app.Config.InputDir)

	outputDirEntry := widget.NewEntry()
	outputDirEntry.SetText(app.Config.OutputDir)

	archiveDirEntry := widget.NewEntry()
	archiveDirEntry.SetText(app.Config.ArchiveDir)

	dirForm := widget.NewForm(
		widget.NewFormItem(msg.InputDirLabel, inputDirEntry),
		widget.NewFormItem(msg.OutputDirLabel, outputDirEntry),
		widget.NewFormItem(msg.ArchiveDirLabel, archiveDirEntry),
	)

	// LLM settings
	llmEnabledCheck := widget.NewCheck("", nil)
	llmEnabledCheck.SetChecked(app.Config.LLMSummaryEnabled)

	llmAPIKeyEntry := widget.NewPasswordEntry()
	llmAPIKeyEntry.SetText(app.Config.LLMAPIKey)

	llmModelSelect := widget.NewSelect([]string{"gpt-4o", "gpt-4-turbo", "gpt-3.5-turbo"}, nil)
	llmModelSelect.SetSelected(app.Config.LLMModel)

	llmForm := widget.NewForm(
		widget.NewFormItem(msg.LLMEnabledLabel, llmEnabledCheck),
		widget.NewFormItem(msg.APIKeyLabel, llmAPIKeyEntry),
		widget.NewFormItem(msg.ModelLabel, llmModelSelect),
	)

	// Recording settings
	recordingForm := app.createRecordingForm()

	// Create tabs
	tabs := container.NewAppTabs(
		container.NewTabItem(msg.BasicTab, basicForm),
		container.NewTabItem(msg.DirectoriesTab, dirForm),
		container.NewTabItem(msg.LLMTab, llmForm),
		container.NewTabItem(msg.RecordingTab, recordingForm),
	)

	// Create dialog content
	content := container.NewVBox(
		widget.NewLabel(msg.SettingsTitle),
		tabs,
	)

	// Create dialog with Save/Cancel buttons
	configDialog := dialog.NewCustomConfirm(msg.SettingsTitle, msg.SaveBtn, msg.CancelBtn, content,
		func(save bool) {
			if save {
				// Save configuration changes only when Save is clicked
				app.saveConfigFromDialog(whisperModelSelect, languageSelect, uiLanguageSelect,
					scanIntervalEntry, inputDirEntry, outputDirEntry,
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
		logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "Failed to list recording devices: %v", err)
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

	msg := ui.GetMessages(app.Config)
	return widget.NewForm(
		widget.NewFormItem(msg.RecordingDeviceLabel, deviceSelect),
	)
}

// saveConfigFromDialog saves the configuration from dialog form entries
func (app *GUIApp) saveConfigFromDialog(whisperModel, language *widget.Select,
	uiLanguage *widget.Select, scanInterval *widget.Entry,
	inputDir, outputDir, archiveDir *widget.Entry, llmEnabled *widget.Check,
	llmAPIKey *widget.Entry, llmModel *widget.Select) {

	// Language code mapping for saving to config
	languageCodeMap := map[string]string{
		"Auto（自動検出）": "auto",
		"日本語":        "ja",
		"English":    "en",
		"中文（简体）":     "zh",
		"한국어":        "ko",
		"Español":    "es",
		"Français":   "fr",
		"Deutsch":    "de",
		"Русский":    "ru",
		"العربية":    "ar",
		"हिन्दी":     "hi",
		"Italiano":   "it",
		"Português":  "pt",
	}

	// UI Language code mapping
	uiLanguageCodeMap := map[string]string{
		"English": "en",
		"日本語":     "ja",
	}

	// Update configuration
	app.Config.WhisperModel = whisperModel.Selected

	// Convert display name back to language code
	if languageCode, exists := languageCodeMap[language.Selected]; exists {
		app.Config.Language = languageCode
	} else {
		app.Config.Language = "ja" // Default fallback
	}

	// Convert UI language display name back to code
	if uiLangCode, exists := uiLanguageCodeMap[uiLanguage.Selected]; exists {
		app.Config.UILanguage = uiLangCode
	} else {
		app.Config.UILanguage = "ja" // Default fallback
	}

	if interval, err := strconv.Atoi(scanInterval.Text); err == nil {
		app.Config.ScanIntervalMinutes = interval
	}

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
		logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "Failed to save config: %v", err)
		dialog.ShowError(err, app.window)
	} else {
		logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "Configuration saved successfully")
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

// showDependencyErrorDialog shows an error dialog for dependency issues
func (app *GUIApp) showDependencyErrorDialog(err error) {
	message := fmt.Sprintf("FasterWhisper is not available: %v\n\nThe application will continue with limited functionality.\n\nTo fix this issue, please install:\npip install faster-whisper whisper-ctranslate2", err)
	
	// Show error dialog
	dialog.ShowError(fmt.Errorf(message), app.window)
	
	// Log the error
	logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "Dependency error dialog shown: %v", err)
}

// showConfigErrorDialog shows an error dialog for configuration loading issues
func (app *GUIApp) showConfigErrorDialog(err error) {
	title := "Configuration Error"
	message := fmt.Sprintf("Failed to load configuration: %v\n\nUsing default settings instead.", err)
	
	// Show error dialog
	dialog.ShowInformation(title, message, app.window)
	
	// Log the error
	logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "Config error dialog shown: %v", err)
}
