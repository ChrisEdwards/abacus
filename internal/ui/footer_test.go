package ui

import (
	"strings"
	"testing"
	"time"
)

func TestKeyPill(t *testing.T) {
	pill := keyPill("↑↓", "Navigate")

	t.Run("ContainsKey", func(t *testing.T) {
		if !strings.Contains(pill, "↑↓") {
			t.Error("expected pill to contain key")
		}
	})

	t.Run("ContainsDesc", func(t *testing.T) {
		if !strings.Contains(pill, "Navigate") {
			t.Error("expected pill to contain description")
		}
	})
}

func TestRenderFooter(t *testing.T) {
	m := &App{
		width:    160, // Wider to accommodate all hints including v/View and m/Comment
		repoName: "abacus",
		focus:    FocusTree,
	}

	footer := m.renderFooter()

	t.Run("ContainsNavigationKeys", func(t *testing.T) {
		if !strings.Contains(footer, "↑↓") {
			t.Error("expected footer to contain navigation arrows")
		}
		if !strings.Contains(footer, "Navigate") {
			t.Error("expected footer to contain 'Navigate'")
		}
	})

	t.Run("ContainsExpandKeys", func(t *testing.T) {
		if !strings.Contains(footer, "←→") {
			t.Error("expected footer to contain expand arrows")
		}
		if !strings.Contains(footer, "Expand") {
			t.Error("expected footer to contain 'Expand'")
		}
	})

	t.Run("ContainsGlobalKeys", func(t *testing.T) {
		globalKeys := []string{"/", "n", "⏎", "⇥", "s", "L", "q", "?"}
		for _, key := range globalKeys {
			if !strings.Contains(footer, key) {
				t.Errorf("expected footer to contain global key %q", key)
			}
		}
	})

	t.Run("NoRefreshStatusWhenEmpty", func(t *testing.T) {
		// With no refresh state, footer should have no right content
		if strings.Contains(footer, "Refreshing") || strings.Contains(footer, "error") {
			t.Error("expected footer to have no refresh status when app is idle")
		}
	})
}

func TestRenderFooterDetailsFocus(t *testing.T) {
	m := &App{
		width:       150, // Wider to accommodate 8 global + 1 context hint
		repoName:    "abacus",
		focus:       FocusDetails,
		ShowDetails: true,
	}

	footer := m.renderFooter()

	t.Run("ShowsScrollInsteadOfExpand", func(t *testing.T) {
		if !strings.Contains(footer, "Scroll") {
			t.Error("expected footer to contain 'Scroll' when in details focus")
		}
		// Should NOT contain "Expand" in details mode
		if strings.Contains(footer, "Expand") {
			t.Error("expected footer NOT to contain 'Expand' when in details focus")
		}
	})
}

func TestTrimHintsToFit(t *testing.T) {
	m := &App{width: 100}

	t.Run("PreservesHintsWhenSpaceAvailable", func(t *testing.T) {
		hints := []footerHint{
			{"↑↓", "Navigate"},
			{"/", "Search"},
		}
		result := m.trimHintsToFit(hints, 200)
		if len(result) != 2 {
			t.Errorf("expected 2 hints, got %d", len(result))
		}
	})

	t.Run("RemovesHintsWhenTooNarrow", func(t *testing.T) {
		// Create hints similar to full footer
		hints := []footerHint{
			{"↑↓", "Navigate"}, // context
			{"←→", "Expand"},   // context
			{"/", "Search"},    // global
			{"⏎", "Detail"},    // global
			{"⇥", "Focus"},     // global
			{"q", "Quit"},      // global
			{"?", "Help"},      // global
		}
		// Very narrow - should remove some hints
		result := m.trimHintsToFit(hints, 50)
		if len(result) >= len(hints) {
			t.Errorf("expected fewer hints when width is narrow, got %d", len(result))
		}
	})
}

func TestFooterNarrowTerminal(t *testing.T) {
	m := &App{
		width:    40, // Very narrow
		repoName: "abacus",
		focus:    FocusTree,
	}

	footer := m.renderFooter()

	// Should not panic and should produce output
	if footer == "" {
		t.Error("expected non-empty footer for narrow terminal")
	}

	// Should contain at least some key hints (search key "/" is commonly preserved)
	if !strings.Contains(footer, "/") && !strings.Contains(footer, "Search") {
		t.Errorf("expected footer to contain at least search key in narrow mode, got: %q", footer)
	}
}

func TestFooterHintSlices(t *testing.T) {
	t.Run("GlobalHintsCount", func(t *testing.T) {
		if len(globalFooterHints) != 10 {
			t.Errorf("expected 10 global hints, got %d", len(globalFooterHints))
		}
	})

	t.Run("TreeHintsCount", func(t *testing.T) {
		if len(treeFooterHints) != 2 {
			t.Errorf("expected 2 tree hints, got %d", len(treeFooterHints))
		}
	})

	t.Run("DetailsHintsCount", func(t *testing.T) {
		if len(detailsFooterHints) != 1 {
			t.Errorf("expected 1 details hint, got %d", len(detailsFooterHints))
		}
	})
}

func TestRenderRefreshStatus(t *testing.T) {
	t.Run("ErrorState", func(t *testing.T) {
		m := &App{lastError: "connection failed"}
		status := m.renderRefreshStatus()
		if !strings.Contains(status, "⚠") || !strings.Contains(status, "Error") {
			t.Errorf("expected error indicator, got: %q", status)
		}
	})

	t.Run("RefreshingState", func(t *testing.T) {
		m := &App{refreshInFlight: true}
		status := m.renderRefreshStatus()
		if !strings.Contains(status, "Refreshing") {
			t.Errorf("expected refreshing indicator, got: %q", status)
		}
	})

	t.Run("DeltaMetricsVisible", func(t *testing.T) {
		m := &App{
			lastRefreshStats: "+1 / Δ0 / -0",
			lastRefreshTime:  time.Now(),
		}
		status := m.renderRefreshStatus()
		if !strings.Contains(status, "Δ") || !strings.Contains(status, "+1") {
			t.Errorf("expected delta metrics, got: %q", status)
		}
	})

	t.Run("NoChangeHidden", func(t *testing.T) {
		m := &App{
			lastRefreshStats: "+0 / Δ0 / -0",
			lastRefreshTime:  time.Now(),
		}
		status := m.renderRefreshStatus()
		if status != "" {
			t.Errorf("expected empty status when no changes, got: %q", status)
		}
	})

	t.Run("DeltaMetricsExpired", func(t *testing.T) {
		m := &App{
			lastRefreshStats: "+1 / Δ0 / -0",
			lastRefreshTime:  time.Now().Add(-refreshDisplayDuration - time.Second),
		}
		status := m.renderRefreshStatus()
		if status != "" {
			t.Errorf("expected empty status after display duration, got: %q", status)
		}
	})

	t.Run("ErrorTakesPriority", func(t *testing.T) {
		m := &App{
			lastError:       "some error",
			refreshInFlight: true, // Also refreshing, but error should take priority
		}
		status := m.renderRefreshStatus()
		if !strings.Contains(status, "Error") {
			t.Errorf("expected error to take priority over refreshing, got: %q", status)
		}
	})
}

func TestRenderHintsWidth(t *testing.T) {
	hints := []footerHint{
		{"↑↓", "Navigate"},
	}

	width := renderHintsWidth(hints)

	if width <= 0 {
		t.Error("expected positive width for rendered hints")
	}

	// Adding more hints should increase width
	hints = append(hints, footerHint{"/", "Search"})
	newWidth := renderHintsWidth(hints)

	if newWidth <= width {
		t.Error("expected width to increase with more hints")
	}
}
