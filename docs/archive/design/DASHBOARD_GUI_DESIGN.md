# KoeMoji-Go GUI改善設計書

## 現在のGUI構造

```
┌─────────────────────────────────────┐
│ 🟢 稼働中 | 待機: 0 | 処理中: なし      │
│ 📁 入力: 2 → 出力: 5 → アーカイブ: 10   │
│ ⏰ 最終: 15:04:05 | 次回: 15:05:05     │
├─────────────────────────────────────┤
│ [INFO] 15:04:05 メッセージ...          │
│ [PROC] 15:04:10 処理中...             │
├─────────────────────────────────────┤
│ [設定] [ログ] [スキャン] [録音] [終了]   │
└─────────────────────────────────────┘
```

**構造の特徴:**
- 3行ステータス：システム状態、ファイル数、タイミング表示
- ログ表示：時系列の処理状況
- ボタンパネル：主要操作機能

## 改善計画

### 1. 設定画面の日本語化

**項目名の変更:**
- "Whisper Model" → "Whisperモデル"
- "Language" → "音声認識言語"
- "UI Language" → "表示言語"
- "Scan Interval (min)" → "スキャン間隔（分）"
- "Use Colors" → "色を使用"
- "Input Directory" → "入力フォルダ"
- "Output Directory" → "出力フォルダ"
- "Archive Directory" → "アーカイブフォルダ"
- "Enable LLM Summary" → "AI要約を有効化"
- "API Key" → "APIキー"
- "Model" → "モデル"
- "Recording Device" → "録音デバイス"

**タブ名の変更:**
- "Basic" → "基本設定"
- "Directories" → "フォルダ設定"
- "LLM" → "AI要約"
- "Recording" → "録音設定"

**ダイアログの変更:**
- "Settings" → "設定"
- "Save" → "保存"
- "Cancel" → "キャンセル"
- "KoeMoji-Go Configuration" → "KoeMoji-Go 設定"

### 2. プルダウン化改善

**Whisperモデル選択:**
```go
whisperModels := []string{
    "tiny", "tiny.en", "base", "base.en", 
    "small", "small.en", "medium", "medium.en",
    "large", "large-v1", "large-v2", "large-v3",
}
```

**音声認識言語選択:**
```go
languages := []string{"ja", "en", "zh", "ko", "es", "fr", "de"}
```

### 3. ボタンレイアウト改善

**中央揃え配置:**
- 左右にスペーサーを追加
- セパレーターで機能グループを分離
- 統一されたボタン間隔

### 4. ログ表示名変更

- "Recent Logs" → "ログ"

## 実装計画

### 変更対象ファイル
- `internal/gui/dialogs.go` - 設定画面の日本語化とプルダウン化
- `internal/gui/components.go` - ボタンレイアウト、ログ表示名変更

### 変更しないファイル
- `internal/gui/app.go` - 更新ロジック維持
- `internal/config/config.go` - 設定項目構造維持
- 全ての業務ロジックパッケージ

## 品質保証

### 安全性
- UI表示のみの変更で業務ロジックに影響なし
- 既存設定ファイルとの完全互換性
- 5秒間隔更新システム維持

### テスト項目
- 全機能の動作確認
- 設定保存・読み込み確認
- 日本語・英語UI環境テスト
- 全ボタン・フォーム操作確認

