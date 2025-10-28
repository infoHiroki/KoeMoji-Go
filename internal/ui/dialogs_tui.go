package ui

import (
	"fmt"
	"runtime"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/infoHiroki/KoeMoji-Go/internal/config"
	"github.com/rivo/tview"
)

// ConfigDialog represents a configuration dialog for TUI
type ConfigDialog struct {
	app    *tview.Application
	pages  *tview.Pages
	config *config.Config

	// UI components
	tabList     *tview.List
	tabPages    *tview.Pages
	mainFlex    *tview.Flex
	buttonFlex  *tview.Flex
	saveButton  *tview.Button
	cancelButton *tview.Button

	// Form fields - Basic settings
	uiLanguageDropDown   *tview.DropDown
	whisperModelDropDown *tview.DropDown
	languageDropDown     *tview.DropDown
	scanIntervalField    *tview.InputField

	// Form fields - Directories
	inputDirField   *tview.InputField
	outputDirField  *tview.InputField
	archiveDirField *tview.InputField

	// Form fields - LLM settings
	llmEnabledCheckbox *tview.Checkbox
	llmAPIKeyField     *tview.InputField
	llmModelDropDown   *tview.DropDown

	// Form fields - Recording settings
	recordingDeviceField  *tview.InputField
	dualRecordingCheckbox *tview.Checkbox

	// Form fields - Advanced settings
	computeTypeDropDown  *tview.DropDown
	outputFormatDropDown *tview.DropDown
	maxCpuField          *tview.InputField

	// Callbacks
	onSave   func()
	onCancel func()
}

// NewConfigDialog creates a new configuration dialog
func NewConfigDialog(app *tview.Application, pages *tview.Pages, cfg *config.Config) *ConfigDialog {
	return &ConfigDialog{
		app:    app,
		pages:  pages,
		config: cfg,
	}
}

// Show displays the configuration dialog
func (d *ConfigDialog) Show(onSave, onCancel func()) {
	d.onSave = onSave
	d.onCancel = onCancel

	msg := GetMessages(d.config)

	// Create tab pages
	d.tabPages = tview.NewPages()

	// Create tabs
	d.createBasicTab(msg)
	d.createDirectoriesTab(msg)
	d.createLLMTab(msg)
	d.createRecordingTab(msg)
	d.createAdvancedTab(msg)

	// Create tab list (left side)
	d.tabList = tview.NewList().
		AddItem(msg.BasicTab, "", '1', nil).
		AddItem(msg.DirectoriesTab, "", '2', nil).
		AddItem(msg.LLMTab, "", '3', nil).
		AddItem(msg.RecordingTab, "", '4', nil).
		AddItem("詳細設定", "", '5', nil)

	d.tabList.SetBorder(true).
		SetTitle(" タブ ").
		SetTitleAlign(tview.AlignCenter)

	// ↑↓で即座にタブ切り替え（Enterキー不要）
	d.tabList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		d.switchTab(index)
	})

	// Set initial tab
	d.switchTab(0)

	// Create buttons
	d.saveButton = tview.NewButton(msg.SaveBtn)
	d.saveButton.SetSelectedFunc(func() {
		d.saveConfig()
		d.close()
		if d.onSave != nil {
			d.onSave()
		}
	})

	d.cancelButton = tview.NewButton(msg.CancelBtn)
	d.cancelButton.SetSelectedFunc(func() {
		d.close()
		if d.onCancel != nil {
			d.onCancel()
		}
	})

	// Button layout
	d.buttonFlex = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(d.saveButton, 12, 0, false).
		AddItem(nil, 2, 0, false).
		AddItem(d.cancelButton, 14, 0, false).
		AddItem(nil, 0, 1, false)

	// Help text
	helpText := tview.NewTextView().
		SetText("操作: ↑↓で項目選択 | Tabでタブ切替 | Enterで決定 | Spaceでチェックボックス | ESCで閉じる").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	// Main layout: tab list (left) + tab content (right)
	contentFlex := tview.NewFlex().
		AddItem(d.tabList, 20, 0, true).
		AddItem(d.tabPages, 0, 1, false)

	// Overall layout: content + help + buttons
	d.mainFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(contentFlex, 0, 1, true).
		AddItem(helpText, 1, 0, false).
		AddItem(d.buttonFlex, 3, 0, false)

	d.mainFlex.SetBorder(true).
		SetTitle(" " + msg.SettingsTitle + " ").
		SetTitleAlign(tview.AlignCenter)

	// Handle keyboard shortcuts
	d.mainFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			d.close()
			if d.onCancel != nil {
				d.onCancel()
			}
			return nil
		case tcell.KeyTab:
			// Tabキーでフォーカス移動: form -> tabList -> buttons -> form
			currentFocus := d.app.GetFocus()
			_, primitive := d.tabPages.GetFrontPage()

			if currentFocus == primitive {
				// form -> tabList
				d.app.SetFocus(d.tabList)
			} else if currentFocus == d.tabList {
				// tabList -> buttons
				d.app.SetFocus(d.saveButton)
			} else {
				// buttons -> form
				d.app.SetFocus(primitive)
			}
			return nil
		case tcell.KeyBacktab: // Shift+Tab
			// 逆方向に移動
			currentFocus := d.app.GetFocus()
			_, primitive := d.tabPages.GetFrontPage()

			if currentFocus == primitive {
				// form -> buttons
				d.app.SetFocus(d.cancelButton)
			} else if currentFocus == d.tabList {
				// tabList -> form
				d.app.SetFocus(primitive)
			} else {
				// buttons -> tabList
				d.app.SetFocus(d.tabList)
			}
			return nil
		}
		return event
	})

	// Add to pages
	d.pages.AddPage("config", d.mainFlex, true, true)

	// 初期フォーカスを設定項目に設定（すぐに編集できるように）
	_, primitive := d.tabPages.GetFrontPage()
	if primitive != nil {
		d.app.SetFocus(primitive)
	}
}

// switchTab changes the active tab
func (d *ConfigDialog) switchTab(index int) {
	tabs := []string{"basic", "directories", "llm", "recording", "advanced"}
	if index >= 0 && index < len(tabs) {
		d.tabPages.SwitchToPage(tabs[index])
		d.tabList.SetCurrentItem(index)

		// 切り替え後、自動的にそのタブのコンテンツにフォーカスを移動
		// これにより、すぐに設定を編集できる
		_, primitive := d.tabPages.GetFrontPage()
		if primitive != nil {
			d.app.SetFocus(primitive)
		}
	}
}

// createBasicTab creates the basic settings tab
func (d *ConfigDialog) createBasicTab(msg *Messages) {
	// UI Language dropdown
	uiLanguageOptions := []string{"English", "日本語"}
	uiLanguageCodes := []string{"en", "ja"}
	currentUILangIndex := 0
	for i, code := range uiLanguageCodes {
		if code == d.config.UILanguage {
			currentUILangIndex = i
			break
		}
	}
	d.uiLanguageDropDown = tview.NewDropDown().
		SetLabel(msg.LanguageLabel + ": ").
		SetOptions(uiLanguageOptions, nil).
		SetCurrentOption(currentUILangIndex)

	// Whisper model dropdown
	whisperModels := []string{
		"tiny", "tiny.en", "base", "base.en",
		"small", "small.en", "medium", "medium.en",
		"large", "large-v1", "large-v2", "large-v3",
	}
	currentModelIndex := 0
	for i, model := range whisperModels {
		if model == d.config.WhisperModel {
			currentModelIndex = i
			break
		}
	}
	d.whisperModelDropDown = tview.NewDropDown().
		SetLabel(msg.WhisperModelLabel + ": ").
		SetOptions(whisperModels, nil).
		SetCurrentOption(currentModelIndex)

	// Language dropdown
	languageOptions := []string{
		"Auto（自動検出）", "日本語", "English", "中文（简体）", "한국어",
		"Español", "Français", "Deutsch", "Русский", "العربية",
		"हिन्दी", "Italiano", "Português",
	}
	languageCodes := []string{
		"auto", "ja", "en", "zh", "ko", "es", "fr", "de", "ru", "ar", "hi", "it", "pt",
	}
	currentLangIndex := 0
	for i, code := range languageCodes {
		if code == d.config.Language {
			currentLangIndex = i
			break
		}
	}
	d.languageDropDown = tview.NewDropDown().
		SetLabel(msg.SpeechLanguageLabel + ": ").
		SetOptions(languageOptions, nil).
		SetCurrentOption(currentLangIndex)

	// Scan interval field
	d.scanIntervalField = tview.NewInputField().
		SetLabel(msg.ScanIntervalLabel + " (分): ").
		SetText(strconv.Itoa(d.config.ScanIntervalMinutes)).
		SetFieldWidth(10).
		SetAcceptanceFunc(tview.InputFieldInteger)

	// Create form
	form := tview.NewForm().
		AddFormItem(d.uiLanguageDropDown).
		AddFormItem(d.whisperModelDropDown).
		AddFormItem(d.languageDropDown).
		AddFormItem(d.scanIntervalField)

	form.SetBorder(true).
		SetTitle(" " + msg.BasicTab + " ").
		SetBorderPadding(0, 1, 1, 1) // Add bottom padding to avoid button overlap
	d.tabPages.AddPage("basic", form, true, true)
}

// createDirectoriesTab creates the directories settings tab
func (d *ConfigDialog) createDirectoriesTab(msg *Messages) {
	// Input directory field
	d.inputDirField = tview.NewInputField().
		SetLabel(msg.InputDirLabel + ": ").
		SetText(config.GetRelativePath(d.config.InputDir)).
		SetFieldWidth(50)

	// Output directory field
	d.outputDirField = tview.NewInputField().
		SetLabel(msg.OutputDirLabel + ": ").
		SetText(config.GetRelativePath(d.config.OutputDir)).
		SetFieldWidth(50)

	// Archive directory field
	d.archiveDirField = tview.NewInputField().
		SetLabel(msg.ArchiveDirLabel + ": ").
		SetText(config.GetRelativePath(d.config.ArchiveDir)).
		SetFieldWidth(50)

	// Create form
	form := tview.NewForm().
		AddFormItem(d.inputDirField).
		AddFormItem(d.outputDirField).
		AddFormItem(d.archiveDirField)

	form.SetBorder(true).
		SetTitle(" " + msg.DirectoriesTab + " ").
		SetBorderPadding(0, 1, 1, 1)
	d.tabPages.AddPage("directories", form, true, false)
}

// createLLMTab creates the LLM settings tab
func (d *ConfigDialog) createLLMTab(msg *Messages) {
	// LLM enabled checkbox
	d.llmEnabledCheckbox = tview.NewCheckbox().
		SetLabel(msg.LLMEnabledLabel + ": ").
		SetChecked(d.config.LLMSummaryEnabled)

	// LLM API key field
	d.llmAPIKeyField = tview.NewInputField().
		SetLabel(msg.APIKeyLabel + ": ").
		SetText(d.config.LLMAPIKey).
		SetFieldWidth(50).
		SetMaskCharacter('*')

	// LLM model dropdown
	llmModels := []string{"gpt-4o", "gpt-4-turbo", "gpt-3.5-turbo"}
	currentLLMModelIndex := 0
	for i, model := range llmModels {
		if model == d.config.LLMModel {
			currentLLMModelIndex = i
			break
		}
	}
	d.llmModelDropDown = tview.NewDropDown().
		SetLabel(msg.ModelLabel + ": ").
		SetOptions(llmModels, nil).
		SetCurrentOption(currentLLMModelIndex)

	// Create form
	form := tview.NewForm().
		AddFormItem(d.llmEnabledCheckbox).
		AddFormItem(d.llmAPIKeyField).
		AddFormItem(d.llmModelDropDown)

	form.SetBorder(true).
		SetTitle(" " + msg.LLMTab + " ").
		SetBorderPadding(0, 1, 1, 1)
	d.tabPages.AddPage("llm", form, true, false)
}

// createRecordingTab creates the recording settings tab
func (d *ConfigDialog) createRecordingTab(msg *Messages) {
	// Recording device field
	d.recordingDeviceField = tview.NewInputField().
		SetLabel(msg.RecordingDeviceLabel + ": ").
		SetText(d.config.RecordingDeviceName).
		SetFieldWidth(40)

	// Dual recording checkbox (macOS only)
	d.dualRecordingCheckbox = tview.NewCheckbox().
		SetLabel(msg.DualRecordingLabel + ": ").
		SetChecked(d.config.DualRecordingEnabled)

	// Create form
	form := tview.NewForm().
		AddFormItem(d.recordingDeviceField)

	// Add dual recording checkbox for macOS only
	if runtime.GOOS == "darwin" {
		form.AddFormItem(d.dualRecordingCheckbox)
	}

	form.SetBorder(true).
		SetTitle(" " + msg.RecordingTab + " ").
		SetBorderPadding(0, 1, 1, 1)
	d.tabPages.AddPage("recording", form, true, false)
}

// createAdvancedTab creates the advanced settings tab
func (d *ConfigDialog) createAdvancedTab(msg *Messages) {
	// Compute type dropdown
	computeTypes := []string{"int8", "int8_float16", "int16", "float16", "float32"}
	currentComputeTypeIndex := 0
	for i, ctype := range computeTypes {
		if ctype == d.config.ComputeType {
			currentComputeTypeIndex = i
			break
		}
	}
	d.computeTypeDropDown = tview.NewDropDown().
		SetLabel("計算精度: ").
		SetOptions(computeTypes, nil).
		SetCurrentOption(currentComputeTypeIndex)

	// Output format dropdown
	outputFormats := []string{"txt", "vtt", "srt", "tsv", "json"}
	currentFormatIndex := 0
	for i, format := range outputFormats {
		if format == d.config.OutputFormat {
			currentFormatIndex = i
			break
		}
	}
	d.outputFormatDropDown = tview.NewDropDown().
		SetLabel("出力形式: ").
		SetOptions(outputFormats, nil).
		SetCurrentOption(currentFormatIndex)

	// Max CPU field
	d.maxCpuField = tview.NewInputField().
		SetLabel("最大CPU使用率 (%): ").
		SetText(strconv.Itoa(d.config.MaxCpuPercent)).
		SetFieldWidth(10).
		SetAcceptanceFunc(tview.InputFieldInteger)

	// Create form
	form := tview.NewForm().
		AddFormItem(d.computeTypeDropDown).
		AddFormItem(d.outputFormatDropDown).
		AddFormItem(d.maxCpuField)

	form.SetBorder(true).
		SetTitle(" 詳細設定 ").
		SetBorderPadding(0, 1, 1, 1)
	d.tabPages.AddPage("advanced", form, true, false)
}

// saveConfig saves the configuration from form fields
func (d *ConfigDialog) saveConfig() {
	// Basic settings
	uiLanguageCodes := []string{"en", "ja"}
	uiLangIndex, _ := d.uiLanguageDropDown.GetCurrentOption()
	if uiLangIndex >= 0 && uiLangIndex < len(uiLanguageCodes) {
		d.config.UILanguage = uiLanguageCodes[uiLangIndex]
	}

	_, whisperModel := d.whisperModelDropDown.GetCurrentOption()
	d.config.WhisperModel = whisperModel

	languageCodes := []string{
		"auto", "ja", "en", "zh", "ko", "es", "fr", "de", "ru", "ar", "hi", "it", "pt",
	}
	langIndex, _ := d.languageDropDown.GetCurrentOption()
	if langIndex >= 0 && langIndex < len(languageCodes) {
		d.config.Language = languageCodes[langIndex]
	}

	if interval, err := strconv.Atoi(d.scanIntervalField.GetText()); err == nil && interval > 0 {
		d.config.ScanIntervalMinutes = interval
	}

	// Directories
	d.config.InputDir = d.inputDirField.GetText()
	d.config.OutputDir = d.outputDirField.GetText()
	d.config.ArchiveDir = d.archiveDirField.GetText()

	// LLM settings
	d.config.LLMSummaryEnabled = d.llmEnabledCheckbox.IsChecked()
	d.config.LLMAPIKey = d.llmAPIKeyField.GetText()

	llmModelIndex, llmModel := d.llmModelDropDown.GetCurrentOption()
	if llmModelIndex >= 0 {
		d.config.LLMModel = llmModel
	}

	// Recording settings
	d.config.RecordingDeviceName = d.recordingDeviceField.GetText()

	// Get dual recording setting (macOS only)
	if runtime.GOOS == "darwin" {
		d.config.DualRecordingEnabled = d.dualRecordingCheckbox.IsChecked()
	}

	// Advanced settings
	_, computeType := d.computeTypeDropDown.GetCurrentOption()
	d.config.ComputeType = computeType

	_, outputFormat := d.outputFormatDropDown.GetCurrentOption()
	d.config.OutputFormat = outputFormat

	if maxCpu, err := strconv.Atoi(d.maxCpuField.GetText()); err == nil && maxCpu >= 1 && maxCpu <= 100 {
		d.config.MaxCpuPercent = maxCpu
	}

	// Save to file
	if err := config.SaveConfig(d.config, "config.json"); err != nil {
		// TODO: Show error dialog
		fmt.Printf("設定の保存に失敗しました: %v\n", err)
	}
}

// close removes the dialog from pages
func (d *ConfigDialog) close() {
	d.pages.RemovePage("config")
}
