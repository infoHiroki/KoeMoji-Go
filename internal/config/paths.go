package config

import (
	"os"
	"path/filepath"
	"strings"
)

// GetExecutablePath returns the directory containing the executable
func GetExecutablePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exePath), nil
}

// ResolvePath resolves a path relative to the application base directory
// For .app: relative to ~/Documents/KoeMoji-Go/
// For CLI: relative to executable directory
func ResolvePath(path string) string {
	// If already absolute, return as-is
	if filepath.IsAbs(path) {
		return path
	}

	// Get base directory (depends on .app vs CLI)
	baseDir := GetAppBaseDir()

	// Join with base directory
	return filepath.Join(baseDir, path)
}

// ResolveConfigPaths updates all relative paths in the config
// For .app: resolves to ~/Documents/KoeMoji-Go/
// For CLI: resolves relative to executable directory
func ResolveConfigPaths(config *Config) {
	config.InputDir = ResolvePath(config.InputDir)
	config.OutputDir = ResolvePath(config.OutputDir)
	config.ArchiveDir = ResolvePath(config.ArchiveDir)
}

// GetRelativePath converts absolute path to relative path from executable directory
func GetRelativePath(absolutePath string) string {
	exeDir, err := GetExecutablePath()
	if err != nil {
		return absolutePath
	}

	// Try to get relative path
	relPath, err := filepath.Rel(exeDir, absolutePath)
	if err != nil {
		return absolutePath
	}

	// Add "./" prefix if not present and not going up directories
	if !strings.HasPrefix(relPath, ".") && !strings.HasPrefix(relPath, "..") {
		return "./" + relPath
	}

	return relPath
}

// IsRunningAsApp checks if the application is running as a .app bundle
func IsRunningAsApp() bool {
	exePath, err := os.Executable()
	if err != nil {
		return false
	}

	// Check if executable path contains .app/Contents/MacOS/
	return strings.Contains(exePath, ".app/Contents/MacOS/")
}

// GetAppBaseDir returns the base directory for application data
// For .app: ~/Documents/KoeMoji-Go/
// For CLI: executable directory
func GetAppBaseDir() string {
	if IsRunningAsApp() {
		// .app version: use ~/Documents/KoeMoji-Go/
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback to current directory
			return "."
		}
		return filepath.Join(homeDir, "Documents", "KoeMoji-Go")
	}

	// CLI version: use executable directory
	exeDir, err := GetExecutablePath()
	if err != nil {
		return "."
	}
	return exeDir
}

// GetLogFilePath returns the appropriate log file path
func GetLogFilePath() string {
	baseDir := GetAppBaseDir()
	return filepath.Join(baseDir, "koemoji.log")
}

// GetConfigFilePath returns the appropriate config file path
func GetConfigFilePath() string {
	baseDir := GetAppBaseDir()
	return filepath.Join(baseDir, "config.json")
}
