package theme

import "github.com/charmbracelet/lipgloss"

// Rosé Pine color palette
// https://rosepinetheme.com/
var rosepine = struct {
	Base         string
	Surface      string
	Overlay      string
	Muted        string
	Subtle       string
	Text         string
	Love         string
	Gold         string
	Rose         string
	Pine         string
	Foam         string
	Iris         string
	HighlightLow string
	HighlightMed string
	DawnBase     string
	DawnSurface  string
	DawnOverlay  string
	DawnMuted    string
	DawnText     string
}{
	Base:         "#191724",
	Surface:      "#1f1d2e",
	Overlay:      "#26233a",
	Muted:        "#6e6a86",
	Subtle:       "#908caa",
	Text:         "#e0def4",
	Love:         "#eb6f92",
	Gold:         "#f6c177",
	Rose:         "#ebbcba",
	Pine:         "#31748f",
	Foam:         "#9ccfd8",
	Iris:         "#c4a7e7",
	HighlightLow: "#21202e",
	HighlightMed: "#403d52",
	DawnBase:     "#faf4ed",
	DawnSurface:  "#fffaf3",
	DawnOverlay:  "#f2e9e1",
	DawnMuted:    "#9893a5",
	DawnText:     "#575279",
}

// RosepineTheme implements Theme with the Rosé Pine color palette.
type RosepineTheme struct{}

// Base colors

func (t RosepineTheme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: rosepine.Pine, Dark: rosepine.Foam}
}

func (t RosepineTheme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#907aa9", Dark: rosepine.Iris}
}

func (t RosepineTheme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#d7827e", Dark: rosepine.Rose}
}

// Status colors

func (t RosepineTheme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#b4637a", Dark: rosepine.Love}
}

func (t RosepineTheme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#ea9d34", Dark: rosepine.Gold}
}

func (t RosepineTheme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#286983", Dark: rosepine.Pine}
}

func (t RosepineTheme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#56949f", Dark: rosepine.Foam}
}

// Text colors

func (t RosepineTheme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: rosepine.DawnText, Dark: rosepine.Text}
}

func (t RosepineTheme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: rosepine.DawnMuted, Dark: rosepine.Muted}
}

func (t RosepineTheme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#000000", Dark: rosepine.Text}
}

// Background colors

func (t RosepineTheme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: rosepine.DawnBase, Dark: rosepine.Base}
}

func (t RosepineTheme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: rosepine.DawnSurface, Dark: rosepine.Overlay}
}

func (t RosepineTheme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: rosepine.DawnOverlay, Dark: rosepine.Overlay}
}

// Border colors

func (t RosepineTheme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#dfdad9", Dark: rosepine.HighlightMed}
}

func (t RosepineTheme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: rosepine.Pine, Dark: rosepine.Foam}
}

func (t RosepineTheme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#f4ede8", Dark: rosepine.HighlightLow}
}

func init() {
	RegisterTheme("rosepine", RosepineTheme{})
}
