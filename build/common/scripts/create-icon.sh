#!/bin/bash

# macOS icon creation script
set -e

ICONS_DIR="../assets/icons"
ICONSET_DIR="$ICONS_DIR/icon.iconset"
ICON_PNG="$ICONS_DIR/icon.png"
ICON_ICO="$ICONS_DIR/icon.ico"
ICON_ICNS="$ICONS_DIR/icon.icns"

echo "🎨 Creating macOS icon from PNG source..."

# Create iconset directory if it doesn't exist
mkdir -p "$ICONSET_DIR"

# Check if icon.png exists, fallback to icon.ico
if [ -f "$ICON_PNG" ]; then
    ICON_SOURCE="$ICON_PNG"
    echo "✅ Using high-resolution PNG source: $ICON_PNG"
elif [ -f "$ICON_ICO" ]; then
    ICON_SOURCE="$ICON_ICO"
    echo "⚠️ Using lower-resolution ICO source: $ICON_ICO"
else
    echo "❌ Error: Neither $ICON_PNG nor $ICON_ICO found"
    exit 1
fi

# Use sips to convert and resize to various sizes
echo "🔄 Converting to multiple sizes..."

# Generate all required icon sizes
sips -z 16 16     "$ICON_SOURCE" --out "$ICONSET_DIR/icon_16x16.png" 2>/dev/null || echo "⚠️ Failed to create 16x16"
sips -z 32 32     "$ICON_SOURCE" --out "$ICONSET_DIR/icon_16x16@2x.png" 2>/dev/null || echo "⚠️ Failed to create 16x16@2x"
sips -z 32 32     "$ICON_SOURCE" --out "$ICONSET_DIR/icon_32x32.png" 2>/dev/null || echo "⚠️ Failed to create 32x32"
sips -z 64 64     "$ICON_SOURCE" --out "$ICONSET_DIR/icon_32x32@2x.png" 2>/dev/null || echo "⚠️ Failed to create 32x32@2x"
sips -z 128 128   "$ICON_SOURCE" --out "$ICONSET_DIR/icon_128x128.png" 2>/dev/null || echo "⚠️ Failed to create 128x128"
sips -z 256 256   "$ICON_SOURCE" --out "$ICONSET_DIR/icon_128x128@2x.png" 2>/dev/null || echo "⚠️ Failed to create 128x128@2x"
sips -z 256 256   "$ICON_SOURCE" --out "$ICONSET_DIR/icon_256x256.png" 2>/dev/null || echo "⚠️ Failed to create 256x256"
sips -z 512 512   "$ICON_SOURCE" --out "$ICONSET_DIR/icon_256x256@2x.png" 2>/dev/null || echo "⚠️ Failed to create 256x256@2x"
sips -z 512 512   "$ICON_SOURCE" --out "$ICONSET_DIR/icon_512x512.png" 2>/dev/null || echo "⚠️ Failed to create 512x512"
sips -z 1024 1024 "$ICON_SOURCE" --out "$ICONSET_DIR/icon_512x512@2x.png" 2>/dev/null || echo "⚠️ Failed to create 512x512@2x"

# Create icns file
echo "📦 Creating icon.icns..."
iconutil -c icns "$ICONSET_DIR" -o "$ICON_ICNS"

if [ -f "$ICON_ICNS" ]; then
    echo "✅ Successfully created $ICON_ICNS"
    
    # Clean up iconset directory
    rm -rf "$ICONSET_DIR"
    echo "🧹 Cleaned up temporary iconset directory"
else
    echo "❌ Failed to create icon.icns"
    exit 1
fi

echo "🎉 Icon conversion completed!"