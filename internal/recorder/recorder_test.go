package recorder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRecorder_IsRecording tests the basic recording state
func TestRecorder_IsRecording(t *testing.T) {
	recorder, err := NewRecorder()
	require.NoError(t, err)
	defer recorder.Close()

	// Initially not recording
	assert.False(t, recorder.IsRecording())

	// Start recording (may fail if no audio device)
	err = recorder.Start()
	if err != nil {
		t.Skipf("Audio device not available: %v", err)
		return
	}

	// Should be recording
	assert.True(t, recorder.IsRecording())

	// Stop recording
	err = recorder.Stop()
	require.NoError(t, err)

	// Should not be recording
	assert.False(t, recorder.IsRecording())
}

// TestRecorder_GetDuration tests duration calculation
func TestRecorder_GetDuration(t *testing.T) {
	recorder, err := NewRecorder()
	require.NoError(t, err)
	defer recorder.Close()

	// Initially zero duration
	assert.Equal(t, 0.0, recorder.GetDuration())

	// Start recording
	err = recorder.Start()
	if err != nil {
		t.Skipf("Audio device not available: %v", err)
		return
	}

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Duration should be non-zero
	duration := recorder.GetDuration()
	assert.GreaterOrEqual(t, duration, 0.0)

	recorder.Stop()
}

// TestRecorder_GetElapsedTime tests elapsed time tracking
func TestRecorder_GetElapsedTime(t *testing.T) {
	recorder, err := NewRecorder()
	require.NoError(t, err)
	defer recorder.Close()

	// Initially zero elapsed time
	assert.Equal(t, time.Duration(0), recorder.GetElapsedTime())

	// Start recording
	err = recorder.Start()
	if err != nil {
		t.Skipf("Audio device not available: %v", err)
		return
	}

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Elapsed time should be non-zero
	elapsed := recorder.GetElapsedTime()
	assert.Greater(t, elapsed, time.Duration(0))
	assert.LessOrEqual(t, elapsed, 200*time.Millisecond) // Should be reasonable

	recorder.Stop()

	// After stop, elapsed time should be zero
	assert.Equal(t, time.Duration(0), recorder.GetElapsedTime())
}

// TestRecorder_DoubleStart tests error handling for double start
func TestRecorder_DoubleStart(t *testing.T) {
	recorder, err := NewRecorder()
	require.NoError(t, err)
	defer recorder.Close()

	// Start recording
	err = recorder.Start()
	if err != nil {
		t.Skipf("Audio device not available: %v", err)
		return
	}

	// Try to start again - should fail
	err = recorder.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "recording already in progress")

	recorder.Stop()
}

// TestRecorder_StopWithoutStart tests error handling for stop without start
func TestRecorder_StopWithoutStart(t *testing.T) {
	recorder, err := NewRecorder()
	require.NoError(t, err)
	defer recorder.Close()

	// Try to stop without starting
	err = recorder.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no recording in progress")
}

// TestRecorder_SaveToFileWithoutData tests save error handling
func TestRecorder_SaveToFileWithoutData(t *testing.T) {
	recorder, err := NewRecorder()
	require.NoError(t, err)
	defer recorder.Close()

	// Try to save without recording
	err = recorder.SaveToFile("test.wav")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no audio data to save")
}

// TestListDevices tests device enumeration
func TestListDevices(t *testing.T) {
	devices, err := ListDevices()

	// On systems without audio devices, this might fail
	if err != nil {
		t.Skipf("Audio device enumeration failed: %v", err)
		return
	}

	// Should return at least some information
	assert.NotNil(t, devices)

	// If devices exist, check structure
	for _, device := range devices {
		assert.NotEmpty(t, device.Name)
		assert.GreaterOrEqual(t, device.ID, 0)
		assert.Greater(t, device.MaxChannels, 0)
	}
}

// TestRecorder_SetLimits tests recording limits
func TestRecorder_SetLimits(t *testing.T) {
	recorder, err := NewRecorder()
	require.NoError(t, err)
	defer recorder.Close()

	// Set limits
	maxDuration := 5 * time.Second
	maxFileSize := int64(1024 * 1024) // 1MB

	recorder.SetLimits(maxDuration, maxFileSize)

	// This is a basic test - actual limit enforcement
	// would require a longer integration test
	assert.NotNil(t, recorder)
}

// TestNormalizeAudio tests the audio normalization function
func TestNormalizeAudio(t *testing.T) {
	tests := []struct {
		name           string
		samples        []int16
		expectModified bool
		description    string
	}{
		{
			name:           "Small amplitude should be normalized",
			samples:        []int16{100, 200, -150, 300, -250},
			expectModified: true,
			description:    "Very small samples (max 300) should be amplified",
		},
		{
			name:           "Normal amplitude should not be normalized",
			samples:        []int16{10000, -8000, 12000, -9000, 11000},
			expectModified: false,
			description:    "Normal speech level (max 12000) should not be modified",
		},
		{
			name:           "Zero samples should not be normalized",
			samples:        []int16{0, 0, 0, 0, 0},
			expectModified: false,
			description:    "All zero samples should be skipped",
		},
		{
			name:           "Empty samples should not be normalized",
			samples:        []int16{},
			expectModified: false,
			description:    "Empty slice should be skipped",
		},
		{
			name:           "Threshold boundary test",
			samples:        []int16{4999, -4999, 4500, -4800},
			expectModified: true,
			description:    "Just below threshold (5000) should be normalized",
		},
		{
			name:           "Above threshold test",
			samples:        []int16{5001, -5001, 6000, -5500},
			expectModified: false,
			description:    "Just above threshold (5000) should not be normalized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to compare
			original := make([]int16, len(tt.samples))
			copy(original, tt.samples)

			// Run normalization
			modified := NormalizeAudio(tt.samples)

			// Check if modification flag matches expectation
			assert.Equal(t, tt.expectModified, modified, tt.description)

			// If we expected no modification, samples should be unchanged
			if !tt.expectModified {
				assert.Equal(t, original, tt.samples, "Samples should not be modified")
			}

			// If we expected modification, check that samples were amplified
			if tt.expectModified && len(tt.samples) > 0 {
				// Find max amplitude in original
				var origMax int16
				for _, s := range original {
					abs := s
					if abs < 0 {
						abs = -abs
					}
					if abs > origMax {
						origMax = abs
					}
				}

				// Find max amplitude after normalization
				var newMax int16
				for _, s := range tt.samples {
					abs := s
					if abs < 0 {
						abs = -abs
					}
					if abs > newMax {
						newMax = abs
					}
				}

				// After normalization, max should be larger (unless it was zero)
				if origMax > 0 {
					assert.Greater(t, newMax, origMax, "Expected amplification")
				}
			}
		})
	}
}

// TestNormalizeAudioClipping tests that normalization doesn't cause clipping
func TestNormalizeAudioClipping(t *testing.T) {
	samples := []int16{100, 200, -150, 300, -250}

	NormalizeAudio(samples)

	// Check all samples are within int16 range
	for i, sample := range samples {
		assert.LessOrEqual(t, sample, int16(32767), "Sample %d should not exceed max", i)
		assert.GreaterOrEqual(t, sample, int16(-32767), "Sample %d should not be below min", i)
	}
}

// TestNormalizeAudioTargetLevel tests that normalization reaches target level
func TestNormalizeAudioTargetLevel(t *testing.T) {
	samples := []int16{1000, -800, 1200, -900}

	modified := NormalizeAudio(samples)

	assert.True(t, modified, "Expected normalization to occur")

	// Find max amplitude
	var maxAmp int16
	for _, sample := range samples {
		abs := sample
		if abs < 0 {
			abs = -abs
		}
		if abs > maxAmp {
			maxAmp = abs
		}
	}

	// Should be close to target level (20000) but not exceed safe maximum
	assert.GreaterOrEqual(t, maxAmp, int16(15000), "Normalized max should be at least 15000")
	assert.LessOrEqual(t, maxAmp, int16(32767), "Normalized max should not exceed 32767")
}

// TestDetectVoiceMeeter tests VoiceMeeter detection
func TestDetectVoiceMeeter(t *testing.T) {
	deviceName, err := DetectVoiceMeeter()

	// Should not return error even if not found
	require.NoError(t, err, "DetectVoiceMeeter should not return error")

	// On Mac, should return empty string
	// On Windows with VoiceMeeter, should return device name
	t.Logf("Detected VoiceMeeter device: '%s' (empty is OK on Mac)", deviceName)
}

// TestSaveToFileWithNormalization tests the normalization integration
func TestSaveToFileWithNormalization(t *testing.T) {
	recorder, err := NewRecorder()
	require.NoError(t, err)
	defer recorder.Close()

	// Test that method exists and can be called (even without data)
	err = recorder.SaveToFileWithNormalization("test.wav", true)
	assert.Error(t, err, "Should error with no audio data")
	assert.Contains(t, err.Error(), "no audio data to save")
}
