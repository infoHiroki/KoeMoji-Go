# TUIライブラリ評価ドキュメント

**作成日**: 2025-01-27
**対象バージョン**: v1.9.0+
**目的**: KoeMoji-Go macOS版のTUI強化のためのライブラリ選定

---

## 📋 背景

### 現状の課題

KoeMoji-Go macOS版は、Apple Developer Program年間費用（約12,000円）を避けるため、当面はターミナルからの起動が必要です。この制約を活かし、**リッチなTUI（Terminal User Interface）体験**を提供することで、GUI版に劣らない使いやすさを実現します。

### 現在のTUI実装

`internal/ui/ui.go`で実装されている既存TUIは以下の特徴があります：

**技術仕様:**
- `fmt.Printf`による単純な描画
- `\033[2J\033[H`で画面クリア→全再描画（チラつき）
- ANSIカラーコード対応
- 絵文字使用（🟢🟡📁🔴⏰）
- 12行固定のログバッファ
- 1文字コマンド入力（`bufio.Reader`）

**制限事項:**
- スクロール不可
- リアルタイム更新時にチラつき
- マウス操作不可
- ウィジェット/レイアウトシステムなし
- 複雑なUI構造の構築が困難

---

## 🔍 Go TUIライブラリの比較

### 全6ライブラリの評価マトリックス

| 項目 | tview | Bubble Tea | termui | gocui | pterm | promptui |
|------|-------|------------|--------|-------|-------|----------|
| **スター数** | 11.3k | 29.2k | 13.2k | 10.4k | 5.1k | 6.1k |
| **使用プロジェクト** | 3,423 | 10,000+ | 多数 | 1,100+ | 2,263 | 3,430 |
| **カテゴリ** | フルTUI | フルTUI | ダッシュボード | ビュー管理 | CLI強化 | プロンプト |
| **アーキテクチャ** | ウィジェット | 関数型（Elm） | ウィジェット | ビューベース | プリンター | プロンプト |
| **学習曲線** | 🟢 低い | 🟡 中程度 | 🟢 低い | 🟡 中程度 | 🟢 非常に低い | 🟢 非常に低い |
| **ウィジェット** | ⭐⭐⭐ 豊富 | ⭐⭐ Bubbles | ⭐⭐⭐ 13種類 | ⭐ ビューのみ | ⭐⭐ 30+ | ⭐ 2種類 |
| **スタイリング** | 組み込み | Lipgloss統合 | 組み込み | カスタム | ⭐⭐⭐ 豊富 | 基本的 |
| **マウス対応** | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| **日本語対応** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **インラインモード** | ❌ フル画面のみ | ✅ 対応 | ❌ フル画面のみ | ❌ フル画面のみ | ✅ 対応 | ✅ 対応 |
| **保守性** | ⭐⭐⭐ 活発 | ⭐⭐⭐ 非常に活発 | ⭐ 停滞（2021～） | ⭐⭐ やや停滞 | ⭐⭐⭐ 活発 | ⭐⭐ 停滞（2021～） |
| **用途** | 複雑なTUIアプリ | モダンTUIアプリ | 監視ダッシュボード | カスタムTUI | CLI出力強化 | 対話的入力 |
| **KoeMoji-Go適合度** | ⭐⭐⭐ 最適 | ⭐⭐ 良い | ⭐ 限定的 | ⭐⭐ 可能 | ❌ 不適 | ❌ 不適 |

### 実績あるアプリケーション

#### tviewを使用
- **K9s** - Kubernetes CLI管理ツール（最も有名）
- **lazysql** - データベース管理ツール
- **podman-tui** - コンテナ管理インターフェース

#### Bubble Teaを使用
- **gh-dash** - GitHub CLI拡張（GitHub公式）
- **glow** - マークダウンリーダー
- **chezmoi** - dotfilesマネージャー
- **Microsoft Azure Aztfy** - Azure移行ツール
- **AWS EKS Node Viewer** - AWS監視ツール

---

## 🎨 tview - ウィジェットベースTUI

### 概要

tviewは**tcellをベースにした高レベルウィジェットライブラリ**で、GUIフレームワークに近い感覚で使えます。

### 主要ウィジェット

| ウィジェット | 説明 | KoeMoji-Goでの用途 |
|-------------|------|-------------------|
| **TextView** | スクロール可能テキスト表示 | ログビューア（色付き、1000行保持） |
| **Table** | テーブル表示 | ファイル一覧（名前/サイズ/状態） |
| **Form** | フォーム入力 | 設定画面（Whisperモデル選択等） |
| **List** | リスト選択 | 録音デバイス選択 |
| **Modal** | モーダルダイアログ | 録音中終了警告 |
| **Flex/Grid** | レイアウトマネージャー | 3ペイン構成（ステータス/ログ/コマンド） |

### コード例：基本構造

```go
package main

import (
    "github.com/rivo/tview"
    "github.com/gdamore/tcell/v2"
)

func main() {
    app := tview.NewApplication()

    // ステータスパネル
    statusView := tview.NewTextView().
        SetDynamicColors(true).
        SetText("🟢 稼働中 | 待機: 0")
    statusView.SetBorder(true).SetTitle(" KoeMoji-Go v1.8.0 ")

    // ログビューア（スクロール可能）
    logView := tview.NewTextView().
        SetDynamicColors(true).
        SetScrollable(true).
        SetMaxLines(1000)
    logView.SetBorder(true).SetTitle(" ログ [↑↓でスクロール] ")

    // コマンドパネル
    commandView := tview.NewTextView().
        SetDynamicColors(true).
        SetTextAlign(tview.AlignCenter).
        SetText("[yellow]F1[white]設定 [yellow]F2[white]ログ [yellow]F3[white]スキャン")
    commandView.SetBorder(true).SetTitle(" コマンド ")

    // レイアウト（3行構成）
    layout := tview.NewFlex().
        SetDirection(tview.FlexRow).
        AddItem(statusView, 5, 0, false).      // 固定高さ5行
        AddItem(logView, 0, 1, true).          // 可変高さ（メイン）
        AddItem(commandView, 3, 0, false)      // 固定高さ3行

    // キーバインド
    app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        switch event.Key() {
        case tcell.KeyF1:
            // 設定画面を表示
            return nil
        case tcell.KeyF3:
            // スキャン実行
            return nil
        case tcell.KeyRune:
            if event.Rune() == 'q' {
                app.Stop()
                return nil
            }
        }
        return event
    })

    // 実行
    if err := app.SetRoot(layout, true).Run(); err != nil {
        panic(err)
    }
}
```

### リアルタイム更新

```go
// 別のgoroutineから安全にUI更新
app.QueueUpdateDraw(func() {
    statusView.SetText("🟡 処理中 | 待機: 3")
    logView.Clear()
    fmt.Fprintf(logView, "[blue]情報[white] 15:04:05 処理開始\n")
})
```

### 設定画面（モーダルフォーム）

```go
form := tview.NewForm().
    AddDropDown("Whisperモデル", []string{"tiny", "base", "large-v3"}, 0, nil).
    AddInputField("スキャン間隔（分）", "5", 10, nil, nil).
    AddCheckbox("AI要約を有効化", false, nil).
    AddButton("保存", func() {
        // 設定保存処理
        pages.SwitchToPage("main")
    }).
    AddButton("キャンセル", func() {
        pages.SwitchToPage("main")
    })

// モーダル表示（画面中央）
modal := tview.NewFlex().
    AddItem(nil, 0, 1, false).
    AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
        AddItem(nil, 0, 1, false).
        AddItem(form, 20, 1, true).
        AddItem(nil, 0, 1, false), 60, 1, true).
    AddItem(nil, 0, 1, false)

pages.AddPage("config", modal, true, false)
```

### メリット・デメリット

**メリット:**
- ✅ GUIフレームワークに近い感覚で学習コスト低い
- ✅ 豊富なウィジェットがすぐ使える
- ✅ K9sなど実績豊富
- ✅ ドキュメント・サンプルが充実
- ✅ 複雑なレイアウトも簡単

**デメリット:**
- ❌ OOP的な設計でGo的ではない部分がある
- ❌ カスタマイズ性はBubble Teaより低い
- ❌ アニメーション機能は限定的

---

## 🫧 Bubble Tea - 関数型TUI

### 概要

Bubble Teaは**Elm Architectureに基づく関数型TUIフレームワーク**で、関数型プログラミングの思想を取り入れています。

### アーキテクチャ（MVU: Model-View-Update）

```go
type Model struct {
    status      string
    queueCount  int
    logs        []string
    cursor      int
}

func (m Model) Init() tea.Cmd {
    return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q":
            return m, tea.Quit
        case "s":
            // スキャン実行
            m.logs = append(m.logs, "スキャンを実行しました")
            return m, nil
        }
    case tickMsg:
        // 定期更新
        m.status = "稼働中"
        return m, tick()
    }
    return m, nil
}

func (m Model) View() string {
    s := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#00FF00")).
        Render("🟢 " + m.status) + "\n\n"

    for _, log := range m.logs {
        s += log + "\n"
    }

    return s
}
```

### Lipgloss統合（スタイリング）

```go
var (
    // CSS風のスタイル定義
    headerStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FAFAFA")).
        Background(lipgloss.Color("#7D56F4")).
        Padding(1, 2).
        Bold(true)

    logStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("#874BFD")).
        Padding(1, 2).
        Width(80)
)

func (m Model) View() string {
    header := headerStyle.Render("KoeMoji-Go v1.8.0")
    logs := logStyle.Render(strings.Join(m.logs, "\n"))

    return lipgloss.JoinVertical(lipgloss.Left, header, logs)
}
```

### Bubblesコンポーネント

公式が提供する再利用可能コンポーネント：

- **textarea**: 複数行テキスト入力
- **textinput**: 1行テキスト入力
- **list**: 選択可能リスト
- **table**: テーブル表示
- **progress**: プログレスバー
- **spinner**: ローディングアニメーション
- **viewport**: スクロール可能ビュー

```go
import "github.com/charmbracelet/bubbles/list"

type Model struct {
    list list.Model
}

func (m Model) Init() tea.Cmd {
    items := []list.Item{
        item{title: "recording_001.wav", desc: "処理待ち"},
        item{title: "recording_002.wav", desc: "処理中"},
    }
    m.list = list.New(items, list.NewDefaultDelegate(), 0, 0)
    return nil
}
```

### メリット・デメリット

**メリット:**
- ✅ 関数型プログラミングで状態管理が明確
- ✅ テスタビリティが高い（純粋関数）
- ✅ カスタマイズ性が非常に高い
- ✅ アニメーション・エフェクト対応
- ✅ 最新のGo TUIトレンド
- ✅ Lipglossによる美しいスタイリング
- ✅ Microsoft、AWS等大手企業採用

**デメリット:**
- ❌ 学習曲線がやや急（Elm Architectureの理解が必要）
- ❌ ボイラープレート多め
- ❌ 簡単なUIでも記述量が多い

---

## 🎨 termui - ダッシュボード特化TUI

### 概要

termuiは**termbox-goをベースにしたダッシュボード特化型ライブラリ**で、監視ツールやシステムメトリクス表示に最適化されています。

### 主要ウィジェット（13種類）

| ウィジェット | 説明 |
|-------------|------|
| **BarChart, StackedBarChart** | 棒グラフ（単一・積み上げ） |
| **Canvas** | ブレイル点描画 |
| **Gauge** | ゲージメーター |
| **Image** | 画像表示 |
| **List, Tree** | リスト・ツリー表示 |
| **Paragraph** | テキスト段落 |
| **PieChart** | 円グラフ |
| **Plot** | 散布図・折れ線グラフ |
| **Sparkline** | スパークライン |
| **Table** | テーブル表示 |
| **Tabs** | タブ切り替え |

### レイアウトシステム

```go
grid := ui.NewGrid()
grid.SetRect(0, 0, termWidth, termHeight)
grid.Set(
    ui.NewRow(1.0/2,
        ui.NewCol(1.0/2, gaugeWidget),
        ui.NewCol(1.0/2, sparklineWidget),
    ),
    ui.NewRow(1.0/2,
        ui.NewCol(1.0, listWidget),
    ),
)
```

### 実績アプリケーション

- **gotop** - システムモニター
- **dockdash** - Dockerダッシュボード
- **updo** - タスク管理ツール

### メリット・デメリット

**メリット:**
- ✅ グラフ・チャート機能が豊富
- ✅ 12列グリッドレイアウトが便利
- ✅ ダッシュボードUI向き

**デメリット:**
- ❌ メンテナンス停滞（2021年以降）
- ❌ フォーム入力機能が弱い
- ❌ KoeMoji-Goには不要な機能が多い

---

## 🪟 gocui - ミニマリストビュー管理

### 概要

gocuiは**ミニマリストなビュー管理ライブラリ**で、ウィジェットではなく「ビュー」を組み合わせてUIを構築します。

### 主要機能

- **ビュー（View）** - `io.ReadWriter`実装の基本ウィンドウ
- **重なり合うビュー** - 柔軟なレイアウト
- **キーバインド** - グローバル・ビュー単位
- **マウス対応** - クリック・ドラッグ
- **ランタイム変更** - UIを動的に変更可能

### コード例

```go
func layout(g *gocui.Gui) error {
    maxX, maxY := g.Size()

    // ログビュー
    if v, err := g.SetView("logs", 0, 0, maxX-1, maxY-5); err != nil {
        if err != gocui.ErrUnknownView {
            return err
        }
        v.Title = "Logs"
        v.Autoscroll = true
    }

    // コマンドビュー
    if v, err := g.SetView("command", 0, maxY-4, maxX-1, maxY-1); err != nil {
        if err != gocui.ErrUnknownView {
            return err
        }
        v.Title = "Commands"
    }

    return nil
}
```

### メリット・デメリット

**メリット:**
- ✅ シンプルなAPI
- ✅ 細かい制御が可能
- ✅ 1,100+プロジェクトで使用

**デメリット:**
- ❌ ウィジェットがない（自作必要）
- ❌ メンテナンスやや停滞
- ❌ 学習コストやや高い

---

## 🎨 pterm - CLI出力強化ライブラリ

### 概要

ptermは**美しいCLI出力に特化したライブラリ**で、フルTUIアプリではなく、コマンドラインツールの出力を美しくすることに焦点を当てています。

### 主要コンポーネント（30+）

**表示系:**
- Area, BarChart, BigText, Box, BulletList, Header, Heatmap, Panel, Paragraph, Section, Table, Tree

**プログレス系:**
- ProgressBar, Spinner, Logger

**インタラクティブ系:**
- InteractiveConfirm（Yes/No確認）
- InteractiveSelect（単一選択）
- InteractiveMultiselect（複数選択）
- InteractiveTextInput（テキスト入力）

### コード例

```go
// 美しいテーブル表示
tableData := pterm.TableData{
    {"Name", "Size", "Status"},
    {"recording_001.wav", "5.2MB", "処理中"},
    {"recording_002.wav", "3.1MB", "完了"},
}
pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()

// プログレスバー
progressbar, _ := pterm.DefaultProgressbar.WithTotal(100).Start()
for i := 0; i < 100; i++ {
    progressbar.Increment()
    time.Sleep(50 * time.Millisecond)
}
```

### メリット・デメリット

**メリット:**
- ✅ 非常に美しい出力
- ✅ 学習コスト超低い
- ✅ 30+コンポーネント
- ✅ インタラクティブ機能あり

**デメリット:**
- ❌ フルTUIアプリには不向き
- ❌ スクロール可能ビューなし
- ❌ 常時表示のダッシュボードに不適

---

## 💬 promptui - プロンプト特化ライブラリ

### 概要

promptuiは**対話的なプロンプト入力に特化したライブラリ**で、ユーザーからの入力を取得するためのシンプルなインターフェースを提供します。

### 2つのモード

**1. Promptモード（1行入力）:**
```go
prompt := promptui.Prompt{
    Label:    "録音デバイス名",
    Default:  "デフォルトデバイス",
    Validate: validateDeviceName,
}
result, _ := prompt.Run()
```

**2. Selectモード（選択）:**
```go
devices := []string{"デフォルトデバイス", "マイク (USB)", "システム音声"}
prompt := promptui.Select{
    Label: "録音デバイスを選択",
    Items: devices,
}
_, result, _ := prompt.Run()
```

### メリット・デメリット

**メリット:**
- ✅ 超シンプルなAPI
- ✅ Cobra/urfave/cli統合
- ✅ パスワードマスク対応
- ✅ 検索・ページング対応

**デメリット:**
- ❌ フルTUIアプリには不向き
- ❌ メンテナンス停滞（2021年以降）
- ❌ KoeMoji-Goには限定的

---

## 🔍 重要な技術的違い

### インライン vs フルスクリーンモード

これは**Bubble Tea vs tview**を選ぶ上で最も重要な違いです：

| ライブラリ | インライン | フルスクリーン | 影響 |
|-----------|-----------|--------------|------|
| **Bubble Tea** | ✅ 対応 | ✅ 対応 | ターミナル履歴と共存可能 |
| **tview** | ❌ 非対応 | ✅ 対応のみ | 画面全体を占有 |
| **termui** | ❌ 非対応 | ✅ 対応のみ | 画面全体を占有 |
| **gocui** | ❌ 非対応 | ✅ 対応のみ | 画面全体を占有 |
| **pterm** | ✅ 対応 | ❌ 非対応 | 通常CLI出力のみ |
| **promptui** | ✅ 対応 | ❌ 非対応 | プロンプト入力のみ |

**インラインモードの例（Bubble Tea）:**
```
$ koemoji-go --tui
KoeMoji-Goを開始しました...
🟢 稼働中 | 待機: 0
[INFO] 15:04:05 設定を読み込みました
↑ この後も新しいコマンドを入力できる
$ ls
```

**フルスクリーンモードの例（tview）:**
```
┏━━━━━━━━━━━━━━━━━━━━┓
┃ KoeMoji-Go v1.9.0  ┃
┃ （画面全体を占有）   ┃
┗━━━━━━━━━━━━━━━━━━━━┛
```

### 開発者の実体験（2025年）

Hacker Newsのディスカッション（2024年9月）より：

> 「シンプルなニーズ（動的リストビュー、ファイルピッカー）のために、Bubble Teaを30分触ってみたが放棄してtviewに切り替えた。tviewなら即座に動くものが作れた。」

**KoeMoji-Goへの影響:**

- ✅ **フルスクリーンアプリとして使う** → tviewで問題なし
- ❌ **ターミナル履歴と共存させたい** → Bubble Tea必須

KoeMoji-Goは**常駐型ダッシュボード**なので、フルスクリーンモードで十分です。

---

## 📊 KoeMoji-Goへの適用比較

### シナリオ別評価

| 要件 | tview | Bubble Tea | 推奨 |
|------|-------|------------|------|
| スクロール可能ログビューア | ⭐⭐⭐ 簡単 | ⭐⭐ Viewport使用 | tview |
| ファイル一覧テーブル | ⭐⭐⭐ Table標準 | ⭐⭐ Bubblesのtable | tview |
| 設定フォーム | ⭐⭐⭐ Form標準 | ⭐ 自作必要 | tview |
| リアルタイム更新 | ⭐⭐⭐ QueueUpdateDraw | ⭐⭐⭐ tick Cmd | 同等 |
| 録音中UI（動的更新） | ⭐⭐⭐ 簡単 | ⭐⭐⭐ spinnerで美しい | 同等 |
| 進捗表示 | ⭐⭐ 自作 | ⭐⭐⭐ progress標準 | Bubble Tea |
| モーダルダイアログ | ⭐⭐⭐ Modal標準 | ⭐⭐ Pagesで実装 | tview |
| 美しいスタイリング | ⭐⭐ 基本的 | ⭐⭐⭐ Lipgloss | Bubble Tea |
| 学習コスト | ⭐⭐⭐ 低い | ⭐⭐ 中程度 | tview |
| 保守性 | ⭐⭐⭐ 構造化容易 | ⭐⭐⭐ 関数型で明確 | 同等 |

---

## 🎯 推奨事項

### 最終推奨: **tview**（6ライブラリ比較後）

**調査したライブラリ:**
1. ✅ **tview** - フルTUI（推奨）
2. ⭐ **Bubble Tea** - フルTUI（将来検討）
3. ❌ **termui** - ダッシュボード特化（不要な機能多い）
4. ❌ **gocui** - ミニマリスト（ウィジェット自作必要）
5. ❌ **pterm** - CLI出力強化（フルTUI不向き）
6. ❌ **promptui** - プロンプト特化（用途限定的）

**tviewを選ぶ理由:**

1. **即座に使える豊富なウィジェット**
   - KoeMoji-Goが必要とする全ウィジェット（TextView, Table, Form, Modal）が標準で揃っている
   - Bubble Teaは美しいが、フォーム・テーブルを自作する手間がかかる
   - termuiはグラフ機能が充実しているが、KoeMoji-Goには不要
   - gocuiは全てを自作する必要があり、開発コストが高い

2. **K9sの実績**
   - Kubernetes管理という複雑なドメインで成功している
   - KoeMoji-Goのユースケース（ステータス表示、ファイル一覧、設定）に類似

3. **学習コスト**
   - GUIフレームワークに近い感覚で、既存のFyne GUI実装者が理解しやすい
   - Elm Architectureを学ぶ必要がない（Bubble Teaとの違い）
   - ptermやpromptuiは学習コストは低いが、フルTUIアプリには不向き

4. **開発速度**
   - プロトタイプを素早く作れる
   - v1.9.0で早期リリース可能
   - 実体験: 「Bubble Teaで30分苦労→tviewで即座に完成」（Hacker News 2024）

5. **フルスクリーンモードで十分**
   - KoeMoji-Goは**常駐型ダッシュボード**なので、画面全体を占有するフルスクリーンモードで問題なし
   - Bubble Teaの**インラインモード**は魅力的だが、本プロジェクトには不要

6. **活発なメンテナンス**
   - tview: ⭐⭐⭐ 活発（2025年も更新継続）
   - termui: ⭐ 停滞（2021年以降更新なし）
   - promptui: ⭐⭐ 停滞（2021年以降更新なし）

### 各ライブラリが優れる点

| ライブラリ | 優れる点 | KoeMoji-Goでの必要性 |
|-----------|---------|---------------------|
| **tview** | 豊富なウィジェット、即戦力 | ⭐⭐⭐ 最適 |
| **Bubble Tea** | 美しいスタイリング、関数型、インラインモード | ⭐⭐ 将来v2.0で検討 |
| **termui** | グラフ・チャート機能 | ❌ 不要 |
| **gocui** | 細かい制御 | ❌ オーバーキル |
| **pterm** | 美しいCLI出力 | ❌ フルTUI不向き |
| **promptui** | シンプルなプロンプト | ❌ 用途限定的 |

### Bubble Teaを将来検討する理由

- ✨ **Lipgloss統合** - CSS風の美しいスタイリング
- 🎬 **アニメーション** - spinner、progressバーのエフェクト
- 🏗️ **関数型設計** - 状態管理が明確、テスタビリティ高い
- 🏢 **大手企業採用** - Microsoft、AWS、GitHub等の実績
- 📱 **インラインモード** - 将来的にCLIモードと統合する可能性

→ **v2.0全面リライト時、またはユーザー要望が増えた際に再検討**

---

## 🚀 実装ロードマップ

### Phase 1: プロトタイプ（v1.9.0-alpha）

**期間**: 3日
**目標**: 基本的なtview TUIの動作確認

```bash
# 依存関係追加
go get github.com/rivo/tview
go get github.com/gdamore/tcell/v2
```

**実装内容:**
- `internal/ui/tui_rich.go` 新規作成
- 3ペインレイアウト（ステータス/ログ/コマンド）
- 基本的なキーバインド（F1-F4, q）
- リアルタイムログ表示（スクロール対応）

**フラグ:**
```bash
./koemoji-go --tui           # 既存TUI（後方互換）
./koemoji-go --tui-rich      # 新tview TUI
```

### Phase 2: 機能追加（v1.9.0-beta）

**期間**: 5日
**目標**: GUI版と同等の機能

**実装内容:**
- ファイル一覧テーブル（Table）
- 設定画面（Form）
- 録音デバイス選択（DropDown）
- モーダルダイアログ（録音中終了警告）
- ヘルプ画面（Modal）

### Phase 3: 安定化（v1.9.0）

**期間**: 2日
**目標**: 正式リリース

**実装内容:**
- 既存TUIをtview版に置き換え
- `--tui-rich`フラグ削除（標準化）
- ドキュメント更新
- 手動テスト

### Phase 4: 拡張（v1.10.0以降）

**将来機能:**
- ファイル進捗表示（処理中ファイルの%表示）
- キーボードショートカットカスタマイズ
- テーマ切り替え（ダーク/ライト）
- ログフィルタリング（INFO/ERROR等で絞り込み）

---

## 📚 参考リンク

### tview（推奨）
- **GitHub**: https://github.com/rivo/tview
- **ドキュメント**: https://pkg.go.dev/github.com/rivo/tview
- **デモ**: https://github.com/rivo/tview/tree/master/demos
- **Wiki**: https://github.com/rivo/tview/wiki
- **スター**: 11.3k | **使用**: 3,423プロジェクト

### Bubble Tea（将来検討）
- **GitHub**: https://github.com/charmbracelet/bubbletea
- **ドキュメント**: https://pkg.go.dev/github.com/charmbracelet/bubbletea
- **チュートリアル**: https://github.com/charmbracelet/bubbletea/tree/master/tutorials
- **Bubbles**: https://github.com/charmbracelet/bubbles（コンポーネント集）
- **Lipgloss**: https://github.com/charmbracelet/lipgloss（スタイリング）
- **スター**: 29.2k | **使用**: 10,000+プロジェクト

### termui（調査済み・不採用）
- **GitHub**: https://github.com/gizak/termui
- **ドキュメント**: https://pkg.go.dev/github.com/gizak/termui
- **スター**: 13.2k | **状態**: ⚠️ メンテナンス停滞（2021～）
- **理由**: ダッシュボード特化、KoeMoji-Goには不要な機能が多い

### gocui（調査済み・不採用）
- **GitHub**: https://github.com/jroimartin/gocui
- **ドキュメント**: https://pkg.go.dev/github.com/jroimartin/gocui
- **スター**: 10.4k | **使用**: 1,100+プロジェクト
- **理由**: ウィジェット自作必要、開発コスト高い

### pterm（調査済み・不採用）
- **GitHub**: https://github.com/pterm/pterm
- **ドキュメント**: https://pkg.go.dev/github.com/pterm/pterm
- **公式サイト**: https://pterm.sh/
- **スター**: 5.1k | **使用**: 2,263プロジェクト
- **理由**: CLI出力強化特化、フルTUIアプリ不向き

### promptui（調査済み・不採用）
- **GitHub**: https://github.com/manifoldco/promptui
- **ドキュメント**: https://pkg.go.dev/github.com/manifoldco/promptui
- **スター**: 6.1k | **使用**: 3,430パッケージ
- **状態**: ⚠️ メンテナンス停滞（2021～）
- **理由**: プロンプト入力特化、用途限定的

---

### 実装例

**tview使用プロジェクト:**
- **K9s**: https://github.com/derailed/k9s（Kubernetes管理）
- **lazysql**: https://github.com/jorgerojas26/lazysql（データベース管理）
- **podman-tui**: https://github.com/containers/podman-tui（コンテナ管理）

**Bubble Tea使用プロジェクト:**
- **gh-dash**: https://github.com/dlvhdr/gh-dash（GitHub CLI拡張）
- **glow**: https://github.com/charmbracelet/glow（マークダウンリーダー）
- **chezmoi**: https://github.com/twpayne/chezmoi（dotfilesマネージャー）

**その他の参考リソース:**
- **Go TUIライブラリ比較**: https://leapcell.io/blog/exploring-tui-libraries-in-go
- **awesome-tuis**: https://github.com/rothgar/awesome-tuis（TUIアプリ一覧）

---

## 🔧 技術スタック（v1.9.0予定）

```
KoeMoji-Go TUI Stack:
├─ tview (v0.0.0-latest)
│  └─ tcell/v2 (v2.x)
├─ gordonklaus/portaudio (録音)
├─ go-ole/go-ole (Windows VoiceMeeter)
└─ 既存internal/ui/ui.go (段階的置き換え)
```

---

## 📝 調査履歴

| 日付 | 調査内容 | 結果 |
|------|---------|------|
| 2025-01-27 | tview, Bubble Tea, tcell初期調査 | tview推奨（3ライブラリ比較） |
| 2025-01-27 | termui, gocui, pterm, promptui追加調査 | tview推奨維持（6ライブラリ比較完了） |

## 🔄 変更履歴

- **v1.0** (2025-01-27): 初版作成（tview/Bubble Tea/tcell比較）
- **v2.0** (2025-01-27): 6ライブラリに拡張（termui/gocui/pterm/promptui追加）
  - インライン vs フルスクリーンモードの重要性を追記
  - 各ライブラリの詳細セクション追加
  - 参考リンク大幅拡充

---

**作成者**: Claude + @infoHiroki
**レビュー**: 未
**ステータス**: ドラフト v2.0
**最終更新**: 2025-01-27
**調査ライブラリ数**: 6個（tview, Bubble Tea, termui, gocui, pterm, promptui）
**最終推奨**: **tview**
