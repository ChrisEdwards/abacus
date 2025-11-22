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
	currentID       string
	expandedIDs     map[string]bool
	filterText      string
	viewportYOffset int
	cursorIndex     int
	focus           FocusArea
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
	filterLower := strings.ToLower(m.filterText)

	var traverse func(nodes []*graph.Node)
	traverse = func(nodes []*graph.Node) {
		for _, node := range nodes {
			isMatch := nodeMatchesFilter(filterLower, node)
			hasMatchingChild := false
			if filterLower != "" {
				var checkChildren func([]*graph.Node) bool
				checkChildren = func(kids []*graph.Node) bool {
					for _, k := range kids {
						if nodeMatchesFilter(filterLower, k) || checkChildren(k.Children) {
							return true
						}
					}
					return false
				}
				hasMatchingChild = checkChildren(node.Children)
			}

			if isMatch || hasMatchingChild {
				m.visibleRows = append(m.visibleRows, node)
				if (filterLower == "" && node.Expanded) || (filterLower != "" && hasMatchingChild) {
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
		filterText:  m.filterText,
		cursorIndex: m.cursor,
		expandedIDs: m.collectExpandedIDs(),
		focus:       m.focus,
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
	for idx, node := range m.visibleRows {
		if node.Issue.ID == id {
			m.cursor = idx
			return
		}
	}
	m.cursor = prev
	m.clampCursor()
}
