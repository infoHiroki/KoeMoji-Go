# 🔐 ライセンス認証MVP

**方式**: オフライン認証
**期間**: 5週間
**投資**: ¥845,000

---

## 🎯 認証ロジック

### ライセンスキー認証の仕組み

```
起動時
  ↓
ライセンスキー確認
  ├─ あり → 検証
  │   ├─ ✅ 有効 → 起動
  │   └─ ❌ 無効 → 入力画面
  └─ なし → 入力画面
```

### ライセンスキーの構造

```
KOEMOJI-XXXXX-XXXXX-XXXXX-XXXXX-XXXXX
        ↓     ↓           ↓     ↓
      タイプ  HWID       期限  署名
```

**検証手順**:
1. **フォーマット確認**: `KOEMOJI-`で始まる
2. **ハードウェアID照合**: このPCのUUIDと一致？
3. **署名検証**: Ed25519で改ざんチェック
4. **有効期限**: 今日 < 期限日？

---

## 📂 実装ファイル

### 新規作成

```
internal/license/license.go
  ├─ GetHardwareID() → macOS/Windows UUID取得
  ├─ ValidateLicense() → キー検証
  └─ GenerateLicenseKey() → キー生成（販売者用）

cmd/license-generator/main.go
  └─ CLI: license-generator -hwid XXX -days 365
```

### 変更ファイル

```
config.go: +LicenseKey, +LicenseActivated, +LicenseExpiry
main.go: +起動時チェック
dialogs.go: +認証ダイアログ
```

---

## 🛠️ 技術選択

| 要素 | 選択 | 理由 |
|------|------|------|
| 署名 | Ed25519 | Go標準、高速 |
| HW ID | macOS: IOPlatformUUID<br>Windows: WMIC | OS標準 |
| 保存 | config.json | 既存システム |

---

## 📅 開発スケジュール

| 週 | タスク | 時間 |
|----|--------|------|
| 1-2 | license.go実装 + テスト | 64h |
| 3 | UI統合（GUI/TUI） | 28h |
| 4 | 販売者ツール + テスト | 56h |
| 5 | マニュアル + 販売準備 | 48h |
| 6 | 🚀 販売開始 | - |

**合計**: 220時間（5週間）

---

## 💰 ROI

| 価格 | 損益分岐点 | 回収期間 |
|------|-----------|---------|
| ¥5,980 | 141本 | 6ヶ月 |
| ¥9,800 | 86本 | 4ヶ月 |

**運用費**: ¥1,500/年（ドメインのみ）

---

## 🚫 実装しないもの（YAGNI）

- オンライン認証
- サブスクリプション
- 管理画面
- 統計
- 自動更新

→ 顧客要望があってから追加

---

## 🎬 次のステップ

1. 🔑 Ed25519鍵ペア生成
2. 🌿 `feature/license-authentication`ブランチ作成
3. 📝 `internal/license/license.go`から実装開始

---

**更新**: 2025-09-30