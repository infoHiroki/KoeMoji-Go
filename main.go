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
)

const version = "1.0.0"

// Color constants
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m" // ERROR
	ColorGreen  = "\033[32m" // DONE
	ColorYellow = "\033[33m" // PROC
	ColorBlue   = "\033[34m" // INFO
	ColorGray   = "\033[37m" // DEBUG
)

type LogEntry struct {
	Level     string
	Message   string
	Timestamp time.Time
}

type App struct {
	config         *Config
	configPath     string
	logger         *log.Logger
	debugMode      bool
	wg             sync.WaitGroup
	processedFiles map[string]bool
	mu             sync.Mutex

	// UI related fields
	startTime      time.Time
	lastScanTime   time.Time
	logBuffer      []LogEntry
	logMutex     sync.RWMutex
	inputCount   int
	outputCount  int
	archiveCount int

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
		logBuffer:      make([]LogEntry, 0, 12),
		queuedFiles:    make([]string, 0),
	}

	app.initLogger()
	app.loadConfig(configPath)

	if configMode {
		app.configureSettings(app.configPath)
		return
	}

	app.ensureDirectories()
	app.ensureDependencies()
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
	msg := app.getMessages()
	app.logInfo(msg.AppStarted, version)
}

func (app *App) run() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go app.startProcessing()
	go app.handleUserInput()

	app.logInfo("KoeMoji-Go is running. Use commands below to interact.")
	app.logInfo("Monitoring %s directory every %d minutes", app.config.InputDir, app.config.ScanIntervalMinutes)

	<-sigChan
	msg := app.getMessages()
	app.logInfo(msg.ShuttingDown)
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
			if app.config.UIMode == "enhanced" {
				app.refreshDisplay()
			}
		case "c":
			app.configureSettings(app.configPath)
			if app.config.UIMode == "enhanced" {
				app.refreshDisplay()
			}
		case "l":
			app.displayLogs()
			fmt.Print("Press Enter to continue...")
			bufio.NewReader(os.Stdin).ReadString('\n')
			if app.config.UIMode == "enhanced" {
				app.refreshDisplay()
			}
		case "s":
			app.logInfo("Manual scan triggered")
			go app.scanAndProcess()
		case "q":
			app.logInfo("Shutting down KoeMoji-Go...")
			os.Exit(0)
		default:
			if strings.TrimSpace(input) != "" {
				fmt.Printf("Invalid command '%s' (use c/l/s/q or Enter to refresh)\n", strings.TrimSpace(input))
			}
		}
	}
}
