package ui

import (
	"fmt"

	"abacus/internal/ui/theme"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DeleteOverlay is a confirmation modal for deleting a bead.
// Uses the unified overlay framework for consistent sizing and layout.
type DeleteOverlay struct {
	issueID       string
	issueTitle    string
	children      []ChildInfo
	descendantIDs []string
}

// DeleteConfirmedMsg is sent when deletion is confirmed.
type DeleteConfirmedMsg struct {
	IssueID  string
	Cascade  bool
	Children []string
}

// DeleteCancelledMsg is sent when the overlay is dismissed without deletion.
type DeleteCancelledMsg struct{}

// NewDeleteOverlay creates a new delete confirmation overlay.
func NewDeleteOverlay(issueID, issueTitle string, children []ChildInfo, descendantIDs []string) *DeleteOverlay {
	return &DeleteOverlay{
		issueID:       issueID,
		issueTitle:    issueTitle,
		children:      children,
		descendantIDs: descendantIDs,
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
		case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
			return m, m.confirm()
		case key.Matches(msg, key.NewBinding(key.WithKeys("c", "esc"))):
			return m, func() tea.Msg { return DeleteCancelledMsg{} }
		}
	}
	return m, nil
}

func (m *DeleteOverlay) confirm() tea.Cmd {
	cascade := len(m.children) > 0
	children := m.descendantIDs
	if !cascade {
		children = nil
	}
	return func() tea.Msg {
		return DeleteConfirmedMsg{IssueID: m.issueID, Cascade: cascade, Children: children}
	}
}

// View implements tea.Model using the unified overlay framework.
func (m *DeleteOverlay) View() string {
	b := NewOverlayBuilder(OverlaySizeStandard, 0)

	overlayBg := theme.Current().BackgroundSecondary()
	contentWidth := b.ContentWidth()

	// Title
	titleStyle := lipgloss.NewStyle().
		Background(overlayBg).
		Foreground(theme.Current().Error()).
		Bold(true)
	b.Line(titleStyle.Render("Delete"))
	b.Line(b.Divider())
	b.BlankLine()

	// Danger block
	b.Lines(m.renderDangerBlock(overlayBg, contentWidth)...)

	// Footer
	b.BlankLine()
	b.Footer(m.footerHints())

	return b.BuildDanger()
}

// footerHints returns the footer hints for the delete overlay.
func (m *DeleteOverlay) footerHints() []footerHint {
	deleteLabel := m.deleteLabel()
	return []footerHint{
		{"d", deleteLabel},
		{"c/esc", "Cancel"},
	}
}

func (m *DeleteOverlay) deleteLabel() string {
	if len(m.children) == 0 {
		return "Delete"
	}
	return fmt.Sprintf("Delete All (%d)", len(m.children)+1)
}

func (m *DeleteOverlay) renderDangerBlock(overlayBg lipgloss.AdaptiveColor, _ int) []string {
	var lines []string

	body := lipgloss.NewStyle().
		Background(overlayBg).
		Foreground(currentThemeWrapper().Text())
	warning := lipgloss.NewStyle().
		Background(overlayBg).
		Foreground(theme.Current().Warning())
	dangerIcon := lipgloss.NewStyle().Foreground(theme.Current().Error()).Bold(true).Render("✖")
	warningIcon := lipgloss.NewStyle().Foreground(theme.Current().Warning()).Bold(true).Render("⚠")

	lines = append(lines, dangerIcon+" "+body.Bold(true).Render("Delete this bead?"))
	lines = append(lines, "")

	beadTitle := truncateTitle(m.issueTitle, 38)
	beadLine := "  " + body.Render("● ") + styleID().Background(overlayBg).Render(m.issueID) + body.Render("  "+beadTitle)
	lines = append(lines, beadLine)
	lines = append(lines, "")
	lines = append(lines, warning.Render("This action cannot be undone."))

	if len(m.children) > 0 {
		summary := fmt.Sprintf("This will also delete %d %s:", len(m.children), childWord(len(m.children)))
		lines = append(lines, "", warningIcon+" "+body.Render(summary))
		lines = append(lines, m.renderChildLines(overlayBg)...)
	}

	return lines
}

func (m *DeleteOverlay) renderChildLines(overlayBg lipgloss.AdaptiveColor) []string {
	var lines []string
	idStyle := styleID().Background(overlayBg)
	textStyle := lipgloss.NewStyle().
		Background(overlayBg).
		Foreground(currentThemeWrapper().TextMuted())
	for _, child := range m.children {
		indent := ""
		for i := 0; i < child.Depth; i++ {
			indent += "  "
		}
		prefix := indent + "└─ "
		entry := prefix + idStyle.Render(child.ID) + textStyle.Render("  "+truncateTitle(child.Title, 32))
		lines = append(lines, entry)
	}
	return lines
}

// Layer returns a centered layer for the delete overlay.
// Uses the shared BaseOverlayLayer to eliminate boilerplate.
func (m *DeleteOverlay) Layer(width, height, topMargin, bottomMargin int) Layer {
	return BaseOverlayLayer(m.View, width, height, topMargin, bottomMargin)
}
