package ui

import "testing"

func TestFormatStatusLabel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"open", "Open"},
		{"in_progress", "In Progress"},
		{"closed", "Closed"},
		{"blocked", "Blocked"},
		{"deferred", "Deferred"},
		{"unknown", "unknown"}, // Unknown statuses pass through unchanged
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatStatusLabel(tt.input)
			if result != tt.expected {
				t.Errorf("formatStatusLabel(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
