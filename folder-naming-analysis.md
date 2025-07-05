# 解凍後フォルダ名変更の検討記録

## 現状分析

### 現在のフォルダ名構成
- **Windows**: `koemoji-go-windows-1.5.4`
- **macOS**: `koemoji-go-macos-arm64-1.5.4`

### アーカイブファイル名（既に変更済み）
- **Windows**: `KoeMoji-Go-v1.5.4-win.zip`
- **macOS**: `KoeMoji-Go-v1.5.4-mac.tar.gz`

## 検討した解決策

### Option A: 固定フォルダ名 `KoeMoji-Go`
**メリット:**
- アップデート時に同じ場所に上書きインストール可能
- シンプルなフォルダ名

**デメリット:**
- **データ消失リスク**: 上書きインストール時に以下が削除される
  - `input/` フォルダ（処理待ちファイル）
  - `output/` フォルダ（文字起こし結果）
  - `archive/` フォルダ（処理済みファイル）
  - `config.json`（ユーザー設定）

### Option B: バージョン付きフォルダ名短縮
**提案:**
- Windows: `koemoji-go-windows-1.5.4` → `KoeMoji-Go-v1.5.4`
- macOS: `koemoji-go-macos-arm64-1.5.4` → `KoeMoji-Go-v1.5.4`

**メリット:**
- データ消失リスク回避
- 複数バージョン並行インストール可能
- 統一された命名規則
- 既存のポータブル性維持

**デメリット:**
- 手動でのデータ移行が必要（アップデート時）

### Option C: データフォルダ分離
**提案:**
```
KoeMoji-Go/          ← 実行ファイルのみ
KoeMoji-Go-Data/     ← データフォルダ
```

**デメリット:**
- ポータブル性の喪失
- 設定ファイルの大幅変更が必要

## データ消失リスク分析

### 上書きインストール時の挙動

#### Windows での上書き
```
新しいKoeMoji-Go.zipを解凍 → 「フォルダを置き換えますか？」

選択肢：
- 「はい」→ フォルダ全体削除 → データ消失！
- 「いいえ」→ 上書きされない
```

#### macOS での上書き
```
新しいKoeMoji-Go.tar.gzを解凍 → 既存フォルダに上書き

結果：
- tar.gzの内容のみ残る
- 既存のinput/, output/, archive/ が削除される
```

### 現在のデータフォルダ構成
```
KoeMoji-Go/
├── koemoji-go.exe
├── config.json
├── README.md
├── input/     ← アプリが実行時に作成
├── output/    ← アプリが実行時に作成
└── archive/   ← アプリが実行時に作成
```

## 技術的実装詳細

### Option B採用時の修正箇所

#### Windows build.bat 修正
- **Line 154**: `mkdir %APP_NAME%-windows-%VERSION%` → `mkdir KoeMoji-Go-v%VERSION%`
- **Line 155-160**: コピー先パスを `KoeMoji-Go-v%VERSION%` に変更
- **Line 165**: 圧縮元パスを `KoeMoji-Go-v%VERSION%` に変更
- **Line 172**: 削除パスを `KoeMoji-Go-v%VERSION%` に変更

#### macOS build.sh 修正
- **Line 47**: `package_name="${APP_NAME}-macos-${arch}-${VERSION}"` → `package_name="KoeMoji-Go-v${VERSION}"`
- 対応するコピー・圧縮・削除パスも修正

## 次回セッションでの作業指示

### 1. 新ブランチ作成（完了）
```bash
git checkout -b feature/folder-naming-convention
```

### 2. 修正すべきファイル
- `/build/windows/build.bat` - Line 154, 155-160, 165, 172
- `/build/macos/build.sh` - Line 47 および関連パス

### 3. テスト手順
- macOS ビルドテスト実行
- Windows ビルドテスト実行（可能であれば）
- 解凍後フォルダ名の確認

### 4. 期待される成果物
- `KoeMoji-Go-v1.5.4-mac.tar.gz`
- `KoeMoji-Go-v1.5.4-win.zip`
- 解凍後フォルダ名: `KoeMoji-Go-v1.5.4/`

## 暫定結論

**推奨案: Option B（バージョン付きフォルダ名短縮）**
- 安全性を優先
- データ消失リスク回避
- 命名規則の統一

**最終決定**: 保留中（計画段階）

## 前回の作業完了状況

### アーカイブファイル名変更（完了済み）
- Windows: `KoeMoji-Go_Windows版.zip` → `KoeMoji-Go-v1.5.4-win.zip`
- macOS: `KoeMoji-Go_Mac_M1M2版.tar.gz` → `KoeMoji-Go-v1.5.4-mac.tar.gz`

### ビルドスクリプト修正（完了済み）
- `/build/windows/build.bat` - Line 164, 169, 182でRELEASE_NAME変数使用
- `/build/macos/build.sh` - Line 66, 67, 72, 84, 132でrelease_name変数使用

### 残り作業：解凍後フォルダ名の変更
**現在の状況：** アーカイブファイル名は新形式だが、解凍後フォルダ名は旧形式のまま

## 現在のブランチ状況
- **元ブランチ**: `fix/whisper-auto-installation-flow`
- **作業ブランチ**: `feature/folder-naming-convention` (作成済み)
- **Git状況**: 
  - `folder-naming-analysis.md` 作成済み（未コミット）
  - `WINDOWS_WHISPER_ISSUES.md` 削除済み（未コミット）

## TodoList状況
- ✅ **完了**: macOS build script with new naming convention
- ✅ **完了**: Windows build script with new naming convention  
- ⏳ **保留**: Rename existing release files to new format
- 🆕 **新規**: 解凍後フォルダ名の変更（本検討の対象）

## 作成日時
2025-07-05

## 作成者
Claude Code による検討記録