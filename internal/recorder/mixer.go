// +build darwin

package recorder

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
)

// WAV file format structures
type wavHeader struct {
	// RIFF header
	ChunkID       [4]byte // "RIFF"
	ChunkSize     uint32
	Format        [4]byte // "WAVE"

	// fmt sub-chunk
	Subchunk1ID   [4]byte // "fmt "
	Subchunk1Size uint32  // 16 for PCM
	AudioFormat   uint16  // 1 for PCM, 3 for IEEE float
	NumChannels   uint16  // 1 = Mono, 2 = Stereo
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16

	// data sub-chunk
	Subchunk2ID   [4]byte // "data"
	Subchunk2Size uint32  // Size of audio data
}

// AudioData represents decoded audio samples
type AudioData struct {
	SampleRate    int
	NumChannels   int
	BitsPerSample int
	AudioFormat   int     // 1 = PCM Int, 3 = IEEE Float
	Samples       []float64 // Normalized to [-1.0, 1.0]
}

// ResampleInt16 resamples audio from one sample rate to another using linear interpolation
// Input and output are in float64 format normalized to [-1.0, 1.0]
func ResampleInt16(input []float64, fromRate, toRate int) []float64 {
	if fromRate == toRate {
		return input
	}

	ratio := float64(fromRate) / float64(toRate)
	length := int(float64(len(input)) / ratio)
	output := make([]float64, length)

	for i := 0; i < length; i++ {
		srcPos := float64(i) * ratio
		srcIdx := int(srcPos)

		if srcIdx+1 < len(input) {
			// Linear interpolation
			frac := srcPos - float64(srcIdx)
			output[i] = input[srcIdx]*(1-frac) + input[srcIdx+1]*frac
		} else if srcIdx < len(input) {
			output[i] = input[srcIdx]
		}
	}

	return output
}

// ConvertFloat32ToFloat64 converts Float32 samples to normalized Float64
func ConvertFloat32ToFloat64(samples []byte) []float64 {
	count := len(samples) / 4 // 4 bytes per Float32
	output := make([]float64, count)

	for i := 0; i < count; i++ {
		bits := binary.LittleEndian.Uint32(samples[i*4 : i*4+4])
		f := math.Float32frombits(bits)
		output[i] = float64(f)
	}

	return output
}

// ConvertInt16ToFloat64 converts Int16 samples to normalized Float64
func ConvertInt16ToFloat64(samples []byte) []float64 {
	count := len(samples) / 2 // 2 bytes per Int16
	output := make([]float64, count)

	for i := 0; i < count; i++ {
		s := int16(binary.LittleEndian.Uint16(samples[i*2 : i*2+2]))
		output[i] = float64(s) / 32768.0 // Normalize to [-1.0, 1.0]
	}

	return output
}

// ConvertFloat64ToInt16 converts normalized Float64 samples to Int16 bytes
func ConvertFloat64ToInt16(samples []float64) []byte {
	output := make([]byte, len(samples)*2)

	for i, f := range samples {
		// Clamp to [-1.0, 1.0]
		if f > 1.0 {
			f = 1.0
		} else if f < -1.0 {
			f = -1.0
		}

		// Convert to Int16
		s := int16(f * 32767.0)
		binary.LittleEndian.PutUint16(output[i*2:i*2+2], uint16(s))
	}

	return output
}

// MixStereoAndMono mixes stereo and mono audio
// stereo: [L0, R0, L1, R1, ...]
// mono: [M0, M1, M2, ...]
// Returns: [L0+M0, R0+M0, L1+M1, R1+M1, ...]
func MixStereoAndMono(stereo, mono []float64, stereoVol, monoVol float64) []float64 {
	// Ensure mono is half the length of stereo (since stereo has 2 channels)
	monoLen := len(stereo) / 2
	if len(mono) > monoLen {
		mono = mono[:monoLen]
	} else if len(mono) < monoLen {
		// Pad mono with zeros
		padded := make([]float64, monoLen)
		copy(padded, mono)
		mono = padded
	}

	output := make([]float64, len(stereo))

	for i := 0; i < monoLen; i++ {
		// Apply volume and mix
		monoSample := mono[i] * monoVol

		// Mix with left channel
		output[i*2] = stereo[i*2]*stereoVol + monoSample

		// Mix with right channel
		output[i*2+1] = stereo[i*2+1]*stereoVol + monoSample

		// Clipping prevention (soft clipping)
		output[i*2] = clamp(output[i*2])
		output[i*2+1] = clamp(output[i*2+1])
	}

	return output
}

// clamp applies soft clipping to prevent distortion
func clamp(sample float64) float64 {
	if sample > 1.0 {
		return 1.0
	} else if sample < -1.0 {
		return -1.0
	}
	return sample
}

// ReadWAVFile reads a WAV file and returns audio data
// Supports WAV files with extra chunks (FLLR, LIST, JUNK, etc.)
func ReadWAVFile(filename string) (*AudioData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read RIFF header (12 bytes)
	var riffHeader struct {
		ChunkID   [4]byte // "RIFF"
		ChunkSize uint32
		Format    [4]byte // "WAVE"
	}
	if err := binary.Read(file, binary.LittleEndian, &riffHeader); err != nil {
		return nil, fmt.Errorf("failed to read RIFF header: %w", err)
	}

	// Verify RIFF header
	if string(riffHeader.ChunkID[:]) != "RIFF" || string(riffHeader.Format[:]) != "WAVE" {
		return nil, fmt.Errorf("not a valid WAV file")
	}

	// Read chunks until we find fmt and data
	var fmtChunk struct {
		AudioFormat   uint16
		NumChannels   uint16
		SampleRate    uint32
		ByteRate      uint32
		BlockAlign    uint16
		BitsPerSample uint16
	}
	var dataSize uint32
	var audioBytes []byte
	foundFmt := false
	foundData := false

	for !foundFmt || !foundData {
		// Read chunk header
		var chunkID [4]byte
		var chunkSize uint32
		if err := binary.Read(file, binary.LittleEndian, &chunkID); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read chunk ID: %w", err)
		}
		if err := binary.Read(file, binary.LittleEndian, &chunkSize); err != nil {
			return nil, fmt.Errorf("failed to read chunk size: %w", err)
		}

		chunkName := string(chunkID[:])

		switch chunkName {
		case "fmt ":
			// Read fmt chunk data (we expect at least 16 bytes for PCM)
			if err := binary.Read(file, binary.LittleEndian, &fmtChunk); err != nil {
				return nil, fmt.Errorf("failed to read fmt chunk: %w", err)
			}
			// Skip any extra bytes in fmt chunk (e.g., for non-PCM formats)
			if chunkSize > 16 {
				extraBytes := int(chunkSize) - 16
				if _, err := file.Seek(int64(extraBytes), io.SeekCurrent); err != nil {
					return nil, fmt.Errorf("failed to skip fmt extra bytes: %w", err)
				}
			}
			foundFmt = true

		case "data":
			// Read audio data
			dataSize = chunkSize
			audioBytes = make([]byte, dataSize)
			if _, err := io.ReadFull(file, audioBytes); err != nil {
				return nil, fmt.Errorf("failed to read audio data: %w", err)
			}
			foundData = true

		default:
			// Skip unknown chunks (FLLR, LIST, JUNK, etc.)
			if _, err := file.Seek(int64(chunkSize), io.SeekCurrent); err != nil {
				return nil, fmt.Errorf("failed to skip chunk %s: %w", chunkName, err)
			}
		}
	}

	if !foundFmt {
		return nil, fmt.Errorf("fmt chunk not found")
	}
	if !foundData {
		return nil, fmt.Errorf("data chunk not found")
	}

	// Convert to normalized Float64
	var samples []float64
	if fmtChunk.AudioFormat == 1 {
		// PCM Int16
		samples = ConvertInt16ToFloat64(audioBytes)
	} else if fmtChunk.AudioFormat == 3 {
		// IEEE Float32
		samples = ConvertFloat32ToFloat64(audioBytes)
	} else {
		return nil, fmt.Errorf("unsupported audio format: %d", fmtChunk.AudioFormat)
	}

	return &AudioData{
		SampleRate:    int(fmtChunk.SampleRate),
		NumChannels:   int(fmtChunk.NumChannels),
		BitsPerSample: int(fmtChunk.BitsPerSample),
		AudioFormat:   int(fmtChunk.AudioFormat),
		Samples:       samples,
	}, nil
}

// WriteWAVFile writes audio data to a WAV file (Int16 PCM format)
func WriteWAVFile(filename string, data *AudioData) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Convert samples to Int16
	audioBytes := ConvertFloat64ToInt16(data.Samples)

	// Create WAV header
	header := wavHeader{
		ChunkID:       [4]byte{'R', 'I', 'F', 'F'},
		ChunkSize:     uint32(36 + len(audioBytes)),
		Format:        [4]byte{'W', 'A', 'V', 'E'},
		Subchunk1ID:   [4]byte{'f', 'm', 't', ' '},
		Subchunk1Size: 16,
		AudioFormat:   1, // PCM
		NumChannels:   uint16(data.NumChannels),
		SampleRate:    uint32(data.SampleRate),
		BitsPerSample: 16,
		Subchunk2ID:   [4]byte{'d', 'a', 't', 'a'},
		Subchunk2Size: uint32(len(audioBytes)),
	}

	header.ByteRate = header.SampleRate * uint32(header.NumChannels) * uint32(header.BitsPerSample) / 8
	header.BlockAlign = header.NumChannels * header.BitsPerSample / 8

	// Write header
	if err := binary.Write(file, binary.LittleEndian, &header); err != nil {
		return fmt.Errorf("failed to write WAV header: %w", err)
	}

	// Write audio data
	if _, err := file.Write(audioBytes); err != nil {
		return fmt.Errorf("failed to write audio data: %w", err)
	}

	return nil
}

// MixAudioFiles mixes two WAV files (system audio + microphone) into one
// systemFile: Stereo Float32 (48kHz)
// micFile: Mono Int16 (44.1kHz)
// outputFile: Stereo Int16 (48kHz)
// systemVol: Volume multiplier for system audio (e.g., 0.7)
// micVol: Volume multiplier for microphone (e.g., 1.0)
func MixAudioFiles(systemFile, micFile, outputFile string, systemVol, micVol float64) error {
	// Read system audio (stereo)
	systemData, err := ReadWAVFile(systemFile)
	if err != nil {
		return fmt.Errorf("failed to read system audio: %w", err)
	}

	// Read microphone audio (mono)
	micData, err := ReadWAVFile(micFile)
	if err != nil {
		return fmt.Errorf("failed to read microphone audio: %w", err)
	}

	// Resample microphone to 48kHz if needed
	var micResampled []float64
	if micData.SampleRate != 48000 {
		micResampled = ResampleInt16(micData.Samples, micData.SampleRate, 48000)
	} else {
		micResampled = micData.Samples
	}

	// Mix stereo and mono
	mixed := MixStereoAndMono(systemData.Samples, micResampled, systemVol, micVol)

	// Create output audio data
	outputData := &AudioData{
		SampleRate:    48000,
		NumChannels:   2,
		BitsPerSample: 16,
		AudioFormat:   1, // PCM
		Samples:       mixed,
	}

	// Write output file
	if err := WriteWAVFile(outputFile, outputData); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
