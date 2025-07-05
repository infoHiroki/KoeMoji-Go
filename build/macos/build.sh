#!/bin/bash
set -e

# バージョン情報をversion.goから動的に取得
VERSION=$(grep -o 'const Version = "[^"]*"' ../../version.go | cut -d'"' -f2)
APP_NAME="koemoji-go"
DIST_DIR="dist"
SOURCE_DIR="../../cmd/koemoji-go"
COMMON_DIR="../common"

# Function to show usage
show_usage() {
    echo "Usage: $0 [options]"
    echo "Options:"
    echo "  (no args)   Build Apple Silicon tar.gz package"
    echo "  arm64       Build Apple Silicon version only"
    echo "  clean       Clean build artifacts"
    echo "  help        Show this help message"
}

# Function to build for specific architecture
build_arch() {
    local arch="$1"
    local binary_name="${APP_NAME}-${arch}"
    
    echo "🍎 Building macOS $arch..."
    
    if [ "$arch" = "arm64" ]; then
        GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.version=$VERSION" -o "$DIST_DIR/$binary_name" "$SOURCE_DIR"
    else
        echo "❌ Unsupported architecture: $arch"
        return 1
    fi
    
    if [ $? -eq 0 ]; then
        echo "✅ $arch build completed"
    else
        echo "❌ $arch build failed"
        return 1
    fi
}

# Function to create distribution package
create_package() {
    local arch="$1"
    local binary_name="${APP_NAME}-${arch}"
    local package_name="${APP_NAME}-macos-${arch}-${VERSION}"
    
    echo "📦 Creating $arch package..."
    
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
    
    # Create tar.gz with new naming convention
    release_name="KoeMoji-Go-v${VERSION}-mac"
    tar -czf "../releases/${release_name}.tar.gz" "$package_name"
    
    # Clean up temporary directory
    rm -rf "$package_name"
    
    echo "✅ Package created: ../releases/${release_name}.tar.gz"
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
        rm -rf ../releases/KoeMoji-Go-v*-mac.tar.gz
        echo "✅ Clean completed"
        exit 0
        ;;
    "arm64")
        BUILD_ARM64_ONLY=true
        ;;
    "")
        # Default behavior - build Apple Silicon only
        ;;
    *)
        echo "❌ Unknown option: $1"
        show_usage
        exit 1
        ;;
esac

echo "🚀 Building KoeMoji-Go for macOS..."

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

# Build architectures
if [ "$BUILD_ARM64_ONLY" = true ]; then
    build_arch "arm64"
    create_package "arm64"
else
    # Default - build Apple Silicon only
    build_arch "arm64"
    create_package "arm64"
fi

echo ""
echo "🎉 macOS build completed successfully!"
echo ""
echo "📦 Distribution file created in: ../releases/"
echo "   - KoeMoji-Go-v${VERSION}-mac.tar.gz"
echo ""