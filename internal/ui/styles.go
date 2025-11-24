package ui

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

var (
	cPurple     = lipgloss.Color("99")
	cCyan       = lipgloss.Color("39")
	cNeonGreen  = lipgloss.Color("118")
	cRed        = lipgloss.Color("196")
	cOrange     = lipgloss.Color("208")
	cGold       = lipgloss.Color("220")
	cGray       = lipgloss.Color("240")
	cBrightGray = lipgloss.Color("246")
	cLightGray  = lipgloss.Color("250")
	cWhite      = lipgloss.Color("255")
	cHighlight  = lipgloss.Color("57")
	cField      = lipgloss.Color("63")

	styleInProgressText = lipgloss.NewStyle().Foreground(cCyan).Bold(true)
	styleNormalText     = lipgloss.NewStyle().Foreground(cWhite)
	styleDoneText       = lipgloss.NewStyle().Foreground(cBrightGray)
	styleBlockedText    = lipgloss.NewStyle().Foreground(cRed)
	styleStatsDim       = lipgloss.NewStyle().Foreground(cBrightGray)

	styleIconOpen       = lipgloss.NewStyle().Foreground(cWhite)
	styleIconInProgress = lipgloss.NewStyle().Foreground(cNeonGreen)
	styleIconDone       = lipgloss.NewStyle().Foreground(cBrightGray)
	styleIconBlocked    = lipgloss.NewStyle().Foreground(cRed)

	styleID = lipgloss.NewStyle().Foreground(cGold).Bold(true)

	styleSelected = lipgloss.NewStyle().
			Background(cHighlight).
			Foreground(cWhite).
			Bold(true)

	styleAppHeader = lipgloss.NewStyle().
			Foreground(cWhite).
			Background(cPurple).
			Bold(true).
			Padding(0, 1)

	styleFilterInfo = lipgloss.NewStyle().
			Foreground(cLightGray).
			Background(cPurple)

	stylePane = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(cGray)

	stylePaneFocused = lipgloss.NewStyle().
				Border(lipgloss.ThickBorder()).
				BorderForeground(cPurple)

	styleDetailHeaderBlock = lipgloss.NewStyle().
				Background(cHighlight).
				Foreground(cWhite).
				Bold(true).
				Padding(0, 1)

	styleDetailHeaderCombined = lipgloss.NewStyle().
					Background(cHighlight).
					Bold(true)

	styleField = lipgloss.NewStyle().
			Foreground(cField).
			Bold(true).
			Width(12)

	styleVal = lipgloss.NewStyle().Foreground(cWhite)

	styleSectionHeader = lipgloss.NewStyle().
				Foreground(cGold).
				Bold(true).
				MarginLeft(detailSectionLabelIndent)

	styleLabel = lipgloss.NewStyle().
			Foreground(cWhite).
			Background(lipgloss.Color("25")).
			Padding(0, 1).
			MarginRight(1).
			Bold(true)

	stylePrio = lipgloss.NewStyle().
			Foreground(cWhite).
			Background(cOrange).
			Padding(0, 1).
			Bold(true)

	styleCommentHeader = lipgloss.NewStyle().
				Foreground(cBrightGray).
				Bold(true)
)

func buildMarkdownRenderer(format string, width int) func(string) string {
	fallback := func(input string) string {
		return wordwrap.String(input, width)
	}

	style := strings.ToLower(strings.TrimSpace(format))
	if style == "" || style == "rich" || style == "dark" {
		style = "dark"
	}
	if style == "plain" {
		return fallback
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return fallback
	}
	return func(input string) string {
		out, err := renderer.Render(input)
		if err != nil {
			return fallback(input)
		}
		return strings.TrimSpace(out)
	}
}
