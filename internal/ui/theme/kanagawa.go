package theme

import "github.com/charmbracelet/lipgloss"

// Kanagawa color palette
// https://github.com/rebelot/kanagawa.nvim
var kanagawa = struct {
	SumiInk0    string
	SumiInk1    string
	SumiInk2    string
	SumiInk3    string
	FujiWhite   string
	FujiGray    string
	OniViolet   string
	CrystalBlue string
	CarpYellow  string
	SakuraPink  string
	WaveAqua    string
	RoninYellow string
	DragonRed   string
	LotusGreen  string
	WaveBlue    string
	LightBg     string
	LightPaper  string
	LightText   string
	LightGray   string
}{
	SumiInk0:    "#1F1F28",
	SumiInk1:    "#2A2A37",
	SumiInk2:    "#363646",
	SumiInk3:    "#54546D",
	FujiWhite:   "#DCD7BA",
	FujiGray:    "#727169",
	OniViolet:   "#957FB8",
	CrystalBlue: "#7E9CD8",
	CarpYellow:  "#C38D9D",
	SakuraPink:  "#D27E99",
	WaveAqua:    "#76946A",
	RoninYellow: "#D7A657",
	DragonRed:   "#E82424",
	LotusGreen:  "#98BB6C",
	WaveBlue:    "#2D4F67",
	LightBg:     "#F2E9DE",
	LightPaper:  "#EAE4D7",
	LightText:   "#54433A",
	LightGray:   "#9E9389",
}

// KanagawaTheme implements Theme with the Kanagawa color palette.
type KanagawaTheme struct{}

// Base colors

func (t KanagawaTheme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: kanagawa.WaveBlue, Dark: kanagawa.CrystalBlue}
}

func (t KanagawaTheme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: kanagawa.OniViolet, Dark: kanagawa.OniViolet}
}

func (t KanagawaTheme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: kanagawa.SakuraPink, Dark: kanagawa.SakuraPink}
}

// Status colors

func (t KanagawaTheme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: kanagawa.DragonRed, Dark: kanagawa.DragonRed}
}

func (t KanagawaTheme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: kanagawa.RoninYellow, Dark: kanagawa.RoninYellow}
}

func (t KanagawaTheme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: kanagawa.LotusGreen, Dark: kanagawa.LotusGreen}
}

func (t KanagawaTheme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: kanagawa.WaveAqua, Dark: kanagawa.WaveAqua}
}

// Text colors

func (t KanagawaTheme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: kanagawa.LightText, Dark: kanagawa.FujiWhite}
}

func (t KanagawaTheme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: kanagawa.LightGray, Dark: kanagawa.FujiGray}
}

func (t KanagawaTheme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#000000", Dark: kanagawa.FujiWhite}
}

// Background colors

func (t KanagawaTheme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: kanagawa.LightBg, Dark: kanagawa.SumiInk0}
}

func (t KanagawaTheme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: kanagawa.LightPaper, Dark: kanagawa.SumiInk1}
}

func (t KanagawaTheme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#E3DCD2", Dark: kanagawa.SumiInk2}
}

// Border colors

func (t KanagawaTheme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#D4CBBF", Dark: kanagawa.SumiInk3}
}

func (t KanagawaTheme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: kanagawa.CarpYellow, Dark: kanagawa.CarpYellow}
}

func (t KanagawaTheme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#DCD4C9", Dark: kanagawa.SumiInk2}
}

func init() {
	RegisterTheme("kanagawa", KanagawaTheme{})
}
