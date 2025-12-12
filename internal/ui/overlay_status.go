package ui

import (
	"abacus/internal/domain"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// StatusOverlay is a compact popup for changing a bead's status.
// Uses the unified overlay framework for consistent sizing and layout.
type StatusOverlay struct {
	issueID       string
	issueTitle    string
	currentStatus string
	selected      int
	options       []statusOption
}

type statusOption struct {
	value    string // "open", "in_progress", "closed"
	label    string // Display label
	disabled bool   // True if transition is not allowed
}

// StatusChangedMsg is sent when a status change is confirmed.
type StatusChangedMsg struct {
	IssueID   string
	NewStatus string
}

// StatusCancelledMsg is sent when the overlay is dismissed without changes.
type StatusCancelledMsg struct{}

// NewStatusOverlay creates a new status overlay for the given issue.
func NewStatusOverlay(issueID, issueTitle, currentStatus string) *StatusOverlay {
	current := domain.Status(currentStatus)

	options := []statusOption{
		{value: "open", label: "Open"},
		{value: "in_progress", label: "In Progress"},
		{value: "closed", label: "Closed"},
	}

	// Mark invalid transitions as disabled
	// Special case: allow closed → open (reopen) even though domain model disallows it
	for i := range options {
		target := domain.Status(options[i].value)
		if current.CanTransitionTo(target) != nil && current != target {
			// Allow reopening: closed → open
			if currentStatus == "closed" && options[i].value == "open" {
				continue
			}
			options[i].disabled = true
		}
	}

	// Find the index of the current status
	selected := 0
	for i, opt := range options {
		if opt.value == currentStatus {
			selected = i
			break
		}
	}

	return &StatusOverlay{
		issueID:       issueID,
		issueTitle:    issueTitle,
		currentStatus: currentStatus,
		selected:      selected,
		options:       options,
	}
}

// Init implements tea.Model.
func (m *StatusOverlay) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m *StatusOverlay) Update(msg tea.Msg) (*StatusOverlay, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
			m.moveDown()
		case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
			m.moveUp()
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			return m, m.confirm()
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return m, func() tea.Msg { return StatusCancelledMsg{} }
		// Hotkeys: o=Open, i=In Progress, c=Closed
		case key.Matches(msg, key.NewBinding(key.WithKeys("o"))):
			return m, m.selectByValue("open")
		case key.Matches(msg, key.NewBinding(key.WithKeys("i"))):
			return m, m.selectByValue("in_progress")
		case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
			return m, m.selectByValue("closed")
		}
	}
	return m, nil
}

// selectByValue selects and confirms the option with the given value if enabled.
func (m *StatusOverlay) selectByValue(value string) tea.Cmd {
	for i, opt := range m.options {
		if opt.value == value && !opt.disabled {
			m.selected = i
			return m.confirm()
		}
	}
	return nil
}

func (m *StatusOverlay) moveDown() {
	// Move to next enabled option
	for i := 1; i <= len(m.options); i++ {
		next := (m.selected + i) % len(m.options)
		if !m.options[next].disabled {
			m.selected = next
			return
		}
	}
}

func (m *StatusOverlay) moveUp() {
	// Move to previous enabled option
	for i := 1; i <= len(m.options); i++ {
		prev := (m.selected - i + len(m.options)) % len(m.options)
		if !m.options[prev].disabled {
			m.selected = prev
			return
		}
	}
}

func (m *StatusOverlay) confirm() tea.Cmd {
	selectedOpt := m.options[m.selected]
	if selectedOpt.disabled {
		return nil
	}
	return func() tea.Msg {
		return StatusChangedMsg{
			IssueID:   m.issueID,
			NewStatus: selectedOpt.value,
		}
	}
}

// View implements tea.Model using the unified overlay framework.
func (m *StatusOverlay) View() string {
	b := NewOverlayBuilder(OverlaySizeNarrow, 0)

	// Header with context
	header := styleID().Render(m.issueID) + styleStatsDim().Render(" › ") + styleStatsDim().Render("Status")
	b.Line(header)
	b.Line(b.Divider())

	// Status options
	for i, opt := range m.options {
		indicator := "○"
		if opt.value == m.currentStatus {
			indicator = "●"
		}

		label := opt.label
		if i == m.selected {
			label += "  ←"
		}

		var line string
		if opt.disabled {
			line = styleStatusDisabled().Render("  " + indicator + " " + label)
		} else if i == m.selected {
			line = styleStatusSelected().Render("  " + indicator + " " + label)
		} else {
			line = styleStatusOption().Render("  " + indicator + " " + label)
		}

		b.Line(line)
	}

	return b.Build()
}

// Layer returns a centered layer for the status overlay.
// Uses the shared BaseOverlayLayer to eliminate boilerplate.
func (m *StatusOverlay) Layer(width, height, topMargin, bottomMargin int) Layer {
	return BaseOverlayLayer(m.View, width, height, topMargin, bottomMargin)
}

// truncateTitle shortens a title to maxLen characters with ellipsis.
func truncateTitle(title string, maxLen int) string {
	if len(title) <= maxLen {
		return title
	}
	return title[:maxLen-3] + "..."
}
