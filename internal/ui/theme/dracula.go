package theme

import "github.com/charmbracelet/lipgloss"

// Dracula color palette
// https://draculatheme.com/contribute
var dracula = struct {
	Background  string
	CurrentLine string
	Foreground  string
	Comment     string
	Cyan        string
	Green       string
	Orange      string
	Pink        string
	Purple      string
	Red         string
	Yellow      string
}{
	Background:  "#282a36",
	CurrentLine: "#44475a",
	Foreground:  "#f8f8f2",
	Comment:     "#6272a4",
	Cyan:        "#8be9fd",
	Green:       "#50fa7b",
	Orange:      "#ffb86c",
	Pink:        "#ff79c6",
	Purple:      "#bd93f9",
	Red:         "#ff5555",
	Yellow:      "#f1fa8c",
}

// DraculaTheme implements Theme with the Dracula color palette.
type DraculaTheme struct{}

// Base colors

func (d DraculaTheme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#7e57c2", Dark: dracula.Purple}
}

func (d DraculaTheme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#0097a7", Dark: dracula.Cyan}
}

func (d DraculaTheme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#f9a825", Dark: dracula.Yellow}
}

// Status colors

func (d DraculaTheme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#d32f2f", Dark: dracula.Red}
}

func (d DraculaTheme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#ef6c00", Dark: dracula.Orange}
}

func (d DraculaTheme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#388e3c", Dark: dracula.Green}
}

func (d DraculaTheme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#1976d2", Dark: dracula.Cyan}
}

// Text colors

func (d DraculaTheme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#212121", Dark: dracula.Foreground}
}

func (d DraculaTheme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#757575", Dark: dracula.Comment}
}

func (d DraculaTheme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#000000", Dark: dracula.Foreground}
}

// Background colors

func (d DraculaTheme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#ffffff", Dark: dracula.Background}
}

func (d DraculaTheme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#e0e0e0", Dark: dracula.CurrentLine}
}

func (d DraculaTheme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#bdbdbd", Dark: "#1e1f29"}
}

// Border colors

func (d DraculaTheme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#bdbdbd", Dark: dracula.Comment}
}

func (d DraculaTheme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#7e57c2", Dark: dracula.Purple}
}

func (d DraculaTheme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#e0e0e0", Dark: dracula.CurrentLine}
}

func init() {
	RegisterTheme("dracula", DraculaTheme{})
}
