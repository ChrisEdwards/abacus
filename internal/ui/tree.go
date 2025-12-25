package ui

import (
	"fmt"
	"strings"

	"abacus/internal/config"
	"abacus/internal/domain"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

const treeScrollMargin = 1

// renderTreeView renders the tree list. Theme is managed by the caller (view.go)
// which sets dimmed theme when an overlay is active.
func (m *App) renderTreeView() string {
	listHeight := clampDimension(m.height-4, minListHeight, m.height-2)
	if len(m.visibleRows) == 0 {
		m.treeTopLine = 0
		// Show empty state message with hint to create first bead
		emptyMsg := styleStatsDim().Render("No beads yet. Press ") +
			styleID().Render("n") +
			styleStatsDim().Render(" to add your first bead.")
		return emptyMsg
	}

	totalWidth := m.width - 2
	if m.ShowDetails {
		totalWidth = m.width - m.viewport.Width - 4
	}
	totalWidth = clampDimension(totalWidth, minTreeWidth, m.width-2)

	lines, cursorStart, cursorEnd := m.buildTreeLines(totalWidth)
	totalLines := len(lines)
	if totalLines == 0 {
		return ""
	}

	m.ensureTreeSelectionVisible(listHeight, totalLines, cursorStart, cursorEnd)

	maxTop := totalLines - listHeight
	if maxTop < 0 {
		maxTop = 0
	}
	start := m.treeTopLine
	if start < 0 {
		start = 0
	} else if start > maxTop {
		start = maxTop
	}
	end := start + listHeight
	if end > totalLines {
		end = totalLines
	}

	visible := append([]string{}, lines[start:end]...)
	for len(visible) < listHeight {
		visible = append(visible, "")
	}

	return strings.Join(visible, "\n")
}

func (m *App) buildTreeLines(totalWidth int) ([]string, int, int) {
	lines := make([]string, 0, len(m.visibleRows))
	cursorStart, cursorEnd := -1, -1
	columns, treeWidth := prepareColumnState(totalWidth)
	showColumns := columns.enabled()

	// Track which nodes are selected for cross-highlighting
	var selectedID string
	if m.cursor >= 0 && m.cursor < len(m.visibleRows) {
		selectedID = m.visibleRows[m.cursor].Node.Issue.ID
	}

	for i, row := range m.visibleRows {
		node := row.Node
		indent := strings.Repeat("  ", row.Depth)
		marker := " •"
		if len(node.Children) > 0 {
			if m.isNodeExpandedInView(row) {
				marker = " ▼"
			} else {
				marker = " ▶"
			}
		}

		iconStr, iconStyle, textStyle := "○", styleNormalText(), styleNormalText()
		domainIssue, err := domain.NewIssueFromFull(node.Issue, node.IsBlocked)
		status := node.Issue.Status
		if err == nil {
			status = string(domainIssue.Status())
		}
		switch status {
		case "in_progress":
			iconStr, iconStyle, textStyle = "◐", styleIconInProgress(), styleInProgressText()
		case "closed":
			iconStr, iconStyle, textStyle = "✔", styleIconDone(), styleDoneText()
		default:
			if node.IsBlocked {
				iconStr, iconStyle, textStyle = "⛔", styleIconBlocked(), styleBlockedText()
			}
		}

		// Add * indicator for multi-parent items
		idDisplay := node.Issue.ID
		if row.HasMultipleParents() {
			idDisplay = node.Issue.ID + "*"
		}

		totalPrefixWidth := treePrefixWidth(indent, marker, iconStr, idDisplay)
		titleLines := []string{node.Issue.Title}

		if showColumns {
			availableWidth := treeWidth - totalPrefixWidth
			titleLines[0] = truncateWithEllipsis(node.Issue.Title, availableWidth)
		} else {
			wrapWidth := treeWidth - 4
			if wrapWidth < 1 {
				wrapWidth = 1
			}
			wrappedTitle := wrapWithHangingIndent(totalPrefixWidth, node.Issue.Title, wrapWidth)
			titleLines = strings.Split(wrappedTitle, "\n")
			if len(titleLines) == 0 {
				titleLines = []string{""}
			}
		}

		// Cross-highlighting: same node appears under multiple parents
		isCrossHighlight := i != m.cursor && node.Issue.ID == selectedID

		// Style for spacing/separators to maintain background
		sp := styleNormalText().Render(" ")

		if i == m.cursor {
			cursorStart = len(lines)
			// Build full-width selected row
			line := buildSelectedRow(indent, marker, iconStr, iconStyle, idDisplay, titleLines[0], textStyle, treeWidth)
			lines = append(lines, line)
			// Handle wrapped continuation lines with selection background
			for k := 1; k < len(titleLines); k++ {
				contLine := buildSelectedContinuation(titleLines[k], textStyle, treeWidth)
				lines = append(lines, contLine)
			}
			cursorEnd = len(lines)
		} else if isCrossHighlight {
			// Cross-highlight style for duplicate instances - also full width
			line := buildCrossHighlightRow(indent, marker, iconStr, iconStyle, idDisplay, titleLines[0], textStyle, treeWidth)
			lines = append(lines, line)
			for k := 1; k < len(titleLines); k++ {
				lines = append(lines, sp+textStyle.Render(titleLines[k]))
			}
		} else {
			// Style the indent and all spacing with background
			styledIndent := styleNormalText().Render(" " + indent)
			line1 := styledIndent + iconStyle.Render(marker) + sp + iconStyle.Render(iconStr) + sp + styleID().Render(idDisplay) + sp + textStyle.Render(titleLines[0])
			lines = append(lines, line1)
			for k := 1; k < len(titleLines); k++ {
				lines = append(lines, sp+textStyle.Render(titleLines[k]))
			}
		}
	}
	return lines, cursorStart, cursorEnd
}

func (m *App) ensureTreeSelectionVisible(listHeight, totalLines, cursorStart, cursorEnd int) {
	if listHeight < 1 {
		listHeight = 1
	}
	maxTop := totalLines - listHeight
	if maxTop < 0 {
		maxTop = 0
	}
	if m.treeTopLine < 0 {
		m.treeTopLine = 0
	} else if m.treeTopLine > maxTop {
		m.treeTopLine = maxTop
	}
	if cursorStart < 0 {
		return
	}

	margin := treeScrollMargin
	if margin > listHeight/2 {
		margin = listHeight / 2
	}
	top := m.treeTopLine
	if cursorStart < top+margin {
		top = cursorStart - margin
	}

	cursorBottom := cursorEnd - 1
	if cursorBottom < cursorStart {
		cursorBottom = cursorStart
	}
	maxVisible := top + listHeight - 1 - margin
	if cursorBottom > maxVisible {
		top = cursorBottom - (listHeight - 1 - margin)
	}

	if top < 0 {
		top = 0
	} else if top > maxTop {
		top = maxTop
	}
	m.treeTopLine = top
}

func wrapWithHangingIndent(prefixWidth int, text string, maxWidth int) string {
	if maxWidth <= prefixWidth {
		return text
	}

	contentWidth := maxWidth - prefixWidth
	if contentWidth <= 0 {
		contentWidth = 10
	}

	wrapped := wordwrap.String(text, contentWidth)

	lines := strings.Split(wrapped, "\n")
	if len(lines) <= 1 {
		return text
	}

	var sb strings.Builder
	sb.WriteString(lines[0])

	padding := strings.Repeat(" ", prefixWidth)
	for i := 1; i < len(lines); i++ {
		sb.WriteString("\n")
		sb.WriteString(padding)
		sb.WriteString(lines[i])
	}
	return sb.String()
}

func truncateWithEllipsis(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	if lipgloss.Width(text) <= maxWidth {
		return text
	}

	ellipsis := "..."
	ellipsisWidth := lipgloss.Width(ellipsis)
	if maxWidth <= ellipsisWidth {
		return strings.Repeat(".", maxWidth)
	}

	runes := []rune(text)
	for i := len(runes); i >= 0; i-- {
		candidate := string(runes[:i])
		if lipgloss.Width(candidate)+ellipsisWidth <= maxWidth {
			return candidate + ellipsis
		}
	}

	return strings.Repeat(".", maxWidth)
}

func treePrefixWidth(indent, marker, icon, id string) int {
	raw := fmt.Sprintf(" %s%s %s %s ", indent, marker, icon, id)
	width := lipgloss.Width(raw)
	if width < 0 {
		return 0
	}
	return width
}

// buildSelectedRow creates a full-width row with selection background.
// It preserves the icon's status color while applying selection background to all elements.
func buildSelectedRow(indent, marker, icon string, iconStyle lipgloss.Style, id, title string, textStyle lipgloss.Style, width int) string {
	t := currentThemeWrapper()
	bg := t.BackgroundSecondary()

	// Create styles with selection background
	selectedBase := lipgloss.NewStyle().Background(bg)
	selectedPrefix := selectedBase.Bold(true).Foreground(t.Primary())
	selectedIcon := selectedBase.Foreground(iconStyle.GetForeground())
	selectedID := selectedBase.Foreground(t.Accent()).Bold(true)
	selectedText := selectedBase.Bold(true).Foreground(textStyle.GetForeground())

	// Build the row content
	content := selectedPrefix.Render(fmt.Sprintf(" %s%s ", indent, marker)) +
		selectedIcon.Render(icon) + selectedBase.Render(" ") +
		selectedID.Render(id) + selectedBase.Render(" ") +
		selectedText.Render(title)

	// Pad to full width with selection background
	return lipgloss.NewStyle().
		Background(bg).
		Width(width).
		Render(content)
}

// buildSelectedContinuation creates a continuation line for wrapped titles with selection background.
func buildSelectedContinuation(text string, textStyle lipgloss.Style, width int) string {
	t := currentThemeWrapper()
	bg := t.BackgroundSecondary()

	selectedText := lipgloss.NewStyle().
		Background(bg).
		Bold(true).
		Foreground(textStyle.GetForeground())

	content := lipgloss.NewStyle().Background(bg).Render(" ") + selectedText.Render(text)

	return lipgloss.NewStyle().
		Background(bg).
		Width(width).
		Render(content)
}

// buildCrossHighlightRow creates a full-width row with cross-highlight background.
func buildCrossHighlightRow(indent, marker, icon string, iconStyle lipgloss.Style, id, title string, textStyle lipgloss.Style, width int) string {
	t := currentThemeWrapper()
	bg := t.BorderNormal()

	// Create styles with cross-highlight background
	crossBase := lipgloss.NewStyle().Background(bg)
	crossPrefix := crossBase.Foreground(t.TextMuted())
	crossIcon := crossBase.Foreground(iconStyle.GetForeground())
	crossID := crossBase.Foreground(t.Accent()).Bold(true)
	crossText := crossBase.Foreground(textStyle.GetForeground())

	// Build the row content
	content := crossPrefix.Render(fmt.Sprintf(" %s%s ", indent, marker)) +
		crossIcon.Render(icon) + crossBase.Render(" ") +
		crossID.Render(id) + crossBase.Render(" ") +
		crossText.Render(title)

	// Pad to full width with cross-highlight background
	return lipgloss.NewStyle().
		Background(bg).
		Width(width).
		Render(content)
}
