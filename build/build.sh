#!/bin/bash
set -e

VERSION="1.1.0"
APP_NAME="koemoji-go"
DIST_DIR="dist"
SOURCE_DIR="../cmd/koemoji-go"

echo "🚀 Building KoeMoji-Go..."

# Clean and prepare
rm -rf $DIST_DIR
mkdir -p $DIST_DIR

# Windows with icon
echo "🪟 Building Windows with icon..."
if ! command -v goversioninfo &> /dev/null; then
    echo "📦 Installing goversioninfo..."
    go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
    export PATH=$PATH:$(go env GOPATH)/bin
fi

echo "🎨 Generating Windows resource file..."
$(go env GOPATH)/bin/goversioninfo -o resource.syso versioninfo.json

echo "🔨 Building Windows executable..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $DIST_DIR/${APP_NAME}.exe $SOURCE_DIR

# Clean up resource file
rm -f resource.syso

echo "✅ Windows build completed"

# macOS
echo "🍎 Building macOS..."
echo "🔨 Building macOS Intel..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $DIST_DIR/${APP_NAME}-darwin-amd64 $SOURCE_DIR

echo "🔨 Building macOS Apple Silicon..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $DIST_DIR/${APP_NAME}-darwin-arm64 $SOURCE_DIR

echo "✅ macOS builds completed"

# Linux
echo "🐧 Building Linux..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $DIST_DIR/${APP_NAME}-linux-amd64 $SOURCE_DIR

echo "✅ Linux build completed"

# Create distribution packages
echo "📦 Creating distribution packages..."

cd $DIST_DIR

# Windows package
echo "📦 Creating Windows package..."
mkdir -p koemoji-go-windows-$VERSION
cp ${APP_NAME}.exe koemoji-go-windows-$VERSION/
cp ../../config.example.json koemoji-go-windows-$VERSION/config.json
cp ../../README.md koemoji-go-windows-$VERSION/
zip -r koemoji-go-windows-$VERSION.zip koemoji-go-windows-$VERSION/
rm -rf koemoji-go-windows-$VERSION

# macOS package
echo "📦 Creating macOS package..."
mkdir -p koemoji-go-macos-$VERSION
cp ${APP_NAME}-darwin-amd64 koemoji-go-macos-$VERSION/
cp ${APP_NAME}-darwin-arm64 koemoji-go-macos-$VERSION/
cp ../../config.example.json koemoji-go-macos-$VERSION/config.json
cp ../../README.md koemoji-go-macos-$VERSION/
tar -czf koemoji-go-macos-$VERSION.tar.gz koemoji-go-macos-$VERSION/
rm -rf koemoji-go-macos-$VERSION

# Linux package
echo "📦 Creating Linux package..."
mkdir -p koemoji-go-linux-$VERSION
cp ${APP_NAME}-linux-amd64 koemoji-go-linux-$VERSION/${APP_NAME}
cp ../../config.example.json koemoji-go-linux-$VERSION/config.json
cp ../../README.md koemoji-go-linux-$VERSION/
tar -czf koemoji-go-linux-$VERSION.tar.gz koemoji-go-linux-$VERSION/
rm -rf koemoji-go-linux-$VERSION

cd ..

echo ""
echo "🎉 Build completed successfully!"
echo ""
echo "📁 Distribution files created in $DIST_DIR/:"
ls -la $DIST_DIR/
echo ""
echo "🚀 Ready for distribution!"