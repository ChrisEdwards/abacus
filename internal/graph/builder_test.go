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
				{ID: "ab-2", Type: "parent-child"},
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
			Dependents:   []beads.Dependent{{ID: "ab-4", Type: "parent-child"}},
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
			Dependents: []beads.Dependent{{ID: "ab-child", Type: "parent-child"}},
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
				{ID: "ab-202", Type: "parent-child"},
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
				{ID: "ab-302", Type: "parent-child"},
				{ID: "ab-303", Type: "parent-child"},
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

func TestBuilderIgnoresNonParentChildDependents(t *testing.T) {
	// Dependents with type != "parent-child" should not create parent relationships.
	// Only "parent-child" types should appear in the "Part Of" section.
	issues := []beads.FullIssue{
		{
			ID:        "ab-blocker",
			Title:     "Blocker Issue",
			Status:    "open",
			CreatedAt: "2024-01-01T00:00:00Z",
			UpdatedAt: "2024-01-01T00:00:00Z",
			// This issue has ab-blocked as a dependent with "blocks" type
			// This should NOT make ab-blocker a parent of ab-blocked
			Dependents: []beads.Dependent{
				{ID: "ab-blocked", Type: "blocks"},
			},
		},
		{
			ID:        "ab-blocked",
			Title:     "Blocked Issue",
			Status:    "open",
			CreatedAt: "2024-01-02T00:00:00Z",
			UpdatedAt: "2024-01-02T00:00:00Z",
			Dependencies: []beads.Dependency{
				{Type: "blocks", TargetID: "ab-blocker"},
			},
		},
		{
			ID:        "ab-parent",
			Title:     "Real Parent",
			Status:    "open",
			CreatedAt: "2024-01-03T00:00:00Z",
			UpdatedAt: "2024-01-03T00:00:00Z",
			// This is a proper parent-child relationship
			Dependents: []beads.Dependent{
				{ID: "ab-child", Type: "parent-child"},
			},
		},
		{
			ID:        "ab-child",
			Title:     "Child Issue",
			Status:    "open",
			CreatedAt: "2024-01-04T00:00:00Z",
			UpdatedAt: "2024-01-04T00:00:00Z",
			Dependencies: []beads.Dependency{
				{Type: "parent-child", TargetID: "ab-parent"},
			},
		},
	}

	roots, err := NewBuilder().Build(issues)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	// Find the blocked issue
	var blocked *Node
	for _, r := range roots {
		if r.Issue.ID == "ab-blocked" {
			blocked = r
			break
		}
	}
	if blocked == nil {
		t.Fatalf("blocked issue not found in roots")
	}

	// blocked should have NO parents (the "blocks" dependent should be ignored)
	if len(blocked.Parents) != 0 {
		parentIDs := make([]string, len(blocked.Parents))
		for i, p := range blocked.Parents {
			parentIDs[i] = p.Issue.ID
		}
		t.Fatalf("expected blocked to have 0 parents, got %d: %v", len(blocked.Parents), parentIDs)
	}

	// Find the child issue
	var child *Node
	for _, r := range roots {
		for _, c := range r.Children {
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
		t.Fatalf("child issue not found")
	}

	// child should have exactly 1 parent (ab-parent)
	if len(child.Parents) != 1 {
		t.Fatalf("expected child to have 1 parent, got %d", len(child.Parents))
	}
	if child.Parents[0].Issue.ID != "ab-parent" {
		t.Fatalf("expected child's parent to be ab-parent, got %s", child.Parents[0].Issue.ID)
	}
}

func TestBuilderRelatedRelationships(t *testing.T) {
	issues := []beads.FullIssue{
		{
			ID:        "ab-a",
			Title:     "Issue A",
			Status:    "open",
			CreatedAt: "2024-01-01T00:00:00Z",
			UpdatedAt: "2024-01-01T00:00:00Z",
			Dependencies: []beads.Dependency{
				{Type: "related", TargetID: "ab-b"},
			},
		},
		{
			ID:        "ab-b",
			Title:     "Issue B",
			Status:    "open",
			CreatedAt: "2024-01-02T00:00:00Z",
			UpdatedAt: "2024-01-02T00:00:00Z",
		},
	}

	roots, err := NewBuilder().Build(issues)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	var nodeA, nodeB *Node
	for _, r := range roots {
		switch r.Issue.ID {
		case "ab-a":
			nodeA = r
		case "ab-b":
			nodeB = r
		}
	}
	if nodeA == nil || nodeB == nil {
		t.Fatalf("expected both nodes in roots")
	}

	// Related is bidirectional - both should reference each other
	if len(nodeA.Related) != 1 || nodeA.Related[0].Issue.ID != "ab-b" {
		t.Fatalf("expected nodeA.Related to contain ab-b")
	}
	if len(nodeB.Related) != 1 || nodeB.Related[0].Issue.ID != "ab-a" {
		t.Fatalf("expected nodeB.Related to contain ab-a")
	}

	// Related should NOT affect parent-child or tree structure
	if len(nodeA.Parents) != 0 {
		t.Fatalf("expected nodeA to have no parents from related relationship")
	}
	if len(nodeB.Children) != 0 {
		t.Fatalf("expected nodeB to have no children from related relationship")
	}
}

func TestBuilderDiscoveredFromRelationships(t *testing.T) {
	issues := []beads.FullIssue{
		{
			ID:        "ab-discovered",
			Title:     "Discovered Issue",
			Status:    "open",
			CreatedAt: "2024-01-02T00:00:00Z",
			UpdatedAt: "2024-01-02T00:00:00Z",
			Dependencies: []beads.Dependency{
				{Type: "discovered-from", TargetID: "ab-source"},
			},
		},
		{
			ID:        "ab-source",
			Title:     "Source Issue",
			Status:    "open",
			CreatedAt: "2024-01-01T00:00:00Z",
			UpdatedAt: "2024-01-01T00:00:00Z",
		},
	}

	roots, err := NewBuilder().Build(issues)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	var discovered *Node
	for _, r := range roots {
		if r.Issue.ID == "ab-discovered" {
			discovered = r
			break
		}
	}
	if discovered == nil {
		t.Fatalf("discovered node not found")
	}

	// DiscoveredFrom should link to the source
	if len(discovered.DiscoveredFrom) != 1 || discovered.DiscoveredFrom[0].Issue.ID != "ab-source" {
		t.Fatalf("expected discovered.DiscoveredFrom to contain ab-source")
	}

	// DiscoveredFrom should NOT affect parent-child or tree structure
	if len(discovered.Parents) != 0 {
		t.Fatalf("expected discovered to have no parents from discovered-from relationship")
	}
}

func TestBuilderDuplicateOfResolution(t *testing.T) {
	issues := []beads.FullIssue{
		{
			ID:        "ab-dup",
			Title:     "Duplicate Issue",
			Status:    "closed",
			CreatedAt: "2024-01-02T00:00:00Z",
			UpdatedAt: "2024-01-02T00:00:00Z",
			// DuplicateOf is now indicated via "duplicates" dependency type
			Dependencies: []beads.Dependency{{TargetID: "ab-canonical", Type: "duplicates"}},
		},
		{
			ID:        "ab-canonical",
			Title:     "Canonical Issue",
			Status:    "open",
			CreatedAt: "2024-01-01T00:00:00Z",
			UpdatedAt: "2024-01-01T00:00:00Z",
		},
	}

	roots, err := NewBuilder().Build(issues)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	var dup *Node
	for _, r := range roots {
		if r.Issue.ID == "ab-dup" {
			dup = r
			break
		}
	}
	if dup == nil {
		t.Fatalf("duplicate node not found")
	}

	// DuplicateOf should be resolved to canonical node
	if dup.DuplicateOf == nil {
		t.Fatalf("expected DuplicateOf to be resolved")
	}
	if dup.DuplicateOf.Issue.ID != "ab-canonical" {
		t.Fatalf("expected DuplicateOf to point to ab-canonical, got %s", dup.DuplicateOf.Issue.ID)
	}

	// DuplicateOf should NOT affect parent-child relationship
	if len(dup.Parents) != 0 {
		t.Fatalf("expected dup to have no parents from duplicate relationship")
	}
}

func TestBuilderSupersededByResolution(t *testing.T) {
	issues := []beads.FullIssue{
		{
			ID:        "ab-old",
			Title:     "Old Issue",
			Status:    "closed",
			CreatedAt: "2024-01-01T00:00:00Z",
			UpdatedAt: "2024-01-01T00:00:00Z",
		},
		{
			ID:        "ab-new",
			Title:     "New Issue",
			Status:    "open",
			CreatedAt: "2024-01-02T00:00:00Z",
			UpdatedAt: "2024-01-02T00:00:00Z",
			// SupersededBy is now indicated via "supersedes" dependency type
			// ab-new supersedes ab-old (the NEW one has the dependency pointing to OLD)
			Dependencies: []beads.Dependency{{TargetID: "ab-old", Type: "supersedes"}},
		},
	}

	roots, err := NewBuilder().Build(issues)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	var old *Node
	for _, r := range roots {
		if r.Issue.ID == "ab-old" {
			old = r
			break
		}
	}
	if old == nil {
		t.Fatalf("old node not found")
	}

	// SupersededBy should be resolved to replacement node
	if old.SupersededBy == nil {
		t.Fatalf("expected SupersededBy to be resolved")
	}
	if old.SupersededBy.Issue.ID != "ab-new" {
		t.Fatalf("expected SupersededBy to point to ab-new, got %s", old.SupersededBy.Issue.ID)
	}

	// SupersededBy should NOT affect parent-child relationship
	if len(old.Parents) != 0 {
		t.Fatalf("expected old to have no parents from supersede relationship")
	}
}

func TestBuilderGraphLinksWithMissingTarget(t *testing.T) {
	// When "duplicates" or "supersedes" dependency targets are not in the current view,
	// the pointers should remain nil (graceful degradation)
	issues := []beads.FullIssue{
		{
			ID:        "ab-orphan",
			Title:     "Orphan Issue",
			Status:    "closed",
			CreatedAt: "2024-01-01T00:00:00Z",
			UpdatedAt: "2024-01-01T00:00:00Z",
			// Dependencies pointing to issues not in the issue list
			Dependencies: []beads.Dependency{
				{TargetID: "ab-missing-canonical", Type: "duplicates"},
			},
		},
		{
			ID:        "ab-superseder",
			Title:     "Superseder Issue",
			Status:    "open",
			CreatedAt: "2024-01-02T00:00:00Z",
			UpdatedAt: "2024-01-02T00:00:00Z",
			// Supersedes a missing target
			Dependencies: []beads.Dependency{
				{TargetID: "ab-missing-obsolete", Type: "supersedes"},
			},
		},
	}

	roots, err := NewBuilder().Build(issues)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	var orphan *Node
	for _, r := range roots {
		if r.Issue.ID == "ab-orphan" {
			orphan = r
			break
		}
	}
	if orphan == nil {
		t.Fatalf("orphan node not found")
	}

	// Pointers should be nil when target is not found
	if orphan.DuplicateOf != nil {
		t.Fatalf("expected DuplicateOf to be nil when target missing")
	}
	if orphan.SupersededBy != nil {
		t.Fatalf("expected SupersededBy to be nil when target missing")
	}
}

func TestBuilderGraphLinksBackwardCompatibility(t *testing.T) {
	// Issues without duplicate/supersede dependencies should have nil pointers
	issues := []beads.FullIssue{
		{
			ID:           "ab-normal",
			Title:        "Normal Issue",
			Status:       "open",
			CreatedAt:    "2024-01-01T00:00:00Z",
			UpdatedAt:    "2024-01-01T00:00:00Z",
			Dependencies: []beads.Dependency{}, // No duplicates/supersedes
		},
	}

	roots, err := NewBuilder().Build(issues)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	var normal *Node
	for _, r := range roots {
		if r.Issue.ID == "ab-normal" {
			normal = r
			break
		}
	}
	if normal == nil {
		t.Fatalf("normal node not found")
	}

	// Pointers should be nil when no duplicates/supersedes dependencies
	if normal.DuplicateOf != nil {
		t.Fatalf("expected DuplicateOf to be nil when no duplicates dependency")
	}
	if normal.SupersededBy != nil {
		t.Fatalf("expected SupersededBy to be nil when not superseded")
	}
}

func TestBuilderHandlesRelatesToDependency(t *testing.T) {
	// relates-to is bidirectional (beads creates both edges via `bd relate`)
	// The builder should handle the dedup correctly
	issues := []beads.FullIssue{
		{
			ID:        "ab-a",
			Title:     "Issue A",
			Status:    "open",
			CreatedAt: "2024-01-01T00:00:00Z",
			UpdatedAt: "2024-01-01T00:00:00Z",
			Dependencies: []beads.Dependency{
				{Type: "relates-to", TargetID: "ab-b"},
			},
		},
		{
			ID:        "ab-b",
			Title:     "Issue B",
			Status:    "open",
			CreatedAt: "2024-01-02T00:00:00Z",
			UpdatedAt: "2024-01-02T00:00:00Z",
			Dependencies: []beads.Dependency{
				{Type: "relates-to", TargetID: "ab-a"},
			},
		},
	}

	roots, err := NewBuilder().Build(issues)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	var nodeA, nodeB *Node
	for _, r := range roots {
		switch r.Issue.ID {
		case "ab-a":
			nodeA = r
		case "ab-b":
			nodeB = r
		}
	}
	if nodeA == nil || nodeB == nil {
		t.Fatalf("expected both nodes in roots")
	}

	// relates-to should populate Related bidirectionally
	if len(nodeA.Related) != 1 || nodeA.Related[0].Issue.ID != "ab-b" {
		t.Fatalf("expected nodeA.Related to contain ab-b, got %v", nodeA.Related)
	}
	if len(nodeB.Related) != 1 || nodeB.Related[0].Issue.ID != "ab-a" {
		t.Fatalf("expected nodeB.Related to contain ab-a, got %v", nodeB.Related)
	}

	// Should NOT create duplicates (both edges stored, but dedup should prevent double-linking)
	// This is the key difference from "related" - relates-to already has both edges stored
}

func TestBuilderHandlesMixedRelatedTypes(t *testing.T) {
	// Mix of legacy "related" and new "relates-to" types should both work
	issues := []beads.FullIssue{
		{
			ID:        "ab-a",
			Title:     "Issue A (legacy related to B)",
			Status:    "open",
			CreatedAt: "2024-01-01T00:00:00Z",
			UpdatedAt: "2024-01-01T00:00:00Z",
			Dependencies: []beads.Dependency{
				{Type: "related", TargetID: "ab-b"},
			},
		},
		{
			ID:        "ab-b",
			Title:     "Issue B",
			Status:    "open",
			CreatedAt: "2024-01-02T00:00:00Z",
			UpdatedAt: "2024-01-02T00:00:00Z",
		},
		{
			ID:        "ab-c",
			Title:     "Issue C (relates-to D)",
			Status:    "open",
			CreatedAt: "2024-01-03T00:00:00Z",
			UpdatedAt: "2024-01-03T00:00:00Z",
			Dependencies: []beads.Dependency{
				{Type: "relates-to", TargetID: "ab-d"},
			},
		},
		{
			ID:        "ab-d",
			Title:     "Issue D (relates-to C)",
			Status:    "open",
			CreatedAt: "2024-01-04T00:00:00Z",
			UpdatedAt: "2024-01-04T00:00:00Z",
			Dependencies: []beads.Dependency{
				{Type: "relates-to", TargetID: "ab-c"},
			},
		},
	}

	roots, err := NewBuilder().Build(issues)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	nodeMap := make(map[string]*Node)
	for _, r := range roots {
		nodeMap[r.Issue.ID] = r
	}

	// Legacy related: A<->B should be bidirectional
	nodeA := nodeMap["ab-a"]
	nodeB := nodeMap["ab-b"]
	if len(nodeA.Related) != 1 || nodeA.Related[0].Issue.ID != "ab-b" {
		t.Fatalf("expected nodeA.Related to contain ab-b")
	}
	if len(nodeB.Related) != 1 || nodeB.Related[0].Issue.ID != "ab-a" {
		t.Fatalf("expected nodeB.Related to contain ab-a")
	}

	// New relates-to: C<->D should also be bidirectional (with dedup)
	nodeC := nodeMap["ab-c"]
	nodeD := nodeMap["ab-d"]
	if len(nodeC.Related) != 1 || nodeC.Related[0].Issue.ID != "ab-d" {
		t.Fatalf("expected nodeC.Related to contain ab-d")
	}
	if len(nodeD.Related) != 1 || nodeD.Related[0].Issue.ID != "ab-c" {
		t.Fatalf("expected nodeD.Related to contain ab-c")
	}
}

func TestNodeSelfSortKeyBlockedDeferred(t *testing.T) {
	tests := []struct {
		name             string
		status           string
		isBlocked        bool
		expectedPriority int
	}{
		{"in_progress", "in_progress", false, sortPriorityInProgress},
		{"open_ready", "open", false, sortPriorityReady},
		{"open_blocked_by_deps", "open", true, sortPriorityBlocked},
		{"blocked_status", "blocked", false, sortPriorityBlocked},
		{"deferred_status", "deferred", false, sortPriorityDeferred},
		{"closed", "closed", false, sortPriorityClosed},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			node := &Node{
				Issue:     beads.FullIssue{ID: "ab-test", Status: tc.status, CreatedAt: "2024-01-01T00:00:00Z"},
				IsBlocked: tc.isBlocked,
			}
			priority, _ := NodeSelfSortKey(node)
			if priority != tc.expectedPriority {
				t.Errorf("NodeSelfSortKey(%s, isBlocked=%v) = %d, want %d",
					tc.status, tc.isBlocked, priority, tc.expectedPriority)
			}
		})
	}
}

func TestSortOrderBlockedDeferredClosed(t *testing.T) {
	// Test that the full sort order is: in_progress → ready → blocked → deferred → closed
	inProgress := &Node{Issue: beads.FullIssue{ID: "ab-1", Status: "in_progress", CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-01T00:00:00Z"}}
	ready := &Node{Issue: beads.FullIssue{ID: "ab-2", Status: "open", CreatedAt: "2024-01-01T00:00:00Z"}}
	blockedByDeps := &Node{Issue: beads.FullIssue{ID: "ab-3", Status: "open", CreatedAt: "2024-01-01T00:00:00Z"}, IsBlocked: true}
	blockedStatus := &Node{Issue: beads.FullIssue{ID: "ab-4", Status: "blocked", CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-01T00:00:00Z"}}
	deferred := &Node{Issue: beads.FullIssue{ID: "ab-5", Status: "deferred", CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-01T00:00:00Z"}}
	closed := &Node{Issue: beads.FullIssue{ID: "ab-6", Status: "closed", CreatedAt: "2024-01-01T00:00:00Z", ClosedAt: "2024-01-01T00:00:00Z"}}

	// Shuffle them in wrong order
	parent := &Node{
		Issue:    beads.FullIssue{ID: "ab-0", Status: "open", CreatedAt: "2024-01-01T00:00:00Z"},
		Children: []*Node{closed, deferred, blockedStatus, blockedByDeps, ready, inProgress},
	}

	computeSortMetrics(parent)

	// Verify sort order
	wantOrder := []string{"ab-1", "ab-2", "ab-3", "ab-4", "ab-5", "ab-6"}
	for i, want := range wantOrder {
		if parent.Children[i].Issue.ID != want {
			got := make([]string, len(parent.Children))
			for j, c := range parent.Children {
				got[j] = c.Issue.ID
			}
			t.Fatalf("sort order mismatch at position %d: got %v, want %v", i, got, wantOrder)
		}
	}
}
