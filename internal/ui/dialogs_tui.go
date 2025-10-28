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
	settingsList *tview.List
	mainFlex     *tview.Flex
	buttonFlex   *tview.Flex
	saveButton   *tview.Button
	cancelButton *tview.Button

	// Callbacks
	onSave   func()
	onCancel func()
}

// settingItem represents a configuration item
type settingItem struct {
	key         string
	label       string
	getValue    func() string
	editFunc    func(*ConfigDialog)
	showOnMacOS bool // true if this setting should only show on macOS
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

	// Define all settings
	settings := d.getSettings(msg)

	// Create settings list
	d.settingsList = tview.NewList()
	d.settingsList.ShowSecondaryText(false)
	d.settingsList.SetBorder(true).
		SetTitle(" 設定項目 ").
		SetTitleAlign(tview.AlignCenter)

	// Add all settings to list
	for _, setting := range settings {
		// Skip macOS-only settings on other platforms
		if setting.showOnMacOS && runtime.GOOS != "darwin" {
			continue
		}

		mainText := fmt.Sprintf("%s: %s", setting.label, setting.getValue())
		d.settingsList.AddItem(mainText, "", 0, nil)
	}

	// Handle Enter key to edit selected setting
	d.settingsList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		// Find the actual setting index (accounting for skipped macOS settings)
		actualIndex := d.getActualSettingIndex(settings, index)
		if actualIndex >= 0 && actualIndex < len(settings) {
			settings[actualIndex].editFunc(d)
			// Update the list item display after editing
			d.updateListItem(index, settings[actualIndex])
		}
	})

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
		SetText("操作: ↑↓で項目選択 | Enterで編集 | Tabでボタンへ移動 | ESCで閉じる").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	// Overall layout: list + help + buttons
	d.mainFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(d.settingsList, 0, 1, true).
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
		case tcell.KeyRune:
			// Handle q/Q to close dialog (prevent app shutdown)
			if event.Rune() == 'q' || event.Rune() == 'Q' {
				d.close()
				if d.onCancel != nil {
					d.onCancel()
				}
				return nil
			}
		case tcell.KeyTab:
			// Tab: list -> save button
			if d.app.GetFocus() == d.settingsList {
				d.app.SetFocus(d.saveButton)
			} else if d.app.GetFocus() == d.saveButton {
				d.app.SetFocus(d.cancelButton)
			} else {
				d.app.SetFocus(d.settingsList)
			}
			return nil
		case tcell.KeyBacktab: // Shift+Tab
			// Reverse direction
			if d.app.GetFocus() == d.settingsList {
				d.app.SetFocus(d.cancelButton)
			} else if d.app.GetFocus() == d.cancelButton {
				d.app.SetFocus(d.saveButton)
			} else {
				d.app.SetFocus(d.settingsList)
			}
			return nil
		}
		return event
	})

	// Add to pages
	d.pages.AddPage("config", d.mainFlex, true, true)

	// Set initial focus to settings list
	d.app.SetFocus(d.settingsList)
}

// getSettings returns all configuration settings
func (d *ConfigDialog) getSettings(msg *Messages) []settingItem {
	return []settingItem{
		// Basic settings
		{
			key:   "ui_language",
			label: msg.LanguageLabel,
			getValue: func() string {
				if d.config.UILanguage == "ja" {
					return "日本語"
				}
				return "English"
			},
			editFunc: func(_ *ConfigDialog) { d.editUILanguage() },
		},
		{
			key:   "whisper_model",
			label: msg.WhisperModelLabel,
			getValue: func() string {
				return d.config.WhisperModel
			},
			editFunc: func(_ *ConfigDialog) { d.editWhisperModel() },
		},
		{
			key:   "language",
			label: msg.SpeechLanguageLabel,
			getValue: func() string {
				langMap := map[string]string{
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
				if display, ok := langMap[d.config.Language]; ok {
					return display
				}
				return d.config.Language
			},
			editFunc: func(_ *ConfigDialog) { d.editLanguage() },
		},
		{
			key:   "scan_interval",
			label: msg.ScanIntervalLabel,
			getValue: func() string {
				return fmt.Sprintf("%d分", d.config.ScanIntervalMinutes)
			},
			editFunc: func(_ *ConfigDialog) { d.editScanInterval() },
		},

		// Directory settings
		{
			key:   "input_dir",
			label: msg.InputDirLabel,
			getValue: func() string {
				return config.GetRelativePath(d.config.InputDir)
			},
			editFunc: func(_ *ConfigDialog) { d.editInputDir() },
		},
		{
			key:   "output_dir",
			label: msg.OutputDirLabel,
			getValue: func() string {
				return config.GetRelativePath(d.config.OutputDir)
			},
			editFunc: func(_ *ConfigDialog) { d.editOutputDir() },
		},
		{
			key:   "archive_dir",
			label: msg.ArchiveDirLabel,
			getValue: func() string {
				return config.GetRelativePath(d.config.ArchiveDir)
			},
			editFunc: func(_ *ConfigDialog) { d.editArchiveDir() },
		},

		// LLM settings
		{
			key:   "llm_enabled",
			label: msg.LLMEnabledLabel,
			getValue: func() string {
				if d.config.LLMSummaryEnabled {
					return "有効"
				}
				return "無効"
			},
			editFunc: func(_ *ConfigDialog) { d.editLLMEnabled() },
		},
		{
			key:   "llm_api_key",
			label: msg.APIKeyLabel,
			getValue: func() string {
				if d.config.LLMAPIKey == "" {
					return "(未設定)"
				}
				return "********"
			},
			editFunc: func(_ *ConfigDialog) { d.editLLMAPIKey() },
		},
		{
			key:   "llm_model",
			label: msg.ModelLabel,
			getValue: func() string {
				return d.config.LLMModel
			},
			editFunc: func(_ *ConfigDialog) { d.editLLMModel() },
		},

		// Recording settings
		{
			key:   "recording_device",
			label: msg.RecordingDeviceLabel,
			getValue: func() string {
				if d.config.RecordingDeviceName == "" {
					return "(デフォルト)"
				}
				return d.config.RecordingDeviceName
			},
			editFunc: func(_ *ConfigDialog) { d.editRecordingDevice() },
		},
		{
			key:         "dual_recording",
			label:       msg.DualRecordingLabel,
			showOnMacOS: true, // macOS only
			getValue: func() string {
				if d.config.DualRecordingEnabled {
					return "有効"
				}
				return "無効"
			},
			editFunc: func(_ *ConfigDialog) { d.editDualRecording() },
		},

		// Advanced settings
		{
			key:   "compute_type",
			label: "計算精度",
			getValue: func() string {
				return d.config.ComputeType
			},
			editFunc: func(_ *ConfigDialog) { d.editComputeType() },
		},
		{
			key:   "output_format",
			label: "出力形式",
			getValue: func() string {
				return d.config.OutputFormat
			},
			editFunc: func(_ *ConfigDialog) { d.editOutputFormat() },
		},
		{
			key:   "max_cpu",
			label: "最大CPU使用率",
			getValue: func() string {
				return fmt.Sprintf("%d%%", d.config.MaxCpuPercent)
			},
			editFunc: func(_ *ConfigDialog) { d.editMaxCPU() },
		},
	}
}

// getActualSettingIndex returns the actual setting index accounting for skipped macOS settings
func (d *ConfigDialog) getActualSettingIndex(settings []settingItem, listIndex int) int {
	visibleIndex := 0
	for i, setting := range settings {
		// Skip macOS-only settings on other platforms
		if setting.showOnMacOS && runtime.GOOS != "darwin" {
			continue
		}
		if visibleIndex == listIndex {
			return i
		}
		visibleIndex++
	}
	return -1
}

// updateListItem updates a list item display after editing
func (d *ConfigDialog) updateListItem(index int, setting settingItem) {
	mainText := fmt.Sprintf("%s: %s", setting.label, setting.getValue())
	d.settingsList.SetItemText(index, mainText, "")
}

// Edit functions for each setting type

func (d *ConfigDialog) editUILanguage() {
	options := []string{"English", "日本語"}
	codes := []string{"en", "ja"}
	currentIndex := 0
	for i, code := range codes {
		if code == d.config.UILanguage {
			currentIndex = i
			break
		}
	}

	d.showDropDownModal("UI言語", options, currentIndex, func(index int) {
		if index >= 0 && index < len(codes) {
			d.config.UILanguage = codes[index]
		}
	})
}

func (d *ConfigDialog) editWhisperModel() {
	options := []string{
		"tiny", "tiny.en", "base", "base.en",
		"small", "small.en", "medium", "medium.en",
		"large", "large-v1", "large-v2", "large-v3",
	}
	currentIndex := 0
	for i, model := range options {
		if model == d.config.WhisperModel {
			currentIndex = i
			break
		}
	}

	d.showDropDownModal("Whisperモデル", options, currentIndex, func(index int) {
		if index >= 0 && index < len(options) {
			d.config.WhisperModel = options[index]
		}
	})
}

func (d *ConfigDialog) editLanguage() {
	options := []string{
		"Auto（自動検出）", "日本語", "English", "中文（简体）", "한국어",
		"Español", "Français", "Deutsch", "Русский", "العربية",
		"हिन्दी", "Italiano", "Português",
	}
	codes := []string{
		"auto", "ja", "en", "zh", "ko", "es", "fr", "de", "ru", "ar", "hi", "it", "pt",
	}
	currentIndex := 0
	for i, code := range codes {
		if code == d.config.Language {
			currentIndex = i
			break
		}
	}

	d.showDropDownModal("音声認識言語", options, currentIndex, func(index int) {
		if index >= 0 && index < len(codes) {
			d.config.Language = codes[index]
		}
	})
}

func (d *ConfigDialog) editScanInterval() {
	d.showInputModal("スキャン間隔（分）", strconv.Itoa(d.config.ScanIntervalMinutes), tview.InputFieldInteger, func(text string) {
		if interval, err := strconv.Atoi(text); err == nil && interval > 0 {
			d.config.ScanIntervalMinutes = interval
		}
	})
}

func (d *ConfigDialog) editInputDir() {
	d.showInputModal("入力フォルダ", config.GetRelativePath(d.config.InputDir), nil, func(text string) {
		d.config.InputDir = text
	})
}

func (d *ConfigDialog) editOutputDir() {
	d.showInputModal("出力フォルダ", config.GetRelativePath(d.config.OutputDir), nil, func(text string) {
		d.config.OutputDir = text
	})
}

func (d *ConfigDialog) editArchiveDir() {
	d.showInputModal("アーカイブフォルダ", config.GetRelativePath(d.config.ArchiveDir), nil, func(text string) {
		d.config.ArchiveDir = text
	})
}

func (d *ConfigDialog) editLLMEnabled() {
	d.showBooleanModal("AI要約を有効化", d.config.LLMSummaryEnabled, func(checked bool) {
		d.config.LLMSummaryEnabled = checked
	})
}

func (d *ConfigDialog) editLLMAPIKey() {
	d.showInputModal("APIキー", d.config.LLMAPIKey, nil, func(text string) {
		d.config.LLMAPIKey = text
	})
}

func (d *ConfigDialog) editLLMModel() {
	options := []string{"gpt-4o", "gpt-4-turbo", "gpt-3.5-turbo"}
	currentIndex := 0
	for i, model := range options {
		if model == d.config.LLMModel {
			currentIndex = i
			break
		}
	}

	d.showDropDownModal("LLMモデル", options, currentIndex, func(index int) {
		if index >= 0 && index < len(options) {
			d.config.LLMModel = options[index]
		}
	})
}

func (d *ConfigDialog) editRecordingDevice() {
	d.showInputModal("録音デバイス", d.config.RecordingDeviceName, nil, func(text string) {
		d.config.RecordingDeviceName = text
	})
}

func (d *ConfigDialog) editDualRecording() {
	d.showBooleanModal("デュアル録音", d.config.DualRecordingEnabled, func(checked bool) {
		d.config.DualRecordingEnabled = checked
	})
}

func (d *ConfigDialog) editComputeType() {
	options := []string{"int8", "int8_float16", "int16", "float16", "float32"}
	currentIndex := 0
	for i, ctype := range options {
		if ctype == d.config.ComputeType {
			currentIndex = i
			break
		}
	}

	d.showDropDownModal("計算精度", options, currentIndex, func(index int) {
		if index >= 0 && index < len(options) {
			d.config.ComputeType = options[index]
		}
	})
}

func (d *ConfigDialog) editOutputFormat() {
	options := []string{"txt", "vtt", "srt", "tsv", "json"}
	currentIndex := 0
	for i, format := range options {
		if format == d.config.OutputFormat {
			currentIndex = i
			break
		}
	}

	d.showDropDownModal("出力形式", options, currentIndex, func(index int) {
		if index >= 0 && index < len(options) {
			d.config.OutputFormat = options[index]
		}
	})
}

func (d *ConfigDialog) editMaxCPU() {
	d.showInputModal("最大CPU使用率 (%)", strconv.Itoa(d.config.MaxCpuPercent), tview.InputFieldInteger, func(text string) {
		if maxCpu, err := strconv.Atoi(text); err == nil && maxCpu >= 1 && maxCpu <= 100 {
			d.config.MaxCpuPercent = maxCpu
		}
	})
}

// Modal dialog helpers

func (d *ConfigDialog) showDropDownModal(title string, options []string, currentIndex int, onSave func(int)) {
	dropdown := tview.NewDropDown().
		SetLabel(title + ": ").
		SetOptions(options, nil).
		SetCurrentOption(currentIndex)

	form := tview.NewForm().
		AddFormItem(dropdown).
		AddButton("保存", func() {
			index, _ := dropdown.GetCurrentOption()
			onSave(index)
			d.pages.RemovePage("modal")
			d.app.SetFocus(d.settingsList)
		}).
		AddButton("キャンセル", func() {
			d.pages.RemovePage("modal")
			d.app.SetFocus(d.settingsList)
		})

	form.SetBorder(true).
		SetTitle(" " + title + " ").
		SetTitleAlign(tview.AlignCenter)

	// Center the form
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 10, 1, true).
			AddItem(nil, 0, 1, false), 60, 1, true).
		AddItem(nil, 0, 1, false)

	d.pages.AddPage("modal", flex, true, true)
	d.app.SetFocus(form)
}

func (d *ConfigDialog) showInputModal(title string, currentValue string, acceptFunc func(textToCheck string, lastChar rune) bool, onSave func(string)) {
	inputField := tview.NewInputField().
		SetLabel(title + ": ").
		SetText(currentValue).
		SetFieldWidth(50)

	if acceptFunc != nil {
		inputField.SetAcceptanceFunc(acceptFunc)
	}

	form := tview.NewForm().
		AddFormItem(inputField).
		AddButton("保存", func() {
			onSave(inputField.GetText())
			d.pages.RemovePage("modal")
			d.app.SetFocus(d.settingsList)
		}).
		AddButton("キャンセル", func() {
			d.pages.RemovePage("modal")
			d.app.SetFocus(d.settingsList)
		})

	form.SetBorder(true).
		SetTitle(" " + title + " ").
		SetTitleAlign(tview.AlignCenter)

	// Center the form
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 8, 1, true).
			AddItem(nil, 0, 1, false), 60, 1, true).
		AddItem(nil, 0, 1, false)

	d.pages.AddPage("modal", flex, true, true)
	d.app.SetFocus(inputField)
}

func (d *ConfigDialog) showBooleanModal(title string, currentValue bool, onSave func(bool)) {
	checkbox := tview.NewCheckbox().
		SetLabel(title + ": ").
		SetChecked(currentValue)

	form := tview.NewForm().
		AddFormItem(checkbox).
		AddButton("保存", func() {
			onSave(checkbox.IsChecked())
			d.pages.RemovePage("modal")
			d.app.SetFocus(d.settingsList)
		}).
		AddButton("キャンセル", func() {
			d.pages.RemovePage("modal")
			d.app.SetFocus(d.settingsList)
		})

	form.SetBorder(true).
		SetTitle(" " + title + " ").
		SetTitleAlign(tview.AlignCenter)

	// Center the form
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 8, 1, true).
			AddItem(nil, 0, 1, false), 60, 1, true).
		AddItem(nil, 0, 1, false)

	d.pages.AddPage("modal", flex, true, true)
	d.app.SetFocus(form)
}

// saveConfig saves the configuration to file
func (d *ConfigDialog) saveConfig() {
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
