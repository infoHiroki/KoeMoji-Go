package whisper

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hirokitakamura/koemoji-go/internal/config"
	"github.com/hirokitakamura/koemoji-go/internal/logger"
	"github.com/hirokitakamura/koemoji-go/internal/ui"
)

func getWhisperCommand() string {
	// 1. 通常のPATHで試す
	if _, err := exec.LookPath("whisper-ctranslate2"); err == nil {
		return "whisper-ctranslate2"
	}

	// 2. 標準的なインストール場所を検索
	standardPaths := []string{
		filepath.Join(os.Getenv("HOME"), ".local", "bin", "whisper-ctranslate2"),                    // macOS user install
		"/usr/local/bin/whisper-ctranslate2",                                                        // macOS system
		filepath.Join(os.Getenv("HOME"), "Library", "Python", "3.12", "bin", "whisper-ctranslate2"), // macOS Python 3.12
		filepath.Join(os.Getenv("HOME"), "Library", "Python", "3.11", "bin", "whisper-ctranslate2"), // macOS Python 3.11
		filepath.Join(os.Getenv("HOME"), "Library", "Python", "3.10", "bin", "whisper-ctranslate2"), // macOS Python 3.10
		filepath.Join(os.Getenv("HOME"), "Library", "Python", "3.9", "bin", "whisper-ctranslate2"),  // macOS Python 3.9
	}

	for _, path := range standardPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "whisper-ctranslate2" // フォールバック
}

func isFasterWhisperAvailable() bool {
	whisperCmd := getWhisperCommand()
	// If we fall back to the command name (not an absolute path), check if it's in PATH
	if whisperCmd == "whisper-ctranslate2" {
		if _, err := exec.LookPath(whisperCmd); err != nil {
			return false
		}
	}
	
	cmd := exec.Command(whisperCmd, "--help")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// IsFasterWhisperAvailableForTesting exports the availability check for testing
func IsFasterWhisperAvailableForTesting() bool {
	return isFasterWhisperAvailable()
}

func installFasterWhisper(log *log.Logger, logBuffer *[]logger.LogEntry, logMutex *sync.RWMutex) error {
	logger.LogInfo(log, logBuffer, logMutex, "Installing faster-whisper and whisper-ctranslate2...")
	cmd := exec.Command("pip", "install", "faster-whisper", "whisper-ctranslate2")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pip install failed: %w", err)
	}
	logger.LogInfo(log, logBuffer, logMutex, "FasterWhisper installed successfully")
	return nil
}

func TranscribeAudio(config *config.Config, log *log.Logger, logBuffer *[]logger.LogEntry,
	logMutex *sync.RWMutex, debugMode bool, inputFile string) error {

	// セキュリティチェック: inputディレクトリ内のファイルのみ許可
	absPath, err := filepath.Abs(inputFile)
	if err != nil {
		msg := ui.GetMessages(config)
		return fmt.Errorf(msg.InvalidPath, err)
	}
	inputDir, err := filepath.Abs(config.InputDir)
	if err != nil {
		msg := ui.GetMessages(config)
		return fmt.Errorf(msg.InvalidPath, err)
	}
	if !strings.HasPrefix(absPath, inputDir+string(os.PathSeparator)) {
		msg := ui.GetMessages(config)
		return fmt.Errorf(msg.InvalidPath, inputFile)
	}

	// 入力ファイルの存在チェック
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputFile)
	}

	// 入力ディレクトリの存在チェック
	if _, err := os.Stat(config.InputDir); os.IsNotExist(err) {
		return fmt.Errorf("input directory does not exist: %s", config.InputDir)
	}

	// 出力ディレクトリの作成確認
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	whisperCmd := getWhisperCommand()

	cmd := exec.Command(whisperCmd,
		"--model", config.WhisperModel,
		"--language", config.Language,
		"--output_dir", config.OutputDir,
		"--output_format", config.OutputFormat,
		"--compute_type", config.ComputeType,
		"--verbose", "True", // Enable verbose for progress
		inputFile,
	)

	logger.LogDebug(log, logBuffer, logMutex, debugMode, "Whisper command: %s", strings.Join(cmd.Args, " "))

	// Start progress monitoring
	startTime := time.Now()
	done := make(chan bool)

	// Monitor progress in background
	go monitorProgress(log, logBuffer, logMutex, filepath.Base(inputFile), startTime, done)

	// Capture and display output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		done <- true
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		done <- true
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		done <- true
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Read output in background
	go readCommandOutput(log, logBuffer, logMutex, debugMode, stdout, "STDOUT")
	go readCommandOutput(log, logBuffer, logMutex, debugMode, stderr, "STDERR")

	// Wait for completion
	err = cmd.Wait()

	// Stop progress monitoring
	done <- true

	if err != nil {
		msg := ui.GetMessages(config)
		return fmt.Errorf(msg.TranscribeFail, err)
	}

	return nil
}

func readCommandOutput(log *log.Logger, logBuffer *[]logger.LogEntry, logMutex *sync.RWMutex,
	debugMode bool, pipe io.ReadCloser, source string) {

	defer pipe.Close()
	scanner := bufio.NewScanner(pipe)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			// Log other output for debugging
			logger.LogDebug(log, logBuffer, logMutex, debugMode, "[%s] %s", source, line)
		}
	}
}

func monitorProgress(log *log.Logger, logBuffer *[]logger.LogEntry, logMutex *sync.RWMutex,
	filename string, startTime time.Time, done chan bool) {

	ticker := time.NewTicker(30 * time.Second) // 30秒ごとに進行状況を報告
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			elapsed := time.Since(startTime)
			logger.LogInfo(log, logBuffer, logMutex, "Still processing %s (elapsed: %s)", filename, formatDuration(elapsed))
		}
	}
}

func EnsureDependencies(config *config.Config, log *log.Logger, logBuffer *[]logger.LogEntry,
	logMutex *sync.RWMutex, debugMode bool) {

	if !isFasterWhisperAvailable() {
		logger.LogInfo(log, logBuffer, logMutex, "FasterWhisper not found. Attempting to install...")
		if err := installFasterWhisper(log, logBuffer, logMutex); err != nil {
			logger.LogError(log, logBuffer, logMutex, "FasterWhisper installation failed: %v", err)
			logger.LogError(log, logBuffer, logMutex, "Please install manually: pip install faster-whisper whisper-ctranslate2")
			os.Exit(1)
		}
	} else {
		logger.LogDebug(log, logBuffer, logMutex, debugMode, "FasterWhisper is available")
	}
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	} else {
		return fmt.Sprintf("%ds", seconds)
	}
}
