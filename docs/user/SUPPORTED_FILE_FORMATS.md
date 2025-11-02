# KoeMoji-Go 対応ファイル形式

KoeMoji-Goで文字起こし可能なファイル形式の詳細ドキュメントです。

## 対応ファイル形式一覧

### 音声ファイル（6形式）

| 拡張子 | 形式名 | 説明 | 推奨用途 |
|-------|--------|------|---------|
| `.mp3` | MP3 | 最も一般的な圧縮音声形式 | 汎用的な音声ファイル |
| `.wav` | WAV | 無圧縮音声形式 | 高品質音声、破損ファイルの復旧 |
| `.m4a` | M4A | Apple標準の音声形式（AAC） | iPhoneボイスメモ、Apple製品 |
| `.flac` | FLAC | 可逆圧縮音声形式 | 高品質でファイルサイズ削減 |
| `.ogg` | Ogg Vorbis | オープンソース圧縮音声形式 | ゲーム音声、配信音声 |
| `.aac` | AAC | 高効率圧縮音声形式 | スマートフォン録音 |

### 動画ファイル（3形式）

| 拡張子 | 形式名 | 説明 | 推奨用途 |
|-------|--------|------|---------|
| `.mp4` | MP4 | 最も一般的な動画形式 | スマホ撮影動画、Web動画 |
| `.mov` | QuickTime | Apple標準の動画形式 | Mac録画、iPhone撮影 |
| `.avi` | AVI | 古い標準動画形式 | Windows録画、古い動画 |

**合計9形式**に対応しています。

## 技術仕様

### 処理の仕組み

KoeMoji-GoはFasterWhisper（whisper-ctranslate2）を使用しており、内部的にFFmpegを利用して音声を抽出・変換します。

```
入力ファイル → FFmpeg（自動変換） → Whisper → 出力テキスト
```

### ファイル判定ロジック

ファイル形式は拡張子で判定されます（`internal/ui/ui.go:302-311`）：

```go
func IsAudioFile(filename string) bool {
    ext := strings.ToLower(filepath.Ext(filename))
    audioExts := []string{".mp3", ".wav", ".m4a", ".flac", ".ogg", ".aac", ".mp4", ".mov", ".avi"}
    for _, audioExt := range audioExts {
        if ext == audioExt {
            return true
        }
    }
    return false
}
```

**特徴**:
- 大文字小文字を区別しない（`.MP3`も`.mp3`も同様に処理）
- 動画ファイルは音声トラックのみを抽出して文字起こし
- 事前変換は不要（すべて自動処理）

## 使用方法

### 基本的な使い方

1. **ファイルを配置**: `input`フォルダに音声/動画ファイルを置く
2. **自動処理**: KoeMoji-Goが自動的に検出して文字起こし
3. **結果確認**: `output`フォルダに出力ファイルが生成
4. **アーカイブ**: 処理完了後、元ファイルは`archive`フォルダに移動

### 動画ファイルの処理

動画ファイル（`.mp4`, `.mov`, `.avi`）は音声トラックのみが抽出されます：

```bash
# 例: 会議録画の文字起こし
input/meeting_2025-11-02.mp4 → output/meeting_2025-11-02.txt
```

## 破損ファイルへの対応（v1.8.2以降）

### AAC/M4Aファイルの破損問題

スマートフォンで録音したAACファイル（`.aac`, `.m4a`）は、録音中に強制終了されると**終端マーカー（END element）が欠損**することがあります。

#### 症状

```
[aac @ 0x11e607b80] Input buffer exhausted before END element found
```

このようなファイルをWhisperが処理すると：
- エラーコード0で終了（一見成功）
- **0バイトの出力ファイル**を生成
- 従来は元ファイルが削除されてデータ消失

### データ消失防止機能（v1.8.2）

v1.8.2以降、以下の機能が追加されました（`internal/whisper/whisper.go:408-429`）：

#### 1. 出力ファイル検証

```go
// validateOutputFile checks if the output file exists and has content
func validateOutputFile(outputPath string, config *config.Config) error {
    fileInfo, err := os.Stat(outputPath)

    // ケース1: 出力ファイルが存在しない
    if os.IsNotExist(err) {
        return fmt.Errorf("出力ファイルが生成されませんでした...")
    }

    // ケース2: 出力ファイルが0バイト
    if fileInfo.Size() == 0 {
        return fmt.Errorf("出力ファイルが空です（0バイト）...")
    }

    return nil
}
```

#### 2. エラーメッセージ（日本語）

**ケース1: 出力ファイルが生成されない**
```
出力ファイルが生成されませんでした: output/file.txt

考えられる原因:
・音声/動画ファイルが破損している可能性があります
・音声認識エンジンが処理できない形式です

対処方法:
・WAV形式に変換してから再度処理してください

元ファイルはinputフォルダに保持されています。
```

**ケース2: 出力ファイルが0バイト**
```
出力ファイルが空です（0バイト）: output/file.txt

考えられる原因:
・音声/動画ファイルが破損している可能性があります
・音声が検出されませんでした

対処方法:
・WAV形式に変換してから再度処理してください

元ファイルはinputフォルダに保持されています。
```

#### 3. データ保護

- エラー検出時、元ファイルは**inputフォルダに保持**
- アーカイブ移動は行われない
- ユーザーは再試行可能

### 破損ファイルの復旧方法

#### macOS: FFmpegで変換

```bash
# 基本的な変換
ffmpeg -i input.aac output.wav

# Whisper向け最適化（16kHz、モノラル）
ffmpeg -i input.aac -ar 16000 -ac 1 output.wav

# 一括変換（カレントディレクトリの全AACファイル）
for file in *.aac; do
  ffmpeg -i "$file" "${file%.aac}.wav"
done
```

#### macOS: afconvert（標準ツール）

```bash
# 基本的な変換
afconvert input.aac output.wav -d LEI16 -f WAVE

# Whisper向け最適化
afconvert input.aac output.wav -d LEI16 -f WAVE -r 16000 -c 1
```

#### Windows: VLC（GUIツール）

1. VLCメディアプレーヤーを開く
2. メディア > 変換/保存
3. ファイルを追加してAACファイルを選択
4. 変換/保存をクリック
5. プロファイルで"Audio - CD"を選択
6. 保存先を指定して開始

## よくある質問

### Q1. 事前にWAV形式に変換する必要はありますか？

**A: いいえ、不要です。** 対応形式のファイルはそのまま`input`フォルダに置くだけで処理できます。

ただし、以下の場合はWAV変換を推奨：
- ファイルが破損している場合
- 文字起こしが失敗する場合
- スマートフォン録音のAACファイルでエラーが出る場合

### Q2. 動画ファイルはどのように処理されますか？

**A: 音声トラックのみが抽出されます。** 動画の映像部分は処理されず、音声のみが文字起こしされます。

### Q3. WAV形式のサイズが大きいのですが？

**A: WAVは無圧縮形式のためサイズが大きくなります。**

例:
- AAC: 36MB → WAV: 284MB（約8倍）
- AAC: 217MB → WAV: 1.7GB（約8倍）

**対策**:
- 文字起こし後、WAVファイルを削除
- 元のAACファイルをアーカイブに保持
- ディスク容量に余裕を持つ

### Q4. FFmpegはmacOSに標準インストールされていますか？

**A: いいえ、標準ではインストールされていません。**

- **FFmpeg**: Homebrew経由でインストール必要（`brew install ffmpeg`）
- **afconvert**: macOS標準ツール（常に利用可能）

### Q5. どの形式が最も推奨されますか？

**A: 用途により異なります。**

- **汎用性重視**: MP3（最も互換性が高い）
- **品質重視**: FLAC、WAV（ロスレス、無圧縮）
- **安定性重視**: WAV（破損に強い、復旧しやすい）
- **iPhone録音**: M4A（iPhoneボイスメモのデフォルト）

## 関連コミット履歴

### v1.8.2 (2025-11-02)

#### データ消失防止機能

**コミット**: `e444e5e` - 🔧 fix: 文字起こし失敗時のデータ消失を防止

**変更内容**:
- `validateOutputFile()`関数を追加
  - 出力ファイルの存在チェック
  - 0バイトファイルの検出
  - 多言語エラーメッセージ（日本語/英語）
- `TranscribeAudio()`関数にファイル検証処理を追加
- エラー時は元ファイルをinputフォルダに保持

**問題の背景**:
スマートフォン録音のAACファイル（終端マーカー欠損）を
whisper-ctranslate2が処理すると、エラーコード0で終了するが
0バイトファイルを生成する。従来は成功と判断され、元ファイルが
アーカイブに移動されてデータが消失していた。

**効果**:
- データ消失の完全防止
- ユーザーに明確な対処方法を提示
- 元ファイルの安全な保持

#### エラーメッセージ改善

**コミット**: `06cf7c4` - 🔧 fix: エラーメッセージをユーザーフレンドリーに改善

**変更内容**:
- 技術用語（FFmpeg、コマンド例）を削除
- 確証のない情報を削除
- シンプルで分かりやすい表現に統一
- 「音声ファイル」→「音声/動画ファイル」に汎用化
- ケース1とケース2のメッセージ構造を統一

**効果**:
- 非技術者にも分かりやすいエラーメッセージ
- 必要な情報のみを提示
- ユーザー体験の向上

## トラブルシューティング

### 問題1: 「Input buffer exhausted before END element found」エラー

**症状**:
```
[aac @ 0x11e607b80] Input buffer exhausted before END element found
```

**原因**: AACファイルの終端マーカー欠損（スマートフォン録音の強制終了など）

**解決策**:
1. FFmpegでWAV形式に変換
2. 変換したWAVファイルで再度文字起こし

### 問題2: 出力ファイルが0バイト

**症状**: 処理は成功するが、出力ファイルが空（0バイト）

**原因**: v1.8.2以降は自動検出、元ファイルを保持

**解決策**:
1. エラーメッセージを確認
2. WAV形式に変換して再試行
3. それでも失敗する場合、音声が含まれていない可能性

### 問題3: 処理が始まらない

**症状**: `input`フォルダにファイルを置いても処理されない

**原因**:
- サポート対象外の拡張子
- 大文字小文字の問題（稀）
- ファイル権限の問題

**解決策**:
1. 拡張子を確認（9形式のいずれかか）
2. デバッグモードで詳細確認: `./koemoji-go --debug`
3. ファイル権限を確認: `ls -la input/`

## 参考情報

### コード実装箇所

- **ファイル判定**: `internal/ui/ui.go:302-311`
- **出力ファイル検証**: `internal/whisper/whisper.go:408-429`
- **処理パイプライン**: `internal/processor/processor.go:92-104`

### 関連ドキュメント

- [基本的な使い方](BASIC_USAGE.md)
- [トラブルシューティング](TROUBLESHOOTING.md)
- [FasterWhisper技術仕様](../technical/FASTER_WHISPER.md)

### 外部リソース

- [FFmpeg公式ドキュメント](https://ffmpeg.org/documentation.html)
- [FasterWhisper GitHub](https://github.com/SYSTRAN/faster-whisper)
- [OpenAI Whisper公式](https://github.com/openai/whisper)

---

**最終更新**: 2025-11-02
**バージョン**: 1.8.2
**作成者**: @infoHiroki
