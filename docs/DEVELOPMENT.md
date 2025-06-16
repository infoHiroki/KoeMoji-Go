# 開発ガイド

## 環境構築

### 必要条件
- Go 1.21+
- Python 3.9+ (faster-whisper用)

### 依存関係
```bash
# faster-whisperのインストール（初回実行時に自動）
pip install faster-whisper whisper-ctranslate2
```

## ビルド

### 開発用ビルド
```bash
go build -o koemoji-go ./cmd/koemoji-go
```

### リリースビルド
```bash
cd build
./build.sh
```

## テスト

### 基本動作確認
```bash
./koemoji-go --version
./koemoji-go --help
./koemoji-go --configure
```

### 処理テスト
1. `input/`ディレクトリに音声ファイルを配置
2. アプリケーション実行
3. `output/`で結果確認
4. `archive/`で処理済みファイル確認

## 開発フロー

### プロジェクト構造
プロジェクトはGoの標準的なレイアウトに従って構成されています：

```
KoeMoji-Go/
├── cmd/koemoji-go/       # アプリケーション実行ファイル
│   ├── main.go          # エントリーポイント
│   └── koemoji-go       # ビルド済みバイナリ
├── internal/             # 内部パッケージ（外部importを制限）
│   ├── config/          # 設定管理・JSON操作
│   ├── logger/          # ログシステム・バッファ管理
│   ├── processor/       # ファイル監視・処理制御
│   ├── ui/              # ターミナルUI・多言語対応
│   └── whisper/         # 音声認識・Whisper連携
├── build/               # ビルド・配布用ファイル
│   ├── build.sh         # マルチプラットフォームビルドスクリプト
│   ├── icon.ico         # Windows用アイコン
│   ├── versioninfo.json # Windows用バージョン情報
│   └── dist/            # ビルド成果物出力先
├── docs/                # プロジェクトドキュメント
│   ├── ARCHITECTURE.md  # システム設計書
│   ├── DEVELOPMENT.md   # 開発ガイド（このファイル）
│   └── USAGE.md         # 使用方法・操作マニュアル
├── input/               # 音声・動画ファイル配置場所
├── output/              # 文字起こし結果出力先
├── archive/             # 処理済みファイル保管場所
├── config.example.json  # 設定ファイルテンプレート
├── go.mod              # Go モジュール定義
└── koemoji.log         # アプリケーションログ
```

### 各パッケージの役割

#### `cmd/koemoji-go/`
- **main.go**: アプリケーションのエントリーポイント
- 各パッケージの初期化と統合
- CLIオプションの処理

#### `internal/config/`
- 設定ファイル（JSON）の読み書き
- 対話式設定エディタ
- デフォルト設定の管理

#### `internal/logger/`
- 構造化ログ出力
- ログバッファ管理（最大12エントリ）
- UIとの連携

#### `internal/processor/`
- ディレクトリ監視（goroutine）
- ファイル処理キューの管理
- 処理フローの制御

#### `internal/ui/`
- リアルタイムターミナルUI
- キーボード入力処理
- 多言語メッセージシステム

#### `internal/whisper/`
- FasterWhisper（whisper-ctranslate2）との連携
- 音声認識処理の実行
- 結果ファイルの出力

### 新機能追加
1. 適切な`internal/`パッケージに実装
2. 必要に応じて`messages.go`に多言語メッセージ追加
3. `main.go`で統合
4. テスト・ビルド確認

## リリース準備

### バージョン更新
1. `cmd/koemoji-go/main.go`の`version`定数
2. `build/build.sh`の`VERSION`変数
3. `build/versioninfo.json`のバージョン情報

### 配布パッケージ
- Windows: `.exe` + `config.json` + `README.md`
- macOS Intel: バイナリ + 設定ファイル
- macOS Apple Silicon: バイナリ + 設定ファイル