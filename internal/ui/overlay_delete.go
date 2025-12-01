package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// DeleteOverlay is a confirmation modal for deleting a bead.
type DeleteOverlay struct {
	issueID    string
	issueTitle string
	selected   int // 0=No (default), 1=Yes
}

// DeleteConfirmedMsg is sent when deletion is confirmed.
type DeleteConfirmedMsg struct {
	IssueID string
}

// DeleteCancelledMsg is sent when the overlay is dismissed without deletion.
type DeleteCancelledMsg struct{}

// NewDeleteOverlay creates a new delete confirmation overlay.
func NewDeleteOverlay(issueID, issueTitle string) *DeleteOverlay {
	return &DeleteOverlay{
		issueID:    issueID,
		issueTitle: issueTitle,
		selected:   0, // Default to "No" for safety
	}
}

// Init implements tea.Model.
func (m *DeleteOverlay) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m *DeleteOverlay) Update(msg tea.Msg) (*DeleteOverlay, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("y"))):
			// Direct confirm with 'y'
			return m, m.confirm()
		case key.Matches(msg, key.NewBinding(key.WithKeys("n"))):
			// Direct cancel with 'n'
			return m, func() tea.Msg { return DeleteCancelledMsg{} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return m, func() tea.Msg { return DeleteCancelledMsg{} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if m.selected == 1 {
				return m, m.confirm()
			}
			// Enter on "No" cancels
			return m, func() tea.Msg { return DeleteCancelledMsg{} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down", "l", "right", "tab"))):
			m.selected = 1 // Move to Yes
		case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up", "h", "left", "shift+tab"))):
			m.selected = 0 // Move to No
		}
	}
	return m, nil
}

func (m *DeleteOverlay) confirm() tea.Cmd {
	return func() tea.Msg {
		return DeleteConfirmedMsg{IssueID: m.issueID}
	}
}

// View implements tea.Model.
func (m *DeleteOverlay) View() string {
	const maxWidth = 50

	var b strings.Builder

	// Header: ab-xxx › Delete
	header := styleID.Render(m.issueID) + styleStatsDim.Render(" › ") + styleErrorIndicator.Render("Delete")
	b.WriteString(header)
	b.WriteString("\n")

	// Divider - match content width
	dividerWidth := maxWidth
	divider := styleStatusDivider.Render(strings.Repeat("─", dividerWidth))
	b.WriteString(divider)
	b.WriteString("\n")

	// Title with word wrapping
	wrappedTitle := wrapText(m.issueTitle, maxWidth)
	b.WriteString(styleNormalText.Render(wrappedTitle))
	b.WriteString("\n\n")

	// Warning message
	warning := styleStatsDim.Render("This cannot be undone.")
	b.WriteString(warning)
	b.WriteString("\n\n")

	// Options: [n]o  [y]es
	noLabel := "[n]o"
	yesLabel := "[y]es"

	if m.selected == 0 {
		noLabel = styleStatusSelected.Render(noLabel + " ←")
		yesLabel = styleStatusOption.Render(yesLabel)
	} else {
		noLabel = styleStatusOption.Render(noLabel)
		yesLabel = styleErrorIndicator.Render(yesLabel + " ←")
	}

	b.WriteString("  ")
	b.WriteString(noLabel)
	b.WriteString("    ")
	b.WriteString(yesLabel)

	content := b.String()
	return styleStatusOverlay.Render(content)
}
