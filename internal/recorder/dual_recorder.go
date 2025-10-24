// +build windows

package recorder

import (
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/gordonklaus/portaudio"
	"github.com/moutend/go-wca/pkg/wca"
)

// DualRecorder records system audio and microphone simultaneously
// Windows only: Uses WASAPI Loopback for system audio + PortAudio for microphone
type DualRecorder struct {
	// System audio (WASAPI Loopback)
	systemAudioEnabled bool
	systemDevice       *wca.IMMDevice
	systemClient       *wca.IAudioClient
	systemCapture      *wca.IAudioCaptureClient
	systemFormat       *wca.WAVEFORMATEX
	systemEvent        uintptr

	// Microphone (PortAudio)
	micStream    *portaudio.Stream
	micDevice    *portaudio.DeviceInfo
	micEnabled   bool

	// Mixing
	systemSamples chan int16
	micSamples    chan int16
	mixedSamples  []int16
	done          chan bool

	// State
	recording bool
	mutex     sync.Mutex
	wg        sync.WaitGroup

	// Settings
	systemVolume float64 // 0.0 - 1.0
	micVolume    float64 // 0.0 - 1.0

	// Recording limits (same as single recorder)
	startTime   time.Time
	maxDuration time.Duration
	maxFileSize int64
}

// NewDualRecorder creates a new dual recorder with default settings
func NewDualRecorder() (*DualRecorder, error) {
	// Initialize PortAudio
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("portaudio init failed: %w", err)
	}

	return &DualRecorder{
		systemAudioEnabled: true,
		micEnabled:         true,
		systemVolume:       0.7,  // 70% system audio
		micVolume:          1.0,  // 100% microphone
		systemSamples:      make(chan int16, 96000),
		micSamples:         make(chan int16, 88200),
		done:               make(chan bool),
		mixedSamples:       make([]int16, 0, 480000),
		maxDuration:        0, // Unlimited
		maxFileSize:        0, // Unlimited
	}, nil
}

// NewDualRecorderWithDevices creates a dual recorder with specific devices
func NewDualRecorderWithDevices(micDeviceName string) (*DualRecorder, error) {
	dr, err := NewDualRecorder()
	if err != nil {
		return nil, err
	}

	// Find microphone device
	if micDeviceName != "" {
		devices, err := portaudio.Devices()
		if err != nil {
			portaudio.Terminate()
			return nil, err
		}

		for _, device := range devices {
			if device.Name == micDeviceName && device.MaxInputChannels > 0 {
				dr.micDevice = device
				break
			}
		}

		if dr.micDevice == nil {
			portaudio.Terminate()
			return nil, fmt.Errorf("microphone device not found: '%s'", micDeviceName)
		}
	}

	return dr, nil
}

// SetVolumes sets the volume levels for system audio and microphone
func (dr *DualRecorder) SetVolumes(systemVol, micVol float64) {
	dr.mutex.Lock()
	defer dr.mutex.Unlock()

	// Allow volume range 0.0 - 2.0 (0% - 200%)
	if systemVol >= 0.0 && systemVol <= 2.0 {
		dr.systemVolume = systemVol
	}
	if micVol >= 0.0 && micVol <= 2.0 {
		dr.micVolume = micVol
	}
}

// SetLimits configures recording limits
func (dr *DualRecorder) SetLimits(maxDuration time.Duration, maxFileSize int64) {
	dr.mutex.Lock()
	defer dr.mutex.Unlock()
	dr.maxDuration = maxDuration
	dr.maxFileSize = maxFileSize
}

// Start begins dual recording
func (dr *DualRecorder) Start() error {
	dr.mutex.Lock()
	defer dr.mutex.Unlock()

	if dr.recording {
		return errors.New("recording already in progress")
	}

	// Reset state
	dr.mixedSamples = make([]int16, 0, 480000)
	dr.startTime = time.Now()
	dr.done = make(chan bool)

	// Recreate channels in case they were closed
	dr.systemSamples = make(chan int16, 96000)
	dr.micSamples = make(chan int16, 88200)

	// Start system audio capture
	if dr.systemAudioEnabled {
		dr.wg.Add(1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("System audio capture panic: %v\n", r)
				}
				dr.wg.Done()
			}()
			if err := dr.captureSystemAudio(); err != nil {
				fmt.Printf("System audio capture error: %v\n", err)
			}
		}()
	}

	// Start microphone capture
	if dr.micEnabled {
		dr.wg.Add(1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Microphone capture panic: %v\n", r)
				}
				dr.wg.Done()
			}()
			if err := dr.captureMicrophone(); err != nil {
				fmt.Printf("Microphone capture error: %v\n", err)
			}
		}()
	}

	// Start mixer
	dr.wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Mixer panic: %v\n", r)
			}
			dr.wg.Done()
		}()
		dr.mixAudio()
	}()

	dr.recording = true
	return nil
}

// Stop ends dual recording
func (dr *DualRecorder) Stop() error {
	dr.mutex.Lock()
	if !dr.recording {
		dr.mutex.Unlock()
		return errors.New("no recording in progress")
	}
	dr.mutex.Unlock()

	// Signal stop
	close(dr.done)

	// Wait for all goroutines
	dr.wg.Wait()

	// Close channels
	close(dr.systemSamples)
	close(dr.micSamples)

	dr.mutex.Lock()
	dr.recording = false
	dr.mutex.Unlock()

	return nil
}

// SaveToFile saves the mixed recording to a WAV file
func (dr *DualRecorder) SaveToFile(filename string) error {
	dr.mutex.Lock()
	defer dr.mutex.Unlock()

	if len(dr.mixedSamples) == 0 {
		return errors.New("no audio data to save")
	}

	return SaveWAV(filename, dr.mixedSamples, SampleRate, Channels)
}

// SaveToFileWithNormalization saves recording with optional audio normalization
func (dr *DualRecorder) SaveToFileWithNormalization(filename string, enableNormalization bool) error {
	dr.mutex.Lock()

	if len(dr.mixedSamples) == 0 {
		dr.mutex.Unlock()
		return errors.New("no audio data to save")
	}

	// Apply normalization if enabled
	normalized := false
	if enableNormalization {
		normalized = NormalizeAudio(dr.mixedSamples)
	}

	dr.mutex.Unlock()

	// Save to file
	err := dr.SaveToFile(filename)
	if err != nil {
		return err
	}

	// Log if normalization was applied
	if normalized {
		// Normalization was applied (logged by caller if needed)
	}

	return nil
}

// Close releases all resources
func (dr *DualRecorder) Close() error {
	if dr.recording {
		if err := dr.Stop(); err != nil {
			return err
		}
	}

	portaudio.Terminate()
	return nil
}

// IsRecording returns whether recording is in progress
func (dr *DualRecorder) IsRecording() bool {
	dr.mutex.Lock()
	defer dr.mutex.Unlock()
	return dr.recording
}

// GetDuration returns the duration of recorded audio in seconds
func (dr *DualRecorder) GetDuration() float64 {
	dr.mutex.Lock()
	defer dr.mutex.Unlock()
	return float64(len(dr.mixedSamples)) / float64(SampleRate)
}

// GetElapsedTime returns elapsed recording time
func (dr *DualRecorder) GetElapsedTime() time.Duration {
	dr.mutex.Lock()
	defer dr.mutex.Unlock()
	if dr.recording {
		return time.Since(dr.startTime)
	}
	return 0
}

// captureSystemAudio captures system audio using WASAPI Loopback
func (dr *DualRecorder) captureSystemAudio() error {
	// COM initialization (ignore error if already initialized)
	ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	defer ole.CoUninitialize()

	// Get device enumerator
	var deviceEnumerator *wca.IMMDeviceEnumerator
	if err := wca.CoCreateInstance(
		wca.CLSID_MMDeviceEnumerator,
		0,
		wca.CLSCTX_ALL,
		wca.IID_IMMDeviceEnumerator,
		&deviceEnumerator,
	); err != nil {
		return fmt.Errorf("device enumerator failed: %w", err)
	}
	defer deviceEnumerator.Release()

	// Get default output device
	var defaultDevice *wca.IMMDevice
	if err := deviceEnumerator.GetDefaultAudioEndpoint(
		wca.ERender,
		wca.EConsole,
		&defaultDevice,
	); err != nil {
		return fmt.Errorf("get default endpoint failed: %w", err)
	}
	defer defaultDevice.Release()

	// Activate audio client
	var audioClient *wca.IAudioClient
	if err := defaultDevice.Activate(
		wca.IID_IAudioClient,
		wca.CLSCTX_ALL,
		nil,
		&audioClient,
	); err != nil {
		return fmt.Errorf("audio client activation failed: %w", err)
	}
	defer audioClient.Release()

	// Get mix format
	var mixFormat *wca.WAVEFORMATEX
	if err := audioClient.GetMixFormat(&mixFormat); err != nil {
		return fmt.Errorf("get mix format failed: %w", err)
	}

	// Get device period
	var defaultPeriod wca.REFERENCE_TIME
	if err := audioClient.GetDevicePeriod(&defaultPeriod, nil); err != nil {
		return fmt.Errorf("get device period failed: %w", err)
	}

	// Initialize audio client in loopback mode
	if err := audioClient.Initialize(
		wca.AUDCLNT_SHAREMODE_SHARED,
		wca.AUDCLNT_STREAMFLAGS_LOOPBACK|wca.AUDCLNT_STREAMFLAGS_EVENTCALLBACK,
		defaultPeriod,
		0,
		mixFormat,
		nil,
	); err != nil {
		return fmt.Errorf("audio client init failed: %w", err)
	}

	// Create event
	audioReadyEvent := wca.CreateEventExA(0, 0, 0, wca.EVENT_MODIFY_STATE|wca.SYNCHRONIZE)
	defer wca.CloseHandle(audioReadyEvent)

	if err := audioClient.SetEventHandle(audioReadyEvent); err != nil {
		return fmt.Errorf("set event handle failed: %w", err)
	}

	// Get capture client
	var captureClient *wca.IAudioCaptureClient
	if err := audioClient.GetService(
		wca.IID_IAudioCaptureClient,
		&captureClient,
	); err != nil {
		return fmt.Errorf("get capture client failed: %w", err)
	}
	defer captureClient.Release()

	// Start capture
	if err := audioClient.Start(); err != nil {
		return fmt.Errorf("start capture failed: %w", err)
	}
	defer audioClient.Stop()

	// Capture loop
	for {
		select {
		case <-dr.done:
			return nil
		default:
		}

		wca.WaitForSingleObject(audioReadyEvent, 100)

		var numFramesToRead uint32
		var flags uint32
		var data *byte

		if err := captureClient.GetBuffer(&data, &numFramesToRead, &flags, nil, nil); err != nil {
			continue
		}

		if numFramesToRead > 0 {
			bytesPerSample := int(mixFormat.WBitsPerSample / 8)
			p := unsafe.Pointer(data)

			// Convert 32bit float stereo → 16bit int mono
			if bytesPerSample == 4 && mixFormat.NChannels == 2 {
				for i := 0; i < int(numFramesToRead); i++ {
					leftFloat := *(*float32)(unsafe.Pointer(uintptr(p) + uintptr(i*8)))
					rightFloat := *(*float32)(unsafe.Pointer(uintptr(p) + uintptr(i*8+4)))

					// Stereo to mono
					monoFloat := (leftFloat + rightFloat) / 2.0

					// Apply volume
					monoFloat *= float32(dr.systemVolume)

					// Convert to int16
					intSample := int16(monoFloat * 32767.0)

					select {
					case dr.systemSamples <- intSample:
					default:
						// Channel full, skip
					}
				}
			}
		}

		captureClient.ReleaseBuffer(numFramesToRead)
	}
}

// captureMicrophone captures microphone audio using PortAudio
func (dr *DualRecorder) captureMicrophone() error {
	var stream *portaudio.Stream
	var err error

	if dr.micDevice != nil {
		params := portaudio.StreamParameters{
			Input: portaudio.StreamDeviceParameters{
				Device:   dr.micDevice,
				Channels: Channels,
				Latency:  dr.micDevice.DefaultLowInputLatency,
			},
			SampleRate:      SampleRate,
			FramesPerBuffer: BufferSize,
		}
		stream, err = portaudio.OpenStream(params, func(in []int16) {
			for _, sample := range in {
				// Apply volume
				adjusted := int32(float64(sample) * dr.micVolume)
				if adjusted > 32767 {
					adjusted = 32767
				} else if adjusted < -32768 {
					adjusted = -32768
				}

				select {
				case dr.micSamples <- int16(adjusted):
				default:
					// Channel full, skip
				}
			}
		})
	} else {
		stream, err = portaudio.OpenDefaultStream(
			Channels,
			0,
			SampleRate,
			BufferSize,
			func(in []int16) {
				for _, sample := range in {
					// Apply volume
					adjusted := int32(float64(sample) * dr.micVolume)
					if adjusted > 32767 {
						adjusted = 32767
					} else if adjusted < -32768 {
						adjusted = -32768
					}

					select {
					case dr.micSamples <- int16(adjusted):
					default:
						// Channel full, skip
					}
				}
			},
		)
	}

	if err != nil {
		return fmt.Errorf("open microphone stream failed: %w", err)
	}
	defer stream.Close()

	if err := stream.Start(); err != nil {
		return fmt.Errorf("start microphone stream failed: %w", err)
	}
	defer stream.Stop()

	// Wait for done signal
	<-dr.done
	return nil
}

// mixAudio mixes system audio and microphone in real-time
func (dr *DualRecorder) mixAudio() {
	systemBuf := make([]int16, 0, 10000)
	micBuf := make([]int16, 0, 10000)

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-dr.done:
			// Final mix of remaining buffers
			dr.mixBuffers(&systemBuf, &micBuf)
			return

		case s, ok := <-dr.systemSamples:
			if ok {
				systemBuf = append(systemBuf, s)
			}

		case m, ok := <-dr.micSamples:
			if ok {
				micBuf = append(micBuf, m)
			}

		case <-ticker.C:
			// Periodic mix
			dr.mixBuffers(&systemBuf, &micBuf)
		}

		// Check recording limits
		if dr.exceedsLimits() {
			return
		}
	}
}

// mixBuffers mixes accumulated samples from both sources
func (dr *DualRecorder) mixBuffers(systemBuf, micBuf *[]int16) {
	minLen := len(*systemBuf)
	if len(*micBuf) < minLen {
		minLen = len(*micBuf)
	}

	if minLen > 100 {
		dr.mutex.Lock()
		for i := 0; i < minLen; i++ {
			// Simple additive mixing with clipping prevention
			mixed := int32((*systemBuf)[i]) + int32((*micBuf)[i])
			if mixed > 32767 {
				mixed = 32767
			} else if mixed < -32768 {
				mixed = -32768
			}
			dr.mixedSamples = append(dr.mixedSamples, int16(mixed))
		}
		dr.mutex.Unlock()

		// Remove processed samples
		*systemBuf = (*systemBuf)[minLen:]
		*micBuf = (*micBuf)[minLen:]
	}
}

// exceedsLimits checks if recording limits are exceeded
func (dr *DualRecorder) exceedsLimits() bool {
	// Check duration limit
	if dr.maxDuration > 0 && time.Since(dr.startTime) >= dr.maxDuration {
		return true
	}

	// Check file size limit (approximate)
	if dr.maxFileSize > 0 {
		estimatedSize := int64(len(dr.mixedSamples))*2 + 44 // 2 bytes per sample + WAV header
		if estimatedSize >= dr.maxFileSize {
			return true
		}
	}

	return false
}
