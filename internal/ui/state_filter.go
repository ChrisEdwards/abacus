package ui

import (
	"strings"

	"abacus/internal/domain"
	"abacus/internal/graph"
)

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

// nodeMatchesViewMode checks if a node matches the current view mode filter.
func nodeMatchesViewMode(mode ViewMode, node *graph.Node) bool {
	if mode == ViewModeAll {
		return true
	}

	// Try to convert to domain issue for proper status checking
	domainIssue, err := domain.NewIssueFromFull(node.Issue, node.IsBlocked)
	if err != nil {
		// Fallback to direct string comparison
		switch mode {
		case ViewModeActive:
			return node.Issue.Status != "closed"
		case ViewModeReady:
			return node.Issue.Status != "closed" && !node.IsBlocked
		default:
			return true
		}
	}

	switch mode {
	case ViewModeActive:
		// Show non-closed issues
		return domainIssue.Status() != domain.StatusClosed
	case ViewModeReady:
		// Show ready issues (open + not blocked)
		return domainIssue.IsReady()
	default:
		return true
	}
}

func (m *App) computeFilterEval(filterLower string) map[string]filterEvaluation {
	evals := make(map[string]filterEvaluation)
	var walk func(node *graph.Node) bool
	walk = func(node *graph.Node) bool {
		// Check BOTH ViewMode AND text filter
		viewModeMatch := nodeMatchesViewMode(m.viewMode, node)
		textMatch := nodeMatchesFilter(filterLower, node)
		directMatch := viewModeMatch && textMatch // Node itself matches both filters

		hasChildMatch := false
		for _, child := range node.Children {
			if walk(child) {
				hasChildMatch = true
			}
		}
		evals[node.Issue.ID] = filterEvaluation{
			matches:          directMatch,
			hasMatchingChild: hasChildMatch,
		}
		return directMatch || hasChildMatch
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
