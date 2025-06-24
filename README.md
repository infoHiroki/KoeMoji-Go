# KoeMoji-Go

音声・動画ファイル自動文字起こしツール

**日本語版 README** | [English README](README_EN.md)

## 概要

KoeMoji-Goは、音声・動画ファイルを自動で文字起こしするGoアプリケーションです。
Python版のKoeMojiAuto-cliをGoに移植し、シングルバイナリでの配布と順次処理による安定動作を実現しています。

## 📚 ドキュメント

- **[⚡ クイックスタート](QUICKSTART.md)** - 5分で始める簡潔ガイド
- **[🔧 トラブル解決](TROUBLESHOOTING.md)** - 問題解決とFAQ

詳細なドキュメントは [docs/](./docs/) ディレクトリに整理されています：

- **[使用方法](./docs/user/USAGE.md)** - 詳細な操作ガイド
- **[開発者向け](./docs/developer/)** - アーキテクチャと開発情報  
- **[設計文書](./docs/design/)** - システム設計仕様
- **[技術ノート](./docs/technical/)** - 実装課題と解決策

→ **[ドキュメント一覧を見る](./docs/README.md)**

![KoeMoji-Go Dashboard](build/screenshot01.png)

## 特徴

- **シングルバイナリ**: 実行ファイル1つで動作
- **順次処理**: 1ファイルずつ安定した処理
- **FasterWhisper連携**: 高精度な音声認識
- **AI要約機能**: OpenAI APIによる自動要約生成（v1.2.0新機能）
- **クロスプラットフォーム**: Windows/Mac対応
- **自動監視**: フォルダを定期的に監視して自動処理
- **リアルタイムUI**: 処理状況をリアルタイム表示

## 1. 動作要件の確認

### システム要件
- **OS**: Windows 10/11, macOS 10.15+
- **CPU**: Intel/AMD 64bit, Apple Silicon
- **メモリ**: 4GB以上推奨（8GB以上でより快適）
- **ストレージ**: 5GB以上（Whisperモデルファイル含む）

### 必須の前提条件

#### Python 3.8以上のインストール
KoeMoji-GoはFasterWhisperを使用するため、**Python 3.8以上が必須**です。

**Pythonのバージョン確認：**
```bash
python --version
# または
python3 --version
```

**Pythonがインストールされていない場合：**

**Windows:**
1. [Python公式サイト](https://www.python.org/downloads/windows/)からダウンロード
2. インストール時に「Add Python to PATH」をチェック
3. 推奨バージョン: Python 3.11以上

**macOS:**
```bash
# Homebrewを使用
brew install python

# または公式サイトからダウンロード
# https://www.python.org/downloads/macos/
```


#### pipの確認
```bash
pip --version
# または
pip3 --version
```

pipが利用できない場合：
```bash
# macOS
python3 -m ensurepip --upgrade

# Windows
python -m ensurepip --upgrade
```

## 2. インストール

### ダウンロード
1. **[GitHubリリースページ](https://github.com/hirokitakamura/koemoji-go/releases)から対応OS版をダウンロード**

**Windows版**: `koemoji-go-windows-1.5.0.zip`
```
📁 koemoji-go-windows-1.5.0.zip
├── koemoji-go.exe          # アイコン付き実行ファイル
├── libportaudio.dll        # 録音機能用ライブラリ
├── libgcc_s_seh-1.dll      # GCCランタイム
├── libwinpthread-1.dll     # スレッドサポート
├── config.json             # 設定ファイル
└── README.md               # 説明書
```

**macOS Intel版**: `koemoji-go-macos-intel-1.5.0.tar.gz`
```
📁 koemoji-go-macos-intel-1.5.0.tar.gz  
├── koemoji-go         # Intel Mac用実行ファイル
├── config.json        # 設定ファイル
└── README.md          # 説明書
```

**macOS Apple Silicon版**: `koemoji-go-macos-arm64-1.5.0.tar.gz`
```
📁 koemoji-go-macos-arm64-1.5.0.tar.gz  
├── koemoji-go         # Apple Silicon用実行ファイル
├── config.json        # 設定ファイル
└── README.md          # 説明書
```


2. **ダウンロードファイルを展開**

3. **初回実行時、FasterWhisperが自動インストールされます**

### 手動インストールが必要な場合
```bash
pip install faster-whisper whisper-ctranslate2
```

## 3. 初回実行

### 基本実行

**Windows:**
```cmd
koemoji-go.exe
```

**macOS:**
```bash
./koemoji-go
```


初回実行時は自動的にデフォルト設定で起動します。設定は実行後に`c`キーで変更可能です。

### PATHに追加（オプション）

どこからでも実行できるようにPATHに追加できます：

**macOS:**
```bash
# バイナリをPATHに追加
sudo cp koemoji-go /usr/local/bin/koemoji-go
sudo chmod +x /usr/local/bin/koemoji-go

# エイリアス設定（お好みで）
echo 'alias koe="koemoji-go"' >> ~/.zshrc  # zshの場合
echo 'alias koe="koemoji-go"' >> ~/.bashrc # bashの場合
source ~/.zshrc  # 設定を反映
```

**Windows:**
```cmd
# PATHの設定は環境変数から手動で設定してください
# または実行ファイルのあるフォルダでコマンドプロンプトを開いてください
```

これで`koemoji-go`または`koe`コマンドで実行できます。

初回実行時に以下が自動作成されます：
- `input/` - 処理対象ファイル置き場
- `output/` - 処理結果出力先  
- `archive/` - 処理済みファイル保管
- `koemoji.log` - ログファイル

## 4. 基本的な使い方

### ステップ1: 音声ファイルを準備
- 対応ファイルを`input/`フォルダに配置
- 複数ファイルも同時に処理可能

### ステップ2: 処理状況を確認
- リアルタイムでUI画面に処理状況が表示されます
- 完了したファイルは自動的に`archive/`に移動
- 文字起こし結果は`output/`フォルダに保存

### 処理の流れ
```
[input/音声ファイル] → [文字起こし処理] → [output/テキストファイル]
                                    ↓              ↓ (AI要約有効時)
                            [archive/処理済みファイル]   [output/要約ファイル]
```

- 1分間隔で自動的に`input/`フォルダをスキャン（デフォルト）
- 新しいファイルが見つかると順次処理を開始
- 処理完了後、元ファイルは`archive/`に移動

## 5. UI モード選択（v1.5.0新機能）

**GUI モード（デフォルト）**: グラフィカルな画面で操作
```bash
./koemoji-go
```

**TUI モード**: ターミナル画面で操作
```bash
./koemoji-go --tui
```

## 6. 対話操作

実行中に以下のキーで操作できます：
- `c` - 設定変更
- `l` - ログ表示
- `r` - 録音開始/停止（v1.4.0新機能）
- `s` - 手動スキャン
- `i` - 入力ディレクトリを開く
- `o` - 出力ディレクトリを開く
- `q` - 終了

## 7. 対応ファイル形式

- **音声**: MP3, WAV, M4A, FLAC, OGG, AAC
- **動画**: MP4, MOV, AVI

## 8. 設定のカスタマイズ

`config.json`で動作をカスタマイズできます：

```json
{
  "whisper_model": "large-v3",
  "language": "ja",
  "scan_interval_minutes": 1,
  "max_cpu_percent": 95,
  "compute_type": "int8",
  "use_colors": true,
  "ui_mode": "enhanced",
  "llm_summary_enabled": false,
  "llm_api_provider": "openai",
  "llm_api_key": "",
  "llm_model": "gpt-4o",
  "llm_max_tokens": 4096
}
```

### 設定項目

**基本設定：**
- `whisper_model`: Whisperモデル（tiny, base, small, medium, large, large-v2, large-v3）
- `language`: 言語コード（ja, en等）
- `scan_interval_minutes`: フォルダ監視間隔（分）
- `max_cpu_percent`: CPU使用率上限（現在未使用）
- `compute_type`: 量子化タイプ（int8, float16等）
- `use_colors`: カラー表示の有効/無効
- `ui_mode`: UI表示モード（enhanced/simple）

**AI要約設定（v1.2.0新機能）：**
- `llm_summary_enabled`: AI要約機能の有効/無効
- `llm_api_provider`: APIプロバイダー（現在はopenaiのみ）
- `llm_api_key`: OpenAI APIキー
- `llm_model`: 使用するモデル（gpt-4o, gpt-4-turbo, gpt-3.5-turbo等）
- `llm_max_tokens`: 最大トークン数（要約の長さ）

### Whisperモデルの選択

| モデル | サイズ | 処理速度 | 精度 | 推奨用途 |
|--------|--------|----------|------|----------|
| tiny | 最小 | 最速 | 低 | テスト用 |
| base | 小 | 速い | 中 | 簡単な音声 |
| small | 中 | 普通 | 中 | バランス重視 |
| medium | 大 | 遅い | 高 | 品質重視 |
| large | 最大 | 最遅 | 最高 | 高精度（旧版） |
| large-v2 | 最大 | 最遅 | 最高 | 多言語改善版 |
| large-v3 | 最大 | 最遅 | 最高 | **日本語推奨** |

**推奨**: 日本語の文字起こしには`large-v3`が最適です（ハルシネーション大幅減少）。

### AI要約機能の設定（v1.2.0新機能）

1. **OpenAI APIキーの取得**:
   - [OpenAI Platform](https://platform.openai.com/)でアカウント作成
   - APIキーを生成

2. **設定方法**:
   - `c`キーで設定画面を開く
   - 項目14でAPIキーを設定
   - 項目15でモデル選択（gpt-4o推奨）
   - または`config.json`を直接編集

3. **使用方法**:
   - `a`キーでAI要約をオン/オフ
   - 文字起こし完了後、自動的に要約が生成される
   - 要約は`output/ファイル名_summary.txt`として保存

## 9. コマンドラインオプション

```bash
./koemoji-go -config custom.json  # カスタム設定ファイル
./koemoji-go -debug               # デバッグモード
./koemoji-go -version             # バージョン表示
./koemoji-go -help                # ヘルプ表示
./koemoji-go -configure           # 設定モード
```

## 10. トラブルシューティング

問題が発生した場合は [TROUBLESHOOTING.md](TROUBLESHOOTING.md) をご確認ください。

主な対応内容：
- 環境・インストール関連のエラー
- 実行・操作時の問題
- パフォーマンス・品質の改善
- エラー・異常終了の対処法
- 高度な設定・カスタマイズ方法

---

## 開発者向け情報

### ビルド方法

#### 簡単ビルド（アイコン付き・推奨）

**Windows:**
```cmd
cd build
build.bat
```

**macOS/Linux:**
```bash
# 全プラットフォーム向けアイコン付きビルド
./build.sh

# 特定プラットフォームのみ
./build.sh windows   # Windows版のみ
./build.sh macos     # macOS版のみ

# ビルド成果物のクリーンアップ
./build.sh clean
```

**生成されるファイル:**
- Windows: `koemoji-go-windows-1.5.0.zip` (アイコン付き.exe)
- macOS Intel: `koemoji-go-macos-intel-1.5.0.tar.gz` (Intel Mac専用)
- macOS Apple Silicon: `koemoji-go-macos-arm64-1.5.0.tar.gz` (M1/M2 Mac専用)

#### 開発用シンプルビルド
```bash
go build -o koemoji-go main.go
```

#### 手動ビルド（アイコンなし）
```bash
# Windows 64bit
GOOS=windows GOARCH=amd64 go build -o koemoji-go-windows-amd64.exe main.go

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o koemoji-go-darwin-amd64 main.go

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o koemoji-go-darwin-arm64 main.go
```

### 開発環境セットアップ

#### 必要ツール
- Go 1.21以上
- Python 3.8以上 + FasterWhisper（テスト用）
- Git

**Windows追加要件:**
- MSYS2（MinGW-w64 GCCツールチェーン）
- 詳細は[Windows Build Guide](./docs/WINDOWS_BUILD_GUIDE.md)を参照

#### セットアップ手順
```bash
git clone https://github.com/hirokitakamura/koemoji-go.git
cd koemoji-go
go mod tidy
go build -o koemoji-go main.go
```

### 技術仕様

#### アーキテクチャ
- **言語**: Go 1.21
- **依存関係**: 標準ライブラリのみ
- **外部連携**: FasterWhisper（whisper-ctranslate2）
- **処理方式**: 順次処理（1ファイルずつ）

#### 主要機能
- 自動ディレクトリ監視
- リアルタイムUI表示
- ログ管理
- 設定ファイル管理
- クロスプラットフォーム対応

## ライセンス

**個人利用**: 自由に使用可能  
**商用利用**: 事前連絡が必要

詳細は[LICENSE](LICENSE)ファイルをご確認ください。

## 作者

KoeMoji-Go開発チーム
連絡先: koemoji2024@gmail.com