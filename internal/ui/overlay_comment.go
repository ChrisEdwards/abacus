package ui

import (
	"fmt"
	"strings"

	"abacus/internal/ui/theme"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	commentModalWidth    = 54 // Content width
	commentTextareaLines = 6  // Visible lines
	commentCharLimit     = 2000
)

// CommentAddedMsg is sent when the user submits a comment.
type CommentAddedMsg struct {
	IssueID string
	Comment string
}

// CommentCancelledMsg is sent when the user cancels the comment modal.
type CommentCancelledMsg struct{}

// CommentOverlay manages the comment modal state.
type CommentOverlay struct {
	issueID   string
	beadTitle string
	textarea  textarea.Model
	errorMsg  string
	termWidth int
}

// NewCommentOverlay creates a new comment modal overlay.
func NewCommentOverlay(issueID, beadTitle string) *CommentOverlay {
	ta := textarea.New()
	ta.Placeholder = "Type your comment here..."
	ta.CharLimit = commentCharLimit
	ta.SetWidth(commentModalWidth)
	ta.SetHeight(commentTextareaLines)
	ta.ShowLineNumbers = false
	ta.Focus()

	return &CommentOverlay{
		issueID:   issueID,
		beadTitle: beadTitle,
		textarea:  ta,
	}
}

// Init returns the initial command for the comment overlay.
func (m *CommentOverlay) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages for the comment overlay.
func (m *CommentOverlay) Update(msg tea.Msg) (*CommentOverlay, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case msg.Type == tea.KeyEsc:
			// Multi-stage escape: clear text first, then cancel
			if strings.TrimSpace(m.textarea.Value()) != "" {
				m.textarea.SetValue("")
				m.errorMsg = ""
				return m, nil
			}
			return m, func() tea.Msg { return CommentCancelledMsg{} }

		case msg.Type == tea.KeyCtrlJ, msg.String() == "ctrl+enter":
			// Ctrl+Enter to submit (KeyCtrlJ is what bubbletea sends for Ctrl+Enter)
			return m.submit()
		}
	}

	// Pass to textarea
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

// submit validates and submits the comment.
func (m *CommentOverlay) submit() (*CommentOverlay, tea.Cmd) {
	text := strings.TrimSpace(m.textarea.Value())
	if text == "" {
		m.errorMsg = "Comment cannot be empty"
		return m, nil
	}
	m.errorMsg = ""
	return m, func() tea.Msg {
		return CommentAddedMsg{
			IssueID: m.issueID,
			Comment: text,
		}
	}
}

// View renders the comment overlay.
func (m *CommentOverlay) View() string {
	var b strings.Builder

	// Header
	header := styleHelpTitle().Render("ADD COMMENT")
	divider := styleHelpDivider().Render(strings.Repeat("─", commentModalWidth+4))

	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(divider)
	b.WriteString("\n\n")

	// Bead context line using common ID + title renderer
	// styleID() renders the bead ID in gold/bold
	// Truncate title if too long
	title := m.beadTitle
	if len(title) > 40 {
		title = title[:37] + "..."
	}
	contextLine := styleID().Render(m.issueID) + " " + title
	b.WriteString(contextLine)
	b.WriteString("\n\n")

	// Textarea with border
	taStyle := styleCommentTextarea(commentModalWidth + 4)
	b.WriteString(taStyle.Render(m.textarea.View()))
	b.WriteString("\n")

	// Character count
	count := len(m.textarea.Value())
	countStyle := lipgloss.NewStyle().
		Background(theme.Current().BackgroundSecondary()).
		Foreground(theme.Current().TextMuted())
	if count > commentCharLimit-100 {
		countStyle = lipgloss.NewStyle().
			Background(theme.Current().BackgroundSecondary()).
			Foreground(theme.Current().Warning())
	}
	b.WriteString(countStyle.Render(fmt.Sprintf("  %d/%d", count, commentCharLimit)))
	b.WriteString("\n")

	// Error message
	if m.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().
			Background(theme.Current().BackgroundSecondary()).
			Foreground(theme.Current().Error())
		b.WriteString(errorStyle.Render("  ⚠ " + m.errorMsg))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(divider)
	b.WriteString("\n")

	// Footer
	hints := []footerHint{
		{"^⏎", "Submit"},
		{"esc", "Cancel"},
	}
	b.WriteString(overlayFooterLine(hints, commentModalWidth+4))

	return styleHelpOverlay().Render(b.String())
}

// Layer returns a Layer for rendering the comment overlay.
func (m *CommentOverlay) Layer(width, height, topMargin, bottomMargin int) Layer {
	return LayerFunc(func() *Canvas {
		content := m.View()
		if strings.TrimSpace(content) == "" {
			return nil
		}

		overlayWidth := lipgloss.Width(content)
		overlayHeight := lipgloss.Height(content)

		surface := NewSecondarySurface(overlayWidth, overlayHeight)
		surface.Draw(0, 0, content)

		x, y := centeredOffsets(width, height, overlayWidth, overlayHeight, topMargin, bottomMargin)
		surface.Canvas.SetOffset(x, y)
		return surface.Canvas
	})
}

// SetSize updates the terminal dimensions for the overlay.
func (m *CommentOverlay) SetSize(width, _ int) {
	m.termWidth = width
}

// styleCommentTextarea returns the style for the comment textarea container.
func styleCommentTextarea(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().BorderNormal()).
		BorderBackground(theme.Current().BackgroundSecondary()).
		Background(theme.Current().BackgroundSecondary())
}
