package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hirokitakamura/koemoji-go/internal/config"
	"github.com/hirokitakamura/koemoji-go/internal/gui"
	"github.com/hirokitakamura/koemoji-go/internal/logger"
	"github.com/hirokitakamura/koemoji-go/internal/processor"
	"github.com/hirokitakamura/koemoji-go/internal/recorder"
	"github.com/hirokitakamura/koemoji-go/internal/ui"
	"github.com/hirokitakamura/koemoji-go/internal/whisper"
)

// version は暫定的に定数として定義（ビルド時に -X フラグで上書き）
var version = "dev"

// Color constants
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m" // ERROR
	ColorGreen  = "\033[32m" // DONE
	ColorYellow = "\033[33m" // PROC
	ColorBlue   = "\033[34m" // INFO
	ColorGray   = "\033[37m" // DEBUG
)

type App struct {
	*config.Config
	configPath     string
	logger         *log.Logger
	debugMode      bool
	wg             sync.WaitGroup
	processedFiles map[string]bool
	mu             sync.Mutex

	// UI related fields
	startTime    time.Time
	lastScanTime time.Time
	logBuffer    []logger.LogEntry
	logMutex     sync.RWMutex
	inputCount   int
	outputCount  int
	archiveCount int

	// Queue management for sequential processing
	queuedFiles    []string // 処理待ちファイルキュー
	processingFile string   // 現在処理中のファイル名（表示用）
	isProcessing   bool     // 処理中フラグ

	// Recording related fields
	recorder           *recorder.Recorder
	isRecording        bool
	recordingStartTime time.Time
}

func main() {
	configPath, debugMode, showVersion, showHelp, configMode, tuiMode := parseFlags()

	if showVersion {
		fmt.Printf("KoeMoji-Go v%s\n", version)
		return
	}

	if showHelp {
		showHelpText()
		return
	}

	// Handle TUI mode (non-default)
	if tuiMode {
		runTUIMode(configPath, debugMode, configMode)
		return
	}

	// Default: GUI mode
	gui.Run(configPath, debugMode)
}

func runTUIMode(configPath string, debugMode bool, configMode bool) {
	app := &App{
		configPath:     configPath,
		debugMode:      debugMode,
		processedFiles: make(map[string]bool),
		startTime:      time.Now(),
		logBuffer:      make([]logger.LogEntry, 0, 12),
		queuedFiles:    make([]string, 0),
	}

	app.initLogger()
	cfg, err := config.LoadConfig(configPath, app.logger)
	if err != nil {
		logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "Failed to load config: %v", err)
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config: %v\n", err)
		fmt.Fprintf(os.Stderr, "Using default configuration.\n")
		cfg = config.GetDefaultConfig()
	}
	app.Config = cfg

	if configMode {
		config.ConfigureSettings(app.Config, configPath, app.logger)
		return
	}

	if err := processor.EnsureDirectories(app.Config, app.logger); err != nil {
		logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "Failed to create directories: %v", err)
		fmt.Fprintf(os.Stderr, "Warning: Failed to create directories: %v\n", err)
		fmt.Fprintf(os.Stderr, "The application will continue with limited functionality.\n")
	}

	if err := whisper.EnsureDependencies(app.Config, app.logger, &app.logBuffer, &app.logMutex, app.debugMode); err != nil {
		logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "FasterWhisper dependency check failed: %v", err)
		fmt.Fprintf(os.Stderr, "Warning: FasterWhisper is not available: %v\n", err)
		fmt.Fprintf(os.Stderr, "Please install manually: pip install faster-whisper whisper-ctranslate2\n")
		fmt.Fprintf(os.Stderr, "The application will continue with limited functionality.\n")
	}
	app.run()
}

func parseFlags() (string, bool, bool, bool, bool, bool) {
	configPath := flag.String("config", "config.json", "Path to config file")
	debugMode := flag.Bool("debug", false, "Enable debug mode")
	showVersion := flag.Bool("version", false, "Show version")
	showHelp := flag.Bool("help", false, "Show help")
	configMode := flag.Bool("configure", false, "Enter configuration mode")
	tuiMode := flag.Bool("tui", false, "Run in Terminal UI (TUI) mode")
	flag.Parse()
	return *configPath, *debugMode, *showVersion, *showHelp, *configMode, *tuiMode
}

func showHelpText() {
	fmt.Println("KoeMoji-Go - Audio/Video Transcription Tool")
	fmt.Printf("Version: %s\n\n", version)
	fmt.Println("Usage: koemoji-go [options]")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nModes:")
	fmt.Println("  Default - Graphical User Interface (GUI) mode")
	fmt.Println("  --tui   - Terminal UI (TUI) mode")
	fmt.Println("\nInteractive commands (TUI mode only):")
	fmt.Println("  c - Configure settings")
	fmt.Println("  l - Display all logs")
	fmt.Println("  s - Scan now")
	fmt.Println("  i - Open input directory")
	fmt.Println("  o - Open output directory")
	fmt.Println("  r - Start/stop recording")
	fmt.Println("  q - Quit")
	fmt.Println("  Enter - Refresh display")
}

func (app *App) initLogger() {
	logFile, err := os.OpenFile("koemoji.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// Only timestamps for file logging, no prefix for console
	app.logger = log.New(io.MultiWriter(logFile), "", log.LstdFlags)
	logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "KoeMoji-Go started")
}

func (app *App) run() {
	// Phase 2: Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	go processor.StartProcessing(ctx, app.Config, app.logger, &app.logBuffer, &app.logMutex,
		&app.lastScanTime, &app.queuedFiles, &app.processingFile, &app.isProcessing,
		&app.processedFiles, &app.mu, &app.wg, app.debugMode)
	go app.handleUserInput(ctx)

	logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "KoeMoji-Go is running. Use commands below to interact.")
	logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "Monitoring %s directory every %d minutes", app.Config.InputDir, app.Config.ScanIntervalMinutes)

	// Display initial dashboard
	// Brief wait for initialization
	time.Sleep(100 * time.Millisecond)
	ui.RefreshDisplay(app.Config, app.startTime, app.lastScanTime, &app.logBuffer,
		&app.logMutex, app.inputCount, app.outputCount, app.archiveCount,
		&app.queuedFiles, app.processingFile, app.isProcessing, &app.mu,
		app.isRecording, app.recordingStartTime)

	<-sigChan
	logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "Graceful shutdown initiated...")
	cancel() // Signal all goroutines to stop

	// Wait for graceful shutdown with timeout
	done := make(chan bool, 1)
	go func() {
		app.wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "Shutdown completed successfully")
	case <-time.After(10 * time.Second):
		logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "Shutdown timeout, forcing exit")
	}
}

func (app *App) handleUserInput(ctx context.Context) {
	reader := bufio.NewReader(os.Stdin)
	for {
		select {
		case <-ctx.Done():
			logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "User input handler stopped")
			return
		default:
			// Non-blocking input check would be complex, so we keep the blocking read
			// The context cancellation will be handled when the process exits
		}
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return
			}
			continue
		}

		switch strings.TrimSpace(strings.ToLower(input)) {
		case "":
			// Empty Enter = manual refresh
			ui.RefreshDisplay(app.Config, app.startTime, app.lastScanTime, &app.logBuffer,
				&app.logMutex, app.inputCount, app.outputCount, app.archiveCount,
				&app.queuedFiles, app.processingFile, app.isProcessing, &app.mu,
				app.isRecording, app.recordingStartTime)
		case "c":
			config.ConfigureSettings(app.Config, app.configPath, app.logger)
			ui.RefreshDisplay(app.Config, app.startTime, app.lastScanTime, &app.logBuffer,
				&app.logMutex, app.inputCount, app.outputCount, app.archiveCount,
				&app.queuedFiles, app.processingFile, app.isProcessing, &app.mu,
				app.isRecording, app.recordingStartTime)
		case "l":
			ui.DisplayLogs(app.Config)
			fmt.Print("Press Enter to continue...")
			bufio.NewReader(os.Stdin).ReadString('\n')
			ui.RefreshDisplay(app.Config, app.startTime, app.lastScanTime, &app.logBuffer,
				&app.logMutex, app.inputCount, app.outputCount, app.archiveCount,
				&app.queuedFiles, app.processingFile, app.isProcessing, &app.mu,
				app.isRecording, app.recordingStartTime)
		case "s":
			logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "Manual scan triggered")
			go processor.ScanAndProcess(app.Config, app.logger, &app.logBuffer, &app.logMutex,
				&app.lastScanTime, &app.queuedFiles, &app.processingFile, &app.isProcessing,
				&app.processedFiles, &app.mu, &app.wg, app.debugMode)
		case "i":
			if err := ui.OpenDirectory(app.Config.InputDir); err != nil {
				logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "Failed to open input directory: %v", err)
			}
		case "o":
			if err := ui.OpenDirectory(app.Config.OutputDir); err != nil {
				logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "Failed to open output directory: %v", err)
			}
		case "r":
			app.handleRecordingToggle()
		case "q":
			logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "Shutting down KoeMoji-Go...")
			os.Exit(0)
		default:
			if strings.TrimSpace(input) != "" {
				fmt.Printf("Invalid command '%s' (use c/l/s/i/o/a/r/q or Enter to refresh)\n", strings.TrimSpace(input))
			}
		}
	}
}

func (app *App) handleRecordingToggle() {
	if app.isRecording {
		// Stop recording
		app.stopRecording()
	} else {
		// Start recording
		app.startRecording()
	}
}

func (app *App) startRecording() {
	// Initialize recorder if not already done
	if app.recorder == nil {
		var err error
		// Use device name if specified, otherwise use default device
		if app.Config.RecordingDeviceName != "" {
			app.recorder, err = recorder.NewRecorderWithDeviceName(app.Config.RecordingDeviceName)
		} else {
			app.recorder, err = recorder.NewRecorder()
		}

		if err != nil {
			logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "録音の初期化に失敗: %v", err)
			return
		}

		// Phase 1: Set recording limits
		var maxDuration time.Duration
		var maxFileSize int64

		if app.Config.RecordingMaxHours > 0 {
			maxDuration = time.Duration(app.Config.RecordingMaxHours) * time.Hour
		}

		if app.Config.RecordingMaxFileMB > 0 {
			maxFileSize = int64(app.Config.RecordingMaxFileMB) * 1024 * 1024 // Convert MB to bytes
		}

		app.recorder.SetLimits(maxDuration, maxFileSize)
	}

	// Start recording
	err := app.recorder.Start()
	if err != nil {
		logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "録音の開始に失敗: %v", err)
		return
	}

	app.isRecording = true
	app.recordingStartTime = time.Now()
	logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "録音を開始しました")

	// Refresh display to show recording status
	ui.RefreshDisplay(app.Config, app.startTime, app.lastScanTime, &app.logBuffer,
		&app.logMutex, app.inputCount, app.outputCount, app.archiveCount,
		&app.queuedFiles, app.processingFile, app.isProcessing, &app.mu,
		app.isRecording, app.recordingStartTime)
}

func (app *App) stopRecording() {
	if app.recorder == nil {
		logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "録音が初期化されていません")
		return
	}

	// Stop recording
	err := app.recorder.Stop()
	if err != nil {
		logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "録音の停止に失敗: %v", err)
		return
	}

	// Generate filename with current timestamp
	now := time.Now()
	filename := fmt.Sprintf("recording_%s.wav", now.Format("20060102_1504"))

	// Save to input directory with normalization
	outputPath := filepath.Join(app.Config.InputDir, filename)
	err = app.recorder.SaveToFileWithNormalization(outputPath, app.Config.AudioNormalizationEnabled)
	if err != nil {
		logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "録音ファイルの保存に失敗: %v", err)
		return
	}

	app.isRecording = false
	duration := time.Since(app.recordingStartTime)
	logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "録音を停止しました: %s (時間: %s)", filename, duration.Round(time.Second))

	// Refresh display to remove recording status
	ui.RefreshDisplay(app.Config, app.startTime, app.lastScanTime, &app.logBuffer,
		&app.logMutex, app.inputCount, app.outputCount, app.archiveCount,
		&app.queuedFiles, app.processingFile, app.isProcessing, &app.mu,
		app.isRecording, app.recordingStartTime)
}
