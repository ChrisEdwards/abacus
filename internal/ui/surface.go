package ui

import (
	"github.com/charmbracelet/lipgloss"

	"abacus/internal/ui/theme"
)

// Surface wraps a Canvas plus theme-aware styles for drawing UI regions with the
// correct background baked into every glyph.
type Surface struct {
	Canvas *Canvas
	Styles SurfaceStyles
}

// SurfaceStyles mirrors the design doc, ensuring each style already includes
// the target background so callers don't need to chain baseStyle helpers.
type SurfaceStyles struct {
	Text          lipgloss.Style
	TextMuted     lipgloss.Style
	Accent        lipgloss.Style
	Error         lipgloss.Style
	Success       lipgloss.Style
	BorderNeutral lipgloss.Style
	BorderError   lipgloss.Style
}

// NewPrimarySurface returns a Surface that uses the primary application
// background from the active theme.
func NewPrimarySurface(width, height int) Surface {
	return newSurface(width, height, theme.Current().Background())
}

// NewSecondarySurface returns a Surface backed by the secondary background
// (used by overlays, dialogs, etc.).
func NewSecondarySurface(width, height int) Surface {
	return newSurface(width, height, theme.Current().BackgroundSecondary())
}

func newSurface(width, height int, bg lipgloss.TerminalColor) Surface {
	canvas := NewCanvas(width, height)
	canvas.Fill(bg)
	styles := SurfaceStyles{
		Text:          lipgloss.NewStyle().Background(bg).Foreground(theme.Current().Text()),
		TextMuted:     lipgloss.NewStyle().Background(bg).Foreground(theme.Current().TextMuted()),
		Accent:        lipgloss.NewStyle().Background(bg).Foreground(theme.Current().Accent()).Bold(true),
		Error:         lipgloss.NewStyle().Background(bg).Foreground(theme.Current().Error()).Bold(true),
		Success:       lipgloss.NewStyle().Background(bg).Foreground(theme.Current().Success()).Bold(true),
		BorderNeutral: lipgloss.NewStyle().Background(bg).Foreground(theme.Current().BorderNormal()),
		BorderError:   lipgloss.NewStyle().Background(bg).Foreground(theme.Current().Error()),
	}
	return Surface{
		Canvas: canvas,
		Styles: styles,
	}
}

// Draw writes the provided block starting at x,y.
func (s Surface) Draw(x, y int, block string) {
	if s.Canvas == nil {
		return
	}
	s.Canvas.DrawStringAt(x, y, block)
}

// Render flushes the surface to a string (ANSI frame).
func (s Surface) Render() string {
	if s.Canvas == nil {
		return ""
	}
	return s.Canvas.Render()
}
