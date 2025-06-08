#!/bin/bash

# KoeMoji-Go Build Script with Icon Support
# Builds application with icons for Windows, macOS, and Linux

set -e

VERSION="1.0.0"
APP_NAME="koemoji-go"
DIST_DIR="dist"

echo "🚀 Starting KoeMoji-Go build with icon support..."

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -rf $DIST_DIR
mkdir -p $DIST_DIR

# Check for required tools
check_tools() {
    echo "🔍 Checking required tools..."
    
    if ! command -v go &> /dev/null; then
        echo "❌ Go is not installed"
        exit 1
    fi
    
    echo "✅ Go found: $(go version)"
}

# Build Windows with icon
build_windows() {
    echo "🪟 Building Windows version with icon..."
    
    # Install goversioninfo if not present
    if ! command -v goversioninfo &> /dev/null; then
        echo "📦 Installing goversioninfo..."
        go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
        # Add GOPATH/bin to PATH if not already there
        export PATH=$PATH:$(go env GOPATH)/bin
    fi
    
    # Generate Windows resource file
    echo "🎨 Generating Windows resource file..."
    $(go env GOPATH)/bin/goversioninfo -o resource.syso versioninfo.json
    
    # Build Windows executable
    echo "🔨 Building Windows executable..."
    GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $DIST_DIR/${APP_NAME}-windows-amd64.exe main.go
    
    # Clean up resource file
    rm -f resource.syso
    
    echo "✅ Windows build completed"
}

# Build macOS
build_macos() {
    echo "🍎 Building macOS versions..."
    
    # Build for Intel
    echo "🔨 Building macOS Intel..."
    GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $DIST_DIR/${APP_NAME}-darwin-amd64 main.go
    
    # Build for Apple Silicon
    echo "🔨 Building macOS Apple Silicon..."
    GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $DIST_DIR/${APP_NAME}-darwin-arm64 main.go
    
    # Create universal binary
    echo "🔗 Creating universal binary..."
    lipo -create -output $DIST_DIR/${APP_NAME}-darwin-universal $DIST_DIR/${APP_NAME}-darwin-amd64 $DIST_DIR/${APP_NAME}-darwin-arm64
    
    echo "✅ macOS builds completed"
}


# Build Linux
build_linux() {
    echo "🐧 Building Linux version..."
    
    GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $DIST_DIR/${APP_NAME}-linux-amd64 main.go
    
    echo "✅ Linux build completed"
}

# Create distribution packages
create_packages() {
    echo "📦 Creating distribution packages..."
    
    cd $DIST_DIR
    
    # Windows package
    echo "📦 Creating Windows package..."
    mkdir -p koemoji-go-windows-$VERSION
    cp ${APP_NAME}-windows-amd64.exe koemoji-go-windows-$VERSION/
    cp ../config.json koemoji-go-windows-$VERSION/
    cp ../README.md koemoji-go-windows-$VERSION/
    zip -r koemoji-go-windows-$VERSION.zip koemoji-go-windows-$VERSION/
    rm -rf koemoji-go-windows-$VERSION
    
    # macOS package
    echo "📦 Creating macOS package..."
    mkdir -p koemoji-go-macos-$VERSION
    cp -R KoeMoji-Go.app koemoji-go-macos-$VERSION/
    cp ../config.json koemoji-go-macos-$VERSION/
    cp ../README.md koemoji-go-macos-$VERSION/
    tar -czf koemoji-go-macos-$VERSION.tar.gz koemoji-go-macos-$VERSION/
    rm -rf koemoji-go-macos-$VERSION
    
    # Linux package
    echo "📦 Creating Linux package..."
    mkdir -p koemoji-go-linux-$VERSION
    cp ${APP_NAME}-linux-amd64 koemoji-go-linux-$VERSION/$APP_NAME
    cp ${APP_NAME}.desktop koemoji-go-linux-$VERSION/
    cp ../icon.ico koemoji-go-linux-$VERSION/icon.png  # Use ico as png for Linux
    cp ../config.json koemoji-go-linux-$VERSION/
    cp ../README.md koemoji-go-linux-$VERSION/
    
    # Create install script for Linux
    cat > koemoji-go-linux-$VERSION/install.sh << 'EOF'
#!/bin/bash
# KoeMoji-Go Linux installer

echo "Installing KoeMoji-Go..."

# Copy binary
sudo cp koemoji-go /usr/local/bin/
sudo chmod +x /usr/local/bin/koemoji-go

# Copy desktop file
mkdir -p ~/.local/share/applications
cp koemoji-go.desktop ~/.local/share/applications/

# Copy icon
mkdir -p ~/.local/share/icons
cp icon.png ~/.local/share/icons/koemoji-go.png

echo "Installation completed!"
echo "You can now run 'koemoji-go' from terminal or find it in your applications menu."
EOF
    chmod +x koemoji-go-linux-$VERSION/install.sh
    
    tar -czf koemoji-go-linux-$VERSION.tar.gz koemoji-go-linux-$VERSION/
    rm -rf koemoji-go-linux-$VERSION
    
    cd ..
    echo "✅ Distribution packages created"
}

# Show build summary
show_summary() {
    echo ""
    echo "🎉 Build completed successfully!"
    echo ""
    echo "📁 Distribution files created in $DIST_DIR/:"
    ls -la $DIST_DIR/
    echo ""
    echo "🚀 Ready for distribution!"
}

# Main build process
main() {
    check_tools
    
    # Ensure GOPATH/bin is in PATH
    export PATH=$PATH:$(go env GOPATH)/bin
    
    build_windows
    build_macos
    build_linux
    create_packages
    show_summary
}

# Handle command line arguments
case "$1" in
    "windows")
        check_tools
        build_windows
        ;;
    "macos")
        check_tools
        build_macos
        ;;
    "linux")
        check_tools
        build_linux
        ;;
    "clean")
        echo "🧹 Cleaning build artifacts..."
        rm -rf $DIST_DIR
        echo "✅ Clean completed"
        ;;
    *)
        main
        ;;
esac