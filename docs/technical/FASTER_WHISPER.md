# FasterWhisper 技術ドキュメント

最終更新: 2025-10-24
調査日: 2025-10-24
対象バージョン: faster-whisper 1.2.0, whisper-ctranslate2 0.5.4

---

## 目次

1. [概要](#概要)
2. [faster-whisper vs whisper-ctranslate2](#faster-whisper-vs-whisper-ctranslate2)
3. [Python要件](#python要件)
4. [インストール方法](#インストール方法)
5. [パフォーマンス](#パフォーマンス)
6. [KoeMoji-Goでの実装](#koemoji-goでの実装)
7. [トラブルシューティング](#トラブルシューティング)

---

## 概要

### FasterWhisperとは

**faster-whisper**は、OpenAIのWhisperモデルを**CTranslate2**推論エンジンで再実装したPythonライブラリです。

**開発元**: SYSTRAN社
**公式リポジトリ**: https://github.com/SYSTRAN/faster-whisper
**PyPI**: https://pypi.org/project/faster-whisper/

### 主な特徴

- ⚡ **最大4倍高速** - OpenAI公式実装比（同精度）
- 💾 **メモリ使用量削減** - 8bit量子化対応
- 🎯 **バッチ処理で12.5倍高速化** - 最適化時
- 📦 **FFmpeg不要** - PyAVがFFmpegをバンドル

### OpenAI Whisperとの比較

| 項目 | OpenAI Whisper | faster-whisper |
|------|---------------|----------------|
| 速度 | 1x（基準） | 4x（最大12.5x） |
| メモリ | 高 | 低 |
| 精度 | 基準 | 同等 |
| FFmpeg | 必要 | 不要 |
| GPU対応 | CUDA | CUDA + 量子化 |

---

## faster-whisper vs whisper-ctranslate2

### 2つのパッケージの関係

```
faster-whisper (ライブラリ)
    ↑ 依存
whisper-ctranslate2 (CLIツール)
```

### 違い

| 項目 | faster-whisper | whisper-ctranslate2 |
|------|---------------|---------------------|
| **種類** | Pythonライブラリ | コマンドラインツール |
| **用途** | プログラムから呼び出し | ターミナルから実行 |
| **インストール** | `pip install faster-whisper` | `pip install whisper-ctranslate2` |
| **依存関係** | ctranslate2, PyAV等 | faster-whisperに依存 |
| **使用例** | `from faster_whisper import WhisperModel` | `whisper-ctranslate2 audio.mp3` |

### whisper-ctranslate2の特徴

**OpenAI Whisper CLIとの互換性**:
```bash
# OpenAI Whisper
whisper audio.mp3 --model large-v3 --language ja

# whisper-ctranslate2（同じコマンド）
whisper-ctranslate2 audio.mp3 --model large-v3 --language ja
```

**追加機能**:
- バッチ処理モード（`--batched True`）
- 話者分離（Speaker Diarization）
- VAD（音声検出）フィルター
- カラーコード付き出力

### KoeMoji-Goでの使用

KoeMoji-Goは**whisper-ctranslate2コマンド**を実行します：

```go
// internal/whisper/whisper.go
whisperCmd := "whisper-ctranslate2"
args := []string{
    "--model", config.WhisperModel,
    "--language", config.Language,
    "--compute_type", config.ComputeType,
    inputFile,
}
cmd := exec.Command(whisperCmd, args...)
```

**重要**: 両方のパッケージが必要
```bash
pip install faster-whisper whisper-ctranslate2
```

---

## Python要件

### サポートバージョン（2025年10月時点）

| Pythonバージョン | faster-whisper | ctranslate2 | 状態 |
|-----------------|----------------|-------------|------|
| **3.8** | ❌ | ❌ | 非サポート |
| **3.9** | ✅ | ✅ | サポート |
| **3.10** | ✅ | ✅ | サポート |
| **3.11** | ✅ | ✅ | サポート |
| **3.12** | ✅ | ✅ | **推奨** |
| **3.13** | ❌ | ❌ | **非サポート** |

### Python 3.13が非サポートの理由

- **ctranslate2**がPython 3.13用のwheelを提供していない
- ソース配布もなし
- 2025年10月時点で対応予定なし

### KoeMoji-Goの「Python 3.12」推奨の妥当性

✅ **完全に妥当**:
- Python 3.12は現時点で最新の**安定サポートバージョン**
- Python 3.13は非サポート
- 動作確認済みバージョンを推奨するのは正しい方針
- 3.9-3.11でも動作するが、3.12が最も新しく推奨

---

## インストール方法

### 基本インストール

```bash
pip install faster-whisper whisper-ctranslate2
```

### FFmpegについて

**重要**: faster-whisperは**FFmpegのシステムインストール不要**

- **PyAV**ライブラリがFFmpegライブラリをバンドル
- OpenAI Whisperと異なり、より簡単なセットアップ

### GPU対応（CUDA）

#### CUDA 12 + cuDNN 9（最新、推奨）

```bash
pip install nvidia-cublas-cu12 nvidia-cudnn-cu12==9.*

# Linux環境変数設定
export LD_LIBRARY_PATH=`python3 -c 'import os; import nvidia.cublas.lib; import nvidia.cudnn.lib; print(os.path.dirname(nvidia.cublas.lib.__file__) + ":" + os.path.dirname(nvidia.cudnn.lib.__file__))'`
```

#### CUDA 11 + cuDNN 8（旧バージョン）

```bash
pip install ctranslate2==3.24.0  # ダウングレード
pip install faster-whisper whisper-ctranslate2
```

#### CUDA 12 + cuDNN 8

```bash
pip install ctranslate2==4.4.0  # ダウングレード
pip install faster-whisper whisper-ctranslate2
```

### インストール確認

```bash
# パッケージ確認
pip show faster-whisper
pip show whisper-ctranslate2

# コマンド確認
whisper-ctranslate2 --help

# Pythonから確認
python -c "from faster_whisper import WhisperModel; print('OK')"
```

---

## パフォーマンス

### 速度比較（2025年）

| 実装 | 相対速度 | 備考 |
|------|---------|------|
| OpenAI Whisper | 1x（基準） | オリジナル実装 |
| **faster-whisper** | **4x** | 単純な置き換え |
| **faster-whisper（バッチ）** | **12.5x** | 最適化時 |
| Whisper Large V3 Turbo | 5.4x | OpenAI新モデル（2024-2025） |

### メモリ使用量

- **faster-whisper**: 低 - 8bit量子化可能（`--compute_type int8`）
- **OpenAI Whisper**: 高 - フルプレシジョン

### 精度

- **同等** - faster-whisperは精度を犠牲にしない
- 品質テストで確認済み
- 前のセグメントテキストを含めると品質がさらに向上

---

## KoeMoji-Goでの実装

### インストール処理

```go
// internal/whisper/whisper.go:136-144
func installFasterWhisper(log *log.Logger, logBuffer *[]logger.LogEntry, logMutex *sync.RWMutex) error {
    logger.LogInfo(log, logBuffer, logMutex, "Installing faster-whisper and whisper-ctranslate2...")
    cmd := createCommand("pip", "install", "faster-whisper", "whisper-ctranslate2")
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("pip install failed: %w", err)
    }
    logger.LogInfo(log, logBuffer, logMutex, "FasterWhisper installed successfully")
    return nil
}
```

✅ **正しい実装**: 両方のパッケージをインストール

### コマンド検索

```go
// internal/whisper/whisper.go:22-114
func getWhisperCommand() string {
    // 1. PATHで検索
    if _, err := exec.LookPath("whisper-ctranslate2"); err == nil {
        return "whisper-ctranslate2"
    }

    // 2. 標準的なインストール場所を検索
    // Windows: C:\Users\...\Python312\Scripts\whisper-ctranslate2.exe
    // macOS: ~/Library/Python/3.12/bin/whisper-ctranslate2
    // ...
}
```

✅ **堅牢な実装**: 複数の場所を検索

### コマンド実行

```go
// internal/whisper/whisper.go:180-198
whisperCmd := getWhisperCommandWithDebug(log, logBuffer, logMutex, debugMode)

args := []string{
    "--model", config.WhisperModel,
    "--language", config.Language,
    "--output_dir", config.OutputDir,
    "--output_format", config.OutputFormat,
    "--compute_type", config.ComputeType,
}

// CPU使用を明示（int8の場合）
if config.ComputeType == "int8" {
    args = append(args, "--device", "cpu")
}

args = append(args, "--verbose", "True", inputFile)
cmd := createCommand(whisperCmd, args...)
```

✅ **正しいオプション使用**: whisper-ctranslate2の標準オプション

### 利用可能なモデル

KoeMoji-Goがサポートするモデル：

| モデル | サイズ | 速度 | 精度 | 用途 |
|--------|--------|------|------|------|
| tiny | 39M | 最速 | 低 | テスト用 |
| base | 74M | 高速 | 中 | 軽量処理 |
| small | 244M | 中速 | 中高 | バランス型 |
| medium | 769M | 中低速 | 高 | 高精度 |
| large-v2 | 1550M | 低速 | 最高 | 旧最高精度 |
| **large-v3** | 1550M | 低速 | **最高** | **デフォルト（推奨）** |

**KoeMoji-Goのデフォルト**: `large-v3`
**設定場所**: `config.json` → `whisper_model`

---

## トラブルシューティング

### 1. "whisper-ctranslate2 not found"

**原因**:
- Python未インストール
- faster-whisper/whisper-ctranslate2未インストール
- PATHに含まれていない

**解決方法**:
```bash
# インストール
pip install faster-whisper whisper-ctranslate2

# 確認
pip show faster-whisper
pip show whisper-ctranslate2

# コマンド確認（Windows）
where whisper-ctranslate2

# コマンド確認（macOS/Linux）
which whisper-ctranslate2
```

### 2. Python 3.13で動作しない

**原因**: ctranslate2がPython 3.13未対応

**解決方法**:
```bash
# Python 3.12をインストール
# KoeMoji-Goを再起動（自動インストールが再試行される）
```

### 3. GPU使用時のエラー

**エラー例**:
```
Device or backend do not support efficient int8_float16 computation
```

**解決方法**:
```json
// config.json
{
  "compute_type": "int8"  // CPU使用（最も安定）
}
```

KoeMoji-Goは`int8`設定時に自動的に`--device cpu`を追加します。

### 4. ネットワークエラー

**エラー例**:
```
pip install failed: connection timeout
```

**解決方法**:
1. ネットワーク接続を確認
2. プロキシ設定を確認
3. KoeMoji-Goを再起動（自動インストールが再試行される）

### 5. 権限エラー

**エラー例**:
```
pip install failed: permission denied
```

**解決方法**:
```bash
# 管理者権限で手動インストール（Windows）
# PowerShellを管理者として実行
pip install faster-whisper whisper-ctranslate2

# macOS/Linux
sudo pip install faster-whisper whisper-ctranslate2
# または
pip install --user faster-whisper whisper-ctranslate2
```

---

## 参考リンク

### 公式ドキュメント
- [faster-whisper GitHub](https://github.com/SYSTRAN/faster-whisper)
- [faster-whisper PyPI](https://pypi.org/project/faster-whisper/)
- [whisper-ctranslate2 GitHub](https://github.com/Softcatala/whisper-ctranslate2)
- [whisper-ctranslate2 PyPI](https://pypi.org/project/whisper-ctranslate2/)
- [CTranslate2 Documentation](https://opennmt.net/CTranslate2/)

### KoeMoji-Go関連
- [README.md](../../README.md)
- [TROUBLESHOOTING.md](../user/TROUBLESHOOTING.md)
- [内部実装: internal/whisper/whisper.go](../../internal/whisper/whisper.go)

---

## 更新履歴

| 日付 | バージョン | 変更内容 |
|------|-----------|---------|
| 2025-10-24 | 1.0 | 初版作成（faster-whisper 1.2.0, Python 3.9-3.12対応） |
