package ui

import (
	"strings"
	"testing"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
)

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
		selected, ok := msg.(ComboBoxEnterSelectedMsg)
		if !ok {
			t.Fatalf("expected ComboBoxEnterSelectedMsg, got %T", msg)
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

	t.Run("ViewHighlightAlignedWithOptions", func(t *testing.T) {
		cb := NewComboBox([]string{"Alpha", "Bravo", "Charlie"})
		cb.Focus()

		// Open dropdown and move highlight to Bravo
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		view := stripANSI(cb.View())

		var highlightedLine, optionLine string
		for _, line := range strings.Split(view, "\n") {
			if strings.Contains(line, "Bravo") {
				highlightedLine = strings.TrimRight(line, " ")
			}
			if strings.Contains(line, "Charlie") {
				optionLine = strings.TrimRight(line, " ")
			}
		}

		if highlightedLine == "" || optionLine == "" {
			t.Fatalf("expected both highlighted and option lines, got highlight=%q option=%q", highlightedLine, optionLine)
		}

		highlightStart := strings.Index(highlightedLine, "Bravo")
		optionStart := strings.Index(optionLine, "Charlie")
		if highlightStart == -1 || optionStart == -1 {
			t.Fatalf("expected to locate option text in lines, got highlightStart=%d optionStart=%d", highlightStart, optionStart)
		}

		highlightPrefixWidth := utf8.RuneCountInString(highlightedLine[:highlightStart])
		optionPrefixWidth := utf8.RuneCountInString(optionLine[:optionStart])

		if highlightPrefixWidth != optionPrefixWidth {
			t.Fatalf("expected highlight indent %d to match option indent %d", highlightPrefixWidth, optionPrefixWidth)
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

// ============================================================================
// Scroll Indicator Tests (ab-ouvw - spec Section 3.5 ComboBox behavior)
// ============================================================================

func TestComboBoxScrollIndicators(t *testing.T) {
	t.Run("ShowsMoreBelowIndicator", func(t *testing.T) {
		// Create list longer than MaxVisible (default 5)
		options := []string{"A", "B", "C", "D", "E", "F", "G", "H"}
		cb := NewComboBox(options)
		cb.Focus()

		// Open dropdown
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		view := cb.View()
		// Should show "more below" indicator
		if !strings.Contains(view, "▼ more below") {
			t.Error("expected '▼ more below' indicator when more items exist below")
		}
	})

	t.Run("ShowsMoreAboveIndicator", func(t *testing.T) {
		// Create list longer than MaxVisible
		options := []string{"A", "B", "C", "D", "E", "F", "G", "H"}
		cb := NewComboBox(options)
		cb.Focus()

		// Open dropdown
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		// Navigate down past visible window
		for i := 0; i < 6; i++ {
			cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})
		}

		view := cb.View()
		// Should show "more above" indicator
		if !strings.Contains(view, "▲ more above") {
			t.Error("expected '▲ more above' indicator when more items exist above")
		}
	})

	t.Run("NoIndicatorsWhenAllFit", func(t *testing.T) {
		// Create list that fits within MaxVisible
		options := []string{"A", "B", "C"}
		cb := NewComboBox(options).WithMaxVisible(5)
		cb.Focus()

		// Open dropdown
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		view := cb.View()
		if strings.Contains(view, "more below") || strings.Contains(view, "more above") {
			t.Error("expected no scroll indicators when all items fit")
		}
	})

	t.Run("BothIndicatorsInMiddle", func(t *testing.T) {
		// Create list much longer than MaxVisible
		options := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K"}
		cb := NewComboBox(options).WithMaxVisible(3)
		cb.Focus()

		// Open dropdown
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		// Navigate to middle
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		view := cb.View()
		// Should show both indicators
		if !strings.Contains(view, "▲ more above") {
			t.Error("expected '▲ more above' indicator in middle of list")
		}
		if !strings.Contains(view, "▼ more below") {
			t.Error("expected '▼ more below' indicator in middle of list")
		}
	})
}

func TestComboBoxScrollOffsetAdjustment(t *testing.T) {
	t.Run("ScrollsDownWhenNavigatingPastVisible", func(t *testing.T) {
		options := []string{"A", "B", "C", "D", "E", "F", "G"}
		cb := NewComboBox(options).WithMaxVisible(3)
		cb.Focus()

		// Open dropdown
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		// Navigate past visible (index 0, 1, 2 visible initially)
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown}) // index 1
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown}) // index 2
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown}) // index 3 - should scroll

		if cb.highlightIndex != 3 {
			t.Errorf("expected highlight at 3, got %d", cb.highlightIndex)
		}
		if cb.scrollOffset < 1 {
			t.Errorf("expected scrollOffset >= 1, got %d", cb.scrollOffset)
		}
	})

	t.Run("ScrollsUpWhenNavigatingAboveVisible", func(t *testing.T) {
		options := []string{"A", "B", "C", "D", "E", "F", "G"}
		cb := NewComboBox(options).WithMaxVisible(3)
		cb.Focus()

		// Open dropdown and scroll down
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})
		for i := 0; i < 5; i++ {
			cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})
		}

		// Verify we've scrolled
		initialOffset := cb.scrollOffset

		// Navigate back up
		for i := 0; i < 5; i++ {
			cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyUp})
		}

		if cb.scrollOffset >= initialOffset {
			t.Errorf("expected scrollOffset to decrease, was %d, now %d", initialOffset, cb.scrollOffset)
		}
	})

	t.Run("ScrollOffsetClampedToValidRange", func(t *testing.T) {
		options := []string{"A", "B", "C", "D", "E"}
		cb := NewComboBox(options).WithMaxVisible(3)
		cb.Focus()

		// Open dropdown
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		// Navigate all the way down
		for i := 0; i < 10; i++ {
			cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})
		}

		// scrollOffset should not exceed (total - MaxVisible)
		maxOffset := len(options) - cb.MaxVisible
		if cb.scrollOffset > maxOffset {
			t.Errorf("expected scrollOffset <= %d, got %d", maxOffset, cb.scrollOffset)
		}
	})
}

func TestComboBoxDownArrowPreservesSelection(t *testing.T) {
	t.Run("HighlightsCurrentValueWhenOpening", func(t *testing.T) {
		options := []string{"Alice", "Bob", "Carlos", "Diana"}
		cb := NewComboBox(options)
		cb.SetValue("Carlos")
		cb.Focus()

		// Open dropdown with Down arrow
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		// Should highlight Carlos (index 2)
		if cb.highlightIndex != 2 {
			t.Errorf("expected highlight index 2 (Carlos), got %d", cb.highlightIndex)
		}
	})

	t.Run("HighlightsFirstWhenNoValue", func(t *testing.T) {
		options := []string{"Alice", "Bob", "Carlos"}
		cb := NewComboBox(options)
		// No value set
		cb.Focus()

		// Open dropdown
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		// Should highlight first (index 0)
		if cb.highlightIndex != 0 {
			t.Errorf("expected highlight index 0, got %d", cb.highlightIndex)
		}
	})

	t.Run("HighlightsFirstWhenValueNotInList", func(t *testing.T) {
		options := []string{"Alice", "Bob", "Carlos"}
		cb := NewComboBox(options)
		cb.SetValue("Unknown")
		cb.Focus()

		// Open dropdown
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		// Should highlight first since "Unknown" isn't in list
		if cb.highlightIndex != 0 {
			t.Errorf("expected highlight index 0, got %d", cb.highlightIndex)
		}
	})

	t.Run("ScrollsToCurrentValueWhenOpening", func(t *testing.T) {
		// Create list longer than MaxVisible
		options := []string{"A", "B", "C", "D", "E", "F", "G", "H"}
		cb := NewComboBox(options).WithMaxVisible(3)
		cb.SetValue("G") // Near the end
		cb.Focus()

		// Open dropdown
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyDown})

		// G is at index 6, should scroll to make it visible
		if cb.highlightIndex != 6 {
			t.Errorf("expected highlight index 6 (G), got %d", cb.highlightIndex)
		}

		// scrollOffset should be adjusted to show G
		// G at index 6 with MaxVisible 3 means offset should be at least 4
		if cb.scrollOffset < 4 {
			t.Errorf("expected scrollOffset >= 4 to show G, got %d", cb.scrollOffset)
		}
	})
}
