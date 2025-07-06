# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

KoeMoji-Goは、Goで書かれた音声・動画ファイル自動文字起こしツールです。GUI（Fyne-based）とTUIの両インターフェースを提供し、FasterWhisperによる音声認識と、オプションのOpenAI API統合によるAI要約機能を備えています。

### 主な特徴
- **クロスプラットフォーム**: Windows、macOS対応（Linux未テスト）
- **二つのUI**: GUI（Fyne）とTUI（Terminal UI）
- **高精度音声認識**: FasterWhisper（OpenAI Whisperの高速版）
- **リアルタイム録音**: PortAudio統合による録音機能
- **AI要約**: OpenAI API連携でテキスト要約生成
- **ポータブル設計**: 単一実行ファイルで動作（Python依存を除く）

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

### リリース成果物
- **macOS**: `KoeMoji-Go-v{VERSION}-mac.tar.gz`
- **Windows**: `KoeMoji-Go-v{VERSION}-win.zip`
- **解凍後フォルダ**: `KoeMoji-Go-v{VERSION}/`

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

### ディレクトリ構造（詳細）
```
KoeMoji-Go/
├── cmd/koemoji-go/         # アプリケーションエントリーポイント
│   └── main.go            # メイン関数、CLI引数処理
├── internal/              # 内部パッケージ（外部公開しない）
│   ├── config/           # 設定管理
│   │   ├── config.go    # 設定構造体と読み込み
│   │   └── validate.go  # 設定バリデーション
│   ├── gui/             # GUIコンポーネント
│   │   ├── app.go      # Fyneアプリケーション
│   │   ├── components.go # UIコンポーネント更新
│   │   ├── window.go   # ウィンドウレイアウト
│   │   ├── resources.go # 埋め込みリソース（アイコン）
│   │   ├── dialogs.go  # ダイアログ機能
│   │   └── icon.png    # アプリケーションアイコン
│   ├── ui/              # TUIコンポーネント
│   │   └── tui.go      # ターミナルUI実装
│   ├── processor/       # ファイル処理エンジン
│   │   └── processor.go # ファイル監視と処理
│   ├── recorder/        # 録音システム
│   │   └── recorder.go  # PortAudio統合
│   ├── whisper/         # 音声認識
│   │   └── whisper.go   # FasterWhisper呼び出し
│   ├── llm/            # AI統合
│   │   └── openai.go   # OpenAI API連携
│   └── logger/         # ログシステム
│       └── logger.go   # バッファード・ロガー
├── build/              # ビルドスクリプトとアセット
│   ├── common/assets/  # 共通リソース
│   │   ├── config.example.json
│   │   └── README_RELEASE.md
│   ├── macos/         # macOSビルド
│   │   └── build.sh
│   └── windows/       # Windowsビルド
│       ├── build.bat
│       ├── *.dll      # 必要なDLLファイル
│       └── icon.ico   # アプリケーションアイコン
├── test/              # テストファイル
│   └── manual-test-commands.md
├── version.go         # バージョン定義
├── go.mod            # Go依存関係
├── README.md         # ユーザー向けドキュメント
├── CLAUDE.md         # このファイル
└── config.json       # ユーザー設定（実行時生成）
```

### 主要技術
- **GUIフレームワーク**: クロスプラットフォームデスクトップインターフェース用Fyne v2.6.1
  - アプリアイコン統合（go:embed）
  - ダークモード自動対応
  - 統一されたアイコンデザイン
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

### 設定ファイル例
```json
{
  "input_dir": "./input",
  "output_dir": "./output",
  "archive_dir": "./archive",
  "whisper": {
    "model": "medium",
    "language": "ja",
    "compute_type": "int8"
  },
  "openai": {
    "api_key": "",
    "enabled": false,
    "prompt": "以下のテキストを要約してください："
  }
}
```

## ビルドシステム

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
- macOS: `KoeMoji-Go-v{VERSION}-mac.tar.gz`
- Windows: `KoeMoji-Go-v{VERSION}-win.zip`
- 解凍後フォルダ: `KoeMoji-Go-v{VERSION}/`

### DLL処理（Windows）
`build.bat`はワイルドカード`*.dll`を使用して自動的にDLLをコピー：
```batch
copy /Y *.dll "%DIST_DIR%\" >nul
```

## よくある問題と解決策

### 1. FasterWhisperインストール問題
**症状**: 「whisper-ctranslate2が見つかりません」エラー
**解決策**:
```bash
pip install faster-whisper
# または
pip3 install faster-whisper
```

### 2. Windows GPU環境での問題
**症状**: `compute_type: int8`設定でもGPUが使用される
**解決策**: config.jsonで明示的にCPU使用を指定（内部で`--device cpu`を自動追加）

### 3. macOS録音許可
**症状**: 録音開始時にエラー
**解決策**: システム環境設定 > セキュリティとプライバシー > マイクでアプリを許可

### 4. ファイル処理が始まらない
**症状**: inputフォルダにファイルを置いても処理されない
**解決策**: 
- ファイル形式を確認（サポート形式のみ）
- ログで詳細確認: `./koemoji-go --debug`
- 権限を確認

## デバッグとトラブルシューティング

### ログレベル
```bash
# 通常ログ
./koemoji-go

# デバッグログ（詳細）
./koemoji-go --debug
```

### ログ出力場所
- **GUI**: アプリケーション内のログエリア
- **TUI**: ターミナル画面下部
- **ファイル**: なし（stdout/stderrのみ）

### よく使うデバッグコマンド
```bash
# Whisperコマンドの確認
which whisper-ctranslate2

# Python環境確認
python --version
pip list | grep faster-whisper

# PortAudio確認（macOS）
brew list portaudio

# DLL確認（Windows）
dir build\windows\*.dll
```

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

## 最近の重要な変更

### v1.5.4での変更
1. **フォルダ命名規則の統一**
   - 旧: `koemoji-go-windows-1.5.4`、`koemoji-go-macos-arm64-1.5.4`
   - 新: `KoeMoji-Go-v1.5.4`（全プラットフォーム共通）
   - 効果: アップデート時のユーザーデータ保護

2. **Windows DLL処理の簡素化**
   - 旧: MSYS2パス検出の複雑な条件分岐（14行）
   - 新: ワイルドカード`*.dll`使用（4行）
   - 効果: 保守性向上、新DLL自動対応

3. **FasterWhisper自動インストール改善**
   - Windows実行ファイルパス対応
   - GPU環境での`--device cpu`自動追加
   - クロスプラットフォーム対応強化

## Git操作時の注意事項

### ブランチ戦略
- `main`: 安定版リリース
- `feature/*`: 新機能開発
- `fix/*`: バグ修正

### コミットメッセージ規則
```
type: 簡潔な説明

- 詳細な変更内容
- 影響範囲

Co-Authored-By: Claude <noreply@anthropic.com>
```

### リリースプロセス
1. version.goのバージョン番号更新
2. ビルドスクリプト実行
3. 成果物確認（releases/フォルダ）
4. GitHubリリース作成
5. アセットアップロード

## 開発のヒント

1. **設定変更時**: 必ずvalidate.goでバリデーション追加
2. **新機能追加時**: config.example.json更新を忘れずに
3. **エラー処理**: ユーザーフレンドリーなメッセージを心がける
4. **クロスプラットフォーム**: OS固有の処理は明確に分離
5. **テスト**: 手動テストコマンドをmanual-test-commands.mdに追加

## 連絡先とサポート

- **GitHub Issues**: バグ報告と機能要望
- **作者**: [@infoHiroki](https://github.com/infoHiroki)
- **ライセンス**: 個人利用自由、商用利用要連絡