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

### コード構造
```
cmd/koemoji-go/main.go    # エントリーポイント
internal/
├── config/               # 設定管理
├── ui/                   # UI・メッセージ
├── processor/            # ファイル処理
├── whisper/             # 音声認識
└── logger/              # ログ管理
```

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
- macOS: Intel/AppleSilicon バイナリ + 設定ファイル
- Linux: バイナリ + 設定ファイル