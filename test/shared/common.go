package testutil

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/hirokitakamura/koemoji-go/internal/config"
	"github.com/hirokitakamura/koemoji-go/internal/logger"
)

// CreateTestConfig creates a basic test configuration
func CreateTestConfig(t *testing.T) *config.Config {
	t.Helper()

	tempDir := t.TempDir()

	return &config.Config{
		WhisperModel:        "base",
		Language:            "ja",
		UILanguage:          "ja",
		ScanIntervalMinutes: 1,
		MaxCpuPercent:       80,
		ComputeType:         "int8",
		UseColors:           false,
		OutputFormat:        "txt",
		InputDir:            filepath.Join(tempDir, "input"),
		OutputDir:           filepath.Join(tempDir, "output"),
		ArchiveDir:          filepath.Join(tempDir, "archive"),
		LLMSummaryEnabled:   false,
		LLMAPIProvider:      "openai",
		LLMAPIKey:           "",
		LLMModel:            "gpt-4o",
		LLMMaxTokens:        4096,
		SummaryLanguage:     "auto",
		RecordingDeviceID:   -1,
		RecordingDeviceName: "既定のマイク",
		RecordingMaxHours:   0,
		RecordingMaxFileMB:  0,
	}
}

// CreateTestLogger creates a test logger with buffer
func CreateTestLogger() (*log.Logger, *[]logger.LogEntry, *sync.RWMutex) {
	testLogger := log.New(os.Stdout, "", log.LstdFlags)
	logBuffer := &[]logger.LogEntry{}
	logMutex := &sync.RWMutex{}
	return testLogger, logBuffer, logMutex
}

// CreateTestDirectories creates test directories
func CreateTestDirectories(t *testing.T, dirs ...string) {
	t.Helper()

	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
}

// AssertFileNotExists checks if a file does not exist
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err == nil {
		t.Errorf("Expected file %s to not exist, but it does", path)
	}
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
