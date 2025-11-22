package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
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
		ready := &Node{Issue: FullIssue{Title: "Ready Task", Status: "open"}}
		inProgress := &Node{
			Issue:    FullIssue{Title: "In Progress", Status: "in_progress"},
			Children: []*Node{ready},
		}
		closed := &Node{Issue: FullIssue{Title: "Closed Task", Status: "closed"}}
		blocked := &Node{Issue: FullIssue{Title: "Blocked Task", Status: "open"}, IsBlocked: true}

		m := model{
			roots: []*Node{inProgress, closed, blocked},
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
		matching := &Node{Issue: FullIssue{Title: "Alpha Ready", Status: "open"}}
		nonMatching := &Node{Issue: FullIssue{Title: "Bravo Active", Status: "in_progress"}}
		m := model{
			roots:      []*Node{matching, nonMatching},
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
	t.Run("marksAllNodesAsLoaded", func(t *testing.T) {
		// Create a simple tree structure
		root := &Node{
			Issue: FullIssue{ID: "ab-001", Title: "Root Issue"},
			Children: []*Node{
				{Issue: FullIssue{ID: "ab-002", Title: "Child Issue"}},
			},
		}

		preloadAllComments([]*Node{root})

		// Verify root node
		if !root.CommentsLoaded {
			t.Errorf("expected root node CommentsLoaded to be true")
		}
		if root.Issue.Comments == nil {
			t.Errorf("expected root node Comments to be initialized")
		}

		// Verify child node
		if !root.Children[0].CommentsLoaded {
			t.Errorf("expected child node CommentsLoaded to be true")
		}
		if root.Children[0].Issue.Comments == nil {
			t.Errorf("expected child node Comments to be initialized")
		}
	})

	t.Run("handlesMultipleRoots", func(t *testing.T) {
		roots := []*Node{
			{Issue: FullIssue{ID: "ab-010", Title: "First Root"}},
			{Issue: FullIssue{ID: "ab-011", Title: "Second Root"}},
			{Issue: FullIssue{ID: "ab-012", Title: "Third Root"}},
		}

		preloadAllComments(roots)

		for i, root := range roots {
			if !root.CommentsLoaded {
				t.Errorf("root %d (%s) not marked as loaded", i, root.Issue.ID)
			}
		}
	})

	t.Run("handlesNestedChildren", func(t *testing.T) {
		// Create a deeper tree structure
		deepChild := &Node{Issue: FullIssue{ID: "ab-023", Title: "Deep Child"}}
		midChild := &Node{
			Issue:    FullIssue{ID: "ab-022", Title: "Mid Child"},
			Children: []*Node{deepChild},
		}
		root := &Node{
			Issue:    FullIssue{ID: "ab-021", Title: "Root"},
			Children: []*Node{midChild},
		}

		preloadAllComments([]*Node{root})

		// Verify all levels are loaded
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
		// Should not panic on empty input
		preloadAllComments([]*Node{})
		preloadAllComments(nil)
	})

	t.Run("initializesEmptyCommentsSlice", func(t *testing.T) {
		node := &Node{Issue: FullIssue{ID: "ab-030", Title: "No Comments"}}
		preloadAllComments([]*Node{node})

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
	child := &Node{Issue: FullIssue{ID: "ab-002"}}
	root := &Node{
		Issue:    FullIssue{ID: "ab-001"},
		Children: []*Node{child},
		Expanded: true,
	}

	m := model{
		roots:       []*Node{root},
		visibleRows: []*Node{root, child},
		cursor:      1,
		filterText:  "alpha",
		showDetails: true,
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
}

func TestRestoreExpandedState(t *testing.T) {
	child := &Node{Issue: FullIssue{ID: "ab-002"}}
	root := &Node{Issue: FullIssue{ID: "ab-001"}, Children: []*Node{child}}
	m := model{roots: []*Node{root}}

	m.restoreExpandedState(map[string]bool{"ab-001": true})

	if !root.Expanded {
		t.Fatalf("expected root expanded")
	}
	if child.Expanded {
		t.Fatalf("expected child collapsed")
	}
}

func TestRestoreCursorToID(t *testing.T) {
	n1 := &Node{Issue: FullIssue{ID: "ab-001"}}
	n2 := &Node{Issue: FullIssue{ID: "ab-002"}}
	m := model{
		visibleRows: []*Node{n1, n2},
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
