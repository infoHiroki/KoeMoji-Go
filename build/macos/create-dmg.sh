#!/bin/bash
set -e

APP_PATH="$1"
OUTPUT_NAME="$2"
DMG_DIR="dmg-temp"
VOLUME_NAME="KoeMoji-Go"
COMMON_DIR="../common"

if [ -z "$APP_PATH" ] || [ -z "$OUTPUT_NAME" ]; then
    echo "Usage: $0 <app_path> <output_name>"
    echo "Example: $0 dist/KoeMoji-Go.app koemoji-go-1.7.0-macos"
    exit 1
fi

if [ ! -d "$APP_PATH" ]; then
    echo "❌ Error: App not found at $APP_PATH"
    exit 1
fi

echo "📦 Creating DMG package: $OUTPUT_NAME.dmg"
echo ""

# Clean up any existing temp directory
rm -rf "$DMG_DIR"
mkdir -p "$DMG_DIR"

# Copy .app to temp directory
echo "Copying .app bundle..."
cp -R "$APP_PATH" "$DMG_DIR/"

# Create symlink to Applications folder
echo "Creating Applications symlink..."
ln -s /Applications "$DMG_DIR/Applications"

# Copy README for .app users
echo "Adding README..."
if [ -f "$COMMON_DIR/assets/README.txt" ]; then
    cp "$COMMON_DIR/assets/README.txt" "$DMG_DIR/README.txt"
else
    # If README.txt doesn't exist yet, create a minimal one
    cat > "$DMG_DIR/README.txt" << 'EOF'
# KoeMoji-Go

音声・動画ファイル自動文字起こしツール

## インストール方法

1. KoeMoji-Go.app を Applications フォルダにドラッグ&ドロップ
2. 初回起動時は右クリック→「開く」を選択
3. 「開く」ボタンをクリック

## 初回起動について

署名なしのアプリケーションのため、初回起動時にセキュリティ警告が表示されます。
これは正常な動作です。以下の手順で起動してください：

1. KoeMoji-Go.app を右クリック
2. メニューから「開く」を選択
3. 警告ダイアログで「開く」ボタンをクリック
4. 2回目以降は通常通りダブルクリックで起動可能

## 基本的な使い方

1. アプリを起動
2. input/ フォルダに音声ファイルを配置
3. 自動的に文字起こしが開始されます
4. 結果は output/ フォルダに保存されます

## サポート

- GitHub: https://github.com/infoHiroki/KoeMoji-Go
- Issues: https://github.com/infoHiroki/KoeMoji-Go/issues
EOF
fi

# Create DMG using hdiutil
echo "Creating DMG file..."
hdiutil create \
    -volname "$VOLUME_NAME" \
    -srcfolder "$DMG_DIR" \
    -ov \
    -format UDZO \
    "../releases/${OUTPUT_NAME}.dmg"

# Clean up temporary directory
echo "Cleaning up..."
rm -rf "$DMG_DIR"

echo ""
echo "✅ DMG package created successfully!"
echo "   Location: ../releases/${OUTPUT_NAME}.dmg"
echo ""

# Show DMG info
DMG_SIZE=$(du -h "../releases/${OUTPUT_NAME}.dmg" | cut -f1)
echo "📊 DMG Size: $DMG_SIZE"
echo ""
