package ui

import (
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
)

// Canvas is a lightweight helper around cellbuf.Screen that lets us compose
// lipgloss-rendered strings into a cell buffer before turning the frame back
// into a string for Bubble Tea.
type Canvas struct {
	screen  *cellbuf.Screen
	writer  *cellbuf.ScreenWriter
	width   int
	height  int
	offsetX int
	offsetY int
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
	c.ensureBackground(bg)
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

// SetOffset stores an overlay offset for later composition.
func (c *Canvas) SetOffset(x, y int) {
	if c == nil {
		return
	}
	c.offsetX = x
	c.offsetY = y
}

// Offset returns the stored overlay offset.
func (c *Canvas) Offset() (int, int) {
	if c == nil {
		return 0, 0
	}
	return c.offsetX, c.offsetY
}

// Overlay draws the provided canvas on top of the receiver using the origin.
func (c *Canvas) Overlay(top *Canvas) {
	c.OverlayAt(0, 0, top)
}

// OverlayAt draws the source canvas on top of the receiver starting at (x, y).
// Empty cells (nil or zero-width) in the source canvas are treated as
// transparent and do not overwrite the destination.
func (c *Canvas) OverlayAt(xOff, yOff int, top *Canvas) {
	if c == nil || top == nil || c.screen == nil || top.screen == nil {
		return
	}

	for y := 0; y < top.height; y++ {
		destY := yOff + y
		if destY < 0 || destY >= c.height {
			continue
		}
		for x := 0; x < top.width; x++ {
			destX := xOff + x
			if destX < 0 || destX >= c.width {
				continue
			}
			cell := top.screen.Cell(x, y)
			if cell == nil || cell.Empty() {
				continue
			}
			c.screen.SetCell(destX, destY, cell.Clone())
		}
	}
}

// OverlayCanvas overlays the provided canvas using its offset.
func (c *Canvas) OverlayCanvas(top *Canvas) {
	if c == nil || top == nil || c.screen == nil || top.screen == nil {
		return
	}
	c.OverlayAt(top.offsetX, top.offsetY, top)
}

// ApplyDimmer marks every non-empty cell as faint, dimming the canvas without
// mutating blank regions.
func (c *Canvas) ApplyDimmer() {
	if c == nil || c.screen == nil {
		return
	}
	for y := 0; y < c.height; y++ {
		for x := 0; x < c.width; x++ {
			cell := c.screen.Cell(x, y)
			if cell == nil || cell.Empty() {
				continue
			}
			cell.Style.Attrs |= cellbuf.FaintAttr
			c.screen.SetCell(x, y, cell)
		}
	}
}

// Width returns the canvas width.
func (c *Canvas) Width() int {
	if c == nil {
		return 0
	}
	return c.width
}

// Height returns the canvas height.
func (c *Canvas) Height() int {
	if c == nil {
		return 0
	}
	return c.height
}

// Cell returns a copy of the cell at the given coordinates, or nil if out of
// bounds. Primarily used in tests to assert background coverage.
func (c *Canvas) Cell(x, y int) *cellbuf.Cell {
	if c == nil || c.screen == nil {
		return nil
	}
	if x < 0 || x >= c.width || y < 0 || y >= c.height {
		return nil
	}
	cell := c.screen.Cell(x, y)
	if cell == nil {
		return nil
	}
	return cell.Clone()
}

func (c *Canvas) ensureBackground(color lipgloss.TerminalColor) {
	if c == nil || c.screen == nil || color == nil {
		return
	}
	bg, ok := lipglossColorToANSI(color)
	if !ok {
		return
	}
	for y := 0; y < c.height; y++ {
		for x := 0; x < c.width; x++ {
			cell := c.screen.Cell(x, y)
			if cell == nil {
				cell = &cellbuf.Cell{}
				cell.Blank()
			}
			if cell.Style.Bg == nil {
				cell.Style.Bg = bg
			}
			c.screen.SetCell(x, y, cell)
		}
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

func lipglossColorToANSI(color lipgloss.TerminalColor) (ansi.Color, bool) {
	if color == nil {
		return nil, false
	}
	r, g, b, _ := color.RGBA()
	return ansi.RGBColor{
		R: toUint8(r),
		G: toUint8(g),
		B: toUint8(b),
	}, true
}

func toUint8(v uint32) uint8 {
	const maxUint8 = 1<<8 - 1
	converted := v / 0x101 // Map 0-65535 to 0-255
	if converted > maxUint8 {
		return maxUint8
	}
	return uint8(converted)
}
