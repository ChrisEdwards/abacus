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

// overlayCenterOnContent overlays the provided block in the center of the base surface.
// topMargin and bottomMargin reserve rows (e.g., header/footer) where the overlay should not render.
func overlayCenterOnContent(base, overlay string, width, height, topMargin, bottomMargin int) string {
	if overlay == "" || width <= 0 || height <= 0 {
		return base
	}

	lines := splitLinesWithPadding(base, height, width)
	overlayLines := strings.Split(overlay, "\n")
	overlayHeight := len(overlayLines)
	overlayWidth := maxLineWidth(overlayLines)
	if overlayWidth > width {
		overlayWidth = width
	}

	if topMargin < 0 {
		topMargin = 0
	}
	if bottomMargin < 0 {
		bottomMargin = 0
	}

	usableHeight := height - topMargin - bottomMargin
	if usableHeight < overlayHeight {
		usableHeight = overlayHeight
	}

	startRow := topMargin + (usableHeight-overlayHeight)/2
	minRow := topMargin
	maxRow := height - bottomMargin - overlayHeight
	if maxRow < minRow {
		maxRow = minRow
	}
	if startRow < minRow {
		startRow = minRow
	}
	if startRow > maxRow {
		startRow = maxRow
	}

	startCol := (width - overlayWidth) / 2
	if startCol < 0 {
		startCol = 0
	}

	for i, line := range overlayLines {
		row := startRow + i
		if row < 0 || row >= len(lines) {
			continue
		}
		lines[row] = overlayLineAt(lines[row], line, startCol, width)
	}

	return strings.Join(lines, "\n")
}

func splitLinesWithPadding(content string, height, width int) []string {
	lines := strings.Split(content, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for len(lines) < height {
		lines = append(lines, blankLine(width))
	}
	for i := range lines {
		lines[i] = padLineToWidth(lines[i], width)
	}
	return lines
}

func overlayLineAt(baseLine, overlayLine string, startCol, width int) string {
	if overlayLine == "" || width <= 0 {
		return baseLine
	}

	if startCol < 0 {
		startCol = 0
	}
	if startCol >= width {
		return baseLine
	}

	availableWidth := width - startCol
	if availableWidth <= 0 {
		return baseLine
	}

	lineWidth := lipgloss.Width(overlayLine)
	if lineWidth == 0 {
		return baseLine
	}
	if lineWidth > availableWidth {
		overlayLine = ansi.Truncate(overlayLine, availableWidth, "")
		lineWidth = lipgloss.Width(overlayLine)
	}

	baseLine = padLineToWidth(baseLine, width)
	left := ansi.Cut(baseLine, 0, startCol)
	left = padLineToWidth(left, startCol)

	endCol := startCol + lineWidth
	if endCol > width {
		endCol = width
	}

	right := ""
	if endCol < width {
		right = ansi.Cut(baseLine, endCol, width)
		right = padLineToWidth(right, width-endCol)
	}

	return left + overlayLine + right
}

func maxLineWidth(lines []string) int {
	max := 0
	for _, line := range lines {
		if w := lipgloss.Width(line); w > max {
			max = w
		}
	}
	return max
}
