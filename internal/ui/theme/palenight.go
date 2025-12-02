package theme

import "github.com/charmbracelet/lipgloss"

// Palenight color palette
// https://github.com/whizkydee/vscode-palenight-theme
var palenight = struct {
	Background      string
	BackgroundAlt   string
	BackgroundPanel string
	Foreground      string
	Comment         string
	Red             string
	Orange          string
	Yellow          string
	Green           string
	Cyan            string
	Blue            string
	Purple          string
}{
	Background:      "#292d3e",
	BackgroundAlt:   "#1e2132",
	BackgroundPanel: "#32364a",
	Foreground:      "#a6accd",
	Comment:         "#676e95",
	Red:             "#f07178",
	Orange:          "#f78c6c",
	Yellow:          "#ffcb6b",
	Green:           "#c3e88d",
	Cyan:            "#89ddff",
	Blue:            "#82aaff",
	Purple:          "#c792ea",
}

// PalenightTheme implements Theme with the Palenight color palette.
type PalenightTheme struct{}

// Base colors

func (t PalenightTheme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#4976eb", Dark: palenight.Blue}
}

func (t PalenightTheme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#a854f2", Dark: palenight.Purple}
}

func (t PalenightTheme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#00acc1", Dark: palenight.Cyan}
}

// Status colors

func (t PalenightTheme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#e53935", Dark: palenight.Red}
}

func (t PalenightTheme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#ffb300", Dark: palenight.Yellow}
}

func (t PalenightTheme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#91b859", Dark: palenight.Green}
}

func (t PalenightTheme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#f4511e", Dark: palenight.Orange}
}

// Text colors

func (t PalenightTheme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#292d3e", Dark: palenight.Foreground}
}

func (t PalenightTheme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#8796b0", Dark: palenight.Comment}
}

func (t PalenightTheme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#000000", Dark: "#bfc7d5"}
}

// Background colors

func (t PalenightTheme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#fafafa", Dark: palenight.Background}
}

func (t PalenightTheme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#f5f5f5", Dark: palenight.BackgroundAlt}
}

func (t PalenightTheme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#e7e7e8", Dark: palenight.BackgroundPanel}
}

// Border colors

func (t PalenightTheme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#e0e0e0", Dark: palenight.BackgroundPanel}
}

func (t PalenightTheme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#4976eb", Dark: palenight.Blue}
}

func (t PalenightTheme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#eeeeee", Dark: palenight.BackgroundAlt}
}

func init() {
	RegisterTheme("palenight", PalenightTheme{})
}
