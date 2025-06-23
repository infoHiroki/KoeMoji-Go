#!/bin/bash
set -e

VERSION="1.5.0"
APP_NAME="koemoji-go"
DIST_DIR="dist"
SOURCE_DIR="../cmd/koemoji-go"

# Function to show usage
show_usage() {
    echo "Usage: $0 [options]"
    echo "Options:"
    echo "  (no args)   Build standard binary packages"
    echo "  app         Build macOS app bundles"
    echo "  dmg         Build macOS DMG packages"
    echo "  clean       Clean build artifacts"
    echo "  help        Show this help message"
}

# Function to create macOS app bundle
create_macos_app() {
    local arch="$1"
    local binary_name="$2"
    local app_name="KoeMoji-Go"
    
    echo "📱 Creating macOS app bundle for $arch..."
    
    # Create app bundle structure
    local app_dir="$DIST_DIR/${app_name}-${arch}.app"
    mkdir -p "$app_dir/Contents/MacOS"
    mkdir -p "$app_dir/Contents/Resources"
    
    # Copy binary
    cp "$DIST_DIR/$binary_name" "$app_dir/Contents/MacOS/$APP_NAME"
    chmod +x "$app_dir/Contents/MacOS/$APP_NAME"
    
    # Create Info.plist from template
    sed "s/VERSION_PLACEHOLDER/$VERSION/g" "templates/macos/Info.plist.template" > "$app_dir/Contents/Info.plist"
    
    # Copy PkgInfo
    cp "templates/macos/PkgInfo" "$app_dir/Contents/PkgInfo"
    
    # Copy icon if it exists, create if needed
    if [ -f "assets/icons/icon.icns" ]; then
        cp "assets/icons/icon.icns" "$app_dir/Contents/Resources/"
        echo "✅ Icon copied successfully"
    else
        echo "⚠️ Creating icon from PNG/ICO source..."
        cd scripts && ./create-icon.sh && cd ..
        if [ -f "assets/icons/icon.icns" ]; then
            cp "assets/icons/icon.icns" "$app_dir/Contents/Resources/"
            echo "✅ Icon created and copied successfully"
        else
            echo "⚠️ Warning: Could not create icon, app will use default icon"
        fi
    fi
    
    # Copy default config
    cp "assets/config.example.json" "$app_dir/Contents/Resources/config.json"
    
    # Sign the app bundle (adhoc signature for local use)
    echo "🔏 Signing app bundle..."
    codesign --force --deep --sign - "$app_dir" 2>/dev/null || echo "⚠️ Signing failed, but app may still work"
    
    echo "✅ App bundle created: $app_dir"
}

# Function to create DMG
create_dmg() {
    local arch="$1"
    local app_name="KoeMoji-Go"
    local app_dir="$DIST_DIR/${app_name}-${arch}.app"
    local dmg_name="KoeMoji-Go-${arch}-${VERSION}.dmg"
    
    echo "💿 Creating DMG for $arch..."
    
    if [ ! -d "$app_dir" ]; then
        echo "❌ Error: App bundle not found: $app_dir"
        echo "Run '$0 app' first to create app bundles"
        return 1
    fi
    
    # Create temporary DMG directory
    local temp_dmg_dir="$DIST_DIR/dmg-temp-$arch"
    mkdir -p "$temp_dmg_dir"
    
    # Copy app bundle to temp directory
    cp -R "$app_dir" "$temp_dmg_dir/"
    
    # Create DMG
    hdiutil create -volname "$app_name" -srcfolder "$temp_dmg_dir" -ov -format UDZO "$DIST_DIR/$dmg_name"
    
    # Clean up temp directory
    rm -rf "$temp_dmg_dir"
    
    echo "✅ DMG created: $DIST_DIR/$dmg_name"
}

# Parse command line arguments
case "${1:-}" in
    "help"|"-h"|"--help")
        show_usage
        exit 0
        ;;
    "clean")
        echo "🧹 Cleaning build artifacts..."
        rm -rf $DIST_DIR
        rm -rf temp/*
        echo "✅ Clean completed"
        exit 0
        ;;
    "app")
        BUILD_APPS=true
        ;;
    "dmg")
        BUILD_DMG=true
        ;;
    "")
        # Default behavior - build standard packages
        ;;
    *)
        echo "❌ Unknown option: $1"
        show_usage
        exit 1
        ;;
esac

echo "🚀 Building KoeMoji-Go..."

# Clean and prepare
rm -rf $DIST_DIR
mkdir -p $DIST_DIR

# Skip Windows build when only building apps
if [ "$BUILD_APPS" != true ] && [ "$BUILD_DMG" != true ]; then
    # Windows with icon
    echo "🪟 Building Windows with icon..."
    if ! command -v goversioninfo &> /dev/null; then
        echo "📦 Installing goversioninfo..."
        go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
        export PATH=$PATH:$(go env GOPATH)/bin
    fi

    echo "🎨 Generating Windows resource file..."
    $(go env GOPATH)/bin/goversioninfo -o temp/resource.syso templates/windows/versioninfo.json

    echo "🔨 Building Windows executable..."
    GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $DIST_DIR/${APP_NAME}.exe $SOURCE_DIR

    # Clean up resource file
    rm -f temp/resource.syso

    echo "✅ Windows build completed"
fi

# macOS builds
echo "🍎 Building macOS..."

# Build for current architecture only when building apps (due to CGO dependencies)
if [ "$BUILD_APPS" = true ] || [ "$BUILD_DMG" = true ]; then
    echo "🔨 Building macOS (current architecture)..."
    go build -ldflags="-s -w" -o $DIST_DIR/${APP_NAME}-native $SOURCE_DIR
    echo "✅ Native macOS build completed"
else
    echo "🔨 Building macOS Intel..."
    GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $DIST_DIR/${APP_NAME}-darwin-amd64 $SOURCE_DIR

    echo "🔨 Building macOS Apple Silicon..."
    GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $DIST_DIR/${APP_NAME}-darwin-arm64 $SOURCE_DIR

    echo "✅ macOS builds completed"
fi


# Create distribution packages (only if not building apps/dmg only)
if [ "$BUILD_APPS" != true ] && [ "$BUILD_DMG" != true ]; then
    echo "📦 Creating distribution packages..."

    cd $DIST_DIR

    # Windows package
    if [ -f "${APP_NAME}.exe" ]; then
        echo "📦 Creating Windows package..."
        mkdir -p koemoji-go-windows-$VERSION
        cp ${APP_NAME}.exe koemoji-go-windows-$VERSION/
        cp ../assets/config.example.json koemoji-go-windows-$VERSION/config.json
        cp ../../README.md koemoji-go-windows-$VERSION/
        zip -r koemoji-go-windows-$VERSION.zip koemoji-go-windows-$VERSION/
        rm -rf koemoji-go-windows-$VERSION
    fi

    # macOS Intel package
    if [ -f "${APP_NAME}-darwin-amd64" ]; then
        echo "📦 Creating macOS Intel package..."
        mkdir -p koemoji-go-macos-intel-$VERSION
        cp ${APP_NAME}-darwin-amd64 koemoji-go-macos-intel-$VERSION/${APP_NAME}
        cp ../assets/config.example.json koemoji-go-macos-intel-$VERSION/config.json
        cp ../../README.md koemoji-go-macos-intel-$VERSION/
        tar -czf koemoji-go-macos-intel-$VERSION.tar.gz koemoji-go-macos-intel-$VERSION/
        rm -rf koemoji-go-macos-intel-$VERSION
    fi

    # macOS Apple Silicon package
    if [ -f "${APP_NAME}-darwin-arm64" ]; then
        echo "📦 Creating macOS Apple Silicon package..."
        mkdir -p koemoji-go-macos-arm64-$VERSION
        cp ${APP_NAME}-darwin-arm64 koemoji-go-macos-arm64-$VERSION/${APP_NAME}
        cp ../assets/config.example.json koemoji-go-macos-arm64-$VERSION/config.json
        cp ../../README.md koemoji-go-macos-arm64-$VERSION/
        tar -czf koemoji-go-macos-arm64-$VERSION.tar.gz koemoji-go-macos-arm64-$VERSION/
        rm -rf koemoji-go-macos-arm64-$VERSION
    fi
else
    cd $DIST_DIR
fi


cd ..

# Build app bundles if requested
if [ "$BUILD_APPS" = true ]; then
    echo ""
    echo "📱 Creating macOS app bundle..."
    
    # Determine current architecture
    CURRENT_ARCH=$(uname -m)
    if [ "$CURRENT_ARCH" = "arm64" ]; then
        ARCH_NAME="arm64"
    else
        ARCH_NAME="intel"
    fi
    
    create_macos_app "$ARCH_NAME" "${APP_NAME}-native"
fi

# Build DMGs if requested
if [ "$BUILD_DMG" = true ]; then
    echo ""
    echo "💿 Creating DMG package..."
    
    # Use the same architecture as the app bundle
    CURRENT_ARCH=$(uname -m)
    if [ "$CURRENT_ARCH" = "arm64" ]; then
        ARCH_NAME="arm64"
    else
        ARCH_NAME="intel"
    fi
    
    create_dmg "$ARCH_NAME"
fi

echo ""
echo "🎉 Build completed successfully!"
echo ""
echo "📁 Distribution files created in $DIST_DIR/:"
ls -la $DIST_DIR/
echo ""
echo "🚀 Ready for distribution!"

if [ "$BUILD_APPS" = true ]; then
    CURRENT_ARCH=$(uname -m)
    if [ "$CURRENT_ARCH" = "arm64" ]; then
        ARCH_NAME="arm64"
    else
        ARCH_NAME="intel"
    fi
    
    echo ""
    echo "📱 App bundle can be launched by:"
    echo "  - Double-clicking in Finder"
    echo "  - Running: open $DIST_DIR/KoeMoji-Go-$ARCH_NAME.app"
fi

if [ "$BUILD_DMG" = true ]; then
    echo ""
    echo "💿 DMG files ready for distribution:"
    echo "  - Mount and drag to Applications folder"
    echo "  - Or distribute directly to users"
fi