package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/hirokitakamura/koemoji-go/internal/recorder"
)

func main() {
	fmt.Println("KoeMoji-Go Recorder Test v1.0")
	fmt.Println("==============================")

	// Test 1: List available devices
	fmt.Println("\n1. Testing device listing...")
	devices, err := recorder.ListDevices()
	if err != nil {
		log.Fatalf("Failed to list devices: %v", err)
	}

	if len(devices) == 0 {
		log.Fatal("No input devices found")
	}

	fmt.Printf("Found %d input devices:\n", len(devices))
	for i, device := range devices {
		prefix := fmt.Sprintf("  %d: %s", i, device.Name)
		if device.IsDefault {
			prefix += " (default)"
		}
		if device.IsVirtual {
			prefix += fmt.Sprintf(" [%s]", device.VirtualType)
		}
		fmt.Printf("%s (%d channels, %s)\n", prefix, device.MaxChannels, device.HostAPI)
	}

	// Test 2: Select device
	fmt.Println("\n2. Device selection...")
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Select device (number) or press Enter for default: ")
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())

	var rec *recorder.Recorder
	if input == "" {
		fmt.Println("Using default device...")
		rec, err = recorder.NewRecorder()
	} else {
		deviceIndex, err2 := strconv.Atoi(input)
		if err2 != nil || deviceIndex < 0 || deviceIndex >= len(devices) {
			log.Fatalf("Invalid device selection: %s", input)
		}
		selectedDevice := devices[deviceIndex]
		fmt.Printf("Using device: %s\n", selectedDevice.Name)
		rec, err = recorder.NewRecorderWithDevice(selectedDevice.ID)
	}

	if err != nil {
		log.Fatalf("Failed to create recorder: %v", err)
	}
	defer rec.Close()

	// Test 3: Interactive recording
	fmt.Println("\n3. Interactive recording test...")
	fmt.Println("Controls:")
	fmt.Println("  Enter - Start/Stop recording")
	fmt.Println("  q     - Quit")
	fmt.Println()

	recording := false
	recordingCount := 0

	for {
		if recording {
			duration := rec.GetElapsedTime()
			fmt.Printf("Recording... %.1fs (Enter to stop, q to quit): ", duration.Seconds())
		} else {
			fmt.Print("Ready (Enter to start, q to quit): ")
		}

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		if input == "q" {
			break
		}

		if input == "" {
			if recording {
				// Stop recording
				fmt.Println("Stopping recording...")
				err = rec.Stop()
				if err != nil {
					log.Printf("Stop failed: %v", err)
					continue
				}

				// Save file
				recordingCount++
				filename := fmt.Sprintf("test_recording_%d.wav", recordingCount)
				fmt.Printf("Saving to %s...", filename)
				err = rec.SaveToFile(filename)
				if err != nil {
					log.Printf("Save failed: %v", err)
					continue
				}

				duration := rec.GetDuration()
				fmt.Printf(" Done! (%.1fs)\n", duration)
				recording = false
			} else {
				// Start recording
				fmt.Println("Starting recording...")
				err = rec.Start()
				if err != nil {
					log.Printf("Start failed: %v", err)
					continue
				}
				recording = true
			}
		}
	}

	// Cleanup
	if recording {
		fmt.Println("Stopping recording...")
		rec.Stop()
		recordingCount++
		filename := fmt.Sprintf("test_recording_%d.wav", recordingCount)
		rec.SaveToFile(filename)
		fmt.Printf("Final recording saved: %s\n", filename)
	}

	fmt.Println("Test completed!")
}
