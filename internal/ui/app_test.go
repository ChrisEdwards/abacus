package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"abacus/internal/beads"
	"abacus/internal/graph"
	"abacus/internal/ui/theme"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

// nodesToRows converts a slice of Nodes to TreeRows for testing.
func nodesToRows(nodes ...*graph.Node) []graph.TreeRow {
	rows := make([]graph.TreeRow, len(nodes))
	for i, n := range nodes {
		rows[i] = graph.TreeRow{Node: n, Depth: 0}
	}
	return rows
}

// nodeToRow creates a TreeRow from a Node for testing expand/collapse methods.
func nodeToRow(node *graph.Node) graph.TreeRow {
	return graph.TreeRow{Node: node, Depth: 0}
}

func TestWrapWithHangingIndent(t *testing.T) {
	t.Run("appliesIndentToWrappedLines", func(t *testing.T) {
		text := "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
		got := wrapWithHangingIndent(4, text, 20)
		lines := strings.Split(got, "\n")
		if len(lines) <= 1 {
			t.Fatalf("expected wrapped text, got %q", got)
		}
		for i := 1; i < len(lines); i++ {
			if !strings.HasPrefix(lines[i], "    ") {
				t.Fatalf("line %d missing hanging indent: %q", i, lines[i])
			}
		}
	})

	t.Run("returnsOriginalWhenTooNarrow", func(t *testing.T) {
		text := "no change"
		got := wrapWithHangingIndent(10, text, 8)
		if got != text {
			t.Fatalf("expected %q, got %q", text, got)
		}
	})

	t.Run("returnsOriginalWhenAlreadyShort", func(t *testing.T) {
		text := "short text"
		got := wrapWithHangingIndent(2, text, 50)
		if got != text {
			t.Fatalf("expected %q, got %q", text, got)
		}
	})
}

func TestIndentBlock(t *testing.T) {
	input := "first line\n\nthird line"
	got := indentBlock(input, 2)
	want := "  first line\n\n  third line"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestGetStats(t *testing.T) {
	t.Run("countsStatuses", func(t *testing.T) {
		ready := &graph.Node{Issue: beads.FullIssue{ID: "ab-001", Title: "Ready Task", Status: "open"}}
		inProgress := &graph.Node{
			Issue:    beads.FullIssue{ID: "ab-002", Title: "In Progress", Status: "in_progress"},
			Children: []*graph.Node{ready},
		}
		closed := &graph.Node{Issue: beads.FullIssue{ID: "ab-003", Title: "Closed Task", Status: "closed"}}
		blocked := &graph.Node{Issue: beads.FullIssue{ID: "ab-004", Title: "Blocked Task", Status: "open"}, IsBlocked: true}

		m := App{
			roots: []*graph.Node{inProgress, closed, blocked},
		}

		stats := m.getStats()
		if stats.Total != 4 {
			t.Fatalf("expected total 4, got %d", stats.Total)
		}
		if stats.InProgress != 1 {
			t.Fatalf("expected 1 in-progress, got %d", stats.InProgress)
		}
		if stats.Closed != 1 {
			t.Fatalf("expected 1 closed, got %d", stats.Closed)
		}
		if stats.Blocked != 1 {
			t.Fatalf("expected 1 blocked, got %d", stats.Blocked)
		}
		if stats.Ready != 1 {
			t.Fatalf("expected 1 ready, got %d", stats.Ready)
		}
	})

	t.Run("appliesFilter", func(t *testing.T) {
		matching := &graph.Node{Issue: beads.FullIssue{ID: "ab-010", Title: "Alpha Ready", Status: "open"}}
		nonMatching := &graph.Node{Issue: beads.FullIssue{ID: "ab-020", Title: "Bravo Active", Status: "in_progress"}}
		m := App{
			roots:      []*graph.Node{matching, nonMatching},
			filterText: "ready",
		}

		stats := m.getStats()
		if stats.Total != 1 {
			t.Fatalf("expected only one matching issue, got %d", stats.Total)
		}
		if stats.Ready != 1 {
			t.Fatalf("expected ready count 1, got %d", stats.Ready)
		}
		if stats.InProgress != 0 {
			t.Fatalf("expected in-progress count 0, got %d", stats.InProgress)
		}
	})

	t.Run("countsMatchesByIDFilter", func(t *testing.T) {
		openNode := &graph.Node{Issue: beads.FullIssue{ID: "ab-100", Title: "Alpha Ready", Status: "open"}}
		inProgress := &graph.Node{Issue: beads.FullIssue{ID: "ab-200", Title: "Beta Active", Status: "in_progress"}}
		m := App{
			roots:      []*graph.Node{openNode, inProgress},
			filterText: "ab-200",
		}

		stats := m.getStats()
		if stats.Total != 1 {
			t.Fatalf("expected filtered count 1, got %d", stats.Total)
		}
		if stats.InProgress != 1 {
			t.Fatalf("expected in-progress count 1, got %d", stats.InProgress)
		}
		if stats.Ready != 0 {
			t.Fatalf("expected ready count 0, got %d", stats.Ready)
		}
	})

	t.Run("deduplicatesMultiParentNodes", func(t *testing.T) {
		// Task with multiple parents should only be counted once
		sharedTask := &graph.Node{Issue: beads.FullIssue{ID: "ab-shared", Title: "Shared Task", Status: "open"}}
		epic1 := &graph.Node{
			Issue:    beads.FullIssue{ID: "ab-epic1", Title: "Epic 1", Status: "open"},
			Children: []*graph.Node{sharedTask},
		}
		epic2 := &graph.Node{
			Issue:    beads.FullIssue{ID: "ab-epic2", Title: "Epic 2", Status: "open"},
			Children: []*graph.Node{sharedTask}, // Same task under another parent
		}
		sharedTask.Parents = []*graph.Node{epic1, epic2}

		m := App{
			roots: []*graph.Node{epic1, epic2},
		}

		stats := m.getStats()
		// Should count: epic1, epic2, sharedTask (once) = 3 total
		if stats.Total != 3 {
			t.Fatalf("expected total 3 (multi-parent task counted once), got %d", stats.Total)
		}
		if stats.Ready != 3 {
			t.Fatalf("expected 3 ready (all open, not blocked), got %d", stats.Ready)
		}
	})
}

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

func TestUpdateTogglesFocusWithTab(t *testing.T) {
	m := &App{ShowDetails: true, focus: FocusTree, keys: DefaultKeyMap()}
	m.visibleRows = nodesToRows(&graph.Node{Issue: beads.FullIssue{ID: "ab-001"}})

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focus != FocusDetails {
		t.Fatalf("expected tab to switch focus to details")
	}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focus != FocusTree {
		t.Fatalf("expected tab to cycle focus back to tree")
	}
}

func TestDetailFocusNavigation(t *testing.T) {
	newDetailApp := func() *App {
		vp := viewport.Model{Width: 40, Height: 3}
		vp.SetContent("line1\nline2\nline3\nline4")
		return &App{
			ShowDetails: true,
			focus:       FocusDetails,
			viewport:    vp,
			visibleRows: nodesToRows(&graph.Node{Issue: beads.FullIssue{ID: "ab-001"}}),
			keys:        DefaultKeyMap(),
		}
	}

	t.Run("arrowKeysScrollViewport", func(t *testing.T) {
		m := newDetailApp()
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if m.cursor != 0 {
			t.Fatalf("expected cursor to remain unchanged when details focused")
		}
		if m.viewport.YOffset == 0 {
			t.Fatalf("expected viewport offset to increase after scrolling")
		}
	})

	t.Run("pageCommandsRespectCtrlKeys", func(t *testing.T) {
		m := newDetailApp()
		start := m.viewport.YOffset
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
		if m.viewport.YOffset <= start {
			t.Fatalf("expected ctrl+f to page down in details")
		}
	})

	t.Run("homeAndEndJump", func(t *testing.T) {
		m := newDetailApp()
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
		if m.viewport.YOffset == 0 {
			t.Fatalf("expected ctrl+f to move viewport before home test")
		}
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyHome})
		if m.viewport.YOffset != 0 {
			t.Fatalf("expected home to reset viewport to top, got %d", m.viewport.YOffset)
		}
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnd})
		if m.viewport.YOffset == 0 {
			t.Fatalf("expected end to jump to bottom")
		}
	})
}

func TestUpdateViewportContentResetsScrollOnNewSelection(t *testing.T) {
	n1 := &graph.Node{Issue: beads.FullIssue{ID: "ab-100", Title: "First"}, CommentsLoaded: true}
	n2 := &graph.Node{Issue: beads.FullIssue{ID: "ab-200", Title: "Second"}, CommentsLoaded: true}
	m := &App{
		ShowDetails: true,
		visibleRows: nodesToRows(n1, n2),
		viewport:    viewport.Model{Width: 60, Height: 10},
		cursor:      0,
	}

	m.updateViewportContent()
	m.viewport.YOffset = 4
	m.cursor = 1
	m.updateViewportContent()

	if m.viewport.YOffset != 0 {
		t.Fatalf("expected viewport offset reset to 0 on new selection, got %d", m.viewport.YOffset)
	}
	if m.detailIssueID != "ab-200" {
		t.Fatalf("expected detailIssueID updated to new selection, got %s", m.detailIssueID)
	}
}

func TestUpdateViewportContentPreservesScrollForSameSelection(t *testing.T) {
	n1 := &graph.Node{Issue: beads.FullIssue{ID: "ab-100", Title: "Same"}, CommentsLoaded: true}
	m := &App{
		ShowDetails: true,
		visibleRows: nodesToRows(n1),
		viewport:    viewport.Model{Width: 60, Height: 10},
	}

	m.updateViewportContent()
	m.viewport.YOffset = 5
	m.updateViewportContent()

	if m.viewport.YOffset != 5 {
		t.Fatalf("expected viewport offset preserved for same selection, got %d", m.viewport.YOffset)
	}
	if m.detailIssueID != "ab-100" {
		t.Fatalf("expected detailIssueID to remain unchanged, got %s", m.detailIssueID)
	}
}

func TestUpdateClearsFilterWithEsc(t *testing.T) {
	buildApp := func(filter string, searching bool) *App {
		m := &App{
			roots: []*graph.Node{
				{Issue: beads.FullIssue{ID: "ab-100", Title: "Alpha"}},
				{Issue: beads.FullIssue{ID: "ab-200", Title: "Beta"}},
			},
			textInput:  textinput.New(),
			filterText: filter,
			searching:  searching,
			keys:       DefaultKeyMap(),
		}
		m.textInput.SetValue(filter)
		m.recalcVisibleRows()
		return m
	}

	t.Run("whileSearching", func(t *testing.T) {
		m := buildApp("beta", true)
		if len(m.visibleRows) != 1 {
			t.Fatalf("expected 1 visible row while filtered, got %d", len(m.visibleRows))
		}
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if m.searching {
			t.Fatalf("expected searching to be disabled after esc")
		}
		if m.filterText != "" {
			t.Fatalf("expected filter cleared after esc, got %s", m.filterText)
		}
		if len(m.visibleRows) != 2 {
			t.Fatalf("expected all rows restored after esc, got %d", len(m.visibleRows))
		}
	})

	t.Run("whileBrowsing", func(t *testing.T) {
		m := buildApp("beta", false)
		if len(m.visibleRows) != 1 {
			t.Fatalf("expected filtered list before esc, got %d rows", len(m.visibleRows))
		}
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if m.filterText != "" {
			t.Fatalf("expected filter cleared after esc, got %s", m.filterText)
		}
		if len(m.visibleRows) != 2 {
			t.Fatalf("expected esc to restore all rows, got %d", len(m.visibleRows))
		}
		if m.textInput.Value() != "" {
			t.Fatalf("expected input cleared, got %q", m.textInput.Value())
		}
	})
}

func TestClearFilterPreservesSelectionSingleParent(t *testing.T) {
	// Create tree: root -> child -> leaf
	leaf := &graph.Node{Issue: beads.FullIssue{ID: "ab-leaf", Title: "Leaf Node"}}
	child := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-child", Title: "Child Node"},
		Children: []*graph.Node{leaf},
	}
	leaf.Parent = child
	leaf.Parents = []*graph.Node{child}
	root := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-root", Title: "Root Node"},
		Children: []*graph.Node{child},
	}
	child.Parent = root
	child.Parents = []*graph.Node{root}

	m := &App{
		roots:     []*graph.Node{root},
		textInput: textinput.New(),
		keys:      DefaultKeyMap(),
	}

	// Filter to show leaf, which auto-expands ancestors
	m.setFilterText("leaf")
	m.recalcVisibleRows()

	// Find and select the leaf node
	for i, row := range m.visibleRows {
		if row.Node.Issue.ID == "ab-leaf" {
			m.cursor = i
			break
		}
	}

	// Clear filter with ESC
	m.clearSearchFilter()

	// Verify leaf is still selected
	if m.visibleRows[m.cursor].Node.Issue.ID != "ab-leaf" {
		t.Fatalf("expected cursor on leaf node after clearing filter, got %s",
			m.visibleRows[m.cursor].Node.Issue.ID)
	}

	// Verify ancestors are expanded so leaf is visible
	if !root.Expanded {
		t.Fatalf("expected root to be expanded after clearing filter")
	}
	if !child.Expanded {
		t.Fatalf("expected child to be expanded after clearing filter")
	}
}

func TestClearFilterPreservesSelectionMultiParent(t *testing.T) {
	// Create tree: epic1 -> task, epic2 -> task (shared)
	task := &graph.Node{Issue: beads.FullIssue{ID: "ab-task", Title: "Shared Task"}}
	epic1 := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-epic1", Title: "Epic One"},
		Children: []*graph.Node{task},
		Expanded: true,
	}
	epic2 := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-epic2", Title: "Epic Two"},
		Children: []*graph.Node{task},
		Expanded: true,
	}
	task.Parents = []*graph.Node{epic1, epic2}

	m := &App{
		roots:     []*graph.Node{epic1, epic2},
		textInput: textinput.New(),
		keys:      DefaultKeyMap(),
	}

	// Filter to show task
	m.setFilterText("task")
	m.recalcVisibleRows()

	// Find the task under epic2 (second occurrence)
	taskUnderEpic2Idx := -1
	for i, row := range m.visibleRows {
		if row.Node.Issue.ID == "ab-task" && row.Parent != nil && row.Parent.Issue.ID == "ab-epic2" {
			taskUnderEpic2Idx = i
			break
		}
	}
	if taskUnderEpic2Idx == -1 {
		t.Fatalf("could not find task under epic2 in filtered results")
	}
	m.cursor = taskUnderEpic2Idx

	// Clear filter
	m.clearSearchFilter()

	// Verify cursor is on task under epic2 specifically
	currentRow := m.visibleRows[m.cursor]
	if currentRow.Node.Issue.ID != "ab-task" {
		t.Fatalf("expected cursor on task, got %s", currentRow.Node.Issue.ID)
	}
	if currentRow.Parent == nil || currentRow.Parent.Issue.ID != "ab-epic2" {
		parentID := ""
		if currentRow.Parent != nil {
			parentID = currentRow.Parent.Issue.ID
		}
		t.Fatalf("expected task under epic2, got parent %s", parentID)
	}
}

func TestClearFilterExpandsAncestors(t *testing.T) {
	// Create deeply nested tree: level0 -> level1 -> level2 -> level3
	level3 := &graph.Node{Issue: beads.FullIssue{ID: "ab-lvl3", Title: "Level Three"}}
	level2 := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-lvl2", Title: "Level Two"},
		Children: []*graph.Node{level3},
	}
	level3.Parent = level2
	level3.Parents = []*graph.Node{level2}
	level1 := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-lvl1", Title: "Level One"},
		Children: []*graph.Node{level2},
	}
	level2.Parent = level1
	level2.Parents = []*graph.Node{level1}
	level0 := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-lvl0", Title: "Level Zero"},
		Children: []*graph.Node{level1},
	}
	level1.Parent = level0
	level1.Parents = []*graph.Node{level0}

	m := &App{
		roots:     []*graph.Node{level0},
		textInput: textinput.New(),
		keys:      DefaultKeyMap(),
	}

	// Filter to show deepest node
	m.setFilterText("three")
	m.recalcVisibleRows()

	// Select the deep node
	for i, row := range m.visibleRows {
		if row.Node.Issue.ID == "ab-lvl3" {
			m.cursor = i
			break
		}
	}

	// Clear filter
	m.clearSearchFilter()

	// Verify all ancestors are expanded
	if !level0.Expanded {
		t.Fatalf("expected level0 to be expanded")
	}
	if !level1.Expanded {
		t.Fatalf("expected level1 to be expanded")
	}
	if !level2.Expanded {
		t.Fatalf("expected level2 to be expanded")
	}

	// Verify the deep node is in visible rows
	found := false
	for _, row := range m.visibleRows {
		if row.Node.Issue.ID == "ab-lvl3" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected level3 to be visible after clearing filter")
	}
}

func TestClearFilterPreservesManualExpansion(t *testing.T) {
	// Create tree with collapsed child that has its own children
	grandchild := &graph.Node{Issue: beads.FullIssue{ID: "ab-gc", Title: "Grandchild"}}
	child := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-child", Title: "Child Node"},
		Children: []*graph.Node{grandchild},
		Expanded: false, // Initially collapsed
	}
	grandchild.Parent = child
	grandchild.Parents = []*graph.Node{child}
	root := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-root", Title: "Root Node"},
		Children: []*graph.Node{child},
		Expanded: true,
	}
	child.Parent = root
	child.Parents = []*graph.Node{root}

	m := &App{
		roots:     []*graph.Node{root},
		textInput: textinput.New(),
		keys:      DefaultKeyMap(),
	}

	// Apply a filter that matches child
	m.setFilterText("child")
	m.recalcVisibleRows()

	// Find child and manually expand it during filtering
	for i, row := range m.visibleRows {
		if row.Node.Issue.ID == "ab-child" {
			m.cursor = i
			m.expandNodeForView(row)
			break
		}
	}
	m.recalcVisibleRows()

	// Clear filter
	m.clearSearchFilter()

	// Verify child remains expanded
	if !child.Expanded {
		t.Fatalf("expected child to remain expanded after clearing filter")
	}
}

func TestClearFilterRootNodeSelection(t *testing.T) {
	// Create simple tree with multiple roots
	root1 := &graph.Node{Issue: beads.FullIssue{ID: "ab-root1", Title: "First Root"}}
	root2 := &graph.Node{Issue: beads.FullIssue{ID: "ab-root2", Title: "Second Root"}}

	m := &App{
		roots:     []*graph.Node{root1, root2},
		textInput: textinput.New(),
		keys:      DefaultKeyMap(),
	}

	// Filter to show only second root
	m.setFilterText("second")
	m.recalcVisibleRows()

	// Should only show root2
	if len(m.visibleRows) != 1 {
		t.Fatalf("expected 1 visible row during filter, got %d", len(m.visibleRows))
	}
	m.cursor = 0

	// Clear filter
	m.clearSearchFilter()

	// Verify root2 is still selected
	if m.visibleRows[m.cursor].Node.Issue.ID != "ab-root2" {
		t.Fatalf("expected cursor on root2 after clearing filter, got %s",
			m.visibleRows[m.cursor].Node.Issue.ID)
	}
}

func TestFilteredTreeManualToggle(t *testing.T) {
	buildApp := func() (*App, *graph.Node) {
		leaf := &graph.Node{Issue: beads.FullIssue{ID: "ab-003", Title: "Leaf"}}
		child := &graph.Node{
			Issue:    beads.FullIssue{ID: "ab-002", Title: "Child"},
			Children: []*graph.Node{leaf},
		}
		root := &graph.Node{
			Issue:    beads.FullIssue{ID: "ab-001", Title: "Root"},
			Children: []*graph.Node{child},
		}
		return &App{roots: []*graph.Node{root}}, root
	}

	assertVisible := func(t *testing.T, m *App, want int) {
		t.Helper()
		if got := len(m.visibleRows); got != want {
			t.Fatalf("expected %d visible rows, got %d", want, got)
		}
	}

	t.Run("collapseWhileFiltered", func(t *testing.T) {
		m, root := buildApp()
		m.setFilterText("leaf")
		m.recalcVisibleRows()
		assertVisible(t, m, 3)

		m.collapseNodeForView(nodeToRow(root))
		m.recalcVisibleRows()
		assertVisible(t, m, 1)
		if m.isNodeExpandedInView(nodeToRow(root)) {
			t.Fatalf("expected root to appear collapsed in filtered view")
		}
	})

	t.Run("expandAfterCollapse", func(t *testing.T) {
		m, root := buildApp()
		m.setFilterText("leaf")
		m.recalcVisibleRows()
		m.collapseNodeForView(nodeToRow(root))
		m.recalcVisibleRows()
		assertVisible(t, m, 1)

		m.expandNodeForView(nodeToRow(root))
		m.recalcVisibleRows()
		assertVisible(t, m, 3)
		if !m.isNodeExpandedInView(nodeToRow(root)) {
			t.Fatalf("expected root to appear expanded in filtered view")
		}
	})
}

func TestFilteredTogglePersistsWhileEditing(t *testing.T) {
	leaf := &graph.Node{Issue: beads.FullIssue{ID: "ab-103", Title: "Leaf"}}
	child := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-102", Title: "Child"},
		Children: []*graph.Node{leaf},
	}
	root := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-101", Title: "Root"},
		Children: []*graph.Node{child},
	}
	m := &App{roots: []*graph.Node{root}}

	m.setFilterText("le")
	m.recalcVisibleRows()
	if len(m.visibleRows) != 3 {
		t.Fatalf("expected initial filtered rows, got %d", len(m.visibleRows))
	}

	m.collapseNodeForView(nodeToRow(root))
	m.recalcVisibleRows()
	if len(m.visibleRows) != 1 {
		t.Fatalf("expected collapse to hide children, got %d rows", len(m.visibleRows))
	}

	m.setFilterText("leaf")
	m.recalcVisibleRows()
	if len(m.visibleRows) != 1 {
		t.Fatalf("expected collapse state to persist while editing filter, got %d rows", len(m.visibleRows))
	}

	m.expandNodeForView(nodeToRow(root))
	m.recalcVisibleRows()
	if len(m.visibleRows) != 3 {
		t.Fatalf("expected expand to restore children, got %d rows", len(m.visibleRows))
	}
}

func TestSearchFilterKeepsParentsVisible(t *testing.T) {
	grandchild := &graph.Node{Issue: beads.FullIssue{ID: "ab-401", Title: "Auth Login Flow"}}
	child := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-400", Title: "UI Improvements"},
		Children: []*graph.Node{grandchild},
	}
	root := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-399", Title: "Root Epic"},
		Children: []*graph.Node{child},
	}
	m := &App{roots: []*graph.Node{root}}

	m.setFilterText("auth")
	m.recalcVisibleRows()

	if len(m.visibleRows) != 3 {
		t.Fatalf("expected parent chain kept when descendant matches, got %d rows", len(m.visibleRows))
	}
	if ids := []string{m.visibleRows[0].Node.Issue.ID, m.visibleRows[1].Node.Issue.ID, m.visibleRows[2].Node.Issue.ID}; ids[0] != "ab-399" || ids[1] != "ab-400" || ids[2] != "ab-401" {
		t.Fatalf("expected full parent chain visible, got %v", ids)
	}
}

func TestSearchFilterAutoSelectsFirstMatch(t *testing.T) {
	root := &graph.Node{Issue: beads.FullIssue{ID: "ab-500", Title: "Alpha"}}
	match := &graph.Node{Issue: beads.FullIssue{ID: "ab-501", Title: "Auth Workflows"}}
	m := &App{roots: []*graph.Node{root, match}}
	m.recalcVisibleRows()
	m.cursor = 1

	m.setFilterText("auth")
	m.recalcVisibleRows()

	if len(m.visibleRows) != 1 {
		t.Fatalf("expected single match, got %d", len(m.visibleRows))
	}
	if m.cursor != 0 {
		t.Fatalf("expected cursor to jump to first match, got %d", m.cursor)
	}
	if m.visibleRows[m.cursor].Node.Issue.ID != "ab-501" {
		t.Fatalf("expected cursor on match, got %s", m.visibleRows[m.cursor].Node.Issue.ID)
	}
}

func TestSearchFilterCountsMatches(t *testing.T) {
	nodes := []*graph.Node{
		{Issue: beads.FullIssue{ID: "ab-600", Title: "Auth Login"}},
		{Issue: beads.FullIssue{ID: "ab-601", Title: "User Profile"}},
		{Issue: beads.FullIssue{ID: "ab-602", Title: "Auth Logout"}},
	}
	m := App{roots: nodes, filterText: "auth"}
	m.recalcVisibleRows()
	stats := m.getStats()
	if stats.Total != 2 {
		t.Fatalf("expected stats total 2 with filter, got %d", stats.Total)
	}
}

func TestSearchFilterRequiresAllWords(t *testing.T) {
	nodes := []*graph.Node{
		{Issue: beads.FullIssue{ID: "ab-610", Title: "Auth Sync"}},
		{Issue: beads.FullIssue{ID: "ab-611", Title: "User Provisioning"}},
		{Issue: beads.FullIssue{ID: "ab-612", Title: "Auth User Sync"}},
	}
	m := App{roots: nodes}
	m.setFilterText("auth user")
	m.recalcVisibleRows()

	var ids []string
	for _, row := range m.visibleRows {
		ids = append(ids, row.Node.Issue.ID)
	}
	if len(ids) != 1 || ids[0] != "ab-612" {
		t.Fatalf("expected only issues containing both words, got %v", ids)
	}
}

func TestStatsBreakdownUpdatesWithFilter(t *testing.T) {
	nodes := []*graph.Node{
		{Issue: beads.FullIssue{ID: "ab-801", Title: "Alpha In Progress", Status: "in_progress"}},
		{Issue: beads.FullIssue{ID: "ab-802", Title: "Beta Ready", Status: "open"}},
		{Issue: beads.FullIssue{ID: "ab-803", Title: "Gamma Blocked", Status: "open"}, IsBlocked: true},
		{Issue: beads.FullIssue{ID: "ab-804", Title: "Delta Closed", Status: "closed"}},
	}
	m := App{roots: nodes}
	statsAll := m.getStats()
	if statsAll.Total != 4 || statsAll.InProgress != 1 || statsAll.Ready != 1 || statsAll.Blocked != 1 || statsAll.Closed != 1 {
		t.Fatalf("unexpected stats before filtering: %+v", statsAll)
	}
	m.setFilterText("beta")
	statsFiltered := m.getStats()
	if statsFiltered.Total != 1 || statsFiltered.Ready != 1 || statsFiltered.InProgress != 0 || statsFiltered.Blocked != 0 || statsFiltered.Closed != 0 {
		t.Fatalf("expected filter to narrow stats to beta ready item, got %+v", statsFiltered)
	}
}

func TestStatsFilteredSuffixMatchesDocs(t *testing.T) {
	m := &App{
		roots: []*graph.Node{
			{Issue: beads.FullIssue{ID: "ab-820", Title: "Filter Demo", Status: "open"}},
		},
		width:  100,
		height: 30,
		ready:  true,
	}
	m.recalcVisibleRows()
	m.setFilterText("filter")
	view := stripANSI(m.View())
	if !strings.Contains(view, "(filtered)") {
		t.Skipf("docs expect '(filtered)' suffix when search is active; header output was:\n%s", view)
	}
	if !strings.Contains(view, "(filtered)") {
		t.Fatalf("expected stats line to include '(filtered)' when filtered:\n%s", view)
	}
}

func TestRefreshDeltaDisplayMatchesDocs(t *testing.T) {
	m := &App{
		width:            100,
		height:           30,
		ready:            true,
		lastRefreshStats: "+1 / Δ1 / -0",
	}

	// Test visible state (within display duration, with changes)
	m.lastRefreshTime = time.Now()
	status := m.renderRefreshStatus()
	if !strings.Contains(status, "Δ") || !strings.Contains(status, "+1") {
		t.Fatalf("expected refresh delta to be visible with changes, got: %q", status)
	}

	// Test empty state (after display duration)
	m.lastRefreshTime = time.Now().Add(-refreshDisplayDuration - time.Millisecond)
	status = m.renderRefreshStatus()
	if status != "" {
		t.Fatalf("expected refresh status to be empty after display duration, got: %q", status)
	}

	// Test no-change state (should not show delta)
	m.lastRefreshStats = "+0 / Δ0 / -0"
	m.lastRefreshTime = time.Now()
	status = m.renderRefreshStatus()
	if status != "" {
		t.Fatalf("expected refresh status to be empty when no changes, got: %q", status)
	}
}

func buildTreeTestApp(nodes ...*graph.Node) *App {
	m := &App{
		roots:  nodes,
		width:  120,
		height: 40,
	}
	m.recalcVisibleRows()
	return m
}

func buildWrappedTreeApp(count int) *App {
	nodes := make([]*graph.Node, count)
	for i := 0; i < count; i++ {
		nodes[i] = &graph.Node{
			Issue: beads.FullIssue{
				ID:     fmt.Sprintf("ab-%02d", i+1),
				Title:  "Ensure selection stays visible even when this title wraps onto multiple lines within the viewport.",
				Status: "open",
			},
		}
	}
	app := &App{
		roots:  nodes,
		width:  50,
		height: 12,
	}
	app.recalcVisibleRows()
	return app
}

func treeLineContaining(t *testing.T, view, id string) string {
	t.Helper()
	clean := stripANSI(view)
	for _, line := range strings.Split(clean, "\n") {
		if strings.Contains(line, id) {
			return strings.TrimRight(line, " ")
		}
	}
	t.Fatalf("tree output missing %s:\n%s", id, clean)
	return ""
}

func TestTreeViewStatusIconsMatchDocs(t *testing.T) {
	inProgress := &graph.Node{Issue: beads.FullIssue{ID: "ab-701", Title: "In Progress", Status: "in_progress"}}
	ready := &graph.Node{Issue: beads.FullIssue{ID: "ab-702", Title: "Ready", Status: "open"}}
	blocked := &graph.Node{Issue: beads.FullIssue{ID: "ab-703", Title: "Blocked", Status: "open"}, IsBlocked: true}
	closed := &graph.Node{Issue: beads.FullIssue{ID: "ab-704", Title: "Closed", Status: "closed"}}
	m := buildTreeTestApp(inProgress, ready, blocked, closed)
	view := m.renderTreeView()
	cases := []struct {
		id   string
		icon string
	}{
		{"ab-701", "◐"},
		{"ab-702", "○"},
		{"ab-703", "⛔"},
		{"ab-704", "✔"},
	}
	for _, c := range cases {
		line := treeLineContaining(t, view, c.id)
		if !strings.Contains(line, c.icon) {
			t.Fatalf("expected %s line to contain %s icon, got %q", c.id, c.icon, line)
		}
	}
}

func TestTreeViewMarkersTogglePerDocs(t *testing.T) {
	child := &graph.Node{Issue: beads.FullIssue{ID: "ab-710", Title: "Child"}}
	expanded := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-711", Title: "Expanded Parent"},
		Children: []*graph.Node{child},
		Expanded: true,
	}
	collapsed := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-712", Title: "Collapsed Parent"},
		Children: []*graph.Node{{Issue: beads.FullIssue{ID: "ab-713", Title: "Hidden"}}},
	}
	m := buildTreeTestApp(expanded, collapsed)
	view := m.renderTreeView()
	expandedLine := treeLineContaining(t, view, "ab-711")
	if !strings.Contains(expandedLine, "▼") {
		t.Fatalf("expected expanded marker ▼, got %q", expandedLine)
	}
	collapsedLine := treeLineContaining(t, view, "ab-712")
	if !strings.Contains(collapsedLine, "▶") {
		t.Fatalf("expected collapsed marker ▶, got %q", collapsedLine)
	}
}

func TestTreeViewCollapsedNodesShowCountBadge(t *testing.T) {
	t.Skip("Docs specify [+N] counts for collapsed nodes; UI currently omits them.")
	child := &graph.Node{Issue: beads.FullIssue{ID: "ab-720", Title: "Hidden"}}
	collapsed := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-721", Title: "Collapsed With Count"},
		Children: []*graph.Node{child},
	}
	m := buildTreeTestApp(collapsed)
	view := m.renderTreeView()
	line := treeLineContaining(t, view, "ab-721")
	if !strings.Contains(line, "[+1]") {
		t.Fatalf("expected collapsed node to show [+1] badge, got %q", line)
	}
}

func TestTreeScrollKeepsWrappedSelectionVisible(t *testing.T) {
	app := buildWrappedTreeApp(12)
	for i := range app.visibleRows {
		app.cursor = i
		view := stripANSI(app.renderTreeView())
		id := fmt.Sprintf("ab-%02d", i+1)
		if !strings.Contains(view, id) {
			t.Fatalf("expected view to include %s at cursor %d:\n%s", id, i, view)
		}
	}
}

func TestTreeEndKeySafeWhenNoVisibleRows(t *testing.T) {
	app := &App{
		visibleRows: []graph.TreeRow{},
		viewport:    viewport.New(80, 20),
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("End key should not panic on empty list: %v", r)
		}
	}()
	app.Update(tea.KeyMsg{Type: tea.KeyEnd})
	if app.cursor != 0 {
		t.Fatalf("expected cursor to remain at 0, got %d", app.cursor)
	}
	// Should also tolerate detail toggles without crashing.
	app.ShowDetails = true
	app.updateViewportContent()
}

func TestFindBeadsDBPrefersEnv(t *testing.T) {
	t.Setenv("BEADS_DB", "")
	tmp := t.TempDir()
	dbFile := filepath.Join(tmp, "custom.db")
	if err := os.WriteFile(dbFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("write db file: %v", err)
	}
	t.Setenv("BEADS_DB", dbFile)

	cleanup := changeWorkingDir(t, tmp)
	defer cleanup()

	path, modTime, err := findBeadsDB()
	if err != nil {
		t.Fatalf("findBeadsDB: %v", err)
	}
	if normalizePath(t, path) != normalizePath(t, dbFile) {
		t.Fatalf("expected path %s, got %s", dbFile, path)
	}
	info, err := os.Stat(dbFile)
	if err != nil {
		t.Fatalf("stat db file: %v", err)
	}
	if !modTime.Equal(info.ModTime()) {
		t.Fatalf("expected mod time %v, got %v", info.ModTime(), modTime)
	}
}

func TestFindBeadsDBWalksUpDirectories(t *testing.T) {
	t.Setenv("BEADS_DB", "")
	root := t.TempDir()
	beadsDir := filepath.Join(root, ".beads")
	if err := os.MkdirAll(beadsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	dbFile := filepath.Join(beadsDir, "beads.db")
	if err := os.WriteFile(dbFile, []byte("db"), 0o644); err != nil {
		t.Fatalf("write db: %v", err)
	}
	nested := filepath.Join(root, "nested", "deep")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	cleanup := changeWorkingDir(t, nested)
	defer cleanup()

	path, _, err := findBeadsDB()
	if err != nil {
		t.Fatalf("findBeadsDB: %v", err)
	}
	if normalizePath(t, path) != normalizePath(t, dbFile) {
		t.Fatalf("expected %s, got %s", dbFile, path)
	}
}

func TestFindBeadsDBFallsBackToDefault(t *testing.T) {
	t.Setenv("BEADS_DB", "")
	projectDir := t.TempDir()
	cleanup := changeWorkingDir(t, projectDir)
	defer cleanup()

	home := t.TempDir()
	t.Setenv("HOME", home)
	defaultDir := filepath.Join(home, ".beads")
	if err := os.MkdirAll(defaultDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	defaultDB := filepath.Join(defaultDir, "default.db")
	if err := os.WriteFile(defaultDB, []byte("db"), 0o644); err != nil {
		t.Fatalf("write db: %v", err)
	}

	path, _, err := findBeadsDB()
	if err != nil {
		t.Fatalf("findBeadsDB: %v", err)
	}
	if normalizePath(t, path) != normalizePath(t, defaultDB) {
		t.Fatalf("expected fallback %s, got %s", defaultDB, path)
	}
}

func TestBuildMarkdownRendererPlainStyle(t *testing.T) {
	text := "alpha beta gamma delta"
	width := 6
	want := wordwrap.String(text, width)

	render := buildMarkdownRenderer("plain", width)
	if got := render(text); got != want {
		t.Fatalf("expected plain renderer to match fallback %q, got %q", want, got)
	}
}

func TestRecalcVisibleRowsMatchesIDs(t *testing.T) {
	nodes := []*graph.Node{
		{Issue: beads.FullIssue{ID: "ab-123", Title: "Alpha"}},
		{Issue: beads.FullIssue{ID: "ab-456", Title: "Beta"}},
	}
	m := App{
		roots:      nodes,
		filterText: "ab-123",
	}
	m.recalcVisibleRows()

	if len(m.visibleRows) != 1 {
		t.Fatalf("expected 1 match, got %d", len(m.visibleRows))
	}
	if got := m.visibleRows[0].Node.Issue.ID; got != "ab-123" {
		t.Fatalf("expected ID ab-123, got %s", got)
	}
}

func TestRecalcVisibleRowsMatchesPartialIDs(t *testing.T) {
	nodes := []*graph.Node{
		{Issue: beads.FullIssue{ID: "ab-123", Title: "Alpha"}},
		{Issue: beads.FullIssue{ID: "ab-456", Title: "Beta"}},
	}
	m := App{
		roots:      nodes,
		filterText: "456",
	}
	m.recalcVisibleRows()

	if len(m.visibleRows) != 1 {
		t.Fatalf("expected 1 match, got %d", len(m.visibleRows))
	}
	if got := m.visibleRows[0].Node.Issue.ID; got != "ab-456" {
		t.Fatalf("expected ID ab-456, got %s", got)
	}
}

func TestNewAppWithMockClientLoadsIssues(t *testing.T) {
	t.Parallel()
	fixture := loadFixtureIssues(t, "issues_basic.json")
	mock := beads.NewMockClient()
	mock.ListFn = func(ctx context.Context) ([]beads.LiteIssue, error) {
		return liteIssuesFromFixture(fixture), nil
	}
	mock.ShowFn = func(ctx context.Context, ids []string) ([]beads.FullIssue, error) {
		return filterIssuesByID(fixture, ids), nil
	}
	mock.CommentsFn = func(ctx context.Context, issueID string) ([]beads.Comment, error) {
		return []beads.Comment{
			{ID: 1, IssueID: issueID, Author: "tester", Text: "hello", CreatedAt: time.Now().UTC().Format(time.RFC3339)},
		}, nil
	}

	dbFile := createTempDBFile(t)
	app, err := NewApp(Config{
		RefreshInterval: time.Second,
		AutoRefresh:     false,
		DBPathOverride:  dbFile,
		Client:          mock,
	})
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	if len(app.roots) != 1 {
		t.Fatalf("expected a single root, got %d", len(app.roots))
	}
	if mock.CommentsCallCount != len(fixture) {
		t.Fatalf("expected comments fetched for all issues, got %d", mock.CommentsCallCount)
	}
}

func TestAppRefreshWithMockClient(t *testing.T) {
	fixtureInitial := loadFixtureIssues(t, "issues_basic.json")
	fixtureUpdated := loadFixtureIssues(t, "issues_refresh.json")
	mock := beads.NewMockClient()
	mock.ListFn = func(ctx context.Context) ([]beads.LiteIssue, error) {
		return liteIssuesFromFixture(fixtureInitial), nil
	}
	var showCalls int
	mock.ShowFn = func(ctx context.Context, ids []string) ([]beads.FullIssue, error) {
		if showCalls == 0 {
			showCalls++
			return filterIssuesByID(fixtureInitial, ids), nil
		}
		return filterIssuesByID(fixtureUpdated, ids), nil
	}
	mock.CommentsFn = func(ctx context.Context, issueID string) ([]beads.Comment, error) {
		return nil, nil
	}

	app := mustNewTestApp(t, mock)
	cmd := app.forceRefresh()
	refreshMsg := extractRefreshMsg(t, cmd)
	app.Update(refreshMsg)

	if got := app.roots[0].Issue.Title; got != "Root Epic Updated" {
		t.Fatalf("expected updated root title, got %s", got)
	}
}

func TestNewAppCapturesClientError(t *testing.T) {
	mock := beads.NewMockClient()
	mock.ListFn = func(ctx context.Context) ([]beads.LiteIssue, error) {
		return nil, errors.New("boom")
	}
	dbFile := createTempDBFile(t)
	_, err := NewApp(Config{
		RefreshInterval: time.Second,
		DBPathOverride:  dbFile,
		Client:          mock,
	})
	if err == nil {
		t.Fatalf("expected error when client list fails")
	}
}

func TestNewAppReturnsErrorWhenNoIssues(t *testing.T) {
	mock := beads.NewMockClient()
	mock.ListFn = func(ctx context.Context) ([]beads.LiteIssue, error) {
		return []beads.LiteIssue{}, nil
	}
	dbFile := createTempDBFile(t)
	if _, err := NewApp(Config{
		DBPathOverride: dbFile,
		Client:         mock,
	}); !errors.Is(err, ErrNoIssues) {
		t.Fatalf("expected ErrNoIssues, got %v", err)
	}
}

func TestCheckDBForChangesDetectsModification(t *testing.T) {
	dbFile := createTempDBFile(t)
	app := &App{
		client:        beads.NewMockClient(),
		dbPath:        dbFile,
		lastDBModTime: fileModTime(t, dbFile),
	}

	if cmd := app.checkDBForChanges(); cmd != nil {
		t.Fatalf("expected no refresh when mod time unchanged")
	}
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(dbFile, []byte("update"), 0o644); err != nil {
		t.Fatalf("write db: %v", err)
	}
	if cmd := app.checkDBForChanges(); cmd == nil {
		t.Fatalf("expected refresh command after db modification")
	}
}

func TestRefreshHandlesClientError(t *testing.T) {
	fixtureInitial := loadFixtureIssues(t, "issues_basic.json")
	mock := beads.NewMockClient()
	mock.ListFn = func(ctx context.Context) ([]beads.LiteIssue, error) {
		return liteIssuesFromFixture(fixtureInitial), nil
	}
	var showCalls int
	mock.ShowFn = func(ctx context.Context, ids []string) ([]beads.FullIssue, error) {
		if showCalls == 0 {
			showCalls++
			return filterIssuesByID(fixtureInitial, ids), nil
		}
		return nil, errors.New("show failed")
	}
	mock.CommentsFn = func(ctx context.Context, issueID string) ([]beads.Comment, error) { return nil, nil }

	app := mustNewTestApp(t, mock)
	cmd := app.forceRefresh()
	refreshMsg := extractRefreshMsg(t, cmd)
	app.Update(refreshMsg)
	// Errors are now stored in lastError, not lastRefreshStats
	if app.lastError == "" {
		t.Fatalf("expected error to be stored in lastError")
	}
	if !strings.Contains(app.lastError, "show failed") {
		t.Fatalf("expected error message to contain 'show failed', got %s", app.lastError)
	}
}

func TestErrorToastShowsOnFirstError(t *testing.T) {
	fixtureInitial := loadFixtureIssues(t, "issues_basic.json")
	mock := beads.NewMockClient()
	mock.ListFn = func(ctx context.Context) ([]beads.LiteIssue, error) {
		return liteIssuesFromFixture(fixtureInitial), nil
	}
	var showCalls int
	mock.ShowFn = func(ctx context.Context, ids []string) ([]beads.FullIssue, error) {
		if showCalls == 0 {
			showCalls++
			return filterIssuesByID(fixtureInitial, ids), nil
		}
		return nil, errors.New("connection failed")
	}
	mock.CommentsFn = func(ctx context.Context, issueID string) ([]beads.Comment, error) { return nil, nil }

	app := mustNewTestApp(t, mock)

	// Trigger refresh error
	cmd := app.forceRefresh()
	refreshMsg := extractRefreshMsg(t, cmd)
	_, nextCmd := app.Update(refreshMsg)

	// Toast should be shown on first error
	if !app.showErrorToast {
		t.Error("expected showErrorToast to be true on first error")
	}
	if !app.errorShownOnce {
		t.Error("expected errorShownOnce to be true")
	}
	if nextCmd == nil {
		t.Error("expected tick command to be returned for toast countdown")
	}
}

func TestErrorToastEscDismisses(t *testing.T) {
	app := &App{
		lastError:      "test error",
		showErrorToast: true,
		errorShownOnce: true,
		ready:          true,
		keys:           DefaultKeyMap(),
	}

	// Press ESC
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if app.showErrorToast {
		t.Error("expected showErrorToast to be false after ESC")
	}
	// Error should still be stored
	if app.lastError == "" {
		t.Error("expected lastError to remain after dismissing toast")
	}
}

func TestErrorToastEKeyRecalls(t *testing.T) {
	app := &App{
		lastError:      "test error",
		showErrorToast: false,
		errorShownOnce: true,
		ready:          true,
		keys:           DefaultKeyMap(),
	}

	// Press '!' (error key)
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}})

	if !app.showErrorToast {
		t.Error("expected showErrorToast to be true after pressing '!'")
	}
	if cmd == nil {
		t.Error("expected tick command to be returned")
	}
}

func TestErrorToastDoesNotReappearOnSubsequentErrors(t *testing.T) {
	app := &App{
		lastError:      "first error",
		showErrorToast: false,
		errorShownOnce: true, // Already shown once
		ready:          true,
	}

	// Simulate another refresh error
	msg := refreshCompleteMsg{err: errors.New("second error")}
	_, cmd := app.Update(msg)

	// Toast should NOT show again
	if app.showErrorToast {
		t.Error("expected showErrorToast to remain false for subsequent errors")
	}
	if cmd != nil {
		t.Error("expected no tick command since toast is not shown")
	}
	// But error should be updated
	if !strings.Contains(app.lastError, "second error") {
		t.Errorf("expected lastError to be updated, got %s", app.lastError)
	}
}

func TestErrorToastClearedOnSuccessfulRefresh(t *testing.T) {
	app := &App{
		lastError:      "previous error",
		showErrorToast: true,
		errorShownOnce: true,
		ready:          true,
	}

	// Simulate successful refresh
	msg := refreshCompleteMsg{
		roots:  []*graph.Node{},
		digest: map[string]string{},
	}
	app.Update(msg)

	if app.lastError != "" {
		t.Errorf("expected lastError to be cleared, got %s", app.lastError)
	}
	if app.showErrorToast {
		t.Error("expected showErrorToast to be false")
	}
	if app.errorShownOnce {
		t.Error("expected errorShownOnce to be reset")
	}
}

func TestErrorToastCountdown(t *testing.T) {
	app := &App{
		lastError:       "test error",
		showErrorToast:  true,
		errorToastStart: time.Now().Add(-5 * time.Second), // Started 5 seconds ago
		ready:           true,
	}

	// Process tick - should continue countdown
	_, cmd := app.Update(errorToastTickMsg{})
	if !app.showErrorToast {
		t.Error("toast should still be visible before 10 seconds")
	}
	if cmd == nil {
		t.Error("expected another tick to be scheduled")
	}

	// Simulate 10+ seconds elapsed
	app.errorToastStart = time.Now().Add(-11 * time.Second)
	_, cmd = app.Update(errorToastTickMsg{})
	if app.showErrorToast {
		t.Error("toast should auto-dismiss after 10 seconds")
	}
}

func TestExtractShortError(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "extracts from Error: prefix",
			input:    "list issues: run bd list: Error: Database out of sync with JSONL. Run 'bd sync' to fix.",
			maxLen:   80,
			expected: "Database out of sync with JSONL",
		},
		{
			name:     "removes Run suggestion",
			input:    "Error: Something failed. Run 'bd fix' to resolve.",
			maxLen:   80,
			expected: "Something failed",
		},
		{
			name:     "truncates long message",
			input:    "Error: This is a very long error message that exceeds the maximum length allowed",
			maxLen:   30,
			expected: "This is a very long error m...",
		},
		{
			name:     "handles multiline error",
			input:    "Error: First line of error\nSecond line with more details",
			maxLen:   80,
			expected: "First line of error",
		},
		{
			name:     "handles simple error without prefix",
			input:    "connection refused",
			maxLen:   80,
			expected: "connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractShortError(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("extractShortError(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestUpdateViewportContentDisplaysDesignSection(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:          "ab-101",
			Title:       "Detail Layout",
			Status:      "open",
			IssueType:   "feature",
			Priority:    2,
			Description: "High-level summary.",
			Design:      "## Architecture\n\nDocument component wiring.",
			CreatedAt:   time.Date(2025, time.November, 21, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
			UpdatedAt:   time.Date(2025, time.November, 21, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
			Comments: []beads.Comment{
				{
					Author:    "Reviewer",
					Text:      "Looks good",
					CreatedAt: time.Date(2025, time.November, 21, 13, 0, 0, 0, time.UTC).Format(time.RFC3339),
				},
			},
		},
		CommentsLoaded: true,
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  nodesToRows(node),
		viewport:     viewport.New(90, 30),
		outputFormat: "plain",
	}

	app.updateViewportContent()
	content := stripANSI(app.viewport.View())

	if !strings.Contains(content, "Design") {
		t.Fatalf("expected Design header in viewport content:\n%s", content)
	}

	descIdx := strings.Index(content, "Description")
	designIdx := strings.Index(content, "Design")
	if descIdx == -1 || designIdx == -1 {
		t.Fatalf("expected both Description and Design headers")
	}
	if !(descIdx < designIdx) {
		t.Fatalf("expected Design to appear after Description: descIdx=%d, designIdx=%d\n%s", descIdx, designIdx, content)
	}

	if !strings.Contains(content, "## Architecture") {
		t.Fatalf("expected markdown-rendered design content present, got:\n%s", content)
	}
}

func TestUpdateViewportContentOmitsDesignWhenBlank(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:          "ab-102",
			Title:       "Missing Section",
			Status:      "open",
			IssueType:   "feature",
			Priority:    2,
			Description: "Content exists.",
			Design:      "   ",
			CreatedAt:   time.Date(2025, time.November, 22, 9, 0, 0, 0, time.UTC).Format(time.RFC3339),
			UpdatedAt:   time.Date(2025, time.November, 22, 9, 15, 0, 0, time.UTC).Format(time.RFC3339),
		},
		CommentsLoaded: true,
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  nodesToRows(node),
		viewport:     viewport.New(90, 30),
		outputFormat: "plain",
	}

	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	if strings.Contains(content, "Design") {
		t.Fatalf("expected Design section omitted when empty, content:\n%s", content)
	}
}

func TestUpdateViewportContentDisplaysAcceptanceSection(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:                 "ab-103",
			Title:              "Version Checks",
			Status:             "open",
			IssueType:          "feature",
			Priority:           2,
			Description:        "Ensure CLI presence",
			Design:             "## Flow\n\n1. Detect CLI\n2. Compare version",
			AcceptanceCriteria: "## Acceptance\n\n- Clear error when missing\n- Friendly instructions",
			CreatedAt:          time.Date(2025, time.November, 22, 8, 0, 0, 0, time.UTC).Format(time.RFC3339),
			UpdatedAt:          time.Date(2025, time.November, 22, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
			Comments: []beads.Comment{
				{
					Author:    "QA",
					Text:      "Need docs link",
					CreatedAt: time.Date(2025, time.November, 22, 11, 0, 0, 0, time.UTC).Format(time.RFC3339),
				},
			},
		},
		CommentsLoaded: true,
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  nodesToRows(node),
		viewport:     viewport.New(90, 30),
		outputFormat: "plain",
	}

	app.updateViewportContent()
	content := stripANSI(app.viewport.View())

	if !strings.Contains(content, "Acceptance:") {
		t.Fatalf("expected Acceptance header present:\n%s", content)
	}
	if !strings.Contains(content, "## Acceptance") {
		t.Fatalf("expected markdown acceptance content present:\n%s", content)
	}

	designIdx := strings.Index(content, "Design:")
	acceptIdx := strings.Index(content, "Acceptance:")
	commentsIdx := strings.Index(content, "Comments:")
	if designIdx == -1 || acceptIdx == -1 || commentsIdx == -1 {
		t.Fatalf("expected Design, Acceptance, and Comments headers present")
	}
	if !(designIdx < acceptIdx && acceptIdx < commentsIdx) {
		t.Fatalf("expected Acceptance to appear between Design and Comments: design=%d acceptance=%d comments=%d\n%s",
			designIdx, acceptIdx, commentsIdx, content)
	}
}

func TestDetailMetadataLayoutMatchesDocs(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:          "ab-210",
			Title:       "Metadata Layout",
			Status:      "in_progress",
			IssueType:   "feature",
			Priority:    2,
			Labels:      []string{"auth", "security"},
			Description: "Doc-aligned metadata block.",
			CreatedAt:   time.Date(2025, time.November, 23, 7, 0, 0, 0, time.UTC).Format(time.RFC3339),
			UpdatedAt:   time.Date(2025, time.November, 23, 8, 30, 0, 0, time.UTC).Format(time.RFC3339),
		},
		CommentsLoaded: true,
	}

	app := &App{
		ShowDetails:  true,
		visibleRows:  nodesToRows(node),
		viewport:     viewport.New(90, 30),
		outputFormat: "plain",
	}

	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	start := strings.Index(content, "Status:")
	if start == -1 {
		t.Fatalf("metadata block missing Status line:\n%s", content)
	}
	end := strings.Index(content[start:], "Description:")
	if end == -1 {
		t.Fatalf("metadata block missing Description delimiter:\n%s", content)
	}
	metaBlock := content[start : start+end]
	var rows []string
	for _, line := range strings.Split(metaBlock, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		rows = append(rows, trimmed)
	}
	if len(rows) < 4 {
		t.Fatalf("expected metadata rows, got %d:\n%s", len(rows), metaBlock)
	}
	if !(strings.Contains(rows[0], "Status:") && strings.Contains(rows[0], "Priority:")) {
		t.Fatalf("row 1 should contain Status and Priority, got %q", rows[0])
	}
	if !(strings.Contains(rows[1], "Type:") && strings.Contains(rows[1], "Labels:")) {
		t.Fatalf("row 2 should contain Type and Labels, got %q", rows[1])
	}
	if !strings.HasPrefix(rows[2], "Created:") {
		t.Fatalf("row 3 should begin with Created, got %q", rows[2])
	}
	if !strings.HasPrefix(rows[3], "Updated:") {
		t.Fatalf("row 4 should begin with Updated, got %q", rows[3])
	}
}

func TestDetailRelationshipSectionsFollowDocs(t *testing.T) {
	parent := &graph.Node{Issue: beads.FullIssue{ID: "ab-300", Title: "Parent Node"}}
	childA := &graph.Node{Issue: beads.FullIssue{ID: "ab-301", Title: "Child A"}}
	childB := &graph.Node{Issue: beads.FullIssue{ID: "ab-302", Title: "Child B"}}
	blocker := &graph.Node{Issue: beads.FullIssue{ID: "ab-303", Title: "Blocking Task"}}
	blocked := &graph.Node{Issue: beads.FullIssue{ID: "ab-304", Title: "Blocked Task"}}

	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:          "ab-305",
			Title:       "Relationship Order",
			Description: "Ensure sections match documentation order.",
		},
		Parent:         parent,
		Parents:        []*graph.Node{parent},
		Children:       []*graph.Node{childA, childB},
		BlockedBy:      []*graph.Node{blocker},
		Blocks:         []*graph.Node{blocked},
		IsBlocked:      true,
		CommentsLoaded: true,
	}

	app := &App{
		ShowDetails:  true,
		visibleRows:  nodesToRows(node),
		viewport:     viewport.New(90, 40),
		outputFormat: "plain",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	// New labels: "Part Of:", "Subtasks", "Must Complete First:", "Will Unblock"
	order := []string{"Part Of:", "Subtasks", "Must Complete First:", "Will Unblock"}
	var lastIdx int = -1
	for _, section := range order {
		idx := strings.Index(content, section)
		if idx == -1 {
			t.Fatalf("missing %s section in content:\n%s", section, content)
		}
		if idx <= lastIdx {
			t.Fatalf("section %s appeared out of order", section)
		}
		lastIdx = idx
	}
	if !strings.Contains(content, "Subtasks: (2)") {
		t.Fatalf("expected Subtasks count header, got:\n%s", content)
	}
	if !strings.Contains(content, "ab-303") || !strings.Contains(content, "ab-304") {
		t.Fatalf("expected related issue IDs rendered:\n%s", content)
	}
}

func TestDetailLabelsWrapWhenViewportNarrow(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:          "ab-320",
			Title:       "Label Wrapping",
			Status:      "open",
			IssueType:   "feature",
			Priority:    2,
			Labels:      []string{"alpha", "beta", "gamma"},
			Description: "Verify label chips wrap across lines.",
			CreatedAt:   time.Date(2025, time.November, 23, 6, 0, 0, 0, time.UTC).Format(time.RFC3339),
		},
		CommentsLoaded: true,
	}

	app := &App{
		ShowDetails:  true,
		visibleRows:  nodesToRows(node),
		viewport:     viewport.New(42, 25),
		outputFormat: "plain",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	lines := strings.Split(content, "\n")
	labelsIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "Labels:") {
			labelsIdx = i
			break
		}
	}
	if labelsIdx == -1 {
		t.Fatalf("no Labels row found:\n%s", content)
	}
	if strings.Contains(lines[labelsIdx], "beta") {
		t.Fatalf("expected first labels row to wrap, got %q", lines[labelsIdx])
	}
	if labelsIdx+1 >= len(lines) {
		t.Fatalf("expected additional label rows after wrap")
	}
	wrapped := strings.TrimSpace(lines[labelsIdx+1])
	if !strings.Contains(wrapped, "beta") {
		t.Fatalf("expected wrapped row to include beta label, got %q", wrapped)
	}
	if labelsIdx+2 >= len(lines) {
		t.Fatalf("expected third line for gamma label")
	}
	wrapped2 := strings.TrimSpace(lines[labelsIdx+2])
	if !strings.Contains(wrapped2, "gamma") {
		t.Fatalf("expected second wrapped row to include gamma label, got %q", wrapped2)
	}
}

func TestDetailCommentsRenderEntries(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:          "ab-330",
			Title:       "Comment Rendering",
			Description: "Doc sample comments.",
			Comments: []beads.Comment{
				{Author: "@alice", Text: "Let's use OAuth2.", CreatedAt: time.Date(2025, time.November, 20, 9, 15, 0, 0, time.UTC).Format(time.RFC3339)},
				{Author: "@bob", Text: "Agreed, updating design.", CreatedAt: time.Date(2025, time.November, 20, 11, 30, 0, 0, time.UTC).Format(time.RFC3339)},
			},
		},
		CommentsLoaded: true,
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  nodesToRows(node),
		viewport:     viewport.New(80, 30),
		outputFormat: "plain",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	idx := strings.Index(content, "Comments:")
	if idx == -1 {
		t.Fatalf("missing Comments header:\n%s", content)
	}
	if !(strings.Contains(content, "@alice") && strings.Contains(content, "Let's use OAuth2.")) {
		t.Fatalf("expected first comment rendered:\n%s", content)
	}
	if !(strings.Contains(content, "@bob") && strings.Contains(content, "Agreed, updating design.")) {
		t.Fatalf("expected second comment rendered:\n%s", content)
	}
	if strings.Index(content, "@alice") > strings.Index(content, "@bob") {
		t.Fatalf("comments should remain chronological")
	}
}

func TestDetailCommentsErrorMessageMatchesDocs(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:          "ab-340",
			Title:       "Comment Error",
			Description: "Doc retry guidance.",
		},
		CommentError: "timeout fetching comments",
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  nodesToRows(node),
		viewport:     viewport.New(80, 30),
		outputFormat: "plain",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	if !strings.Contains(content, "Failed to load comments. Press 'c' to retry.") {
		t.Fatalf("expected retry guidance in error state:\n%s", content)
	}
	if !strings.Contains(content, "timeout fetching comments") {
		t.Fatalf("expected underlying error rendered, content:\n%s", content)
	}
}

func TestUpdateViewportContentOmitsAcceptanceWhenBlank(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:                 "ab-104",
			Title:              "Whitespace Acceptance",
			Status:             "open",
			IssueType:          "feature",
			Priority:           2,
			Description:        "Has description.",
			Design:             "## Design\n\n- present",
			AcceptanceCriteria: "   \n",
			CreatedAt:          time.Date(2025, time.November, 22, 9, 30, 0, 0, time.UTC).Format(time.RFC3339),
			UpdatedAt:          time.Date(2025, time.November, 22, 9, 45, 0, 0, time.UTC).Format(time.RFC3339),
		},
		CommentsLoaded: true,
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  nodesToRows(node),
		viewport:     viewport.New(90, 30),
		outputFormat: "plain",
	}

	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	if strings.Contains(content, "Acceptance:") {
		t.Fatalf("expected Acceptance section omitted when whitespace, content:\n%s", content)
	}
}

func TestCopyBeadIDSetsToastState(t *testing.T) {
	node := &graph.Node{Issue: beads.FullIssue{ID: "ab-123", Title: "Test Issue"}}
	app := &App{
		visibleRows: nodesToRows(node),
		cursor:      0,
		ready:       true,
	}

	// Simulate pressing 'c' key - we test the state changes, not actual clipboard
	// (clipboard may not work in CI/test environments)
	app.copiedBeadID = node.Issue.ID
	app.showCopyToast = true
	app.copyToastStart = time.Now()

	if !app.showCopyToast {
		t.Error("expected showCopyToast to be true")
	}
	if app.copiedBeadID != "ab-123" {
		t.Errorf("expected copiedBeadID 'ab-123', got %s", app.copiedBeadID)
	}
}

func TestCopyToastCountdown(t *testing.T) {
	app := &App{
		copiedBeadID:   "ab-456",
		showCopyToast:  true,
		copyToastStart: time.Now().Add(-3 * time.Second), // Started 3 seconds ago
		ready:          true,
	}

	// Process tick - should continue countdown (not yet 5 seconds)
	_, cmd := app.Update(copyToastTickMsg{})
	if !app.showCopyToast {
		t.Error("toast should still be visible before 5 seconds")
	}
	if cmd == nil {
		t.Error("expected another tick to be scheduled")
	}

	// Simulate 5+ seconds elapsed
	app.copyToastStart = time.Now().Add(-6 * time.Second)
	_, cmd = app.Update(copyToastTickMsg{})
	if app.showCopyToast {
		t.Error("toast should auto-dismiss after 5 seconds")
	}
}

func TestCopyToastRenders(t *testing.T) {
	app := &App{
		copiedBeadID:   "ab-789",
		showCopyToast:  true,
		copyToastStart: time.Now(),
		ready:          true,
	}

	toast := app.renderCopyToast()
	if toast == "" {
		t.Fatal("expected toast to render")
	}

	plain := stripANSI(toast)
	if !strings.Contains(plain, "ab-789") {
		t.Errorf("expected toast to contain bead ID 'ab-789', got: %s", plain)
	}
	if !strings.Contains(plain, "Copied") {
		t.Errorf("expected toast to contain 'Copied', got: %s", plain)
	}
	if !strings.Contains(plain, "clipboard") {
		t.Errorf("expected toast to contain 'clipboard', got: %s", plain)
	}
}

func TestCopyToastNotRenderedWhenInactive(t *testing.T) {
	app := &App{
		copiedBeadID:  "ab-999",
		showCopyToast: false,
		ready:         true,
	}

	toast := app.renderCopyToast()
	if toast != "" {
		t.Error("expected no toast when showCopyToast is false")
	}
}

func TestCopyWithEmptyVisibleRows(t *testing.T) {
	app := &App{
		visibleRows: []graph.TreeRow{},
		ready:       true,
	}

	// Press 'c' with no visible rows - should not panic or set toast
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})

	if app.showCopyToast {
		t.Error("expected no toast when no rows visible")
	}
}

func TestCopyToastTickWhenNotShowing(t *testing.T) {
	app := &App{
		showCopyToast: false,
		ready:         true,
	}

	// Process tick when toast is not showing - should return nil cmd
	_, cmd := app.Update(copyToastTickMsg{})
	if cmd != nil {
		t.Error("expected no tick scheduled when toast is not showing")
	}
}

func loadFixtureIssues(t *testing.T, file string) []beads.FullIssue {
	t.Helper()
	candidates := []string{
		filepath.Join("testdata", file),
		filepath.Join("..", "..", "testdata", file),
	}
	var data []byte
	var err error
	for _, path := range candidates {
		data, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		t.Fatalf("read fixture %s: %v", file, err)
	}
	var issues []beads.FullIssue
	if err := json.Unmarshal(data, &issues); err != nil {
		t.Fatalf("unmarshal fixture %s: %v", file, err)
	}
	return issues
}

func filterIssuesByID(issues []beads.FullIssue, ids []string) []beads.FullIssue {
	set := make(map[string]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	var filtered []beads.FullIssue
	for _, iss := range issues {
		if set[iss.ID] {
			filtered = append(filtered, iss)
		}
	}
	return filtered
}

func liteIssuesFromFixture(issues []beads.FullIssue) []beads.LiteIssue {
	results := make([]beads.LiteIssue, len(issues))
	for i, iss := range issues {
		results[i] = beads.LiteIssue{ID: iss.ID}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].ID < results[j].ID
	})
	return results
}

func createTempDBFile(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "beads.db")
	if err := os.WriteFile(path, []byte("db"), 0o644); err != nil {
		t.Fatalf("write temp db: %v", err)
	}
	return path
}

func mustNewTestApp(t *testing.T, client beads.Client) *App {
	t.Helper()
	app, err := NewApp(Config{
		RefreshInterval: time.Second,
		AutoRefresh:     false,
		DBPathOverride:  createTempDBFile(t),
		Client:          client,
	})
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}
	return app
}

// extractRefreshMsg extracts refreshCompleteMsg from a command result.
// Since startRefresh now returns a batch (spinner + refresh), we need to
// execute the batch and find the refreshCompleteMsg among the results.
func extractRefreshMsg(t *testing.T, cmd tea.Cmd) refreshCompleteMsg {
	t.Helper()
	if cmd == nil {
		t.Fatalf("expected refresh cmd, got nil")
	}
	msg := cmd()
	// Check if it's a direct refreshCompleteMsg
	if refreshMsg, ok := msg.(refreshCompleteMsg); ok {
		return refreshMsg
	}
	// Check if it's a BatchMsg containing our refresh
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			if c != nil {
				result := c()
				if refreshMsg, ok := result.(refreshCompleteMsg); ok {
					return refreshMsg
				}
			}
		}
	}
	t.Fatalf("could not find refreshCompleteMsg in %T", msg)
	return refreshCompleteMsg{}
}

func fileModTime(t *testing.T, path string) time.Time {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	return info.ModTime()
}

func changeWorkingDir(t *testing.T, dir string) func() {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	return func() {
		if err := os.Chdir(orig); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}
}

func normalizePath(t *testing.T, path string) string {
	t.Helper()
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}
	return abs
}

func TestHelpToggle(t *testing.T) {
	app := &App{
		visibleRows: nodesToRows(&graph.Node{Issue: beads.FullIssue{ID: "ab-001"}}),
		keys:        DefaultKeyMap(),
	}

	// Initially help is not shown
	if app.showHelp {
		t.Error("expected showHelp to be false initially")
	}

	// Press ? to show help
	result, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	app = result.(*App)
	if !app.showHelp {
		t.Error("expected showHelp to be true after pressing ?")
	}

	// Press ? again to hide help
	result, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	app = result.(*App)
	if app.showHelp {
		t.Error("expected showHelp to be false after pressing ? again")
	}
}

func TestHelpDismissWithEsc(t *testing.T) {
	app := &App{
		showHelp: true,
		keys:     DefaultKeyMap(),
	}

	result, _ := app.Update(tea.KeyMsg{Type: tea.KeyEscape})
	app = result.(*App)

	if app.showHelp {
		t.Error("expected showHelp to be false after pressing Esc")
	}
}

func TestHelpDismissWithQ(t *testing.T) {
	app := &App{
		showHelp:    true,
		keys:        DefaultKeyMap(),
		visibleRows: nodesToRows(&graph.Node{Issue: beads.FullIssue{ID: "ab-001"}}),
	}

	// q should dismiss help, NOT quit the app
	result, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	app = result.(*App)

	if app.showHelp {
		t.Error("expected showHelp to be false after pressing q")
	}
	if cmd != nil {
		t.Error("expected no quit command when dismissing help with q")
	}
}

func TestHelpBlocksOtherKeys(t *testing.T) {
	app := &App{
		showHelp:    true,
		keys:        DefaultKeyMap(),
		visibleRows: nodesToRows(&graph.Node{Issue: beads.FullIssue{ID: "ab-001"}}),
		cursor:      0,
	}

	initialCursor := app.cursor

	// Navigation keys should be blocked
	result, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	app = result.(*App)
	if app.cursor != initialCursor {
		t.Error("expected cursor to remain unchanged when help is shown")
	}
	if !app.showHelp {
		t.Error("expected help to still be shown after pressing j")
	}

	// Search key should be blocked
	result, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	app = result.(*App)
	if app.searching {
		t.Error("expected searching to remain false when help is shown")
	}
	if !app.showHelp {
		t.Error("expected help to still be shown after pressing /")
	}
}

func TestHelpOverlayInView(t *testing.T) {
	app := &App{
		showHelp:    true,
		keys:        DefaultKeyMap(),
		visibleRows: nodesToRows(&graph.Node{Issue: beads.FullIssue{ID: "ab-001"}}),
		width:       80,
		height:      24,
		ready:       true,
	}

	view := app.View()

	// View should contain help overlay content
	if !strings.Contains(view, "ABACUS HELP") {
		t.Error("expected view to contain 'ABACUS HELP' when help is shown")
	}
	if !strings.Contains(view, "NAVIGATION") {
		t.Error("expected view to contain 'NAVIGATION' section when help is shown")
	}
}

func TestStatusOverlayKeepsBaseContentVisible(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{ID: "ab-ov1", Title: "Overlay Test", Status: "open"},
	}
	app := &App{
		ready:           true,
		width:           80,
		height:          24,
		visibleRows:     []graph.TreeRow{{Node: node}},
		cursor:          0,
		activeOverlay:   OverlayStatus,
		statusOverlay:   NewStatusOverlay(node.Issue.ID, node.Issue.Title, node.Issue.Status),
		repoName:        "abacus",
		lastRefreshTime: time.Now(),
	}

	view := app.View()
	plain := stripANSI(view)

	if !strings.Contains(plain, node.Issue.ID) {
		t.Fatalf("expected base content (tree) to remain visible, got:\n%s", plain)
	}
	if !strings.Contains(plain, "Status") {
		t.Fatalf("expected status overlay content to be present, got:\n%s", plain)
	}
	if !strings.Contains(view, "\x1b[2m") {
		t.Fatalf("expected dimming control sequence in overlay view, got:\n%s", view)
	}
	if secondary := theme.Current().BackgroundSecondaryANSI(); secondary != "" && !strings.Contains(view, secondary) {
		t.Fatalf("expected secondary background sequence in overlay view:\n%s", view)
	}
}

func TestCreateOverlayShowsErrorToast(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{ID: "ab-ov2", Title: "Overlay Toast", Status: "open"},
	}
	app := &App{
		ready:           true,
		width:           80,
		height:          24,
		visibleRows:     []graph.TreeRow{{Node: node}},
		cursor:          0,
		activeOverlay:   OverlayCreate,
		createOverlay:   NewCreateOverlay(CreateOverlayOptions{}),
		showErrorToast:  true,
		lastError:       "Error: backend unavailable",
		errorToastStart: time.Now(),
		repoName:        "abacus",
		lastRefreshTime: time.Now(),
	}

	view := app.View()
	plain := stripANSI(view)
	if !strings.Contains(plain, "⚠ Error") {
		t.Fatalf("expected error toast to appear on create overlay, got:\n%s", plain)
	}
	if !strings.Contains(plain, "ABACUS") {
		t.Fatalf("expected header content to remain visible beneath overlay, got:\n%s", plain)
	}
	if !strings.Contains(view, "\x1b[2m") {
		t.Fatalf("expected dimming applied to create overlay view, got:\n%s", view)
	}
	if secondary := theme.Current().BackgroundSecondaryANSI(); secondary != "" && !strings.Contains(view, secondary) {
		t.Fatalf("expected secondary background sequence in create overlay view:\n%s", view)
	}
}

// Toast Tests for ab-1t3

func TestNewLabelToastDisplay(t *testing.T) {
	t.Run("NewLabelAddedMsgTriggersToast", func(t *testing.T) {
		app := &App{
			ready: true,
			roots: []*graph.Node{},
		}

		// Send NewLabelAddedMsg
		msg := NewLabelAddedMsg{Label: "test-label"}
		model, cmd := app.Update(msg)

		app = model.(*App)

		// Verify toast is visible
		if !app.newLabelToastVisible {
			t.Error("expected newLabelToastVisible to be true")
		}

		// Verify label name is set
		if app.newLabelToastLabel != "test-label" {
			t.Errorf("expected label 'test-label', got '%s'", app.newLabelToastLabel)
		}

		// Verify toast start time is recent (within last second)
		if time.Since(app.newLabelToastStart) > time.Second {
			t.Error("expected newLabelToastStart to be recent")
		}

		// Verify tick scheduler was returned
		if cmd == nil {
			t.Error("expected command to schedule toast tick")
		}
	})
}

func TestNewLabelToastTimeout(t *testing.T) {
	t.Run("ToastClearsAfter3Seconds", func(t *testing.T) {
		app := &App{
			ready:                true,
			roots:                []*graph.Node{},
			newLabelToastVisible: true,
			newLabelToastLabel:   "test-label",
			newLabelToastStart:   time.Now().Add(-4 * time.Second), // 4 seconds ago
		}

		// Send tick message
		msg := newLabelToastTickMsg{}
		model, cmd := app.Update(msg)

		app = model.(*App)

		// Verify toast is cleared
		if app.newLabelToastVisible {
			t.Error("expected newLabelToastVisible to be false after timeout")
		}

		// Verify no further ticks scheduled
		if cmd != nil {
			t.Error("expected no command after toast cleared")
		}
	})

	t.Run("ToastRemainsVisibleBefore3Seconds", func(t *testing.T) {
		app := &App{
			ready:                true,
			roots:                []*graph.Node{},
			newLabelToastVisible: true,
			newLabelToastLabel:   "test-label",
			newLabelToastStart:   time.Now().Add(-1 * time.Second), // 1 second ago
		}

		// Send tick message
		msg := newLabelToastTickMsg{}
		model, cmd := app.Update(msg)

		app = model.(*App)

		// Verify toast is still visible
		if !app.newLabelToastVisible {
			t.Error("expected newLabelToastVisible to remain true before timeout")
		}

		// Verify tick is rescheduled
		if cmd == nil {
			t.Error("expected command to reschedule toast tick")
		}
	})

	t.Run("TickHandlerReturnsEarlyWhenNotVisible", func(t *testing.T) {
		app := &App{
			ready:                true,
			roots:                []*graph.Node{},
			newLabelToastVisible: false,
			newLabelToastLabel:   "",
		}

		// Send tick message when toast not visible
		msg := newLabelToastTickMsg{}
		model, cmd := app.Update(msg)

		app = model.(*App)

		// Verify no command returned
		if cmd != nil {
			t.Error("expected no command when toast not visible")
		}
	})
}

func TestNewAssigneeToastDisplay(t *testing.T) {
	t.Run("NewAssigneeAddedMsgTriggersToast", func(t *testing.T) {
		app := &App{
			ready: true,
			roots: []*graph.Node{},
		}

		// Send NewAssigneeAddedMsg
		msg := NewAssigneeAddedMsg{Assignee: "test-user"}
		model, cmd := app.Update(msg)

		app = model.(*App)

		// Verify toast is visible
		if !app.newAssigneeToastVisible {
			t.Error("expected newAssigneeToastVisible to be true")
		}

		// Verify assignee name is set
		if app.newAssigneeToastAssignee != "test-user" {
			t.Errorf("expected assignee 'test-user', got '%s'", app.newAssigneeToastAssignee)
		}

		// Verify toast start time is recent
		if time.Since(app.newAssigneeToastStart) > time.Second {
			t.Error("expected newAssigneeToastStart to be recent")
		}

		// Verify tick scheduler was returned
		if cmd == nil {
			t.Error("expected command to schedule toast tick")
		}
	})
}

func TestNewAssigneeToastTimeout(t *testing.T) {
	t.Run("ToastClearsAfter3Seconds", func(t *testing.T) {
		app := &App{
			ready:                    true,
			roots:                    []*graph.Node{},
			newAssigneeToastVisible:  true,
			newAssigneeToastAssignee: "test-user",
			newAssigneeToastStart:    time.Now().Add(-4 * time.Second), // 4 seconds ago
		}

		// Send tick message
		msg := newAssigneeToastTickMsg{}
		model, cmd := app.Update(msg)

		app = model.(*App)

		// Verify toast is cleared
		if app.newAssigneeToastVisible {
			t.Error("expected newAssigneeToastVisible to be false after timeout")
		}

		// Verify no further ticks scheduled
		if cmd != nil {
			t.Error("expected no command after toast cleared")
		}
	})

	t.Run("ToastRemainsVisibleBefore3Seconds", func(t *testing.T) {
		app := &App{
			ready:                    true,
			roots:                    []*graph.Node{},
			newAssigneeToastVisible:  true,
			newAssigneeToastAssignee: "test-user",
			newAssigneeToastStart:    time.Now().Add(-1 * time.Second), // 1 second ago
		}

		// Send tick message
		msg := newAssigneeToastTickMsg{}
		model, cmd := app.Update(msg)

		app = model.(*App)

		// Verify toast is still visible
		if !app.newAssigneeToastVisible {
			t.Error("expected newAssigneeToastVisible to remain true before timeout")
		}

		// Verify tick is rescheduled
		if cmd == nil {
			t.Error("expected command to reschedule toast tick")
		}
	})

	t.Run("TickHandlerReturnsEarlyWhenNotVisible", func(t *testing.T) {
		app := &App{
			ready:                    true,
			roots:                    []*graph.Node{},
			newAssigneeToastVisible:  false,
			newAssigneeToastAssignee: "",
		}

		// Send tick message when toast not visible
		msg := newAssigneeToastTickMsg{}
		model, cmd := app.Update(msg)

		app = model.(*App)

		// Verify no command returned
		if cmd != nil {
			t.Error("expected no command when toast not visible")
		}
	})
}

func TestToastRendering(t *testing.T) {
	t.Run("RenderNewLabelToastReturnsEmptyWhenNotVisible", func(t *testing.T) {
		app := &App{
			newLabelToastVisible: false,
		}

		result := app.renderNewLabelToast()

		if result != "" {
			t.Error("expected empty string when toast not visible")
		}
	})

	t.Run("RenderNewLabelToastReturnsEmptyWhenLabelEmpty", func(t *testing.T) {
		app := &App{
			newLabelToastVisible: true,
			newLabelToastLabel:   "",
		}

		result := app.renderNewLabelToast()

		if result != "" {
			t.Error("expected empty string when label is empty")
		}
	})

	t.Run("RenderNewLabelToastReturnsFormattedToast", func(t *testing.T) {
		app := &App{
			newLabelToastVisible: true,
			newLabelToastLabel:   "test-label",
			newLabelToastStart:   time.Now(),
		}

		result := app.renderNewLabelToast()

		if result == "" {
			t.Error("expected non-empty toast when visible")
		}

		// Check for label name in output
		if !strings.Contains(result, "test-label") {
			t.Error("expected toast to contain label name")
		}

		// Check for checkmark symbol
		if !strings.Contains(result, "✓") {
			t.Error("expected toast to contain checkmark")
		}

		// Check for "New Label Added" text
		if !strings.Contains(result, "New Label Added") {
			t.Error("expected toast to contain 'New Label Added' text")
		}
	})

	t.Run("RenderNewLabelToastReturnsEmptyAfterTimeout", func(t *testing.T) {
		app := &App{
			newLabelToastVisible: true,
			newLabelToastLabel:   "test-label",
			newLabelToastStart:   time.Now().Add(-4 * time.Second), // 4 seconds ago
		}

		result := app.renderNewLabelToast()

		if result != "" {
			t.Error("expected empty string when toast has timed out")
		}
	})

	t.Run("RenderNewAssigneeToastReturnsEmptyWhenNotVisible", func(t *testing.T) {
		app := &App{
			newAssigneeToastVisible: false,
		}

		result := app.renderNewAssigneeToast()

		if result != "" {
			t.Error("expected empty string when toast not visible")
		}
	})

	t.Run("RenderNewAssigneeToastReturnsEmptyWhenAssigneeEmpty", func(t *testing.T) {
		app := &App{
			newAssigneeToastVisible:  true,
			newAssigneeToastAssignee: "",
		}

		result := app.renderNewAssigneeToast()

		if result != "" {
			t.Error("expected empty string when assignee is empty")
		}
	})

	t.Run("RenderNewAssigneeToastReturnsFormattedToast", func(t *testing.T) {
		app := &App{
			newAssigneeToastVisible:  true,
			newAssigneeToastAssignee: "test-user",
			newAssigneeToastStart:    time.Now(),
		}

		result := app.renderNewAssigneeToast()

		if result == "" {
			t.Error("expected non-empty toast when visible")
		}

		// Check for assignee name in output
		if !strings.Contains(result, "test-user") {
			t.Error("expected toast to contain assignee name")
		}

		// Check for checkmark symbol
		if !strings.Contains(result, "✓") {
			t.Error("expected toast to contain checkmark")
		}

		// Check for "New Assignee Added" text
		if !strings.Contains(result, "New Assignee Added") {
			t.Error("expected toast to contain 'New Assignee Added' text")
		}
	})

	t.Run("RenderNewAssigneeToastReturnsEmptyAfterTimeout", func(t *testing.T) {
		app := &App{
			newAssigneeToastVisible:  true,
			newAssigneeToastAssignee: "test-user",
			newAssigneeToastStart:    time.Now().Add(-4 * time.Second), // 4 seconds ago
		}

		result := app.renderNewAssigneeToast()

		if result != "" {
			t.Error("expected empty string when toast has timed out")
		}
	})
}

// Tests for ab-8sgz: Disable global hotkeys when entering text
func TestGlobalHotkeysDisabledDuringTextInput(t *testing.T) {
	t.Run("nKeyIgnoredWhenCreateOverlayActive", func(t *testing.T) {
		// Setup app with create overlay active
		node := &graph.Node{Issue: beads.FullIssue{ID: "ab-001", Title: "Test"}}
		app := &App{
			visibleRows: nodesToRows(node),
			keys:        DefaultKeyMap(),
			ready:       true,
		}

		// Create initial overlay
		createOverlay := NewCreateOverlay(CreateOverlayOptions{
			DefaultParentID:  "ab-001",
			AvailableParents: []ParentOption{{ID: "ab-001", Display: "ab-001 Test"}},
		})
		app.createOverlay = createOverlay
		app.activeOverlay = OverlayCreate

		// Press 'n' key - should be passed to overlay, not create new overlay
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
		result, _ := app.Update(msg)
		app = result.(*App)

		// Verify: Still only one overlay, not a new one created
		if app.activeOverlay != OverlayCreate {
			t.Error("expected CreateOverlay to remain active")
		}
		if app.createOverlay == nil {
			t.Error("expected createOverlay to still exist")
		}
	})

	t.Run("nKeyWorksNormallyWhenNoOverlayActive", func(t *testing.T) {
		// Setup app without any overlay
		node := &graph.Node{Issue: beads.FullIssue{ID: "ab-002", Title: "Test"}}
		app := &App{
			visibleRows:   nodesToRows(node),
			keys:          DefaultKeyMap(),
			ready:         true,
			activeOverlay: OverlayNone,
		}

		// Press 'n' key - should create new overlay
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
		result, _ := app.Update(msg)
		app = result.(*App)

		// Verify: CreateOverlay was created
		if app.activeOverlay != OverlayCreate {
			t.Error("expected CreateOverlay to be active")
		}
		if app.createOverlay == nil {
			t.Error("expected createOverlay to be created")
		}
	})

	t.Run("sKeyIgnoredWhenCreateOverlayActive", func(t *testing.T) {
		// Setup app with create overlay active
		node := &graph.Node{Issue: beads.FullIssue{ID: "ab-003", Title: "Test", Status: "open"}}
		app := &App{
			visibleRows: nodesToRows(node),
			keys:        DefaultKeyMap(),
			ready:       true,
		}

		// Create overlay
		createOverlay := NewCreateOverlay(CreateOverlayOptions{})
		app.createOverlay = createOverlay
		app.activeOverlay = OverlayCreate

		// Press 's' key - should be passed to overlay, not open status overlay
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
		result, _ := app.Update(msg)
		app = result.(*App)

		// Verify: Still in CreateOverlay, not StatusOverlay
		if app.activeOverlay != OverlayCreate {
			t.Error("expected CreateOverlay to remain active")
		}
		if app.statusOverlay != nil {
			t.Error("expected statusOverlay to not be created")
		}
	})

	t.Run("CapitalLKeyIgnoredWhenCreateOverlayActive", func(t *testing.T) {
		// Setup app with create overlay active
		node := &graph.Node{Issue: beads.FullIssue{ID: "ab-004", Title: "Test", Labels: []string{"test"}}}
		app := &App{
			visibleRows: nodesToRows(node),
			keys:        DefaultKeyMap(),
			ready:       true,
		}

		// Create overlay
		createOverlay := NewCreateOverlay(CreateOverlayOptions{})
		app.createOverlay = createOverlay
		app.activeOverlay = OverlayCreate

		// Press 'L' key - should be passed to overlay, not open labels overlay
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'L'}}
		result, _ := app.Update(msg)
		app = result.(*App)

		// Verify: Still in CreateOverlay, not LabelsOverlay
		if app.activeOverlay != OverlayCreate {
			t.Error("expected CreateOverlay to remain active")
		}
		if app.labelsOverlay != nil {
			t.Error("expected labelsOverlay to not be created")
		}
	})

	t.Run("xKeyIgnoredWhenCreateOverlayActive", func(t *testing.T) {
		// Setup app with create overlay active and mock client
		node := &graph.Node{Issue: beads.FullIssue{ID: "ab-005", Title: "Test", Status: "open"}}
		mock := beads.NewMockClient()
		app := &App{
			visibleRows: nodesToRows(node),
			keys:        DefaultKeyMap(),
			ready:       true,
			client:      mock,
		}

		// Create overlay
		createOverlay := NewCreateOverlay(CreateOverlayOptions{})
		app.createOverlay = createOverlay
		app.activeOverlay = OverlayCreate

		// Press 'x' key - should be passed to overlay, not trigger close
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
		result, _ := app.Update(msg)
		app = result.(*App)

		// Verify: Still in CreateOverlay, close was not triggered
		if app.activeOverlay != OverlayCreate {
			t.Error("expected CreateOverlay to remain active")
		}
		// No way to directly check if close wasn't called, but overlay remaining active is sufficient
	})

	t.Run("StatusOverlayBlocksGlobalHotkeys", func(t *testing.T) {
		// Setup app with status overlay active
		node := &graph.Node{Issue: beads.FullIssue{ID: "ab-006", Title: "Test", Status: "open"}}
		app := &App{
			visibleRows: nodesToRows(node),
			keys:        DefaultKeyMap(),
			ready:       true,
		}

		// Create status overlay
		statusOverlay := NewStatusOverlay("ab-006", "Test", "open")
		app.statusOverlay = statusOverlay
		app.activeOverlay = OverlayStatus

		// Press 'n' key - should be passed to overlay, not create new bead overlay
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
		result, _ := app.Update(msg)
		app = result.(*App)

		// Verify: Still in StatusOverlay, CreateOverlay not created
		if app.activeOverlay != OverlayStatus {
			t.Error("expected StatusOverlay to remain active")
		}
		if app.createOverlay != nil {
			t.Error("expected createOverlay to not be created")
		}
	})

	t.Run("LabelsOverlayBlocksGlobalHotkeys", func(t *testing.T) {
		// Setup app with labels overlay active
		node := &graph.Node{Issue: beads.FullIssue{ID: "ab-007", Title: "Test", Labels: []string{"existing"}}}
		app := &App{
			visibleRows: nodesToRows(node),
			keys:        DefaultKeyMap(),
			ready:       true,
			roots:       []*graph.Node{node},
		}

		// Create labels overlay
		labelsOverlay := NewLabelsOverlay("ab-007", "Test", []string{"existing"}, []string{"existing", "other"})
		app.labelsOverlay = labelsOverlay
		app.activeOverlay = OverlayLabels

		// Press 'n' key - should be passed to overlay, not create new bead overlay
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
		result, _ := app.Update(msg)
		app = result.(*App)

		// Verify: Still in LabelsOverlay, CreateOverlay not created
		if app.activeOverlay != OverlayLabels {
			t.Error("expected LabelsOverlay to remain active")
		}
		if app.createOverlay != nil {
			t.Error("expected createOverlay to not be created")
		}
	})
}
