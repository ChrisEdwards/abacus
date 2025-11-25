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

func TestBuilderMultiParentChildrenAppearUnderAllParents(t *testing.T) {
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

	// Leaf should appear under both mid and rootB (multi-parent support)
	if len(mid.Children) != 1 || mid.Children[0].Issue.ID != "ab-4" {
		t.Fatalf("expected mid to have leaf as child")
	}
	leaf := mid.Children[0]

	// Leaf should also appear under rootB
	if len(rootB.Children) != 1 || rootB.Children[0].Issue.ID != "ab-4" {
		t.Fatalf("expected rootB to also have leaf as child (multi-parent)")
	}

	// Leaf should have 2 parents
	if len(leaf.Parents) != 2 {
		t.Fatalf("expected leaf to have 2 parents, got %d", len(leaf.Parents))
	}
}

func TestBuilderMultiParentWithThreeParents(t *testing.T) {
	issues := []beads.FullIssue{
		{ID: "ab-p1", Title: "Parent 1", Status: "open", CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-01T00:00:00Z"},
		{ID: "ab-p2", Title: "Parent 2", Status: "open", CreatedAt: "2024-01-02T00:00:00Z", UpdatedAt: "2024-01-02T00:00:00Z"},
		{ID: "ab-p3", Title: "Parent 3", Status: "open", CreatedAt: "2024-01-03T00:00:00Z", UpdatedAt: "2024-01-03T00:00:00Z"},
		{
			ID: "ab-child", Title: "Shared Child", Status: "open",
			Dependencies: []beads.Dependency{
				{Type: "parent-child", TargetID: "ab-p1"},
				{Type: "parent-child", TargetID: "ab-p2"},
				{Type: "parent-child", TargetID: "ab-p3"},
			},
			CreatedAt: "2024-01-04T00:00:00Z",
			UpdatedAt: "2024-01-04T00:00:00Z",
		},
	}

	roots, err := NewBuilder().Build(issues)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	// Find all parent nodes
	parents := make(map[string]*Node)
	for _, r := range roots {
		parents[r.Issue.ID] = r
	}

	// Find the shared child
	var child *Node
	for _, p := range parents {
		for _, c := range p.Children {
			if c.Issue.ID == "ab-child" {
				child = c
				break
			}
		}
		if child != nil {
			break
		}
	}
	if child == nil {
		t.Fatalf("child not found in any parent")
	}

	// Verify child has 3 parents
	if len(child.Parents) != 3 {
		t.Fatalf("expected child to have 3 parents, got %d", len(child.Parents))
	}

	// Verify child appears under all 3 parents
	for id, parent := range parents {
		if id == "ab-child" {
			continue
		}
		found := false
		for _, c := range parent.Children {
			if c.Issue.ID == "ab-child" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected parent %s to have ab-child in Children", id)
		}
	}
}

func TestBuilderNoDuplicateChildrenWithinParent(t *testing.T) {
	// When both Dependencies and Dependents reference the same parent-child relationship,
	// the child should not be duplicated in the parent's Children slice
	issues := []beads.FullIssue{
		{
			ID:         "ab-parent",
			Title:      "Parent",
			Status:     "open",
			CreatedAt:  "2024-01-01T00:00:00Z",
			UpdatedAt:  "2024-01-01T00:00:00Z",
			Dependents: []beads.Dependent{{ID: "ab-child"}},
		},
		{
			ID:    "ab-child",
			Title: "Child",
			Dependencies: []beads.Dependency{
				{Type: "parent-child", TargetID: "ab-parent"},
			},
			CreatedAt: "2024-01-02T00:00:00Z",
			UpdatedAt: "2024-01-02T00:00:00Z",
		},
	}

	roots, err := NewBuilder().Build(issues)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	var parent *Node
	for _, r := range roots {
		if r.Issue.ID == "ab-parent" {
			parent = r
			break
		}
	}
	if parent == nil {
		t.Fatalf("parent not found in roots")
	}

	// Should have exactly 1 child, not duplicated
	if len(parent.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(parent.Children))
	}
	if parent.Children[0].Issue.ID != "ab-child" {
		t.Fatalf("expected child ab-child, got %s", parent.Children[0].Issue.ID)
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

func TestComputeSortMetricsAggregatesDescendantStatus(t *testing.T) {
	child := &Node{Issue: beads.FullIssue{ID: "ab-020", Status: "in_progress", UpdatedAt: "2024-01-02T00:00:00Z"}}
	parent := &Node{Issue: beads.FullIssue{ID: "ab-010", Status: "open", CreatedAt: "2024-01-01T00:00:00Z"}, Children: []*Node{child}}
	computeSortMetrics(parent)
	if parent.SortPriority != sortPriorityInProgress {
		t.Fatalf("expected parent priority to cascade to in-progress, got %d", parent.SortPriority)
	}
	if !parent.SortTimestamp.Equal(child.SortTimestamp) {
		t.Fatalf("expected parent timestamp to match earliest in-progress descendant")
	}
}

func TestSortNodesOrdersByPriorityAndDate(t *testing.T) {
	readyOld := &Node{Issue: beads.FullIssue{ID: "ab-101", Status: "open", CreatedAt: "2024-01-01T00:00:00Z"}}
	readyNew := &Node{Issue: beads.FullIssue{ID: "ab-102", Status: "open", CreatedAt: "2024-02-01T00:00:00Z"}}
	blocked := &Node{Issue: beads.FullIssue{ID: "ab-103", Status: "open", CreatedAt: "2024-01-15T00:00:00Z"}, IsBlocked: true}
	parent := &Node{Issue: beads.FullIssue{ID: "ab-100", Status: "open", CreatedAt: "2024-01-01T00:00:00Z"}, Children: []*Node{blocked, readyNew, readyOld}}
	computeSortMetrics(parent)
	got := []string{parent.Children[0].Issue.ID, parent.Children[1].Issue.ID, parent.Children[2].Issue.ID}
	want := []string{"ab-101", "ab-102", "ab-103"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("child order mismatch: got %v want %v", got, want)
		}
	}
}

func TestBuildSortsRootsByCascadingPriority(t *testing.T) {
	issues := []beads.FullIssue{
		{
			ID:        "ab-201",
			Title:     "ParentActive",
			Status:    "open",
			CreatedAt: "2024-01-01T00:00:00Z",
			Dependents: []beads.Dependent{
				{ID: "ab-202"},
			},
		},
		{
			ID:        "ab-202",
			Title:     "Child In Progress",
			Status:    "in_progress",
			CreatedAt: "2024-01-02T00:00:00Z",
			UpdatedAt: "2024-01-03T00:00:00Z",
			Dependencies: []beads.Dependency{
				{Type: "parent-child", TargetID: "ab-201"},
			},
		},
		{
			ID:        "ab-203",
			Title:     "ParentReady",
			Status:    "open",
			CreatedAt: "2023-12-01T00:00:00Z",
		},
		{
			ID:        "ab-204",
			Title:     "ParentClosed",
			Status:    "closed",
			CreatedAt: "2024-01-10T00:00:00Z",
			ClosedAt:  "2024-01-11T00:00:00Z",
		},
	}
	roots, err := NewBuilder().Build(issues)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if len(roots) != 3 {
		t.Fatalf("expected 3 roots, got %d", len(roots))
	}
	order := []string{roots[0].Issue.ID, roots[1].Issue.ID, roots[2].Issue.ID}
	want := []string{"ab-201", "ab-203", "ab-204"}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("root order mismatch: got %v want %v", order, want)
		}
	}
}

func TestBuildSortsChildrenByOldestRelevantDate(t *testing.T) {
	issues := []beads.FullIssue{
		{
			ID:        "ab-301",
			Title:     "Parent",
			Status:    "open",
			CreatedAt: "2024-01-01T00:00:00Z",
			Dependents: []beads.Dependent{
				{ID: "ab-302"},
				{ID: "ab-303"},
			},
		},
		{
			ID:        "ab-302",
			Title:     "Ready Old",
			Status:    "open",
			CreatedAt: "2023-12-01T00:00:00Z",
			Dependencies: []beads.Dependency{
				{Type: "parent-child", TargetID: "ab-301"},
			},
		},
		{
			ID:        "ab-303",
			Title:     "Ready New",
			Status:    "open",
			CreatedAt: "2024-02-01T00:00:00Z",
			Dependencies: []beads.Dependency{
				{Type: "parent-child", TargetID: "ab-301"},
			},
		},
	}
	roots, err := NewBuilder().Build(issues)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	var parent *Node
	for _, r := range roots {
		if r.Issue.ID == "ab-301" {
			parent = r
			break
		}
	}
	if parent == nil {
		t.Fatalf("parent node not found")
	}
	if len(parent.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(parent.Children))
	}
	if parent.Children[0].Issue.ID != "ab-302" {
		t.Fatalf("expected oldest ready child first, got %s", parent.Children[0].Issue.ID)
	}
}
