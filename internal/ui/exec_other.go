//go:build !windows
// +build !windows

package ui

import "os/exec"

// CreateCommand creates a command (no special handling needed on non-Windows platforms)
func CreateCommand(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}
