# macOS ビルドガイド

KoeMoji-Go の macOS 版をビルドするための完全ガイドです。

## 📋 目次

- [前提条件](#前提条件)
- [ビルド環境のセットアップ](#ビルド環境のセットアップ)
- [ビルド手順](#ビルド手順)
- [トラブルシューティング](#トラブルシューティング)
- [リリース成果物](#リリース成果物)

---

## 前提条件

### 必須ソフトウェア

| ソフトウェア | バージョン | 確認コマンド |
|------------|----------|------------|
| **macOS** | 12.0 (Monterey) 以降 | `sw_vers` |
| **Go** | 1.21 以降 | `go version` |
| **PortAudio** | 最新版 | `brew list portaudio` |
| **Python** | 3.12（推奨、3.13は非対応） | `python3 --version` |
| **FasterWhisper** | 最新版 | `pip3 list \| grep faster-whisper` |

### Homebrewのインストール

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

---

## ビルド環境のセットアップ

### 1. Go のインストール

```bash
# Homebrew経由でインストール
brew install go

# インストール確認
go version
# 期待される出力: go version go1.21.x darwin/arm64
```

### 2. PortAudio のインストール

```bash
# Homebrew経由でインストール
brew install portaudio

# インストール確認
brew list portaudio
pkg-config --modversion portaudio-2.0
```

### 3. Python環境のセットアップ

```bash
# Python3確認
python3 --version

# FasterWhisperインストール
pip3 install faster-whisper

# インストール確認
pip3 list | grep faster-whisper
which whisper-ctranslate2
```

### 4. リポジトリのクローン

```bash
git clone https://github.com/hirokitakamura/koemoji-go.git
cd koemoji-go
```

---

## ビルド手順

### ステップ1: ビルドディレクトリに移動

```bash
cd build/macos
```

### ステップ2: ビルドスクリプト実行

```bash
# Apple Silconビルド（デフォルト）
./build.sh

# または明示的にarm64指定
./build.sh arm64
```

### ステップ3: ビルド成果物の確認

```bash
# 配布用tar.gzの確認
ls -lh ../releases/

# 期待される出力:
# koemoji-go-macos-1.6.1.tar.gz
```

### ステップ4: 動作確認

```bash
# 解凍
cd ../releases
tar -xzf koemoji-go-macos-1.6.1.tar.gz
cd koemoji-go-1.6.1

# バージョン確認
./koemoji-go --version
# 期待される出力: KoeMoji-Go v1.6.1

# GUIモード起動
./koemoji-go
```

---

## ビルドスクリプトのオプション

### 使用可能なコマンド

```bash
# ヘルプ表示
./build.sh help

# クリーンビルド（成果物削除）
./build.sh clean

# 通常ビルド
./build.sh
```

### ビルドプロセスの内訳

1. **バージョン取得**: `version.go` から動的に取得
2. **Goコンパイル**: Apple Silicon (arm64) 向けにビルド
3. **パッケージング**: 実行ファイル + config.json + README.md
4. **圧縮**: tar.gz形式で配布パッケージ作成
5. **配置**: `build/releases/` に移動

---

## トラブルシューティング

### エラー1: `go: command not found`

**原因**: Go がインストールされていない、またはPATHが通っていない

**解決策**:
```bash
# Goインストール確認
which go

# PATHに追加（.zshrcまたは.bash_profileに追記）
export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:$HOME/go/bin

# 設定反映
source ~/.zshrc
```

### エラー2: `pkg-config: command not found`

**原因**: pkg-config がインストールされていない

**解決策**:
```bash
brew install pkg-config
```

### エラー3: `portaudio.h: No such file or directory`

**原因**: PortAudio がインストールされていない

**解決策**:
```bash
brew install portaudio
```

### エラー4: `whisper-ctranslate2 not found`

**原因**: FasterWhisper がインストールされていない

**解決策**:
```bash
pip3 install faster-whisper

# インストール確認
which whisper-ctranslate2
```

### エラー5: 実行時に「開発元を確認できないため開けません」

**原因**: macOS Gatekeeper による実行制限

**解決策**:
```bash
# アプリに実行権限付与
xattr -cr koemoji-go-1.6.1/

# または個別に
xattr -d com.apple.quarantine koemoji-go-1.6.1/koemoji-go
```

### エラー6: マイクアクセス許可エラー

**原因**: 録音機能にマイク権限が必要

**解決策**:
1. システム環境設定 → セキュリティとプライバシー
2. プライバシー → マイク
3. アプリケーションを許可

---

## リリース成果物

### ファイル構成

```
koemoji-go-macos-1.6.1.tar.gz
└── koemoji-go-1.6.1/
    ├── koemoji-go           # 実行ファイル
    ├── config.json          # 設定ファイル（サンプル）
    └── README.md            # リリースノート
```

### ファイルサイズ目安

| ファイル | サイズ |
|---------|-------|
| koemoji-go (実行ファイル) | 約 15-20 MB |
| tar.gz (圧縮後) | 約 8-10 MB |

### 配布方法

1. **GitHub Releases にアップロード**
   ```bash
   # GitHubリポジトリページに移動
   # Releases → Draft a new release
   # アセットとしてアップロード: koemoji-go-macos-1.6.1.tar.gz
   ```

2. **ダウンロード後の手順**（ユーザー向け）
   ```bash
   # 解凍
   tar -xzf koemoji-go-macos-1.6.1.tar.gz

   # ディレクトリ移動
   cd koemoji-go-1.6.1

   # 実行権限付与（必要に応じて）
   chmod +x koemoji-go

   # 起動
   ./koemoji-go
   ```

---

## アーキテクチャ対応

### Apple Silicon (M1/M2/M3) - arm64

- **ビルドコマンド**: `./build.sh` または `./build.sh arm64`
- **GOARCH**: `arm64`
- **対応デバイス**: MacBook Air/Pro (M1以降), Mac mini (M1以降), iMac (M1以降)

### Intel Mac - amd64（現在非対応）

現在のビルドスクリプトは Apple Silicon (arm64) のみをサポートしています。
Intel Mac 向けビルドが必要な場合は、`build.sh` を以下のように変更してください：

```bash
# build_arch関数に追加
elif [ "$arch" = "amd64" ]; then
    GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.version=$VERSION" -o "$DIST_DIR/$binary_name" "$SOURCE_DIR"
```

---

## 開発者向けメモ

### クロスコンパイル

```bash
# macOS上で他プラットフォーム向けにビルド（参考）
GOOS=linux GOARCH=amd64 go build -o koemoji-go-linux ./cmd/koemoji-go
GOOS=windows GOARCH=amd64 go build -o koemoji-go.exe ./cmd/koemoji-go
```

ただし、CGO依存（PortAudio）があるため、実際のクロスコンパイルには各プラットフォームのCライブラリが必要です。

### ビルドフラグ説明

```bash
-ldflags="-s -w -X main.version=$VERSION"
```

- `-s`: シンボルテーブル削除（ファイルサイズ削減）
- `-w`: DWARFデバッグ情報削除（ファイルサイズ削減）
- `-X main.version=$VERSION`: バージョン番号を動的に注入

---

## 関連ドキュメント

- [バージョン更新チェックリスト](VERSION_UPDATE_CHECKLIST.md)
- [Windows ビルドガイド](WINDOWS_BUILD_GUIDE.md)
- [開発ガイド](DEVELOPMENT.md)
- [CLAUDE.md](../../CLAUDE.md)

---

**最終更新**: 2025-01-22
**対象バージョン**: v1.6.1以降
