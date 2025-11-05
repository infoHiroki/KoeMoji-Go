package diagnostics

import (
	"fmt"
)

// Run executes the diagnostic checks and prints results to stdout
func Run() {
	printHeader()

	// System Information
	fmt.Println("\n[System Information]")
	checkSystem()

	// Audio Devices
	fmt.Println("\n[Audio Devices]")
	checkAudioDevices()

	// Dual Recording Support
	fmt.Println("\n[Dual Recording Support]")
	checkDualRecording()

	// Configuration File
	fmt.Println("\n[Configuration]")
	checkConfiguration()

	// Summary
	fmt.Println("\n[Summary]")
	printSummary()
}

func printHeader() {
	version := getVersion()
	fmt.Println("====================================")
	fmt.Printf("  KoeMoji-Go Doctor v%s\n", version)
	fmt.Println("====================================")
}
