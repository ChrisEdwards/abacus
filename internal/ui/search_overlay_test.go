package ui

import (
	"strings"
	"testing"

	"abacus/internal/beads"
	"abacus/internal/graph"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
	overlay.SetSuggestions([]suggestionEntry{{Display: "status:open", Value: "open"}, {Display: "assignee:me", Value: "me"}})
	overlay.UpdateInput("status:open")

	output := overlay.View("/ status:", 100, 30)
	for _, snippet := range []string{"Smart Filter Search", "STATUS", "open", "status:open", "/ status:"} {
		if !strings.Contains(output, snippet) {
			t.Fatalf("expected overlay output to contain %q\n%s", snippet, output)
		}
	}
}

func TestSearchOverlayCursorMovement(t *testing.T) {
	overlay := NewSearchOverlay()
	overlay.SetSuggestions([]suggestionEntry{{Display: "status:open", Value: "open"}, {Display: "status:closed", Value: "closed"}})
	if got := overlay.SelectedSuggestion(); got != "status:open" {
		t.Fatalf("expected initial selection status:open, got %q", got)
	}
	overlay.CursorDown()
	if got := overlay.SelectedSuggestion(); got != "status:closed" {
		t.Fatalf("expected cursor down to select status:closed, got %q", got)
	}
	overlay.CursorUp()
	if got := overlay.SelectedSuggestion(); got != "status:open" {
		t.Fatalf("expected cursor up to return to status:open, got %q", got)
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

func TestAppSearchNavigationUsesSuggestionList(t *testing.T) {
	ti := textinput.New()
	ti.SetValue("status:")
	root := &graph.Node{Issue: beads.FullIssue{ID: "ab-010", Title: "Root"}}
	app := &App{
		roots:       []*graph.Node{root},
		visibleRows: []*graph.Node{root},
		textInput:   ti,
		searching:   true,
		overlay:     NewSearchOverlay(),
		ready:       true,
		width:       80,
		height:      25,
	}
	app.overlay.SetSuggestions([]suggestionEntry{{Display: "status:open", Value: "open"}, {Display: "status:closed", Value: "closed"}})

	app.Update(tea.KeyMsg{Type: tea.KeyDown})
	if got := app.overlay.SelectedSuggestion(); got != "status:closed" {
		t.Fatalf("expected down key to advance selection, got %q", got)
	}
	app.Update(tea.KeyMsg{Type: tea.KeyUp})
	if got := app.overlay.SelectedSuggestion(); got != "status:open" {
		t.Fatalf("expected up key to move selection back, got %q", got)
	}
}

func TestSearchOverlayShowsParseError(t *testing.T) {
	overlay := NewSearchOverlay()
	overlay.UpdateInput(`status:"open`)
	view := overlay.View("/ status:", 80, 24)
	if !strings.Contains(view, "unterminated quote") {
		t.Fatalf("expected parse error rendered in overlay:\n%s", view)
	}
}
