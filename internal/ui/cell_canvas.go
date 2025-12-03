package ui

import (
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/cellbuf"
)

// Canvas is a lightweight helper around cellbuf.Screen that lets us compose
// lipgloss-rendered strings into a cell buffer before turning the frame back
// into a string for Bubble Tea.
type Canvas struct {
	screen *cellbuf.Screen
	writer *cellbuf.ScreenWriter
	width  int
	height int
}

func NewCanvas(width, height int) *Canvas {
	if width <= 0 {
		width = 1
	}
	if height <= 0 {
		height = 1
	}
	screen := cellbuf.NewScreen(io.Discard, width, height, &cellbuf.ScreenOptions{
		ShowCursor: false,
		AltScreen:  false,
	})
	return &Canvas{
		screen: screen,
		writer: cellbuf.NewScreenWriter(screen),
		width:  width,
		height: height,
	}
}

// Fill paints the entire canvas with the provided background color.
func (c *Canvas) Fill(bg lipgloss.TerminalColor) {
	if c == nil {
		return
	}
	fill := lipgloss.NewStyle().
		Background(bg).
		Width(c.width).
		Height(c.height).
		Render("")
	c.DrawStringAt(0, 0, fill)
}

// DrawStringAt writes the provided block starting at x,y. Newlines are
// normalized so each line begins at column 0 relative to x.
func (c *Canvas) DrawStringAt(x, y int, content string) {
	if content == "" || c == nil || c.writer == nil {
		return
	}
	normalized := normalizeForCellbuf(content)
	c.writer.PrintCropAt(x, y, normalized, "")
}

// centerOverlay renders the provided overlay centered within the canvas,
// respecting the top/bottom margins so headers/footers remain visible.
func (c *Canvas) centerOverlay(overlay string, topMargin, bottomMargin int) {
	lines := splitOverlayLines(overlay)
	if len(lines) == 0 || c == nil {
		return
	}

	overlayHeight := len(lines)
	overlayWidth := maxLineWidth(lines)
	if overlayWidth > c.width {
		overlayWidth = c.width
	}

	if topMargin < 0 {
		topMargin = 0
	}
	if bottomMargin < 0 {
		bottomMargin = 0
	}

	usableHeight := c.height - topMargin - bottomMargin
	if usableHeight < overlayHeight {
		usableHeight = overlayHeight
	}

	startY := topMargin
	if usableHeight > overlayHeight {
		startY = topMargin + (usableHeight-overlayHeight)/2
	}
	maxStartY := c.height - bottomMargin - overlayHeight
	if startY > maxStartY {
		startY = maxStartY
	}
	if startY < topMargin {
		startY = topMargin
	}
	if startY < 0 {
		startY = 0
	}

	startX := (c.width - overlayWidth) / 2
	if startX < 0 {
		startX = 0
	}

	c.drawBlockAt(startX, startY, lines)
}

// bottomRightOverlay positions the overlay anchored to the bottom-right corner
// with the provided padding inside the canvas.
func (c *Canvas) bottomRightOverlay(overlay string, padding int) {
	lines := splitOverlayLines(overlay)
	if len(lines) == 0 || c == nil {
		return
	}
	if padding < 0 {
		padding = 0
	}

	overlayHeight := len(lines)
	startY := c.height - overlayHeight - padding
	if startY < 0 {
		startY = 0
	}

	overlayWidth := maxLineWidth(lines)
	startX := c.width - overlayWidth - padding
	if startX < 0 {
		startX = 0
	}

	c.drawBlockAt(startX, startY, lines)
}

func (c *Canvas) drawBlockAt(x, y int, lines []string) {
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	for i, line := range lines {
		row := y + i
		if row >= c.height {
			break
		}
		if line == "" {
			continue
		}
		c.writer.PrintCropAt(x, row, line, "")
	}
}

// Render returns the composed frame as a newline-delimited string suitable for
// Bubble Tea consumption.
func (c *Canvas) Render() string {
	if c == nil || c.screen == nil {
		return ""
	}
	raw := cellbuf.Render(c.screen)
	_ = c.screen.Close()
	return strings.ReplaceAll(raw, "\r\n", "\n")
}

func normalizeForCellbuf(content string) string {
	if content == "" {
		return ""
	}
	content = strings.ReplaceAll(content, "\r\n", "\n")
	return strings.ReplaceAll(content, "\n", "\r\n")
}

func splitOverlayLines(content string) []string {
	if content == "" {
		return nil
	}
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	return strings.Split(normalized, "\n")
}
