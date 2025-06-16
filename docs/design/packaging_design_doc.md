# KoeMoji-Go 配布戦略設計書

## 1. 概要

### 1.1 基本方針
KoeMoji-GoはCLIアプリケーションであり、ターミナルからの実行が前提である。
この特性を踏まえ、プラットフォーム別に最適化された配布方法を採用する。

### 1.2 プラットフォーム別配布方針
- **Windows**: アイコン付きEXEファイル（Explorer表示の向上）
- **macOS/Linux**: シンプルなバイナリファイル（ターミナル実行）

## 2. 技術仕様

### 2.1 Windows用アイコン埋め込み

#### 実装理由
- ExplorerでファイルアイコンとしてKoeMoji-Goロゴが表示される
- タスクバーでプロセス実行時にアイコンが表示される
- プロフェッショナルな見た目の向上

#### versioninfo.json（最小構成）
```json
{
    "FixedFileInfo": {
        "FileVersion": {"Major": 1, "Minor": 0, "Patch": 0, "Build": 0},
        "ProductVersion": {"Major": 1, "Minor": 0, "Patch": 0, "Build": 0}
    },
    "StringFileInfo": {
        "FileDescription": "KoeMoji-Go Audio/Video Transcription Tool",
        "ProductName": "KoeMoji-Go",
        "ProductVersion": "1.0.0",
        "FileVersion": "1.0.0.0",
        "OriginalFilename": "koemoji-go.exe",
        "InternalName": "koemoji-go",
        "CompanyName": "KoeMoji-Go Development Team",
        "LegalCopyright": "Copyright (c) 2025 KoeMoji-Go Development Team"
    },
    "IconPath": "icon.ico"
}
```

#### 実装方法
1. goversioninfoツールのインストール
   ```bash
   go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
   ```
2. リソースファイル生成
   ```bash
   goversioninfo -o resource.syso versioninfo.json
   ```
3. Windowsビルド
   ```bash
   GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o koemoji-go.exe main.go
   ```

### 2.2 macOS/Linux用シンプルバイナリ

#### 実装理由
- CLIアプリとして適切な配布形式
- ユーザーの期待値と実際の動作が一致
- 不要な複雑さを避ける

#### ビルド方法
```bash
# macOS
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o koemoji-go-darwin-amd64 main.go
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o koemoji-go-darwin-arm64 main.go

# Linux
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o koemoji-go-linux-amd64 main.go
```

## 3. ビルドスクリプト設計

### 3.1 シンプルなbuild.sh
```bash
#!/bin/bash
set -e

VERSION="1.0.0"
APP_NAME="koemoji-go"
DIST_DIR="dist"

echo "🚀 Building KoeMoji-Go..."

# Clean and prepare
rm -rf $DIST_DIR
mkdir -p $DIST_DIR

# Windows with icon
echo "🪟 Building Windows with icon..."
if ! command -v goversioninfo &> /dev/null; then
    go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
fi

$(go env GOPATH)/bin/goversioninfo -o resource.syso versioninfo.json
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $DIST_DIR/${APP_NAME}.exe main.go
rm -f resource.syso

# macOS
echo "🍎 Building macOS..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $DIST_DIR/${APP_NAME}-darwin-amd64 main.go
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $DIST_DIR/${APP_NAME}-darwin-arm64 main.go

# Linux
echo "🐧 Building Linux..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $DIST_DIR/${APP_NAME}-linux-amd64 main.go

echo "✅ Build completed!"
ls -la $DIST_DIR/
```

## 4. 配布パッケージ構成

### 4.1 Windows配布
```
koemoji-go-windows-v1.0.0.zip
├── koemoji-go.exe       # アイコン付きEXE
├── config.json
└── README.md
```

### 4.2 macOS配布
```
koemoji-go-macos-v1.0.0.tar.gz
├── koemoji-go-darwin-amd64    # Intel Mac用
├── koemoji-go-darwin-arm64    # Apple Silicon用
├── config.json
└── README.md
```

### 4.3 Linux配布
```
koemoji-go-linux-v1.0.0.tar.gz
├── koemoji-go-linux-amd64
├── config.json
└── README.md
```

## 5. 使用方法

### 5.1 Windows
1. ZIPファイルを解凍
2. PowerShellまたはコマンドプロンプトで実行
   ```cmd
   cd koemoji-go-windows-v1.0.0
   .\koemoji-go.exe
   ```

### 5.2 macOS
1. tar.gzファイルを解凍
2. ターミナルで実行
   ```bash
   cd koemoji-go-macos-v1.0.0
   # Intel Mac
   ./koemoji-go-darwin-amd64
   # Apple Silicon
   ./koemoji-go-darwin-arm64
   ```

### 5.3 Linux
1. tar.gzファイルを解凍
2. ターミナルで実行
   ```bash
   cd koemoji-go-linux-v1.0.0
   ./koemoji-go-linux-amd64
   ```

## 6. 実装影響範囲

### 6.1 変更なし
- ✅ `main.go` - コード変更なし
- ✅ `config.json` - 設定変更なし
- ✅ 実行時動作 - 機能変更なし
- ✅ 依存関係 - ライブラリ追加なし

### 6.2 追加ファイル
- `build.sh` - ビルドスクリプト
- `versioninfo.json` - Windows用リソース設定（復活時）

## 7. 期待される効果

### 7.1 ユーザー体験
- **Windows**: アイコン表示によるプロフェッショナルな見た目
- **macOS/Linux**: CLIアプリとして適切な配布形式
- **全体**: プラットフォームに適した配布方法

### 7.2 開発・保守性
- シンプルな構成で保守しやすい
- 各プラットフォームの特性に合わせた最適化
- 不要な複雑さを排除

## 8. 実装手順

1. versioninfo.jsonの作成（Windows用）
2. シンプルなbuild.shの作成
3. 各プラットフォームでのビルドテスト
4. 配布パッケージ作成の自動化
5. READMEの更新

---

作成日: 2025/06/08
更新日: 2025/06/08  
バージョン: 2.0 (CLI特性に基づく再設計)