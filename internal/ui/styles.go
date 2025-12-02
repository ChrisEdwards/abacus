package ui

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"

	"abacus/internal/ui/theme"
)

// Status text styles

func styleInProgressText() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().Success()).Bold(true)
}

func styleNormalText() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().Text())
}

func styleDoneText() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().TextMuted())
}

func styleBlockedText() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().Error())
}

func styleStatsDim() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().TextMuted())
}

// Status icon styles

func styleIconOpen() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().Text())
}

func styleIconInProgress() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().Success())
}

func styleIconDone() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().TextMuted())
}

func styleIconBlocked() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().Error())
}

// Tree and list styles

func styleID() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().Accent()).Bold(true)
}

func styleSelected() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(theme.Current().BackgroundSecondary()).
		Foreground(theme.Current().Primary()).
		Bold(true)
}

// styleCrossHighlight is a muted version for duplicate instances of the same node
func styleCrossHighlight() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(theme.Current().BorderNormal()).
		Foreground(theme.Current().TextMuted())
}

// App header styles

func styleAppHeader() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Accent()).
		Background(theme.Current().BackgroundSecondary()).
		Bold(true).
		Padding(0, 1)
}

func styleFilterInfo() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Secondary()).
		Background(theme.Current().BackgroundSecondary())
}

// Pane styles

func stylePane() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(theme.Current().BorderNormal())
}

func stylePaneFocused() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(theme.Current().BorderFocused())
}

// Detail view styles

func styleDetailHeaderBlock() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(theme.Current().BackgroundSecondary()).
		Foreground(theme.Current().Text()).
		Bold(true).
		Padding(0, 1)
}

func styleDetailHeaderCombined() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(theme.Current().BackgroundSecondary()).
		Bold(true)
}

func styleField() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Secondary()).
		Bold(true).
		Width(12)
}

func styleVal() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().Text())
}

func styleSectionHeader() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Accent()).
		Bold(true).
		MarginLeft(detailSectionLabelIndent)
}

// Label and priority badge styles

func styleLabel() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Text()).
		Background(theme.Current().BackgroundDarker()).
		Padding(0, 1).
		MarginRight(1).
		Bold(true)
}

func stylePrio() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Text()).
		Background(theme.Current().Warning()).
		Padding(0, 1).
		Bold(true)
}

func styleCommentHeader() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().TextMuted()).
		Bold(true)
}

// Toast styles

func styleErrorToast() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().Error()).
		Foreground(theme.Current().Text()).
		Padding(0, 1)
}

func styleErrorIndicator() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Error()).
		Bold(true)
}

func styleSuccessToast() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().Success()).
		Foreground(theme.Current().Text()).
		Padding(0, 1)
}

// Help overlay styles

func styleHelpOverlay() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().BorderFocused()).
		Padding(1, 2)
}

func styleHelpTitle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Accent()).
		Bold(true)
}

func styleHelpDivider() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Primary())
}

func styleHelpSectionHeader() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Secondary()).
		Bold(true)
}

func styleHelpUnderline() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Secondary())
}

func styleHelpKey() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Accent()).
		Bold(true)
}

func styleHelpDesc() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().TextMuted())
}

func styleHelpFooter() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().TextMuted()).
		Italic(true)
}

// Footer bar styles

func styleKeyPill() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(theme.Current().BackgroundSecondary()).
		Foreground(theme.Current().Accent()).
		Bold(true)
}

func styleKeyDesc() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().TextMuted())
}

func styleFooterMuted() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().TextMuted())
}

// Status overlay styles

func styleStatusOverlay() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().BorderFocused()).
		Padding(1, 2)
}

func styleDeleteOverlay() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().Error()).
		Padding(1, 2)
}

func styleStatusDivider() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Primary())
}

func styleStatusOption() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Text())
}

func styleStatusSelected() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Primary()).
		Bold(true)
}

func styleStatusDisabled() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().BorderNormal())
}

// Labels overlay styles

func styleLabelChecked() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Success()).
		Bold(true)
}

func styleLabelUnchecked() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Text())
}

func styleLabelCursor() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(theme.Current().BackgroundSecondary()).
		Foreground(theme.Current().Text()).
		Bold(true)
}

func styleLabelNewOption() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Accent())
}

// Chip styles for label tokens

func styleChip() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Text()).
		Background(theme.Current().BackgroundDarker())
}

func styleChipHighlight() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Text()).
		Background(theme.Current().BackgroundSecondary()).
		Bold(true)
}

func styleChipFlash() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Text()).
		Background(theme.Current().Warning()).
		Bold(true)
}

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
