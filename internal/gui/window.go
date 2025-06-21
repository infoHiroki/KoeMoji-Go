package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/hirokitakamura/koemoji-go/internal/ui"
)

// createWindow creates and configures the main application window
func (app *GUIApp) createWindow() {
	// Create the main window
	app.window = app.fyneApp.NewWindow("KoeMoji-Go v1.3.0")
	app.window.Resize(fyne.NewSize(800, 700))
	app.window.CenterOnScreen()
	app.window.SetMaster()

	// Set window icon (will be implemented later)
	// app.window.SetIcon(resourceIconPng)

	// Create UI components
	app.createComponents()

	// Set up the main layout using BorderLayout
	content := container.NewBorder(
		app.statusWidget, // top
		app.buttonWidget, // bottom
		nil,              // left
		nil,              // right
		app.logWidget,    // center
	)

	app.window.SetContent(content)

	// Set up window close behavior (immediate exit as per design)
	app.window.SetCloseIntercept(func() {
		app.fyneApp.Quit()
	})

	// Periodic updates are started in app.go after window creation
}

// createComponents creates all UI components
func (app *GUIApp) createComponents() {
	// Get messages for the current language
	msg := ui.GetMessages(app.Config)

	// Create status panel (top)
	app.statusWidget = app.createStatusPanel(msg)

	// Create log viewer (center)
	app.logWidget = app.createLogViewer(msg)

	// Create button panel (bottom)
	app.buttonWidget = app.createButtonPanel(msg)
}

// createStatusPanel creates the status display panel
func (app *GUIApp) createStatusPanel(msg *ui.Messages) fyne.CanvasObject {
	// Status line 1: Active/Processing state and queue info
	statusLabel := widget.NewLabel("🟢 " + msg.Active + " | " + msg.Queue + ": 0 | " + msg.Processing + ": " + msg.None)
	statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Status line 2: File counts
	filesLabel := widget.NewLabel("📁 " + msg.Input + ": 0 → " + msg.Output + ": 0 → " + msg.Archive + ": 0")

	// Status line 3: Timing info
	timingLabel := widget.NewLabel("⏰ " + msg.Last + ": " + msg.Never + " | " + msg.Next + ": " + msg.Soon + " | " + msg.Uptime + ": 0s")

	// Store references for updates
	app.statusLabel = statusLabel
	app.filesLabel = filesLabel
	app.timingLabel = timingLabel

	// Create a card container for the status panel
	statusCard := widget.NewCard("", "", container.NewVBox(
		statusLabel,
		filesLabel,
		timingLabel,
	))

	return statusCard
}

// createLogViewer creates the scrollable log display
func (app *GUIApp) createLogViewer(msg *ui.Messages) fyne.CanvasObject {
	// Create a rich text widget for log display
	app.logText = widget.NewRichTextFromMarkdown("**Logs will appear here...**")
	app.logText.Wrapping = fyne.TextWrapWord

	// Create scrollable container
	logScroll := container.NewVScroll(app.logText)
	logScroll.SetMinSize(fyne.NewSize(750, 400))

	// Create a card container for the log viewer
	logCard := widget.NewCard("Recent Logs", "", logScroll)

	return logCard
}

// createButtonPanel creates the action buttons panel
func (app *GUIApp) createButtonPanel(msg *ui.Messages) fyne.CanvasObject {
	// Create buttons with localized labels
	configBtn := widget.NewButton(msg.ConfigCmd, func() {
		app.onConfigPressed()
	})
	configBtn.Importance = widget.MediumImportance

	logsBtn := widget.NewButton(msg.LogsCmd, func() {
		app.onLogsPressed()
	})

	scanBtn := widget.NewButton(msg.ScanCmd, func() {
		app.onScanPressed()
	})
	scanBtn.Importance = widget.HighImportance

	// Create recording button with dynamic text and importance
	recordBtn := widget.NewButton("🎤 "+msg.RecordCmd, func() {
		app.onRecordPressed()
	})
	recordBtn.Importance = widget.WarningImportance

	// Store reference for updating button text and appearance
	app.recordButton = recordBtn

	inputBtn := widget.NewButton(msg.InputDirCmd, func() {
		app.onInputDirPressed()
	})

	outputBtn := widget.NewButton(msg.OutputDirCmd, func() {
		app.onOutputDirPressed()
	})

	quitBtn := widget.NewButton(msg.QuitCmd, func() {
		app.onQuitPressed()
	})
	quitBtn.Importance = widget.DangerImportance

	// Create primary and secondary button groups for better organization
	primaryButtons := container.NewHBox(
		scanBtn,
		recordBtn,
	)

	configButtons := container.NewHBox(
		configBtn,
		logsBtn,
	)

	directoryButtons := container.NewHBox(
		inputBtn,
		outputBtn,
	)

	// Arrange buttons in a horizontal container with better spacing
	buttonContainer := container.NewHBox(
		primaryButtons,
		widget.NewSeparator(),
		configButtons,
		widget.NewSeparator(),
		directoryButtons,
		widget.NewSeparator(),
		quitBtn,
	)

	return buttonContainer
}
