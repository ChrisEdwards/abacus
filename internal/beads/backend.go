// Package beads provides clients for interacting with beads issue trackers.
// This file contains backend detection logic for choosing between bd and br CLIs.
package beads

import (
	"os"
	"os/exec"

	"golang.org/x/term"
)

// commandExists checks if a binary exists on PATH using exec.LookPath.
// Returns true if the binary is found, false otherwise.
// Does not execute the binary or check its version - just existence.
//
//nolint:unused // Will be used by detectBackend() in ab-pccw.3.2
func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// isInteractiveTTY checks if stdin is connected to an interactive terminal.
// Used to determine if we can prompt the user for backend selection.
//
//nolint:unused // Will be used by detectBackend() in ab-pccw.3.2
func isInteractiveTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}
