package ui

import (
	"testing"

	"abacus/internal/beads"
	"abacus/internal/graph"
)

func TestCollectChildInfo(t *testing.T) {
	child2 := &graph.Node{Issue: beads.FullIssue{ID: "ab-2", Title: "Second"}}
	child1 := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-1", Title: "First"},
		Children: []*graph.Node{child2},
	}
	root := &graph.Node{Issue: beads.FullIssue{ID: "ab-parent"}, Children: []*graph.Node{child1}}

	infos, ids := collectChildInfo(root)
	if len(infos) != 2 {
		t.Fatalf("expected 2 child infos, got %d", len(infos))
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 ids, got %d", len(ids))
	}
	if infos[0].ID != "ab-1" || infos[1].ID != "ab-2" {
		t.Fatalf("unexpected order: %+v", infos)
	}
	if infos[1].Depth != 1 {
		t.Fatalf("expected depth 1 for nested child, got %d", infos[1].Depth)
	}
}
