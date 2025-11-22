package graph

import "abacus/internal/beads"

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
}
