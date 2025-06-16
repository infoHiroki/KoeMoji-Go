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

	for {
		fmt.Println("\n=== KoeMoji-Go Configuration ===")
		fmt.Printf("1. Whisper Model: %s\n", app.config.WhisperModel)
		fmt.Printf("2. Language: %s\n", app.config.Language)
		fmt.Printf("3. UI Language: %s\n", app.config.UILanguage)
		fmt.Printf("4. Scan Interval: %d minutes\n", app.config.ScanIntervalMinutes)
		fmt.Printf("5. Max CPU Percent: %d%%\n", app.config.MaxCpuPercent)
		fmt.Printf("6. Compute Type: %s\n", app.config.ComputeType)
		fmt.Printf("7. Use Colors: %t\n", app.config.UseColors)
		fmt.Printf("8. UI Mode: %s\n", app.config.UIMode)
		fmt.Printf("9. Output Format: %s\n", app.config.OutputFormat)
		fmt.Printf("10. Input Directory: %s\n", app.config.InputDir)
		fmt.Printf("11. Output Directory: %s\n", app.config.OutputDir)
		fmt.Printf("12. Archive Directory: %s\n", app.config.ArchiveDir)
		fmt.Println("r. Reset to defaults")
		fmt.Println("s. Save and exit")
		fmt.Println("q. Quit without saving")
		fmt.Print("\nSelect option (1-12, r, s, q): ")

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
				fmt.Println("Configuration saved successfully!")
			} else {
				fmt.Println("No changes to save.")
			}
			return
		case "q":
			if modified {
				fmt.Print("You have unsaved changes. Are you sure you want to quit? (y/N): ")
				confirm, _ := reader.ReadString('\n')
				if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
					continue
				}
			}
			return
		default:
			fmt.Println("Invalid option. Please try again.")
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

	fmt.Println("\nAvailable Whisper models:")
	for i, model := range models {
		fmt.Printf("%d. %s", i+1, model)
		if model == app.config.WhisperModel {
			fmt.Print(" (current)")
		}
		fmt.Println()
	}
	fmt.Print("Select model (1-12) or press Enter to keep current: ")

	input, _ := reader.ReadString('\n')
	choice := strings.TrimSpace(input)

	if choice == "" {
		return false
	}

	if idx, err := strconv.Atoi(choice); err == nil && idx >= 1 && idx <= len(models) {
		app.config.WhisperModel = models[idx-1]
		fmt.Printf("Whisper model set to: %s\n", app.config.WhisperModel)
		return true
	}

	fmt.Println("Invalid selection.")
	return false
}

func (app *App) configureLanguage(reader *bufio.Reader) bool {
	fmt.Printf("Current language: %s\n", app.config.Language)
	fmt.Print("Enter new language code (e.g., ja, en, zh) or press Enter to keep current: ")

	input, _ := reader.ReadString('\n')
	newLang := strings.TrimSpace(input)

	if newLang == "" {
		return false
	}

	app.config.Language = newLang
	fmt.Printf("Language set to: %s\n", app.config.Language)
	return true
}

func (app *App) configureScanInterval(reader *bufio.Reader) bool {
	fmt.Printf("Current scan interval: %d minutes\n", app.config.ScanIntervalMinutes)
	fmt.Print("Enter new scan interval (minutes) or press Enter to keep current: ")

	input, _ := reader.ReadString('\n')
	newInterval := strings.TrimSpace(input)

	if newInterval == "" {
		return false
	}

	if interval, err := strconv.Atoi(newInterval); err == nil && interval > 0 {
		app.config.ScanIntervalMinutes = interval
		fmt.Printf("Scan interval set to: %d minutes\n", app.config.ScanIntervalMinutes)
		return true
	}

	fmt.Println("Invalid input. Please enter a positive number.")
	return false
}

func (app *App) configureMaxCpuPercent(reader *bufio.Reader) bool {
	fmt.Printf("Current max CPU percent: %d%%\n", app.config.MaxCpuPercent)
	fmt.Print("Enter new max CPU percent (1-100) or press Enter to keep current: ")

	input, _ := reader.ReadString('\n')
	newPercent := strings.TrimSpace(input)

	if newPercent == "" {
		return false
	}

	if percent, err := strconv.Atoi(newPercent); err == nil && percent >= 1 && percent <= 100 {
		app.config.MaxCpuPercent = percent
		fmt.Printf("Max CPU percent set to: %d%%\n", app.config.MaxCpuPercent)
		return true
	}

	fmt.Println("Invalid input. Please enter a number between 1 and 100.")
	return false
}

func (app *App) configureComputeType(reader *bufio.Reader) bool {
	types := []string{"int8", "int8_float16", "int16", "float16", "float32"}

	fmt.Println("\nAvailable compute types:")
	for i, ctype := range types {
		fmt.Printf("%d. %s", i+1, ctype)
		if ctype == app.config.ComputeType {
			fmt.Print(" (current)")
		}
		fmt.Println()
	}
	fmt.Print("Select compute type (1-5) or press Enter to keep current: ")

	input, _ := reader.ReadString('\n')
	choice := strings.TrimSpace(input)

	if choice == "" {
		return false
	}

	if idx, err := strconv.Atoi(choice); err == nil && idx >= 1 && idx <= len(types) {
		app.config.ComputeType = types[idx-1]
		fmt.Printf("Compute type set to: %s\n", app.config.ComputeType)
		return true
	}

	fmt.Println("Invalid selection.")
	return false
}

func (app *App) configureUILanguage(reader *bufio.Reader) bool {
	languages := []string{"en", "ja"}
	languageNames := []string{"English", "日本語"}

	fmt.Println("\nAvailable UI languages:")
	for i, lang := range languages {
		fmt.Printf("%d. %s (%s)", i+1, lang, languageNames[i])
		if lang == app.config.UILanguage {
			fmt.Print(" (current)")
		}
		fmt.Println()
	}
	fmt.Print("Select UI language (1-2) or press Enter to keep current: ")

	input, _ := reader.ReadString('\n')
	choice := strings.TrimSpace(input)

	if choice == "" {
		return false
	}

	if idx, err := strconv.Atoi(choice); err == nil && idx >= 1 && idx <= len(languages) {
		app.config.UILanguage = languages[idx-1]
		fmt.Printf("UI language set to: %s\n", app.config.UILanguage)
		return true
	}

	fmt.Println("Invalid selection.")
	return false
}

func (app *App) configureUseColors(reader *bufio.Reader) bool {
	fmt.Printf("Current use colors: %t\n", app.config.UseColors)
	fmt.Print("Enable colors? (y/n) or press Enter to keep current: ")

	input, _ := reader.ReadString('\n')
	choice := strings.ToLower(strings.TrimSpace(input))

	if choice == "" {
		return false
	}

	if choice == "y" || choice == "yes" {
		app.config.UseColors = true
		fmt.Println("Colors enabled")
		return true
	} else if choice == "n" || choice == "no" {
		app.config.UseColors = false
		fmt.Println("Colors disabled")
		return true
	}

	fmt.Println("Invalid input. Please enter y or n.")
	return false
}

func (app *App) configureUIMode(reader *bufio.Reader) bool {
	modes := []string{"simple", "enhanced"}

	fmt.Println("\nAvailable UI modes:")
	for i, mode := range modes {
		fmt.Printf("%d. %s", i+1, mode)
		if mode == app.config.UIMode {
			fmt.Print(" (current)")
		}
		fmt.Println()
	}
	fmt.Print("Select UI mode (1-2) or press Enter to keep current: ")

	input, _ := reader.ReadString('\n')
	choice := strings.TrimSpace(input)

	if choice == "" {
		return false
	}

	if idx, err := strconv.Atoi(choice); err == nil && idx >= 1 && idx <= len(modes) {
		app.config.UIMode = modes[idx-1]
		fmt.Printf("UI mode set to: %s\n", app.config.UIMode)
		return true
	}

	fmt.Println("Invalid selection.")
	return false
}

func (app *App) configureOutputFormat(reader *bufio.Reader) bool {
	formats := []string{"txt", "vtt", "srt", "tsv", "json"}

	fmt.Println("\nAvailable output formats:")
	for i, format := range formats {
		fmt.Printf("%d. %s", i+1, format)
		if format == app.config.OutputFormat {
			fmt.Print(" (current)")
		}
		fmt.Println()
	}
	fmt.Print("Select output format (1-5) or press Enter to keep current: ")

	input, _ := reader.ReadString('\n')
	choice := strings.TrimSpace(input)

	if choice == "" {
		return false
	}

	if idx, err := strconv.Atoi(choice); err == nil && idx >= 1 && idx <= len(formats) {
		app.config.OutputFormat = formats[idx-1]
		fmt.Printf("Output format set to: %s\n", app.config.OutputFormat)
		return true
	}

	fmt.Println("Invalid selection.")
	return false
}

func (app *App) configureInputDir(reader *bufio.Reader) bool {
	fmt.Printf("Current input directory: %s\n", app.config.InputDir)
	fmt.Print("Press Enter to select folder with dialog, or type path manually: ")

	input, _ := reader.ReadString('\n')
	newDir := strings.TrimSpace(input)

	if newDir == "" {
		// Use folder selection dialog
		selectedDir, err := app.selectFolder("Select Input Directory")
		if err != nil {
			fmt.Printf("Folder selection failed: %v\n", err)
			return false
		}
		newDir = selectedDir
	}

	app.config.InputDir = newDir
	fmt.Printf("Input directory set to: %s\n", app.config.InputDir)
	return true
}

func (app *App) configureOutputDir(reader *bufio.Reader) bool {
	fmt.Printf("Current output directory: %s\n", app.config.OutputDir)
	fmt.Print("Press Enter to select folder with dialog, or type path manually: ")

	input, _ := reader.ReadString('\n')
	newDir := strings.TrimSpace(input)

	if newDir == "" {
		// Use folder selection dialog
		selectedDir, err := app.selectFolder("Select Output Directory")
		if err != nil {
			fmt.Printf("Folder selection failed: %v\n", err)
			return false
		}
		newDir = selectedDir
	}

	app.config.OutputDir = newDir
	fmt.Printf("Output directory set to: %s\n", app.config.OutputDir)
	return true
}

func (app *App) configureArchiveDir(reader *bufio.Reader) bool {
	fmt.Printf("Current archive directory: %s\n", app.config.ArchiveDir)
	fmt.Print("Press Enter to select folder with dialog, or type path manually: ")

	input, _ := reader.ReadString('\n')
	newDir := strings.TrimSpace(input)

	if newDir == "" {
		// Use folder selection dialog
		selectedDir, err := app.selectFolder("Select Archive Directory")
		if err != nil {
			fmt.Printf("Folder selection failed: %v\n", err)
			return false
		}
		newDir = selectedDir
	}

	app.config.ArchiveDir = newDir
	fmt.Printf("Archive directory set to: %s\n", app.config.ArchiveDir)
	return true
}

func (app *App) resetToDefaults(reader *bufio.Reader) bool {
	fmt.Print("Are you sure you want to reset all settings to defaults? (y/N): ")

	input, _ := reader.ReadString('\n')
	choice := strings.ToLower(strings.TrimSpace(input))

	if choice == "y" || choice == "yes" {
		defaultConfig := getDefaultConfig()
		*app.config = *defaultConfig
		fmt.Println("Configuration reset to defaults.")
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
		fmt.Printf("Failed to create config file: %v\n", err)
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(app.config); err != nil {
		fmt.Printf("Failed to save config: %v\n", err)
		return err
	}

	return nil
}
