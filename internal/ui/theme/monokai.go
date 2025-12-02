package theme

import "github.com/charmbracelet/lipgloss"

// MonokaiTheme implements the Monokai Pro color scheme.
type MonokaiTheme struct{}

func (t MonokaiTheme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#78dce8", Light: "#0095a8"}
}

func (t MonokaiTheme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#ab9df2", Light: "#6e5494"}
}

func (t MonokaiTheme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#ffd866", Light: "#c18401"}
}

func (t MonokaiTheme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#ff6188", Light: "#d32f2f"}
}

func (t MonokaiTheme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#fc9867", Light: "#e65100"}
}

func (t MonokaiTheme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#a9dc76", Light: "#388e3c"}
}

func (t MonokaiTheme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#78dce8", Light: "#0095a8"}
}

func (t MonokaiTheme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#fcfcfa", Light: "#2d2a2e"}
}

func (t MonokaiTheme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#727072", Light: "#939293"}
}

func (t MonokaiTheme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#ffd866", Light: "#c18401"}
}

func (t MonokaiTheme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#2d2a2e", Light: "#fafafa"}
}

func (t MonokaiTheme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#403e41", Light: "#e8e8e8"}
}

func (t MonokaiTheme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#221f22", Light: "#f0f0f0"}
}

func (t MonokaiTheme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#5b595c", Light: "#d0d0d0"}
}

func (t MonokaiTheme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#78dce8", Light: "#0095a8"}
}

func (t MonokaiTheme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: "#403e41", Light: "#e8e8e8"}
}

func init() {
	RegisterTheme("monokai", MonokaiTheme{})
}
