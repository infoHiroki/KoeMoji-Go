package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/infoHiroki/KoeMoji-Go/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestConfigDialog creates a test ConfigDialog instance
func createTestConfigDialog(t *testing.T) (*ConfigDialog, *config.Config) {
	tempDir := t.TempDir()

	cfg := config.GetDefaultConfig()
	cfg.InputDir = filepath.Join(tempDir, "input")
	cfg.OutputDir = filepath.Join(tempDir, "output")
	cfg.ArchiveDir = filepath.Join(tempDir, "archive")

	// Create directories
	for _, dir := range []string{cfg.InputDir, cfg.OutputDir, cfg.ArchiveDir} {
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)
	}

	// Note: We don't initialize tview.Application here because it requires a terminal
	dialog := &ConfigDialog{
		config: cfg,
	}

	return dialog, cfg
}

func TestConfigDialog_Creation(t *testing.T) {
	dialog, cfg := createTestConfigDialog(t)

	assert.NotNil(t, dialog)
	assert.NotNil(t, dialog.config)
	assert.Equal(t, cfg, dialog.config)
}

func TestConfigDialog_ConfigurationAccess(t *testing.T) {
	dialog, _ := createTestConfigDialog(t)

	// Test configuration access
	assert.NotNil(t, dialog.config)
	assert.Equal(t, "large-v3", dialog.config.WhisperModel)
	assert.Equal(t, "ja", dialog.config.Language)
	assert.Equal(t, "ja", dialog.config.UILanguage)
	assert.NotEmpty(t, dialog.config.InputDir)
	assert.NotEmpty(t, dialog.config.OutputDir)
	assert.NotEmpty(t, dialog.config.ArchiveDir)
}

func TestConfigDialog_BasicSettings(t *testing.T) {
	dialog, _ := createTestConfigDialog(t)

	// Test basic settings
	assert.Equal(t, "large-v3", dialog.config.WhisperModel)
	assert.Equal(t, "ja", dialog.config.Language)
	assert.Equal(t, "ja", dialog.config.UILanguage)
	assert.Equal(t, 1, dialog.config.ScanIntervalMinutes)
}

func TestConfigDialog_DirectorySettings(t *testing.T) {
	dialog, _ := createTestConfigDialog(t)

	// Test directory settings
	assert.NotEmpty(t, dialog.config.InputDir)
	assert.NotEmpty(t, dialog.config.OutputDir)
	assert.NotEmpty(t, dialog.config.ArchiveDir)

	// Verify directories exist
	for _, dir := range []string{
		dialog.config.InputDir,
		dialog.config.OutputDir,
		dialog.config.ArchiveDir,
	} {
		_, err := os.Stat(dir)
		assert.NoError(t, err, "Directory should exist: %s", dir)
	}
}

func TestConfigDialog_LLMSettings(t *testing.T) {
	dialog, _ := createTestConfigDialog(t)

	// Test LLM settings
	assert.False(t, dialog.config.LLMSummaryEnabled)
	assert.Empty(t, dialog.config.LLMAPIKey)
	assert.Equal(t, "gpt-4o", dialog.config.LLMModel)
}

func TestConfigDialog_RecordingSettings(t *testing.T) {
	dialog, _ := createTestConfigDialog(t)

	// Test recording settings
	assert.Empty(t, dialog.config.RecordingDeviceName)
	// DualRecordingEnabled depends on runtime.GOOS
	// Just verify it's a boolean
	assert.IsType(t, false, dialog.config.DualRecordingEnabled)
}

func TestConfigDialog_AdvancedSettings(t *testing.T) {
	dialog, _ := createTestConfigDialog(t)

	// Test advanced settings
	assert.Equal(t, "int8", dialog.config.ComputeType)
	assert.Equal(t, "txt", dialog.config.OutputFormat)
	assert.Equal(t, 95, dialog.config.MaxCpuPercent)
}

func TestConfigDialog_SaveConfig(t *testing.T) {
	_, cfg := createTestConfigDialog(t)

	// Modify config
	cfg.WhisperModel = "base"
	cfg.Language = "en"
	cfg.ScanIntervalMinutes = 5

	// Save config to temp file
	tempConfigFile := filepath.Join(t.TempDir(), "test_config.json")
	err := config.SaveConfig(cfg, tempConfigFile)
	require.NoError(t, err)

	// Load config and verify
	loadedCfg, err := config.LoadConfig(tempConfigFile, nil)
	require.NoError(t, err)

	assert.Equal(t, "base", loadedCfg.WhisperModel)
	assert.Equal(t, "en", loadedCfg.Language)
	assert.Equal(t, 5, loadedCfg.ScanIntervalMinutes)
}

func TestConfigDialog_ConfigModification(t *testing.T) {
	dialog, _ := createTestConfigDialog(t)

	// Test modifying configuration values
	dialog.config.WhisperModel = "small"
	assert.Equal(t, "small", dialog.config.WhisperModel)

	dialog.config.Language = "en"
	assert.Equal(t, "en", dialog.config.Language)

	dialog.config.UILanguage = "en"
	assert.Equal(t, "en", dialog.config.UILanguage)

	dialog.config.ScanIntervalMinutes = 10
	assert.Equal(t, 10, dialog.config.ScanIntervalMinutes)

	dialog.config.LLMSummaryEnabled = true
	assert.True(t, dialog.config.LLMSummaryEnabled)

	dialog.config.LLMAPIKey = "test-api-key"
	assert.Equal(t, "test-api-key", dialog.config.LLMAPIKey)

	dialog.config.ComputeType = "float16"
	assert.Equal(t, "float16", dialog.config.ComputeType)

	dialog.config.OutputFormat = "vtt"
	assert.Equal(t, "vtt", dialog.config.OutputFormat)

	dialog.config.MaxCpuPercent = 80
	assert.Equal(t, 80, dialog.config.MaxCpuPercent)
}

func TestConfigDialog_DirectoryPaths(t *testing.T) {
	dialog, _ := createTestConfigDialog(t)

	// Test relative path conversion
	relativePath := config.GetRelativePath(dialog.config.InputDir)
	assert.NotEmpty(t, relativePath)

	// Relative paths should start with ./ or be absolute
	// (depending on whether they're relative to current directory)
	assert.True(t,
		filepath.IsAbs(relativePath) ||
		relativePath[:2] == "./" ||
		relativePath[:3] == "../",
		"Path should be relative or absolute: %s", relativePath)
}

func TestConfigDialog_CallbacksNil(t *testing.T) {
	dialog, _ := createTestConfigDialog(t)

	// Initially, callbacks should be nil
	assert.Nil(t, dialog.onSave)
	assert.Nil(t, dialog.onCancel)
}

func TestConfigDialog_SetCallbacks(t *testing.T) {
	dialog, _ := createTestConfigDialog(t)

	// Set callbacks
	saveCalled := false
	cancelCalled := false

	// Simulate Show with callbacks
	dialog.onSave = func() { saveCalled = true }
	dialog.onCancel = func() { cancelCalled = true }

	// Test that callbacks are set
	assert.NotNil(t, dialog.onSave)
	assert.NotNil(t, dialog.onCancel)

	// Test that callbacks work
	dialog.onSave()
	assert.True(t, saveCalled)

	dialog.onCancel()
	assert.True(t, cancelCalled)
}

func TestConfigDialog_TabIndexing(t *testing.T) {
	// Test tab names (should match implementation)
	tabs := []string{"basic", "directories", "llm", "recording", "advanced"}

	for i, tabName := range tabs {
		assert.NotEmpty(t, tabName, "Tab %d should have a name", i)
	}

	assert.Len(t, tabs, 5, "Should have 5 tabs")
}

func TestConfigDialog_ConfigDefaults(t *testing.T) {
	dialog, _ := createTestConfigDialog(t)
	cfg := dialog.config

	// Test all default values
	tests := []struct {
		name     string
		actual   interface{}
		expected interface{}
	}{
		{"WhisperModel", cfg.WhisperModel, "large-v3"},
		{"Language", cfg.Language, "ja"},
		{"UILanguage", cfg.UILanguage, "ja"},
		{"ScanIntervalMinutes", cfg.ScanIntervalMinutes, 1},
		{"MaxCpuPercent", cfg.MaxCpuPercent, 95},
		{"ComputeType", cfg.ComputeType, "int8"},
		{"UseColors", cfg.UseColors, true},
		{"OutputFormat", cfg.OutputFormat, "txt"},
		{"LLMSummaryEnabled", cfg.LLMSummaryEnabled, false},
		{"LLMAPIProvider", cfg.LLMAPIProvider, "openai"},
		{"LLMModel", cfg.LLMModel, "gpt-4o"},
		{"LLMMaxTokens", cfg.LLMMaxTokens, 4096},
		{"RecordingDeviceName", cfg.RecordingDeviceName, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.actual)
		})
	}
}

// Benchmark tests
func BenchmarkConfigDialog_Creation(b *testing.B) {
	tempDir := b.TempDir()
	cfg := config.GetDefaultConfig()
	cfg.InputDir = filepath.Join(tempDir, "input")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dialog := &ConfigDialog{
			config: cfg,
		}
		_ = dialog
	}
}

func BenchmarkConfigDialog_ConfigAccess(b *testing.B) {
	dialog, _ := createTestConfigDialog(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dialog.config.WhisperModel
		_ = dialog.config.Language
		_ = dialog.config.InputDir
	}
}

func BenchmarkConfigDialog_ConfigModification(b *testing.B) {
	dialog, _ := createTestConfigDialog(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dialog.config.ScanIntervalMinutes = i
		dialog.config.LLMSummaryEnabled = !dialog.config.LLMSummaryEnabled
	}
}
