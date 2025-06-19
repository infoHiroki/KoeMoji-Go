package gui

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/hirokitakamura/koemoji-go/internal/config"
	"github.com/hirokitakamura/koemoji-go/internal/logger"
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
	
	// Create tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Basic", basicForm),
		container.NewTabItem("Directories", dirForm),
		container.NewTabItem("LLM", llmForm),
	)
	
	// Create dialog content
	content := container.NewVBox(
		widget.NewLabel("KoeMoji-Go Configuration"),
		tabs,
	)
	
	// Create dialog
	configDialog := dialog.NewCustom("Settings", "Save", content, app.window)
	configDialog.Resize(fyne.NewSize(500, 400))
	
	// Handle save button
	configDialog.SetOnClosed(func() {
		// Save configuration changes
		app.saveConfigFromDialog(whisperModelEntry, languageEntry, uiLanguageSelect, 
			scanIntervalEntry, colorsCheck, inputDirEntry, outputDirEntry, 
			archiveDirEntry, llmEnabledCheck, llmAPIKeyEntry, llmModelSelect)
	})
	
	configDialog.Show()
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
	
	// Save to file
	if err := config.SaveConfig(app.Config, app.configPath); err != nil {
		logger.LogError(nil, &app.logBuffer, &app.logMutex, "Failed to save config: %v", err)
		dialog.ShowError(err, app.window)
	} else {
		logger.LogInfo(nil, &app.logBuffer, &app.logMutex, "Configuration saved successfully")
		dialog.ShowInformation("Success", "Configuration saved successfully!", app.window)
	}
}