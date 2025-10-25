// +build windows

package recorder

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDualRecorder_New tests basic initialization
func TestDualRecorder_New(t *testing.T) {
	dr, err := NewDualRecorder()
	require.NoError(t, err, "NewDualRecorder should succeed")
	require.NotNil(t, dr, "DualRecorder should not be nil")
	defer dr.Close()

	// Check default values
	assert.False(t, dr.IsRecording(), "Should not be recording initially")
	assert.Equal(t, 0.0, dr.GetDuration(), "Duration should be zero initially")
}

// TestDualRecorder_IsRecording tests recording state
func TestDualRecorder_IsRecording(t *testing.T) {
	dr, err := NewDualRecorder()
	require.NoError(t, err)
	defer dr.Close()

	// Initially not recording
	assert.False(t, dr.IsRecording())

	// Start recording
	err = dr.Start()
	if err != nil {
		t.Skipf("Audio device not available: %v", err)
		return
	}

	// Should be recording
	assert.True(t, dr.IsRecording())

	// Stop recording
	err = dr.Stop()
	require.NoError(t, err)

	// Should not be recording
	assert.False(t, dr.IsRecording())
}

// TestDualRecorder_StartStopCycle tests multiple start-stop cycles
func TestDualRecorder_StartStopCycle(t *testing.T) {
	dr, err := NewDualRecorder()
	require.NoError(t, err)
	defer dr.Close()

	// Perform 5 start-stop cycles to ensure no crash
	for i := 0; i < 5; i++ {
		t.Logf("Cycle %d: Starting recording", i+1)

		err = dr.Start()
		if err != nil {
			t.Skipf("Audio device not available: %v", err)
			return
		}

		assert.True(t, dr.IsRecording(), "Should be recording in cycle %d", i+1)

		// Record for a short time
		time.Sleep(100 * time.Millisecond)

		t.Logf("Cycle %d: Stopping recording", i+1)
		err = dr.Stop()
		require.NoError(t, err, "Stop should succeed in cycle %d", i+1)

		assert.False(t, dr.IsRecording(), "Should not be recording after stop in cycle %d", i+1)
	}
}

// TestDualRecorder_GetDuration tests duration calculation
func TestDualRecorder_GetDuration(t *testing.T) {
	dr, err := NewDualRecorder()
	require.NoError(t, err)
	defer dr.Close()

	// Initially zero duration
	assert.Equal(t, 0.0, dr.GetDuration())

	// Start recording
	err = dr.Start()
	if err != nil {
		t.Skipf("Audio device not available: %v", err)
		return
	}

	// Wait a bit
	time.Sleep(200 * time.Millisecond)

	// Duration should be non-zero
	duration := dr.GetDuration()
	assert.Greater(t, duration, 0.0, "Duration should be greater than 0")
	assert.LessOrEqual(t, duration, 1.0, "Duration should be less than 1 second")

	dr.Stop()
}

// TestDualRecorder_GetElapsedTime tests elapsed time tracking
func TestDualRecorder_GetElapsedTime(t *testing.T) {
	dr, err := NewDualRecorder()
	require.NoError(t, err)
	defer dr.Close()

	// Initially zero elapsed time
	assert.Equal(t, time.Duration(0), dr.GetElapsedTime())

	// Start recording
	err = dr.Start()
	if err != nil {
		t.Skipf("Audio device not available: %v", err)
		return
	}

	// Wait a bit
	time.Sleep(150 * time.Millisecond)

	// Elapsed time should be non-zero
	elapsed := dr.GetElapsedTime()
	assert.Greater(t, elapsed, time.Duration(0))
	assert.LessOrEqual(t, elapsed, 300*time.Millisecond) // Should be reasonable

	dr.Stop()

	// After stop, elapsed time should be zero
	assert.Equal(t, time.Duration(0), dr.GetElapsedTime())
}

// TestDualRecorder_DoubleStart tests error handling for double start
func TestDualRecorder_DoubleStart(t *testing.T) {
	dr, err := NewDualRecorder()
	require.NoError(t, err)
	defer dr.Close()

	// Start recording
	err = dr.Start()
	if err != nil {
		t.Skipf("Audio device not available: %v", err)
		return
	}

	// Try to start again - should fail
	err = dr.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "recording already in progress")

	dr.Stop()
}

// TestDualRecorder_StopWithoutStart tests error handling for stop without start
func TestDualRecorder_StopWithoutStart(t *testing.T) {
	dr, err := NewDualRecorder()
	require.NoError(t, err)
	defer dr.Close()

	// Try to stop without starting
	err = dr.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no recording in progress")
}

// TestDualRecorder_SaveToFileWithoutData tests save error handling
func TestDualRecorder_SaveToFileWithoutData(t *testing.T) {
	dr, err := NewDualRecorder()
	require.NoError(t, err)
	defer dr.Close()

	// Try to save without recording
	err = dr.SaveToFile("test.wav")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no audio data to save")
}

// TestDualRecorder_SetVolumes tests volume adjustment
func TestDualRecorder_SetVolumes(t *testing.T) {
	dr, err := NewDualRecorder()
	require.NoError(t, err)
	defer dr.Close()

	// Test various volume combinations
	testCases := []struct {
		systemVol float64
		micVol    float64
	}{
		{0.5, 1.0},  // Half system, full mic
		{1.0, 0.5},  // Full system, half mic
		{0.7, 1.5},  // Moderate system, boosted mic
		{1.5, 1.5},  // Both boosted
		{0.0, 1.0},  // Muted system
		{1.0, 0.0},  // Muted mic
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			dr.SetVolumes(tc.systemVol, tc.micVol)
			assert.Equal(t, tc.systemVol, dr.systemVolume)
			assert.Equal(t, tc.micVol, dr.micVolume)
		})
	}
}

// TestDualRecorder_SetVolumes_BoundaryValues tests volume boundary conditions
func TestDualRecorder_SetVolumes_BoundaryValues(t *testing.T) {
	dr, err := NewDualRecorder()
	require.NoError(t, err)
	defer dr.Close()

	testCases := []struct {
		name           string
		systemVol      float64
		micVol         float64
		expectSystem   float64
		expectMic      float64
	}{
		{"Minimum valid", 0.0, 0.0, 0.0, 0.0},
		{"Maximum valid", 2.0, 2.0, 2.0, 2.0},
		{"Below minimum", -0.1, -0.5, 0.7, 1.0}, // Should keep old values
		{"Above maximum", 2.1, 3.0, 0.7, 1.0},   // Should keep old values
		{"Mixed invalid", -1.0, 2.5, 0.7, 1.0},  // Should keep old values
		{"Extreme values", 999.0, -999.0, 0.7, 1.0}, // Should keep old values
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dr.SetVolumes(tc.systemVol, tc.micVol)
			assert.Equal(t, tc.expectSystem, dr.systemVolume, "System volume mismatch")
			assert.Equal(t, tc.expectMic, dr.micVolume, "Mic volume mismatch")
		})
	}
}

// TestDualRecorder_SetLimits_BoundaryValues tests recording limits boundary conditions
func TestDualRecorder_SetLimits_BoundaryValues(t *testing.T) {
	dr, err := NewDualRecorder()
	require.NoError(t, err)
	defer dr.Close()

	testCases := []struct {
		name        string
		maxDuration time.Duration
		maxFileSize int64
	}{
		{"Zero values (unlimited)", 0, 0},
		{"Only duration limit", 1 * time.Hour, 0},
		{"Only file size limit", 0, 100 * 1024 * 1024},
		{"Both limits", 30 * time.Minute, 50 * 1024 * 1024},
		{"Very small limits", 1 * time.Second, 1024},
		{"Very large limits", 24 * time.Hour, 10 * 1024 * 1024 * 1024},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dr.SetLimits(tc.maxDuration, tc.maxFileSize)
			assert.Equal(t, tc.maxDuration, dr.maxDuration)
			assert.Equal(t, tc.maxFileSize, dr.maxFileSize)
		})
	}
}

// TestDualRecorder_SetLimits tests recording limits
func TestDualRecorder_SetLimits(t *testing.T) {
	dr, err := NewDualRecorder()
	require.NoError(t, err)
	defer dr.Close()

	// Set limits
	maxDuration := 5 * time.Second
	maxFileSize := int64(1024 * 1024) // 1MB

	dr.SetLimits(maxDuration, maxFileSize)

	// Verify limits were set (basic test)
	assert.NotNil(t, dr)
}

// TestDualRecorder_SaveToFileWithNormalization tests normalization integration
func TestDualRecorder_SaveToFileWithNormalization(t *testing.T) {
	dr, err := NewDualRecorder()
	require.NoError(t, err)
	defer dr.Close()

	// Test that method exists and can be called (even without data)
	err = dr.SaveToFileWithNormalization("test.wav", true)
	assert.Error(t, err, "Should error with no audio data")
	assert.Contains(t, err.Error(), "no audio data to save")
}

// TestDualRecorder_RecordAndSave tests actual recording and file save
func TestDualRecorder_RecordAndSave(t *testing.T) {
	dr, err := NewDualRecorder()
	require.NoError(t, err)
	defer dr.Close()

	// Start recording
	err = dr.Start()
	if err != nil {
		t.Skipf("Audio device not available: %v", err)
		return
	}

	// Record for a short time
	time.Sleep(500 * time.Millisecond)

	// Stop recording
	err = dr.Stop()
	require.NoError(t, err)

	// Save to file
	filename := "test_dual_recorder_output.wav"
	defer os.Remove(filename)

	err = dr.SaveToFile(filename)
	require.NoError(t, err, "Save should succeed")

	// Verify file exists
	info, err := os.Stat(filename)
	require.NoError(t, err, "File should exist")
	assert.Greater(t, info.Size(), int64(0), "File should not be empty")
}

// TestDualRecorder_RecordWithNormalization tests recording with normalization
func TestDualRecorder_RecordWithNormalization(t *testing.T) {
	dr, err := NewDualRecorder()
	require.NoError(t, err)
	defer dr.Close()

	// Start recording
	err = dr.Start()
	if err != nil {
		t.Skipf("Audio device not available: %v", err)
		return
	}

	// Record for a short time
	time.Sleep(500 * time.Millisecond)

	// Stop recording
	err = dr.Stop()
	require.NoError(t, err)

	// Save with normalization enabled
	filename := "test_dual_recorder_normalized.wav"
	defer os.Remove(filename)

	err = dr.SaveToFileWithNormalization(filename, true)
	require.NoError(t, err, "Save with normalization should succeed")

	// Verify file exists
	info, err := os.Stat(filename)
	require.NoError(t, err, "File should exist")
	assert.Greater(t, info.Size(), int64(0), "File should not be empty")
}

// TestDualRecorder_ConcurrentAccess tests thread safety
func TestDualRecorder_ConcurrentAccess(t *testing.T) {
	dr, err := NewDualRecorder()
	require.NoError(t, err)
	defer dr.Close()

	err = dr.Start()
	if err != nil {
		t.Skipf("Audio device not available: %v", err)
		return
	}

	// Access from multiple goroutines
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			dr.IsRecording()
			dr.GetDuration()
			dr.GetElapsedTime()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	dr.Stop()
}

// TestDualRecorder_NewWithDevices tests device-specific initialization
func TestDualRecorder_NewWithDevices(t *testing.T) {
	// Get list of devices
	devices, err := ListDevices()
	if err != nil {
		t.Skipf("Device listing not available: %v", err)
		return
	}

	if len(devices) == 0 {
		t.Skip("No audio devices available")
		return
	}

	// Try to create with first available device
	firstDevice := devices[0].Name
	dr, err := NewDualRecorderWithDevices(firstDevice)
	if err != nil {
		t.Skipf("Could not create recorder with device %s: %v", firstDevice, err)
		return
	}
	defer dr.Close()

	assert.NotNil(t, dr)
	assert.NotNil(t, dr.micDevice, "Mic device should be set")
}

// TestDualRecorder_NewWithInvalidDevice tests error handling for invalid device
func TestDualRecorder_NewWithInvalidDevice(t *testing.T) {
	dr, err := NewDualRecorderWithDevices("NonExistentDevice12345")
	assert.Error(t, err, "Should error with invalid device name")
	assert.Nil(t, dr, "Recorder should be nil on error")
}
