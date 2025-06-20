package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDefaultConfig(t *testing.T) {
	config := GetDefaultConfig()

	assert.Equal(t, "large-v3", config.WhisperModel)
	assert.Equal(t, "ja", config.Language)
	assert.Equal(t, "ja", config.UILanguage)
	assert.Equal(t, 1, config.ScanIntervalMinutes)
	assert.Equal(t, 95, config.MaxCpuPercent)
	assert.Equal(t, "int8", config.ComputeType)
	assert.True(t, config.UseColors)
	assert.Equal(t, "txt", config.OutputFormat)
	assert.Equal(t, "./input", config.InputDir)
	assert.Equal(t, "./output", config.OutputDir)
	assert.Equal(t, "./archive", config.ArchiveDir)
	assert.False(t, config.LLMSummaryEnabled)
	assert.Equal(t, "openai", config.LLMAPIProvider)
	assert.Equal(t, "gpt-4o", config.LLMModel)
	assert.Equal(t, 4096, config.LLMMaxTokens)
	assert.Equal(t, "auto", config.SummaryLanguage)
}

func TestLoadConfig_ValidFile(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.json")

	testConfig := map[string]interface{}{
		"whisper_model":         "base",
		"language":              "en",
		"ui_language":           "en",
		"scan_interval_minutes": 5,
		"max_cpu_percent":       80,
		"compute_type":          "float16",
		"use_colors":            false,
		"output_format":         "srt",
		"input_dir":             "./test_input",
		"output_dir":            "./test_output",
		"archive_dir":           "./test_archive",
		"llm_summary_enabled":   true,
		"llm_api_provider":      "openai",
		"llm_api_key":           "test-key",
		"llm_model":             "gpt-3.5-turbo",
		"llm_max_tokens":        500,
		"summary_language":      "en",
	}

	data, err := json.Marshal(testConfig)
	require.NoError(t, err)

	err = os.WriteFile(configFile, data, 0644)
	require.NoError(t, err)

	logger := log.New(os.Stdout, "", log.LstdFlags)
	config := LoadConfig(configFile, logger)

	assert.Equal(t, "base", config.WhisperModel)
	assert.Equal(t, "en", config.Language)
	assert.Equal(t, "en", config.UILanguage)
	assert.Equal(t, 5, config.ScanIntervalMinutes)
	assert.Equal(t, 80, config.MaxCpuPercent)
	assert.Equal(t, "float16", config.ComputeType)
	assert.False(t, config.UseColors)
	assert.Equal(t, "srt", config.OutputFormat)
	assert.Equal(t, "./test_input", config.InputDir)
	assert.Equal(t, "./test_output", config.OutputDir)
	assert.Equal(t, "./test_archive", config.ArchiveDir)
	assert.True(t, config.LLMSummaryEnabled)
	assert.Equal(t, "openai", config.LLMAPIProvider)
	assert.Equal(t, "test-key", config.LLMAPIKey)
	assert.Equal(t, "gpt-3.5-turbo", config.LLMModel)
	assert.Equal(t, 500, config.LLMMaxTokens)
	assert.Equal(t, "en", config.SummaryLanguage)
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "nonexistent.json")

	logger := log.New(os.Stdout, "", log.LstdFlags)
	config := LoadConfig(configFile, logger)

	defaultConfig := GetDefaultConfig()
	assert.Equal(t, defaultConfig.WhisperModel, config.WhisperModel)
	assert.Equal(t, defaultConfig.Language, config.Language)
	assert.Equal(t, defaultConfig.UILanguage, config.UILanguage)
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid.json")

	err := os.WriteFile(configFile, []byte("{invalid json"), 0644)
	require.NoError(t, err)

	// Invalid JSON will cause the program to exit, so we can't test this directly
	// This test demonstrates the expected behavior documented in LoadConfig
	t.Skip("LoadConfig calls os.Exit(1) on invalid JSON - cannot test directly")
}

func TestSaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.json")

	config := GetDefaultConfig()
	config.WhisperModel = "medium"
	config.Language = "en"

	err := SaveConfig(config, configFile)
	require.NoError(t, err)

	assert.FileExists(t, configFile)

	logger := log.New(os.Stdout, "", log.LstdFlags)
	loadedConfig := LoadConfig(configFile, logger)

	assert.Equal(t, "medium", loadedConfig.WhisperModel)
	assert.Equal(t, "en", loadedConfig.Language)
}

// Note: ValidateConfig function is not exported in the actual config package
// These tests demonstrate expected validation behavior that could be implemented

// Note: GetMessages function is not exported in the actual config package
// These tests demonstrate expected multilingual message behavior

// Note: CreateDirectoriesIfNotExist function is not exported in the actual config package
// This test demonstrates expected directory creation behavior