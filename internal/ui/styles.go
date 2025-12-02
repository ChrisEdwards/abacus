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
	cRed        = lipgloss.Color("203")
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

	// styleCrossHighlight is a muted version for duplicate instances of the same node
	styleCrossHighlight = lipgloss.NewStyle().
				Background(lipgloss.Color("240")).
				Foreground(cLightGray)

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

	styleErrorToast = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cRed).
			Foreground(cWhite).
			Padding(0, 1)

	styleErrorIndicator = lipgloss.NewStyle().
				Foreground(cRed).
				Bold(true)

	styleSuccessToast = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#00FF00")). // Green
				Foreground(cWhite).
				Padding(0, 1)

	// Help overlay styles
	styleHelpOverlay = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(cPurple).
				Padding(1, 2)

	styleHelpTitle = lipgloss.NewStyle().
			Foreground(cGold).
			Bold(true)

	styleHelpDivider = lipgloss.NewStyle().
				Foreground(cPurple)

	styleHelpSectionHeader = lipgloss.NewStyle().
				Foreground(cField).
				Bold(true)

	styleHelpUnderline = lipgloss.NewStyle().
				Foreground(cField)

	styleHelpKey = lipgloss.NewStyle().
			Foreground(cCyan).
			Bold(true)

	styleHelpDesc = lipgloss.NewStyle().
			Foreground(cLightGray)

	styleHelpFooter = lipgloss.NewStyle().
			Foreground(cBrightGray).
			Italic(true)

	// Footer bar styles
	styleKeyPill = lipgloss.NewStyle().
			Background(cPurple).
			Foreground(cWhite).
			Bold(true)

	styleKeyDesc = lipgloss.NewStyle().
			Foreground(cBrightGray)

	styleFooterMuted = lipgloss.NewStyle().
			Foreground(cBrightGray)

	// Status overlay styles
	styleStatusOverlay = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(cPurple).
				Padding(1, 2)

	// Delete overlay style
	styleDeleteOverlay = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(cRed).
				Padding(1, 2)

	styleStatusDivider = lipgloss.NewStyle().
				Foreground(cPurple)

	styleStatusOption = lipgloss.NewStyle().
				Foreground(cWhite)

	styleStatusSelected = lipgloss.NewStyle().
				Foreground(cCyan).
				Bold(true)

	styleStatusDisabled = lipgloss.NewStyle().
				Foreground(cGray)

	// Labels overlay styles
	styleLabelChecked = lipgloss.NewStyle().
				Foreground(cNeonGreen).
				Bold(true)

	styleLabelUnchecked = lipgloss.NewStyle().
				Foreground(cWhite)

	styleLabelCursor = lipgloss.NewStyle().
				Background(cHighlight).
				Foreground(cWhite).
				Bold(true)

	styleLabelNewOption = lipgloss.NewStyle().
				Foreground(cGold)

	// Chip styles for label tokens
	styleChip = lipgloss.NewStyle().
			Foreground(cWhite).
			Background(lipgloss.Color("25")) // Same as styleLabel

	styleChipHighlight = lipgloss.NewStyle().
				Foreground(cWhite).
				Background(cHighlight). // Purple
				Bold(true)

	styleChipFlash = lipgloss.NewStyle().
			Foreground(cWhite).
			Background(cOrange). // Orange flash for duplicate
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
