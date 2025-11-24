package ui

import (
	"strings"
	"testing"
	"time"

	"abacus/internal/beads"
	"abacus/internal/graph"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

func TestRenderRefRow_HangingIndentForWrappedTitles(t *testing.T) {
	id := "ab-ltb"
	title := "Research Beads repository structure and CI/CD setup"
	widths := []int{34, 40, 60}
	for _, targetWidth := range widths {
		row := renderRefRow(
			id,
			title,
			targetWidth,
			lipgloss.NewStyle(),
			lipgloss.NewStyle(),
			lipgloss.Color(""),
		)
		lines := splitAndTrim(row)
		validateWrappedLines(t, lines, id)
	}
}

func TestDetailHeaderWrappingAcrossWidths(t *testing.T) {
	id := "ab-ltb"
	title := "Research Beads repository structure and CI/CD setup"
	widths := []int{34, 40, 50, 60, 80, 120}
	for _, width := range widths {
		headerWidth := width - styleDetailHeaderBlock.GetHorizontalFrameSize()
		if headerWidth < 1 {
			headerWidth = 1
		}
		headerContent := renderRefRow(
			id,
			title,
			headerWidth,
			styleDetailHeaderCombined.Foreground(cGold),
			styleDetailHeaderCombined.Foreground(cWhite),
			cHighlight,
		)
		headerBlock := styleDetailHeaderBlock.Width(width).Render(headerContent)
		lines := splitStripANSI(headerBlock)
		validateWrappedLines(t, lines, id)
	}
}

func splitAndTrim(content string) []string {
	raw := strings.Split(content, "\n")
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		lines = append(lines, strings.TrimRight(line, " "))
	}
	return lines
}

func splitStripANSI(content string) []string {
	raw := strings.Split(content, "\n")
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		clean := strings.TrimRight(stripANSI(line), " ")
		if strings.HasPrefix(clean, " ") {
			clean = clean[1:]
		}
		lines = append(lines, clean)
	}
	return lines
}

func validateWrappedLines(t *testing.T, lines []string, id string) {
	t.Helper()
	if len(lines) == 0 {
		t.Fatalf("no lines to validate")
	}

	prefix := id + "  "
	if !strings.HasPrefix(lines[0], prefix) {
		t.Fatalf("line 1 missing prefix %q: %q", prefix, lines[0])
	}

	indentWidth := len(prefix)
	expectedIndent := strings.Repeat(" ", indentWidth)

	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			t.Fatalf("blank wrapped line at index %d", i)
		}
		if !strings.HasPrefix(line, expectedIndent) {
			t.Fatalf("wrapped line %d missing indent %q: %q", i, expectedIndent, line)
		}
		if strings.Contains(line, id) {
			t.Fatalf("ID should not appear on wrapped line %d: %q", i, line)
		}
		if strings.HasSuffix(strings.TrimRight(line, " "), "-") {
			t.Fatalf("wrapped line %d should not end with hyphen: %q", i, line)
		}
	}
}

func TestDetailHeaderRegression_ab176(t *testing.T) {
	id := "ab-176"
	title := "Add incremental background refresh for real-time updates"
	cases := map[int][]string{
		40: {
			"ab-176  Add incremental background",
			"        refresh for real-time updates",
		},
		60: {
			"ab-176  Add incremental background refresh for real-time",
			"        updates",
		},
		120: {
			"ab-176  Add incremental background refresh for real-time updates",
		},
	}

	for width, want := range cases {
		headerWidth := width - styleDetailHeaderBlock.GetHorizontalFrameSize()
		if headerWidth < 1 {
			headerWidth = 1
		}
		headerContent := renderRefRow(
			id,
			title,
			headerWidth,
			styleDetailHeaderCombined.Foreground(cGold),
			styleDetailHeaderCombined.Foreground(cWhite),
			cHighlight,
		)
		block := styleDetailHeaderBlock.Width(width).Render(headerContent)
		lines := splitStripANSI(block)
		if len(lines) != len(want) {
			t.Fatalf("width %d: expected %d lines, got %d: %v", width, len(want), len(lines), lines)
		}
		for i := range want {
			if lines[i] != want[i] {
				t.Fatalf("width %d line %d mismatch:\nwant: %q\ngot:  %q", width, i, want[i], lines[i])
			}
		}
	}
}

func TestUpdateViewportContentSkipsWhenNoSelection(t *testing.T) {
	app := &App{
		ShowDetails: true,
		viewport:    viewport.New(80, 20),
	}
	app.updateViewportContent()
	content := strings.TrimSpace(stripANSI(app.viewport.View()))
	if content != "" {
		t.Fatalf("expected blank viewport when no selection, got %q", content)
	}
}

func TestDetailSectionsHaveNoBlankLineBetweenLabelAndContent(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:                 "ab-200",
			Title:              "Spacing Check",
			Description:        "Line one\n\nLine two",
			Design:             "\nDesign body",
			AcceptanceCriteria: " \n- Acceptance item",
			Comments: []beads.Comment{
				{Author: "qa", Text: "Looks good", CreatedAt: time.Now().Format(time.RFC3339)},
			},
		},
		CommentsLoaded: true,
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  []*graph.Node{node},
		viewport:     viewport.New(80, 30),
		outputFormat: "plain",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	for _, label := range []string{"Description:", "Design:", "Acceptance:", "Comments:"} {
		assertNoWhitespaceLineAfterLabel(t, content, label)
	}
}

func TestDetailSectionsWithRichMarkdownHaveNoLeadingBlankLine(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:          "ab-201",
			Title:       "Rich markdown spacing",
			Description: "- Bullet item\n- Second item",
		},
		CommentsLoaded: true,
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  []*graph.Node{node},
		viewport:     viewport.New(80, 30),
		outputFormat: "rich",
	}
	app.updateViewportContent()
	content := app.viewport.View()
	assertNoWhitespaceLineAfterLabel(t, content, "Description:")
}

func TestDetailPaneLimitsBlankLinesBetweenSections(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	parent := &graph.Node{
		Issue: beads.FullIssue{
			ID:        "ab-prt",
			Title:     "Parent work",
			Status:    "open",
			IssueType: "task",
			Priority:  2,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	child := &graph.Node{
		Issue: beads.FullIssue{
			ID:        "ab-chl",
			Title:     "Child work",
			Status:    "open",
			IssueType: "task",
			Priority:  2,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	blockedBy := &graph.Node{
		Issue: beads.FullIssue{
			ID:        "ab-blk",
			Title:     "Blocking issue",
			Status:    "open",
			IssueType: "bug",
			Priority:  1,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	blocks := &graph.Node{
		Issue: beads.FullIssue{
			ID:        "ab-blo",
			Title:     "Blocked issue",
			Status:    "open",
			IssueType: "task",
			Priority:  3,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:                 "ab-210",
			Title:              "Ensure consistent spacing",
			Status:             "open",
			IssueType:          "bug",
			Priority:           2,
			Description:        "Description content",
			Design:             "Design details",
			AcceptanceCriteria: "Acceptance details",
			CreatedAt:          now,
			UpdatedAt:          now,
			ExternalRef:        "jira-210",
			Labels:             []string{"ui", "spacing"},
			Comments: []beads.Comment{
				{Author: "qa", Text: "Looks good to me", CreatedAt: now},
			},
		},
		Parent:         parent,
		Children:       []*graph.Node{child},
		BlockedBy:      []*graph.Node{blockedBy},
		Blocks:         []*graph.Node{blocks},
		IsBlocked:      true,
		CommentsLoaded: true,
	}

	app := &App{
		ShowDetails:  true,
		visibleRows:  []*graph.Node{node},
		viewport:     viewport.New(120, 60),
		outputFormat: "plain",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	if strings.Contains(content, "\n\n\n") {
		t.Fatalf("expected at most one blank line between sections:\n%s", content)
	}
}

func assertNoWhitespaceLineAfterLabel(t *testing.T, content, label string) {
	t.Helper()
	idx := strings.Index(content, label)
	if idx == -1 {
		t.Fatalf("label %q not found in content:\n%s", label, content)
	}
	afterLabel := content[idx+len(label):]
	lineBreak := strings.Index(afterLabel, "\n")
	if lineBreak == -1 {
		t.Fatalf("no newline found after label %s", label)
	}
	nextStart := idx + len(label) + lineBreak + 1
	if nextStart >= len(content) {
		t.Fatalf("no content after label %s", label)
	}
	rest := content[nextStart:]
	nextLineEnd := strings.Index(rest, "\n")
	if nextLineEnd == -1 {
		nextLineEnd = len(rest)
	}
	nextLine := rest[:nextLineEnd]
	if strings.TrimSpace(stripANSI(nextLine)) == "" {
		t.Fatalf("blank line detected immediately after label %s:\n%s", label, content)
	}
}

func TestDetailRelationshipsShowStatusIcons(t *testing.T) {
	child := &graph.Node{Issue: beads.FullIssue{ID: "ab-601", Title: "Child Active", Status: "in_progress"}, CommentsLoaded: true}
	parent := &graph.Node{Issue: beads.FullIssue{ID: "ab-600", Title: "Parent", Status: "open"}, Children: []*graph.Node{child}, CommentsLoaded: true}
	app := &App{
		ShowDetails:  true,
		visibleRows:  []*graph.Node{parent},
		viewport:     viewport.New(90, 30),
		outputFormat: "plain",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	if !strings.Contains(content, "Depends On") {
		t.Fatalf("expected Depends On section in detail view:\n%s", content)
	}
	if !strings.Contains(content, "◐") || !strings.Contains(content, "ab-601") {
		t.Fatalf("expected in-progress icon with child id, got:\n%s", content)
	}
}

func TestDetailBlockedByShowsBlockedIcon(t *testing.T) {
	blocker := &graph.Node{Issue: beads.FullIssue{ID: "ab-611", Title: "Blocker", Status: "open"}, IsBlocked: true, CommentsLoaded: true}
	node := &graph.Node{
		Issue:          beads.FullIssue{ID: "ab-612", Title: "Blocked Node", Status: "open"},
		BlockedBy:      []*graph.Node{blocker},
		IsBlocked:      true,
		CommentsLoaded: true,
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  []*graph.Node{node},
		viewport:     viewport.New(90, 30),
		outputFormat: "plain",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	if !strings.Contains(content, "Blocked By") {
		t.Fatalf("expected Blocked By section in detail view:\n%s", content)
	}
	if !strings.Contains(content, "⛔") || !strings.Contains(content, "ab-611") {
		t.Fatalf("expected blocked icon rendered for dependency:\n%s", content)
	}
}
