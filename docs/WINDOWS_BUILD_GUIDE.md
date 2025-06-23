# KoeMoji-Go Windows ビルドガイド

## 概要
KoeMoji-GoをWindows環境でビルドして、完全機能版（録音・GUI・TUI・LLM要約）のバイナリを作成する手順書です。

## 前提条件
- Windows 10/11 (64bit)
- インターネット接続
- 管理者権限

## 手順

### 1. Git for Windows インストール
1. https://git-scm.com/download/win からダウンロード
2. インストール（デフォルト設定でOK）
3. Git Bashを起動して確認：
```bash
git --version
```

### 2. Go 言語環境セットアップ
1. https://golang.org/dl/ から Go 1.21+ をダウンロード
2. インストーラーを実行（デフォルト設定）
3. コマンドプロンプトで確認：
```cmd
go version
```

### 3. Visual Studio Build Tools インストール
1. https://visualstudio.microsoft.com/ja/downloads/ から「Build Tools for Visual Studio 2022」をダウンロード
2. インストール時に「C++ build tools」をチェック
3. 必要コンポーネント：
   - MSVC v143 - VS 2022 C++ x64/x86 build tools
   - Windows 10/11 SDK

### 4. PortAudio 依存関係セットアップ

#### 方法A: vcpkg 使用（推奨）
```cmd
# vcpkg をインストール
git clone https://github.com/Microsoft/vcpkg.git
cd vcpkg
.\bootstrap-vcpkg.bat
.\vcpkg.exe integrate install

# PortAudio をインストール
.\vcpkg.exe install portaudio:x64-windows

# 環境変数設定
set CGO_ENABLED=1
set PKG_CONFIG_PATH=%CD%\installed\x64-windows\lib\pkgconfig
```

#### 方法B: 手動インストール
1. http://files.portaudio.com/download.html からPortAudioソースをダウンロード
2. Visual Studioでビルド
3. 生成されたライブラリを適切な場所に配置

### 5. リポジトリクローンとビルド
```cmd
# プロジェクトをクローン
git clone https://github.com/infoHiroki/KoeMoji-Go.git
cd KoeMoji-Go

# goversioninfo インストール（Windowsアイコン用）
go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest

# Windowsリソースファイル生成
cd build
goversioninfo -o resource_windows_amd64.syso versioninfo.json
copy resource_windows_amd64.syso ..\cmd\koemoji-go\

# ビルド実行
cd ..
go build -ldflags="-s -w" -o koemoji-go.exe .\cmd\koemoji-go

# クリーンアップ
del cmd\koemoji-go\resource_windows_amd64.syso
del build\resource_windows_amd64.syso
```

### 6. 動作確認
```cmd
# バージョン確認
.\koemoji-go.exe --version

# ヘルプ表示
.\koemoji-go.exe --help

# 設定モード起動
.\koemoji-go.exe --configure
```

## トラブルシューティング

### CGO_ENABLED エラー
```cmd
set CGO_ENABLED=1
go env CGO_ENABLED
```

### PortAudio 関連エラー
```cmd
# pkg-config パス確認
echo %PKG_CONFIG_PATH%

# vcpkg 再インストール
vcpkg remove portaudio:x64-windows
vcpkg install portaudio:x64-windows
```

### ビルドエラー: 'gcc' not found
Visual Studio Build Toolsが正しくインストールされていません。
1. Visual Studio Installer を再実行
2. 「C++ build tools」を追加インストール

### Fyne GUI関連エラー
```cmd
# OpenGL ドライバーが古い場合
# グラフィックドライバーを最新に更新
```

## 成果物
ビルド成功時に以下が生成されます：
- `koemoji-go.exe` - 完全機能版バイナリ（約50-80MB）

## 機能確認
以下の機能が利用可能になります：
- ✅ **録音機能**: マイクからの直接録音
- ✅ **GUI**: Fyne ベースのグラフィカルUI
- ✅ **TUI**: ターミナルでの操作
- ✅ **文字起こし**: FasterWhisper による高精度変換
- ✅ **LLM要約**: OpenAI API との連携

## 配布準備
```cmd
# 配布用フォルダ作成
mkdir koemoji-go-windows-v1.5.0
copy koemoji-go.exe koemoji-go-windows-v1.5.0\
copy config.example.json koemoji-go-windows-v1.5.0\config.json
copy README.md koemoji-go-windows-v1.5.0\

# ZIP圧縮
powershell Compress-Archive -Path koemoji-go-windows-v1.5.0 -DestinationPath koemoji-go-windows-v1.5.0.zip
```

## 参考情報
- [Go CGO Documentation](https://golang.org/cmd/cgo/)
- [PortAudio Documentation](http://files.portaudio.com/docs/v19-doxydocs/)
- [Fyne Windows Building](https://docs.fyne.io/started/cross-compiling)

---
作成日: 2025-06-23  
対象バージョン: KoeMoji-Go v1.5.0