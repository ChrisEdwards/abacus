package theme

import "github.com/charmbracelet/lipgloss"

// Material color palette
// https://material.io/design/color
var material = struct {
	DarkBg       string
	DarkBgAlt    string
	DarkBgPanel  string
	DarkFg       string
	DarkFgMuted  string
	DarkRed      string
	DarkOrange   string
	DarkYellow   string
	DarkGreen    string
	DarkCyan     string
	DarkBlue     string
	DarkPurple   string
	LightBg      string
	LightBgAlt   string
	LightBgPanel string
	LightFg      string
	LightFgMuted string
	LightRed     string
	LightOrange  string
	LightYellow  string
	LightGreen   string
	LightCyan    string
	LightBlue    string
	LightPurple  string
}{
	DarkBg:       "#263238",
	DarkBgAlt:    "#1e272c",
	DarkBgPanel:  "#37474f",
	DarkFg:       "#eeffff",
	DarkFgMuted:  "#546e7a",
	DarkRed:      "#f07178",
	DarkOrange:   "#ffcb6b",
	DarkYellow:   "#ffcb6b",
	DarkGreen:    "#c3e88d",
	DarkCyan:     "#89ddff",
	DarkBlue:     "#82aaff",
	DarkPurple:   "#c792ea",
	LightBg:      "#fafafa",
	LightBgAlt:   "#f5f5f5",
	LightBgPanel: "#e7e7e8",
	LightFg:      "#263238",
	LightFgMuted: "#90a4ae",
	LightRed:     "#e53935",
	LightOrange:  "#f4511e",
	LightYellow:  "#ffb300",
	LightGreen:   "#91b859",
	LightCyan:    "#39adb5",
	LightBlue:    "#6182b8",
	LightPurple:  "#7c4dff",
}

// MaterialTheme implements Theme with the Material color palette.
type MaterialTheme struct{}

// Base colors

func (t MaterialTheme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: material.LightBlue, Dark: material.DarkBlue}
}

func (t MaterialTheme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: material.LightPurple, Dark: material.DarkPurple}
}

func (t MaterialTheme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: material.LightCyan, Dark: material.DarkCyan}
}

// Status colors

func (t MaterialTheme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: material.LightRed, Dark: material.DarkRed}
}

func (t MaterialTheme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: material.LightYellow, Dark: material.DarkYellow}
}

func (t MaterialTheme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: material.LightGreen, Dark: material.DarkGreen}
}

func (t MaterialTheme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: material.LightOrange, Dark: material.DarkOrange}
}

// Text colors

func (t MaterialTheme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: material.LightFg, Dark: material.DarkFg}
}

func (t MaterialTheme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: material.LightFgMuted, Dark: material.DarkFgMuted}
}

func (t MaterialTheme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#000000", Dark: material.DarkFg}
}

// Background colors

func (t MaterialTheme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: material.LightBg, Dark: material.DarkBg}
}

func (t MaterialTheme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: material.LightBgAlt, Dark: material.DarkBgPanel}
}

func (t MaterialTheme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: material.LightBgPanel, Dark: material.DarkBgAlt}
}

// Border colors

func (t MaterialTheme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#e0e0e0", Dark: "#37474f"}
}

func (t MaterialTheme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: material.LightBlue, Dark: material.DarkBlue}
}

func (t MaterialTheme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#eeeeee", Dark: "#1e272c"}
}

func init() {
	RegisterTheme("material", MaterialTheme{})
}
