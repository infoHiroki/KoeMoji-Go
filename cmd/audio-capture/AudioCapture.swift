import Foundation
import AVFoundation
import AudioToolbox
import CoreAudio

// MARK: - Error Handling

enum AudioCaptureError: Error {
    case setupFailed(String)
    case recordingFailed(String)
    case tapCreationFailed(OSStatus)
    case aggregateDeviceCreationFailed(OSStatus)
    case tapAssignmentFailed(OSStatus)
}

// MARK: - Audio Property Utilities

func getPropertyAddress(
    selector: AudioObjectPropertySelector,
    scope: AudioObjectPropertyScope = kAudioObjectPropertyScopeGlobal,
    element: AudioObjectPropertyElement = kAudioObjectPropertyElementMain
) -> AudioObjectPropertyAddress {
    return AudioObjectPropertyAddress(mSelector: selector, mScope: scope, mElement: element)
}

// MARK: - Audio Tap Manager

class AudioTapManager {
    private var tapID: AudioObjectID?
    private var deviceID: AudioObjectID?

    deinit {
        cleanup()
    }

    func setupSystemAudioTap() throws -> (deviceID: AudioObjectID, format: AudioStreamBasicDescription) {
        // Create system audio tap
        tapID = try createSystemAudioTap()

        // Create aggregate device
        deviceID = try createAggregateDevice()

        guard let tapID = tapID, let deviceID = deviceID else {
            throw AudioCaptureError.setupFailed("Failed to create tap or device")
        }

        // Add tap to aggregate device
        try addTapToAggregateDevice(tapID: tapID, deviceID: deviceID)

        // Get audio format
        let format = try getAudioFormat(from: deviceID)

        return (deviceID, format)
    }

    private func createSystemAudioTap() throws -> AudioObjectID {
        let description = CATapDescription()

        description.name = "koemoji-system-audio-tap"
        description.processes = []  // Empty array = all system audio
        description.isPrivate = true
        description.muteBehavior = .unmuted
        description.isMixdown = true
        description.isMono = false  // Stereo
        description.isExclusive = false
        description.deviceUID = nil  // System default
        description.stream = 0  // First stream

        var tapID = AudioObjectID(kAudioObjectUnknown)
        let status = AudioHardwareCreateProcessTap(description, &tapID)

        guard status == kAudioHardwareNoError else {
            throw AudioCaptureError.tapCreationFailed(status)
        }

        return tapID
    }

    private func createAggregateDevice() throws -> AudioObjectID {
        let uid = UUID().uuidString
        let description: [String: Any] = [
            kAudioAggregateDeviceNameKey: "koemoji-aggregate-device",
            kAudioAggregateDeviceUIDKey: uid,
            kAudioAggregateDeviceSubDeviceListKey: [] as CFArray,
            kAudioAggregateDeviceMasterSubDeviceKey: 0,
            kAudioAggregateDeviceIsPrivateKey: true,
            kAudioAggregateDeviceIsStackedKey: false,
        ]

        var deviceID: AudioObjectID = 0
        let status = AudioHardwareCreateAggregateDevice(description as CFDictionary, &deviceID)

        guard status == kAudioHardwareNoError else {
            throw AudioCaptureError.aggregateDeviceCreationFailed(status)
        }

        return deviceID
    }

    private func addTapToAggregateDevice(tapID: AudioObjectID, deviceID: AudioObjectID) throws {
        // Get tap UID
        var propertyAddress = getPropertyAddress(selector: kAudioTapPropertyUID)
        var propertySize = UInt32(MemoryLayout<CFString>.stride)
        var tapUID: CFString = "" as CFString

        _ = withUnsafeMutablePointer(to: &tapUID) { tapUIDPtr in
            AudioObjectGetPropertyData(tapID, &propertyAddress, 0, nil, &propertySize, tapUIDPtr)
        }

        // Add tap to aggregate device
        propertyAddress = getPropertyAddress(selector: kAudioAggregateDevicePropertyTapList)
        let tapArray = [tapUID] as CFArray
        propertySize = UInt32(MemoryLayout<CFArray>.stride)

        let status = withUnsafePointer(to: tapArray) { ptr in
            AudioObjectSetPropertyData(deviceID, &propertyAddress, 0, nil, propertySize, ptr)
        }

        guard status == kAudioHardwareNoError else {
            throw AudioCaptureError.tapAssignmentFailed(status)
        }
    }

    private func getAudioFormat(from deviceID: AudioObjectID) throws -> AudioStreamBasicDescription {
        // Wait for device to become ready
        let deviceReadyTimeout = 2.0
        let pollInterval = 0.1
        let maxPolls = Int(deviceReadyTimeout / pollInterval)

        for poll in 1...maxPolls {
            if isDeviceReady(deviceID) {
                break
            }
            if poll == maxPolls {
                // Continue anyway, maybe it will work
                break
            }
            Thread.sleep(forTimeInterval: pollInterval)
        }

        // Retry getting format
        let maxRetries = 3
        let retryDelayMs = 20

        for attempt in 1...maxRetries {
            var propertyAddress = getPropertyAddress(
                selector: kAudioDevicePropertyStreamFormat,
                scope: kAudioDevicePropertyScopeInput
            )
            var propertySize = UInt32(MemoryLayout<AudioStreamBasicDescription>.stride)
            var format = AudioStreamBasicDescription()

            let status = AudioObjectGetPropertyData(
                deviceID,
                &propertyAddress,
                0,
                nil,
                &propertySize,
                &format
            )

            if status == kAudioHardwareNoError {
                return format
            }

            if attempt < maxRetries {
                Thread.sleep(forTimeInterval: Double(retryDelayMs) / 1000.0)
            }
        }

        throw AudioCaptureError.setupFailed("Failed to get audio format after retries")
    }

    private func isDeviceReady(_ deviceID: AudioObjectID) -> Bool {
        var address = getPropertyAddress(selector: kAudioDevicePropertyDeviceIsAlive)
        var isAlive: UInt32 = 0
        var size = UInt32(MemoryLayout<UInt32>.size)
        let status = AudioObjectGetPropertyData(deviceID, &address, 0, nil, &size, &isAlive)
        return status == kAudioHardwareNoError && isAlive == 1
    }

    func cleanup() {
        if let tapID = tapID {
            AudioHardwareDestroyProcessTap(tapID)
            self.tapID = nil
        }

        if let deviceID = deviceID {
            AudioHardwareDestroyAggregateDevice(deviceID)
            self.deviceID = nil
        }
    }
}

// MARK: - WAV File Writer

class WAVFileWriter {
    private var fileHandle: FileHandle?
    private let url: URL
    private var dataSize: UInt32 = 0
    private let format: AudioStreamBasicDescription

    init(url: URL, format: AudioStreamBasicDescription) throws {
        self.url = url
        self.format = format

        // Create file
        FileManager.default.createFile(atPath: url.path, contents: nil)
        fileHandle = try FileHandle(forWritingTo: url)

        // Write WAV header (placeholder, will update at end)
        try writeWAVHeader(dataSize: 0)
    }

    func write(_ data: Data) throws {
        guard let fileHandle = fileHandle else {
            throw AudioCaptureError.recordingFailed("File handle is nil")
        }

        fileHandle.write(data)
        dataSize += UInt32(data.count)
    }

    func close() throws {
        guard let fileHandle = fileHandle else { return }

        // Update WAV header with final data size
        try fileHandle.seek(toOffset: 0)
        try writeWAVHeader(dataSize: dataSize)

        try fileHandle.close()
        self.fileHandle = nil
    }

    private func writeWAVHeader(dataSize: UInt32) throws {
        guard let fileHandle = fileHandle else {
            throw AudioCaptureError.recordingFailed("File handle is nil")
        }

        let channels = UInt16(format.mChannelsPerFrame)
        let sampleRate = UInt32(format.mSampleRate)
        let bitsPerSample: UInt16 = 16
        let byteRate = sampleRate * UInt32(channels) * UInt32(bitsPerSample / 8)
        let blockAlign = channels * bitsPerSample / 8

        var header = Data()

        // RIFF chunk
        header.append("RIFF".data(using: .ascii)!)
        header.append(withUnsafeBytes(of: dataSize + 36) { Data($0) })
        header.append("WAVE".data(using: .ascii)!)

        // fmt chunk
        header.append("fmt ".data(using: .ascii)!)
        header.append(withUnsafeBytes(of: UInt32(16)) { Data($0) })  // Chunk size
        header.append(withUnsafeBytes(of: UInt16(1)) { Data($0) })   // Audio format (PCM)
        header.append(withUnsafeBytes(of: channels) { Data($0) })
        header.append(withUnsafeBytes(of: sampleRate) { Data($0) })
        header.append(withUnsafeBytes(of: byteRate) { Data($0) })
        header.append(withUnsafeBytes(of: blockAlign) { Data($0) })
        header.append(withUnsafeBytes(of: bitsPerSample) { Data($0) })

        // data chunk
        header.append("data".data(using: .ascii)!)
        header.append(withUnsafeBytes(of: dataSize) { Data($0) })

        fileHandle.write(header)
    }

    deinit {
        try? close()
    }
}

// MARK: - Audio Recorder

class AudioRecorder {
    private let deviceID: AudioObjectID
    private let format: AudioStreamBasicDescription
    private var ioProcID: AudioDeviceIOProcID?
    private var wavWriter: WAVFileWriter?
    private var isRecording = false

    init(deviceID: AudioObjectID, format: AudioStreamBasicDescription, outputPath: String) throws {
        self.deviceID = deviceID
        self.format = format

        let url = URL(fileURLWithPath: outputPath)
        self.wavWriter = try WAVFileWriter(url: url, format: format)
    }

    func startRecording() throws {
        let unmanagedSelf = Unmanaged.passUnretained(self).toOpaque()

        var status = AudioDeviceCreateIOProcID(
            deviceID,
            { (_, _, inInputData, _, _, _, inClientData) -> OSStatus in
                guard let clientData = inClientData else { return noErr }
                let recorder = Unmanaged<AudioRecorder>.fromOpaque(clientData).takeUnretainedValue()
                return recorder.processAudio(inInputData)
            },
            unmanagedSelf,
            &ioProcID
        )

        guard status == noErr else {
            throw AudioCaptureError.recordingFailed("Failed to create IO proc: OSStatus \(status)")
        }

        status = AudioDeviceStart(deviceID, ioProcID)
        guard status == noErr else {
            if let procID = ioProcID {
                AudioDeviceDestroyIOProcID(deviceID, procID)
            }
            throw AudioCaptureError.recordingFailed("Failed to start device: OSStatus \(status)")
        }

        isRecording = true
    }

    private func processAudio(_ inputData: UnsafePointer<AudioBufferList>) -> OSStatus {
        let bufferList = inputData.pointee
        let firstBuffer = bufferList.mBuffers

        guard firstBuffer.mData != nil && firstBuffer.mDataByteSize > 0 else {
            return noErr
        }

        let audioData = Data(bytes: firstBuffer.mData!, count: Int(firstBuffer.mDataByteSize))

        do {
            try wavWriter?.write(audioData)
        } catch {
            print("Error writing audio data: \(error)", to: &standardError)
        }

        return noErr
    }

    func stopRecording() {
        guard isRecording else { return }

        if let procID = ioProcID {
            AudioDeviceStop(deviceID, procID)
            AudioDeviceDestroyIOProcID(deviceID, procID)
        }

        do {
            try wavWriter?.close()
        } catch {
            print("Error closing WAV file: \(error)", to: &standardError)
        }

        isRecording = false
    }

    deinit {
        stopRecording()
    }
}
