package graph

import (
	"testing"

	"abacus/internal/beads"
)

func TestBuilderBuildSimpleGraph(t *testing.T) {
	issues := []beads.FullIssue{
		{
			ID:        "ab-1",
			Title:     "Root",
			Status:    "in_progress",
			CreatedAt: "2024-01-01T00:00:00Z",
			UpdatedAt: "2024-01-01T00:00:00Z",
			Dependents: []beads.Dependent{
				{ID: "ab-2"},
			},
		},
		{
			ID:     "ab-2",
			Title:  "Child",
			Status: "open",
			Dependencies: []beads.Dependency{
				{Type: "parent-child", TargetID: "ab-1"},
				{Type: "blocks", TargetID: "ab-4"},
			},
			CreatedAt: "2024-01-02T00:00:00Z",
			UpdatedAt: "2024-01-02T00:00:00Z",
		},
		{
			ID:        "ab-3",
			Title:     "Independent",
			Status:    "open",
			CreatedAt: "2024-01-03T00:00:00Z",
			UpdatedAt: "2024-01-03T00:00:00Z",
		},
		{
			ID:        "ab-4",
			Title:     "Blocker",
			Status:    "open",
			CreatedAt: "2024-01-04T00:00:00Z",
			UpdatedAt: "2024-01-04T00:00:00Z",
		},
	}

	b := NewBuilder()
	roots, err := b.Build(issues)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	if len(roots) != 3 {
		t.Fatalf("expected 3 roots, got %d", len(roots))
	}

	var root *Node
	for _, r := range roots {
		if r.Issue.ID == "ab-1" {
			root = r
			break
		}
	}
	if root == nil {
		t.Fatalf("root ab-1 not found")
	}

	if len(root.Children) != 1 {
		ids := []string{}
		for _, c := range root.Children {
			ids = append(ids, c.Issue.ID)
		}
		t.Fatalf("expected root to have 1 child, got %d (%v)", len(root.Children), ids)
	}
	child := root.Children[0]
	if child.Issue.ID != "ab-2" {
		t.Fatalf("expected child ab-2, got %s", child.Issue.ID)
	}
	if child.Parent != root {
		t.Fatalf("expected child parent to be root")
	}
	if !child.IsBlocked {
		t.Fatalf("expected child to be blocked")
	}
	if len(child.BlockedBy) != 1 || child.BlockedBy[0].Issue.ID != "ab-4" {
		t.Fatalf("expected child blocked by ab-4")
	}
	if len(child.BlockedBy[0].Blocks) != 1 || child.BlockedBy[0].Blocks[0] != child {
		t.Fatalf("expected blocker to reference child")
	}
	if child.Depth != 1 {
		t.Fatalf("expected child depth 1, got %d", child.Depth)
	}
	if root.HasInProgress != true {
		t.Fatalf("expected root HasInProgress true")
	}
	if root.Expanded != true {
		t.Fatalf("expected root expanded due to in-progress descendant")
	}
	if root.HasReady {
		t.Fatalf("expected root HasReady false because child blocked")
	}
}

func TestBuilderPrefersDeepestParent(t *testing.T) {
	issues := []beads.FullIssue{
		{ID: "ab-1", Title: "RootA", Status: "open", CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-01T00:00:00Z"},
		{
			ID: "ab-2", Title: "Mid", Status: "open",
			Dependencies: []beads.Dependency{{Type: "parent-child", TargetID: "ab-1"}},
			CreatedAt:    "2024-01-02T00:00:00Z",
			UpdatedAt:    "2024-01-02T00:00:00Z",
			Dependents:   []beads.Dependent{{ID: "ab-4"}},
		},
		{ID: "ab-3", Title: "RootB", Status: "open", CreatedAt: "2024-01-03T00:00:00Z", UpdatedAt: "2024-01-03T00:00:00Z"},
		{
			ID: "ab-4", Title: "Leaf", Status: "open",
			Dependencies: []beads.Dependency{
				{Type: "parent-child", TargetID: "ab-2"},
				{Type: "parent-child", TargetID: "ab-3"},
			},
			CreatedAt: "2024-01-04T00:00:00Z",
			UpdatedAt: "2024-01-04T00:00:00Z",
		},
	}

	roots, err := NewBuilder().Build(issues)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	var rootA, rootB *Node
	for _, r := range roots {
		switch r.Issue.ID {
		case "ab-1":
			rootA = r
		case "ab-3":
			rootB = r
		}
	}
	if rootA == nil || rootB == nil {
		t.Fatalf("expected both rootA and rootB in roots slice")
	}

	var mid *Node
	for _, child := range rootA.Children {
		if child.Issue.ID == "ab-2" {
			mid = child
			break
		}
	}
	if mid == nil {
		t.Fatalf("expected rootA to own mid child")
	}

	if len(mid.Children) != 1 || mid.Children[0].Issue.ID != "ab-4" {
		t.Fatalf("expected mid to own leaf child")
	}
	leaf := mid.Children[0]
	if leaf.Parent != mid {
		t.Fatalf("expected leaf parent to be mid")
	}
	if len(rootB.Children) != 0 {
		t.Fatalf("expected rootB to have no children after deepest selection")
	}
	if leaf.Depth != mid.Depth+1 {
		t.Fatalf("expected leaf depth to be parent depth + 1")
	}
}

func TestBuilderDetectsCycles(t *testing.T) {
	issues := []beads.FullIssue{
		{
			ID:    "ab-1",
			Title: "Root",
			Dependencies: []beads.Dependency{
				{Type: "parent-child", TargetID: "ab-2"},
			},
		},
		{
			ID:    "ab-2",
			Title: "Child",
			Dependencies: []beads.Dependency{
				{Type: "parent-child", TargetID: "ab-1"},
			},
		},
	}

	if _, err := NewBuilder().Build(issues); err == nil {
		t.Fatalf("expected cycle detection error")
	}
}
