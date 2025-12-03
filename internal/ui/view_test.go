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
