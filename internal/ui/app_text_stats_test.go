package ui

import (
	"strings"
	"testing"

	"abacus/internal/beads"
	"abacus/internal/config"
	"abacus/internal/graph"

	"github.com/charmbracelet/lipgloss"
)

func TestWrapWithHangingIndent(t *testing.T) {
	t.Run("appliesIndentToWrappedLines", func(t *testing.T) {
		text := "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
		got := wrapWithHangingIndent(4, text, 20)
		lines := strings.Split(got, "\n")
		if len(lines) <= 1 {
			t.Fatalf("expected wrapped text, got %q", got)
		}
		for i := 1; i < len(lines); i++ {
			if !strings.HasPrefix(lines[i], "    ") {
				t.Fatalf("line %d missing hanging indent: %q", i, lines[i])
			}
		}
	})

	t.Run("returnsOriginalWhenTooNarrow", func(t *testing.T) {
		text := "no change"
		got := wrapWithHangingIndent(10, text, 8)
		if got != text {
			t.Fatalf("expected %q, got %q", text, got)
		}
	})

	t.Run("returnsOriginalWhenAlreadyShort", func(t *testing.T) {
		text := "short text"
		got := wrapWithHangingIndent(2, text, 50)
		if got != text {
			t.Fatalf("expected %q, got %q", text, got)
		}
	})
}

func TestIndentBlock(t *testing.T) {
	input := "first line\n\nthird line"
	got := indentBlock(input, 2)
	want := "  first line\n\n  third line"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestTruncateWithEllipsis(t *testing.T) {
	t.Run("returnsOriginalWhenFits", func(t *testing.T) {
		text := "short"
		got := truncateWithEllipsis(text, 10)
		if got != text {
			t.Fatalf("expected %q, got %q", text, got)
		}
	})

	t.Run("truncatesAndAppendsEllipsis", func(t *testing.T) {
		text := "this title should be truncated"
		got := truncateWithEllipsis(text, 12)
		if !strings.HasSuffix(got, "...") {
			t.Fatalf("expected ellipsis suffix, got %q", got)
		}
		if lipgloss.Width(got) > 12 {
			t.Fatalf("expected truncated text to fit within width, got width %d", lipgloss.Width(got))
		}
	})

	t.Run("handlesVeryNarrowWidths", func(t *testing.T) {
		if got := truncateWithEllipsis("wide", 2); got != ".." {
			t.Fatalf("expected fallback to dots for narrow width, got %q", got)
		}
	})
}

func TestBuildTreeLines_TruncatesWhenColumnsEnabled(t *testing.T) {
	previous := config.GetBool(config.KeyTreeShowColumns)
	if err := config.Set(config.KeyTreeShowColumns, true); err != nil {
		t.Fatalf("failed to set showColumns: %v", err)
	}
	t.Cleanup(func() {
		_ = config.Set(config.KeyTreeShowColumns, previous)
	})

	node := &graph.Node{Issue: beads.FullIssue{ID: "ab-111", Title: "This is a very long title that should wrap or truncate for testing purposes", Status: "open"}}
	m := App{
		visibleRows: []graph.TreeRow{{Node: node}},
		cursor:      -1,
	}

	lines, _, _ := m.buildTreeLines(30)
	if len(lines) != 1 {
		t.Fatalf("expected single line when columns enabled, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "...") {
		t.Fatalf("expected ellipsis in truncated title, got %q", lines[0])
	}

	if err := config.Set(config.KeyTreeShowColumns, false); err != nil {
		t.Fatalf("failed to disable showColumns: %v", err)
	}
	wrappedLines, _, _ := m.buildTreeLines(30)
	if len(wrappedLines) <= 1 {
		t.Fatalf("expected wrapped lines when columns disabled, got %d", len(wrappedLines))
	}
}

func TestGetStats(t *testing.T) {
	t.Run("countsStatuses", func(t *testing.T) {
		ready := &graph.Node{Issue: beads.FullIssue{ID: "ab-001", Title: "Ready Task", Status: "open"}}
		inProgress := &graph.Node{
			Issue:    beads.FullIssue{ID: "ab-002", Title: "In Progress", Status: "in_progress"},
			Children: []*graph.Node{ready},
		}
		closed := &graph.Node{Issue: beads.FullIssue{ID: "ab-003", Title: "Closed Task", Status: "closed"}}
		blocked := &graph.Node{Issue: beads.FullIssue{ID: "ab-004", Title: "Blocked Task", Status: "open"}, IsBlocked: true}

		m := App{
			roots: []*graph.Node{inProgress, closed, blocked},
		}

		stats := m.getStats()
		if stats.Total != 4 {
			t.Fatalf("expected total 4, got %d", stats.Total)
		}
		if stats.InProgress != 1 {
			t.Fatalf("expected 1 in-progress, got %d", stats.InProgress)
		}
		if stats.Closed != 1 {
			t.Fatalf("expected 1 closed, got %d", stats.Closed)
		}
		if stats.Blocked != 1 {
			t.Fatalf("expected 1 blocked, got %d", stats.Blocked)
		}
		if stats.Ready != 1 {
			t.Fatalf("expected 1 ready, got %d", stats.Ready)
		}
	})

	t.Run("appliesFilter", func(t *testing.T) {
		matching := &graph.Node{Issue: beads.FullIssue{ID: "ab-010", Title: "Alpha Ready", Status: "open"}}
		nonMatching := &graph.Node{Issue: beads.FullIssue{ID: "ab-020", Title: "Bravo Active", Status: "in_progress"}}
		m := App{
			roots:      []*graph.Node{matching, nonMatching},
			filterText: "ready",
		}

		stats := m.getStats()
		if stats.Total != 1 {
			t.Fatalf("expected only one matching issue, got %d", stats.Total)
		}
		if stats.Ready != 1 {
			t.Fatalf("expected ready count 1, got %d", stats.Ready)
		}
		if stats.InProgress != 0 {
			t.Fatalf("expected in-progress count 0, got %d", stats.InProgress)
		}
	})

	t.Run("countsMatchesByIDFilter", func(t *testing.T) {
		openNode := &graph.Node{Issue: beads.FullIssue{ID: "ab-100", Title: "Alpha Ready", Status: "open"}}
		inProgress := &graph.Node{Issue: beads.FullIssue{ID: "ab-200", Title: "Beta Active", Status: "in_progress"}}
		m := App{
			roots:      []*graph.Node{openNode, inProgress},
			filterText: "ab-200",
		}

		stats := m.getStats()
		if stats.Total != 1 {
			t.Fatalf("expected filtered count 1, got %d", stats.Total)
		}
		if stats.InProgress != 1 {
			t.Fatalf("expected in-progress count 1, got %d", stats.InProgress)
		}
		if stats.Ready != 0 {
			t.Fatalf("expected ready count 0, got %d", stats.Ready)
		}
	})

	t.Run("deduplicatesMultiParentNodes", func(t *testing.T) {
		// Task with multiple parents should only be counted once
		sharedTask := &graph.Node{Issue: beads.FullIssue{ID: "ab-shared", Title: "Shared Task", Status: "open"}}
		epic1 := &graph.Node{
			Issue:    beads.FullIssue{ID: "ab-epic1", Title: "Epic 1", Status: "open"},
			Children: []*graph.Node{sharedTask},
		}
		epic2 := &graph.Node{
			Issue:    beads.FullIssue{ID: "ab-epic2", Title: "Epic 2", Status: "open"},
			Children: []*graph.Node{sharedTask}, // Same task under another parent
		}
		sharedTask.Parents = []*graph.Node{epic1, epic2}

		m := App{
			roots: []*graph.Node{epic1, epic2},
		}

		stats := m.getStats()
		// Should count: epic1, epic2, sharedTask (once) = 3 total
		if stats.Total != 3 {
			t.Fatalf("expected total 3 (multi-parent task counted once), got %d", stats.Total)
		}
		if stats.Ready != 3 {
			t.Fatalf("expected 3 ready (all open, not blocked), got %d", stats.Ready)
		}
	})
}
