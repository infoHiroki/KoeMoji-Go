# macOSデュアル録音機能 手動テスト手順

## 前提条件
- ✅ macOS 13以上
- ✅ Swift CLIバイナリ: `cmd/audio-capture/audio-capture` (171KB)
- ✅ 録音デバイス: 5台検出（koemoji集約デバイス含む）

## テスト1: GUIモード（デュアル録音無効）

### 手順
1. config.jsonで設定確認：
   ```json
   {
     "dual_recording_enabled": false  // または項目自体がない
   }
   ```

2. GUI起動：
   ```bash
   ./koemoji-go
   ```

3. 確認事項：
   - [ ] アプリが正常に起動する
   - [ ] 録音ボタンが表示される
   - [ ] 録音開始→3秒待機→録音停止
   - [ ] ログに「録音を開始しました」表示
   - [ ] `input/` ディレクトリに1ファイル生成
   - [ ] ファイル名: `recording-YYYYMMDD-HHMMSS.wav`

### 期待結果
- 単一ファイル録音成功（マイク音声のみ）
- ファイルサイズ: 約250KB/3秒

---

## テスト2: GUIモード（デュアル録音有効）

### 手順
1. config.jsonで設定変更：
   ```json
   {
     "dual_recording_enabled": true,
     "recording_device_name": "koemoji"
   }
   ```

2. GUI再起動：
   ```bash
   ./koemoji-go
   ```

3. 確認事項：
   - [ ] アプリが正常に起動する
   - [ ] ログに「デュアル録音モード: システム音声 + マイク」表示
   - [ ] システム音声を再生しながら録音開始
   - [ ] 同時にマイクに向かって話す
   - [ ] 3秒待機→録音停止
   - [ ] `input/` ディレクトリに**1ファイル**生成（自動ミックス済み）

### 期待結果
- **1ファイル生成（ミックス済み）**:
  - `recording-YYYYMMDD-HHMMSS.wav`
     - サイズ: 約500KB/3秒
     - フォーマット: 48kHz, Int16, Stereo
     - 内容: システム音声（70%）+ マイク音声（100%）の自動ミックス

### 確認コマンド
```bash
# ファイル確認
ls -lh input/

# オーディオ情報確認（macOS）
afinfo input/recording-*.wav

# 再生テスト
afplay input/recording-*-system.wav  # システム音声
afplay input/recording-*.wav         # マイク音声（systemなし）
```

---

## テスト3: TUIモード（デュアル録音有効）

### 手順
1. config.jsonでデュアル録音有効を確認

2. TUI起動：
   ```bash
   ./koemoji-go --tui
   ```

3. 操作：
   - [ ] アプリが正常に起動する
   - [ ] ログエリアに「デュアル録音モード: システム音声 + マイク」表示
   - [ ] `[R]キーで録音開始`を押す
   - [ ] 3秒待機
   - [ ] 再度`[R]キーで録音停止`を押す
   - [ ] Ctrl+Cで終了

### 期待結果
- GUIモードと同様に1ファイル生成（自動ミックス済み）
- TUIログに録音開始/停止メッセージ表示

---

## テスト4: エラーハンドリング

### テスト4-1: Swift CLIバイナリがない場合
```bash
# バイナリを一時的にリネーム
mv cmd/audio-capture/audio-capture cmd/audio-capture/audio-capture.bak

# GUI起動
./koemoji-go
```

**期待結果**:
- [ ] エラーメッセージ表示: "デュアル録音の初期化に失敗: audio-capture binary not found"
- [ ] アプリは起動するが録音不可

```bash
# 元に戻す
mv cmd/audio-capture/audio-capture.bak cmd/audio-capture/audio-capture
```

### テスト4-2: 不正なデバイス名
config.jsonで設定：
```json
{
  "dual_recording_enabled": true,
  "recording_device_name": "存在しないデバイス"
}
```

**期待結果**:
- [ ] エラーメッセージ表示: "recording device not found: '存在しないデバイス'"
- [ ] アプリは起動するが録音不可

---

## テスト5: 画面収録権限の確認

### 初回起動時
1. デュアル録音有効でGUI起動
2. 録音開始

**期待結果**:
- [ ] macOSシステムダイアログ表示
- [ ] 「"koemoji-go"が画面を収録しようとしています」
- [ ] 「システム環境設定」で許可を促すメッセージ
- [ ] 許可後、録音が正常動作

### 権限確認方法
```
システム設定 > プライバシーとセキュリティ > 画面収録
→ koemoji-go がリストに表示され、チェックがON
```

---

## テスト結果サマリー

### ✅ 成功基準
- [ ] デュアル録音無効: 単一ファイル生成
- [ ] デュアル録音有効: 1ファイル生成（自動ミックス済み）
- [ ] GUIモード動作正常
- [ ] TUIモード動作正常
- [ ] エラーハンドリング適切
- [ ] 画面収録権限の取得成功

### 📊 テスト実施日時
- 実施日: YYYY-MM-DD
- 実施者:
- macOSバージョン:

### 🐛 発見した問題
（あれば記載）

---

## 参考: ファイルフォーマット確認コマンド

```bash
# マイク音声
afinfo input/recording-20250127-120000.wav
# → Format: Linear PCM, 44100 Hz, Mono, 16-bit

# システム音声
afinfo input/recording-20250127-120000-system.wav
# → Format: Linear PCM (Float32), 48000 Hz, Stereo

# ファイルサイズ比較
ls -lh input/recording-*
```

## クリーンアップ

テスト後、生成されたファイルを削除：
```bash
rm -f input/recording-*.wav
rm -f /tmp/gui-integration-test*.wav
```
