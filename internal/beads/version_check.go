package beads

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// MinBeadsVersion defines the minimum supported Beads CLI version.
const MinBeadsVersion = "0.24.0"

// VersionInfo captures metadata about the Beads CLI discovered during checks.
type VersionInfo struct {
	Bin       string
	Installed string
	Required  string
}

// VersionErrorKind categorizes failures encountered while validating the CLI.
type VersionErrorKind string

const (
	VersionErrorUnknown       VersionErrorKind = "unknown"
	VersionErrorNotInstalled  VersionErrorKind = "not_installed"
	VersionErrorCommandFailed VersionErrorKind = "command_failed"
	VersionErrorParse         VersionErrorKind = "parse_failed"
	VersionErrorTooOld        VersionErrorKind = "too_old"
)

// VersionError wraps failures with their category and optional inner error.
type VersionError struct {
	Kind VersionErrorKind
	Info VersionInfo
	Err  error
}

// Error implements the error interface.
func (e VersionError) Error() string {
	switch {
	case e.Err != nil:
		return e.Err.Error()
	case e.Kind == VersionErrorNotInstalled:
		return "beads cli not found"
	case e.Kind == VersionErrorTooOld:
		return "beads cli version too old"
	case e.Kind == VersionErrorParse:
		return "failed to parse beads cli version"
	case e.Kind == VersionErrorCommandFailed:
		return "failed to run beads cli command"
	default:
		return "beads version check failed"
	}
}

// Unwrap exposes the wrapped error.
func (e VersionError) Unwrap() error {
	return e.Err
}

// CommandRunner executes external commands, allowing tests to inject stubs.
type CommandRunner interface {
	Run(ctx context.Context, bin string, args ...string) ([]byte, error)
}

type execCommandRunner struct{}

func (execCommandRunner) Run(ctx context.Context, bin string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, bin, args...)
	return cmd.CombinedOutput()
}

// LookPathFunc resolves a binary reference to an executable path.
type LookPathFunc func(bin string) (string, error)

// VersionCheckOptions configure how CheckVersion validates the CLI.
type VersionCheckOptions struct {
	Bin        string
	MinVersion string
	Runner     CommandRunner
	LookPath   LookPathFunc
}

// CheckVersion validates that the Beads CLI exists and satisfies the minimum version.
// It returns VersionInfo populated with discovered details and, when applicable,
// a VersionError describing the failure category.
func CheckVersion(ctx context.Context, opts VersionCheckOptions) (VersionInfo, error) {
	bin := strings.TrimSpace(opts.Bin)
	if bin == "" {
		bin = "bd"
	}
	minVersion := strings.TrimSpace(opts.MinVersion)
	if minVersion == "" {
		minVersion = MinBeadsVersion
	}
	minSemver, normalizedMin, err := parseSemver(minVersion)
	if err != nil {
		return VersionInfo{}, VersionError{
			Kind: VersionErrorParse,
			Info: VersionInfo{Bin: bin, Required: normalizedMin},
			Err:  fmt.Errorf("parse minimum version %q: %w", minVersion, err),
		}
	}

	info := VersionInfo{
		Bin:      bin,
		Required: normalizedMin,
	}

	lookPath := opts.LookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	runner := opts.Runner
	if runner == nil {
		runner = execCommandRunner{}
	}

	resolvedBin, err := lookPath(bin)
	if err != nil {
		return info, VersionError{
			Kind: VersionErrorNotInstalled,
			Info: info,
			Err:  err,
		}
	}

	out, err := runner.Run(ctx, resolvedBin, "--version")
	if err != nil {
		return info, VersionError{
			Kind: VersionErrorCommandFailed,
			Info: info,
			Err:  err,
		}
	}

	installedSemver, normalizedInstalled, err := parseSemver(string(out))
	if err != nil {
		return info, VersionError{
			Kind: VersionErrorParse,
			Info: info,
			Err:  err,
		}
	}
	info.Installed = normalizedInstalled

	if installedSemver.compare(minSemver) < 0 {
		return info, VersionError{
			Kind: VersionErrorTooOld,
			Info: info,
		}
	}

	return info, nil
}

var semverRegex = regexp.MustCompile(`v?(\d+)\.(\d+)\.(\d+)`)

type semver struct {
	major int
	minor int
	patch int
}

func (s semver) compare(other semver) int {
	if s.major != other.major {
		return compareInt(s.major, other.major)
	}
	if s.minor != other.minor {
		return compareInt(s.minor, other.minor)
	}
	if s.patch != other.patch {
		return compareInt(s.patch, other.patch)
	}
	return 0
}

func compareInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func parseSemver(input string) (semver, string, error) {
	match := semverRegex.FindStringSubmatch(strings.TrimSpace(input))
	if match == nil {
		return semver{}, "", fmt.Errorf("no semantic version found in %q", strings.TrimSpace(input))
	}
	major, err := strconv.Atoi(match[1])
	if err != nil {
		return semver{}, "", fmt.Errorf("parse major %q: %w", match[1], err)
	}
	minor, err := strconv.Atoi(match[2])
	if err != nil {
		return semver{}, "", fmt.Errorf("parse minor %q: %w", match[2], err)
	}
	patch, err := strconv.Atoi(match[3])
	if err != nil {
		return semver{}, "", fmt.Errorf("parse patch %q: %w", match[3], err)
	}
	return semver{major: major, minor: minor, patch: patch}, fmt.Sprintf("v%d.%d.%d", major, minor, patch), nil
}
