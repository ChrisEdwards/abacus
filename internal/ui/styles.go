package ui

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"

	"abacus/internal/ui/theme"
)

type styleThemeState struct {
	override theme.ThemeWrapper
	active   bool
	dimmed   bool
}

var currentStyleTheme styleThemeState

// useDimmedTheme sets dimmed palette for background content behind overlays.
// Returns a restore function to revert to the previous theme state.
func useDimmedTheme() func() {
	prev := currentStyleTheme
	currentStyleTheme = styleThemeState{
		override: theme.Current().Dimmed(),
		active:   true,
		dimmed:   true,
	}
	return func() {
		currentStyleTheme = prev
	}
}

func currentThemeWrapper() theme.ThemeWrapper {
	if currentStyleTheme.active {
		return currentStyleTheme.override
	}
	return theme.Current()
}

func stylesDimmed() bool {
	return currentStyleTheme.active && currentStyleTheme.dimmed
}

func applyBold(style lipgloss.Style, always bool) lipgloss.Style {
	if stylesDimmed() && !always {
		return style
	}
	return style.Bold(true)
}

// baseStyle returns a style with the theme background - use as foundation for all styles
func baseStyle() lipgloss.Style {
	bg := currentThemeWrapper().Background()
	return lipgloss.NewStyle().Background(bg)
}

// Status text styles

func styleInProgressText() lipgloss.Style {
	style := baseStyle().Foreground(currentThemeWrapper().Success())
	return applyBold(style, true)
}

func styleNormalText() lipgloss.Style {
	return baseStyle().Foreground(currentThemeWrapper().Text())
}

func styleDoneText() lipgloss.Style {
	return baseStyle().Foreground(currentThemeWrapper().TextMuted())
}

func styleBlockedText() lipgloss.Style {
	return baseStyle().Foreground(currentThemeWrapper().Error())
}

//nolint:unused // Used by ab-lqny and ab-meoq beads for deferred status rendering
func styleDeferredText() lipgloss.Style {
	return baseStyle().Foreground(currentThemeWrapper().TextMuted())
}

func styleStatsDim() lipgloss.Style {
	return baseStyle().Foreground(currentThemeWrapper().TextMuted())
}

func stylePriority() lipgloss.Style {
	return baseStyle().Foreground(currentThemeWrapper().TextMuted())
}

func styleColumnSeparator() lipgloss.Style {
	return baseStyle().Foreground(currentThemeWrapper().BorderNormal())
}

func styleColumnText() lipgloss.Style {
	return baseStyle().Foreground(currentThemeWrapper().TextMuted())
}

// Status icon styles

func styleIconOpen() lipgloss.Style {
	return baseStyle().Foreground(currentThemeWrapper().Text())
}

func styleIconInProgress() lipgloss.Style {
	return baseStyle().Foreground(currentThemeWrapper().Success())
}

func styleIconDone() lipgloss.Style {
	return baseStyle().Foreground(currentThemeWrapper().TextMuted())
}

func styleIconBlocked() lipgloss.Style {
	return baseStyle().Foreground(currentThemeWrapper().Error())
}

//nolint:unused // Used by ab-lqny and ab-meoq beads for deferred status icon
func styleIconDeferred() lipgloss.Style {
	return baseStyle().Foreground(currentThemeWrapper().Secondary())
}

// Tree and list styles

func styleID() lipgloss.Style {
	style := baseStyle().Foreground(currentThemeWrapper().Accent())
	return applyBold(style, false)
}

// App header styles

func styleAppHeader() lipgloss.Style {
	style := lipgloss.NewStyle().
		Foreground(currentThemeWrapper().Accent()).
		Background(currentThemeWrapper().BackgroundSecondary()).
		Padding(0, 1)
	return applyBold(style, false)
}

func styleFilterInfo() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(currentThemeWrapper().Secondary()).
		Background(currentThemeWrapper().BackgroundSecondary())
}

// Pane styles

func stylePane() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(currentThemeWrapper().BorderNormal()).
		BorderBackground(currentThemeWrapper().Background()).
		Background(currentThemeWrapper().Background())
}

func stylePaneFocused() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(currentThemeWrapper().BorderFocused()).
		BorderBackground(currentThemeWrapper().Background()).
		Background(currentThemeWrapper().Background())
}

// Detail view styles

func styleDetailHeaderBlock() lipgloss.Style {
	style := lipgloss.NewStyle().
		Background(currentThemeWrapper().BackgroundSecondary()).
		Foreground(currentThemeWrapper().Text()).
		Padding(0, 1)
	return applyBold(style, false)
}

func styleDetailHeaderCombined() lipgloss.Style {
	style := lipgloss.NewStyle().
		Background(currentThemeWrapper().BackgroundSecondary())
	return applyBold(style, false)
}

func styleField() lipgloss.Style {
	style := baseStyle().
		Foreground(currentThemeWrapper().Secondary()).
		Width(12)
	return applyBold(style, false)
}

func styleVal() lipgloss.Style {
	return baseStyle().Foreground(currentThemeWrapper().Text())
}

func styleSectionHeader() lipgloss.Style {
	style := baseStyle().
		Foreground(currentThemeWrapper().Secondary())
	return applyBold(style, false)
}

// Label and priority badge styles

func stylePrio() lipgloss.Style {
	style := lipgloss.NewStyle().
		Foreground(currentThemeWrapper().Text()).
		Background(currentThemeWrapper().Warning()).
		Padding(0, 1)
	return applyBold(style, false)
}

func styleCommentHeader() lipgloss.Style {
	style := baseStyle().
		Foreground(currentThemeWrapper().TextMuted())
	return applyBold(style, false)
}

// Toast styles

func styleErrorToast() lipgloss.Style {
	return baseStyle().
		Border(lipgloss.RoundedBorder()).
		BorderBackground(currentThemeWrapper().Background()).
		BorderForeground(currentThemeWrapper().Error()).
		Foreground(currentThemeWrapper().Text()).
		Padding(0, 1)
}

func styleErrorIndicator() lipgloss.Style {
	style := lipgloss.NewStyle().
		Foreground(currentThemeWrapper().Error())
	return applyBold(style, true)
}

func styleSuccessToast() lipgloss.Style {
	return baseStyle().
		Border(lipgloss.RoundedBorder()).
		BorderBackground(currentThemeWrapper().Background()).
		BorderForeground(currentThemeWrapper().Success()).
		Foreground(currentThemeWrapper().Text()).
		Padding(0, 1)
}

// Help overlay styles

func styleHelpOverlay() lipgloss.Style {
	return baseStyle().
		Background(currentThemeWrapper().BackgroundSecondary()).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(currentThemeWrapper().BorderFocused()).
		Padding(1, 2)
}

func styleHelpTitle() lipgloss.Style {
	style := lipgloss.NewStyle().
		Foreground(currentThemeWrapper().Accent())
	return applyBold(style, false)
}

func styleHelpDivider() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(currentThemeWrapper().Primary())
}

func styleHelpSectionHeader() lipgloss.Style {
	style := lipgloss.NewStyle().
		Foreground(currentThemeWrapper().Secondary())
	return applyBold(style, false)
}

func styleHelpUnderline() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(currentThemeWrapper().Secondary())
}

func styleHelpKey() lipgloss.Style {
	style := lipgloss.NewStyle().
		Foreground(currentThemeWrapper().Accent())
	return applyBold(style, false)
}

func styleHelpDesc() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(currentThemeWrapper().TextMuted())
}

func styleHelpFooter() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(currentThemeWrapper().TextMuted()).
		Italic(true)
}

// Footer bar styles

func styleKeyPill() lipgloss.Style {
	style := lipgloss.NewStyle().
		Background(currentThemeWrapper().BackgroundSecondary()).
		Foreground(currentThemeWrapper().Accent())
	return applyBold(style, false)
}

func styleKeyDesc() lipgloss.Style {
	return baseStyle().
		Foreground(currentThemeWrapper().TextMuted())
}

func styleFooterMuted() lipgloss.Style {
	return baseStyle().
		Foreground(currentThemeWrapper().TextMuted())
}

// Status overlay styles (moved to overlay_base.go for unified overlay framework)

func styleStatusOption() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(currentThemeWrapper().Text())
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

func buildMarkdownRenderer(format string, width int) func(string) string {
	fallback := func(input string) string {
		// When dimmed, render with muted text style for consistency
		wrapped := wordwrap.String(input, width)
		if stylesDimmed() {
			return styleNormalText().Render(wrapped)
		}
		return wrapped
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
		// When dimmed, use plain text to respect dimmed theme colors
		if stylesDimmed() {
			return fallback(input)
		}
		out, err := renderer.Render(input)
		if err != nil {
			return fallback(input)
		}
		return strings.TrimSpace(out)
	}
}
