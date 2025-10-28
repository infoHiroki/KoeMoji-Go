#!/bin/bash
# リリース自動化スクリプト
# Usage: ./scripts/release.sh

set -e

# GitHub CLIをPATHに追加
export PATH="/c/Program Files/GitHub CLI:$PATH"

# バージョン取得
VERSION=$(grep 'const Version' version.go | cut -d'"' -f2)
echo "📦 Building KoeMoji-Go v${VERSION}..."

# 1. Windowsビルド
echo ""
echo "🪟 Building Windows version..."
cd build/windows
./build.bat
cd ../..

# 2. リリースディレクトリの確認
RELEASE_DIR="build/releases"
if [ ! -d "$RELEASE_DIR" ]; then
    mkdir -p "$RELEASE_DIR"
fi

# 3. ビルド成果物の確認
WINDOWS_ZIP="build/releases/koemoji-go-${VERSION}.zip"

if [ ! -f "$WINDOWS_ZIP" ]; then
    echo "❌ Error: Windows build not found at $WINDOWS_ZIP"
    exit 1
fi

echo ""
echo "✅ Build artifacts ready:"
echo "  - $WINDOWS_ZIP"

# 4. Gitタグ作成
echo ""
echo "🏷️  Creating Git tag v${VERSION}..."
git tag -a "v${VERSION}" -m "v${VERSION}" 2>/dev/null || echo "Tag v${VERSION} already exists"
git push origin "v${VERSION}" 2>/dev/null || echo "Tag already pushed"

# 5. GitHub Release作成
echo ""
echo "🚀 Creating GitHub Release..."

# リリースノート生成
RELEASE_NOTES="## KoeMoji-Go v${VERSION}

### 主な機能
- デュアル録音機能（システム音声+マイク同時録音）※Windows版
- 音量設定5段階相対スケール（-2～+2）
- 音量自動調整（内部処理化、常時有効）
- プラットフォーム別デフォルト設定

### ダウンロード
- **Windows版**: koemoji-go-${VERSION}.zip

### システム要件
- Windows 10/11（64bit）
- Python 3.8以上（FasterWhisper用）

---

**Full Changelog**: https://github.com/infoHiroki/KoeMoji-Go/compare/v1.6.1...v${VERSION}"

# GitHub Releaseを作成してファイルをアップロード
gh release create "v${VERSION}" \
    --title "v${VERSION}" \
    --notes "$RELEASE_NOTES" \
    "$WINDOWS_ZIP"

echo ""
echo "✅ Release v${VERSION} created successfully!"
echo "🌐 View at: https://github.com/infoHiroki/KoeMoji-Go/releases/tag/v${VERSION}"
