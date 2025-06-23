# KoeMoji-Go クイックスタートガイド

5分でKoeMoji-Goを始められる簡潔なガイドです。

## 前提条件

**Python 3.8以上** が必要です。
```bash
python --version  # 3.8以上であることを確認
```

Pythonがない場合：
- **Windows**: [Python公式サイト](https://www.python.org/downloads/windows/)
- **macOS**: `brew install python`

## 1. ダウンロード

[GitHubリリース](https://github.com/hirokitakamura/koemoji-go/releases)から対応OS版をダウンロード：

- **Windows**: `koemoji-go-windows-1.5.0.zip`
- **macOS Intel**: `koemoji-go-macos-intel-1.5.0.tar.gz`
- **macOS Apple Silicon**: `koemoji-go-macos-arm64-1.5.0.tar.gz`

## 2. 実行

ダウンロードファイルを展開し、実行：

**Windows**:
```cmd
koemoji-go.exe
```

**macOS**:
```bash
./koemoji-go
```

初回実行時、FasterWhisperが自動インストールされます（数分かかります）。

## 3. 基本的な使い方

### 音声ファイルを処理する
1. `input/` フォルダに音声ファイル（MP3, WAV等）を置く
2. 自動的に処理が開始されます
3. 結果は `output/` フォルダに保存されます
4. 処理済みファイルは `archive/` に移動されます

### 主要な操作
**GUI モード（デフォルト）**: ボタンクリックで操作
**TUI モード（--tui）**: キー操作
- `c` - 設定変更
- `l` - ログ表示  
- `s` - 手動スキャン
- `r` - 録音開始/停止（v1.4.0+）
- `q` - 終了

## 4. UI モード選択（v1.5.0+）

**デフォルト**: GUI（グラフィカル画面）で起動
```bash
./koemoji-go
```

**ターミナル**: コマンドライン画面で使用したい場合
```bash
./koemoji-go --tui
```

## 設定のカスタマイズ

実行中に `c` キーで設定画面を開き、以下を調整できます：

- **Whisperモデル**: `large-v3`（日本語推奨）
- **言語**: `ja`（日本語）
- **スキャン間隔**: `1`分
- **録音デバイス**: 使用するマイク
- **AI要約**: OpenAI APIキー設定

## トラブル時

- **起動しない**: [TROUBLESHOOTING.md](TROUBLESHOOTING.md) を確認
- **処理されない**: 対応ファイル形式（MP3, WAV, M4A, FLAC, OGG, AAC, MP4, MOV, AVI）か確認
- **遅い**: 設定でモデルを `medium` や `small` に変更

## AI要約機能（オプション）

1. [OpenAI Platform](https://platform.openai.com/)でAPIキー取得
2. 設定画面（`c`）で項目14にAPIキー入力
3. `a` キーでAI要約を有効化
4. 文字起こし後に自動で要約が生成されます

## 詳細情報

- **完全なマニュアル**: [README.md](README.md)
- **トラブル解決**: [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
- **開発者向け**: [docs/](docs/)

---

これで準備完了です！音声ファイルを `input/` に置いて、文字起こしを始めましょう。