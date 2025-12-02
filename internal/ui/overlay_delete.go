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
		case key.Matches(msg, key.NewBinding(key.WithKeys("y", "d"))):
			// Direct confirm with 'y' or 'd' (Delete)
			return m, m.confirm()
		case key.Matches(msg, key.NewBinding(key.WithKeys("n", "c"))):
			// Direct cancel with 'n' or 'c' (Cancel)
			return m, func() tea.Msg { return DeleteCancelledMsg{} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return m, func() tea.Msg { return DeleteCancelledMsg{} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if m.selected == 1 {
				return m, m.confirm()
			}
			// Enter on "Cancel" cancels
			return m, func() tea.Msg { return DeleteCancelledMsg{} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down", "l", "right", "tab"))):
			m.selected = 1 // Move to Delete
		case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up", "h", "left", "shift+tab"))):
			m.selected = 0 // Move to Cancel
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
	var b strings.Builder

	// Title
	b.WriteString(styleErrorIndicator.Render("Delete Bead"))
	b.WriteString("\n")

	// Divider
	b.WriteString(styleStatusDivider.Render(strings.Repeat("─", 38)))
	b.WriteString("\n\n")

	// Prompt
	b.WriteString("Are you sure you want to delete:\n\n")

	// Bead line using same pattern as tree view: icon + ID + title
	icon := styleIconOpen.Render("●")
	id := styleID.Render(m.issueID)
	title := m.issueTitle
	if len(title) > 25 {
		title = title[:22] + "..."
	}
	b.WriteString(icon + " " + id + "  " + styleNormalText.Render(title))
	b.WriteString("\n\n")

	// Warning
	b.WriteString(styleStatsDim.Render("This action cannot be undone."))
	b.WriteString("\n\n")

	// Buttons with underlined hotkeys (C and D)
	var cancelBtn, deleteBtn string
	if m.selected == 0 {
		// Cancel selected
		cancelBtn = styleStatusSelected.Render("[ ") +
			styleStatusSelected.Underline(true).Render("C") +
			styleStatusSelected.Render("ancel ]")
		deleteBtn = styleStatsDim.Render("[ ") +
			styleStatsDim.Underline(true).Render("D") +
			styleStatsDim.Render("elete ]")
	} else {
		// Delete selected
		cancelBtn = styleStatsDim.Render("[ ") +
			styleStatsDim.Underline(true).Render("C") +
			styleStatsDim.Render("ancel ]")
		deleteBtn = styleErrorIndicator.Render("[ ") +
			styleErrorIndicator.Underline(true).Render("D") +
			styleErrorIndicator.Render("elete ]")
	}
	b.WriteString("        " + cancelBtn + "  " + deleteBtn)

	// Use lipgloss border style (same as other overlays)
	return styleDeleteOverlay.Render(b.String())
}
