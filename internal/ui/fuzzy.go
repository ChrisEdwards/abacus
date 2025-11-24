package ui

import (
	"strings"

	"github.com/sahilm/fuzzy"
)

func rankSuggestions(entries []suggestionEntry, query string) []suggestionEntry {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" || len(entries) == 0 {
		return entries
	}
	targets := make([]string, len(entries))
	for i, entry := range entries {
		targets[i] = strings.ToLower(entry.Value)
	}
	matches := fuzzy.Find(query, targets)
	if len(matches) == 0 {
		return entries
	}
	used := make([]bool, len(entries))
	ranked := make([]suggestionEntry, 0, len(entries))
	for _, match := range matches {
		if match.Index >= 0 && match.Index < len(entries) {
			ranked = append(ranked, entries[match.Index])
			used[match.Index] = true
		}
	}
	for i, entry := range entries {
		if !used[i] {
			ranked = append(ranked, entry)
		}
	}
	return ranked
}
