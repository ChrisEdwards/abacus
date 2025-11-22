package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"abacus/internal/graph"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

func (m *App) updateViewportContent() {
	if !m.ShowDetails || m.cursor >= len(m.visibleRows) {
		return
	}
	node := m.visibleRows[m.cursor]

	if !node.CommentsLoaded && node.CommentError == "" {
		if err := fetchCommentsForNode(context.Background(), m.client, node); err != nil {
			node.CommentError = err.Error()
		}
	}

	iss := node.Issue
	vpWidth := m.viewport.Width

	headerContent := renderRefRow(
		iss.ID,
		iss.Title,
		vpWidth,
		styleDetailHeaderCombined.Foreground(cGold),
		styleDetailHeaderCombined.Foreground(cWhite),
		cHighlight,
	)
	headerBlock := styleDetailHeaderBlock.Width(vpWidth).Render(headerContent)

	makeRow := func(k, v string) string {
		return lipgloss.JoinHorizontal(lipgloss.Left, styleField.Render(k), styleVal.Render(v))
	}

	col1 := []string{
		makeRow("Status:", iss.Status),
		makeRow("Type:", iss.IssueType),
		makeRow("Created:", formatTime(iss.CreatedAt)),
	}
	if iss.UpdatedAt != iss.CreatedAt {
		col1 = append(col1, makeRow("Updated:", formatTime(iss.UpdatedAt)))
	}
	if iss.Status == "closed" {
		col1 = append(col1, makeRow("Closed:", formatTime(iss.ClosedAt)))
	}

	prioLabel := fmt.Sprintf("P%d", iss.Priority)
	col2 := []string{
		makeRow("Priority:", stylePrio.Render(prioLabel)),
	}
	if iss.ExternalRef != "" {
		col2 = append(col2, makeRow("Ext Ref:", iss.ExternalRef))
	}

	if len(iss.Labels) > 0 {
		var labelRows []string
		var currentRow string
		currentLen := 0
		labelPrefixWidth := 12
		availableLabelWidth := (vpWidth / 2) - labelPrefixWidth
		if availableLabelWidth < 10 {
			availableLabelWidth = 10
		}

		for _, l := range iss.Labels {
			rendered := styleLabel.Render(l)
			w := lipgloss.Width(rendered)
			if currentLen+w > availableLabelWidth && currentLen > 0 {
				labelRows = append(labelRows, currentRow)
				currentRow = ""
				currentLen = 0
			}
			currentRow += rendered
			currentLen += w
		}
		if currentRow != "" {
			labelRows = append(labelRows, currentRow)
		}

		firstRow := lipgloss.JoinHorizontal(lipgloss.Left, styleField.Render("Labels:"), labelRows[0])
		finalLabelBlock := firstRow
		padding := strings.Repeat(" ", labelPrefixWidth)
		for i := 1; i < len(labelRows); i++ {
			finalLabelBlock += "\n" + padding + labelRows[i]
		}
		col2 = append(col2, finalLabelBlock)
	} else {
		col2 = append(col2, makeRow("Labels:", "-"))
	}

	leftStack := lipgloss.JoinVertical(lipgloss.Left, col1...)
	rightStack := lipgloss.JoinVertical(lipgloss.Left, col2...)

	var metaBlock string
	if vpWidth < 60 {
		metaBlock = lipgloss.JoinVertical(lipgloss.Left, leftStack, rightStack)
	} else {
		metaBlock = lipgloss.JoinHorizontal(lipgloss.Top, leftStack, "    ", rightStack)
	}
	metaBlock = lipgloss.NewStyle().MarginLeft(1).PaddingTop(1).PaddingBottom(1).Render(metaBlock)

	relBuilder := strings.Builder{}

	if iss.ExternalRef != "" {
		relBuilder.WriteString(styleSectionHeader.Render("External Reference") + "\n")
		relBuilder.WriteString(fmt.Sprintf("  🔗 %s\n\n", iss.ExternalRef))
	}

	renderRelSection := func(title string, items []*graph.Node) {
		relBuilder.WriteString(styleSectionHeader.Render(title) + "\n")
		const indentSpaces = 2
		rowWidth := vpWidth - indentSpaces - 2
		if rowWidth < 1 {
			rowWidth = 1
		}
		for _, item := range items {
			row := renderRefRow(
				item.Issue.ID,
				item.Issue.Title,
				rowWidth,
				styleID,
				styleVal,
				lipgloss.Color(""),
			)
			relBuilder.WriteString(indentBlock(row, indentSpaces) + "\n")
		}
	}

	if node.Parent != nil {
		renderRelSection("Parent", []*graph.Node{node.Parent})
	}
	if len(node.Children) > 0 {
		renderRelSection(fmt.Sprintf("Depends On (%d)", len(node.Children)), node.Children)
	}
	if node.IsBlocked {
		renderRelSection("Blocked By", node.BlockedBy)
	}
	if len(node.Blocks) > 0 {
		renderRelSection(fmt.Sprintf("Blocks (%d)", len(node.Blocks)), node.Blocks)
	}

	relBlock := ""
	if relBuilder.Len() > 0 {
		relBlock = relBuilder.String()
	}

	descBuilder := strings.Builder{}
	descBuilder.WriteString(styleSectionHeader.Render("Description") + "\n")
	renderMarkdown := buildMarkdownRenderer(m.outputFormat, vpWidth-2)
	if iss.Description == "" {
		descBuilder.WriteString(indentBlock("(no description)", 2))
	} else {
		descBuilder.WriteString(indentBlock(renderMarkdown(iss.Description), 2))
	}

	if node.CommentError != "" {
		descBuilder.WriteString("\n" + styleSectionHeader.Render("Comments") + "\n")
		descBuilder.WriteString(styleBlockedText.Render("Failed to load comments. Press 'c' to retry.") + "\n")
		wrappedErr := wordwrap.String(node.CommentError, vpWidth-4)
		descBuilder.WriteString(indentBlock(wrappedErr, 2) + "\n")
	} else if len(iss.Comments) > 0 {
		descBuilder.WriteString("\n" + styleSectionHeader.Render("Comments") + "\n")
		for _, c := range iss.Comments {
			header := fmt.Sprintf("  %s  %s", c.Author, formatTime(c.CreatedAt))
			descBuilder.WriteString(styleCommentHeader.Render(header) + "\n")

			renderedComment := renderMarkdown(c.Text)
			descBuilder.WriteString(indentBlock(renderedComment, 2) + "\n\n")
		}
	}

	finalContent := lipgloss.JoinVertical(lipgloss.Left,
		headerBlock,
		metaBlock,
		relBlock,
		descBuilder.String(),
	)

	m.viewport.SetContent(finalContent)
}

func renderRefRow(id, title string, targetWidth int, idStyle, titleStyle lipgloss.Style, bgColor lipgloss.Color) string {
	gap := "  "
	gapWidth := 2

	idRendered := idStyle.Background(bgColor).Render(id)
	idWidth := lipgloss.Width(idRendered)

	titleWidth := targetWidth - idWidth - gapWidth
	if titleWidth < 1 {
		titleWidth = 1
	}

	titleRendered := titleStyle.
		Background(bgColor).
		Width(titleWidth).
		Render(wordwrap.String(title, titleWidth))

	h := lipgloss.Height(titleRendered)
	idRendered = idStyle.Background(bgColor).Height(h).Render(id)
	gapRendered := lipgloss.NewStyle().Background(bgColor).Height(h).Render(gap)

	return lipgloss.JoinHorizontal(lipgloss.Top, idRendered, gapRendered, titleRendered)
}

func indentBlock(text string, spaces int) string {
	padding := strings.Repeat(" ", spaces)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = padding + line
		}
	}
	return strings.Join(lines, "\n")
}

func formatTime(isoStr string) string {
	if isoStr == "" {
		return "-"
	}
	t, err := time.Parse(time.RFC3339, isoStr)
	if err != nil {
		return isoStr
	}
	return t.Local().Format("Jan 02, 3:04 PM")
}
