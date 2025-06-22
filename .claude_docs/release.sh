#!/bin/bash
# /release カスタムコマンドスクリプト
# KoeMoji-Go の手動ビルドとリリース作成を自動化

set -e

echo "🚀 /release コマンドを実行します..."

# 1. 現在の状態を確認
echo "📋 現在のGit状態を確認..."
git status --short

# 未コミットの変更がある場合は警告
if [[ -n $(git status --porcelain) ]]; then
    echo "⚠️  警告: 未コミットの変更があります"
    read -p "続行しますか？ (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ リリースを中止しました"
        exit 1
    fi
fi

# 2. バージョン情報を取得
CURRENT_VERSION=$(grep 'VERSION=' build/build.sh | cut -d'"' -f2)
echo "📌 現在のバージョン: $CURRENT_VERSION"
read -p "新しいバージョンを入力してください (現在: $CURRENT_VERSION): " NEW_VERSION

if [[ -z "$NEW_VERSION" ]]; then
    NEW_VERSION=$CURRENT_VERSION
fi

# 3. バージョンを更新
if [[ "$NEW_VERSION" != "$CURRENT_VERSION" ]]; then
    echo "📝 バージョンを $NEW_VERSION に更新..."
    sed -i '' "s/VERSION=\"$CURRENT_VERSION\"/VERSION=\"$NEW_VERSION\"/" build/build.sh
    
    # versioninfo.jsonも更新
    sed -i '' "s/\"ProductVersion\": \"$CURRENT_VERSION\"/\"ProductVersion\": \"$NEW_VERSION\"/" build/versioninfo.json
    sed -i '' "s/\"FileVersion\": \"$CURRENT_VERSION\"/\"FileVersion\": \"$NEW_VERSION\"/" build/versioninfo.json
    
    # 変更をコミット
    git add build/build.sh build/versioninfo.json
    git commit -m "chore: bump version to $NEW_VERSION"
fi

# 4. ビルドを実行
echo "🔨 ビルドを開始..."
cd build
./build.sh
cd ..

# 5. ビルド成果物を確認
echo "📦 ビルド成果物:"
ls -la build/dist/*.{zip,tar.gz} 2>/dev/null || echo "エラー: ビルド成果物が見つかりません"

# 6. リリースを作成するか確認
read -p "GitHubリリースを作成しますか？ (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    # タグを作成
    TAG="v$NEW_VERSION"
    echo "🏷️  タグ $TAG を作成..."
    git tag -a "$TAG" -m "Release $TAG"
    git push origin "$TAG"
    
    # リリースノートを入力
    echo "📝 リリースノートを入力してください (Ctrl+Dで終了):"
    RELEASE_NOTES=$(cat)
    
    # GitHubリリースを作成
    echo "🚀 GitHubリリースを作成..."
    gh release create "$TAG" \
        build/dist/koemoji-go-windows-$NEW_VERSION.zip \
        build/dist/koemoji-go-macos-intel-$NEW_VERSION.tar.gz \
        build/dist/koemoji-go-macos-arm64-$NEW_VERSION.tar.gz \
        --title "KoeMoji-Go $TAG" \
        --notes "$RELEASE_NOTES"
    
    echo "✅ リリース $TAG が作成されました！"
    echo "🔗 https://github.com/infoHiroki/KoeMoji-Go/releases/tag/$TAG"
else
    echo "ℹ️  リリース作成をスキップしました"
    echo "📁 ビルド成果物は build/dist/ にあります"
fi

echo "🎉 /release コマンドが完了しました！"