//go:build windows
// +build windows

package ui

import (
	"os/exec"
	"syscall"
)

// CreateCommand creates a command that runs without showing a console window on Windows
func CreateCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	return cmd
}
