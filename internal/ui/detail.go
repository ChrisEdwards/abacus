package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"abacus/internal/domain"
	"abacus/internal/graph"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

const (
	detailSectionLabelIndent   = 1
	detailSectionContentIndent = detailSectionLabelIndent + 1
)

func (m *App) updateViewportContent() {
	if !m.ShowDetails {
		return
	}
	if len(m.visibleRows) == 0 || m.cursor < 0 || m.cursor >= len(m.visibleRows) {
		m.viewport.SetContent("")
		return
	}
	node := m.visibleRows[m.cursor].Node

	if !node.CommentsLoaded && node.CommentError == "" {
		if err := fetchCommentsForNode(context.Background(), m.client, node); err != nil {
			node.CommentError = err.Error()
		}
	}

	iss := node.Issue
	if m.detailIssueID != iss.ID {
		m.viewport.GotoTop()
	}
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
	metaBlock = lipgloss.NewStyle().MarginLeft(1).Render(metaBlock)

	relSections := make([]string, 0, 6)

	renderRelSection := func(title string, items []*graph.Node) string {
		if len(items) == 0 {
			return ""
		}
		const extraPadding = 2
		rowWidth := vpWidth - detailSectionContentIndent - extraPadding
		if rowWidth < 1 {
			rowWidth = 1
		}
		rows := make([]string, 0, len(items))
		for _, item := range items {
			icon, iconStyle, titleStyle := relatedStatusPresentation(item)
			row := renderRefRowWithIcon(
				icon,
				iconStyle,
				item.Issue.ID,
				item.Issue.Title,
				rowWidth,
				styleID,
				titleStyle,
			)
			rows = append(rows, row)
		}
		return renderContentSection(title, strings.Join(rows, "\n"))
	}

	// Part Of - show ALL parents (parent-child relationships)
	if len(node.Parents) > 0 {
		if section := renderRelSection("Part Of", node.Parents); section != "" {
			relSections = append(relSections, section)
		}
	}
	// Subtasks - children of this node (sorted: in_progress → ready → blocked → closed)
	if len(node.Children) > 0 {
		sorted := sortSubtasks(node.Children)
		if section := renderRelSection(fmt.Sprintf("Subtasks (%d)", len(node.Children)), sorted); section != "" {
			relSections = append(relSections, section)
		}
	}
	// Must Complete First - blockers (sorted: topological order, things to do first)
	if len(node.BlockedBy) > 0 {
		sorted := sortBlockers(node.BlockedBy)
		if section := renderRelSection("Must Complete First", sorted); section != "" {
			relSections = append(relSections, section)
		}
	}
	// Will Unblock - what this issue blocks (sorted: items becoming ready first)
	if len(node.Blocks) > 0 {
		sorted := sortBlocked(node.Blocks)
		if section := renderRelSection(fmt.Sprintf("Will Unblock (%d)", len(node.Blocks)), sorted); section != "" {
			relSections = append(relSections, section)
		}
	}
	// See Also - related issues (bidirectional soft links)
	if len(node.Related) > 0 {
		if section := renderRelSection("See Also", node.Related); section != "" {
			relSections = append(relSections, section)
		}
	}
	// Discovered While Working On - issues that led to discovering this one
	if len(node.DiscoveredFrom) > 0 {
		if section := renderRelSection("Discovered While Working On", node.DiscoveredFrom); section != "" {
			relSections = append(relSections, section)
		}
	}
	relBlock := joinDetailSections(relSections...)

	renderMarkdown := buildMarkdownRenderer(m.outputFormat, vpWidth-2)
	descSections := []string{
		renderContentSection("Description:", renderMarkdown(iss.Description)),
	}
	if strings.TrimSpace(iss.Design) != "" {
		descSections = append(descSections, renderContentSection("Design:", renderMarkdown(iss.Design)))
	}
	if strings.TrimSpace(iss.AcceptanceCriteria) != "" {
		descSections = append(descSections, renderContentSection("Acceptance:", renderMarkdown(iss.AcceptanceCriteria)))
	}
	if node.CommentError != "" {
		errorBody := styleBlockedText.Render("Failed to load comments. Press 'c' to retry.") + "\n" +
			indentBlock(wordwrap.String(node.CommentError, vpWidth-4), 2)
		descSections = append(descSections, renderContentSection("Comments:", errorBody))
	} else if len(iss.Comments) > 0 {
		var commentBlocks []string
		for _, c := range iss.Comments {
			header := fmt.Sprintf("  %s  %s", c.Author, formatTime(c.CreatedAt))
			body := styleCommentHeader.Render(header) + "\n" + indentBlock(renderMarkdown(c.Text), 2)
			commentBlocks = append(commentBlocks, body)
		}
		descSections = append(descSections, renderContentSection("Comments:", strings.Join(commentBlocks, "\n\n")))
	}
	descBlock := joinDetailSections(descSections...)

	finalContent := joinDetailSections(
		headerBlock,
		metaBlock,
		relBlock,
		descBlock,
	)

	m.viewport.SetContent(finalContent)
	m.detailIssueID = iss.ID
}

func renderContentSection(label, body string) string {
	cleanBody := normalizeSectionBody(body)
	indentedBody := alignSectionBody(cleanBody, detailSectionContentIndent)
	var sb strings.Builder
	sb.WriteString(styleSectionHeader.Render(label))
	sb.WriteString("\n")
	sb.WriteString(indentedBody)
	return sb.String()
}

func normalizeSectionBody(body string) string {
	body = strings.TrimRight(body, "\r\n")
	return trimLeadingWhitespaceLines(body)
}

func joinDetailSections(sections ...string) string {
	cleaned := make([]string, 0, len(sections))
	for _, section := range sections {
		if strings.TrimSpace(section) == "" {
			continue
		}
		cleaned = append(cleaned, strings.Trim(section, "\n\r"))
	}
	return strings.Join(cleaned, "\n\n")
}

func trimLeadingWhitespaceLines(body string) string {
	body = strings.TrimLeft(body, "\r\n")
	for len(body) > 0 {
		lineEnd := strings.IndexByte(body, '\n')
		line := body
		nextStart := len(body)
		if lineEnd != -1 {
			line = body[:lineEnd]
			nextStart = lineEnd + 1
		}
		if !isVisualBlankLine(line) {
			break
		}
		body = strings.TrimLeft(body[nextStart:], "\r\n")
	}
	return body
}

func alignSectionBody(body string, indent int) string {
	lines := strings.Split(body, "\n")
	if len(lines) == 0 {
		return ""
	}
	padding := strings.Repeat(" ", indent)
	common := commonLeadingSpaces(lines)
	for i, line := range lines {
		if strings.TrimSpace(stripANSI(line)) == "" {
			lines[i] = ""
			continue
		}
		trimmed := trimANSIIndent(line, common)
		lines[i] = padding + trimmed
	}
	return strings.Join(lines, "\n")
}

func commonLeadingSpaces(lines []string) int {
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(stripANSI(line)) == "" {
			continue
		}
		count := countLeadingSpacesANSI(line)
		if minIndent == -1 || count < minIndent {
			minIndent = count
		}
	}
	if minIndent < 0 {
		return 0
	}
	return minIndent
}

func countLeadingSpacesANSI(line string) int {
	count := 0
	i := 0
	for i < len(line) {
		if line[i] == '\x1b' {
			end := i + 1
			for end < len(line) && line[end-1] != 'm' {
				end++
			}
			if end >= len(line) {
				break
			}
			i = end
			continue
		}
		if line[i] == ' ' {
			count++
			i++
			continue
		}
		break
	}
	return count
}

func trimANSIIndent(line string, spaces int) string {
	if spaces <= 0 {
		return line
	}
	var b strings.Builder
	remaining := spaces
	i := 0
	for i < len(line) {
		if line[i] == '\x1b' {
			end := i + 1
			for end < len(line) && line[end-1] != 'm' {
				end++
			}
			if end > len(line) {
				end = len(line)
			}
			b.WriteString(line[i:end])
			i = end
			continue
		}
		if line[i] == ' ' && remaining > 0 {
			remaining--
			i++
			continue
		}
		break
	}
	b.WriteString(line[i:])
	return b.String()
}

func isVisualBlankLine(line string) bool {
	return strings.TrimSpace(stripANSI(line)) == ""
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

func renderRefRowWithIcon(icon string, iconStyle lipgloss.Style, id, title string, targetWidth int, idStyle, titleStyle lipgloss.Style) string {
	const gap = "  "
	iconRendered := iconStyle.Render(icon)
	idRendered := idStyle.Render(id)
	gapRendered := gap
	iconWidth := lipgloss.Width(iconRendered)
	idWidth := lipgloss.Width(idRendered)
	gapWidth := lipgloss.Width(gapRendered)
	prefixWidth := iconWidth + 1 + idWidth + gapWidth
	titleWidth := targetWidth - prefixWidth
	if titleWidth < 1 {
		titleWidth = 1
	}
	titleLines := wrapTitleWithoutHyphenBreaks(title, titleWidth)
	if len(titleLines) == 0 {
		titleLines = []string{""}
	}
	prefixFirst := lipgloss.JoinHorizontal(lipgloss.Left,
		iconRendered,
		lipgloss.NewStyle().Render(" "),
		idRendered,
		gapRendered,
	)
	prefixBlank := strings.Repeat(" ", lipgloss.Width(prefixFirst))
	lines := make([]string, len(titleLines))
	for i, line := range titleLines {
		prefix := prefixBlank
		if i == 0 {
			prefix = prefixFirst
		}
		lines[i] = lipgloss.JoinHorizontal(lipgloss.Left, prefix, titleStyle.Width(titleWidth).Render(line))
	}
	return strings.Join(lines, "\n")
}

func relatedStatusPresentation(node *graph.Node) (string, lipgloss.Style, lipgloss.Style) {
	domainIssue, err := domain.NewIssueFromFull(node.Issue, node.IsBlocked)
	status := node.Issue.Status
	if err == nil {
		status = string(domainIssue.Status())
	}
	switch status {
	case "in_progress":
		return "◐", styleIconInProgress, styleInProgressText
	case "closed":
		return "✔", styleIconDone, styleDoneText
	default:
		if node.IsBlocked {
			return "⛔", styleIconBlocked, styleBlockedText
		}
		return "○", styleIconOpen, styleNormalText
	}
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
