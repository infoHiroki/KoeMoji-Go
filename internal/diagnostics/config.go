package diagnostics

import (
	"fmt"
	"os"

	"github.com/infoHiroki/KoeMoji-Go/internal/config"
)

var (
	configOK       = true
	configWarnings []string
	configErrors   []string
)

func checkConfiguration() {
	// Load configuration
	configPath := config.GetConfigFilePath()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configWarnings = append(configWarnings, "config.json not found (will use defaults)")
		fmt.Println("⚠ config.json not found (defaults will be used)")
		return
	}

	fmt.Println("✓ config.json found")

	// Load configuration
	cfg, err := config.LoadConfig(configPath, nil)
	if err != nil {
		configOK = false
		configErrors = append(configErrors, fmt.Sprintf("Failed to load config: %v", err))
		fmt.Printf("✗ Failed to load config.json: %v\n", err)
		return
	}

	fmt.Println("✓ JSON format valid")

	// Check if configured device exists
	if cfg.RecordingDeviceName != "" && cfg.RecordingDeviceName != "デフォルトデバイス" {
		deviceExists := false
		for _, device := range audioDevices {
			if device.Name == cfg.RecordingDeviceName {
				deviceExists = true
				break
			}
		}

		if deviceExists {
			fmt.Printf("✓ Recording device \"%s\" exists\n", cfg.RecordingDeviceName)
		} else {
			configWarnings = append(configWarnings, fmt.Sprintf("Device \"%s\" not found", cfg.RecordingDeviceName))
			fmt.Printf("⚠ Recording device \"%s\" not found\n", cfg.RecordingDeviceName)
		}
	}

	// Check dual recording configuration
	if cfg.DualRecordingEnabled {
		if systemInfo.OS == "windows" || systemInfo.OS == "darwin" {
			fmt.Println("✓ Dual recording: enabled")
		} else {
			configWarnings = append(configWarnings, "Dual recording not supported on this OS")
			fmt.Println("⚠ Dual recording enabled but not supported on this OS")
		}
	}
}
