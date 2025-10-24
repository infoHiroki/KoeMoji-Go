package recorder

import "time"

// AudioRecorder is the common interface for both Recorder and DualRecorder
type AudioRecorder interface {
	Start() error
	Stop() error
	IsRecording() bool
	GetDuration() float64
	GetElapsedTime() time.Duration
	SaveToFile(filename string) error
	SaveToFileWithNormalization(filename string, enableNormalization bool) error
	SetLimits(maxDuration time.Duration, maxFileSize int64)
	Close() error
}
