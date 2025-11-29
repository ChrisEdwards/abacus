package ui

import (
	"strings"

	"abacus/internal/domain"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// StatusOverlay is a compact popup for changing a bead's status.
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
	for i := range options {
		target := domain.Status(options[i].value)
		if current.CanTransitionTo(target) != nil && current != target {
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
		}
	}
	return m, nil
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

// View implements tea.Model.
func (m *StatusOverlay) View() string {
	var b strings.Builder

	// Title with bead context
	title := styleStatusTitle.Render("CHANGE STATUS")
	b.WriteString(title)
	b.WriteString("\n")

	// Bead info line
	beadInfo := styleStatusBeadID.Render(m.issueID) + " " + styleStatusBeadTitle.Render(truncateTitle(m.issueTitle, 30))
	b.WriteString(beadInfo)
	b.WriteString("\n")

	// Divider
	divider := styleStatusDivider.Render(strings.Repeat("─", 36))
	b.WriteString(divider)
	b.WriteString("\n")

	// Options
	for i, opt := range m.options {
		var line string
		indicator := "○"
		if opt.value == m.currentStatus {
			indicator = "●"
		}

		label := opt.label
		if i == m.selected {
			label += "  ←"
		}

		if opt.disabled {
			line = styleStatusDisabled.Render("  " + indicator + " " + label)
		} else if i == m.selected {
			line = styleStatusSelected.Render("  " + indicator + " " + label)
		} else {
			line = styleStatusOption.Render("  " + indicator + " " + label)
		}

		b.WriteString(line)
		if i < len(m.options)-1 {
			b.WriteString("\n")
		}
	}

	content := b.String()
	return styleStatusOverlay.Render(content)
}

// truncateTitle shortens a title to maxLen characters with ellipsis.
func truncateTitle(title string, maxLen int) string {
	if len(title) <= maxLen {
		return title
	}
	return title[:maxLen-3] + "..."
}
