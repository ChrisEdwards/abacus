package theme

import "github.com/charmbracelet/lipgloss"

// Nord color palette
// https://www.nordtheme.com/docs/colors-and-palettes
var nord = struct {
	Nord0  string // Polar Night
	Nord1  string
	Nord2  string
	Nord3  string
	Nord4  string // Snow Storm
	Nord5  string
	Nord6  string
	Nord7  string // Frost
	Nord8  string
	Nord9  string
	Nord10 string
	Nord11 string // Aurora
	Nord12 string
	Nord13 string
	Nord14 string
	Nord15 string
}{
	Nord0:  "#2E3440",
	Nord1:  "#3B4252",
	Nord2:  "#434C5E",
	Nord3:  "#4C566A",
	Nord4:  "#D8DEE9",
	Nord5:  "#E5E9F0",
	Nord6:  "#ECEFF4",
	Nord7:  "#8FBCBB",
	Nord8:  "#88C0D0",
	Nord9:  "#81A1C1",
	Nord10: "#5E81AC",
	Nord11: "#BF616A",
	Nord12: "#D08770",
	Nord13: "#EBCB8B",
	Nord14: "#A3BE8C",
	Nord15: "#B48EAD",
}

// NordTheme implements Theme with the Nord color palette.
type NordTheme struct{}

// Base colors

func (t NordTheme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nord.Nord10, Dark: nord.Nord8}
}

func (t NordTheme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nord.Nord9, Dark: nord.Nord9}
}

func (t NordTheme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nord.Nord7, Dark: nord.Nord7}
}

// Status colors

func (t NordTheme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nord.Nord11, Dark: nord.Nord11}
}

func (t NordTheme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nord.Nord12, Dark: nord.Nord12}
}

func (t NordTheme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nord.Nord14, Dark: nord.Nord14}
}

func (t NordTheme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nord.Nord10, Dark: nord.Nord8}
}

// Text colors

func (t NordTheme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nord.Nord0, Dark: nord.Nord6}
}

func (t NordTheme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nord.Nord1, Dark: "#8B95A7"}
}

func (t NordTheme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#000000", Dark: nord.Nord6}
}

// Background colors

func (t NordTheme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nord.Nord6, Dark: nord.Nord0}
}

func (t NordTheme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nord.Nord5, Dark: nord.Nord1}
}

func (t NordTheme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nord.Nord4, Dark: nord.Nord2}
}

// Border colors

func (t NordTheme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nord.Nord3, Dark: nord.Nord2}
}

func (t NordTheme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nord.Nord2, Dark: nord.Nord3}
}

func (t NordTheme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nord.Nord3, Dark: nord.Nord2}
}

func init() {
	RegisterTheme("nord", NordTheme{})
}
