package ui

import "abacus/internal/graph"

// ChildInfo captures descendant metadata for display in the delete overlay.
type ChildInfo struct {
	ID    string
	Title string
	Depth int
}

// collectChildInfo returns flattened child metadata and a list of descendant IDs
// for cascading delete operations. Nodes are deduplicated to avoid infinite loops
// with multi-parent relationships.
func collectChildInfo(node *graph.Node) ([]ChildInfo, []string) {
	if node == nil {
		return nil, nil
	}

	infos := []ChildInfo{}
	ids := []string{}
	seen := map[string]bool{}

	var walk func(children []*graph.Node, depth int)
	walk = func(children []*graph.Node, depth int) {
		for _, child := range children {
			if child == nil {
				continue
			}
			id := child.Issue.ID
			if seen[id] {
				continue
			}
			seen[id] = true
			infos = append(infos, ChildInfo{ID: id, Title: child.Issue.Title, Depth: depth})
			ids = append(ids, id)
			if len(child.Children) > 0 {
				walk(child.Children, depth+1)
			}
		}
	}

	walk(node.Children, 0)
	return infos, ids
}

func childWord(count int) string {
	if count == 1 {
		return "child"
	}
	return "children"
}
