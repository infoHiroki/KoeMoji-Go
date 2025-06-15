package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	WhisperModel        string `json:"whisper_model"`
	Language            string `json:"language"`
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

func (app *App) displayConfig() {
	fmt.Println("\n--- Configuration (config.json) ---")
	fmt.Printf("Whisper model: %s\n", app.config.WhisperModel)
	fmt.Printf("Language: %s\n", app.config.Language)
	fmt.Printf("Scan interval: %d minutes\n", app.config.ScanIntervalMinutes)
	fmt.Printf("Max CPU percent: %d%%\n", app.config.MaxCpuPercent)
	fmt.Printf("Compute type: %s\n", app.config.ComputeType)
	fmt.Printf("Use colors: %t\n", app.config.UseColors)
	fmt.Printf("UI mode: %s\n", app.config.UIMode)
	fmt.Printf("Output format: %s\n", app.config.OutputFormat)
	fmt.Println("\nDirectories:")
	fmt.Println("  Input: ./input/")
	fmt.Println("  Output: ./output/")
	fmt.Println("  Archive: ./archive/")
	fmt.Println("---")
}