# GitHub CLI ガイド

## 概要

GitHub CLI (`gh`) は、コマンドラインからGitHubの操作を自動化するツールです。
ブラウザでの手動作業を、コマンド一発で実行できます。

## 基本的な仕組み

### 従来の手動方法 vs GitHub CLI

| 作業 | 手動（ブラウザ） | GitHub CLI |
|------|------------------|------------|
| リリース作成 | 5-10分 | 3秒 |
| ファイルアップロード | 1つずつドラッグ&ドロップ | 複数ファイル一括 |
| リリースノート記入 | フォームに手入力 | テキストファイルから自動設定 |

### 内部動作

1. **GitHub API呼び出し**: `gh`コマンドがGitHub REST APIを使用
2. **認証**: GitHubトークンで自動認証
3. **操作実行**: ブラウザで行う操作をAPI経由で自動実行

## インストール・設定

### インストール
```bash
# macOS
brew install gh

# Windows
winget install --id GitHub.cli

# 他の方法
# https://cli.github.com/
```

### 初期設定
```bash
# GitHubにログイン
gh auth login

# 設定確認
gh auth status
```

## リリース作成

### 基本的な使い方

```bash
# 1. タグ作成・プッシュ
git tag v1.0.0
git push origin v1.0.0

# 2. リリース作成（ファイルなし）
gh release create v1.0.0 --title "リリースタイトル" --notes "リリース内容"

# 3. リリース作成（ファイル付き）
gh release create v1.0.0 \
  --title "リリースタイトル" \
  --notes "リリース内容" \
  file1.zip file2.tar.gz
```

### 実際の例（KoeMoji-Go）

```bash
gh release create v1.1.1 \
  --title "KoeMoji-Go v1.1.1 - Windows環境問題修正版" \
  --notes "$(cat <<'EOF'
## Windows環境問題修正版

### 修正内容
- Windows標準cmdでの色表示問題を修正
- ログ表示コマンド（l）をnotepadに変更
- EXEファイルのアイコン表示問題を修正
EOF
)" \
  build/dist/koemoji-go-1.1.0.zip \
  build/dist/koemoji-go-macos-arm64-1.1.0.tar.gz
```

### リリースノートのオプション

```bash
# ファイルから読み込み
gh release create v1.0.0 --notes-file RELEASE_NOTES.md

# 自動生成（前回リリースからの変更）
gh release create v1.0.0 --generate-notes

# プレリリースとして作成
gh release create v1.0.0-beta --prerelease

# ドラフトとして作成
gh release create v1.0.0 --draft
```

## その他の便利なコマンド

### リリース関連

```bash
# リリース一覧表示
gh release list

# 特定のリリース詳細表示
gh release view v1.0.0

# リリース編集
gh release edit v1.0.0 --notes "更新されたリリースノート"

# リリース削除
gh release delete v1.0.0

# アセット追加
gh release upload v1.0.0 new-file.zip
```

### リポジトリ操作

```bash
# リポジトリ作成
gh repo create my-project --public

# リポジトリクローン
gh repo clone owner/repo

# フォーク
gh repo fork owner/repo

# リポジトリ表示（ブラウザで開く）
gh repo view --web
```

### Issue・Pull Request

```bash
# Issue作成
gh issue create --title "バグ報告" --body "詳細説明"

# Issue一覧
gh issue list

# Pull Request作成
gh pr create --title "機能追加" --body "詳細説明"

# Pull Request一覧
gh pr list
```

## 自動化への活用

### GitHub Actionsとの連携

```yaml
# .github/workflows/release.yml
name: Release
on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Build
        run: ./build.sh
        
      - name: Create Release
        run: |
          gh release create ${{ github.ref_name }} \
            --title "Release ${{ github.ref_name }}" \
            --generate-notes \
            dist/*.zip dist/*.tar.gz
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### スクリプト化例

```bash
#!/bin/bash
# release.sh

VERSION=$1
if [ -z "$VERSION" ]; then
  echo "Usage: ./release.sh v1.0.0"
  exit 1
fi

# ビルド
./build.sh

# タグ作成
git tag $VERSION
git push origin $VERSION

# リリース作成
gh release create $VERSION \
  --title "Release $VERSION" \
  --generate-notes \
  dist/*.zip dist/*.tar.gz

echo "✅ Release $VERSION created successfully!"
```

## トラブルシューティング

### よくある問題

**認証エラー**
```bash
# 再ログイン
gh auth logout
gh auth login
```

**ファイルが見つからない**
```bash
# ファイルパスを絶対パスで指定
gh release create v1.0.0 /full/path/to/file.zip
```

**リリースが作成されない**
```bash
# タグが存在するか確認
git tag -l

# リモートにプッシュ済みか確認
git ls-remote --tags origin
```

## まとめ

GitHub CLIを使うことで：

- ✅ **効率化**: 手動作業が数秒で完了
- ✅ **自動化**: スクリプトやCI/CDに組み込み可能
- ✅ **一貫性**: 人的ミスを防止
- ✅ **再現性**: 同じコマンドで同じ結果

特にリリース作業では、複数ファイルのアップロードや詳細なリリースノートの設定が、コマンド一発で完了するため、開発効率が大幅に向上します。