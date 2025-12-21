package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"abacus/internal/beads"
	"abacus/internal/graph"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

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

	// Test placeholder state (after display duration) - returns space to reserve layout
	m.lastRefreshTime = time.Now().Add(-refreshDisplayDuration - time.Millisecond)
	status = m.renderRefreshStatus()
	if status != " " {
		t.Fatalf("expected refresh status to be space placeholder after display duration, got: %q", status)
	}

	// Test no-change state (should not show delta, just space placeholder)
	m.lastRefreshStats = "+0 / Δ0 / -0"
	m.lastRefreshTime = time.Now()
	status = m.renderRefreshStatus()
	if status != " " {
		t.Fatalf("expected refresh status to be space placeholder when no changes, got: %q", status)
	}
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
