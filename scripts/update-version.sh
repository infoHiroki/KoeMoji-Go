#!/bin/bash

# KoeMoji-Go バージョン一括更新スクリプト
# 使用法: ./scripts/update-version.sh [新しいバージョン]
# 例: ./scripts/update-version.sh 1.6.0

set -e

# バージョン引数チェック
if [ $# -ne 1 ]; then
    echo "使用法: $0 [新しいバージョン]"
    echo "例: $0 1.6.0"
    exit 1
fi

NEW_VERSION="$1"
CURRENT_VERSION=$(grep -o 'const Version = "[^"]*"' version.go | cut -d'"' -f2)

echo "🔄 バージョン更新: $CURRENT_VERSION -> $NEW_VERSION"

# セマンティックバージョニング形式チェック
if ! [[ "$NEW_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "❌ エラー: バージョン形式が正しくありません (例: 1.2.3)"
    exit 1
fi

# 現在のバージョンをバックアップ
echo "📋 現在のバージョン: $CURRENT_VERSION"

# 1. version.go を更新
echo "📝 version.go を更新中..."
sed -i.bak "s/const Version = \"$CURRENT_VERSION\"/const Version = \"$NEW_VERSION\"/" version.go
rm version.go.bak

# 2. macOS ビルドスクリプトを更新
echo "📝 macOS ビルドスクリプトを更新中..."
sed -i.bak "s/VERSION=\"$CURRENT_VERSION\"/VERSION=\"$NEW_VERSION\"/" build/macos/build.sh
rm build/macos/build.sh.bak

# 3. Windows ビルドスクリプトを更新
echo "📝 Windows ビルドスクリプトを更新中..."
sed -i.bak "s/set VERSION=$CURRENT_VERSION/set VERSION=$NEW_VERSION/" build/windows/build.bat
rm build/windows/build.bat.bak

# 4. Windows versioninfo.json を更新
echo "📝 Windows versioninfo.json を更新中..."
VERSIONINFO_FILE="build/common/templates/windows/versioninfo.json"

# メジャー、マイナー、パッチバージョンを分割
IFS='.' read -r MAJOR MINOR PATCH <<< "$NEW_VERSION"

# JSON内のバージョン情報を更新
sed -i.bak "s/\"Major\": [0-9]*/\"Major\": $MAJOR/g" "$VERSIONINFO_FILE"
sed -i.bak "s/\"Minor\": [0-9]*/\"Minor\": $MINOR/g" "$VERSIONINFO_FILE"
sed -i.bak "s/\"Patch\": [0-9]*/\"Patch\": $PATCH/g" "$VERSIONINFO_FILE"
sed -i.bak "s/\"ProductVersion\": \"[^\"]*\"/\"ProductVersion\": \"$NEW_VERSION\"/" "$VERSIONINFO_FILE"
sed -i.bak "s/\"FileVersion\": \"[^\"]*\"/\"FileVersion\": \"$NEW_VERSION.0\"/" "$VERSIONINFO_FILE"
rm "$VERSIONINFO_FILE.bak"

# 5. README.md を更新（該当箇所があれば）
echo "📝 README.md を更新中..."
if grep -q "Version.*$CURRENT_VERSION" README.md 2>/dev/null; then
    sed -i.bak "s/Version.*$CURRENT_VERSION/Version $NEW_VERSION/" README.md
    rm README.md.bak
fi

# 6. その他のドキュメントを更新
echo "📝 その他のドキュメントを更新中..."
for file in docs/developer/WINDOWS_BUILD_GUIDE.md build/common/assets/README_RELEASE.md; do
    if [ -f "$file" ] && grep -q "$CURRENT_VERSION" "$file"; then
        sed -i.bak "s/$CURRENT_VERSION/$NEW_VERSION/g" "$file"
        rm "$file.bak"
    fi
done

# 7. 更新結果を確認
echo ""
echo "✅ バージョン更新完了!"
echo ""
echo "📋 更新されたファイル:"
echo "  - version.go"
echo "  - build/macos/build.sh"
echo "  - build/windows/build.bat"
echo "  - build/common/templates/windows/versioninfo.json"
echo "  - README.md (該当箇所があれば)"
echo "  - その他のドキュメント"
echo ""
echo "🔍 更新内容の確認:"
echo "  git diff"
echo ""
echo "📝 次のステップ:"
echo "  1. 変更内容を確認: git diff"
echo "  2. テストビルド実行"
echo "  3. コミット: git add . && git commit -m \"chore: bump version to $NEW_VERSION\""
echo "  4. タグ作成: git tag v$NEW_VERSION"
echo ""