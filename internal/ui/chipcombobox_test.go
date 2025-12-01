package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewChipComboBox(t *testing.T) {
	t.Run("DefaultValues", func(t *testing.T) {
		options := []string{"backend", "frontend", "api"}
		cc := NewChipComboBox(options)

		if cc.Width != 50 {
			t.Errorf("expected default width 50, got %d", cc.Width)
		}
		if cc.ChipCount() != 0 {
			t.Errorf("expected 0 chips, got %d", cc.ChipCount())
		}
		if cc.InChipNavMode() {
			t.Error("expected not in chip nav mode initially")
		}
		if cc.IsDropdownOpen() {
			t.Error("expected dropdown closed initially")
		}
		if cc.Focused() {
			t.Error("expected not focused initially")
		}
	})

	t.Run("OptionsAreCopied", func(t *testing.T) {
		options := []string{"a", "b", "c"}
		cc := NewChipComboBox(options)
		options[0] = "modified"

		// Internal options should not be affected
		if len(cc.allOptions) != 3 || cc.allOptions[0] == "modified" {
			t.Error("expected options to be copied")
		}
	})
}

func TestChipComboBoxBuilders(t *testing.T) {
	t.Run("WithWidth", func(t *testing.T) {
		cc := NewChipComboBox(nil).WithWidth(80)
		if cc.Width != 80 {
			t.Errorf("expected width 80, got %d", cc.Width)
		}
	})

	t.Run("WithMaxVisible", func(t *testing.T) {
		cc := NewChipComboBox(nil).WithMaxVisible(10)
		if cc.combo.MaxVisible != 10 {
			t.Errorf("expected MaxVisible 10, got %d", cc.combo.MaxVisible)
		}
	})

	t.Run("WithPlaceholder", func(t *testing.T) {
		cc := NewChipComboBox(nil).WithPlaceholder("Select...")
		if cc.combo.Placeholder != "Select..." {
			t.Errorf("expected placeholder 'Select...', got '%s'", cc.combo.Placeholder)
		}
	})

	t.Run("WithAllowNew", func(t *testing.T) {
		cc := NewChipComboBox(nil).WithAllowNew(false, "")
		if cc.combo.AllowNew {
			t.Error("expected AllowNew to be false")
		}
	})
}

func TestChipComboBox_UpArrow_EntersChipNav(t *testing.T) {
	options := []string{"backend", "frontend"}
	cc := NewChipComboBox(options)
	cc.Focus()

	// Add some chips first
	cc.chips.AddChip("backend")
	cc.chips.AddChip("frontend")

	// Press up arrow with empty input and dropdown closed (chips are above input)
	cc, _ = cc.Update(tea.KeyMsg{Type: tea.KeyUp})

	if !cc.InChipNavMode() {
		t.Error("expected to enter chip nav mode")
	}
	// Should highlight last chip
	if cc.chips.HighlightedChip() != "frontend" {
		t.Errorf("expected 'frontend' highlighted, got '%s'", cc.chips.HighlightedChip())
	}
}

func TestChipComboBox_UpArrow_IgnoredWhenNotEmpty(t *testing.T) {
	cc := NewChipComboBox([]string{"a", "b"})
	cc.Focus()
	cc.chips.AddChip("a")

	// Type something first
	cc, _ = cc.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	// Now press up - should not enter chip nav (input not empty)
	cc, _ = cc.Update(tea.KeyMsg{Type: tea.KeyUp})

	if cc.InChipNavMode() {
		t.Error("should not enter chip nav when input is not empty")
	}
}

func TestChipComboBox_UpArrow_IgnoredWhenDropdownOpen(t *testing.T) {
	cc := NewChipComboBox([]string{"backend", "frontend"})
	cc.Focus()
	cc.chips.AddChip("backend")

	// Open dropdown
	cc, _ = cc.Update(tea.KeyMsg{Type: tea.KeyDown})

	if !cc.IsDropdownOpen() {
		t.Fatal("expected dropdown to be open")
	}

	// Press up - should navigate dropdown, not enter chip nav
	cc, _ = cc.Update(tea.KeyMsg{Type: tea.KeyUp})

	if cc.InChipNavMode() {
		t.Error("should not enter chip nav when dropdown is open")
	}
}

func TestChipComboBox_UpArrow_IgnoredWhenNoChips(t *testing.T) {
	cc := NewChipComboBox([]string{"a", "b"})
	cc.Focus()

	// Press up with no chips
	cc, _ = cc.Update(tea.KeyMsg{Type: tea.KeyUp})

	if cc.InChipNavMode() {
		t.Error("should not enter chip nav when no chips exist")
	}
}

func TestChipComboBox_Selection_AddsChip(t *testing.T) {
	cc := NewChipComboBox([]string{"backend", "frontend", "api"})
	cc.Focus()

	// Simulate selection
	cc, cmd := cc.Update(ComboBoxValueSelectedMsg{Value: "backend", IsNew: false})

	if cc.ChipCount() != 1 {
		t.Errorf("expected 1 chip, got %d", cc.ChipCount())
	}
	chips := cc.GetChips()
	if len(chips) != 1 || chips[0] != "backend" {
		t.Errorf("expected chip 'backend', got %v", chips)
	}

	// Should emit ChipAddedMsg
	if cmd == nil {
		t.Fatal("expected command")
	}
	msg := cmd()
	addedMsg, ok := msg.(ChipComboBoxChipAddedMsg)
	if !ok {
		t.Fatalf("expected ChipComboBoxChipAddedMsg, got %T", msg)
	}
	if addedMsg.Label != "backend" {
		t.Errorf("expected label 'backend', got '%s'", addedMsg.Label)
	}
}

func TestChipComboBox_Selection_ClearsInput(t *testing.T) {
	cc := NewChipComboBox([]string{"backend"})
	cc.Focus()

	// Type something
	cc, _ = cc.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})

	// Select
	cc, _ = cc.Update(ComboBoxValueSelectedMsg{Value: "backend", IsNew: false})

	if cc.InputValue() != "" {
		t.Errorf("expected empty input after selection, got '%s'", cc.InputValue())
	}
}

func TestChipComboBox_Selection_FiltersOptions(t *testing.T) {
	cc := NewChipComboBox([]string{"backend", "frontend", "api"})
	cc.Focus()

	// Select backend
	cc, _ = cc.Update(ComboBoxValueSelectedMsg{Value: "backend", IsNew: false})

	// Open dropdown
	cc, _ = cc.Update(tea.KeyMsg{Type: tea.KeyDown})

	// backend should be filtered out
	filtered := cc.combo.FilteredOptions()
	for _, opt := range filtered {
		if opt == "backend" {
			t.Error("expected 'backend' to be filtered out of dropdown")
		}
	}
}

func TestChipComboBox_Selection_StaysInField(t *testing.T) {
	cc := NewChipComboBox([]string{"backend"})
	cc.Focus()

	cc, _ = cc.Update(ComboBoxValueSelectedMsg{Value: "backend", IsNew: false})

	// Should still be focused (multi-select stays in field)
	if !cc.Focused() {
		t.Error("expected to stay focused after selection")
	}
}

func TestChipComboBox_Duplicate_FlashesChip(t *testing.T) {
	cc := NewChipComboBox([]string{"backend"})
	cc.Focus()

	// Add chip first
	cc.chips.AddChip("backend")

	// Try to add duplicate via selection
	cc, cmd := cc.Update(ComboBoxValueSelectedMsg{Value: "backend", IsNew: false})

	// Should still be 1 chip
	if cc.ChipCount() != 1 {
		t.Errorf("expected 1 chip, got %d", cc.ChipCount())
	}

	// Flash should be set
	if cc.FlashIndex() != 0 {
		t.Errorf("expected flashIndex 0, got %d", cc.FlashIndex())
	}

	// Should return flash command
	if cmd == nil {
		t.Error("expected flash command")
	}
}

func TestChipComboBox_Duplicate_DoesNotAdd(t *testing.T) {
	cc := NewChipComboBox([]string{"backend", "frontend"})
	cc.Focus()

	cc.chips.AddChip("backend")
	cc.chips.AddChip("frontend")

	// Try to add duplicate
	cc, _ = cc.Update(ComboBoxValueSelectedMsg{Value: "backend", IsNew: false})

	if cc.ChipCount() != 2 {
		t.Errorf("expected 2 chips (no duplicate added), got %d", cc.ChipCount())
	}
}

func TestChipComboBox_ChipRemoval_RestoresOption(t *testing.T) {
	cc := NewChipComboBox([]string{"backend", "frontend"})
	cc.Focus()

	// Add chips
	cc.chips.AddChip("backend")
	cc.chips.AddChip("frontend")
	cc.updateAvailableOptions()

	// Verify options are filtered
	if len(cc.combo.FilteredOptions()) != 0 {
		t.Errorf("expected 0 options after adding all chips, got %d", len(cc.combo.FilteredOptions()))
	}

	// Simulate chip removal - chip is removed from ChipList first, then message sent
	cc.chips.RemoveChip("backend")
	cc, _ = cc.Update(ChipRemovedMsg{Label: "backend", Index: 0})

	// backend should be back in options
	found := false
	for _, opt := range cc.combo.Options {
		if opt == "backend" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'backend' to be restored to options")
	}
}

func TestChipComboBox_Tab_AddsChipIfText(t *testing.T) {
	cc := NewChipComboBox([]string{"backend", "frontend"})
	cc.Focus()

	// Type something
	cc, _ = cc.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	cc, _ = cc.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	cc, _ = cc.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	cc, _ = cc.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})

	// Press Tab
	cc, _ = cc.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Should have added chip
	if cc.ChipCount() != 1 {
		t.Errorf("expected 1 chip after Tab, got %d", cc.ChipCount())
	}
}

func TestChipComboBox_Tab_SendsTabMsg(t *testing.T) {
	cc := NewChipComboBox([]string{"backend"})
	cc.Focus()

	_, cmd := cc.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Tab should return a command (either TabMsg directly or batched)
	if cmd == nil {
		t.Fatal("expected command from Tab press")
	}

	// The command is returned - actual TabMsg verification is done via
	// integration testing since tea.Batch wraps commands in a complex way
}

func TestChipComboBox_ChipNavExit_ForwardsCharacter(t *testing.T) {
	cc := NewChipComboBox([]string{"backend"})
	cc.Focus()
	cc.chips.AddChip("backend")
	cc.chips.EnterNavigation()

	// Exit with typing
	cc, _ = cc.Update(ChipNavExitMsg{Reason: ChipNavExitTyping, Character: 'x'})

	// Character should be in input
	if !strings.Contains(cc.InputValue(), "x") {
		t.Errorf("expected 'x' in input, got '%s'", cc.InputValue())
	}
}

func TestChipComboBox_View_ChipsAndInput(t *testing.T) {
	cc := NewChipComboBox([]string{"backend", "frontend"})
	cc.chips.AddChip("backend")
	cc.chips.AddChip("frontend")

	view := cc.View()

	if !strings.Contains(view, "backend") {
		t.Error("expected view to contain 'backend'")
	}
	if !strings.Contains(view, "frontend") {
		t.Error("expected view to contain 'frontend'")
	}
}

func TestChipComboBox_View_ChipNavMode(t *testing.T) {
	cc := NewChipComboBox([]string{"backend"})
	cc.chips.AddChip("backend")
	cc.chips.EnterNavigation()

	view := cc.View()

	// Should show highlighted chip marker
	if !strings.Contains(view, "►backend◄") {
		t.Errorf("expected highlighted chip with markers, got: %s", view)
	}
}

func TestChipComboBox_View_WordWrap(t *testing.T) {
	cc := NewChipComboBox(nil).WithWidth(30)
	cc.chips.AddChip("backend")
	cc.chips.AddChip("frontend")
	cc.chips.AddChip("api")
	cc.chips.AddChip("urgent")

	view := cc.View()
	lines := strings.Split(view, "\n")

	if len(lines) < 2 {
		t.Errorf("expected multiple lines at width 30, got %d lines", len(lines))
	}
}

func TestChipComboBox_FocusBlur(t *testing.T) {
	cc := NewChipComboBox(nil)

	if cc.Focused() {
		t.Error("expected not focused initially")
	}

	cc.Focus()
	if !cc.Focused() {
		t.Error("expected focused after Focus()")
	}

	cc.Blur()
	if cc.Focused() {
		t.Error("expected not focused after Blur()")
	}
}

func TestChipComboBox_SetChips(t *testing.T) {
	cc := NewChipComboBox([]string{"a", "b", "c"})

	cc.SetChips([]string{"a", "b"})

	chips := cc.GetChips()
	if len(chips) != 2 {
		t.Errorf("expected 2 chips, got %d", len(chips))
	}
}

func TestChipComboBox_SetOptions(t *testing.T) {
	cc := NewChipComboBox([]string{"a", "b"})

	cc.SetOptions([]string{"x", "y", "z"})

	if len(cc.allOptions) != 3 {
		t.Errorf("expected 3 options, got %d", len(cc.allOptions))
	}
}

func TestChipComboBox_NewLabel(t *testing.T) {
	cc := NewChipComboBox([]string{"backend"})
	cc.Focus()

	// Select a new value (not in original options)
	cc, cmd := cc.Update(ComboBoxValueSelectedMsg{Value: "newlabel", IsNew: true})

	if cc.ChipCount() != 1 {
		t.Errorf("expected 1 chip, got %d", cc.ChipCount())
	}

	// Check IsNew flag in message
	if cmd == nil {
		t.Fatal("expected command")
	}
	msg := cmd()
	addedMsg, ok := msg.(ChipComboBoxChipAddedMsg)
	if !ok {
		t.Fatalf("expected ChipComboBoxChipAddedMsg, got %T", msg)
	}
	if !addedMsg.IsNew {
		t.Error("expected IsNew to be true")
	}
}

func TestChipComboBox_RenderChips(t *testing.T) {
	// Test that ChipList.RenderChips works correctly
	cl := NewChipList()
	cl.AddChip("backend")
	cl.AddChip("frontend")

	chips := cl.RenderChips()

	if len(chips) != 2 {
		t.Errorf("expected 2 rendered chips, got %d", len(chips))
	}

	// Should contain the text
	if !strings.Contains(chips[0], "backend") {
		t.Error("expected first chip to contain 'backend'")
	}
	if !strings.Contains(chips[1], "frontend") {
		t.Error("expected second chip to contain 'frontend'")
	}
}

func TestChipComboBox_RenderChips_Highlighted(t *testing.T) {
	cl := NewChipList()
	cl.AddChip("backend")
	cl.AddChip("frontend")
	cl.EnterNavigation()

	chips := cl.RenderChips()

	// Last chip should have highlight markers
	if !strings.Contains(chips[1], "►frontend◄") {
		t.Errorf("expected highlighted chip, got: %s", chips[1])
	}
}
