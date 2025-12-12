package ui

import (
	"strings"

	"abacus/internal/ui/theme"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
)

// NewBaseTextarea returns a textarea configured for Abacus overlays.
// It removes the default prompt/line numbers so the full interior is usable input space.
func NewBaseTextarea(width, height int) textarea.Model {
	ta := textarea.New()
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.SetWidth(width)
	ta.SetHeight(height)
	return ta
}

// TextareaContentWidth calculates the inner width available for textarea content
// given a container width and horizontal padding.
func TextareaContentWidth(containerWidth, padding int) int {
	inner := containerWidth - (padding * 2)
	if inner < 1 {
		return 1
	}
	return inner
}

// PadTextareaView adds left/right padding to a textarea view using the overlay background.
// This ensures the black input area starts at the text and fills the interior without bleeding into borders.
func PadTextareaView(view string, padding int) string {
	if padding <= 0 {
		return view
	}

	pad := lipgloss.NewStyle().
		Background(theme.Current().BackgroundSecondary()).
		Render(strings.Repeat(" ", padding))

	lines := strings.Split(view, "\n")
	for i, line := range lines {
		// Preserve trailing newline from textarea (last element empty)
		if line == "" && i == len(lines)-1 {
			continue
		}
		lines[i] = pad + line + pad
	}
	return strings.Join(lines, "\n")
}
