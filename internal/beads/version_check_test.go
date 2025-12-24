package beads

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

type stubRunner struct {
	output string
	err    error
}

func (s stubRunner) Run(ctx context.Context, bin string, args ...string) ([]byte, error) {
	if s.err != nil {
		return nil, s.err
	}
	return []byte(s.output), nil
}

func TestParseSemverExtractsFromText(t *testing.T) {
	raw := "Beads CLI v0.24.1 (amd64)"
	sem, normalized, err := parseSemver(raw)
	if err != nil {
		t.Fatalf("parseSemver returned error: %v", err)
	}
	if normalized != "v0.24.1" {
		t.Fatalf("expected normalized v0.24.1, got %s", normalized)
	}
	if sem.major != 0 || sem.minor != 24 || sem.patch != 1 {
		t.Fatalf("unexpected semver %+v", sem)
	}
}

func TestParseSemverRejectsInvalid(t *testing.T) {
	if _, _, err := parseSemver("no version here"); err == nil {
		t.Fatalf("expected error for invalid semver input")
	}
}

func TestSemverCompare(t *testing.T) {
	base := semver{major: 0, minor: 24, patch: 0}
	if cmp := base.compare(semver{major: 0, minor: 24, patch: 0}); cmp != 0 {
		t.Fatalf("expected equality, got %d", cmp)
	}
	if cmp := base.compare(semver{major: 0, minor: 25, patch: 0}); cmp >= 0 {
		t.Fatalf("expected less-than comparison, got %d", cmp)
	}
	if cmp := base.compare(semver{major: 0, minor: 23, patch: 9}); cmp <= 0 {
		t.Fatalf("expected greater-than comparison, got %d", cmp)
	}
}

func TestCheckVersionNoBinary(t *testing.T) {
	opts := VersionCheckOptions{
		LookPath: func(string) (string, error) {
			return "", errors.New("not found")
		},
	}
	_, err := CheckVersion(context.Background(), opts)
	var vErr VersionError
	if !errors.As(err, &vErr) {
		t.Fatalf("expected VersionError, got %v", err)
	}
	if vErr.Kind != VersionErrorNotInstalled {
		t.Fatalf("expected not installed, got %s", vErr.Kind)
	}
}

func TestCheckVersionCommandFailure(t *testing.T) {
	runErr := errors.New("boom")
	opts := VersionCheckOptions{
		LookPath: func(bin string) (string, error) {
			return fmt.Sprintf("/usr/bin/%s", bin), nil
		},
		Runner: stubRunner{err: runErr},
	}
	_, err := CheckVersion(context.Background(), opts)
	var vErr VersionError
	if !errors.As(err, &vErr) {
		t.Fatalf("expected VersionError, got %v", err)
	}
	if vErr.Kind != VersionErrorCommandFailed {
		t.Fatalf("expected command failure, got %s", vErr.Kind)
	}
}

func TestCheckVersionParseFailure(t *testing.T) {
	opts := VersionCheckOptions{
		LookPath: func(bin string) (string, error) {
			return "/usr/bin/" + bin, nil
		},
		Runner: stubRunner{output: "nonsense"},
	}
	_, err := CheckVersion(context.Background(), opts)
	var vErr VersionError
	if !errors.As(err, &vErr) {
		t.Fatalf("expected VersionError, got %v", err)
	}
	if vErr.Kind != VersionErrorParse {
		t.Fatalf("expected parse failure, got %s", vErr.Kind)
	}
}

func TestCheckVersionTooOld(t *testing.T) {
	opts := VersionCheckOptions{
		LookPath: func(bin string) (string, error) {
			return "/usr/bin/" + bin, nil
		},
		Runner:     stubRunner{output: "Beads CLI v0.29.9"},
		MinVersion: "0.30.0",
	}
	info, err := CheckVersion(context.Background(), opts)
	var vErr VersionError
	if !errors.As(err, &vErr) {
		t.Fatalf("expected VersionError, got %v", err)
	}
	if vErr.Kind != VersionErrorTooOld {
		t.Fatalf("expected too old, got %s", vErr.Kind)
	}
	if info.Installed != "v0.29.9" {
		t.Fatalf("expected installed version recorded, got %s", info.Installed)
	}
	if info.Required != "v0.30.0" {
		t.Fatalf("expected required version recorded, got %s", info.Required)
	}
}

func TestCheckVersionSuccess(t *testing.T) {
	opts := VersionCheckOptions{
		LookPath: func(bin string) (string, error) {
			return "/opt/" + bin, nil
		},
		Runner:     stubRunner{output: "Beads CLI 0.30.1"},
		MinVersion: "0.30.0",
	}
	info, err := CheckVersion(context.Background(), opts)
	if err != nil {
		t.Fatalf("expected success, got error %v", err)
	}
	if info.Installed != "v0.30.1" {
		t.Fatalf("expected normalized installed version, got %s", info.Installed)
	}
	if info.Required != "v0.30.0" {
		t.Fatalf("expected normalized required version, got %s", info.Required)
	}
}
