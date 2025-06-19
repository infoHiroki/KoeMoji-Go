# GUI実装テックノート

## Overview
KoeMoji-GoのGUI実装における技術的課題と解決策をまとめたテックノート。特にFyneフレームワークを使用したGUI開発での知見を記録。

## 問題1: Fyneスクロールコンテナの適用範囲

### 発生状況と原因分析
- **症状**: フォーム要素が一行ずつしか表示され、視認性が悪化
- **発生箇所**: 設定ダイアログ、複雑なフォームレイアウト
- **根本原因**: Fyneの`container.NewScroll()`は**仕様として**コンテンツのMinSizeを縮小する

### 技術的背景
```go
// ❌ フォーム要素での問題例
form := container.NewVBox(
    widget.NewForm(...),
)
return container.NewScroll(form) // フォームが圧縮される
```

**Fyne公式ドキュメントより**:
- `container.NewScroll()` は「may cause the MinSize to be smaller than that of the passed object」
- これは意図的な設計であり、バグではない

### 解決策（用途別）

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

### 設計原則の更新
1. **フォーム要素**: スクロール使用を避け、`container.NewVBox()`や`container.NewBorder()`を使用
2. **ログ・テキスト表示**: `container.NewVScroll()`や適切なウィジェット選択で対応
3. **用途に応じた使い分け**: コンテンツの性質に応じてスクロール適用を判断

## 問題2: Fyne Preferences API警告

### 発生状況
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

## GUI設定ダイアログ実装アーキテクチャ

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

### 1. レイアウト設計
- **スクロールの使い分け**: フォーム要素では回避、ログ・テキスト表示では適用
- **Border layoutで適切な領域分割**
- **VBoxで自然な垂直配置**
- **ダイアログサイズは内容に応じて調整**

### 2. 設定管理
- **一時設定オブジェクトで編集**
- **保存時にバリデーション実行**
- **キャンセル時は変更を破棄**

### 3. ユーザビリティ
- **リアルタイムバリデーション**
- **エラーメッセージの適切な表示**
- **多言語対応の統一**

### 4. 状態管理
- **関連ウィジェットの連動**
- **設定変更時のUI更新**
- **ログ出力による操作記録**

## 参考実装

### メインウィンドウレイアウト (window.go)
```go
// 成功例: Border layoutでの適切な領域分割
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

## まとめ

FyneでのGUI開発では、**コンテンツの性質に応じたスクロール適用**が最も重要：

1. **フォーム要素**: `container.NewVBox()`や`container.NewBorder()`を優先
2. **ログ・テキスト表示**: `container.NewVScroll()`や`widget.RichText`を活用
3. **大量データ**: `widget.List`や適切なスクロール対応ウィジェットを選択

Fyneのスクロールコンテナは「MinSize縮小」が仕様であることを理解し、用途に応じて適切に使い分けることで、ユーザビリティの高いGUIが実現できる。