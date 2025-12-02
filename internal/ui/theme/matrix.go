package theme

import "github.com/charmbracelet/lipgloss"

// Matrix color palette - inspired by The Matrix films
var matrix = struct {
	MatrixInk0   string
	MatrixInk1   string
	MatrixInk2   string
	MatrixInk3   string
	RainGreen    string
	RainGreenDim string
	RainGreenHi  string
	RainCyan     string
	RainTeal     string
	RainPurple   string
	RainOrange   string
	AlertRed     string
	AlertYellow  string
	AlertBlue    string
	RainGray     string
	LightBg      string
	LightPaper   string
	LightInk1    string
	LightText    string
	LightGray    string
}{
	MatrixInk0:   "#0a0e0a",
	MatrixInk1:   "#0e130d",
	MatrixInk2:   "#141c12",
	MatrixInk3:   "#1e2a1b",
	RainGreen:    "#2eff6a",
	RainGreenDim: "#1cc24b",
	RainGreenHi:  "#62ff94",
	RainCyan:     "#00efff",
	RainTeal:     "#24f6d9",
	RainPurple:   "#c770ff",
	RainOrange:   "#ffa83d",
	AlertRed:     "#ff4b4b",
	AlertYellow:  "#e6ff57",
	AlertBlue:    "#30b3ff",
	RainGray:     "#8ca391",
	LightBg:      "#eef3ea",
	LightPaper:   "#e4ebe1",
	LightInk1:    "#dae1d7",
	LightText:    "#203022",
	LightGray:    "#748476",
}

// MatrixTheme implements Theme with the Matrix color palette.
type MatrixTheme struct{}

// Base colors

func (t MatrixTheme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: matrix.RainGreenDim, Dark: matrix.RainGreen}
}

func (t MatrixTheme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: matrix.RainTeal, Dark: matrix.RainCyan}
}

func (t MatrixTheme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: matrix.RainPurple, Dark: matrix.RainPurple}
}

// Status colors

func (t MatrixTheme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: matrix.AlertRed, Dark: matrix.AlertRed}
}

func (t MatrixTheme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: matrix.AlertYellow, Dark: matrix.AlertYellow}
}

func (t MatrixTheme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: matrix.RainGreenDim, Dark: matrix.RainGreenHi}
}

func (t MatrixTheme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: matrix.AlertBlue, Dark: matrix.AlertBlue}
}

// Text colors

func (t MatrixTheme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: matrix.LightText, Dark: matrix.RainGreenHi}
}

func (t MatrixTheme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: matrix.LightGray, Dark: matrix.RainGray}
}

func (t MatrixTheme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#000000", Dark: matrix.RainGreenHi}
}

// Background colors

func (t MatrixTheme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: matrix.LightBg, Dark: matrix.MatrixInk0}
}

func (t MatrixTheme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: matrix.LightPaper, Dark: matrix.MatrixInk1}
}

func (t MatrixTheme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: matrix.LightInk1, Dark: matrix.MatrixInk2}
}

// Border colors

func (t MatrixTheme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: matrix.LightGray, Dark: matrix.MatrixInk3}
}

func (t MatrixTheme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: matrix.RainGreenDim, Dark: matrix.RainGreen}
}

func (t MatrixTheme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: matrix.LightInk1, Dark: matrix.MatrixInk2}
}

func init() {
	RegisterTheme("matrix", MatrixTheme{})
}
