# KoeMoji-Go

音声・動画ファイル自動文字起こしツール

## 前提条件

**Python 3.12** を推奨します（3.13以降は非対応）。
```bash
python --version  # 3.12であることを確認
```

> **重要**: Python 3.13/3.14や3.8など古いバージョンでは動作しません。

Pythonがない場合は [Python公式ダウンロード](https://www.python.org/downloads/) からインストールしてください。

## 使い方

### 1. 実行

**Windows**:
```cmd
koemoji-go.exe
```

**macOS**:
```bash
./koemoji-go
```

初回実行時、FasterWhisperが自動インストールされます（数分かかります）。

### 2. 音声ファイルを処理する

1. `input/` フォルダに音声ファイル（MP3, WAV等）を置く
2. 自動的に処理が開始されます
3. 結果は `output/` フォルダに保存されます
4. 処理済みファイルは `archive/` に移動されます

### 3. UIモード選択

- **GUI モード（デフォルト）**: グラフィカル画面
  ```bash
  ./koemoji-go
  ```
- **TUI モード（macOS専用）**: ターミナル画面
  ```bash
  ./koemoji-go --tui
  ```

> **注意**: TUIモードはmacOS専用です。WindowsではGUIモードをご利用ください。

### 4. 録音機能（オプション）

**シングル録音**: マイクのみの録音
**デュアル録音**: システム音声（YouTube等）+マイクの同時録音（macOS 13以降）

> ⚠️ **デュアル録音使用時の注意**
> ヘッドホン/イヤホンの使用を推奨します。
> スピーカーだと、マイクがスピーカーの音を拾い、音が二重になる場合があります。

### 5. 主要な操作（TUIモード）

- `c` - 設定変更
- `l` - ログ表示  
- `s` - 手動スキャン
- `r` - 録音開始/停止
- `q` - 終了

## 対応ファイル形式

- **音声**: MP3, WAV, M4A, FLAC, OGG, AAC
- **動画**: MP4, MOV, AVI

## AI要約機能（オプション）

1. [OpenAI Platform](https://platform.openai.com/)でAPIキー取得
2. 設定画面（`c`）でAPIキーを入力
3. 文字起こし後に自動で要約が生成されます

## トラブル時

- **起動しない**: Pythonがインストールされているか確認
- **処理されない**: 対応ファイル形式か確認
- **遅い**: 設定でモデルを `medium` や `small` に変更

## 詳細情報

- **完全なマニュアル**: https://github.com/infoHiroki/KoeMoji-Go
- **トラブル解決**: https://github.com/infoHiroki/KoeMoji-Go/issues

## ライセンス

**個人利用**: 自由に使用可能  
**商用利用**: 事前連絡が必要

## 作者

KoeMoji-Go開発チーム  
連絡先: koemoji2024@gmail.com