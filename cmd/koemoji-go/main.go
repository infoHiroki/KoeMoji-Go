package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hirokitakamura/koemoji-go/internal/config"
	"github.com/hirokitakamura/koemoji-go/internal/logger"
	"github.com/hirokitakamura/koemoji-go/internal/processor"
	"github.com/hirokitakamura/koemoji-go/internal/ui"
	"github.com/hirokitakamura/koemoji-go/internal/whisper"
)

const version = "1.2.0"

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
	startTime      time.Time
	lastScanTime   time.Time
	logBuffer      []logger.LogEntry
	logMutex       sync.RWMutex
	inputCount     int
	outputCount    int
	archiveCount   int

	// Queue management for sequential processing
	queuedFiles    []string // 処理待ちファイルキュー
	processingFile string   // 現在処理中のファイル名（表示用）
	isProcessing   bool     // 処理中フラグ
}

func main() {
	configPath, debugMode, showVersion, showHelp, configMode := parseFlags()

	if showVersion {
		fmt.Printf("KoeMoji-Go v%s\n", version)
		return
	}

	if showHelp {
		showHelpText()
		return
	}

	app := &App{
		configPath:     configPath,
		debugMode:      debugMode,
		processedFiles: make(map[string]bool),
		startTime:      time.Now(),
		logBuffer:      make([]logger.LogEntry, 0, 12),
		queuedFiles:    make([]string, 0),
	}

	app.initLogger()
	cfg := config.LoadConfig(configPath, app.logger)
	app.Config = cfg

	if configMode {
		config.ConfigureSettings(app.Config, configPath, app.logger)
		return
	}

	processor.EnsureDirectories(app.Config, app.logger)
	whisper.EnsureDependencies(app.Config, app.logger, &app.logBuffer, &app.logMutex, app.debugMode)
	app.run()
}

func parseFlags() (string, bool, bool, bool, bool) {
	configPath := flag.String("config", "config.json", "Path to config file")
	debugMode := flag.Bool("debug", false, "Enable debug mode")
	showVersion := flag.Bool("version", false, "Show version")
	showHelp := flag.Bool("help", false, "Show help")
	configMode := flag.Bool("configure", false, "Enter configuration mode")
	flag.Parse()
	return *configPath, *debugMode, *showVersion, *showHelp, *configMode
}

func showHelpText() {
	fmt.Println("KoeMoji-Go - Audio/Video Transcription Tool")
	fmt.Printf("Version: %s\n\n", version)
	fmt.Println("Usage: koemoji-go [options]")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nInteractive commands:")
	fmt.Println("  c - Configure settings")
	fmt.Println("  l - Display all logs")
	fmt.Println("  s - Scan now")
	fmt.Println("  i - Open input directory")
	fmt.Println("  o - Open output directory")
	fmt.Println("  a - Toggle AI summary")
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
	logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "KoeMoji-Go v%s started", version)
}

func (app *App) run() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go processor.StartProcessing(app.Config, app.logger, &app.logBuffer, &app.logMutex, 
		&app.lastScanTime, &app.queuedFiles, &app.processingFile, &app.isProcessing, 
		&app.processedFiles, &app.mu, &app.wg, app.debugMode)
	go app.handleUserInput()

	logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "KoeMoji-Go is running. Use commands below to interact.")
	logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "Monitoring %s directory every %d minutes", app.Config.InputDir, app.Config.ScanIntervalMinutes)
	
	// Display initial dashboard
	// Brief wait for initialization
	time.Sleep(100 * time.Millisecond)
	ui.RefreshDisplay(app.Config, app.startTime, app.lastScanTime, &app.logBuffer, 
		&app.logMutex, app.inputCount, app.outputCount, app.archiveCount,
		&app.queuedFiles, app.processingFile, app.isProcessing, &app.mu)

	<-sigChan
	logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "Shutting down KoeMoji-Go...")
	app.wg.Wait()
}

func (app *App) handleUserInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
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
				&app.queuedFiles, app.processingFile, app.isProcessing, &app.mu)
		case "c":
			config.ConfigureSettings(app.Config, app.configPath, app.logger)
			ui.RefreshDisplay(app.Config, app.startTime, app.lastScanTime, &app.logBuffer, 
				&app.logMutex, app.inputCount, app.outputCount, app.archiveCount,
				&app.queuedFiles, app.processingFile, app.isProcessing, &app.mu)
		case "l":
			ui.DisplayLogs(app.Config)
			fmt.Print("Press Enter to continue...")
			bufio.NewReader(os.Stdin).ReadString('\n')
			ui.RefreshDisplay(app.Config, app.startTime, app.lastScanTime, &app.logBuffer, 
				&app.logMutex, app.inputCount, app.outputCount, app.archiveCount,
				&app.queuedFiles, app.processingFile, app.isProcessing, &app.mu)
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
		case "a":
			// Toggle AI summary
			app.Config.LLMSummaryEnabled = !app.Config.LLMSummaryEnabled
			status := "disabled"
			if app.Config.LLMSummaryEnabled {
				status = "enabled"
			}
			logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "AI summary %s", status)
			// Save the configuration change
			if err := config.SaveConfig(app.Config, app.configPath); err != nil {
				logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "Failed to save config: %v", err)
			}
			ui.RefreshDisplay(app.Config, app.startTime, app.lastScanTime, &app.logBuffer, 
				&app.logMutex, app.inputCount, app.outputCount, app.archiveCount,
				&app.queuedFiles, app.processingFile, app.isProcessing, &app.mu)
		case "q":
			logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "Shutting down KoeMoji-Go...")
			os.Exit(0)
		default:
			if strings.TrimSpace(input) != "" {
				fmt.Printf("Invalid command '%s' (use c/l/s/i/o/a/q or Enter to refresh)\n", strings.TrimSpace(input))
			}
		}
	}
}