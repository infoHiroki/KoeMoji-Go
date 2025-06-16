# KoeMoji-Go プラットフォーム対応方針

## 対応OS
- ✅ macOS (Apple Silicon/Intel)
- ✅ Windows (10/11)
- ❌ Linux

## Linux非対応の理由
1. **GUI前提の機能追加**: フォルダ選択ダイアログなど
2. **主要ユーザー層**: デスクトップ環境のライトユーザー
3. **開発リソース**: macOS/Windowsに集中
4. **FasterWhisperの制約**: サーバー用途に不向き（長時間処理、中断不可）

## 実装方針
```go
// ビルド制約を使用
// +build darwin windows

// Linux向けビルドは提供しない
```