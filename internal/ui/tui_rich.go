package ui

import (
	"fmt"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/infoHiroki/KoeMoji-Go/internal/config"
	"github.com/infoHiroki/KoeMoji-Go/internal/logger"
	"github.com/rivo/tview"
)

// RichTUI represents the tview-based Terminal UI
type RichTUI struct {
	app    *tview.Application
	config *config.Config
	pages  *tview.Pages

	// Widgets
	statusView  *tview.TextView
	logView     *tview.TextView
	commandView *tview.TextView

	// State
	startTime    time.Time
	lastScanTime time.Time
	logBuffer    *[]logger.LogEntry
	logMutex     *sync.RWMutex

	// Callbacks
	onScan   func()
	onRecord func()
	onConfig func()
	onLogs   func()
	onInput  func()
	onOutput func()
	onQuit   func()

	// Status tracking
	inputCount     int
	outputCount    int
	archiveCount   int
	queuedFiles    *[]string
	processingFile string
	isProcessing   bool
	mu             *sync.Mutex
	isRecording    bool
	recordingStart time.Time
}

// NewRichTUI creates a new rich terminal UI
func NewRichTUI(cfg *config.Config, logBuffer *[]logger.LogEntry, logMutex *sync.RWMutex,
	queuedFiles *[]string, mu *sync.Mutex) *RichTUI {

	tui := &RichTUI{
		app:          tview.NewApplication(),
		config:       cfg,
		pages:        tview.NewPages(),
		startTime:    time.Now(),
		logBuffer:    logBuffer,
		logMutex:     logMutex,
		queuedFiles:  queuedFiles,
		mu:           mu,
	}

	tui.createUI()
	return tui
}

func (t *RichTUI) createUI() {
	msg := GetMessages(t.config)

	// Create status panel (top) - 5 lines fixed
	t.statusView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	t.statusView.SetBorder(true).
		SetTitle(" KoeMoji-Go v1.9.0-alpha ").
		SetBorderPadding(0, 0, 1, 1)

	// Create log viewer (center) - scrollable
	t.logView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			t.app.Draw()
		})
	t.logView.SetBorder(true).
		SetTitle(" " + msg.LogTitle + " [↑↓でスクロール] ").
		SetBorderPadding(0, 0, 1, 1)

	// Create command panel (bottom) - 3 lines fixed
	t.commandView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	t.commandView.SetBorder(true).
		SetTitle(" " + msg.ConfigCmd + " ").
		SetBorderPadding(0, 0, 1, 1)
	t.updateCommandView()

	// Create main layout (3 rows)
	mainLayout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(t.statusView, 7, 0, false).     // Fixed height 7 lines
		AddItem(t.logView, 0, 1, true).         // Flexible (main area)
		AddItem(t.commandView, 3, 0, false)     // Fixed height 3 lines

	// Add to pages
	t.pages.AddPage("main", mainLayout, true, true)

	// Set up key bindings
	t.setupKeyBindings()

	// Set root
	t.app.SetRoot(t.pages, true).SetFocus(t.logView)
}

func (t *RichTUI) setupKeyBindings() {
	t.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF1:
			if t.onConfig != nil {
				t.onConfig()
			}
			return nil
		case tcell.KeyF2:
			if t.onLogs != nil {
				t.onLogs()
			}
			return nil
		case tcell.KeyF3:
			if t.onScan != nil {
				t.onScan()
			}
			return nil
		case tcell.KeyF4:
			if t.onRecord != nil {
				t.onRecord()
			}
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q', 'Q':
				if t.onQuit != nil {
					t.onQuit()
				}
				t.app.Stop()
				return nil
			case 'c', 'C':
				if t.onConfig != nil {
					t.onConfig()
				}
				return nil
			case 'l', 'L':
				if t.onLogs != nil {
					t.onLogs()
				}
				return nil
			case 's', 'S':
				if t.onScan != nil {
					t.onScan()
				}
				return nil
			case 'r', 'R':
				if t.onRecord != nil {
					t.onRecord()
				}
				return nil
			case 'i', 'I':
				if t.onInput != nil {
					t.onInput()
				}
				return nil
			case 'o', 'O':
				if t.onOutput != nil {
					t.onOutput()
				}
				return nil
			}
		}
		return event
	})
}

// UpdateStatus updates the status panel
func (t *RichTUI) UpdateStatus(inputCount, outputCount, archiveCount int, processingFile string, isProcessing bool, isRecording bool, recordingStart time.Time) {
	t.inputCount = inputCount
	t.outputCount = outputCount
	t.archiveCount = archiveCount
	t.processingFile = processingFile
	t.isProcessing = isProcessing
	t.isRecording = isRecording
	t.recordingStart = recordingStart

	t.app.QueueUpdateDraw(func() {
		msg := GetMessages(t.config)

		// Status line
		statusIcon := "🟢"
		statusText := msg.Active
		if isProcessing {
			statusIcon = "🟡"
			statusText = msg.Processing
		}

		t.mu.Lock()
		queueCount := len(*t.queuedFiles)
		processingDisplay := msg.None
		if processingFile != "" {
			processingDisplay = processingFile
		}
		t.mu.Unlock()

		// Calculate timing
		uptime := time.Since(t.startTime)
		lastScanStr := msg.Never
		nextScanStr := msg.Soon
		if !t.lastScanTime.IsZero() {
			lastScanStr = t.lastScanTime.Format("15:04:05")
			nextScan := t.lastScanTime.Add(time.Duration(t.config.ScanIntervalMinutes) * time.Minute)
			nextScanStr = nextScan.Format("15:04:05")
		}

		statusContent := fmt.Sprintf(
			"%s [yellow]%s[white] | %s: %d | %s: %s\n"+
				"📁 %s: %d → %s: %d → %s: %d\n"+
				"⏰ %s: %s | %s: %s | %s: %s",
			statusIcon, statusText, msg.Queue, queueCount, msg.Processing, processingDisplay,
			msg.Input, inputCount, msg.Output, outputCount, msg.Archive, archiveCount,
			msg.Last, lastScanStr, msg.Next, nextScanStr, msg.Uptime, formatDuration(uptime),
		)

		if isRecording {
			elapsed := time.Since(recordingStart)
			statusContent += fmt.Sprintf("\n🔴 [red]%s[white] - %s", msg.Recording, formatDuration(elapsed))
		}

		t.statusView.SetText(statusContent)
	})
}

// UpdateLogs updates the log viewer
func (t *RichTUI) UpdateLogs() {
	t.app.QueueUpdateDraw(func() {
		t.logMutex.RLock()
		defer t.logMutex.RUnlock()

		msg := GetMessages(t.config)
		t.logView.Clear()

		if len(*t.logBuffer) == 0 {
			t.logView.SetText("[gray]" + msg.LogPlaceholder)
			return
		}

		for _, entry := range *t.logBuffer {
			timestamp := entry.Timestamp.Format("15:04:05")
			color := t.getLogColorCode(entry.Level)

			// Localize log level
			localizedLevel := entry.Level
			switch entry.Level {
			case "INFO":
				localizedLevel = msg.LogInfo
			case "PROC":
				localizedLevel = msg.LogProc
			case "DONE":
				localizedLevel = msg.LogDone
			case "ERROR":
				localizedLevel = msg.LogError
			case "DEBUG":
				localizedLevel = msg.LogDebug
			}

			// tview color tags
			fmt.Fprintf(t.logView, "[%s][%s][white] %s %s\n",
				color, localizedLevel, timestamp, entry.Message)
		}

		// Auto-scroll to bottom
		t.logView.ScrollToEnd()
	})
}

func (t *RichTUI) getLogColorCode(level string) string {
	switch level {
	case "INFO":
		return "blue"
	case "PROC":
		return "yellow"
	case "DONE":
		return "green"
	case "ERROR":
		return "red"
	case "DEBUG":
		return "gray"
	default:
		return "white"
	}
}

func (t *RichTUI) updateCommandView() {
	msg := GetMessages(t.config)
	commands := fmt.Sprintf(
		"[yellow]F1[white]/c=%s  [yellow]F2[white]/l=%s  [yellow]F3[white]/s=%s  [yellow]F4[white]/r=%s  i=%s  o=%s  [yellow]q[white]=%s",
		msg.ConfigCmd, msg.LogsCmd, msg.ScanCmd, msg.RecordCmd, msg.InputDirCmd, msg.OutputDirCmd, msg.QuitCmd,
	)
	t.commandView.SetText(commands)
}

// SetCallbacks sets the callback functions
func (t *RichTUI) SetCallbacks(onScan, onRecord, onConfig, onLogs, onInput, onOutput, onQuit func()) {
	t.onScan = onScan
	t.onRecord = onRecord
	t.onConfig = onConfig
	t.onLogs = onLogs
	t.onInput = onInput
	t.onOutput = onOutput
	t.onQuit = onQuit
}

// SetLastScanTime sets the last scan time
func (t *RichTUI) SetLastScanTime(lastScanTime time.Time) {
	t.lastScanTime = lastScanTime
}

// Run starts the TUI application
func (t *RichTUI) Run() error {
	return t.app.Run()
}

// Stop stops the TUI application
func (t *RichTUI) Stop() {
	t.app.Stop()
}
