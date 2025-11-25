package graph

import (
	"time"

	"abacus/internal/beads"
)

// Node represents a Beads issue within the dependency graph used by the UI.
type Node struct {
	Issue    beads.FullIssue
	Children []*Node
	Parents  []*Node
	Parent   *Node

	BlockedBy      []*Node
	Blocks         []*Node
	Related        []*Node
	DiscoveredFrom []*Node

	IsBlocked      bool
	CommentsLoaded bool
	CommentError   string

	Expanded      bool
	Depth         int
	TreeDepth     int
	HasInProgress bool
	HasReady      bool

	SortPriority  int
	SortTimestamp time.Time
}

// TreeRow represents a single row in the tree view.
//
// Multi-Parent Support: A Node may appear in multiple TreeRows when it has
// multiple parents. For example, if task "T" is a child of both "Epic A" and
// "Epic B", there will be two TreeRows for T - one under each parent. This
// allows users to see the task when browsing any of its parent epics.
//
// The tree view uses TreeRow.Parent to determine which parent's context the
// row is displayed under, while TreeRow.Node.Parents contains all parents.
// Cross-highlighting uses Node.Issue.ID to identify duplicate instances.
type TreeRow struct {
	Node   *Node // The underlying node (shared across duplicate rows)
	Parent *Node // The parent context for this specific row (nil for roots)
	Depth  int   // Visual depth in the tree
}

// HasMultipleParents returns true if the node has more than one parent.
func (r TreeRow) HasMultipleParents() bool {
	return len(r.Node.Parents) > 1
}
