//go:build darwin

package diagnostics

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	dualRecordingOK       = true
	dualRecordingWarnings []string
	dualRecordingErrors   []string
)

func checkDualRecording() {
	// Check macOS version
	version, err := getMacOSVersion()
	if err != nil {
		dualRecordingWarnings = append(dualRecordingWarnings, "Could not determine macOS version")
		fmt.Printf("⚠ Could not determine macOS version: %v\n", err)
	} else {
		major, _ := strconv.Atoi(strings.Split(version, ".")[0])
		if major >= 13 {
			fmt.Printf("✓ macOS version: %s (ScreenCaptureKit supported)\n", version)
		} else {
			dualRecordingWarnings = append(dualRecordingWarnings, "macOS 13+ required for dual recording")
			fmt.Printf("⚠ macOS version: %s (macOS 13+ required for dual recording)\n", version)
		}
	}

	// Check for audio-capture binary
	binaryPath := findAudioCaptureBinary()
	if binaryPath != "" {
		fmt.Printf("✓ audio-capture binary found: %s\n", binaryPath)
	} else {
		dualRecordingOK = false
		dualRecordingErrors = append(dualRecordingErrors, "audio-capture binary not found")
		fmt.Println("✗ audio-capture binary not found")
		fmt.Println("  (Required for system audio recording)")
	}
}

func getMacOSVersion() (string, error) {
	cmd := exec.Command("sw_vers", "-productVersion")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func findAudioCaptureBinary() string {
	// Get executable directory
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	exeDir := filepath.Dir(exePath)

	// Search paths
	searchPaths := []string{
		filepath.Join(exeDir, "audio-capture"),
		"/usr/local/bin/audio-capture",
		"audio-capture", // Check PATH
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Check PATH
	if path, err := exec.LookPath("audio-capture"); err == nil {
		return path
	}

	return ""
}
