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

	headerContentWidth := vpWidth - styleDetailHeaderBlock.GetHorizontalFrameSize()
	if headerContentWidth < 1 {
		headerContentWidth = 1
	}

	headerContent := renderRefRow(
		iss.ID,
		iss.Title,
		headerContentWidth,
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

	if strings.TrimSpace(iss.Design) != "" {
		descBuilder.WriteString("\n" + styleSectionHeader.Render("Design") + "\n")
		descBuilder.WriteString(indentBlock(renderMarkdown(iss.Design), 2))
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
	const gap = "  "

	idStyled := idStyle.Background(bgColor)
	gapStyled := lipgloss.NewStyle().Background(bgColor)
	titleStyled := titleStyle.Background(bgColor)

	idRendered := idStyled.Render(id)
	idWidth := lipgloss.Width(idRendered)
	gapRendered := gapStyled.Render(gap)
	gapWidth := lipgloss.Width(gapRendered)
	prefixWidth := idWidth + gapWidth

	titleWidth := targetWidth - prefixWidth
	if titleWidth < 1 {
		titleWidth = 1
	}

	titleLines := wrapTitleWithoutHyphenBreaks(title, titleWidth)
	if len(titleLines) == 0 {
		titleLines = []string{""}
	}

	idBlank := idStyled.Width(idWidth).Render("")
	gapBlank := gapStyled.Width(gapWidth).Render("")

	lines := make([]string, len(titleLines))
	for i, line := range titleLines {
		idSegment := idBlank
		gapSegment := gapBlank
		if i == 0 {
			idSegment = idStyled.Width(idWidth).Render(id)
			gapSegment = gapStyled.Width(gapWidth).Render(gap)
		}
		titleSegment := titleStyled.Width(titleWidth).Render(line)
		lines[i] = lipgloss.JoinHorizontal(lipgloss.Left, idSegment, gapSegment, titleSegment)
	}

	return strings.Join(lines, "\n")
}

// wrapTitleWithoutHyphenBreaks wraps the title text while treating the ID as a
// fixed-width prefix so the ID never wraps and continuation lines align with
// the start of the title. The wrapping only happens at whitespace boundaries
// to avoid splitting hyphenated words like "real-time".
func wrapTitleWithoutHyphenBreaks(title string, width int) []string {
	if width < 1 {
		width = 1
	}
	words := strings.Fields(title)
	if len(words) == 0 {
		return []string{""}
	}

	lines := make([]string, 0, len(words))
	var current strings.Builder
	currentWidth := 0

	flush := func() {
		if current.Len() == 0 {
			return
		}
		lines = append(lines, current.String())
		current.Reset()
		currentWidth = 0
	}

	for _, word := range words {
		wordWidth := lipgloss.Width(word)
		if currentWidth == 0 {
			current.WriteString(word)
			currentWidth = wordWidth
			continue
		}

		nextWidth := currentWidth + 1 + wordWidth
		if nextWidth > width {
			flush()
			current.WriteString(word)
			currentWidth = wordWidth
			continue
		}

		current.WriteByte(' ')
		current.WriteString(word)
		currentWidth = nextWidth
	}

	flush()
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
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
