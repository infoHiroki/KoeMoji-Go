package gui

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"

	"github.com/hirokitakamura/koemoji-go/internal/config"
	"github.com/hirokitakamura/koemoji-go/internal/logger"
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
	statusLabel *widget.Label
	filesLabel  *widget.Label
	timingLabel *widget.Label
	logText     *widget.RichText
}

// Run starts the GUI application
func Run(configPath string, debugMode bool) {
	guiApp := &GUIApp{
		configPath:     configPath,
		debugMode:      debugMode,
		processedFiles: make(map[string]bool),
		startTime:      time.Now(),
		logBuffer:      make([]logger.LogEntry, 0, 12),
		queuedFiles:    make([]string, 0),
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
