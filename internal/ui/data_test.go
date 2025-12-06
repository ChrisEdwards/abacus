package ui

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"abacus/internal/beads"
	"abacus/internal/graph"
)

func TestLoadDataUsesExport(t *testing.T) {
	mock := beads.NewMockClient()
	mock.ExportFn = func(ctx context.Context) ([]beads.FullIssue, error) {
		return []beads.FullIssue{
			{ID: "ab-001", Title: "Issue 1", Status: "open", CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-01T00:00:00Z"},
			{ID: "ab-002", Title: "Issue 2", Status: "open", CreatedAt: "2024-01-02T00:00:00Z", UpdatedAt: "2024-01-02T00:00:00Z"},
		}, nil
	}
	mock.CommentsFn = func(ctx context.Context, issueID string) ([]beads.Comment, error) {
		return []beads.Comment{
			{ID: 1, IssueID: issueID, Author: "tester", Text: "hi", CreatedAt: "2024-01-01T00:00:00Z"},
		}, nil
	}

	roots, err := loadData(context.Background(), mock, nil)
	if err != nil {
		t.Fatalf("loadData returned error: %v", err)
	}
	if len(roots) != 2 {
		t.Fatalf("expected 2 root nodes, got %d", len(roots))
	}
	if mock.ExportCallCount != 1 {
		t.Fatalf("expected Export called once, got %d", mock.ExportCallCount)
	}
}

func TestLoadDataPreloadsComments(t *testing.T) {
	mock := beads.NewMockClient()
	mock.ExportFn = func(ctx context.Context) ([]beads.FullIssue, error) {
		return []beads.FullIssue{
			{ID: "ab-001", Title: "Issue 1", Status: "open", CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-01T00:00:00Z"},
			{ID: "ab-002", Title: "Issue 2", Status: "open", CreatedAt: "2024-01-02T00:00:00Z", UpdatedAt: "2024-01-02T00:00:00Z"},
		}, nil
	}
	mock.CommentsFn = func(ctx context.Context, issueID string) ([]beads.Comment, error) {
		return []beads.Comment{
			{ID: 1, IssueID: issueID, Author: "tester", Text: "hi", CreatedAt: "2024-01-01T00:00:00Z"},
		}, nil
	}

	roots, err := loadData(context.Background(), mock, nil)
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
	mock.ExportFn = func(ctx context.Context) ([]beads.FullIssue, error) {
		return nil, fmt.Errorf("bd export returned no issues")
	}
	if _, err := loadData(context.Background(), mock, nil); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadDataReturnsErrorForEmptyExport(t *testing.T) {
	mock := beads.NewMockClient()
	mock.ExportFn = func(ctx context.Context) ([]beads.FullIssue, error) {
		return []beads.FullIssue{}, nil
	}
	if _, err := loadData(context.Background(), mock, nil); !errors.Is(err, ErrNoIssues) {
		t.Fatalf("expected ErrNoIssues, got %v", err)
	}
}

func TestLoadDataReportsStartupStages(t *testing.T) {
	mock := beads.NewMockClient()
	mock.ExportFn = func(ctx context.Context) ([]beads.FullIssue, error) {
		return []beads.FullIssue{
			{ID: "ab-001", Title: "Issue 1", Status: "open", CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-01T00:00:00Z"},
		}, nil
	}
	mock.CommentsFn = func(ctx context.Context, issueID string) ([]beads.Comment, error) {
		return nil, nil
	}

	reporter := &recordingReporter{}
	if _, err := loadData(context.Background(), mock, reporter); err != nil {
		t.Fatalf("loadData returned error: %v", err)
	}

	// With Export, we have fewer stages since we don't have intermediate progress reporting
	want := []StartupStage{
		StartupStageLoadingIssues,  // "Loading issues..."
		StartupStageLoadingIssues,  // "Loaded X issues"
		StartupStageBuildingGraph,  // "Building dependency graph..."
		StartupStageOrganizingTree, // "Loading comments..."
		StartupStageOrganizingTree, // "Loading comments... X/Y" (progress)
	}
	if len(reporter.stages) != len(want) {
		t.Fatalf("expected %d stage events, got %d: %#v", len(want), len(reporter.stages), reporter.stages)
	}
	for i, stage := range want {
		if reporter.stages[i] != stage {
			t.Fatalf("stage[%d] = %v, want %v", i, reporter.stages[i], stage)
		}
	}
}

type recordingReporter struct {
	stages []StartupStage
}

func (r *recordingReporter) Stage(stage StartupStage, detail string) {
	r.stages = append(r.stages, stage)
}
