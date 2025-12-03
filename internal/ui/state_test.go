package ui

import (
	"testing"

	"abacus/internal/beads"
	"abacus/internal/graph"
)

func TestRemoveNodeFromTree(t *testing.T) {
	t.Run("removes root node", func(t *testing.T) {
		app := &App{
			roots: []*graph.Node{
				{Issue: beads.FullIssue{ID: "ab-001", Title: "Root 1"}},
				{Issue: beads.FullIssue{ID: "ab-002", Title: "Root 2"}},
				{Issue: beads.FullIssue{ID: "ab-003", Title: "Root 3"}},
			},
		}

		app.removeNodeFromTree("ab-002")

		if len(app.roots) != 2 {
			t.Errorf("expected 2 roots after removal, got %d", len(app.roots))
		}
		for _, root := range app.roots {
			if root.Issue.ID == "ab-002" {
				t.Error("root ab-002 should have been removed")
			}
		}
	})

	t.Run("removes child node", func(t *testing.T) {
		child := &graph.Node{Issue: beads.FullIssue{ID: "ab-child", Title: "Child"}}
		parent := &graph.Node{
			Issue:    beads.FullIssue{ID: "ab-parent", Title: "Parent"},
			Children: []*graph.Node{child},
		}
		app := &App{roots: []*graph.Node{parent}}

		app.removeNodeFromTree("ab-child")

		if len(parent.Children) != 0 {
			t.Errorf("expected 0 children after removal, got %d", len(parent.Children))
		}
	})

	t.Run("removes multi-parent node from all parents", func(t *testing.T) {
		child := &graph.Node{Issue: beads.FullIssue{ID: "ab-shared", Title: "Shared Child"}}
		parent1 := &graph.Node{
			Issue:    beads.FullIssue{ID: "ab-p1", Title: "Parent 1"},
			Children: []*graph.Node{child},
		}
		parent2 := &graph.Node{
			Issue:    beads.FullIssue{ID: "ab-p2", Title: "Parent 2"},
			Children: []*graph.Node{child},
		}
		child.Parents = []*graph.Node{parent1, parent2}
		app := &App{roots: []*graph.Node{parent1, parent2}}

		app.removeNodeFromTree("ab-shared")

		if len(parent1.Children) != 0 {
			t.Errorf("expected 0 children in parent1 after removal, got %d", len(parent1.Children))
		}
		if len(parent2.Children) != 0 {
			t.Errorf("expected 0 children in parent2 after removal, got %d", len(parent2.Children))
		}
	})

	t.Run("removes nested node", func(t *testing.T) {
		grandchild := &graph.Node{Issue: beads.FullIssue{ID: "ab-gc", Title: "Grandchild"}}
		child := &graph.Node{
			Issue:    beads.FullIssue{ID: "ab-c", Title: "Child"},
			Children: []*graph.Node{grandchild},
		}
		root := &graph.Node{
			Issue:    beads.FullIssue{ID: "ab-r", Title: "Root"},
			Children: []*graph.Node{child},
		}
		app := &App{roots: []*graph.Node{root}}

		app.removeNodeFromTree("ab-gc")

		if len(child.Children) != 0 {
			t.Errorf("expected 0 children in child after removal, got %d", len(child.Children))
		}
	})

	t.Run("no-op for non-existent node", func(t *testing.T) {
		app := &App{
			roots: []*graph.Node{
				{Issue: beads.FullIssue{ID: "ab-001", Title: "Root 1"}},
			},
		}

		app.removeNodeFromTree("ab-nonexistent")

		if len(app.roots) != 1 {
			t.Errorf("expected 1 root (unchanged), got %d", len(app.roots))
		}
	})
}
