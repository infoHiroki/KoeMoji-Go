# KoeMoji-Go

Automatic Audio/Video Transcription Tool

[日本語版 README](README.md) | **English README**

## Overview

KoeMoji-Go is an application that automatically transcribes audio and video files.
It's a Go port of the Python-based KoeMojiAuto-cli, providing single binary distribution and stable sequential processing.

### Features

- **Single Binary**: Works with just one executable file
- **Sequential Processing**: Stable processing one file at a time
- **FasterWhisper Integration**: High-accuracy speech recognition
- **AI Summary**: Automatic summary generation using OpenAI API
- **Recording Feature**: Built-in microphone recording
- **GUI/TUI Support**: Both graphical and terminal interfaces
- **Auto Monitoring**: Automatically monitors folders and processes files

## ⚡ Quick Start

### Prerequisites

**Python 3.8+** is required.
```bash
python --version  # Check version is 3.8+
```

If Python is not installed, download from [Python Official Downloads](https://www.python.org/downloads/).

### Installation

Download and extract the version for your OS from [GitHub Releases](https://github.com/infoHiroki/KoeMoji-Go/releases).

#### 🪟 Windows

1. **Download**: `koemoji-go-windows-1.5.0.zip`
2. **Extracted contents**:
   ```
   📁 koemoji-go-windows-1.5.0
   ├── koemoji-go.exe          # Executable with icon
   ├── libportaudio.dll        # Audio recording library
   ├── libgcc_s_seh-1.dll      # GCC runtime
   ├── libwinpthread-1.dll     # Thread support
   ├── config.json             # Configuration file
   └── README.md               # Documentation
   ```
3. **Run**:
   ```cmd
   koemoji-go.exe
   ```

#### 🍎 macOS

1. **Download**:
   - **Intel Mac**: `koemoji-go-macos-intel-1.5.0.tar.gz`
   - **Apple Silicon (M1/M2)**: `koemoji-go-macos-arm64-1.5.0.tar.gz`

2. **Extracted contents**:
   ```
   📁 koemoji-go-macos-*-1.5.0
   ├── koemoji-go         # Executable file
   ├── config.json        # Configuration file
   └── README.md          # Documentation
   ```

3. **Run**:
   ```bash
   ./koemoji-go
   ```

> **First run**: FasterWhisper will be automatically installed (takes a few minutes)

### Basic Usage

#### 1. Process Audio Files
1. Place audio files (MP3, WAV, etc.) in the `input/` folder
2. Processing will start automatically
3. Results are saved in the `output/` folder
4. Processed files are moved to `archive/`

#### 2. UI Mode Selection
- **GUI Mode (Default)**: Click buttons to operate
  ```bash
  ./koemoji-go
  ```
- **TUI Mode**: Keyboard controls
  ```bash
  ./koemoji-go --tui
  ```

#### 3. Main Controls (TUI Mode)
- `c` - Configure settings
- `l` - Display logs
- `s` - Manual scan
- `r` - Start/stop recording
- `q` - Quit

#### 4. Supported File Formats
- **Audio**: MP3, WAV, M4A, FLAC, OGG, AAC
- **Video**: MP4, MOV, AVI

#### 5. AI Summary Feature (Optional)
1. Get API key from [OpenAI Platform](https://platform.openai.com/)
2. Enter API key in settings (`c`)
3. Summary will be automatically generated after transcription

## 📚 Additional Information

- **[🔧 Troubleshooting](TROUBLESHOOTING.md)** - Problem solving and FAQ
- **[📖 Developer Documentation](docs/)** - Build instructions, architecture, technical specifications

## License

**Personal Use**: Free to use  
**Commercial Use**: Contact required

See [LICENSE](LICENSE) file for details.

## Author

KoeMoji-Go Development Team  
Contact: koemoji2024@gmail.com