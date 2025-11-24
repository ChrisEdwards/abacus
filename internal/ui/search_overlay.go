package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	overlayMinWidth       = 60
	overlayMaxWidth       = 80
	overlayHorizontalPad  = 4 // styleSearchOverlayModal padding left+right
	overlayVerticalTarget = 0.25
)

// SearchOverlay renders the modal search UI with a dimmed scrim background.
type SearchOverlay struct {
	Title       string
	Suggestions []string
	emptyHint   string

	tokens       []SearchToken
	mode         SuggestionMode
	pendingField string
	pendingText  string
}

// NewSearchOverlay builds a modal with sensible defaults.
func NewSearchOverlay() SearchOverlay {
	overlay := SearchOverlay{
		Title:     "Smart Filter Search",
		emptyHint: "Type to search by title or bead id",
	}
	overlay.UpdateInput("")
	return overlay
}

// SetSuggestions replaces the current suggestion list.
func (o *SearchOverlay) SetSuggestions(items []string) {
	if o == nil {
		return
	}
	o.Suggestions = append([]string(nil), items...)
}

// UpdateInput parses the raw input and updates tokens/mode metadata.
func (o *SearchOverlay) UpdateInput(input string) {
	if o == nil {
		return
	}
	result := parseSearchInput(input)
	o.tokens = result.tokens
	o.mode = result.mode
	o.pendingField = result.pendingField
	o.pendingText = result.pendingText
}

// Tokens returns a copy of the parsed tokens slice.
func (o SearchOverlay) Tokens() []SearchToken {
	if len(o.tokens) == 0 {
		return nil
	}
	cloned := make([]SearchToken, len(o.tokens))
	copy(cloned, o.tokens)
	return cloned
}

// SuggestionMode reports whether the next suggestions should target fields or values.
func (o SearchOverlay) SuggestionMode() SuggestionMode {
	return o.mode
}

// PendingField returns the field for which a value is currently being entered.
func (o SearchOverlay) PendingField() string {
	return o.pendingField
}

// PendingText returns the raw text that has not yet been committed to a token.
func (o SearchOverlay) PendingText() string {
	return o.pendingText
}

// InputWidth returns the usable width for the text input based on the current
// terminal width and modal padding.
func (o SearchOverlay) InputWidth(containerWidth int) int {
	width := o.modalWidth(containerWidth) - overlayHorizontalPad
	if width < 20 {
		width = 20
	}
	return width
}

// View renders the modal overlay with the provided input content.
func (o SearchOverlay) View(inputView string, containerWidth, containerHeight int) string {
	if containerWidth <= 0 {
		containerWidth = overlayMaxWidth
	}
	if containerHeight <= 0 {
		containerHeight = 24
	}
	modalWidth := o.modalWidth(containerWidth)
	body := lipgloss.JoinVertical(lipgloss.Left, o.modalSections(inputView, modalWidth)...)
	modal := styleSearchOverlayModal.Width(modalWidth).Render(body)
	return lipgloss.Place(
		containerWidth,
		containerHeight,
		lipgloss.Center,
		lipgloss.Position(overlayVerticalTarget),
		modal,
		lipgloss.WithWhitespaceBackground(cScrim),
	)
}

func (o SearchOverlay) modalSections(inputView string, modalWidth int) []string {
	sections := []string{}
	if title := strings.TrimSpace(o.Title); title != "" {
		sections = append(sections, styleSearchOverlayTitle.Render(title))
	}
	sections = append(sections, inputView)
	if suggestions := o.renderSuggestions(modalWidth); suggestions != "" {
		sections = append(sections, suggestions)
	}
	return sections
}

func (o SearchOverlay) renderSuggestions(modalWidth int) string {
	if len(o.Suggestions) == 0 {
		return styleSearchOverlayHint.Render(o.emptyHint)
	}
	lines := make([]string, len(o.Suggestions))
	for i, suggestion := range o.Suggestions {
		lines[i] = styleSearchOverlaySuggestion.Render(fmt.Sprintf("• %s", suggestion))
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (o SearchOverlay) modalWidth(containerWidth int) int {
	if containerWidth <= 0 {
		return overlayMaxWidth
	}
	maxAllowed := containerWidth - 4
	if maxAllowed < overlayMinWidth {
		maxAllowed = containerWidth - 2
	}
	defaultTarget := containerWidth - 10
	width := clampDimension(defaultTarget, overlayMinWidth, overlayMaxWidth)
	if maxAllowed > 0 && width > maxAllowed {
		width = clampDimension(maxAllowed, overlayMinWidth, overlayMaxWidth)
	}
	return width
}
