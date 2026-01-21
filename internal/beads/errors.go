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

// CLIError wraps errors coming from invoking a beads CLI (bd or br).
type CLIError struct {
	Binary  string
	Command []string
	Output  string
	Err     error
}

func (e CLIError) Error() string {
	bin := e.Binary
	if bin == "" {
		bin = "bd" // default for backward compatibility
	}
	if e.Output != "" {
		return fmt.Sprintf("%s %v failed: %s", bin, e.Command, e.Output)
	}
	return fmt.Sprintf("%s %v failed: %v", bin, e.Command, e.Err)
}

func (e CLIError) Unwrap() error {
	return e.Err
}

func classifyCLIError(binary string, cmd []string, err error, output string) error {
	if errors.Is(err, exec.ErrNotFound) {
		return appErrors.New(appErrors.CodeCLINotFound, fmt.Sprintf("%s binary not found in PATH", binary), err)
	}
	if errors.Is(err, ErrNotFound) {
		return appErrors.New(appErrors.CodeNotFound, "issue not found", err)
	}
	return appErrors.New(appErrors.CodeCLIFailed, CLIError{Binary: binary, Command: cmd, Output: output, Err: err}.Error(), err)
}
