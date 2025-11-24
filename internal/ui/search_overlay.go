package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
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
	Title     string
	emptyHint string

	tokens       []SearchToken
	mode         SuggestionMode
	pendingField string
	pendingText  string

	listModel *suggestionList
}

// NewSearchOverlay builds a modal with sensible defaults.
func NewSearchOverlay() SearchOverlay {
	items := []list.Item{}
	delegate := newSuggestionDelegate()
	model := list.New(items, delegate, overlayMaxWidth, suggestionListMaxHeight)
	model.DisableQuitKeybindings()
	model.SetShowTitle(false)
	model.SetShowStatusBar(false)
	model.SetShowHelp(false)
	model.SetShowPagination(false)
	model.SetShowFilter(false)
	model.SetFilteringEnabled(false)
	model.SetHeight(suggestionListMaxHeight)
	listWrapper := &suggestionList{model: &model}
	m := SearchOverlay{
		Title:     "Smart Filter Search",
		emptyHint: "Type to search by title or bead id",
		listModel: listWrapper,
	}
	m.UpdateInput("")
	return m
}

// SetSuggestions replaces the current suggestion list.
func (o *SearchOverlay) SetSuggestions(items []string) {
	if o == nil {
		return
	}
	converted := make([]list.Item, len(items))
	for i, suggestion := range items {
		converted[i] = textSuggestion{Text: suggestion}
	}
	if o.listModel != nil {
		o.listModel.SetItems(converted)
	}
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
	sections = append(sections, renderOverlayDivider(modalWidth))
	if suggestions := o.renderSuggestions(modalWidth); suggestions != "" {
		sections = append(sections, suggestions)
	}
	return sections
}

func (o SearchOverlay) renderSuggestions(modalWidth int) string {
	if o.listModel == nil || o.listModel.ItemCount() == 0 {
		return styleSearchOverlayHint.Render(o.emptyHint)
	}
	return o.listModel.ViewWithWidth(modalWidth)
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

func renderOverlayDivider(width int) string {
	if width < 1 {
		width = 1
	}
	line := strings.Repeat("─", clampDimension(width-4, 1, width))
	return styleSearchOverlayDivider.Width(width).Render(line)
}

func (o SearchOverlay) HasSuggestions() bool {
	return o.listModel != nil && o.listModel.ItemCount() > 0
}

func (o *SearchOverlay) CursorUp() {
	if o == nil || o.listModel == nil {
		return
	}
	o.listModel.CursorUp()
}

func (o *SearchOverlay) CursorDown() {
	if o == nil || o.listModel == nil {
		return
	}
	o.listModel.CursorDown()
}

func (o SearchOverlay) SelectedSuggestion() string {
	if o.listModel == nil {
		return ""
	}
	return o.listModel.SelectedText()
}
