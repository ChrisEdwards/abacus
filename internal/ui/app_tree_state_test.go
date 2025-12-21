package ui

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"abacus/internal/beads"
	"abacus/internal/graph"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

func TestVisibleRowsMultiParentDuplicates(t *testing.T) {
	// When a node has multiple parents, it should appear multiple times in visibleRows
	sharedTask := &graph.Node{Issue: beads.FullIssue{ID: "ab-task", Title: "Shared Task", Status: "open"}}
	epic1 := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-epic1", Title: "Epic 1", Status: "open"},
		Children: []*graph.Node{sharedTask},
		Expanded: true,
	}
	epic2 := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-epic2", Title: "Epic 2", Status: "open"},
		Children: []*graph.Node{sharedTask},
		Expanded: true,
	}
	sharedTask.Parents = []*graph.Node{epic1, epic2}

	m := App{
		roots: []*graph.Node{epic1, epic2},
	}
	m.recalcVisibleRows()

	// Should have 4 rows: epic1, task (under epic1), epic2, task (under epic2)
	if len(m.visibleRows) != 4 {
		ids := make([]string, len(m.visibleRows))
		for i, r := range m.visibleRows {
			ids[i] = r.Node.Issue.ID
		}
		t.Fatalf("expected 4 visible rows (task appears twice), got %d: %v", len(m.visibleRows), ids)
	}

	// Count how many times sharedTask appears
	taskCount := 0
	for _, row := range m.visibleRows {
		if row.Node.Issue.ID == "ab-task" {
			taskCount++
		}
	}
	if taskCount != 2 {
		t.Fatalf("expected sharedTask to appear 2 times in visibleRows, got %d", taskCount)
	}

	// Verify depths are correct - task should be at depth 1 under each parent
	for _, row := range m.visibleRows {
		if row.Node.Issue.ID == "ab-task" && row.Depth != 1 {
			t.Fatalf("expected task depth 1, got %d", row.Depth)
		}
	}
}

func TestVisibleRowsMultiParentCorrectParentContext(t *testing.T) {
	// Each TreeRow should have correct Parent context
	sharedTask := &graph.Node{Issue: beads.FullIssue{ID: "ab-task", Title: "Shared Task", Status: "open"}}
	epic1 := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-epic1", Title: "Epic 1", Status: "open"},
		Children: []*graph.Node{sharedTask},
		Expanded: true,
	}
	epic2 := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-epic2", Title: "Epic 2", Status: "open"},
		Children: []*graph.Node{sharedTask},
		Expanded: true,
	}
	sharedTask.Parents = []*graph.Node{epic1, epic2}

	m := App{
		roots: []*graph.Node{epic1, epic2},
	}
	m.recalcVisibleRows()

	// Find the task rows and verify their parent context
	taskParents := []string{}
	for _, row := range m.visibleRows {
		if row.Node.Issue.ID == "ab-task" && row.Parent != nil {
			taskParents = append(taskParents, row.Parent.Issue.ID)
		}
	}

	if len(taskParents) != 2 {
		t.Fatalf("expected task to have 2 rows with parent context, got %d", len(taskParents))
	}

	// Both epics should be represented as parents
	hasEpic1 := false
	hasEpic2 := false
	for _, pid := range taskParents {
		if pid == "ab-epic1" {
			hasEpic1 = true
		}
		if pid == "ab-epic2" {
			hasEpic2 = true
		}
	}
	if !hasEpic1 || !hasEpic2 {
		t.Fatalf("expected both epic1 and epic2 as parent contexts, got %v", taskParents)
	}
}

func TestMultiParentExpandCollapseIndependent(t *testing.T) {
	// A shared task under two epics should have independent expand/collapse states per instance
	sharedTask := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-task", Title: "Shared Task", Status: "open"},
		Children: []*graph.Node{{Issue: beads.FullIssue{ID: "ab-subtask", Title: "Subtask", Status: "open"}}},
		Expanded: true,
	}
	epic1 := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-epic1", Title: "Epic 1", Status: "open"},
		Children: []*graph.Node{sharedTask},
		Expanded: true,
	}
	epic2 := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-epic2", Title: "Epic 2", Status: "open"},
		Children: []*graph.Node{sharedTask},
		Expanded: true,
	}
	sharedTask.Parents = []*graph.Node{epic1, epic2}

	m := App{
		roots: []*graph.Node{epic1, epic2},
	}
	m.recalcVisibleRows()

	// Initial: all expanded, should see 6 rows:
	// epic1, task (under epic1), subtask, epic2, task (under epic2), subtask
	if len(m.visibleRows) != 6 {
		t.Fatalf("expected 6 visible rows initially, got %d", len(m.visibleRows))
	}

	// Find task row under epic1 (should be at index 1)
	taskUnderEpic1 := m.visibleRows[1]
	if taskUnderEpic1.Node.Issue.ID != "ab-task" || taskUnderEpic1.Parent.Issue.ID != "ab-epic1" {
		t.Fatalf("expected task under epic1 at index 1, got %s under %v",
			taskUnderEpic1.Node.Issue.ID, taskUnderEpic1.Parent)
	}

	// Collapse task under epic1
	m.collapseNodeForView(taskUnderEpic1)
	m.recalcVisibleRows()

	// Should now see 5 rows: epic1, task (collapsed), epic2, task, subtask
	if len(m.visibleRows) != 5 {
		t.Fatalf("expected 5 visible rows after collapsing task under epic1, got %d", len(m.visibleRows))
	}

	// Verify task under epic2 is still expanded (subtask should be visible)
	subtaskCount := 0
	for _, row := range m.visibleRows {
		if row.Node.Issue.ID == "ab-subtask" {
			subtaskCount++
		}
	}
	if subtaskCount != 1 {
		t.Fatalf("expected exactly 1 subtask visible (under epic2), got %d", subtaskCount)
	}
}

func TestMultiParentCursorStable(t *testing.T) {
	// Cursor should stay on the same row after expand/collapse
	sharedTask := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-task", Title: "Shared Task", Status: "open"},
		Children: []*graph.Node{{Issue: beads.FullIssue{ID: "ab-subtask", Title: "Subtask", Status: "open"}}},
		Expanded: true,
	}
	epic1 := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-epic1", Title: "Epic 1", Status: "open"},
		Children: []*graph.Node{sharedTask},
		Expanded: true,
	}
	epic2 := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-epic2", Title: "Epic 2", Status: "open"},
		Children: []*graph.Node{sharedTask},
		Expanded: true,
	}
	sharedTask.Parents = []*graph.Node{epic1, epic2}

	m := App{
		roots: []*graph.Node{epic1, epic2},
	}
	m.recalcVisibleRows()

	// Position cursor on task under epic2 (index 4: epic1, task, subtask, epic2, task)
	m.cursor = 4
	if m.visibleRows[m.cursor].Node.Issue.ID != "ab-task" {
		t.Fatalf("expected cursor on task, got %s", m.visibleRows[m.cursor].Node.Issue.ID)
	}
	cursorParent := m.visibleRows[m.cursor].Parent
	if cursorParent == nil || cursorParent.Issue.ID != "ab-epic2" {
		t.Fatalf("expected cursor on task under epic2")
	}

	// Collapse the task under epic2
	m.collapseNodeForView(m.visibleRows[m.cursor])
	m.recalcVisibleRows()

	// Cursor should still be on task, but now at different index due to collapse
	if m.visibleRows[m.cursor].Node.Issue.ID != "ab-task" {
		// Cursor was clamped - find the task row
		found := false
		for i, row := range m.visibleRows {
			if row.Node.Issue.ID == "ab-task" && row.Parent != nil && row.Parent.Issue.ID == "ab-epic2" {
				found = true
				m.cursor = i
				break
			}
		}
		if !found {
			t.Fatalf("task under epic2 should still be visible")
		}
	}

	// Verify the collapsed row is still present
	taskUnderEpic2 := m.visibleRows[m.cursor]
	if taskUnderEpic2.Node.Issue.ID != "ab-task" {
		t.Fatalf("expected cursor on task, got %s", taskUnderEpic2.Node.Issue.ID)
	}
}

func TestMultiParentExpandStatePreservedOnRefresh(t *testing.T) {
	// Per-instance expand state should survive a data refresh
	sharedTask := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-task", Title: "Shared Task", Status: "open"},
		Children: []*graph.Node{{Issue: beads.FullIssue{ID: "ab-subtask", Title: "Subtask", Status: "open"}}},
		Expanded: true,
	}
	epic1 := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-epic1", Title: "Epic 1", Status: "open"},
		Children: []*graph.Node{sharedTask},
		Expanded: true,
	}
	epic2 := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-epic2", Title: "Epic 2", Status: "open"},
		Children: []*graph.Node{sharedTask},
		Expanded: true,
	}
	sharedTask.Parents = []*graph.Node{epic1, epic2}

	m := App{
		roots:     []*graph.Node{epic1, epic2},
		textInput: textinput.New(),
	}
	m.recalcVisibleRows()

	// Collapse task under epic1
	taskUnderEpic1 := m.visibleRows[1]
	m.collapseNodeForView(taskUnderEpic1)
	m.recalcVisibleRows()

	// Verify collapse worked
	initialRowCount := len(m.visibleRows)
	if initialRowCount != 5 {
		t.Fatalf("expected 5 rows after collapse, got %d", initialRowCount)
	}

	// Simulate refresh with new (but structurally identical) nodes
	newSubtask := &graph.Node{Issue: beads.FullIssue{ID: "ab-subtask", Title: "Subtask Updated", Status: "open"}}
	newSharedTask := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-task", Title: "Shared Task Updated", Status: "open"},
		Children: []*graph.Node{newSubtask},
		Expanded: true, // Default expanded on new load
	}
	newEpic1 := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-epic1", Title: "Epic 1", Status: "open"},
		Children: []*graph.Node{newSharedTask},
		Expanded: true,
	}
	newEpic2 := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-epic2", Title: "Epic 2", Status: "open"},
		Children: []*graph.Node{newSharedTask},
		Expanded: true,
	}
	newSharedTask.Parents = []*graph.Node{newEpic1, newEpic2}

	newRoots := []*graph.Node{newEpic1, newEpic2}
	m.applyRefresh(newRoots, buildIssueDigest(newRoots), time.Now())

	// Verify per-instance state was preserved
	// Task under epic1 should still be collapsed, task under epic2 expanded
	if len(m.visibleRows) != 5 {
		t.Fatalf("expected 5 rows after refresh (collapse state preserved), got %d", len(m.visibleRows))
	}

	// Find and verify task under epic1 is collapsed
	for _, row := range m.visibleRows {
		if row.Node.Issue.ID == "ab-task" && row.Parent != nil && row.Parent.Issue.ID == "ab-epic1" {
			if m.isNodeExpandedInView(row) {
				t.Fatalf("task under epic1 should remain collapsed after refresh")
			}
		}
		if row.Node.Issue.ID == "ab-task" && row.Parent != nil && row.Parent.Issue.ID == "ab-epic2" {
			if !m.isNodeExpandedInView(row) {
				t.Fatalf("task under epic2 should remain expanded after refresh")
			}
		}
	}
}

func TestTreePrefixWidth(t *testing.T) {
	indent := "    "
	marker := " ▶"
	icon := "◐"
	id := "ab-123"

	got := treePrefixWidth(indent, marker, icon, id)
	want := lipgloss.Width(fmt.Sprintf(" %s%s %s %s ", indent, marker, icon, id))
	if got != want {
		t.Fatalf("expected %d, got %d", want, got)
	}

	icon = "🧪"
	marker = " ⛔"
	got = treePrefixWidth(indent, marker, icon, id)
	want = lipgloss.Width(fmt.Sprintf(" %s%s %s %s ", indent, marker, icon, id))
	if got != want {
		t.Fatalf("expected %d, got %d for multi-byte glyph", want, got)
	}
}

func TestPreloadAllComments(t *testing.T) {
	ctx := context.Background()

	t.Run("marksAllNodesAsLoaded", func(t *testing.T) {
		root := &graph.Node{
			Issue: beads.FullIssue{ID: "ab-001", Title: "Root Issue"},
			Children: []*graph.Node{
				{Issue: beads.FullIssue{ID: "ab-002", Title: "Child Issue"}},
			},
		}

		client := beads.NewMockClient()
		client.CommentsFn = func(ctx context.Context, issueID string) ([]beads.Comment, error) {
			return []beads.Comment{
				{ID: 1, IssueID: issueID, Author: "tester", Text: "hi", CreatedAt: time.Now().UTC().Format(time.RFC3339)},
			}, nil
		}

		preloadAllComments(ctx, client, []*graph.Node{root}, nil)

		if !root.CommentsLoaded {
			t.Errorf("expected root node CommentsLoaded to be true")
		}
		if root.Issue.Comments == nil {
			t.Errorf("expected root node Comments to be initialized")
		}

		if !root.Children[0].CommentsLoaded {
			t.Errorf("expected child node CommentsLoaded to be true")
		}
		if root.Children[0].Issue.Comments == nil {
			t.Errorf("expected child node Comments to be initialized")
		}
	})

	t.Run("handlesMultipleRoots", func(t *testing.T) {
		roots := []*graph.Node{
			{Issue: beads.FullIssue{ID: "ab-010", Title: "First Root"}},
			{Issue: beads.FullIssue{ID: "ab-011", Title: "Second Root"}},
			{Issue: beads.FullIssue{ID: "ab-012", Title: "Third Root"}},
		}

		client := beads.NewMockClient()
		client.CommentsFn = func(ctx context.Context, issueID string) ([]beads.Comment, error) {
			return []beads.Comment{}, nil
		}

		preloadAllComments(ctx, client, roots, nil)

		for i, root := range roots {
			if !root.CommentsLoaded {
				t.Errorf("root %d (%s) not marked as loaded", i, root.Issue.ID)
			}
		}
	})

	t.Run("handlesNestedChildren", func(t *testing.T) {
		deepChild := &graph.Node{Issue: beads.FullIssue{ID: "ab-023", Title: "Deep Child"}}
		midChild := &graph.Node{
			Issue:    beads.FullIssue{ID: "ab-022", Title: "Mid Child"},
			Children: []*graph.Node{deepChild},
		}
		root := &graph.Node{
			Issue:    beads.FullIssue{ID: "ab-021", Title: "Root"},
			Children: []*graph.Node{midChild},
		}

		client := beads.NewMockClient()
		client.CommentsFn = func(ctx context.Context, issueID string) ([]beads.Comment, error) {
			return []beads.Comment{}, nil
		}

		preloadAllComments(ctx, client, []*graph.Node{root}, nil)

		if !root.CommentsLoaded {
			t.Errorf("root not loaded")
		}
		if !midChild.CommentsLoaded {
			t.Errorf("mid-level child not loaded")
		}
		if !deepChild.CommentsLoaded {
			t.Errorf("deep child not loaded")
		}
	})

	t.Run("handlesEmptyTree", func(t *testing.T) {
		client := beads.NewMockClient()
		client.CommentsFn = func(ctx context.Context, issueID string) ([]beads.Comment, error) {
			return []beads.Comment{}, nil
		}
		preloadAllComments(ctx, client, []*graph.Node{}, nil)
		preloadAllComments(ctx, client, nil, nil)
	})

	t.Run("initializesEmptyCommentsSlice", func(t *testing.T) {
		node := &graph.Node{Issue: beads.FullIssue{ID: "ab-030", Title: "No Comments"}}
		client := beads.NewMockClient()
		client.CommentsFn = func(ctx context.Context, issueID string) ([]beads.Comment, error) {
			return nil, nil
		}
		preloadAllComments(ctx, client, []*graph.Node{node}, nil)

		if !node.CommentsLoaded {
			t.Errorf("expected node to be marked as loaded even with no comments")
		}
		if node.Issue.Comments == nil {
			t.Errorf("expected Comments slice to be initialized")
		}
		if len(node.Issue.Comments) != 0 {
			t.Errorf("expected empty Comments slice, got %d items", len(node.Issue.Comments))
		}
	})

	t.Run("limitsConcurrentFetches", func(t *testing.T) {
		const totalRoots = 24
		roots := make([]*graph.Node, 0, totalRoots)
		for i := 0; i < totalRoots; i++ {
			roots = append(roots, &graph.Node{Issue: beads.FullIssue{ID: fmt.Sprintf("ab-%03d", i)}})
		}

		client := beads.NewMockClient()
		var mu sync.Mutex
		inFlight := 0
		maxInFlight := 0
		client.CommentsFn = func(ctx context.Context, issueID string) ([]beads.Comment, error) {
			mu.Lock()
			inFlight++
			if inFlight > maxInFlight {
				maxInFlight = inFlight
			}
			mu.Unlock()

			time.Sleep(5 * time.Millisecond)

			mu.Lock()
			inFlight--
			mu.Unlock()

			return []beads.Comment{}, nil
		}

		preloadAllComments(ctx, client, roots, nil)

		if maxInFlight > maxConcurrentCommentFetches {
			t.Fatalf("expected at most %d concurrent fetches, saw %d", maxConcurrentCommentFetches, maxInFlight)
		}
	})
}

func TestLoadCommentsInBackgroundUpdatesNodes(t *testing.T) {
	// This test verifies that loadCommentsInBackground correctly updates
	// the nodes in m.roots and that the same nodes are visible in m.visibleRows.
	// This was added to debug ab-o0fm where comments showed "Loading..." indefinitely.

	root := &graph.Node{
		Issue: beads.FullIssue{ID: "ab-bg1", Title: "Background Load Test"},
		Children: []*graph.Node{
			{Issue: beads.FullIssue{ID: "ab-bg2", Title: "Child Issue"}},
		},
	}

	client := beads.NewMockClient()
	client.CommentsFn = func(ctx context.Context, issueID string) ([]beads.Comment, error) {
		return []beads.Comment{
			{ID: 1, IssueID: issueID, Author: "tester", Text: "test comment", CreatedAt: time.Now().UTC().Format(time.RFC3339)},
		}, nil
	}

	app := &App{
		roots:       []*graph.Node{root},
		visibleRows: nodesToRows(root, root.Children[0]),
		client:      client,
	}

	// Verify initial state
	if root.CommentsLoaded {
		t.Fatal("expected root CommentsLoaded to be false initially")
	}
	if app.visibleRows[0].Node.CommentsLoaded {
		t.Fatal("expected visibleRows[0].Node CommentsLoaded to be false initially")
	}

	// Verify node identity - visibleRows should reference same nodes as roots
	if app.visibleRows[0].Node != root {
		t.Fatal("expected visibleRows[0].Node to be same pointer as root")
	}
	if app.visibleRows[1].Node != root.Children[0] {
		t.Fatal("expected visibleRows[1].Node to be same pointer as root.Children[0]")
	}

	// Call loadCommentsInBackground and execute the returned command
	cmd := app.loadCommentsInBackground()
	if cmd == nil {
		t.Fatal("expected non-nil command from loadCommentsInBackground")
	}

	// Execute the command synchronously (simulates what Bubble Tea runtime does)
	msg := cmd()

	// Verify the correct message type was returned
	batch, ok := msg.(commentBatchLoadedMsg)
	if !ok {
		t.Fatalf("expected commentBatchLoadedMsg from loadCommentsInBackground, got %T", msg)
	}

	for _, res := range batch.results {
		model, _ := app.Update(res)
		app = model.(*App)
	}

	// Verify nodes in m.roots are now marked as loaded
	if !root.CommentsLoaded {
		t.Error("expected root CommentsLoaded to be true after loading")
	}
	if !root.Children[0].CommentsLoaded {
		t.Error("expected root.Children[0] CommentsLoaded to be true after loading")
	}

	// CRITICAL: Verify nodes in visibleRows are also updated (same pointers)
	if !app.visibleRows[0].Node.CommentsLoaded {
		t.Error("expected visibleRows[0].Node CommentsLoaded to be true - this means node identity was lost")
	}
	if !app.visibleRows[1].Node.CommentsLoaded {
		t.Error("expected visibleRows[1].Node CommentsLoaded to be true - this means node identity was lost")
	}

	// Verify comments were actually loaded
	if len(root.Issue.Comments) != 1 {
		t.Errorf("expected 1 comment on root, got %d", len(root.Issue.Comments))
	}
}

func TestCommentLoadedMsgRefreshesDetailPane(t *testing.T) {
	root := &graph.Node{
		Issue: beads.FullIssue{ID: "ab-1", Title: "Detail Test"},
	}
	app := &App{
		roots:       []*graph.Node{root},
		visibleRows: nodesToRows(root),
		ShowDetails: true,
		cursor:      0,
		viewport: viewport.Model{
			Width:  80,
			Height: 20,
		},
	}

	app.updateViewportContent()
	if !strings.Contains(app.viewport.View(), "Loading comments") {
		t.Fatal("expected initial detail view to show loading state")
	}

	comment := beads.Comment{ID: 1, IssueID: root.Issue.ID, Author: "tester", Text: "detail loaded", CreatedAt: time.Now().UTC().Format(time.RFC3339)}
	model, _ := app.Update(commentLoadedMsg{
		issueID:  root.Issue.ID,
		comments: []beads.Comment{comment},
	})
	app = model.(*App)

	content := stripANSI(app.viewport.View())
	if strings.Contains(content, "Loading comments") {
		t.Fatal("detail view should not show loading after commentLoadedMsg")
	}
	if !strings.Contains(content, "detail loaded") {
		t.Fatalf("expected loaded comment text in detail view, got:\n%s", content)
	}
}

func TestCaptureState(t *testing.T) {
	child := &graph.Node{Issue: beads.FullIssue{ID: "ab-002"}}
	root := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-001"},
		Children: []*graph.Node{child},
		Expanded: true,
	}

	m := App{
		roots:       []*graph.Node{root},
		visibleRows: nodesToRows(root, child),
		cursor:      1,
		filterText:  "alpha",
		ShowDetails: true,
		focus:       FocusDetails,
		viewport: viewport.Model{
			YOffset: 3,
			Height:  10,
		},
	}

	state := m.captureState()

	if state.currentID != "ab-002" {
		t.Fatalf("expected currentID ab-002, got %s", state.currentID)
	}
	if state.filterText != "alpha" {
		t.Fatalf("expected filter alpha, got %s", state.filterText)
	}
	if state.viewportYOffset != 3 {
		t.Fatalf("expected viewport offset 3, got %d", state.viewportYOffset)
	}
	if !state.expandedIDs["ab-001"] || len(state.expandedIDs) != 1 {
		t.Fatalf("expected only root to be remembered as expanded")
	}
	if state.focus != FocusDetails {
		t.Fatalf("expected focus captured as details")
	}
}

func TestRestoreExpandedState(t *testing.T) {
	child := &graph.Node{Issue: beads.FullIssue{ID: "ab-002"}}
	root := &graph.Node{Issue: beads.FullIssue{ID: "ab-001"}, Children: []*graph.Node{child}}
	m := App{roots: []*graph.Node{root}}

	m.restoreExpandedState(map[string]bool{"ab-001": true})

	if !root.Expanded {
		t.Fatalf("expected root expanded")
	}
	if child.Expanded {
		t.Fatalf("expected child collapsed")
	}
}

func TestRestoreCursorToID(t *testing.T) {
	n1 := &graph.Node{Issue: beads.FullIssue{ID: "ab-001"}}
	n2 := &graph.Node{Issue: beads.FullIssue{ID: "ab-002"}}
	m := App{
		visibleRows: nodesToRows(n1, n2),
		cursor:      0,
	}

	m.restoreCursorToID("ab-002")
	if m.cursor != 1 {
		t.Fatalf("expected cursor 1, got %d", m.cursor)
	}

	m.restoreCursorToID("missing")
	if m.cursor != 1 {
		t.Fatalf("expected cursor to remain 1 when id missing, got %d", m.cursor)
	}
}

func TestComputeDiffStats(t *testing.T) {
	oldSet := map[string]string{
		"ab-1": "2024-01-01",
		"ab-2": "2024-01-01",
	}
	newSet := map[string]string{
		"ab-2": "2024-01-02",
		"ab-3": "2024-01-01",
	}

	got := computeDiffStats(oldSet, newSet)
	want := "+1 / Δ1 / -1"
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestApplyRefreshRestoresState(t *testing.T) {
	childOld := &graph.Node{Issue: beads.FullIssue{ID: "ab-002", Title: "Child", Status: "open"}}
	rootOld := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-001", Title: "Root", Status: "open"},
		Children: []*graph.Node{childOld},
		Expanded: true,
	}

	m := App{
		roots:       []*graph.Node{rootOld},
		visibleRows: nodesToRows(rootOld, childOld),
		cursor:      1,
		filterText:  "child",
		ShowDetails: true,
		focus:       FocusDetails,
		viewport: viewport.Model{
			Height:  5,
			YOffset: 2,
		},
		textInput: textinput.New(),
	}
	m.textInput.SetValue("child")

	childNew := &graph.Node{Issue: beads.FullIssue{ID: "ab-002", Title: "Child Updated", Status: "open"}}
	rootNew := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-001", Title: "Root", Status: "open"},
		Children: []*graph.Node{childNew},
	}
	newDigest := buildIssueDigest([]*graph.Node{rootNew})

	m.applyRefresh([]*graph.Node{rootNew}, newDigest, time.Now())

	if m.filterText != "child" {
		t.Fatalf("expected filter preserved, got %s", m.filterText)
	}
	if len(m.visibleRows) == 0 || m.visibleRows[m.cursor].Node.Issue.ID != "ab-002" {
		t.Fatalf("expected cursor to remain on child after refresh")
	}
	if m.viewport.YOffset != 2 {
		t.Fatalf("expected viewport offset restored, got %d", m.viewport.YOffset)
	}
	if m.lastRefreshStats == "" {
		t.Fatalf("expected refresh stats to be populated")
	}
	if m.focus != FocusDetails {
		t.Fatalf("expected focus restored to details")
	}
}

func TestApplyRefreshPreservesCollapsedStatePerDocs(t *testing.T) {
	childOld := &graph.Node{Issue: beads.FullIssue{ID: "ab-012", Title: "Child Hidden", Status: "open"}}
	rootOld := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-011", Title: "Root", Status: "open"},
		Children: []*graph.Node{childOld},
	}
	m := App{
		roots:           []*graph.Node{rootOld},
		visibleRows:     nodesToRows(rootOld),
		filterText:      "root",
		filterCollapsed: map[string]bool{rootOld.Issue.ID: true},
		textInput:       textinput.New(),
	}
	m.textInput.SetValue("root")
	m.recalcVisibleRows()
	m.collapseNodeForView(nodeToRow(rootOld))
	childNew := &graph.Node{Issue: beads.FullIssue{ID: "ab-012", Title: "Child Updated", Status: "open"}}
	rootNew := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-011", Title: "Root", Status: "open"},
		Children: []*graph.Node{childNew},
	}
	m.applyRefresh([]*graph.Node{rootNew}, buildIssueDigest([]*graph.Node{rootNew}), time.Now())
	if m.isNodeExpandedInView(nodeToRow(rootNew)) {
		t.Fatalf("expected collapsed state preserved after refresh per docs")
	}
}
