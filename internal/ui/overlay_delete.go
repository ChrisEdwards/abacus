package ui

import (
	"fmt"
	"strings"

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
	selected      int // 0=Cancel, 1=Delete
	cascadeMode   bool
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
		selected:      0,
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
			m.setSelected(1)
			return m, m.confirm()
		case key.Matches(msg, key.NewBinding(key.WithKeys("n", "c"))):
			m.setSelected(0)
			return m, func() tea.Msg { return DeleteCancelledMsg{} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			m.setSelected(0)
			return m, func() tea.Msg { return DeleteCancelledMsg{} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if m.selected == 1 {
				return m, m.confirm()
			}
			return m, func() tea.Msg { return DeleteCancelledMsg{} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down", "l", "right", "tab"))):
			m.setSelected(1)
		case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up", "h", "left", "shift+tab"))):
			m.setSelected(0)
		}
	}
	return m, nil
}

func (m *DeleteOverlay) confirm() tea.Cmd {
	cascade := m.cascadeMode && len(m.children) > 0
	children := m.descendantIDs
	if !cascade {
		children = nil
	}
	return func() tea.Msg {
		return DeleteConfirmedMsg{IssueID: m.issueID, Cascade: cascade, Children: children}
	}
}

func (m *DeleteOverlay) setSelected(value int) {
	if value != 0 && value != 1 {
		return
	}
	m.selected = value
	if m.selected == 1 && len(m.children) > 0 {
		m.cascadeMode = true
	} else {
		m.cascadeMode = false
	}
}

// View implements tea.Model.
func (m *DeleteOverlay) View() string {
	return styleDeleteOverlay().Render(strings.Join(m.renderLines(), "\n"))
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

	lines = append(lines, styleDeleteTitle().Render("Delete Bead"))
	lines = append(lines, styleDeleteTitle().Render(strings.Repeat("─", 38)), "")

	if len(m.children) == 0 {
		lines = append(lines, styleOverlayText().Render("Are you sure you want to delete:"), "")
		lines = append(lines, m.renderBeadLine("●", m.issueID, m.issueTitle), "")
		lines = append(lines, styleOverlayTextMuted().Render("This action cannot be undone."), "")
	} else {
		warning := fmt.Sprintf("⚠ This bead has %d %s that will also be deleted:", len(m.children), childWord(len(m.children)))
		lines = append(lines, styleOverlayText().Render(warning), "")
		lines = append(lines, m.renderBeadLine("●", m.issueID, m.issueTitle))
		childLines := m.renderChildLines()
		if len(childLines) > 0 {
			lines = append(lines, childLines...)
		}
		lines = append(lines, "")
		lines = append(lines, styleOverlayTextMuted().Render("This action cannot be undone."), "")
	}

	deleteLabel := "Delete"
	if len(m.children) > 0 {
		deleteLabel = fmt.Sprintf("Delete All (%d)", len(m.children)+1)
	}

	cancelBtn := styleOverlayTextMuted().Render(fmt.Sprintf("[ %s ]", cancelLabel))
	deleteBtn := styleOverlayTextMuted().Render(fmt.Sprintf("[ %s ]", deleteLabel))

	if m.selected == 0 {
		cancelBtn = styleOverlayButtonSelected().Render(fmt.Sprintf("[ %s ]", cancelLabel))
	} else {
		deleteBtn = styleOverlayButtonDanger().Render(fmt.Sprintf("[ %s ]", deleteLabel))
	}

	lines = append(lines, styleOverlayText().Render("        ")+cancelBtn+styleOverlayText().Render("  ")+deleteBtn)

	return lines
}

func (m *DeleteOverlay) renderBeadLine(icon, id, title string) string {
	trimmed := title
	if len(trimmed) > 32 {
		trimmed = trimmed[:29] + "..."
	}
	return styleOverlayIcon().Render(icon) + " " + styleOverlayID().Render(id) + "  " + styleOverlayText().Render(trimmed)
}

func (m *DeleteOverlay) renderChildLines() []string {
	var lines []string
	for _, child := range m.children {
		indent := strings.Repeat("  ", child.Depth)
		prefix := indent + "└─ "
		entry := prefix + styleOverlayID().Render(child.ID) + "  " + styleOverlayText().Render(truncateTitle(child.Title, 24))
		lines = append(lines, entry)
	}
	return lines
}

const cancelLabel = "Cancel"
