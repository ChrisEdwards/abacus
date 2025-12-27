package ui

import (
	"testing"
	"time"

	"abacus/internal/beads"
	"abacus/internal/graph"
)

func TestConstructNodeFromIssue(t *testing.T) {
	tests := []struct {
		name              string
		issue             beads.FullIssue
		wantPriority      int
		wantIsBlocked     bool
		wantHasReady      bool
		wantHasInProgress bool
	}{
		{
			name: "in_progress status",
			issue: beads.FullIssue{
				ID:        "ab-123",
				Title:     "Test task",
				Status:    "in_progress",
				Priority:  2,
				CreatedAt: "2025-12-01T10:00:00Z",
				UpdatedAt: "2025-12-01T11:00:00Z",
			},
			wantPriority:      1, // sortPriorityInProgress
			wantIsBlocked:     false,
			wantHasReady:      false,
			wantHasInProgress: true,
		},
		{
			name: "open status (not blocked)",
			issue: beads.FullIssue{
				ID:        "ab-456",
				Title:     "Another task",
				Status:    "open",
				Priority:  2,
				CreatedAt: "2025-12-01T10:00:00Z",
			},
			wantPriority:      2, // sortPriorityReady
			wantIsBlocked:     false,
			wantHasReady:      true,
			wantHasInProgress: false,
		},
		{
			name: "closed status",
			issue: beads.FullIssue{
				ID:        "ab-789",
				Title:     "Closed task",
				Status:    "closed",
				Priority:  2,
				CreatedAt: "2025-12-01T10:00:00Z",
				ClosedAt:  "2025-12-01T12:00:00Z",
			},
			wantPriority:      5, // sortPriorityClosed
			wantIsBlocked:     false,
			wantHasReady:      false,
			wantHasInProgress: false,
		},
		{
			name: "blocked status",
			issue: beads.FullIssue{
				ID:        "ab-blocked",
				Title:     "Blocked task",
				Status:    "blocked",
				Priority:  2,
				CreatedAt: "2025-12-01T10:00:00Z",
			},
			wantPriority:      3, // sortPriorityBlocked
			wantIsBlocked:     false,
			wantHasReady:      false,
			wantHasInProgress: false,
		},
		{
			name: "deferred status",
			issue: beads.FullIssue{
				ID:        "ab-deferred",
				Title:     "Deferred task",
				Status:    "deferred",
				Priority:  2,
				CreatedAt: "2025-12-01T10:00:00Z",
			},
			wantPriority:      4, // sortPriorityDeferred
			wantIsBlocked:     false,
			wantHasReady:      false,
			wantHasInProgress: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := constructNodeFromIssue(tt.issue)

			if node == nil {
				t.Fatal("constructNodeFromIssue returned nil")
			}

			if node.Issue.ID != tt.issue.ID {
				t.Errorf("Issue.ID = %v, want %v", node.Issue.ID, tt.issue.ID)
			}

			if node.SortPriority != tt.wantPriority {
				t.Errorf("SortPriority = %v, want %v", node.SortPriority, tt.wantPriority)
			}

			if node.IsBlocked != tt.wantIsBlocked {
				t.Errorf("IsBlocked = %v, want %v", node.IsBlocked, tt.wantIsBlocked)
			}

			if node.HasReady != tt.wantHasReady {
				t.Errorf("HasReady = %v, want %v", node.HasReady, tt.wantHasReady)
			}

			if node.HasInProgress != tt.wantHasInProgress {
				t.Errorf("HasInProgress = %v, want %v", node.HasInProgress, tt.wantHasInProgress)
			}

			// Check that collections are initialized
			if node.Children == nil {
				t.Error("Children is nil, expected empty slice")
			}
			if node.Parents == nil {
				t.Error("Parents is nil, expected empty slice")
			}
		})
	}
}

func TestFindInsertPosition(t *testing.T) {
	// Create test nodes with different priorities and timestamps
	node1 := &graph.Node{
		Issue:         beads.FullIssue{ID: "ab-1"},
		SortPriority:  1, // in_progress
		SortTimestamp: time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC),
	}
	node2 := &graph.Node{
		Issue:         beads.FullIssue{ID: "ab-2"},
		SortPriority:  2, // ready
		SortTimestamp: time.Date(2025, 12, 1, 9, 0, 0, 0, time.UTC),
	}
	node3 := &graph.Node{
		Issue:         beads.FullIssue{ID: "ab-3"},
		SortPriority:  2, // ready
		SortTimestamp: time.Date(2025, 12, 1, 11, 0, 0, 0, time.UTC),
	}
	node4 := &graph.Node{
		Issue:         beads.FullIssue{ID: "ab-4"},
		SortPriority:  3, // open
		SortTimestamp: time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name    string
		nodes   []*graph.Node
		newNode *graph.Node
		wantPos int
	}{
		{
			name:    "insert at beginning (highest priority)",
			nodes:   []*graph.Node{node2, node3, node4},
			newNode: node1,
			wantPos: 0,
		},
		{
			name:    "insert in middle (same priority, earlier timestamp)",
			nodes:   []*graph.Node{node1, node3, node4},
			newNode: node2,
			wantPos: 1,
		},
		{
			name:    "insert at end (lowest priority)",
			nodes:   []*graph.Node{node1, node2, node3},
			newNode: node4,
			wantPos: 3,
		},
		{
			name:    "insert into empty list",
			nodes:   []*graph.Node{},
			newNode: node1,
			wantPos: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos := findInsertPosition(tt.nodes, tt.newNode)
			if pos != tt.wantPos {
				t.Errorf("findInsertPosition() = %v, want %v", pos, tt.wantPos)
			}
		})
	}
}

func TestInsertNodeIntoParent(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (*App, *graph.Node, string)
		wantErr  bool
		validate func(*testing.T, *App, *graph.Node)
	}{
		{
			name: "insert as root",
			setup: func() (*App, *graph.Node, string) {
				app := &App{
					roots: []*graph.Node{},
				}
				newNode := &graph.Node{
					Issue:         beads.FullIssue{ID: "ab-new"},
					SortPriority:  2,
					SortTimestamp: time.Now(),
				}
				return app, newNode, ""
			},
			wantErr: false,
			validate: func(t *testing.T, app *App, newNode *graph.Node) {
				if len(app.roots) != 1 {
					t.Errorf("len(roots) = %v, want 1", len(app.roots))
				}
				if len(app.roots) > 0 && app.roots[0] != newNode {
					t.Error("newNode not found in roots")
				}
			},
		},
		{
			name: "insert as child of existing node",
			setup: func() (*App, *graph.Node, string) {
				parent := &graph.Node{
					Issue:     beads.FullIssue{ID: "ab-parent"},
					Children:  []*graph.Node{},
					Depth:     0,
					TreeDepth: 0,
				}
				app := &App{
					roots: []*graph.Node{parent},
				}
				newNode := &graph.Node{
					Issue:         beads.FullIssue{ID: "ab-child"},
					SortPriority:  2,
					SortTimestamp: time.Now(),
				}
				return app, newNode, "ab-parent"
			},
			wantErr: false,
			validate: func(t *testing.T, app *App, newNode *graph.Node) {
				parent := app.roots[0]
				if len(parent.Children) != 1 {
					t.Errorf("len(parent.Children) = %v, want 1", len(parent.Children))
				}
				if len(parent.Children) > 0 && parent.Children[0] != newNode {
					t.Error("newNode not found in parent.Children")
				}
				if newNode.Parent != parent {
					t.Error("newNode.Parent not set correctly")
				}
				if newNode.Depth != parent.Depth+1 {
					t.Errorf("newNode.Depth = %v, want %v", newNode.Depth, parent.Depth+1)
				}
			},
		},
		{
			name: "parent not found",
			setup: func() (*App, *graph.Node, string) {
				app := &App{
					roots: []*graph.Node{},
				}
				newNode := &graph.Node{
					Issue: beads.FullIssue{ID: "ab-orphan"},
				}
				return app, newNode, "ab-nonexistent"
			},
			wantErr: true,
			validate: func(t *testing.T, app *App, newNode *graph.Node) {
				// Error case - nothing to validate
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, newNode, parentID := tt.setup()
			err := app.insertNodeIntoParent(newNode, parentID)

			if (err != nil) != tt.wantErr {
				t.Errorf("insertNodeIntoParent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				tt.validate(t, app, newNode)
			}
		})
	}
}

func TestPropagateStateChanges(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *graph.Node
		validate func(*testing.T, *graph.Node)
	}{
		{
			name: "propagate in_progress status",
			setup: func() *graph.Node {
				grandparent := &graph.Node{
					Issue:         beads.FullIssue{ID: "ab-gp"},
					HasInProgress: false,
					Expanded:      false,
				}
				parent := &graph.Node{
					Issue:         beads.FullIssue{ID: "ab-p"},
					Parent:        grandparent,
					HasInProgress: false,
					Expanded:      false,
				}
				grandparent.Children = []*graph.Node{parent}

				child := &graph.Node{
					Issue:     beads.FullIssue{ID: "ab-c", Status: "in_progress"},
					Parent:    parent,
					IsBlocked: false,
				}
				parent.Children = []*graph.Node{child}

				return child
			},
			validate: func(t *testing.T, child *graph.Node) {
				parent := child.Parent
				if !parent.HasInProgress {
					t.Error("parent.HasInProgress should be true")
				}
				if !parent.Expanded {
					t.Error("parent.Expanded should be true (auto-expanded)")
				}

				grandparent := parent.Parent
				if !grandparent.HasInProgress {
					t.Error("grandparent.HasInProgress should be true (propagated)")
				}
			},
		},
		{
			name: "propagate ready status",
			setup: func() *graph.Node {
				parent := &graph.Node{
					Issue:    beads.FullIssue{ID: "ab-p"},
					HasReady: false,
				}
				child := &graph.Node{
					Issue:     beads.FullIssue{ID: "ab-c", Status: "open"},
					Parent:    parent,
					IsBlocked: false,
				}
				parent.Children = []*graph.Node{child}

				return child
			},
			validate: func(t *testing.T, child *graph.Node) {
				parent := child.Parent
				if !parent.HasReady {
					t.Error("parent.HasReady should be true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := tt.setup()
			propagateStateChanges(node)
			tt.validate(t, node)
		})
	}
}

func TestPropagateStateChangesSortPriority(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *graph.Node
		validate func(*testing.T, *graph.Node)
	}{
		{
			name: "in_progress child bubbles up sort priority to open parent",
			setup: func() *graph.Node {
				childTs := time.Date(2025, 12, 1, 11, 0, 0, 0, time.UTC)
				parent := &graph.Node{
					Issue:         beads.FullIssue{ID: "ab-p", Status: "open"},
					SortPriority:  3, // open
					SortTimestamp: time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC),
				}
				child := &graph.Node{
					Issue:         beads.FullIssue{ID: "ab-c", Status: "in_progress"},
					Parent:        parent,
					SortPriority:  1, // in_progress
					SortTimestamp: childTs,
				}
				parent.Children = []*graph.Node{child}
				return child
			},
			validate: func(t *testing.T, child *graph.Node) {
				parent := child.Parent
				if parent.SortPriority != 1 {
					t.Errorf("parent.SortPriority = %d, want 1 (in_progress)", parent.SortPriority)
				}
				if !parent.SortTimestamp.Equal(child.SortTimestamp) {
					t.Error("parent.SortTimestamp should match child's timestamp")
				}
			},
		},
		{
			name: "ready child bubbles up sort priority to open parent",
			setup: func() *graph.Node {
				childTs := time.Date(2025, 12, 1, 9, 0, 0, 0, time.UTC)
				parent := &graph.Node{
					Issue:         beads.FullIssue{ID: "ab-p", Status: "open"},
					SortPriority:  3, // open (blocked or just open)
					SortTimestamp: time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC),
				}
				child := &graph.Node{
					Issue:         beads.FullIssue{ID: "ab-c", Status: "open"},
					Parent:        parent,
					IsBlocked:     false,
					SortPriority:  2, // ready
					SortTimestamp: childTs,
				}
				parent.Children = []*graph.Node{child}
				return child
			},
			validate: func(t *testing.T, child *graph.Node) {
				parent := child.Parent
				if parent.SortPriority != 2 {
					t.Errorf("parent.SortPriority = %d, want 2 (ready)", parent.SortPriority)
				}
				if !parent.SortTimestamp.Equal(child.SortTimestamp) {
					t.Error("parent.SortTimestamp should match child's timestamp")
				}
			},
		},
		{
			name: "sort priority bubbles up recursively to grandparent",
			setup: func() *graph.Node {
				childTs := time.Date(2025, 12, 1, 12, 0, 0, 0, time.UTC)
				grandparent := &graph.Node{
					Issue:         beads.FullIssue{ID: "ab-gp", Status: "open"},
					SortPriority:  3, // open
					SortTimestamp: time.Date(2025, 12, 1, 8, 0, 0, 0, time.UTC),
				}
				parent := &graph.Node{
					Issue:         beads.FullIssue{ID: "ab-p", Status: "open"},
					Parent:        grandparent,
					SortPriority:  3, // open
					SortTimestamp: time.Date(2025, 12, 1, 9, 0, 0, 0, time.UTC),
				}
				grandparent.Children = []*graph.Node{parent}
				child := &graph.Node{
					Issue:         beads.FullIssue{ID: "ab-c", Status: "in_progress"},
					Parent:        parent,
					SortPriority:  1, // in_progress
					SortTimestamp: childTs,
				}
				parent.Children = []*graph.Node{child}
				return child
			},
			validate: func(t *testing.T, child *graph.Node) {
				parent := child.Parent
				grandparent := parent.Parent
				if grandparent.SortPriority != 1 {
					t.Errorf("grandparent.SortPriority = %d, want 1 (in_progress)", grandparent.SortPriority)
				}
				if !grandparent.SortTimestamp.Equal(child.SortTimestamp) {
					t.Error("grandparent.SortTimestamp should match child's timestamp")
				}
			},
		},
		{
			name: "earlier timestamp bubbles up when same priority",
			setup: func() *graph.Node {
				earlierTs := time.Date(2025, 12, 1, 8, 0, 0, 0, time.UTC)
				parent := &graph.Node{
					Issue:         beads.FullIssue{ID: "ab-p", Status: "open"},
					SortPriority:  2, // ready
					SortTimestamp: time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC),
				}
				child := &graph.Node{
					Issue:         beads.FullIssue{ID: "ab-c", Status: "open"},
					Parent:        parent,
					IsBlocked:     false,
					SortPriority:  2, // ready (same as parent)
					SortTimestamp: earlierTs,
				}
				parent.Children = []*graph.Node{child}
				return child
			},
			validate: func(t *testing.T, child *graph.Node) {
				parent := child.Parent
				if parent.SortPriority != 2 {
					t.Errorf("parent.SortPriority = %d, want 2 (ready)", parent.SortPriority)
				}
				if !parent.SortTimestamp.Equal(child.SortTimestamp) {
					t.Errorf("parent.SortTimestamp should be updated to earlier child timestamp")
				}
			},
		},
		{
			name: "worse priority does not bubble up",
			setup: func() *graph.Node {
				parent := &graph.Node{
					Issue:         beads.FullIssue{ID: "ab-p", Status: "in_progress"},
					SortPriority:  1, // in_progress
					SortTimestamp: time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC),
				}
				child := &graph.Node{
					Issue:         beads.FullIssue{ID: "ab-c", Status: "open"},
					Parent:        parent,
					SortPriority:  3, // open (worse than parent)
					SortTimestamp: time.Date(2025, 12, 1, 8, 0, 0, 0, time.UTC),
				}
				parent.Children = []*graph.Node{child}
				return child
			},
			validate: func(t *testing.T, child *graph.Node) {
				parent := child.Parent
				if parent.SortPriority != 1 {
					t.Errorf("parent.SortPriority = %d, want 1 (unchanged)", parent.SortPriority)
				}
				// Parent timestamp should remain unchanged
				expectedTs := time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC)
				if !parent.SortTimestamp.Equal(expectedTs) {
					t.Error("parent.SortTimestamp should remain unchanged")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := tt.setup()
			propagateStateChanges(node)
			tt.validate(t, node)
		})
	}
}

func TestFastInjectBeadUsesParentHint(t *testing.T) {
	parent := &graph.Node{
		Issue: beads.FullIssue{ID: "ab-parent", Status: "open"},
		Depth: 0,
	}
	app := &App{roots: []*graph.Node{parent}}
	app.recalcVisibleRows()

	issue := beads.FullIssue{
		ID:     "ab-child",
		Status: "open",
	}

	if err := app.fastInjectBead(issue, "ab-parent"); err != nil {
		t.Fatalf("fastInjectBead returned error: %v", err)
	}

	if len(parent.Children) != 1 {
		t.Fatalf("expected parent to have 1 child, got %d", len(parent.Children))
	}
	child := parent.Children[0]
	if child.Issue.ID != "ab-child" {
		t.Fatalf("child ID = %s, want ab-child", child.Issue.ID)
	}
	if child.Parent == nil || child.Parent.Issue.ID != "ab-parent" {
		t.Fatalf("child parent mismatch: %+v", child.Parent)
	}
	if len(app.roots) != 1 {
		t.Fatalf("expected roots to stay at 1, got %d", len(app.roots))
	}
}

func TestFastInjectBeadFallsBackToDependencies(t *testing.T) {
	parent := &graph.Node{
		Issue: beads.FullIssue{ID: "ab-parent", Status: "open"},
		Depth: 0,
	}
	app := &App{roots: []*graph.Node{parent}}
	app.recalcVisibleRows()

	issue := beads.FullIssue{
		ID:     "ab-child",
		Status: "in_progress",
		Dependencies: []beads.Dependency{
			{TargetID: "ab-parent", Type: "parent-child"},
		},
	}

	if err := app.fastInjectBead(issue, ""); err != nil {
		t.Fatalf("fastInjectBead returned error: %v", err)
	}
	if len(parent.Children) != 1 {
		t.Fatalf("expected parent to have 1 child, got %d", len(parent.Children))
	}
	child := parent.Children[0]
	if child.Issue.ID != "ab-child" {
		t.Fatalf("child ID = %s, want ab-child", child.Issue.ID)
	}
	if !parent.Expanded {
		t.Fatalf("expected parent to auto-expand when child in progress")
	}
}
