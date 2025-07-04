//go:build windows
// +build windows

package ui

import (
	"os/exec"
	"syscall"
)

// createCommand creates a command that runs without showing a console window on Windows
func createCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	
	// Special handling for explorer.exe - it doesn't work well with HideWindow
	if name != "explorer" && name != "explorer.exe" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow: true,
			CreationFlags: 0x08000000, // CREATE_NO_WINDOW
		}
	}
	
	return cmd
}
