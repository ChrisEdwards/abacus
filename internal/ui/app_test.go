package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
		ready := &graph.Node{Issue: beads.FullIssue{Title: "Ready Task", Status: "open"}}
		inProgress := &graph.Node{
			Issue:    beads.FullIssue{Title: "In Progress", Status: "in_progress"},
			Children: []*graph.Node{ready},
		}
		closed := &graph.Node{Issue: beads.FullIssue{Title: "Closed Task", Status: "closed"}}
		blocked := &graph.Node{Issue: beads.FullIssue{Title: "Blocked Task", Status: "open"}, IsBlocked: true}

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
		matching := &graph.Node{Issue: beads.FullIssue{Title: "Alpha Ready", Status: "open"}}
		nonMatching := &graph.Node{Issue: beads.FullIssue{Title: "Bravo Active", Status: "in_progress"}}
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

		preloadAllComments(ctx, client, []*graph.Node{root})

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

		preloadAllComments(ctx, client, roots)

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

		preloadAllComments(ctx, client, []*graph.Node{root})

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
		preloadAllComments(ctx, client, []*graph.Node{})
		preloadAllComments(ctx, client, nil)
	})

	t.Run("initializesEmptyCommentsSlice", func(t *testing.T) {
		node := &graph.Node{Issue: beads.FullIssue{ID: "ab-030", Title: "No Comments"}}
		client := beads.NewMockClient()
		client.CommentsFn = func(ctx context.Context, issueID string) ([]beads.Comment, error) {
			return nil, nil
		}
		preloadAllComments(ctx, client, []*graph.Node{node})

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
		visibleRows: []*graph.Node{root, child},
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
		visibleRows: []*graph.Node{n1, n2},
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
		visibleRows: []*graph.Node{rootOld, childOld},
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
	if len(m.visibleRows) == 0 || m.visibleRows[m.cursor].Issue.ID != "ab-002" {
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

func TestUpdateTogglesFocusWithTab(t *testing.T) {
	m := &App{ShowDetails: true, focus: FocusTree}
	m.visibleRows = []*graph.Node{{Issue: beads.FullIssue{ID: "ab-001"}}}

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
			visibleRows: []*graph.Node{{Issue: beads.FullIssue{ID: "ab-001"}}},
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
	if got := m.visibleRows[0].Issue.ID; got != "ab-123" {
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
	if got := m.visibleRows[0].Issue.ID; got != "ab-456" {
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
	if app.err != nil {
		t.Fatalf("expected app.err nil, got %v", app.err)
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
	app, err := NewApp(Config{
		RefreshInterval: time.Second,
		DBPathOverride:  dbFile,
		Client:          mock,
	})
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	if app.err == nil {
		t.Fatalf("expected app.err to capture client failure")
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

func TestOutputIssuesJSONUsesMockClient(t *testing.T) {
	fixture := loadFixtureIssues(t, "issues_basic.json")
	mock := beads.NewMockClient()
	mock.ListFn = func(ctx context.Context) ([]beads.LiteIssue, error) {
		return liteIssuesFromFixture(fixture), nil
	}
	mock.ShowFn = func(ctx context.Context, ids []string) ([]beads.FullIssue, error) {
		return filterIssuesByID(fixture, ids), nil
	}

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	if err := OutputIssuesJSON(context.Background(), mock); err != nil {
		t.Fatalf("OutputIssuesJSON: %v", err)
	}
	w.Close()
	os.Stdout = origStdout
	if _, err := io.ReadAll(r); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if mock.ShowCallCount == 0 {
		t.Fatalf("expected show to be called")
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
		visibleRows:  []*graph.Node{node},
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
		visibleRows:  []*graph.Node{node},
		viewport:     viewport.New(90, 30),
		outputFormat: "plain",
	}

	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	if strings.Contains(content, "Design") {
		t.Fatalf("expected Design section omitted when empty, content:\n%s", content)
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
