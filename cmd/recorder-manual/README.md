# Recorder Manual Test Tool

This is an interactive command-line tool for manually testing the audio recording functionality of KoeMoji-Go.

## Usage

```bash
go run main.go
```

## Features

1. **Device Listing**: Lists all available audio input devices
2. **Device Selection**: Allows you to select a specific device or use the default
3. **Interactive Recording**: Start/stop recording with Enter key
4. **File Output**: Saves recordings as WAV files in the current directory

## Controls

- **Enter**: Start/Stop recording
- **q**: Quit the application

## Output

The tool creates numbered WAV files (`test_recording_1.wav`, `test_recording_2.wav`, etc.) in the current directory.

## Requirements

- Go 1.21+
- PortAudio library installed
- Audio input device available

## Installation Note

This tool was moved from `test/recorder_manual.go` to resolve package conflicts in the test directory. It maintains the same functionality as before.