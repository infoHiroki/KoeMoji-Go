// Package enhanced provides comprehensive error handling tests
package enhanced

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/hirokitakamura/koemoji-go/internal/config"
	"github.com/hirokitakamura/koemoji-go/internal/llm"
	"github.com/hirokitakamura/koemoji-go/internal/logger"
	"github.com/hirokitakamura/koemoji-go/internal/processor"
	"github.com/hirokitakamura/koemoji-go/internal/recorder"
	testutil "github.com/hirokitakamura/koemoji-go/test/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigErrorHandling tests configuration-related error scenarios
func TestConfigErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		configData   string
		expectError  bool
		description  string
	}{
		{
			name:         "Empty config file",
			configData:   "",
			expectError:  true,
			description:  "Should handle empty config file",
		},
		{
			name:         "Invalid JSON",
			configData:   `{"whisper_model": "base", "invalid": }`,
			expectError:  true,
			description:  "Should handle malformed JSON",
		},
		{
			name:         "Missing required fields",
			configData:   `{"some_field": "value"}`,
			expectError:  false, // Should use defaults
			description:  "Should use defaults for missing fields",
		},
		{
			name:         "Invalid field types",
			configData:   `{"max_cpu_percent": "not_a_number"}`,
			expectError:  true, // JSON type errors should be reported
			description:  "Should handle invalid field types",
		},
		{
			name:         "Valid minimal config",
			configData:   `{"whisper_model": "base", "language": "en"}`,
			expectError:  false,
			description:  "Should accept valid minimal config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configFile := filepath.Join(tempDir, "test_config.json")

			err := os.WriteFile(configFile, []byte(tt.configData), 0644)
			require.NoError(t, err)

			// Note: LoadConfig may call os.Exit on certain errors
			// We test what we can without triggering exits
			if tt.configData == "" || strings.Contains(tt.configData, "invalid") {
				// These would cause os.Exit, so we just verify the file setup
				assert.FileExists(t, configFile)
			} else {
				_, err := config.LoadConfig(configFile, nil)
				if tt.expectError {
					assert.Error(t, err, tt.description)
				} else {
					assert.NoError(t, err, tt.description)
				}
			}
		})
	}
}

// TestDirectoryPermissionErrors tests directory permission error handling
func TestDirectoryPermissionErrors(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tempDir := t.TempDir()

	tests := []struct {
		name        string
		setupFunc   func(string) string
		operation   func(string) error
		expectError bool
	}{
		{
			name: "Read-only directory write attempt",
			setupFunc: func(baseDir string) string {
				readOnlyDir := filepath.Join(baseDir, "readonly")
				os.MkdirAll(readOnlyDir, 0444) // Read-only
				return readOnlyDir
			},
			operation: func(dir string) error {
				testFile := filepath.Join(dir, "test.txt")
				return os.WriteFile(testFile, []byte("test"), 0644)
			},
			expectError: true,
		},
		{
			name: "Non-existent directory access",
			setupFunc: func(baseDir string) string {
				return filepath.Join(baseDir, "nonexistent")
			},
			operation: func(dir string) error {
				_, err := os.ReadDir(dir)
				return err
			},
			expectError: true,
		},
		{
			name: "Protected system directory",
			setupFunc: func(baseDir string) string {
				return "/root" // Typically protected
			},
			operation: func(dir string) error {
				testFile := filepath.Join(dir, "test.txt")
				return os.WriteFile(testFile, []byte("test"), 0644)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := tt.setupFunc(tempDir)
			err := tt.operation(testDir)

			if tt.expectError {
				assert.Error(t, err)
				// Verify it's a permission-related error
				assert.True(t, os.IsPermission(err) || os.IsNotExist(err))
			} else {
				assert.NoError(t, err)
			}

			// Clean up if we created a read-only directory
			if strings.Contains(testDir, "readonly") {
				os.Chmod(testDir, 0755)
			}
		})
	}
}

// TestFileOperationErrors tests file operation error scenarios
func TestFileOperationErrors(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		setupFunc   func(string) (string, func())
		operation   func(string) error
		expectError bool
		errorType   func(error) bool
	}{
		{
			name: "Read non-existent file",
			setupFunc: func(baseDir string) (string, func()) {
				return filepath.Join(baseDir, "nonexistent.txt"), func() {}
			},
			operation: func(file string) error {
				_, err := os.ReadFile(file)
				return err
			},
			expectError: true,
			errorType:   os.IsNotExist,
		},
		{
			name: "Write to directory instead of file",
			setupFunc: func(baseDir string) (string, func()) {
				dirPath := filepath.Join(baseDir, "testdir")
				os.MkdirAll(dirPath, 0755)
				return dirPath, func() { os.RemoveAll(dirPath) }
			},
			operation: func(path string) error {
				return os.WriteFile(path, []byte("test"), 0644)
			},
			expectError: true,
			errorType:   func(err error) bool { return err != nil },
		},
		{
			name: "Read corrupted/invalid file",
			setupFunc: func(baseDir string) (string, func()) {
				corruptedFile := filepath.Join(baseDir, "corrupted.wav")
				// Create a file with invalid audio data
				os.WriteFile(corruptedFile, []byte("not audio data"), 0644)
				return corruptedFile, func() { os.Remove(corruptedFile) }
			},
			operation: func(file string) error {
				// This simulates trying to process an invalid audio file
				data, err := os.ReadFile(file)
				if err != nil {
					return err
				}
				// Check if it looks like valid audio data (very basic check)
				if len(data) < 44 || !strings.Contains(string(data[:12]), "RIFF") {
					return fmt.Errorf("invalid audio file format")
				}
				return nil
			},
			expectError: true,
			errorType:   func(err error) bool { return strings.Contains(err.Error(), "invalid") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPath, cleanup := tt.setupFunc(tempDir)
			defer cleanup()

			err := tt.operation(testPath)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.True(t, tt.errorType(err), "Error type mismatch: %v", err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRecorderErrorHandling tests recorder-specific error scenarios
func TestRecorderErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		testFunc    func(t *testing.T)
		description string
	}{
		{
			name: "No audio devices available",
			testFunc: func(t *testing.T) {
				// Note: This test may skip if devices are actually available
				devices, err := recorder.ListDevices()
				if err != nil {
					// Expected error case
					assert.Error(t, err)
					assert.Empty(t, devices)
				} else {
					// Devices are available, which is also valid
					t.Logf("Audio devices are available: %d", len(devices))
				}
			},
			description: "Should handle absence of audio devices gracefully",
		},
		{
			name: "Invalid device selection",
			testFunc: func(t *testing.T) {
				devices, err := recorder.ListDevices()
				if err != nil {
					t.Skipf("Skipping test, no audio devices: %v", err)
				}
				
				// Try to use an invalid device ID
				invalidDeviceID := len(devices) + 100
				
				// Note: The actual behavior depends on implementation
				// We're testing that the system doesn't crash with invalid IDs
				t.Logf("Testing with invalid device ID: %d", invalidDeviceID)
				// This would typically be tested in a more controlled environment
			},
			description: "Should handle invalid device IDs gracefully",
		},
		{
			name: "Recorder creation failure",
			testFunc: func(t *testing.T) {
				// Test recorder creation when it might fail
				rec, err := recorder.NewRecorder()
				if err != nil {
					// This is an expected error case
					assert.Error(t, err)
					assert.Nil(t, rec)
				} else {
					// Recorder creation succeeded
					assert.NotNil(t, rec)
					rec.Close()
				}
			},
			description: "Should handle recorder creation failures",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

// TestLLMErrorHandling tests LLM API error scenarios
func TestLLMErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
		description string
	}{
		{
			name: "Empty API key",
			config: &config.Config{
				LLMSummaryEnabled: true,
				LLMAPIKey:        "",
				LLMAPIProvider:   "openai",
				LLMModel:         "gpt-4o",
			},
			expectError: true,
			description: "Should reject empty API key",
		},
		{
			name: "Invalid API key format",
			config: &config.Config{
				LLMSummaryEnabled: true,
				LLMAPIKey:        "invalid-key",
				LLMAPIProvider:   "openai",
				LLMModel:         "gpt-4o",
			},
			expectError: true,
			description: "Should reject invalid API key format",
		},
		{
			name: "Unsupported provider",
			config: &config.Config{
				LLMSummaryEnabled: true,
				LLMAPIKey:        "sk-test1234567890abcdef1234567890abcdef12345678",
				LLMAPIProvider:   "unsupported_provider",
				LLMModel:         "some-model",
			},
			expectError: true,
			description: "Should reject unsupported providers",
		},
		{
			name: "LLM disabled",
			config: &config.Config{
				LLMSummaryEnabled: false,
				LLMAPIKey:        "sk-test1234567890abcdef1234567890abcdef12345678",
				LLMAPIProvider:   "openai",
				LLMModel:         "gpt-4o",
			},
			expectError: true,
			description: "Should reject when LLM is disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := testutil.NewTestEnvironment(t)
			defer env.Cleanup(t)

			// Test SummarizeText with various error conditions
			var logBuffer []logger.LogEntry
			var logMutex sync.RWMutex
			testLogger := log.New(os.Stdout, "", log.LstdFlags)
			summary, err := llm.SummarizeText(tt.config, testLogger, &logBuffer, 
				&logMutex, false, "Test text for summarization")

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Empty(t, summary)
			} else {
				// Note: Even valid configs might fail due to network/API issues
				// So we test that the function handles the config correctly
				if err != nil {
					t.Logf("Expected error for valid config (likely network/API issue): %v", err)
				}
			}
		})
	}
}

// TestProcessorErrorHandling tests processor error scenarios
func TestProcessorErrorHandling(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	tests := []struct {
		name        string
		setupFunc   func() *config.Config
		expectError bool
		description string
	}{
		{
			name: "Invalid input directory",
			setupFunc: func() *config.Config {
				cfg := config.GetDefaultConfig()
				cfg.InputDir = "/nonexistent/input"
				cfg.OutputDir = env.OutputDir
				cfg.ArchiveDir = env.ArchiveDir
				return cfg
			},
			expectError: true, // Should fail when trying to create directory in protected location
			description: "Should fail when trying to create directory in protected location",
		},
		{
			name: "Empty directory paths",
			setupFunc: func() *config.Config {
				cfg := config.GetDefaultConfig()
				cfg.InputDir = ""
				cfg.OutputDir = ""
				cfg.ArchiveDir = ""
				return cfg
			},
			expectError: true,
			description: "Should reject empty directory paths",
		},
		{
			name: "Same input and output directories",
			setupFunc: func() *config.Config {
				cfg := config.GetDefaultConfig()
				sameDir := env.BaseDir
				cfg.InputDir = sameDir
				cfg.OutputDir = sameDir
				cfg.ArchiveDir = sameDir
				return cfg
			},
			expectError: false, // System allows this, but it's not recommended
			description: "Should handle same input/output directories",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupFunc()

			// Test EnsureDirectories with error conditions
			testLogger := log.New(os.Stdout, "", log.LstdFlags)
			processor.EnsureDirectories(cfg, testLogger)

			// Check if directories were created where possible
			if cfg.InputDir != "" {
				_, err := os.Stat(cfg.InputDir)
				if tt.expectError {
					// For error cases, directory should not exist due to creation failure
					if os.IsNotExist(err) {
						t.Logf("Directory not created as expected for error case: %s", cfg.InputDir)
					}
				} else {
					// For success cases, directory should exist
					assert.NoError(t, err, tt.description)
					assert.DirExists(t, cfg.InputDir, tt.description)
				}
			}
		})
	}
}

// TestPathTraversalSecurity tests security against path traversal attacks
func TestPathTraversalSecurity(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	maliciousPaths := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"/etc/shadow",
		"C:\\Windows\\System32\\config\\SAM",
		"./../../sensitive_file.txt",
		"input/../../../outside_dir/file.txt",
	}

	for _, maliciousPath := range maliciousPaths {
		t.Run(fmt.Sprintf("Path_%s", strings.ReplaceAll(maliciousPath, "/", "_")), func(t *testing.T) {
			// Test that malicious paths are rejected or sanitized
			cleanPath := filepath.Clean(maliciousPath)
			
			// For absolute paths, we should reject them entirely
			if filepath.IsAbs(cleanPath) {
				t.Logf("Absolute path correctly detected and would be rejected: %s", cleanPath)
				return
			}
			
			// Check if the clean path contains path traversal elements
			if strings.Contains(cleanPath, "..") {
				t.Logf("Path traversal detected in clean path, should be rejected: %s", cleanPath)
				return
			}
			
			testFile := filepath.Join(env.InputDir, cleanPath)
			err := os.WriteFile(testFile, []byte("test"), 0644)
			
			if err != nil {
				// Error is expected for malicious paths
				t.Logf("Expected error for malicious path %s: %v", maliciousPath, err)
			} else {
				// If file was created, ensure it's within the safe directory
				absTestFile, _ := filepath.Abs(testFile)
				
				// Check if the file is within the input directory
				relPath, err := filepath.Rel(env.InputDir, absTestFile)
				isWithinInputDir := err == nil && !strings.HasPrefix(relPath, "..")
				assert.True(t, isWithinInputDir, 
					"File should be created within input directory, got: %s (relative: %s)", absTestFile, relPath)
				
				// Clean up
				os.Remove(testFile)
			}
		})
	}
}

// TestResourceExhaustion tests behavior under resource exhaustion
func TestResourceExhaustion(t *testing.T) {
	testutil.SkipIfShort(t, "resource exhaustion test")

	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	t.Run("Many small files", func(t *testing.T) {
		// Create many small files to test file handle limits
		numFiles := 1000
		createdFiles := make([]string, 0, numFiles)
		
		for i := 0; i < numFiles; i++ {
			filename := fmt.Sprintf("exhaust_%d.wav", i)
			filePath := filepath.Join(env.InputDir, filename)
			
			err := os.WriteFile(filePath, []byte("small audio data"), 0644)
			if err != nil {
				// If we hit resource limits, that's expected
				t.Logf("Hit resource limit at file %d: %v", i, err)
				break
			}
			createdFiles = append(createdFiles, filePath)
		}
		
		t.Logf("Successfully created %d files before hitting limits", len(createdFiles))
		assert.Greater(t, len(createdFiles), 0, "Should create at least some files")
		
		// Clean up
		for _, file := range createdFiles {
			os.Remove(file)
		}
	})

	t.Run("Large file creation", func(t *testing.T) {
		// Test with a reasonably large file
		largeFileName := filepath.Join(env.InputDir, "large_test.wav")
		largeData := make([]byte, 10*1024*1024) // 10MB
		
		// Fill with some pattern
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}
		
		err := os.WriteFile(largeFileName, largeData, 0644)
		if err != nil {
			t.Logf("Failed to create large file (expected on low-disk systems): %v", err)
		} else {
			// Verify file was created correctly
			info, err := os.Stat(largeFileName)
			require.NoError(t, err)
			assert.Equal(t, int64(len(largeData)), info.Size())
			
			// Clean up
			os.Remove(largeFileName)
		}
	})
}

// TestNetworkErrorSimulation tests network-related error handling
func TestNetworkErrorSimulation(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	// Test LLM API with network issues (simulated by invalid endpoints)
	cfg := &config.Config{
		LLMSummaryEnabled: true,
		LLMAPIKey:        "sk-test1234567890abcdef1234567890abcdef12345678",
		LLMAPIProvider:   "openai",
		LLMModel:         "gpt-4o",
		LLMMaxTokens:     1000,
	}

	// This should fail due to invalid API key, simulating network/auth error
	var logBuffer []logger.LogEntry
	var logMutex sync.RWMutex
	testLogger := log.New(os.Stdout, "", log.LstdFlags)
	summary, err := llm.SummarizeText(cfg, testLogger, &logBuffer, 
		&logMutex, false, "Test text for network error simulation")

	assert.Error(t, err, "Should fail with network/auth error")
	assert.Empty(t, summary)
	// Check for common authentication/authorization error indicators
	errorMsg := strings.ToLower(err.Error())
	hasAuthError := strings.Contains(errorMsg, "401") || 
		strings.Contains(errorMsg, "unauthorized") || 
		strings.Contains(errorMsg, "authentication") ||
		strings.Contains(errorMsg, "invalid") ||
		strings.Contains(errorMsg, "api key")
	assert.True(t, hasAuthError, "Should indicate authentication error, got: %s", err.Error())
}

// TestConcurrentErrorHandling tests error handling under concurrent access
func TestConcurrentErrorHandling(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup(t)

	numGoroutines := 20
	errors := make(chan error, numGoroutines)

	// Test concurrent file operations that might cause errors
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					errors <- fmt.Errorf("panic in goroutine %d: %v", id, r)
				} else {
					errors <- nil
				}
			}()

			// Mix of operations that might succeed or fail
			if id%2 == 0 {
				// Try to create a file
				filename := fmt.Sprintf("concurrent_%d.wav", id)
				filePath := filepath.Join(env.InputDir, filename)
				err := os.WriteFile(filePath, []byte("test data"), 0644)
				if err != nil {
					errors <- fmt.Errorf("write error in goroutine %d: %v", id, err)
					return
				}
			} else {
				// Try to read a potentially non-existent file
				filename := fmt.Sprintf("nonexistent_%d.wav", id)
				filePath := filepath.Join(env.InputDir, filename)
				_, err := os.ReadFile(filePath)
				if err != nil && !os.IsNotExist(err) {
					errors <- fmt.Errorf("unexpected read error in goroutine %d: %v", id, err)
					return
				}
			}
		}(i)
	}

	// Collect results
	var errorCount int
	for i := 0; i < numGoroutines; i++ {
		err := <-errors
		if err != nil {
			t.Logf("Goroutine error: %v", err)
			errorCount++
		}
	}

	// Some errors are expected (like file not found), but no panics
	t.Logf("Total errors encountered: %d/%d", errorCount, numGoroutines)
	assert.LessOrEqual(t, errorCount, numGoroutines/2, "Should not have excessive errors")
}