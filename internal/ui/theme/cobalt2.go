package theme

import "github.com/charmbracelet/lipgloss"

// Cobalt2 color palette
// https://github.com/wesbos/cobalt2-vscode
var cobalt2 = struct {
	Background      string
	BackgroundAlt   string
	BackgroundPanel string
	Foreground      string
	ForegroundMuted string
	Yellow          string
	Orange          string
	Mint            string
	Blue            string
	Pink            string
	Green           string
	Purple          string
	Red             string
}{
	Background:      "#193549",
	BackgroundAlt:   "#122738",
	BackgroundPanel: "#1f4662",
	Foreground:      "#ffffff",
	ForegroundMuted: "#adb7c9",
	Yellow:          "#ffc600",
	Orange:          "#ff9d00",
	Mint:            "#2affdf",
	Blue:            "#0088ff",
	Pink:            "#ff628c",
	Green:           "#9eff80",
	Purple:          "#9a5feb",
	Red:             "#ff0088",
}

// Cobalt2Theme implements Theme with the Cobalt2 color palette.
type Cobalt2Theme struct{}

// Base colors

func (t Cobalt2Theme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#0066cc", Dark: cobalt2.Blue}
}

func (t Cobalt2Theme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#7c4dff", Dark: cobalt2.Purple}
}

func (t Cobalt2Theme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#00acc1", Dark: cobalt2.Mint}
}

// Status colors

func (t Cobalt2Theme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#e91e63", Dark: cobalt2.Red}
}

func (t Cobalt2Theme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#ff9800", Dark: cobalt2.Yellow}
}

func (t Cobalt2Theme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#4caf50", Dark: cobalt2.Green}
}

func (t Cobalt2Theme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#ff5722", Dark: cobalt2.Orange}
}

// Text colors

func (t Cobalt2Theme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#193549", Dark: cobalt2.Foreground}
}

func (t Cobalt2Theme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#5c6b7d", Dark: cobalt2.ForegroundMuted}
}

func (t Cobalt2Theme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#000000", Dark: cobalt2.Foreground}
}

// Background colors

func (t Cobalt2Theme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#ffffff", Dark: cobalt2.Background}
}

func (t Cobalt2Theme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#f5f7fa", Dark: cobalt2.BackgroundAlt}
}

func (t Cobalt2Theme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#e8ecf1", Dark: cobalt2.BackgroundPanel}
}

// Border colors

func (t Cobalt2Theme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#d3dae3", Dark: cobalt2.BackgroundPanel}
}

func (t Cobalt2Theme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#0066cc", Dark: cobalt2.Blue}
}

func (t Cobalt2Theme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#e8ecf1", Dark: "#0e1e2e"}
}

func init() {
	RegisterTheme("cobalt2", Cobalt2Theme{})
}
