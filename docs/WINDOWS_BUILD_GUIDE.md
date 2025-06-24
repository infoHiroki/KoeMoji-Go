# Windows Build Guide for KoeMoji-Go

最終更新: 2025-06-24

## 概要

このドキュメントは、Windows環境でKoeMoji-Goをビルドするための実践的なガイドです。MSYS2を使用したビルド環境の構築から、配布パッケージの作成まで、実際に成功した手順を記載しています。

## 前提条件

### 必須要件
- Windows 10/11 (64bit)
- Go 1.21以上
- Python 3.8以上（FasterWhisper用）
- MSYS2（GCCツールチェーン用）

## セットアップ手順

### 1. Go言語のインストール

1. [Go公式サイト](https://golang.org/dl/)から最新版をダウンロード
2. インストーラーを実行
3. 確認:
   ```cmd
   go version
   ```

### 2. MSYS2のインストールとセットアップ

1. [MSYS2公式サイト](https://www.msys2.org/)からインストーラーをダウンロード
2. デフォルト設定でインストール（通常は`C:\msys64`）
3. MSYS2を起動し、以下のコマンドを実行:

```bash
# パッケージデータベースを更新
pacman -Syu

# MinGW-w64 GCCツールチェーンをインストール
pacman -S mingw-w64-x86_64-gcc mingw-w64-x86_64-make mingw-w64-x86_64-pkg-config

# PortAudioをインストール（録音機能に必要）
pacman -S mingw-w64-x86_64-portaudio
```

### 3. プロジェクトのセットアップ

1. プロジェクトをクローン:
   ```cmd
   git clone https://github.com/hirokitakamura/koemoji-go.git
   cd koemoji-go
   ```

2. goversioninfoをインストール（Windowsアイコン用）:
   ```cmd
   go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
   ```

## ビルド手順

### 自動ビルド（推奨）

`build`ディレクトリに作成された`build.bat`を使用:

```cmd
cd build
build.bat
```

このスクリプトは以下を自動的に実行します:
- 環境のチェック
- Windowsリソースファイルの生成（アイコン埋め込み）
- CGOを有効にしたビルド
- 必要なDLLのコピー
- 配布用ZIPパッケージの作成

### 手動ビルド

必要に応じて手動でビルドする場合:

```cmd
# MSYS2のパスを一時的に追加
set PATH=C:\msys64\mingw64\bin;%PATH%
set PKG_CONFIG_PATH=C:\msys64\mingw64\lib\pkgconfig

# GOPATHを設定（未設定の場合）
for /f "tokens=*" %i in ('go env GOPATH') do set GOPATH=%i

# リソースファイルを生成
cd build\templates\windows
%GOPATH%\bin\goversioninfo.exe -o ..\..\temp\resource.syso versioninfo.json
cd ..\..\..

# ビルド実行
cd cmd\koemoji-go
go build -ldflags="-s -w -H=windowsgui" -o ..\..\build\dist\koemoji-go.exe .
cd ..\..

# 必要なDLLをコピー
copy C:\msys64\mingw64\bin\libportaudio.dll build\dist\
copy C:\msys64\mingw64\bin\libgcc_s_seh-1.dll build\dist\
copy C:\msys64\mingw64\bin\libwinpthread-1.dll build\dist\
```

## 必要なDLLファイル

ビルドした実行ファイルには以下のDLLが必要です:
- `libportaudio.dll` - 音声入出力
- `libgcc_s_seh-1.dll` - GCCランタイム
- `libwinpthread-1.dll` - スレッドサポート

これらはMSYS2の`mingw64\bin`ディレクトリにあります。

## トラブルシューティング

### PortAudioが見つからないエラー

```
Package portaudio-2.0 was not found in the pkg-config search path
```

解決方法:
1. MSYS2でPortAudioをインストール:
   ```bash
   pacman -S mingw-w64-x86_64-portaudio
   ```

2. PKG_CONFIG_PATHを設定:
   ```cmd
   set PKG_CONFIG_PATH=C:\msys64\mingw64\lib\pkgconfig
   ```

### DLLが見つからないエラー

実行時に「libportaudio.dllが見つかりません」などのエラーが出る場合:

1. 必要なDLLを実行ファイルと同じディレクトリにコピー
2. または、`C:\msys64\mingw64\bin`をシステムのPATHに追加

### 録音デバイスのエラー

「no default input device」エラーの場合:
1. Windowsのサウンド設定でマイクが有効か確認
2. プライバシー設定でマイクへのアクセスが許可されているか確認

## 配布パッケージの構成

ビルド成功後、以下の構成で配布パッケージが作成されます:

```
koemoji-go-windows-1.5.0.zip
├── koemoji-go.exe          # 実行ファイル（アイコン付き）
├── libportaudio.dll        # PortAudioライブラリ
├── libgcc_s_seh-1.dll      # GCCランタイム
├── libwinpthread-1.dll     # Pthreadライブラリ
├── config.json             # 設定ファイル
└── README.md               # 説明書
```

## ビルドスクリプト

### build.bat

プロジェクトの`build`ディレクトリに配置されたWindows用ビルドスクリプトです。

主な機能:
- 環境の自動チェック
- MSYS2パスの自動設定
- goversioninfoの自動インストール
- DLLの自動コピー
- 配布パッケージの自動作成

使用方法:
```cmd
build.bat         # ビルド実行
build.bat clean   # クリーンアップ
build.bat help    # ヘルプ表示
```

### install_portaudio.bat

PortAudioを簡単にインストールするためのヘルパースクリプトです。

## 既知の問題と回避策

### GUI起動時のコンソールウィンドウ表示

## 参考リンク

- [MSYS2](https://www.msys2.org/)
- [Go CGO Documentation](https://golang.org/cmd/cgo/)
- [PortAudio](http://www.portaudio.com/)
- [Fyne Framework](https://fyne.io/)


**問題**: GUI版で「i」（入力ディレクトリを開く）、「o」（出力ディレクトリを開く）、「l」（ログを開く）を押すと、一時的にコンソールウィンドウが表示される。

**原因**: `exec.Command`で外部プログラム（explorer.exe、notepad.exe等）を起動する際、Windowsではデフォルトでコンソールウィンドウが表示される。

**回避策**:
1. 現状では仕様として受け入れる（一瞬表示されるだけで実害はない）
2. 将来の修正案:
   - `syscall.SysProcAttr`の`CREATE_NO_WINDOW`フラグを使用
   - VBScriptラッパーを作成して静かに実行

### 相対パスの解決問題

**問題**: プログラムを異なる場所から実行すると、`./input`、`./output`などの相対パスが意図しない場所を指す可能性がある。

**原因**: 相対パスは現在の作業ディレクトリ（CWD）を基準に解決されるが、実行方法によってCWDが異なる。

**回避策**:

1. **ショートカットを使用**（推奨）
   - koemoji-go.exeのショートカットを作成
   - プロパティで「作業フォルダー」を実行ファイルのあるディレクトリに設定

2. **バッチファイルでラップ**
   ```batch
   @echo off
   cd /d "%~dp0"
   start koemoji-go.exe
   ```
   このバッチファイルをkoemoji-go.exeと同じディレクトリに配置

3. **設定ファイルで絶対パスを使用**
   ```json
   {
     "input_dir": "C:\\path\\to\\koemoji-go\\input",
     "output_dir": "C:\\path\\to\\koemoji-go\\output",
     "archive_dir": "C:\\path\\to\\koemoji-go\\archive"
   }
   ```

**将来の修正案**:
- プログラム起動時に実行ファイルの場所を基準とした相対パス解決を実装
- `filepath.Abs()`と`os.Executable()`を組み合わせて使用

## 今後の改善点

1. **コンソールウィンドウの非表示化**
   - Windows固有のプロセス起動オプションを実装
   - 対象ファイル: `internal/ui/ui.go`

2. **パス解決の改善**
   - 実行ファイル基準の相対パス解決
   - 対象ファイル: `internal/config/config.go`

3. **GitHub Actions対応**: CI/CDパイプラインでの自動ビルド
4. **静的リンク**: DLLを含まない単一実行ファイルの作成
5. **インストーラー**: NSIS等を使用したインストーラーの作成
