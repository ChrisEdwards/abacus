package main

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"abacus/internal/beads"
)

const beadsRepoURL = "https://github.com/steveyegge/beads"

func handleVersionCheckResult(w io.Writer, info beads.VersionInfo, err error) bool {
	if err == nil {
		return false
	}

	var vErr beads.VersionError
	if errors.As(err, &vErr) {
		switch vErr.Kind {
		case beads.VersionErrorNotInstalled:
			_, _ = fmt.Fprint(w, formatBeadsNotInstalledMessage(info.Bin))
			return true
		case beads.VersionErrorTooOld:
			_, _ = fmt.Fprint(w, formatBeadsVersionTooOldMessage(info))
			return true
		case beads.VersionErrorCommandFailed, beads.VersionErrorParse:
			_, _ = fmt.Fprint(w, formatBeadsVersionWarning(info, err))
			return false
		default:
			_, _ = fmt.Fprintf(w, "Warning: Beads version check failed: %v\n", err)
			return false
		}
	}

	_, _ = fmt.Fprintf(w, "Warning: Beads version check failed: %v\n", err)
	return false
}

func formatBeadsNotInstalledMessage(bin string) string {
	if strings.TrimSpace(bin) == "" {
		bin = "bd"
	}
	return fmt.Sprintf(`Error: Beads CLI is required but not found

Abacus requires the Beads CLI to access your issue database.

What is Beads?
  Beads is a lightweight, git-based issue tracking system.
  Learn more: %[1]s

Installation:
  See download and installation instructions at:
  %[1]s

After installation, ensure the %[2]s command is in your PATH:
  export PATH="$PATH:$(go env GOPATH)/bin"

`, beadsRepoURL, bin)
}

func formatBeadsVersionTooOldMessage(info beads.VersionInfo) string {
	installed := info.Installed
	if strings.TrimSpace(installed) == "" {
		installed = "unknown"
	}
	required := info.Required
	if strings.TrimSpace(required) == "" {
		required = "v" + beads.MinBeadsVersion
	}
	bin := info.Bin
	if strings.TrimSpace(bin) == "" {
		bin = "bd"
	}
	return fmt.Sprintf(`Error: Beads CLI version too old

Your version: %s
Required:     %s or later

Upgrade Beads:
  See installation instructions at:
  %[3]s

Verify upgrade:
  %s --version

`, installed, required, beadsRepoURL, bin)
}

func formatBeadsVersionWarning(info beads.VersionInfo, err error) string {
	bin := info.Bin
	if strings.TrimSpace(bin) == "" {
		bin = "bd"
	}
	required := info.Required
	if strings.TrimSpace(required) == "" {
		required = "v" + beads.MinBeadsVersion
	}
	errorText := "unknown error"
	if err != nil {
		errorText = strings.TrimSpace(err.Error())
		if errorText == "" {
			errorText = "unknown error"
		}
	}
	return fmt.Sprintf(`Warning: Could not verify Beads CLI version

Attempted to run: %s --version
Error: %s

Abacus requires Beads CLI %s or later.

Troubleshooting:
  - Ensure %s is in your PATH: which %s
  - Check Beads installation: %s
  - Check Beads works: %s list

Continuing anyway, but you may encounter errors...

`, bin, errorText, required, bin, bin, beadsRepoURL, bin)
}
