package ui

import (
	"reflect"
	"strings"
	"testing"

	"abacus/internal/beads"
	"abacus/internal/graph"

	"github.com/charmbracelet/bubbles/textinput"
)

func TestBuildValueCacheCollectsUniqueValues(t *testing.T) {
	root := &graph.Node{Issue: beads.FullIssue{Status: "open", Priority: 1, IssueType: "bug", Labels: []string{"alpha", "beta"}}}
	child := &graph.Node{Issue: beads.FullIssue{Status: "closed", Priority: 2, IssueType: "feature", Labels: []string{"beta", "gamma"}}}
	root.Children = []*graph.Node{child}
	app := &App{roots: []*graph.Node{root}}
	app.buildValueCache()

	expect := map[string][]string{
		"status":   {"closed", "open"},
		"priority": {"1", "2"},
		"type":     {"bug", "feature"},
		"labels":   {"alpha", "beta", "gamma"},
		"label":    {"alpha", "beta", "gamma"},
	}
	for field, want := range expect {
		got := app.valueCache[field]
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s cache mismatch: got %v want %v", field, got, want)
		}
	}
}

func TestSuggestionsForFieldAppliesFormatter(t *testing.T) {
	app := &App{
		valueCache: map[string][]string{
			"status": {"open"},
		},
		suggestionFormatter: map[string]func(string) string{
			"status": strings.ToUpper,
		},
	}
	entries := app.suggestionsForField("status")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Display != "OPEN" || entries[0].Value != "open" {
		t.Fatalf("unexpected entry: %+v", entries[0])
	}
}

func TestApplySuggestionReplacesInputValue(t *testing.T) {
	app := &App{
		textInput: textinput.New(),
	}
	app.textInput.SetValue("status:")
	app.overlay.pendingField = "status"
	app.overlay.pendingText = "status:"
	app.applySuggestion("open")
	if got := app.textInput.Value(); got != "status:open " {
		t.Fatalf("expected text input updated, got %q", got)
	}
}
