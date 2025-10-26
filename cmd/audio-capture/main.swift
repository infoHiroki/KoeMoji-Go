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

        do {
            // Simply start recording (duration is handled inside recorder)
            try await recorder.startRecording()
            print("✓ Recording completed successfully")

            // Convert CAF to WAV if output is .wav
            if args.outputPath.hasSuffix(".wav") {
                print("Converting to WAV format...", to: &standardError)
                let cafPath = args.outputPath.replacingOccurrences(of: ".wav", with: ".caf")

                // Rename the output to .caf temporarily
                try FileManager.default.moveItem(atPath: args.outputPath, toPath: cafPath)

                // Convert using afconvert
                let process = Process()
                process.executableURL = URL(fileURLWithPath: "/usr/bin/afconvert")
                process.arguments = [
                    "-f", "WAVE",      // WAV format
                    "-d", "LEF32",     // Little-endian float 32
                    cafPath,
                    args.outputPath
                ]

                try process.run()
                process.waitUntilExit()

                if process.terminationStatus == 0 {
                    // Remove the temporary CAF file
                    try FileManager.default.removeItem(atPath: cafPath)
                    print("✓ Converted to WAV successfully")
                } else {
                    print("Warning: Conversion failed, keeping CAF format", to: &standardError)
                    try FileManager.default.moveItem(atPath: cafPath, toPath: args.outputPath)
                }
            }

        } catch {
            print("Error: \(error)", to: &standardError)
            Foundation.exit(1)
        }
    }
}
