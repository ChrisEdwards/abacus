package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"abacus/internal/beads"
	"abacus/internal/graph"
	"abacus/internal/ui/theme"

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
		headerWidth := width - styleDetailHeaderBlock().GetHorizontalFrameSize()
		if headerWidth < 1 {
			headerWidth = 1
		}
		headerContent := renderRefRow(
			id,
			title,
			headerWidth,
			styleDetailHeaderCombined().Foreground(theme.Current().Warning()),
			styleDetailHeaderCombined().Foreground(theme.Current().Text()),
			theme.Current().BackgroundSecondary(),
		)
		headerBlock := styleDetailHeaderBlock().Width(width).Render(headerContent)
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
		headerWidth := width - styleDetailHeaderBlock().GetHorizontalFrameSize()
		if headerWidth < 1 {
			headerWidth = 1
		}
		headerContent := renderRefRow(
			id,
			title,
			headerWidth,
			styleDetailHeaderCombined().Foreground(theme.Current().Warning()),
			styleDetailHeaderCombined().Foreground(theme.Current().Text()),
			theme.Current().BackgroundSecondary(),
		)
		block := styleDetailHeaderBlock().Width(width).Render(headerContent)
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
			Notes:              "Some notes here",
			Comments: []beads.Comment{
				{Author: "qa", Text: "Looks good", CreatedAt: time.Now().Format(time.RFC3339)},
			},
		},
		CommentsLoaded: true,
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  []graph.TreeRow{{Node: node}},
		viewport:     viewport.New(80, 30),
		outputFormat: "plain",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	for _, label := range []string{"Description:", "Design:", "Acceptance:", "Notes:", "Comments:"} {
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
		visibleRows:  []graph.TreeRow{{Node: node}},
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
		Parents:        []*graph.Node{parent},
		Children:       []*graph.Node{child},
		BlockedBy:      []*graph.Node{blockedBy},
		Blocks:         []*graph.Node{blocks},
		IsBlocked:      true,
		CommentsLoaded: true,
	}

	app := &App{
		ShowDetails:  true,
		visibleRows:  []graph.TreeRow{{Node: node}},
		viewport:     viewport.New(120, 60),
		outputFormat: "plain",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	if strings.Contains(content, "\n\n\n") {
		t.Fatalf("expected at most one blank line between sections:\n%s", content)
	}
}

func TestDetailViewportWhitespaceUsesThemeBackground(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:        "ab-bg1",
			Title:     "Background padding audit",
			Status:    "open",
			IssueType: "task",
			Priority:  2,
			CreatedAt: now,
			UpdatedAt: now,
		},
		CommentsLoaded: true,
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  []graph.TreeRow{{Node: node}},
		viewport:     viewport.New(70, 40),
		outputFormat: "plain",
	}
	app.updateViewportContent()

	bgSeq := theme.Current().BackgroundANSI()
	if bgSeq == "" {
		t.Fatalf("theme did not provide background sequence")
	}

	content := app.viewport.View()
	lines := strings.Split(content, "\n")

	var missing int
	for _, line := range lines {
		if strings.TrimSpace(stripANSI(line)) != "" {
			continue
		}
		if !strings.Contains(line, bgSeq) {
			missing++
		}
	}
	if missing > 0 {
		t.Fatalf("found %d whitespace lines without theme background:\n%s", missing, content)
	}
}

func TestDetailViewportAllLinesHaveThemeBackground(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:        "ab-bg2",
			Title:     "Full background coverage",
			Status:    "open",
			IssueType: "bug",
			Priority:  1,
			CreatedAt: now,
			UpdatedAt: now,
			Comments: []beads.Comment{
				{Author: "qa", Text: "Looks good", CreatedAt: now},
			},
			Labels: []string{"release"},
		},
		CommentsLoaded: true,
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  []graph.TreeRow{{Node: node}},
		viewport:     viewport.New(90, 45),
		outputFormat: "rich",
	}
	app.updateViewportContent()

	bgSeq := theme.Current().BackgroundANSI()
	if bgSeq == "" {
		t.Fatalf("theme did not provide background sequence")
	}

	content := app.viewport.View()
	lines := strings.Split(content, "\n")
	var missing []string
	for i, line := range lines {
		if !strings.Contains(line, bgSeq) {
			missing = append(missing, fmt.Sprintf("%d:%q", i, line))
		}
	}
	if len(missing) > 0 {
		t.Fatalf("lines missing theme background:\n%s", strings.Join(missing, "\n"))
	}
}

func TestDetailViewportWhitespaceLinesFillViewportWidth(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	const width = 76
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:        "ab-bg3",
			Title:     "Whitespace padding audit",
			Status:    "open",
			IssueType: "task",
			Priority:  3,
			CreatedAt: now,
			UpdatedAt: now,
			Description: strings.Join([]string{
				"Line one of description.",
				"",
				"Line three to create whitespace gap.",
			}, "\n"),
		},
		CommentsLoaded: true,
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  []graph.TreeRow{{Node: node}},
		viewport:     viewport.New(width, 40),
		outputFormat: "plain",
	}
	app.updateViewportContent()

	content := app.viewport.View()
	lines := strings.Split(content, "\n")
	var shortLines []string
	for i, line := range lines {
		if strings.TrimSpace(stripANSI(line)) != "" {
			continue
		}
		if lipgloss.Width(stripANSI(line)) < width {
			shortLines = append(shortLines, fmt.Sprintf("%d:%q", i, line))
		}
	}
	if len(shortLines) > 0 {
		t.Fatalf("whitespace lines not padded to viewport width %d:\n%s", width, strings.Join(shortLines, "\n"))
	}
}

func TestDetailSectionsUseConsistentIndentation(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	parent := &graph.Node{Issue: beads.FullIssue{ID: "ab-par", Title: "Parent", Status: "open"}}
	child := &graph.Node{Issue: beads.FullIssue{ID: "ab-chd", Title: "Child", Status: "open"}}
	blocker := &graph.Node{Issue: beads.FullIssue{ID: "ab-blk", Title: "Blocker", Status: "open"}}
	blocks := &graph.Node{Issue: beads.FullIssue{ID: "ab-bls", Title: "Blocks", Status: "open"}}

	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:                 "ab-701",
			Title:              "Indentation audit",
			Status:             "open",
			IssueType:          "bug",
			Priority:           1,
			Description:        "Paragraph one\n\nParagraph two",
			Design:             "High level design",
			AcceptanceCriteria: "- first item",
			Notes:              "Implementation notes",
			ExternalRef:        "jira-701",
			CreatedAt:          now,
			UpdatedAt:          now,
			Comments: []beads.Comment{
				{Author: "qa", Text: "Looks good", CreatedAt: now},
			},
		},
		Parent:         parent,
		Parents:        []*graph.Node{parent},
		Children:       []*graph.Node{child},
		BlockedBy:      []*graph.Node{blocker},
		Blocks:         []*graph.Node{blocks},
		IsBlocked:      true,
		CommentsLoaded: true,
	}

	app := &App{
		ShowDetails:  true,
		visibleRows:  []graph.TreeRow{{Node: node}},
		viewport:     viewport.New(100, 50),
		outputFormat: "rich",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())

	labels := []string{
		"Description:",
		"Design:",
		"Acceptance:",
		"Notes:",
		"Comments:",
		fmt.Sprintf("Part Of: (%d)", len(node.Parents)),
		fmt.Sprintf("Subtasks: (%d)", len(node.Children)),
		fmt.Sprintf("Must Complete First: (%d)", len(node.BlockedBy)),
		fmt.Sprintf("Will Unblock: (%d)", len(node.Blocks)),
	}
	for _, label := range labels {
		assertSectionIndentSpacing(t, content, label, detailSectionLabelIndent, detailSectionContentIndent)
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

func assertSectionIndentSpacing(t *testing.T, content, label string, wantLabel, wantContent int) {
	t.Helper()
	lines := strings.Split(content, "\n")
	labelIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == label {
			labelIdx = i
			break
		}
	}
	if labelIdx == -1 {
		t.Fatalf("label %q not found in content", label)
	}
	labelLine := lines[labelIdx]
	if got := leadingSpaces(labelLine); got != wantLabel {
		t.Fatalf("label %q indent=%d, want %d (line=%q)", label, got, wantLabel, labelLine)
	}
	contentLine := ""
	for i := labelIdx + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		contentLine = lines[i]
		break
	}
	if contentLine == "" {
		t.Fatalf("no content line found after label %q", label)
	}
	if got := leadingSpaces(contentLine); got != wantContent {
		t.Fatalf("content indent for %q=%d, want %d (line=%q)", label, got, wantContent, contentLine)
	}
}

func leadingSpaces(line string) int {
	count := 0
	for _, ch := range line {
		if ch != ' ' {
			break
		}
		count++
	}
	return count
}

func TestDetailRelationshipsShowStatusIcons(t *testing.T) {
	child := &graph.Node{Issue: beads.FullIssue{ID: "ab-601", Title: "Child Active", Status: "in_progress"}, CommentsLoaded: true}
	parent := &graph.Node{Issue: beads.FullIssue{ID: "ab-600", Title: "Parent", Status: "open"}, Children: []*graph.Node{child}, CommentsLoaded: true}
	app := &App{
		ShowDetails:  true,
		visibleRows:  []graph.TreeRow{{Node: parent}},
		viewport:     viewport.New(90, 30),
		outputFormat: "plain",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	if !strings.Contains(content, "Subtasks") {
		t.Fatalf("expected Subtasks section in detail view:\n%s", content)
	}
	if !strings.Contains(content, "◐") || !strings.Contains(content, "ab-601") {
		t.Fatalf("expected in-progress icon with child id, got:\n%s", content)
	}
}

func TestDetailPartOfShowsAllParents(t *testing.T) {
	parent1 := &graph.Node{Issue: beads.FullIssue{ID: "ab-p1", Title: "Parent Epic 1", Status: "open"}}
	parent2 := &graph.Node{Issue: beads.FullIssue{ID: "ab-p2", Title: "Parent Epic 2", Status: "open"}}
	parent3 := &graph.Node{Issue: beads.FullIssue{ID: "ab-p3", Title: "Parent Epic 3", Status: "in_progress"}}

	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:     "ab-child",
			Title:  "Multi-Parent Child",
			Status: "open",
		},
		Parent:         parent1, // Legacy single parent
		Parents:        []*graph.Node{parent1, parent2, parent3},
		CommentsLoaded: true,
	}

	app := &App{
		ShowDetails:  true,
		visibleRows:  []graph.TreeRow{{Node: node}},
		viewport:     viewport.New(90, 40),
		outputFormat: "plain",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())

	// "Part Of:" section should be present
	if !strings.Contains(content, "Part Of:") {
		t.Fatalf("expected 'Part Of' section for multi-parent node:\n%s", content)
	}

	// All parent IDs should be visible
	for _, parent := range node.Parents {
		if !strings.Contains(content, parent.Issue.ID) {
			t.Fatalf("expected parent %s to appear in Part Of section:\n%s", parent.Issue.ID, content)
		}
	}

	// Verify status icons are shown - parent3 is in_progress
	if !strings.Contains(content, "◐") {
		t.Fatalf("expected in-progress icon for parent3:\n%s", content)
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
		visibleRows:  []graph.TreeRow{{Node: node}},
		viewport:     viewport.New(90, 30),
		outputFormat: "plain",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	if !strings.Contains(content, "Must Complete First:") {
		t.Fatalf("expected Must Complete First section in detail view:\n%s", content)
	}
	if !strings.Contains(content, "⛔") || !strings.Contains(content, "ab-611") {
		t.Fatalf("expected blocked icon rendered for dependency:\n%s", content)
	}
}

func TestDetailViewShowsDuplicateOf(t *testing.T) {
	canonical := &graph.Node{
		Issue:          beads.FullIssue{ID: "ab-canonical", Title: "Original Feature Request", Status: "open"},
		CommentsLoaded: true,
	}
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:          "ab-dup",
			Title:       "Duplicate Report",
			Status:      "closed",
			DuplicateOf: "ab-canonical",
		},
		DuplicateOf:    canonical,
		CommentsLoaded: true,
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  []graph.TreeRow{{Node: node}},
		viewport:     viewport.New(90, 30),
		outputFormat: "plain",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())

	// "Duplicate Of:" section should be present
	if !strings.Contains(content, "Duplicate Of:") {
		t.Fatalf("expected 'Duplicate Of' section for duplicate node:\n%s", content)
	}
	// Canonical issue ID should be visible
	if !strings.Contains(content, "ab-canonical") {
		t.Fatalf("expected canonical issue ID in Duplicate Of section:\n%s", content)
	}
}

func TestDetailViewShowsSupersededBy(t *testing.T) {
	replacement := &graph.Node{
		Issue:          beads.FullIssue{ID: "ab-new", Title: "Updated Design Doc v2", Status: "in_progress"},
		CommentsLoaded: true,
	}
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:           "ab-old",
			Title:        "Old Design Doc",
			Status:       "closed",
			SupersededBy: "ab-new",
		},
		SupersededBy:   replacement,
		CommentsLoaded: true,
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  []graph.TreeRow{{Node: node}},
		viewport:     viewport.New(90, 30),
		outputFormat: "plain",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())

	// "Superseded By:" section should be present
	if !strings.Contains(content, "Superseded By:") {
		t.Fatalf("expected 'Superseded By' section for superseded node:\n%s", content)
	}
	// Replacement issue ID should be visible
	if !strings.Contains(content, "ab-new") {
		t.Fatalf("expected replacement issue ID in Superseded By section:\n%s", content)
	}
	// Status icon should be present (in_progress = ◐)
	if !strings.Contains(content, "◐") {
		t.Fatalf("expected in-progress icon for replacement:\n%s", content)
	}
}

func TestDetailViewHidesEmptyGraphLinks(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:     "ab-normal",
			Title:  "Normal Issue",
			Status: "open",
		},
		DuplicateOf:    nil, // No duplicate_of
		SupersededBy:   nil, // No superseded_by
		CommentsLoaded: true,
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  []graph.TreeRow{{Node: node}},
		viewport:     viewport.New(90, 30),
		outputFormat: "plain",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())

	// Neither section should appear when pointers are nil
	if strings.Contains(content, "Duplicate Of:") {
		t.Fatalf("did not expect 'Duplicate Of' section when DuplicateOf is nil:\n%s", content)
	}
	if strings.Contains(content, "Superseded By:") {
		t.Fatalf("did not expect 'Superseded By' section when SupersededBy is nil:\n%s", content)
	}
}
