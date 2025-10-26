import Foundation
import AVFoundation
import AudioToolbox

// MARK: - Command Line Arguments Parser

struct Arguments {
    var outputPath: String = ""
    var duration: TimeInterval = 0  // 0 = continuous until SIGINT
    var sampleRate: Double = 44100

    static func parse() -> Arguments? {
        var args = Arguments()
        let arguments = CommandLine.arguments

        var i = 1
        while i < arguments.count {
            let arg = arguments[i]

            switch arg {
            case "--output", "-o":
                guard i + 1 < arguments.count else {
                    print("Error: --output requires a path", to: &standardError)
                    return nil
                }
                args.outputPath = arguments[i + 1]
                i += 2
            case "--duration", "-d":
                guard i + 1 < arguments.count else {
                    print("Error: --duration requires a number (seconds)", to: &standardError)
                    return nil
                }
                guard let duration = TimeInterval(arguments[i + 1]) else {
                    print("Error: Invalid duration value", to: &standardError)
                    return nil
                }
                args.duration = duration
                i += 2
            case "--sample-rate", "-s":
                guard i + 1 < arguments.count else {
                    print("Error: --sample-rate requires a number", to: &standardError)
                    return nil
                }
                guard let rate = Double(arguments[i + 1]) else {
                    print("Error: Invalid sample rate value", to: &standardError)
                    return nil
                }
                args.sampleRate = rate
                i += 2
            case "--help", "-h":
                printUsage()
                return nil
            default:
                print("Error: Unknown argument: \(arg)", to: &standardError)
                printUsage()
                return nil
            }
        }

        if args.outputPath.isEmpty {
            print("Error: --output is required", to: &standardError)
            printUsage()
            return nil
        }

        return args
    }

    static func printUsage() {
        print("""
        Usage: audio-capture --output <path> [options]

        Options:
          --output, -o <path>      Output WAV file path (required)
          --duration, -d <seconds> Recording duration in seconds (0 = until Ctrl+C, default: 0)
          --sample-rate, -s <rate> Sample rate in Hz (default: 44100)
          --help, -h               Show this help message

        Example:
          audio-capture --output recording.wav --duration 10
          audio-capture -o output.wav  # Record until Ctrl+C
        """)
    }
}

// MARK: - Stderr Output

var standardError = FileHandle.standardError

extension FileHandle: @retroactive TextOutputStream {
    public func write(_ string: String) {
        guard let data = string.data(using: .utf8) else { return }
        self.write(data)
    }
}

// MARK: - Main Entry Point

guard let args = Arguments.parse() else {
    exit(1)
}

// Check macOS version (14.4+ required for CATap API)
if #unavailable(macOS 14.4) {
    print("Error: This tool requires macOS 14.4 or later", to: &standardError)
    exit(1)
}

print("Starting system audio capture...", to: &standardError)
print("Output: \(args.outputPath)", to: &standardError)
if args.duration > 0 {
    print("Duration: \(args.duration) seconds", to: &standardError)
} else {
    print("Duration: Continuous (press Ctrl+C to stop)", to: &standardError)
}
print("Sample Rate: \(args.sampleRate) Hz", to: &standardError)

// TODO: Implement system audio capture using CATap API
// This will be implemented in the next step

print("System audio capture implementation coming soon...", to: &standardError)
exit(0)
