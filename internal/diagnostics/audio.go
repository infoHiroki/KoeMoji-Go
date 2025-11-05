package diagnostics

import (
	"fmt"
	"strings"

	"github.com/infoHiroki/KoeMoji-Go/internal/recorder"
)

var (
	audioDevices    []recorder.DeviceInfo
	audioOK         = true
	audioErrors     []string
	defaultDeviceID int = -1
)

func checkAudioDevices() {
	// List all audio devices
	devices, err := recorder.ListDevices()
	if err != nil {
		audioOK = false
		audioErrors = append(audioErrors, fmt.Sprintf("Failed to list devices: %v", err))
		fmt.Printf("✗ PortAudio initialization failed: %v\n", err)
		return
	}

	audioDevices = devices
	fmt.Println("✓ PortAudio initialized successfully")

	if len(devices) == 0 {
		audioOK = false
		audioErrors = append(audioErrors, "No input devices found")
		fmt.Println("✗ No input devices found")
		return
	}

	fmt.Printf("✓ Found %d input device(s):\n", len(devices))
	fmt.Println()

	// Display each device
	for _, device := range devices {
		displayDevice(device)
	}
}

func displayDevice(device recorder.DeviceInfo) {
	// Device number and name
	prefix := " "
	if device.IsDefault {
		prefix = "●"
		defaultDeviceID = device.ID
	}

	fmt.Printf("  %s %s", prefix, device.Name)

	if device.IsDefault {
		fmt.Print(" [DEFAULT]")
	}
	fmt.Println()

	// Device details
	fmt.Printf("     Host API: %s\n", device.HostAPI)
	fmt.Printf("     Channels: %d\n", device.MaxChannels)

	// Virtual device detection
	if device.IsVirtual {
		fmt.Printf("     Type: Virtual (%s)\n", device.VirtualType)
	} else {
		// Fallback: check for common virtual device names
		if isVirtualDeviceName(device.Name) {
			fmt.Printf("     Type: Virtual\n")
		} else {
			fmt.Printf("     Type: Physical\n")
		}
	}

	fmt.Println()
}

// isVirtualDeviceName checks if the device name contains virtual device keywords
func isVirtualDeviceName(name string) bool {
	name = strings.ToLower(name)
	virtualKeywords := []string{
		"voicemeeter", "stereo mix", "what u hear", "rec. playback",
		"blackhole", "aggregate", "集約", "multi-output", "マルチ出力",
		"cable", "virtual",
	}

	for _, keyword := range virtualKeywords {
		if strings.Contains(name, keyword) {
			return true
		}
	}
	return false
}
