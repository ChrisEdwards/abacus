package ui

import (
	"regexp"
	"strings"
	"testing"

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

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiRegexp.ReplaceAllString(s, "")
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
