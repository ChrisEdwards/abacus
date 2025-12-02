package theme

import "github.com/charmbracelet/lipgloss"

// Aura color palette
// https://github.com/daltonmenezes/aura-theme
var aura = struct {
	DarkBg      string
	DarkBgPanel string
	DarkBorder  string
	DarkFgMuted string
	DarkFg      string
	Purple      string
	Pink        string
	Blue        string
	Red         string
	Orange      string
	Cyan        string
	Green       string
}{
	DarkBg:      "#0f0f0f",
	DarkBgPanel: "#15141b",
	DarkBorder:  "#2d2d2d",
	DarkFgMuted: "#6d6d6d",
	DarkFg:      "#edecee",
	Purple:      "#a277ff",
	Pink:        "#f694ff",
	Blue:        "#82e2ff",
	Red:         "#ff6767",
	Orange:      "#ffca85",
	Cyan:        "#61ffca",
	Green:       "#9dff65",
}

// AuraTheme implements Theme with the Aura color palette.
type AuraTheme struct{}

// Base colors

func (t AuraTheme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#7c4dff", Dark: aura.Purple}
}

func (t AuraTheme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#e040fb", Dark: aura.Pink}
}

func (t AuraTheme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#7c4dff", Dark: aura.Purple}
}

// Status colors

func (t AuraTheme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#d32f2f", Dark: aura.Red}
}

func (t AuraTheme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#ff9800", Dark: aura.Orange}
}

func (t AuraTheme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#00bfa5", Dark: aura.Cyan}
}

func (t AuraTheme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#7c4dff", Dark: aura.Purple}
}

// Text colors

func (t AuraTheme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#212121", Dark: aura.DarkFg}
}

func (t AuraTheme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#757575", Dark: aura.DarkFgMuted}
}

func (t AuraTheme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#000000", Dark: aura.DarkFg}
}

// Background colors

func (t AuraTheme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#ffffff", Dark: aura.DarkBg}
}

func (t AuraTheme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#f5f5f5", Dark: aura.DarkBgPanel}
}

func (t AuraTheme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#eeeeee", Dark: aura.DarkBgPanel}
}

// Border colors

func (t AuraTheme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#e0e0e0", Dark: aura.DarkBorder}
}

func (t AuraTheme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#757575", Dark: aura.DarkFgMuted}
}

func (t AuraTheme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#eeeeee", Dark: aura.DarkBorder}
}

func init() {
	RegisterTheme("aura", AuraTheme{})
}
