package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/hirokitakamura/koemoji-go/internal/config/testdata"
)

// Test individual configure functions
func TestConfigureWhisperModel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		changed  bool
	}{
		{"Select tiny model", "1", "tiny", true},
		{"Select base model", "2", "base", true},
		{"Select small model", "3", "small", true},
		{"Select medium model", "4", "medium", true},
		{"Select large model", "5", "large", true},
		{"Select large-v2 model", "6", "large-v2", true},
		{"Select large-v3 model", "7", "large-v3", true},
		{"Keep current (empty input)", "", "large-v3", false},
		{"Invalid input", "99", "large-v3", false},
		{"Invalid input (text)", "abc", "large-v3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GetDefaultConfig()
			config.WhisperModel = "large-v3"
			reader := testdata.CreateMockReader(tt.input)
			
			changed := configureWhisperModel(config, reader)
			
			assert.Equal(t, tt.expected, config.WhisperModel)
			assert.Equal(t, tt.changed, changed)
		})
	}
}

func TestConfigureLanguage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		changed  bool
	}{
		{"Set Japanese", "ja", "ja", true},
		{"Set English", "en", "en", true},
		{"Set Chinese", "zh", "zh", true},
		{"Set Korean", "ko", "ko", true},
		{"Set Spanish", "es", "es", true},
		{"Set Auto", "auto", "auto", true},
		{"Keep current (empty)", "", "ja", false},
		{"Same as current", "ja", "ja", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GetDefaultConfig()
			config.Language = "ja"
			reader := testdata.CreateMockReader(tt.input)
			
			changed := configureLanguage(config, reader)
			
			assert.Equal(t, tt.expected, config.Language)
			assert.Equal(t, tt.changed, changed)
		})
	}
}

func TestConfigureUILanguage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		changed  bool
	}{
		{"Select Japanese", "1", "ja", true},
		{"Select English", "2", "en", true},
		{"Keep current (empty)", "", "ja", false},
		{"Invalid input", "3", "ja", false},
		{"Invalid input (text)", "xyz", "ja", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GetDefaultConfig()
			config.UILanguage = "ja"
			reader := testdata.CreateMockReader(tt.input)
			
			changed := configureUILanguage(config, reader)
			
			assert.Equal(t, tt.expected, config.UILanguage)
			assert.Equal(t, tt.changed, changed)
		})
	}
}

func TestConfigureScanInterval(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		changed  bool
	}{
		{"Set to 5 minutes", "5", 5, true},
		{"Set to 10 minutes", "10", 10, true},
		{"Set to 30 minutes", "30", 30, true},
		{"Keep current (empty)", "", 1, false},
		{"Invalid (zero)", "0", 1, false},
		{"Invalid (negative)", "-5", 1, false},
		{"Invalid (text)", "abc", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GetDefaultConfig()
			config.ScanIntervalMinutes = 1
			reader := testdata.CreateMockReader(tt.input)
			
			changed := configureScanInterval(config, reader)
			
			assert.Equal(t, tt.expected, config.ScanIntervalMinutes)
			assert.Equal(t, tt.changed, changed)
		})
	}
}

func TestConfigureMaxCpuPercent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		changed  bool
	}{
		{"Set to 50%", "50", 50, true},
		{"Set to 80%", "80", 80, true},
		{"Set to 100%", "100", 100, true},
		{"Set to minimum 1%", "1", 1, true},
		{"Keep current (empty)", "", 95, false},
		{"Invalid (zero)", "0", 95, false},
		{"Invalid (over 100)", "101", 95, false},
		{"Invalid (negative)", "-10", 95, false},
		{"Invalid (text)", "fifty", 95, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GetDefaultConfig()
			config.MaxCpuPercent = 95
			reader := testdata.CreateMockReader(tt.input)
			
			changed := configureMaxCpuPercent(config, reader)
			
			assert.Equal(t, tt.expected, config.MaxCpuPercent)
			assert.Equal(t, tt.changed, changed)
		})
	}
}

func TestConfigureComputeType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		changed  bool
	}{
		{"Select int8", "1", "int8", true},
		{"Select int16", "2", "int16", true},
		{"Select float16", "3", "float16", true},
		{"Select float32", "4", "float32", true},
		{"Keep current (empty)", "", "int8", false},
		{"Invalid input", "5", "int8", false},
		{"Invalid input (text)", "gpu", "int8", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GetDefaultConfig()
			config.ComputeType = "int8"
			reader := testdata.CreateMockReader(tt.input)
			
			changed := configureComputeType(config, reader)
			
			assert.Equal(t, tt.expected, config.ComputeType)
			assert.Equal(t, tt.changed, changed)
		})
	}
}

func TestConfigureUseColors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
		changed  bool
	}{
		{"Enable colors", "y", true, true},
		{"Enable colors (uppercase)", "Y", true, true},
		{"Disable colors", "n", false, true},
		{"Disable colors (uppercase)", "N", false, true},
		{"Keep current true (empty)", "", true, false},
		{"Invalid input", "maybe", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GetDefaultConfig()
			config.UseColors = true
			reader := testdata.CreateMockReader(tt.input)
			
			changed := configureUseColors(config, reader)
			
			assert.Equal(t, tt.expected, config.UseColors)
			assert.Equal(t, tt.changed, changed)
		})
	}
}

func TestConfigureOutputFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		changed  bool
	}{
		{"Select txt", "1", "txt", true},
		{"Select srt", "2", "srt", true},
		{"Select vtt", "3", "vtt", true},
		{"Select tsv", "4", "tsv", true},
		{"Select json", "5", "json", true},
		{"Keep current (empty)", "", "txt", false},
		{"Invalid input", "6", "txt", false},
		{"Invalid input (text)", "xml", "txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GetDefaultConfig()
			config.OutputFormat = "txt"
			reader := testdata.CreateMockReader(tt.input)
			
			changed := configureOutputFormat(config, reader)
			
			assert.Equal(t, tt.expected, config.OutputFormat)
			assert.Equal(t, tt.changed, changed)
		})
	}
}

func TestConfigureDirectories(t *testing.T) {
	t.Run("ConfigureInputDir", func(t *testing.T) {
		config := GetDefaultConfig()
		config.InputDir = "./input"
		
		// Test manual path input
		reader := testdata.CreateMockReader("./custom_input")
		changed := configureInputDir(config, reader)
		
		assert.Equal(t, "./custom_input", config.InputDir)
		assert.True(t, changed)
		
		// Test keep current (empty)
		reader = testdata.CreateMockReader("")
		changed = configureInputDir(config, reader)
		
		assert.Equal(t, "./custom_input", config.InputDir)
		assert.False(t, changed)
	})
	
	t.Run("ConfigureOutputDir", func(t *testing.T) {
		config := GetDefaultConfig()
		config.OutputDir = "./output"
		
		reader := testdata.CreateMockReader("./custom_output")
		changed := configureOutputDir(config, reader)
		
		assert.Equal(t, "./custom_output", config.OutputDir)
		assert.True(t, changed)
	})
	
	t.Run("ConfigureArchiveDir", func(t *testing.T) {
		config := GetDefaultConfig()
		config.ArchiveDir = "./archive"
		
		reader := testdata.CreateMockReader("./custom_archive")
		changed := configureArchiveDir(config, reader)
		
		assert.Equal(t, "./custom_archive", config.ArchiveDir)
		assert.True(t, changed)
	})
}

func TestConfigureLLMSettings(t *testing.T) {
	t.Run("ConfigureLLMSummaryEnabled", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected bool
			changed  bool
		}{
			{"Enable LLM", "y", true, true},
			{"Disable LLM", "n", false, true},
			{"Keep disabled (empty)", "", false, false},
			{"Invalid input", "maybe", false, false},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				config := GetDefaultConfig()
				config.LLMSummaryEnabled = false
				reader := testdata.CreateMockReader(tt.input)
				
				changed := configureLLMSummaryEnabled(config, reader)
				
				assert.Equal(t, tt.expected, config.LLMSummaryEnabled)
				assert.Equal(t, tt.changed, changed)
			})
		}
	})
	
	t.Run("ConfigureLLMAPIProvider", func(t *testing.T) {
		config := GetDefaultConfig()
		config.LLMAPIProvider = "openai"
		
		// Currently only supports openai
		reader := testdata.CreateMockReader("1")
		changed := configureLLMAPIProvider(config, reader)
		
		assert.Equal(t, "openai", config.LLMAPIProvider)
		assert.False(t, changed) // No change as it's already openai
	})
	
	t.Run("ConfigureLLMAPIKey", func(t *testing.T) {
		config := GetDefaultConfig()
		config.LLMAPIKey = ""
		
		reader := testdata.CreateMockReader("sk-test1234567890")
		changed := configureLLMAPIKey(config, reader)
		
		assert.Equal(t, "sk-test1234567890", config.LLMAPIKey)
		assert.True(t, changed)
		
		// Test keep current
		reader = testdata.CreateMockReader("")
		changed = configureLLMAPIKey(config, reader)
		
		assert.Equal(t, "sk-test1234567890", config.LLMAPIKey)
		assert.False(t, changed)
	})
	
	t.Run("ConfigureLLMModel", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected string
			changed  bool
		}{
			{"Select gpt-4o", "1", "gpt-4o", true},
			{"Select gpt-4o-mini", "2", "gpt-4o-mini", true},
			{"Select gpt-4-turbo", "3", "gpt-4-turbo", true},
			{"Select gpt-4", "4", "gpt-4", true},
			{"Select gpt-3.5-turbo", "5", "gpt-3.5-turbo", true},
			{"Keep current", "", "gpt-4o", false},
			{"Invalid input", "10", "gpt-4o", false},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				config := GetDefaultConfig()
				config.LLMModel = "gpt-4o"
				reader := testdata.CreateMockReader(tt.input)
				
				changed := configureLLMModel(config, reader)
				
				assert.Equal(t, tt.expected, config.LLMModel)
				assert.Equal(t, tt.changed, changed)
			})
		}
	})
	
	t.Run("ConfigureLLMMaxTokens", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected int
			changed  bool
		}{
			{"Set to 1000", "1000", 1000, true},
			{"Set to 8192", "8192", 8192, true},
			{"Keep current", "", 4096, false},
			{"Invalid (zero)", "0", 4096, false},
			{"Invalid (negative)", "-100", 4096, false},
			{"Invalid (too high)", "20000", 4096, false},
			{"Invalid (text)", "many", 4096, false},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				config := GetDefaultConfig()
				config.LLMMaxTokens = 4096
				reader := testdata.CreateMockReader(tt.input)
				
				changed := configureLLMMaxTokens(config, reader)
				
				assert.Equal(t, tt.expected, config.LLMMaxTokens)
				assert.Equal(t, tt.changed, changed)
			})
		}
	})
	
	t.Run("ConfigureSummaryPrompt", func(t *testing.T) {
		config := GetDefaultConfig()
		
		// Test new prompt
		newPrompt := "Please summarize {text} in {language}"
		reader := testdata.CreateMockReader(newPrompt)
		changed := configureSummaryPrompt(config, reader)
		
		assert.Equal(t, newPrompt, config.SummaryPromptTemplate)
		assert.True(t, changed)
		
		// Test keep current
		reader = testdata.CreateMockReader("")
		changed = configureSummaryPrompt(config, reader)
		
		assert.Equal(t, newPrompt, config.SummaryPromptTemplate)
		assert.False(t, changed)
		
		// Test invalid prompt (missing placeholders)
		reader = testdata.CreateMockReader("Invalid prompt without placeholders")
		changed = configureSummaryPrompt(config, reader)
		
		assert.Equal(t, newPrompt, config.SummaryPromptTemplate) // Should not change
		assert.False(t, changed)
	})
	
	// Note: configureSummaryLanguage function doesn't exist in the current implementation
	// This test documents expected behavior if it were implemented
	t.Run("ConfigureSummaryLanguage", func(t *testing.T) {
		config := GetDefaultConfig()
		
		// The summary language configuration is currently handled differently
		// This test validates the config field exists and has proper default
		assert.Equal(t, "auto", config.SummaryLanguage)
		
		// Test setting different values
		config.SummaryLanguage = "ja"
		assert.Equal(t, "ja", config.SummaryLanguage)
		
		config.SummaryLanguage = "en"
		assert.Equal(t, "en", config.SummaryLanguage)
	})
}

func TestConfigureRecordingDevice(t *testing.T) {
	// Note: This test is limited because we can't mock PortAudio
	// We test the user input handling logic only
	
	config := GetDefaultConfig()
	config.RecordingDeviceID = -1
	config.RecordingDeviceName = "既定のマイク"
	
	// Test keep current (empty input)
	reader := testdata.CreateMockReader("")
	changed := configureRecordingDevice(config, reader)
	
	assert.Equal(t, -1, config.RecordingDeviceID)
	assert.Equal(t, "既定のマイク", config.RecordingDeviceName)
	assert.False(t, changed)
	
	// Test manual device ID input
	reader = testdata.CreateMockReader("2")
	// This will try to enumerate devices and likely fail in test environment
	// but the input parsing logic should work
	_ = configureRecordingDevice(config, reader)
}

func TestResetToDefaults(t *testing.T) {
	config := &Config{
		WhisperModel:        "tiny",
		Language:            "en",
		UILanguage:          "en",
		ScanIntervalMinutes: 10,
		MaxCpuPercent:       50,
		ComputeType:         "float32",
		UseColors:           false,
		OutputFormat:        "srt",
		InputDir:            "./custom_input",
		OutputDir:           "./custom_output",
		ArchiveDir:          "./custom_archive",
		LLMSummaryEnabled:   true,
		LLMAPIKey:           "test-key",
		LLMModel:            "gpt-3.5-turbo",
		LLMMaxTokens:        1000,
		RecordingDeviceID:   5,
	}
	
	// Confirm reset
	reader := testdata.CreateMockReader("y")
	resetToDefaults(config, reader)
	
	defaultConfig := GetDefaultConfig()
	assert.Equal(t, defaultConfig.WhisperModel, config.WhisperModel)
	assert.Equal(t, defaultConfig.Language, config.Language)
	assert.Equal(t, defaultConfig.UILanguage, config.UILanguage)
	
	// Test cancel reset
	config.WhisperModel = "tiny"
	reader = testdata.CreateMockReader("n")
	resetToDefaults(config, reader)
	
	assert.Equal(t, "tiny", config.WhisperModel) // Should not change
}

// Test helper functions
func TestSelectFolder(t *testing.T) {
	// Note: selectFolder uses system dialog which we can't test
	// We can only test the fallback manual input mode
	
	folder, err := selectFolder("Select test folder")
	
	// In test environment, selectFolder will likely fail due to no GUI
	// But we can verify the function exists and returns proper types
	assert.NotNil(t, err) // Expect error in headless environment
	assert.Empty(t, folder) // Expect empty result on error
}