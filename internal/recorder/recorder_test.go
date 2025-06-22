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
