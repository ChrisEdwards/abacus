package ui

import (
	"strings"

	"abacus/internal/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

// Overlay Width System
//
// This file establishes a consistent overlay sizing system to eliminate
// ad-hoc width calculations scattered across individual overlays.
//
// LIPGLOSS WIDTH SEMANTICS (learned from painful experience - see commit 00d370c):
//
//   Style.Width(n)   → sets the CONTENT width (visible area inside padding)
//   Padding(v, h)    → adds to visual width, but is INSIDE the Width value
//   Border           → adds 2 to visual width, but is OUTSIDE the Width value
//
// Example: A box with Width(48), Padding(1,2), RoundedBorder:
//   - Visual width on screen: 48 + 2 (border left+right) = 50 chars
//   - Content width inside padding: 48 - 4 (2 left + 2 right) = 44 chars
//
// The OverlayBuilder handles these calculations automatically.

// Standard overlay content widths (before padding/border).
// These are the widths passed to lipgloss Style.Width().
const (
	// OverlayWidthNarrow is for simple selections (status picker).
	// Visual width ≈ 28 chars.
	OverlayWidthNarrow = 24

	// OverlayWidthStandard is for most modals (labels, delete confirmation).
	// Visual width ≈ 52 chars.
	OverlayWidthStandard = 48

	// OverlayWidthWide is for forms with textareas (comment, create).
	// Visual width ≈ 68 chars.
	OverlayWidthWide = 64

	// overlayHPadding is the horizontal padding used by styleOverlay().
	// Matches the Padding(1, 2) in the style definition.
	overlayHPadding = 2
)

// OverlaySize represents standard overlay sizing presets.
type OverlaySize int

const (
	OverlaySizeNarrow OverlaySize = iota
	OverlaySizeStandard
	OverlaySizeWide
	OverlaySizeResponsive // Uses terminal width
)

// OverlayWidth returns the content width for a given size preset.
// For responsive sizing, pass the terminal width.
func OverlayWidth(size OverlaySize, termWidth int) int {
	switch size {
	case OverlaySizeNarrow:
		return OverlayWidthNarrow
	case OverlaySizeStandard:
		return OverlayWidthStandard
	case OverlaySizeWide:
		return OverlayWidthWide
	case OverlaySizeResponsive:
		return responsiveOverlayWidth(termWidth)
	default:
		return OverlayWidthStandard
	}
}

// responsiveOverlayWidth calculates width based on terminal size.
// Formula: min(120, max(48, int(0.7 * termWidth)))
func responsiveOverlayWidth(termWidth int) int {
	if termWidth == 0 {
		return OverlayWidthStandard
	}
	width := int(float64(termWidth) * 0.7)
	if width < OverlayWidthStandard {
		width = OverlayWidthStandard
	}
	if width > 120 {
		width = 120
	}
	return width
}

// OverlayContentWidth returns the usable width inside an overlay's padding.
// This is what you should use for sizing text content, dividers, etc.
//
//	boxWidth: The lipgloss Width value (e.g., OverlayWidthStandard)
//	Returns:  The width available for content after padding
func OverlayContentWidth(boxWidth int) int {
	inner := boxWidth - (overlayHPadding * 2)
	if inner < 1 {
		return 1
	}
	return inner
}

// OverlayBuilder helps construct consistent overlay content.
// It manages width calculations and provides a fluent API for building
// the standard header/body/footer structure.
type OverlayBuilder struct {
	boxWidth     int      // lipgloss Width value
	contentWidth int      // usable width after padding
	lines        []string // accumulated content lines
}

// NewOverlayBuilder creates a builder with the specified size preset.
func NewOverlayBuilder(size OverlaySize, termWidth int) *OverlayBuilder {
	boxWidth := OverlayWidth(size, termWidth)
	return &OverlayBuilder{
		boxWidth:     boxWidth,
		contentWidth: OverlayContentWidth(boxWidth),
		lines:        make([]string, 0, 16),
	}
}

// NewOverlayBuilderWithWidth creates a builder with an explicit width.
// Use this when migrating existing overlays that need specific widths.
func NewOverlayBuilderWithWidth(boxWidth int) *OverlayBuilder {
	return &OverlayBuilder{
		boxWidth:     boxWidth,
		contentWidth: OverlayContentWidth(boxWidth),
		lines:        make([]string, 0, 16),
	}
}

// BoxWidth returns the lipgloss Width value for styling containers.
func (b *OverlayBuilder) BoxWidth() int {
	return b.boxWidth
}

// ContentWidth returns the usable width for text content.
func (b *OverlayBuilder) ContentWidth() int {
	return b.contentWidth
}

// Header adds a styled title and divider.
func (b *OverlayBuilder) Header(title string) *OverlayBuilder {
	b.lines = append(b.lines, styleOverlayTitle().Render(title))
	b.lines = append(b.lines, b.Divider())
	b.lines = append(b.lines, "")
	return b
}

// HeaderWithContext adds a title, context line, and divider.
// Useful for showing which item the overlay is acting on.
func (b *OverlayBuilder) HeaderWithContext(title, issueID, issueTitle string) *OverlayBuilder {
	b.lines = append(b.lines, styleOverlayTitle().Render(title))
	// Truncate title if needed
	maxTitleLen := b.contentWidth - lipgloss.Width(issueID) - 3
	if maxTitleLen < 10 {
		maxTitleLen = 10
	}
	displayTitle := issueTitle
	if len(displayTitle) > maxTitleLen {
		displayTitle = displayTitle[:maxTitleLen-3] + "..."
	}
	contextLine := styleID().Render(issueID) + styleStatsDim().Render(": ") + styleStatsDim().Render(displayTitle)
	b.lines = append(b.lines, contextLine)
	b.lines = append(b.lines, b.Divider())
	b.lines = append(b.lines, "")
	return b
}

// Divider returns a styled horizontal divider line.
func (b *OverlayBuilder) Divider() string {
	return styleOverlayDivider().Render(strings.Repeat("─", b.contentWidth))
}

// Line adds a content line.
func (b *OverlayBuilder) Line(content string) *OverlayBuilder {
	b.lines = append(b.lines, content)
	return b
}

// Lines adds multiple content lines.
func (b *OverlayBuilder) Lines(content ...string) *OverlayBuilder {
	b.lines = append(b.lines, content...)
	return b
}

// BlankLine adds an empty line for spacing.
func (b *OverlayBuilder) BlankLine() *OverlayBuilder {
	return b.Line("")
}

// Section adds a labeled section with header and content.
func (b *OverlayBuilder) Section(label string, content string) *OverlayBuilder {
	b.lines = append(b.lines, styleOverlaySectionLabel().Render(label))
	b.lines = append(b.lines, content)
	b.lines = append(b.lines, "")
	return b
}

// Footer adds a divider and centered footer hints.
func (b *OverlayBuilder) Footer(hints []footerHint) *OverlayBuilder {
	b.lines = append(b.lines, b.Divider())
	b.lines = append(b.lines, overlayFooterLine(hints, b.contentWidth))
	return b
}

// FooterText adds a divider and custom footer text.
func (b *OverlayBuilder) FooterText(text string) *OverlayBuilder {
	b.lines = append(b.lines, b.Divider())
	centered := lipgloss.NewStyle().
		Width(b.contentWidth).
		Align(lipgloss.Center).
		Background(theme.Current().BackgroundSecondary()).
		Foreground(theme.Current().TextMuted()).
		Render(text)
	b.lines = append(b.lines, centered)
	return b
}

// Build returns the final styled overlay content.
// The overlay is sized to boxWidth to ensure dividers span the full width.
func (b *OverlayBuilder) Build() string {
	content := strings.Join(b.lines, "\n")
	return styleOverlay().Width(b.boxWidth).Render(content)
}

// BuildRaw returns the content without the overlay wrapper style.
// Use this when you need to apply a custom wrapper style.
func (b *OverlayBuilder) BuildRaw() string {
	return strings.Join(b.lines, "\n")
}

// BuildDanger returns the content with danger styling (red border).
// The overlay is sized to boxWidth to ensure dividers span the full width.
func (b *OverlayBuilder) BuildDanger() string {
	content := strings.Join(b.lines, "\n")
	return styleOverlayDanger().Width(b.boxWidth).Render(content)
}

// Overlay Styles
//
// These are THE canonical overlay styles. All overlays should use these
// for consistency. Do not create new overlay style functions.

// styleOverlay is the primary overlay wrapper style.
// Use this for all centered modal overlays.
func styleOverlay() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(theme.Current().BackgroundSecondary()).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().BorderFocused()).
		Padding(1, overlayHPadding)
}

// styleOverlayDanger is for destructive action overlays (delete confirmation).
func styleOverlayDanger() lipgloss.Style {
	return styleOverlay().
		BorderForeground(theme.Current().Error())
}

// styleOverlayTitle styles the overlay header title.
func styleOverlayTitle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Accent()).
		Bold(true)
}

// styleOverlayDivider styles horizontal dividers.
func styleOverlayDivider() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Primary())
}

// styleOverlaySectionLabel styles section labels within overlays.
func styleOverlaySectionLabel() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Secondary()).
		Bold(true)
}

// BaseOverlayLayer creates a centered Layer from an overlay's View function.
// This eliminates the need for copy-pasted Layer() methods in each overlay.
//
// Overlays render with the standard theme. The caller (view.go) is responsible
// for restoring the standard theme before rendering overlays - dimming is only
// applied to background content, not overlays.
func BaseOverlayLayer(viewFn func() string, width, height, topMargin, bottomMargin int) Layer {
	return LayerFunc(func() *Canvas {
		content := viewFn()

		if strings.TrimSpace(content) == "" {
			return nil
		}

		overlayWidth := lipgloss.Width(content)
		if overlayWidth <= 0 {
			return nil
		}
		overlayHeight := lipgloss.Height(content)
		if overlayHeight <= 0 {
			return nil
		}

		surface := NewSecondarySurface(overlayWidth, overlayHeight)
		surface.Draw(0, 0, content)

		x, y := centeredOffsets(width, height, overlayWidth, overlayHeight, topMargin, bottomMargin)
		surface.Canvas.SetOffset(x, y)
		return surface.Canvas
	})
}

// Form Field Helpers
//
// These extend textarea_helpers.go with overlay-aware sizing.

// OverlayTextareaWidth returns the width for a textarea inside an overlay.
// Accounts for the textarea's own border and padding.
//
//	boxWidth: The overlay's boxWidth (from OverlayBuilder.BoxWidth())
//	Returns:  The width to pass to textarea.SetWidth()
func OverlayTextareaWidth(boxWidth int) int {
	// Textarea sits inside overlay padding, and has its own border (2) and padding (2)
	// Content width inside overlay = boxWidth - 4 (overlay padding)
	// Textarea content width = above - 2 (textarea border is outside, but we need to fit)
	// Actually, textarea.SetWidth sets the viewport width, so we just need content width
	return OverlayContentWidth(boxWidth)
}

// OverlayInputStyle returns a bordered input style sized for the overlay.
func OverlayInputStyle(boxWidth int, focused bool) lipgloss.Style {
	borderColor := theme.Current().BorderDim()
	if focused {
		borderColor = theme.Current().Success()
	}
	// The input box should fill the content width
	// Width value = contentWidth, which includes space for the padding we add
	contentWidth := OverlayContentWidth(boxWidth)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(contentWidth + 4) // +4 for our padding (2) inside the border
}

// OverlayInputErrorStyle returns a bordered input style with error highlighting.
func OverlayInputErrorStyle(boxWidth int) lipgloss.Style {
	contentWidth := OverlayContentWidth(boxWidth)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().Error()).
		Padding(0, 1).
		Width(contentWidth + 4)
}
