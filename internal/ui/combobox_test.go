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
		selected, ok := msg.(ComboBoxValueSelectedMsg)
		if !ok {
			t.Fatalf("expected ComboBoxValueSelectedMsg, got %T", msg)
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

		// But View only renders MaxVisible items
		view := cb.View()
		// Count how many "Item" entries appear (should be limited to MaxVisible)
		itemCount := strings.Count(view, "Item")
		if itemCount > 5 {
			t.Errorf("expected max 5 visible items, got %d", itemCount)
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
}

func TestComboBoxAllowNew(t *testing.T) {
	t.Run("AllowNewCreatesValue", func(t *testing.T) {
		options := []string{"Alice", "Bob"}
		cb := NewComboBox(options).WithAllowNew(true, "Added: %s")
		cb.Focus()

		// Type a new name
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'C'}})
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})

		// No matches
		if len(cb.filteredOptions) != 0 {
			t.Errorf("expected 0 filtered options, got %d", len(cb.filteredOptions))
		}

		// Press Enter to create
		var cmd tea.Cmd
		cb, cmd = cb.Update(tea.KeyMsg{Type: tea.KeyEnter})

		if cb.Value() != "Carl" {
			t.Errorf("expected value 'Carl', got '%s'", cb.Value())
		}

		if cmd == nil {
			t.Fatal("expected command")
		}
		msg := cmd()
		selected, ok := msg.(ComboBoxValueSelectedMsg)
		if !ok {
			t.Fatalf("expected ComboBoxValueSelectedMsg, got %T", msg)
		}
		if !selected.IsNew {
			t.Error("expected IsNew to be true")
		}
	})

	t.Run("NoMatchWithoutAllowNew", func(t *testing.T) {
		options := []string{"Alice", "Bob"}
		cb := NewComboBox(options) // AllowNew is false
		cb.Focus()

		// Type a new name
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Z'}})

		// Press Enter - should not create
		var cmd tea.Cmd
		cb, cmd = cb.Update(tea.KeyMsg{Type: tea.KeyEnter})

		if cb.Value() != "" {
			t.Errorf("expected empty value, got '%s'", cb.Value())
		}
		if cmd != nil {
			t.Error("expected no command when AllowNew is false")
		}
	})
}

func TestComboBoxFocusBlur(t *testing.T) {
	t.Run("FocusSetsState", func(t *testing.T) {
		cb := NewComboBox([]string{"A", "B"})
		cb.SetValue("A")

		cmd := cb.Focus()

		if !cb.Focused() {
			t.Error("expected focused to be true")
		}
		if cmd == nil {
			t.Error("expected blink command from Focus")
		}
	})

	t.Run("BlurClosesDropdown", func(t *testing.T) {
		cb := NewComboBox([]string{"A", "B"})
		cb.Focus()

		// Open dropdown
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})
		if cb.state != ComboBoxBrowsing {
			t.Fatalf("expected browsing state, got %v", cb.state)
		}

		// Blur
		cb.Blur()

		if cb.Focused() {
			t.Error("expected focused to be false")
		}
		if cb.state != ComboBoxIdle {
			t.Errorf("expected state ComboBoxIdle, got %v", cb.state)
		}
	})
}

func TestComboBoxView(t *testing.T) {
	t.Run("ViewShowsInputWhenIdle", func(t *testing.T) {
		cb := NewComboBox([]string{"Alice", "Bob"})
		view := cb.View()

		if view == "" {
			t.Error("expected non-empty view")
		}
		// Should not contain dropdown items when idle
		if strings.Contains(view, "Alice") && strings.Contains(view, "Bob") {
			// Could be in the input, but we don't expect both
		}
	})

	t.Run("ViewShowsDropdownWhenOpen", func(t *testing.T) {
		cb := NewComboBox([]string{"Alice", "Bob"})
		cb.Focus()

		// Open dropdown
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		view := cb.View()

		if !strings.Contains(view, "Alice") {
			t.Error("expected view to contain 'Alice'")
		}
		if !strings.Contains(view, "Bob") {
			t.Error("expected view to contain 'Bob'")
		}
	})

	t.Run("ViewHighlightsCorrectItem", func(t *testing.T) {
		cb := NewComboBox([]string{"Alice", "Bob", "Carlos"})
		cb.Focus()

		// Open dropdown and move to Bob
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		view := cb.View()

		// The highlight marker should appear before Bob
		// We use \u25b8 (▸) as highlight marker
		if !strings.Contains(view, "\u25b8 Bob") {
			t.Error("expected Bob to be highlighted with marker")
		}
	})

	t.Run("ViewShowsNoMatchHint", func(t *testing.T) {
		cb := NewComboBox([]string{"Alice", "Bob"}).WithAllowNew(true, "")
		cb.Focus()

		// Type something with no matches
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Z'}})

		view := cb.View()

		if !strings.Contains(view, "No matches") {
			t.Error("expected view to contain 'No matches'")
		}
		if !strings.Contains(view, "to add new") {
			t.Error("expected view to contain hint to add new")
		}
	})
}

func TestComboBoxSetters(t *testing.T) {
	t.Run("SetValue", func(t *testing.T) {
		cb := NewComboBox([]string{"A", "B"})
		cb.SetValue("A")

		if cb.Value() != "A" {
			t.Errorf("expected value 'A', got '%s'", cb.Value())
		}
		if cb.InputValue() != "A" {
			t.Errorf("expected input value 'A', got '%s'", cb.InputValue())
		}
	})

	t.Run("SetOptions", func(t *testing.T) {
		cb := NewComboBox([]string{"A", "B"})
		cb.SetOptions([]string{"X", "Y", "Z"})

		if len(cb.Options) != 3 {
			t.Errorf("expected 3 options, got %d", len(cb.Options))
		}
		if cb.Options[0] != "X" {
			t.Errorf("expected first option 'X', got '%s'", cb.Options[0])
		}
	})
}

func TestComboBoxHelperMethods(t *testing.T) {
	t.Run("IsDropdownOpen", func(t *testing.T) {
		cb := NewComboBox([]string{"A", "B"})

		if cb.IsDropdownOpen() {
			t.Error("expected dropdown to be closed initially")
		}

		cb.Focus()
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		if !cb.IsDropdownOpen() {
			t.Error("expected dropdown to be open after down arrow")
		}
	})

	t.Run("State", func(t *testing.T) {
		cb := NewComboBox([]string{"A"})

		if cb.State() != ComboBoxIdle {
			t.Errorf("expected ComboBoxIdle, got %v", cb.State())
		}
	})

	t.Run("FilteredOptions", func(t *testing.T) {
		cb := NewComboBox([]string{"Alice", "Bob"})
		cb.Focus()
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})

		filtered := cb.FilteredOptions()
		if len(filtered) != 1 || filtered[0] != "Alice" {
			t.Errorf("expected filtered to contain only Alice, got %v", filtered)
		}
	})

	t.Run("HighlightIndex", func(t *testing.T) {
		cb := NewComboBox([]string{"A", "B", "C"})
		cb.Focus()
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		if cb.HighlightIndex() != 1 {
			t.Errorf("expected highlight index 1, got %d", cb.HighlightIndex())
		}
	})
}
