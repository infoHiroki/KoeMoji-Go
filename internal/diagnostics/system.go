package diagnostics

import (
	"fmt"
	"os"
	"runtime"
)

// SystemInfo holds system information
type SystemInfo struct {
	OS           string
	Arch         string
	Version      string
	ExecutablePath string
}

var (
	systemInfo SystemInfo
	systemOK   = true
)

func checkSystem() {
	// Get executable path
	exePath, err := os.Executable()
	if err != nil {
		exePath = "unknown"
	}

	systemInfo = SystemInfo{
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		Version:      getVersion(),
		ExecutablePath: exePath,
	}

	fmt.Printf("✓ OS: %s (%s)\n", getOSName(), systemInfo.Arch)
	fmt.Printf("✓ Version: %s\n", systemInfo.Version)
	fmt.Printf("✓ Path: %s\n", systemInfo.ExecutablePath)
}

func getOSName() string {
	switch runtime.GOOS {
	case "windows":
		return "Windows"
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	default:
		return runtime.GOOS
	}
}

var appVersion string

func SetVersion(version string) {
	appVersion = version
}

func getVersion() string {
	if appVersion == "" {
		return "unknown"
	}
	return appVersion
}
