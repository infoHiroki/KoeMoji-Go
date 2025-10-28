// +build darwin

package recorder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DualRecorder records system audio and microphone simultaneously on macOS
// Architecture: 2-stream approach (separate files for system audio and microphone)
// - System audio: ScreenCaptureKit (Swift CLI) → 48kHz, Float32, Stereo
// - Microphone: PortAudio → 44.1kHz, Int16, Mono (configurable)
// - Output: Mixed to single WAV file (48kHz, Int16, Stereo)
//
// IMPORTANT: Headphones/earphones are recommended.
// In speaker environments, the microphone may pick up system audio from speakers,
// resulting in doubled system audio in the recording.
// This is a physical limitation (acoustic coupling) and cannot be fully resolved by software.
type DualRecorder struct {
	// System audio (ScreenCaptureKit via Swift CLI)
	systemRecorder *SystemAudioRecorder
	systemEnabled  bool

	// Microphone (PortAudio)
	micRecorder *Recorder
	micEnabled  bool

	// State
	recording   bool
	mutex       sync.Mutex
	startTime   time.Time
	baseDir     string // Output directory for recording files
	sessionID   string // Unique session identifier (timestamp)
	wg          sync.WaitGroup
	stopChan    chan struct{}
	recordError error

	// Recording limits
	maxDuration time.Duration
	maxFileSize int64

	// Output paths
	systemOutputPath string
	micOutputPath    string
	mixedOutputPath  string
}

// NewDualRecorder creates a new dual recorder with default settings
func NewDualRecorder() (*DualRecorder, error) {
	// Create system audio recorder
	systemRecorder, err := NewSystemAudioRecorder()
	if err != nil {
		return nil, fmt.Errorf("failed to create system audio recorder: %w", err)
	}

	// Create microphone recorder (use default device)
	micRecorder, err := NewRecorder()
	if err != nil {
		return nil, fmt.Errorf("failed to create microphone recorder: %w", err)
	}

	return &DualRecorder{
		systemRecorder: systemRecorder,
		systemEnabled:  true,
		micRecorder:    micRecorder,
		micEnabled:     true,
		recording:      false,
		maxDuration:    0, // Unlimited
		maxFileSize:    0, // Unlimited
	}, nil
}

// NewDualRecorderWithDevices creates a dual recorder with specific microphone device
func NewDualRecorderWithDevices(micDeviceName string) (*DualRecorder, error) {
	// Create system audio recorder
	systemRecorder, err := NewSystemAudioRecorder()
	if err != nil {
		return nil, fmt.Errorf("failed to create system audio recorder: %w", err)
	}

	// Create microphone recorder with specified device
	var micRecorder *Recorder
	if micDeviceName != "" {
		micRecorder, err = NewRecorderWithDeviceName(micDeviceName)
		if err != nil {
			return nil, fmt.Errorf("failed to create microphone recorder: %w", err)
		}
	} else {
		micRecorder, err = NewRecorder()
		if err != nil {
			return nil, fmt.Errorf("failed to create microphone recorder: %w", err)
		}
	}

	return &DualRecorder{
		systemRecorder: systemRecorder,
		systemEnabled:  true,
		micRecorder:    micRecorder,
		micEnabled:     true,
		recording:      false,
		maxDuration:    0, // Unlimited
		maxFileSize:    0, // Unlimited
	}, nil
}

// SetLimits configures recording limits
func (dr *DualRecorder) SetLimits(maxDuration time.Duration, maxFileSize int64) {
	dr.mutex.Lock()
	defer dr.mutex.Unlock()
	dr.maxDuration = maxDuration
	dr.maxFileSize = maxFileSize

	// Apply limits to sub-recorders
	if dr.micRecorder != nil {
		dr.micRecorder.SetLimits(maxDuration, maxFileSize)
	}
}

// Start begins dual recording
func (dr *DualRecorder) Start() error {
	dr.mutex.Lock()
	defer dr.mutex.Unlock()

	if dr.recording {
		return fmt.Errorf("recording already in progress")
	}

	// Create temporary directory for this session
	dr.sessionID = fmt.Sprintf("%d", time.Now().Unix())
	dr.baseDir = os.TempDir()
	dr.stopChan = make(chan struct{})
	dr.recordError = nil

	// Generate output paths
	dr.systemOutputPath = filepath.Join(dr.baseDir, fmt.Sprintf("system-%s.wav", dr.sessionID))
	dr.micOutputPath = filepath.Join(dr.baseDir, fmt.Sprintf("mic-%s.wav", dr.sessionID))

	dr.startTime = time.Now()
	dr.recording = true

	// Start system audio recording
	if dr.systemEnabled {
		dr.wg.Add(1)
		go func() {
			defer dr.wg.Done()
			if err := dr.systemRecorder.Start(dr.systemOutputPath); err != nil {
				dr.mutex.Lock()
				dr.recordError = fmt.Errorf("system audio: %w", err)
				dr.mutex.Unlock()
				return
			}
		}()
	}

	// Start microphone recording
	if dr.micEnabled {
		dr.wg.Add(1)
		go func() {
			defer dr.wg.Done()
			if err := dr.micRecorder.Start(); err != nil {
				dr.mutex.Lock()
				dr.recordError = fmt.Errorf("microphone: %w", err)
				dr.mutex.Unlock()
				return
			}
		}()
	}

	// Start limit monitor if limits are set
	if dr.maxDuration > 0 {
		dr.wg.Add(1)
		go dr.monitorLimits()
	}

	return nil
}

// Stop ends dual recording and saves files
func (dr *DualRecorder) Stop() error {
	dr.mutex.Lock()
	if !dr.recording {
		dr.mutex.Unlock()
		return fmt.Errorf("no recording in progress")
	}
	dr.recording = false
	dr.mutex.Unlock()

	// Signal stop to all goroutines
	close(dr.stopChan)

	// Stop system audio
	var systemErr error
	if dr.systemEnabled {
		if _, err := dr.systemRecorder.Stop(); err != nil {
			systemErr = fmt.Errorf("stop system audio failed: %w", err)
		}
	}

	// Stop microphone
	var micErr error
	if dr.micEnabled {
		if err := dr.micRecorder.Stop(); err != nil {
			micErr = fmt.Errorf("stop microphone failed: %w", err)
		}
	}

	// Wait for all goroutines
	dr.wg.Wait()

	// Check for errors
	if dr.recordError != nil {
		return dr.recordError
	}
	if systemErr != nil {
		return systemErr
	}
	if micErr != nil {
		return micErr
	}

	return nil
}

// SaveToFile saves the recorded audio to a file
// By default, mixes system audio and microphone into a single file (FFmpeg-free)
// For separate files, use SaveSeparateFiles() instead
func (dr *DualRecorder) SaveToFile(filename string) error {
	dr.mutex.Lock()
	defer dr.mutex.Unlock()

	if dr.recording {
		return fmt.Errorf("cannot save while recording")
	}

	// If both system and mic are enabled, mix them
	if dr.systemEnabled && dr.micEnabled {
		// Save mic to temporary file
		micTempPath := filepath.Join(dr.baseDir, fmt.Sprintf("mic-temp-%s.wav", dr.sessionID))
		if err := dr.micRecorder.SaveToFile(micTempPath); err != nil {
			return fmt.Errorf("failed to save microphone audio: %w", err)
		}
		defer os.Remove(micTempPath)

		// Mix system and mic audio (FFmpeg-free)
		// System: 70% volume, Mic: 100% volume
		if err := MixAudioFiles(dr.systemOutputPath, micTempPath, filename, 0.7, 1.0); err != nil {
			return fmt.Errorf("failed to mix audio files: %w", err)
		}

		// Clean up temporary system audio file
		os.Remove(dr.systemOutputPath)

		return nil
	}

	// If only one stream is enabled, save it directly
	if dr.micEnabled {
		if err := dr.micRecorder.SaveToFile(filename); err != nil {
			return fmt.Errorf("failed to save microphone audio: %w", err)
		}
	} else if dr.systemEnabled {
		if err := copyFile(dr.systemOutputPath, filename); err != nil {
			return fmt.Errorf("failed to save system audio: %w", err)
		}
		os.Remove(dr.systemOutputPath)
	}

	return nil
}

// SaveToFileWithNormalization saves recording with optional audio normalization
// Normalization is applied only to microphone audio before mixing
func (dr *DualRecorder) SaveToFileWithNormalization(filename string, enableNormalization bool) error {
	dr.mutex.Lock()
	defer dr.mutex.Unlock()

	if dr.recording {
		return fmt.Errorf("cannot save while recording")
	}

	// If both system and mic are enabled, mix them
	if dr.systemEnabled && dr.micEnabled {
		// Save mic to temporary file with normalization
		micTempPath := filepath.Join(dr.baseDir, fmt.Sprintf("mic-temp-%s.wav", dr.sessionID))
		if err := dr.micRecorder.SaveToFileWithNormalization(micTempPath, enableNormalization); err != nil {
			return fmt.Errorf("failed to save microphone audio: %w", err)
		}
		defer os.Remove(micTempPath)

		// Mix system and mic audio (FFmpeg-free)
		// System: 70% volume, Mic: 100% volume
		if err := MixAudioFiles(dr.systemOutputPath, micTempPath, filename, 0.7, 1.0); err != nil {
			return fmt.Errorf("failed to mix audio files: %w", err)
		}

		// Clean up temporary system audio file
		os.Remove(dr.systemOutputPath)

		return nil
	}

	// If only one stream is enabled, save it directly
	if dr.micEnabled {
		if err := dr.micRecorder.SaveToFileWithNormalization(filename, enableNormalization); err != nil {
			return fmt.Errorf("failed to save microphone audio: %w", err)
		}
	} else if dr.systemEnabled {
		if err := copyFile(dr.systemOutputPath, filename); err != nil {
			return fmt.Errorf("failed to save system audio: %w", err)
		}
		os.Remove(dr.systemOutputPath)
	}

	return nil
}

// SaveSeparateFiles saves system audio and microphone as separate files
// This is useful for advanced use cases (e.g., speaker diarization)
// System audio: filename-system.wav
// Microphone: filename.wav
func (dr *DualRecorder) SaveSeparateFiles(filename string) error {
	dr.mutex.Lock()
	defer dr.mutex.Unlock()

	if dr.recording {
		return fmt.Errorf("cannot save while recording")
	}

	// Save microphone audio to the requested filename
	if dr.micEnabled {
		if err := dr.micRecorder.SaveToFile(filename); err != nil {
			return fmt.Errorf("failed to save microphone audio: %w", err)
		}
	}

	// Save system audio with "-system" suffix
	if dr.systemEnabled {
		systemFilename := generateSystemFilename(filename)
		if err := copyFile(dr.systemOutputPath, systemFilename); err != nil {
			return fmt.Errorf("failed to save system audio: %w", err)
		}
		// Clean up temporary file
		os.Remove(dr.systemOutputPath)
	}

	return nil
}

// MixToFile mixes system audio and microphone into a single file using FFmpeg
// DEPRECATED: Use SaveToFile() instead (FFmpeg-free)
// This method is kept for backward compatibility but requires FFmpeg
func (dr *DualRecorder) MixToFile(filename string) error {
	dr.mutex.Lock()
	defer dr.mutex.Unlock()

	if dr.recording {
		return fmt.Errorf("cannot mix while recording")
	}

	// Check if FFmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found (required for mixing): %w", err)
	}

	// Ensure both files exist
	if _, err := os.Stat(dr.systemOutputPath); err != nil {
		return fmt.Errorf("system audio file not found: %w", err)
	}

	micTempPath := filepath.Join(dr.baseDir, fmt.Sprintf("mic-temp-%s.wav", dr.sessionID))
	if err := dr.micRecorder.SaveToFile(micTempPath); err != nil {
		return fmt.Errorf("failed to save mic audio: %w", err)
	}
	defer os.Remove(micTempPath)

	if _, err := os.Stat(micTempPath); err != nil {
		return fmt.Errorf("microphone audio file not found: %w", err)
	}

	// Mix with FFmpeg
	// System: 70% volume, Mic: 100% volume (convert mono to stereo)
	cmd := exec.Command("ffmpeg",
		"-i", dr.systemOutputPath,
		"-i", micTempPath,
		"-filter_complex",
		"[0:a]volume=0.7[sys];[1:a]volume=1.0,pan=stereo|c0=c0|c1=c0[mic];[sys][mic]amix=inputs=2:duration=longest",
		"-ar", "48000",
		"-ac", "2",
		"-y", // Overwrite output file
		filename,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg mixing failed: %w, output: %s", err, string(output))
	}

	dr.mixedOutputPath = filename
	return nil
}

// Close releases all resources
func (dr *DualRecorder) Close() error {
	if dr.recording {
		if err := dr.Stop(); err != nil {
			return err
		}
	}

	// Clean up temporary files
	if dr.systemOutputPath != "" {
		os.Remove(dr.systemOutputPath)
	}

	// Close sub-recorders
	if dr.systemRecorder != nil {
		dr.systemRecorder.Close()
	}
	if dr.micRecorder != nil {
		dr.micRecorder.Close()
	}

	return nil
}

// IsRecording returns whether recording is in progress
func (dr *DualRecorder) IsRecording() bool {
	dr.mutex.Lock()
	defer dr.mutex.Unlock()
	return dr.recording
}

// GetDuration returns the duration of recorded audio in seconds
func (dr *DualRecorder) GetDuration() float64 {
	dr.mutex.Lock()
	defer dr.mutex.Unlock()

	// Use microphone duration (should be similar to system audio)
	if dr.micRecorder != nil {
		return dr.micRecorder.GetDuration()
	}
	return 0
}

// GetElapsedTime returns elapsed recording time
func (dr *DualRecorder) GetElapsedTime() time.Duration {
	dr.mutex.Lock()
	defer dr.mutex.Unlock()
	if dr.recording {
		return time.Since(dr.startTime)
	}
	return 0
}

// monitorLimits monitors recording limits and stops automatically
func (dr *DualRecorder) monitorLimits() {
	defer dr.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-dr.stopChan:
			return
		case <-ticker.C:
			if dr.exceedsLimits() {
				dr.Stop()
				return
			}
		}
	}
}

// exceedsLimits checks if recording limits are exceeded
func (dr *DualRecorder) exceedsLimits() bool {
	dr.mutex.Lock()
	defer dr.mutex.Unlock()

	// Check duration limit
	if dr.maxDuration > 0 && time.Since(dr.startTime) >= dr.maxDuration {
		return true
	}

	// Check file size limit (approximate)
	if dr.maxFileSize > 0 {
		duration := time.Since(dr.startTime).Seconds()
		// Estimate: 48kHz * 2ch * 2bytes = ~192KB/sec for system audio
		//          + 44.1kHz * 1ch * 2bytes = ~88KB/sec for mic
		estimatedSize := int64(duration * (192000 + 88000))
		if estimatedSize >= dr.maxFileSize {
			return true
		}
	}

	return false
}

// Helper functions

// generateSystemFilename generates a filename for system audio
// Example: "output.wav" → "output-system.wav"
func generateSystemFilename(filename string) string {
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	return base + "-system" + ext
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
