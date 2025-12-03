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

	// Title (error color with secondary background)
	b.WriteString(styleDeleteTitle().Render("Delete Bead"))
	b.WriteString("\n")

	// Divider (red to match border, with secondary background)
	b.WriteString(styleDeleteTitle().Render(strings.Repeat("─", 38)))
	b.WriteString("\n\n")

	// Prompt (with secondary background)
	b.WriteString(styleOverlayText().Render("Are you sure you want to delete:"))
	b.WriteString("\n\n")

	// Bead line using same pattern as tree view: icon + ID + title (all with secondary bg)
	icon := styleOverlayIcon().Render("●")
	id := styleOverlayID().Render(m.issueID)
	title := m.issueTitle
	if len(title) > 25 {
		title = title[:22] + "..."
	}
	b.WriteString(icon + " " + id + "  " + styleOverlayText().Render(title))
	b.WriteString("\n\n")

	// Warning (muted with secondary background)
	b.WriteString(styleOverlayTextMuted().Render("This action cannot be undone."))
	b.WriteString("\n\n")

	// Buttons with underlined hotkeys (C and D) - all with secondary background
	var cancelBtn, deleteBtn string
	if m.selected == 0 {
		// Cancel selected (primary highlight)
		cancelBtn = styleOverlayButtonSelected().Render("[ ") +
			styleOverlayButtonSelected().Underline(true).Render("C") +
			styleOverlayButtonSelected().Render("ancel ]")
		deleteBtn = styleOverlayTextMuted().Render("[ ") +
			styleOverlayTextMuted().Underline(true).Render("D") +
			styleOverlayTextMuted().Render("elete ]")
	} else {
		// Delete selected (danger/error highlight)
		cancelBtn = styleOverlayTextMuted().Render("[ ") +
			styleOverlayTextMuted().Underline(true).Render("C") +
			styleOverlayTextMuted().Render("ancel ]")
		deleteBtn = styleOverlayButtonDanger().Render("[ ") +
			styleOverlayButtonDanger().Underline(true).Render("D") +
			styleOverlayButtonDanger().Render("elete ]")
	}
	b.WriteString(styleOverlayText().Render("        ") + cancelBtn + styleOverlayText().Render("  ") + deleteBtn)

	// Use lipgloss border style (same as other overlays)
	return styleDeleteOverlay().Render(b.String())
}
