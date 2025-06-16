package main

import (
	"bufio"
	"encoding/json"
	"fmt"
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
	UIMode              string `json:"ui_mode"`
	OutputFormat        string `json:"output_format"`
	InputDir            string `json:"input_dir"`
	OutputDir           string `json:"output_dir"`
	ArchiveDir          string `json:"archive_dir"`
}

func getDefaultConfig() *Config {
	return &Config{
		WhisperModel:        "medium",
		Language:            "ja",
		UILanguage:          "en",
		ScanIntervalMinutes: 10,
		MaxCpuPercent:       95,
		ComputeType:         "int8",
		UseColors:           true,
		UIMode:              "enhanced",
		OutputFormat:        "txt",
		InputDir:            "./input",
		OutputDir:           "./output",
		ArchiveDir:          "./archive",
	}
}

func (app *App) loadConfig(configPath string) {
	// Default config
	app.config = getDefaultConfig()

	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			app.logInfo("Config file not found, using defaults")
			return
		}
		app.logError("Failed to load config: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(app.config); err != nil {
		app.logError("Failed to parse config: %v", err)
		os.Exit(1)
	}
}

func (app *App) configureSettings(configPath string) {
	reader := bufio.NewReader(os.Stdin)
	modified := false
	msg := app.getMessages()

	for {
		fmt.Printf("\n=== %s ===\n", msg.ConfigTitle)
		fmt.Printf("1. %s: %s\n", msg.WhisperModel, app.config.WhisperModel)
		fmt.Printf("2. %s: %s\n", msg.Language, app.config.Language)
		fmt.Printf("3. %s: %s\n", msg.UILanguage, app.config.UILanguage)
		fmt.Printf("4. %s: %d %s\n", msg.ScanInterval, app.config.ScanIntervalMinutes, msg.Minutes)
		fmt.Printf("5. %s: %d%%\n", msg.MaxCPUPercent, app.config.MaxCpuPercent)
		fmt.Printf("6. %s: %s\n", msg.ComputeType, app.config.ComputeType)
		fmt.Printf("7. %s: %t\n", msg.UseColors, app.config.UseColors)
		fmt.Printf("8. %s: %s\n", msg.UIMode, app.config.UIMode)
		fmt.Printf("9. %s: %s\n", msg.OutputFormat, app.config.OutputFormat)
		fmt.Printf("10. %s: %s\n", msg.InputDirectory, app.config.InputDir)
		fmt.Printf("11. %s: %s\n", msg.OutputDirectory, app.config.OutputDir)
		fmt.Printf("12. %s: %s\n", msg.ArchiveDirectory, app.config.ArchiveDir)
		fmt.Printf("r. %s\n", msg.ResetDefaults)
		fmt.Printf("s. %s\n", msg.SaveAndExit)
		fmt.Printf("q. %s\n", msg.QuitWithoutSave)
		fmt.Printf("\n%s (1-12, r, s, q): ", msg.SelectOption)

		input, _ := reader.ReadString('\n')
		choice := strings.TrimSpace(input)

		switch choice {
		case "1":
			if app.configureWhisperModel(reader) {
				modified = true
			}
		case "2":
			if app.configureLanguage(reader) {
				modified = true
			}
		case "3":
			if app.configureUILanguage(reader) {
				modified = true
			}
		case "4":
			if app.configureScanInterval(reader) {
				modified = true
			}
		case "5":
			if app.configureMaxCpuPercent(reader) {
				modified = true
			}
		case "6":
			if app.configureComputeType(reader) {
				modified = true
			}
		case "7":
			if app.configureUseColors(reader) {
				modified = true
			}
		case "8":
			if app.configureUIMode(reader) {
				modified = true
			}
		case "9":
			if app.configureOutputFormat(reader) {
				modified = true
			}
		case "10":
			if app.configureInputDir(reader) {
				modified = true
			}
		case "11":
			if app.configureOutputDir(reader) {
				modified = true
			}
		case "12":
			if app.configureArchiveDir(reader) {
				modified = true
			}
		case "r":
			if app.resetToDefaults(reader) {
				modified = true
			}
		case "s":
			if modified {
				app.saveConfig(configPath)
				fmt.Println(msg.ConfigSaved)
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

func (app *App) configureWhisperModel(reader *bufio.Reader) bool {
	models := []string{
		"tiny", "tiny.en",
		"base", "base.en",
		"small", "small.en",
		"medium", "medium.en",
		"large", "large-v1", "large-v2", "large-v3",
	}
	msg := app.getMessages()

	fmt.Println("\nAvailable Whisper models:")
	for i, model := range models {
		fmt.Printf("%d. %s", i+1, model)
		if model == app.config.WhisperModel {
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
		app.config.WhisperModel = models[idx-1]
		msg2 := app.getMessages()
		fmt.Printf(msg2.ModelSet+"\n", app.config.WhisperModel)
		return true
	}

	msg2 := app.getMessages()
	fmt.Println(msg2.InvalidOption)
	return false
}

func (app *App) configureLanguage(reader *bufio.Reader) bool {
	msg := app.getMessages()
	fmt.Printf("%s %s: %s\n", msg.Current, msg.Language, app.config.Language)
	fmt.Printf("%s ", msg.EnterLanguage)

	input, _ := reader.ReadString('\n')
	newLang := strings.TrimSpace(input)

	if newLang == "" {
		return false
	}

	app.config.Language = newLang
	fmt.Printf(msg.LanguageSet+"\n", app.config.Language)
	return true
}

func (app *App) configureScanInterval(reader *bufio.Reader) bool {
	msg := app.getMessages()
	fmt.Printf("%s %s: %d %s\n", msg.Current, msg.ScanInterval, app.config.ScanIntervalMinutes, msg.Minutes)
	fmt.Printf("%s ", msg.EnterInterval)

	input, _ := reader.ReadString('\n')
	newInterval := strings.TrimSpace(input)

	if newInterval == "" {
		return false
	}

	if interval, err := strconv.Atoi(newInterval); err == nil && interval > 0 {
		app.config.ScanIntervalMinutes = interval
		fmt.Printf(msg.IntervalSet+"\n", app.config.ScanIntervalMinutes)
		return true
	}

	fmt.Println(msg.InvalidInput)
	return false
}

func (app *App) configureMaxCpuPercent(reader *bufio.Reader) bool {
	msg := app.getMessages()
	fmt.Printf("%s %s: %d%%\n", msg.Current, msg.MaxCPUPercent, app.config.MaxCpuPercent)
	fmt.Printf("%s ", msg.EnterCPU)

	input, _ := reader.ReadString('\n')
	newPercent := strings.TrimSpace(input)

	if newPercent == "" {
		return false
	}

	if percent, err := strconv.Atoi(newPercent); err == nil && percent >= 1 && percent <= 100 {
		app.config.MaxCpuPercent = percent
		fmt.Printf(msg.CPUSet+"\n", app.config.MaxCpuPercent)
		return true
	}

	fmt.Println(msg.InvalidInput)
	return false
}

func (app *App) configureComputeType(reader *bufio.Reader) bool {
	types := []string{"int8", "int8_float16", "int16", "float16", "float32"}
	msg := app.getMessages()

	fmt.Println("\nAvailable compute types:")
	for i, ctype := range types {
		fmt.Printf("%d. %s", i+1, ctype)
		if ctype == app.config.ComputeType {
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
		app.config.ComputeType = types[idx-1]
		fmt.Printf(msg.ComputeSet+"\n", app.config.ComputeType)
		return true
	}

	fmt.Println(msg.InvalidOption)
	return false
}

func (app *App) configureUILanguage(reader *bufio.Reader) bool {
	languages := []string{"en", "ja"}
	languageNames := []string{"English", "日本語"}
	msg := app.getMessages()

	fmt.Println("\nAvailable UI languages:")
	for i, lang := range languages {
		fmt.Printf("%d. %s (%s)", i+1, lang, languageNames[i])
		if lang == app.config.UILanguage {
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
		app.config.UILanguage = languages[idx-1]
		fmt.Printf(msg.UILanguageSet+"\n", app.config.UILanguage)
		return true
	}

	fmt.Println(msg.InvalidOption)
	return false
}

func (app *App) configureUseColors(reader *bufio.Reader) bool {
	msg := app.getMessages()
	fmt.Printf("%s %s: %t\n", msg.Current, msg.UseColors, app.config.UseColors)
	fmt.Printf("%s ", msg.EnableColors)

	input, _ := reader.ReadString('\n')
	choice := strings.ToLower(strings.TrimSpace(input))

	if choice == "" {
		return false
	}

	if choice == "y" || choice == "yes" {
		app.config.UseColors = true
		fmt.Println(msg.ColorsEnabled)
		return true
	} else if choice == "n" || choice == "no" {
		app.config.UseColors = false
		fmt.Println(msg.ColorsDisabled)
		return true
	}

	fmt.Println(msg.InvalidInput)
	return false
}

func (app *App) configureUIMode(reader *bufio.Reader) bool {
	modes := []string{"simple", "enhanced"}
	msg := app.getMessages()

	fmt.Println("\nAvailable UI modes:")
	for i, mode := range modes {
		fmt.Printf("%d. %s", i+1, mode)
		if mode == app.config.UIMode {
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
		app.config.UIMode = modes[idx-1]
		fmt.Printf(msg.UIModeSet+"\n", app.config.UIMode)
		return true
	}

	fmt.Println(msg.InvalidOption)
	return false
}

func (app *App) configureOutputFormat(reader *bufio.Reader) bool {
	formats := []string{"txt", "vtt", "srt", "tsv", "json"}
	msg := app.getMessages()

	fmt.Println("\nAvailable output formats:")
	for i, format := range formats {
		fmt.Printf("%d. %s", i+1, format)
		if format == app.config.OutputFormat {
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
		app.config.OutputFormat = formats[idx-1]
		fmt.Printf(msg.FormatSet+"\n", app.config.OutputFormat)
		return true
	}

	fmt.Println(msg.InvalidOption)
	return false
}

func (app *App) configureInputDir(reader *bufio.Reader) bool {
	msg := app.getMessages()
	fmt.Printf("%s %s: %s\n", msg.Current, msg.InputDirectory, app.config.InputDir)
	fmt.Printf("%s ", msg.SelectFolder)

	input, _ := reader.ReadString('\n')
	newDir := strings.TrimSpace(input)

	if newDir == "" {
		// Use folder selection dialog
		selectedDir, err := app.selectFolder("Select Input Directory")
		if err != nil {
			fmt.Printf(msg.FolderSelectFail+"\n", err)
			return false
		}
		newDir = selectedDir
	}

	app.config.InputDir = newDir
	fmt.Printf(msg.InputDirSet+"\n", app.config.InputDir)
	return true
}

func (app *App) configureOutputDir(reader *bufio.Reader) bool {
	msg := app.getMessages()
	fmt.Printf("%s %s: %s\n", msg.Current, msg.OutputDirectory, app.config.OutputDir)
	fmt.Printf("%s ", msg.SelectFolder)

	input, _ := reader.ReadString('\n')
	newDir := strings.TrimSpace(input)

	if newDir == "" {
		// Use folder selection dialog
		selectedDir, err := app.selectFolder("Select Output Directory")
		if err != nil {
			fmt.Printf(msg.FolderSelectFail+"\n", err)
			return false
		}
		newDir = selectedDir
	}

	app.config.OutputDir = newDir
	fmt.Printf(msg.OutputDirSet+"\n", app.config.OutputDir)
	return true
}

func (app *App) configureArchiveDir(reader *bufio.Reader) bool {
	msg := app.getMessages()
	fmt.Printf("%s %s: %s\n", msg.Current, msg.ArchiveDirectory, app.config.ArchiveDir)
	fmt.Printf("%s ", msg.SelectFolder)

	input, _ := reader.ReadString('\n')
	newDir := strings.TrimSpace(input)

	if newDir == "" {
		// Use folder selection dialog
		selectedDir, err := app.selectFolder("Select Archive Directory")
		if err != nil {
			fmt.Printf(msg.FolderSelectFail+"\n", err)
			return false
		}
		newDir = selectedDir
	}

	app.config.ArchiveDir = newDir
	fmt.Printf(msg.ArchiveDirSet+"\n", app.config.ArchiveDir)
	return true
}

func (app *App) resetToDefaults(reader *bufio.Reader) bool {
	msg := app.getMessages()
	fmt.Printf("%s ", msg.ResetConfirm)

	input, _ := reader.ReadString('\n')
	choice := strings.ToLower(strings.TrimSpace(input))

	if choice == "y" || choice == "yes" {
		defaultConfig := getDefaultConfig()
		*app.config = *defaultConfig
		fmt.Println(msg.ConfigReset)
		return true
	}

	return false
}

func (app *App) selectFolder(title string) (string, error) {
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

func (app *App) saveConfig(configPath string) error {
	file, err := os.Create(configPath)
	if err != nil {
		msg := app.getMessages()
		fmt.Printf(msg.ConfigSaveError+"\n", err)
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(app.config); err != nil {
		msg := app.getMessages()
		fmt.Printf(msg.ConfigSaveError+"\n", err)
		return err
	}

	return nil
}
