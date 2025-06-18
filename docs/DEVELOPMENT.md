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

## Windows環境問題修正計画

### 進捗状況
- ✅ **1. ログ表示問題** - 修正完了
- ✅ **2. 色表示問題** - 修正完了  
- ⏳ **3. アイコン問題** - Windows実機テスト待ち

### 1. ログ表示問題 ✅ 決定済み

**現象:** `l` コマンドでログファイルが開けない

**解決方法:** notepad一択を採用
```go
// 修正箇所: internal/ui/ui.go 140行目
cmd = exec.Command("notepad", "koemoji.log")
```

**採用理由:**
- Windows 95以降、全バージョンに標準搭載（100%互換性）
- システムコアコンポーネントで削除不可
- シンプルで確実、日本語ログファイル表示も問題なし

### 2. 色表示問題 ✅ 決定済み

**現象:** `[情報 ]` のようにカラーコードが `[]` で表示され、スペースが含まれる

**解決方法:** 積極的カラー化を採用

**修正内容:**
1. **カラー表示の強制有効化**
```go
// 修正箇所: internal/ui/ui.go 173-175行目
if runtime.GOOS == "windows" {
    return true // Windows 10以降は強制有効
}
```

2. **フォーマット修正（スペース削除）**
```go
// 修正箇所: internal/ui/ui.go 113行目  
fmt.Printf("[%s] %s %s\n", localizedLevel, timestamp, entry.Message)
```

**採用理由:**
- KoeMoji-Goユーザー層は技術に詳しく、新しい環境を使用
- Windows 7/8.1はサポート終了済み（極少数ユーザー）
- 95%のユーザー体験向上を優先

### 3. アイコン問題 🔄 調査中

**現象:** EXEファイルにアイコンが表示されない

**調査ログ:**
- ✅ **2025-06-18 13:20** - versioninfo.jsonのバージョンを1.0.0→1.1.0に更新
- ✅ **2025-06-18 13:20** - ビルド成功、goversioninfo正常動作確認
- ✅ **2025-06-18 13:21** - goversioninfoツール動作確認（最新版使用）
- ✅ **2025-06-18 13:21** - icon.icoファイル形式確認（256x256 PNG、正常なWindows ICO）

**考えられる原因:**
- versioninfo.jsonのバージョンが古い（1.0.0 → 1.1.0） ← **修正済み**
- goversioninfoツールの問題
- アイコンファイルの形式問題  
- ビルドプロセスの問題

**対応方針:**
1. ✅ バージョン更新 → テスト
2. ✅ goversioninfo更新 → テスト
3. ⏳ Windows実機での動作確認（Windows環境でEXEファイルのアイコン表示テスト）
4. 🔄 代替手段（ショートカット作成バッチ）の検討

**次のステップ:**
Windows環境での実際のテストが必要。ビルドプロセスは正常で、アイコンも適切に埋め込まれている可能性が高い。