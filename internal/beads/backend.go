// Package beads provides clients for interacting with beads issue trackers.
// This file contains backend detection logic for choosing between bd and br CLIs.
package beads

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/charmbracelet/huh"
	"golang.org/x/term"

	"abacus/internal/config"
)

// Backend constants
const (
	BackendBd = "bd" // beads Go CLI
	BackendBr = "br" // beads_rust CLI
)

// MinBrVersion defines the minimum supported br CLI version.
const MinBrVersion = "0.1.7"

// MaxSupportedBdVersion defines the maximum officially supported bd CLI version.
// Versions above this trigger a one-time warning (abacus development focuses on br).
const MaxSupportedBdVersion = "0.38.0"

// ErrNoBackendAvailable indicates neither bd nor br was found on PATH.
var ErrNoBackendAvailable = errors.New("neither bd nor br found in PATH")

// ErrBackendAmbiguous indicates both backends exist but no preference is set in non-TTY mode.
var ErrBackendAmbiguous = errors.New("both bd and br found in PATH; use --backend flag or set beads.backend in .abacus/config.yaml")

// Test hooks - these function variables allow tests to mock dependencies.
// In production, they point to the real implementations.
var (
	// commandExistsFunc is used to check if a binary exists on PATH.
	commandExistsFunc = commandExists

	// isInteractiveTTYFunc is used to check if stdin is a TTY.
	isInteractiveTTYFunc = isInteractiveTTY

	// checkBackendVersionFunc is used to validate backend versions.
	checkBackendVersionFunc = checkBackendVersion

	// configGetProjectStringFunc is used to read project config values.
	configGetProjectStringFunc = config.GetProjectString

	// configSaveBackendFunc is used to save backend preference to project config.
	configSaveBackendFunc = config.SaveBackend

	// promptUserForBackendFunc is used for interactive backend selection.
	promptUserForBackendFunc = promptUserForBackend

	// promptSwitchBackendFunc is used for interactive switch confirmation.
	promptSwitchBackendFunc = promptSwitchBackend
)

// DetectBackendOptions configures DetectBackend behavior.
type DetectBackendOptions struct {
	// CLIFlag is the value of --backend flag (empty if not provided).
	CLIFlag string
	// BeforePrompt is called before any interactive prompt (e.g., to stop animations).
	// Pass nil if no pre-prompt callback is needed.
	BeforePrompt func()
	// SkipVersionCheck skips version validation but still performs detection.
	// Use this when the user wants faster startup and accepts version risk.
	SkipVersionCheck bool
}

// DetectBackend determines which backend (bd or br) to use.
// Returns the backend name ("bd" or "br") or an error if detection fails.
//
// Priority order (regardless of SkipVersionCheck):
//  1. CLI flag (--backend)
//  2. Stored preference (.abacus/config.yaml beads.backend)
//  3. Auto-detection (which backend exists on PATH)
func DetectBackend(ctx context.Context, opts DetectBackendOptions) (string, error) {
	// 0. CLI flag override (highest priority, one-time, no save)
	if opts.CLIFlag != "" {
		if opts.CLIFlag != BackendBd && opts.CLIFlag != BackendBr {
			return "", fmt.Errorf("invalid --backend value: %q (must be 'bd' or 'br')", opts.CLIFlag)
		}
		if !commandExistsFunc(opts.CLIFlag) {
			return "", fmt.Errorf("--backend %s specified but %s not found in PATH", opts.CLIFlag, opts.CLIFlag)
		}
		if !opts.SkipVersionCheck {
			if err := checkBackendVersionFunc(ctx, opts.CLIFlag); err != nil {
				return "", fmt.Errorf("--backend %s version check failed: %w", opts.CLIFlag, err)
			}
		}
		// Don't save - CLI flag is a one-time override
		return opts.CLIFlag, nil
	}

	// 1. Check stored preference (project config ONLY - no env var support)
	storedPref := configGetProjectStringFunc(config.KeyBeadsBackend)
	if storedPref != "" {
		// Verify the stored preference is still valid (binary exists)
		if commandExistsFunc(storedPref) {
			// Version check for stored preference (already saved, just validate)
			if !opts.SkipVersionCheck {
				if err := checkBackendVersionFunc(ctx, storedPref); err != nil {
					return "", fmt.Errorf("stored backend '%s' version check failed: %w", storedPref, err)
				}
			}
			return storedPref, nil
		}
		// 1b. Stale preference - prompt user before clearing
		return handleStalePreference(ctx, storedPref, opts.BeforePrompt, opts.SkipVersionCheck)
	}

	// 2. Check binary availability (PATH only, no probing)
	brExists := commandExistsFunc(BackendBr)
	bdExists := commandExistsFunc(BackendBd)

	var choice string
	var userPrompted bool
	switch {
	case !brExists && !bdExists:
		return "", ErrNoBackendAvailable
	case brExists && !bdExists:
		choice = BackendBr
	case bdExists && !brExists:
		choice = BackendBd
	case brExists && bdExists:
		// Both exist - need user input
		if !isInteractiveTTYFunc() {
			return "", ErrBackendAmbiguous
		}
		// Stop any animations before prompting
		if opts.BeforePrompt != nil {
			opts.BeforePrompt()
		}
		choice = promptUserForBackendFunc()
		userPrompted = true
	}

	// 3. Version check BEFORE saving - allows user to switch if version fails
	if !opts.SkipVersionCheck {
		var err error
		choice, err = validateWithFallback(ctx, choice, brExists, bdExists, opts.BeforePrompt)
		if err != nil {
			return "", err
		}
	}

	// 4. Save validated choice
	// Note: SaveBackend may fail if no .beads/ directory exists, but main.go
	// validates .beads/ presence before calling DetectBackend(), so this is
	// defense-in-depth. Log warning but continue since detection succeeded.
	if err := configSaveBackendFunc(choice); err != nil {
		log.Printf("warning: could not save backend preference: %v", err)
	} else if userPrompted {
		fmt.Println("Saved to .abacus/config.yaml - edit beads.backend to change later.")
	}

	return choice, nil
}

// checkBackendVersion validates the backend meets minimum version requirements.
func checkBackendVersion(ctx context.Context, backend string) error {
	minVersion := MinBeadsVersion // default for bd
	if backend == BackendBr {
		minVersion = MinBrVersion
	}

	_, err := CheckVersion(ctx, VersionCheckOptions{
		Bin:        backend,
		MinVersion: minVersion,
	})
	return err
}

// handleStalePreference handles the case where stored preference points to
// a binary that's no longer on PATH.
func handleStalePreference(ctx context.Context, storedPref string, beforePrompt func(), skipVersionCheck bool) (string, error) {
	// Determine which binary (if any) is available as alternative
	other := BackendBd
	if storedPref == BackendBd {
		other = BackendBr
	}
	otherExists := commandExistsFunc(other)

	if !otherExists {
		return "", fmt.Errorf("this project is configured for '%s' but neither bd nor br found in PATH", storedPref)
	}

	// In non-TTY mode, we can't prompt - return error
	if !isInteractiveTTYFunc() {
		return "", fmt.Errorf("this project is configured for '%s' but %s is not found in PATH; use --backend %s to override", storedPref, storedPref, other)
	}

	// Stop any animations before prompting
	if beforePrompt != nil {
		beforePrompt()
	}

	// Prompt user: their configured backend is missing, offer to switch
	fmt.Printf("This project is configured for '%s' but %s is not found in PATH.\n", storedPref, storedPref)
	if confirmed := promptSwitchBackendFunc(other); confirmed {
		// Version check BEFORE saving (unless skipped)
		if !skipVersionCheck {
			if err := checkBackendVersionFunc(ctx, other); err != nil {
				return "", fmt.Errorf("cannot switch to %s: %w", other, err)
			}
		}
		if err := configSaveBackendFunc(other); err != nil {
			log.Printf("warning: could not save backend preference: %v", err)
		}
		return other, nil
	}

	// User declined to switch - exit with helpful message
	return "", fmt.Errorf("cannot continue: '%s' not found in PATH (add it to PATH or accept switch to '%s')", storedPref, other)
}

// validateWithFallback validates the chosen backend's version and offers
// to switch to the alternative if validation fails.
func validateWithFallback(ctx context.Context, choice string, brExists, bdExists bool, beforePrompt func()) (string, error) {
	if err := checkBackendVersionFunc(ctx, choice); err == nil {
		return choice, nil // Version check passed
	}

	// Version check failed - is there an alternative?
	other := BackendBd
	if choice == BackendBd {
		other = BackendBr
	}
	otherExists := (other == BackendBr && brExists) || (other == BackendBd && bdExists)

	if !otherExists {
		return "", fmt.Errorf("%s version is too old (see requirements) and no alternative backend available", choice)
	}

	// In non-TTY mode, we can't prompt - return error
	if !isInteractiveTTYFunc() {
		return "", fmt.Errorf("%s version is too old; use --backend %s to try alternative", choice, other)
	}

	// Stop any animations before prompting
	if beforePrompt != nil {
		beforePrompt()
	}

	// Offer to switch to the other backend
	fmt.Printf("%s version is too old. Would you like to use %s instead?\n", choice, other)
	if confirmed := promptSwitchBackendFunc(other); confirmed {
		// Check the alternative's version too
		if err := checkBackendVersionFunc(ctx, other); err != nil {
			return "", fmt.Errorf("both backends have version issues: %s and %s", choice, other)
		}
		return other, nil
	}

	return "", fmt.Errorf("cannot continue: %s version too old and user declined switch to %s", choice, other)
}

// promptUserForBackend prompts the user to select between bd and br backends.
// Uses huh library for a nice interactive selection UI.
// Returns the selected backend ("bd" or "br").
func promptUserForBackend() string {
	var choice string
	form := huh.NewSelect[string]().
		Title("Both bd and br are available. Which backend does this project use?").
		Options(
			huh.NewOption("br (recommended)", BackendBr),
			huh.NewOption("bd", BackendBd),
		).
		Value(&choice)

	if err := form.Run(); err != nil {
		// If form fails (e.g., interrupted), default to br
		return BackendBr
	}

	return choice
}

// promptSwitchBackend asks the user to confirm switching to an alternative backend.
// Returns true if user confirms the switch.
func promptSwitchBackend(other string) bool {
	var confirmed bool
	form := huh.NewConfirm().
		Title(fmt.Sprintf("Switch to %s?", other)).
		Description("Your preference will be updated.").
		Affirmative("Yes").
		Negative("No").
		Value(&confirmed)

	if err := form.Run(); err != nil {
		return false
	}
	return confirmed
}

// commandExists checks if a binary exists on PATH using exec.LookPath.
// Returns true if the binary is found, false otherwise.
// Does not execute the binary or check its version - just existence.
func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// isInteractiveTTY checks if stdin is connected to an interactive terminal.
// Used to determine if we can prompt the user for backend selection.
func isInteractiveTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// NewClientForBackend creates the appropriate Client based on backend string.
// backend must be "bd" or "br". dbPath is the path to the SQLite database.
// Returns an error for unknown backends or empty dbPath.
func NewClientForBackend(backend, dbPath string) (Client, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("dbPath is required")
	}
	switch backend {
	case BackendBd:
		return NewBdSQLiteClient(dbPath), nil
	case BackendBr:
		return NewBrSQLiteClient(dbPath), nil
	default:
		return nil, fmt.Errorf("unknown backend: %q (must be %q or %q)", backend, BackendBd, BackendBr)
	}
}

// CheckBdVersionWarning shows a one-time warning if bd version > MaxSupportedBdVersion.
// The warning is non-blocking and only shown once per user (stored in ~/.abacus/config.yaml).
// Call this after DetectBackend returns "bd" successfully.
func CheckBdVersionWarning(ctx context.Context) {
	// Only applies to bd backend
	// Check if warning already shown
	if config.GetBool(config.KeyBdUnsupportedVersionWarnShown) {
		return
	}

	// Get bd version
	info, err := CheckVersion(ctx, VersionCheckOptions{
		Bin:        BackendBd,
		MinVersion: MinBeadsVersion,
	})
	if err != nil {
		// Version check failed - don't show warning (probably already shown error elsewhere)
		return
	}

	// Compare with max supported version
	installedSemver, _, err := parseSemver(info.Installed)
	if err != nil {
		return
	}
	maxSemver, _, err := parseSemver(MaxSupportedBdVersion)
	if err != nil {
		return
	}

	// If installed version <= max supported, no warning needed
	if installedSemver.compare(maxSemver) <= 0 {
		return
	}

	// Show one-time warning
	fmt.Printf("\n")
	fmt.Printf("Note: abacus officially supports beads (bd) up to version %s.\n", MaxSupportedBdVersion)
	fmt.Printf("You have version %s installed. Newer versions may work but are\n", info.Installed)
	fmt.Printf("not officially supported.\n")
	fmt.Printf("\n")
	fmt.Printf("Future development is focused on beads_rust (br):\n")
	fmt.Printf("https://github.com/Dicklesworthstone/beads_rust\n")
	fmt.Printf("\n")
	fmt.Printf("This message will not be shown again.\n")
	fmt.Printf("\n")

	// Save flag to prevent showing again
	if err := config.SaveBdUnsupportedVersionWarned(); err != nil {
		log.Printf("warning: could not save bd version warning flag: %v", err)
	}
}
