// +build darwin

package recorder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// SystemAudioRecorder records macOS system audio using ScreenCaptureKit
// This recorder wraps the Swift CLI tool (cmd/audio-capture/audio-capture)
type SystemAudioRecorder struct {
	binaryPath string
	cmd        *exec.Cmd
	outputPath string
	recording  bool
	startTime  time.Time
	mutex      sync.Mutex
}

// NewSystemAudioRecorder creates a new system audio recorder
func NewSystemAudioRecorder() (*SystemAudioRecorder, error) {
	// Find the audio-capture binary
	binaryPath, err := findAudioCaptureBinary()
	if err != nil {
		return nil, fmt.Errorf("audio-capture binary not found: %w", err)
	}

	return &SystemAudioRecorder{
		binaryPath: binaryPath,
		recording:  false,
	}, nil
}

// Start begins system audio recording in the background
// The recording continues until Stop() is called
// Note: Swift CLI always outputs CAF format, conversion to WAV happens in Stop()
func (r *SystemAudioRecorder) Start(outputPath string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.recording {
		return fmt.Errorf("recording already in progress")
	}

	// Ensure output path is absolute
	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("failed to resolve output path: %w", err)
	}

	// Swift CLI outputs CAF format, so we request a .caf file
	cafPath := absPath
	if filepath.Ext(absPath) == ".wav" {
		cafPath = absPath[:len(absPath)-4] + ".caf"
	}

	r.outputPath = absPath // Store the desired final path (might be .wav)
	r.startTime = time.Now()

	// Create command: audio-capture -o output.caf -d 0 (0 = infinite duration)
	r.cmd = exec.Command(r.binaryPath, "-o", cafPath, "-d", "0")

	// Redirect stderr for debugging (optional)
	r.cmd.Stderr = os.Stderr

	// Start the process in the background
	if err := r.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start audio capture: %w", err)
	}

	r.recording = true
	return nil
}

// Stop ends system audio recording and returns the output file path
func (r *SystemAudioRecorder) Stop() (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.recording {
		return "", fmt.Errorf("no recording in progress")
	}

	// Send SIGTERM to gracefully stop the Swift CLI
	if r.cmd != nil && r.cmd.Process != nil {
		if err := r.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			return "", fmt.Errorf("failed to send SIGTERM: %w", err)
		}

		// Wait for the process to finish (with timeout)
		done := make(chan error, 1)
		go func() {
			done <- r.cmd.Wait()
		}()

		select {
		case err := <-done:
			if err != nil {
				// Ignore "signal: terminated" error (expected)
				if exitErr, ok := err.(*exec.ExitError); ok {
					if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
						if status.Signal() == syscall.SIGTERM {
							// This is expected, not an error
							err = nil
						}
					}
				}
				if err != nil {
					return "", fmt.Errorf("process wait failed: %w", err)
				}
			}
		case <-time.After(5 * time.Second):
			// Force kill if graceful shutdown takes too long
			r.cmd.Process.Kill()
			return "", fmt.Errorf("recording process did not stop gracefully, killed")
		}
	}

	r.recording = false

	// Swift CLI outputs CAF format, convert to WAV if needed
	cafPath := r.outputPath
	if filepath.Ext(cafPath) == ".caf" {
		cafPath = r.outputPath[:len(r.outputPath)-4] + ".caf"
	} else {
		// Assume it's a CAF file with wrong extension
		cafPath = r.outputPath
	}

	wavPath := r.outputPath
	if filepath.Ext(wavPath) != ".wav" {
		wavPath = r.outputPath[:len(r.outputPath)-len(filepath.Ext(r.outputPath))] + ".wav"
	}

	// Swift CLI writes CAF format internally
	// Check if CAF file exists (Swift writes with .caf extension)
	actualCafPath := r.outputPath[:len(r.outputPath)-len(filepath.Ext(r.outputPath))] + ".caf"
	if _, err := os.Stat(actualCafPath); err == nil {
		// CAF file exists, convert to WAV
		if err := convertCAFtoWAV(actualCafPath, wavPath); err != nil {
			return "", fmt.Errorf("failed to convert CAF to WAV: %w", err)
		}
		// Remove CAF file after successful conversion
		os.Remove(actualCafPath)
		r.outputPath = wavPath
	} else if _, err := os.Stat(r.outputPath); err == nil {
		// File exists with requested extension, assume it's ready
		// (this shouldn't happen with current Swift implementation)
	} else {
		return "", fmt.Errorf("output file not created: %w", err)
	}

	return r.outputPath, nil
}

// convertCAFtoWAV converts a CAF file to WAV using afconvert
func convertCAFtoWAV(cafPath, wavPath string) error {
	cmd := exec.Command("afconvert",
		"-f", "WAVE",   // WAV format
		"-d", "LEF32",  // Little-endian float 32
		cafPath,
		wavPath,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("afconvert failed: %w, output: %s", err, string(output))
	}

	return nil
}

// IsRecording returns whether recording is in progress
func (r *SystemAudioRecorder) IsRecording() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.recording
}

// GetElapsedTime returns the elapsed recording time
func (r *SystemAudioRecorder) GetElapsedTime() time.Duration {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.recording {
		return time.Since(r.startTime)
	}
	return 0
}

// Close cleans up resources
func (r *SystemAudioRecorder) Close() error {
	if r.IsRecording() {
		_, err := r.Stop()
		return err
	}
	return nil
}

// findAudioCaptureBinary searches for the audio-capture binary
// Search order:
// 1. ./cmd/audio-capture/audio-capture (development)
// 2. Same directory as the executable (production)
// 3. /tmp/audio-capture (go:embed extraction, future)
func findAudioCaptureBinary() (string, error) {
	candidates := []string{
		// Development: relative to project root
		"./cmd/audio-capture/audio-capture",
		"cmd/audio-capture/audio-capture",
	}

	// Production: same directory as the executable
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates, filepath.Join(exeDir, "audio-capture"))
	}

	// Check each candidate
	for _, path := range candidates {
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}

		if info, err := os.Stat(absPath); err == nil {
			// Check if it's executable
			if info.Mode()&0111 != 0 {
				return absPath, nil
			}
		}
	}

	return "", fmt.Errorf("audio-capture binary not found in: %v", candidates)
}
