package theme

import "github.com/charmbracelet/lipgloss"

// Night Owl color palette
// https://github.com/sdras/night-owl-vscode-theme
var nightowl = struct {
	Background string
	Foreground string
	Blue       string
	Cyan       string
	Green      string
	Yellow     string
	Orange     string
	Red        string
	Pink       string
	Purple     string
	Muted      string
	Gray       string
	Panel      string
}{
	Background: "#011627",
	Foreground: "#d6deeb",
	Blue:       "#82AAFF",
	Cyan:       "#7fdbca",
	Green:      "#c5e478",
	Yellow:     "#ecc48d",
	Orange:     "#F78C6C",
	Red:        "#EF5350",
	Pink:       "#ff5874",
	Purple:     "#c792ea",
	Muted:      "#5f7e97",
	Gray:       "#637777",
	Panel:      "#0b253a",
}

// NightOwlTheme implements Theme with the Night Owl color palette.
type NightOwlTheme struct{}

// Base colors

func (t NightOwlTheme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nightowl.Blue, Dark: nightowl.Blue}
}

func (t NightOwlTheme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nightowl.Cyan, Dark: nightowl.Cyan}
}

func (t NightOwlTheme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nightowl.Purple, Dark: nightowl.Purple}
}

// Status colors

func (t NightOwlTheme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nightowl.Red, Dark: nightowl.Red}
}

func (t NightOwlTheme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nightowl.Yellow, Dark: nightowl.Yellow}
}

func (t NightOwlTheme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nightowl.Green, Dark: nightowl.Green}
}

func (t NightOwlTheme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nightowl.Blue, Dark: nightowl.Blue}
}

// Text colors

func (t NightOwlTheme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nightowl.Foreground, Dark: nightowl.Foreground}
}

func (t NightOwlTheme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nightowl.Muted, Dark: nightowl.Muted}
}

func (t NightOwlTheme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#ffffff", Dark: nightowl.Foreground}
}

// Background colors

func (t NightOwlTheme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nightowl.Background, Dark: nightowl.Background}
}

func (t NightOwlTheme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nightowl.Panel, Dark: nightowl.Panel}
}

func (t NightOwlTheme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nightowl.Panel, Dark: nightowl.Panel}
}

// Border colors

func (t NightOwlTheme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nightowl.Muted, Dark: nightowl.Muted}
}

func (t NightOwlTheme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nightowl.Blue, Dark: nightowl.Blue}
}

func (t NightOwlTheme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: nightowl.Muted, Dark: nightowl.Muted}
}

func init() {
	RegisterTheme("nightowl", NightOwlTheme{})
}
