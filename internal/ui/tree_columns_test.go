package ui

import (
	"testing"

	"abacus/internal/beads"
	"abacus/internal/config"
	"abacus/internal/graph"
)

func makeTestNodeWithComments(count int) *graph.Node {
	node := &graph.Node{CommentsLoaded: true}
	node.Issue.Comments = make([]beads.Comment, count)
	return node
}

func TestPrepareColumnState_ResponsiveHiding(t *testing.T) {
	// Save original config and restore after test
	origShowColumns := config.GetBool(config.KeyTreeShowColumns)
	origLastUpdated := config.GetBool(config.KeyTreeColumnsLastUpdated)
	origAssignee := config.GetBool(config.KeyTreeColumnsAssignee)
	origComments := config.GetBool(config.KeyTreeColumnsComments)
	defer func() {
		_ = config.Set(config.KeyTreeShowColumns, origShowColumns)
		_ = config.Set(config.KeyTreeColumnsLastUpdated, origLastUpdated)
		_ = config.Set(config.KeyTreeColumnsAssignee, origAssignee)
		_ = config.Set(config.KeyTreeColumnsComments, origComments)
	}()

	// Enable all columns for testing
	_ = config.Set(config.KeyTreeShowColumns, true)
	_ = config.Set(config.KeyTreeColumnsLastUpdated, true)
	_ = config.Set(config.KeyTreeColumnsAssignee, true)
	_ = config.Set(config.KeyTreeColumnsComments, true)

	// Column widths: lastUpdated=8, assignee=10, comments=5, separator=3 (total=26)
	// minTreeWidth=18
	// All 3:  minTreeWidth(18) + separator(3) + lastUpdated(8) + assignee(10) + comments(5) = 44
	// 2 cols: minTreeWidth(18) + separator(3) + lastUpdated(8) + assignee(10) = 39
	// 1 col:  minTreeWidth(18) + separator(3) + lastUpdated(8) = 29

	t.Run("wide_terminal_shows_all_columns", func(t *testing.T) {
		// 100 chars should easily fit all 3 columns
		state, treeWidth := prepareColumnState(100)
		if !state.enabled() {
			t.Fatal("expected columns to be enabled with wide terminal")
		}
		if len(state.columns) != 3 {
			t.Fatalf("expected 3 columns, got %d", len(state.columns))
		}
		if treeWidth < minTreeWidth {
			t.Fatalf("expected treeWidth >= %d, got %d", minTreeWidth, treeWidth)
		}
	})

	t.Run("medium_terminal_hides_rightmost_columns", func(t *testing.T) {
		// Width 32: fits lastUpdated (29 needed) but not assignee+comments
		state, treeWidth := prepareColumnState(32)
		if !state.enabled() {
			t.Fatal("expected columns to be enabled with medium terminal")
		}
		if len(state.columns) != 1 {
			t.Fatalf("expected 1 column (assignee+comments hidden), got %d", len(state.columns))
		}
		// Should have lastUpdated (leftmost, highest priority)
		if state.columns[0].ConfigKey != config.KeyTreeColumnsLastUpdated {
			t.Fatalf("expected lastUpdated column to remain, got %s", state.columns[0].ConfigKey)
		}
		if treeWidth < minTreeWidth {
			t.Fatalf("expected treeWidth >= %d, got %d", minTreeWidth, treeWidth)
		}
	})

	t.Run("medium_terminal_shows_two_columns", func(t *testing.T) {
		// Width 40: fits lastUpdated+assignee (39 needed) but not comments
		state, treeWidth := prepareColumnState(40)
		if !state.enabled() {
			t.Fatal("expected columns to be enabled")
		}
		if len(state.columns) != 2 {
			t.Fatalf("expected 2 columns (comments hidden), got %d", len(state.columns))
		}
		if state.columns[0].ConfigKey != config.KeyTreeColumnsLastUpdated {
			t.Fatalf("expected columns[0] = lastUpdated, got %s", state.columns[0].ConfigKey)
		}
		if state.columns[1].ConfigKey != config.KeyTreeColumnsAssignee {
			t.Fatalf("expected columns[1] = assignee, got %s", state.columns[1].ConfigKey)
		}
		if treeWidth < minTreeWidth {
			t.Fatalf("expected treeWidth >= %d, got %d", minTreeWidth, treeWidth)
		}
	})

	t.Run("narrow_terminal_hides_all_columns", func(t *testing.T) {
		// Test with width that can't even fit one column + minTreeWidth
		// minTreeWidth(18) + separator(3) + lastUpdated(8) = 29
		state, treeWidth := prepareColumnState(25)
		if state.enabled() {
			t.Fatalf("expected no columns with narrow terminal, got %d columns", len(state.columns))
		}
		if treeWidth != 25 {
			t.Fatalf("expected full width returned when no columns, got %d", treeWidth)
		}
	})

	t.Run("columns_disabled_returns_empty", func(t *testing.T) {
		_ = config.Set(config.KeyTreeShowColumns, false)
		state, treeWidth := prepareColumnState(100)
		if state.enabled() {
			t.Fatal("expected no columns when showColumns is false")
		}
		if treeWidth != 100 {
			t.Fatalf("expected full width returned, got %d", treeWidth)
		}
	})
}

func TestRenderCommentsColumn(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		expected string
	}{
		{name: "zero comments", count: 0, expected: ""},
		{name: "one comment no space", count: 1, expected: "💬1"},
		{name: "five comments no space", count: 5, expected: "💬5"},
		{name: "nine comments no space", count: 9, expected: "💬9"},
		{name: "ten comments", count: 10, expected: "💬10"},
		{name: "fifty comments", count: 50, expected: "💬50"},
		{name: "ninety nine comments", count: 99, expected: "💬99"},
		{name: "over ninety nine capped", count: 100, expected: "💬99+"},
		{name: "way over ninety nine", count: 500, expected: "💬99+"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := makeTestNodeWithComments(tt.count)
			got := renderCommentsColumn(node)
			if got != tt.expected {
				t.Errorf("renderCommentsColumn(%d comments) = %q, want %q", tt.count, got, tt.expected)
			}
		})
	}
}

func TestRenderCommentsColumn_NoNode(t *testing.T) {
	if got := renderCommentsColumn(nil); got != "" {
		t.Errorf("renderCommentsColumn(nil) = %q, want empty", got)
	}

	node := &graph.Node{CommentsLoaded: false}
	if got := renderCommentsColumn(node); got != "" {
		t.Errorf("renderCommentsColumn(not loaded) = %q, want empty", got)
	}
}

func TestRenderAssigneeColumn(t *testing.T) {
	tests := []struct {
		name     string
		assignee string
		expected string
	}{
		{name: "empty assignee", assignee: "", expected: ""},
		{name: "short name", assignee: "alice", expected: "alice"},
		{name: "exactly 10 chars", assignee: "1234567890", expected: "1234567890"},
		{name: "11 chars truncated", assignee: "12345678901", expected: "1234567..."},
		{name: "long name truncated", assignee: "Christopher Edwards", expected: "Christo..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &graph.Node{}
			node.Issue.Assignee = tt.assignee
			got := renderAssigneeColumn(node)
			if got != tt.expected {
				t.Errorf("renderAssigneeColumn(%q) = %q, want %q", tt.assignee, got, tt.expected)
			}
		})
	}
}

func TestRenderAssigneeColumn_NoNode(t *testing.T) {
	if got := renderAssigneeColumn(nil); got != "" {
		t.Errorf("renderAssigneeColumn(nil) = %q, want empty", got)
	}
}

func TestPrepareColumnState_AssigneeColumnOrder(t *testing.T) {
	origShowColumns := config.GetBool(config.KeyTreeShowColumns)
	origLastUpdated := config.GetBool(config.KeyTreeColumnsLastUpdated)
	origAssignee := config.GetBool(config.KeyTreeColumnsAssignee)
	origComments := config.GetBool(config.KeyTreeColumnsComments)
	defer func() {
		_ = config.Set(config.KeyTreeShowColumns, origShowColumns)
		_ = config.Set(config.KeyTreeColumnsLastUpdated, origLastUpdated)
		_ = config.Set(config.KeyTreeColumnsAssignee, origAssignee)
		_ = config.Set(config.KeyTreeColumnsComments, origComments)
	}()

	_ = config.Set(config.KeyTreeShowColumns, true)
	_ = config.Set(config.KeyTreeColumnsLastUpdated, true)
	_ = config.Set(config.KeyTreeColumnsAssignee, true)
	_ = config.Set(config.KeyTreeColumnsComments, true)

	// Wide terminal: all 3 columns should be present in order lastUpdated, assignee, comments
	state, _ := prepareColumnState(120)
	if len(state.columns) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(state.columns))
	}
	if state.columns[0].ConfigKey != config.KeyTreeColumnsLastUpdated {
		t.Errorf("expected columns[0] = lastUpdated, got %s", state.columns[0].ConfigKey)
	}
	if state.columns[1].ConfigKey != config.KeyTreeColumnsAssignee {
		t.Errorf("expected columns[1] = assignee, got %s", state.columns[1].ConfigKey)
	}
	if state.columns[2].ConfigKey != config.KeyTreeColumnsComments {
		t.Errorf("expected columns[2] = comments, got %s", state.columns[2].ConfigKey)
	}
}
