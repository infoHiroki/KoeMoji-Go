// Package testutil provides common testing utilities for KoeMoji-Go project
package testutil

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/infoHiroki/KoeMoji-Go/internal/config"
	customLogger "github.com/infoHiroki/KoeMoji-Go/internal/logger"
	"github.com/stretchr/testify/require"
)

// TestContext holds common test dependencies
type TestContext struct {
	TempDir   string
	Config    *config.Config
	Logger    *log.Logger
	LogBuffer []customLogger.LogEntry
	LogMutex  sync.RWMutex
	Cleanup   func()
}

// NewTestContext creates a new test context with common setup
func NewTestContext(t *testing.T) *TestContext {
	tempDir := t.TempDir()

	// Create test directories
	inputDir := filepath.Join(tempDir, "input")
	outputDir := filepath.Join(tempDir, "output")
	archiveDir := filepath.Join(tempDir, "archive")

	require.NoError(t, os.MkdirAll(inputDir, 0755))
	require.NoError(t, os.MkdirAll(outputDir, 0755))
	require.NoError(t, os.MkdirAll(archiveDir, 0755))

	// Create test config
	cfg := GetTestConfig()
	cfg.InputDir = inputDir
	cfg.OutputDir = outputDir
	cfg.ArchiveDir = archiveDir

	// Create test logger
	logger := log.New(io.Discard, "", log.LstdFlags)
	if testing.Verbose() {
		logger = log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	}

	return &TestContext{
		TempDir:   tempDir,
		Config:    cfg,
		Logger:    logger,
		LogBuffer: make([]customLogger.LogEntry, 0),
		LogMutex:  sync.RWMutex{},
		Cleanup:   func() {}, // No-op cleanup by default
	}
}

// GetTestConfig returns a configuration optimized for testing
func GetTestConfig() *config.Config {
	cfg := config.GetDefaultConfig()

	// Override settings for testing
	cfg.ScanIntervalMinutes = 1
	cfg.MaxCpuPercent = 50        // Reduce CPU usage during tests
	cfg.ComputeType = "int8"      // Faster for testing
	cfg.UseColors = false         // Disable colors in test output
	cfg.LLMSummaryEnabled = false // Disable LLM by default in tests
	cfg.RecordingMaxHours = 1     // Limit recording time
	cfg.RecordingMaxFileMB = 10   // Limit file size

	return cfg
}

// CreateTestConfigFile creates a temporary config file with test settings
func CreateTestConfigFile(t *testing.T, cfg *config.Config) string {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.json")

	data, err := json.MarshalIndent(cfg, "", "  ")
	require.NoError(t, err)

	err = os.WriteFile(configFile, data, 0644)
	require.NoError(t, err)

	return configFile
}

// CreateTestAudioFiles creates test audio files in the specified directory
func CreateTestAudioFiles(t *testing.T, dir string, files ...string) []string {
	createdFiles := make([]string, 0, len(files))

	for _, filename := range files {
		filePath := filepath.Join(dir, filename)

		// Create a minimal WAV file header for testing
		wavData := CreateMinimalWAVData()
		err := os.WriteFile(filePath, wavData, 0644)
		require.NoError(t, err)

		createdFiles = append(createdFiles, filePath)
	}

	return createdFiles
}

// CreateMinimalWAVData creates minimal WAV file data for testing
func CreateMinimalWAVData() []byte {
	// Minimal WAV header (44 bytes) + minimal audio data
	wavHeader := []byte{
		// RIFF header
		0x52, 0x49, 0x46, 0x46, // "RIFF"
		0x28, 0x00, 0x00, 0x00, // File size - 8 (40 bytes)
		0x57, 0x41, 0x56, 0x45, // "WAVE"

		// fmt chunk
		0x66, 0x6d, 0x74, 0x20, // "fmt "
		0x10, 0x00, 0x00, 0x00, // Chunk size (16)
		0x01, 0x00, // Audio format (PCM)
		0x01, 0x00, // Number of channels (1)
		0x44, 0xac, 0x00, 0x00, // Sample rate (44100)
		0x88, 0x58, 0x01, 0x00, // Byte rate
		0x02, 0x00, // Block align
		0x10, 0x00, // Bits per sample (16)

		// data chunk
		0x64, 0x61, 0x74, 0x61, // "data"
		0x04, 0x00, 0x00, 0x00, // Data size (4 bytes)
		0x00, 0x00, 0x00, 0x00, // Minimal audio data
	}

	return wavHeader
}

// MockHTTPServer creates a mock HTTP server for testing API calls
type MockHTTPServer struct {
	Server   *httptest.Server
	Requests []MockRequest
	mutex    sync.Mutex
}

// MockRequest captures request details
type MockRequest struct {
	Method  string
	Path    string
	Body    string
	Headers http.Header
	Time    time.Time
}

// NewMockHTTPServer creates a new mock HTTP server
func NewMockHTTPServer() *MockHTTPServer {
	mock := &MockHTTPServer{
		Requests: make([]MockRequest, 0),
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mock.mutex.Lock()
		defer mock.mutex.Unlock()

		// Read request body
		body, _ := io.ReadAll(r.Body)

		// Store request details
		mock.Requests = append(mock.Requests, MockRequest{
			Method:  r.Method,
			Path:    r.URL.Path,
			Body:    string(body),
			Headers: r.Header.Clone(),
			Time:    time.Now(),
		})

		// Default response
		mock.handleDefaultResponse(w, r)
	})

	mock.Server = httptest.NewServer(handler)
	return mock
}

// handleDefaultResponse provides default responses for common endpoints
func (m *MockHTTPServer) handleDefaultResponse(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/v1/chat/completions":
		// Mock OpenAI API response
		response := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"content": "Mock summary of the audio content.",
						"role":    "assistant",
					},
				},
			},
			"usage": map[string]interface{}{
				"total_tokens": 100,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Mock response"))
	}
}

// Close shuts down the mock server
func (m *MockHTTPServer) Close() {
	m.Server.Close()
}

// GetRequestCount returns the number of requests received
func (m *MockHTTPServer) GetRequestCount() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return len(m.Requests)
}

// GetLastRequest returns the last request received
func (m *MockHTTPServer) GetLastRequest() *MockRequest {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.Requests) == 0 {
		return nil
	}

	return &m.Requests[len(m.Requests)-1]
}

// ClearRequests clears all recorded requests
func (m *MockHTTPServer) ClearRequests() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Requests = m.Requests[:0]
}

// WaitForPort waits for a port to become available
func WaitForPort(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}

	return fmt.Errorf("port %d did not become available within %v", port, timeout)
}

// CreateTestEnvironment sets up environment variables for testing
func CreateTestEnvironment(t *testing.T, envVars map[string]string) func() {
	originalVars := make(map[string]string)

	// Save original values and set test values
	for key, value := range envVars {
		if original := os.Getenv(key); original != "" {
			originalVars[key] = original
		}
		os.Setenv(key, value)
	}

	// Return cleanup function
	return func() {
		for key := range envVars {
			if original, exists := originalVars[key]; exists {
				os.Setenv(key, original)
			} else {
				os.Unsetenv(key)
			}
		}
	}
}

// AssertFileExists checks if a file exists and optionally validates its content
func AssertFileExists(t *testing.T, filePath string, expectedContentContains ...string) {
	require.FileExists(t, filePath)

	if len(expectedContentContains) > 0 {
		content, err := os.ReadFile(filePath)
		require.NoError(t, err)

		contentStr := string(content)
		for _, expected := range expectedContentContains {
			require.Contains(t, contentStr, expected,
				"File %s should contain %s", filePath, expected)
		}
	}
}

// CreateLargeTestFile creates a test file of specified size for performance testing
func CreateLargeTestFile(t *testing.T, filePath string, sizeMB int) {
	file, err := os.Create(filePath)
	require.NoError(t, err)
	defer file.Close()

	// Write WAV header first
	wavData := CreateMinimalWAVData()
	_, err = file.Write(wavData)
	require.NoError(t, err)

	// Fill with dummy data to reach desired size
	chunkSize := 1024 * 1024 // 1MB chunks
	chunk := make([]byte, chunkSize)

	remainingBytes := (sizeMB * 1024 * 1024) - len(wavData)
	for remainingBytes > 0 {
		writeSize := chunkSize
		if remainingBytes < chunkSize {
			writeSize = remainingBytes
		}

		_, err = file.Write(chunk[:writeSize])
		require.NoError(t, err)

		remainingBytes -= writeSize
	}
}

// SkipIfShort skips the test if running in short mode
func SkipIfShort(t *testing.T, reason string) {
	if testing.Short() {
		t.Skipf("Skipping in short mode: %s", reason)
	}
}
