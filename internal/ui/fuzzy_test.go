package ui

import "testing"

func TestRankSuggestionsUsesFuzzyMatch(t *testing.T) {
	entries := []suggestionEntry{
		{Display: "Open", Value: "open"},
		{Display: "Blocked", Value: "blocked"},
		{Display: "Closed", Value: "closed"},
	}
	ranked := rankSuggestions(entries, "blkd")
	if len(ranked) != len(entries) {
		t.Fatalf("expected %d entries, got %d", len(entries), len(ranked))
	}
	if ranked[0].Value != "blocked" {
		t.Fatalf("expected 'blocked' ranked first, got %q", ranked[0].Value)
	}
}
