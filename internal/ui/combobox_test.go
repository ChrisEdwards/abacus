package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewComboBox(t *testing.T) {
	t.Run("DefaultValues", func(t *testing.T) {
		cb := NewComboBox(nil)
		if cb.Width != 40 {
			t.Errorf("expected default width 40, got %d", cb.Width)
		}
		if cb.MaxVisible != 5 {
			t.Errorf("expected default MaxVisible 5, got %d", cb.MaxVisible)
		}
		if cb.AllowNew {
			t.Error("expected AllowNew to be false by default")
		}
		if cb.state != ComboBoxIdle {
			t.Errorf("expected initial state ComboBoxIdle, got %v", cb.state)
		}
		if cb.focused {
			t.Error("expected focused to be false initially")
		}
	})

	t.Run("WithOptions", func(t *testing.T) {
		options := []string{"Alice", "Bob", "Carlos"}
		cb := NewComboBox(options)
		if len(cb.Options) != 3 {
			t.Errorf("expected 3 options, got %d", len(cb.Options))
		}
		if len(cb.filteredOptions) != 3 {
			t.Errorf("expected 3 filtered options, got %d", len(cb.filteredOptions))
		}
	})
}

func TestComboBoxBuilders(t *testing.T) {
	t.Run("WithPlaceholder", func(t *testing.T) {
		cb := NewComboBox(nil).WithPlaceholder("Select...")
		if cb.Placeholder != "Select..." {
			t.Errorf("expected placeholder 'Select...', got %s", cb.Placeholder)
		}
	})

	t.Run("WithWidth", func(t *testing.T) {
		cb := NewComboBox(nil).WithWidth(60)
		if cb.Width != 60 {
			t.Errorf("expected width 60, got %d", cb.Width)
		}
	})

	t.Run("WithMaxVisible", func(t *testing.T) {
		cb := NewComboBox(nil).WithMaxVisible(10)
		if cb.MaxVisible != 10 {
			t.Errorf("expected MaxVisible 10, got %d", cb.MaxVisible)
		}
	})

	t.Run("WithAllowNew", func(t *testing.T) {
		cb := NewComboBox(nil).WithAllowNew(true, "Added: %s")
		if !cb.AllowNew {
			t.Error("expected AllowNew to be true")
		}
		if cb.NewItemLabel != "Added: %s" {
			t.Errorf("expected NewItemLabel 'Added: %%s', got %s", cb.NewItemLabel)
		}
	})
}

func TestComboBoxStateTransitions(t *testing.T) {
	t.Run("IdleToBrowsing_OnDownArrow", func(t *testing.T) {
		options := []string{"Alice", "Bob", "Carlos"}
		cb := NewComboBox(options)
		cb.Focus()

		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		if cb.state != ComboBoxBrowsing {
			t.Errorf("expected state ComboBoxBrowsing, got %v", cb.state)
		}
		if len(cb.filteredOptions) != 3 {
			t.Errorf("expected 3 filtered options (full list), got %d", len(cb.filteredOptions))
		}
	})

	t.Run("IdleToFiltering_OnTyping", func(t *testing.T) {
		options := []string{"Alice", "Bob", "Carlos"}
		cb := NewComboBox(options)
		cb.Focus()

		// Type 'a'
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

		if cb.state != ComboBoxFiltering {
			t.Errorf("expected state ComboBoxFiltering, got %v", cb.state)
		}
	})

	t.Run("BrowsingToFiltering_OnTyping", func(t *testing.T) {
		options := []string{"Alice", "Bob", "Carlos"}
		cb := NewComboBox(options)
		cb.Focus()

		// Open dropdown
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})
		if cb.state != ComboBoxBrowsing {
			t.Fatalf("expected state ComboBoxBrowsing, got %v", cb.state)
		}

		// Type 'b'
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})

		if cb.state != ComboBoxFiltering {
			t.Errorf("expected state ComboBoxFiltering, got %v", cb.state)
		}
	})

	t.Run("FilteringToIdle_OnEscape", func(t *testing.T) {
		options := []string{"Alice", "Bob", "Carlos"}
		cb := NewComboBox(options)
		cb.Focus()

		// Type to enter filtering
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		if cb.state != ComboBoxFiltering {
			t.Fatalf("expected state ComboBoxFiltering, got %v", cb.state)
		}

		// Press Escape
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyEsc})

		if cb.state != ComboBoxIdle {
			t.Errorf("expected state ComboBoxIdle, got %v", cb.state)
		}
		// Text should be preserved (first Esc)
		if cb.InputValue() != "a" {
			t.Errorf("expected input value 'a' preserved, got '%s'", cb.InputValue())
		}
	})

	t.Run("BrowsingToIdle_OnEscape", func(t *testing.T) {
		options := []string{"Alice", "Bob", "Carlos"}
		cb := NewComboBox(options)
		cb.Focus()

		// Open dropdown
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		// Press Escape
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyEsc})

		if cb.state != ComboBoxIdle {
			t.Errorf("expected state ComboBoxIdle, got %v", cb.state)
		}
	})
}

func TestComboBoxTwoStageEscape(t *testing.T) {
	t.Run("SecondEscapeRevertsValue", func(t *testing.T) {
		options := []string{"Alice", "Bob", "Carlos"}
		cb := NewComboBox(options)
		cb.SetValue("Alice")
		cb.Focus()

		// Type something different
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})

		// First Esc - close dropdown
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if cb.state != ComboBoxIdle {
			t.Fatalf("expected state ComboBoxIdle after first Esc, got %v", cb.state)
		}
		// Text should still be "bo"
		if cb.InputValue() != "bo" {
			t.Errorf("expected input 'bo' after first Esc, got '%s'", cb.InputValue())
		}

		// Second Esc - revert to original
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if cb.InputValue() != "Alice" {
			t.Errorf("expected input reverted to 'Alice', got '%s'", cb.InputValue())
		}
	})
}

func TestComboBoxSelection(t *testing.T) {
	t.Run("EnterSelectsHighlighted", func(t *testing.T) {
		options := []string{"Alice", "Bob", "Carlos"}
		cb := NewComboBox(options)
		cb.Focus()

		// Open dropdown
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})
		// Move to Bob
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		// Press Enter
		var cmd tea.Cmd
		cb, cmd = cb.Update(tea.KeyMsg{Type: tea.KeyEnter})

		if cb.Value() != "Bob" {
			t.Errorf("expected value 'Bob', got '%s'", cb.Value())
		}
		if cb.state != ComboBoxIdle {
			t.Errorf("expected state ComboBoxIdle after selection, got %v", cb.state)
		}
		// Should send message
		if cmd == nil {
			t.Fatal("expected command to be returned")
		}
		msg := cmd()
		selected, ok := msg.(ComboBoxEnterSelectedMsg)
		if !ok {
			t.Fatalf("expected ComboBoxEnterSelectedMsg, got %T", msg)
		}
		if selected.Value != "Bob" {
			t.Errorf("expected selected value 'Bob', got '%s'", selected.Value)
		}
		if selected.IsNew {
			t.Error("expected IsNew to be false")
		}
	})

	t.Run("TabSelectsHighlighted", func(t *testing.T) {
		options := []string{"Alice", "Bob", "Carlos"}
		cb := NewComboBox(options)
		cb.Focus()

		// Open dropdown and select
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyTab})

		if cb.Value() != "Alice" {
			t.Errorf("expected value 'Alice', got '%s'", cb.Value())
		}
	})

	t.Run("TabReturnsCommand", func(t *testing.T) {
		options := []string{"Alice", "Bob", "Carlos"}
		cb := NewComboBox(options)
		cb.Focus()

		// Open dropdown
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})
		// Navigate to Bob
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		// Press Tab - should return ComboBoxTabSelectedMsg
		var cmd tea.Cmd
		cb, cmd = cb.Update(tea.KeyMsg{Type: tea.KeyTab})

		if cb.Value() != "Bob" {
			t.Errorf("expected value 'Bob', got '%s'", cb.Value())
		}
		if cmd == nil {
			t.Fatal("expected command to be returned from Tab")
		}
		msg := cmd()
		selected, ok := msg.(ComboBoxTabSelectedMsg)
		if !ok {
			t.Fatalf("expected ComboBoxTabSelectedMsg, got %T", msg)
		}
		if selected.Value != "Bob" {
			t.Errorf("expected selected value 'Bob', got '%s'", selected.Value)
		}
	})

	t.Run("TabWithFilteringReturnsCommand", func(t *testing.T) {
		options := []string{"Alice", "Bob"}
		cb := NewComboBox(options).WithAllowNew(true, "New: %s")
		cb.Focus()

		// Type to filter
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})

		// Press Tab - should select Alice and return command
		var cmd tea.Cmd
		cb, cmd = cb.Update(tea.KeyMsg{Type: tea.KeyTab})

		if cb.Value() != "Alice" {
			t.Errorf("expected value 'Alice', got '%s'", cb.Value())
		}
		if cmd == nil {
			t.Fatal("expected command to be returned")
		}
		msg := cmd()
		selected, ok := msg.(ComboBoxTabSelectedMsg)
		if !ok {
			t.Fatalf("expected ComboBoxTabSelectedMsg, got %T", msg)
		}
		if selected.IsNew {
			t.Error("expected IsNew to be false for existing option")
		}
	})

	t.Run("TabWithNoMatchCreatesNew", func(t *testing.T) {
		options := []string{"Alice", "Bob"}
		cb := NewComboBox(options).WithAllowNew(true, "New: %s")
		cb.Focus()

		// Type new value with no match
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Z'}})
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

		// Press Tab - should create new value
		var cmd tea.Cmd
		cb, cmd = cb.Update(tea.KeyMsg{Type: tea.KeyTab})

		if cb.Value() != "Zed" {
			t.Errorf("expected value 'Zed', got '%s'", cb.Value())
		}
		if cmd == nil {
			t.Fatal("expected command to be returned")
		}
		msg := cmd()
		selected, ok := msg.(ComboBoxTabSelectedMsg)
		if !ok {
			t.Fatalf("expected ComboBoxTabSelectedMsg, got %T", msg)
		}
		if !selected.IsNew {
			t.Error("expected IsNew to be true for new value")
		}
	})

	t.Run("SelectionPreservesValue", func(t *testing.T) {
		options := []string{"Alice", "Bob", "Carlos"}
		cb := NewComboBox(options)
		cb.SetValue("Carlos")
		cb.Focus()

		// Open dropdown - should highlight Carlos
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		if cb.highlightIndex != 2 {
			t.Errorf("expected highlight index 2 (Carlos), got %d", cb.highlightIndex)
		}
	})
}

func TestComboBoxNavigation(t *testing.T) {
	t.Run("UpDownNavigatesOptions", func(t *testing.T) {
		options := []string{"Alice", "Bob", "Carlos"}
		cb := NewComboBox(options)
		cb.Focus()

		// Open dropdown
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})
		if cb.highlightIndex != 0 {
			t.Errorf("expected highlight at 0, got %d", cb.highlightIndex)
		}

		// Move down
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})
		if cb.highlightIndex != 1 {
			t.Errorf("expected highlight at 1, got %d", cb.highlightIndex)
		}

		// Move up
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyUp})
		if cb.highlightIndex != 0 {
			t.Errorf("expected highlight at 0, got %d", cb.highlightIndex)
		}
	})

	t.Run("NavigationStopsAtBounds", func(t *testing.T) {
		options := []string{"Alice", "Bob", "Carlos"}
		cb := NewComboBox(options)
		cb.Focus()

		// Open dropdown
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		// Try to go up from 0
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyUp})
		if cb.highlightIndex != 0 {
			t.Errorf("expected highlight to stay at 0, got %d", cb.highlightIndex)
		}

		// Go to end
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown}) // Try past end

		if cb.highlightIndex != 2 {
			t.Errorf("expected highlight to stay at 2, got %d", cb.highlightIndex)
		}
	})
}

func TestComboBoxFiltering(t *testing.T) {
	t.Run("FilterUpdatesOnTyping", func(t *testing.T) {
		options := []string{"Alice", "Albert", "Bob", "Carlos"}
		cb := NewComboBox(options)
		cb.Focus()

		// Type 'al'
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})

		if len(cb.filteredOptions) != 2 {
			t.Errorf("expected 2 filtered options (Alice, Albert), got %d", len(cb.filteredOptions))
		}
	})

	t.Run("FilterIsCaseInsensitive", func(t *testing.T) {
		options := []string{"Alice", "ALBERT", "Bob"}
		cb := NewComboBox(options)
		cb.Focus()

		// Type 'AL' (uppercase)
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'L'}})

		if len(cb.filteredOptions) != 2 {
			t.Errorf("expected 2 filtered options, got %d", len(cb.filteredOptions))
		}
	})

	t.Run("FilterLimitsToMaxVisible", func(t *testing.T) {
		options := make([]string, 20)
		for i := range options {
			options[i] = "Item" + string(rune('A'+i))
		}
		cb := NewComboBox(options).WithMaxVisible(5)
		cb.Focus()

		// Type 'I' to match all
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'I'}})

		// All matches are stored in filteredOptions for navigation
		if len(cb.filteredOptions) != 20 {
			t.Errorf("expected 20 filtered options, got %d", len(cb.filteredOptions))
		}

		// But View only renders MaxVisible items in the dropdown
		view := cb.View()
		// Count dropdown items by looking for the marker pattern (▸ or indented items)
		// The dropdown starts after the input box closes (after ╯)
		dropdownStart := strings.Index(view, "╯")
		if dropdownStart == -1 {
			t.Fatal("expected dropdown in view")
		}
		dropdownPart := view[dropdownStart:]
		// Count how many "Item" entries appear in dropdown (should be limited to MaxVisible)
		itemCount := strings.Count(dropdownPart, "Item")
		if itemCount > 5 {
			t.Errorf("expected max 5 visible items in dropdown, got %d", itemCount)
		}
	})

	t.Run("FilterResetsHighlightToZero", func(t *testing.T) {
		options := []string{"Alice", "Bob", "Carlos"}
		cb := NewComboBox(options)
		cb.Focus()

		// Open and navigate
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		// Type to filter
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})

		if cb.highlightIndex != 0 {
			t.Errorf("expected highlight reset to 0, got %d", cb.highlightIndex)
		}
	})

	t.Run("FilterHighlightsExactMatch", func(t *testing.T) {
		// Regression test for ab-mod2/ab-qa72: exact matches should be highlighted
		options := []string{"build", "UI", "ui-redesign"}
		cb := NewComboBox(options)
		cb.Focus()

		// Type 'ui' - should match all three but highlight "UI" (exact match)
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

		// Should have 3 matches
		if len(cb.filteredOptions) != 3 {
			t.Errorf("expected 3 filtered options, got %d: %v", len(cb.filteredOptions), cb.filteredOptions)
		}

		// "UI" should be highlighted (it's the exact match)
		// Find index of "UI" in filtered options
		uiIndex := -1
		for i, opt := range cb.filteredOptions {
			if opt == "UI" {
				uiIndex = i
				break
			}
		}
		if uiIndex == -1 {
			t.Fatal("UI should be in filtered options")
		}
		if cb.highlightIndex != uiIndex {
			t.Errorf("expected highlight on UI (index %d), got index %d (value: %s)",
				uiIndex, cb.highlightIndex, cb.filteredOptions[cb.highlightIndex])
		}
	})

	t.Run("FilterHighlightsFirstWhenNoExactMatch", func(t *testing.T) {
		options := []string{"build", "builder", "rebuild"}
		cb := NewComboBox(options)
		cb.Focus()

		// Type 'buil' - matches "build" and "builder" but no exact match
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})

		// Should highlight first match (index 0)
		if cb.highlightIndex != 0 {
			t.Errorf("expected highlight on first match (index 0), got %d", cb.highlightIndex)
		}
	})
}
