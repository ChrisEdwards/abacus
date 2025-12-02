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
	const innerWidth = 39 // Content width inside borders

	// Helper to pad line to innerWidth
	padLine := func(content string, contentLen int) string {
		padding := innerWidth - contentLen
		if padding < 0 {
			padding = 0
		}
		return content + strings.Repeat(" ", padding)
	}

	var b strings.Builder

	// Top border: ╭───────────────────────────────────────╮
	b.WriteString("╭" + strings.Repeat("─", innerWidth) + "╮\n")

	// Title row: │ Delete Bead                           │
	title := styleErrorIndicator.Render("Delete Bead")
	titleRow := " " + title
	b.WriteString("│" + padLine(titleRow, 12) + "│\n") // "Delete Bead" = 11 chars + 1 space

	// Divider: ├───────────────────────────────────────┤
	b.WriteString("├" + strings.Repeat("─", innerWidth) + "┤\n")

	// Empty line
	b.WriteString("│" + strings.Repeat(" ", innerWidth) + "│\n")

	// "Are you sure you want to delete:" line
	prompt := "  Are you sure you want to delete:"
	b.WriteString("│" + padLine(prompt, len(prompt)) + "│\n")

	// Empty line
	b.WriteString("│" + strings.Repeat(" ", innerWidth) + "│\n")

	// Bead line: ●  ab-fg2  test flag check
	icon := "●"
	beadContent := "  " + icon + " " + m.issueID + "  " + m.issueTitle
	// Truncate if too long
	if len(beadContent) > innerWidth {
		beadContent = beadContent[:innerWidth-3] + "..."
	}
	b.WriteString("│" + padLine(beadContent, len(beadContent)) + "│\n")

	// Empty line
	b.WriteString("│" + strings.Repeat(" ", innerWidth) + "│\n")

	// Warning line
	warning := "  This action cannot be undone."
	b.WriteString("│" + padLine(warning, len(warning)) + "│\n")

	// Empty line
	b.WriteString("│" + strings.Repeat(" ", innerWidth) + "│\n")

	// Buttons: [ Cancel ]  [ Delete ]
	var cancelBtn, deleteBtn string
	if m.selected == 0 {
		// Cancel selected
		cancelBtn = styleStatusSelected.Render("[ Cancel ]")
		deleteBtn = styleStatsDim.Render("[ Delete ]")
	} else {
		// Delete selected
		cancelBtn = styleStatsDim.Render("[ Cancel ]")
		deleteBtn = styleErrorIndicator.Render("[ Delete ]")
	}
	buttons := cancelBtn + "  " + deleteBtn
	// Center buttons (10 + 2 + 10 = 22 chars visible)
	btnPadding := (innerWidth - 22) / 2
	btnLine := strings.Repeat(" ", btnPadding) + buttons
	// Pad to innerWidth (account for ANSI codes by using fixed padding)
	btnLinePadded := btnLine + strings.Repeat(" ", innerWidth-btnPadding-22)
	b.WriteString("│" + btnLinePadded + "│\n")

	// Bottom border: ╰───────────────────────────────────────╯
	b.WriteString("╰" + strings.Repeat("─", innerWidth) + "╯")

	return b.String()
}
