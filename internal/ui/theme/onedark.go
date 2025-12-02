package theme

import "github.com/charmbracelet/lipgloss"

// OneDarkTheme implements the One Dark color scheme.
type OneDarkTheme struct{}

func (t OneDarkTheme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#61afef", Light: "#4078f2"}
}

func (t OneDarkTheme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#c678dd", Light: "#a626a4"}
}

func (t OneDarkTheme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#e5c07b", Light: "#c18401"}
}

func (t OneDarkTheme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#e06c75", Light: "#e45649"}
}

func (t OneDarkTheme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#d19a66", Light: "#da8548"}
}

func (t OneDarkTheme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#98c379", Light: "#50a14f"}
}

func (t OneDarkTheme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#56b6c2", Light: "#0184bc"}
}

func (t OneDarkTheme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#abb2bf", Light: "#383a42"}
}

func (t OneDarkTheme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#5c6370", Light: "#a0a1a7"}
}

func (t OneDarkTheme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#e5c07b", Light: "#c18401"}
}

func (t OneDarkTheme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#282c34", Light: "#fafafa"}
}

func (t OneDarkTheme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#3e4451", Light: "#e5e5e6"}
}

func (t OneDarkTheme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#21252b", Light: "#f0f0f0"}
}

func (t OneDarkTheme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#3b4048", Light: "#d3d3d3"}
}

func (t OneDarkTheme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#61afef", Light: "#4078f2"}
}

func (t OneDarkTheme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#2c313c", Light: "#e5e5e6"}
}

func init() {
	RegisterTheme("onedark", OneDarkTheme{})
}
