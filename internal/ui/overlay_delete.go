package ui

import (
	"fmt"
	"strings"

	"abacus/internal/ui/theme"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DeleteOverlay is a confirmation modal for deleting a bead.
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

// View implements tea.Model.
func (m *DeleteOverlay) View() string {
	return styleStatusOverlay().
		BorderForeground(theme.Current().Error()).
		Render(strings.Join(m.renderLines(), "\n"))
}

// Layer returns a centered layer for the delete overlay.
func (m *DeleteOverlay) Layer(width, height, topMargin, bottomMargin int) Layer {
	return LayerFunc(func() *Canvas {
		content := m.View()
		if strings.TrimSpace(content) == "" {
			return nil
		}

		overlayWidth := lipgloss.Width(content)
		if overlayWidth <= 0 {
			return nil
		}
		overlayHeight := lipgloss.Height(content)
		if overlayHeight <= 0 {
			return nil
		}

		surface := NewSecondarySurface(overlayWidth, overlayHeight)
		surface.Draw(0, 0, content)

		x, y := centeredOffsets(width, height, overlayWidth, overlayHeight, topMargin, bottomMargin)
		surface.Canvas.SetOffset(x, y)
		return surface.Canvas
	})
}

func (m *DeleteOverlay) renderLines() []string {
	var lines []string

	overlayBg := currentThemeWrapper().BackgroundSecondary()
	divider := styleStatusDivider().Background(overlayBg).Render(strings.Repeat("─", 44))

	titleStyle := lipgloss.NewStyle().
		Background(overlayBg).
		Foreground(theme.Current().Error()).
		Bold(true)

	lines = append(lines, titleStyle.Render("Delete"))
	lines = append(lines, divider, "")
	lines = append(lines, m.renderDangerBlock(overlayBg)...)

	deleteLabel := m.deleteLabel()
	lines = append(lines, "", divider, m.renderFooter(deleteLabel))

	return lines
}

func (m *DeleteOverlay) deleteLabel() string {
	if len(m.children) == 0 {
		return "Delete"
	}
	return fmt.Sprintf("Delete All (%d)", len(m.children)+1)
}

func (m *DeleteOverlay) renderDangerBlock(overlayBg lipgloss.AdaptiveColor) []string {
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

func (m *DeleteOverlay) renderFooter(deleteLabel string) string {
	hints := []footerHint{
		{"d", deleteLabel},
		{"c/esc", "Cancel"},
	}

	return overlayFooterLine(hints, 44)
}

func (m *DeleteOverlay) renderChildLines(overlayBg lipgloss.AdaptiveColor) []string {
	var lines []string
	idStyle := styleID().Background(overlayBg)
	textStyle := lipgloss.NewStyle().
		Background(overlayBg).
		Foreground(currentThemeWrapper().TextMuted())
	for _, child := range m.children {
		indent := strings.Repeat("  ", child.Depth)
		prefix := indent + "└─ "
		entry := prefix + idStyle.Render(child.ID) + textStyle.Render("  "+truncateTitle(child.Title, 32))
		lines = append(lines, entry)
	}
	return lines
}
