package ui

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/infoHiroki/KoeMoji-Go/internal/config"
	"github.com/infoHiroki/KoeMoji-Go/internal/logger"
	"github.com/infoHiroki/KoeMoji-Go/internal/recorder"
	"github.com/rivo/tview"
)

// Helper functions for volume conversion (Phase 14)
// Convert float64 volume (0.0-1.0) to index (0-4)
func volumeFloatToIndex(vol float64) int {
	systemVolumeScale := []float64{0.1, 0.2, 0.3, 0.5, 0.7}
	micVolumeScale := []float64{1.0, 1.3, 1.6, 1.9, 2.2}

	// Try system scale first
	for i, v := range systemVolumeScale {
		if vol <= v+0.05 { // Tolerance
			return i
		}
	}

	// Try mic scale
	for i, v := range micVolumeScale {
		if vol <= v+0.05 {
			return i
		}
	}

	return 2 // Default to middle (0)
}

// Convert index (0-4) to float64 volume for system audio
func volumeIndexToSystemFloat(idx int) float64 {
	systemVolumeScale := []float64{0.1, 0.2, 0.3, 0.5, 0.7}
	if idx >= 0 && idx < len(systemVolumeScale) {
		return systemVolumeScale[idx]
	}
	return 0.3 // Default
}

// Convert index (0-4) to float64 volume for microphone
func volumeIndexToMicFloat(idx int) float64 {
	micVolumeScale := []float64{1.0, 1.3, 1.6, 1.9, 2.2}
	if idx >= 0 && idx < len(micVolumeScale) {
		return micVolumeScale[idx]
	}
	return 1.6 // Default
}

// RichTUICallbacks contains callback functions for RichTUI actions (Phase 11)
type RichTUICallbacks struct {
	OnRecordingToggle func() error        // 録音開始/停止
	OnScanTrigger     func() error        // 手動スキャン実行
	OnOpenLogFile     func() error        // ログファイルを開く
	OnOpenDirectory   func(dir string) error // フォルダを開く
	OnRefreshFileList func() error        // ファイルリスト更新
}

// RichTUI represents a rich terminal UI (LazyGit/k9s style)
// Phase 11: Integrated with actual application functions
type RichTUI struct {
	app         *tview.Application
	config      *config.Config
	callbacks   *RichTUICallbacks // Phase 11: Callbacks for actions
	menuList    *tview.List
	statusBar   *tview.TextView
	helpBar     *tview.TextView
	contentArea *tview.Pages // Phase 8: Changed from TextView to Pages
	mainFlex    *tview.Flex

	// Content pages (Phase 8)
	dashboardPage *tview.TextView // Phase 12: Real-time log display
	settingsPage  *tview.TextView
	logsPage      *tview.TextView
	scanPage      *tview.TextView
	recordPage    *tview.TextView
	inputPage     *tview.List // Phase 9: Changed to List for file display
	outputPage    *tview.List // Phase 9: Changed to List for file display
	archivePage   *tview.List // Phase 9: Archive folder file display

	// Status tracking (Phase 7)
	startTime      time.Time
	inputCount     int
	outputCount    int
	archiveCount   int
	isProcessing   bool
	processingFile string
	isRecording    bool
	recordingStart time.Time
	mu             sync.RWMutex
}

// NewRichTUI creates a new rich TUI (Phase 11: with callbacks)
func NewRichTUI(cfg *config.Config, callbacks *RichTUICallbacks) *RichTUI {
	app := tview.NewApplication()

	// Create status bar (top, 3 lines)
	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[green]KoeMoji-Go Rich TUI[white] | Phase 7\n行2: ファイル数\n行3: タイミング情報")
	statusBar.SetBorder(false)

	// Create menu list (left side, fixed width) - Phase 12: 9 items
	list := tview.NewList().ShowSecondaryText(false)
	list.AddItem("1. ダッシュボード", "", 0, nil) // Phase 12: Real-time logs
	list.AddItem("2. 設定", "", 0, nil)
	list.AddItem("3. ログファイル", "", 0, nil)
	list.AddItem("4. スキャン", "", 0, nil)
	list.AddItem("5. 録音", "", 0, nil)
	list.AddItem("6. 入力", "", 0, nil)
	list.AddItem("7. 出力", "", 0, nil)
	list.AddItem("8. アーカイブ", "", 0, nil)
	list.AddItem("9. 終了", "", 0, nil)

	list.SetBorder(true).
		SetTitle(" メニュー ").
		SetTitleAlign(tview.AlignCenter)

	// Create content area with Pages (Phase 8)
	contentArea := tview.NewPages()

	// Create individual pages for each menu item (Phase 8/9/12)
	dashboardPage := createBorderedTextView(" ダッシュボード ", "[yellow]1. ダッシュボード[white]\n\nリアルタイムログ（最新12件）\n\n起動中...")
	settingsPage := createBorderedTextView(" 設定 ", "[yellow]2. 設定[white]\n\n設定画面の内容がここに表示されます\n\n• Whisperモデル\n• 入力/出力ディレクトリ\n• OpenAI API設定")
	logsPage := createBorderedTextView(" ログファイル ", "[yellow]3. ログファイル[white]\n\nログファイルを開きます\n\n• Enterキーでログファイルを開く")
	scanPage := createBorderedTextView(" スキャン ", "[yellow]4. スキャン[white]\n\n入力フォルダをスキャンして音声ファイルを検出します\n\n• 手動スキャン実行\n• ファイル検出")
	recordPage := createBorderedTextView(" 録音 ", "[yellow]5. 録音[white]\n\n音声録音機能\n\n• 録音開始/停止\n• デバイス選択\n• 音量調整")

	// Phase 9: Create file lists for input/output folders
	inputPage, inputErr := CreateFileList(cfg.InputDir, app)
	if inputErr != nil {
		// Fallback: create empty list with error message
		inputPage = tview.NewList().ShowSecondaryText(false)
		inputPage.AddItem(fmt.Sprintf("[red]エラー:[white] %v", inputErr), "", 0, nil)
	}
	inputPage.SetBorder(true).
		SetTitle(GetFileListTitle("input", cfg.InputDir)).
		SetTitleAlign(tview.AlignCenter)

	outputPage, outputErr := CreateFileList(cfg.OutputDir, app)
	if outputErr != nil {
		// Fallback: create empty list with error message
		outputPage = tview.NewList().ShowSecondaryText(false)
		outputPage.AddItem(fmt.Sprintf("[red]エラー:[white] %v", outputErr), "", 0, nil)
	}
	outputPage.SetBorder(true).
		SetTitle(GetFileListTitle("output", cfg.OutputDir)).
		SetTitleAlign(tview.AlignCenter)

	archivePage, archiveErr := CreateFileList(cfg.ArchiveDir, app)
	if archiveErr != nil {
		// Fallback: create empty list with error message
		archivePage = tview.NewList().ShowSecondaryText(false)
		archivePage.AddItem(fmt.Sprintf("[red]エラー:[white] %v", archiveErr), "", 0, nil)
	}
	archivePage.SetBorder(true).
		SetTitle(GetFileListTitle("archive", cfg.ArchiveDir)).
		SetTitleAlign(tview.AlignCenter)

	// Add pages to content area (Phase 12: dashboard first)
	contentArea.AddPage("dashboard", dashboardPage, true, true)
	contentArea.AddPage("settings", settingsPage, true, false)
	contentArea.AddPage("logs", logsPage, true, false)
	contentArea.AddPage("scan", scanPage, true, false)
	contentArea.AddPage("record", recordPage, true, false)
	contentArea.AddPage("input", inputPage, true, false)
	contentArea.AddPage("output", outputPage, true, false)
	contentArea.AddPage("archive", archivePage, true, false)

	// Create help bar (bottom, 1 line)
	helpBar := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[yellow]↑↓/j/k[white]:移動 [yellow]Enter[white]:選択 [yellow]q[white]:終了 [yellow]?[white]:ヘルプ")
	helpBar.SetBorder(false)

	// Create left-right split layout
	middleFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(list, 20, 0, true).         // Menu: fixed 20 chars width
		AddItem(contentArea, 0, 1, false)   // Content: expand to fill

	// Create 3-row layout (status / middle / help)
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(statusBar, 3, 0, false).    // Fixed 3 lines (Phase 7)
		AddItem(middleFlex, 0, 1, true).    // Expand to fill
		AddItem(helpBar, 1, 0, false)       // Fixed 1 line

	// Handle cursor movement: switch pages on selection change (Phase 8/12)
	list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		// Page names corresponding to menu indices (Phase 12: 9 items)
		pageNames := []string{"dashboard", "settings", "logs", "scan", "record", "input", "output", "archive", "quit"}
		if index >= 0 && index < len(pageNames) && index != 8 {
			contentArea.SwitchToPage(pageNames[index])
		}
	})

	// Create RichTUI struct early to pass mainFlex to showRichHelpDialog
	tui := &RichTUI{
		app:           app,
		config:        cfg,
		callbacks:     callbacks,     // Phase 11
		menuList:      list,
		statusBar:     statusBar,
		helpBar:       helpBar,
		contentArea:   contentArea,
		mainFlex:      mainFlex,
		dashboardPage: dashboardPage, // Phase 12
		settingsPage:  settingsPage,  // Phase 8
		logsPage:      logsPage,      // Phase 8
		scanPage:      scanPage,      // Phase 8
		recordPage:    recordPage,    // Phase 8
		inputPage:     inputPage,     // Phase 9
		outputPage:    outputPage,    // Phase 9
		archivePage:   archivePage,   // Phase 9
		startTime:     time.Now(),    // Phase 7
	}

	// Initial status update (Phase 7)
	tui.updateStatusBar()

	// Handle custom key bindings (j/k/q/?)
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j', 'J':
				// j: Move down (same as ↓)
				return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
			case 'k', 'K':
				// k: Move up (same as ↑)
				return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
			case 'q', 'Q':
				// q: Quit application
				app.Stop()
				return nil
			case '?':
				// ?: Show help dialog
				showRichHelpDialog(app, mainFlex)
				return nil
			}
		}
		// Return event for default behavior (arrow keys, Enter, etc.)
		return event
	})

	// Handle Enter key selection (Phase 11/12: integrated functions)
	list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		switch index {
		case 0:
			// ダッシュボード - Phase 12: Just display (no action needed)
		case 1:
			// 設定 - Phase 10: Show config dialog
			tui.showConfigDialog()
		case 2:
			// ログファイル - Phase 11: Open log file
			if tui.callbacks != nil && tui.callbacks.OnOpenLogFile != nil {
				tui.callbacks.OnOpenLogFile()
			}
		case 3:
			// スキャン - Phase 11: Trigger manual scan
			if tui.callbacks != nil && tui.callbacks.OnScanTrigger != nil {
				tui.callbacks.OnScanTrigger()
			}
		case 4:
			// 録音 - Phase 11: Toggle recording
			if tui.callbacks != nil && tui.callbacks.OnRecordingToggle != nil {
				tui.callbacks.OnRecordingToggle()
			}
		case 5:
			// 入力フォルダ - Phase 11: Open input directory
			if tui.callbacks != nil && tui.callbacks.OnOpenDirectory != nil {
				tui.callbacks.OnOpenDirectory(tui.config.InputDir)
			}
		case 6:
			// 出力フォルダ - Phase 11: Open output directory
			if tui.callbacks != nil && tui.callbacks.OnOpenDirectory != nil {
				tui.callbacks.OnOpenDirectory(tui.config.OutputDir)
			}
		case 7:
			// アーカイブ - Phase 11: Open archive directory
			if tui.callbacks != nil && tui.callbacks.OnOpenDirectory != nil {
				tui.callbacks.OnOpenDirectory(tui.config.ArchiveDir)
			}
		case 8:
			// 終了
			app.Stop()
		}
	})

	return tui
}

// Run starts the rich TUI
func (t *RichTUI) Run() error {
	return t.app.SetRoot(t.mainFlex, true).Run()
}

// Stop stops the rich TUI
func (t *RichTUI) Stop() {
	t.app.Stop()
}

// showRichHelpDialog shows a help dialog with key bindings
func showRichHelpDialog(app *tview.Application, mainFlex *tview.Flex) {
	helpText := `[yellow]KoeMoji-Go Rich TUI - ヘルプ[white]

[yellow]キー操作:[white]
  ↑ / k     : 上に移動
  ↓ / j     : 下に移動
  Enter     : 選択 / フォルダを開く
  r         : ファイルリスト再読み込み
  q         : 終了
  ?         : このヘルプを表示

[yellow]Phase 9の状態:[white]
  • 入力/出力/アーカイブフォルダに実際のファイル一覧を表示
  • Enterでフォルダを開く、rで再読み込み
  • 3行のステータスバー（リアルタイム更新）
  • 8メニュー項目

[yellow]メニュー:[white]
  1. 設定        - アプリケーション設定
  2. ログ        - ログ表示
  3. スキャン    - 入力フォルダスキャン
  4. 録音        - 音声録音
  5. 入力        - 入力フォルダ一覧（実ファイル表示）
  6. 出力        - 出力フォルダ一覧（実ファイル表示）
  7. アーカイブ  - アーカイブフォルダ一覧（実ファイル表示）
  8. 終了        - アプリケーション終了

[green]Escキーまたは閉じるボタンで戻る[white]`

	modal := tview.NewModal().
		SetText(helpText).
		AddButtons([]string{"閉じる"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			app.SetRoot(mainFlex, true)
		})

	// Handle Esc key to close
	modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			app.SetRoot(mainFlex, true)
			return nil
		}
		return event
	})

	app.SetRoot(modal, true)
}

// updateStatusBar updates the status bar display (Phase 7)
func (t *RichTUI) updateStatusBar() {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Line 1: Status and processing info
	statusIcon := "[green]●[white]"
	statusText := "待機中"

	if t.isRecording {
		statusIcon = "[red]●[white]"
		elapsed := time.Since(t.recordingStart)
		statusText = fmt.Sprintf("録音中 (%s)", formatDuration(elapsed))
	} else if t.isProcessing {
		statusIcon = "[yellow]●[white]"
		if t.processingFile != "" {
			statusText = fmt.Sprintf("処理中: %s", t.processingFile)
		} else {
			statusText = "処理中"
		}
	}

	line1 := fmt.Sprintf("%s %s | Phase 7", statusIcon, statusText)

	// Line 2: File counts
	line2 := fmt.Sprintf("[blue]入力:[white]%d → [green]出力:[white]%d → [gray]保存:[white]%d",
		t.inputCount, t.outputCount, t.archiveCount)

	// Line 3: Timing info
	uptime := time.Since(t.startTime)
	line3 := fmt.Sprintf("[yellow]起動時間:[white] %s", formatDuration(uptime))

	// Update status bar
	t.statusBar.SetText(fmt.Sprintf("%s\n%s\n%s", line1, line2, line3))
}

// UpdateStatus updates status information from main goroutine (Phase 7)
func (t *RichTUI) UpdateStatus(inputCount, outputCount, archiveCount int,
	processingFile string, isProcessing bool, isRecording bool, recordingStart time.Time) {

	t.mu.Lock()
	t.inputCount = inputCount
	t.outputCount = outputCount
	t.archiveCount = archiveCount
	t.processingFile = processingFile
	t.isProcessing = isProcessing
	t.isRecording = isRecording
	t.recordingStart = recordingStart
	t.mu.Unlock()

	// Queue UI update
	t.app.QueueUpdateDraw(func() {
		t.updateStatusBar()
	})
}

// UpdateFileLists updates the file lists for input/output/archive directories (Phase 11)
func (t *RichTUI) UpdateFileLists() {
	t.app.QueueUpdateDraw(func() {
		// Update input file list
		inputList, inputErr := CreateFileList(t.config.InputDir, t.app)
		if inputErr == nil {
			inputList.SetBorder(true).
				SetTitle(GetFileListTitle("input", t.config.InputDir)).
				SetTitleAlign(tview.AlignCenter)
			t.contentArea.AddPage("input", inputList, true, false)
			t.inputPage = inputList
		}

		// Update output file list
		outputList, outputErr := CreateFileList(t.config.OutputDir, t.app)
		if outputErr == nil {
			outputList.SetBorder(true).
				SetTitle(GetFileListTitle("output", t.config.OutputDir)).
				SetTitleAlign(tview.AlignCenter)
			t.contentArea.AddPage("output", outputList, true, false)
			t.outputPage = outputList
		}

		// Update archive file list
		archiveList, archiveErr := CreateFileList(t.config.ArchiveDir, t.app)
		if archiveErr == nil {
			archiveList.SetBorder(true).
				SetTitle(GetFileListTitle("archive", t.config.ArchiveDir)).
				SetTitleAlign(tview.AlignCenter)
			t.contentArea.AddPage("archive", archiveList, true, false)
			t.archivePage = archiveList
		}
	})
}

// UpdateDashboard updates the dashboard page with real-time logs (Phase 12)
func (t *RichTUI) UpdateDashboard(logBuffer []logger.LogEntry) {
	t.app.QueueUpdateDraw(func() {
		// Build log text with colors
		logText := "[yellow]リアルタイムログ（最新12件）[white]\n\n"

		if len(logBuffer) == 0 {
			logText += "[gray]ログがありません[white]"
		} else {
			for _, entry := range logBuffer {
				// Get color based on log level
				color := getLogColorTUI(entry.Level)
				timestamp := entry.Timestamp.Format("15:04:05")

				// Format: [COLOR]LEVEL[white] HH:MM:SS Message
				logText += fmt.Sprintf("%s%-5s[white] %s %s\n", color, entry.Level, timestamp, entry.Message)
			}
		}

		t.dashboardPage.SetText(logText)
	})
}

// getLogColorTUI returns tview color tag for log level (Phase 12)
func getLogColorTUI(level string) string {
	switch level {
	case "INFO":
		return "[blue]"
	case "PROC":
		return "[yellow]"
	case "DONE":
		return "[green]"
	case "ERROR":
		return "[red]"
	case "DEBUG":
		return "[gray]"
	default:
		return "[white]"
	}
}

// UpdateScanPage updates the scan page with real-time scan status (Phase 13)
func (t *RichTUI) UpdateScanPage(lastScanTime time.Time, fileCount int, isScanning bool) {
	t.app.QueueUpdateDraw(func() {
		var statusText string

		if isScanning {
			statusText = "[yellow]スキャン実行中...[white]\n\n• 入力フォルダをスキャンしています"
		} else if !lastScanTime.IsZero() {
			timeStr := lastScanTime.Format("15:04:05")
			statusText = fmt.Sprintf("[green]スキャン完了[white]\n\n• 最終スキャン: %s\n• 検出ファイル数: %d件", timeStr, fileCount)
		} else {
			statusText = "Enterキーで手動スキャン実行\n\n• 入力フォルダをスキャンして音声ファイルを検出します"
		}

		fullText := "[yellow]4. スキャン[white]\n\n" + statusText
		t.scanPage.SetText(fullText)
	})
}

// UpdateRecordPage updates the record page with real-time recording status (Phase 13)
func (t *RichTUI) UpdateRecordPage(isRecording bool, recordingStart time.Time, deviceName string) {
	t.app.QueueUpdateDraw(func() {
		var statusText string

		if isRecording {
			elapsed := time.Since(recordingStart)
			elapsedStr := formatDuration(elapsed)
			now := time.Now()
			filename := fmt.Sprintf("recording_%s.wav", now.Format("20060102_1504"))

			statusText = fmt.Sprintf("[red]🔴 録音中: %s[white]\n\n• Enterキーで停止\n• ファイル: %s", elapsedStr, filename)
		} else {
			deviceText := "デフォルトデバイス"
			if deviceName != "" && deviceName != "デフォルトデバイス" {
				deviceText = deviceName
			}
			statusText = fmt.Sprintf("Enterキーで録音開始\n\n• 録音デバイス: %s\n• 録音準備完了", deviceText)
		}

		fullText := "[yellow]5. 録音[white]\n\n" + statusText
		t.recordPage.SetText(fullText)
	})
}

// createBorderedTextView creates a bordered TextView with title (Phase 8 helper)
func createBorderedTextView(title, text string) *tview.TextView {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetText(text)
	textView.SetBorder(true).
		SetTitle(title).
		SetTitleAlign(tview.AlignCenter)
	return textView
}

// showConfigDialog shows configuration settings dialog (Phase 10 - split layout)
func (t *RichTUI) showConfigDialog() {
	// Whisper models
	whisperModels := []string{
		"tiny", "tiny.en", "base", "base.en",
		"small", "small.en", "medium", "medium.en",
		"large", "large-v1", "large-v2", "large-v3",
	}
	currentModelIndex := 0
	for i, model := range whisperModels {
		if model == t.config.WhisperModel {
			currentModelIndex = i
			break
		}
	}

	// Languages
	languages := []string{"auto", "ja", "en", "zh", "ko", "es", "fr", "de", "ru"}
	languageNames := []string{"自動", "日本語", "English", "中文", "한국어", "Español", "Français", "Deutsch", "Русский"}
	currentLangIndex := 0
	for i, lang := range languages {
		if lang == t.config.Language {
			currentLangIndex = i
			break
		}
	}

	// Create category menu (left side) - Phase 14: 4 categories
	categoryList := tview.NewList().ShowSecondaryText(false)
	categoryList.AddItem("1. 基本設定", "", 0, nil)
	categoryList.AddItem("2. ディレクトリ", "", 0, nil)
	categoryList.AddItem("3. LLM設定", "", 0, nil)
	categoryList.AddItem("4. 録音設定", "", 0, nil) // Phase 14: NEW
	categoryList.AddItem("", "", 0, nil)            // Separator
	categoryList.AddItem("保存", "", 's', nil)
	categoryList.AddItem("キャンセル", "", 'q', nil)
	categoryList.SetBorder(true).
		SetTitle(" カテゴリ ").
		SetTitleAlign(tview.AlignCenter)

	// Create content area (right side) with Pages - List based for better navigation
	contentArea := tview.NewPages()

	// Language code to display name mapping
	codeToDisplayMap := map[string]string{
		"auto": "自動",
		"ja":   "日本語",
		"en":   "English",
		"zh":   "中文",
		"ko":   "한국어",
		"es":   "Español",
		"fr":   "Français",
		"de":   "Deutsch",
		"ru":   "Русский",
	}

	// Get current language display name
	langDisplay := "日本語"
	if display, exists := codeToDisplayMap[t.config.Language]; exists {
		langDisplay = display
	}

	// === Page 1: Basic Settings List ===
	// UI Language display
	uiLangDisplay := "日本語"
	if t.config.UILanguage == "en" {
		uiLangDisplay = "English"
	}

	basicList := tview.NewList().ShowSecondaryText(true)
	basicList.AddItem("UI言語", uiLangDisplay, 0, nil)
	basicList.AddItem("Whisperモデル", t.config.WhisperModel, 0, nil)
	basicList.AddItem("認識言語", langDisplay, 0, nil)
	basicList.AddItem("スキャン間隔", fmt.Sprintf("%d分", t.config.ScanIntervalMinutes), 0, nil)
	basicList.SetBorder(true).
		SetTitle(" 基本設定 (Enterで編集) ").
		SetTitleAlign(tview.AlignCenter)

	// === Page 2: Directories List ===
	dirList := tview.NewList().ShowSecondaryText(true)
	dirList.AddItem("入力フォルダ", t.config.InputDir, 0, nil)
	dirList.AddItem("出力フォルダ", t.config.OutputDir, 0, nil)
	dirList.AddItem("保存フォルダ", t.config.ArchiveDir, 0, nil)
	dirList.SetBorder(true).
		SetTitle(" ディレクトリ設定 (Enterで編集) ").
		SetTitleAlign(tview.AlignCenter)

	// === Page 3: LLM Settings List ===
	llmStatusText := "無効"
	if t.config.LLMSummaryEnabled {
		llmStatusText = "有効"
	}
	apiKeyDisplay := "未設定"
	if t.config.LLMAPIKey != "" {
		if len(t.config.LLMAPIKey) >= 10 {
			apiKeyDisplay = t.config.LLMAPIKey[:4] + "..." + t.config.LLMAPIKey[len(t.config.LLMAPIKey)-4:]
		} else {
			apiKeyDisplay = "設定済み"
		}
	}
	// Prompt template display (first 30 chars)
	promptDisplay := t.config.SummaryPromptTemplate
	if len(promptDisplay) > 30 {
		promptDisplay = promptDisplay[:30] + "..."
	}
	if promptDisplay == "" {
		promptDisplay = "デフォルト"
	}

	llmList := tview.NewList().ShowSecondaryText(true)
	llmList.AddItem("LLM要約機能", llmStatusText, 0, nil)
	llmList.AddItem("OpenAI APIキー", apiKeyDisplay, 0, nil)
	llmList.AddItem("LLMモデル", t.config.LLMModel, 0, nil)
	llmList.AddItem("プロンプト", promptDisplay, 0, nil)
	llmList.SetBorder(true).
		SetTitle(" LLM設定 (Enterで編集) ").
		SetTitleAlign(tview.AlignCenter)

	// === Page 4: Recording Settings List ===
	deviceDisplay := t.config.RecordingDeviceName
	if deviceDisplay == "" {
		deviceDisplay = "デフォルト"
	}
	modeDisplay := "シングル"
	if t.config.DualRecordingEnabled {
		modeDisplay = "デュアル"
	}

	recordingList := tview.NewList().ShowSecondaryText(true)
	recordingList.AddItem("録音デバイス", deviceDisplay, 0, nil)
	recordingList.AddItem("録音モード", modeDisplay, 0, nil)

	// Platform-specific settings: Volume controls only on Windows
	if runtime.GOOS == "windows" {
		// Convert float64 (0.0-1.0) to display scale (-2 to +2)
		// SystemAudioVolume: 0.1, 0.2, 0.3, 0.5, 0.7 -> display as -2, -1, 0, +1, +2
		sysVolumeIdx := volumeFloatToIndex(t.config.SystemAudioVolume)
		micVolumeIdx := volumeFloatToIndex(t.config.MicrophoneVolume)
		sysVolumeDisplay := fmt.Sprintf("%+d", sysVolumeIdx-2)
		micVolumeDisplay := fmt.Sprintf("%+d", micVolumeIdx-2)

		recordingList.AddItem("システム音声音量", sysVolumeDisplay, 0, nil)
		recordingList.AddItem("マイク音量", micVolumeDisplay, 0, nil)
	}

	recordingList.SetBorder(true).
		SetTitle(" 録音設定 (Enterで編集) ").
		SetTitleAlign(tview.AlignCenter)

	// Add pages
	contentArea.AddPage("basic", basicList, true, true)
	contentArea.AddPage("directories", dirList, true, false)
	contentArea.AddPage("llm", llmList, true, false)
	contentArea.AddPage("recording", recordingList, true, false)

	// Handle category selection (cursor movement)
	categoryList.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		pageNames := []string{"basic", "directories", "llm", "recording"}
		if index >= 0 && index < len(pageNames) {
			contentArea.SwitchToPage(pageNames[index])
		}
	})

	// Create main container (will be populated later)
	var mainContainer *tview.Flex

	// Helper function to show edit dialog
	showEditDialog := func(title string, widget tview.Primitive) {

		// Create modal container
		modal := tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().
				SetDirection(tview.FlexColumn).
				AddItem(nil, 0, 1, false).
				AddItem(widget, 100, 1, true).
				AddItem(nil, 0, 1, false), 20, 1, true).
			AddItem(nil, 0, 1, false)

		t.app.SetRoot(modal, true).SetFocus(widget)
	}

	// Close edit dialog and return to main
	closeEditDialog := func() {
		t.app.SetRoot(mainContainer, true).SetFocus(contentArea)
	}

	// Edit handlers for Basic Settings
	basicList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		switch index {
		case 0: // UI Language
			dropdown := tview.NewDropDown().
				SetLabel("UI言語: ").
				SetOptions([]string{"日本語", "English"}, nil)

			// Set current selection
			if t.config.UILanguage == "en" {
				dropdown.SetCurrentOption(1)
			} else {
				dropdown.SetCurrentOption(0)
			}

			dropdown.SetBorder(true).
				SetTitle(" UI言語を選択 ").
				SetTitleAlign(tview.AlignCenter)

			dropdown.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEscape {
					closeEditDialog()
					return nil
				}
				if event.Key() == tcell.KeyEnter {
					idx, displayName := dropdown.GetCurrentOption()
					if idx == 0 {
						t.config.UILanguage = "ja"
					} else {
						t.config.UILanguage = "en"
					}
					basicList.SetItemText(0, "UI言語", displayName)
					closeEditDialog()
					return nil
				}
				return event
			})

			showEditDialog("UI言語", dropdown)

		case 1: // Whisper Model
			dropdown := tview.NewDropDown().
				SetLabel("Whisperモデル: ").
				SetOptions(whisperModels, nil).
				SetCurrentOption(currentModelIndex)

			dropdown.SetBorder(true).
				SetTitle(" Whisperモデルを選択 ").
				SetTitleAlign(tview.AlignCenter)

			dropdown.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEscape {
					closeEditDialog()
					return nil
				}
				if event.Key() == tcell.KeyEnter {
					idx, modelName := dropdown.GetCurrentOption()
					currentModelIndex = idx
					t.config.WhisperModel = modelName
					basicList.SetItemText(1, "Whisperモデル", modelName)
					closeEditDialog()
					return nil
				}
				return event
			})

			showEditDialog("Whisperモデル", dropdown)

		case 2: // Language
			dropdown := tview.NewDropDown().
				SetLabel("認識言語: ").
				SetOptions(languageNames, nil).
				SetCurrentOption(currentLangIndex)

			dropdown.SetBorder(true).
				SetTitle(" 認識言語を選択 ").
				SetTitleAlign(tview.AlignCenter)

			dropdown.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEscape {
					closeEditDialog()
					return nil
				}
				if event.Key() == tcell.KeyEnter {
					idx, _ := dropdown.GetCurrentOption()
					currentLangIndex = idx
					t.config.Language = languages[idx]
					basicList.SetItemText(2, "認識言語", languageNames[idx])
					closeEditDialog()
					return nil
				}
				return event
			})

			showEditDialog("認識言語", dropdown)

		case 3: // Scan Interval
			field := tview.NewInputField().
				SetLabel("スキャン間隔（分）: ").
				SetText(fmt.Sprintf("%d", t.config.ScanIntervalMinutes)).
				SetFieldWidth(10)

			field.SetBorder(true).
				SetTitle(" スキャン間隔を編集 ").
				SetTitleAlign(tview.AlignCenter)

			field.SetDoneFunc(func(key tcell.Key) {
				if key == tcell.KeyEscape {
					closeEditDialog()
				} else if key == tcell.KeyEnter {
					text := field.GetText()
					if interval, err := strconv.Atoi(text); err == nil && interval > 0 {
						t.config.ScanIntervalMinutes = interval
						basicList.SetItemText(3, "スキャン間隔", fmt.Sprintf("%d分", interval))
					}
					closeEditDialog()
				}
			})

			showEditDialog("スキャン間隔", field)
		}
	})

	// Edit handlers for Directories
	dirList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		var field *tview.InputField
		var targetConfig *string
		var itemIndex int = index

		switch index {
		case 0: // Input Dir
			field = tview.NewInputField().
				SetLabel("入力フォルダ: ").
				SetText(t.config.InputDir).
				SetFieldWidth(70)
			targetConfig = &t.config.InputDir
		case 1: // Output Dir
			field = tview.NewInputField().
				SetLabel("出力フォルダ: ").
				SetText(t.config.OutputDir).
				SetFieldWidth(70)
			targetConfig = &t.config.OutputDir
		case 2: // Archive Dir
			field = tview.NewInputField().
				SetLabel("保存フォルダ: ").
				SetText(t.config.ArchiveDir).
				SetFieldWidth(70)
			targetConfig = &t.config.ArchiveDir
		}

		field.SetBorder(true).
			SetTitle(" " + mainText + " を編集 ").
			SetTitleAlign(tview.AlignCenter)

		field.SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEscape {
				closeEditDialog()
			} else if key == tcell.KeyEnter {
				newValue := field.GetText()
				*targetConfig = newValue
				dirList.SetItemText(itemIndex, mainText, newValue)
				closeEditDialog()
			}
		})

		showEditDialog(mainText, field)
	})

	// Edit handlers for LLM Settings
	llmList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		switch index {
		case 0: // LLM Status toggle
			list := tview.NewList().ShowSecondaryText(false)
			list.AddItem("有効", "", '1', nil)
			list.AddItem("無効", "", '2', nil)

			// Set current selection
			if t.config.LLMSummaryEnabled {
				list.SetCurrentItem(0)
			} else {
				list.SetCurrentItem(1)
			}

			list.SetBorder(true).
				SetTitle(" LLM要約機能 ").
				SetTitleAlign(tview.AlignCenter)

			list.SetSelectedFunc(func(idx int, text, secondary string, r rune) {
				t.config.LLMSummaryEnabled = (idx == 0)
				statusText := "無効"
				if t.config.LLMSummaryEnabled {
					statusText = "有効"
				}
				llmList.SetItemText(0, "LLM要約機能", statusText)
				closeEditDialog()
			})

			list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEscape {
					closeEditDialog()
					return nil
				}
				return event
			})

			showEditDialog("LLM要約機能", list)

		case 1: // API Key
			field := tview.NewInputField().
				SetLabel("APIキー: ").
				SetText(t.config.LLMAPIKey).
				SetFieldWidth(70).
				SetMaskCharacter('*')

			field.SetBorder(true).
				SetTitle(" OpenAI APIキー を編集 ").
				SetTitleAlign(tview.AlignCenter)

			field.SetDoneFunc(func(key tcell.Key) {
				if key == tcell.KeyEscape {
					closeEditDialog()
				} else if key == tcell.KeyEnter {
					newValue := field.GetText()
					t.config.LLMAPIKey = newValue

					apiKeyDisplay := "未設定"
					if newValue != "" {
						if len(newValue) >= 10 {
							apiKeyDisplay = newValue[:4] + "..." + newValue[len(newValue)-4:]
						} else {
							apiKeyDisplay = "設定済み"
						}
					}
					llmList.SetItemText(1, "OpenAI APIキー", apiKeyDisplay)
					closeEditDialog()
				}
			})

			showEditDialog("APIキー", field)

		case 2: // LLM Model
			llmModels := []string{"gpt-4o", "gpt-4-turbo", "gpt-3.5-turbo"}
			currentModelIdx := 0
			for i, model := range llmModels {
				if model == t.config.LLMModel {
					currentModelIdx = i
					break
				}
			}

			dropdown := tview.NewDropDown().
				SetLabel("LLMモデル: ").
				SetOptions(llmModels, nil).
				SetCurrentOption(currentModelIdx)

			dropdown.SetBorder(true).
				SetTitle(" LLMモデルを選択 ").
				SetTitleAlign(tview.AlignCenter)

			dropdown.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEscape {
					closeEditDialog()
					return nil
				}
				if event.Key() == tcell.KeyEnter {
					_, modelName := dropdown.GetCurrentOption()
					t.config.LLMModel = modelName
					llmList.SetItemText(2, "LLMモデル", modelName)
					closeEditDialog()
					return nil
				}
				return event
			})

			showEditDialog("LLMモデル", dropdown)

		case 3: // Prompt Template
			textArea := tview.NewTextArea().
				SetText(t.config.SummaryPromptTemplate, true)

			textArea.SetBorder(true).
				SetTitle(" プロンプトテンプレートを編集 (Ctrl+S: 保存, Esc: キャンセル) ").
				SetTitleAlign(tview.AlignCenter)

			textArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEscape {
					closeEditDialog()
					return nil
				}
				// Ctrl+S to save
				if event.Key() == tcell.KeyCtrlS {
					newValue := textArea.GetText()
					t.config.SummaryPromptTemplate = newValue

					promptDisplay := newValue
					if len(promptDisplay) > 30 {
						promptDisplay = promptDisplay[:30] + "..."
					}
					if promptDisplay == "" {
						promptDisplay = "デフォルト"
					}
					llmList.SetItemText(3, "プロンプト", promptDisplay)
					closeEditDialog()
					return nil
				}
				return event
			})

			showEditDialog("プロンプト", textArea)
		}
	})

	// Edit handlers for Recording Settings
	recordingList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		switch index {
		case 0: // Recording Device
			// Get device list from recorder
			devices, err := recorder.ListDevices()
			if err != nil {
				// Show error
				t.statusBar.SetText(fmt.Sprintf("[red]デバイス取得エラー: %v[white]", err))
				time.AfterFunc(3*time.Second, func() {
					t.app.QueueUpdateDraw(func() {
						t.updateStatusBar()
					})
				})
				return
			}

			deviceNames := make([]string, len(devices)+1)
			deviceNames[0] = "デフォルト"
			for i, dev := range devices {
				deviceNames[i+1] = dev.Name
			}

			currentDeviceIdx := 0
			if t.config.RecordingDeviceName != "" {
				for i, name := range deviceNames {
					if name == t.config.RecordingDeviceName {
						currentDeviceIdx = i
						break
					}
				}
			}

			dropdown := tview.NewDropDown().
				SetLabel("録音デバイス: ").
				SetOptions(deviceNames, nil).
				SetCurrentOption(currentDeviceIdx)

			dropdown.SetBorder(true).
				SetTitle(" 録音デバイスを選択 ").
				SetTitleAlign(tview.AlignCenter)

			dropdown.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEscape {
					closeEditDialog()
					return nil
				}
				if event.Key() == tcell.KeyEnter {
					idx, deviceName := dropdown.GetCurrentOption()
					if idx == 0 {
						t.config.RecordingDeviceName = ""
						recordingList.SetItemText(0, "録音デバイス", "デフォルト")
					} else {
						t.config.RecordingDeviceName = deviceName
						recordingList.SetItemText(0, "録音デバイス", deviceName)
					}
					closeEditDialog()
					return nil
				}
				return event
			})

			showEditDialog("録音デバイス", dropdown)

		case 1: // Recording Mode
			list := tview.NewList().ShowSecondaryText(false)
			list.AddItem("シングル録音（マイクのみ）", "", '1', nil)
			list.AddItem("デュアル録音（システム音声+マイク）", "", '2', nil)

			// Set current selection
			if t.config.DualRecordingEnabled {
				list.SetCurrentItem(1)
			} else {
				list.SetCurrentItem(0)
			}

			list.SetBorder(true).
				SetTitle(" 録音モードを選択 ").
				SetTitleAlign(tview.AlignCenter)

			list.SetSelectedFunc(func(idx int, text, secondary string, r rune) {
				t.config.DualRecordingEnabled = (idx == 1)
				modeDisplay := "シングル"
				if t.config.DualRecordingEnabled {
					modeDisplay = "デュアル"
				}
				recordingList.SetItemText(1, "録音モード", modeDisplay)
				closeEditDialog()
			})

			list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEscape {
					closeEditDialog()
					return nil
				}
				return event
			})

			showEditDialog("録音モード", list)

		case 2: // System Volume
			list := tview.NewList().ShowSecondaryText(false)
			volumeOptions := []string{"-2", "-1", "0", "+1", "+2"}
			for _, vol := range volumeOptions {
				list.AddItem(vol, "", 0, nil)
			}

			// Set current selection
			currentIdx := volumeFloatToIndex(t.config.SystemAudioVolume)
			if currentIdx >= 0 && currentIdx < 5 {
				list.SetCurrentItem(currentIdx)
			}

			list.SetBorder(true).
				SetTitle(" システム音声音量を選択 ").
				SetTitleAlign(tview.AlignCenter)

			list.SetSelectedFunc(func(idx int, text, secondary string, r rune) {
				t.config.SystemAudioVolume = volumeIndexToSystemFloat(idx)
				recordingList.SetItemText(2, "システム音声音量", fmt.Sprintf("%+d", idx-2))
				closeEditDialog()
			})

			list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEscape {
					closeEditDialog()
					return nil
				}
				return event
			})

			showEditDialog("システム音声音量", list)

		case 3: // Mic Volume
			list := tview.NewList().ShowSecondaryText(false)
			volumeOptions := []string{"-2", "-1", "0", "+1", "+2"}
			for _, vol := range volumeOptions {
				list.AddItem(vol, "", 0, nil)
			}

			// Set current selection
			currentIdx := volumeFloatToIndex(t.config.MicrophoneVolume)
			if currentIdx >= 0 && currentIdx < 5 {
				list.SetCurrentItem(currentIdx)
			}

			list.SetBorder(true).
				SetTitle(" マイク音量を選択 ").
				SetTitleAlign(tview.AlignCenter)

			list.SetSelectedFunc(func(idx int, text, secondary string, r rune) {
				t.config.MicrophoneVolume = volumeIndexToMicFloat(idx)
				recordingList.SetItemText(3, "マイク音量", fmt.Sprintf("%+d", idx-2))
				closeEditDialog()
			})

			list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEscape {
					closeEditDialog()
					return nil
				}
				return event
			})

			showEditDialog("マイク音量", list)
		}
	})

	// Save function
	saveConfig := func() {
		// Configuration is already saved in real-time during editing

		// Save to file
		if err := config.SaveConfig(t.config, "config.json"); err != nil {
			// Show error message temporarily
			t.statusBar.SetText(fmt.Sprintf("[red]設定の保存に失敗: %v[white]", err))
			time.AfterFunc(3*time.Second, func() {
				t.app.QueueUpdateDraw(func() {
					t.updateStatusBar()
				})
			})
		} else {
			// Show success message temporarily
			t.statusBar.SetText("[green]設定を保存しました[white]")
			time.AfterFunc(2*time.Second, func() {
				t.app.QueueUpdateDraw(func() {
					t.updateStatusBar()
				})
			})
		}

		// Close dialog
		t.app.SetRoot(t.mainFlex, true)
	}

	// Handle Enter key on menu items - move focus to right side
	categoryList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		switch index {
		case 0:
			t.app.SetFocus(basicList)
		case 1:
			t.app.SetFocus(dirList)
		case 2:
			t.app.SetFocus(llmList)
		case 4:
			// Save
			saveConfig()
		case 5:
			// Cancel
			t.app.SetRoot(t.mainFlex, true)
		}
	})

	// Create split layout (left: menu, right: content)
	splitFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(categoryList, 22, 0, true).
		AddItem(contentArea, 0, 1, false)

	// Create help bar
	helpBar := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[yellow]↑↓/j/k[white]:移動 [yellow]←→[white]:左右移動 [yellow]Enter[white]:編集/決定 [yellow]s[white]:保存 [yellow]q/Esc[white]:閉じる")
	helpBar.SetBorder(false)

	// Populate main container
	mainContainer = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(splitFlex, 0, 1, true).
		AddItem(helpBar, 1, 0, false)

	// Handle keyboard shortcuts on left side
	categoryList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j', 'J':
				return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
			case 'k', 'K':
				return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
			case 's', 'S':
				saveConfig()
				return nil
			case 'q', 'Q':
				t.app.SetRoot(t.mainFlex, true)
				return nil
			}
		case tcell.KeyEscape:
			t.app.SetRoot(t.mainFlex, true)
			return nil
		case tcell.KeyRight:
			// Move to right side
			currentIndex := categoryList.GetCurrentItem()
			switch currentIndex {
			case 0:
				t.app.SetFocus(basicList)
			case 1:
				t.app.SetFocus(dirList)
			case 2:
				t.app.SetFocus(llmList)
			}
			return nil
		}
		return event
	})

	// Handle keyboard on right side lists
	listInputCapture := func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyLeft:
			// Move back to left menu
			t.app.SetFocus(categoryList)
			return nil
		case tcell.KeyEscape:
			// Close dialog
			t.app.SetRoot(t.mainFlex, true)
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j', 'J':
				return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
			case 'k', 'K':
				return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
			case 's', 'S':
				saveConfig()
				return nil
			case 'q', 'Q':
				t.app.SetRoot(t.mainFlex, true)
				return nil
			}
		}
		// Allow up/down arrows for list navigation
		return event
	}

	basicList.SetInputCapture(listInputCapture)
	dirList.SetInputCapture(listInputCapture)
	llmList.SetInputCapture(listInputCapture)

	// Show dialog
	t.app.SetRoot(mainContainer, true).SetFocus(categoryList)
}

