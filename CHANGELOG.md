# CHANGELOG

すべての重要な変更はこのファイルに記録されます。

フォーマットは [Keep a Changelog](https://keepachangelog.com/ja/1.0.0/) に基づいています。

## [1.5.0] - 2025-06-24

### 追加
- Windowsビルド時のDLLファイル同梱
  - libportaudio.dll（録音機能用）
  - libgcc_s_seh-1.dll（GCCランタイム）
  - libwinpthread-1.dll（スレッドサポート）

### 修正
- Windows環境で外部プログラム起動時にコンソールウィンドウが表示される問題を修正
  - `CREATE_NO_WINDOW`フラグを使用してバックグラウンドで実行
  - ただし、`explorer.exe`は例外として通常のコマンド実行を使用
- 相対パスの解決問題を修正
  - 実行ファイルのディレクトリを基準にinput/output/archiveディレクトリを解決
  - どこから実行しても正しいディレクトリを参照するように改善
- GUI版で入力/出力ディレクトリボタンが反応しない問題を修正
  - `explorer.exe`に対して`HideWindow`フラグを使用しないように変更

### 変更
- ビルドプロセスの改善
  - Windows用のプラットフォーム固有コードを分離（exec_windows.go）
  - パス解決ロジックを専用モジュールに分離（paths.go）

### ドキュメント
- README.mdの配布パッケージ構成を更新（DLLファイルを追加）
- WINDOWS_BUILD_GUIDE.mdの既知の問題を更新（解決済みの問題をマーク）

## [1.4.0] - 以前

### 追加
- 録音機能の実装

[1.5.0]: https://github.com/hirokitakamura/koemoji-go/releases/tag/v1.5.0
