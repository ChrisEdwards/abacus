package ui

import (
	"strings"
	"testing"

	"abacus/internal/beads"
	"abacus/internal/graph"

	"github.com/charmbracelet/bubbles/textinput"
)

func TestSearchOverlayInputWidthRespectsBounds(t *testing.T) {
	overlay := NewSearchOverlay()

	if got, want := overlay.InputWidth(200), overlayMaxWidth-overlayHorizontalPad; got != want {
		t.Fatalf("expected max input width %d, got %d", want, got)
	}

	if got, want := overlay.InputWidth(50), overlayMinWidth-overlayHorizontalPad; got != want {
		t.Fatalf("expected min input width %d, got %d", want, got)
	}
}

func TestSearchOverlayViewRendersSuggestions(t *testing.T) {
	overlay := NewSearchOverlay()
	overlay.SetSuggestions([]string{"status:open", "assignee:me"})

	output := overlay.View("/ status:", 100, 30)
	for _, snippet := range []string{"Smart Filter Search", "• status:open", "/ status:"} {
		if !strings.Contains(output, snippet) {
			t.Fatalf("expected overlay output to contain %q\n%s", snippet, output)
		}
	}
}

func TestAppViewShowsOverlayWhileSearching(t *testing.T) {
	ti := textinput.New()
	ti.SetValue("proj")

	root := &graph.Node{Issue: beads.FullIssue{ID: "ab-001", Title: "Root"}}
	app := &App{
		roots:       []*graph.Node{root},
		visibleRows: []*graph.Node{root},
		textInput:   ti,
		searching:   true,
		overlay:     NewSearchOverlay(),
		ready:       true,
		width:       100,
		height:      30,
	}

	view := app.View()
	if !strings.Contains(view, "Smart Filter Search") {
		t.Fatalf("expected search overlay view, got:\n%s", view)
	}
	if strings.Contains(view, "ABACUS") {
		t.Fatalf("expected base UI hidden behind overlay, got:\n%s", view)
	}
}
