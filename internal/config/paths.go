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

// ResolvePath resolves a path relative to the executable directory
func ResolvePath(path string) string {
	// If already absolute, return as-is
	if filepath.IsAbs(path) {
		return path
	}
	
	// Get executable directory
	exeDir, err := GetExecutablePath()
	if err != nil {
		// Fallback to current directory
		return path
	}
	
	// Join with executable directory
	return filepath.Join(exeDir, path)
}

// ResolveConfigPaths updates all relative paths in the config to be relative to the executable
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
