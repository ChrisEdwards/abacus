package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"abacus/internal/beads"
)

func TestHandleVersionCheckResultNoError(t *testing.T) {
	var buf bytes.Buffer
	if exit := handleVersionCheckResult(&buf, beads.VersionInfo{}, nil); exit {
		t.Fatalf("expected no exit on nil error")
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no output, got %q", buf.String())
	}
}

func TestHandleVersionCheckResultNotInstalled(t *testing.T) {
	var buf bytes.Buffer
	err := beads.VersionError{Kind: beads.VersionErrorNotInstalled, Info: beads.VersionInfo{Bin: "bd"}}
	if exit := handleVersionCheckResult(&buf, err.Info, err); !exit {
		t.Fatalf("expected exit for missing CLI")
	}
	out := buf.String()
	if !strings.Contains(out, "Beads CLI is required but not found") {
		t.Fatalf("expected missing CLI messaging, got %q", out)
	}
}

func TestHandleVersionCheckResultTooOld(t *testing.T) {
	var buf bytes.Buffer
	info := beads.VersionInfo{Bin: "bd", Installed: "v0.20.0", Required: "v0.25.0"}
	err := beads.VersionError{Kind: beads.VersionErrorTooOld, Info: info}
	if exit := handleVersionCheckResult(&buf, info, err); !exit {
		t.Fatalf("expected exit for outdated CLI")
	}
	out := buf.String()
	if !strings.Contains(out, "version too old") {
		t.Fatalf("expected version too old text, got %q", out)
	}
	if !strings.Contains(out, "v0.20.0") || !strings.Contains(out, "v0.25.0") {
		t.Fatalf("expected version numbers in output: %q", out)
	}
}

func TestHandleVersionCheckResultWarning(t *testing.T) {
	var buf bytes.Buffer
	info := beads.VersionInfo{Bin: "bd", Required: "v0.25.0"}
	err := beads.VersionError{Kind: beads.VersionErrorCommandFailed, Info: info, Err: errors.New("timeout")}
	if exit := handleVersionCheckResult(&buf, info, err); exit {
		t.Fatalf("expected to continue on warning")
	}
	out := buf.String()
	if !strings.Contains(out, "Warning: Could not verify Beads CLI version") {
		t.Fatalf("expected warning header, got %q", out)
	}
	if !strings.Contains(out, "timeout") {
		t.Fatalf("expected original error included, got %q", out)
	}
}
