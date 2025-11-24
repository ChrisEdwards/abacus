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
			matches := m.matchesFilter(n)

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

func (m *App) matchesFilter(node *graph.Node) bool {
	textActive := strings.TrimSpace(m.filterText) != ""
	if !textActive && len(m.filterTokens) == 0 {
		return true
	}
	words := append([]string{}, m.filterFreeText...)
	if len(words) == 0 && len(m.filterTokens) == 0 && textActive {
		words = strings.Fields(strings.ToLower(m.filterText))
	}
	tokens := append([]SearchToken{}, m.filterTokens...)
	return nodeMatchesTokenFilter(words, tokens, node)
}

func nodeMatchesTokenFilter(words []string, tokens []SearchToken, node *graph.Node) bool {
	if len(words) == 0 && len(tokens) == 0 {
		return true
	}
	if len(tokens) > 0 && !matchesTokens(tokens, node) {
		return false
	}
	if len(words) == 0 {
		return true
	}
	title := strings.ToLower(node.Issue.Title)
	id := strings.ToLower(node.Issue.ID)
	for _, word := range words {
		word = strings.TrimSpace(strings.ToLower(word))
		if word == "" {
			continue
		}
		if !strings.Contains(title, word) && !strings.Contains(id, word) {
			return false
		}
	}
	return true
}
