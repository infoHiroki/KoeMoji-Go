// Package testutil provides common utilities for integration and benchmark tests
package testutil

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/hirokitakamura/koemoji-go/internal/config"
)

// TestEnvironment represents a test environment with temporary directories
type TestEnvironment struct {
	BaseDir    string
	InputDir   string
	OutputDir  string
	ArchiveDir string
	ConfigFile string
	Config     *config.Config
	cleanup    func()
	mu         sync.Mutex
}

// NewTestEnvironment creates a new test environment with temporary directories
func NewTestEnvironment(t *testing.T) *TestEnvironment {
	t.Helper()

	baseDir := t.TempDir()
	inputDir := filepath.Join(baseDir, "input")
	outputDir := filepath.Join(baseDir, "output")
	archiveDir := filepath.Join(baseDir, "archive")
	configFile := filepath.Join(baseDir, "config.json")

	// Create directories
	for _, dir := range []string{inputDir, outputDir, archiveDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test configuration
	cfg := &config.Config{
		WhisperModel:        "tiny",
		Language:            "ja",
		UILanguage:          "ja",
		ScanIntervalMinutes: 1,
		MaxCpuPercent:       50, // Lower for tests
		ComputeType:         "int8",
		UseColors:           false, // Disable for consistent test output
		OutputFormat:        "txt",
		InputDir:            inputDir,
		OutputDir:           outputDir,
		ArchiveDir:          archiveDir,
		RecordingDeviceName: "Test Device",
	}

	// Save config to file
	if err := saveConfig(configFile, cfg); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	return &TestEnvironment{
		BaseDir:    baseDir,
		InputDir:   inputDir,
		OutputDir:  outputDir,
		ArchiveDir: archiveDir,
		ConfigFile: configFile,
		Config:     cfg,
	}
}

// CreateTestAudioFile creates a test audio file with specified duration (in seconds)
func (env *TestEnvironment) CreateTestAudioFile(t *testing.T, filename string, durationSec int) string {
	t.Helper()

	filePath := filepath.Join(env.InputDir, filename)

	// Create a simple WAV file for testing
	// This is a minimal WAV file with silence
	wavData := createSilentWAV(durationSec)

	if err := os.WriteFile(filePath, wavData, 0644); err != nil {
		t.Fatalf("Failed to create test audio file %s: %v", filePath, err)
	}

	return filePath
}

// createSilentWAV creates a minimal WAV file with silence
func createSilentWAV(durationSec int) []byte {
	sampleRate := 44100
	samples := sampleRate * durationSec
	dataSize := samples * 2 // 16-bit mono

	header := []byte{
		// RIFF header
		'R', 'I', 'F', 'F',
		0, 0, 0, 0, // File size (will be filled)
		'W', 'A', 'V', 'E',

		// fmt chunk
		'f', 'm', 't', ' ',
		16, 0, 0, 0, // fmt chunk size
		1, 0, // PCM format
		1, 0, // mono
		0x44, 0xAC, 0, 0, // sample rate (44100)
		0x88, 0x58, 1, 0, // byte rate
		2, 0, // block align
		16, 0, // bits per sample

		// data chunk
		'd', 'a', 't', 'a',
		0, 0, 0, 0, // data size (will be filled)
	}

	// Fill in sizes
	fileSize := len(header) + dataSize - 8
	header[4] = byte(fileSize)
	header[5] = byte(fileSize >> 8)
	header[6] = byte(fileSize >> 16)
	header[7] = byte(fileSize >> 24)

	header[40] = byte(dataSize)
	header[41] = byte(dataSize >> 8)
	header[42] = byte(dataSize >> 16)
	header[43] = byte(dataSize >> 24)

	// Create silent data
	data := make([]byte, dataSize)

	return append(header, data...)
}

// WaitForFile waits for a file to appear with timeout
func (env *TestEnvironment) WaitForFile(t *testing.T, filePath string, timeout time.Duration) bool {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(filePath); err == nil {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// WaitForFileContent waits for a file to have non-empty content
func (env *TestEnvironment) WaitForFileContent(t *testing.T, filePath string, timeout time.Duration) bool {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if info, err := os.Stat(filePath); err == nil && info.Size() > 0 {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// GetFileContent reads file content safely
func (env *TestEnvironment) GetFileContent(t *testing.T, filePath string) string {
	t.Helper()

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", filePath, err)
	}
	return string(content)
}

// ListFiles lists all files in a directory
func (env *TestEnvironment) ListFiles(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Failed to read directory %s: %v", dir, err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files
}

// Cleanup performs cleanup operations
func (env *TestEnvironment) Cleanup(t *testing.T) {
	t.Helper()
	env.mu.Lock()
	defer env.mu.Unlock()

	if env.cleanup != nil {
		env.cleanup()
	}
}

// saveConfig saves configuration to file
func saveConfig(filename string, cfg *config.Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// MockLLMServer provides a mock HTTP server for LLM testing
type MockLLMServer struct {
	Server interface{} // Will be httptest.Server
	URL    string
}

// BenchmarkMetrics holds benchmark measurement results
type BenchmarkMetrics struct {
	Duration     time.Duration
	MemoryUsage  int64
	Allocations  int64
	Goroutines   int
	FileSize     int64
	ProcessSpeed float64 // MB/s
}

// MeasurePerformance measures performance metrics during benchmark execution
func MeasurePerformance(b *testing.B, fn func()) *BenchmarkMetrics {
	var startMem, endMem runtime.MemStats
	var startGoroutines, endGoroutines int

	runtime.GC()
	runtime.ReadMemStats(&startMem)
	startGoroutines = runtime.NumGoroutine()

	start := time.Now()
	fn()
	duration := time.Since(start)

	runtime.GC()
	runtime.ReadMemStats(&endMem)
	endGoroutines = runtime.NumGoroutine()

	return &BenchmarkMetrics{
		Duration:    duration,
		MemoryUsage: int64(endMem.Alloc - startMem.Alloc),
		Allocations: int64(endMem.TotalAlloc - startMem.TotalAlloc),
		Goroutines:  endGoroutines - startGoroutines,
	}
}

// ReportMetrics reports benchmark metrics in a structured format
func (m *BenchmarkMetrics) ReportMetrics(b *testing.B) {
	b.ReportMetric(float64(m.Duration.Nanoseconds()), "ns/op")
	b.ReportMetric(float64(m.MemoryUsage), "B/op")
	b.ReportMetric(float64(m.Allocations), "allocs/op")
	b.ReportMetric(float64(m.Goroutines), "goroutines")
	if m.ProcessSpeed > 0 {
		b.ReportMetric(m.ProcessSpeed, "MB/s")
	}
}

// SetProcessSpeed calculates and sets processing speed in MB/s
func (m *BenchmarkMetrics) SetProcessSpeed(fileSize int64) {
	if m.Duration > 0 && fileSize > 0 {
		mbSize := float64(fileSize) / 1024 / 1024
		seconds := m.Duration.Seconds()
		m.ProcessSpeed = mbSize / seconds
		m.FileSize = fileSize
	}
}

// CreateLargeTestFileForBenchmark creates a large test file for performance testing
func CreateLargeTestFileForBenchmark(t *testing.T, dir, filename string, sizeMB int) string {
	t.Helper()

	filePath := filepath.Join(dir, filename)

	// Create a large WAV file
	sampleRate := 44100
	durationSec := int(math.Ceil(float64(sizeMB) * 1024 * 1024 / (float64(sampleRate) * 2)))
	wavData := createSilentWAV(durationSec)

	if err := os.WriteFile(filePath, wavData, 0644); err != nil {
		t.Fatalf("Failed to create large test file %s: %v", filePath, err)
	}

	return filePath
}

// SkipIfShortForBenchmark skips test if running in short mode
func SkipIfShortForBenchmark(t *testing.T, reason string) {
	if testing.Short() {
		t.Skipf("Skipping in short mode: %s", reason)
	}
}

// SkipIfNoWhisper skips test if whisper is not available
func SkipIfNoWhisper(t *testing.T) {
	// This would check if whisper-ctranslate2 is available
	// For now, we'll assume it's available in CI/local testing
	// In real implementation, you'd check the PATH or run a test command
}

// LogCapture captures log output for testing
type LogCapture struct {
	mu      sync.Mutex
	entries []string
}

// NewLogCapture creates a new log capture
func NewLogCapture() *LogCapture {
	return &LogCapture{
		entries: make([]string, 0),
	}
}

// Write implements io.Writer interface
func (lc *LogCapture) Write(p []byte) (n int, err error) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.entries = append(lc.entries, string(p))
	return len(p), nil
}

// GetEntries returns captured log entries
func (lc *LogCapture) GetEntries() []string {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	entries := make([]string, len(lc.entries))
	copy(entries, lc.entries)
	return entries
}

// Clear clears captured entries
func (lc *LogCapture) Clear() {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.entries = lc.entries[:0]
}

// MultiWriter creates a writer that writes to multiple destinations
func MultiWriter(writers ...io.Writer) io.Writer {
	return io.MultiWriter(writers...)
}

// TestReporter provides structured test reporting
type TestReporter struct {
	t         *testing.T
	startTime time.Time
	steps     []string
}

// NewTestReporter creates a new test reporter
func NewTestReporter(t *testing.T) *TestReporter {
	return &TestReporter{
		t:         t,
		startTime: time.Now(),
		steps:     make([]string, 0),
	}
}

// Step records a test step
func (tr *TestReporter) Step(step string) {
	tr.steps = append(tr.steps, fmt.Sprintf("[%s] %s", time.Since(tr.startTime).Round(time.Millisecond), step))
	tr.t.Logf("STEP: %s", step)
}

// Report provides a final test report
func (tr *TestReporter) Report() {
	tr.t.Logf("TEST COMPLETED in %s", time.Since(tr.startTime).Round(time.Millisecond))
	tr.t.Logf("Steps executed:")
	for _, step := range tr.steps {
		tr.t.Logf("  %s", step)
	}
}
