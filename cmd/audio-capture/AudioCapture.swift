import Foundation
import ScreenCaptureKit
import AVFoundation

// MARK: - Error Handling

enum ScreenCaptureError: Error {
    case noDisplayAvailable
    case streamCreationFailed(String)
    case captureFailed(String)
    case fileWriteFailed(String)
}

// MARK: - Audio Capture Manager

@available(macOS 13.0, *)
class ScreenCaptureAudioRecorder: NSObject, SCStreamOutput {
    private var stream: SCStream?
    private var audioFile: AVAudioFile?
    private let outputURL: URL
    private var duration: TimeInterval
    private var startTime: Date?
    private var isRecording = false

    init(outputPath: String, duration: TimeInterval = 0) {
        self.outputURL = URL(fileURLWithPath: outputPath)
        self.duration = duration
        super.init()
    }

    func startRecording() async throws {
        print("Starting ScreenCaptureKit audio recording...", to: &standardError)

        // Get available content (displays)
        let availableContent = try await SCShareableContent.excludingDesktopWindows(
            false,
            onScreenWindowsOnly: true
        )

        guard let display = availableContent.displays.first else {
            throw ScreenCaptureError.noDisplayAvailable
        }

        print("Found display: \(display.width)x\(display.height)", to: &standardError)

        // Create content filter (we need a display for ScreenCaptureKit, even for audio-only)
        let filter = SCContentFilter(display: display, excludingWindows: [])

        // Configure stream for audio capture
        let streamConfig = SCStreamConfiguration()

        // Audio settings
        streamConfig.capturesAudio = true
        streamConfig.excludesCurrentProcessAudio = true
        streamConfig.sampleRate = 48000
        streamConfig.channelCount = 2

        // Minimal video settings (required even for audio-only)
        streamConfig.width = display.width
        streamConfig.height = display.height
        streamConfig.minimumFrameInterval = CMTime(value: 1, timescale: 1)
        streamConfig.queueDepth = 5

        print("Creating stream with config: \(streamConfig.width)x\(streamConfig.height), audio: \(streamConfig.capturesAudio)", to: &standardError)

        // Create stream
        let newStream = SCStream(filter: filter, configuration: streamConfig, delegate: nil)
        stream = newStream

        // Add audio output BEFORE starting capture
        let audioQueue = DispatchQueue(label: "com.koemoji.audiocapture", qos: .userInteractive)
        try newStream.addStreamOutput(self, type: .audio, sampleHandlerQueue: audioQueue)

        print("Starting stream capture...", to: &standardError)

        // Start capture
        try await newStream.startCapture()

        isRecording = true
        startTime = Date()
        print("✓ Recording started successfully", to: &standardError)

        // Wait for duration or until stopped
        if duration > 0 {
            try await Task.sleep(nanoseconds: UInt64(duration * 1_000_000_000))
            try await stopRecording()
        }
    }

    func stopRecording() async throws {
        guard isRecording else { return }

        isRecording = false

        if let stream = stream {
            try await stream.stopCapture()
            print("✓ Recording stopped", to: &standardError)
        }

        stream = nil
        audioFile = nil
    }

    // MARK: - SCStreamOutput Protocol

    func stream(_ stream: SCStream, didOutputSampleBuffer sampleBuffer: CMSampleBuffer, of type: SCStreamOutputType) {
        guard isRecording, type == .audio else { return }

        // Check if duration has elapsed
        if duration > 0, let startTime = startTime {
            let elapsed = Date().timeIntervalSince(startTime)
            if elapsed >= duration {
                Task {
                    try? await stopRecording()
                }
                return
            }
        }

        // Process audio sample buffer
        do {
            try processSampleBuffer(sampleBuffer)
        } catch {
            print("Error processing sample buffer: \(error)", to: &standardError)
        }
    }

    private func processSampleBuffer(_ sampleBuffer: CMSampleBuffer) throws {
        // Extract audio format from sample buffer
        guard let formatDescription = sampleBuffer.formatDescription else {
            throw ScreenCaptureError.captureFailed("No format description")
        }

        let audioStreamBasicDescription = CMAudioFormatDescriptionGetStreamBasicDescription(formatDescription)
        guard let asbd = audioStreamBasicDescription else {
            throw ScreenCaptureError.captureFailed("Failed to get audio stream description")
        }

        // Create AVAudioFormat from ASBD
        guard let format = AVAudioFormat(streamDescription: asbd) else {
            throw ScreenCaptureError.captureFailed("Failed to create AVAudioFormat")
        }

        // Initialize audio file on first buffer
        if audioFile == nil {
            // Use the format directly from the stream (works best with CAF)
            audioFile = try AVAudioFile(forWriting: outputURL, settings: format.settings)
            print("Audio format: \(format.sampleRate)Hz, \(format.channelCount)ch, Float32", to: &standardError)
        }

        // Convert CMSampleBuffer to AVAudioPCMBuffer
        try sampleBuffer.withAudioBufferList { audioBufferList, blockBuffer in
            guard let pcmBuffer = AVAudioPCMBuffer(pcmFormat: format, bufferListNoCopy: audioBufferList.unsafePointer) else {
                throw ScreenCaptureError.captureFailed("Failed to create PCM buffer")
            }

            // Write to file
            try audioFile?.write(from: pcmBuffer)
        }
    }
}

// MARK: - Standard Error Extension

var standardError = FileHandle.standardError

extension FileHandle: @retroactive TextOutputStream {
    public func write(_ string: String) {
        let data = Data(string.utf8)
        self.write(data)
    }
}
