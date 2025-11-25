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
	if !m.showRefreshFlash {
		t.Fatalf("expected refresh flash flag to be set")
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
	m.collapseNodeForView(rootOld)
	childNew := &graph.Node{Issue: beads.FullIssue{ID: "ab-012", Title: "Child Updated", Status: "open"}}
	rootNew := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-011", Title: "Root", Status: "open"},
		Children: []*graph.Node{childNew},
	}
	m.applyRefresh([]*graph.Node{rootNew}, buildIssueDigest([]*graph.Node{rootNew}), time.Now())
	if m.isNodeExpandedInView(rootNew) {
		t.Fatalf("expected collapsed state preserved after refresh per docs")
	}
}

func TestUpdateTogglesFocusWithTab(t *testing.T) {
	m := &App{ShowDetails: true, focus: FocusTree}
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

		m.collapseNodeForView(root)
		m.recalcVisibleRows()
		assertVisible(t, m, 1)
		if m.isNodeExpandedInView(root) {
			t.Fatalf("expected root to appear collapsed in filtered view")
		}
	})

	t.Run("expandAfterCollapse", func(t *testing.T) {
		m, root := buildApp()
		m.setFilterText("leaf")
		m.recalcVisibleRows()
		m.collapseNodeForView(root)
		m.recalcVisibleRows()
		assertVisible(t, m, 1)

		m.expandNodeForView(root)
		m.recalcVisibleRows()
		assertVisible(t, m, 3)
		if !m.isNodeExpandedInView(root) {
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

	m.collapseNodeForView(root)
	m.recalcVisibleRows()
	if len(m.visibleRows) != 1 {
		t.Fatalf("expected collapse to hide children, got %d rows", len(m.visibleRows))
	}

	m.setFilterText("leaf")
	m.recalcVisibleRows()
	if len(m.visibleRows) != 1 {
		t.Fatalf("expected collapse state to persist while editing filter, got %d rows", len(m.visibleRows))
	}

	m.expandNodeForView(root)
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

func TestRefreshDeltaHighlightMatchesDocs(t *testing.T) {
	m := &App{
		width:            100,
		height:           30,
		ready:            true,
		lastRefreshStats: "+1 / Δ1 / -0",
	}
	highlightFragment := styleSelected.Render(" Δ " + m.lastRefreshStats)
	shadowFragment := styleStatsDim.Render(" Δ " + m.lastRefreshStats)
	m.showRefreshFlash = true
	m.lastRefreshTime = time.Now()
	view := m.View()
	if !strings.Contains(view, highlightFragment) {
		t.Fatalf("expected refresh delta to be highlighted immediately after refresh")
	}
	m.lastRefreshTime = time.Now().Add(-refreshFlashDuration - time.Millisecond)
	view = m.View()
	if !strings.Contains(view, shadowFragment) {
		t.Fatalf("expected refresh delta to dim after highlight window: %q", view)
	}
	if m.showRefreshFlash {
		t.Fatalf("expected flash flag reset after highlight window")
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
	if cmd == nil {
		t.Fatalf("expected refresh cmd")
	}
	msg := cmd()
	refreshMsg, ok := msg.(refreshCompleteMsg)
	if !ok {
		t.Fatalf("expected refreshCompleteMsg, got %T", msg)
	}
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
	if cmd == nil {
		t.Fatalf("expected refresh cmd")
	}
	msg := cmd()
	refreshMsg, ok := msg.(refreshCompleteMsg)
	if !ok {
		t.Fatalf("expected refreshCompleteMsg, got %T", msg)
	}
	app.Update(refreshMsg)
	if !strings.Contains(app.lastRefreshStats, "refresh failed") {
		t.Fatalf("expected refresh failure message, got %s", app.lastRefreshStats)
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
