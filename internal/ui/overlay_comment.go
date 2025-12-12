package ui

import (
	"fmt"
	"runtime"
	"strings"

	"abacus/internal/ui/theme"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	commentModalWidth    = 54 // Base content width (without padding)
	commentTextareaLines = 6  // Visible lines
	commentCharLimit     = 2000
	commentTextareaPad   = 1 // Space between border and text area content
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
	ta.Prompt = "" // Remove default left prompt to align background with input text
	ta.CharLimit = commentCharLimit
	ta.SetWidth(commentModalWidth + 6) // Temporary default; updated in View based on padding
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
		switch msg.Type {
		case tea.KeyEsc:
			// Multi-stage escape: clear text first, then cancel
			if strings.TrimSpace(m.textarea.Value()) != "" {
				m.textarea.SetValue("")
				m.errorMsg = ""
				return m, nil
			}
			return m, func() tea.Msg { return CommentCancelledMsg{} }

		case tea.KeyCtrlS:
			// Ctrl+S to submit/save
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

	containerWidth := commentModalWidth + 11
	taContentWidth := containerWidth - (commentTextareaPad * 2)
	m.textarea.SetWidth(taContentWidth)

	leftPad := lipgloss.NewStyle().
		Background(theme.Current().BackgroundSecondary()).
		Render(strings.Repeat(" ", commentTextareaPad))
	rightPad := leftPad
	taViewLines := strings.Split(m.textarea.View(), "\n")
	for i, line := range taViewLines {
		if line == "" && i == len(taViewLines)-1 {
			continue
		}
		taViewLines[i] = leftPad + line + rightPad
	}
	taView := strings.Join(taViewLines, "\n")

	// Header
	header := styleHelpTitle().Render("ADD COMMENT")
	divider := styleHelpDivider().Render(strings.Repeat("─", containerWidth))

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
	taStyle := styleCommentTextarea(containerWidth)
	b.WriteString(taStyle.Render(taView))
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

	// Footer - show ⌘S on Mac, ^S elsewhere
	saveKey := "^S"
	if runtime.GOOS == "darwin" {
		saveKey = "⌘S"
	}
	hints := []footerHint{
		{saveKey, "Save"},
		{"esc", "Cancel"},
	}
	b.WriteString(overlayFooterLine(hints, containerWidth))

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
