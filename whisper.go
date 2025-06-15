package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func (app *App) getWhisperCommand() string {
	// 1. 通常のPATHで試す
	if _, err := exec.LookPath("whisper-ctranslate2"); err == nil {
		app.logDebug("Found whisper-ctranslate2 in PATH")
		return "whisper-ctranslate2"
	}

	// 2. 標準的なインストール場所を検索
	standardPaths := []string{
		filepath.Join(os.Getenv("HOME"), ".local", "bin", "whisper-ctranslate2"),                // Linux/macOS user install
		"/usr/local/bin/whisper-ctranslate2",                                                    // Linux/macOS system
		filepath.Join(os.Getenv("HOME"), "Library", "Python", "3.12", "bin", "whisper-ctranslate2"), // macOS Python 3.12
		filepath.Join(os.Getenv("HOME"), "Library", "Python", "3.11", "bin", "whisper-ctranslate2"), // macOS Python 3.11
		filepath.Join(os.Getenv("HOME"), "Library", "Python", "3.10", "bin", "whisper-ctranslate2"), // macOS Python 3.10
		filepath.Join(os.Getenv("HOME"), "Library", "Python", "3.9", "bin", "whisper-ctranslate2"),  // macOS Python 3.9
	}

	for _, path := range standardPaths {
		if _, err := os.Stat(path); err == nil {
			app.logDebug("Found whisper-ctranslate2 at: %s", path)
			return path
		}
	}

	app.logError("whisper-ctranslate2 not found in any standard location")
	return "whisper-ctranslate2" // フォールバック
}

func (app *App) isFasterWhisperAvailable() bool {
	cmd := exec.Command(app.getWhisperCommand(), "--help")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

func (app *App) installFasterWhisper() error {
	app.logInfo("Installing faster-whisper and whisper-ctranslate2...")
	cmd := exec.Command("pip", "install", "faster-whisper", "whisper-ctranslate2")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pip install failed: %w", err)
	}
	app.logInfo("FasterWhisper installed successfully")
	return nil
}

func (app *App) transcribeAudio(inputFile string) error {
	// セキュリティチェック: inputディレクトリ内のファイルのみ許可
	absPath, err := filepath.Abs(inputFile)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}
	inputDir, _ := filepath.Abs("input")
	if !strings.HasPrefix(absPath, inputDir+string(os.PathSeparator)) {
		return fmt.Errorf("file must be in input directory: %s", inputFile)
	}

	whisperCmd := app.getWhisperCommand()

	cmd := exec.Command(whisperCmd,
		"--model", app.config.WhisperModel,
		"--language", app.config.Language,
		"--output_dir", "./output",
		"--output_format", app.config.OutputFormat,
		"--compute_type", app.config.ComputeType,
		"--verbose", "True", // Enable verbose for progress
		inputFile,
	)

	app.logDebug("Whisper command: %s", strings.Join(cmd.Args, " "))

	// Start progress monitoring
	startTime := time.Now()
	done := make(chan bool)
	
	// Monitor progress in background
	go app.monitorProgress(filepath.Base(inputFile), startTime, done)

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
	go app.readCommandOutput(stdout, "STDOUT")
	go app.readCommandOutput(stderr, "STDERR")

	// Wait for completion
	err = cmd.Wait()
	
	// Stop progress monitoring
	done <- true
	
	if err != nil {
		return fmt.Errorf("whisper execution failed: %w", err)
	}

	return nil
}

func (app *App) readCommandOutput(pipe io.ReadCloser, source string) {
	defer pipe.Close()
	scanner := bufio.NewScanner(pipe)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			// Log other output for debugging
			app.logDebug("[%s] %s", source, line)
		}
	}
}

func (app *App) monitorProgress(filename string, startTime time.Time, done chan bool) {
	ticker := time.NewTicker(30 * time.Second) // 30秒ごとに進行状況を報告
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			elapsed := time.Since(startTime)
			app.logInfo("Still processing %s (elapsed: %s)", filename, app.formatDuration(elapsed))
		}
	}
}

func (app *App) ensureDependencies() {
	if !app.isFasterWhisperAvailable() {
		app.logInfo("FasterWhisper not found. Attempting to install...")
		if err := app.installFasterWhisper(); err != nil {
			app.logError("FasterWhisper installation failed: %v", err)
			app.logError("Please install manually: pip install faster-whisper whisper-ctranslate2")
			os.Exit(1)
		}
	} else {
		app.logDebug("FasterWhisper is available")
	}
}