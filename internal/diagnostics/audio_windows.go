//go:build windows

package diagnostics

import (
	"fmt"
	"strings"

	"github.com/go-ole/go-ole"
)

var (
	dualRecordingOK       = true
	dualRecordingWarnings []string
	dualRecordingErrors   []string
)

func checkDualRecording() {
	// Check COM initialization
	err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	if err != nil {
		// S_FALSE (already initialized) is acceptable, only fail on other errors
		oleErr, ok := err.(*ole.OleError)
		if !ok || oleErr.Code() != 0x00000001 { // S_FALSE = 1
			dualRecordingOK = false
			dualRecordingErrors = append(dualRecordingErrors, fmt.Sprintf("COM initialization failed: %v", err))
			fmt.Printf("✗ COM initialization failed: %v\n", err)
			return
		}
	}
	defer ole.CoUninitialize()

	fmt.Println("✓ Windows WASAPI available")
	fmt.Println("✓ COM initialization successful")

	// Check for virtual audio devices (VoiceMeeter, Stereo Mix)
	hasVirtualDevices := false
	virtualDeviceTypes := make(map[string]bool)

	for _, device := range audioDevices {
		// Check both the IsVirtual flag and device name
		if device.IsVirtual && device.VirtualType != "" {
			hasVirtualDevices = true
			virtualDeviceTypes[device.VirtualType] = true
		} else if isVirtualDeviceName(device.Name) {
			hasVirtualDevices = true
			// Determine type from name
			name := strings.ToLower(device.Name)
			if strings.Contains(name, "voicemeeter") {
				virtualDeviceTypes["VoiceMeeter"] = true
			} else if strings.Contains(name, "cable") {
				virtualDeviceTypes["Virtual Cable"] = true
			} else if strings.Contains(name, "stereo mix") {
				virtualDeviceTypes["Stereo Mix"] = true
			}
		}
	}

	if hasVirtualDevices {
		// Convert map keys to slice
		types := make([]string, 0, len(virtualDeviceTypes))
		for t := range virtualDeviceTypes {
			types = append(types, t)
		}
		fmt.Printf("✓ Virtual audio devices detected: %v\n", types)
	} else {
		dualRecordingWarnings = append(dualRecordingWarnings, "No virtual audio devices detected (VoiceMeeter or Stereo Mix recommended for dual recording)")
		fmt.Println("⚠ No virtual audio devices detected")
		fmt.Println("  (VoiceMeeter or Stereo Mix recommended for dual recording)")
	}
}
