package theme

import "github.com/charmbracelet/lipgloss"

// Synthwave '84 color palette
// https://github.com/robb0wen/synthwave-vscode
var synthwave84 = struct {
	Background      string
	BackgroundAlt   string
	BackgroundPanel string
	Foreground      string
	ForegroundMuted string
	Pink            string
	Cyan            string
	Yellow          string
	Orange          string
	Purple          string
	Red             string
	Green           string
}{
	Background:      "#262335",
	BackgroundAlt:   "#1e1a29",
	BackgroundPanel: "#2a2139",
	Foreground:      "#ffffff",
	ForegroundMuted: "#848bbd",
	Pink:            "#ff7edb",
	Cyan:            "#36f9f6",
	Yellow:          "#fede5d",
	Orange:          "#ff8b39",
	Purple:          "#b084eb",
	Red:             "#fe4450",
	Green:           "#72f1b8",
}

// Synthwave84Theme implements Theme with the Synthwave '84 color palette.
type Synthwave84Theme struct{}

// Base colors

func (t Synthwave84Theme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#00bcd4", Dark: synthwave84.Cyan}
}

func (t Synthwave84Theme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#e91e63", Dark: synthwave84.Pink}
}

func (t Synthwave84Theme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#9c27b0", Dark: synthwave84.Purple}
}

// Status colors

func (t Synthwave84Theme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#f44336", Dark: synthwave84.Red}
}

func (t Synthwave84Theme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#ff9800", Dark: synthwave84.Yellow}
}

func (t Synthwave84Theme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#4caf50", Dark: synthwave84.Green}
}

func (t Synthwave84Theme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#ff5722", Dark: synthwave84.Orange}
}

// Text colors

func (t Synthwave84Theme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#262335", Dark: synthwave84.Foreground}
}

func (t Synthwave84Theme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#5c5c8a", Dark: synthwave84.ForegroundMuted}
}

func (t Synthwave84Theme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#000000", Dark: synthwave84.Foreground}
}

// Background colors

func (t Synthwave84Theme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#fafafa", Dark: synthwave84.Background}
}

func (t Synthwave84Theme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#f5f5f5", Dark: synthwave84.BackgroundAlt}
}

func (t Synthwave84Theme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#eeeeee", Dark: synthwave84.BackgroundPanel}
}

// Border colors

func (t Synthwave84Theme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#e0e0e0", Dark: "#495495"}
}

func (t Synthwave84Theme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#00bcd4", Dark: synthwave84.Cyan}
}

func (t Synthwave84Theme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#f0f0f0", Dark: "#241b2f"}
}

func init() {
	RegisterTheme("synthwave84", Synthwave84Theme{})
}
