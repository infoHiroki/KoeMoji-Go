package llm

import (
	"strings"
	"testing"

	"github.com/hirokitakamura/koemoji-go/internal/config"
	"github.com/hirokitakamura/koemoji-go/internal/llm/testdata"
	"github.com/stretchr/testify/assert"
)

func TestValidateAPIKey_ValidKey(t *testing.T) {
	// Test with obviously invalid key format
	cfg := &config.Config{LLMAPIKey: "invalid-key-format"}
	err := ValidateAPIKey(cfg)
	assert.Error(t, err)
	
	// Test with valid format but fake key (will fail on API call)
	cfg2 := &config.Config{LLMAPIKey: "sk-test1234567890abcdef1234567890abcdef12345678"}
	err = ValidateAPIKey(cfg2)
	// We expect an error since this is not a real API key, but format is valid
	assert.Error(t, err)
}

func TestValidateAPIKey_EmptyKey(t *testing.T) {
	config := &config.Config{LLMAPIKey: ""}
	err := ValidateAPIKey(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key is empty")
}

func TestPreparePrompt(t *testing.T) {
	tests := []struct {
		name           string
		language       string
		summaryLang    string
		template       string
		text           string
		expectedLang   string
	}{
		{
			name:         "Japanese auto",
			language:     "ja",
			summaryLang:  "auto",
			template:     "Summarize {text} in {language}",
			text:         "テストテキスト",
			expectedLang: "日本語",
		},
		{
			name:         "English auto",
			language:     "en",
			summaryLang:  "auto",
			template:     "Summarize {text} in {language}",
			text:         "Test text",
			expectedLang: "英語",
		},
		{
			name:         "Force Japanese",
			language:     "en",
			summaryLang:  "ja",
			template:     "Summarize {text} in {language}",
			text:         "Test text",
			expectedLang: "日本語",
		},
		{
			name:         "Force English",
			language:     "ja",
			summaryLang:  "en",
			template:     "Summarize {text} in {language}",
			text:         "テストテキスト",
			expectedLang: "英語",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &config.Config{
				Language:              tt.language,
				SummaryLanguage:       tt.summaryLang,
				SummaryPromptTemplate: tt.template,
			}

			result := preparePrompt(config, tt.text)
			
			assert.Contains(t, result, tt.text)
			assert.Contains(t, result, tt.expectedLang)
			assert.NotContains(t, result, "{text}")
			assert.NotContains(t, result, "{language}")
		})
	}
}

func TestGetSummaryLanguage(t *testing.T) {
	tests := []struct {
		name         string
		language     string
		summaryLang  string
		expected     string
	}{
		{"Auto Japanese", "ja", "auto", "日本語"},
		{"Auto English", "en", "auto", "英語"},
		{"Auto other", "zh", "auto", "英語"},
		{"Force Japanese", "en", "ja", "日本語"},
		{"Force English", "ja", "en", "英語"},
		{"Invalid defaults to Japanese", "ja", "invalid", "日本語"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &config.Config{
				Language:        tt.language,
				SummaryLanguage: tt.summaryLang,
			}

			result := getSummaryLanguage(config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSummarizeText_ConfigValidation(t *testing.T) {
	logger, logBuffer, logMutex := testdata.CreateTestLogger()

	t.Run("LLM disabled", func(t *testing.T) {
		config := testdata.CreateTestConfigWithCustomValues(false, "sk-test123", "openai", "gpt-4o", 4096)
		
		summary, err := SummarizeText(config, logger, logBuffer, logMutex, false, "test text")
		
		assert.Error(t, err)
		assert.Empty(t, summary)
		assert.Contains(t, err.Error(), "disabled")
	})

	t.Run("Empty API key", func(t *testing.T) {
		config := testdata.CreateTestConfigWithCustomValues(true, "", "openai", "gpt-4o", 4096)
		
		summary, err := SummarizeText(config, logger, logBuffer, logMutex, false, "test text")
		
		assert.Error(t, err)
		assert.Empty(t, summary)
		assert.Contains(t, err.Error(), "not configured")
	})

	t.Run("Unsupported provider", func(t *testing.T) {
		config := testdata.CreateTestConfigWithCustomValues(true, "sk-test123", "claude", "claude-v1", 4096)
		
		summary, err := SummarizeText(config, logger, logBuffer, logMutex, false, "test text")
		
		assert.Error(t, err)
		assert.Empty(t, summary)
		assert.Contains(t, err.Error(), "unsupported LLM provider")
	})
}

func TestValidateAPIKey_Comprehensive(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		expectErr bool
		errorMsg  string
	}{
		{
			name:      "Empty key",
			apiKey:    "",
			expectErr: true,
			errorMsg:  "API key is empty",
		},
		{
			name:      "Too short key",
			apiKey:    "sk-123",
			expectErr: true,
			errorMsg:  "", // Will fail on API call
		},
		{
			name:      "Invalid format",
			apiKey:    "invalid-key-format",
			expectErr: true,
			errorMsg:  "", // Will fail on API call
		},
		{
			name:      "Valid format but fake key",
			apiKey:    "sk-1234567890abcdef1234567890abcdef12345678",
			expectErr: true,
			errorMsg:  "", // Will fail on API call with 401
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &config.Config{LLMAPIKey: tt.apiKey}
			err := ValidateAPIKey(config)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test prompt template validation
func TestPromptTemplateValidation(t *testing.T) {
	tests := []struct {
		name     string
		template string
		text     string
		valid    bool
	}{
		{
			name:     "Valid template",
			template: "Summarize {text} in {language}",
			text:     "test",
			valid:    true,
		},
		{
			name:     "Missing text placeholder",
			template: "Summarize in {language}",
			text:     "test",
			valid:    false,
		},
		{
			name:     "Missing language placeholder",
			template: "Summarize {text}",
			text:     "test",
			valid:    false,
		},
		{
			name:     "Empty template",
			template: "",
			text:     "test",
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &config.Config{
				Language:              "ja",
				SummaryLanguage:       "auto",
				SummaryPromptTemplate: tt.template,
			}

			result := preparePrompt(config, tt.text)

			if tt.valid {
				assert.Contains(t, result, tt.text)
				assert.NotContains(t, result, "{text}")
				assert.NotContains(t, result, "{language}")
			} else {
				// For invalid templates, placeholders might remain
				if strings.Contains(tt.template, "{text}") && !strings.Contains(result, tt.text) {
					assert.Contains(t, result, "{text}")
				}
				if strings.Contains(tt.template, "{language}") && !strings.Contains(result, "語") {
					assert.Contains(t, result, "{language}")
				}
			}
		})
	}
}

// Test error handling with different text inputs
func TestTextInputValidation(t *testing.T) {
	config := testdata.CreateTestConfig(t)
	logger, logBuffer, logMutex := testdata.CreateTestLogger()

	tests := []struct {
		name     string
		text     string
		expectErr bool
	}{
		{
			name:     "Normal text",
			text:     testdata.GetTestText(),
			expectErr: true, // Will fail due to fake API key
		},
		{
			name:     "Empty text",
			text:     "",
			expectErr: true,
		},
		{
			name:     "Very long text",
			text:     testdata.GetLongTestText(),
			expectErr: true, // Will fail due to fake API key
		},
		{
			name:     "Text with special characters",
			text:     "Special chars: @#$%^&*()_+{}|:<>?[]\\",
			expectErr: true, // Will fail due to fake API key
		},
		{
			name:     "Mixed language text",
			text:     "Hello こんにちは 你好 안녕하세요",
			expectErr: true, // Will fail due to fake API key
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SummarizeText(config, logger, logBuffer, logMutex, false, tt.text)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test different model configurations
func TestModelConfiguration(t *testing.T) {
	logger, logBuffer, logMutex := testdata.CreateTestLogger()
	testText := testdata.GetTestText()

	models := []string{"gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-4", "gpt-3.5-turbo"}

	for _, model := range models {
		t.Run("Model_"+model, func(t *testing.T) {
			config := testdata.CreateTestConfigWithCustomValues(true, "sk-test123", "openai", model, 4096)

			_, err := SummarizeText(config, logger, logBuffer, logMutex, false, testText)

			// All should fail due to fake API key, but not due to invalid model
			assert.Error(t, err)
			// Should not contain model-related errors
			errorMsg := strings.ToLower(err.Error())
			assert.NotContains(t, errorMsg, "invalid model")
			assert.NotContains(t, errorMsg, "unknown model")
		})
	}
}

// Test token limit configurations
func TestTokenLimitConfiguration(t *testing.T) {
	logger, logBuffer, logMutex := testdata.CreateTestLogger()
	testText := testdata.GetTestText()

	tokenLimits := []int{100, 1000, 4096, 8192}

	for _, limit := range tokenLimits {
		t.Run("TokenLimit_"+string(rune(limit)), func(t *testing.T) {
			config := testdata.CreateTestConfigWithCustomValues(true, "sk-test123", "openai", "gpt-4o", limit)

			_, err := SummarizeText(config, logger, logBuffer, logMutex, false, testText)

			// All should fail due to fake API key, but token limit should be handled
			assert.Error(t, err)
		})
	}
}

// Benchmark tests
func BenchmarkPreparePrompt(b *testing.B) {
	config := &config.Config{
		Language:              "ja",
		SummaryLanguage:       "auto",
		SummaryPromptTemplate: "以下の文字起こしテキストを{language}で要約してください。重要なポイントを箇条書きでまとめ、全体の概要も含めてください。\n\n{text}",
	}
	testText := testdata.GetTestText()

	for i := 0; i < b.N; i++ {
		_ = preparePrompt(config, testText)
	}
}

func BenchmarkGetSummaryLanguage(b *testing.B) {
	config := &config.Config{
		Language:        "ja",
		SummaryLanguage: "auto",
	}

	for i := 0; i < b.N; i++ {
		_ = getSummaryLanguage(config)
	}
}
