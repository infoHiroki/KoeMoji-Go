# KoeMoji-Go

音声・動画ファイル自動文字起こしツール

## 概要

KoeMoji-Goは、音声・動画ファイルを自動で文字起こしするアプリケーションです。
Python版のKoeMojiAuto-cliをGoに移植し、シングルバイナリでの配布と順次処理による安定動作を実現しています。

### 特徴

- **シングルバイナリ**: 実行ファイル1つで動作
- **順次処理**: 1ファイルずつ安定した処理
- **FasterWhisper連携**: 高精度な音声認識
- **AI要約機能**: OpenAI APIによる自動要約生成
- **録音機能**: 内蔵マイク録音機能
- **GUI/TUI対応**: グラフィカル画面とターミナル画面の両対応
- **自動監視**: フォルダを定期的に監視して自動処理

## ⚡ クイックスタート

### 前提条件

**Python 3.8以上** が必要です。
```bash
python --version  # 3.8以上であることを確認
```

Pythonがない場合は [Python公式ダウンロード](https://www.python.org/downloads/) からインストールしてください。

### インストール

[GitHubリリース](https://github.com/infoHiroki/KoeMoji-Go/releases)から対応OS版をダウンロードして展開してください。

#### 🪟 Windows

1. **ダウンロード**: `koemoji-go-windows-1.5.1.zip`
2. **展開後の構成**:
   ```
   📁 koemoji-go-windows-1.5.1
   ├── koemoji-go.exe          # 実行ファイル（アイコン付き）
   ├── libportaudio.dll        # 録音機能用ライブラリ
   ├── libgcc_s_seh-1.dll      # GCCランタイム
   ├── libwinpthread-1.dll     # スレッドサポート
   ├── config.json             # 設定ファイル
   └── README.md               # 説明書
   ```
3. **実行**:
   ```cmd
   koemoji-go.exe
   ```

#### 🍎 macOS

1. **ダウンロード**:
   - **Apple Silicon (M1/M2)**: `koemoji-go-macos-arm64-1.5.1.tar.gz`

2. **展開後の構成**:
   ```
   📁 koemoji-go-macos-*-1.5.1
   ├── koemoji-go         # 実行ファイル
   ├── config.json        # 設定ファイル
   └── README.md          # 説明書
   ```

3. **実行**:
   ```bash
   ./koemoji-go
   ```

> **初回実行時**: FasterWhisperが自動インストールされます（数分かかります）

### 基本的な使い方

#### 1. 音声ファイルを処理する
1. `input/` フォルダに音声ファイル（MP3, WAV等）を置く
2. 自動的に処理が開始されます
3. 結果は `output/` フォルダに保存されます
4. 処理済みファイルは `archive/` に移動されます

#### 2. UIモード選択
- **GUI モード（デフォルト）**: ボタンクリックで操作
  ```bash
  ./koemoji-go
  ```
- **TUI モード**: キーボード操作
  ```bash
  ./koemoji-go --tui
  ```

#### 3. 主要な操作（TUIモード）
- `c` - 設定変更
- `l` - ログ表示  
- `s` - 手動スキャン
- `r` - 録音開始/停止
- `q` - 終了

#### 4. 対応ファイル形式
- **音声**: MP3, WAV, M4A, FLAC, OGG, AAC
- **動画**: MP4, MOV, AVI

#### 5. AI要約機能（オプション）
1. [OpenAI Platform](https://platform.openai.com/)でAPIキー取得
2. 設定画面（`c`）でAPIキーを入力
3. 文字起こし後に自動で要約が生成されます

## 📚 詳細情報

- **[🔧 トラブル解決](docs/user/TROUBLESHOOTING.md)** - 問題解決とFAQ
- **[🎤 システム音声＋マイク録音設定（Windows）](docs/user/SYSTEM_MIC_RECORDING_WINDOWS.md)** - Windows環境での同時録音設定
- **[📖 開発者向けドキュメント](docs/)** - ビルド方法、アーキテクチャ、技術仕様

## ライセンス

**個人利用**: 自由に使用可能  
**商用利用**: 事前連絡が必要

詳細は[LICENSE](LICENSE)ファイルをご確認ください。

## 作者

KoeMoji-Go開発チーム  
連絡先: koemoji2024@gmail.com