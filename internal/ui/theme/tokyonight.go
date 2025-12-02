package theme

import "github.com/charmbracelet/lipgloss"

// TokyoNightTheme implements the Tokyo Night color scheme.
type TokyoNightTheme struct{}

func (t TokyoNightTheme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#82aaff", Light: "#2e7de9"}
}

func (t TokyoNightTheme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#c099ff", Light: "#9854f1"}
}

func (t TokyoNightTheme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#ff966c", Light: "#b15c00"}
}

func (t TokyoNightTheme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#ff757f", Light: "#f52a65"}
}

func (t TokyoNightTheme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#ff966c", Light: "#b15c00"}
}

func (t TokyoNightTheme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#c3e88d", Light: "#587539"}
}

func (t TokyoNightTheme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#7dcfff", Light: "#0db9d7"}
}

func (t TokyoNightTheme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#c8d3f5", Light: "#3760bf"}
}

func (t TokyoNightTheme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#636da6", Light: "#848cb5"}
}

func (t TokyoNightTheme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#ffc777", Light: "#8c6c3e"}
}

func (t TokyoNightTheme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#222436", Light: "#e1e2e7"}
}

func (t TokyoNightTheme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#2f334d", Light: "#c8c9ce"}
}

func (t TokyoNightTheme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#1e2030", Light: "#d5d6db"}
}

func (t TokyoNightTheme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#3b4261", Light: "#a8aecb"}
}

func (t TokyoNightTheme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#82aaff", Light: "#2e7de9"}
}

func (t TokyoNightTheme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#292e42", Light: "#c8c9ce"}
}

func init() {
	RegisterTheme("tokyonight", TokyoNightTheme{})
}
