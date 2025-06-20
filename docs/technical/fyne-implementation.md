# Fyne実装テクニカルノート

## 概要
KoeMoji-GoのGUI実装における技術的課題と解決策を記録したテクニカルノート。Fyneフレームワークを使用したクロスプラットフォームGUI開発での重要な知見を体系化。

## Fyneダイアログボタン制御の課題

### 問題の詳細
`dialog.NewCustom("設定", "", content, window)`で空文字を指定しても、Fyneが内部的にデフォルトの閉じるボタンを自動生成し、ダイアログの中央に小さなボタンとして表示される。

#### 症状
- 設定ダイアログの中央に謎の小さなグレーボタンが表示
- クリックするとダイアログが閉じる動作
- 独自ボタン（キャンセル、保存等）とは別に表示される

#### 根本原因
`dialog.NewCustom()`の第2引数（dismissText）に空文字`""`を指定しても、Fyneフレームワークが内部的に標準の閉じるボタンを生成する仕様のため。

#### 解決策
```go
// ❌ 問題のあるコード
cd.dialog = dialog.NewCustom("設定", "", content, cd.app.window)

// ✅ 修正後のコード  
cd.dialog = dialog.NewCustomWithoutButtons("設定", content, cd.app.window)
```

#### 学習事項
- Fyneダイアログで完全にカスタムボタンのみを使用したい場合は`NewCustomWithoutButtons()`を使用
- `NewCustom()`は標準ボタンが必ず生成される仕様
- 当て推量での修正ではなく、原因を特定してから適切な修正を実施することの重要性

## Fyneスクロールコンテナの適用範囲

### 技術的背景
Fyneの`container.NewScroll()`は**仕様として**コンテンツのMinSizeを縮小する。これはFyne公式ドキュメントにも明記されており、意図的な設計でありバグではない。

### 発生する問題
```go
// ❌ フォーム要素での問題例
form := container.NewVBox(
    widget.NewForm(...),
)
return container.NewScroll(form) // フォームが圧縮される
```

- **症状**: フォーム要素が一行ずつしか表示され、視認性が悪化
- **発生箇所**: 設定ダイアログ、複雑なフォームレイアウト

### 用途別解決策

#### A. フォーム要素（推奨：スクロール回避）
```go
// ✅ フォーム要素の場合
form := widget.NewForm(...)
return container.NewVBox(form) // スクロールを使わない
```

#### B. ログ表示（推奨：VScroll使用）
```go
// ✅ ログ表示の場合
logText := widget.NewRichTextWithText(logContent)
return container.NewVScroll(logText) // VScrollは適用可能
```

#### C. テキストエリア（推奨：Entry + Scroll）
```go
// ✅ 大量テキスト表示の場合
entry := widget.NewMultiLineEntry()
scrollContainer := container.NewScroll(entry)
scrollContainer.ScrollToBottom() // 新しい内容で下スクロール
```

## Fyne Preferences API設定

### 発生する警告
```
Fyne error: Preferences API requires a unique ID, use app.NewWithID()
```

### 解決策
```go
// ❌ 問題のあるコード
fyneApp := app.New()

// ✅ 修正後のコード
fyneApp := app.NewWithID("com.hirokitakamura.koemoji-go")
```

## GUI設定ダイアログアーキテクチャ

### 構造設計
```
ConfigDialog
├── Basic Settings Tab (基本設定)
│   ├── Whisper Model (Select)
│   ├── Language (Entry)
│   ├── UI Language (Select)
│   ├── Scan Interval (Entry + validation)
│   ├── Max CPU Percent (Entry + validation)
│   ├── Compute Type (Select)
│   ├── Use Colors (Check)
│   └── Output Format (Select)
├── Directory Settings Tab (ディレクトリ設定)
│   ├── Input Directory (Entry + Browse button)
│   ├── Output Directory (Entry + Browse button)
│   └── Archive Directory (Entry + Browse button)
└── LLM Settings Tab (LLM設定)
    ├── LLM Summary Enabled (Check)
    ├── LLM API Provider (Select)
    ├── LLM API Key (Password Entry)
    ├── LLM Model (Select)
    ├── LLM Max Tokens (Entry + validation)
    ├── Summary Language (Entry)
    └── Summary Prompt (MultiLine Entry)
```

### 重要な実装パターン

#### 1. 設定の一時保存パターン
```go
type ConfigDialog struct {
    config     *config.Config  // 元の設定
    tempConfig *config.Config  // 編集用の一時設定
    // ...
}

// 編集開始時に一時コピーを作成
cd.tempConfig = &config.Config{}
*cd.tempConfig = *app.config

// 保存時に元の設定に反映
*cd.config = *cd.tempConfig
```

#### 2. バリデーション統合パターン
```go
func (cd *ConfigDialog) saveConfig() {
    // バリデーション実行
    if err := cd.validateConfig(); err != nil {
        dialog.ShowError(err, cd.app.window)
        return
    }
    
    // 保存処理
    // ...
}
```

#### 3. ウィジェット状態管理パターン
```go
func (cd *ConfigDialog) updateLLMWidgetStates() {
    enabled := cd.tempConfig.LLMSummaryEnabled
    
    if enabled {
        cd.llmProviderSelect.Enable()
        cd.llmAPIKeyEntry.Enable()
        // ...
    } else {
        cd.llmProviderSelect.Disable()
        cd.llmAPIKeyEntry.Disable()
        // ...
    }
}
```

## ベストプラクティス

### 1. レイアウト設計原則
- **スクロールの使い分け**: フォーム要素では回避、ログ・テキスト表示では適用
- **Border layoutで適切な領域分割**
- **VBoxで自然な垂直配置**
- **ダイアログサイズは内容に応じて調整**

### 2. 設定管理ベストプラクティス
- **一時設定オブジェクトで編集**
- **保存時にバリデーション実行**
- **キャンセル時は変更を破棄**

### 3. ユーザビリティ向上
- **リアルタイムバリデーション**
- **エラーメッセージの適切な表示**
- **多言語対応の統一**

### 4. 状態管理
- **関連ウィジェットの連動**
- **設定変更時のUI更新**
- **ログ出力による操作記録**

## 参考実装例

### メインウィンドウレイアウト
```go
// ✅ Border layoutでの適切な領域分割
content := container.NewBorder(
    topSection,    // top - 固定ヘッダー
    bottomSection, // bottom - 固定ボタン
    nil,          // left
    nil,          // right
    logContent,   // center - 残りスペースをログに割り当て
)
```

## トラブルシューティング

### 問題: フォーム要素が見切れる
- **原因**: スクロールコンテナの不適切な使用
- **解決**: VBoxまたはBorder layoutに変更

### 問題: ダイアログが小さすぎる
- **原因**: デフォルトサイズが内容に対して不十分
- **解決**: `dialog.Resize(fyne.NewSize(width, height))`で調整

### 問題: ウィジェットが応答しない
- **原因**: 無効化されたウィジェットまたはイベントハンドラの未設定
- **解決**: Enable/Disableの状態確認、OnChangedハンドラの設定

## 設計原則まとめ

FyneでのGUI開発では、**コンテンツの性質に応じたスクロール適用**が最も重要：

1. **フォーム要素**: `container.NewVBox()`や`container.NewBorder()`を優先
2. **ログ・テキスト表示**: `container.NewVScroll()`や`widget.RichText`を活用
3. **大量データ**: `widget.List`や適切なスクロール対応ウィジェットを選択

Fyneのスクロールコンテナは「MinSize縮小」が仕様であることを理解し、用途に応じて適切に使い分けることで、ユーザビリティの高いGUIが実現できる。

## 更新履歴
- 2025-06-18: Fyneダイアログボタン制御の問題と解決策を記録
- 2025-06-20: スクロールコンテナの適用範囲とGUI設定ダイアログのアーキテクチャを統合