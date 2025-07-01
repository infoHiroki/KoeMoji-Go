# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

KoeMoji-Goは、Goで書かれた音声・動画ファイル自動文字起こしツールです。GUI（Fyne-based）とTUIの両インターフェースを提供し、FasterWhisperによる音声認識と、オプションのOpenAI API統合によるAI要約機能を備えています。

## 開発コマンド

### ビルド
```bash
# 開発用ビルド
go build -o koemoji-go ./cmd/koemoji-go

# macOS配布用ビルド
cd build/macos && ./build.sh

# Windows配布用ビルド（MSYS2/MinGW64が必要）
cd build/windows && build.bat

# ビルド成果物のクリーンアップ
cd build/macos && ./build.sh clean
cd build/windows && build.bat clean
```

### テスト
```bash
# 単体テスト実行
go test ./...

# 特定パッケージのテスト実行
go test ./internal/config
go test ./internal/logger
go test ./internal/processor

# 統合テスト実行
go test ./test

# 手動テスト（詳細な手順はtest/manual-test-commands.mdを参照）
./koemoji-go --version
./koemoji-go --help
./koemoji-go --gui     # GUIモード
./koemoji-go --tui     # ターミナルUIモード
./koemoji-go --debug   # デバッグログ
```

### アプリケーション実行
```bash
# GUIモード（デフォルト）
./koemoji-go

# ターミナルUIモード
./koemoji-go --tui

# 設定モード
./koemoji-go --configure

# デバッグモード
./koemoji-go --debug
```

## アーキテクチャ概要

### 中核設計パターン
- **単一責任**: 各internalパッケージは専門機能に特化
- **順次処理**: 安定性のための1ファイルずつ処理方式
- **コンテキストベース停止**: Goのcontextパッケージによる優雅な終了処理
- **依存性注入**: 設定とロガーは関数パラメータ経由で受け渡し

### ディレクトリ構造
- `/cmd/koemoji-go/` - アプリケーションエントリーポイント
- `/internal/config/` - JSON永続化による設定管理
- `/internal/gui/` - Fyneベースの録音状態管理機能付きGUIコンポーネント
- `/internal/ui/` - キーボードショートカット付きターミナルUIコンポーネント
- `/internal/processor/` - ファイル監視と順次処理エンジン
- `/internal/recorder/` - PortAudioベースの録音デバイス選択機能付き録音システム
- `/internal/whisper/` - FasterWhisper音声認識統合
- `/internal/llm/` - OpenAI API AI要約統合
- `/internal/logger/` - スレッドセーフバッファード・ログシステム
- `/build/` - プラットフォーム固有ビルドスクリプトとアセット
- `/test/` - 統合テストと手動テスト手順

### 主要技術
- **GUIフレームワーク**: クロスプラットフォームデスクトップインターフェース用Fyne v2.6.1
- **音声処理**: リアルタイム音声録音用PortAudio
- **音声認識**: FasterWhisper（Python依存）
- **AI統合**: テキスト要約用OpenAI API
- **テスト**: 単体テスト用Testify v1.10.0

### サポート形式
- **入力音声**: MP3, WAV, M4A, FLAC, OGG, AAC
- **入力動画**: MP4, MOV, AVI
- **出力**: TXT, VTT, SRT, TSV, JSON

## 設定システム

設定は`config.json`で管理され、以下の主要領域があります：
- **Whisper設定**: モデル選択（tinyからlarge-v3）、言語設定、計算精度
- **パフォーマンス**: CPU使用制限、処理オプション
- **ディレクトリ**: 入力/出力/アーカイブパスの設定
- **録音**: デバイス選択、録音時間/サイズ制限
- **AI機能**: OpenAI APIキーとカスタマイズ可能プロンプト
- **UI**: カラーサポート、言語設定

新規インストール時は`config.example.json`をテンプレートとして使用してください。

## ビルドシステム注意事項

### 前提条件
- **Go 1.21+** コンパイル用
- **Python 3.8+** FasterWhisperバックエンド用
- **MSYS2/MinGW64** Windows ビルド用（CGO依存）
- **PortAudio** 開発ライブラリ

### プラットフォーム固有の考慮事項
- **macOS**: Apple Silicon最適化、マイクアクセス許可が必要
- **Windows**: 必要なDLL（libportaudio.dll、GCCランタイム）を同梱
- **CGO**: PortAudio統合に必須

### 配布パッケージ化
ビルドスクリプトは全依存関係を含む実行可能パッケージを作成：
- macOS: 実行ファイルと設定テンプレートを含む`.tar.gz`
- Windows: 実行ファイル、DLL、設定テンプレートを含む`.zip`

## 重要な実装詳細

### 録音状態管理
録音機能は、GUIコンポーネントとレコーダーバックエンド間の慎重な状態同期を使用します。以下に注意：
- チャネルを使ったスレッドセーフな状態更新
- 録音開始/停止操作でのUI状態一貫性
- 特に録音中のアプリケーション終了時の適切なクリーンアップ

### ファイル処理パイプライン
リソース競合を防ぐため、ファイルは順次処理されます：
1. ディレクトリ監視が新ファイルを検出
2. 単一ファイル処理でメモリ問題を防止
3. 処理成功後、ファイルをアーカイブに移動
4. エラーハンドリングで元ファイルを保護

### テストアプローチ
重要なテストシナリオには以下が含まれます：
- 録音状態管理とUI同期
- 録音中終了警告
- クロスプラットフォームビルド検証
- FasterWhisper Pythonバックエンドとの統合
- 長時間録音でのメモリ使用量

包括的なテスト手順については`test/manual-test-commands.md`を参照してください。