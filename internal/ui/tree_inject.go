package ui

import (
	"fmt"
	"sort"
	"time"

	"abacus/internal/beads"
	"abacus/internal/graph"
)

// fastInjectBead performs fast tree injection of a newly created bead.
// parentHint allows callers to pass the known parent ID when dependency
// metadata isn't yet populated (e.g., immediate create responses).
// Returns error if injection fails (caller should fall back to full refresh).
func (m *App) fastInjectBead(issue beads.FullIssue, parentHint string) error {
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		if elapsed > 50*time.Millisecond {
			// Log warning if injection took > 50ms
			m.lastError = fmt.Sprintf("Injection took %v (target: <50ms)", elapsed)
			m.lastErrorSource = errorSourceOperation
		}
	}()

	// 1. Construct node from issue
	newNode := constructNodeFromIssue(issue)

	// 2. Determine parent ID: prefer explicit hint, fall back to dependencies
	parentID := parentHint
	if parentID == "" {
		for _, dep := range issue.Dependencies {
			if dep.Type == "parent-child" {
				parentID = dep.TargetID
				break
			}
		}
	}

	// 3. Insert into tree
	if err := m.insertNodeIntoParent(newNode, parentID); err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	// 4. Propagate state changes up the tree
	propagateStateChanges(newNode)

	// 5. Recalculate visible rows (incremental traversal)
	m.recalcVisibleRows()

	// 6. Move cursor to new node
	m.restoreCursorToID(newNode.Issue.ID)

	// 7. Auto-expand parent if new node is in_progress
	if newNode.Issue.Status == "in_progress" && parentID != "" {
		if parent := m.findNodeByID(parentID); parent != nil {
			parent.Expanded = true
			m.recalcVisibleRows() // Recalc again to show expanded children
		}
	}

	return nil
}

// constructNodeFromIssue creates a properly initialized Node from FullIssue.
// Sets all computed fields (SortPriority, SortTimestamp, IsBlocked, etc.)
func constructNodeFromIssue(issue beads.FullIssue) *graph.Node {
	node := &graph.Node{
		Issue:    issue,
		Children: []*graph.Node{},
		Parents:  []*graph.Node{},

		// Dependency relationships (initially empty for new nodes)
		BlockedBy:      []*graph.Node{},
		Blocks:         []*graph.Node{},
		Related:        []*graph.Node{},
		DiscoveredFrom: []*graph.Node{},

		// Computed states
		IsBlocked:      false, // New nodes have no blocking deps yet
		CommentsLoaded: false,
		CommentError:   "",

		// UI state
		Expanded:      false, // Don't auto-expand new nodes
		Depth:         0,     // Will be set by parent context
		TreeDepth:     0,     // Will be computed after insertion
		HasInProgress: issue.Status == "in_progress",
		HasReady:      issue.Status == "open", // New nodes can't be blocked yet

		// Sort metrics - computed below
		SortPriority:  0,
		SortTimestamp: time.Time{},
	}

	// Compute sort metrics using existing builder logic
	priority, ts := graph.NodeSelfSortKey(node)
	node.SortPriority = priority
	node.SortTimestamp = ts

	return node
}

// findInsertPosition performs binary search to find correct insert position
// in a sorted slice based on Node's sort metrics.
func findInsertPosition(nodes []*graph.Node, node *graph.Node) int {
	return sort.Search(len(nodes), func(i int) bool {
		existing := nodes[i]

		// Sort by: Priority → Timestamp → ID (matches builder.go:264 sortNodes)
		if existing.SortPriority != node.SortPriority {
			return existing.SortPriority > node.SortPriority
		}
		if !existing.SortTimestamp.Equal(node.SortTimestamp) {
			return existing.SortTimestamp.After(node.SortTimestamp)
		}
		return existing.Issue.ID > node.Issue.ID
	})
}

// insertNodeIntoParent inserts a child node into parent's Children array
// at the correct sorted position. Handles both root insertion and child insertion.
func (m *App) insertNodeIntoParent(child *graph.Node, parentID string) error {
	if parentID == "" {
		// Root insertion
		pos := findInsertPosition(m.roots, child)
		m.roots = append(m.roots[:pos], append([]*graph.Node{child}, m.roots[pos:]...)...)
		return nil
	}

	// Find parent node
	parent := m.findNodeByID(parentID)
	if parent == nil {
		return fmt.Errorf("parent node not found: %s", parentID)
	}

	// Set up parent-child relationships
	child.Parents = []*graph.Node{parent}
	child.Parent = parent
	child.Depth = parent.Depth + 1

	// Calculate TreeDepth (depth from any root)
	child.TreeDepth = parent.TreeDepth + 1

	// Insert into parent's Children at sorted position
	pos := findInsertPosition(parent.Children, child)
	parent.Children = append(parent.Children[:pos],
		append([]*graph.Node{child}, parent.Children[pos:]...)...)

	return nil
}

// propagateStateChanges updates ancestor HasInProgress/HasReady flags and
// SortPriority/SortTimestamp after inserting a new node. This ensures parents
// with active children sort appropriately (e.g., a parent with an in_progress
// child sorts as in_progress).
func propagateStateChanges(node *graph.Node) {
	// Walk up the parent chain updating flags and sort metrics
	current := node.Parent
	if current == nil && len(node.Parents) > 0 {
		current = node.Parents[0]
	}

	// Track the child's sort metrics as we bubble up
	childPriority := node.SortPriority
	childTimestamp := node.SortTimestamp

	for current != nil {
		// Update parent's flags based on new child
		if node.Issue.Status == "in_progress" {
			current.HasInProgress = true
			// Auto-expand parent when in_progress child added
			current.Expanded = true
		}
		if node.Issue.Status == "open" && !node.IsBlocked {
			current.HasReady = true
		}

		// Bubble up sort metrics if child has better (lower) priority
		// or same priority with earlier timestamp
		if childPriority < current.SortPriority ||
			(childPriority == current.SortPriority &&
				childTimestamp.Before(current.SortTimestamp)) {
			current.SortPriority = childPriority
			current.SortTimestamp = childTimestamp
		}

		// Move up the chain
		next := current.Parent
		if next == nil && len(current.Parents) > 0 {
			next = current.Parents[0]
		}
		current = next
	}
}
