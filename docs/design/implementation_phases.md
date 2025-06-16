# 対話型設定機能 - 実装フェーズ

## Phase 1: フォルダパス設定の基盤実装
**目的**: JSONでフォルダパスを設定可能にする

### 実装内容
1. Config構造体に追加
   ```go
   InputDir    string `json:"input_dir"`
   OutputDir   string `json:"output_dir"`
   ArchiveDir  string `json:"archive_dir"`
   ```

2. デフォルト値設定
   ```go
   InputDir:    "./input",
   OutputDir:   "./output", 
   ArchiveDir:  "./archive",
   ```

3. ハードコードされたパスを置換
   - `"input"` → `app.config.InputDir`
   - `"output"` → `app.config.OutputDir`
   - `"archive"` → `app.config.ArchiveDir`

### 動作確認
- 既存の動作に影響なし
- config.jsonでパス変更可能

---

## Phase 2: 対話型設定UI
**目的**: `--config`フラグで設定変更UIを提供

### 実装内容
1. フラグ追加
   ```go
   configMode := flag.Bool("config", false, "Enter configuration mode")
   ```

2. 対話型UIの基本実装
   - メニュー表示
   - 各設定項目の変更機能
   - 設定保存機能

3. 手動パス入力対応
   - フォルダパスの手動入力
   - 存在確認とバリデーション

### 動作確認
- `./koemoji-go --config`で設定モード起動
- 全設定項目の変更が可能
- config.jsonへの保存確認

---

## Phase 3: フォルダ選択ダイアログ
**目的**: GUIでフォルダ選択を可能に

### 実装内容
1. OS別実装
   ```go
   // macOS: osascript
   // Windows: PowerShell
   ```

2. 対話型UIとの統合
   - 手動入力 or ダイアログ選択
   - エラー時のフォールバック

### 動作確認
- macOS/Windowsでダイアログ表示
- 選択したパスの反映

---

## Phase 4: 仕上げ
**目的**: 完成度向上

### 実装内容
1. エラーハンドリング強化
2. ヘルプメッセージ更新
3. README更新

### 動作確認
- 各OS環境でのテスト
- エッジケースの確認

---

## 各フェーズの所要時間目安
- Phase 1: 30分（基盤実装）
- Phase 2: 1-2時間（対話型UI）
- Phase 3: 1時間（ダイアログ実装）
- Phase 4: 30分（仕上げ）

## リスクと対策
- **Phase 1**: 既存機能への影響 → 十分なテスト
- **Phase 2**: UI設計の複雑化 → シンプルに保つ
- **Phase 3**: OS依存の問題 → フォールバック実装