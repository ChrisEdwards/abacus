package ui

import (
	"fmt"
	"strings"

	"abacus/internal/domain"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

func (m *App) renderTreeView() string {
	listHeight := clampDimension(m.height-4, minListHeight, m.height-2)
	if len(m.visibleRows) == 0 {
		return ""
	}

	var treeLines []string
	start, end := 0, len(m.visibleRows)
	if end > listHeight {
		if m.cursor > listHeight/2 {
			start = m.cursor - listHeight/2
		}
		if start+listHeight < end {
			end = start + listHeight
		} else {
			start = end - listHeight
			if start < 0 {
				start = 0
			}
		}
	}

	treeWidth := m.width - 2
	if m.ShowDetails {
		treeWidth = m.width - m.viewport.Width - 4
	}
	treeWidth = clampDimension(treeWidth, minTreeWidth, m.width-2)

	visualLinesCount := 0
	for i := start; i < end; i++ {
		if visualLinesCount >= listHeight {
			break
		}
		node := m.visibleRows[i]

		indent := strings.Repeat("  ", node.Depth)
		marker := " •"
		if len(node.Children) > 0 {
			if m.isNodeExpandedInView(node) {
				marker = " ▼"
			} else {
				marker = " ▶"
			}
		}

		iconStr, iconStyle, textStyle := "○", styleNormalText, styleNormalText
		domainIssue, err := domain.NewIssueFromFull(node.Issue, node.IsBlocked)
		status := node.Issue.Status
		if err == nil {
			status = string(domainIssue.Status())
		}
		switch status {
		case "in_progress":
			iconStr, iconStyle, textStyle = "◐", styleIconInProgress, styleInProgressText
		case "closed":
			iconStr, iconStyle, textStyle = "✔", styleIconDone, styleDoneText
		default:
			if node.IsBlocked {
				iconStr, iconStyle, textStyle = "⛔", styleIconBlocked, styleBlockedText
			}
		}

		wrapWidth := treeWidth - 4
		if wrapWidth < 1 {
			wrapWidth = 1
		}
		totalPrefixWidth := treePrefixWidth(indent, marker, iconStr, node.Issue.ID)
		wrappedTitle := wrapWithHangingIndent(totalPrefixWidth, node.Issue.Title, wrapWidth)
		titleLines := strings.Split(wrappedTitle, "\n")

		if i == m.cursor {
			highlightedPrefix := styleSelected.Render(fmt.Sprintf(" %s%s", indent, marker))
			line1Rest := fmt.Sprintf(" %s %s %s", iconStyle.Render(iconStr), styleID.Render(node.Issue.ID), textStyle.Render(titleLines[0]))
			treeLines = append(treeLines, highlightedPrefix+line1Rest)
			visualLinesCount++
		} else {
			line1Prefix := fmt.Sprintf(" %s%s %s ", indent, iconStyle.Render(marker), iconStyle.Render(iconStr))
			line1 := fmt.Sprintf("%s%s %s", line1Prefix, styleID.Render(node.Issue.ID), textStyle.Render(titleLines[0]))
			treeLines = append(treeLines, line1)
			visualLinesCount++
		}

		for k := 1; k < len(titleLines); k++ {
			if visualLinesCount >= listHeight {
				break
			}
			treeLines = append(treeLines, " "+textStyle.Render(titleLines[k]))
			visualLinesCount++
		}
	}

	for visualLinesCount < listHeight {
		treeLines = append(treeLines, "")
		visualLinesCount++
	}

	return strings.Join(treeLines, "\n")
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

func treePrefixWidth(indent, marker, icon, id string) int {
	raw := fmt.Sprintf(" %s%s %s %s ", indent, marker, icon, id)
	width := lipgloss.Width(raw)
	if width < 0 {
		return 0
	}
	return width
}
