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

	BlockedBy []*Node
	Blocks    []*Node

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

// TreeRow represents a single row in the tree view. A Node may appear in
// multiple TreeRows when it has multiple parents.
type TreeRow struct {
	Node   *Node // The underlying node
	Parent *Node // The parent context for this row (nil for roots)
	Depth  int   // Visual depth in the tree
}

// HasMultipleParents returns true if the node has more than one parent.
func (r TreeRow) HasMultipleParents() bool {
	return len(r.Node.Parents) > 1
}
