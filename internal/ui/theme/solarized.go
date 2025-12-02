package theme

import "github.com/charmbracelet/lipgloss"

// Solarized color palette
// https://ethanschoonover.com/solarized/
var solarized = struct {
	Base03  string
	Base02  string
	Base01  string
	Base00  string
	Base0   string
	Base1   string
	Base2   string
	Base3   string
	Yellow  string
	Orange  string
	Red     string
	Magenta string
	Violet  string
	Blue    string
	Cyan    string
	Green   string
}{
	Base03:  "#002b36",
	Base02:  "#073642",
	Base01:  "#586e75",
	Base00:  "#657b83",
	Base0:   "#839496",
	Base1:   "#93a1a1",
	Base2:   "#eee8d5",
	Base3:   "#fdf6e3",
	Yellow:  "#b58900",
	Orange:  "#cb4b16",
	Red:     "#dc322f",
	Magenta: "#d33682",
	Violet:  "#6c71c4",
	Blue:    "#268bd2",
	Cyan:    "#2aa198",
	Green:   "#859900",
}

// SolarizedTheme implements Theme with the Solarized color palette.
type SolarizedTheme struct{}

// Base colors

func (t SolarizedTheme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: solarized.Blue, Dark: solarized.Blue}
}

func (t SolarizedTheme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: solarized.Violet, Dark: solarized.Violet}
}

func (t SolarizedTheme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: solarized.Cyan, Dark: solarized.Cyan}
}

// Status colors

func (t SolarizedTheme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: solarized.Red, Dark: solarized.Red}
}

func (t SolarizedTheme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: solarized.Yellow, Dark: solarized.Yellow}
}

func (t SolarizedTheme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: solarized.Green, Dark: solarized.Green}
}

func (t SolarizedTheme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: solarized.Orange, Dark: solarized.Orange}
}

// Text colors

func (t SolarizedTheme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: solarized.Base00, Dark: solarized.Base0}
}

func (t SolarizedTheme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: solarized.Base1, Dark: solarized.Base01}
}

func (t SolarizedTheme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#000000", Dark: solarized.Base0}
}

// Background colors

func (t SolarizedTheme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: solarized.Base3, Dark: solarized.Base03}
}

func (t SolarizedTheme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: solarized.Base2, Dark: solarized.Base02}
}

func (t SolarizedTheme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#eee8d5", Dark: "#073642"}
}

// Border colors

func (t SolarizedTheme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: solarized.Base2, Dark: solarized.Base02}
}

func (t SolarizedTheme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: solarized.Base1, Dark: solarized.Base01}
}

func (t SolarizedTheme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#eee8d5", Dark: "#073642"}
}

func init() {
	RegisterTheme("solarized", SolarizedTheme{})
}
