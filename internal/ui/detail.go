package ui

import (
	"fmt"
	"strings"

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

	// Comments are loaded asynchronously in background (ab-fkyz).
	// Do NOT block navigation - show loading state if comments not ready.

	iss := node.Issue
	if m.detailIssueID != iss.ID {
		m.viewport.GotoTop()
	}
	vpWidth := m.viewport.Width

	headerContentWidth := vpWidth - styleDetailHeaderBlock().GetHorizontalFrameSize()
	if headerContentWidth < 1 {
		headerContentWidth = 1
	}

	headerContent := renderRefRow(
		iss.ID,
		iss.Title,
		headerContentWidth,
		styleDetailHeaderCombined().Foreground(currentThemeWrapper().Accent()),
		styleDetailHeaderCombined().Foreground(currentThemeWrapper().Text()),
		currentThemeWrapper().BackgroundSecondary(),
	)
	headerBlock := styleDetailHeaderBlock().Width(vpWidth).Render(headerContent)

	makeRow := func(k, v string) string {
		return lipgloss.JoinHorizontal(lipgloss.Left, styleField().Render(k), styleVal().Render(v))
	}
	bgStyle := baseStyle()

	col1 := []string{
		makeRow("Status:", formatStatusDisplay(iss.Status, node.IsBlocked)),
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
		makeRow("Priority:", stylePrio().Render(prioLabel)),
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

		labelSpacer := bgStyle.Render(" ")
		for _, l := range iss.Labels {
			rendered := renderPillChip(l, chipStateNormal) + labelSpacer
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

		firstRow := lipgloss.JoinHorizontal(lipgloss.Left, styleField().Render("Labels:"), labelRows[0])
		finalLabelBlock := firstRow
		padding := bgStyle.Render(strings.Repeat(" ", labelPrefixWidth))
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
		metaBlock = lipgloss.JoinHorizontal(lipgloss.Top, leftStack, bgStyle.Render("    "), rightStack)
	}
	metaBlock = bgStyle.MarginLeft(1).Render(metaBlock)

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
				styleID(),
				titleStyle,
			)
			rows = append(rows, row)
		}
		return renderContentSection(title, strings.Join(rows, "\n"))
	}

	// Part Of - show ALL parents (parent-child relationships)
	if len(node.Parents) > 0 {
		if section := renderRelSection(fmt.Sprintf("Part Of: (%d)", len(node.Parents)), node.Parents); section != "" {
			relSections = append(relSections, section)
		}
	}
	// Subtasks - children of this node (sorted: in_progress → ready → blocked → closed)
	if len(node.Children) > 0 {
		sorted := sortSubtasks(node.Children)
		if section := renderRelSection(fmt.Sprintf("Subtasks: (%d)", len(node.Children)), sorted); section != "" {
			relSections = append(relSections, section)
		}
	}
	// Must Complete First - blockers (sorted: topological order, things to do first)
	if len(node.BlockedBy) > 0 {
		sorted := sortBlockers(node.BlockedBy)
		if section := renderRelSection(fmt.Sprintf("Must Complete First: (%d)", len(node.BlockedBy)), sorted); section != "" {
			relSections = append(relSections, section)
		}
	}
	// Will Unblock - what this issue blocks (sorted: items becoming ready first)
	if len(node.Blocks) > 0 {
		sorted := sortBlocked(node.Blocks)
		if section := renderRelSection(fmt.Sprintf("Will Unblock: (%d)", len(node.Blocks)), sorted); section != "" {
			relSections = append(relSections, section)
		}
	}
	// See Also - related issues (bidirectional soft links)
	if len(node.Related) > 0 {
		if section := renderRelSection(fmt.Sprintf("See Also: (%d)", len(node.Related)), node.Related); section != "" {
			relSections = append(relSections, section)
		}
	}
	// Discovered While Working On - issues that led to discovering this one
	if len(node.DiscoveredFrom) > 0 {
		if section := renderRelSection(fmt.Sprintf("Discovered While Working On: (%d)", len(node.DiscoveredFrom)), node.DiscoveredFrom); section != "" {
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
	if strings.TrimSpace(iss.Notes) != "" {
		descSections = append(descSections, renderContentSection("Notes:", renderMarkdown(iss.Notes)))
	}
	if node.CommentError != "" {
		errorBody := styleBlockedText().Render("Failed to load comments. Press 'c' to retry.") + "\n" +
			indentBlock(wordwrap.String(node.CommentError, vpWidth-4), 2)
		descSections = append(descSections, renderContentSection("Comments:", errorBody))
	} else if !node.CommentsLoaded {
		// Show loading state while comments are fetched in background (ab-o0fm)
		loadingBody := styleStatsDim().Render("Loading comments...")
		descSections = append(descSections, renderContentSection("Comments:", loadingBody))
	} else if len(iss.Comments) > 0 {
		var commentBlocks []string
		for _, c := range iss.Comments {
			header := fmt.Sprintf("  %s  %s", c.Author, formatTime(c.CreatedAt))
			body := styleCommentHeader().Render(header) + "\n" + indentBlock(renderMarkdown(c.Text), 2)
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

	// Fill background gaps before applying placement padding
	finalContent = padLinesToWidth(finalContent, vpWidth)
	finalContent = fillBackground(finalContent)

	// Also use lipgloss.Place for outer padding
	contentHeight := lipgloss.Height(finalContent)
	targetHeight := contentHeight
	if m.viewport.Height > targetHeight {
		targetHeight = m.viewport.Height
	}
	if targetHeight == 0 {
		targetHeight = 1
	}

	if vpWidth > 0 && targetHeight > 0 {
		width := vpWidth
		if width < 1 {
			width = 1
		}
		finalContent = lipgloss.Place(
			width, targetHeight,
			lipgloss.Left, lipgloss.Top,
			finalContent,
			lipgloss.WithWhitespaceBackground(currentThemeWrapper().Background()),
		)
		// lipgloss.Place can insert additional escape sequences, so reapply background fills
		finalContent = padLinesToWidth(finalContent, width)
		finalContent = fillBackground(finalContent)
	}

	m.viewport.SetContent(finalContent)
	m.detailIssueID = iss.ID
}

func renderContentSection(label, body string) string {
	cleanBody := normalizeSectionBody(body)
	indentedBody := alignSectionBody(cleanBody, detailSectionContentIndent)
	var sb strings.Builder
	// Add styled indentation before section header
	indent := baseStyle().Render(strings.Repeat(" ", detailSectionLabelIndent))
	sb.WriteString(indent)
	sb.WriteString(styleSectionHeader().Render(label))
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
	padding := baseStyle().Render(strings.Repeat(" ", indent))
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
