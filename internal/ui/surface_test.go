package ui

import (
	"strings"
	"testing"

	"abacus/internal/ui/theme"
)

func TestNewPrimarySurfaceInitializesStyles(t *testing.T) {
	surface := NewPrimarySurface(10, 3)
	if surface.Canvas == nil {
		t.Fatal("expected canvas to be initialized")
	}
	assertAdaptiveColor(t, surface.Styles.Text.GetBackground(), theme.Current().Background(), "primary text background")
	if surface.Styles.Accent.GetForeground() != theme.Current().Accent() {
		t.Fatalf("expected Accent foreground %q, got %q", theme.Current().Accent(), surface.Styles.Accent.GetForeground())
	}
}

func TestSurfaceDrawWritesContent(t *testing.T) {
	surface := NewSecondarySurface(8, 4)
	surface.Draw(0, 1, surface.Styles.Accent.Render("HI"))

	got := stripANSI(surface.Render())
	lines := strings.Split(got, "\n")
	if len(lines) < 2 || !strings.Contains(lines[1], "HI") {
		t.Fatalf("expected drawn content on second line, got %q", got)
	}
}
