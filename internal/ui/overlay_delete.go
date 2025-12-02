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

	// Title: "Delete Bead?" centered
	title := styleErrorIndicator.Render("Delete Bead?")
	titlePadding := (maxWidth - 12) / 2 // "Delete Bead?" is 12 chars
	b.WriteString(strings.Repeat(" ", titlePadding))
	b.WriteString(title)
	b.WriteString("\n")

	// Divider
	divider := styleStatusDivider.Render(strings.Repeat("─", maxWidth))
	b.WriteString(divider)
	b.WriteString("\n")

	// Bead info: icon + ID + title (like tree view)
	icon := "○" // Open status icon
	beadLine := styleIconOpen.Render(icon) + " " + styleID.Render(m.issueID) + "  " + styleNormalText.Render(m.issueTitle)

	// Wrap if too long
	if len(m.issueID)+len(m.issueTitle)+4 > maxWidth {
		// Show ID on first line, wrapped title below
		b.WriteString(styleIconOpen.Render(icon) + " " + styleID.Render(m.issueID))
		b.WriteString("\n")
		wrappedTitle := wrapText(m.issueTitle, maxWidth-2)
		// Indent continuation lines
		for i, line := range strings.Split(wrappedTitle, "\n") {
			if i == 0 {
				b.WriteString("  " + styleNormalText.Render(line))
			} else {
				b.WriteString("\n  " + styleNormalText.Render(line))
			}
		}
	} else {
		b.WriteString(beadLine)
	}
	b.WriteString("\n\n")

	// Warning message
	warning := styleStatsDim.Render("This cannot be undone.")
	b.WriteString(warning)
	b.WriteString("\n\n")

	// Centered options with underlined hotkey letters (N and Y)
	// Use lipgloss Underline for the hotkey letter

	var noLabel, yesLabel string
	if m.selected == 0 {
		// No selected - fully highlighted, N underlined
		noLabel = styleStatusSelected.Copy().Underline(true).Render("N") + styleStatusSelected.Render("o")
		yesLabel = styleStatsDim.Copy().Underline(true).Render("Y") + styleStatsDim.Render("es")
	} else {
		// Yes selected - highlighted in red, Y underlined
		noLabel = styleStatsDim.Copy().Underline(true).Render("N") + styleStatsDim.Render("o")
		yesLabel = styleErrorIndicator.Copy().Underline(true).Render("Y") + styleErrorIndicator.Render("es")
	}

	// Center the options (No=2 chars, gap=6 spaces, Yes=3 chars)
	optionsWidth := 2 + 6 + 3
	optionsPadding := (maxWidth - optionsWidth) / 2
	b.WriteString(strings.Repeat(" ", optionsPadding))
	b.WriteString(noLabel + "      " + yesLabel)

	content := b.String()
	return styleStatusOverlay.Render(content)
}
