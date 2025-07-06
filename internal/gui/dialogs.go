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
		// Default fallback based on current language
		if app.Config.UILanguage == "en" {
			uiLanguageSelect.SetSelected("English")
		} else {
			uiLanguageSelect.SetSelected("日本語")
		}
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
		// Default fallback based on current language
		if app.Config.UILanguage == "en" {
			languageSelect.SetSelected("English")
		} else {
			languageSelect.SetSelected("日本語")
		}
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

	// Directory settings - show relative paths for user-friendly display
	inputDirEntry := widget.NewEntry()
	inputDirEntry.SetText(config.GetRelativePath(app.Config.InputDir))
	inputDirEntry.Resize(fyne.NewSize(400, 40)) // Set minimum width for path visibility
	inputDirBrowseBtn := widget.NewButton(msg.BrowseBtn, func() {
		app.showFolderSelectDialog(inputDirEntry)
	})
	inputDirBrowseBtn.Resize(fyne.NewSize(80, 40))
	inputDirContainer := container.NewHBox(inputDirEntry, inputDirBrowseBtn)

	outputDirEntry := widget.NewEntry()
	outputDirEntry.SetText(config.GetRelativePath(app.Config.OutputDir))
	outputDirEntry.Resize(fyne.NewSize(400, 40)) // Set minimum width for path visibility
	outputDirBrowseBtn := widget.NewButton(msg.BrowseBtn, func() {
		app.showFolderSelectDialog(outputDirEntry)
	})
	outputDirBrowseBtn.Resize(fyne.NewSize(80, 40))
	outputDirContainer := container.NewHBox(outputDirEntry, outputDirBrowseBtn)

	archiveDirEntry := widget.NewEntry()
	archiveDirEntry.SetText(config.GetRelativePath(app.Config.ArchiveDir))
	archiveDirEntry.Resize(fyne.NewSize(400, 40)) // Set minimum width for path visibility
	archiveDirBrowseBtn := widget.NewButton(msg.BrowseBtn, func() {
		app.showFolderSelectDialog(archiveDirEntry)
	})
	archiveDirBrowseBtn.Resize(fyne.NewSize(80, 40))
	archiveDirContainer := container.NewHBox(archiveDirEntry, archiveDirBrowseBtn)

	dirForm := widget.NewForm(
		widget.NewFormItem(msg.InputDirLabel, inputDirContainer),
		widget.NewFormItem(msg.OutputDirLabel, outputDirContainer),
		widget.NewFormItem(msg.ArchiveDirLabel, archiveDirContainer),
	)

	// LLM settings
	llmEnabledCheck := widget.NewCheck("", nil)
	llmEnabledCheck.SetChecked(app.Config.LLMSummaryEnabled)

	llmAPIKeyEntry := widget.NewPasswordEntry()
	llmAPIKeyEntry.SetText(app.Config.LLMAPIKey)

	llmModelSelect := widget.NewSelect([]string{"gpt-4o", "gpt-4-turbo", "gpt-3.5-turbo"}, nil)
	llmModelSelect.SetSelected(app.Config.LLMModel)

	// Prompt template entry (multi-line)
	llmPromptEntry := widget.NewMultiLineEntry()
	llmPromptEntry.SetText(app.Config.SummaryPromptTemplate)
	llmPromptEntry.SetMinRowsVisible(8) // Show 8 rows
	llmPromptEntry.Wrapping = fyne.TextWrapWord

	llmForm := widget.NewForm(
		widget.NewFormItem(msg.LLMEnabledLabel, llmEnabledCheck),
		widget.NewFormItem(msg.APIKeyLabel, llmAPIKeyEntry),
		widget.NewFormItem(msg.ModelLabel, llmModelSelect),
		widget.NewFormItem(msg.PromptTemplateLabel, llmPromptEntry),
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
					archiveDirEntry, llmEnabledCheck, llmAPIKeyEntry, llmModelSelect, llmPromptEntry)
			}
			// If Cancel is clicked, changes are discarded automatically
		}, app.window)
	configDialog.Resize(fyne.NewSize(800, 550))

	configDialog.Show()
}

// createRecordingForm creates the recording settings form
func (app *GUIApp) createRecordingForm() *widget.Form {
	// Get available recording devices
	devices, err := recorder.ListDevices()
	msg := ui.GetMessages(app.Config)
	if err != nil {
		logger.LogError(app.logger, &app.logBuffer, &app.logMutex, msg.RecordingDeviceListError, err)
		return widget.NewForm(
			widget.NewFormItem(msg.ConfigError, widget.NewLabel(msg.DeviceLoadError)),
		)
	}

	// Create device options
	var deviceNames []string
	var deviceMap = make(map[string]int)
	var selectedDevice string

	deviceNames = append(deviceNames, msg.DefaultDevice)
	deviceMap[msg.DefaultDevice] = -1
	selectedDevice = msg.DefaultDevice

	for _, device := range devices {
		deviceNames = append(deviceNames, device.Name)
		deviceMap[device.Name] = device.ID
		if device.ID == app.Config.RecordingDeviceID {
			selectedDevice = device.Name
		}
	}

	// If current device ID is -1, keep "Default Device" selected
	if app.Config.RecordingDeviceID == -1 {
		selectedDevice = msg.DefaultDevice
	}

	// Create device selection widget
	deviceSelect := widget.NewSelectEntry(deviceNames)
	deviceSelect.SetText(selectedDevice)

	// Store reference for saving
	app.recordingDeviceSelect = deviceSelect
	app.recordingDeviceMap = deviceMap

	return widget.NewForm(
		widget.NewFormItem(msg.RecordingDeviceLabel, deviceSelect),
	)
}

// saveConfigFromDialog saves the configuration from dialog form entries
func (app *GUIApp) saveConfigFromDialog(whisperModel, language *widget.Select,
	uiLanguage *widget.Select, scanInterval *widget.Entry,
	inputDir, outputDir, archiveDir *widget.Entry, llmEnabled *widget.Check,
	llmAPIKey *widget.Entry, llmModel *widget.Select, llmPromptTemplate *widget.Entry) {

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

	// Resolve relative paths to absolute paths for internal storage
	app.Config.InputDir = config.ResolvePath(inputDir.Text)
	app.Config.OutputDir = config.ResolvePath(outputDir.Text)
	app.Config.ArchiveDir = config.ResolvePath(archiveDir.Text)
	app.Config.LLMSummaryEnabled = llmEnabled.Checked
	app.Config.LLMAPIKey = llmAPIKey.Text
	app.Config.LLMModel = llmModel.Selected
	app.Config.SummaryPromptTemplate = llmPromptTemplate.Text

	// Update recording configuration
	if app.recordingDeviceSelect != nil && app.recordingDeviceMap != nil {
		selectedDevice := app.recordingDeviceSelect.Text
		if deviceID, exists := app.recordingDeviceMap[selectedDevice]; exists {
			app.Config.RecordingDeviceID = deviceID
			app.Config.RecordingDeviceName = selectedDevice
		} else if selectedDevice == "" || selectedDevice == ui.GetMessages(app.Config).DefaultDevice {
			// Handle empty or default selection
			app.Config.RecordingDeviceID = -1
			app.Config.RecordingDeviceName = ui.GetMessages(app.Config).DefaultDevice
		}
		// If device not found and not empty/default, keep current settings
	}

	// Save to file
	msg := ui.GetMessages(app.Config)
	if err := config.SaveConfig(app.Config, app.configPath); err != nil {
		logger.LogError(app.logger, &app.logBuffer, &app.logMutex, msg.ConfigSaveError, err)
		dialog.ShowError(err, app.window)
	} else {
		logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, msg.ConfigSaved)
		dialog.ShowInformation(msg.Success, msg.ConfigSaved, app.window)
	}
}

// showRecordingExitWarning shows a warning dialog when user tries to exit while recording
func (app *GUIApp) showRecordingExitWarning() {
	// KISS Design: Get duration directly from recorder
	elapsed := app.getRecordingDuration()
	msg := ui.GetMessages(app.Config)

	// Create warning message with elapsed time
	warningMessage := fmt.Sprintf(msg.RecordingExitWarning,
		formatRecordingDuration(elapsed))

	// Create warning dialog
	confirmDialog := dialog.NewConfirm(
		msg.RecordingInProgress,
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
	msg := ui.GetMessages(app.Config)
	message := fmt.Sprintf(msg.DependencyError, err)
	
	// Show error dialog
	dialog.ShowError(fmt.Errorf(message), app.window)
	
	// Log the error
	logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "依存関係エラーダイアログを表示しました: %v", err)
}

// showConfigErrorDialog shows an error dialog for configuration loading issues
func (app *GUIApp) showConfigErrorDialog(err error) {
	msg := ui.GetMessages(app.Config)
	title := msg.ConfigError
	message := fmt.Sprintf(msg.ConfigLoadErrorDialog, err)
	
	// Show error dialog
	dialog.ShowInformation(title, message, app.window)
	
	// Log the error
	logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "設定エラーダイアログを表示しました: %v", err)
}

// showFolderSelectDialog shows a folder selection dialog and updates the entry field
func (app *GUIApp) showFolderSelectDialog(entry *widget.Entry) {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			// User cancelled or error occurred
			return
		}
		if uri == nil {
			// No folder selected
			return
		}
		
		// Convert URI to path string
		selectedPath := uri.Path()
		
		// Convert to relative path for display
		relativePath := config.GetRelativePath(selectedPath)
		
		// Update the entry field
		entry.SetText(relativePath)
	}, app.window)
}
