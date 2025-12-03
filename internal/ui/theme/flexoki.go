package theme

import "github.com/charmbracelet/lipgloss"

// Flexoki color palette
// https://stephango.com/flexoki
var flexoki = struct {
	Black     string
	Base950   string
	Base900   string
	Base800   string
	Base700   string
	Base600   string
	Base300   string
	Base200   string
	Paper     string
	Red400    string
	Red600    string
	Orange400 string
	Orange600 string
	Yellow400 string
	Yellow600 string
	Green400  string
	Green600  string
	Cyan400   string
	Cyan600   string
	Blue400   string
	Blue600   string
	Purple400 string
	Purple600 string
}{
	Black:     "#100F0F",
	Base950:   "#1C1B1A",
	Base900:   "#282726",
	Base800:   "#403E3C",
	Base700:   "#575653",
	Base600:   "#6F6E69",
	Base300:   "#B7B5AC",
	Base200:   "#CECDC3",
	Paper:     "#FFFCF0",
	Red400:    "#D14D41",
	Red600:    "#AF3029",
	Orange400: "#DA702C",
	Orange600: "#BC5215",
	Yellow400: "#D0A215",
	Yellow600: "#AD8301",
	Green400:  "#879A39",
	Green600:  "#66800B",
	Cyan400:   "#3AA99F",
	Cyan600:   "#24837B",
	Blue400:   "#4385BE",
	Blue600:   "#205EA6",
	Purple400: "#8B7EC8",
	Purple600: "#5E409D",
}

// FlexokiTheme implements Theme with the Flexoki color palette.
type FlexokiTheme struct{}

// Base colors

func (t FlexokiTheme) Primary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: flexoki.Blue600, Dark: flexoki.Orange400}
}

func (t FlexokiTheme) Secondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: flexoki.Purple600, Dark: flexoki.Blue400}
}

func (t FlexokiTheme) Accent() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: flexoki.Orange600, Dark: flexoki.Purple400}
}

// Status colors

func (t FlexokiTheme) Error() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: flexoki.Red600, Dark: flexoki.Red400}
}

func (t FlexokiTheme) Warning() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: flexoki.Orange600, Dark: flexoki.Orange400}
}

func (t FlexokiTheme) Success() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: flexoki.Green600, Dark: flexoki.Green400}
}

func (t FlexokiTheme) Info() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: flexoki.Cyan600, Dark: flexoki.Cyan400}
}

// Text colors

func (t FlexokiTheme) Text() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: flexoki.Black, Dark: flexoki.Base200}
}

func (t FlexokiTheme) TextMuted() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: flexoki.Base600, Dark: flexoki.Base600}
}

func (t FlexokiTheme) TextEmphasized() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: flexoki.Black, Dark: flexoki.Base200}
}

// Background colors

func (t FlexokiTheme) Background() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: flexoki.Paper, Dark: flexoki.Black}
}

func (t FlexokiTheme) BackgroundSecondary() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#F2F0E5", Dark: flexoki.Base900}
}

func (t FlexokiTheme) BackgroundDarker() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#E6E4D9", Dark: flexoki.Base900}
}

// Border colors

func (t FlexokiTheme) BorderNormal() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: flexoki.Base300, Dark: flexoki.Base700}
}

func (t FlexokiTheme) BorderFocused() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#878580", Dark: flexoki.Base600}
}

func (t FlexokiTheme) BorderDim() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: flexoki.Base200, Dark: flexoki.Base800}
}

func init() {
	RegisterTheme("flexoki", FlexokiTheme{})
}
