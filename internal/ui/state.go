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
	viewMode             ViewMode
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
	filterActive := m.isFilterActive()

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
					expanded = m.shouldExpandFilteredRow(row, hasMatchingChild)
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
		viewMode:             m.viewMode,
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

// tallLayoutSplit computes the tree and detail pane heights for tall layout.
// sharedBudget = listHeight - 2: two panes each have 1 border top/bottom (4 total),
// but must match the wide-mode mainBody height of listHeight+2. So inner content
// available = (listHeight+2) - 4 = listHeight - 2.
func (m *App) tallLayoutSplit() (treeH, detailH int) {
	listHeight := clampDimension(m.height-4, minListHeight, m.height-2)
	sharedBudget := listHeight - 2
	detailH = int(float64(sharedBudget) * 0.6)
	detailH = clampDimension(detailH, minViewportHeight, sharedBudget-minListHeight)
	treeH = sharedBudget - detailH
	return
}

// treePaneHeight returns the height available for the tree list.
// In wide mode (or when details are closed), this is the full body height.
// In tall mode with details open, it is the tree portion of the shared budget.
func (m *App) treePaneHeight() int {
	if m.layout != LayoutTall || !m.ShowDetails {
		return clampDimension(m.height-4, minListHeight, m.height-2)
	}
	treeH, _ := m.tallLayoutSplit()
	return treeH
}

// recalcViewportSize recomputes viewport width and height based on current layout.
// Called on resize and on layout toggle.
func (m *App) recalcViewportSize() {
	if m.layout == LayoutTall {
		m.viewport.Width = m.width - 2
		_, detailH := m.tallLayoutSplit()
		m.viewport.Height = detailH
	} else {
		rawViewportWidth := int(float64(m.width)*0.45) - 2
		maxViewportWidth := m.width - minTreeWidth - 4
		m.viewport.Width = clampDimension(rawViewportWidth, minViewportWidth, maxViewportWidth)
		rawViewportHeight := m.height - 5
		maxViewportHeight := m.height - 2
		m.viewport.Height = clampDimension(rawViewportHeight, minViewportHeight, maxViewportHeight)
	}
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
