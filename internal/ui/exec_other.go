//go:build !windows
// +build !windows

package ui

import "os/exec"

// createCommand creates a command (no special handling needed on non-Windows platforms)
func createCommand(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}
