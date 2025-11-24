package ui

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"abacus/internal/beads"
	"abacus/internal/graph"
)

func TestFetchFullIssuesBatchesCalls(t *testing.T) {
	mock := beads.NewMockClient()
	mock.ListFn = func(ctx context.Context) ([]beads.LiteIssue, error) {
		issues := make([]beads.LiteIssue, 25)
		for i := 0; i < 25; i++ {
			issues[i] = beads.LiteIssue{ID: fmt.Sprintf("ab-%03d", i+1)}
		}
		return issues, nil
	}
	mock.ShowFn = func(ctx context.Context, ids []string) ([]beads.FullIssue, error) {
		batch := make([]beads.FullIssue, len(ids))
		for i, id := range ids {
			batch[i] = beads.FullIssue{
				ID:        id,
				Title:     "Issue " + id,
				CreatedAt: "2024-01-01T00:00:00Z",
				UpdatedAt: "2024-01-01T00:00:00Z",
			}
		}
		return batch, nil
	}

	issues, err := fetchFullIssues(context.Background(), mock)
	if err != nil {
		t.Fatalf("fetchFullIssues returned error: %v", err)
	}
	if len(issues) != 25 {
		t.Fatalf("expected 25 issues, got %d", len(issues))
	}
	if mock.ShowCallCount != 2 {
		t.Fatalf("expected 2 show calls, got %d", mock.ShowCallCount)
	}
	if got := len(mock.ShowCallArgs[0]); got != 20 {
		t.Fatalf("expected first batch len 20, got %d", got)
	}
	if got := len(mock.ShowCallArgs[1]); got != 5 {
		t.Fatalf("expected second batch len 5, got %d", got)
	}
}

func TestLoadDataPreloadsComments(t *testing.T) {
	mock := beads.NewMockClient()
	mock.ListFn = func(ctx context.Context) ([]beads.LiteIssue, error) {
		return []beads.LiteIssue{
			{ID: "ab-001"},
			{ID: "ab-002"},
		}, nil
	}
	mock.ShowFn = func(ctx context.Context, ids []string) ([]beads.FullIssue, error) {
		batch := make([]beads.FullIssue, len(ids))
		for i, id := range ids {
			batch[i] = beads.FullIssue{
				ID:        id,
				Title:     "Issue " + id,
				Status:    "open",
				CreatedAt: fmt.Sprintf("2024-01-%02dT00:00:00Z", i+1),
				UpdatedAt: fmt.Sprintf("2024-01-%02dT00:00:00Z", i+1),
			}
		}
		return batch, nil
	}
	mock.CommentsFn = func(ctx context.Context, issueID string) ([]beads.Comment, error) {
		return []beads.Comment{
			{ID: 1, IssueID: issueID, Author: "tester", Text: "hi", CreatedAt: "2024-01-01T00:00:00Z"},
		}, nil
	}

	roots, err := loadData(context.Background(), mock)
	if err != nil {
		t.Fatalf("loadData returned error: %v", err)
	}
	if len(roots) != 2 {
		t.Fatalf("expected 2 root nodes, got %d", len(roots))
	}
	if mock.CommentsCallCount != 2 {
		t.Fatalf("expected comments called twice, got %d", mock.CommentsCallCount)
	}

	var check func(nodes []*graph.Node)
	check = func(nodes []*graph.Node) {
		for _, n := range nodes {
			if !n.CommentsLoaded {
				t.Fatalf("node %s not marked comments loaded", n.Issue.ID)
			}
			if n.CommentError != "" {
				t.Fatalf("unexpected comment error for %s: %s", n.Issue.ID, n.CommentError)
			}
			if len(n.Issue.Comments) != 1 {
				t.Fatalf("expected exactly one comment for %s", n.Issue.ID)
			}
			check(n.Children)
		}
	}
	check(roots)
}

func TestLoadDataReturnsErrorWhenNoIssues(t *testing.T) {
	mock := beads.NewMockClient()
	mock.ListFn = func(ctx context.Context) ([]beads.LiteIssue, error) {
		return nil, nil
	}
	if _, err := loadData(context.Background(), mock); !errors.Is(err, ErrNoIssues) {
		t.Fatalf("expected ErrNoIssues, got %v", err)
	}
}
