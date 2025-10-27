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

	// Form fields
	whisperModelDropDown *tview.DropDown
	languageDropDown     *tview.DropDown
	scanIntervalField    *tview.InputField
	llmEnabledCheckbox   *tview.Checkbox
	llmAPIKeyField       *tview.InputField
	llmModelDropDown     *tview.DropDown

	// Recording fields
	recordingDeviceField  *tview.InputField
	dualRecordingCheckbox *tview.Checkbox

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
		SetLabel(msg.ScanIntervalLabel + ": ").
		SetText(strconv.Itoa(d.config.ScanIntervalMinutes)).
		SetFieldWidth(10).
		SetAcceptanceFunc(tview.InputFieldInteger)

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
		AddFormItem(d.whisperModelDropDown).
		AddFormItem(d.languageDropDown).
		AddFormItem(d.scanIntervalField).
		AddFormItem(d.llmEnabledCheckbox).
		AddFormItem(d.llmAPIKeyField).
		AddFormItem(d.llmModelDropDown).
		AddFormItem(d.recordingDeviceField)

	// Add dual recording checkbox for macOS only
	if runtime.GOOS == "darwin" {
		form.AddFormItem(d.dualRecordingCheckbox)
	}

	form.
		AddButton(msg.SaveBtn, func() {
			d.saveConfig()
			d.close()
			if d.onSave != nil {
				d.onSave()
			}
		}).
		AddButton(msg.CancelBtn, func() {
			d.close()
			if d.onCancel != nil {
				d.onCancel()
			}
		})

	form.SetBorder(true).
		SetTitle(" " + msg.SettingsTitle + " ").
		SetTitleAlign(tview.AlignCenter)

	// Handle Escape key to close dialog
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			d.close()
			if d.onCancel != nil {
				d.onCancel()
			}
			return nil
		}
		return event
	})

	// Add to pages
	d.pages.AddPage("config", form, true, true)
}

// saveConfig saves the configuration from form fields
func (d *ConfigDialog) saveConfig() {
	// Get Whisper model
	_, whisperModel := d.whisperModelDropDown.GetCurrentOption()
	d.config.WhisperModel = whisperModel

	// Get language
	languageCodes := []string{
		"auto", "ja", "en", "zh", "ko", "es", "fr", "de", "ru", "ar", "hi", "it", "pt",
	}
	langIndex, _ := d.languageDropDown.GetCurrentOption()
	if langIndex >= 0 && langIndex < len(languageCodes) {
		d.config.Language = languageCodes[langIndex]
	}

	// Get scan interval
	if interval, err := strconv.Atoi(d.scanIntervalField.GetText()); err == nil {
		if interval > 0 {
			d.config.ScanIntervalMinutes = interval
		}
	}

	// Get LLM settings
	d.config.LLMSummaryEnabled = d.llmEnabledCheckbox.IsChecked()
	d.config.LLMAPIKey = d.llmAPIKeyField.GetText()

	llmModelIndex, llmModel := d.llmModelDropDown.GetCurrentOption()
	if llmModelIndex >= 0 {
		d.config.LLMModel = llmModel
	}

	// Get recording settings
	d.config.RecordingDeviceName = d.recordingDeviceField.GetText()

	// Get dual recording setting (macOS only)
	if runtime.GOOS == "darwin" {
		d.config.DualRecordingEnabled = d.dualRecordingCheckbox.IsChecked()
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
