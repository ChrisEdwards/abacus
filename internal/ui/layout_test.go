package ui

import (
	"testing"
)

// TestSizingInvariantTallMode verifies that at all terminal heights the shared
// budget invariant holds: (treeContent+2) + (detailContent+2) == listHeight.
func TestSizingInvariantTallMode(t *testing.T) {
	heights := []int{24, 15, 8}
	for _, h := range heights {
		m := &App{
			height:      h,
			width:       80,
			ShowDetails: true,
			layout:      LayoutTall,
		}
		// Compute values using same formulas as treePaneHeight / recalcViewportSize
		listHeight := clampDimension(h-4, minListHeight, h-2)
		sharedBudget := listHeight - 4
		detailContent := int(float64(sharedBudget) * 0.6)
		detailContent = clampDimension(detailContent, minViewportHeight, sharedBudget-minListHeight)
		treeContent := sharedBudget - detailContent

		sum := (treeContent + 2) + (detailContent + 2)
		if sum != listHeight {
			t.Errorf("height=%d: invariant failed: (treeContent+2)+(detailContent+2)=%d, want %d",
				h, sum, listHeight)
		}

		// Also verify treePaneHeight returns treeContent
		if got := m.treePaneHeight(); got != treeContent {
			t.Errorf("height=%d: treePaneHeight()=%d, want %d", h, got, treeContent)
		}
	}
}

// TestToggleCycle verifies Wide→Tall→Wide restores original viewport dimensions.
func TestToggleCycle(t *testing.T) {
	m := &App{
		height:      24,
		width:       120,
		ShowDetails: true,
		layout:      LayoutWide,
	}
	m.recalcViewportSize()
	origW := m.viewport.Width
	origH := m.viewport.Height

	// Toggle to Tall
	m.layout = LayoutTall
	m.recalcViewportSize()
	if m.viewport.Width == origW && m.viewport.Height == origH {
		t.Error("expected viewport to change after switching to LayoutTall")
	}

	// Toggle back to Wide
	m.layout = LayoutWide
	m.recalcViewportSize()
	if m.viewport.Width != origW {
		t.Errorf("after toggle cycle: viewport.Width=%d, want %d", m.viewport.Width, origW)
	}
	if m.viewport.Height != origH {
		t.Errorf("after toggle cycle: viewport.Height=%d, want %d", m.viewport.Height, origH)
	}
}

// TestResizeInTallMode ensures that after a resize in tall mode neither
// pane height is zero and the sizes are consistent.
func TestResizeInTallMode(t *testing.T) {
	sizes := [][2]int{{80, 24}, {40, 12}, {120, 50}}
	for _, s := range sizes {
		w, h := s[0], s[1]
		m := &App{
			width:       w,
			height:      h,
			ShowDetails: true,
			layout:      LayoutTall,
		}
		m.recalcViewportSize()

		if m.viewport.Width <= 0 {
			t.Errorf("w=%d h=%d: viewport.Width=%d, must be > 0", w, h, m.viewport.Width)
		}
		if m.viewport.Height <= 0 {
			t.Errorf("w=%d h=%d: viewport.Height=%d, must be > 0", w, h, m.viewport.Height)
		}
		treeH := m.treePaneHeight()
		if treeH <= 0 {
			t.Errorf("w=%d h=%d: treePaneHeight=%d, must be > 0", w, h, treeH)
		}
	}
}

// TestTreePaneHeightWideMode checks that in wide mode treePaneHeight returns
// the same value as the legacy inline calculation.
func TestTreePaneHeightWideMode(t *testing.T) {
	heights := []int{24, 15, 8}
	for _, h := range heights {
		m := &App{height: h, width: 80, ShowDetails: true, layout: LayoutWide}
		want := clampDimension(h-4, minListHeight, h-2)
		got := m.treePaneHeight()
		if got != want {
			t.Errorf("height=%d: treePaneHeight()=%d, want %d", h, got, want)
		}
	}
}

// TestTreePaneHeightTallMode checks heights are computed correctly in tall mode.
func TestTreePaneHeightTallMode(t *testing.T) {
	heights := []int{24, 15, 8}
	for _, h := range heights {
		m := &App{height: h, width: 80, ShowDetails: true, layout: LayoutTall}

		listHeight := clampDimension(h-4, minListHeight, h-2)
		sharedBudget := listHeight - 4
		detailContent := int(float64(sharedBudget) * 0.6)
		detailContent = clampDimension(detailContent, minViewportHeight, sharedBudget-minListHeight)
		want := sharedBudget - detailContent

		got := m.treePaneHeight()
		if got != want {
			t.Errorf("height=%d: treePaneHeight()=%d, want %d", h, got, want)
		}
	}
}

// TestLayoutNoopWhenDetailsClosed ensures treePaneHeight is unchanged when detail pane is closed.
func TestLayoutNoopWhenDetailsClosed(t *testing.T) {
	m := &App{height: 24, width: 80, ShowDetails: false, layout: LayoutTall}
	want := clampDimension(m.height-4, minListHeight, m.height-2)
	if got := m.treePaneHeight(); got != want {
		t.Errorf("treePaneHeight with ShowDetails=false: got %d, want %d", got, want)
	}
}
