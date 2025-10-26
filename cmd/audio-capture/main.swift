import Foundation

// MARK: - Command Line Arguments

struct Arguments {
    var outputPath: String = ""
    var duration: TimeInterval = 0  // 0 = continuous until Ctrl+C
    var help: Bool = false
}

func parseArguments() -> Arguments {
    var args = Arguments()
    let arguments = CommandLine.arguments

    var i = 1
    while i < arguments.count {
        let arg = arguments[i]

        switch arg {
        case "--output", "-o":
            if i + 1 < arguments.count {
                args.outputPath = arguments[i + 1]
                i += 1
            }
        case "--duration", "-d":
            if i + 1 < arguments.count {
                args.duration = TimeInterval(arguments[i + 1]) ?? 0
                i += 1
            }
        case "--help", "-h":
            args.help = true
        default:
            break
        }

        i += 1
    }

    return args
}

func printUsage() {
    print("""
    Usage: audio-capture [OPTIONS]

    Options:
        --output, -o PATH      Output WAV file path (required)
        --duration, -d SECONDS Recording duration in seconds (0 = continuous)
        --help, -h             Show this help message

    Examples:
        audio-capture --output recording.wav --duration 10
        audio-capture -o recording.wav -d 60
    """)
}

// MARK: - Signal Handling

actor SignalHandler {
    static let shared = SignalHandler()
    private var shouldStop = false
    private var recorder: ScreenCaptureAudioRecorder?
    private var sigintSource: DispatchSourceSignal?
    private var sigtermSource: DispatchSourceSignal?

    private init() {}

    func setup(recorder: ScreenCaptureAudioRecorder) {
        self.recorder = recorder

        // Ignore default signal handling
        signal(SIGINT, SIG_IGN)
        signal(SIGTERM, SIG_IGN)

        // Set up dispatch sources for signals
        let queue = DispatchQueue.global(qos: .userInteractive)

        sigintSource = DispatchSource.makeSignalSource(signal: SIGINT, queue: queue)
        sigintSource?.setEventHandler {
            Task {
                await self.handleSignal()
            }
        }
        sigintSource?.resume()

        sigtermSource = DispatchSource.makeSignalSource(signal: SIGTERM, queue: queue)
        sigtermSource?.setEventHandler {
            Task {
                await self.handleSignal()
            }
        }
        sigtermSource?.resume()
    }

    func handleSignal() async {
        guard !shouldStop else { return }
        shouldStop = true

        print("Received stop signal, finalizing recording...", to: &standardError)

        if let recorder = recorder {
            do {
                try await recorder.stopRecording()
                print("✓ Recording stopped gracefully", to: &standardError)
            } catch {
                print("Error stopping recording: \(error)", to: &standardError)
            }
        }

        Foundation.exit(0)
    }

    func waitForever() async {
        // Keep the program running
        while !shouldStop {
            try? await Task.sleep(nanoseconds: 100_000_000) // 0.1 seconds
        }
    }
}

// MARK: - Main

@available(macOS 13.0, *)
@main
struct Main {
    static func main() async {
        let args = parseArguments()

        if args.help {
            printUsage()
            return
        }

        if args.outputPath.isEmpty {
            print("Error: Output path is required")
            printUsage()
            Foundation.exit(1)
        }

        print("Starting system audio capture...")
        print("Output: \(args.outputPath)")
        if args.duration > 0 {
            print("Duration: \(args.duration) seconds")
        } else {
            print("Duration: Continuous (press Ctrl+C to stop)")
        }

        let recorder = ScreenCaptureAudioRecorder(
            outputPath: args.outputPath,
            duration: args.duration
        )

        // Set up signal handling
        await SignalHandler.shared.setup(recorder: recorder)

        do {
            if args.duration > 0 {
                // Duration-based recording
                try await recorder.startRecording()
                print("✓ Recording completed successfully (CAF format)", to: &standardError)
            } else {
                // Continuous recording - wait for signal
                try await recorder.startRecording()
                // Keep running until signal
                await SignalHandler.shared.waitForever()
            }

        } catch {
            print("Error: \(error)", to: &standardError)
            Foundation.exit(1)
        }
    }
}
