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

func TestLabelsOverlay_EnterAddsChipFromDropdown(t *testing.T) {
	// Regression test for ab-mod2: Enter should select highlighted item and add chip
	overlay := NewLabelsOverlay("test-123", "Test", []string{}, []string{"build", "UI", "ui-redesign"})
	overlay.chipCombo.Focus()

	// Type "ui" - should highlight "UI" (exact match)
	overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	// Verify dropdown is open and not idle
	if overlay.IsIdle() {
		t.Fatal("expected overlay to not be idle after typing")
	}

	// Press Enter - should select the highlighted item
	overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Process returned commands to simulate Bubble Tea runtime
	if cmd != nil {
		msg := cmd()
		if batchMsg, ok := msg.(tea.BatchMsg); ok {
			for _, c := range batchMsg {
				if c != nil {
					innerMsg := c()
					overlay, _ = overlay.Update(innerMsg)
				}
			}
		} else {
			overlay, _ = overlay.Update(msg)
		}
	}

	// Should have added "UI" chip
	chips := overlay.GetChips()
	if len(chips) != 1 {
		t.Errorf("expected 1 chip after Enter, got %d: %v", len(chips), chips)
	}
	if len(chips) > 0 && chips[0] != "UI" {
		t.Errorf("expected chip 'UI', got '%s'", chips[0])
	}

	// Input should be cleared
	if overlay.chipCombo.InputValue() != "" {
		t.Errorf("expected empty input after selection, got '%s'", overlay.chipCombo.InputValue())
	}

	// Now should be idle (dropdown closed, input empty)
	if !overlay.IsIdle() {
		t.Error("expected overlay to be idle after selection")
	}
}

func TestLabelsOverlay_EnterSelectsBeforeConfirm(t *testing.T) {
	// Test that first Enter selects, second Enter (if idle) confirms
	// This ensures Enter doesn't prematurely confirm when selecting a label
	overlay := NewLabelsOverlay("test-123", "Test", []string{}, []string{"build", "UI", "ui-redesign"})
	overlay.chipCombo.Focus()

	// Type "ui" to filter
	overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	t.Logf("After typing 'ui': inputValue=%q, isDropdownOpen=%v, filtered=%v, highlightIdx=%d",
		overlay.chipCombo.InputValue(),
		overlay.chipCombo.IsDropdownOpen(),
		overlay.chipCombo.combo.FilteredOptions(),
		overlay.chipCombo.combo.HighlightIndex())

	// Verify ComboBox state before Enter
	if overlay.chipCombo.combo.State() != ComboBoxFiltering {
		t.Fatalf("expected ComboBox in Filtering state, got %v", overlay.chipCombo.combo.State())
	}

	// Verify highlighted option
	filtered := overlay.chipCombo.combo.FilteredOptions()
	highlightIdx := overlay.chipCombo.combo.HighlightIndex()
	if highlightIdx < 0 || highlightIdx >= len(filtered) {
		t.Fatalf("invalid highlightIndex %d for %d options", highlightIdx, len(filtered))
	}
	highlighted := filtered[highlightIdx]
	if highlighted != "UI" {
		t.Errorf("expected highlighted='UI', got '%s'", highlighted)
	}

	// Press Enter - this should:
	// 1. Select "UI" in the ComboBox (set value, transition to Idle)
	// 2. Return a cmd that produces ComboBoxEnterSelectedMsg
	// 3. NOT confirm the overlay (that would send LabelsUpdatedMsg)
	overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// ComboBox should now be Idle with "UI" as value
	if overlay.chipCombo.combo.State() != ComboBoxIdle {
		t.Errorf("expected ComboBox in Idle state after Enter, got %v", overlay.chipCombo.combo.State())
	}
	if overlay.chipCombo.combo.Value() != "UI" {
		t.Errorf("expected ComboBox value='UI' after Enter, got '%s'", overlay.chipCombo.combo.Value())
	}

	// Should NOT have chip yet (cmd not executed)
	if overlay.chipCombo.ChipCount() != 0 {
		t.Errorf("expected 0 chips before cmd execution, got %d", overlay.chipCombo.ChipCount())
	}

	// Verify cmd was returned
	if cmd == nil {
		t.Fatal("expected cmd to be returned from Enter")
	}

	// Execute cmd and dispatch messages (simulating Bubble Tea runtime)
	msg := cmd()
	foundValueSelected := false
	if batchMsg, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batchMsg {
			if c != nil {
				innerMsg := c()
				if vsm, ok := innerMsg.(ComboBoxEnterSelectedMsg); ok {
					foundValueSelected = true
					if vsm.Value != "UI" {
						t.Errorf("expected ComboBoxEnterSelectedMsg.Value='UI', got '%s'", vsm.Value)
					}
				}
				overlay, _ = overlay.Update(innerMsg)
			}
		}
	} else if vsm, ok := msg.(ComboBoxEnterSelectedMsg); ok {
		foundValueSelected = true
		if vsm.Value != "UI" {
			t.Errorf("expected ComboBoxEnterSelectedMsg.Value='UI', got '%s'", vsm.Value)
		}
		overlay, _ = overlay.Update(msg)
	}

	if !foundValueSelected {
		t.Error("expected ComboBoxEnterSelectedMsg to be in the cmd result")
	}

	// Now should have chip
	if overlay.chipCombo.ChipCount() != 1 {
		t.Errorf("expected 1 chip after cmd execution, got %d", overlay.chipCombo.ChipCount())
	}

	// Input should be cleared (handleSelection calls combo.SetValue(""))
	if overlay.chipCombo.InputValue() != "" {
		t.Errorf("expected empty input after cmd execution, got '%s'", overlay.chipCombo.InputValue())
	}

	// ComboBox value should be cleared
	if overlay.chipCombo.combo.Value() != "" {
		t.Errorf("expected ComboBox value cleared after chip added, got '%s'", overlay.chipCombo.combo.Value())
	}

	// Should be idle (ready for next input or confirm)
	if !overlay.IsIdle() {
		t.Error("expected overlay to be idle after selection completed")
	}
}

func TestLabelsOverlay_SimulateAppMessageFlow(t *testing.T) {
	// This test simulates how the App routes messages more accurately
	// The App receives KeyMsg, returns cmd, then receives ComboBoxEnterSelectedMsg separately
	overlay := NewLabelsOverlay("test-123", "Test", []string{}, []string{"build", "UI", "ui-redesign"})
	overlay.chipCombo.Focus()

	// Type "ui" to filter
	overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	t.Logf("Before Enter: chips=%v, inputValue=%q, comboValue=%q",
		overlay.GetChips(),
		overlay.chipCombo.InputValue(),
		overlay.chipCombo.combo.Value())

	// Press Enter - this returns cmd but doesn't execute it yet
	overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})

	t.Logf("After Enter (before cmd execution): chips=%v, inputValue=%q, comboValue=%q, isIdle=%v",
		overlay.GetChips(),
		overlay.chipCombo.InputValue(),
		overlay.chipCombo.combo.Value(),
		overlay.IsIdle())

	// Verify chip not added yet
	if len(overlay.GetChips()) != 0 {
		t.Errorf("expected 0 chips before cmd execution, got %d", len(overlay.GetChips()))
	}

	// Verify cmd was returned
	if cmd == nil {
		t.Fatal("expected cmd to be returned")
	}

	// Now simulate what Bubble Tea does:
	// 1. Execute the cmd
	// 2. Dispatch resulting messages to App.Update
	// 3. App.Update routes ComboBoxEnterSelectedMsg to overlay
	msg := cmd()
	t.Logf("cmd() returned type: %T", msg)

	// Process batch if returned
	processMsg := func(m tea.Msg) {
		switch m := m.(type) {
		case ComboBoxEnterSelectedMsg:
			t.Logf("Processing ComboBoxEnterSelectedMsg: Value=%q, IsNew=%v", m.Value, m.IsNew)
			// This is what App.Update does - route to overlay
			overlay, _ = overlay.Update(m)
		default:
			t.Logf("Skipping message type: %T", m)
		}
	}

	if batchMsg, ok := msg.(tea.BatchMsg); ok {
		t.Logf("Processing batch of %d commands", len(batchMsg))
		for i, c := range batchMsg {
			if c != nil {
				innerMsg := c()
				t.Logf("Batch[%d] returned type: %T", i, innerMsg)
				processMsg(innerMsg)
			}
		}
	} else {
		processMsg(msg)
	}

	t.Logf("After cmd execution: chips=%v, inputValue=%q, comboValue=%q, isIdle=%v",
		overlay.GetChips(),
		overlay.chipCombo.InputValue(),
		overlay.chipCombo.combo.Value(),
		overlay.IsIdle())

	// Now verify chip was added
	chips := overlay.GetChips()
	if len(chips) != 1 {
		t.Errorf("expected 1 chip after message processing, got %d: %v", len(chips), chips)
	}
	if len(chips) > 0 && chips[0] != "UI" {
		t.Errorf("expected chip 'UI', got '%s'", chips[0])
	}

	// Input should be cleared
	if overlay.chipCombo.InputValue() != "" {
		t.Errorf("expected empty input, got '%s'", overlay.chipCombo.InputValue())
	}
}

func TestLabelsOverlay_ChipVisibleAfterEnter(t *testing.T) {
	// Test that chip is properly visible through all accessor methods
	// This verifies no aliasing/copy issues are hiding the chip
	overlay := NewLabelsOverlay("test-123", "Test", []string{}, []string{"build", "UI", "api"})
	overlay.chipCombo.Focus()

	// Type and press Enter
	overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Process the cmd
	if cmd != nil {
		msg := cmd()
		if vsm, ok := msg.(ComboBoxEnterSelectedMsg); ok {
			overlay, _ = overlay.Update(vsm)
		}
	}

	// Verify through multiple paths
	// 1. Through overlay.GetChips()
	chipsFromOverlay := overlay.GetChips()
	if len(chipsFromOverlay) != 1 || chipsFromOverlay[0] != "UI" {
		t.Errorf("overlay.GetChips() = %v, want [UI]", chipsFromOverlay)
	}

	// 2. Through chipCombo.GetChips()
	chipsFromCombo := overlay.chipCombo.GetChips()
	if len(chipsFromCombo) != 1 || chipsFromCombo[0] != "UI" {
		t.Errorf("chipCombo.GetChips() = %v, want [UI]", chipsFromCombo)
	}

	// 3. Through chipCombo.chips.Chips directly
	chipsFromList := overlay.chipCombo.chips.Chips
	if len(chipsFromList) != 1 || chipsFromList[0] != "UI" {
		t.Errorf("chipCombo.chips.Chips = %v, want [UI]", chipsFromList)
	}

	// 4. Verify ChipCount
	if overlay.chipCombo.ChipCount() != 1 {
		t.Errorf("ChipCount() = %d, want 1", overlay.chipCombo.ChipCount())
	}

	// 5. Verify computeDiff returns the chip as added
	added, removed := overlay.computeDiff()
	if len(added) != 1 || added[0] != "UI" {
		t.Errorf("computeDiff() added=%v, want [UI]", added)
	}
	if len(removed) != 0 {
		t.Errorf("computeDiff() removed=%v, want []", removed)
	}
}

func TestLabelsOverlayTab(t *testing.T) {
	t.Run("TabDoesNotConfirmInLabelsOverlay", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", "Test", []string{"ui"}, []string{"ui", "bug"})
		// Add a chip
		overlay.chipCombo.SetChips([]string{"ui", "bug"})

		// Tab when idle should NOT confirm (unlike create modal)
		// It should pass through to ChipComboBox which sends ChipComboBoxTabMsg
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyTab})

		// Verify it doesn't send LabelsUpdatedMsg
		if cmd != nil {
			msg := cmd()
			if _, ok := msg.(LabelsUpdatedMsg); ok {
				t.Error("expected Tab NOT to send LabelsUpdatedMsg in labels overlay")
			}
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

func TestLabelsOverlay_TwoFastEnters(t *testing.T) {
	// Simulates user pressing Enter twice quickly:
	// 1. First Enter with highlighted item - should select
	// 2. Second Enter before cmd processed - what happens?
	overlay := NewLabelsOverlay("test-123", "Test", []string{}, []string{"build", "UI", "ui-redesign"})
	overlay.chipCombo.Focus()

	// Type "ui" to filter and highlight "UI"
	overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	t.Logf("Before first Enter: chips=%v, inputValue=%q, isIdle=%v",
		overlay.GetChips(), overlay.chipCombo.InputValue(), overlay.IsIdle())

	// First Enter - this should select "UI" and return a cmd
	overlay, cmd1 := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})

	t.Logf("After first Enter: chips=%v, inputValue=%q, isIdle=%v, cmd1!=nil=%v",
		overlay.GetChips(), overlay.chipCombo.InputValue(), overlay.IsIdle(), cmd1 != nil)

	// Second Enter BEFORE processing cmd1 - simulating fast typing
	overlay, cmd2 := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})

	t.Logf("After second Enter (before cmd1 processed): chips=%v, inputValue=%q, isIdle=%v, cmd2!=nil=%v",
		overlay.GetChips(), overlay.chipCombo.InputValue(), overlay.IsIdle(), cmd2 != nil)

	// Check if cmd2 is LabelsUpdatedMsg (form submit)
	if cmd2 != nil {
		msg := cmd2()
		if labelsMsg, ok := msg.(LabelsUpdatedMsg); ok {
			t.Logf("Second Enter submitted form! Added=%v, Removed=%v", labelsMsg.Added, labelsMsg.Removed)
			if len(labelsMsg.Added) == 0 {
				t.Error("BUG: Form submitted without the chip! This is the user-reported issue.")
			}
		}
	}

	// Now process cmd1 (the ComboBoxEnterSelectedMsg)
	if cmd1 != nil {
		msg := cmd1()
		t.Logf("cmd1 returned: %T", msg)
		if vsm, ok := msg.(ComboBoxEnterSelectedMsg); ok {
			overlay, _ = overlay.Update(vsm)
			t.Logf("After processing cmd1: chips=%v, inputValue=%q, isIdle=%v",
				overlay.GetChips(), overlay.chipCombo.InputValue(), overlay.IsIdle())
		}
	}

	// Third Enter - this should now have the chip
	overlay, cmd3 := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd3 != nil {
		msg := cmd3()
		if labelsMsg, ok := msg.(LabelsUpdatedMsg); ok {
			t.Logf("Third Enter: Added=%v, Removed=%v", labelsMsg.Added, labelsMsg.Removed)
			if len(labelsMsg.Added) != 1 || labelsMsg.Added[0] != "UI" {
				t.Errorf("Expected Added=['UI'], got %v", labelsMsg.Added)
			}
		}
	}
}

func TestLabelsOverlayChipComboBoxTabMsg(t *testing.T) {
	t.Run("TabMsgDoesNotConfirmInLabelsOverlay", func(t *testing.T) {
		overlay := NewLabelsOverlay("test-123", "Test", []string{"ui"}, []string{"ui", "bug"})
		overlay.chipCombo.SetChips([]string{"ui", "bug"})

		// ChipComboBoxTabMsg in labels overlay should NOT confirm (returns nil)
		// Unlike create modal where Tab moves to next field
		_, cmd := overlay.Update(ChipComboBoxTabMsg{})
		if cmd != nil {
			msg := cmd()
			if _, ok := msg.(LabelsUpdatedMsg); ok {
				t.Error("expected ChipComboBoxTabMsg NOT to confirm in labels overlay")
			}
		}
	})
}

func TestLabelsOverlay_Tab_AddsChipFromDropdown(t *testing.T) {
	overlay := NewLabelsOverlay("test-123", "Test", []string{}, []string{"backend", "frontend", "api"})
	overlay.chipCombo.Focus()

	t.Logf("Initial: Chips=%v, InputValue=%q, IsIdle=%v",
		overlay.GetChips(), overlay.chipCombo.InputValue(), overlay.IsIdle())

	// Type "back"
	for _, r := range "back" {
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	t.Logf("After 'back': Chips=%v, InputValue=%q, IsIdle=%v, IsDropdownOpen=%v",
		overlay.GetChips(), overlay.chipCombo.InputValue(), overlay.IsIdle(), overlay.chipCombo.IsDropdownOpen())

	// Press Tab - this should select the highlighted item
	overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyTab})

	t.Logf("After Tab (before cmd): Chips=%v, InputValue=%q",
		overlay.GetChips(), overlay.chipCombo.InputValue())

	// Execute the command - this should produce ComboBoxTabSelectedMsg
	if cmd == nil {
		t.Fatal("No command returned from Tab!")
	}

	msg := cmd()
	t.Logf("Tab cmd returned: %T = %+v", msg, msg)

	// The message needs to go back to the overlay (simulating App routing)
	overlay, cmd = overlay.Update(msg)
	t.Logf("After routing message back: Chips=%v, InputValue=%q",
		overlay.GetChips(), overlay.chipCombo.InputValue())

	// Process any follow-up commands
	if cmd != nil {
		msg2 := cmd()
		t.Logf("Second cmd returned: %T", msg2)
		if batchMsg, ok := msg2.(tea.BatchMsg); ok {
			for i, c := range batchMsg {
				if c != nil {
					innerMsg := c()
					t.Logf("  Batch[%d]: %T", i, innerMsg)
				}
			}
		}
	}

	t.Logf("Final: Chips=%v", overlay.GetChips())

	chips := overlay.GetChips()
	if len(chips) != 1 || chips[0] != "backend" {
		t.Errorf("Expected ['backend'], got %v", chips)
	}
}
