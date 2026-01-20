package beads

import (
	"errors"
	"fmt"
	"os/exec"

	appErrors "abacus/internal/errors"
)

var (
	// ErrNotFound indicates the CLI could not find the requested issue.
	ErrNotFound = errors.New("beads: issue not found")
)

// CLIError wraps errors coming from invoking the bd CLI.
type CLIError struct {
	Command []string
	Output  string
	Err     error
}

func (e CLIError) Error() string {
	if e.Output != "" {
		return fmt.Sprintf("bd %v failed: %s", e.Command, e.Output)
	}
	return fmt.Sprintf("bd %v failed: %v", e.Command, e.Err)
}

func (e CLIError) Unwrap() error {
	return e.Err
}

func classifyCLIError(cmd []string, err error, output string) error {
	if errors.Is(err, exec.ErrNotFound) {
		return appErrors.New(appErrors.CodeCLINotFound, "bd binary not found in PATH", err)
	}
	if errors.Is(err, ErrNotFound) {
		return appErrors.New(appErrors.CodeNotFound, "issue not found", err)
	}
	return appErrors.New(appErrors.CodeCLIFailed, CLIError{Command: cmd, Output: output, Err: err}.Error(), err)
}
