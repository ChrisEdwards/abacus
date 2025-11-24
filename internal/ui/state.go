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
	m.visibleRows = []*graph.Node{}
	filterActive := strings.TrimSpace(m.filterText) != "" || len(m.filterTokens) > 0

	if filterActive {
		m.filterEval = m.computeFilterEval()
	} else {
		m.filterEval = nil
	}

	var traverse func(nodes []*graph.Node)
	traverse = func(nodes []*graph.Node) {
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
				m.visibleRows = append(m.visibleRows, node)
				if !filterActive && node.Expanded {
					traverse(node.Children)
				} else if filterActive && m.shouldExpandFilteredNode(node, hasMatchingChild) {
					traverse(node.Children)
				}
			}
		}
	}
	traverse(m.roots)
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
		filterCollapsed:      copyBoolMap(m.filterCollapsed),
		filterForcedExpanded: copyBoolMap(m.filterForcedExpanded),
		focus:                m.focus,
	}

	if m.ShowDetails && m.viewport.Height > 0 {
		state.viewportYOffset = m.viewport.YOffset
	}

	if len(m.visibleRows) > 0 && m.cursor >= 0 && m.cursor < len(m.visibleRows) {
		state.currentID = m.visibleRows[m.cursor].Issue.ID
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

func (m *App) restoreCursorToID(id string) {
	prev := m.cursor
	if id == "" {
		m.clampCursor()
		return
	}
	for idx, node := range m.visibleRows {
		if node.Issue.ID == id {
			m.cursor = idx
			return
		}
	}
	m.cursor = prev
	m.clampCursor()
}

func (m *App) computeFilterEval() map[string]filterEvaluation {
	words := append([]string{}, m.filterFreeText...)
	if len(words) == 0 && len(m.filterTokens) == 0 {
		trimmed := strings.TrimSpace(strings.ToLower(m.filterText))
		if trimmed != "" {
			words = strings.Fields(trimmed)
		}
	}
	tokens := append([]SearchToken{}, m.filterTokens...)
	evals := make(map[string]filterEvaluation)
	var walk func(node *graph.Node) bool
	walk = func(node *graph.Node) bool {
		matches := nodeMatchesTokenFilter(words, tokens, node)
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

func (m *App) isNodeExpandedInView(node *graph.Node) bool {
	if len(node.Children) == 0 {
		return false
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

func (m *App) expandNodeForView(node *graph.Node) {
	node.Expanded = true
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

func (m *App) collapseNodeForView(node *graph.Node) {
	node.Expanded = false
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
