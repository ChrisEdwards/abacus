package ui

import (
	"strings"
	"time"

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
	{"s", "Status"},
}

var detailsFooterHints = []footerHint{
	{"↑↓", "Scroll"},
}

var statusOverlayFooterHints = []footerHint{
	{"o", "Open"},
	{"i", "In Progress"},
	{"c", "Close"},
	{"esc", "Cancel"},
}

// renderFooter renders the footer bar with pill-style key hints.
func (m *App) renderFooter() string {
	var hints []footerHint

	// Status overlay gets its own footer (no global hints)
	if m.activeOverlay == OverlayStatus {
		hints = statusOverlayFooterHints
	} else {
		// Context-specific keys (shown first, leftmost)
		switch m.focus {
		case FocusTree:
			hints = append(hints, treeFooterHints...)
		case FocusDetails:
			hints = append(hints, detailsFooterHints...)
		}

		// Global keys
		hints = append(hints, globalFooterHints...)
	}

	// Calculate available width for hints
	rightContent := m.renderRefreshStatus()
	rightWidth := lipgloss.Width(rightContent)
	availableWidth := m.width - rightWidth - 4 // padding

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
	spacing := m.width - leftWidth - rightWidth
	if spacing < 2 {
		spacing = 2
	}

	return left + strings.Repeat(" ", spacing) + rightContent
}

// renderRefreshStatus returns the current refresh status for the footer.
// Priority: error > refreshing > delta metrics (if changed) > empty
func (m *App) renderRefreshStatus() string {
	if m.lastError != "" {
		return styleErrorIndicator.Render("⚠ Refresh error (e)")
	}
	if m.refreshInFlight {
		return styleFooterMuted.Render(m.spinner.View() + " Refreshing...")
	}
	// Only show delta if something changed and within display duration
	if m.lastRefreshStats != "" &&
		m.lastRefreshStats != "+0 / Δ0 / -0" &&
		time.Since(m.lastRefreshTime) < refreshDisplayDuration {
		return styleFooterMuted.Render("Δ " + m.lastRefreshStats)
	}
	return ""
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
