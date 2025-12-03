package theme

import "github.com/charmbracelet/lipgloss"

// GruvboxTheme implements the Gruvbox color scheme.
type GruvboxTheme struct{}

func (t GruvboxTheme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#83a598", Light: "#076678"}
}

func (t GruvboxTheme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#d3869b", Light: "#8f3f71"}
}

func (t GruvboxTheme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#fabd2f", Light: "#b57614"}
}

func (t GruvboxTheme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#fb4934", Light: "#9d0006"}
}

func (t GruvboxTheme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#fe8019", Light: "#af3a03"}
}

func (t GruvboxTheme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#b8bb26", Light: "#79740e"}
}

func (t GruvboxTheme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#83a598", Light: "#076678"}
}

func (t GruvboxTheme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#ebdbb2", Light: "#3c3836"}
}

func (t GruvboxTheme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#a89984", Light: "#7c6f64"}
}

func (t GruvboxTheme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#fabd2f", Light: "#b57614"}
}

func (t GruvboxTheme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#282828", Light: "#fbf1c7"}
}

func (t GruvboxTheme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#504945", Light: "#ebdbb2"}
}

func (t GruvboxTheme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#1d2021", Light: "#d5c4a1"}
}

func (t GruvboxTheme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#504945", Light: "#bdae93"}
}

func (t GruvboxTheme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#83a598", Light: "#076678"}
}

func (t GruvboxTheme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#3c3836", Light: "#d5c4a1"}
}

func init() {
	RegisterTheme("gruvbox", GruvboxTheme{})
}
