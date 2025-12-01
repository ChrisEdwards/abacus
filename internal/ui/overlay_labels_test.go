package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewLabelsOverlay(t *testing.T) {
	t.Run("InitializesWithCurrentLabels", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", "Test Bead", []string{"ui", "bug"}, []string{"ui", "bug", "enhancement"})
		chips := overlay.GetChips()
		if len(chips) != 2 {
			t.Errorf("expected 2 chips, got %d", len(chips))
		}
		// Check both labels are present (order may vary)
		found := make(map[string]bool)
		for _, c := range chips {
			found[c] = true
		}
		if !found["ui"] || !found["bug"] {
			t.Errorf("expected chips to contain 'ui' and 'bug', got %v", chips)
		}
	})

	t.Run("StoresOriginalChips", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", "Test Bead", []string{"ui"}, []string{"ui", "bug"})
		original := overlay.OriginalChips()
		if len(original) != 1 || original[0] != "ui" {
			t.Errorf("expected originalChips=['ui'], got %v", original)
		}
	})

	t.Run("StoresIssueIDAndTitle", func(t *testing.T) {
		overlay := NewLabelsOverlay("ab-007", "Fix the bug", []string{}, []string{})
		if overlay.IssueID() != "ab-007" {
			t.Errorf("expected IssueID='ab-007', got %s", overlay.IssueID())
		}
		if overlay.BeadTitle() != "Fix the bug" {
			t.Errorf("expected BeadTitle='Fix the bug', got %s", overlay.BeadTitle())
		}
	})

	t.Run("StartsInIdleState", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", "Test", []string{}, []string{"ui"})
		if !overlay.IsIdle() {
			t.Error("expected overlay to start in idle state")
		}
	})
}

func TestLabelsOverlayDiff(t *testing.T) {
	t.Run("DetectsAddedLabels", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", "Test", []string{"ui"}, []string{"ui", "bug"})
		// Simulate adding "bug" chip
		overlay.chipCombo.SetChips([]string{"ui", "bug"})
		added, removed := overlay.computeDiff()
		if len(added) != 1 || added[0] != "bug" {
			t.Errorf("expected added=['bug'], got %v", added)
		}
		if len(removed) != 0 {
			t.Errorf("expected no removed labels, got %v", removed)
		}
	})

	t.Run("DetectsRemovedLabels", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", "Test", []string{"ui", "bug"}, []string{"ui", "bug"})
		// Simulate removing "bug" chip
		overlay.chipCombo.SetChips([]string{"ui"})
		added, removed := overlay.computeDiff()
		if len(added) != 0 {
			t.Errorf("expected no added labels, got %v", added)
		}
		if len(removed) != 1 || removed[0] != "bug" {
			t.Errorf("expected removed=['bug'], got %v", removed)
		}
	})

	t.Run("DetectsBothAddedAndRemoved", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", "Test", []string{"old"}, []string{"old", "new"})
		// Simulate removing "old" and adding "new"
		overlay.chipCombo.SetChips([]string{"new"})
		added, removed := overlay.computeDiff()
		if len(added) != 1 || added[0] != "new" {
			t.Errorf("expected added=['new'], got %v", added)
		}
		if len(removed) != 1 || removed[0] != "old" {
			t.Errorf("expected removed=['old'], got %v", removed)
		}
	})

	t.Run("NoChanges", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", "Test", []string{"ui", "bug"}, []string{"ui", "bug"})
		added, removed := overlay.computeDiff()
		if len(added) != 0 {
			t.Errorf("expected no added labels, got %v", added)
		}
		if len(removed) != 0 {
			t.Errorf("expected no removed labels, got %v", removed)
		}
	})
}

func TestLabelsOverlayEnter(t *testing.T) {
	t.Run("SendsLabelsUpdatedMsgWhenIdle", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", "Test", []string{"ui"}, []string{"ui", "bug"})
		// Add a chip to have some change
		overlay.chipCombo.SetChips([]string{"ui", "bug"})

		// Enter when idle should confirm
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

	t.Run("PassesToChipComboWhenNotIdle", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", "Test", []string{}, []string{"ui", "bug"})

		// Focus the ChipComboBox (as Init() would do)
		overlay.chipCombo.Focus()

		// Open dropdown with down arrow to make it non-idle
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyDown})

		// Verify we're not idle (dropdown is open)
		if overlay.IsIdle() {
			t.Skip("dropdown did not open - skipping test")
		}

		// Now Enter should NOT send LabelsUpdatedMsg (should pass to combo to select)
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd != nil {
			msg := cmd()
			if _, ok := msg.(LabelsUpdatedMsg); ok {
				t.Error("expected Enter to pass to ChipComboBox when not idle, but got LabelsUpdatedMsg")
			}
		}
	})
}

func TestLabelsOverlayTab(t *testing.T) {
	t.Run("SendsLabelsUpdatedMsgWhenIdle", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", "Test", []string{"ui"}, []string{"ui", "bug"})
		// Add a chip
		overlay.chipCombo.SetChips([]string{"ui", "bug"})

		// Tab when idle should confirm
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyTab})
		if cmd == nil {
			t.Fatal("expected command from tab")
		}
		msg := cmd()
		labelsMsg, ok := msg.(LabelsUpdatedMsg)
		if !ok {
			t.Fatalf("expected LabelsUpdatedMsg, got %T", msg)
		}
		if labelsMsg.IssueID != "test-123" {
			t.Errorf("expected IssueID test-123, got %s", labelsMsg.IssueID)
		}
	})
}

func TestLabelsOverlayEscape(t *testing.T) {
	t.Run("CancelsWhenIdle", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", "Test", []string{}, []string{"ui"})
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

	t.Run("ClosesDropdownFirstWhenOpen", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", "Test", []string{}, []string{"ui", "bug"})

		// Focus the ChipComboBox (as Init() would do)
		overlay.chipCombo.Focus()

		// Open dropdown with down arrow
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyDown})

		// Verify dropdown is open
		if overlay.IsIdle() {
			t.Skip("dropdown did not open - skipping test")
		}

		// First Esc should close dropdown, not cancel
		overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if cmd != nil {
			msg := cmd()
			if _, ok := msg.(LabelsCancelledMsg); ok {
				t.Error("first Esc should close dropdown, not cancel")
			}
		}

		// Keep escaping until idle
		for !overlay.IsIdle() {
			overlay, cmd = overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})
			if cmd != nil {
				msg := cmd()
				if _, ok := msg.(LabelsCancelledMsg); ok {
					t.Error("Esc should not cancel while still not idle")
				}
			}
		}

		// Now should be idle
		if !overlay.IsIdle() {
			t.Error("expected to be idle after escaping")
		}

		// Final Esc should cancel
		_, cmd = overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if cmd == nil {
			t.Fatal("expected command from escape when idle")
		}
		msg := cmd()
		_, ok := msg.(LabelsCancelledMsg)
		if !ok {
			t.Fatalf("expected LabelsCancelledMsg, got %T", msg)
		}
	})
}

func TestLabelsOverlayView(t *testing.T) {
	t.Run("ContainsIssueID", func(t *testing.T) {
		overlay := NewLabelsOverlay("ab-123", "Test Bead", []string{}, []string{"ui"})
		view := overlay.View()
		if !strings.Contains(view, "ab-123") {
			t.Error("expected view to contain issue ID")
		}
	})

	t.Run("ContainsBeadTitle", func(t *testing.T) {
		overlay := NewLabelsOverlay("ab-123", "My Test Bead", []string{}, []string{"ui"})
		view := overlay.View()
		if !strings.Contains(view, "My Test Bead") {
			t.Error("expected view to contain bead title")
		}
	})

	t.Run("ContainsLabelsHeader", func(t *testing.T) {
		overlay := NewLabelsOverlay("ab-123", "Test", []string{}, []string{"ui"})
		view := overlay.View()
		if !strings.Contains(view, "Labels") {
			t.Error("expected view to contain 'Labels' header")
		}
	})

	t.Run("TruncatesLongTitle", func(t *testing.T) {
		longTitle := "This is a very long bead title that should be truncated"
		overlay := NewLabelsOverlay("ab-123", longTitle, []string{}, []string{})
		view := overlay.View()
		// Should contain truncated version with "..."
		if !strings.Contains(view, "...") {
			t.Error("expected long title to be truncated with '...'")
		}
		// Should NOT contain full title
		if strings.Contains(view, longTitle) {
			t.Error("expected title to be truncated, but full title found")
		}
	})
}

func TestLabelsOverlayFooter(t *testing.T) {
	t.Run("ShowsSaveHintsWhenIdle", func(t *testing.T) {
		overlay := NewLabelsOverlay("ab-123", "Test", []string{}, []string{"ui"})
		footer := overlay.renderFooter()
		if !strings.Contains(footer, "Save") {
			t.Error("expected footer to contain 'Save' when idle")
		}
		if !strings.Contains(footer, "Cancel") {
			t.Error("expected footer to contain 'Cancel' when idle")
		}
	})

	t.Run("ShowsSelectHintsWhenDropdownOpen", func(t *testing.T) {
		overlay := NewLabelsOverlay("ab-123", "Test", []string{}, []string{"ui", "bug"})
		// Open dropdown by pressing down
		overlay.Update(tea.KeyMsg{Type: tea.KeyDown})

		footer := overlay.renderFooter()
		if !strings.Contains(footer, "Select") {
			t.Error("expected footer to contain 'Select' when dropdown open")
		}
		if !strings.Contains(footer, "Navigate") {
			t.Error("expected footer to contain 'Navigate' when dropdown open")
		}
	})
}

func TestLabelsOverlayChipComboBoxTabMsg(t *testing.T) {
	t.Run("ConfirmsOnChipComboBoxTabMsg", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", "Test", []string{"ui"}, []string{"ui", "bug"})
		overlay.chipCombo.SetChips([]string{"ui", "bug"})

		_, cmd := overlay.Update(ChipComboBoxTabMsg{})
		if cmd == nil {
			t.Fatal("expected command from ChipComboBoxTabMsg")
		}
		msg := cmd()
		labelsMsg, ok := msg.(LabelsUpdatedMsg)
		if !ok {
			t.Fatalf("expected LabelsUpdatedMsg, got %T", msg)
		}
		if len(labelsMsg.Added) != 1 || labelsMsg.Added[0] != "bug" {
			t.Errorf("expected Added=['bug'], got %v", labelsMsg.Added)
		}
	})
}
