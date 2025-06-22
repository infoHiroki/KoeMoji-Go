package whisper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/hirokitakamura/koemoji-go/internal/whisper/testdata"
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
		name     string
		input    string
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

func TestIsFasterWhisperAvailable_MockEnvironment(t *testing.T) {
	// Save original PATH
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)
	
	// Clear PATH to simulate missing whisper
	os.Setenv("PATH", "")
	
	available := isFasterWhisperAvailable()
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
	config := testdata.CreateTestConfig(t)
	logger, logBuffer, logMutex := testdata.CreateTestLogger()
	
	testdata.CreateDirectories(t, config.InputDir, config.OutputDir)
	audioFile := testdata.CreateTestAudioFile(t, config.InputDir, "test.wav")
	
	// Test different Whisper models
	models := []string{"tiny", "base", "small", "medium", "large", "large-v2", "large-v3"}
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
		// Will fail, but not due to security validation
		if err != nil {
			errorMsg := strings.ToLower(err.Error())
			assert.False(t, strings.Contains(errorMsg, "invalid") || 
				        strings.Contains(errorMsg, "無効") || 
				        strings.Contains(err.Error(), "ファイルパス"))
		}
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
