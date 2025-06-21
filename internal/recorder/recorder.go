package recorder

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
)

const (
	SampleRate  = 44100
	Channels    = 1
	BufferSize  = 4096
)

type DeviceInfo struct {
	ID           int
	Name         string
	IsDefault    bool
	MaxChannels  int
	HostAPI      string
	IsVirtual    bool
	VirtualType  string
}

type Recorder struct {
	stream     *portaudio.Stream
	samples    []int16
	recording  bool
	sampleRate float64
	deviceInfo *portaudio.DeviceInfo
	mutex      sync.Mutex
	startTime  time.Time
}

func NewRecorder() (*Recorder, error) {
	err := portaudio.Initialize()
	if err != nil {
		return nil, err
	}

	return &Recorder{
		samples:    make([]int16, 0),
		recording:  false,
		sampleRate: SampleRate,
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
		samples:    make([]int16, 0),
		recording:  false,
		sampleRate: SampleRate,
		deviceInfo: selectedDevice,
	}, nil
}

func (r *Recorder) Start() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.recording {
		return errors.New("recording already in progress")
	}

	r.samples = make([]int16, 0)
	r.startTime = time.Now()

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
	
	if r.recording {
		r.samples = append(r.samples, in...)
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

	r.recording = false
	r.stream = nil
	return nil
}

func (r *Recorder) SaveToFile(filename string) error {
	r.mutex.Lock()
	samples := make([]int16, len(r.samples))
	copy(samples, r.samples)
	r.mutex.Unlock()

	if len(samples) == 0 {
		return errors.New("no audio data to save")
	}

	return SaveWAV(filename, samples, int(r.sampleRate), Channels)
}

func (r *Recorder) Close() error {
	if r.recording {
		err := r.Stop()
		if err != nil {
			return err
		}
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
	return float64(len(r.samples)) / r.sampleRate
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