package theme

import "github.com/charmbracelet/lipgloss"

// CatppuccinTheme implements the Catppuccin Mocha color scheme.
type CatppuccinTheme struct{}

func (t CatppuccinTheme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#89b4fa", Light: "#1e66f5"}
}

func (t CatppuccinTheme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#cba6f7", Light: "#8839ef"}
}

func (t CatppuccinTheme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#fab387", Light: "#fe640b"}
}

func (t CatppuccinTheme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#f38ba8", Light: "#d20f39"}
}

func (t CatppuccinTheme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#fab387", Light: "#fe640b"}
}

func (t CatppuccinTheme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#a6e3a1", Light: "#40a02b"}
}

func (t CatppuccinTheme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#89b4fa", Light: "#1e66f5"}
}

func (t CatppuccinTheme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#cdd6f4", Light: "#4c4f69"}
}

func (t CatppuccinTheme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#6c7086", Light: "#9ca0b0"}
}

func (t CatppuccinTheme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#f5e0dc", Light: "#dc8a78"}
}

func (t CatppuccinTheme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#1e1e2e", Light: "#eff1f5"}
}

func (t CatppuccinTheme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#313244", Light: "#e6e9ef"}
}

func (t CatppuccinTheme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#181825", Light: "#dce0e8"}
}

func (t CatppuccinTheme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#6c7086", Light: "#9ca0b0"}
}

func (t CatppuccinTheme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#89b4fa", Light: "#1e66f5"}
}

func (t CatppuccinTheme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#45475a", Light: "#ccd0da"}
}

func init() {
	RegisterTheme("catppuccin", CatppuccinTheme{})
}
