package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	suggestionListMaxItems  = 7
	suggestionListMaxHeight = suggestionListMaxItems
)

type suggestionList struct {
	model *list.Model
}

func (s *suggestionList) SetItems(items []list.Item) {
	if s == nil || s.model == nil {
		return
	}
	_ = s.model.SetItems(items)
	s.model.SetHeight(clampDimension(len(items), 1, suggestionListMaxHeight))
	s.model.ResetSelected()
}

func (s *suggestionList) ItemCount() int {
	if s == nil || s.model == nil {
		return 0
	}
	return len(s.model.Items())
}

func (s *suggestionList) ViewWithWidth(width int) string {
	if s == nil || s.model == nil {
		return ""
	}
	if width < 1 {
		width = 1
	}
	s.model.SetWidth(width)
	return s.model.View()
}

func (s *suggestionList) CursorUp() {
	if s == nil || s.model == nil {
		return
	}
	s.model.CursorUp()
}

func (s *suggestionList) CursorDown() {
	if s == nil || s.model == nil {
		return
	}
	s.model.CursorDown()
}

func (s *suggestionList) SelectedText() string {
	if s == nil || s.model == nil {
		return ""
	}
	item, ok := s.model.SelectedItem().(textSuggestion)
	if !ok {
		return ""
	}
	return item.Text
}

type textSuggestion struct {
	Text string
}

func (t textSuggestion) FilterValue() string {
	return t.Text
}

type suggestionDelegate struct{}

func newSuggestionDelegate() list.ItemDelegate {
	return suggestionDelegate{}
}

func (d suggestionDelegate) Height() int { return 1 }

func (d suggestionDelegate) Spacing() int { return 0 }

func (d suggestionDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d suggestionDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	suggestion, _ := item.(textSuggestion)
	content := suggestion.Text
	row := styleSuggestionBullet.Render("• ") + styleSuggestionItem.Render(content)
	if index == m.Index() {
		row = styleSuggestionItemSelected.Render(row)
	}
	_, _ = fmt.Fprintln(w, row)
}
