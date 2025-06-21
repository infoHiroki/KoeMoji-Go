package recorder

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
)

const (
	SampleRate = 44100
	Channels   = 1
	BufferSize = 4096
	// Phase 1: Memory-efficient recording with buffering
	MemoryBufferSize = 88200 * 5  // 5 seconds worth of samples (44.1kHz * 2bytes * 5sec)
	FlushThreshold   = 88200 * 2  // Flush every 2 seconds
)

type DeviceInfo struct {
	ID          int
	Name        string
	IsDefault   bool
	MaxChannels int
	HostAPI     string
	IsVirtual   bool
	VirtualType string
}

type Recorder struct {
	stream     *portaudio.Stream
	samples    []int16          // Memory buffer (limited size)
	recording  bool
	sampleRate float64
	deviceInfo *portaudio.DeviceInfo
	mutex      sync.Mutex
	startTime  time.Time
	
	// Phase 1: Memory-efficient recording
	tempFile     *os.File        // Temporary file for overflow
	totalSamples int64           // Total samples written
	lastFlush    time.Time       // Last flush time
	maxDuration  time.Duration   // Maximum recording duration (0 = unlimited)
	maxFileSize  int64          // Maximum file size in bytes (0 = unlimited)
}

func NewRecorder() (*Recorder, error) {
	err := portaudio.Initialize()
	if err != nil {
		return nil, err
	}

	return &Recorder{
		samples:      make([]int16, 0, MemoryBufferSize),
		recording:    false,
		sampleRate:   SampleRate,
		maxDuration:  0,  // Unlimited by default
		maxFileSize:  0,  // Unlimited by default
	}, nil
}

func NewRecorderWithDevice(deviceID int) (*Recorder, error) {
	err := portaudio.Initialize()
	if err != nil {
		return nil, err
	}

	devices, err := portaudio.Devices()
	if err != nil {
		portaudio.Terminate()
		return nil, err
	}

	var selectedDevice *portaudio.DeviceInfo
	for _, device := range devices {
		if device.Index == deviceID && device.MaxInputChannels > 0 {
			selectedDevice = device
			break
		}
	}

	if selectedDevice == nil {
		portaudio.Terminate()
		return nil, fmt.Errorf("device not found or has no input channels: index %d", deviceID)
	}

	return &Recorder{
		samples:      make([]int16, 0, MemoryBufferSize),
		recording:    false,
		sampleRate:   SampleRate,
		deviceInfo:   selectedDevice,
		maxDuration:  0,  // Unlimited by default
		maxFileSize:  0,  // Unlimited by default
	}, nil
}

func (r *Recorder) Start() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.recording {
		return errors.New("recording already in progress")
	}

	r.samples = make([]int16, 0, MemoryBufferSize)
	r.startTime = time.Now()
	r.lastFlush = time.Now()
	r.totalSamples = 0
	
	// Create temporary file for overflow data
	if r.tempFile != nil {
		r.tempFile.Close()
		r.tempFile = nil
	}

	var stream *portaudio.Stream
	var err error

	if r.deviceInfo != nil {
		params := portaudio.StreamParameters{
			Input: portaudio.StreamDeviceParameters{
				Device:   r.deviceInfo,
				Channels: Channels,
				Latency:  r.deviceInfo.DefaultLowInputLatency,
			},
			SampleRate:      r.sampleRate,
			FramesPerBuffer: BufferSize,
		}
		stream, err = portaudio.OpenStream(params, r.recordCallback)
	} else {
		stream, err = portaudio.OpenDefaultStream(
			Channels,
			0,
			r.sampleRate,
			BufferSize,
			r.recordCallback,
		)
	}

	if err != nil {
		return err
	}

	err = stream.Start()
	if err != nil {
		stream.Close()
		return err
	}

	r.stream = stream
	r.recording = true
	return nil
}

func (r *Recorder) recordCallback(in []int16) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.recording {
		return
	}
	
	// Check recording limits
	if r.exceedsLimits() {
		r.recording = false
		return
	}
	
	// Add samples to memory buffer
	r.samples = append(r.samples, in...)
	r.totalSamples += int64(len(in))
	
	// Check if buffer needs flushing
	if len(r.samples) >= FlushThreshold || time.Since(r.lastFlush) > 2*time.Second {
		r.flushToTempFile()
	}
}

func (r *Recorder) Stop() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.recording {
		return errors.New("no recording in progress")
	}

	err := r.stream.Stop()
	if err != nil {
		return err
	}

	err = r.stream.Close()
	if err != nil {
		return err
	}

	// Final flush of remaining samples
	if len(r.samples) > 0 {
		r.flushToTempFile()
	}

	r.recording = false
	r.stream = nil
	return nil
}

func (r *Recorder) SaveToFile(filename string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.totalSamples == 0 {
		return errors.New("no audio data to save")
	}

	// If we have a temp file, consolidate all data
	if r.tempFile != nil {
		return r.consolidateToFile(filename)
	}
	
	// If only memory samples, use existing method
	if len(r.samples) == 0 {
		return errors.New("no audio data to save")
	}

	return SaveWAV(filename, r.samples, int(r.sampleRate), Channels)
}

func (r *Recorder) Close() error {
	if r.recording {
		err := r.Stop()
		if err != nil {
			return err
		}
	}
	
	// Clean up temp file
	if r.tempFile != nil {
		r.tempFile.Close()
		os.Remove(r.tempFile.Name())  // Remove temp file
		r.tempFile = nil
	}

	portaudio.Terminate()
	return nil
}

func (r *Recorder) IsRecording() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.recording
}

func (r *Recorder) GetDuration() float64 {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return float64(r.totalSamples) / r.sampleRate
}

func (r *Recorder) GetElapsedTime() time.Duration {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.recording {
		return time.Since(r.startTime)
	}
	return 0
}

func detectVirtualDevice(device *portaudio.DeviceInfo) (bool, string) {
	name := strings.ToLower(device.Name)

	switch runtime.GOOS {
	case "darwin":
		if strings.Contains(name, "blackhole") {
			return true, "BlackHole"
		}
		if strings.Contains(name, "aggregate") || strings.Contains(name, "集約") {
			return true, "Aggregate"
		}
		if strings.Contains(name, "multi-output") || strings.Contains(name, "マルチ出力") {
			return true, "Multi-Output"
		}
	case "windows":
		if strings.Contains(name, "stereo mix") ||
			strings.Contains(name, "what u hear") ||
			strings.Contains(name, "rec. playback") {
			return true, "Stereo Mix"
		}
	}

	return false, ""
}

func ListDevices() ([]DeviceInfo, error) {
	err := portaudio.Initialize()
	if err != nil {
		return nil, err
	}
	defer portaudio.Terminate()

	devices, err := portaudio.Devices()
	if err != nil {
		return nil, err
	}

	defaultInput, err := portaudio.DefaultInputDevice()
	if err != nil {
		defaultInput = nil
	}

	var deviceList []DeviceInfo
	for _, device := range devices {
		if device.MaxInputChannels > 0 {
			isVirtual, virtualType := detectVirtualDevice(device)
			isDefault := defaultInput != nil && device.Index == defaultInput.Index

			deviceInfo := DeviceInfo{
				ID:          device.Index,
				Name:        device.Name,
				IsDefault:   isDefault,
				MaxChannels: device.MaxInputChannels,
				HostAPI:     device.HostApi.Name,
				IsVirtual:   isVirtual,
				VirtualType: virtualType,
			}
			deviceList = append(deviceList, deviceInfo)
		}
	}

	return deviceList, nil
}

// Phase 1: Memory-efficient recording helper methods

// exceedsLimits checks if recording limits are exceeded
func (r *Recorder) exceedsLimits() bool {
	// Check duration limit
	if r.maxDuration > 0 && time.Since(r.startTime) >= r.maxDuration {
		return true
	}
	
	// Check file size limit (approximate)
	if r.maxFileSize > 0 {
		estimatedSize := r.totalSamples * 2 + 44 // 2 bytes per sample + WAV header
		if estimatedSize >= r.maxFileSize {
			return true
		}
	}
	
	return false
}

// flushToTempFile flushes memory samples to temporary file
func (r *Recorder) flushToTempFile() error {
	if len(r.samples) == 0 {
		return nil
	}
	
	// Create temp file if it doesn't exist
	if r.tempFile == nil {
		tempFile, err := os.CreateTemp("", "koemoji_recording_*.raw")
		if err != nil {
			return err
		}
		r.tempFile = tempFile
	}
	
	// Write samples to temp file as raw bytes
	err := binary.Write(r.tempFile, binary.LittleEndian, r.samples)
	if err != nil {
		return err
	}
	
	// Clear memory buffer but preserve capacity
	r.samples = r.samples[:0]
	r.lastFlush = time.Now()
	
	return nil
}

// consolidateToFile combines temp file and memory buffer into final WAV file
func (r *Recorder) consolidateToFile(filename string) error {
	// Create output file
	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()
	
	// Calculate total data size
	dataSize := r.totalSamples * 2 // 2 bytes per sample
	fileSize := 36 + dataSize
	
	// Write WAV header
	header := WAVHeader{
		ChunkID:       [4]byte{'R', 'I', 'F', 'F'},
		ChunkSize:     uint32(fileSize),
		Format:        [4]byte{'W', 'A', 'V', 'E'},
		Subchunk1ID:   [4]byte{'f', 'm', 't', ' '},
		Subchunk1Size: 16,
		AudioFormat:   1, // PCM
		NumChannels:   uint16(Channels),
		SampleRate:    uint32(r.sampleRate),
		ByteRate:      uint32(r.sampleRate * Channels * 16 / 8),
		BlockAlign:    uint16(Channels * 16 / 8),
		BitsPerSample: 16,
		Subchunk2ID:   [4]byte{'d', 'a', 't', 'a'},
		Subchunk2Size: uint32(dataSize),
	}
	
	err = binary.Write(outFile, binary.LittleEndian, header)
	if err != nil {
		return err
	}
	
	// Copy temp file data if it exists
	if r.tempFile != nil {
		r.tempFile.Seek(0, 0) // Seek to beginning
		_, err = io.Copy(outFile, r.tempFile)
		if err != nil {
			return err
		}
	}
	
	// Write remaining memory samples
	if len(r.samples) > 0 {
		err = binary.Write(outFile, binary.LittleEndian, r.samples)
		if err != nil {
			return err
		}
	}
	
	return nil
}

// SetLimits configures recording limits
func (r *Recorder) SetLimits(maxDuration time.Duration, maxFileSize int64) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.maxDuration = maxDuration
	r.maxFileSize = maxFileSize
}
