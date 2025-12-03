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

	out := app.renderThemeToast()
	if out == "" {
		t.Fatal("expected theme toast to render")
	}

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
