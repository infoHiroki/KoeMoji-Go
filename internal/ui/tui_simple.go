package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/infoHiroki/KoeMoji-Go/internal/config"
	"github.com/rivo/tview"
)

// SimpleTUI represents a simple terminal UI for testing
// Phase 4: Custom key bindings (j/k/q/?)
type SimpleTUI struct {
	app         *tview.Application
	config      *config.Config
	menuList    *tview.List
	statusBar   *tview.TextView
	helpBar     *tview.TextView
	contentArea *tview.TextView
	mainFlex    *tview.Flex
}

// NewSimpleTUI creates a new simple TUI (Phase 4)
func NewSimpleTUI(cfg *config.Config) *SimpleTUI {
	app := tview.NewApplication()

	// Create status bar (top, 1 line)
	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[green]KoeMoji-Go Simple TUI[white] | Phase 4")
	statusBar.SetBorder(false)

	// Create menu list (left side, fixed width)
	list := tview.NewList().ShowSecondaryText(false)
	list.AddItem("1. 設定", "", 0, nil)
	list.AddItem("2. ログ", "", 0, nil)
	list.AddItem("3. 終了", "", 0, nil)

	list.SetBorder(true).
		SetTitle(" メニュー ").
		SetTitleAlign(tview.AlignCenter)

	// Create content area (right side, expands)
	contentArea := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[yellow]1. 設定[white]\n\n設定画面の内容がここに表示されます")
	contentArea.SetBorder(true).
		SetTitle(" コンテンツ ").
		SetTitleAlign(tview.AlignCenter)

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
		AddItem(statusBar, 1, 0, false).    // Fixed 1 line
		AddItem(middleFlex, 0, 1, true).    // Expand to fill
		AddItem(helpBar, 1, 0, false)       // Fixed 1 line

	// Handle cursor movement (Phase 3: update content on selection change)
	list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		switch index {
		case 0:
			contentArea.SetText("[yellow]1. 設定[white]\n\n設定画面の内容がここに表示されます\n\n• Whisperモデル\n• 入力/出力ディレクトリ\n• OpenAI API設定")
		case 1:
			contentArea.SetText("[yellow]2. ログ[white]\n\nログ画面の内容がここに表示されます\n\n• アプリケーションログ\n• 処理履歴\n• エラーメッセージ")
		case 2:
			contentArea.SetText("[yellow]3. 終了[white]\n\nEnterキーまたはqキーでアプリケーションを終了します")
		}
	})

	// Create SimpleTUI struct early to pass mainFlex to showHelpDialog
	tui := &SimpleTUI{
		app:         app,
		config:      cfg,
		menuList:    list,
		statusBar:   statusBar,
		helpBar:     helpBar,
		contentArea: contentArea,
		mainFlex:    mainFlex,
	}

	// Handle custom key bindings (Phase 4: j/k/q/?)
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
				showHelpDialog(app, mainFlex)
				return nil
			}
		}
		// Return event for default behavior (arrow keys, Enter, etc.)
		return event
	})

	// Handle Enter key selection
	list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		switch index {
		case 0:
			// 設定 - Phase 4では何もしない
		case 1:
			// ログ - Phase 4では何もしない
		case 2:
			// 終了
			app.Stop()
		}
	})

	return tui
}

// Run starts the simple TUI
func (t *SimpleTUI) Run() error {
	return t.app.SetRoot(t.mainFlex, true).Run()
}

// Stop stops the simple TUI
func (t *SimpleTUI) Stop() {
	t.app.Stop()
}

// showHelpDialog shows a help dialog with key bindings
func showHelpDialog(app *tview.Application, mainFlex *tview.Flex) {
	helpText := `[yellow]KoeMoji-Go Simple TUI - ヘルプ[white]

[yellow]キー操作:[white]
  ↑ / k     : 上に移動
  ↓ / j     : 下に移動
  Enter     : 選択
  q         : 終了
  ?         : このヘルプを表示

[yellow]Phase 4の機能:[white]
  • 左右分割レイアウト
  • メニュー選択でコンテンツ自動更新
  • Vim風キーバインディング (j/k)
  • カスタムキー (q/?)

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
