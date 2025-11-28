package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// footerHint defines a key hint for the footer bar.
// These are intentionally shorter than the KeyMap help text.
type footerHint struct {
	key  string // Short symbol: "↑↓", "←→", "/", etc.
	desc string // Short description: "Navigate", "Expand", etc.
}

// Global footer hints (always shown)
var globalFooterHints = []footerHint{
	{"/", "Search"},
	{"⏎", "Detail"},
	{"⇥", "Focus"},
	{"q", "Quit"},
	{"?", "Help"},
}

// Context-specific footer hints
var treeFooterHints = []footerHint{
	{"↑↓", "Navigate"},
	{"←→", "Expand"},
}

var detailsFooterHints = []footerHint{
	{"↑↓", "Scroll"},
}

// renderFooter renders the footer bar with pill-style key hints.
func (m *App) renderFooter() string {
	var hints []footerHint

	// Context-specific keys (shown first, leftmost)
	switch m.focus {
	case FocusTree:
		hints = append(hints, treeFooterHints...)
	case FocusDetails:
		hints = append(hints, detailsFooterHints...)
	}

	// Global keys
	hints = append(hints, globalFooterHints...)

	// Calculate available width
	repoText := "Repo: " + m.repoName
	repoRendered := styleFooterMuted.Render(repoText)
	repoWidth := lipgloss.Width(repoRendered)
	availableWidth := m.width - repoWidth - 4 // padding

	// Progressively remove hints if too wide
	hints = m.trimHintsToFit(hints, availableWidth)

	// Render hints as pills
	var parts []string
	for _, h := range hints {
		parts = append(parts, keyPill(h.key, h.desc))
	}

	left := strings.Join(parts, "  ")
	leftWidth := lipgloss.Width(left)

	// Calculate spacing for right-alignment
	spacing := m.width - leftWidth - repoWidth
	if spacing < 2 {
		spacing = 2
	}

	return left + strings.Repeat(" ", spacing) + repoRendered
}

// keyPill renders a single key hint as a pill with description.
func keyPill(key, desc string) string {
	return styleKeyPill.Render(" "+key+" ") + " " + styleKeyDesc.Render(desc)
}

// trimHintsToFit progressively removes hints to fit available width.
// Removes context-specific hints first, then global hints from end.
func (m *App) trimHintsToFit(hints []footerHint, availableWidth int) []footerHint {
	globalCount := len(globalFooterHints)

	for len(hints) > 0 {
		rendered := renderHintsWidth(hints)
		if rendered <= availableWidth {
			break
		}
		// Remove context-specific hints first, keep globals
		if len(hints) > globalCount {
			hints = hints[1:]
		} else {
			// Remove from end (least important globals)
			hints = hints[:len(hints)-1]
		}
	}
	return hints
}

// renderHintsWidth calculates the visual width of rendered hints.
func renderHintsWidth(hints []footerHint) int {
	var parts []string
	for _, h := range hints {
		parts = append(parts, keyPill(h.key, h.desc))
	}
	return lipgloss.Width(strings.Join(parts, "  "))
}
