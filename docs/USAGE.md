# KoeMoji-Go 使用方法

## 基本起動パターン

### 1. 通常モード（転写処理実行）

```bash
./koemoji-go
```

**フロー:**
```
起動 → 設定読み込み → ディレクトリ監視開始 → 対話コマンド待機
                                     ↓
                               ファイル検出時
                                     ↓
                            転写処理 → アーカイブ移動
```

**対話コマンド:**
- `c` → 設定表示 → Enter待機 → メイン画面復帰
- `l` → ログ表示 → Enter待機 → メイン画面復帰  
- `s` → 手動スキャン実行 → 処理続行
- `q` → アプリケーション終了
- `Enter` → 画面リフレッシュ

### 2. 設定モード（対話的設定変更）

```bash
./koemoji-go --configure
```

**メイン設定画面:**
```
=== KoeMoji-Go Configuration ===
1. Whisper Model: medium
2. Language: ja
3. Scan Interval: 10 minutes
4. Max CPU Percent: 95%
5. Compute Type: int8
6. Use Colors: true
7. UI Mode: enhanced
8. Output Format: txt
9. Input Directory: ./input
10. Output Directory: ./output
11. Archive Directory: ./archive
r. Reset to defaults
s. Save and exit
q. Quit without saving

Select option (1-11, r, s, q): 
```

### 3. その他オプション

```bash
./koemoji-go --version    # バージョン表示
./koemoji-go --help       # ヘルプ表示
./koemoji-go --debug      # デバッグモード
./koemoji-go --config custom.json  # カスタム設定ファイル
```

## 対話型設定の詳細使用例

### Whisperモデル選択（オプション1）

```
Select option (1-11, r, s, q): 1

Available Whisper models:
1. tiny
2. tiny.en
3. base
4. base.en
5. small
6. small.en
7. medium (current)
8. medium.en
9. large
10. large-v1
11. large-v2
12. large-v3
Select model (1-12) or press Enter to keep current: 12
Whisper model set to: large-v3
```

### CPU使用率設定（オプション4）

```
Select option (1-11, r, s, q): 4
Current max CPU percent: 95%
Enter new max CPU percent (1-100) or press Enter to keep current: 80
Max CPU percent set to: 80%
```

**エラーケース:**
```
Enter new max CPU percent (1-100) or press Enter to keep current: 150
Invalid input. Please enter a number between 1 and 100.
```

### 計算タイプ選択（オプション5）

```
Select option (1-11, r, s, q): 5

Available compute types:
1. int8 (current)
2. int8_float16
3. int16
4. float16
5. float32
Select compute type (1-5) or press Enter to keep current: 2
Compute type set to: int8_float16
```

### ディレクトリパス設定（オプション9-11）

```
Select option (1-11, r, s, q): 9
Current input directory: ./input
Enter new input directory path or press Enter to keep current: /Users/user/Audio/Input
Input directory set to: /Users/user/Audio/Input
```

### リセット機能（オプションr）

```
Select option (1-11, r, s, q): r
Are you sure you want to reset all settings to defaults? (y/N): y
Configuration reset to defaults.
```

**キャンセル例:**
```
Are you sure you want to reset all settings to defaults? (y/N): n
(メイン画面に戻る)
```

### 保存・終了パターン

**変更ありで保存:**
```
Select option (1-11, r, s, q): s
Configuration saved successfully!
```

**変更なしで保存:**
```
Select option (1-11, r, s, q): s
No changes to save.
```

**未保存で終了しようとした場合:**
```
Select option (1-11, r, s, q): q
You have unsaved changes. Are you sure you want to quit? (y/N): n
(メイン画面に戻る)
```

**未保存で強制終了:**
```
You have unsaved changes. Are you sure you want to quit? (y/N): y
(設定を保存せずに終了)
```

## 実行時の分岐フローチャート

```
./koemoji-go --configure
         ↓
    設定画面表示
         ↓
    ユーザー入力
         ↓
┌────────┬────────┬────────┬────────┐
│ 1-11   │   r    │   s    │   q    │
│設定変更│ リセット│ 保存   │ 終了   │
└────────┴────────┴────────┴────────┘
    ↓        ↓        ↓        ↓
個別設定    確認     変更有?   変更有?
画面表示  ダイアログ    ↓        ↓
    ↓        ↓      保存実行   警告表示
設定更新   全リセット    ↓        ↓
    ↓        ↓      終了    強制終了?
メイン画面   メイン画面              ↓
復帰      復帰              終了orキャンセル
```

## トラブルシューティング

### よくあるエラーと対処法

**1. 実行ファイルが見つからない**
```bash
# ビルドが必要
go build -o koemoji-go .
```

**2. 設定ファイルが見つからない**
```
Config file not found, using defaults
```
→ 正常動作。初回実行時はデフォルト設定が使用されます。

**3. ディレクトリが存在しない**
```
Failed to create directory ./input: permission denied
```
→ ディレクトリ作成権限を確認するか、パスを変更してください。

**4. Whisperが見つからない**
```
whisper-ctranslate2 not found in any standard location
```
→ `pip install faster-whisper whisper-ctranslate2` で依存関係をインストール

### 設定リセット方法

**1. 対話型リセット（推奨）**
```bash
./koemoji-go --configure
# メニューで 'r' を選択
```

**2. 設定ファイル削除**
```bash
rm config.json
# 次回起動時にデフォルト設定で作成される
```

### デバッグ方法

```bash
# デバッグモードで詳細ログを表示
./koemoji-go --debug

# ログファイルで詳細確認
tail -f koemoji.log
```