# Windows Build Guide for KoeMoji-Go

最終更新: 2025-06-24

## 概要

このドキュメントは、Windows環境でKoeMoji-Goをビルドするための実践的なガイドです。MSYS2を使用したビルド環境の構築から、配布パッケージの作成まで、実際に成功した手順を記載しています。

## 前提条件

### 必須要件
- Windows 10/11 (64bit)
- Go 1.21以上
- Python 3.12（推奨、3.13以降は非対応、FasterWhisper用）
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

### 手動ビルド（完璧な12ステップ手順）

アイコン付きリリースパッケージを作成する完全な手順：

#### 前提条件
- versioninfo.jsonのバージョンを更新済み
- build/common/assets/icons/icon.icoが存在
- goversioninfoがインストール済み

#### 手順

```bash
# プロジェクトルートから実行

# Step 1: クリーンアップ
rm -rf build/windows/dist build/windows/temp build/releases/koemoji-go-1.8.4.zip

# Step 2: ディレクトリ準備
mkdir -p build/windows/dist build/windows/temp build/releases

# Step 3: goversioninfoでresource.syso生成
cd build/common/templates/windows
goversioninfo -64 -o ../../../windows/temp/resource.syso versioninfo.json
cd ../../../..

# Step 4: resource.sysoをソースにコピー
cp build/windows/temp/resource.syso cmd/koemoji-go/

# Step 5: Goビルド（アイコン埋め込み）
cd cmd/koemoji-go
go build -ldflags "-s -w -H=windowsgui -X main.version=1.8.4" -o ../../build/windows/dist/koemoji-go.exe .
cd ../..

# Step 6: resource.syso削除（クリーンアップ）
rm -f cmd/koemoji-go/resource.syso

# Step 7: DLLコピー（build/windows/ディレクトリに配置済みのDLL）
cd build/windows
cp *.dll dist/
cd ../..

# Step 8: パッケージディレクトリ作成
cd build/windows/dist
mkdir -p koemoji-go-1.8.4

# Step 9: ファイルコピー
cp koemoji-go.exe koemoji-go-1.8.4/
cp *.dll koemoji-go-1.8.4/
cp ../診断実行.bat koemoji-go-1.8.4/
cp ../../common/assets/config.example.json koemoji-go-1.8.4/config.json
cp ../../common/assets/README_RELEASE.md koemoji-go-1.8.4/README.md

# Step 10: ZIP作成
powershell -Command "Compress-Archive -Path 'koemoji-go-1.8.4' -DestinationPath 'koemoji-go-1.8.4.zip' -Force"

# Step 11: releasesに移動
mv koemoji-go-1.8.4.zip ../../releases/

# Step 12: 確認
cd ../../..
ls -lh build/releases/koemoji-go-1.8.4.zip
```

#### 重要なポイント

1. **versioninfo.json更新**: ビルド前に必ずバージョン番号を更新
   - FileVersion/ProductVersion: `Major.Minor.Patch.Build`
   - StringFileInfo: バージョン文字列

2. **resource.syso**: アイコン埋め込み用、ビルド後は削除

3. **DLL配置**: 事前に`build/windows/`に3つのDLLを配置
   - libportaudio.dll
   - libgcc_s_seh-1.dll
   - libwinpthread-1.dll

4. **診断実行.bat**: v1.8.4以降、パッケージに含まれる

### 簡易手動ビルド（アイコンなし）

アイコン埋め込みが不要な場合の簡易手順:

```cmd
# MSYS2のパスを一時的に追加
set PATH=C:\msys64\mingw64\bin;%PATH%
set PKG_CONFIG_PATH=C:\msys64\mingw64\lib\pkgconfig

# ビルド実行
cd cmd\koemoji-go
go build -ldflags="-s -w -H=windowsgui -X main.version=1.8.4" -o ..\..\build\windows\dist\koemoji-go.exe .
cd ..\..

# 必要なDLLをコピー
copy C:\msys64\mingw64\bin\libportaudio.dll build\windows\dist\
copy C:\msys64\mingw64\bin\libgcc_s_seh-1.dll build\windows\dist\
copy C:\msys64\mingw64\bin\libwinpthread-1.dll build\windows\dist\
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
koemoji-go-1.8.3.zip
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

### ~~GUI起動時のコンソールウィンドウ表示~~ (v1.5.1で修正済み)

**問題**: GUI版で「i」（入力ディレクトリを開く）、「o」（出力ディレクトリを開く）、「l」（ログを開く）を押すと、一時的にコンソールウィンドウが表示される。

**状態**: ✅ **v1.5.1で修正済み** - `syscall.SysProcAttr`の`CREATE_NO_WINDOW`フラグを使用して解決

### ~~相対パスの解決問題~~ (v1.5.1で修正済み)

**問題**: プログラムを異なる場所から実行すると、`./input`、`./output`などの相対パスが意図しない場所を指す可能性がある。

**状態**: ✅ **v1.5.1で修正済み** - 実行ファイルのディレクトリを基準にパスを解決するように改善

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

## GitHub Actions でのビルド無効化について

### 背景
KoeMoji-Go は PortAudio を使用した録音機能を持つため、CGO（C言語バインディング）を必要とします。
これにより、GitHub Actions でのクロスコンパイルが困難になりました。

### 問題点
1. **Windows ビルド**: CGO_ENABLED=1 が必要なため、クロスコンパイルができない
2. **macOS ビルド**: 単一プラットフォームのみのビルドは CI/CD の価値が限定的

### 現在の対応
- GitHub Actions でのビルドワークフローを削除
- 各プラットフォームでのローカルビルドに移行

### 将来的な改善案
1. CGO 不要な代替ライブラリへの移行
2. プラットフォーム固有のランナーを使用した self-hosted runners
3. 録音機能を別プロセスとして分離

## トラブルシューティング

### ビルドスクリプトが途中で落ちる

**症状:**
- `build.bat`をダブルクリックするとウィンドウが即座に閉じる
- エラーメッセージが確認できない

**原因:**
- goversioninfo実行時のエラー
- パス指定の問題
- バッチファイル内のコマンドエラー

**解決方法:**

1. **コマンドプロンプトから実行してエラー確認**
   ```cmd
   cd C:\dev\KoeMoji-Go\build\windows
   build.bat
   ```

2. **環境チェックツールの実行**
   ```cmd
   cd C:\dev\KoeMoji-Go\build\windows
   check_env.bat
   ```
   すべての項目が`[OK]`であることを確認

3. **段階的テスト**
   - Goビルドのみ: `test_go_build.bat`
   - パッケージングのみ: `test_packaging_only.bat`

### goversioninfo エラー

**症状:**
```
Error: Failed to generate Windows resource file
```

**原因:**
- goversioninfoがインストールされていない
- アイコンファイルが見つからない

**解決方法:**

1. **goversioninfoの手動インストール**
   ```cmd
   go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
   ```

2. **アイコンなしでビルド（v1.6.0以降）**
   - build.batは自動的にアイコンなしで続行
   - `[WARNING] goversioninfo failed - continuing without icon`が表示される
   - 機能には影響なし

### DLLコピーエラー

**症状:**
```
Warning: Failed to copy DLL files
```

**原因:**
- DLLファイルが`build/windows/`ディレクトリに存在しない
- PortAudioがインストールされていない

**解決方法:**

1. **PortAudioの再インストール**
   ```bash
   # MSYS2 MinGW64で実行
   pacman -S mingw-w64-x86_64-portaudio
   ```

2. **DLLの手動コピー**
   ```cmd
   copy C:\msys64\mingw64\bin\libportaudio.dll build\windows\
   copy C:\msys64\mingw64\bin\libgcc_s_seh-1.dll build\windows\
   copy C:\msys64\mingw64\bin\libwinpthread-1.dll build\windows\
   ```

### ZIP作成エラー

**症状:**
```
Error: Failed to create ZIP package
```

**原因:**
- PowerShellの実行ポリシー制限
- ディスク容量不足

**解決方法:**

1. **実行ポリシーの確認**
   ```powershell
   Get-ExecutionPolicy
   ```
   `Restricted`の場合は変更:
   ```powershell
   Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
   ```

2. **手動でZIP作成**
   ```cmd
   cd build\windows\dist
   powershell -Command "Compress-Archive -Path 'koemoji-go-1.8.3' -DestinationPath 'koemoji-go-1.8.3.zip' -Force"
   ```

### アイコンが埋め込まれない

**症状:**
- exeファイルにアイコンが表示されない
- Windowsエクスプローラーで確認するとデフォルトアイコンが表示される

**原因:**
- goversioninfoが失敗した
- versioninfo.jsonの設定が間違っている

**確認方法:**
```cmd
# exeのプロパティを確認
右クリック → プロパティ → 詳細タブ
```

**解決方法:**

1. **ビルドログでgoversioninfoの状態を確認**
   - `[OK] Icon will be embedded in executable` → 成功
   - `[WARNING] goversioninfo failed` → 失敗（機能には影響なし）

2. **versioninfo.jsonの確認**
   ```cmd
   type build\common\templates\windows\versioninfo.json
   ```
   `IconPath`が正しく設定されているか確認

### 詳細なトラブルシューティング情報

より詳細な情報は以下のドキュメントを参照してください：
- [v1.6.0 Build System Fix](../progress/v1.6.0-build-system-fix.md) - ビルドシステムの問題と解決の詳細記録

## 今後の改善点

1. **GitHub Actions対応**: CI/CDパイプラインでの自動ビルド
2. **静的リンク**: DLLを含まない単一実行ファイルの作成（将来的な目標）
3. **インストーラー**: NSIS等を使用したインストーラーの作成
4. **デジタル署名**: 実行ファイルへのコード署名でセキュリティ警告を回避

## 参考リンク

- [MSYS2](https://www.msys2.org/)
- [Go CGO Documentation](https://golang.org/cmd/cgo/)
- [PortAudio](http://www.portaudio.com/)
- [Fyne Framework](https://fyne.io/)
- [goversioninfo](https://github.com/josephspurrier/goversioninfo)
