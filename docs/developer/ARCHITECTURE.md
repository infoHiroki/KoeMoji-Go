# KoeMoji-Go アーキテクチャ

## 概要
KoeMoji-Goは音声・動画ファイルの自動文字起こしを行うGo言語製CLIツールです。

## プロジェクト構造

```
KoeMoji-Go/
├── cmd/
│   └── koemoji-go/
│       ├── main.go          # アプリケーションエントリーポイント
│       └── koemoji-go       # 実行バイナリ
├── internal/
│   ├── config/
│   │   └── config.go        # 設定管理・JSON読み書き・対話式設定
│   ├── logger/
│   │   └── logger.go        # ログ管理・バッファリング・出力制御
│   ├── processor/
│   │   └── processor.go     # ファイル監視・処理キュー・フロー制御
│   ├── ui/
│   │   ├── ui.go           # リアルタイムUI表示・キー入力処理
│   │   └── messages.go     # 多言語メッセージ・ローカライゼーション
│   └── whisper/
│       └── whisper.go      # Whisper連携・音声認識実行・出力処理
├── build/
│   ├── build.sh            # マルチプラットフォームビルドスクリプト
│   ├── icon.ico            # Windows用アプリケーションアイコン
│   ├── versioninfo.json    # Windows用バージョン情報
│   └── dist/               # ビルド成果物出力ディレクトリ
├── docs/
│   ├── ARCHITECTURE.md     # システム設計・技術仕様
│   ├── DEVELOPMENT.md      # 開発ガイド・ビルド手順
│   └── USAGE.md           # 使用方法・コマンドリファレンス
├── input/                  # 処理対象ファイル配置ディレクトリ
├── output/                 # 文字起こし結果出力ディレクトリ
├── archive/                # 処理済みファイル保管ディレクトリ
├── config.example.json     # 設定ファイルテンプレート
├── go.mod                  # Go モジュール定義
├── koemoji.log            # アプリケーションログファイル
├── README.md              # プロジェクト概要（日本語）
├── README_EN.md           # プロジェクト概要（英語）
└── LICENSE                # ライセンス情報
```

### ディレクトリ詳細

#### `/cmd/koemoji-go/`
アプリケーションの実行可能ファイルと main パッケージを格納。Go の標準的なプロジェクト構造に従い、実行バイナリは cmd ディレクトリ配下に配置。

#### `/internal/`
アプリケーションの内部パッケージ群。external import を防ぎ、API の安定性を保つために internal ディレクトリを使用。

- **config**: 設定ファイルの読み書き、対話式設定エディタ
- **logger**: 構造化ログ、バッファ管理、リアルタイム表示対応
- **processor**: ファイル監視、処理キュー管理、並行処理制御
- **ui**: ターミナルUI、リアルタイム表示、キーボード入力処理
- **whisper**: faster-whisper連携、音声認識実行、結果出力

#### `/build/`
ビルド関連ファイル群。マルチプラットフォーム対応とアイコン付きバイナリ生成をサポート。

#### `/docs/`
プロジェクトドキュメント群。設計書、開発ガイド、使用方法を整理。

#### 実行時ディレクトリ
- **input**: ユーザーが音声・動画ファイルを配置
- **output**: 文字起こし結果（txt, srt, vtt等）を保存
- **archive**: 処理完了ファイルを移動・保管

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
- **macOS**: Apple Silicon (フォルダ選択ダイアログ対応)

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