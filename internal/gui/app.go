package gui

import (
	"context"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"

	"github.com/hirokitakamura/koemoji-go/internal/config"
	"github.com/hirokitakamura/koemoji-go/internal/logger"
	"github.com/hirokitakamura/koemoji-go/internal/recorder"
)

// GUIApp represents the GUI application
type GUIApp struct {
	// Core application fields (reused from main.go App struct)
	*config.Config
	configPath     string
	debugMode      bool
	processedFiles map[string]bool
	mu             sync.Mutex
	logger         *log.Logger

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

	// GUI specific fields
	fyneApp fyne.App
	window  fyne.Window

	// UI components (will be implemented in components.go)
	statusWidget fyne.CanvasObject
	logWidget    fyne.CanvasObject
	buttonWidget fyne.CanvasObject

	// UI component references for updates
	statusLabel  *widget.Label
	filesLabel   *widget.Label
	timingLabel  *widget.Label
	logText      *widget.RichText
	recordButton *widget.Button

	// Recording related fields
	recorder              *recorder.Recorder
	recordingDeviceSelect *widget.SelectEntry
	recordingDeviceMap    map[string]int

	// UI safety fields
	uiInitialized bool

	// Phase 2: Context cancellation for goroutines
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// Run starts the GUI application
func Run(configPath string, debugMode bool) {
	// Phase 2: Create context for goroutine management
	ctx, cancel := context.WithCancel(context.Background())

	guiApp := &GUIApp{
		configPath:     configPath,
		debugMode:      debugMode,
		processedFiles: make(map[string]bool),
		startTime:      time.Now(),
		logBuffer:      make([]logger.LogEntry, 0, 12),
		queuedFiles:    make([]string, 0),
		ctx:            ctx,
		cancelFunc:     cancel,
	}

	// Initialize the application
	guiApp.fyneApp = app.NewWithID("com.hirokitakamura.koemoji-go")

	// Load configuration
	guiApp.loadConfig()

	// Create and show the main window
	guiApp.createWindow()
	guiApp.window.ShowAndRun()
}

// loadConfig loads the application configuration
func (app *GUIApp) loadConfig() {
	// Initialize logger first (similar to main.go)
	app.initLogger()

	// Load configuration
	cfg, err := config.LoadConfig(app.configPath, app.logger) // Use logger for consistent behavior
	if err != nil {
		logger.LogError(app.logger, &app.logBuffer, &app.logMutex, "Failed to load config: %v", err)
		// In GUI mode, use default config and show error dialog later
		app.Config = config.GetDefaultConfigResolved()
		logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "Using default configuration due to config load error")
		
		// Show error dialog when window is available
		go func() {
			// Wait for UI to be ready before showing dialog
			for app.window == nil {
				time.Sleep(100 * time.Millisecond)
			}
			app.showConfigErrorDialog(err)
		}()
		return
	}
	app.Config = cfg
}

// initLogger initializes the logger (consistent with TUI mode)
func (app *GUIApp) initLogger() {
	logFile, err := os.OpenFile("koemoji.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// Only timestamps for file logging, no prefix for console
	app.logger = log.New(io.MultiWriter(logFile), "", log.LstdFlags)
	logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "KoeMoji-Go started (GUI mode)")
}

// ForceCleanup performs immediate resource cleanup for application exit
func (app *GUIApp) ForceCleanup() {
	// Phase 2: Cancel all goroutines
	if app.cancelFunc != nil {
		app.cancelFunc()
	}

	// Clean up recorder resources (PortAudio)
	if app.recorder != nil {
		app.recorder.Close()
		app.recorder = nil
	}

	// Log cleanup action
	logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "Application resources cleaned up")
}

// isUIReady checks if all essential UI components are initialized
func (app *GUIApp) isUIReady() bool {
	return app.uiInitialized &&
		app.statusLabel != nil &&
		app.filesLabel != nil &&
		app.timingLabel != nil &&
		app.logText != nil &&
		app.recordButton != nil
}
