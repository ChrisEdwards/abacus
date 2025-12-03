package ui

import (
	"strings"
	"testing"
	"time"
)

func TestRenderThemeToastSpacerKeepsBackground(t *testing.T) {
	app := &App{
		themeToastVisible: true,
		themeToastName:    "dracula",
		themeToastStart:   time.Now(),
	}

	layer := app.themeToastLayer(80, 24, 2, 10)
	if layer == nil {
		t.Fatal("expected theme toast to render")
	}

	canvas := layer.Render()
	if canvas == nil {
		t.Fatal("expected canvas from theme toast layer")
	}
	out := canvas.Render()
	if strings.Contains(out, "Theme:\x1b[0m ") {
		t.Fatalf("found raw space with default background: %q", out)
	}
}

func TestHeaderVersionGapUsesBackground(t *testing.T) {
	app := &App{
		ready:    true,
		width:    80,
		height:   20,
		version:  "dev",
		repoName: "abacus",
	}

	view := app.View()
	titleSegment := styleAppHeader().Render("ABACUS vdev")
	gap := baseStyle().Render(" ")
	if !strings.Contains(view, titleSegment+gap) {
		t.Fatalf("expected themed gap after header title, got: %q", view)
	}
}

func TestViewOmitsDefaultResetGaps(t *testing.T) {
	app := &App{
		ready:                true,
		width:                100,
		height:               30,
		repoName:             "abacus",
		activeOverlay:        OverlayStatus,
		statusOverlay:        NewStatusOverlay("ab-smg0", "Snapshot", "in_progress"),
		statusToastVisible:   true,
		statusToastStart:     time.Now(),
		statusToastBeadID:    "ab-smg0",
		statusToastNewStatus: "in_progress",
	}

	view := app.View()
	if strings.Contains(view, "\x1b[0m ") {
		t.Fatalf("view contains default reset gap: %q", view)
	}
}
