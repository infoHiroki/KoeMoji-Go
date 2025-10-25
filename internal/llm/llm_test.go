package llm

import (
	"strings"
	"testing"

	"github.com/infoHiroki/KoeMoji-Go/internal/config"
	"github.com/infoHiroki/KoeMoji-Go/internal/llm/testdata"
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
		name         string
		language     string
		summaryLang  string
		template     string
		text         string
		expectedLang string
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
		name        string
		language    string
		summaryLang string
		expected    string
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
		name      string
		text      string
		expectErr bool
	}{
		{
			name:      "Normal text",
			text:      testdata.GetTestText(),
			expectErr: true, // Will fail due to fake API key
		},
		{
			name:      "Empty text",
			text:      "",
			expectErr: true,
		},
		{
			name:      "Very long text",
			text:      testdata.GetLongTestText(),
			expectErr: true, // Will fail due to fake API key
		},
		{
			name:      "Text with special characters",
			text:      "Special chars: @#$%^&*()_+{}|:<>?[]\\",
			expectErr: true, // Will fail due to fake API key
		},
		{
			name:      "Mixed language text",
			text:      "Hello こんにちは 你好 안녕하세요",
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

// Test new implementation: text auto-append without variables
func TestPreparePrompt_TextAutoAppend(t *testing.T) {
	tests := []struct {
		name     string
		template string
		text     string
		expected string
	}{
		{
			name:     "New default prompt",
			template: "以下の文字起こしテキストを日本語で詳細に要約してください。プレーンテキストで出力し、マークダウンは使用しないでください。",
			text:     "これはテストの音声認識結果です。重要な情報が含まれています。",
			expected: "以下の文字起こしテキストを日本語で詳細に要約してください。プレーンテキストで出力し、マークダウンは使用しないでください。\n\nこれはテストの音声認識結果です。重要な情報が含まれています。",
		},
		{
			name:     "Custom prompt without variables",
			template: "簡潔に要約してください。",
			text:     "長いテキストがここに入ります。",
			expected: "簡潔に要約してください。\n\n長いテキストがここに入ります。",
		},
		{
			name:     "Long text",
			template: "要約してください。",
			text:     testdata.GetTestText(),
			expected: "要約してください。\n\n" + testdata.GetTestText(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &config.Config{
				SummaryPromptTemplate: tt.template,
			}

			result := preparePrompt(config, tt.text)

			// テスト結果を出力（実際のAPI送信データを確認）
			t.Logf("\n=== Prepared Prompt ===\n%s\n======================\n", result)

			assert.Equal(t, tt.expected, result)
			assert.Contains(t, result, tt.template)
			assert.Contains(t, result, tt.text)
			// 変数が含まれていないことを確認
			assert.NotContains(t, result, "{text}")
			assert.NotContains(t, result, "{language}")
		})
	}
}

// Test backward compatibility: old format with {text} and {language} placeholders
func TestPreparePrompt_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name         string
		template     string
		text         string
		language     string
		summaryLang  string
		expectedLang string
	}{
		{
			name:         "Old format with both placeholders (Japanese)",
			template:     "以下の文字起こしテキストを{language}で詳細に要約してください。\n\n{text}",
			text:         "これはテストです。",
			language:     "ja",
			summaryLang:  "auto",
			expectedLang: "日本語",
		},
		{
			name:         "Old format with both placeholders (English)",
			template:     "Summarize the following text in {language}:\n\n{text}",
			text:         "This is a test.",
			language:     "en",
			summaryLang:  "auto",
			expectedLang: "英語",
		},
		{
			name:         "Old format with only {text}",
			template:     "要約してください:\n\n{text}",
			text:         "テキスト内容",
			language:     "ja",
			summaryLang:  "ja",
			expectedLang: "日本語",
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

			// プレースホルダーが置換されていることを確認
			assert.NotContains(t, result, "{text}", "Placeholder {text} should be replaced")
			assert.NotContains(t, result, "{language}", "Placeholder {language} should be replaced")

			// テキストと言語が含まれていることを確認
			assert.Contains(t, result, tt.text, "Result should contain the text")
			if strings.Contains(tt.template, "{language}") {
				assert.Contains(t, result, tt.expectedLang, "Result should contain the language")
			}

			t.Logf("\n=== Backward Compatible Prompt ===\n%s\n==================================\n", result)
		})
	}
}

// 実際のAPIリクエストデータ形式を確認するテスト
func TestAPIRequestDataFormat(t *testing.T) {
	config := &config.Config{
		SummaryPromptTemplate: "以下の文字起こしテキストを日本語で詳細に要約してください。プレーンテキストで出力し、マークダウンは使用しないでください。",
		LLMModel:              "gpt-4o",
		LLMMaxTokens:          4096,
	}

	sampleText := "本日の会議では、新製品の開発スケジュールについて議論しました。来月からプロトタイプの制作を開始し、3ヶ月後にテストを実施する予定です。"

	prompt := preparePrompt(config, sampleText)

	// APIに送信される実際のデータ形式を出力
	t.Logf("\n=== API Request Data ===")
	t.Logf("Model: %s", config.LLMModel)
	t.Logf("MaxTokens: %d", config.LLMMaxTokens)
	t.Logf("\nPrompt (sent to API):\n---\n%s\n---", prompt)
	t.Logf("\nPrompt length: %d characters\n", len(prompt))

	// 基本的な検証
	assert.Contains(t, prompt, config.SummaryPromptTemplate)
	assert.Contains(t, prompt, sampleText)
	assert.True(t, len(prompt) > len(sampleText))
}
