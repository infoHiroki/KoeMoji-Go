# KoeMoji-Go v1.5.0

音声・動画ファイル自動文字起こしツール

## 前提条件

**Python 3.8以上** が必要です。
```bash
python --version  # 3.8以上であることを確認
```

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
- **TUI モード**: ターミナル画面
  ```bash
  ./koemoji-go --tui
  ```

### 4. 主要な操作（TUIモード）

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
- **トラブル解決**: https://github.com/infoHiroki/KoeMoji-Go/blob/main/TROUBLESHOOTING.md

## ライセンス

**個人利用**: 自由に使用可能  
**商用利用**: 事前連絡が必要

## 作者

KoeMoji-Go開発チーム  
連絡先: koemoji2024@gmail.com