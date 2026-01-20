package beads

import (
	"testing"
)

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
