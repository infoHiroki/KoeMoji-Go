package ui

import (
	"testing"
	"time"

	"github.com/infoHiroki/KoeMoji-Go/internal/config"
	"github.com/stretchr/testify/assert"
)

// TestTUICallbacks_Initialization tests TUICallbacks struct initialization
func TestTUICallbacks_Initialization(t *testing.T) {
	callbacks := &TUICallbacks{
		OnRecordingToggle: func() error { return nil },
		OnScanTrigger:     func() error { return nil },
		OnOpenLogFile:     func() error { return nil },
		OnOpenDirectory:   func(dir string) error { return nil },
		OnRefreshFileList: func() error { return nil },
	}

	assert.NotNil(t, callbacks.OnRecordingToggle)
	assert.NotNil(t, callbacks.OnScanTrigger)
	assert.NotNil(t, callbacks.OnOpenLogFile)
	assert.NotNil(t, callbacks.OnOpenDirectory)
	assert.NotNil(t, callbacks.OnRefreshFileList)
}

// TestTUICallbacks_Execution tests callback execution
func TestTUICallbacks_Execution(t *testing.T) {
	recordingCalled := false
	scanCalled := false
	logFileCalled := false
	directoryCalled := false
	refreshCalled := false

	callbacks := &TUICallbacks{
		OnRecordingToggle: func() error {
			recordingCalled = true
			return nil
		},
		OnScanTrigger: func() error {
			scanCalled = true
			return nil
		},
		OnOpenLogFile: func() error {
			logFileCalled = true
			return nil
		},
		OnOpenDirectory: func(dir string) error {
			directoryCalled = true
			return nil
		},
		OnRefreshFileList: func() error {
			refreshCalled = true
			return nil
		},
	}

	// Execute callbacks
	callbacks.OnRecordingToggle()
	assert.True(t, recordingCalled)

	callbacks.OnScanTrigger()
	assert.True(t, scanCalled)

	callbacks.OnOpenLogFile()
	assert.True(t, logFileCalled)

	callbacks.OnOpenDirectory("/test/path")
	assert.True(t, directoryCalled)

	callbacks.OnRefreshFileList()
	assert.True(t, refreshCalled)
}

// TestVolumeFloatToIndex tests volume conversion from float to index
func TestVolumeFloatToIndex(t *testing.T) {
	tests := []struct {
		name     string
		volume   float64
		expected int
	}{
		{"System volume 0.1", 0.1, 0},
		{"System volume 0.2", 0.2, 1},
		{"System volume 0.3", 0.3, 2},
		{"System volume 0.5", 0.5, 3},
		{"System volume 0.7", 0.7, 4},
		{"Mic volume 1.0", 1.0, 0},
		{"Mic volume 1.3", 1.3, 1},
		{"Mic volume 1.6", 1.6, 2},
		{"Mic volume 1.9", 1.9, 3},
		{"Mic volume 2.2", 2.2, 4},
		{"Default fallback", 999.9, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := volumeFloatToIndex(tt.volume)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestVolumeIndexToSystemFloat tests system volume conversion from index to float
func TestVolumeIndexToSystemFloat(t *testing.T) {
	tests := []struct {
		name     string
		index    int
		expected float64
	}{
		{"Index 0", 0, 0.1},
		{"Index 1", 1, 0.2},
		{"Index 2", 2, 0.3},
		{"Index 3", 3, 0.5},
		{"Index 4", 4, 0.7},
		{"Out of range -1", -1, 0.3}, // Default
		{"Out of range 10", 10, 0.3}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := volumeIndexToSystemFloat(tt.index)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestVolumeIndexToMicFloat tests microphone volume conversion from index to float
func TestVolumeIndexToMicFloat(t *testing.T) {
	tests := []struct {
		name     string
		index    int
		expected float64
	}{
		{"Index 0", 0, 1.0},
		{"Index 1", 1, 1.3},
		{"Index 2", 2, 1.6},
		{"Index 3", 3, 1.9},
		{"Index 4", 4, 2.2},
		{"Out of range -1", -1, 1.6}, // Default
		{"Out of range 10", 10, 1.6}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := volumeIndexToMicFloat(tt.index)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestNewTUI_BasicStructure tests NewTUI creates valid structure
// Note: We can't fully test tview.Application without a terminal,
// but we can verify the function doesn't panic and returns non-nil
func TestNewTUI_BasicStructure(t *testing.T) {
	cfg := config.GetDefaultConfig()
	callbacks := &TUICallbacks{
		OnRecordingToggle: func() error { return nil },
		OnScanTrigger:     func() error { return nil },
		OnOpenLogFile:     func() error { return nil },
		OnOpenDirectory:   func(dir string) error { return nil },
		OnRefreshFileList: func() error { return nil },
	}

	// This will create tview components but won't run the application
	tui := NewTUI(cfg, callbacks)

	// Verify basic fields are initialized
	assert.NotNil(t, tui)
	assert.NotNil(t, tui.app)
	assert.NotNil(t, tui.config)
	assert.NotNil(t, tui.callbacks)
	assert.NotNil(t, tui.menuList)
	assert.NotNil(t, tui.statusBar)
	assert.NotNil(t, tui.helpBar)
	assert.NotNil(t, tui.contentArea)
	assert.NotNil(t, tui.mainFlex)
	assert.False(t, tui.startTime.IsZero())
}

// TestTUI_StatusFields tests status tracking fields initialization
func TestTUI_StatusFields(t *testing.T) {
	cfg := config.GetDefaultConfig()
	callbacks := &TUICallbacks{
		OnRecordingToggle: func() error { return nil },
		OnScanTrigger:     func() error { return nil },
		OnOpenLogFile:     func() error { return nil },
		OnOpenDirectory:   func(dir string) error { return nil },
		OnRefreshFileList: func() error { return nil },
	}

	tui := NewTUI(cfg, callbacks)

	// Verify initial status values
	tui.mu.RLock()
	defer tui.mu.RUnlock()

	assert.Equal(t, 0, tui.inputCount)
	assert.Equal(t, 0, tui.outputCount)
	assert.Equal(t, 0, tui.archiveCount)
	assert.Empty(t, tui.processingFile)
	assert.False(t, tui.isProcessing)
	assert.False(t, tui.isRecording)
	assert.True(t, tui.recordingStart.IsZero())
	assert.False(t, tui.startTime.IsZero())
}

// TestTUI_TimeTracking tests time tracking functionality
func TestTUI_TimeTracking(t *testing.T) {
	cfg := config.GetDefaultConfig()
	callbacks := &TUICallbacks{
		OnRecordingToggle: func() error { return nil },
		OnScanTrigger:     func() error { return nil },
		OnOpenLogFile:     func() error { return nil },
		OnOpenDirectory:   func(dir string) error { return nil },
		OnRefreshFileList: func() error { return nil },
	}

	tui := NewTUI(cfg, callbacks)

	// Verify startTime is set
	assert.False(t, tui.startTime.IsZero())
	assert.WithinDuration(t, time.Now(), tui.startTime, time.Second)
}

// Benchmark tests
func BenchmarkVolumeFloatToIndex(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		volumeFloatToIndex(0.3)
	}
}

func BenchmarkVolumeIndexToSystemFloat(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		volumeIndexToSystemFloat(2)
	}
}

func BenchmarkVolumeIndexToMicFloat(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		volumeIndexToMicFloat(2)
	}
}
