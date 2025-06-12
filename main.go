package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
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

type Config struct {
	WhisperModel        string `json:"whisper_model"`
	Language            string `json:"language"`
	ScanIntervalMinutes int    `json:"scan_interval_minutes"`
	MaxCpuPercent       int    `json:"max_cpu_percent"`
	ComputeType         string `json:"compute_type"`
	UseColors           bool   `json:"use_colors"`
	UIMode              string `json:"ui_mode"`
	OutputFormat        string `json:"output_format"`
}

type LogEntry struct {
	Level     string
	Message   string
	Timestamp time.Time
}

type App struct {
	config         *Config
	logger         *log.Logger
	debugMode      bool
	wg             sync.WaitGroup
	processedFiles map[string]bool
	mu             sync.Mutex

	// UI related fields
	startTime    time.Time
	lastScanTime time.Time
	logBuffer    []LogEntry
	logMutex     sync.RWMutex
	totalProcessed int
	inputCount   int
	outputCount  int
	archiveCount int

	// Queue management for sequential processing
	queuedFiles    []string // 処理待ちファイルキュー
	processingFile string   // 現在処理中のファイル名（表示用）
	isProcessing   bool     // 処理中フラグ
}

func main() {
	configPath, debugMode, showVersion, showHelp := parseFlags()

	if showVersion {
		fmt.Printf("KoeMoji-Go v%s\n", version)
		return
	}

	if showHelp {
		showHelpText()
		return
	}

	app := &App{
		debugMode:      debugMode,
		processedFiles: make(map[string]bool),
		startTime:      time.Now(),
		logBuffer:      make([]LogEntry, 0, 12),
		queuedFiles:    make([]string, 0),
	}

	app.initLogger()
	app.loadConfig(configPath)
	app.ensureDirectories()
	app.ensureDependencies()
	app.run()
}

func parseFlags() (string, bool, bool, bool) {
	configPath := flag.String("config", "config.json", "Path to config file")
	debugMode := flag.Bool("debug", false, "Enable debug mode")
	showVersion := flag.Bool("version", false, "Show version")
	showHelp := flag.Bool("help", false, "Show help")
	flag.Parse()
	return *configPath, *debugMode, *showVersion, *showHelp
}

func showHelpText() {
	fmt.Println("KoeMoji-Go - Audio/Video Transcription Tool")
	fmt.Printf("Version: %s\n\n", version)
	fmt.Println("Usage: koemoji-go [options]")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nInteractive commands:")
	fmt.Println("  c - Display configuration")
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
	app.logInfo("Starting KoeMoji-Go v%s", version)
}

func (app *App) loadConfig(configPath string) {
	// Default config
	app.config = &Config{
		WhisperModel:        "medium",
		Language:            "ja",
		ScanIntervalMinutes: 10,
		MaxCpuPercent:       95,
		ComputeType:         "int8",
		UseColors:           true,
		UIMode:              "enhanced",
		OutputFormat:        "txt",
	}

	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			app.logInfo("Config file not found, using defaults")
			return
		}
		app.logError("Failed to load config: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(app.config); err != nil {
		app.logError("Failed to parse config: %v", err)
		os.Exit(1)
	}
}

func (app *App) ensureDirectories() {
	dirs := []string{"input", "output", "archive"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			app.logError("Failed to create directory %s: %v", dir, err)
			os.Exit(1)
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

func (app *App) run() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go app.startProcessing()
	go app.handleUserInput()

	app.logInfo("KoeMoji-Go is running. Use commands below to interact.")
	app.logInfo("Monitoring ./input/ directory every %d minutes", app.config.ScanIntervalMinutes)

	<-sigChan
	app.logInfo("Shutting down KoeMoji-Go...")
	app.wg.Wait()
}

func (app *App) startProcessing() {
	// Initial scan
	app.scanAndProcess()

	// Periodic scan
	ticker := time.NewTicker(time.Duration(app.config.ScanIntervalMinutes) * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		app.scanAndProcess()
	}
}

func (app *App) scanAndProcess() {
	app.lastScanTime = time.Now()
	app.logInfo("Scanning for new files...")

	files, err := filepath.Glob("input/*")
	if err != nil {
		app.logError("Failed to scan input directory: %v", err)
		return
	}

	newFiles := app.filterNewAudioFiles(files)
	if len(newFiles) == 0 {
		app.logDebug("No new files found")
		return
	}

	app.logInfo("Found %d new file(s) to process", len(newFiles))

	// Add files to queue
	app.mu.Lock()
	app.queuedFiles = append(app.queuedFiles, newFiles...)
	app.mu.Unlock()

	// Start processing if not already processing
	if !app.isProcessing {
		app.wg.Add(1)
		go app.processQueue()
	}
}

func (app *App) filterNewAudioFiles(files []string) []string {
	app.mu.Lock()
	defer app.mu.Unlock()

	var newFiles []string
	for _, file := range files {
		if app.isAudioFile(file) && !app.processedFiles[file] {
			app.processedFiles[file] = true
			newFiles = append(newFiles, file)
		}
	}
	return newFiles
}

func (app *App) isAudioFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	audioExts := []string{".mp3", ".wav", ".m4a", ".flac", ".ogg", ".aac", ".mp4", ".mov", ".avi"}
	for _, audioExt := range audioExts {
		if ext == audioExt {
			return true
		}
	}
	return false
}

func (app *App) processQueue() {
	defer app.wg.Done()

	for {
		app.mu.Lock()
		if len(app.queuedFiles) == 0 {
			app.isProcessing = false
			app.processingFile = ""
			app.mu.Unlock()
			return
		}

		// Get next file from queue
		filePath := app.queuedFiles[0]
		app.queuedFiles = app.queuedFiles[1:]
		app.processingFile = filepath.Base(filePath)
		app.isProcessing = true
		app.mu.Unlock()

		// Process the file
		app.logProc("Processing: %s", app.processingFile)
		startTime := time.Now()

		if err := app.transcribeAudio(filePath); err != nil {
			app.logError("Failed to process %s: %v", app.processingFile, err)
		} else {
			duration := time.Since(startTime)
			app.logDone("Completed: %s (%s)", app.processingFile, app.formatDuration(duration))
			app.totalProcessed++

			// Move to archive
			if err := app.moveToArchive(filePath); err != nil {
				app.logError("Failed to archive %s: %v", app.processingFile, err)
			}
		}
	}
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

func (app *App) moveToArchive(sourcePath string) error {
	filename := filepath.Base(sourcePath)
	destPath := filepath.Join("archive", filename)

	// Handle duplicate filenames
	if _, err := os.Stat(destPath); err == nil {
		timestamp := time.Now().Format("20060102_150405")
		ext := filepath.Ext(filename)
		name := strings.TrimSuffix(filename, ext)
		destPath = filepath.Join("archive", fmt.Sprintf("%s_%s%s", name, timestamp, ext))
	}

	if err := os.Rename(sourcePath, destPath); err != nil {
		return err
	}
	return nil
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
			app.displayConfig()
			fmt.Print("Press Enter to continue...")
			bufio.NewReader(os.Stdin).ReadString('\n')
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

// Logging functions
func (app *App) addToLogBuffer(level, message string) {
	app.logMutex.Lock()
	defer app.logMutex.Unlock()

	entry := LogEntry{
		Level:     level,
		Message:   message,
		Timestamp: time.Now(),
	}

	app.logBuffer = append(app.logBuffer, entry)
	if len(app.logBuffer) > 12 {
		app.logBuffer = app.logBuffer[1:]
	}
}

func (app *App) logInfo(format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	app.logger.Printf("[INFO] %s", message)
	app.addToLogBuffer("INFO", message)
	app.refreshDisplay()
}

func (app *App) logError(format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	app.logger.Printf("[ERROR] %s", message)
	app.addToLogBuffer("ERROR", message)
	app.refreshDisplay()
}

func (app *App) logDebug(format string, v ...any) {
	if app.debugMode {
		message := fmt.Sprintf(format, v...)
		app.logger.Printf("[DEBUG] %s", message)
		app.addToLogBuffer("DEBUG", message)
		app.refreshDisplay()
	}
}

func (app *App) logProc(format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	app.logger.Printf("[PROC] %s", message)
	app.addToLogBuffer("PROC", message)
	app.refreshDisplay()
}

func (app *App) logDone(format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	app.logger.Printf("[DONE] %s", message)
	app.addToLogBuffer("DONE", message)
	app.refreshDisplay()
}

// UI functions
func (app *App) refreshDisplay() {
	if app.config == nil || app.config.UIMode != "enhanced" {
		return
	}

	// Clear screen and move cursor to top
	fmt.Print("\033[2J\033[H")

	app.displayHeader()
	app.displayRealtimeLogs()
	app.displayCommands()
}

func (app *App) displayHeader() {
	app.updateFileCounts()

	status := "🟢 Active"
	if app.isProcessing {
		status = "🟡 Processing"
	}

	uptime := time.Since(app.startTime)

	fmt.Println("=== KoeMoji-Go v" + version + " ===")
	
	app.mu.Lock()
	queueCount := len(app.queuedFiles)
	processingDisplay := "None"
	if app.processingFile != "" {
		processingDisplay = app.processingFile
	}
	app.mu.Unlock()

	fmt.Printf("%s | Queue: %d | Processing: %s\n",
		status, queueCount, processingDisplay)
	fmt.Printf("📁 Input: %d → Output: %d → Archive: %d\n",
		app.inputCount, app.outputCount, app.archiveCount)

	lastScanStr := "Never"
	nextScanStr := "Soon"
	if !app.lastScanTime.IsZero() {
		lastScanStr = app.lastScanTime.Format("15:04:05")
		nextScan := app.lastScanTime.Add(time.Duration(app.config.ScanIntervalMinutes) * time.Minute)
		nextScanStr = nextScan.Format("15:04:05")
	}

	fmt.Printf("⏰ Last: %s | Next: %s | Uptime: %s\n",
		lastScanStr, nextScanStr, app.formatDuration(uptime))
	fmt.Println()
}

func (app *App) displayRealtimeLogs() {
	app.logMutex.RLock()
	defer app.logMutex.RUnlock()

	for _, entry := range app.logBuffer {
		color := app.getLogColor(entry.Level)
		timestamp := entry.Timestamp.Format("15:04:05")

		if color != "" {
			fmt.Printf("%s%-5s%s %s %s\n", color, entry.Level, ColorReset, timestamp, entry.Message)
		} else {
			fmt.Printf("[%-5s] %s %s\n", entry.Level, timestamp, entry.Message)
		}
	}

	// Fill remaining lines to maintain 12-line display
	for i := len(app.logBuffer); i < 12; i++ {
		fmt.Println()
	}
}

func (app *App) displayCommands() {
	fmt.Println("c=config l=logs s=scan q=quit")
	fmt.Print("> ")
}

func (app *App) displayConfig() {
	fmt.Println("\n--- Configuration (config.json) ---")
	fmt.Printf("Whisper model: %s\n", app.config.WhisperModel)
	fmt.Printf("Language: %s\n", app.config.Language)
	fmt.Printf("Scan interval: %d minutes\n", app.config.ScanIntervalMinutes)
	fmt.Printf("Max CPU percent: %d%%\n", app.config.MaxCpuPercent)
	fmt.Printf("Compute type: %s\n", app.config.ComputeType)
	fmt.Printf("Use colors: %t\n", app.config.UseColors)
	fmt.Printf("UI mode: %s\n", app.config.UIMode)
	fmt.Printf("Output format: %s\n", app.config.OutputFormat)
	fmt.Println("\nDirectories:")
	fmt.Println("  Input: ./input/")
	fmt.Println("  Output: ./output/")
	fmt.Println("  Archive: ./archive/")
	fmt.Println("---")
}

func (app *App) displayLogs() {
	file, err := os.Open("koemoji.log")
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
	defer file.Close()

	fmt.Println("\n--- Log File Contents ---")
	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		lineCount++
		if lineCount%20 == 0 {
			fmt.Print("Press Enter to continue...")
			bufio.NewReader(os.Stdin).ReadString('\n')
		}
	}
	fmt.Println("--- End of Log ---")
}

// Utility functions
func (app *App) supportsColor() bool {
	if app.config == nil || !app.config.UseColors {
		return false
	}

	if runtime.GOOS == "windows" {
		return os.Getenv("WT_SESSION") != "" || os.Getenv("ConEmuPID") != ""
	}

	term := os.Getenv("TERM")
	return term != "" && term != "dumb"
}

func (app *App) getLogColor(level string) string {
	if !app.supportsColor() {
		return ""
	}

	switch level {
	case "INFO":
		return ColorBlue
	case "PROC":
		return ColorYellow
	case "DONE":
		return ColorGreen
	case "ERROR":
		return ColorRed
	case "DEBUG":
		return ColorGray
	default:
		return ""
	}
}

func (app *App) updateFileCounts() {
	app.inputCount = app.countFiles("input")
	app.outputCount = app.countFiles("output")
	app.archiveCount = app.countFiles("archive")
}

func (app *App) countFiles(dir string) int {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return 0
	}

	count := 0
	for _, file := range files {
		if info, err := os.Stat(file); err == nil && !info.IsDir() {
			count++
		}
	}
	return count
}

func (app *App) formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	if minutes >= 60 {
		hours := minutes / 60
		minutes = minutes % 60
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	return fmt.Sprintf("%dm%ds", minutes, seconds)
}