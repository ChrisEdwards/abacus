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
	{"⏎", "Detail"},
	{"⇥", "Focus"},
	{"/", "Search"},
	{"v", "View"},
	{"n", "New"},
	{"s", "✎ Status"},
	{"L", "Labels"},
	{"m", "Comment"},
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

var statusOverlayFooterHints = []footerHint{
	{"o", "Open"},
	{"i", "In Progress"},
	{"c", "Close"},
	{"esc", "Cancel"},
}

var labelsOverlayFooterHints = []footerHint{
	{"⏎", "Save"},
	{"esc", "Cancel"},
}

var createOverlayFooterHints = []footerHint{
	{"Tab", "Next"},
	{"←→", "Select"},
	{"⏎", "Submit"},
	{"esc", "Cancel"},
}

// renderFooter renders the footer bar with pill-style key hints.
func (m *App) renderFooter() string {
	var hints []footerHint

	// Overlays get their own footer (no global hints)
	switch m.activeOverlay {
	case OverlayStatus:
		hints = statusOverlayFooterHints
	case OverlayLabels:
		hints = labelsOverlayFooterHints
	case OverlayCreate:
		hints = createOverlayFooterHints
	default:
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

	// Join with styled separators
	sp := baseStyle().Render("  ")
	left := strings.Join(parts, sp)
	leftWidth := lipgloss.Width(left)

	// Calculate spacing for right-alignment
	spacing := m.width - leftWidth - rightWidth
	if spacing < 2 {
		spacing = 2
	}

	spacer := baseStyle().Render(strings.Repeat(" ", spacing))
	return baseStyle().Width(m.width).Render(left + spacer + rightContent)
}

// renderRefreshStatus returns the current refresh status for the footer.
// Priority: error > refreshing > delta metrics (if changed) > empty
func (m *App) renderRefreshStatus() string {
	if m.lastError != "" {
		return styleErrorIndicator().Render("⚠ Error (!)")
	}
	if m.refreshInFlight {
		return styleFooterMuted().Render(m.spinner.View())
	}
	// Only show delta if something changed and within display duration
	if m.lastRefreshStats != "" &&
		m.lastRefreshStats != "+0 / Δ0 / -0" &&
		time.Since(m.lastRefreshTime) < refreshDisplayDuration {
		return styleFooterMuted().Render("Δ " + m.lastRefreshStats)
	}
	// Reserve space for spinner to prevent layout shifts when refresh starts
	return baseStyle().Render(" ")
}

// keyPill renders a single key hint as a pill with description.
func keyPill(key, desc string) string {
	return styleKeyPill().Render(" "+key+" ") + baseStyle().Render(" ") + styleKeyDesc().Render(desc)
}

// overlayFooterLine centers overlay footer hints within the given width.
// Width should match the overlay content width (before borders/padding).
func overlayFooterLine(hints []footerHint, width int) string {
	var parts []string
	for _, h := range hints {
		parts = append(parts, overlayKeyPill(h.key, h.desc))
	}
	line := strings.Join(parts, "  ")
	if width <= 0 {
		return line
	}
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Background(currentThemeWrapper().BackgroundSecondary()).
		Render(line)
}

// overlayKeyPill renders a pill for overlays with inverted backgrounds:
// key on dark background, description on overlay background to improve contrast.
func overlayKeyPill(key, desc string) string {
	keyStyle := lipgloss.NewStyle().
		Background(currentThemeWrapper().Background()).
		Foreground(currentThemeWrapper().Accent())
	keyStyle = applyBold(keyStyle, false)

	descStyle := lipgloss.NewStyle().
		Background(currentThemeWrapper().BackgroundSecondary()).
		Foreground(currentThemeWrapper().TextMuted())

	return keyStyle.Render(" "+key+" ") + descStyle.Render(" "+desc)
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
