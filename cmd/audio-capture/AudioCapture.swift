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
    private let finalOutputURL: URL  // User-requested output path
    private var duration: TimeInterval
    private var startTime: Date?
    private var isRecording = false

    init(outputPath: String, duration: TimeInterval = 0) {
        // Always write as CAF format (AVAudioFile default)
        // Conversion to WAV is handled by the caller (Go code)
        self.outputURL = URL(fileURLWithPath: outputPath)
        self.finalOutputURL = self.outputURL
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

        print("Stopping recording...", to: &standardError)
        isRecording = false

        // Stop the stream first
        if let stream = stream {
            try await stream.stopCapture()
        }

        // Give a moment for final buffers to be written
        try await Task.sleep(nanoseconds: 100_000_000) // 0.1 seconds

        // Close audio file explicitly
        audioFile = nil
        stream = nil

        print("✓ Recording stopped", to: &standardError)
    }

    // MARK: - SCStreamOutput Protocol

    func stream(_ stream: SCStream, didOutputSampleBuffer sampleBuffer: CMSampleBuffer, of type: SCStreamOutputType) {
        guard isRecording, type == .audio else { return }

        // Process audio sample buffer
        do {
            try processSampleBuffer(sampleBuffer)
        } catch {
            print("⚠ Error processing sample buffer: \(error)", to: &standardError)
            // Don't crash on individual buffer errors
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
            audioFile = try AVAudioFile(forWriting: outputURL, settings: format.settings)
            print("System audio format: \(format.sampleRate)Hz, \(format.channelCount)ch", to: &standardError)
        }

        // Convert CMSampleBuffer to AVAudioPCMBuffer and write
        try sampleBuffer.withAudioBufferList { audioBufferList, blockBuffer in
            guard let pcmBuffer = AVAudioPCMBuffer(pcmFormat: format, bufferListNoCopy: audioBufferList.unsafePointer) else {
                throw ScreenCaptureError.captureFailed("Failed to create PCM buffer")
            }
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
