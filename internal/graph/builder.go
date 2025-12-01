package graph

import (
	"sort"
	"strings"
	"time"

	"abacus/internal/beads"
)

// Builder constructs dependency graphs from raw Beads issues.
type Builder struct{}

// NewBuilder creates a new Builder instance.
func NewBuilder() *Builder {
	return &Builder{}
}

// Build converts raw issues into a rooted dependency forest with computed states.
func (Builder) Build(issues []beads.FullIssue) ([]*Node, error) {
	if len(issues) == 0 {
		return []*Node{}, nil
	}

	nodeMap := make(map[string]*Node, len(issues))
	for _, iss := range issues {
		nodeMap[iss.ID] = &Node{Issue: iss}
	}

	// Populate relationships from dependencies/dependents metadata.
	for _, node := range nodeMap {
		for _, dep := range node.Issue.Dependencies {
			switch dep.Type {
			case "parent-child":
				if parent, ok := nodeMap[dep.TargetID]; ok {
					node.Parents = append(node.Parents, parent)
				}
			case "blocks":
				if blocker, ok := nodeMap[dep.TargetID]; ok {
					node.BlockedBy = append(node.BlockedBy, blocker)
					if blocker.Issue.Status != "closed" {
						node.IsBlocked = true
					}
					blocker.Blocks = append(blocker.Blocks, node)
				}
			case "related":
				if related, ok := nodeMap[dep.TargetID]; ok {
					node.Related = append(node.Related, related)
					related.Related = append(related.Related, node)
				}
			case "discovered-from":
				if source, ok := nodeMap[dep.TargetID]; ok {
					node.DiscoveredFrom = append(node.DiscoveredFrom, source)
				}
			}
		}
		for _, dep := range node.Issue.Dependents {
			if dep.Type != "parent-child" {
				continue
			}
			if child, ok := nodeMap[dep.ID]; ok {
				child.Parents = append(child.Parents, node)
			}
		}
	}

	for _, node := range nodeMap {
		if len(node.Parents) <= 1 {
			continue
		}
		seen := make(map[string]bool, len(node.Parents))
		uniq := node.Parents[:0]
		for _, parent := range node.Parents {
			if seen[parent.Issue.ID] {
				continue
			}
			seen[parent.Issue.ID] = true
			uniq = append(uniq, parent)
		}
		node.Parents = uniq
	}

	if err := ensureAcyclic(nodeMap); err != nil {
		return nil, err
	}

	for _, node := range nodeMap {
		node.TreeDepth = calculateDepth(node, make(map[string]bool))
	}

	var roots []*Node
	childrenIDs := make(map[string]bool)

	for _, node := range nodeMap {
		if len(node.Parents) == 0 {
			roots = append(roots, node)
			continue
		}

		// Add child to ALL parents' Children slices (multi-parent support)
		for _, p := range node.Parents {
			// Avoid duplicate children within same parent
			alreadyChild := false
			for _, existing := range p.Children {
				if existing.Issue.ID == node.Issue.ID {
					alreadyChild = true
					break
				}
			}
			if !alreadyChild {
				p.Children = append(p.Children, node)
			}
		}
		// Set Parent to first parent for backwards compatibility
		if len(node.Parents) > 0 {
			node.Parent = node.Parents[0]
		}
		childrenIDs[node.Issue.ID] = true
	}

	for _, node := range nodeMap {
		if len(node.Parents) > 0 && !childrenIDs[node.Issue.ID] {
			roots = append(roots, node)
		}
	}

	for _, node := range nodeMap {
		sort.Slice(node.Blocks, func(i, j int) bool {
			return node.Blocks[i].Issue.CreatedAt < node.Blocks[j].Issue.CreatedAt
		})
	}

	for _, root := range roots {
		computeStates(root)
		if root.HasInProgress {
			root.Expanded = true
		}
		computeSortMetrics(root)
	}
	sortNodes(roots)

	return roots, nil
}

func calculateDepth(n *Node, visited map[string]bool) int {
	if visited[n.Issue.ID] {
		return 0
	}
	visited[n.Issue.ID] = true
	if len(n.Parents) == 0 {
		return 0
	}
	maxDepth := 0
	for _, parent := range n.Parents {
		depth := calculateDepth(parent, visited)
		if depth > maxDepth {
			maxDepth = depth
		}
	}
	delete(visited, n.Issue.ID)
	return maxDepth + 1
}

func ensureAcyclic(nodes map[string]*Node) error {
	visited := make(map[string]bool)
	onStack := make(map[string]bool)

	var visit func(n *Node, stack []string) error
	visit = func(n *Node, stack []string) error {
		if onStack[n.Issue.ID] {
			stack = append(stack, n.Issue.ID)
			return cyclicDependencyError(stack)
		}
		if visited[n.Issue.ID] {
			return nil
		}
		onStack[n.Issue.ID] = true
		stack = append(stack, n.Issue.ID)
		for _, parent := range n.Parents {
			if err := visit(parent, stack); err != nil {
				return err
			}
		}
		onStack[n.Issue.ID] = false
		visited[n.Issue.ID] = true
		return nil
	}

	for _, node := range nodes {
		if err := visit(node, nil); err != nil {
			return err
		}
	}
	return nil
}

func computeStates(n *Node) {
	if n.Issue.Status == "in_progress" {
		n.HasInProgress = true
	}
	if n.Issue.Status == "open" && !n.IsBlocked {
		n.HasReady = true
	}
	for _, child := range n.Children {
		child.Depth = n.Depth + 1
		computeStates(child)
		if child.HasInProgress {
			n.HasInProgress = true
			n.Expanded = true
		}
		if child.HasReady {
			n.HasReady = true
		}
	}
}

const (
	sortPriorityInProgress = 1
	sortPriorityReady      = 2
	sortPriorityOpen       = 3
	sortPriorityClosed     = 4
)

var distantFuture = time.Date(9999, time.January, 1, 0, 0, 0, 0, time.UTC)

func computeSortMetrics(node *Node) (int, time.Time) {
	priority, ts := NodeSelfSortKey(node)
	for _, child := range node.Children {
		childPriority, childTime := computeSortMetrics(child)
		if childPriority < priority || (childPriority == priority && childTime.Before(ts)) {
			priority = childPriority
			ts = childTime
		}
	}
	node.SortPriority = priority
	node.SortTimestamp = ts
	sortNodes(node.Children)
	return priority, ts
}

// NodeSelfSortKey computes the sort priority and timestamp for a node based on its status.
// This is exported for use by the fast tree injection logic.
func NodeSelfSortKey(node *Node) (int, time.Time) {
	status := strings.ToLower(strings.TrimSpace(node.Issue.Status))
	switch status {
	case "in_progress":
		return sortPriorityInProgress, pickTimestamp(node.Issue.UpdatedAt, node.Issue.CreatedAt)
	case "closed":
		return sortPriorityClosed, pickTimestamp(node.Issue.ClosedAt, node.Issue.UpdatedAt, node.Issue.CreatedAt)
	}
	if status == "open" && !node.IsBlocked {
		return sortPriorityReady, pickTimestamp(node.Issue.CreatedAt)
	}
	return sortPriorityOpen, pickTimestamp(node.Issue.CreatedAt)
}

func pickTimestamp(values ...string) time.Time {
	for _, v := range values {
		if ts, err := time.Parse(time.RFC3339, strings.TrimSpace(v)); err == nil {
			return ts
		}
	}
	return distantFuture
}

func sortNodes(nodes []*Node) {
	sort.SliceStable(nodes, func(i, j int) bool {
		a, b := nodes[i], nodes[j]
		if a.SortPriority != b.SortPriority {
			return a.SortPriority < b.SortPriority
		}
		if !a.SortTimestamp.Equal(b.SortTimestamp) {
			return a.SortTimestamp.Before(b.SortTimestamp)
		}
		return a.Issue.ID < b.Issue.ID
	})
}
