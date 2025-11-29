package ui

import (
	"strings"
	"testing"
)

func TestTruncateVisualWidth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{
			name:     "returns empty for zero width",
			input:    "hello",
			width:    0,
			expected: "",
		},
		{
			name:     "returns empty for negative width",
			input:    "hello",
			width:    -5,
			expected: "",
		},
		{
			name:     "returns unchanged if shorter than width",
			input:    "hi",
			width:    10,
			expected: "hi",
		},
		{
			name:     "truncates to exact width",
			input:    "hello world",
			width:    5,
			expected: "hello",
		},
		{
			name:     "handles empty string",
			input:    "",
			width:    10,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateVisualWidth(tt.input, tt.width)
			if result != tt.expected {
				t.Errorf("truncateVisualWidth(%q, %d) = %q, want %q",
					tt.input, tt.width, result, tt.expected)
			}
		})
	}
}

func TestOverlayBottomRight(t *testing.T) {
	tests := []struct {
		name           string
		background     string
		overlay        string
		containerWidth int
		padding        int
		checkContains  string // Check that result contains this
	}{
		{
			name:           "returns background unchanged for empty overlay",
			background:     "line1\nline2\nline3",
			overlay:        "",
			containerWidth: 20,
			padding:        1,
			checkContains:  "line1",
		},
		{
			name:           "overlays content on background",
			background:     "aaaaaaaaaaaaaaaaaaaa\nbbbbbbbbbbbbbbbbbbbb\ncccccccccccccccccccc",
			overlay:        "XX",
			containerWidth: 20,
			padding:        1,
			checkContains:  "XX",
		},
		{
			name:           "handles single line overlay",
			background:     "test line 1\ntest line 2\ntest line 3",
			overlay:        "hi",
			containerWidth: 15,
			padding:        0,
			checkContains:  "hi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := overlayBottomRight(tt.background, tt.overlay, tt.containerWidth, tt.padding)
			if !strings.Contains(result, tt.checkContains) {
				t.Errorf("overlayBottomRight result should contain %q, got:\n%s",
					tt.checkContains, result)
			}
		})
	}
}

func TestOverlayBottomRightPreservesEmptyOverlay(t *testing.T) {
	background := "line1\nline2\nline3"
	result := overlayBottomRight(background, "", 20, 1)
	if result != background {
		t.Errorf("expected background unchanged for empty overlay, got:\n%s", result)
	}
}

func TestOverlayBottomRightPositioning(t *testing.T) {
	// Create a 5x5 background
	bg := strings.Repeat(".", 10) + "\n" +
		strings.Repeat(".", 10) + "\n" +
		strings.Repeat(".", 10) + "\n" +
		strings.Repeat(".", 10) + "\n" +
		strings.Repeat(".", 10)

	overlay := "X"
	result := overlayBottomRight(bg, overlay, 10, 0)

	// The overlay should appear somewhere in the result
	if !strings.Contains(result, "X") {
		t.Error("overlay 'X' not found in result")
	}

	// Should still have 5 lines
	lines := strings.Split(result, "\n")
	if len(lines) != 5 {
		t.Errorf("expected 5 lines, got %d", len(lines))
	}
}
