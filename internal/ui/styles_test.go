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

	expected := theme.Current().Background()

	assertAdaptiveColor(t, s.GetBackground(), expected, "body background")

	assertAdaptiveColor(t, s.GetBorderTopBackground(), expected, "border top background")
	assertAdaptiveColor(t, s.GetBorderRightBackground(), expected, "border right background")
	assertAdaptiveColor(t, s.GetBorderBottomBackground(), expected, "border bottom background")
	assertAdaptiveColor(t, s.GetBorderLeftBackground(), expected, "border left background")
}

func assertAdaptiveColor(t *testing.T, got lipgloss.TerminalColor, expected lipgloss.AdaptiveColor, label string) {
	t.Helper()

	adaptive, ok := got.(lipgloss.AdaptiveColor)
	if !ok {
		t.Fatalf("%s should be AdaptiveColor, got %T", label, got)
	}
	if adaptive != expected {
		t.Fatalf("%s mismatch: expected %+v, got %+v", label, expected, adaptive)
	}
}
