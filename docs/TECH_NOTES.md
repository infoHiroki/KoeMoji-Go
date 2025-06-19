# Technical Notes

## Fyne Dialog Button Control Issue

### 問題
`dialog.NewCustom("設定", "", content, window)`で空文字を指定しても、Fyneが内部的にデフォルトの閉じるボタンを自動生成し、ダイアログの中央に小さなボタンとして表示される問題が発生。

### 症状
- 設定ダイアログの中央に謎の小さなグレーボタンが表示
- クリックするとダイアログが閉じる動作
- 独自ボタン（キャンセル、保存等）とは別に表示される

### 原因
`dialog.NewCustom()`の第2引数（dismissText）に空文字`""`を指定しても、Fyneフレームワークが内部的に標準の閉じるボタンを生成するため。

### 解決策
```go
// 問題のあるコード
cd.dialog = dialog.NewCustom("設定", "", content, cd.app.window)

// 修正後のコード  
cd.dialog = dialog.NewCustomWithoutButtons("設定", content, cd.app.window)
```

### 学習事項
- Fyneダイアログで完全にカスタムボタンのみを使用したい場合は`NewCustomWithoutButtons()`を使用
- `NewCustom()`は標準ボタンが必ず生成される仕様
- 当て推量での修正ではなく、原因を特定してから適切な修正を実施することの重要性

### 日付
2025-06-18

### 関連ファイル
- `internal/gui/configdialog.go:271`