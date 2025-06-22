package testdata

import (
	"log"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/hirokitakamura/koemoji-go/internal/config"
	"github.com/hirokitakamura/koemoji-go/internal/logger"
)

// CreateTestConfig creates a test configuration for LLM testing
func CreateTestConfig(t *testing.T) *config.Config {
	t.Helper()

	return &config.Config{
		WhisperModel:          "base",
		Language:              "ja",
		UILanguage:            "ja",
		LLMSummaryEnabled:     true,
		LLMAPIProvider:        "openai",
		LLMAPIKey:             "sk-test1234567890abcdef1234567890abcdef12345678",
		LLMModel:              "gpt-4o",
		LLMMaxTokens:          4096,
		SummaryPromptTemplate: "以下の文字起こしテキストを{language}で要約してください。重要なポイントを箇条書きでまとめ、全体の概要も含めてください。\n\n{text}",
		SummaryLanguage:       "auto",
	}
}

// CreateTestConfigWithCustomValues creates a test config with custom values
func CreateTestConfigWithCustomValues(summaryEnabled bool, apiKey, provider, model string, maxTokens int) *config.Config {
	return &config.Config{
		WhisperModel:          "base",
		Language:              "ja",
		UILanguage:            "ja",
		LLMSummaryEnabled:     summaryEnabled,
		LLMAPIProvider:        provider,
		LLMAPIKey:             apiKey,
		LLMModel:              model,
		LLMMaxTokens:          maxTokens,
		SummaryPromptTemplate: "以下の文字起こしテキストを{language}で要約してください。重要なポイントを箇条書きでまとめ、全体の概要も含めてください。\n\n{text}",
		SummaryLanguage:       "auto",
	}
}

// CreateTestLogger creates a test logger with buffer
func CreateTestLogger() (*log.Logger, *[]logger.LogEntry, *sync.RWMutex) {
	testLogger := log.New(os.Stdout, "", log.LstdFlags)
	logBuffer := &[]logger.LogEntry{}
	logMutex := &sync.RWMutex{}
	return testLogger, logBuffer, logMutex
}

// GetTestText returns sample text for testing
func GetTestText() string {
	return `これは音声文字起こしのテストデータです。
今日の会議では、以下の内容について話し合いました。

1. プロジェクトの進捗状況について
2. 次四半期の目標設定
3. チームメンバーの役割分担
4. 予算とリソースの配分

特に重要な決定事項として、新しいマーケティング戦略の導入が決まりました。
これにより、来月から新しい取り組みを開始する予定です。

質疑応答では、実装スケジュールについて詳細な議論が行われ、
各チームのタイムラインが確認されました。`
}

// GetLongTestText returns a longer text for testing token limits
func GetLongTestText() string {
	baseText := GetTestText()
	longText := ""

	// Repeat the text multiple times to create a long text
	for i := 0; i < 20; i++ {
		longText += baseText + "\n\n"
	}

	return longText
}

// GetTestTextInEnglish returns sample text in English
func GetTestTextInEnglish() string {
	return `This is a test data for audio transcription.
In today's meeting, we discussed the following topics:

1. Project progress status
2. Next quarter goal setting
3. Team member role assignments
4. Budget and resource allocation

As an important decision, we decided to introduce a new marketing strategy.
This will allow us to start new initiatives from next month.

During the Q&A session, detailed discussions were held about the implementation schedule,
and each team's timeline was confirmed.`
}

// AssertContainsJapanese checks if the text contains Japanese characters
func AssertContainsJapanese(t *testing.T, text string) {
	t.Helper()

	hasJapanese := false
	for _, r := range text {
		if (r >= 0x3040 && r <= 0x309F) || // Hiragana
			(r >= 0x30A0 && r <= 0x30FF) || // Katakana
			(r >= 0x4E00 && r <= 0x9FAF) { // Kanji
			hasJapanese = true
			break
		}
	}

	if !hasJapanese {
		t.Errorf("Expected text to contain Japanese characters, but got: %s", text)
	}
}

// AssertValidSummaryFormat checks if the summary has the expected format
func AssertValidSummaryFormat(t *testing.T, summary string) {
	t.Helper()

	if len(summary) == 0 {
		t.Error("Summary should not be empty")
		return
	}

	// Check for bullet points or structured content
	hasBulletPoints := false
	lines := []string{"•", "・", "-", "*", "1.", "2.", "3."}
	for _, line := range lines {
		if strings.Contains(summary, line) {
			hasBulletPoints = true
			break
		}
	}

	if !hasBulletPoints {
		t.Logf("Warning: Summary might not contain bullet points: %s", summary)
	}
}
