package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/hirokitakamura/koemoji-go/internal/ui"
)

// createWindow creates and configures the main application window
func (app *GUIApp) createWindow() {
	// Create the main window
	app.window = app.fyneApp.NewWindow("KoeMoji-Go")
	app.window.Resize(fyne.NewSize(900, 750))
	app.window.CenterOnScreen()
	app.window.SetMaster()

	// Set window icon
	app.window.SetIcon(GetAppIcon())

	// Create UI components
	app.createComponents()

	// Create title with custom app icon
	titleIcon := widget.NewIcon(GetAppIcon())
	titleLabel := widget.NewLabelWithStyle("KoeMoji-Go", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	titleContainer := container.NewCenter(
		container.NewHBox(titleIcon, titleLabel),
	)

	// Create top section with title and status
	topSection := container.NewVBox(
		titleContainer,
		app.statusWidget,
	)

	// Set up the main layout using BorderLayout with padding
	bottomWithPadding := container.NewVBox(
		app.buttonWidget,
		widget.NewLabel(""), // Small spacer for bottom margin
	)

	// Add padding to the entire content
	content := container.NewPadded(
		container.NewBorder(
			topSection,        // top (title + status)
			bottomWithPadding, // bottom with padding
			nil,               // left
			nil,               // right
			app.logWidget,     // center
		),
	)

	app.window.SetContent(content)

	// Set up window close behavior with recording check
	app.window.SetCloseIntercept(func() {
		// KISS Design: Consistent state check across all exit points
		if app.isRecording() {
			// Show warning dialog if recording is in progress
			app.showRecordingExitWarning()
			return
		}
		// Immediate exit if not recording
		app.forceQuit()
	})

	// Start periodic updates (5 seconds as per design)
	app.startPeriodicUpdate()
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

	// Mark UI as initialized
	app.uiInitialized = true
}

// createStatusPanel creates the status display panel
func (app *GUIApp) createStatusPanel(msg *ui.Messages) fyne.CanvasObject {
	// Status line 1: Active/Processing state and queue info
	app.statusIcon = widget.NewIcon(theme.ConfirmIcon())
	app.statusLabel = widget.NewLabel(msg.Active + " | " + msg.Queue + ": 0 | " + msg.Processing + ": " + msg.None)
	app.statusContainer = container.NewHBox(app.statusIcon, app.statusLabel)

	// Status line 2: File counts
	app.filesIcon = widget.NewIcon(theme.FolderIcon())
	app.filesLabel = widget.NewLabel(msg.Input + ": 0 → " + msg.Output + ": 0 → " + msg.Archive + ": 0")
	app.filesContainer = container.NewHBox(app.filesIcon, app.filesLabel)

	// Status line 3: Timing info
	app.timingIcon = widget.NewIcon(theme.SearchIcon())
	app.timingLabel = widget.NewLabel(msg.Last + ": " + msg.Never + " | " + msg.Next + ": " + msg.Soon + " | " + msg.Uptime + ": 0s")
	app.timingContainer = container.NewHBox(app.timingIcon, app.timingLabel)

	// Create a card container for the status panel
	statusCard := widget.NewCard("", "", container.NewVBox(
		app.statusContainer,
		app.filesContainer,
		app.timingContainer,
	))

	return statusCard
}

// createLogViewer creates the scrollable log display
func (app *GUIApp) createLogViewer(msg *ui.Messages) fyne.CanvasObject {
	// Create a rich text widget for log display
	app.logText = widget.NewRichTextFromMarkdown(msg.LogPlaceholder)
	app.logText.Wrapping = fyne.TextWrapWord

	// Create scrollable container
	logScroll := container.NewVScroll(app.logText)
	logScroll.SetMinSize(fyne.NewSize(750, 400))

	// Create a card container for the log viewer
	logCard := widget.NewCard(msg.LogTitle, "", logScroll)

	return logCard
}

// createButtonPanel creates the action buttons panel
func (app *GUIApp) createButtonPanel(msg *ui.Messages) fyne.CanvasObject {
	// Define button size
	buttonSize := fyne.NewSize(80, 40) // Width: 80, Height: 40

	// Create buttons with localized labels
	configBtn := widget.NewButton(msg.ConfigCmd, func() {
		app.onConfigPressed()
	})
	configBtn.Importance = widget.MediumImportance
	configBtn.Resize(buttonSize)

	logsBtn := widget.NewButton(msg.LogsCmd, func() {
		app.onLogsPressed()
	})
	logsBtn.Resize(buttonSize)

	scanBtn := widget.NewButton(msg.ScanCmd, func() {
		app.onScanPressed()
	})
	scanBtn.Importance = widget.HighImportance
	scanBtn.Resize(buttonSize)

	// Create recording button with dynamic text and importance
	recordBtn := widget.NewButton(msg.RecordCmd, func() {
		app.onRecordPressed()
	})
	recordBtn.Importance = widget.WarningImportance
	recordBtn.Resize(buttonSize)

	// Store reference for updating button text and appearance
	app.recordButton = recordBtn

	inputBtn := widget.NewButton(msg.InputDirCmd, func() {
		app.onInputDirPressed()
	})
	inputBtn.Resize(buttonSize)

	outputBtn := widget.NewButton(msg.OutputDirCmd, func() {
		app.onOutputDirPressed()
	})
	outputBtn.Resize(buttonSize)

	quitBtn := widget.NewButton(msg.QuitCmd, func() {
		app.onQuitPressed()
	})
	quitBtn.Importance = widget.DangerImportance
	quitBtn.Resize(buttonSize)

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

	// Arrange buttons with appropriate spacing
	buttonContainer := container.NewHBox(
		layout.NewSpacer(),
		primaryButtons,
		widget.NewLabel("   "), // Fixed spacing between groups
		configButtons,
		widget.NewLabel("   "), // Fixed spacing between groups
		directoryButtons,
		widget.NewLabel("   "), // Fixed spacing between groups
		quitBtn,
		layout.NewSpacer(),
	)

	// Add padding around the button container
	paddedContainer := container.NewPadded(buttonContainer)

	return paddedContainer
}
