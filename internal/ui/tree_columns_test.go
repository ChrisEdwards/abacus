package ui

import (
	"testing"

	"abacus/internal/config"
)

func TestPrepareColumnState_ResponsiveHiding(t *testing.T) {
	// Save original config and restore after test
	origShowColumns := config.GetBool(config.KeyTreeShowColumns)
	origLastUpdated := config.GetBool(config.KeyTreeColumnsLastUpdated)
	origComments := config.GetBool(config.KeyTreeColumnsComments)
	defer func() {
		_ = config.Set(config.KeyTreeShowColumns, origShowColumns)
		_ = config.Set(config.KeyTreeColumnsLastUpdated, origLastUpdated)
		_ = config.Set(config.KeyTreeColumnsComments, origComments)
	}()

	// Enable all columns for testing
	_ = config.Set(config.KeyTreeShowColumns, true)
	_ = config.Set(config.KeyTreeColumnsLastUpdated, true)
	_ = config.Set(config.KeyTreeColumnsComments, true)

	// Column widths: lastUpdated=8, comments=5, separator=3 (total=16)
	// minTreeWidth=18

	t.Run("wide_terminal_shows_all_columns", func(t *testing.T) {
		// 100 chars should easily fit all columns
		state, treeWidth := prepareColumnState(100)
		if !state.enabled() {
			t.Fatal("expected columns to be enabled with wide terminal")
		}
		if len(state.columns) != 2 {
			t.Fatalf("expected 2 columns, got %d", len(state.columns))
		}
		if treeWidth < minTreeWidth {
			t.Fatalf("expected treeWidth >= %d, got %d", minTreeWidth, treeWidth)
		}
	})

	t.Run("medium_terminal_hides_rightmost_column", func(t *testing.T) {
		// Test with width that can fit tree+lastUpdated but not comments
		// minTreeWidth(18) + separator(3) + lastUpdated(8) = 29
		// minTreeWidth(18) + separator(3) + lastUpdated(8) + comments(5) = 34
		state, treeWidth := prepareColumnState(32)
		if !state.enabled() {
			t.Fatal("expected columns to be enabled with medium terminal")
		}
		if len(state.columns) != 1 {
			t.Fatalf("expected 1 column (comments hidden), got %d", len(state.columns))
		}
		// Should have lastUpdated (leftmost, higher priority)
		if state.columns[0].ConfigKey != config.KeyTreeColumnsLastUpdated {
			t.Fatalf("expected lastUpdated column to remain, got %s", state.columns[0].ConfigKey)
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
