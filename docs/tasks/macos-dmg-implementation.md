# macOS DMG配布 実装タスク

**作成日**: 2025-01-23
**対応バージョン**: v1.7.0
**ステータス**: Not Started

関連ドキュメント: [設計ドキュメント](../design/macos-dmg-distribution.md)

---

## 📋 タスク概要

Phase 1: 基本的な.app/DMG生成（署名なし）

**推定作業時間**: 2-3時間
**実装範囲**: ビルドスクリプトとメタデータのみ修正（コアロジックは変更なし）

---

## ✅ タスク一覧

### 1. 環境セットアップ

#### Task 1.1: fyne-cliのインストール
- **担当**: 開発者
- **優先度**: 🔴 High
- **見積もり**: 5分
- **ステータス**: ⬜ Not Started

**作業内容**:
```bash
go install fyne.io/fyne/v2/cmd/fyne@latest
```

**確認方法**:
```bash
fyne version
# 出力例: Fyne version v2.6.1
```

**依存関係**: なし

---

### 2. メタデータ修正

#### Task 2.1: FyneApp.tomlの修正
- **担当**: 開発者
- **優先度**: 🔴 High
- **見積もり**: 5分
- **ステータス**: ⬜ Not Started

**作業内容**:
`FyneApp.toml`を以下のように修正：

```toml
[Details]
Icon = "Icon.png"
Name = "KoeMoji-Go"
ID = "com.hirokitakamura.koemoji-go"  # 変更: 統一
Category = "Productivity"              # 変更: より適切
Version = "1.7.0"                      # 追加
Build = 1                              # 追加
```

**変更理由**:
- ID統一: `internal/gui/app.go`と一致させる
- Category変更: "Developer Tools"より"Productivity"が適切
- Version/Build追加: リリース管理の明確化

**依存関係**: なし

---

### 3. ビルドスクリプト拡張

#### Task 3.1: build.shに.app生成機能を追加
- **担当**: 開発者
- **優先度**: 🔴 High
- **見積もり**: 30分
- **ステータス**: ⬜ Not Started

**作業内容**:
`build/macos/build.sh`に以下の関数を追加：

```bash
# .app生成関数
build_app() {
    local arch="$1"
    local binary_name="${APP_NAME}-${arch}"

    echo "📱 Building .app bundle for $arch..."

    # バイナリを生成
    build_arch "$arch"

    # fyne packageで.appバンドル作成
    cd ../..
    fyne package -os darwin -icon Icon.png \
        -name KoeMoji-Go \
        -appID com.hirokitakamura.koemoji-go \
        -release

    # .appをdistに移動
    mv KoeMoji-Go.app "build/macos/$DIST_DIR/"
    cd build/macos

    echo "✅ .app bundle created"
}
```

**コマンドライン引数の追加**:
- `build.sh app` - .app生成
- `build.sh dmg` - DMG生成（Task 3.2で実装）
- `build.sh cli` - 従来のtar.gz（現在の処理をリネーム）
- `build.sh all` - 両方生成

**依存関係**: Task 1.1, Task 2.1

---

#### Task 3.2: DMG生成機能の追加
- **担当**: 開発者
- **優先度**: 🔴 High
- **見積もり**: 1時間
- **ステータス**: ⬜ Not Started

**作業内容**:
`build/macos/build.sh`に以下の関数を追加：

```bash
# DMG生成関数
build_dmg() {
    local arch="$1"
    local package_name="koemoji-go-${VERSION}"
    local release_name="koemoji-go-${VERSION}-macos"

    echo "💿 Creating DMG package..."

    # .appバンドルを生成
    build_app "$arch"

    # DMG作成スクリプトを呼び出し
    ./create-dmg.sh "$DIST_DIR/KoeMoji-Go.app" "$release_name"

    echo "✅ DMG created: ../releases/${release_name}.dmg"
}
```

**依存関係**: Task 3.1, Task 4.1

---

#### Task 3.3: CLI版ビルドの分離
- **担当**: 開発者
- **優先度**: 🟡 Medium
- **見積もり**: 15分
- **ステータス**: ⬜ Not Started

**作業内容**:
現在のデフォルト動作を`build_cli()`関数として分離：

```bash
# CLI版生成関数（従来のtar.gz）
build_cli() {
    local arch="$1"

    echo "🖥️  Building CLI version for $arch..."

    # 既存の処理をそのまま移動
    build_arch "$arch"

    # tar.gz作成
    local package_name="koemoji-go-${VERSION}"
    local release_name="koemoji-go-${VERSION}-macos-cli"  # 変更: -cli追加

    # ... 以下既存の処理
}
```

**変更点**:
- ファイル名に`-cli`サフィックス追加
- DMG版との差別化

**依存関係**: なし

---

### 4. DMG作成スクリプト

#### Task 4.1: create-dmg.shの新規作成
- **担当**: 開発者
- **優先度**: 🔴 High
- **見積もり**: 1時間
- **ステータス**: ⬜ Not Started

**作業内容**:
`build/macos/create-dmg.sh`を新規作成：

```bash
#!/bin/bash
set -e

APP_PATH="$1"
OUTPUT_NAME="$2"
DMG_DIR="dmg-temp"
VOLUME_NAME="KoeMoji-Go"

# 一時ディレクトリ作成
rm -rf "$DMG_DIR"
mkdir -p "$DMG_DIR"

# .appをコピー
cp -R "$APP_PATH" "$DMG_DIR/"

# Applicationsへのシンボリックリンク作成
ln -s /Applications "$DMG_DIR/Applications"

# README.txtをコピー
cp "$COMMON_DIR/assets/README_APP.md" "$DMG_DIR/README.txt"

# DMG作成
hdiutil create -volname "$VOLUME_NAME" \
    -srcfolder "$DMG_DIR" \
    -ov -format UDZO \
    "../releases/${OUTPUT_NAME}.dmg"

# クリーンアップ
rm -rf "$DMG_DIR"

echo "✅ DMG created successfully"
```

**実行権限付与**:
```bash
chmod +x build/macos/create-dmg.sh
```

**依存関係**: Task 5.1

---

### 5. ドキュメント作成

#### Task 5.1: README_APP.mdの新規作成
- **担当**: 開発者
- **優先度**: 🔴 High
- **見積もり**: 30分
- **ステータス**: ⬜ Not Started

**作業内容**:
`build/common/assets/README_APP.md`を新規作成

**必須内容**:
1. **インストール手順**
   - DMGマウント
   - Applicationsフォルダへドラッグ&ドロップ

2. **初回起動方法**（セキュリティ警告対策）
   - 右クリック→「開く」の手順
   - スクリーンショット風の説明
   - なぜこの警告が出るのか

3. **基本的な使い方**
   - GUIの簡単な説明
   - 設定ファイルの場所（カレントディレクトリ）
   - input/output/archiveフォルダについて

4. **トラブルシューティング**
   - 起動できない場合
   - Python/FasterWhisperエラー
   - 設定ファイルが見つからない

5. **CLI/TUI利用方法**（上級者向け）
   ```bash
   # ターミナルから起動
   /Applications/KoeMoji-Go.app/Contents/MacOS/koemoji-go --tui

   # エイリアス設定（推奨）
   alias koemoji-go='/Applications/KoeMoji-Go.app/Contents/MacOS/koemoji-go'
   ```

**依存関係**: なし

---

#### Task 5.2: CLAUDE.mdの更新
- **担当**: 開発者
- **優先度**: 🟡 Medium
- **見積もり**: 15分
- **ステータス**: ⬜ Not Started

**作業内容**:
`CLAUDE.md`の「ビルドシステム」セクションを更新

**追加内容**:
1. **配布形式の説明**
   - DMG版とCLI版の違い
   - 対象ユーザー

2. **ビルドコマンドの追加**
   ```bash
   # DMG版ビルド
   cd build/macos && ./build.sh dmg

   # CLI版ビルド
   cd build/macos && ./build.sh cli

   # 両方ビルド
   cd build/macos && ./build.sh all
   ```

3. **配布成果物の更新**
   ```
   releases/
   ├── koemoji-go-1.7.0-macos.dmg
   └── koemoji-go-1.7.0-macos-cli.tar.gz
   ```

**依存関係**: なし

---

### 6. テストと検証

#### Task 6.1: .app動作確認
- **担当**: 開発者
- **優先度**: 🔴 High
- **見積もり**: 30分
- **ステータス**: ⬜ Not Started

**テスト項目**:
- [ ] ダブルクリックで起動
- [ ] GUIが正常に表示される
- [ ] ファイル処理が正常動作
- [ ] 録音機能が動作
- [ ] AI要約機能が動作（APIキー設定時）
- [ ] 設定変更が保存される
- [ ] ログが正常に表示される

**確認コマンド**:
```bash
# .appの構造確認
ls -la build/macos/dist/KoeMoji-Go.app/Contents/
ls -la build/macos/dist/KoeMoji-Go.app/Contents/MacOS/
ls -la build/macos/dist/KoeMoji-Go.app/Contents/Resources/

# Info.plistの確認
plutil -p build/macos/dist/KoeMoji-Go.app/Contents/Info.plist
```

**依存関係**: Task 3.1

---

#### Task 6.2: CLI/TUI互換性確認
- **担当**: 開発者
- **優先度**: 🔴 High
- **見積もり**: 15分
- **ステータス**: ⬜ Not Started

**テスト項目**:
- [ ] ターミナルから.app内バイナリを直接実行
- [ ] `--tui`オプションが動作
- [ ] `--help`, `--version`が正常表示
- [ ] `--configure`が動作

**確認コマンド**:
```bash
# TUIモード起動
/Applications/KoeMoji-Go.app/Contents/MacOS/koemoji-go --tui

# バージョン確認
/Applications/KoeMoji-Go.app/Contents/MacOS/koemoji-go --version

# ヘルプ表示
/Applications/KoeMoji-Go.app/Contents/MacOS/koemoji-go --help
```

**依存関係**: Task 6.1

---

#### Task 6.3: DMG動作確認
- **担当**: 開発者
- **優先度**: 🔴 High
- **見積もり**: 15分
- **ステータス**: ⬜ Not Started

**テスト項目**:
- [ ] DMGが正常にマウントされる
- [ ] .appがコピー可能
- [ ] Applicationsリンクが機能
- [ ] README.txtが読める
- [ ] DMGのアンマウント

**確認コマンド**:
```bash
# DMGをマウント
open build/macos/releases/koemoji-go-1.7.0-macos.dmg

# 内容確認
ls -la /Volumes/KoeMoji-Go/

# アンマウント
hdiutil detach /Volumes/KoeMoji-Go/
```

**依存関係**: Task 4.1

---

#### Task 6.4: CLI版動作確認
- **担当**: 開発者
- **優先度**: 🟡 Medium
- **見積もり**: 10分
- **ステータス**: ⬜ Not Started

**テスト項目**:
- [ ] tar.gzが正常に解凍される
- [ ] バイナリが実行可能
- [ ] 従来通りの動作（後方互換性）

**確認コマンド**:
```bash
cd /tmp
tar -xzf ~/path/to/koemoji-go-1.7.0-macos-cli.tar.gz
cd koemoji-go-1.7.0
./koemoji-go --version
```

**依存関係**: Task 3.3

---

### 7. リリース準備

#### Task 7.1: version.goの更新
- **担当**: 開発者
- **優先度**: 🔴 High
- **見積もり**: 1分
- **ステータス**: ⬜ Not Started

**作業内容**:
```go
const Version = "1.7.0"
```

**依存関係**: 全タスク完了後

---

#### Task 7.2: 最終ビルドと成果物確認
- **担当**: 開発者
- **優先度**: 🔴 High
- **見積もり**: 10分
- **ステータス**: ⬜ Not Started

**作業内容**:
```bash
cd build/macos
./build.sh clean
./build.sh all
```

**確認項目**:
- [ ] `releases/koemoji-go-1.7.0-macos.dmg`が存在
- [ ] `releases/koemoji-go-1.7.0-macos-cli.tar.gz`が存在
- [ ] ファイルサイズが妥当（DMG: 15-20MB、tar.gz: 12-15MB）

**依存関係**: Task 7.1, 全テスト完了

---

## 📂 変更ファイル一覧

### 修正するファイル

| ファイルパス | 変更内容 | 優先度 |
|-------------|---------|--------|
| `FyneApp.toml` | ID統一、メタデータ追加 | 🔴 High |
| `build/macos/build.sh` | app/dmg/cli生成機能追加 | 🔴 High |
| `CLAUDE.md` | ビルド手順更新 | 🟡 Medium |
| `version.go` | バージョン1.7.0に更新 | 🔴 High |

### 新規作成するファイル

| ファイルパス | 内容 | 優先度 |
|-------------|-----|--------|
| `build/macos/create-dmg.sh` | DMG生成スクリプト | 🔴 High |
| `build/common/assets/README_APP.md` | .app版ユーザー向けドキュメント | 🔴 High |
| `docs/design/macos-dmg-distribution.md` | 設計ドキュメント | 🟡 Medium |
| `docs/tasks/macos-dmg-implementation.md` | このファイル | 🟡 Medium |

---

## 🎯 完了条件（Definition of Done）

### 機能要件
- [x] fyne-cliがインストールされている
- [ ] .appがダブルクリックで起動する
- [ ] DMGが正常に作成される
- [ ] CLI版（tar.gz）も正常にビルドされる
- [ ] CLI/TUIモードが.appからも利用可能

### 品質要件
- [ ] 全テスト項目がパス
- [ ] 既存機能（文字起こし、録音、AI要約）が正常動作
- [ ] 設定ファイルの読み込み・保存が正常動作
- [ ] Python（FasterWhisper）呼び出しが成功

### ドキュメント要件
- [ ] README_APP.mdが分かりやすい
- [ ] セキュリティ警告の回避方法が明記されている
- [ ] CLAUDE.mdにビルド手順が追加されている

### リリース要件
- [ ] version.goが1.7.0に更新されている
- [ ] 2つの配布ファイルが生成されている
- [ ] GitHub Releaseの準備が整っている

---

## 📊 進捗管理

### 全体進捗
- **未着手**: 14タスク
- **進行中**: 0タスク
- **完了**: 0タスク
- **全体進捗**: 0%

### フェーズ別進捗
| フェーズ | タスク数 | 完了 | 進捗率 |
|---------|---------|------|--------|
| 1. 環境セットアップ | 1 | 0 | 0% |
| 2. メタデータ修正 | 1 | 0 | 0% |
| 3. ビルドスクリプト | 3 | 0 | 0% |
| 4. DMG作成 | 1 | 0 | 0% |
| 5. ドキュメント | 2 | 0 | 0% |
| 6. テスト | 4 | 0 | 0% |
| 7. リリース準備 | 2 | 0 | 0% |

---

## 🚨 既知の課題・リスク

### Issue #1: 設定ファイルの場所
- **優先度**: 🟡 Medium
- **内容**: .app起動時、カレントディレクトリが予測不能
- **対策**: Phase 2で複数パスのフォールバック実装
- **回避策**: README_APP.mdで明確に案内

### Issue #2: Python依存
- **優先度**: 🟡 Medium
- **内容**: FasterWhisper呼び出しが.appから失敗する可能性
- **対策**: Task 6.1で重点的にテスト
- **回避策**: エラーメッセージを分かりやすく

### Issue #3: セキュリティ警告
- **優先度**: 🟢 Low（仕様）
- **内容**: 署名なしの.appは初回起動時に警告
- **対策**: README_APP.mdで詳細に案内
- **将来対応**: Phase 3でコード署名

---

## 📝 メモ・備考

### 開発環境
- macOS Sonoma以降
- Go 1.21+
- Fyne v2.6.1

### 参考コマンド
```bash
# .app内部の確認
open -a TextEdit build/macos/dist/KoeMoji-Go.app/Contents/Info.plist

# DMGのサイズ確認
du -sh build/macos/releases/*.dmg

# バイナリのサイズ確認
ls -lh build/macos/dist/koemoji-go-*
```

---

## 変更履歴

| 日付 | 変更内容 | 担当者 |
|------|---------|--------|
| 2025-01-23 | 初版作成 | Claude Code |
