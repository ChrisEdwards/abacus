package ui

import (
	"testing"

	"abacus/internal/beads"
	"abacus/internal/graph"
)

func TestCollectCommentState(t *testing.T) {
	// Build a tree with comments loaded
	roots := []*graph.Node{
		{
			Issue: beads.FullIssue{
				ID:    "ab-001",
				Title: "Root Issue",
				Comments: []beads.Comment{
					{ID: 1, Text: "comment 1"},
					{ID: 2, Text: "comment 2"},
				},
			},
			CommentsLoaded: true,
			Children: []*graph.Node{
				{
					Issue: beads.FullIssue{
						ID:    "ab-002",
						Title: "Child Issue",
						Comments: []beads.Comment{
							{ID: 3, Text: "child comment"},
						},
					},
					CommentsLoaded: true,
				},
			},
		},
		{
			Issue: beads.FullIssue{
				ID:    "ab-003",
				Title: "Not loaded",
			},
			CommentsLoaded: false,
		},
	}

	state := collectCommentState(roots)

	// Should have entries for ab-001 and ab-002, but not ab-003
	if len(state) != 2 {
		t.Fatalf("expected 2 entries in state, got %d", len(state))
	}

	s1, ok := state["ab-001"]
	if !ok {
		t.Fatal("expected state for ab-001")
	}
	if len(s1.comments) != 2 {
		t.Fatalf("expected 2 comments for ab-001, got %d", len(s1.comments))
	}
	if !s1.commentsLoaded {
		t.Fatal("expected commentsLoaded=true for ab-001")
	}

	s2, ok := state["ab-002"]
	if !ok {
		t.Fatal("expected state for ab-002")
	}
	if len(s2.comments) != 1 {
		t.Fatalf("expected 1 comment for ab-002, got %d", len(s2.comments))
	}

	if _, ok := state["ab-003"]; ok {
		t.Fatal("should not have state for ab-003 (not loaded)")
	}
}

func TestCollectCommentStateWithError(t *testing.T) {
	roots := []*graph.Node{
		{
			Issue: beads.FullIssue{
				ID:    "ab-err",
				Title: "Error Issue",
			},
			CommentsLoaded: false,
			CommentError:   "failed to load",
		},
	}

	state := collectCommentState(roots)

	// Should capture the error state too
	if len(state) != 1 {
		t.Fatalf("expected 1 entry in state, got %d", len(state))
	}
	s, ok := state["ab-err"]
	if !ok {
		t.Fatal("expected state for ab-err")
	}
	if s.commentError != "failed to load" {
		t.Fatalf("expected error message, got %q", s.commentError)
	}
}

func TestTransferCommentState(t *testing.T) {
	// Old state with comments
	oldState := map[string]commentState{
		"ab-001": {
			comments: []beads.Comment{
				{ID: 1, Text: "preserved comment"},
			},
			commentsLoaded: true,
		},
		"ab-002": {
			commentError: "some error",
		},
	}

	// Fresh nodes without comments
	newRoots := []*graph.Node{
		{
			Issue: beads.FullIssue{ID: "ab-001", Title: "Issue 1"},
			Children: []*graph.Node{
				{Issue: beads.FullIssue{ID: "ab-002", Title: "Issue 2"}},
			},
		},
		{Issue: beads.FullIssue{ID: "ab-003", Title: "New Issue"}},
	}

	transferCommentState(newRoots, oldState)

	// Check ab-001 got its comments back
	if !newRoots[0].CommentsLoaded {
		t.Fatal("expected ab-001 to have CommentsLoaded=true")
	}
	if len(newRoots[0].Issue.Comments) != 1 {
		t.Fatalf("expected 1 comment for ab-001, got %d", len(newRoots[0].Issue.Comments))
	}
	if newRoots[0].Issue.Comments[0].Text != "preserved comment" {
		t.Fatalf("expected preserved comment, got %q", newRoots[0].Issue.Comments[0].Text)
	}

	// Check ab-002 got its error state back
	if newRoots[0].Children[0].CommentError != "some error" {
		t.Fatalf("expected error for ab-002, got %q", newRoots[0].Children[0].CommentError)
	}

	// Check ab-003 was not modified (not in old state)
	if newRoots[1].CommentsLoaded {
		t.Fatal("ab-003 should not have comments loaded")
	}
	if len(newRoots[1].Issue.Comments) != 0 {
		t.Fatal("ab-003 should have no comments")
	}
}

func TestCommentStatePreservedAcrossRefresh(t *testing.T) {
	// Simulate a full refresh cycle - this is an integration test of the
	// collect -> transfer flow that happens in applyRefresh

	// Start with nodes that have comments loaded
	originalRoots := []*graph.Node{
		{
			Issue: beads.FullIssue{
				ID:    "ab-001",
				Title: "Original Title",
				Comments: []beads.Comment{
					{ID: 1, Text: "comment A"},
					{ID: 2, Text: "comment B"},
				},
			},
			CommentsLoaded: true,
		},
	}

	// Collect comment state (this happens before replacing roots)
	savedState := collectCommentState(originalRoots)

	// Simulate a refresh - new nodes come in without comments
	refreshedRoots := []*graph.Node{
		{
			Issue: beads.FullIssue{
				ID:       "ab-001",
				Title:    "Updated Title", // Title changed but ID same
				Comments: nil,             // No comments on fresh node
			},
			CommentsLoaded: false,
		},
	}

	// Transfer state to new nodes (this happens after replacing roots)
	transferCommentState(refreshedRoots, savedState)

	// Verify comments were preserved
	if !refreshedRoots[0].CommentsLoaded {
		t.Fatal("CommentsLoaded should be true after transfer")
	}
	if len(refreshedRoots[0].Issue.Comments) != 2 {
		t.Fatalf("expected 2 comments preserved, got %d", len(refreshedRoots[0].Issue.Comments))
	}
	if refreshedRoots[0].Issue.Comments[0].Text != "comment A" {
		t.Fatalf("expected 'comment A', got %q", refreshedRoots[0].Issue.Comments[0].Text)
	}
}
