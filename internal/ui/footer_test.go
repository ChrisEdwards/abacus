package ui

import (
	"strings"
	"testing"
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
		width:    100,
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
		globalKeys := []string{"/", "⏎", "⇥", "q", "?"}
		for _, key := range globalKeys {
			if !strings.Contains(footer, key) {
				t.Errorf("expected footer to contain global key %q", key)
			}
		}
	})

	t.Run("ContainsRepoName", func(t *testing.T) {
		if !strings.Contains(footer, "Repo: abacus") {
			t.Error("expected footer to contain repo name")
		}
	})
}

func TestRenderFooterDetailsFocus(t *testing.T) {
	m := &App{
		width:       100,
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

	// Should always contain repo name
	if !strings.Contains(footer, "Repo:") {
		t.Error("expected footer to contain repo even in narrow mode")
	}
}

func TestFooterHintSlices(t *testing.T) {
	t.Run("GlobalHintsCount", func(t *testing.T) {
		if len(globalFooterHints) != 5 {
			t.Errorf("expected 5 global hints, got %d", len(globalFooterHints))
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
