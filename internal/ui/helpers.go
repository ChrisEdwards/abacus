package ui

import (
	"strings"
	"time"

	"abacus/internal/domain"
	"abacus/internal/graph"
)

// SessionInfo contains session tracking data for the exit summary.
type SessionInfo struct {
	StartTime    time.Time
	InitialStats Stats
}

// GetSessionInfo returns session tracking data for the exit summary.
func (m *App) GetSessionInfo() SessionInfo {
	return SessionInfo{
		StartTime:    m.sessionStart,
		InitialStats: m.initialStats,
	}
}

// GetStats returns issue counts for external use (e.g., exit summary).
// This is the public version of getStats().
func (m *App) GetStats() Stats {
	return m.getStats()
}

// getStats computes issue counts for the status bar.
//
// Count Deduplication: With multi-parent support, the same Node appears in
// multiple parents' Children slices. To avoid double-counting, we track seen
// IDs. The principle is:
//   - Logical counts (stats, progress): count unique issues once
//   - Visual instances (row indices, scrolling): count TreeRow instances
func (m *App) getStats() Stats {
	s := Stats{}
	filterLower := strings.ToLower(m.filterText)
	seen := make(map[string]bool) // Track counted nodes to avoid double-counting multi-parent nodes

	var traverse func(nodes []*graph.Node)
	traverse = func(nodes []*graph.Node) {
		for _, n := range nodes {
			// Skip if already counted (multi-parent case where same node appears under multiple parents)
			if seen[n.Issue.ID] {
				traverse(n.Children) // Still traverse children
				continue
			}
			seen[n.Issue.ID] = true

			matches := nodeMatchesFilter(filterLower, n)

			domainIssue, err := domain.NewIssueFromFull(n.Issue, n.IsBlocked)
			if matches {
				s.Total++
				if err != nil {
					// Fallback for unknown statuses (e.g., "pinned" from br backend).
					// Only count as Ready if explicitly "open" to match IsReady() semantics.
					// Unknown statuses are counted in Total but not in any bucket.
					if n.Issue.Status == "in_progress" {
						s.InProgress++
					} else if n.Issue.Status == "closed" {
						s.Closed++
					} else if n.IsBlocked {
						s.Blocked++
					} else if n.Issue.Status == "open" {
						s.Ready++
					}
				} else {
					switch {
					case domainIssue.Status() == domain.StatusInProgress:
						s.InProgress++
					case domainIssue.Status() == domain.StatusClosed:
						s.Closed++
					case domainIssue.IsBlocked():
						s.Blocked++
					case domainIssue.IsReady():
						s.Ready++
					default:
						s.Ready++
					}
				}
			}
			traverse(n.Children)
		}
	}
	traverse(m.roots)
	return s
}
