import Foundation
import AVFoundation
import AudioToolbox

// MARK: - Core Audio Error Handling

enum AudioCaptureError: Error {
    case setupFailed(String)
    case recordingFailed(String)
    case permissionDenied
    case unsupportedOS
}

extension OSStatus {
    var isSuccess: Bool {
        return self == noErr
    }

    func checkError(_ context: String) throws {
        guard self.isSuccess else {
            throw AudioCaptureError.setupFailed("\(context): OSStatus \(self)")
        }
    }
}

// MARK: - Audio Tap Manager

class AudioTapManager {
    private var tapID: AudioObjectID = 0
    private var aggregateDeviceID: AudioObjectID = 0
    private var format: AudioStreamBasicDescription?

    func setupSystemAudioTap() throws -> (deviceID: AudioObjectID, format: AudioStreamBasicDescription) {
        // Step 1: Create system audio tap
        tapID = try createSystemAudioTap()

        // Step 2: Get audio format from tap
        format = try readAudioFormat(from: tapID)

        // Step 3: Create aggregate device with the tap
        aggregateDeviceID = try createAggregateDevice(withTapID: tapID)

        guard let fmt = format else {
            throw AudioCaptureError.setupFailed("Failed to get audio format")
        }

        return (aggregateDeviceID, fmt)
    }

    private func createSystemAudioTap() throws -> AudioObjectID {
        // Get default output device
        var defaultOutputID = AudioObjectID(kAudioObjectUnknown)
        var propertyAddress = AudioObjectPropertyAddress(
            mSelector: kAudioHardwarePropertyDefaultOutputDevice,
            mScope: kAudioObjectPropertyScopeGlobal,
            mElement: kAudioObjectPropertyElementMain
        )

        var size = UInt32(MemoryLayout<AudioObjectID>.size)
        try AudioObjectGetPropertyData(
            AudioObjectID(kAudioObjectSystemObject),
            &propertyAddress,
            0,
            nil,
            &size,
            &defaultOutputID
        ).checkError("Get default output device")

        // Translate to process object (system-wide)
        // For system-wide capture, use pid 0
        var processObjectID = AudioObjectID(kAudioObjectUnknown)
        var pid: pid_t = 0

        propertyAddress.mSelector = kAudioHardwarePropertyTranslatePIDToProcessObject

        var pidSize = UInt32(MemoryLayout<pid_t>.size)
        var processObjectIDSize = UInt32(MemoryLayout<AudioObjectID>.size)

        try AudioObjectGetPropertyData(
            AudioObjectID(kAudioObjectSystemObject),
            &propertyAddress,
            pidSize,
            &pid,
            &processObjectIDSize,
            &processObjectID
        ).checkError("Translate PID to process object")

        // Create tap description
        var tapDescription = CATapDescription(
            stereoMixdownOfProcesses: [processObjectID],
            andMuteBehavior: .mutedWhenTapped,
            withStream: nil
        )

        // Create process tap
        var tapID = AudioObjectID(kAudioObjectUnknown)
        try AudioHardwareCreateProcessTap(
            &tapDescription,
            &tapID
        ).checkError("Create process tap")

        return tapID
    }

    private func readAudioFormat(from tapID: AudioObjectID) throws -> AudioStreamBasicDescription {
        var format = AudioStreamBasicDescription()
        var propertyAddress = AudioObjectPropertyAddress(
            mSelector: kAudioTapPropertyFormat,
            mScope: kAudioObjectPropertyScopeGlobal,
            mElement: kAudioObjectPropertyElementMain
        )

        var size = UInt32(MemoryLayout<AudioStreamBasicDescription>.size)
        try AudioObjectGetPropertyData(
            tapID,
            &propertyAddress,
            0,
            nil,
            &size,
            &format
        ).checkError("Read audio format from tap")

        return format
    }

    private func createAggregateDevice(withTapID tapID: AudioObjectID) throws -> AudioObjectID {
        // Get tap UUID
        var tapUUID: CFString?
        var propertyAddress = AudioObjectPropertyAddress(
            mSelector: kAudioTapPropertyUIDKey,
            mScope: kAudioObjectPropertyScopeGlobal,
            mElement: kAudioObjectPropertyElementMain
        )

        var size = UInt32(MemoryLayout<CFString>.size)
        try AudioObjectGetPropertyData(
            tapID,
            &propertyAddress,
            0,
            nil,
            &size,
            &tapUUID
        ).checkError("Get tap UUID")

        guard let uuid = tapUUID else {
            throw AudioCaptureError.setupFailed("Failed to get tap UUID")
        }

        // Create aggregate device description
        let uniqueID = UUID().uuidString
        let deviceDict: [String: Any] = [
            kAudioAggregateDeviceNameKey as String: "KoeMoji System Audio Tap",
            kAudioAggregateDeviceUIDKey as String: uniqueID,
            kAudioAggregateDevicePrivateKey as String: 1,  // Private device
            kAudioAggregateDeviceTapListKey as String: [
                [kAudioSubTapUIDKey as String: uuid as String]
            ]
        ]

        // Create aggregate device
        var aggregateDeviceID = AudioObjectID(kAudioObjectUnknown)
        try AudioHardwareCreateAggregateDevice(
            deviceDict as CFDictionary,
            &aggregateDeviceID
        ).checkError("Create aggregate device")

        return aggregateDeviceID
    }

    func cleanup() {
        if aggregateDeviceID != 0 {
            AudioHardwareDestroyAggregateDevice(aggregateDeviceID)
        }
        if tapID != 0 {
            // Tap cleanup (if needed)
        }
    }

    deinit {
        cleanup()
    }
}

// MARK: - Audio Recorder

class AudioRecorder {
    private let deviceID: AudioObjectID
    private let format: AudioStreamBasicDescription
    private let outputURL: URL
    private var audioFile: ExtAudioFileRef?
    private var ioProcID: AudioDeviceIOProcID?
    private var isRecording = false

    init(deviceID: AudioObjectID, format: AudioStreamBasicDescription, outputPath: String) {
        self.deviceID = deviceID
        self.format = format
        self.outputURL = URL(fileURLWithPath: outputPath)
    }

    func startRecording() throws {
        // Create WAV file
        var clientFormat = format
        var fileFormat = AudioStreamBasicDescription(
            mSampleRate: format.mSampleRate,
            mFormatID: kAudioFormatLinearPCM,
            mFormatFlags: kAudioFormatFlagIsSignedInteger | kAudioFormatFlagIsPacked,
            mBytesPerPacket: 4,
            mFramesPerPacket: 1,
            mBytesPerFrame: 4,
            mChannelsPerFrame: 2,
            mBitsPerChannel: 16,
            mReserved: 0
        )

        try ExtAudioFileCreateWithURL(
            outputURL as CFURL,
            kAudioFileWAVEType,
            &fileFormat,
            nil,
            AudioFileFlags.eraseFile.rawValue,
            &audioFile
        ).checkError("Create audio file")

        guard let file = audioFile else {
            throw AudioCaptureError.recordingFailed("Failed to create audio file")
        }

        // Set client format
        try ExtAudioFileSetProperty(
            file,
            kExtAudioFileProperty_ClientDataFormat,
            UInt32(MemoryLayout<AudioStreamBasicDescription>.size),
            &clientFormat
        ).checkError("Set client format")

        // Create IO proc
        let unmanagedSelf = Unmanaged.passUnretained(self).toOpaque()
        try AudioDeviceCreateIOProcID(
            deviceID,
            { (
                inDevice: AudioObjectID,
                inNow: UnsafePointer<AudioTimeStamp>,
                inInputData: UnsafePointer<AudioBufferList>,
                inInputTime: UnsafePointer<AudioTimeStamp>,
                outOutputData: UnsafeMutablePointer<AudioBufferList>,
                inOutputTime: UnsafePointer<AudioTimeStamp>,
                inClientData: UnsafeMutableRawPointer?
            ) -> OSStatus in
                guard let clientData = inClientData else { return noErr }
                let recorder = Unmanaged<AudioRecorder>.fromOpaque(clientData).takeUnretainedValue()
                return recorder.handleAudioBuffer(inInputData)
            },
            unmanagedSelf,
            &ioProcID
        ).checkError("Create IO proc")

        // Start device
        try AudioDeviceStart(deviceID, ioProcID).checkError("Start audio device")

        isRecording = true
        print("Recording started...", to: &standardError)
    }

    private func handleAudioBuffer(_ bufferList: UnsafePointer<AudioBufferList>) -> OSStatus {
        guard let file = audioFile else { return kAudioHardwareUnspecifiedError }

        let buffers = UnsafeBufferPointer<AudioBuffer>(
            start: &UnsafeMutablePointer(mutating: bufferList).pointee.mBuffers,
            count: Int(bufferList.pointee.mNumberBuffers)
        )

        guard let buffer = buffers.first else { return noErr }

        let frameCount = buffer.mDataByteSize / UInt32(format.mBytesPerFrame)
        var mutableBufferList = bufferList.pointee

        let status = ExtAudioFileWrite(
            file,
            frameCount,
            &mutableBufferList
        )

        return status
    }

    func stopRecording() {
        guard isRecording else { return }

        if let procID = ioProcID {
            AudioDeviceStop(deviceID, procID)
            AudioDeviceDestroyIOProcID(deviceID, procID)
        }

        if let file = audioFile {
            ExtAudioFileDispose(file)
        }

        isRecording = false
        print("Recording stopped.", to: &standardError)
        print("Saved to: \(outputURL.path)", to: &standardError)
    }

    deinit {
        stopRecording()
    }
}
