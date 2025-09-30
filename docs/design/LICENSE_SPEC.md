# ライセンス認証システム 詳細仕様書

**対象**: Windows版のみ
**方式**: オフライン認証
**運用**: 手動（対人契約）
**更新**: 2025-09-30

---

## 1. システム概要

### 1.1 目的
- 1ライセンス = 1PC の制限実現
- 改ざん防止
- オフライン動作（サーバー不要）

### 1.2 対象環境
- **OS**: Windows 10/11（64bit）のみ
- **配布形態**: スタンドアロン実行ファイル
- **運用方式**: 手動（対人契約）

### 1.3 実装しないもの（YAGNI）
- ❌ macOS対応
- ❌ トライアル期間
- ❌ オンライン認証
- ❌ 自動ライセンス発行
- ❌ サブスクリプション

---

## 2. アーキテクチャ

### 2.1 全体構成

```
┌─────────────────────────────────────────┐
│  KoeMoji-Go.exe (メインアプリ)           │
│  ┌─────────────────────────────────┐    │
│  │ 起動時チェック                   │    │
│  │  ↓                              │    │
│  │ license.ValidateLicense()       │    │
│  │  ├─ GetHardwareID()             │    │
│  │  ├─ ParseLicenseKey()           │    │
│  │  ├─ Ed25519.Verify()            │    │
│  │  └─ ExpiryCheck()               │    │
│  └─────────────────────────────────┘    │
│                                          │
│  成功 → メインアプリ起動                 │
│  失敗 → ライセンス入力ダイアログ         │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│ license-generator.exe (販売者専用)       │
│  ┌─────────────────────────────────┐    │
│  │ CLI引数受け取り                  │    │
│  │  --hwid <UUID>                  │    │
│  │  --days <日数>                   │    │
│  │  --private-key <ファイル>        │    │
│  │  ↓                              │    │
│  │ license.GenerateLicenseKey()    │    │
│  │  ├─ buildMessage()              │    │
│  │  ├─ Ed25519.Sign()              │    │
│  │  └─ encodeLicenseKey()          │    │
│  └─────────────────────────────────┘    │
│                                          │
│  出力 → ライセンスキー文字列              │
└─────────────────────────────────────────┘
```

### 2.2 パッケージ構成

```
KoeMoji-Go/
├── internal/
│   ├── license/              # 新規作成
│   │   ├── license.go        # コア機能
│   │   └── license_windows.go # Windows UUID取得
│   ├── config/
│   │   └── config.go         # 変更：LicenseKey追加
│   ├── gui/
│   │   └── dialogs.go        # 変更：認証ダイアログ追加
│   └── (その他既存パッケージ)
│
├── cmd/
│   ├── koemoji-go/
│   │   └── main.go           # 変更：起動時チェック追加
│   └── license-generator/    # 新規作成
│       └── main.go           # CLI
│
└── docs/
    └── design/
        ├── LICENSE_MVP.md    # MVP計画
        └── LICENSE_SPEC.md   # この設計書
```

---

## 3. データ構造

### 3.1 ライセンスキーの構造

```
文字列形式:
KOEMOJI-XXXXX-XXXXX-XXXXX-XXXXX-XXXXX-XXXXX-XXXXX-XXXXX

↓ Base32デコード

バイナリ構造（105バイト）:
┌──────────┬─────────────┬─────────────┬─────────────┐
│ Type     │ HWID Hash   │ Expiry      │ Signature   │
│ 1 byte   │ 32 bytes    │ 8 bytes     │ 64 bytes    │
└──────────┴─────────────┴─────────────┴─────────────┘
   'P'        SHA-256      Unix時間      Ed25519署名
```

#### 3.1.1 Type（ライセンスタイプ）
```go
const (
    LicenseType = 'P'  // P = Permanent（買い切り）
)

// 将来の拡張（今は実装しない）:
// 'S' = Subscription（サブスク）
// 'T' = Trial（トライアル）
```

#### 3.1.2 HWID Hash
```go
// 元のハードウェアID（UUID）
hwid := "12345678-1234-1234-1234-123456789ABC"

// SHA-256でハッシュ化（32バイト固定長）
hash := sha256.Sum256([]byte(hwid))

理由:
- UUIDの長さが可変でも固定長に
- 元のUUIDの逆算不可（プライバシー保護）
```

#### 3.1.3 Expiry（有効期限）
```go
// Unix timestamp (8 bytes, uint64)
expiry := time.Now().AddDate(0, 0, 36500) // 100年後
expiryUnix := uint64(expiry.Unix())

// Big Endian形式で格納
binary.BigEndian.PutUint64(bytes, expiryUnix)
```

#### 3.1.4 Signature（署名）
```go
// Ed25519署名（64バイト固定）
message := Type + HWID_Hash + Expiry
signature := ed25519.Sign(privateKey, message)
```

### 3.2 設定ファイル（config.json）

```json
{
  "license_key": "KOEMOJI-XXXXX-XXXXX-...",
  "license_expiry": "2125-09-30T23:59:59Z",

  "whisper_model": "large-v3",
  "language": "ja",
  ...
}
```

**変更点**:
```go
type Config struct {
    // 追加
    LicenseKey    string `json:"license_key"`
    LicenseExpiry string `json:"license_expiry"`

    // 既存
    WhisperModel string `json:"whisper_model"`
    Language     string `json:"language"`
    ...
}
```

### 3.3 鍵ペア

#### 秘密鍵（private.key）
```
形式: PEM形式またはバイナリ
サイズ: 64バイト（Ed25519）
保管場所: USBメモリ（オフライン）
用途: ライセンス生成のみ
```

#### 公開鍵（アプリに埋め込み）
```go
// cmd/koemoji-go/main.go
const PublicKeyHex = "a1b2c3d4e5f6789..." // 64バイトをHEX文字列化

func main() {
    publicKey, _ := hex.DecodeString(PublicKeyHex)
    // ...
}
```

---

## 4. 処理フロー

### 4.1 アプリ起動時のフロー

```
┌─────────────────────┐
│ アプリ起動          │
└──────────┬──────────┘
           ↓
┌─────────────────────┐
│ config.json 読み込み │
└──────────┬──────────┘
           ↓
      license_key
      フィールドあり？
           ↓
    ┌─────┴─────┐
    │           │
   YES          NO
    ↓           ↓
┌─────────┐  ┌──────────────┐
│検証実行  │  │ダイアログ表示│
└────┬────┘  │「キー入力」  │
     │       └──────────────┘
     ↓              ↓
  検証成功？     （入力待ち）
     │              ↓
  ┌──┴──┐      入力されたら
  │     │      検証実行へ
 YES   NO           │
  ↓     ↓           │
┌────┐ ┌──────┐    │
│起動│ │エラー│←───┘
└────┘ │表示  │
       └──────┘
```

### 4.2 ライセンス検証の詳細フロー

```go
func ValidateLicense(licenseKey string, publicKey ed25519.PublicKey) error {

    // 1. フォーマット確認
    if !strings.HasPrefix(licenseKey, "KOEMOJI-") {
        return errors.New("invalid format")
    }

    // 2. Base32デコード
    data := decodeBase32(licenseKey)

    // 3. データ解析
    info := ParseLicenseKey(data)
    //   → Type, HWID, Expiry, Signature

    // 4. ハードウェアID取得
    currentHWID := GetHardwareID()
    //   → Windows: wmic csproduct get uuid

    // 5. HWID照合
    currentHash := sha256.Sum256(currentHWID)
    if currentHash != info.HWID {
        return errors.New("hardware mismatch")
    }

    // 6. 署名検証
    message := buildMessage(info.Type, info.HWID, info.Expiry)
    valid := ed25519.Verify(publicKey, message, info.Signature)
    if !valid {
        return errors.New("invalid signature")
    }

    // 7. 有効期限チェック
    if time.Now().After(info.Expiry) {
        return errors.New("expired")
    }

    return nil
}
```

### 4.3 ライセンス生成フロー（販売者側）

```
┌─────────────────────────┐
│ 購入者から HWID 受領    │
│ 例: 12345678-1234-...   │
└───────────┬─────────────┘
            ↓
┌─────────────────────────┐
│ license-generator.exe   │
│ 実行                    │
│                         │
│ --hwid "12345678..."    │
│ --days 36500            │
│ --private-key private.key│
└───────────┬─────────────┘
            ↓
┌─────────────────────────┐
│ 1. HWID をハッシュ化     │
│    SHA-256              │
└───────────┬─────────────┘
            ↓
┌─────────────────────────┐
│ 2. 有効期限計算         │
│    now + 36500日        │
└───────────┬─────────────┘
            ↓
┌─────────────────────────┐
│ 3. メッセージ構築       │
│    Type + Hash + Expiry │
└───────────┬─────────────┘
            ↓
┌─────────────────────────┐
│ 4. Ed25519 署名         │
│    privateKey で署名    │
└───────────┬─────────────┘
            ↓
┌─────────────────────────┐
│ 5. データ連結           │
│    Message + Signature  │
└───────────┬─────────────┘
            ↓
┌─────────────────────────┐
│ 6. Base32 エンコード    │
└───────────┬─────────────┘
            ↓
┌─────────────────────────┐
│ 7. 整形（5文字ごとに-） │
│    KOEMOJI-XXXXX-...    │
└───────────┬─────────────┘
            ↓
┌─────────────────────────┐
│ ✅ ライセンスキー出力   │
│ KOEMOJI-ABCDE-FGHIJ-... │
└─────────────────────────┘
```

---

## 5. API/インターフェース設計

### 5.1 internal/license/license.go

```go
// ValidateLicense はライセンスキーを検証する
//
// 引数:
//   licenseKey: ライセンスキー文字列
//   publicKey:  Ed25519公開鍵（32バイト）
//
// 戻り値:
//   error: 検証エラー。nilなら成功
//
// エラー種類:
//   - "invalid format": フォーマット不正
//   - "hardware mismatch": HWID不一致
//   - "invalid signature": 署名不正
//   - "expired": 期限切れ
func ValidateLicense(licenseKey string, publicKey ed25519.PublicKey) error

// GenerateLicenseKey は新しいライセンスキーを生成する
//
// 引数:
//   hwid:       ハードウェアID（UUID文字列）
//   days:       有効日数
//   privateKey: Ed25519秘密鍵（64バイト）
//
// 戻り値:
//   string: ライセンスキー
//   error:  生成エラー
func GenerateLicenseKey(hwid string, days int, privateKey ed25519.PrivateKey) (string, error)

// ParseLicenseKey はライセンスキーを解析する
func ParseLicenseKey(licenseKey string) (*LicenseInfo, error)

// GetHardwareID は現在のPCのハードウェアIDを取得する（Windows）
func GetHardwareID() (string, error)
```

### 5.2 internal/license/license_windows.go

```go
// GetHardwareID はWindows環境のハードウェアIDを取得
//
// 実装:
//   cmd := exec.Command("wmic", "csproduct", "get", "uuid")
//   output を解析して UUID を抽出
//
// 戻り値:
//   string: UUID（例: "12345678-1234-1234-1234-123456789ABC"）
//   error:  取得失敗時
func GetHardwareID() (string, error)
```

### 5.3 cmd/license-generator/main.go

```go
// CLI引数:
//   --hwid <string>        必須: ハードウェアID
//   --days <int>           必須: 有効日数
//   --private-key <path>   必須: 秘密鍵ファイル
//   --generate-keys        オプション: 鍵ペア生成モード
//
// 使用例:
//   license-generator.exe --hwid "12345..." --days 36500 --private-key private.key
//   license-generator.exe --generate-keys

func main()
```

---

## 6. エラーハンドリング

### 6.1 エラーコード設計

```go
const (
    ErrInvalidFormat    = "LICERR001: ライセンスキーの形式が不正です"
    ErrHardwareMismatch = "LICERR002: このPCでは使用できないライセンスです"
    ErrInvalidSignature = "LICERR003: ライセンスキーが改ざんされています"
    ErrExpired          = "LICERR004: ライセンスの有効期限が切れています"
    ErrHardwareNotFound = "LICERR005: ハードウェアIDの取得に失敗しました"
)
```

### 6.2 ユーザー向けエラーメッセージ

```
┌──────────────────────────────────────┐
│  ❌ ライセンス認証エラー              │
│                                      │
│  このPCでは使用できないライセンスです  │
│  (LICERR002)                         │
│                                      │
│  別のPCで発行されたライセンスキーの   │
│  可能性があります。                  │
│                                      │
│  サポート: your-email@example.com    │
│                                      │
│  [再入力]  [サポートに連絡]          │
└──────────────────────────────────────┘
```

### 6.3 ログ記録

```go
// エラー発生時のログ
logger.LogError("License validation failed: %v (HWID: %s)", err, maskedHWID)

// maskedHWID = "12345678-****-****-****-********ABC"
// （プライバシー保護のため一部マスク）
```

---

## 7. セキュリティ考慮事項

### 7.1 秘密鍵の保護

```
✅ DO:
- USBメモリに保存（暗号化推奨）
- オフラインPCで管理
- バックアップは別のUSBメモリに

❌ DON'T:
- GitHubにコミット
- クラウドストレージ
- メールで送信
- 通常のPCのHDD/SSD
```

### 7.2 公開鍵の埋め込み

```go
// ❌ NG: 外部ファイルから読み込み
publicKey, _ := ioutil.ReadFile("public.key")  // 改ざん可能

// ✅ OK: ソースコードに直接埋め込み
const PublicKeyHex = "a1b2c3d4e5f6..."
publicKey, _ := hex.DecodeString(PublicKeyHex)  // 改ざんには再ビルド必要
```

### 7.3 HWID のマスク表示

```go
// ログやエラーメッセージでは一部をマスク
func maskHWID(hwid string) string {
    if len(hwid) < 12 {
        return "****"
    }
    return hwid[:8] + "-****-****-****-****" + hwid[len(hwid)-3:]
}

// 出力例: "12345678-****-****-****-****ABC"
```

### 7.4 防げないこと（受容するリスク）

```
❌ 完全には防げない:
1. コードの改変（ライセンスチェックをスキップ）
   → 対策: コード難読化（コスト大）

2. ライセンス生成ツールの流出
   → 対策: 厳重管理のみ

3. ハードウェアIDの偽装（高度な技術）
   → 対策: 不可能（コストに見合わない）

✅ 受容:
- 100%の保護は不可能
- 一般ユーザーには十分な保護レベル
- 不正利用の「ハードル」を上げることが目的
```

---

## 8. テスト計画

### 8.1 単体テスト

```go
// internal/license/license_test.go

func TestGenerateAndValidateLicense(t *testing.T)
func TestInvalidFormat(t *testing.T)
func TestHardwareMismatch(t *testing.T)
func TestExpiredLicense(t *testing.T)
func TestInvalidSignature(t *testing.T)
func TestGetHardwareID(t *testing.T)  // Windows環境必須
```

### 8.2 統合テスト

```
1. 鍵ペア生成
   $ license-generator.exe --generate-keys

2. HWID取得
   $ wmic csproduct get uuid
   → 12345678-1234-1234-1234-123456789ABC

3. ライセンス生成
   $ license-generator.exe \
       --hwid "12345678..." \
       --days 365 \
       --private-key private.key
   → KOEMOJI-XXXXX-XXXXX-...

4. メインアプリで検証
   - ライセンスキー入力
   - 認証成功を確認

5. 別のPCで検証（HWID不一致）
   - 同じライセンスキー入力
   - LICERR002 エラーを確認
```

### 8.3 エッジケース

```
1. 有効期限ギリギリ（明日期限切れ）
2. 極端に長いHWID
3. 不正な文字を含むライセンスキー
4. 空文字列のライセンスキー
5. WMICコマンドが失敗する環境
```

---

## 9. 実装順序

### Phase 1: コア機能（1週間）
```
1. internal/license/license.go
   - データ構造定義
   - buildMessage()
   - hashHWID()
   - encodeLicenseKey()
   - ParseLicenseKey()

2. internal/license/license_windows.go
   - GetHardwareID()

3. 単体テスト作成
   - license_test.go
```

### Phase 2: 生成ツール（1週間）
```
1. cmd/license-generator/main.go
   - CLI引数パース
   - 鍵ペア生成機能
   - ライセンス生成機能
   - エラーハンドリング

2. テスト
   - 実際に鍵ペア生成
   - ライセンス生成テスト
```

### Phase 3: メインアプリ統合（1週間）
```
1. internal/config/config.go
   - LicenseKey, LicenseExpiry 追加

2. cmd/koemoji-go/main.go
   - 起動時チェック追加
   - エラーハンドリング

3. internal/gui/dialogs.go
   - ライセンス入力ダイアログ
   - 認証成功/失敗表示
```

### Phase 4: テスト・ドキュメント（1週間）
```
1. 統合テスト実行
2. マニュアル作成
   - ユーザー向け（ライセンス入力方法）
   - 販売者向け（ライセンス生成手順）
3. README更新
```

**合計: 約4週間（172時間）**

---

## 10. 成果物

### 10.1 配布ファイル（ユーザー向け）

```
KoeMoji-Go-v2.0.0-win.zip
├── KoeMoji-Go.exe          # メインアプリ（公開鍵埋め込み済み）
├── *.dll                   # 必要なDLL
├── config.example.json
└── README.txt              # ライセンス入力方法記載
```

### 10.2 販売者用ツール（非公開）

```
license-tools/
├── license-generator.exe   # ライセンス生成ツール
├── private.key             # 秘密鍵（厳重保管）
├── public.key              # 公開鍵（参照用）
└── MANUAL.txt              # 使用方法
```

### 10.3 ドキュメント

```
docs/
├── design/
│   ├── LICENSE_MVP.md      # MVP計画
│   └── LICENSE_SPEC.md     # この詳細設計書
├── user/
│   └── LICENSE_ACTIVATION.md   # ユーザー向けガイド
└── internal/
    └── LICENSE_GENERATION_MANUAL.md  # 販売者向けマニュアル
```

---

**更新**: 2025-09-30