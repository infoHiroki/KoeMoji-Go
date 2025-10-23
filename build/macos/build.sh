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
    echo "  app         Build .app bundle only"
    echo "  dmg         Build DMG package (includes .app)"
    echo "  cli         Build CLI version (tar.gz)"
    echo "  all         Build both DMG and CLI versions"
    echo "  clean       Clean build artifacts"
    echo "  help        Show this help message"
    echo ""
    echo "Default (no args): Build CLI version"
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

    # Build binary first
    build_arch "$arch"

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
    local binary_name="${APP_NAME}-${arch}"
    local binary_path="$SCRIPT_DIR/$DIST_DIR/$binary_name"

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
}

# Function to build DMG package
build_dmg() {
    local arch="$1"
    local release_name="koemoji-go-${VERSION}-macos"

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

# Function to build CLI version (tar.gz)
build_cli() {
    local arch="$1"
    local binary_name="${APP_NAME}-${arch}"
    local package_name="koemoji-go-${VERSION}"
    local release_name="koemoji-go-${VERSION}-macos-cli"

    echo "🖥️  Building CLI version for $arch..."

    # Build binary
    build_arch "$arch"

    echo "📦 Creating CLI package..."

    # Create package directory
    rm -rf "$package_name"
    mkdir -p "$package_name"

    # Copy binary and make executable
    cp "$DIST_DIR/$binary_name" "$package_name/$APP_NAME"
    chmod +x "$package_name/$APP_NAME"

    # Copy config file
    cp "$COMMON_DIR/assets/config.example.json" "$package_name/config.json"

    # Copy release README
    cp "$COMMON_DIR/assets/README_RELEASE.md" "$package_name/README.md"

    # Create tar.gz
    tar -czf "../releases/${release_name}.tar.gz" "$package_name"

    # Clean up temporary directory
    rm -rf "$package_name"

    echo "✅ CLI package created: ../releases/${release_name}.tar.gz"
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
        rm -rf ../releases/koemoji-go-*-macos.dmg
        rm -rf ../releases/koemoji-go-*-macos-cli.tar.gz
        echo "✅ Clean completed"
        exit 0
        ;;
    "app")
        BUILD_TYPE="app"
        ;;
    "dmg")
        BUILD_TYPE="dmg"
        ;;
    "cli")
        BUILD_TYPE="cli"
        ;;
    "all")
        BUILD_TYPE="all"
        ;;
    "")
        # Default behavior - build CLI version
        BUILD_TYPE="cli"
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
    "app")
        build_app "arm64"
        echo ""
        echo "🎉 macOS .app build completed successfully!"
        echo ""
        echo "📱 App bundle created in: $DIST_DIR/"
        echo "   - KoeMoji-Go.app"
        ;;
    "dmg")
        build_dmg "arm64"
        echo ""
        echo "🎉 macOS DMG build completed successfully!"
        echo ""
        echo "📦 Distribution file created in: ../releases/"
        echo "   - koemoji-go-${VERSION}-macos.dmg"
        ;;
    "cli")
        build_cli "arm64"
        echo ""
        echo "🎉 macOS CLI build completed successfully!"
        echo ""
        echo "📦 Distribution file created in: ../releases/"
        echo "   - koemoji-go-${VERSION}-macos-cli.tar.gz"
        ;;
    "all")
        build_dmg "arm64"
        build_cli "arm64"
        echo ""
        echo "🎉 macOS build completed successfully!"
        echo ""
        echo "📦 Distribution files created in: ../releases/"
        echo "   - koemoji-go-${VERSION}-macos.dmg"
        echo "   - koemoji-go-${VERSION}-macos-cli.tar.gz"
        ;;
esac

echo ""
