package theme

import "github.com/charmbracelet/lipgloss"

// Ayu color palette
// https://github.com/ayu-theme/ayu-colors
var ayu = struct {
	DarkBg      string
	DarkBgAlt   string
	DarkPanel   string
	DarkFg      string
	DarkFgMuted string
	DarkGutter  string
	DarkEntity  string
	DarkAccent  string
	DarkError   string
	DarkAdded   string
	DarkSpecial string
	DarkTag     string
}{
	DarkBg:      "#0B0E14",
	DarkBgAlt:   "#0D1017",
	DarkPanel:   "#0F131A",
	DarkFg:      "#BFBDB6",
	DarkFgMuted: "#565B66",
	DarkGutter:  "#6C7380",
	DarkEntity:  "#59C2FF",
	DarkAccent:  "#E6B450",
	DarkError:   "#D95757",
	DarkAdded:   "#7FD962",
	DarkSpecial: "#E6B673",
	DarkTag:     "#39BAE6",
}

// AyuTheme implements Theme with the Ayu color palette.
type AyuTheme struct{}

// Base colors

func (t AyuTheme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#0d6efd", Dark: ayu.DarkEntity}
}

func (t AyuTheme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#6f42c1", Dark: "#D2A6FF"}
}

func (t AyuTheme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#fd7e14", Dark: ayu.DarkAccent}
}

// Status colors

func (t AyuTheme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#dc3545", Dark: ayu.DarkError}
}

func (t AyuTheme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#fd7e14", Dark: ayu.DarkSpecial}
}

func (t AyuTheme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#198754", Dark: ayu.DarkAdded}
}

func (t AyuTheme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#0dcaf0", Dark: ayu.DarkTag}
}

// Text colors

func (t AyuTheme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#212529", Dark: ayu.DarkFg}
}

func (t AyuTheme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#6c757d", Dark: ayu.DarkFgMuted}
}

func (t AyuTheme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#000000", Dark: ayu.DarkFg}
}

// Background colors

func (t AyuTheme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#ffffff", Dark: ayu.DarkBg}
}

func (t AyuTheme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#f8f9fa", Dark: "#1a1f28"}
}

func (t AyuTheme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#e9ecef", Dark: ayu.DarkBgAlt}
}

// Border colors

func (t AyuTheme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#dee2e6", Dark: ayu.DarkGutter}
}

func (t AyuTheme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#adb5bd", Dark: ayu.DarkGutter}
}

func (t AyuTheme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#f8f9fa", Dark: "#11151C"}
}

func init() {
	RegisterTheme("ayu", AyuTheme{})
}
