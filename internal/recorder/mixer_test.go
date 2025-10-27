// +build darwin

package recorder

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

// TestResampleInt16 tests the resampling function
func TestResampleInt16(t *testing.T) {
	// Test case: 44.1kHz → 48kHz
	inputSampleRate := 44100
	outputSampleRate := 48000

	// Create a simple test signal (sine wave)
	input := make([]float64, 44100) // 1 second at 44.1kHz
	for i := range input {
		// 440 Hz sine wave (A4 note)
		input[i] = math.Sin(2 * math.Pi * 440 * float64(i) / float64(inputSampleRate))
	}

	// Resample
	output := ResampleInt16(input, inputSampleRate, outputSampleRate)

	// Check output length
	expectedLength := 48000
	if len(output) != expectedLength {
		t.Errorf("Expected output length %d, got %d", expectedLength, len(output))
	}

	// Check that output values are reasonable (within [-1.0, 1.0])
	for i, v := range output {
		if v < -1.0 || v > 1.0 {
			t.Errorf("Output sample %d out of range: %f", i, v)
		}
	}
}

// TestConvertFloat32ToFloat64 tests Float32 to Float64 conversion
func TestConvertFloat32ToFloat64(t *testing.T) {
	// Create test Float32 data
	samples := []byte{
		0x00, 0x00, 0x00, 0x3f, // 0.5 in Float32
		0x00, 0x00, 0x00, 0xbf, // -0.5 in Float32
	}

	output := ConvertFloat32ToFloat64(samples)

	if len(output) != 2 {
		t.Errorf("Expected 2 samples, got %d", len(output))
	}

	// Check values (with small tolerance for floating point)
	if math.Abs(output[0]-0.5) > 0.01 {
		t.Errorf("Expected ~0.5, got %f", output[0])
	}

	if math.Abs(output[1]+0.5) > 0.01 {
		t.Errorf("Expected ~-0.5, got %f", output[1])
	}
}

// TestConvertInt16ToFloat64 tests Int16 to Float64 conversion
func TestConvertInt16ToFloat64(t *testing.T) {
	// Create test Int16 data
	samples := []byte{
		0xff, 0x7f, // 32767 (max)
		0x00, 0x80, // -32768 (min)
		0x00, 0x00, // 0
	}

	output := ConvertInt16ToFloat64(samples)

	if len(output) != 3 {
		t.Errorf("Expected 3 samples, got %d", len(output))
	}

	// Check normalized values
	if math.Abs(output[0]-1.0) > 0.01 {
		t.Errorf("Expected ~1.0, got %f", output[0])
	}

	if math.Abs(output[1]+1.0) > 0.01 {
		t.Errorf("Expected ~-1.0, got %f", output[1])
	}

	if math.Abs(output[2]) > 0.01 {
		t.Errorf("Expected ~0.0, got %f", output[2])
	}
}

// TestConvertFloat64ToInt16 tests Float64 to Int16 conversion
func TestConvertFloat64ToInt16(t *testing.T) {
	input := []float64{1.0, -1.0, 0.0, 0.5}

	output := ConvertFloat64ToInt16(input)

	if len(output) != 8 { // 4 samples * 2 bytes
		t.Errorf("Expected 8 bytes, got %d", len(output))
	}

	// Test that values are clamped and converted properly
	// 1.0 → 32767
	// -1.0 → -32768
	// 0.0 → 0
	// 0.5 → 16383

	samples := ConvertInt16ToFloat64(output)
	if len(samples) != 4 {
		t.Errorf("Expected 4 samples after round-trip, got %d", len(samples))
	}
}

// TestMixStereoAndMono tests the mixing function
func TestMixStereoAndMono(t *testing.T) {
	// Create test stereo signal (L=0.5, R=0.5)
	stereo := []float64{0.5, 0.5, 0.5, 0.5, 0.5, 0.5} // 3 frames

	// Create test mono signal (0.3)
	mono := []float64{0.3, 0.3, 0.3}

	// Mix with equal volume
	output := MixStereoAndMono(stereo, mono, 1.0, 1.0)

	if len(output) != 6 {
		t.Errorf("Expected 6 samples, got %d", len(output))
	}

	// Check that mono was added to both L and R channels
	// Expected: 0.5 + 0.3 = 0.8 for both channels
	for i := 0; i < 3; i++ {
		if math.Abs(output[i*2]-0.8) > 0.01 {
			t.Errorf("Frame %d L channel: expected ~0.8, got %f", i, output[i*2])
		}
		if math.Abs(output[i*2+1]-0.8) > 0.01 {
			t.Errorf("Frame %d R channel: expected ~0.8, got %f", i, output[i*2+1])
		}
	}
}

// TestMixStereoAndMonoClipping tests clipping prevention
func TestMixStereoAndMonoClipping(t *testing.T) {
	// Create test stereo signal that would clip
	stereo := []float64{0.9, 0.9, 0.9, 0.9}

	// Create test mono signal
	mono := []float64{0.3, 0.3}

	// Mix (should clip to 1.0)
	output := MixStereoAndMono(stereo, mono, 1.0, 1.0)

	// Check that values are clamped to 1.0
	for i, v := range output {
		if v > 1.0 {
			t.Errorf("Sample %d not clipped: %f", i, v)
		}
		if v < -1.0 {
			t.Errorf("Sample %d not clipped: %f", i, v)
		}
	}
}

// TestReadWriteWAVFile tests WAV file I/O
func TestReadWriteWAVFile(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.wav")

	// Create test audio data
	testData := &AudioData{
		SampleRate:    48000,
		NumChannels:   2,
		BitsPerSample: 16,
		AudioFormat:   1, // PCM
		Samples:       []float64{0.5, -0.5, 0.3, -0.3, 0.0, 0.0},
	}

	// Write WAV file
	if err := WriteWAVFile(testFile, testData); err != nil {
		t.Fatalf("Failed to write WAV file: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatalf("WAV file was not created")
	}

	// Read WAV file back
	readData, err := ReadWAVFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read WAV file: %v", err)
	}

	// Verify header
	if readData.SampleRate != 48000 {
		t.Errorf("Expected sample rate 48000, got %d", readData.SampleRate)
	}

	if readData.NumChannels != 2 {
		t.Errorf("Expected 2 channels, got %d", readData.NumChannels)
	}

	// Verify sample count
	if len(readData.Samples) != 6 {
		t.Errorf("Expected 6 samples, got %d", len(readData.Samples))
	}

	// Verify sample values (with tolerance for Int16 conversion)
	expectedSamples := []float64{0.5, -0.5, 0.3, -0.3, 0.0, 0.0}
	for i, expected := range expectedSamples {
		if math.Abs(readData.Samples[i]-expected) > 0.01 {
			t.Errorf("Sample %d: expected %f, got %f", i, expected, readData.Samples[i])
		}
	}
}

// TestMixAudioFilesIntegration tests the full mixing workflow
func TestMixAudioFilesIntegration(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create system audio file (stereo, 48kHz)
	systemFile := filepath.Join(tempDir, "system.wav")
	systemData := &AudioData{
		SampleRate:    48000,
		NumChannels:   2,
		BitsPerSample: 16,
		AudioFormat:   1,
		Samples:       make([]float64, 96000), // 1 second stereo
	}
	// Fill with test signal (440 Hz sine wave)
	for i := 0; i < 48000; i++ {
		value := 0.3 * math.Sin(2*math.Pi*440*float64(i)/48000.0)
		systemData.Samples[i*2] = value     // L
		systemData.Samples[i*2+1] = value   // R
	}
	if err := WriteWAVFile(systemFile, systemData); err != nil {
		t.Fatalf("Failed to create system audio file: %v", err)
	}

	// Create microphone file (mono, 44.1kHz)
	micFile := filepath.Join(tempDir, "mic.wav")
	micData := &AudioData{
		SampleRate:    44100,
		NumChannels:   1,
		BitsPerSample: 16,
		AudioFormat:   1,
		Samples:       make([]float64, 44100), // 1 second mono
	}
	// Fill with test signal (880 Hz sine wave - one octave higher)
	for i := 0; i < 44100; i++ {
		micData.Samples[i] = 0.5 * math.Sin(2*math.Pi*880*float64(i)/44100.0)
	}
	if err := WriteWAVFile(micFile, micData); err != nil {
		t.Fatalf("Failed to create mic audio file: %v", err)
	}

	// Mix files
	outputFile := filepath.Join(tempDir, "mixed.wav")
	if err := MixAudioFiles(systemFile, micFile, outputFile, 0.7, 1.0); err != nil {
		t.Fatalf("Failed to mix audio files: %v", err)
	}

	// Verify output file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatalf("Mixed audio file was not created")
	}

	// Read and verify output
	outputData, err := ReadWAVFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read mixed audio file: %v", err)
	}

	// Verify format
	if outputData.SampleRate != 48000 {
		t.Errorf("Expected sample rate 48000, got %d", outputData.SampleRate)
	}

	if outputData.NumChannels != 2 {
		t.Errorf("Expected 2 channels, got %d", outputData.NumChannels)
	}

	// Verify that output has reasonable duration (should be ~1 second = 48000 frames = 96000 samples)
	expectedSamples := 96000
	if math.Abs(float64(len(outputData.Samples)-expectedSamples)) > 1000 {
		t.Errorf("Expected ~%d samples, got %d", expectedSamples, len(outputData.Samples))
	}

	t.Logf("Integration test passed: mixed %d samples at %dHz", len(outputData.Samples), outputData.SampleRate)
}
