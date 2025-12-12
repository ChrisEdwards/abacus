package ui

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	commentTextareaLines = 6 // Visible lines
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
// Uses the unified overlay framework for consistent sizing and layout.
type CommentOverlay struct {
	issueID   string
	beadTitle string
	textarea  textarea.Model
	errorMsg  string
	termWidth int
}

// NewCommentOverlay creates a new comment modal overlay.
func NewCommentOverlay(issueID, beadTitle string) *CommentOverlay {
	// Use OverlayWidthWide for comment overlays (textareas need more space)
	boxWidth := OverlayWidthWide
	taWidth := OverlayTextareaWidth(boxWidth)

	taModel := NewBaseTextarea(taWidth, commentTextareaLines)
	taModel.Placeholder = "Type your comment here..."
	taModel.CharLimit = commentCharLimit

	taModel.Focus()

	return &CommentOverlay{
		issueID:   issueID,
		beadTitle: beadTitle,
		textarea:  taModel,
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

// View renders the comment overlay using the unified overlay framework.
func (m *CommentOverlay) View() string {
	b := NewOverlayBuilder(OverlaySizeWide, 0)
	contentWidth := b.ContentWidth()

	// Update textarea width to match builder
	taContentWidth := TextareaContentWidth(contentWidth+4, commentTextareaPad)
	m.textarea.SetWidth(taContentWidth)

	// Header
	b.Header("ADD COMMENT")

	// Bead context line
	title := m.beadTitle
	if len(title) > 40 {
		title = title[:37] + "..."
	}
	contextLine := styleID().Render(m.issueID) + " " + title
	b.Line(contextLine)
	b.BlankLine()

	// Textarea with border
	taView := PadTextareaView(m.textarea.View(), commentTextareaPad)
	taStyle := styleCommentTextarea(contentWidth + 4)
	b.Line(taStyle.Render(taView))

	// Character count
	count := len(m.textarea.Value())
	countStyle := lipgloss.NewStyle().
		Background(currentThemeWrapper().BackgroundSecondary()).
		Foreground(currentThemeWrapper().TextMuted())
	if count > commentCharLimit-100 {
		countStyle = lipgloss.NewStyle().
			Background(currentThemeWrapper().BackgroundSecondary()).
			Foreground(currentThemeWrapper().Warning())
	}
	b.Line(countStyle.Render(fmt.Sprintf("  %d/%d", count, commentCharLimit)))

	// Error message
	if m.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().
			Background(currentThemeWrapper().BackgroundSecondary()).
			Foreground(currentThemeWrapper().Error())
		b.Line(errorStyle.Render("  ⚠ " + m.errorMsg))
	}

	b.BlankLine()

	// Footer - show ⌘S on Mac, ^S elsewhere
	saveKey := "^S"
	if runtime.GOOS == "darwin" {
		saveKey = "⌘S"
	}
	hints := []footerHint{
		{saveKey, "Save"},
		{"esc", "Cancel"},
	}
	b.Footer(hints)

	return b.Build()
}

// Layer returns a Layer for rendering the comment overlay.
// Uses the shared BaseOverlayLayer to eliminate boilerplate.
func (m *CommentOverlay) Layer(width, height, topMargin, bottomMargin int) Layer {
	return BaseOverlayLayer(m.View, width, height, topMargin, bottomMargin)
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
		BorderForeground(currentThemeWrapper().BorderNormal()).
		BorderBackground(currentThemeWrapper().BackgroundSecondary()).
		Background(currentThemeWrapper().BackgroundSecondary())
}
