package ui

import (
	"context"
	"strings"

	"abacus/internal/domain"
	"abacus/internal/graph"
)

func (m *App) retryCommentsForCurrentNode() {
	if m.cursor >= len(m.visibleRows) {
		return
	}
	node := m.visibleRows[m.cursor]
	node.Issue.Comments = nil
	node.CommentsLoaded = false
	node.CommentError = ""
	if err := fetchCommentsForNode(context.Background(), m.client, node); err != nil {
		node.CommentError = err.Error()
	}
}

func (m *App) getStats() Stats {
	s := Stats{}

	var traverse func(nodes []*graph.Node)
	traverse = func(nodes []*graph.Node) {
		for _, n := range nodes {
			matches := m.filterText == "" || strings.Contains(strings.ToLower(n.Issue.Title), strings.ToLower(m.filterText))

			domainIssue, err := domain.NewIssueFromFull(n.Issue, n.IsBlocked)
			if matches {
				s.Total++
				if err != nil {
					if n.Issue.Status == "in_progress" {
						s.InProgress++
					} else if n.Issue.Status == "closed" {
						s.Closed++
					} else if n.IsBlocked {
						s.Blocked++
					} else {
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
