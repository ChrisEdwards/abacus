package theme

import "github.com/charmbracelet/lipgloss"

// Zenburn color palette
// https://github.com/jnurmine/Zenburn
var zenburn = struct {
	Background      string
	BackgroundAlt   string
	BackgroundPanel string
	Foreground      string
	ForegroundMuted string
	Red             string
	RedBright       string
	Green           string
	GreenBright     string
	Yellow          string
	YellowDim       string
	Blue            string
	BlueDim         string
	Magenta         string
	Cyan            string
	Orange          string
}{
	Background:      "#3f3f3f",
	BackgroundAlt:   "#4f4f4f",
	BackgroundPanel: "#5f5f5f",
	Foreground:      "#dcdccc",
	ForegroundMuted: "#9f9f9f",
	Red:             "#cc9393",
	RedBright:       "#dca3a3",
	Green:           "#7f9f7f",
	GreenBright:     "#8fb28f",
	Yellow:          "#f0dfaf",
	YellowDim:       "#e0cf9f",
	Blue:            "#8cd0d3",
	BlueDim:         "#7cb8bb",
	Magenta:         "#dc8cc3",
	Cyan:            "#93e0e3",
	Orange:          "#dfaf8f",
}

// ZenburnTheme implements Theme with the Zenburn color palette.
type ZenburnTheme struct{}

// Base colors

func (t ZenburnTheme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#5f7f8f", Dark: zenburn.Blue}
}

func (t ZenburnTheme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#8f5f8f", Dark: zenburn.Magenta}
}

func (t ZenburnTheme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#5f8f8f", Dark: zenburn.Cyan}
}

// Status colors

func (t ZenburnTheme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#8f5f5f", Dark: zenburn.Red}
}

func (t ZenburnTheme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#8f8f5f", Dark: zenburn.Yellow}
}

func (t ZenburnTheme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#5f8f5f", Dark: zenburn.Green}
}

func (t ZenburnTheme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#8f7f5f", Dark: zenburn.Orange}
}

// Text colors

func (t ZenburnTheme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#3f3f3f", Dark: zenburn.Foreground}
}

func (t ZenburnTheme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#6f6f6f", Dark: zenburn.ForegroundMuted}
}

func (t ZenburnTheme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#000000", Dark: zenburn.Foreground}
}

// Background colors

func (t ZenburnTheme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#ffffef", Dark: zenburn.Background}
}

func (t ZenburnTheme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#f5f5e5", Dark: zenburn.BackgroundAlt}
}

func (t ZenburnTheme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#ebebdb", Dark: zenburn.BackgroundPanel}
}

// Border colors

func (t ZenburnTheme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#d0d0c0", Dark: "#5f5f5f"}
}

func (t ZenburnTheme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#5f7f8f", Dark: zenburn.Blue}
}

func (t ZenburnTheme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#e0e0d0", Dark: "#4f4f4f"}
}

func init() {
	RegisterTheme("zenburn", ZenburnTheme{})
}
