package ui

import "abacus/internal/graph"

// isRowExpandedForTraversal checks if a row should be expanded during tree traversal.
// For multi-parent nodes, it checks per-instance state first.
func (m *App) isRowExpandedForTraversal(row graph.TreeRow) bool {
	node := row.Node
	if len(node.Children) == 0 {
		return false
	}

	// Check per-instance state for multi-parent nodes
	if row.HasMultipleParents() {
		parentID := ""
		if row.Parent != nil {
			parentID = row.Parent.Issue.ID
		}
		key := treeRowKey(parentID, node.Issue.ID)
		if expanded, ok := m.expandedInstances[key]; ok {
			return expanded
		}
		// Fall back to Node.Expanded if no per-instance state set yet
	}

	return node.Expanded
}

func (m *App) isNodeExpandedInView(row graph.TreeRow) bool {
	node := row.Node
	if len(node.Children) == 0 {
		return false
	}

	// Check per-instance state for multi-parent nodes
	if row.HasMultipleParents() {
		parentID := ""
		if row.Parent != nil {
			parentID = row.Parent.Issue.ID
		}
		key := treeRowKey(parentID, node.Issue.ID)
		if expanded, ok := m.expandedInstances[key]; ok {
			// When filtering, also check filter overrides
			if m.filterText != "" {
				if m.filterCollapsed != nil && m.filterCollapsed[node.Issue.ID] {
					return false
				}
				if m.filterForcedExpanded != nil && m.filterForcedExpanded[node.Issue.ID] {
					return true
				}
			}
			return expanded
		}
		// Fall back to Node.Expanded if no per-instance state set yet
	}

	if m.filterText == "" {
		return node.Expanded
	}
	hasMatchingChild := false
	if m.filterEval != nil {
		if eval, ok := m.filterEval[node.Issue.ID]; ok {
			hasMatchingChild = eval.hasMatchingChild
		}
	}
	return m.shouldExpandFilteredNode(node, hasMatchingChild)
}

// treeRowKey creates a composite key for tracking per-instance state of multi-parent nodes.
// Format: "parentID:nodeID" where parentID is empty for root nodes.
func treeRowKey(parentID, nodeID string) string {
	return parentID + ":" + nodeID
}

func copyBoolMap(src map[string]bool) map[string]bool {
	if len(src) == 0 {
		return nil
	}
	dest := make(map[string]bool, len(src))
	for k, v := range src {
		if v {
			dest[k] = true
		}
	}
	if len(dest) == 0 {
		return nil
	}
	return dest
}

// copyBoolMapAll copies all entries from src, preserving both true and false values.
// Used for expandedInstances where false explicitly means "collapsed".
func copyBoolMapAll(src map[string]bool) map[string]bool {
	if len(src) == 0 {
		return nil
	}
	dest := make(map[string]bool, len(src))
	for k, v := range src {
		dest[k] = v
	}
	return dest
}

func (m *App) expandNodeForView(row graph.TreeRow) {
	node := row.Node

	// Track per-instance state for multi-parent nodes
	if row.HasMultipleParents() {
		parentID := ""
		if row.Parent != nil {
			parentID = row.Parent.Issue.ID
		}
		key := treeRowKey(parentID, node.Issue.ID)
		if m.expandedInstances == nil {
			m.expandedInstances = make(map[string]bool)
		}
		m.expandedInstances[key] = true
		// Don't modify shared node.Expanded for multi-parent nodes
	} else {
		node.Expanded = true
	}

	if m.filterText == "" {
		return
	}
	id := node.Issue.ID
	if m.filterCollapsed != nil {
		delete(m.filterCollapsed, id)
		if len(m.filterCollapsed) == 0 {
			m.filterCollapsed = nil
		}
	}
	if m.filterForcedExpanded == nil {
		m.filterForcedExpanded = make(map[string]bool)
	}
	m.filterForcedExpanded[id] = true
}

func (m *App) collapseNodeForView(row graph.TreeRow) {
	node := row.Node

	// Track per-instance state for multi-parent nodes
	if row.HasMultipleParents() {
		parentID := ""
		if row.Parent != nil {
			parentID = row.Parent.Issue.ID
		}
		key := treeRowKey(parentID, node.Issue.ID)
		if m.expandedInstances == nil {
			m.expandedInstances = make(map[string]bool)
		}
		m.expandedInstances[key] = false
		// Don't modify shared node.Expanded for multi-parent nodes
	} else {
		node.Expanded = false
	}

	if m.filterText == "" {
		return
	}
	id := node.Issue.ID
	if m.filterCollapsed == nil {
		m.filterCollapsed = make(map[string]bool)
	}
	m.filterCollapsed[id] = true
	if m.filterForcedExpanded != nil {
		delete(m.filterForcedExpanded, id)
		if len(m.filterForcedExpanded) == 0 {
			m.filterForcedExpanded = nil
		}
	}
}

// findNodeByID searches the tree for a node with the given ID.
func (m *App) findNodeByID(id string) *graph.Node {
	var result *graph.Node
	var walk func(nodes []*graph.Node)
	walk = func(nodes []*graph.Node) {
		for _, n := range nodes {
			if n.Issue.ID == id {
				result = n
				return
			}
			walk(n.Children)
		}
	}
	walk(m.roots)
	return result
}

// restoreCursorToRow finds the exact row matching nodeID and parentID (for multi-parent support).
// Returns true if the row was found and cursor was set.
func (m *App) restoreCursorToRow(nodeID, parentID string) bool {
	for idx, row := range m.visibleRows {
		if row.Node.Issue.ID != nodeID {
			continue
		}
		rowParentID := ""
		if row.Parent != nil {
			rowParentID = row.Parent.Issue.ID
		}
		if rowParentID == parentID {
			m.cursor = idx
			return true
		}
	}
	return false
}

// expandAncestorsForRow expands the ancestor chain for a specific parent context
// so the node will be visible after the filter is cleared.
func (m *App) expandAncestorsForRow(nodeID, parentID string) {
	// Find the node
	node := m.findNodeByID(nodeID)
	if node == nil {
		return
	}

	// If parentID is specified, find that specific parent and expand up from there
	if parentID != "" {
		parent := m.findNodeByID(parentID)
		for parent != nil {
			if len(parent.Parents) > 1 {
				// Multi-parent: use per-instance expansion with root context
				key := treeRowKey("", parent.Issue.ID)
				if m.expandedInstances == nil {
					m.expandedInstances = make(map[string]bool)
				}
				m.expandedInstances[key] = true
			} else {
				parent.Expanded = true
			}
			if len(parent.Parents) > 0 {
				parent = parent.Parents[0]
			} else {
				parent = parent.Parent
			}
		}
	}

	// Also expand up from the node itself
	current := node.Parent
	if current == nil && len(node.Parents) > 0 {
		current = node.Parents[0]
	}
	for current != nil {
		current.Expanded = true
		next := current.Parent
		if next == nil && len(current.Parents) > 0 {
			next = current.Parents[0]
		}
		current = next
	}
}

// transferFilterExpansionState copies filterForcedExpanded state to permanent Node.Expanded state
// so manually expanded nodes stay expanded after clearing the filter.
func (m *App) transferFilterExpansionState() {
	if m.filterForcedExpanded == nil {
		return
	}
	for id := range m.filterForcedExpanded {
		node := m.findNodeByID(id)
		if node != nil {
			node.Expanded = true
		}
	}
}

// removeNodeFromTree removes a node with the given ID from the tree structure.
// It removes the node from roots if it's a root, and from all parents' Children arrays.
// After removal, recalcVisibleRows should be called to update the display.
func (m *App) removeNodeFromTree(id string) {
	// Remove from roots if present
	newRoots := make([]*graph.Node, 0, len(m.roots))
	for _, root := range m.roots {
		if root.Issue.ID != id {
			newRoots = append(newRoots, root)
		}
	}
	m.roots = newRoots

	// Remove from all parent Children arrays (handles multi-parent case)
	var removeFromChildren func(nodes []*graph.Node)
	removeFromChildren = func(nodes []*graph.Node) {
		for _, node := range nodes {
			newChildren := make([]*graph.Node, 0, len(node.Children))
			for _, child := range node.Children {
				if child.Issue.ID != id {
					newChildren = append(newChildren, child)
				}
			}
			node.Children = newChildren
			removeFromChildren(node.Children)
		}
	}
	removeFromChildren(m.roots)
}
