package ui

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"abacus/internal/beads"
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

func TestLoadDataDoesNotPreloadComments(t *testing.T) {
	// Comments are now loaded in background after TUI starts (ab-fkyz)
	// This test verifies loadData doesn't block on comment loading
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
	// Comments should NOT be loaded during loadData - they're loaded in background
	if mock.CommentsCallCount != 0 {
		t.Fatalf("expected no comments calls during loadData, got %d", mock.CommentsCallCount)
	}
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

	// Comments are now loaded in background (ab-fkyz), so no comment loading stages
	want := []StartupStage{
		StartupStageLoadingIssues, // "Loading issues..."
		StartupStageLoadingIssues, // "Loaded X issues"
		StartupStageBuildingGraph, // "Building dependency graph..."
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
