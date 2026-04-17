package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type PriorityOverlay struct {
	issueID         string
	issueTitle      string
	currentPriority int
	selected        int
	options         []priorityOption
}

type priorityOption struct {
	value int
	label string
	name  string
}

type PriorityChangedMsg struct {
	IssueID     string
	NewPriority int
}

type PriorityCancelledMsg struct{}

func NewPriorityOverlay(issueID, issueTitle string, currentPriority int) *PriorityOverlay {
	options := []priorityOption{
		{value: 0, label: "P0", name: "Critical"},
		{value: 1, label: "P1", name: "High"},
		{value: 2, label: "P2", name: "Medium"},
		{value: 3, label: "P3", name: "Low"},
		{value: 4, label: "P4", name: "Backlog"},
	}

	selected := 2
	for i, opt := range options {
		if opt.value == currentPriority {
			selected = i
			break
		}
	}

	return &PriorityOverlay{
		issueID:         issueID,
		issueTitle:      issueTitle,
		currentPriority: currentPriority,
		selected:        selected,
		options:         options,
	}
}

func (m *PriorityOverlay) Init() tea.Cmd {
	return nil
}

func (m *PriorityOverlay) Update(msg tea.Msg) (*PriorityOverlay, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
			m.selected = (m.selected + 1) % len(m.options)
		case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
			m.selected = (m.selected - 1 + len(m.options)) % len(m.options)
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			return m, m.confirm()
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return m, func() tea.Msg { return PriorityCancelledMsg{} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("0"))):
			return m, m.selectByValue(0)
		case key.Matches(msg, key.NewBinding(key.WithKeys("1"))):
			return m, m.selectByValue(1)
		case key.Matches(msg, key.NewBinding(key.WithKeys("2"))):
			return m, m.selectByValue(2)
		case key.Matches(msg, key.NewBinding(key.WithKeys("3"))):
			return m, m.selectByValue(3)
		case key.Matches(msg, key.NewBinding(key.WithKeys("4"))):
			return m, m.selectByValue(4)
		}
	}
	return m, nil
}

func (m *PriorityOverlay) selectByValue(value int) tea.Cmd {
	for i, opt := range m.options {
		if opt.value == value {
			m.selected = i
			return m.confirm()
		}
	}
	return nil
}

func (m *PriorityOverlay) confirm() tea.Cmd {
	newPriority := m.options[m.selected].value
	issueID := m.issueID
	return func() tea.Msg {
		return PriorityChangedMsg{
			IssueID:     issueID,
			NewPriority: newPriority,
		}
	}
}

func (m *PriorityOverlay) View() string {
	b := NewOverlayBuilder(OverlaySizeNarrow, 0)

	header := styleID().Render(m.issueID) + styleStatsDim().Render(" › ") + styleStatsDim().Render("Priority")
	b.Line(header)
	b.Line(b.Divider())

	for i, opt := range m.options {
		indicator := "○"
		if opt.value == m.currentPriority {
			indicator = "●"
		}

		label := opt.label + "  " + opt.name
		if i == m.selected {
			label += "  ←"
		}

		var line string
		if i == m.selected {
			line = styleStatusSelected().Render("  " + indicator + " " + label)
		} else {
			line = styleStatusOption().Render("  " + indicator + " " + label)
		}

		b.Line(line)
	}

	return b.Build()
}

func (m *PriorityOverlay) Layer(width, height, topMargin, bottomMargin int) Layer {
	return BaseOverlayLayer(m.View, width, height, topMargin, bottomMargin)
}
