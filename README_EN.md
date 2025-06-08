# KoeMoji-Go

Automatic Audio/Video Transcription Tool

[日本語版 README](README.md) | **English README**

## Overview

KoeMoji-Go is a Go application that automatically transcribes audio and video files.
It's a Go port of the Python-based KoeMojiAuto-cli, providing single binary distribution and stable sequential processing.

## Features

- **Single Binary**: Works with just one executable file
- **Sequential Processing**: Stable processing one file at a time
- **FasterWhisper Integration**: High-accuracy speech recognition
- **Cross-Platform**: Windows/Mac/Linux support
- **Auto Monitoring**: Automatically monitors folders and processes files
- **Real-time UI**: Real-time display of processing status

## 1. System Requirements

### Hardware Requirements
- **OS**: Windows 10/11, macOS 10.15+, Linux (major distributions)
- **CPU**: Intel/AMD 64bit, Apple Silicon
- **Memory**: 4GB+ recommended (8GB+ for better performance)
- **Storage**: 5GB+ (including Whisper model files)

### Prerequisites

#### Python 3.8+ Installation
KoeMoji-Go uses FasterWhisper, so **Python 3.8+ is required**.

**Check Python version:**
```bash
python --version
# or
python3 --version
```

**If Python is not installed:**

**Windows:**
1. Download from [Python official site](https://www.python.org/downloads/windows/)
2. Check "Add Python to PATH" during installation
3. Recommended: Python 3.11+

**macOS:**
```bash
# Using Homebrew
brew install python

# Or download from official site
# https://www.python.org/downloads/macos/
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt update
sudo apt install python3 python3-pip
```

**Linux (CentOS/RHEL):**
```bash
sudo yum install python3 python3-pip
# or
sudo dnf install python3 python3-pip
```

#### Check pip
```bash
pip --version
# or
pip3 --version
```

If pip is not available:
```bash
# macOS/Linux
python3 -m ensurepip --upgrade

# Windows
python -m ensurepip --upgrade
```

## 2. Installation

### Download
1. **Download the version for your OS from [GitHub Releases](https://github.com/[username]/koemoji-go/releases)**

**Windows**: `koemoji-go-windows-1.0.0.zip`
```
📁 koemoji-go-windows-1.0.0.zip
├── koemoji-go.exe     # Executable with icon
├── config.json        # Configuration file
└── README.md          # Documentation
```

**macOS**: `koemoji-go-macos-1.0.0.tar.gz`
```
📁 koemoji-go-macos-1.0.0.tar.gz  
├── koemoji-go-darwin-amd64    # Intel Mac executable
├── koemoji-go-darwin-arm64    # Apple Silicon executable
├── config.json                # Configuration file
└── README.md                  # Documentation
```

**Linux**: `koemoji-go-linux-1.0.0.tar.gz`
```
📁 koemoji-go-linux-1.0.0.tar.gz
├── koemoji-go         # Executable
├── config.json        # Configuration file
└── README.md          # Documentation
```

2. **Extract the downloaded file**

3. **FasterWhisper will be automatically installed on first run**

### Manual Installation (if needed)
```bash
pip install faster-whisper whisper-ctranslate2
```

## 3. First Run

### Basic Execution

**Windows:**
```cmd
koemoji-go.exe
```

**macOS:**
```bash
# Intel Mac
./koemoji-go-darwin-amd64

# Apple Silicon
./koemoji-go-darwin-arm64
```

**Linux:**
```bash
./koemoji-go
```

### Add to PATH (Optional)

To run from anywhere, add to PATH:

**macOS/Linux:**
```bash
# Add binary to PATH
sudo cp koemoji-go-darwin-arm64 /usr/local/bin/koemoji-go  # Apple Silicon
sudo cp koemoji-go-darwin-amd64 /usr/local/bin/koemoji-go  # Intel Mac
sudo cp koemoji-go /usr/local/bin/koemoji-go               # Linux
sudo chmod +x /usr/local/bin/koemoji-go

# Alias setup (optional)
echo 'alias koe="koemoji-go"' >> ~/.zshrc  # for zsh
echo 'alias koe="koemoji-go"' >> ~/.bashrc # for bash
source ~/.zshrc  # Apply settings
```

**Windows:**
```cmd
# Set PATH manually through environment variables
# Or open command prompt in the executable folder
```

Now you can run with `koemoji-go` or `koe` command.

The following directories will be automatically created on first run:
- `input/` - Place files to process here
- `output/` - Transcription results output here
- `archive/` - Processed files stored here
- `koemoji.log` - Log file

## 4. Basic Usage

### Step 1: Prepare Audio Files
- Place supported files in the `input/` folder
- Multiple files can be processed simultaneously

### Step 2: Monitor Processing Status
- Processing status is displayed in real-time on the UI screen
- Completed files are automatically moved to `archive/`
- Transcription results are saved in the `output/` folder

### Processing Flow
```
[input/audio file] → [transcription] → [output/text file]
                                    ↓
                            [archive/processed file]
```

- Automatically scans the `input/` folder every 10 minutes
- Starts sequential processing when new files are found
- After processing, original files are moved to `archive/`

## 5. Interactive Controls

During execution, you can use the following keys:
- `c` - Show configuration
- `l` - Show all logs
- `s` - Manual scan
- `q` - Quit
- `Enter` - Refresh screen
- `Ctrl+C` - Force quit

## 6. Supported File Formats

- **Audio**: MP3, WAV, M4A, FLAC, OGG, AAC
- **Video**: MP4, MOV, AVI

## 7. Configuration Customization

You can customize behavior with `config.json`:

```json
{
  "whisper_model": "large-v3",
  "language": "ja",
  "scan_interval_minutes": 10,
  "max_cpu_percent": 95,
  "compute_type": "int8",
  "use_colors": true,
  "ui_mode": "enhanced"
}
```

### Configuration Options

- `whisper_model`: Whisper model (tiny, base, small, medium, large, large-v2, large-v3)
- `language`: Language code (ja, en, etc.)
- `scan_interval_minutes`: Folder monitoring interval (minutes)
- `max_cpu_percent`: CPU usage limit (currently unused)
- `compute_type`: Quantization type (int8, float16, etc.)
- `use_colors`: Enable/disable color display
- `ui_mode`: UI display mode (enhanced/simple)

### Whisper Model Selection

| Model | Size | Speed | Accuracy | Recommended Use |
|-------|------|-------|----------|-----------------|
| tiny | Smallest | Fastest | Low | Testing |
| base | Small | Fast | Medium | Simple audio |
| small | Medium | Normal | Medium | Balanced |
| medium | Large | Slow | High | Quality focused |
| large | Largest | Slowest | Highest | High accuracy (old) |
| large-v2 | Largest | Slowest | Highest | Multilingual improved |
| large-v3 | Largest | Slowest | Highest | **Japanese recommended** |

**Recommended**: For Japanese transcription, `large-v3` is optimal (significantly reduced hallucinations).

## 8. Command Line Options

```bash
./koemoji-go --config custom.json  # Custom config file
./koemoji-go --debug               # Debug mode
./koemoji-go --version             # Show version
./koemoji-go --help                # Show help
```

## 9. Troubleshooting

### Common Issues

**Q: "Python not found" error**
- Python is not installed
- Install Python following the "1. System Requirements" section above
- Restart terminal/command prompt after installation

**Q: Python is installed but old version**
```bash
# Check version
python --version

# If below Python 3.8, install newer version
```

**Q: FasterWhisper installation fails**
```bash
# Install manually
pip install faster-whisper whisper-ctranslate2

# If pip is old
pip install --upgrade pip
pip install faster-whisper whisper-ctranslate2

# If permission error
pip install --user faster-whisper whisper-ctranslate2
```

**Q: "whisper-ctranslate2 not found" error**
- Python PATH may not be set correctly
- pip installed packages PATH may not be set
- Check the following:
```bash
# Check if package is installed
pip show whisper-ctranslate2

# Check PATH
which whisper-ctranslate2
# or
where whisper-ctranslate2  # Windows
```

**Q: Processing is slow**
- Change model to `small` or `medium` in `config.json`
- If memory is insufficient, set `compute_type` to `int8`

**Q: Audio files not recognized**
- Supported formats: MP3, WAV, M4A, FLAC, OGG, AAC, MP4, MOV, AVI
- Check if filename contains special characters

**Q: Transcription results are incorrect**
- Recommend using `large-v3` model
- Check audio quality (noise, volume, etc.)

### Check Logs

If problems occur, check `koemoji.log`:
```bash
# Check log file
cat koemoji.log

# Check latest logs only
tail -f koemoji.log
```

---

## Developer Information

### Build Instructions

#### Simple Build (with icons - recommended)
```bash
# Build for all platforms with icons
./build.sh

# Build for specific platform only
./build.sh windows   # Windows only
./build.sh macos     # macOS only
./build.sh linux    # Linux only

# Clean build artifacts
./build.sh clean
```

**Generated files:**
- Windows: `koemoji-go-windows-1.0.0.zip` (executable with icon)
- macOS: `koemoji-go-macos-1.0.0.tar.gz` (Intel/Apple Silicon support)
- Linux: `koemoji-go-linux-1.0.0.tar.gz` (64bit version)

#### Simple Development Build
```bash
go build -o koemoji-go main.go
```

#### Manual Build (without icons)
```bash
# Windows 64bit
GOOS=windows GOARCH=amd64 go build -o koemoji-go-windows-amd64.exe main.go

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o koemoji-go-darwin-amd64 main.go

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o koemoji-go-darwin-arm64 main.go

# Linux 64bit
GOOS=linux GOARCH=amd64 go build -o koemoji-go-linux-amd64 main.go
```

### Development Environment Setup

#### Required Tools
- Go 1.21+
- Python 3.8+ + FasterWhisper (for testing)
- Git

#### Setup Steps
```bash
git clone https://github.com/[username]/koemoji-go.git
cd koemoji-go
go mod tidy
go build -o koemoji-go main.go
```

### Technical Specifications

#### Architecture
- **Language**: Go 1.21
- **Dependencies**: Standard library only
- **External Integration**: FasterWhisper (whisper-ctranslate2)
- **Processing Method**: Sequential processing (one file at a time)

#### Core Features
- Automatic directory monitoring
- Real-time UI display
- Log management
- Configuration file management
- Cross-platform support

## License

**Personal Use**: Free to use  
**Commercial Use**: Contact required

See [LICENSE](LICENSE) file for details.

## Author

KoeMoji-Go Development Team

Contact: dev@habitengineer.com