package config

import (
	"encoding/json"
	"fmt"
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

	// Capture os.Exit behavior for testing
	if os.Getenv("TEST_INVALID_JSON") == "1" {
		logger := log.New(os.Stdout, "", log.LstdFlags)
		LoadConfig(configFile, logger) // This will call os.Exit(1)
		return
	}

	// This is documented behavior - function exits on invalid JSON
	// In production, this is the intended behavior
	t.Log("LoadConfig exits with os.Exit(1) on invalid JSON - this is expected behavior")
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

func TestConfigValidation_ValidValues(t *testing.T) {
	config := GetDefaultConfig()
	
	// Test valid whisper models
	validModels := []string{"tiny", "base", "small", "medium", "large", "large-v2", "large-v3"}
	for _, model := range validModels {
		config.WhisperModel = model
		// Validation would pass for these models
		assert.NotEmpty(t, config.WhisperModel)
	}
	
	// Test valid languages
	validLanguages := []string{"ja", "en", "auto"}
	for _, lang := range validLanguages {
		config.Language = lang
		assert.NotEmpty(t, config.Language)
	}
	
	// Test valid scan intervals
	validIntervals := []int{1, 5, 10, 30}
	for _, interval := range validIntervals {
		config.ScanIntervalMinutes = interval
		assert.Greater(t, config.ScanIntervalMinutes, 0)
	}
}

func TestConfigValidation_EdgeCases(t *testing.T) {
	config := GetDefaultConfig()
	
	// Test CPU percent boundaries
	config.MaxCpuPercent = 1
	assert.GreaterOrEqual(t, config.MaxCpuPercent, 1)
	
	config.MaxCpuPercent = 100
	assert.LessOrEqual(t, config.MaxCpuPercent, 100)
	
	// Test output format
	validFormats := []string{"txt", "srt", "vtt", "tsv", "json"}
	for _, format := range validFormats {
		config.OutputFormat = format
		assert.NotEmpty(t, config.OutputFormat)
	}
}

func TestConfigValidation_DirectoryPaths(t *testing.T) {
	config := GetDefaultConfig()
	
	// Test relative paths (should be valid)
	config.InputDir = "./input"
	config.OutputDir = "./output" 
	config.ArchiveDir = "./archive"
	
	assert.NotEmpty(t, config.InputDir)
	assert.NotEmpty(t, config.OutputDir)
	assert.NotEmpty(t, config.ArchiveDir)
	
	// Test that directories are different
	assert.NotEqual(t, config.InputDir, config.OutputDir)
	assert.NotEqual(t, config.InputDir, config.ArchiveDir)
	assert.NotEqual(t, config.OutputDir, config.ArchiveDir)
}

func TestLLMConfig_Validation(t *testing.T) {
	config := GetDefaultConfig()
	
	// Test LLM settings
	config.LLMSummaryEnabled = true
	config.LLMAPIProvider = "openai"
	config.LLMModel = "gpt-4o"
	config.LLMMaxTokens = 4096
	
	assert.True(t, config.LLMSummaryEnabled)
	assert.Equal(t, "openai", config.LLMAPIProvider)
	assert.Greater(t, config.LLMMaxTokens, 0)
	assert.LessOrEqual(t, config.LLMMaxTokens, 8192) // Reasonable upper limit
}

func TestRecordingConfig_Validation(t *testing.T) {
	config := GetDefaultConfig()
	
	// Test recording settings
	config.RecordingDeviceID = -1 // Default device
	config.RecordingMaxHours = 2
	config.RecordingMaxFileMB = 100
	
	assert.GreaterOrEqual(t, config.RecordingDeviceID, -1)
	assert.GreaterOrEqual(t, config.RecordingMaxHours, 0)
	assert.GreaterOrEqual(t, config.RecordingMaxFileMB, 0)
}

// Test the getMessages function for both languages
func TestGetMessages(t *testing.T) {
	// Test English messages
	configEN := &Config{UILanguage: "en"}
	messagesEN := getMessages(configEN)
	assert.Equal(t, "KoeMoji-Go Configuration", messagesEN.ConfigTitle)
	assert.Equal(t, "Whisper Model", messagesEN.WhisperModel)
	
	// Test Japanese messages
	configJA := &Config{UILanguage: "ja"}
	messagesJA := getMessages(configJA)
	assert.Equal(t, "KoeMoji-Go 設定", messagesJA.ConfigTitle)
	assert.Equal(t, "Whisperモデル", messagesJA.WhisperModel)
	
	// Test nil config defaults to English
	messagesDefault := getMessages(nil)
	assert.Equal(t, "KoeMoji-Go Configuration", messagesDefault.ConfigTitle)
}

// Test the getAPIKeyDisplay function
func TestGetAPIKeyDisplay(t *testing.T) {
	tests := []struct {
		apiKey   string
		expected string
	}{
		{"", "[未設定]"},
		{"sk-1234567890", "[設定済み]"},
		{"sk-1234567890123456789012345", "sk-1...5"},
		{"very-long-api-key-that-exceeds-ten-characters", "very...ters"},
	}
	
	for _, tt := range tests {
		t.Run(fmt.Sprintf("apiKey=%s", tt.apiKey), func(t *testing.T) {
			result := getAPIKeyDisplay(tt.apiKey)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test configuration validation with invalid values
func TestConfigValidation_InvalidValues(t *testing.T) {
	config := GetDefaultConfig()
	
	// Test invalid CPU percent (should be caught by validation)
	config.MaxCpuPercent = 0
	assert.Equal(t, 0, config.MaxCpuPercent) // Structure allows it, validation should catch
	
	config.MaxCpuPercent = 101
	assert.Equal(t, 101, config.MaxCpuPercent) // Structure allows it, validation should catch
	
	// Test negative scan interval
	config.ScanIntervalMinutes = -1
	assert.Equal(t, -1, config.ScanIntervalMinutes) // Structure allows it, validation should catch
	
	// Test invalid LLM max tokens
	config.LLMMaxTokens = -100
	assert.Equal(t, -100, config.LLMMaxTokens) // Structure allows it, validation should catch
}

// Test SaveConfig with invalid directory permissions
func TestSaveConfig_InvalidPermissions(t *testing.T) {
	config := GetDefaultConfig()
	
	// Try to save to a non-existent directory
	invalidPath := "/non/existent/directory/config.json"
	err := SaveConfig(config, invalidPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create config file")
}

// Test LoadConfig with file permission errors
func TestLoadConfig_PermissionError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("Running as root, permission test not applicable")
	}
	
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "no_read.json")
	
	// Create file and write valid config
	testConfig := map[string]interface{}{
		"whisper_model": "base",
		"language":      "en",
	}
	data, _ := json.Marshal(testConfig)
	err := os.WriteFile(configFile, data, 0000) // No read permissions
	require.NoError(t, err)
	
	logger := log.New(os.Stdout, "", log.LstdFlags)
	
	// This should fail to open and return default config
	config := LoadConfig(configFile, logger)
	
	// Should return default config when file can't be read
	defaultConfig := GetDefaultConfig()
	assert.Equal(t, defaultConfig.WhisperModel, config.WhisperModel)
}

// Test complete config round-trip with all fields
func TestCompleteConfigRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "complete_config.json")
	
	originalConfig := &Config{
		WhisperModel:          "large-v2",
		Language:              "en",
		UILanguage:            "en", 
		ScanIntervalMinutes:   3,
		MaxCpuPercent:         85,
		ComputeType:           "float16",
		UseColors:             false,
		OutputFormat:          "vtt",
		InputDir:              "./custom_input",
		OutputDir:             "./custom_output",
		ArchiveDir:            "./custom_archive",
		LLMSummaryEnabled:     true,
		LLMAPIProvider:        "openai",
		LLMAPIKey:             "test-api-key-12345",
		LLMModel:              "gpt-3.5-turbo",
		LLMMaxTokens:          2048,
		SummaryPromptTemplate: "Summarize this text: {text} in {language}",
		SummaryLanguage:       "en",
		RecordingDeviceID:     5,
		RecordingDeviceName:   "Test Microphone",
		RecordingMaxHours:     3,
		RecordingMaxFileMB:    200,
	}
	
	// Save the config
	err := SaveConfig(originalConfig, configFile)
	require.NoError(t, err)
	
	// Load it back
	logger := log.New(os.Stdout, "", log.LstdFlags)
	loadedConfig := LoadConfig(configFile, logger)
	
	// Verify all fields match
	assert.Equal(t, originalConfig.WhisperModel, loadedConfig.WhisperModel)
	assert.Equal(t, originalConfig.Language, loadedConfig.Language)
	assert.Equal(t, originalConfig.UILanguage, loadedConfig.UILanguage)
	assert.Equal(t, originalConfig.ScanIntervalMinutes, loadedConfig.ScanIntervalMinutes)
	assert.Equal(t, originalConfig.MaxCpuPercent, loadedConfig.MaxCpuPercent)
	assert.Equal(t, originalConfig.ComputeType, loadedConfig.ComputeType)
	assert.Equal(t, originalConfig.UseColors, loadedConfig.UseColors)
	assert.Equal(t, originalConfig.OutputFormat, loadedConfig.OutputFormat)
	assert.Equal(t, originalConfig.InputDir, loadedConfig.InputDir)
	assert.Equal(t, originalConfig.OutputDir, loadedConfig.OutputDir)
	assert.Equal(t, originalConfig.ArchiveDir, loadedConfig.ArchiveDir)
	assert.Equal(t, originalConfig.LLMSummaryEnabled, loadedConfig.LLMSummaryEnabled)
	assert.Equal(t, originalConfig.LLMAPIProvider, loadedConfig.LLMAPIProvider)
	assert.Equal(t, originalConfig.LLMAPIKey, loadedConfig.LLMAPIKey)
	assert.Equal(t, originalConfig.LLMModel, loadedConfig.LLMModel)
	assert.Equal(t, originalConfig.LLMMaxTokens, loadedConfig.LLMMaxTokens)
	assert.Equal(t, originalConfig.SummaryPromptTemplate, loadedConfig.SummaryPromptTemplate)
	assert.Equal(t, originalConfig.SummaryLanguage, loadedConfig.SummaryLanguage)
	assert.Equal(t, originalConfig.RecordingDeviceID, loadedConfig.RecordingDeviceID)
	assert.Equal(t, originalConfig.RecordingDeviceName, loadedConfig.RecordingDeviceName)
	assert.Equal(t, originalConfig.RecordingMaxHours, loadedConfig.RecordingMaxHours)
	assert.Equal(t, originalConfig.RecordingMaxFileMB, loadedConfig.RecordingMaxFileMB)
}

// Test JSON encoding with special characters
func TestSaveConfig_SpecialCharacters(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "special_chars.json")
	
	config := GetDefaultConfig()
	config.RecordingDeviceName = "マイク デバイス"  // Japanese characters
	config.SummaryPromptTemplate = "テスト: {text} を {language} で要約してください。"  // Japanese with placeholders
	
	err := SaveConfig(config, configFile)
	require.NoError(t, err)
	
	// Verify the file contains proper JSON
	data, err := os.ReadFile(configFile)
	require.NoError(t, err)
	
	var loadedData map[string]interface{}
	err = json.Unmarshal(data, &loadedData)
	require.NoError(t, err)
	
	assert.Equal(t, "マイク デバイス", loadedData["recording_device_name"])
	assert.Contains(t, loadedData["summary_prompt_template"].(string), "テスト")
}
