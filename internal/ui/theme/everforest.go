package theme

import "github.com/charmbracelet/lipgloss"

// Everforest color palette
// https://github.com/sainnhe/everforest
var everforest = struct {
	DarkBg      string
	DarkBgPanel string
	DarkBgAlt   string
	DarkFg      string
	DarkFgMuted string
	DarkBorder  string
	DarkRed     string
	DarkOrange  string
	DarkGreen   string
	DarkCyan    string
	DarkYellow  string
	LightBg     string
	LightBgAlt  string
	LightFg     string
	LightFgMuted string
	LightBorder string
	LightRed    string
	LightOrange string
	LightGreen  string
	LightCyan   string
	LightYellow string
}{
	DarkBg:       "#2d353b",
	DarkBgPanel:  "#333c43",
	DarkBgAlt:    "#343f44",
	DarkFg:       "#d3c6aa",
	DarkFgMuted:  "#7a8478",
	DarkBorder:   "#859289",
	DarkRed:      "#e67e80",
	DarkOrange:   "#e69875",
	DarkGreen:    "#a7c080",
	DarkCyan:     "#83c092",
	DarkYellow:   "#dbbc7f",
	LightBg:      "#fdf6e3",
	LightBgAlt:   "#efebd4",
	LightFg:      "#5c6a72",
	LightFgMuted: "#a6b0a0",
	LightBorder:  "#939f91",
	LightRed:     "#f85552",
	LightOrange:  "#f57d26",
	LightGreen:   "#8da101",
	LightCyan:    "#35a77c",
	LightYellow:  "#dfa000",
}

// EverforestTheme implements Theme with the Everforest color palette.
type EverforestTheme struct{}

// Base colors

func (t EverforestTheme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: everforest.LightGreen, Dark: everforest.DarkGreen}
}

func (t EverforestTheme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#3a94c5", Dark: "#7fbbb3"}
}

func (t EverforestTheme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#df69ba", Dark: "#d699b6"}
}

// Status colors

func (t EverforestTheme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: everforest.LightRed, Dark: everforest.DarkRed}
}

func (t EverforestTheme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: everforest.LightOrange, Dark: everforest.DarkOrange}
}

func (t EverforestTheme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: everforest.LightGreen, Dark: everforest.DarkGreen}
}

func (t EverforestTheme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: everforest.LightCyan, Dark: everforest.DarkCyan}
}

// Text colors

func (t EverforestTheme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: everforest.LightFg, Dark: everforest.DarkFg}
}

func (t EverforestTheme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: everforest.LightFgMuted, Dark: everforest.DarkFgMuted}
}

func (t EverforestTheme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#000000", Dark: everforest.DarkFg}
}

// Background colors

func (t EverforestTheme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: everforest.LightBg, Dark: everforest.DarkBg}
}

func (t EverforestTheme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: everforest.LightBgAlt, Dark: everforest.DarkBgPanel}
}

func (t EverforestTheme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#f4f0d9", Dark: everforest.DarkBgAlt}
}

// Border colors

func (t EverforestTheme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: everforest.LightBorder, Dark: everforest.DarkBorder}
}

func (t EverforestTheme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#829181", Dark: "#9da9a0"}
}

func (t EverforestTheme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: everforest.LightFgMuted, Dark: everforest.DarkFgMuted}
}

func init() {
	RegisterTheme("everforest", EverforestTheme{})
}
