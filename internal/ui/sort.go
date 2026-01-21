package ui

// Relationship Section Sorting
//
// This file provides sorting functions for the detail pane relationship sections.
// The goal is to surface actionable items first and show logical implementation order.
//
// Sorting principles:
// - sortSubtasks: Status-first (in_progress → ready → blocked → deferred → closed),
//   within ready: high-impact items (that unblock others) first,
//   within blocked: items closest to ready (fewest blockers) first
// - sortBlockers: Topological order - items you can work on now appear first,
//   then items with fewer blockers (easier to unblock), then by priority
// - sortBlocked: Items that will become ready first (fewest other blockers)
//   appear first, showing what gets unblocked when this issue is completed

import (
	"sort"
	"time"

	"abacus/internal/graph"
)

// Status categories for sorting (lower = higher priority)
const (
	statusInProgress = 1
	statusReady      = 2
	statusBlocked    = 3
	statusDeferred   = 4
	statusClosed     = 5
)

// nodeStatusCategory returns the sorting category for a node
func nodeStatusCategory(n *graph.Node) int {
	switch n.Issue.Status {
	case "in_progress":
		return statusInProgress
	case "closed":
		return statusClosed
	case "blocked":
		// Explicit blocked status
		return statusBlocked
	case "deferred":
		// Deferred (on ice) - lower priority than blocked
		return statusDeferred
	default: // open
		if n.IsBlocked {
			return statusBlocked
		}
		return statusReady
	}
}

// countOpenBlockers returns the number of unclosed blockers for a node
func countOpenBlockers(n *graph.Node) int {
	count := 0
	for _, b := range n.BlockedBy {
		if b.Issue.Status != "closed" {
			count++
		}
	}
	return count
}

// parseClosedAt parses the ClosedAt timestamp, returning zero time if unparseable
func parseClosedAt(n *graph.Node) time.Time {
	if n.Issue.ClosedAt == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, n.Issue.ClosedAt)
	if err != nil {
		return time.Time{}
	}
	return t
}

// sortSubtasks orders children: in_progress → ready (unblocking others first) → blocked → closed
func sortSubtasks(nodes []*graph.Node) []*graph.Node {
	if len(nodes) <= 1 {
		return nodes
	}

	result := make([]*graph.Node, len(nodes))
	copy(result, nodes)

	sort.SliceStable(result, func(i, j int) bool {
		catI := nodeStatusCategory(result[i])
		catJ := nodeStatusCategory(result[j])

		// Primary: status category
		if catI != catJ {
			return catI < catJ
		}

		// Secondary within ready: items that unblock others come first
		if catI == statusReady {
			blocksI := len(result[i].Blocks)
			blocksJ := len(result[j].Blocks)
			if blocksI != blocksJ {
				return blocksI > blocksJ // More blocks = higher priority
			}
		}

		// Secondary within blocked: closest to ready (fewest blockers) first
		if catI == statusBlocked {
			openBlockersI := countOpenBlockers(result[i])
			openBlockersJ := countOpenBlockers(result[j])
			if openBlockersI != openBlockersJ {
				return openBlockersI < openBlockersJ // Fewer blockers = higher priority
			}
		}

		// Secondary within closed: reverse chronological (most recent first)
		if catI == statusClosed {
			closedI := parseClosedAt(result[i])
			closedJ := parseClosedAt(result[j])
			if !closedI.IsZero() && !closedJ.IsZero() && !closedI.Equal(closedJ) {
				return closedI.After(closedJ) // More recent = higher priority
			}
		}

		// Tertiary: priority (lower number = higher priority)
		if result[i].Issue.Priority != result[j].Issue.Priority {
			return result[i].Issue.Priority < result[j].Issue.Priority
		}

		// Quaternary: by ID for stability
		return result[i].Issue.ID < result[j].Issue.ID
	})

	return result
}

// sortBlockers orders blockers topologically - things to do first appear first
// Items with no blockers of their own come first, then by priority
func sortBlockers(nodes []*graph.Node) []*graph.Node {
	if len(nodes) <= 1 {
		return nodes
	}

	result := make([]*graph.Node, len(nodes))
	copy(result, nodes)

	sort.SliceStable(result, func(i, j int) bool {
		// Closed items last
		iClosed := result[i].Issue.Status == "closed"
		jClosed := result[j].Issue.Status == "closed"
		if iClosed != jClosed {
			return jClosed // closed items come last
		}

		// Within closed items: reverse chronological (most recent first)
		if iClosed && jClosed {
			closedI := parseClosedAt(result[i])
			closedJ := parseClosedAt(result[j])
			if !closedI.IsZero() && !closedJ.IsZero() && !closedI.Equal(closedJ) {
				return closedI.After(closedJ)
			}
		}

		// Items with fewer open blockers come first (can be worked on sooner)
		openBlockersI := countOpenBlockers(result[i])
		openBlockersJ := countOpenBlockers(result[j])
		if openBlockersI != openBlockersJ {
			return openBlockersI < openBlockersJ
		}

		// In-progress items come first among same-blocker-count items
		iInProgress := result[i].Issue.Status == "in_progress"
		jInProgress := result[j].Issue.Status == "in_progress"
		if iInProgress != jInProgress {
			return iInProgress
		}

		// By priority
		if result[i].Issue.Priority != result[j].Issue.Priority {
			return result[i].Issue.Priority < result[j].Issue.Priority
		}

		return result[i].Issue.ID < result[j].Issue.ID
	})

	return result
}

// sortBlocked orders downstream items - items that become ready first come first
func sortBlocked(nodes []*graph.Node) []*graph.Node {
	if len(nodes) <= 1 {
		return nodes
	}

	result := make([]*graph.Node, len(nodes))
	copy(result, nodes)

	sort.SliceStable(result, func(i, j int) bool {
		// Items that will become ready (only blocked by one thing) come first
		openBlockersI := countOpenBlockers(result[i])
		openBlockersJ := countOpenBlockers(result[j])

		// Items with only 1 open blocker (this issue) will become ready
		if openBlockersI != openBlockersJ {
			return openBlockersI < openBlockersJ
		}

		// By priority
		if result[i].Issue.Priority != result[j].Issue.Priority {
			return result[i].Issue.Priority < result[j].Issue.Priority
		}

		return result[i].Issue.ID < result[j].Issue.ID
	})

	return result
}
