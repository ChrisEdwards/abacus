package ui

import (
	"strings"

	"abacus/internal/graph"
)

type FocusArea int

const (
	FocusTree FocusArea = iota
	FocusDetails
)

type Stats struct {
	Total      int
	InProgress int
	Ready      int
	Blocked    int
	Closed     int
}

type viewState struct {
	currentID            string
	expandedIDs          map[string]bool
	expandedInstances    map[string]bool // per-instance state for multi-parent nodes
	filterText           string
	filterCollapsed      map[string]bool
	filterForcedExpanded map[string]bool
	viewportYOffset      int
	cursorIndex          int
	focus                FocusArea
}

type filterEvaluation struct {
	matches          bool
	hasMatchingChild bool
}

func clampDimension(value, minValue, maxValue int) int {
	if maxValue < 1 {
		maxValue = 1
	}
	if minValue < 1 {
		minValue = 1
	}
	if minValue > maxValue {
		minValue = maxValue
	}
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func (m *App) recalcVisibleRows() {
	m.visibleRows = []graph.TreeRow{}
	filterLower := strings.ToLower(m.filterText)
	filterActive := filterLower != ""

	if filterActive {
		m.filterEval = m.computeFilterEval(filterLower)
	} else {
		m.filterEval = nil
	}

	var traverse func(nodes []*graph.Node, parent *graph.Node, depth int)
	traverse = func(nodes []*graph.Node, parent *graph.Node, depth int) {
		for _, node := range nodes {
			includeNode := true
			hasMatchingChild := false
			if filterActive {
				if eval, ok := m.filterEval[node.Issue.ID]; ok {
					includeNode = eval.matches || eval.hasMatchingChild
					hasMatchingChild = eval.hasMatchingChild
				} else {
					includeNode = false
				}
			}

			if includeNode {
				row := graph.TreeRow{
					Node:   node,
					Parent: parent,
					Depth:  depth,
				}
				m.visibleRows = append(m.visibleRows, row)

				// Use per-instance expansion state for multi-parent nodes
				expanded := false
				if !filterActive {
					expanded = m.isRowExpandedForTraversal(row)
				} else {
					expanded = m.shouldExpandFilteredNode(node, hasMatchingChild)
				}
				if expanded {
					traverse(node.Children, node, depth+1)
				}
			}
		}
	}
	traverse(m.roots, nil, 0)
	m.clampCursor()
}

func (m *App) clampCursor() {
	if len(m.visibleRows) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor >= len(m.visibleRows) {
		m.cursor = len(m.visibleRows) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *App) captureState() viewState {
	state := viewState{
		filterText:           m.filterText,
		cursorIndex:          m.cursor,
		expandedIDs:          m.collectExpandedIDs(),
		expandedInstances:    copyBoolMapAll(m.expandedInstances),
		filterCollapsed:      copyBoolMap(m.filterCollapsed),
		filterForcedExpanded: copyBoolMap(m.filterForcedExpanded),
		focus:                m.focus,
	}

	if m.ShowDetails && m.viewport.Height > 0 {
		state.viewportYOffset = m.viewport.YOffset
	}

	if len(m.visibleRows) > 0 && m.cursor >= 0 && m.cursor < len(m.visibleRows) {
		state.currentID = m.visibleRows[m.cursor].Node.Issue.ID
	}
	return state
}

func (m *App) collectExpandedIDs() map[string]bool {
	expanded := make(map[string]bool)
	var walk func(nodes []*graph.Node)
	walk = func(nodes []*graph.Node) {
		for _, n := range nodes {
			if n.Expanded {
				expanded[n.Issue.ID] = true
			}
			walk(n.Children)
		}
	}
	walk(m.roots)
	return expanded
}

func nodeMatchesFilter(filterLower string, node *graph.Node) bool {
	if filterLower == "" {
		return true
	}

	titleLower := strings.ToLower(node.Issue.Title)
	if strings.Contains(titleLower, filterLower) {
		return true
	}

	idLower := strings.ToLower(node.Issue.ID)
	if strings.Contains(idLower, filterLower) {
		return true
	}

	trimmed := strings.TrimPrefix(idLower, "ab-")
	return strings.Contains(trimmed, filterLower)
}

func (m *App) restoreExpandedState(expanded map[string]bool) {
	if expanded == nil {
		expanded = map[string]bool{}
	}
	var walk func(nodes []*graph.Node)
	walk = func(nodes []*graph.Node) {
		for _, n := range nodes {
			n.Expanded = expanded[n.Issue.ID]
			walk(n.Children)
		}
	}
	walk(m.roots)
}

func (m *App) restoreCursorToID(id string) {
	prev := m.cursor
	if id == "" {
		m.clampCursor()
		return
	}
	for idx, row := range m.visibleRows {
		if row.Node.Issue.ID == id {
			m.cursor = idx
			return
		}
	}
	m.cursor = prev
	m.clampCursor()
}

func (m *App) computeFilterEval(filterLower string) map[string]filterEvaluation {
	evals := make(map[string]filterEvaluation)
	var walk func(node *graph.Node) bool
	walk = func(node *graph.Node) bool {
		matches := nodeMatchesFilter(filterLower, node)
		hasChildMatch := false
		for _, child := range node.Children {
			if walk(child) {
				hasChildMatch = true
			}
		}
		evals[node.Issue.ID] = filterEvaluation{
			matches:          matches,
			hasMatchingChild: hasChildMatch,
		}
		return matches || hasChildMatch
	}
	for _, root := range m.roots {
		walk(root)
	}
	return evals
}

func (m *App) shouldExpandFilteredNode(node *graph.Node, hasMatchingChild bool) bool {
	if len(node.Children) == 0 {
		return false
	}
	id := node.Issue.ID
	if m.filterCollapsed != nil && m.filterCollapsed[id] {
		return false
	}
	if m.filterForcedExpanded != nil && m.filterForcedExpanded[id] {
		return true
	}
	if hasMatchingChild {
		return true
	}
	return node.Expanded
}

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
