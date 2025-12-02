package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// padLinesToWidth ensures every line in the content is padded to the provided width.
func padLinesToWidth(content string, width int) string {
	if width <= 0 || content == "" {
		return content
	}
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = padLineToWidth(line, width)
	}
	return strings.Join(lines, "\n")
}

// padLinesToMaxWidth pads each line so the content reaches its widest visual width.
func padLinesToMaxWidth(content string) string {
	lines := strings.Split(content, "\n")
	width := maxLineWidth(lines)
	if width <= 0 {
		return content
	}
	return padLinesToWidth(content, width)
}

// padLineToWidth pads a single line with the base background so it reaches the provided width.
func padLineToWidth(line string, width int) string {
	if width <= 0 {
		return line
	}
	lineWidth := lipgloss.Width(line)
	if lineWidth >= width {
		return line
	}
	padding := baseStyle().Render(strings.Repeat(" ", width-lineWidth))
	return line + padding
}

// blankLine returns a background-filled blank line with the given width.
func blankLine(width int) string {
	if width <= 0 {
		return ""
	}
	return baseStyle().Render(strings.Repeat(" ", width))
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
