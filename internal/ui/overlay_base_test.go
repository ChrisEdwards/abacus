package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestTruncateByDisplayWidth(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxWidth int
		want     string
	}{
		{
			name:     "no truncation needed",
			text:     "short",
			maxWidth: 10,
			want:     "short",
		},
		{
			name:     "exact fit",
			text:     "exact",
			maxWidth: 5,
			want:     "exact",
		},
		{
			name:     "truncation with ellipsis",
			text:     "this is a long title",
			maxWidth: 10,
			want:     "this is...",
		},
		{
			name:     "very short max width",
			text:     "hello",
			maxWidth: 4,
			want:     "h...", // Can fit 1 char + ellipsis
		},
		{
			name:     "empty string",
			text:     "",
			maxWidth: 10,
			want:     "",
		},
		{
			name:     "zero width",
			text:     "hello",
			maxWidth: 0,
			want:     "",
		},
		{
			name:     "negative width",
			text:     "hello",
			maxWidth: -1,
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateByDisplayWidth(tt.text, tt.maxWidth)
			if got != tt.want {
				t.Errorf("truncateByDisplayWidth(%q, %d) = %q, want %q",
					tt.text, tt.maxWidth, got, tt.want)
			}
			// Verify result fits within max width
			if tt.maxWidth > 0 && lipgloss.Width(got) > tt.maxWidth {
				t.Errorf("result width %d exceeds maxWidth %d",
					lipgloss.Width(got), tt.maxWidth)
			}
		})
	}
}

func TestFormatOverlayBeadLine(t *testing.T) {
	// Use unstyled styles for predictable width calculation in tests
	plainStyle := lipgloss.NewStyle()

	tests := []struct {
		name     string
		prefix   string
		id       string
		title    string
		maxWidth int
		wantID   bool // Should contain full ID
	}{
		{
			name:     "simple bead line",
			prefix:   "  ",
			id:       "ab-123",
			title:    "Test title",
			maxWidth: 40,
			wantID:   true,
		},
		{
			name:     "with tree prefix",
			prefix:   "└─ ",
			id:       "ab-xyz",
			title:    "Child bead",
			maxWidth: 35,
			wantID:   true,
		},
		{
			name:     "long title truncation",
			prefix:   "  ",
			id:       "ab-abc",
			title:    "This is a very long title that should definitely be truncated",
			maxWidth: 30,
			wantID:   true,
		},
		{
			name:     "nested child",
			prefix:   "    └─ ",
			id:       "ab-nested",
			title:    "Deeply nested child",
			maxWidth: 40,
			wantID:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatOverlayBeadLine(tt.prefix, tt.id, tt.title, tt.maxWidth, plainStyle, plainStyle)

			// Check that ID is present and not truncated
			if tt.wantID && !strings.Contains(got, tt.id) {
				t.Errorf("formatOverlayBeadLine() result %q does not contain ID %q", got, tt.id)
			}

			// Check that result starts with prefix
			if !strings.HasPrefix(got, tt.prefix) {
				t.Errorf("formatOverlayBeadLine() result %q does not start with prefix %q", got, tt.prefix)
			}

			// Verify width doesn't exceed max (accounting for ANSI codes if styled)
			gotWidth := lipgloss.Width(got)
			if gotWidth > tt.maxWidth {
				t.Errorf("formatOverlayBeadLine() result width %d exceeds maxWidth %d", gotWidth, tt.maxWidth)
			}
		})
	}
}

func TestFormatOverlayBeadLine_IDNeverTruncated(t *testing.T) {
	plainStyle := lipgloss.NewStyle()

	// Even with very narrow width, the ID should never be truncated
	id := "ab-verylongid123"
	title := "Some title"
	prefix := "└─ "

	// Width that's barely larger than prefix + ID
	narrowWidth := lipgloss.Width(prefix) + lipgloss.Width(id) + 6

	got := formatOverlayBeadLine(prefix, id, title, narrowWidth, plainStyle, plainStyle)

	if !strings.Contains(got, id) {
		t.Errorf("ID %q was truncated in result %q (width=%d)", id, got, narrowWidth)
	}
}
