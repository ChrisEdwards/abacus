package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"

	"abacus/internal/ui/theme"
)

// toasts float above the main UI, so they must explicitly set a background
// color instead of inheriting the terminal default.

func TestStyleSuccessToastUsesThemeBackground(t *testing.T) {
	expectToastBackground(t, styleSuccessToast())
}

func TestStyleErrorToastUsesThemeBackground(t *testing.T) {
	expectToastBackground(t, styleErrorToast())
}

func expectToastBackground(t *testing.T, s lipgloss.Style) {
	t.Helper()

	bg := s.GetBackground()
	adaptive, ok := bg.(lipgloss.AdaptiveColor)
	if !ok {
		t.Fatalf("background should be AdaptiveColor, got %T", bg)
	}

	expected := theme.Current().Background()
	if adaptive != expected {
		t.Fatalf("toast background mismatch: expected %+v, got %+v", expected, adaptive)
	}
}
