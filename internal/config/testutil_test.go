package config

import (
	"bufio"
	"io"
	"strings"
	"testing"
)

// MockReader simulates user input for testing interactive functions
type MockReader struct {
	inputs []string
	index  int
}

func NewMockReader(inputs ...string) *MockReader {
	return &MockReader{
		inputs: inputs,
		index:  0,
	}
}

func (m *MockReader) ReadString(delim byte) (string, error) {
	if m.index >= len(m.inputs) {
		return "", io.EOF
	}
	input := m.inputs[m.index]
	m.index++
	if !strings.HasSuffix(input, string(delim)) {
		input += string(delim)
	}
	return input, nil
}

// Helper function to create a bufio.Reader from mock inputs
func createMockReader(inputs ...string) *bufio.Reader {
	combined := strings.Join(inputs, "\n")
	if !strings.HasSuffix(combined, "\n") {
		combined += "\n"
	}
	return bufio.NewReader(strings.NewReader(combined))
}

// Test helper to validate config fields
func assertConfigEquals(t *testing.T, expected, actual *Config) {
	t.Helper()
	
	if expected.WhisperModel != actual.WhisperModel {
		t.Errorf("WhisperModel: expected %s, got %s", expected.WhisperModel, actual.WhisperModel)
	}
	if expected.Language != actual.Language {
		t.Errorf("Language: expected %s, got %s", expected.Language, actual.Language)
	}
	if expected.UILanguage != actual.UILanguage {
		t.Errorf("UILanguage: expected %s, got %s", expected.UILanguage, actual.UILanguage)
	}
	if expected.ScanIntervalMinutes != actual.ScanIntervalMinutes {
		t.Errorf("ScanIntervalMinutes: expected %d, got %d", expected.ScanIntervalMinutes, actual.ScanIntervalMinutes)
	}
	if expected.MaxCpuPercent != actual.MaxCpuPercent {
		t.Errorf("MaxCpuPercent: expected %d, got %d", expected.MaxCpuPercent, actual.MaxCpuPercent)
	}
	if expected.ComputeType != actual.ComputeType {
		t.Errorf("ComputeType: expected %s, got %s", expected.ComputeType, actual.ComputeType)
	}
	if expected.UseColors != actual.UseColors {
		t.Errorf("UseColors: expected %t, got %t", expected.UseColors, actual.UseColors)
	}
	if expected.OutputFormat != actual.OutputFormat {
		t.Errorf("OutputFormat: expected %s, got %s", expected.OutputFormat, actual.OutputFormat)
	}
	if expected.InputDir != actual.InputDir {
		t.Errorf("InputDir: expected %s, got %s", expected.InputDir, actual.InputDir)
	}
	if expected.OutputDir != actual.OutputDir {
		t.Errorf("OutputDir: expected %s, got %s", expected.OutputDir, actual.OutputDir)
	}
	if expected.ArchiveDir != actual.ArchiveDir {
		t.Errorf("ArchiveDir: expected %s, got %s", expected.ArchiveDir, actual.ArchiveDir)
	}
	if expected.LLMSummaryEnabled != actual.LLMSummaryEnabled {
		t.Errorf("LLMSummaryEnabled: expected %t, got %t", expected.LLMSummaryEnabled, actual.LLMSummaryEnabled)
	}
	if expected.LLMAPIProvider != actual.LLMAPIProvider {
		t.Errorf("LLMAPIProvider: expected %s, got %s", expected.LLMAPIProvider, actual.LLMAPIProvider)
	}
	if expected.LLMAPIKey != actual.LLMAPIKey {
		t.Errorf("LLMAPIKey: expected %s, got %s", expected.LLMAPIKey, actual.LLMAPIKey)
	}
	if expected.LLMModel != actual.LLMModel {
		t.Errorf("LLMModel: expected %s, got %s", expected.LLMModel, actual.LLMModel)
	}
	if expected.LLMMaxTokens != actual.LLMMaxTokens {
		t.Errorf("LLMMaxTokens: expected %d, got %d", expected.LLMMaxTokens, actual.LLMMaxTokens)
	}
	if expected.SummaryLanguage != actual.SummaryLanguage {
		t.Errorf("SummaryLanguage: expected %s, got %s", expected.SummaryLanguage, actual.SummaryLanguage)
	}
	if expected.RecordingDeviceID != actual.RecordingDeviceID {
		t.Errorf("RecordingDeviceID: expected %d, got %d", expected.RecordingDeviceID, actual.RecordingDeviceID)
	}
	if expected.RecordingDeviceName != actual.RecordingDeviceName {
		t.Errorf("RecordingDeviceName: expected %s, got %s", expected.RecordingDeviceName, actual.RecordingDeviceName)
	}
	if expected.RecordingMaxHours != actual.RecordingMaxHours {
		t.Errorf("RecordingMaxHours: expected %d, got %d", expected.RecordingMaxHours, actual.RecordingMaxHours)
	}
	if expected.RecordingMaxFileMB != actual.RecordingMaxFileMB {
		t.Errorf("RecordingMaxFileMB: expected %d, got %d", expected.RecordingMaxFileMB, actual.RecordingMaxFileMB)
	}
}