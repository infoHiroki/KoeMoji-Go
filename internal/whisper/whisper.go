package whisper

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/infoHiroki/KoeMoji-Go/internal/config"
	"github.com/infoHiroki/KoeMoji-Go/internal/logger"
	"github.com/infoHiroki/KoeMoji-Go/internal/ui"
)

func getWhisperCommand() string {
	return getWhisperCommandWithDebug(nil, nil, nil, false)
}

func getWhisperCommandWithDebug(log *log.Logger, logBuffer *[]logger.LogEntry, logMutex *sync.RWMutex, debugMode bool) string {
	// 1. 通常のPATHで試す
	if _, err := exec.LookPath("whisper-ctranslate2"); err == nil {
		if debugMode && log != nil {
			logger.LogDebug(log, logBuffer, logMutex, debugMode, "Found whisper-ctranslate2 in PATH")
		}
		return "whisper-ctranslate2"
	}

	// 2. 標準的なインストール場所を検索
	var standardPaths []string
	
	if runtime.GOOS == "windows" {
		// Windows specific paths
		username := os.Getenv("USERNAME")
		if username == "" {
			username = os.Getenv("USER") // Fallback
		}
		
		// Common Python installation paths on Windows
		pythonVersions := []string{"312", "311", "310", "39", "38"}
		
		for _, version := range pythonVersions {
			// Standard Python installations
			standardPaths = append(standardPaths, 
				filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Python", "Python"+version, "Scripts", "whisper-ctranslate2.exe"),
				filepath.Join(os.Getenv("APPDATA"), "Python", "Python"+version, "Scripts", "whisper-ctranslate2.exe"),
				// pip --user installations (Roaming)
				filepath.Join(os.Getenv("APPDATA"), "Roaming", "Python", "Python"+version, "Scripts", "whisper-ctranslate2.exe"),
				// System-wide installations
				filepath.Join("C:\\", "Python"+version, "Scripts", "whisper-ctranslate2.exe"),
			)
			
			// Also try with dot notation (e.g., Python3.12)
			versionWithDot := string(version[0]) + "." + version[1:]
			standardPaths = append(standardPaths,
				filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Python", "Python"+versionWithDot, "Scripts", "whisper-ctranslate2.exe"),
				filepath.Join("C:\\", "Python"+versionWithDot, "Scripts", "whisper-ctranslate2.exe"),
			)
		}
		
		// User profile paths
		if username != "" {
			userProfilePaths := []string{
				filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", "Programs", "Python", "Python312", "Scripts", "whisper-ctranslate2.exe"),
				filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", "Programs", "Python", "Python311", "Scripts", "whisper-ctranslate2.exe"),
				filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", "Programs", "Python", "Python310", "Scripts", "whisper-ctranslate2.exe"),
			}
			standardPaths = append(standardPaths, userProfilePaths...)
		}
		
		// Anaconda/Miniconda paths
		if username != "" {
			standardPaths = append(standardPaths,
				filepath.Join("C:", "Users", username, "anaconda3", "Scripts", "whisper-ctranslate2.exe"),
				filepath.Join("C:", "Users", username, "miniconda3", "Scripts", "whisper-ctranslate2.exe"),
			)
		}
	} else {
		// macOS/Linux paths
		standardPaths = []string{
			filepath.Join(os.Getenv("HOME"), ".local", "bin", "whisper-ctranslate2"),                    // macOS user install
			"/usr/local/bin/whisper-ctranslate2",                                                        // macOS system
			"/opt/homebrew/bin/whisper-ctranslate2",                                                     // Homebrew Apple Silicon
			"/usr/local/bin/whisper-ctranslate2",                                                        // Homebrew Intel
			filepath.Join(os.Getenv("HOME"), "Library", "Python", "3.12", "bin", "whisper-ctranslate2"), // macOS Python 3.12
			filepath.Join(os.Getenv("HOME"), "Library", "Python", "3.11", "bin", "whisper-ctranslate2"), // macOS Python 3.11
			filepath.Join(os.Getenv("HOME"), "Library", "Python", "3.10", "bin", "whisper-ctranslate2"), // macOS Python 3.10
			filepath.Join(os.Getenv("HOME"), "Library", "Python", "3.9", "bin", "whisper-ctranslate2"),  // macOS Python 3.9
		}
	}

	for _, path := range standardPaths {
		if debugMode && log != nil {
			logger.LogDebug(log, logBuffer, logMutex, debugMode, "Checking for whisper-ctranslate2 at: %s", path)
		}
		if _, err := os.Stat(path); err == nil {
			if debugMode && log != nil {
				logger.LogDebug(log, logBuffer, logMutex, debugMode, "Found whisper-ctranslate2 at: %s", path)
			}
			return path
		}
	}

	if debugMode && log != nil {
		logger.LogDebug(log, logBuffer, logMutex, debugMode, "whisper-ctranslate2 not found in any standard location")
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
	
	cmd := createCommand(whisperCmd, "--help")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// IsFasterWhisperAvailableForTesting exports the availability check for testing
func IsFasterWhisperAvailableForTesting() bool {
	return isFasterWhisperAvailable()
}

func installFasterWhisper(log *log.Logger, logBuffer *[]logger.LogEntry, logMutex *sync.RWMutex) error {
	// Step 1: Upgrade pip to ensure proper dependency resolution
	logger.LogInfo(log, logBuffer, logMutex, "Upgrading pip to latest version...")
	pipUpgradeCmd := createCommand("python", "-m", "pip", "install", "--upgrade", "pip")
	if err := pipUpgradeCmd.Run(); err != nil {
		logger.LogError(log, logBuffer, logMutex, "pip upgrade failed (non-fatal): %v", err)
		// Continue even if pip upgrade fails
	}

	// Step 2: Install faster-whisper with explicit dependencies
	// Include 'requests' explicitly to avoid indirect dependency resolution issues
	// Background: requests is an indirect dependency of huggingface-hub (used by faster-whisper)
	// Older pip versions or certain environments may fail to resolve it automatically
	logger.LogInfo(log, logBuffer, logMutex, "Installing faster-whisper and dependencies...")
	cmd := createCommand("pip", "install", "requests", "faster-whisper", "whisper-ctranslate2")
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

	whisperCmd := getWhisperCommandWithDebug(log, logBuffer, logMutex, debugMode)

	// Build command arguments
	args := []string{
		"--model", config.WhisperModel,
		"--language", config.Language,
		"--output_dir", config.OutputDir,
		"--output_format", config.OutputFormat,
		"--compute_type", config.ComputeType,
	}
	
	// Add device parameter for CPU-specific compute types
	if config.ComputeType == "int8" {
		args = append(args, "--device", "cpu")
	}
	
	// Add verbose and input file
	args = append(args, "--verbose", "True", inputFile)
	
	cmd := createCommand(whisperCmd, args...)

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
		// Check if whisper-ctranslate2 is not found
		if errors.Is(err, exec.ErrNotFound) || strings.Contains(err.Error(), "executable file not found") {
			msg := ui.GetMessages(config)
			return fmt.Errorf("%s\n%s", msg.WhisperNotFound, msg.WhisperLocation)
		}
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
		
		// Check for GPU-related errors and provide detailed guidance
		errorStr := err.Error()
		if isGPURelatedError(errorStr) {
			return createGPUErrorMessage(config, err)
		}
		
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
	logMutex *sync.RWMutex, debugMode bool) error {

	if !isFasterWhisperAvailable() {
		logger.LogInfo(log, logBuffer, logMutex, "FasterWhisper not found. Attempting to install...")
		if err := installFasterWhisper(log, logBuffer, logMutex); err != nil {
			logger.LogError(log, logBuffer, logMutex, "FasterWhisper automatic installation failed: %v", err)
			logger.LogError(log, logBuffer, logMutex, "Please install Python 3.12 and restart, or manually run: pip install faster-whisper whisper-ctranslate2")
			return fmt.Errorf("FasterWhisper installation failed: %v", err)
		}
	} else {
		logger.LogDebug(log, logBuffer, logMutex, debugMode, "FasterWhisper is available")
	}
	return nil
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

// isGPURelatedError checks if the error is related to GPU/CUDA issues
func isGPURelatedError(errorStr string) bool {
	errorLower := strings.ToLower(errorStr)
	
	// Common GPU-related error patterns
	gpuErrorPatterns := []string{
		"cuda",
		"gpu",
		"float16",
		"int8_float16", 
		"device or backend do not support",
		"efficient float16 computation",
		"efficient int8_float16 computation",
		"nvidia",
		"cudnn",
		"cublas",
		"out of memory",
		"insufficient memory",
	}
	
	for _, pattern := range gpuErrorPatterns {
		if strings.Contains(errorLower, pattern) {
			return true
		}
	}
	
	return false
}

// createGPUErrorMessage creates a user-friendly error message with guidance
func createGPUErrorMessage(config *config.Config, originalErr error) error {
	var guidance string
	if config.UILanguage == "ja" {
		guidance = fmt.Sprintf(`GPU処理に失敗しました。

考えられる原因:
• NVIDIA CUDA Toolkit未インストール
• GPU非対応またはVRAMメモリ不足  
• 古いGPUドライバー
• compute_type設定とGPUの不整合

推奨解決策:
1. config.jsonで "compute_type": "int8" に変更 (CPU使用、最も安定)
2. または、NVIDIA CUDA Toolkit をインストール
3. NVIDIAドライバーを最新版に更新

現在の設定: compute_type="%s"
元のエラー: %v`, config.ComputeType, originalErr)
	} else {
		guidance = fmt.Sprintf(`GPU processing failed.

Possible causes:
• NVIDIA CUDA Toolkit not installed
• GPU incompatible or insufficient VRAM
• Outdated GPU drivers
• compute_type incompatible with GPU

Recommended solutions:
1. Change "compute_type": "int8" in config.json (CPU usage, most stable)
2. Or install NVIDIA CUDA Toolkit  
3. Update NVIDIA drivers to latest version

Current setting: compute_type="%s"
Original error: %v`, config.ComputeType, originalErr)
	}
	
	return fmt.Errorf(guidance)
}

// IsGPURelatedErrorForTesting exports isGPURelatedError for testing
func IsGPURelatedErrorForTesting(errorStr string) bool {
	return isGPURelatedError(errorStr)
}

// CreateGPUErrorMessageForTesting exports createGPUErrorMessage for testing
func CreateGPUErrorMessageForTesting(config *config.Config, originalErr error) error {
	return createGPUErrorMessage(config, originalErr)
}
