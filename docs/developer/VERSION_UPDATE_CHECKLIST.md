# バージョン更新チェックリスト

KoeMoji-Goのバージョンを更新する際の完全なチェックリストです。

## 📋 バージョン更新の流れ

### 1️⃣ 事前準備

- [ ] 現在のバージョン番号を確認
- [ ] 次のバージョン番号を決定（セマンティックバージョニング）
  - パッチ: `1.6.0` → `1.6.1` (バグ修正のみ)
  - マイナー: `1.6.0` → `1.7.0` (新機能追加、後方互換あり)
  - メジャー: `1.6.0` → `2.0.0` (破壊的変更)
- [ ] 変更内容をドキュメント化

---

### 2️⃣ コードの更新

#### ✅ 必須ファイル

| ファイル | 変更箇所 | 例 |
|---------|---------|-----|
| **version.go** | `const Version` | `const Version = "1.7.0"` |

**重要**: このファイルが**唯一の真実の情報源**です。他のファイルは自動的にこの値を参照します。

#### ⚠️ 確認ファイル

以下のファイルは**自動的に** `version.go` を参照しているため、**手動変更不要**です：

| ファイル | 仕組み | 備考 |
|---------|--------|------|
| `cmd/koemoji-go/main.go` | ビルド時に `-X` フラグで上書き | `var version = "dev"` のまま |
| `build/windows/build.bat` | `findstr` で動的に取得 | 11-14行目 |
| `build/macos/build.sh` | `go run` で動的に取得 | 該当箇所参照 |

---

### 3️⃣ ドキュメントの更新

#### 必須更新

- [ ] **README.md** - バージョン番号、変更履歴
- [ ] **CHANGELOG.md** (存在する場合) - リリースノート追加
- [ ] **CLAUDE.md** - 「最近の重要な変更」セクション

#### 推奨更新

- [ ] **docs/INDEX.md** - バージョン情報
- [ ] **build/common/assets/README_RELEASE.md** - リリースノート

---

### 4️⃣ ビルドとテスト

#### ローカルビルド

```bash
# Windows
cd build/windows
build.bat

# macOS
cd build/macos
./build.sh
```

#### バージョン確認

```bash
# ビルド後の実行ファイルでバージョン確認
./koemoji-go --version

# 期待される出力
KoeMoji-Go v1.7.0
```

#### テスト実行

```bash
# 全テスト実行
go test ./...

# 特定パッケージのテスト
go test ./internal/config
go test ./internal/processor
```

---

### 5️⃣ Git操作

#### コミット

```bash
# 変更を確認
git status
git diff

# ステージング
git add version.go
git add README.md
git add CLAUDE.md
git add docs/INDEX.md

# コミット（絵文字付き）
git commit -m "🚀 release: v1.7.0 [変更内容の概要]

- 主要な変更点1
- 主要な変更点2
- 主要な変更点3

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

#### タグ作成

```bash
# タグ作成
git tag -a v1.7.0 -m "Release v1.7.0: [変更内容の概要]"

# タグ確認
git tag -l

# リモートにプッシュ
git push origin main
git push origin v1.7.0
```

---

### 6️⃣ リリース作成

#### GitHub Releases

1. GitHubリポジトリページに移動
2. **Releases** → **Draft a new release**
3. タグ選択: `v1.7.0`
4. リリースタイトル: `v1.7.0 - [変更内容の概要]`
5. リリースノート記載:

```markdown
## 🎉 v1.7.0 リリース

### ✨ 新機能
- 機能1の説明
- 機能2の説明

### 🔧 改善
- 改善1の説明
- 改善2の説明

### 🐛 バグ修正
- 修正1の説明
- 修正2の説明

### 📦 ダウンロード
- Windows: `KoeMoji-Go-v1.7.0-win.zip`
- macOS: `KoeMoji-Go-v1.7.0-mac.tar.gz`

### 📝 インストール手順
[README.md](https://github.com/hirokitakamura/koemoji-go#installation) を参照してください。
```

6. アセットをアップロード:
   - `build/releases/KoeMoji-Go-v1.7.0-win.zip`
   - `build/releases/KoeMoji-Go-v1.7.0-mac.tar.gz`

7. **Publish release**

---

### 7️⃣ 最終確認

#### ダウンロードテスト

- [ ] Windows版ZIPをダウンロード・解凍・実行
- [ ] macOS版tar.gzをダウンロード・解凍・実行
- [ ] バージョン番号が正しく表示されるか確認

#### ドキュメント確認

- [ ] GitHubリポジトリページのREADMEが最新か
- [ ] Releasesページが正しく表示されているか
- [ ] ダウンロードリンクが正常に動作するか

---

## 🔍 トラブルシューティング

### バージョン番号が古いまま表示される

**原因**: ビルド時に `version.go` が参照されていない

**解決策**:
```bash
# クリーンビルド
cd build/windows
build.bat clean
build.bat

# または
go clean -cache
```

### ビルドスクリプトがバージョンを取得できない

**原因**: `version.go` のフォーマットが変更された

**解決策**: `version.go` を以下の形式に保つ
```go
package main

// Version はアプリケーションのバージョン情報を管理します
// この値は全てのビルドスクリプトとドキュメントで参照されます
const Version = "1.7.0"
```

### Git タグが既に存在する

**原因**: 同じバージョンのタグが既に作成されている

**解決策**:
```bash
# タグを削除して再作成
git tag -d v1.7.0
git push origin :refs/tags/v1.7.0
git tag -a v1.7.0 -m "Release v1.7.0"
git push origin v1.7.0
```

---

## 📊 バージョン管理のベストプラクティス

### セマンティックバージョニング

```
MAJOR.MINOR.PATCH

例: 1.7.0
```

- **MAJOR (1)**: 破壊的変更（後方互換性なし）
- **MINOR (7)**: 新機能追加（後方互換性あり）
- **PATCH (0)**: バグ修正のみ（後方互換性あり）

### プレリリース版

```
1.7.0-beta.1
1.7.0-rc.2
```

- ベータ版: `-beta.N`
- リリース候補: `-rc.N`

---

## ✅ クイックチェックリスト（印刷用）

```
□ version.go を更新
□ README.md を更新
□ CLAUDE.md を更新
□ ローカルビルド成功
□ バージョン表示確認
□ テスト実行成功
□ git commit
□ git tag 作成
□ git push
□ GitHub Release 作成
□ アセットアップロード
□ ダウンロードテスト
```

---

## 🔗 関連ドキュメント

- [開発ガイド](DEVELOPMENT.md)
- [ビルドガイド (Windows)](WINDOWS_BUILD_GUIDE.md)
- [CLAUDE.md](../../CLAUDE.md)
- [セマンティックバージョニング](https://semver.org/lang/ja/)

---

**最終更新**: 2025-01-21
**対象バージョン**: v1.6.0以降
