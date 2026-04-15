package ui

import (
	"strings"
	"testing"
	"time"

	"abacus/internal/beads"
	"abacus/internal/graph"
)

// TestSizingInvariantTallMode verifies that at all terminal heights the shared
// budget invariant holds: (treeContent+2) + (detailContent+2) == listHeight+2.
// This matches the wide-mode mainBody height of listHeight+2, so the full view
// fills the terminal without a gap at the bottom.
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
		sharedBudget := listHeight - 2
		detailContent := int(float64(sharedBudget) * 0.6)
		detailContent = clampDimension(detailContent, minViewportHeight, sharedBudget-minListHeight)
		treeContent := sharedBudget - detailContent

		sum := (treeContent + 2) + (detailContent + 2)
		if sum != listHeight+2 {
			t.Errorf("height=%d: invariant failed: (treeContent+2)+(detailContent+2)=%d, want %d",
				h, sum, listHeight+2)
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
		sharedBudget := listHeight - 2
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

// --- handleLayoutKey tests ---

// TestHandleLayoutKeyTogglesWideToTall verifies pressing the layout key with details open
// and wide layout switches to tall layout and schedules a toast tick.
func TestHandleLayoutKeyTogglesWideToTall(t *testing.T) {
	m := &App{
		ShowDetails:   true,
		layout:        LayoutWide,
		activeOverlay: OverlayNone,
		keys:          DefaultKeyMap(),
	}
	model, cmd := m.handleLayoutKey()
	result := model.(*App)
	if result.layout != LayoutTall {
		t.Errorf("expected LayoutTall after toggle, got %v", result.layout)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd after layout toggle")
	}
}

// TestHandleLayoutKeyTogglesTallToWide verifies the reverse toggle direction.
func TestHandleLayoutKeyTogglesTallToWide(t *testing.T) {
	m := &App{
		ShowDetails:   true,
		layout:        LayoutTall,
		activeOverlay: OverlayNone,
		keys:          DefaultKeyMap(),
	}
	model, cmd := m.handleLayoutKey()
	result := model.(*App)
	if result.layout != LayoutWide {
		t.Errorf("expected LayoutWide after toggle, got %v", result.layout)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd after layout toggle")
	}
}

// TestHandleLayoutKeyNoopWhenDetailsClosed verifies no toggle occurs when the detail pane is hidden.
func TestHandleLayoutKeyNoopWhenDetailsClosed(t *testing.T) {
	m := &App{
		ShowDetails:   false,
		layout:        LayoutWide,
		activeOverlay: OverlayNone,
		keys:          DefaultKeyMap(),
	}
	model, cmd := m.handleLayoutKey()
	result := model.(*App)
	if result.layout != LayoutWide {
		t.Errorf("expected layout unchanged (LayoutWide), got %v", result.layout)
	}
	if result.layoutToastVisible {
		t.Error("expected no toast when detail pane is closed")
	}
	if cmd != nil {
		t.Errorf("expected nil cmd when no-op, got non-nil")
	}
}

// TestHandleLayoutKeyNoopWhenOverlayActive verifies no toggle occurs when an overlay is open.
func TestHandleLayoutKeyNoopWhenOverlayActive(t *testing.T) {
	m := &App{
		ShowDetails:   true,
		layout:        LayoutWide,
		activeOverlay: OverlayStatus,
		keys:          DefaultKeyMap(),
	}
	model, cmd := m.handleLayoutKey()
	result := model.(*App)
	if result.layout != LayoutWide {
		t.Errorf("expected layout unchanged (LayoutWide), got %v", result.layout)
	}
	if result.layoutToastVisible {
		t.Error("expected no toast when overlay is active")
	}
	if cmd != nil {
		t.Errorf("expected nil cmd when overlay active, got non-nil")
	}
}

// TestHandleLayoutKeySetToastState verifies toast state is set correctly after each toggle direction.
func TestHandleLayoutKeySetToastState(t *testing.T) {
	tests := []struct {
		name     string
		initial  Layout
		wantName string
	}{
		{"wide to tall sets Tall toast", LayoutWide, "Tall"},
		{"tall to wide sets Wide toast", LayoutTall, "Wide"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &App{
				ShowDetails:   true,
				layout:        tt.initial,
				activeOverlay: OverlayNone,
				keys:          DefaultKeyMap(),
			}
			model, _ := m.handleLayoutKey()
			result := model.(*App)
			if !result.layoutToastVisible {
				t.Error("expected layoutToastVisible=true after toggle")
			}
			if result.layoutToastName != tt.wantName {
				t.Errorf("layoutToastName=%q, want %q", result.layoutToastName, tt.wantName)
			}
		})
	}
}

// --- renderTreeView width tests ---

// TestRenderTreeViewTallModeUsesFullWidth verifies that in tall layout the tree uses the full
// terminal width (width-2) rather than subtracting the viewport width as in wide layout.
func TestRenderTreeViewTallModeUsesFullWidth(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:     "ab-01",
			Title:  strings.Repeat("x", 50),
			Status: "open",
		},
	}

	const termWidth = 80
	const vpWidth = 30

	m := &App{
		roots:       []*graph.Node{node},
		width:       termWidth,
		height:      24,
		ShowDetails: true,
	}
	m.viewport.Width = vpWidth
	m.recalcVisibleRows()

	// Tall mode: totalWidth = termWidth-2 = 78
	m.layout = LayoutTall
	renderedTall := m.renderTreeView()

	// Wide mode: totalWidth = termWidth-vpWidth-4 = 46
	m.layout = LayoutWide
	renderedWide := m.renderTreeView()

	// After stripping ANSI codes, tall-mode lines are padded to a wider totalWidth
	// than wide-mode lines. Comparing byte lengths is valid here: both have the same
	// unicode prefix characters, so extra bytes come from additional space padding.
	tallLines := strings.Split(stripANSI(renderedTall), "\n")
	wideLines := strings.Split(stripANSI(renderedWide), "\n")

	var tallLen, wideLen int
	for _, l := range tallLines {
		if l != "" {
			tallLen = len(l)
			break
		}
	}
	for _, l := range wideLines {
		if l != "" {
			wideLen = len(l)
			break
		}
	}

	if tallLen == 0 || wideLen == 0 {
		t.Fatal("no non-empty lines found in rendered tree output")
	}
	if tallLen <= wideLen {
		t.Errorf("tall mode should render wider lines than wide mode: tall=%d, wide=%d", tallLen, wideLen)
	}
}

// --- layoutToastTickMsg handler tests ---

// TestLayoutToastTickClearsAfterTimeout verifies the layout toast is hidden after its 3-second TTL.
func TestLayoutToastTickClearsAfterTimeout(t *testing.T) {
	m := &App{
		layoutToastVisible: true,
		layoutToastStart:   time.Now().Add(-4 * time.Second),
	}
	model, _, handled := m.handleOverlayMsg(layoutToastTickMsg{})
	if !handled {
		t.Fatal("layoutToastTickMsg should be handled")
	}
	result := model.(*App)
	if result.layoutToastVisible {
		t.Error("expected layoutToastVisible=false after timeout")
	}
}

// TestLayoutToastTickContinuesWhileVisible verifies the toast stays visible and reschedules a tick
// while still within its display window.
func TestLayoutToastTickContinuesWhileVisible(t *testing.T) {
	m := &App{
		layoutToastVisible: true,
		layoutToastStart:   time.Now(),
	}
	model, cmd, handled := m.handleOverlayMsg(layoutToastTickMsg{})
	if !handled {
		t.Fatal("layoutToastTickMsg should be handled")
	}
	result := model.(*App)
	if !result.layoutToastVisible {
		t.Error("expected layoutToastVisible=true while within display window")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd to reschedule tick")
	}
}
