package theme

import "github.com/charmbracelet/lipgloss"

// GitHub color palette
// https://primer.style/primitives/colors
var github = struct {
	DarkBg      string
	DarkBgAlt   string
	DarkBgPanel string
	DarkFg      string
	DarkFgMuted string
	DarkBlue    string
	DarkGreen   string
	DarkRed     string
	DarkOrange  string
	DarkPurple  string
	DarkYellow  string
	DarkCyan    string
	LightBg     string
	LightBgAlt  string
	LightBgPanel string
	LightFg     string
	LightFgMuted string
	LightBlue   string
	LightGreen  string
	LightRed    string
	LightOrange string
	LightPurple string
	LightYellow string
	LightCyan   string
}{
	DarkBg:       "#0d1117",
	DarkBgAlt:    "#010409",
	DarkBgPanel:  "#161b22",
	DarkFg:       "#c9d1d9",
	DarkFgMuted:  "#8b949e",
	DarkBlue:     "#58a6ff",
	DarkGreen:    "#3fb950",
	DarkRed:      "#f85149",
	DarkOrange:   "#d29922",
	DarkPurple:   "#bc8cff",
	DarkYellow:   "#e3b341",
	DarkCyan:     "#39c5cf",
	LightBg:      "#ffffff",
	LightBgAlt:   "#f6f8fa",
	LightBgPanel: "#f0f3f6",
	LightFg:      "#24292f",
	LightFgMuted: "#57606a",
	LightBlue:    "#0969da",
	LightGreen:   "#1a7f37",
	LightRed:     "#cf222e",
	LightOrange:  "#bc4c00",
	LightPurple:  "#8250df",
	LightYellow:  "#9a6700",
	LightCyan:    "#1b7c83",
}

// GitHubTheme implements Theme with the GitHub color palette.
type GitHubTheme struct{}

// Base colors

func (t GitHubTheme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: github.LightBlue, Dark: github.DarkBlue}
}

func (t GitHubTheme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: github.LightPurple, Dark: github.DarkPurple}
}

func (t GitHubTheme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: github.LightCyan, Dark: github.DarkCyan}
}

// Status colors

func (t GitHubTheme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: github.LightRed, Dark: github.DarkRed}
}

func (t GitHubTheme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: github.LightYellow, Dark: github.DarkYellow}
}

func (t GitHubTheme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: github.LightGreen, Dark: github.DarkGreen}
}

func (t GitHubTheme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: github.LightOrange, Dark: github.DarkOrange}
}

// Text colors

func (t GitHubTheme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: github.LightFg, Dark: github.DarkFg}
}

func (t GitHubTheme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: github.LightFgMuted, Dark: github.DarkFgMuted}
}

func (t GitHubTheme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#000000", Dark: github.DarkFg}
}

// Background colors

func (t GitHubTheme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: github.LightBg, Dark: github.DarkBg}
}

func (t GitHubTheme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: github.LightBgAlt, Dark: "#21262d"}
}

func (t GitHubTheme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: github.LightBgPanel, Dark: github.DarkBgAlt}
}

// Border colors

func (t GitHubTheme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#d0d7de", Dark: "#30363d"}
}

func (t GitHubTheme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: github.LightBlue, Dark: github.DarkBlue}
}

func (t GitHubTheme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#d8dee4", Dark: "#21262d"}
}

func init() {
	RegisterTheme("github", GitHubTheme{})
}
