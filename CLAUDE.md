# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

KoeMoji-Go is a Go-based audio/video transcription tool that uses FasterWhisper for high-accuracy speech recognition. It provides a real-time terminal UI for monitoring file processing and supports cross-platform distribution (Windows/macOS).

## Key Commands

### Development & Building
```bash
# Development build
go build -o koemoji-go ./cmd/koemoji-go

# Release build with distribution packages
cd build && ./build.sh

# Test basic functionality
./koemoji-go --version
./koemoji-go --help
./koemoji-go --configure

# Run in GUI mode
./koemoji-go --gui
```

### Dependencies
- Go 1.21+ required
- Python 3.8+ with FasterWhisper (auto-installed on first run)
- Manual installation: `pip install faster-whisper whisper-ctranslate2`

## Architecture

### Package Structure
The project follows Go's standard internal package layout:

- **`cmd/koemoji-go/`** - Main application entry point and CLI handling
- **`internal/config/`** - Configuration management (JSON) and interactive settings editor
- **`internal/logger/`** - Structured logging with circular buffer (max 12 entries)  
- **`internal/processor/`** - File monitoring, queue management, and processing orchestration
- **`internal/ui/`** - Real-time terminal UI with multilingual support
- **`internal/whisper/`** - FasterWhisper integration and audio transcription
- **`internal/llm/`** - LLM integration for AI summarization (v1.2.0+)
- **`internal/gui/`** - Fyne-based GUI implementation (v1.3.0+)

### Core Processing Flow
1. **File Monitoring**: Periodic directory scanning (`input/`) with configurable intervals
2. **Queue Management**: Sequential processing to ensure stability (one file at a time)
3. **Transcription**: Shell command execution to `whisper-ctranslate2` with progress monitoring
4. **AI Summarization** (v1.2.0+): Optional LLM-based summary generation via OpenAI API
5. **File Management**: Automatic archiving of processed files to `archive/`
6. **Real-time UI**: Live status updates and interactive controls

### Multilingual Support
The application supports English and Japanese UI languages. Messages are centralized in `internal/config/config.go` with `Messages` struct and language-specific instances (`messagesEN`, `messagesJA`).

## GUI System (v1.3.0+)

### GUI Design Principles
- **Functional Compatibility**: Complete feature parity with existing TUI
- **Simple Design**: Minimal implementation with intuitive interface
- **Code Reuse**: Existing processor/logger/whisper/config packages unchanged
- **Immediate Exit**: Window close/quit button → instant application termination

### GUI Architecture
```bash
# GUI mode launch
./koemoji-go --gui

# Traditional TUI launch (default)
./koemoji-go
```

### Package Structure (`internal/gui/`)
- **`app.go`** - GUI application core
- **`window.go`** - Window layout (BorderLayout: status/logs/buttons)
- **`components.go`** - UI components (status panel, log viewer, button panel)
- **`icons/`** - Application icon resources

### UI Specifications
- **Window**: 800x700 pixels, centered, resizable
- **Update Method**: 5-second periodic refresh (not event-driven)
- **Log Display**: 12-entry circular buffer with scrollable RichText widget
- **Settings Dialog**: Tabbed interface (Basic/Directories/LLM)

### Technical Stack
- **Framework**: Fyne v2 for cross-platform GUI
- **Build Impact**: +30-40MB binary size, +30-40MB memory usage
- **Dependencies**: fyne.io/fyne/v2 packages only

## Configuration System

### Configuration File Structure
```json
{
  "whisper_model": "large-v3",
  "language": "ja", 
  "ui_language": "ja",
  "scan_interval_minutes": 1,
  "max_cpu_percent": 95,
  "compute_type": "int8",
  "use_colors": true,
  "output_format": "txt",
  "input_dir": "./input",
  "output_dir": "./output", 
  "archive_dir": "./archive"
}
```

### Interactive Configuration
The application provides a built-in configuration editor accessible via:
- CLI flag: `--configure`
- Runtime command: `c` key

## Build System

### Cross-Platform Builds
The `build/build.sh` script handles:
- Windows builds with embedded icons using `goversioninfo`
- macOS builds for both Intel and Apple Silicon
- Automatic packaging with config files and documentation

### Windows-Specific Considerations
- Icon embedding requires `resource_windows_amd64.syso` in `cmd/koemoji-go/`
- Color support forced on Windows 10+ for optimal UI experience
- Notepad used for log file viewing (universal Windows compatibility)

## External Dependencies

### FasterWhisper Integration
- Uses `whisper-ctranslate2` command-line tool
- Automatic dependency installation on first run
- Supports various Whisper models (tiny to large-v3)
- Progress monitoring via goroutines and command output parsing

### Audio File Support
Supported formats: MP3, WAV, M4A, FLAC, OGG, AAC, MP4, MOV, AVI

### LLM Integration (v1.2.0+)
- **Package**: `internal/llm/` - OpenAI API integration
- **Features**: Automatic summarization post-transcription
- **API Support**: OpenAI GPT models (gpt-4o, gpt-4-turbo, gpt-3.5-turbo)
- **Error Handling**: Retry logic, rate limit handling, API validation

## Development Notes

### Error Handling
- Comprehensive error logging with UI feedback
- Graceful degradation for missing dependencies
- Path validation for security (input directory restrictions)

### Concurrency
- Goroutine-based file monitoring and processing
- Thread-safe logging buffer management
- Sequential processing queue to prevent resource conflicts

### Testing Approach
- Manual testing with audio files in `input/` directory
- Version and help command verification
- Configuration system testing via `--configure` flag