package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type Config struct {
	WhisperModel        string `json:"whisper_model"`
	Language            string `json:"language"`
	UILanguage          string `json:"ui_language"`
	ScanIntervalMinutes int    `json:"scan_interval_minutes"`
	MaxCpuPercent       int    `json:"max_cpu_percent"`
	ComputeType         string `json:"compute_type"`
	UseColors           bool   `json:"use_colors"`
	OutputFormat        string `json:"output_format"`
	InputDir            string `json:"input_dir"`
	OutputDir           string `json:"output_dir"`
	ArchiveDir          string `json:"archive_dir"`
}

func GetDefaultConfig() *Config {
	return &Config{
		WhisperModel:        "medium",
		Language:            "ja",
		UILanguage:          "en",
		ScanIntervalMinutes: 10,
		MaxCpuPercent:       95,
		ComputeType:         "int8",
		UseColors:           true,
		OutputFormat:        "txt",
		InputDir:            "./input",
		OutputDir:           "./output",
		ArchiveDir:          "./archive",
	}
}

func LoadConfig(configPath string, logger *log.Logger) *Config {
	// Default config
	config := GetDefaultConfig()

	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Printf("[INFO] Config file not found, using defaults")
			return config
		}
		logger.Printf("[ERROR] Failed to load config: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(config); err != nil {
		logger.Printf("[ERROR] Failed to parse config: %v", err)
		os.Exit(1)
	}

	return config
}

func SaveConfig(config *Config, configPath string) error {
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}

func ConfigureSettings(config *Config, configPath string, logger *log.Logger) {
	reader := bufio.NewReader(os.Stdin)
	modified := false
	msg := getMessages(config)

	for {
		fmt.Printf("\n=== %s ===\n", msg.ConfigTitle)
		fmt.Printf("1. %s: %s\n", msg.WhisperModel, config.WhisperModel)
		fmt.Printf("2. %s: %s\n", msg.Language, config.Language)
		fmt.Printf("3. %s: %s\n", msg.UILanguage, config.UILanguage)
		fmt.Printf("4. %s: %d %s\n", msg.ScanInterval, config.ScanIntervalMinutes, msg.Minutes)
		fmt.Printf("5. %s: %d%%\n", msg.MaxCPUPercent, config.MaxCpuPercent)
		fmt.Printf("6. %s: %s\n", msg.ComputeType, config.ComputeType)
		fmt.Printf("7. %s: %t\n", msg.UseColors, config.UseColors)
		fmt.Printf("8. %s: %s\n", msg.UIMode, config.UIMode)
		fmt.Printf("9. %s: %s\n", msg.OutputFormat, config.OutputFormat)
		fmt.Printf("10. %s: %s\n", msg.InputDirectory, config.InputDir)
		fmt.Printf("11. %s: %s\n", msg.OutputDirectory, config.OutputDir)
		fmt.Printf("12. %s: %s\n", msg.ArchiveDirectory, config.ArchiveDir)
		fmt.Printf("r. %s\n", msg.ResetDefaults)
		fmt.Printf("s. %s\n", msg.SaveAndExit)
		fmt.Printf("q. %s\n", msg.QuitWithoutSave)
		fmt.Printf("\n%s (1-12, r, s, q): ", msg.SelectOption)

		input, _ := reader.ReadString('\n')
		choice := strings.TrimSpace(input)

		switch choice {
		case "1":
			if configureWhisperModel(config, reader) {
				modified = true
			}
		case "2":
			if configureLanguage(config, reader) {
				modified = true
			}
		case "3":
			if configureUILanguage(config, reader) {
				modified = true
			}
		case "4":
			if configureScanInterval(config, reader) {
				modified = true
			}
		case "5":
			if configureMaxCpuPercent(config, reader) {
				modified = true
			}
		case "6":
			if configureComputeType(config, reader) {
				modified = true
			}
		case "7":
			if configureUseColors(config, reader) {
				modified = true
			}
		case "8":
			if configureUIMode(config, reader) {
				modified = true
			}
		case "9":
			if configureOutputFormat(config, reader) {
				modified = true
			}
		case "10":
			if configureInputDir(config, reader) {
				modified = true
			}
		case "11":
			if configureOutputDir(config, reader) {
				modified = true
			}
		case "12":
			if configureArchiveDir(config, reader) {
				modified = true
			}
		case "r":
			if resetToDefaults(config, reader) {
				modified = true
			}
		case "s":
			if modified {
				if err := SaveConfig(config, configPath); err != nil {
					fmt.Printf(msg.ConfigSaveError+"\n", err)
				} else {
					fmt.Println(msg.ConfigSaved)
				}
			} else {
				fmt.Println(msg.NoChanges)
			}
			return
		case "q":
			if modified {
				fmt.Printf("%s ", msg.UnsavedChanges)
				confirm, _ := reader.ReadString('\n')
				if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
					continue
				}
			}
			return
		default:
			fmt.Println(msg.InvalidOption)
		}
	}
}

func configureWhisperModel(config *Config, reader *bufio.Reader) bool {
	models := []string{
		"tiny", "tiny.en",
		"base", "base.en",
		"small", "small.en",
		"medium", "medium.en",
		"large", "large-v1", "large-v2", "large-v3",
	}
	msg := getMessages(config)

	fmt.Println("\nAvailable Whisper models:")
	for i, model := range models {
		fmt.Printf("%d. %s", i+1, model)
		if model == config.WhisperModel {
			fmt.Printf(" (%s)", msg.Current)
		}
		fmt.Println()
	}
	fmt.Printf(msg.SelectModel+" ", len(models))

	input, _ := reader.ReadString('\n')
	choice := strings.TrimSpace(input)

	if choice == "" {
		return false
	}

	if idx, err := strconv.Atoi(choice); err == nil && idx >= 1 && idx <= len(models) {
		config.WhisperModel = models[idx-1]
		msg2 := getMessages(config)
		fmt.Printf(msg2.ModelSet+"\n", config.WhisperModel)
		return true
	}

	msg2 := getMessages(config)
	fmt.Println(msg2.InvalidOption)
	return false
}

func configureLanguage(config *Config, reader *bufio.Reader) bool {
	msg := getMessages(config)
	fmt.Printf("%s %s: %s\n", msg.Current, msg.Language, config.Language)
	fmt.Printf("%s ", msg.EnterLanguage)

	input, _ := reader.ReadString('\n')
	newLang := strings.TrimSpace(input)

	if newLang == "" {
		return false
	}

	config.Language = newLang
	fmt.Printf(msg.LanguageSet+"\n", config.Language)
	return true
}

func configureScanInterval(config *Config, reader *bufio.Reader) bool {
	msg := getMessages(config)
	fmt.Printf("%s %s: %d %s\n", msg.Current, msg.ScanInterval, config.ScanIntervalMinutes, msg.Minutes)
	fmt.Printf("%s ", msg.EnterInterval)

	input, _ := reader.ReadString('\n')
	newInterval := strings.TrimSpace(input)

	if newInterval == "" {
		return false
	}

	if interval, err := strconv.Atoi(newInterval); err == nil && interval > 0 {
		config.ScanIntervalMinutes = interval
		fmt.Printf(msg.IntervalSet+"\n", config.ScanIntervalMinutes)
		return true
	}

	fmt.Println(msg.InvalidInput)
	return false
}

func configureMaxCpuPercent(config *Config, reader *bufio.Reader) bool {
	msg := getMessages(config)
	fmt.Printf("%s %s: %d%%\n", msg.Current, msg.MaxCPUPercent, config.MaxCpuPercent)
	fmt.Printf("%s ", msg.EnterCPU)

	input, _ := reader.ReadString('\n')
	newPercent := strings.TrimSpace(input)

	if newPercent == "" {
		return false
	}

	if percent, err := strconv.Atoi(newPercent); err == nil && percent >= 1 && percent <= 100 {
		config.MaxCpuPercent = percent
		fmt.Printf(msg.CPUSet+"\n", config.MaxCpuPercent)
		return true
	}

	fmt.Println(msg.InvalidInput)
	return false
}

func configureComputeType(config *Config, reader *bufio.Reader) bool {
	types := []string{"int8", "int8_float16", "int16", "float16", "float32"}
	msg := getMessages(config)

	fmt.Println("\nAvailable compute types:")
	for i, ctype := range types {
		fmt.Printf("%d. %s", i+1, ctype)
		if ctype == config.ComputeType {
			fmt.Printf(" (%s)", msg.Current)
		}
		fmt.Println()
	}
	fmt.Printf(msg.SelectCompute+" ", len(types))

	input, _ := reader.ReadString('\n')
	choice := strings.TrimSpace(input)

	if choice == "" {
		return false
	}

	if idx, err := strconv.Atoi(choice); err == nil && idx >= 1 && idx <= len(types) {
		config.ComputeType = types[idx-1]
		fmt.Printf(msg.ComputeSet+"\n", config.ComputeType)
		return true
	}

	fmt.Println(msg.InvalidOption)
	return false
}

func configureUILanguage(config *Config, reader *bufio.Reader) bool {
	languages := []string{"en", "ja"}
	languageNames := []string{"English", "日本語"}
	msg := getMessages(config)

	fmt.Println("\nAvailable UI languages:")
	for i, lang := range languages {
		fmt.Printf("%d. %s (%s)", i+1, lang, languageNames[i])
		if lang == config.UILanguage {
			fmt.Printf(" (%s)", msg.Current)
		}
		fmt.Println()
	}
	fmt.Printf("%s ", msg.SelectUILang)

	input, _ := reader.ReadString('\n')
	choice := strings.TrimSpace(input)

	if choice == "" {
		return false
	}

	if idx, err := strconv.Atoi(choice); err == nil && idx >= 1 && idx <= len(languages) {
		config.UILanguage = languages[idx-1]
		fmt.Printf(msg.UILanguageSet+"\n", config.UILanguage)
		return true
	}

	fmt.Println(msg.InvalidOption)
	return false
}

func configureUseColors(config *Config, reader *bufio.Reader) bool {
	msg := getMessages(config)
	fmt.Printf("%s %s: %t\n", msg.Current, msg.UseColors, config.UseColors)
	fmt.Printf("%s ", msg.EnableColors)

	input, _ := reader.ReadString('\n')
	choice := strings.ToLower(strings.TrimSpace(input))

	if choice == "" {
		return false
	}

	if choice == "y" || choice == "yes" {
		config.UseColors = true
		fmt.Println(msg.ColorsEnabled)
		return true
	} else if choice == "n" || choice == "no" {
		config.UseColors = false
		fmt.Println(msg.ColorsDisabled)
		return true
	}

	fmt.Println(msg.InvalidInput)
	return false
}

func configureUIMode(config *Config, reader *bufio.Reader) bool {
	modes := []string{"simple", "enhanced"}
	msg := getMessages(config)

	fmt.Println("\nAvailable UI modes:")
	for i, mode := range modes {
		fmt.Printf("%d. %s", i+1, mode)
		if mode == config.UIMode {
			fmt.Printf(" (%s)", msg.Current)
		}
		fmt.Println()
	}
	fmt.Printf("%s ", msg.SelectUIMode)

	input, _ := reader.ReadString('\n')
	choice := strings.TrimSpace(input)

	if choice == "" {
		return false
	}

	if idx, err := strconv.Atoi(choice); err == nil && idx >= 1 && idx <= len(modes) {
		config.UIMode = modes[idx-1]
		fmt.Printf(msg.UIModeSet+"\n", config.UIMode)
		return true
	}

	fmt.Println(msg.InvalidOption)
	return false
}

func configureOutputFormat(config *Config, reader *bufio.Reader) bool {
	formats := []string{"txt", "vtt", "srt", "tsv", "json"}
	msg := getMessages(config)

	fmt.Println("\nAvailable output formats:")
	for i, format := range formats {
		fmt.Printf("%d. %s", i+1, format)
		if format == config.OutputFormat {
			fmt.Printf(" (%s)", msg.Current)
		}
		fmt.Println()
	}
	fmt.Printf(msg.SelectFormat+" ", len(formats))

	input, _ := reader.ReadString('\n')
	choice := strings.TrimSpace(input)

	if choice == "" {
		return false
	}

	if idx, err := strconv.Atoi(choice); err == nil && idx >= 1 && idx <= len(formats) {
		config.OutputFormat = formats[idx-1]
		fmt.Printf(msg.FormatSet+"\n", config.OutputFormat)
		return true
	}

	fmt.Println(msg.InvalidOption)
	return false
}

func configureInputDir(config *Config, reader *bufio.Reader) bool {
	msg := getMessages(config)
	fmt.Printf("%s %s: %s\n", msg.Current, msg.InputDirectory, config.InputDir)
	fmt.Printf("%s ", msg.SelectFolder)

	input, _ := reader.ReadString('\n')
	newDir := strings.TrimSpace(input)

	if newDir == "" {
		// Use folder selection dialog
		selectedDir, err := selectFolder("Select Input Directory")
		if err != nil {
			fmt.Printf(msg.FolderSelectFail+"\n", err)
			return false
		}
		newDir = selectedDir
	}

	config.InputDir = newDir
	fmt.Printf(msg.InputDirSet+"\n", config.InputDir)
	return true
}

func configureOutputDir(config *Config, reader *bufio.Reader) bool {
	msg := getMessages(config)
	fmt.Printf("%s %s: %s\n", msg.Current, msg.OutputDirectory, config.OutputDir)
	fmt.Printf("%s ", msg.SelectFolder)

	input, _ := reader.ReadString('\n')
	newDir := strings.TrimSpace(input)

	if newDir == "" {
		// Use folder selection dialog
		selectedDir, err := selectFolder("Select Output Directory")
		if err != nil {
			fmt.Printf(msg.FolderSelectFail+"\n", err)
			return false
		}
		newDir = selectedDir
	}

	config.OutputDir = newDir
	fmt.Printf(msg.OutputDirSet+"\n", config.OutputDir)
	return true
}

func configureArchiveDir(config *Config, reader *bufio.Reader) bool {
	msg := getMessages(config)
	fmt.Printf("%s %s: %s\n", msg.Current, msg.ArchiveDirectory, config.ArchiveDir)
	fmt.Printf("%s ", msg.SelectFolder)

	input, _ := reader.ReadString('\n')
	newDir := strings.TrimSpace(input)

	if newDir == "" {
		// Use folder selection dialog
		selectedDir, err := selectFolder("Select Archive Directory")
		if err != nil {
			fmt.Printf(msg.FolderSelectFail+"\n", err)
			return false
		}
		newDir = selectedDir
	}

	config.ArchiveDir = newDir
	fmt.Printf(msg.ArchiveDirSet+"\n", config.ArchiveDir)
	return true
}

func resetToDefaults(config *Config, reader *bufio.Reader) bool {
	msg := getMessages(config)
	fmt.Printf("%s ", msg.ResetConfirm)

	input, _ := reader.ReadString('\n')
	choice := strings.ToLower(strings.TrimSpace(input))

	if choice == "y" || choice == "yes" {
		defaultConfig := GetDefaultConfig()
		*config = *defaultConfig
		fmt.Println(msg.ConfigReset)
		return true
	}

	return false
}

func selectFolder(title string) (string, error) {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("powershell", "-Command",
			"Add-Type -AssemblyName System.Windows.Forms; "+
				"$folder = New-Object System.Windows.Forms.FolderBrowserDialog; "+
				"$folder.Description = '"+title+"'; "+
				"if ($folder.ShowDialog() -eq 'OK') { $folder.SelectedPath }")
	case "darwin":
		cmd = exec.Command("osascript", "-e",
			"POSIX path of (choose folder with prompt \""+title+"\")")
	default:
		return "", fmt.Errorf("folder selection not supported on this platform")
	}
	
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	
	selectedPath := strings.TrimSpace(string(output))
	if selectedPath == "" {
		return "", fmt.Errorf("no folder selected")
	}
	
	return selectedPath, nil
}

// Messages contains all UI text strings
type Messages struct {
	// Config menu
	ConfigTitle     string
	WhisperModel    string
	Language        string
	UILanguage      string
	ScanInterval    string
	MaxCPUPercent   string
	ComputeType     string
	UseColors       string
	UIMode          string
	OutputFormat    string
	InputDirectory  string
	OutputDirectory string
	ArchiveDirectory string
	ResetDefaults   string
	SaveAndExit     string
	QuitWithoutSave string
	SelectOption    string
	Minutes         string
	Current         string
	
	// Config prompts
	SelectModel     string
	EnterLanguage   string
	SelectUILang    string
	EnterInterval   string
	EnterCPU        string
	SelectCompute   string
	EnableColors    string
	SelectUIMode    string
	SelectFormat    string
	SelectFolder    string
	ResetConfirm    string
	UnsavedChanges  string
	
	// Config messages
	ModelSet        string
	LanguageSet     string
	UILanguageSet   string
	IntervalSet     string
	CPUSet          string
	ComputeSet      string
	ColorsEnabled   string
	ColorsDisabled  string
	UIModeSet       string
	FormatSet       string
	InputDirSet     string
	OutputDirSet    string
	ArchiveDirSet   string
	ConfigReset     string
	ConfigSaved     string
	NoChanges       string
	InvalidOption   string
	InvalidInput    string
	FolderSelectFail string
	ConfigSaveError string
}

var messagesEN = Messages{
	// Config menu
	ConfigTitle:     "KoeMoji-Go Configuration",
	WhisperModel:    "Whisper Model",
	Language:        "Language",
	UILanguage:      "UI Language",
	ScanInterval:    "Scan Interval",
	MaxCPUPercent:   "Max CPU Percent",
	ComputeType:     "Compute Type",
	UseColors:       "Use Colors",
	UIMode:          "UI Mode",
	OutputFormat:    "Output Format",
	InputDirectory:  "Input Directory",
	OutputDirectory: "Output Directory",
	ArchiveDirectory: "Archive Directory",
	ResetDefaults:   "Reset to defaults",
	SaveAndExit:     "Save and exit",
	QuitWithoutSave: "Quit without saving",
	SelectOption:    "Select option",
	Minutes:         "minutes",
	Current:         "current",
	
	// Config prompts
	SelectModel:     "Select model (1-%d) or press Enter to keep current:",
	EnterLanguage:   "Enter new language code (e.g., ja, en, zh) or press Enter to keep current:",
	SelectUILang:    "Select UI language (1-2) or press Enter to keep current:",
	EnterInterval:   "Enter new scan interval (minutes) or press Enter to keep current:",
	EnterCPU:        "Enter new max CPU percent (1-100) or press Enter to keep current:",
	SelectCompute:   "Select compute type (1-%d) or press Enter to keep current:",
	EnableColors:    "Enable colors? (y/n) or press Enter to keep current:",
	SelectUIMode:    "Select UI mode (1-2) or press Enter to keep current:",
	SelectFormat:    "Select output format (1-%d) or press Enter to keep current:",
	SelectFolder:    "Press Enter to select folder with dialog, or type path manually:",
	ResetConfirm:    "Are you sure you want to reset all settings to defaults? (y/N):",
	UnsavedChanges:  "You have unsaved changes. Are you sure you want to quit? (y/N):",
	
	// Config messages
	ModelSet:        "Whisper model set to: %s",
	LanguageSet:     "Language set to: %s",
	UILanguageSet:   "UI language set to: %s",
	IntervalSet:     "Scan interval set to: %d minutes",
	CPUSet:          "Max CPU percent set to: %d%%",
	ComputeSet:      "Compute type set to: %s",
	ColorsEnabled:   "Colors enabled",
	ColorsDisabled:  "Colors disabled",
	UIModeSet:       "UI mode set to: %s",
	FormatSet:       "Output format set to: %s",
	InputDirSet:     "Input directory set to: %s",
	OutputDirSet:    "Output directory set to: %s",
	ArchiveDirSet:   "Archive directory set to: %s",
	ConfigReset:     "Configuration reset to defaults.",
	ConfigSaved:     "Configuration saved successfully!",
	NoChanges:       "No changes to save.",
	InvalidOption:   "Invalid option. Please try again.",
	InvalidInput:    "Invalid input.",
	FolderSelectFail: "Folder selection failed: %v",
	ConfigSaveError: "Failed to save config: %v",
}

var messagesJA = Messages{
	// Config menu
	ConfigTitle:     "KoeMoji-Go 設定",
	WhisperModel:    "Whisperモデル",
	Language:        "認識言語",
	UILanguage:      "UI言語",
	ScanInterval:    "スキャン間隔",
	MaxCPUPercent:   "最大CPU使用率",
	ComputeType:     "計算タイプ",
	UseColors:       "色を使用",
	UIMode:          "UIモード",
	OutputFormat:    "出力フォーマット",
	InputDirectory:  "入力ディレクトリ",
	OutputDirectory: "出力ディレクトリ",
	ArchiveDirectory: "アーカイブディレクトリ",
	ResetDefaults:   "デフォルトに戻す",
	SaveAndExit:     "保存して終了",
	QuitWithoutSave: "保存せずに終了",
	SelectOption:    "オプションを選択",
	Minutes:         "分",
	Current:         "現在",
	
	// Config prompts
	SelectModel:     "モデルを選択 (1-%d) またはEnterで現在の設定を維持:",
	EnterLanguage:   "新しい言語コード (例: ja, en, zh) を入力またはEnterで現在の設定を維持:",
	SelectUILang:    "UI言語を選択 (1-2) またはEnterで現在の設定を維持:",
	EnterInterval:   "新しいスキャン間隔（分）を入力またはEnterで現在の設定を維持:",
	EnterCPU:        "新しい最大CPU使用率 (1-100) を入力またはEnterで現在の設定を維持:",
	SelectCompute:   "計算タイプを選択 (1-%d) またはEnterで現在の設定を維持:",
	EnableColors:    "色を有効にしますか？ (y/n) またはEnterで現在の設定を維持:",
	SelectUIMode:    "UIモードを選択 (1-2) またはEnterで現在の設定を維持:",
	SelectFormat:    "出力フォーマットを選択 (1-%d) またはEnterで現在の設定を維持:",
	SelectFolder:    "Enterでフォルダ選択ダイアログを開く、または手動でパスを入力:",
	ResetConfirm:    "本当にすべての設定をデフォルトに戻しますか？ (y/N):",
	UnsavedChanges:  "未保存の変更があります。本当に終了しますか？ (y/N):",
	
	// Config messages
	ModelSet:        "Whisperモデルを設定: %s",
	LanguageSet:     "言語を設定: %s",
	UILanguageSet:   "UI言語を設定: %s",
	IntervalSet:     "スキャン間隔を設定: %d分",
	CPUSet:          "最大CPU使用率を設定: %d%%",
	ComputeSet:      "計算タイプを設定: %s",
	ColorsEnabled:   "色を有効にしました",
	ColorsDisabled:  "色を無効にしました",
	UIModeSet:       "UIモードを設定: %s",
	FormatSet:       "出力フォーマットを設定: %s",
	InputDirSet:     "入力ディレクトリを設定: %s",
	OutputDirSet:    "出力ディレクトリを設定: %s",
	ArchiveDirSet:   "アーカイブディレクトリを設定: %s",
	ConfigReset:     "設定をデフォルトに戻しました。",
	ConfigSaved:     "設定を保存しました！",
	NoChanges:       "変更はありません。",
	InvalidOption:   "無効なオプションです。もう一度お試しください。",
	InvalidInput:    "無効な入力です。",
	FolderSelectFail: "フォルダ選択に失敗: %v",
	ConfigSaveError: "設定の保存に失敗: %v",
}

// getMessages returns the messages for the current UI language
func getMessages(config *Config) *Messages {
	if config != nil && config.UILanguage == "ja" {
		return &messagesJA
	}
	return &messagesEN
}