package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// extractShortError extracts a short, user-friendly error message.
func extractShortError(fullError string, maxLen int) string {
	msg := fullError

	// Look for "Error:" pattern and extract from there
	if idx := strings.Index(msg, "Error:"); idx >= 0 {
		msg = strings.TrimSpace(msg[idx+6:]) // Skip "Error:"
	} else if idx := strings.Index(msg, "error:"); idx >= 0 {
		msg = strings.TrimSpace(msg[idx+6:])
	}

	// Take only the first line/sentence
	if idx := strings.Index(msg, "\n"); idx >= 0 {
		msg = msg[:idx]
	}
	// Also truncate at period if it makes sense
	if idx := strings.Index(msg, ". "); idx >= 0 && idx < maxLen {
		msg = msg[:idx]
	}

	// Remove any "Run 'bd..." suggestions
	if idx := strings.Index(msg, " Run '"); idx >= 0 {
		msg = msg[:idx]
	}
	if idx := strings.Index(msg, " Run \""); idx >= 0 {
		msg = msg[:idx]
	}

	msg = strings.TrimSpace(msg)

	// Truncate if still too long
	if len(msg) > maxLen {
		msg = msg[:maxLen-3] + "..."
	}

	return msg
}

// overlayBottomRight positions the overlay at bottom-right of the background.
// containerWidth specifies the known width of the container for proper right-alignment.
// padding is parameterized for testability even though production always uses 1.
//
//nolint:unparam // padding varies in tests
func overlayBottomRight(background, overlay string, containerWidth, padding int) string {
	if overlay == "" {
		return background
	}

	bgLines := strings.Split(background, "\n")
	overlayLines := strings.Split(overlay, "\n")

	bgHeight := len(bgLines)
	overlayHeight := len(overlayLines)
	overlayWidth := lipgloss.Width(overlay)

	// Calculate position: bottom-right with padding
	startRow := bgHeight - overlayHeight - padding
	if startRow < 0 {
		startRow = 0
	}

	// Insert column: account for border (1 char) plus padding inside
	borderWidth := 1
	insertCol := containerWidth - overlayWidth - padding - borderWidth
	if insertCol < 0 {
		insertCol = 0
	}

	// For each overlay line, merge it with background
	overlayLineWidth := 0
	if len(overlayLines) > 0 {
		overlayLineWidth = lipgloss.Width(overlayLines[0])
	}

	for i, overlayLine := range overlayLines {
		bgRow := startRow + i
		if bgRow >= bgHeight {
			break
		}

		bgLine := bgLines[bgRow]
		bgLineWidth := lipgloss.Width(bgLine)

		// Pad background line to reach insert position
		for lipgloss.Width(bgLine) < insertCol {
			bgLine += " "
		}

		// Build: left part + overlay + right part (preserves border)
		leftPart := truncateVisualWidth(bgLine, insertCol)
		rightStart := insertCol + overlayLineWidth
		rightPart := ""
		if rightStart < bgLineWidth {
			// Extract the right portion after the overlay ends
			rightPart = ansi.Cut(bgLines[bgRow], rightStart, bgLineWidth)
		}
		bgLines[bgRow] = leftPart + overlayLine + rightPart
	}

	return strings.Join(bgLines, "\n")
}

// truncateVisualWidth truncates a string to the specified visual width,
// properly handling ANSI escape sequences.
func truncateVisualWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	// Use proper ANSI-aware truncation
	return ansi.Truncate(s, width, "")
}
