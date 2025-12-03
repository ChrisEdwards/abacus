package theme

import "github.com/charmbracelet/lipgloss"

// Vesper color palette
// https://github.com/raunofreiberg/vesper
var vesper = struct {
	Background string
	Foreground string
	Comment    string
	Keyword    string
	Function   string
	String     string
	Number     string
	Error      string
	Warning    string
	Success    string
	Muted      string
}{
	Background: "#101010",
	Foreground: "#FFF",
	Comment:    "#8b8b8b",
	Keyword:    "#A0A0A0",
	Function:   "#FFC799",
	String:     "#99FFE4",
	Number:     "#FFC799",
	Error:      "#FF8080",
	Warning:    "#FFC799",
	Success:    "#99FFE4",
	Muted:      "#A0A0A0",
}

// VesperTheme implements Theme with the Vesper color palette.
type VesperTheme struct{}

// Base colors

func (t VesperTheme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: vesper.Function, Dark: vesper.Function}
}

func (t VesperTheme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: vesper.String, Dark: vesper.String}
}

func (t VesperTheme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: vesper.Function, Dark: vesper.Function}
}

// Status colors

func (t VesperTheme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: vesper.Error, Dark: vesper.Error}
}

func (t VesperTheme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: vesper.Warning, Dark: vesper.Warning}
}

func (t VesperTheme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: vesper.Success, Dark: vesper.Success}
}

func (t VesperTheme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: vesper.Function, Dark: vesper.Function}
}

// Text colors

func (t VesperTheme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: vesper.Background, Dark: vesper.Foreground}
}

func (t VesperTheme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: vesper.Muted, Dark: vesper.Muted}
}

func (t VesperTheme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#000000", Dark: vesper.Foreground}
}

// Background colors

func (t VesperTheme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#FFF", Dark: vesper.Background}
}

func (t VesperTheme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#F0F0F0", Dark: "#1a1a1a"}
}

func (t VesperTheme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#E0E0E0", Dark: "#0a0a0a"}
}

// Border colors

func (t VesperTheme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#D0D0D0", Dark: "#282828"}
}

func (t VesperTheme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: vesper.Function, Dark: vesper.Function}
}

func (t VesperTheme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#E8E8E8", Dark: "#1C1C1C"}
}

func init() {
	RegisterTheme("vesper", VesperTheme{})
}
