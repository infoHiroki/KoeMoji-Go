#!/bin/bash
set -e

# スクリプトのディレクトリを取得
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# バージョン情報をversion.goから動的に取得
VERSION=$(grep -o 'const Version = "[^"]*"' "$PROJECT_ROOT/version.go" | cut -d'"' -f2)
APP_NAME="koemoji-go"
DIST_DIR="dist"
SOURCE_DIR="$PROJECT_ROOT/cmd/koemoji-go"
COMMON_DIR="$SCRIPT_DIR/../common"

# Function to show usage
show_usage() {
    echo "Usage: $0 [options]"
    echo "Options:"
    echo "  build       Build release version (tar.gz) [default]"
    echo "  clean       Clean build artifacts"
    echo "  help        Show this help message"
    echo ""
    echo "Default (no args): Build release version"
}

# Function to build for specific architecture
build_arch() {
    local arch="$1"
    local binary_name="${APP_NAME}-${arch}"

    echo "🍎 Building macOS $arch binary..."

    if [ "$arch" = "arm64" ]; then
        cd "$PROJECT_ROOT"
        GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.version=$VERSION" -o "$SCRIPT_DIR/$DIST_DIR/$binary_name" ./cmd/koemoji-go
        cd "$SCRIPT_DIR"
    else
        echo "❌ Unsupported architecture: $arch"
        return 1
    fi

    if [ $? -eq 0 ]; then
        echo "✅ $arch binary build completed"
    else
        echo "❌ $arch binary build failed"
        return 1
    fi
}

# Function to build .app bundle
build_app() {
    local arch="$1"

    echo "📱 Building .app bundle for $arch..."

    # Build binary with correct name for .app bundle
    echo "🍎 Building macOS $arch binary for .app..."
    cd "$PROJECT_ROOT"
    local app_binary_path="$SCRIPT_DIR/$DIST_DIR/$APP_NAME"
    GOOS=darwin GOARCH=$arch go build -ldflags="-s -w -X main.version=$VERSION" -o "$app_binary_path" ./cmd/koemoji-go
    cd "$SCRIPT_DIR"

    if [ $? -ne 0 ]; then
        echo "❌ Failed to build binary for .app"
        return 1
    fi
    echo "✅ Binary built for .app bundle"

    # Check if fyne is available
    if ! command -v fyne &> /dev/null; then
        if [ -f "$HOME/go/bin/fyne" ]; then
            FYNE_CMD="$HOME/go/bin/fyne"
        else
            echo "❌ Error: fyne command not found"
            echo "Please install: go install fyne.io/tools/cmd/fyne@latest"
            exit 1
        fi
    else
        FYNE_CMD="fyne"
    fi

    # Use fyne package to create .app bundle from existing binary
    local binary_path="$app_binary_path"

    cd "$PROJECT_ROOT"

    echo "Running fyne package..."
    $FYNE_CMD package -os darwin \
        --icon "$PROJECT_ROOT/Icon.png" \
        --executable "$binary_path" \
        --release

    # Move .app to dist directory
    if [ -d "KoeMoji-Go.app" ]; then
        rm -rf "$SCRIPT_DIR/$DIST_DIR/KoeMoji-Go.app"
        mv "KoeMoji-Go.app" "$SCRIPT_DIR/$DIST_DIR/"
        echo "✅ .app bundle created: $DIST_DIR/KoeMoji-Go.app"
    else
        echo "❌ Failed to create .app bundle"
        exit 1
    fi

    cd "$SCRIPT_DIR"

    # Bundle PortAudio library for self-contained distribution
    echo "📦 Bundling PortAudio library..."
    local app_path="$SCRIPT_DIR/$DIST_DIR/KoeMoji-Go.app"
    local frameworks_dir="$app_path/Contents/Frameworks"
    local portaudio_src="/opt/homebrew/opt/portaudio/lib/libportaudio.2.dylib"

    # Check if PortAudio library exists
    if [ ! -f "$portaudio_src" ]; then
        echo "❌ Error: PortAudio library not found at $portaudio_src"
        echo "Please install: brew install portaudio"
        exit 1
    fi

    # Create Frameworks directory
    mkdir -p "$frameworks_dir"

    # Copy PortAudio library to bundle
    cp "$portaudio_src" "$frameworks_dir/"

    # Update library references in binary using install_name_tool
    install_name_tool -change \
        "$portaudio_src" \
        "@executable_path/../Frameworks/libportaudio.2.dylib" \
        "$app_path/Contents/MacOS/koemoji-go"

    # Update library ID in the bundled library
    install_name_tool -id \
        "@executable_path/../Frameworks/libportaudio.2.dylib" \
        "$frameworks_dir/libportaudio.2.dylib"

    echo "✅ PortAudio library bundled successfully"

    # Add microphone permission to Info.plist
    echo "🎤 Adding microphone permission to Info.plist..."
    /usr/libexec/PlistBuddy -c "Add :NSMicrophoneUsageDescription string 'KoeMoji-Goはマイクを使用して音声を録音し、文字起こしを行います。録音機能を使用するにはマイクへのアクセスを許可してください。'" "$app_path/Contents/Info.plist" 2>/dev/null || \
    /usr/libexec/PlistBuddy -c "Set :NSMicrophoneUsageDescription 'KoeMoji-Goはマイクを使用して音声を録音し、文字起こしを行います。録音機能を使用するにはマイクへのアクセスを許可してください。'" "$app_path/Contents/Info.plist"
    echo "✅ Microphone permission added to Info.plist"

    # Re-sign the app with ad-hoc signature (required after install_name_tool)
    echo "✍️  Re-signing app bundle with ad-hoc signature..."
    codesign --force --deep --sign - "$app_path"

    if [ $? -eq 0 ]; then
        echo "✅ App bundle signed successfully"
    else
        echo "⚠️  Warning: Code signing failed, but continuing..."
    fi

    # Verify the changes
    echo "🔍 Verifying library dependencies..."
    otool -L "$app_path/Contents/MacOS/koemoji-go" | grep portaudio
}

# Function to build DMG package
build_dmg() {
    local arch="$1"
    local release_name="koemoji-go-macos-${VERSION}"

    echo "💿 Building DMG package..."

    # Build .app bundle first
    build_app "$arch"

    # Check if create-dmg.sh exists
    if [ ! -f "create-dmg.sh" ]; then
        echo "❌ Error: create-dmg.sh not found"
        exit 1
    fi

    # Run DMG creation script
    ./create-dmg.sh "$DIST_DIR/KoeMoji-Go.app" "$release_name"

    echo "✅ DMG package created: ../releases/${release_name}.dmg"
}

# Function to build release version (tar.gz)
build_cli() {
    local arch="$1"
    local binary_name="${APP_NAME}-${arch}"
    local package_name="koemoji-go-${VERSION}"
    local release_name="koemoji-go-macos-${VERSION}"

    echo "🖥️  Building release version for $arch..."

    # Build binary
    build_arch "$arch"

    echo "📦 Creating release package..."

    # Create package directory
    rm -rf "$package_name"
    mkdir -p "$package_name"

    # Copy binary and make executable
    cp "$DIST_DIR/$binary_name" "$package_name/$APP_NAME"
    chmod +x "$package_name/$APP_NAME"

    # Copy Swift CLI binary (audio-capture) for macOS system audio recording
    if [ -f "$PROJECT_ROOT/cmd/audio-capture/audio-capture" ]; then
        cp "$PROJECT_ROOT/cmd/audio-capture/audio-capture" "$package_name/audio-capture"
        chmod +x "$package_name/audio-capture"
        echo "✅ Swift CLI (audio-capture) included"
    else
        echo "⚠️  Warning: Swift CLI binary (audio-capture) not found"
        echo "   Dual recording will not work without this binary"
    fi

    # Copy config file
    cp "$COMMON_DIR/assets/config.example.json" "$package_name/config.json"

    # Copy release README
    cp "$COMMON_DIR/assets/README_RELEASE.md" "$package_name/README.md"

    # Copy diagnostic script
    cp "$SCRIPT_DIR/診断実行.command" "$package_name/診断実行.command"
    chmod +x "$package_name/診断実行.command"

    # Create tar.gz
    tar -czf "../releases/${release_name}.tar.gz" "$package_name"

    # Clean up temporary directory
    rm -rf "$package_name"

    echo "✅ Release package created: ../releases/${release_name}.tar.gz"
}

# Parse command line arguments
case "${1:-}" in
    "help"|"-h"|"--help")
        show_usage
        exit 0
        ;;
    "clean")
        echo "🧹 Cleaning macOS build artifacts..."
        rm -rf $DIST_DIR
        rm -rf ../releases/koemoji-go-macos-*.tar.gz
        echo "✅ Clean completed"
        exit 0
        ;;
    "build"|"")
        # Default behavior - build release version (tar.gz)
        BUILD_TYPE="build"
        ;;
    *)
        echo "❌ Unknown option: $1"
        show_usage
        exit 1
        ;;
esac

echo "🚀 Building KoeMoji-Go v${VERSION} for macOS..."
echo ""

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed or not in PATH"
    echo "Please install Go 1.21 or later from https://golang.org/"
    exit 1
fi

echo "Go version: $(go version)"

# Clean and prepare
echo "📁 Preparing directories..."
rm -rf $DIST_DIR
mkdir -p $DIST_DIR
mkdir -p ../releases

# Build based on type
case "$BUILD_TYPE" in
    "build")
        build_cli "arm64"
        echo ""
        echo "🎉 macOS build completed successfully!"
        echo ""
        echo "📦 Distribution file created in: ../releases/"
        echo "   - koemoji-go-macos-${VERSION}.tar.gz"
        ;;
esac

echo ""
