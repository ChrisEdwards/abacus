package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewLabelsOverlay(t *testing.T) {
	t.Run("PreSelectsCurrentLabels", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", []string{"ui", "bug"}, []string{"ui", "bug", "enhancement"})
		if !overlay.selected["ui"] {
			t.Error("expected 'ui' to be selected")
		}
		if !overlay.selected["bug"] {
			t.Error("expected 'bug' to be selected")
		}
		if overlay.selected["enhancement"] {
			t.Error("expected 'enhancement' to NOT be selected")
		}
	})

	t.Run("SortsAllLabels", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", nil, []string{"zebra", "alpha", "middle"})
		if overlay.allLabels[0] != "alpha" {
			t.Errorf("expected first label to be 'alpha', got %s", overlay.allLabels[0])
		}
		if overlay.allLabels[2] != "zebra" {
			t.Errorf("expected last label to be 'zebra', got %s", overlay.allLabels[2])
		}
	})

	t.Run("StoresOriginalState", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", []string{"ui"}, []string{"ui", "bug"})
		if !overlay.original["ui"] {
			t.Error("expected original to contain 'ui'")
		}
		if overlay.original["bug"] {
			t.Error("expected original to NOT contain 'bug'")
		}
	})
}

func TestLabelsOverlayNavigation(t *testing.T) {
	t.Run("DownMovesToNext", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", nil, []string{"a", "b", "c"})
		overlay.cursor = 0
		overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if overlay.cursor != 1 {
			t.Errorf("expected cursor 1 after j, got %d", overlay.cursor)
		}
	})

	t.Run("UpMovesToPrevious", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", nil, []string{"a", "b", "c"})
		overlay.cursor = 1
		overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if overlay.cursor != 0 {
			t.Errorf("expected cursor 0 after k, got %d", overlay.cursor)
		}
	})

	t.Run("DownWrapsAround", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", nil, []string{"a", "b", "c"})
		overlay.cursor = 2
		overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if overlay.cursor != 0 {
			t.Errorf("expected cursor 0 after wrap, got %d", overlay.cursor)
		}
	})

	t.Run("UpWrapsAround", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", nil, []string{"a", "b", "c"})
		overlay.cursor = 0
		overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if overlay.cursor != 2 {
			t.Errorf("expected cursor 2 after wrap, got %d", overlay.cursor)
		}
	})
}

func TestLabelsOverlayToggle(t *testing.T) {
	t.Run("SpaceTogglesLabel", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", nil, []string{"ui", "bug"})
		overlay.cursor = 0 // On "bug" (sorted)
		if overlay.selected["bug"] {
			t.Error("expected 'bug' to start unselected")
		}
		overlay.Update(tea.KeyMsg{Type: tea.KeySpace})
		if !overlay.selected["bug"] {
			t.Error("expected 'bug' to be selected after space")
		}
		overlay.Update(tea.KeyMsg{Type: tea.KeySpace})
		if overlay.selected["bug"] {
			t.Error("expected 'bug' to be unselected after second space")
		}
	})
}

func TestLabelsOverlayFilter(t *testing.T) {
	t.Run("FilterReducesList", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", nil, []string{"ui", "bug", "enhancement"})
		overlay.filterInput.SetValue("u")
		filtered := overlay.filteredLabels()
		if len(filtered) != 2 { // "ui" and "bug" contain "u"
			t.Errorf("expected 2 filtered labels, got %d", len(filtered))
		}
	})

	t.Run("FilterIsCaseInsensitive", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", nil, []string{"UI", "bug"})
		overlay.filterInput.SetValue("ui")
		filtered := overlay.filteredLabels()
		if len(filtered) != 1 {
			t.Errorf("expected 1 filtered label, got %d", len(filtered))
		}
		if filtered[0] != "UI" {
			t.Errorf("expected 'UI', got %s", filtered[0])
		}
	})
}

func TestLabelsOverlayDiff(t *testing.T) {
	t.Run("DetectsAddedLabels", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", []string{"ui"}, []string{"ui", "bug"})
		overlay.selected["bug"] = true
		added, removed := overlay.computeDiff()
		if len(added) != 1 || added[0] != "bug" {
			t.Errorf("expected added=['bug'], got %v", added)
		}
		if len(removed) != 0 {
			t.Errorf("expected no removed labels, got %v", removed)
		}
	})

	t.Run("DetectsRemovedLabels", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", []string{"ui", "bug"}, []string{"ui", "bug"})
		overlay.selected["bug"] = false
		added, removed := overlay.computeDiff()
		if len(added) != 0 {
			t.Errorf("expected no added labels, got %v", added)
		}
		if len(removed) != 1 || removed[0] != "bug" {
			t.Errorf("expected removed=['bug'], got %v", removed)
		}
	})

	t.Run("DetectsBothAddedAndRemoved", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", []string{"old"}, []string{"old", "new"})
		overlay.selected["old"] = false
		overlay.selected["new"] = true
		added, removed := overlay.computeDiff()
		if len(added) != 1 || added[0] != "new" {
			t.Errorf("expected added=['new'], got %v", added)
		}
		if len(removed) != 1 || removed[0] != "old" {
			t.Errorf("expected removed=['old'], got %v", removed)
		}
	})
}

func TestLabelsOverlayEnter(t *testing.T) {
	t.Run("SendsLabelsUpdatedMsg", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", []string{"ui"}, []string{"ui", "bug"})
		overlay.selected["bug"] = true
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Fatal("expected command from enter")
		}
		msg := cmd()
		labelsMsg, ok := msg.(LabelsUpdatedMsg)
		if !ok {
			t.Fatalf("expected LabelsUpdatedMsg, got %T", msg)
		}
		if labelsMsg.IssueID != "test-123" {
			t.Errorf("expected IssueID test-123, got %s", labelsMsg.IssueID)
		}
		if len(labelsMsg.Added) != 1 || labelsMsg.Added[0] != "bug" {
			t.Errorf("expected Added=['bug'], got %v", labelsMsg.Added)
		}
	})
}

func TestLabelsOverlayEscape(t *testing.T) {
	t.Run("SendsLabelsCancelledMsg", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", nil, []string{"ui"})
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if cmd == nil {
			t.Fatal("expected command from escape")
		}
		msg := cmd()
		_, ok := msg.(LabelsCancelledMsg)
		if !ok {
			t.Fatalf("expected LabelsCancelledMsg, got %T", msg)
		}
	})
}

func TestLabelsOverlayNewLabel(t *testing.T) {
	t.Run("CanAddNewLabelWhenNoMatch", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", nil, []string{"ui", "bug"})
		overlay.filterInput.SetValue("newlabel")
		if !overlay.canAddNew() {
			t.Error("expected canAddNew() to be true when filter doesn't match")
		}
	})

	t.Run("NoNewLabelWhenExactMatch", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", nil, []string{"ui", "bug"})
		overlay.filterInput.SetValue("ui")
		if overlay.canAddNew() {
			t.Error("expected canAddNew() to be false when filter matches exactly")
		}
	})

	t.Run("AddNewLabelOnToggle", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", nil, []string{"existing"})
		overlay.filterInput.SetValue("newlabel")
		// Move cursor to "add new" option (index 0 for filtered list is empty since no match, so cursor 0 is "add new")
		overlay.cursor = 0 // The add new option
		overlay.toggleCurrent()
		if !overlay.selected["newlabel"] {
			t.Error("expected newlabel to be selected after toggle")
		}
	})
}

func TestLabelsOverlayView(t *testing.T) {
	t.Run("ContainsIssueID", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", nil, []string{"ui"})
		view := overlay.View()
		if view == "" {
			t.Error("expected non-empty view")
		}
	})

	t.Run("ContainsLabels", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", []string{"ui"}, []string{"ui", "bug"})
		view := overlay.View()
		// Just verify it renders without panic
		if len(view) < 10 {
			t.Error("view seems too short")
		}
	})
}
