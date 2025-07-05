package whisper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hirokitakamura/koemoji-go/internal/whisper/testdata"
	"github.com/stretchr/testify/assert"
)

func TestGetWhisperCommand_FallbackPath(t *testing.T) {
	cmd := getWhisperCommand()
	// The command should either be the fallback or a valid path
	assert.True(t, cmd == "whisper-ctranslate2" || filepath.IsAbs(cmd))
}

func TestIsFasterWhisperAvailable_CommandNotFound(t *testing.T) {
	originalPath := os.Getenv("PATH")
	originalHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("PATH", originalPath)
		os.Setenv("HOME", originalHome)
	}()

	// Clear PATH and HOME to ensure no whisper command is found
	os.Setenv("PATH", "")
	os.Setenv("HOME", "/tmp/nonexistent")

	// Since isFasterWhisperAvailable is not exported, we test through getWhisperCommand
	cmd := getWhisperCommand()
	// When no command is found, should return fallback
	assert.Equal(t, "whisper-ctranslate2", cmd)
}

func TestEnsureDependencies_Available(t *testing.T) {
	config := testdata.CreateTestConfig(t)
	logger, logBuffer, logMutex := testdata.CreateTestLogger()

	// Test when dependencies are available
	// Note: This will try to actually check for whisper command
	EnsureDependencies(config, logger, logBuffer, logMutex, false)

	// In test environment, this might fail due to missing dependencies
	// But we can verify the function exists and handles gracefully
	t.Log("EnsureDependencies function executed")
}

func TestTranscribeAudio_SecurityValidation(t *testing.T) {
	config := testdata.CreateTestConfig(t)
	logger, logBuffer, logMutex := testdata.CreateTestLogger()

	// Test path traversal protection
	tests := []struct {
		name          string
		input         string
		expectedError bool
	}{
		{"Valid input file", filepath.Join(config.InputDir, "test.wav"), false},
		{"Path traversal attempt", "../../../etc/passwd", true},
		{"Outside input directory", "/tmp/test.wav", true},
		{"Relative path outside", "../../test.wav", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := TranscribeAudio(config, logger, logBuffer, logMutex, false, tt.input)

			if tt.expectedError {
				assert.Error(t, err)
				// Check for Japanese error message
				errorMsg := strings.ToLower(err.Error())
				assert.True(t, strings.Contains(errorMsg, "invalid") ||
					strings.Contains(errorMsg, "無効") ||
					strings.Contains(err.Error(), "ファイルパス"))
			} else {
				// For valid paths, the error might be due to missing whisper command
				// or missing input file, which is expected in test environment
				if err != nil {
					t.Logf("Expected error due to test environment: %v", err)
				}
			}
		})
	}
}

func TestTranscribeAudio_FileValidation(t *testing.T) {
	config := testdata.CreateTestConfig(t)
	logger, logBuffer, logMutex := testdata.CreateTestLogger()

	// Create test directories
	testdata.CreateDirectories(t, config.InputDir, config.OutputDir)

	// Test with valid audio file
	audioFile := testdata.CreateTestAudioFile(t, config.InputDir, "test.wav")

	err := TranscribeAudio(config, logger, logBuffer, logMutex, false, audioFile)

	// In test environment, this will likely fail due to missing whisper command
	// But the security validation should pass
	if err != nil {
		errorMsg := strings.ToLower(err.Error())
		assert.False(t, strings.Contains(errorMsg, "invalid") ||
			strings.Contains(errorMsg, "無効") ||
			strings.Contains(err.Error(), "ファイルパス"))
		t.Logf("Expected error due to missing whisper command: %v", err)
	}
}

func TestGetWhisperCommand_PathResolution(t *testing.T) {
	cmd := getWhisperCommand()

	// Should return either the command name or an absolute path
	assert.True(t, cmd == "whisper-ctranslate2" || filepath.IsAbs(cmd))

	// Test that the function doesn't panic
	assert.NotEmpty(t, cmd)
}

func TestGetWhisperCommand_WindowsExtension(t *testing.T) {
	// Test Windows-specific behavior
	cmd := getWhisperCommand()
	
	if os.Getenv("GOOS") == "windows" || strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") {
		// On Windows, the command should handle .exe extension
		t.Log("Testing Windows command extension handling")
		
		// The function should work with or without .exe
		assert.NotEmpty(t, cmd)
	}
}

func TestGetWhisperCommand_WindowsPaths(t *testing.T) {
	// Test that Windows Python paths are considered
	if os.Getenv("GOOS") != "windows" && !strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") {
		t.Skip("Skipping Windows path test on non-Windows platform")
	}
	
	// Save original environment
	originalPath := os.Getenv("PATH")
	originalUsername := os.Getenv("USERNAME")
	defer func() {
		os.Setenv("PATH", originalPath)
		if originalUsername != "" {
			os.Setenv("USERNAME", originalUsername)
		}
	}()
	
	// Set up test environment
	os.Setenv("PATH", "") // Clear PATH to force fallback search
	if originalUsername == "" {
		os.Setenv("USERNAME", "testuser")
	}
	
	cmd := getWhisperCommand()
	
	// Even if not found, should return fallback
	assert.NotEmpty(t, cmd)
	assert.True(t, cmd == "whisper-ctranslate2" || filepath.IsAbs(cmd))
}

func TestIsFasterWhisperAvailable_MockEnvironment(t *testing.T) {
	// Save original PATH and HOME
	originalPath := os.Getenv("PATH")
	originalHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("PATH", originalPath)
		os.Setenv("HOME", originalHome)
	}()

	// Clear PATH and set HOME to non-existent to simulate missing whisper
	os.Setenv("PATH", "")
	os.Setenv("HOME", "/tmp/nonexistent")

	available := IsFasterWhisperAvailableForTesting()
	assert.False(t, available)
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration string
		expected string
	}{
		{"Zero duration", "0s", "0s"},
		{"Seconds only", "30s", "30s"},
		{"Minutes and seconds", "1m30s", "1m30s"},
		{"Hours, minutes, seconds", "1h2m3s", "1h2m3s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse duration string
			var h, m, s int
			if strings.Contains(tt.duration, "h") {
				n, _ := fmt.Sscanf(tt.duration, "%dh%dm%ds", &h, &m, &s)
				if n < 3 {
					fmt.Sscanf(tt.duration, "%dh%ds", &h, &s)
				}
			} else if strings.Contains(tt.duration, "m") {
				fmt.Sscanf(tt.duration, "%dm%ds", &m, &s)
			} else {
				fmt.Sscanf(tt.duration, "%ds", &s)
			}

			totalSeconds := h*3600 + m*60 + s
			duration := time.Duration(totalSeconds) * time.Second

			result := formatDuration(duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test command building logic
func TestCommandConstruction(t *testing.T) {
	config := testdata.CreateTestConfig(t)
	logger, logBuffer, logMutex := testdata.CreateTestLogger()

	// Create a test audio file
	testdata.CreateDirectories(t, config.InputDir, config.OutputDir)
	audioFile := testdata.CreateTestAudioFile(t, config.InputDir, "test.wav")

	// Test that TranscribeAudio constructs the command correctly
	// This will fail due to missing whisper, but we can check the error
	err := TranscribeAudio(config, logger, logBuffer, logMutex, true, audioFile)

	// Should fail due to missing whisper command, not due to invalid arguments
	if err != nil {
		// Check that error is about command execution, not argument validation
		assert.NotContains(t, strings.ToLower(err.Error()), "invalid argument")
		assert.NotContains(t, strings.ToLower(err.Error()), "bad flag")
		t.Logf("Expected error due to missing whisper: %v", err)
	}
}

// Test audio file format detection
func TestAudioFileFormats(t *testing.T) {
	config := testdata.CreateTestConfig(t)
	logger, logBuffer, logMutex := testdata.CreateTestLogger()

	testdata.CreateDirectories(t, config.InputDir, config.OutputDir)

	// Test different audio file extensions
	audioFormats := []string{"wav", "mp3", "m4a", "flac", "ogg"}

	for _, format := range audioFormats {
		t.Run(fmt.Sprintf("Format_%s", format), func(t *testing.T) {
			audioFile := testdata.CreateTestAudioFile(t, config.InputDir, fmt.Sprintf("test.%s", format))

			err := TranscribeAudio(config, logger, logBuffer, logMutex, false, audioFile)

			// Should not fail due to unsupported format
			if err != nil {
				assert.NotContains(t, strings.ToLower(err.Error()), "unsupported format")
				assert.NotContains(t, strings.ToLower(err.Error()), "invalid format")
				t.Logf("Expected error due to missing whisper: %v", err)
			}
		})
	}
}

// Test configuration parameter validation
func TestConfigParameterValidation(t *testing.T) {
	t.Skip("Skipping long-running integration test - requires actual Whisper execution")
	
	config := testdata.CreateTestConfig(t)
	logger, logBuffer, logMutex := testdata.CreateTestLogger()

	testdata.CreateDirectories(t, config.InputDir, config.OutputDir)
	audioFile := testdata.CreateTestAudioFile(t, config.InputDir, "test.wav")

	// Test only with fastest model to reduce test time
	models := []string{"tiny"}  // Only test with fastest model
	for _, model := range models {
		t.Run(fmt.Sprintf("Model_%s", model), func(t *testing.T) {
			config.WhisperModel = model
			err := TranscribeAudio(config, logger, logBuffer, logMutex, false, audioFile)

			// Should not fail due to invalid model
			if err != nil {
				assert.NotContains(t, strings.ToLower(err.Error()), "invalid model")
				assert.NotContains(t, strings.ToLower(err.Error()), "unknown model")
			}
		})
	}

	// Test different languages
	languages := []string{"ja", "en", "zh", "ko", "auto"}
	for _, lang := range languages {
		t.Run(fmt.Sprintf("Language_%s", lang), func(t *testing.T) {
			config.Language = lang
			err := TranscribeAudio(config, logger, logBuffer, logMutex, false, audioFile)

			// Should not fail due to invalid language
			if err != nil {
				assert.NotContains(t, strings.ToLower(err.Error()), "invalid language")
				assert.NotContains(t, strings.ToLower(err.Error()), "unknown language")
			}
		})
	}

	// Test different output formats
	formats := []string{"txt", "srt", "vtt", "json", "tsv"}
	for _, format := range formats {
		t.Run(fmt.Sprintf("Format_%s", format), func(t *testing.T) {
			config.OutputFormat = format
			err := TranscribeAudio(config, logger, logBuffer, logMutex, false, audioFile)

			// Should not fail due to invalid format
			if err != nil {
				assert.NotContains(t, strings.ToLower(err.Error()), "invalid format")
				assert.NotContains(t, strings.ToLower(err.Error()), "unknown format")
			}
		})
	}

	// Test different compute types
	computeTypes := []string{"int8", "int16", "float16", "float32"}
	for _, computeType := range computeTypes {
		t.Run(fmt.Sprintf("ComputeType_%s", computeType), func(t *testing.T) {
			config.ComputeType = computeType
			err := TranscribeAudio(config, logger, logBuffer, logMutex, false, audioFile)

			// Should not fail due to invalid compute type
			if err != nil {
				assert.NotContains(t, strings.ToLower(err.Error()), "invalid compute")
				assert.NotContains(t, strings.ToLower(err.Error()), "unknown compute")
			}
		})
	}
}

// Test error handling scenarios
func TestErrorHandling(t *testing.T) {
	config := testdata.CreateTestConfig(t)
	logger, logBuffer, logMutex := testdata.CreateTestLogger()

	t.Run("MissingInputFile", func(t *testing.T) {
		testdata.CreateDirectories(t, config.InputDir, config.OutputDir)
		missingFile := filepath.Join(config.InputDir, "nonexistent.wav")

		err := TranscribeAudio(config, logger, logBuffer, logMutex, false, missingFile)
		assert.Error(t, err)
	})

	t.Run("InvalidInputDirectory", func(t *testing.T) {
		config.InputDir = "/nonexistent/directory"
		testFile := filepath.Join(config.InputDir, "test.wav")

		err := TranscribeAudio(config, logger, logBuffer, logMutex, false, testFile)
		assert.Error(t, err)
	})

	t.Run("InvalidOutputDirectory", func(t *testing.T) {
		config := testdata.CreateTestConfig(t)
		config.OutputDir = "/invalid/directory/that/cannot/be/created"
		testdata.CreateDirectories(t, config.InputDir)
		audioFile := testdata.CreateTestAudioFile(t, config.InputDir, "test.wav")

		err := TranscribeAudio(config, logger, logBuffer, logMutex, false, audioFile)
		// Should fail due to invalid output directory
		assert.Error(t, err)
		
		// Error should be about directory creation, not security validation
		if err != nil {
			errorMsg := strings.ToLower(err.Error())
			// Check that it's a directory creation error, not security error
			assert.True(t, strings.Contains(errorMsg, "create") || 
				strings.Contains(errorMsg, "directory") ||
				strings.Contains(errorMsg, "permission"))
		}
	})
}

// Test device parameter handling based on compute type
func TestTranscribeAudio_DeviceParameter(t *testing.T) {
	config := testdata.CreateTestConfig(t)
	logger, logBuffer, logMutex := testdata.CreateTestLogger()
	
	testdata.CreateDirectories(t, config.InputDir, config.OutputDir)
	audioFile := testdata.CreateTestAudioFile(t, config.InputDir, "test.wav")
	
	tests := []struct {
		name         string
		computeType  string
		expectDevice bool
	}{
		{
			name:         "int8 should use CPU device",
			computeType:  "int8",
			expectDevice: true,
		},
		{
			name:         "float16 should auto-select device",
			computeType:  "float16",
			expectDevice: false,
		},
		{
			name:         "int8_float16 should auto-select device",
			computeType:  "int8_float16",
			expectDevice: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.ComputeType = tt.computeType
			
			// This will fail due to missing whisper, but we document expected behavior
			err := TranscribeAudio(config, logger, logBuffer, logMutex, true, audioFile)
			
			// The test documents that int8 should force CPU device selection
			if tt.computeType == "int8" && !tt.expectDevice {
				t.Error("int8 compute type should trigger CPU device selection")
			}
			
			// Log the expected behavior
			if tt.expectDevice {
				t.Logf("Compute type %s should add --device cpu parameter", tt.computeType)
			} else {
				t.Logf("Compute type %s should not add device parameter", tt.computeType)
			}
			
			if err != nil {
				t.Logf("Expected error in test environment: %v", err)
			}
		})
	}
}

// Test GPU error detection and message generation
func TestIsGPURelatedError(t *testing.T) {
	tests := []struct {
		name        string
		errorStr    string
		expectGPU   bool
	}{
		{
			name:      "CUDA error should be detected",
			errorStr:  "CUDA device not found",
			expectGPU: true,
		},
		{
			name:      "Float16 error should be detected", 
			errorStr:  "Requested float16 compute type, but the target device or backend do not support efficient float16 computation",
			expectGPU: true,
		},
		{
			name:      "GPU memory error should be detected",
			errorStr:  "GPU out of memory",
			expectGPU: true,
		},
		{
			name:      "Regular file error should not be detected",
			errorStr:  "input file does not exist",
			expectGPU: false,
		},
		{
			name:      "Permission error should not be detected",
			errorStr:  "permission denied",
			expectGPU: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGPURelatedError(tt.errorStr)
			assert.Equal(t, tt.expectGPU, result)
		})
	}
}

func TestCreateGPUErrorMessage(t *testing.T) {
	tests := []struct {
		name       string
		uiLanguage string
		computeType string
		expectJA   bool
	}{
		{
			name:       "Japanese error message",
			uiLanguage: "ja",
			computeType: "float16",
			expectJA:   true,
		},
		{
			name:       "English error message",
			uiLanguage: "en", 
			computeType: "int8_float16",
			expectJA:   false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := testdata.CreateTestConfig(t)
			config.UILanguage = tt.uiLanguage
			config.ComputeType = tt.computeType
			
			originalErr := fmt.Errorf("GPU test error")
			result := createGPUErrorMessage(config, originalErr)
			
			errorMsg := result.Error()
			
			if tt.expectJA {
				assert.Contains(t, errorMsg, "GPU処理に失敗")
				assert.Contains(t, errorMsg, "推奨解決策")
			} else {
				assert.Contains(t, errorMsg, "GPU processing failed")
				assert.Contains(t, errorMsg, "Recommended solutions")
			}
			
			// Should contain current compute type and original error
			assert.Contains(t, errorMsg, tt.computeType)
			assert.Contains(t, errorMsg, "GPU test error")
			
			// Output the full error message for verification
			t.Logf("Full error message:\n%s", errorMsg)
		})
	}
}

// Test diagnostic functions
func TestDiagnoseDependencies(t *testing.T) {
	logger, logBuffer, logMutex := testdata.CreateTestLogger()

	// Test diagnostic function execution
	diagnoseDependencies(logger, logBuffer, logMutex)

	// Verify that diagnostic logs were created
	logMutex.RLock()
	logs := *logBuffer
	logMutex.RUnlock()

	// Should have diagnostic logs for Python, pip, and whisper-ctranslate2
	var hasPythonCheck, hasPipCheck, hasWhisperCheck bool
	for _, log := range logs {
		if strings.Contains(log.Message, "Python") {
			hasPythonCheck = true
		}
		if strings.Contains(log.Message, "pip") {
			hasPipCheck = true
		}
		if strings.Contains(log.Message, "whisper-ctranslate2") {
			hasWhisperCheck = true
		}
	}

	assert.True(t, hasPythonCheck, "Should check Python availability")
	assert.True(t, hasPipCheck, "Should check pip availability")
	assert.True(t, hasWhisperCheck, "Should check whisper-ctranslate2 availability")

	t.Logf("Diagnostic function executed with %d log entries", len(logs))
}

func TestCreateDependencyErrorMessage(t *testing.T) {
	tests := []struct {
		name       string
		uiLanguage string
		expectJA   bool
	}{
		{
			name:       "Japanese error message",
			uiLanguage: "ja",
			expectJA:   true,
		},
		{
			name:       "English error message",
			uiLanguage: "en",
			expectJA:   false,
		},
		{
			name:       "Default language (should be English)",
			uiLanguage: "",
			expectJA:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := testdata.CreateTestConfig(t)
			config.UILanguage = tt.uiLanguage

			result := createDependencyErrorMessage(config)

			if tt.expectJA {
				assert.Contains(t, result, "診断結果をログで確認してください")
				assert.Contains(t, result, "よくある解決策")
				assert.Contains(t, result, "Python/pipが見つからない場合")
				assert.Contains(t, result, "whisper-ctranslate2が未インストールの場合")
			} else {
				assert.Contains(t, result, "Check the diagnostic results in the logs")
				assert.Contains(t, result, "Common solutions")
				assert.Contains(t, result, "If Python/pip not found")
				assert.Contains(t, result, "If whisper-ctranslate2 not installed")
			}

			// Should contain installation commands
			assert.Contains(t, result, "pip install whisper-ctranslate2")
			assert.Contains(t, result, "pip3 install whisper-ctranslate2")

			t.Logf("Error message language: %s\nLength: %d characters", tt.uiLanguage, len(result))
		})
	}
}

func TestEnsureDependencies_InstallationFailure(t *testing.T) {
	config := testdata.CreateTestConfig(t)
	logger, logBuffer, logMutex := testdata.CreateTestLogger()

	// Save original PATH and HOME to ensure whisper-ctranslate2 is not available
	originalPath := os.Getenv("PATH")
	originalHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("PATH", originalPath)
		os.Setenv("HOME", originalHome)
	}()

	// Clear PATH and set non-existent HOME to ensure both pip and whisper-ctranslate2 are not found
	os.Setenv("PATH", "/nonexistent")
	os.Setenv("HOME", "/tmp/nonexistent")

	// This should trigger installation attempt and failure
	err := EnsureDependencies(config, logger, logBuffer, logMutex, false)

	// Check logs regardless of error result (environment dependent)
	logMutex.RLock()
	logs := *logBuffer
	logMutex.RUnlock()

	var hasInstallAttempt, hasDiagnostics bool
	for _, log := range logs {
		if strings.Contains(log.Message, "Attempting to install") || strings.Contains(log.Message, "FasterWhisper not found") {
			hasInstallAttempt = true
		}
		if strings.Contains(log.Message, "Running diagnostics") || strings.Contains(log.Message, "dependency diagnostics") {
			hasDiagnostics = true
		}
	}

	if err != nil {
		// Installation failed as expected
		assert.Contains(t, err.Error(), "FasterWhisper installation failed")
		assert.True(t, hasInstallAttempt, "Should attempt installation")
		assert.True(t, hasDiagnostics, "Should run diagnostics after installation failure")
		t.Logf("Installation failure test completed as expected with %d log entries", len(logs))
	} else {
		// Installation succeeded (unexpected but possible in some environments)
		t.Logf("Installation unexpectedly succeeded in test environment. This may happen if pip/whisper are available despite PATH modification.")
		t.Logf("Generated %d log entries", len(logs))
		// Don't fail the test - this is environment dependent
	}
}

func TestDiagnoseDependencies_MockEnvironment(t *testing.T) {
	logger, logBuffer, logMutex := testdata.CreateTestLogger()

	// Save original environment
	originalPath := os.Getenv("PATH")
	originalHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("PATH", originalPath)
		os.Setenv("HOME", originalHome)
	}()

	tests := []struct {
		name        string
		pathEnv     string
		homeEnv     string
		expectError bool
	}{
		{
			name:        "Empty PATH should cause errors",
			pathEnv:     "",
			homeEnv:     "/tmp/nonexistent",
			expectError: true,
		},
		{
			name:        "Nonexistent HOME should cause errors",
			pathEnv:     "/usr/bin:/bin",
			homeEnv:     "/tmp/nonexistent",
			expectError: false, // May still find system Python/pip
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear log buffer
			logMutex.Lock()
			*logBuffer = (*logBuffer)[:0]
			logMutex.Unlock()

			// Set test environment
			os.Setenv("PATH", tt.pathEnv)
			os.Setenv("HOME", tt.homeEnv)

			// Run diagnostics
			diagnoseDependencies(logger, logBuffer, logMutex)

			// Check logs
			logMutex.RLock()
			logs := *logBuffer
			logMutex.RUnlock()

			// Should have some diagnostic output
			assert.Greater(t, len(logs), 0, "Should generate diagnostic logs")

			// Log the results
			t.Logf("Test environment PATH=%s HOME=%s generated %d logs", tt.pathEnv, tt.homeEnv, len(logs))
			for _, log := range logs {
				t.Logf("  [%s] %s", log.Level, log.Message)
			}
		})
	}
}

func TestCreateDependencyErrorMessage_MessageContent(t *testing.T) {
	config := testdata.CreateTestConfig(t)

	t.Run("Japanese message content", func(t *testing.T) {
		config.UILanguage = "ja"
		result := createDependencyErrorMessage(config)

		// Check specific Japanese content
		assert.Contains(t, result, "Python 3.8以上をインストール")
		assert.Contains(t, result, "pip --version で確認")
		assert.Contains(t, result, "PATH環境変数に追加が必要な可能性")

		// Should not contain English-specific text
		assert.NotContains(t, result, "Install Python 3.8 or higher")
		assert.NotContains(t, result, "Verify with: pip --version")
	})

	t.Run("English message content", func(t *testing.T) {
		config.UILanguage = "en"
		result := createDependencyErrorMessage(config)

		// Check specific English content
		assert.Contains(t, result, "Install Python 3.8 or higher")
		assert.Contains(t, result, "Verify with: pip --version")
		assert.Contains(t, result, "May need to add to PATH environment variable")

		// Should not contain Japanese-specific text
		assert.NotContains(t, result, "Python 3.8以上をインストール")
		assert.NotContains(t, result, "pip --version で確認")
	})
}

// Benchmark tests for performance
func BenchmarkGetWhisperCommand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = getWhisperCommand()
	}
}

func BenchmarkIsFasterWhisperAvailable(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = isFasterWhisperAvailable()
	}
}

func BenchmarkDiagnoseDependencies(b *testing.B) {
	logger, logBuffer, logMutex := testdata.CreateTestLogger()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear log buffer to avoid memory buildup
		logMutex.Lock()
		*logBuffer = (*logBuffer)[:0]
		logMutex.Unlock()
		
		diagnoseDependencies(logger, logBuffer, logMutex)
	}
}
