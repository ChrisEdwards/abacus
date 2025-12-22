package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// ============================================================================
// Tab/Enter Behavior Tests (ab-inht bug fix)
// ============================================================================

func TestChipComboBox_TabSelectedMsg_AddsChipAndAdvances(t *testing.T) {
	cc := NewChipComboBox([]string{"UI", "Bug", "Feature"})
	cc.Focus()

	// Simulate receiving TabSelectedMsg
	cc, cmd := cc.Update(ComboBoxTabSelectedMsg{Value: "UI", IsNew: false})

	// Chip added
	if cc.ChipCount() != 1 {
		t.Errorf("expected 1 chip, got %d", cc.ChipCount())
	}
	chips := cc.GetChips()
	if len(chips) != 1 || chips[0] != "UI" {
		t.Errorf("expected chip 'UI', got %v", chips)
	}

	// Should return batch with ChipAddedMsg AND TabMsg
	if cmd == nil {
		t.Fatal("expected command to be returned")
	}
	msg := cmd()
	if batchMsg, ok := msg.(tea.BatchMsg); ok {
		foundChipAdded := false
		foundTabMsg := false
		for _, c := range batchMsg {
			if c != nil {
				innerMsg := c()
				if _, ok := innerMsg.(ChipComboBoxChipAddedMsg); ok {
					foundChipAdded = true
				}
				if _, ok := innerMsg.(ChipComboBoxTabMsg); ok {
					foundTabMsg = true
				}
			}
		}
		if !foundChipAdded {
			t.Error("expected ChipComboBoxChipAddedMsg in batch")
		}
		if !foundTabMsg {
			t.Error("expected ChipComboBoxTabMsg in batch (should advance)")
		}
	} else {
		t.Errorf("expected tea.BatchMsg, got %T", msg)
	}
}

func TestChipComboBox_EnterSelectedMsg_AddsChipStaysInField(t *testing.T) {
	cc := NewChipComboBox([]string{"UI", "Bug", "Feature"})
	cc.Focus()

	// Simulate receiving EnterSelectedMsg
	cc, cmd := cc.Update(ComboBoxEnterSelectedMsg{Value: "UI", IsNew: false})

	// Chip added
	if cc.ChipCount() != 1 {
		t.Errorf("expected 1 chip, got %d", cc.ChipCount())
	}

	// Should return ChipAddedMsg but NOT TabMsg
	if cmd == nil {
		t.Fatal("expected command to be returned")
	}
	msg := cmd()
	// Enter should NOT return a batch with TabMsg
	if _, ok := msg.(ChipComboBoxChipAddedMsg); !ok {
		// Check if it's a batch
		if batchMsg, ok := msg.(tea.BatchMsg); ok {
			for _, c := range batchMsg {
				if c != nil {
					innerMsg := c()
					if _, ok := innerMsg.(ChipComboBoxTabMsg); ok {
						t.Error("Enter should NOT return ChipComboBoxTabMsg - should stay in field")
					}
				}
			}
		}
	}
}

func TestChipComboBox_TabSelectedMsg_DuplicateStaysInField(t *testing.T) {
	cc := NewChipComboBox([]string{"UI", "Bug", "Feature"})
	cc.Focus()
	cc.chips.AddChip("UI") // Pre-add chip

	// Tab on duplicate
	cc, cmd := cc.Update(ComboBoxTabSelectedMsg{Value: "UI", IsNew: false})

	// Still only 1 chip
	if cc.ChipCount() != 1 {
		t.Errorf("expected 1 chip (no duplicate added), got %d", cc.ChipCount())
	}

	// Should return FlashCmd, NOT TabMsg (stay in field)
	if cmd == nil {
		t.Fatal("expected command")
	}
	msg := cmd()
	// Should be flash message, not batch with TabMsg
	if batchMsg, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batchMsg {
			if c != nil {
				innerMsg := c()
				if _, ok := innerMsg.(ChipComboBoxTabMsg); ok {
					t.Error("duplicate on Tab should NOT advance - should stay in field")
				}
			}
		}
	}
	// Flash index should be set
	if cc.FlashIndex() != 0 {
		t.Errorf("expected flashIndex 0, got %d", cc.FlashIndex())
	}
}

func TestChipComboBox_TabSelectedMsg_EmptyValueStillAdvances(t *testing.T) {
	cc := NewChipComboBox([]string{"UI", "Bug"})
	cc.Focus()

	// Tab with no selection (empty value)
	cc, cmd := cc.Update(ComboBoxTabSelectedMsg{Value: "", IsNew: false})

	// No chips added
	if cc.ChipCount() != 0 {
		t.Errorf("expected 0 chips, got %d", cc.ChipCount())
	}

	// But Tab signal should still be sent (advance anyway)
	if cmd == nil {
		t.Fatal("expected command")
	}
	msg := cmd()
	if _, ok := msg.(ChipComboBoxTabMsg); !ok {
		t.Errorf("expected ChipComboBoxTabMsg for empty selection, got %T", msg)
	}
}

func TestChipComboBox_TabSelectedMsg_NewValueAddedAndAdvances(t *testing.T) {
	cc := NewChipComboBox([]string{"UI", "Bug"})
	cc.Focus()

	// Tab with new value (not in options)
	cc, cmd := cc.Update(ComboBoxTabSelectedMsg{Value: "NewLabel", IsNew: true})

	// Chip added
	if cc.ChipCount() != 1 {
		t.Errorf("expected 1 chip, got %d", cc.ChipCount())
	}
	chips := cc.GetChips()
	if len(chips) != 1 || chips[0] != "NewLabel" {
		t.Errorf("expected chip 'NewLabel', got %v", chips)
	}

	// Should include both ChipAddedMsg with IsNew: true and TabMsg
	if cmd == nil {
		t.Fatal("expected command")
	}
	msg := cmd()
	if batchMsg, ok := msg.(tea.BatchMsg); ok {
		foundChipAdded := false
		foundTabMsg := false
		for _, c := range batchMsg {
			if c != nil {
				innerMsg := c()
				if addedMsg, ok := innerMsg.(ChipComboBoxChipAddedMsg); ok {
					foundChipAdded = true
					if !addedMsg.IsNew {
						t.Error("expected IsNew to be true for new value")
					}
				}
				if _, ok := innerMsg.(ChipComboBoxTabMsg); ok {
					foundTabMsg = true
				}
			}
		}
		if !foundChipAdded {
			t.Error("expected ChipComboBoxChipAddedMsg in batch")
		}
		if !foundTabMsg {
			t.Error("expected ChipComboBoxTabMsg in batch")
		}
	}
}

func TestChipComboBox_Tab_RawInputWhenDropdownClosed(t *testing.T) {
	cc := NewChipComboBox([]string{"UI", "Bug"})
	cc.Focus()

	// Manually set raw input without opening dropdown
	cc.combo.SetValue("Backend")

	// Press Tab directly (dropdown closed path)
	cc, cmd := cc.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Chip added from raw input
	if cc.ChipCount() != 1 {
		t.Errorf("expected 1 chip after Tab with raw input, got %d", cc.ChipCount())
	}
	chips := cc.GetChips()
	if len(chips) != 1 || chips[0] != "Backend" {
		t.Errorf("expected chip 'Backend', got %v", chips)
	}

	// Tab signal should be in batch
	if cmd == nil {
		t.Fatal("expected command")
	}
	msg := cmd()
	if batchMsg, ok := msg.(tea.BatchMsg); ok {
		foundTabMsg := false
		for _, c := range batchMsg {
			if c != nil {
				innerMsg := c()
				if _, ok := innerMsg.(ChipComboBoxTabMsg); ok {
					foundTabMsg = true
				}
			}
		}
		if !foundTabMsg {
			t.Error("expected ChipComboBoxTabMsg in batch")
		}
	}
}

func TestChipComboBox_Tab_DuplicateRawInputStaysInField(t *testing.T) {
	cc := NewChipComboBox([]string{"UI", "Bug"})
	cc.Focus()
	cc.chips.AddChip("Backend")

	// Manually set raw input to duplicate without opening dropdown
	cc.combo.SetValue("Backend")

	// Press Tab directly (dropdown closed path)
	cc, cmd := cc.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Still only 1 chip
	if cc.ChipCount() != 1 {
		t.Errorf("expected 1 chip (no duplicate), got %d", cc.ChipCount())
	}

	// Should NOT advance (flash instead)
	if cmd == nil {
		t.Fatal("expected command")
	}
	msg := cmd()
	if batchMsg, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batchMsg {
			if c != nil {
				innerMsg := c()
				if _, ok := innerMsg.(ChipComboBoxTabMsg); ok {
					t.Error("duplicate on raw Tab should NOT advance")
				}
			}
		}
	}
	// Flash should be set
	if cc.FlashIndex() != 0 {
		t.Errorf("expected flashIndex 0, got %d", cc.FlashIndex())
	}
}

func TestChipComboBox_Tab_AddsChipWhenDropdownOpen(t *testing.T) {
	cc := NewChipComboBox([]string{"backend", "frontend", "api"})
	cc.Focus()

	t.Logf("Initial: ChipCount=%d, InputValue=%q, IsDropdownOpen=%v",
		cc.ChipCount(), cc.InputValue(), cc.IsDropdownOpen())

	// Type "back"
	for _, r := range "back" {
		cc, _ = cc.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	t.Logf("After 'back': ChipCount=%d, InputValue=%q, IsDropdownOpen=%v, ComboState=%v",
		cc.ChipCount(), cc.InputValue(), cc.IsDropdownOpen(), cc.combo.State())

	// Press Tab
	cc, cmd := cc.Update(tea.KeyMsg{Type: tea.KeyTab})

	t.Logf("After Tab: ChipCount=%d, InputValue=%q, IsDropdownOpen=%v",
		cc.ChipCount(), cc.InputValue(), cc.IsDropdownOpen())

	// Execute the command
	if cmd != nil {
		msg := cmd()
		t.Logf("Command returned: %T = %+v", msg, msg)

		// If it's a batch, process all
		if batchMsg, ok := msg.(tea.BatchMsg); ok {
			t.Logf("BatchMsg with %d commands", len(batchMsg))
			for i, c := range batchMsg {
				if c != nil {
					innerMsg := c()
					t.Logf("  Batch[%d]: %T = %+v", i, innerMsg, innerMsg)
					cc, _ = cc.Update(innerMsg)
				}
			}
		} else {
			// Process single message
			cc, _ = cc.Update(msg)
		}
	} else {
		t.Error("No command returned!")
	}

	t.Logf("Final: ChipCount=%d, InputValue=%q, Chips=%v",
		cc.ChipCount(), cc.InputValue(), cc.GetChips())

	if cc.ChipCount() != 1 {
		t.Errorf("Expected 1 chip, got %d", cc.ChipCount())
	}
	chips := cc.GetChips()
	if len(chips) == 0 || chips[0] != "backend" {
		t.Errorf("Expected chip 'backend', got %v", chips)
	}
}
