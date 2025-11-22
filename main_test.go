package main

import (
	"fmt"
	"strings"
	"testing"

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

func TestGetStats(t *testing.T) {
	t.Run("countsStatuses", func(t *testing.T) {
		ready := &Node{Issue: FullIssue{Title: "Ready Task", Status: "open"}}
		inProgress := &Node{
			Issue:    FullIssue{Title: "In Progress", Status: "in_progress"},
			Children: []*Node{ready},
		}
		closed := &Node{Issue: FullIssue{Title: "Closed Task", Status: "closed"}}
		blocked := &Node{Issue: FullIssue{Title: "Blocked Task", Status: "open"}, IsBlocked: true}

		m := model{
			roots: []*Node{inProgress, closed, blocked},
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
		matching := &Node{Issue: FullIssue{Title: "Alpha Ready", Status: "open"}}
		nonMatching := &Node{Issue: FullIssue{Title: "Bravo Active", Status: "in_progress"}}
		m := model{
			roots:      []*Node{matching, nonMatching},
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
}

func TestTreePrefixWidth(t *testing.T) {
	indent := "    "
	marker := " ▶"
	icon := "◐"
	id := "ab-123"

	got := treePrefixWidth(indent, marker, icon, id)
	want := lipgloss.Width(fmt.Sprintf(" %s%s %s %s ", indent, marker, icon, id))
	if got != want {
		t.Fatalf("expected %d, got %d", want, got)
	}

	icon = "🧪"
	marker = " ⛔"
	got = treePrefixWidth(indent, marker, icon, id)
	want = lipgloss.Width(fmt.Sprintf(" %s%s %s %s ", indent, marker, icon, id))
	if got != want {
		t.Fatalf("expected %d, got %d for multi-byte glyph", want, got)
	}
}
