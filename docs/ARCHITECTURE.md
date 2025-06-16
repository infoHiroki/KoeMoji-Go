# KoeMoji-Go アーキテクチャ

## 概要
KoeMoji-Goは音声・動画ファイルの自動文字起こしを行うGo言語製CLIツールです。

## システム構成

### コアコンポーネント
- **Processor**: ファイル監視・処理キュー管理
- **Whisper**: faster-whisper統合・音声認識実行
- **Config**: 設定管理・対話式設定エディタ
- **UI**: リアルタイム表示・多言語対応
- **Logger**: 構造化ログ・バッファ管理

### データフロー
```
入力ディレクトリ → ファイル検出 → 処理キュー → Whisper実行 → 出力・アーカイブ
       ↓                                           ↑
    UI表示 ←→ ログ管理 ←→ 設定管理 ←→ 多言語メッセージ
```

## 技術仕様

### 対応プラットフォーム
- **Windows**: x64 (フォルダ選択ダイアログ対応)
- **macOS**: Intel/Apple Silicon (フォルダ選択ダイアログ対応)
- **Linux**: x64 (コマンドライン設定のみ)

### 対応ファイル形式
**入力**: mp3, wav, m4a, flac, ogg, aac, mp4, mov, avi
**出力**: txt, vtt, srt, tsv, json

### UI機能
- **Enhanced Mode**: リアルタイム表示・カラー対応
- **Simple Mode**: 基本的なログ出力
- **多言語**: 日本語・英語対応

## 設定システム

### 設定項目
```json
{
  "whisper_model": "medium",      // tiny〜large-v3
  "language": "ja",               // 認識言語
  "ui_language": "en",            // UI言語 (en/ja)
  "scan_interval_minutes": 10,    // 監視間隔
  "max_cpu_percent": 95,          // CPU使用率制限
  "compute_type": "int8",         // 計算精度
  "use_colors": true,             // 色表示
  "ui_mode": "enhanced",          // UIモード
  "output_format": "txt",         // 出力形式
  "input_dir": "./input",         // 入力ディレクトリ
  "output_dir": "./output",       // 出力ディレクトリ
  "archive_dir": "./archive"      // アーカイブディレクトリ
}
```

## パフォーマンス設計

### 並行処理
- **ファイル監視**: 独立goroutine
- **UI更新**: 独立goroutine  
- **ファイル処理**: シーケンシャル（CPU負荷制御）

### メモリ管理
- **ログバッファ**: 最大12エントリでローテーション
- **処理キュー**: 動的配列で管理
- **設定**: シングルトンパターン

## セキュリティ

### ファイルアクセス制御
- 入力ディレクトリ内のファイルのみ処理許可
- パス正規化による不正アクセス防止
- 実行時権限チェック

### 依存関係管理
- faster-whisper自動インストール
- 外部コマンド実行の安全性確保