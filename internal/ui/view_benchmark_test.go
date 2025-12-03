package ui

import (
	"testing"
	"time"

	"abacus/internal/beads"
	"abacus/internal/graph"
	"abacus/internal/ui/theme"
)

func BenchmarkAppViewLayering(b *testing.B) {
	if !theme.SetTheme("dracula") {
		b.Fatal("expected dracula theme to be registered")
	}

	root := &graph.Node{
		Issue: beads.FullIssue{
			ID:     "ab-root",
			Title:  "Root Epic",
			Status: "open",
		},
		Expanded: true,
	}
	child := &graph.Node{
		Issue: beads.FullIssue{
			ID:     "ab-child",
			Title:  "Child Task",
			Status: "in_progress",
		},
		Expanded: true,
		Parents:  []*graph.Node{root},
	}
	root.Children = []*graph.Node{child}

	app := &App{
		ready:    true,
		width:    120,
		height:   40,
		repoName: "abacus",
		roots:    []*graph.Node{root},
		visibleRows: []graph.TreeRow{
			{Node: root, Depth: 0},
			{Node: child, Parent: root, Depth: 1},
		},
		activeOverlay:        OverlayStatus,
		statusOverlay:        NewStatusOverlay("ab-child", "Child Task", "in_progress"),
		statusToastVisible:   true,
		statusToastNewStatus: "in_progress",
		statusToastBeadID:    "ab-child",
		statusToastStart:     time.Now(),
		showHelp:             false,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if out := app.View(); len(out) == 0 {
			b.Fatal("View() returned empty string")
		}
	}
}
