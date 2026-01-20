package beads

import (
	"testing"
)

func TestBackendConstants(t *testing.T) {
	if BackendBd != "bd" {
		t.Errorf("BackendBd = %q, want 'bd'", BackendBd)
	}
	if BackendBr != "br" {
		t.Errorf("BackendBr = %q, want 'br'", BackendBr)
	}
}

func TestMinBrVersion(t *testing.T) {
	if MinBrVersion == "" {
		t.Error("MinBrVersion should not be empty")
	}
}

func TestErrorVariables(t *testing.T) {
	if ErrNoBackendAvailable == nil {
		t.Error("ErrNoBackendAvailable should not be nil")
	}
	if ErrBackendAmbiguous == nil {
		t.Error("ErrBackendAmbiguous should not be nil")
	}
}

func TestCommandExists(t *testing.T) {
	tests := []struct {
		name   string
		binary string
		want   bool
	}{
		{
			name:   "go binary should exist",
			binary: "go",
			want:   true,
		},
		{
			name:   "nonexistent binary should not exist",
			binary: "definitely-not-a-real-binary-xyz123",
			want:   false,
		},
		{
			name:   "empty string should not exist",
			binary: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := commandExists(tt.binary)
			if got != tt.want {
				t.Errorf("commandExists(%q) = %v, want %v", tt.binary, got, tt.want)
			}
		})
	}
}

func TestIsInteractiveTTY(t *testing.T) {
	// In test environment, stdin is typically not a TTY
	// This test just ensures the function runs without panicking
	_ = isInteractiveTTY()
}

// Note: Full integration tests for DetectBackend would require mocking:
// - exec.LookPath for binary existence checks
// - version check calls
// - config file reads/writes
// - TTY detection
// - huh form interactions
//
// These are better suited for integration/e2e tests with proper fixtures.
