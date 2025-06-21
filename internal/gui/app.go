package gui

import (
	"context"
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
	recordingDeviceSelect *widget.Select
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
	cfg := config.LoadConfig(app.configPath, nil) // GUI mode doesn't need file logger
	app.Config = cfg
}

// initLogger initializes the logger (simplified for GUI mode)
func (app *GUIApp) initLogger() {
	// For GUI mode, we'll use in-memory logging only
	// The log buffer will be displayed in the GUI
	logger.LogInfo(nil, &app.logBuffer, &app.logMutex, "KoeMoji-Go v1.3.0 started")
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
	logger.LogInfo(nil, &app.logBuffer, &app.logMutex, "Application resources cleaned up")
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
