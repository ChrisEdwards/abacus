package ui

import (
	"context"
	"testing"
	"time"

	"abacus/internal/beads"
	"abacus/internal/config"
	"abacus/internal/graph"

	tea "github.com/charmbracelet/bubbletea"
)

func TestViewModeCycling(t *testing.T) {
	t.Run("NextCyclesForward", func(t *testing.T) {
		mode := ViewModeAll
		mode = mode.Next()
		if mode != ViewModeActive {
			t.Errorf("expected ViewModeActive after Next(), got %v", mode)
		}
		mode = mode.Next()
		if mode != ViewModeReady {
			t.Errorf("expected ViewModeReady after Next(), got %v", mode)
		}
		mode = mode.Next()
		if mode != ViewModeAll {
			t.Errorf("expected ViewModeAll after wrapping, got %v", mode)
		}
	})

	t.Run("PrevCyclesBackward", func(t *testing.T) {
		mode := ViewModeAll
		mode = mode.Prev()
		if mode != ViewModeReady {
			t.Errorf("expected ViewModeReady after Prev(), got %v", mode)
		}
		mode = mode.Prev()
		if mode != ViewModeActive {
			t.Errorf("expected ViewModeActive after Prev(), got %v", mode)
		}
		mode = mode.Prev()
		if mode != ViewModeAll {
			t.Errorf("expected ViewModeAll after wrapping, got %v", mode)
		}
	})

	t.Run("String", func(t *testing.T) {
		if ViewModeAll.String() != "All" {
			t.Errorf("expected 'All', got %s", ViewModeAll.String())
		}
		if ViewModeActive.String() != "Active" {
			t.Errorf("expected 'Active', got %s", ViewModeActive.String())
		}
		if ViewModeReady.String() != "Ready" {
			t.Errorf("expected 'Ready', got %s", ViewModeReady.String())
		}
	})
}

func TestViewModeActiveHidesClosedIssues(t *testing.T) {
	openNode := &graph.Node{
		Issue: beads.FullIssue{ID: "ab-open", Title: "Open Issue", Status: "open"},
	}
	closedNode := &graph.Node{
		Issue: beads.FullIssue{ID: "ab-closed", Title: "Closed Issue", Status: "closed"},
	}
	inProgressNode := &graph.Node{
		Issue: beads.FullIssue{ID: "ab-wip", Title: "In Progress", Status: "in_progress"},
	}

	app := &App{
		roots:    []*graph.Node{openNode, closedNode, inProgressNode},
		viewMode: ViewModeAll,
		keys:     DefaultKeyMap(),
	}

	// ViewModeAll should show all 3 nodes
	app.recalcVisibleRows()
	if len(app.visibleRows) != 3 {
		t.Errorf("ViewModeAll: expected 3 visible rows, got %d", len(app.visibleRows))
	}

	// ViewModeActive should hide closed node (show 2)
	app.viewMode = ViewModeActive
	app.recalcVisibleRows()
	if len(app.visibleRows) != 2 {
		t.Errorf("ViewModeActive: expected 2 visible rows, got %d", len(app.visibleRows))
	}

	// Verify the closed node is not in visible rows
	for _, row := range app.visibleRows {
		if row.Node.Issue.Status == "closed" {
			t.Error("ViewModeActive should hide closed issues")
		}
	}
}

func TestViewModeReadyShowsOnlyReadyIssues(t *testing.T) {
	readyNode := &graph.Node{
		Issue:     beads.FullIssue{ID: "ab-ready", Title: "Ready Issue", Status: "open"},
		IsBlocked: false,
	}
	blockedNode := &graph.Node{
		Issue:     beads.FullIssue{ID: "ab-blocked", Title: "Blocked Issue", Status: "open"},
		IsBlocked: true,
	}
	closedNode := &graph.Node{
		Issue: beads.FullIssue{ID: "ab-closed", Title: "Closed Issue", Status: "closed"},
	}
	inProgressNode := &graph.Node{
		Issue: beads.FullIssue{ID: "ab-wip", Title: "In Progress", Status: "in_progress"},
	}

	app := &App{
		roots:    []*graph.Node{readyNode, blockedNode, closedNode, inProgressNode},
		viewMode: ViewModeReady,
		keys:     DefaultKeyMap(),
	}

	app.recalcVisibleRows()

	// ViewModeReady should only show the ready node (open + not blocked)
	if len(app.visibleRows) != 1 {
		t.Errorf("ViewModeReady: expected 1 visible row, got %d", len(app.visibleRows))
	}
	if len(app.visibleRows) > 0 && app.visibleRows[0].Node.Issue.ID != "ab-ready" {
		t.Errorf("ViewModeReady: expected ab-ready to be visible, got %s", app.visibleRows[0].Node.Issue.ID)
	}
}

func TestViewModePreservesTreeHierarchy(t *testing.T) {
	// Child is open, parent is closed
	// ViewModeActive should show BOTH (parent shown because child matches)
	openChild := &graph.Node{
		Issue: beads.FullIssue{ID: "ab-child", Title: "Open Child", Status: "open"},
	}
	closedParent := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-parent", Title: "Closed Parent", Status: "closed"},
		Children: []*graph.Node{openChild},
		Expanded: true,
	}
	openChild.Parent = closedParent

	app := &App{
		roots:    []*graph.Node{closedParent},
		viewMode: ViewModeActive,
		keys:     DefaultKeyMap(),
	}

	app.recalcVisibleRows()

	// Both should be visible: parent (due to child match) and child
	if len(app.visibleRows) != 2 {
		t.Errorf("ViewModeActive with tree hierarchy: expected 2 visible rows (parent+child), got %d", len(app.visibleRows))
	}
}

func TestViewModeReadyPreservesTreeHierarchy(t *testing.T) {
	// Child is ready (open + not blocked), parent is closed
	// ViewModeReady should show BOTH (parent shown because child matches)
	readyChild := &graph.Node{
		Issue:     beads.FullIssue{ID: "ab-child", Title: "Ready Child", Status: "open"},
		IsBlocked: false,
	}
	closedParent := &graph.Node{
		Issue:    beads.FullIssue{ID: "ab-parent", Title: "Closed Parent", Status: "closed"},
		Children: []*graph.Node{readyChild},
		Expanded: true,
	}
	readyChild.Parent = closedParent

	app := &App{
		roots:    []*graph.Node{closedParent},
		viewMode: ViewModeReady,
		keys:     DefaultKeyMap(),
	}

	app.recalcVisibleRows()

	// Both should be visible: parent (due to child match) and child
	if len(app.visibleRows) != 2 {
		t.Errorf("ViewModeReady with tree hierarchy: expected 2 visible rows (parent+child), got %d", len(app.visibleRows))
	}
	// Verify child is actually in the list
	hasChild := false
	for _, row := range app.visibleRows {
		if row.Node.Issue.ID == "ab-child" {
			hasChild = true
			break
		}
	}
	if !hasChild {
		t.Error("ViewModeReady: expected ready child to be visible")
	}
}

func TestViewModeWithSearchFilter(t *testing.T) {
	matchingOpen := &graph.Node{
		Issue: beads.FullIssue{ID: "ab-1", Title: "Bug fix for login", Status: "open"},
	}
	matchingClosed := &graph.Node{
		Issue: beads.FullIssue{ID: "ab-2", Title: "Bug fix for settings", Status: "closed"},
	}
	nonMatching := &graph.Node{
		Issue: beads.FullIssue{ID: "ab-3", Title: "Feature request", Status: "open"},
	}

	app := &App{
		roots:      []*graph.Node{matchingOpen, matchingClosed, nonMatching},
		viewMode:   ViewModeActive,
		filterText: "bug",
		keys:       DefaultKeyMap(),
	}

	app.recalcVisibleRows()

	// Should only show matching open issue (combines ViewMode AND text filter)
	if len(app.visibleRows) != 1 {
		t.Errorf("ViewModeActive + search: expected 1 visible row, got %d", len(app.visibleRows))
	}
	if len(app.visibleRows) > 0 && app.visibleRows[0].Node.Issue.ID != "ab-1" {
		t.Errorf("ViewModeActive + search: expected ab-1, got %s", app.visibleRows[0].Node.Issue.ID)
	}
}

func TestViewModeKeyHandler(t *testing.T) {
	app := &App{
		ready:    true,
		viewMode: ViewModeAll,
		keys:     DefaultKeyMap(),
		roots:    []*graph.Node{},
	}

	// Press 'v' to cycle forward
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}}
	result, _ := app.Update(msg)
	app = result.(*App)

	if app.viewMode != ViewModeActive {
		t.Errorf("expected ViewModeActive after pressing 'v', got %v", app.viewMode)
	}

	// Press 'V' (shift+v) to cycle backward
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'V'}}
	result, _ = app.Update(msg)
	app = result.(*App)

	if app.viewMode != ViewModeAll {
		t.Errorf("expected ViewModeAll after pressing 'V', got %v", app.viewMode)
	}
}

func TestTickAlwaysReschedules(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*App)
	}{
		{
			name:  "normal state",
			setup: func(app *App) {},
		},
		{
			name: "autoRefresh disabled",
			setup: func(app *App) {
				app.autoRefresh = false
			},
		},
		{
			name: "refresh in flight",
			setup: func(app *App) {
				app.refreshInFlight = true
			},
		},
		{
			name: "db path empty",
			setup: func(app *App) {
				app.dbPath = ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{
				refreshInterval: 3 * time.Second,
				autoRefresh:     true,
				dbPath:          "/nonexistent/path",
				keys:            DefaultKeyMap(),
			}
			tt.setup(app)

			_, cmd := app.Update(tickMsg{})

			if cmd == nil {
				t.Fatalf("tick must always reschedule, got nil command")
			}
		})
	}
}

func TestNewAppUsesConfiguredRefreshInterval(t *testing.T) {
	cleanup := config.ResetForTesting(t)
	defer cleanup()

	// Set a custom refresh interval via config (not the default)
	customSeconds := 42
	if err := config.Set(config.KeyAutoRefreshSeconds, customSeconds); err != nil {
		t.Fatalf("failed to set config: %v", err)
	}

	// Verify config returns our custom value
	if got := config.GetInt(config.KeyAutoRefreshSeconds); got != customSeconds {
		t.Fatalf("config not set correctly: expected %d, got %d", customSeconds, got)
	}

	// Set up mock client with required functions
	mock := beads.NewMockClient()
	mock.ExportFn = func(ctx context.Context) ([]beads.FullIssue, error) {
		return []beads.FullIssue{
			{ID: "ab-001", Title: "Test Issue", Status: "open"},
		}, nil
	}
	mock.CommentsFn = func(ctx context.Context, issueID string) ([]beads.Comment, error) {
		return nil, nil
	}

	// Create App with RefreshInterval=0 to trigger fallback
	dbFile := createTempDBFile(t)
	cfg := Config{
		RefreshInterval: 0, // Should fall back to config value
		DBPathOverride:  dbFile,
		Client:          mock,
	}

	app, err := NewApp(cfg)
	if err != nil {
		t.Fatalf("NewApp failed: %v", err)
	}

	// Verify the app uses the configured value, not the hardcoded default
	expected := time.Duration(customSeconds) * time.Second
	if app.refreshInterval != expected {
		t.Errorf("NewApp should use configured refresh interval: expected %v, got %v", expected, app.refreshInterval)
	}
}
