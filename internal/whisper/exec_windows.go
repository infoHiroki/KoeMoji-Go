//go:build windows
// +build windows

package whisper

import (
	"os/exec"
	"syscall"
)

// createCommand creates a command that runs without showing a console window on Windows
func createCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	return cmd
}
