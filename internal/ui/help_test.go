package ui

import (
	"strings"
	"testing"
)

func TestRenderHelpOverlay(t *testing.T) {
	keys := DefaultKeyMap()
	overlay := renderHelpOverlay(keys)

	t.Run("ContainsTitle", func(t *testing.T) {
		if !strings.Contains(overlay, "ABACUS HELP") {
			t.Error("expected overlay to contain 'ABACUS HELP'")
		}
		if !strings.Contains(overlay, "✦") {
			t.Error("expected overlay to contain sparkle symbol")
		}
	})

	t.Run("ContainsAllSections", func(t *testing.T) {
		sections := []string{"NAVIGATION", "ACTIONS", "BEAD ACTIONS", "SEARCH"}
		for _, section := range sections {
			if !strings.Contains(overlay, section) {
				t.Errorf("expected overlay to contain section %q", section)
			}
		}
	})

	t.Run("ContainsKeyHintsFromKeyMap", func(t *testing.T) {
		// Key hints should be derived from KeyMap
		if !strings.Contains(overlay, keys.Up.Help().Key) {
			t.Errorf("expected overlay to contain Up key hint %q", keys.Up.Help().Key)
		}
		if !strings.Contains(overlay, keys.Enter.Help().Key) {
			t.Errorf("expected overlay to contain Enter key hint %q", keys.Enter.Help().Key)
		}
		if !strings.Contains(overlay, keys.Tab.Help().Key) {
			t.Errorf("expected overlay to contain Tab key hint %q", keys.Tab.Help().Key)
		}
	})

	t.Run("ContainsFooter", func(t *testing.T) {
		if !strings.Contains(overlay, "Press ? or Esc to close") {
			t.Error("expected overlay to contain footer instruction")
		}
	})
}

func TestGetHelpSections(t *testing.T) {
	keys := DefaultKeyMap()
	sections := getHelpSections(keys)

	t.Run("ReturnsFourSections", func(t *testing.T) {
		if len(sections) != 4 {
			t.Errorf("expected 4 sections, got %d", len(sections))
		}
	})

	t.Run("SectionTitles", func(t *testing.T) {
		expected := []string{"NAVIGATION", "ACTIONS", "BEAD ACTIONS", "SEARCH"}
		for i, section := range sections {
			if section.title != expected[i] {
				t.Errorf("section %d: expected title %q, got %q", i, expected[i], section.title)
			}
		}
	})

	t.Run("NavigationHas7Rows", func(t *testing.T) {
		if len(sections[0].rows) != 7 {
			t.Errorf("Navigation section: expected 7 rows, got %d", len(sections[0].rows))
		}
	})

	t.Run("ActionsHas5Rows", func(t *testing.T) {
		if len(sections[1].rows) != 5 {
			t.Errorf("Actions section: expected 5 rows, got %d", len(sections[1].rows))
		}
	})

	t.Run("BeadActionsHas8Rows", func(t *testing.T) {
		if len(sections[2].rows) != 8 {
			t.Errorf("Bead Actions section: expected 8 rows, got %d", len(sections[2].rows))
		}
	})

	t.Run("SearchHas3Rows", func(t *testing.T) {
		if len(sections[3].rows) != 3 {
			t.Errorf("Search section: expected 3 rows, got %d", len(sections[3].rows))
		}
	})

	t.Run("TextDerivedFromKeyMap", func(t *testing.T) {
		// First navigation row should be Up's help text
		if sections[0].rows[0][0] != keys.Up.Help().Key {
			t.Errorf("expected first navigation key to be %q, got %q",
				keys.Up.Help().Key, sections[0].rows[0][0])
		}
		if sections[0].rows[0][1] != keys.Up.Help().Desc {
			t.Errorf("expected first navigation desc to be %q, got %q",
				keys.Up.Help().Desc, sections[0].rows[0][1])
		}
	})
}

func TestRenderHelpSectionTable(t *testing.T) {
	section := helpSection{
		title: "TEST",
		rows: [][]string{
			{"key1", "desc1"},
			{"key2", "desc2"},
		},
	}

	rendered := renderHelpSectionTable(section)

	t.Run("ContainsTitle", func(t *testing.T) {
		if !strings.Contains(rendered, "TEST") {
			t.Error("expected rendered section to contain title")
		}
	})

	t.Run("ContainsUnderline", func(t *testing.T) {
		if !strings.Contains(rendered, "───") {
			t.Error("expected rendered section to contain underline")
		}
	})

	t.Run("ContainsKeys", func(t *testing.T) {
		if !strings.Contains(rendered, "key1") {
			t.Error("expected rendered section to contain 'key1'")
		}
		if !strings.Contains(rendered, "key2") {
			t.Error("expected rendered section to contain 'key2'")
		}
	})

	t.Run("ContainsDescriptions", func(t *testing.T) {
		if !strings.Contains(rendered, "desc1") {
			t.Error("expected rendered section to contain 'desc1'")
		}
		if !strings.Contains(rendered, "desc2") {
			t.Error("expected rendered section to contain 'desc2'")
		}
	})
}

func TestHelpOverlayDimensions(t *testing.T) {
	keys := DefaultKeyMap()

	t.Run("SmallTerminal", func(t *testing.T) {
		layer := newHelpOverlayLayer(keys, 60, 20, 1, 1)
		if layer == nil {
			t.Fatal("expected layer for small terminal")
		}
		canvas := layer.Render()
		if canvas == nil {
			t.Fatal("expected canvas for small terminal layer")
		}
		if output := canvas.Render(); output == "" {
			t.Error("expected non-empty overlay for small terminal")
		}
	})

	t.Run("LargeTerminal", func(t *testing.T) {
		layer := newHelpOverlayLayer(keys, 200, 60, 2, 2)
		if layer == nil {
			t.Fatal("expected layer for large terminal")
		}
		canvas := layer.Render()
		if canvas == nil {
			t.Fatal("expected canvas for large terminal layer")
		}
		if output := canvas.Render(); output == "" {
			t.Error("expected non-empty overlay for large terminal")
		}
	})
}
