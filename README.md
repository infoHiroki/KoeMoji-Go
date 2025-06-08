# KoeMoji-Go

音声・動画ファイル自動文字起こしツール

**日本語版 README** | [English README](README_EN.md)

## 概要

KoeMoji-Goは、音声・動画ファイルを自動で文字起こしするGoアプリケーションです。
Python版のKoeMojiAuto-cliをGoに移植し、シングルバイナリでの配布と順次処理による安定動作を実現しています。

## 特徴

- **シングルバイナリ**: 実行ファイル1つで動作
- **順次処理**: 1ファイルずつ安定した処理
- **FasterWhisper連携**: 高精度な音声認識
- **クロスプラットフォーム**: Windows/Mac/Linux対応
- **自動監視**: フォルダを定期的に監視して自動処理
- **リアルタイムUI**: 処理状況をリアルタイム表示

## 1. 動作要件の確認

### システム要件
- **OS**: Windows 10/11, macOS 10.15+, Linux (主要ディストリビューション)
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

**Linux (Ubuntu/Debian):**
```bash
sudo apt update
sudo apt install python3 python3-pip
```

**Linux (CentOS/RHEL):**
```bash
sudo yum install python3 python3-pip
# または
sudo dnf install python3 python3-pip
```

#### pipの確認
```bash
pip --version
# または
pip3 --version
```

pipが利用できない場合：
```bash
# macOS/Linux
python3 -m ensurepip --upgrade

# Windows
python -m ensurepip --upgrade
```

## 2. インストール

### ダウンロード
1. **[GitHubリリースページ](https://github.com/[username]/koemoji-go/releases)から対応OS版をダウンロード**

**Windows版**: `koemoji-go-windows-1.0.0.zip`
```
📁 koemoji-go-windows-1.0.0.zip
├── koemoji-go.exe     # アイコン付き実行ファイル
├── config.json        # 設定ファイル
└── README.md          # 説明書
```

**macOS版**: `koemoji-go-macos-1.0.0.tar.gz`
```
📁 koemoji-go-macos-1.0.0.tar.gz  
├── koemoji-go-darwin-amd64    # Intel Mac用実行ファイル
├── koemoji-go-darwin-arm64    # Apple Silicon用実行ファイル
├── config.json                # 設定ファイル
└── README.md                  # 説明書
```

**Linux版**: `koemoji-go-linux-1.0.0.tar.gz`
```
📁 koemoji-go-linux-1.0.0.tar.gz
├── koemoji-go         # 実行ファイル
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
# Intel Mac
./koemoji-go-darwin-amd64

# Apple Silicon
./koemoji-go-darwin-arm64
```

**Linux:**
```bash
./koemoji-go
```

### PATHに追加（オプション）

どこからでも実行できるようにPATHに追加できます：

**macOS/Linux:**
```bash
# バイナリをPATHに追加
sudo cp koemoji-go-darwin-arm64 /usr/local/bin/koemoji-go  # Apple Silicon
sudo cp koemoji-go-darwin-amd64 /usr/local/bin/koemoji-go  # Intel Mac
sudo cp koemoji-go /usr/local/bin/koemoji-go               # Linux
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
                                    ↓
                            [archive/処理済みファイル]
```

- 10分間隔で自動的に`input/`フォルダをスキャン
- 新しいファイルが見つかると順次処理を開始
- 処理完了後、元ファイルは`archive/`に移動

## 5. 対話操作

実行中に以下のキーで操作できます：
- `c` - 設定表示
- `l` - 全ログ表示
- `s` - 手動スキャン実行
- `q` - 終了
- `Enter` - 画面更新
- `Ctrl+C` - 強制終了

## 6. 対応ファイル形式

- **音声**: MP3, WAV, M4A, FLAC, OGG, AAC
- **動画**: MP4, MOV, AVI

## 7. 設定のカスタマイズ

`config.json`で動作をカスタマイズできます：

```json
{
  "whisper_model": "large-v3",
  "language": "ja",
  "scan_interval_minutes": 10,
  "max_cpu_percent": 95,
  "compute_type": "int8",
  "use_colors": true,
  "ui_mode": "enhanced"
}
```

### 設定項目

- `whisper_model`: Whisperモデル（tiny, base, small, medium, large, large-v2, large-v3）
- `language`: 言語コード（ja, en等）
- `scan_interval_minutes`: フォルダ監視間隔（分）
- `max_cpu_percent`: CPU使用率上限（現在未使用）
- `compute_type`: 量子化タイプ（int8, float16等）
- `use_colors`: カラー表示の有効/無効
- `ui_mode`: UI表示モード（enhanced/simple）

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

## 8. コマンドラインオプション

```bash
./koemoji-go --config custom.json  # カスタム設定ファイル
./koemoji-go --debug               # デバッグモード
./koemoji-go --version             # バージョン表示
./koemoji-go --help                # ヘルプ表示
```

## 9. トラブルシューティング

### よくある問題

**Q: "Python not found" エラー**
- Pythonがインストールされていません
- 上記「1. 動作要件の確認」に従ってPythonをインストールしてください
- インストール後、ターミナル/コマンドプロンプトを再起動してください

**Q: Pythonはあるが古いバージョン**
```bash
# バージョン確認
python --version

# Python 3.8未満の場合、新しいバージョンをインストール
```

**Q: FasterWhisperのインストールに失敗する**
```bash
# 手動でインストールしてください
pip install faster-whisper whisper-ctranslate2

# pipが古い場合
pip install --upgrade pip
pip install faster-whisper whisper-ctranslate2

# 権限エラーの場合
pip install --user faster-whisper whisper-ctranslate2
```

**Q: "whisper-ctranslate2 not found" エラー**
- Pythonのパスが通っていない可能性があります
- pipでインストールしたパッケージのパスが通っていない可能性があります
- 以下を確認してください：
```bash
# パッケージがインストールされているか確認
pip show whisper-ctranslate2

# パスの確認
which whisper-ctranslate2
# または
where whisper-ctranslate2  # Windows
```

**Q: 処理が遅い**
- `config.json`でモデルを`small`や`medium`に変更
- デフォルトで既に最高速設定（`compute_type`: `int8`）済み

**Q: 音声ファイルが認識されない**
- 対応形式: MP3, WAV, M4A, FLAC, OGG, AAC, MP4, MOV, AVI
- ファイル名に特殊文字が含まれていないか確認

**Q: 文字起こし結果がおかしい**
- `large-v3`モデルの使用を推奨
- 音声品質を確認（ノイズ、音量等）

### ログの確認

問題が発生した場合は`koemoji.log`を確認してください：
```bash
# ログファイルの確認
cat koemoji.log

# 最新のログのみ確認
tail -f koemoji.log
```

---

## 開発者向け情報

### ビルド方法

#### 簡単ビルド（アイコン付き・推奨）
```bash
# 全プラットフォーム向けアイコン付きビルド
./build.sh

# 特定プラットフォームのみ
./build.sh windows   # Windows版のみ
./build.sh macos     # macOS版のみ
./build.sh linux    # Linux版のみ

# ビルド成果物のクリーンアップ
./build.sh clean
```

**生成されるファイル:**
- Windows: `koemoji-go-windows-1.0.0.zip` (アイコン付き.exe)
- macOS: `koemoji-go-macos-1.0.0.tar.gz` (Intel/Apple Silicon両対応)
- Linux: `koemoji-go-linux-1.0.0.tar.gz` (64bit版)

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

# Linux 64bit
GOOS=linux GOARCH=amd64 go build -o koemoji-go-linux-amd64 main.go
```

### 開発環境セットアップ

#### 必要ツール
- Go 1.21以上
- Python 3.8以上 + FasterWhisper（テスト用）
- Git

#### セットアップ手順
```bash
git clone https://github.com/[username]/koemoji-go.git
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
連絡先: dev@habitengineer.com