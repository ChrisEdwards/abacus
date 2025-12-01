package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewChipList(t *testing.T) {
	t.Run("DefaultValues", func(t *testing.T) {
		cl := NewChipList()
		if cl.Width != 40 {
			t.Errorf("expected default width 40, got %d", cl.Width)
		}
		if len(cl.Chips) != 0 {
			t.Errorf("expected empty chips, got %d", len(cl.Chips))
		}
		if cl.state != ChipListInput {
			t.Errorf("expected ChipListInput state, got %v", cl.state)
		}
		if cl.navIndex != -1 {
			t.Errorf("expected navIndex -1, got %d", cl.navIndex)
		}
		if cl.focused {
			t.Error("expected focused to be false")
		}
		if cl.flashIndex != -1 {
			t.Errorf("expected flashIndex -1, got %d", cl.flashIndex)
		}
	})
}

func TestChipListBuilders(t *testing.T) {
	t.Run("WithWidth", func(t *testing.T) {
		cl := NewChipList().WithWidth(80)
		if cl.Width != 80 {
			t.Errorf("expected width 80, got %d", cl.Width)
		}
	})
}

func TestChipList_AddChip(t *testing.T) {
	t.Run("AddSingleChip", func(t *testing.T) {
		cl := NewChipList()
		ok := cl.AddChip("backend")
		if !ok {
			t.Error("expected AddChip to return true")
		}
		if len(cl.Chips) != 1 {
			t.Errorf("expected 1 chip, got %d", len(cl.Chips))
		}
		if cl.Chips[0] != "backend" {
			t.Errorf("expected 'backend', got '%s'", cl.Chips[0])
		}
	})

	t.Run("AddMultipleChips", func(t *testing.T) {
		cl := NewChipList()
		cl.AddChip("backend")
		cl.AddChip("frontend")
		cl.AddChip("api")
		if len(cl.Chips) != 3 {
			t.Errorf("expected 3 chips, got %d", len(cl.Chips))
		}
	})

	t.Run("AddEmptyChip", func(t *testing.T) {
		cl := NewChipList()
		ok := cl.AddChip("")
		if ok {
			t.Error("expected AddChip to return false for empty string")
		}
		ok = cl.AddChip("   ")
		if ok {
			t.Error("expected AddChip to return false for whitespace")
		}
		if len(cl.Chips) != 0 {
			t.Errorf("expected 0 chips, got %d", len(cl.Chips))
		}
	})
}

func TestChipList_AddChip_Duplicate(t *testing.T) {
	t.Run("DuplicateReturnsFalse", func(t *testing.T) {
		cl := NewChipList()
		cl.AddChip("backend")
		ok := cl.AddChip("backend")
		if ok {
			t.Error("expected AddChip to return false for duplicate")
		}
		if len(cl.Chips) != 1 {
			t.Errorf("expected 1 chip, got %d", len(cl.Chips))
		}
	})

	t.Run("DuplicateCaseInsensitive", func(t *testing.T) {
		cl := NewChipList()
		cl.AddChip("Backend")
		ok := cl.AddChip("BACKEND")
		if ok {
			t.Error("expected AddChip to return false for case-insensitive duplicate")
		}
		ok = cl.AddChip("backend")
		if ok {
			t.Error("expected AddChip to return false for case-insensitive duplicate")
		}
	})

	t.Run("DuplicateSetsFlashIndex", func(t *testing.T) {
		cl := NewChipList()
		cl.AddChip("first")
		cl.AddChip("second")
		cl.AddChip("third")
		cl.AddChip("second") // duplicate
		if cl.flashIndex != 1 {
			t.Errorf("expected flashIndex 1, got %d", cl.flashIndex)
		}
	})
}

func TestChipList_RemoveChip(t *testing.T) {
	t.Run("RemoveExisting", func(t *testing.T) {
		cl := NewChipList()
		cl.AddChip("backend")
		cl.AddChip("frontend")
		cl.RemoveChip("backend")
		if len(cl.Chips) != 1 {
			t.Errorf("expected 1 chip, got %d", len(cl.Chips))
		}
		if cl.Chips[0] != "frontend" {
			t.Errorf("expected 'frontend', got '%s'", cl.Chips[0])
		}
	})

	t.Run("RemoveNonExisting", func(t *testing.T) {
		cl := NewChipList()
		cl.AddChip("backend")
		cl.RemoveChip("nonexistent")
		if len(cl.Chips) != 1 {
			t.Errorf("expected 1 chip, got %d", len(cl.Chips))
		}
	})

	t.Run("RemoveCaseInsensitive", func(t *testing.T) {
		cl := NewChipList()
		cl.AddChip("Backend")
		cl.RemoveChip("BACKEND")
		if len(cl.Chips) != 0 {
			t.Errorf("expected 0 chips, got %d", len(cl.Chips))
		}
	})
}

func TestChipList_Contains(t *testing.T) {
	cl := NewChipList()
	cl.AddChip("backend")
	cl.AddChip("frontend")

	if !cl.Contains("backend") {
		t.Error("expected Contains to return true for 'backend'")
	}
	if !cl.Contains("BACKEND") {
		t.Error("expected Contains to be case-insensitive")
	}
	if cl.Contains("api") {
		t.Error("expected Contains to return false for 'api'")
	}
}

func TestChipList_GetChips(t *testing.T) {
	cl := NewChipList()
	cl.AddChip("a")
	cl.AddChip("b")

	chips := cl.GetChips()
	if len(chips) != 2 {
		t.Errorf("expected 2 chips, got %d", len(chips))
	}

	// Verify it's a copy
	chips[0] = "modified"
	if cl.Chips[0] == "modified" {
		t.Error("expected GetChips to return a copy")
	}
}

func TestChipList_EnterNavigation_HighlightsLast(t *testing.T) {
	cl := NewChipList()
	cl.AddChip("first")
	cl.AddChip("second")
	cl.AddChip("third")

	ok := cl.EnterNavigation()
	if !ok {
		t.Error("expected EnterNavigation to return true")
	}
	if cl.state != ChipListNavigation {
		t.Errorf("expected ChipListNavigation state, got %v", cl.state)
	}
	if cl.navIndex != 2 {
		t.Errorf("expected navIndex 2 (last chip), got %d", cl.navIndex)
	}
	if cl.HighlightedChip() != "third" {
		t.Errorf("expected highlighted chip 'third', got '%s'", cl.HighlightedChip())
	}
}

func TestChipList_EnterNavigation_EmptyList(t *testing.T) {
	cl := NewChipList()

	ok := cl.EnterNavigation()
	if ok {
		t.Error("expected EnterNavigation to return false for empty list")
	}
	if cl.state != ChipListInput {
		t.Errorf("expected ChipListInput state, got %v", cl.state)
	}
}

func TestChipList_Navigation_Left(t *testing.T) {
	cl := NewChipList()
	cl.AddChip("first")
	cl.AddChip("second")
	cl.AddChip("third")
	cl.EnterNavigation() // starts at index 2

	// Move left
	cl, _ = cl.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if cl.navIndex != 1 {
		t.Errorf("expected navIndex 1, got %d", cl.navIndex)
	}

	// Move left again
	cl, _ = cl.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if cl.navIndex != 0 {
		t.Errorf("expected navIndex 0, got %d", cl.navIndex)
	}

	// Try to move left past first - should stay at 0
	cl, _ = cl.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if cl.navIndex != 0 {
		t.Errorf("expected navIndex to stay at 0, got %d", cl.navIndex)
	}
}

func TestChipList_Navigation_Right(t *testing.T) {
	cl := NewChipList()
	cl.AddChip("first")
	cl.AddChip("second")
	cl.AddChip("third")
	cl.EnterNavigation()

	// Move to first
	cl, _ = cl.Update(tea.KeyMsg{Type: tea.KeyLeft})
	cl, _ = cl.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if cl.navIndex != 0 {
		t.Fatalf("expected navIndex 0, got %d", cl.navIndex)
	}

	// Move right
	cl, _ = cl.Update(tea.KeyMsg{Type: tea.KeyRight})
	if cl.navIndex != 1 {
		t.Errorf("expected navIndex 1, got %d", cl.navIndex)
	}
}

func TestChipList_Navigation_RightPastEnd(t *testing.T) {
	cl := NewChipList()
	cl.AddChip("first")
	cl.AddChip("second")
	cl.EnterNavigation() // at index 1

	var cmd tea.Cmd
	cl, cmd = cl.Update(tea.KeyMsg{Type: tea.KeyRight})

	// Should exit navigation
	if cl.state != ChipListInput {
		t.Errorf("expected ChipListInput state, got %v", cl.state)
	}
	if cl.navIndex != -1 {
		t.Errorf("expected navIndex -1, got %d", cl.navIndex)
	}

	// Should send exit message
	if cmd == nil {
		t.Fatal("expected command to be returned")
	}
	msg := cmd()
	exitMsg, ok := msg.(ChipNavExitMsg)
	if !ok {
		t.Fatalf("expected ChipNavExitMsg, got %T", msg)
	}
	if exitMsg.Reason != ChipNavExitRight {
		t.Errorf("expected ChipNavExitRight reason, got %v", exitMsg.Reason)
	}
}

func TestChipList_Delete_LastChip(t *testing.T) {
	cl := NewChipList()
	cl.AddChip("first")
	cl.AddChip("second")
	cl.AddChip("third")
	cl.EnterNavigation() // at index 2 (third)

	var cmd tea.Cmd
	cl, cmd = cl.Update(tea.KeyMsg{Type: tea.KeyBackspace})

	// Should remove "third" and highlight "second" (index 1)
	if len(cl.Chips) != 2 {
		t.Errorf("expected 2 chips, got %d", len(cl.Chips))
	}
	if cl.navIndex != 1 {
		t.Errorf("expected navIndex 1 (previous), got %d", cl.navIndex)
	}
	if cl.HighlightedChip() != "second" {
		t.Errorf("expected highlighted 'second', got '%s'", cl.HighlightedChip())
	}

	// Should send removed message
	if cmd == nil {
		t.Fatal("expected command")
	}
	msg := cmd()
	removedMsg, ok := msg.(ChipRemovedMsg)
	if !ok {
		t.Fatalf("expected ChipRemovedMsg, got %T", msg)
	}
	if removedMsg.Label != "third" {
		t.Errorf("expected removed label 'third', got '%s'", removedMsg.Label)
	}
}

func TestChipList_Delete_FirstChip(t *testing.T) {
	cl := NewChipList()
	cl.AddChip("first")
	cl.AddChip("second")
	cl.AddChip("third")
	cl.EnterNavigation()
	// Navigate to first
	cl, _ = cl.Update(tea.KeyMsg{Type: tea.KeyLeft})
	cl, _ = cl.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if cl.navIndex != 0 {
		t.Fatalf("expected navIndex 0, got %d", cl.navIndex)
	}

	cl, _ = cl.Update(tea.KeyMsg{Type: tea.KeyDelete})

	// Should remove "first", stay at index 0 (now "second")
	if len(cl.Chips) != 2 {
		t.Errorf("expected 2 chips, got %d", len(cl.Chips))
	}
	if cl.navIndex != 0 {
		t.Errorf("expected navIndex 0 (same), got %d", cl.navIndex)
	}
	if cl.HighlightedChip() != "second" {
		t.Errorf("expected highlighted 'second', got '%s'", cl.HighlightedChip())
	}
}

func TestChipList_Delete_MiddleChip(t *testing.T) {
	cl := NewChipList()
	cl.AddChip("first")
	cl.AddChip("second")
	cl.AddChip("third")
	cl.EnterNavigation()
	// Navigate to middle
	cl, _ = cl.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if cl.navIndex != 1 {
		t.Fatalf("expected navIndex 1, got %d", cl.navIndex)
	}

	cl, _ = cl.Update(tea.KeyMsg{Type: tea.KeyBackspace})

	// Should remove "second", stay at index 1 (now "third")
	if len(cl.Chips) != 2 {
		t.Errorf("expected 2 chips, got %d", len(cl.Chips))
	}
	if cl.navIndex != 1 {
		t.Errorf("expected navIndex 1 (same), got %d", cl.navIndex)
	}
	if cl.HighlightedChip() != "third" {
		t.Errorf("expected highlighted 'third', got '%s'", cl.HighlightedChip())
	}
}

func TestChipList_Delete_OnlyChip(t *testing.T) {
	cl := NewChipList()
	cl.AddChip("only")
	cl.EnterNavigation()

	var cmd tea.Cmd
	cl, cmd = cl.Update(tea.KeyMsg{Type: tea.KeyBackspace})

	// Should remove chip and exit navigation
	if len(cl.Chips) != 0 {
		t.Errorf("expected 0 chips, got %d", len(cl.Chips))
	}
	if cl.state != ChipListInput {
		t.Errorf("expected ChipListInput state, got %v", cl.state)
	}
	if cl.navIndex != -1 {
		t.Errorf("expected navIndex -1, got %d", cl.navIndex)
	}

	// Should send removed message
	if cmd == nil {
		t.Fatal("expected command")
	}
	msg := cmd()
	removedMsg, ok := msg.(ChipRemovedMsg)
	if !ok {
		t.Fatalf("expected ChipRemovedMsg, got %T", msg)
	}
	if removedMsg.Label != "only" {
		t.Errorf("expected removed label 'only', got '%s'", removedMsg.Label)
	}
}

func TestChipList_Exit_Escape(t *testing.T) {
	cl := NewChipList()
	cl.AddChip("chip")
	cl.EnterNavigation()

	var cmd tea.Cmd
	cl, cmd = cl.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cl.state != ChipListInput {
		t.Errorf("expected ChipListInput state, got %v", cl.state)
	}

	if cmd == nil {
		t.Fatal("expected command")
	}
	msg := cmd()
	exitMsg, ok := msg.(ChipNavExitMsg)
	if !ok {
		t.Fatalf("expected ChipNavExitMsg, got %T", msg)
	}
	if exitMsg.Reason != ChipNavExitEscape {
		t.Errorf("expected ChipNavExitEscape, got %v", exitMsg.Reason)
	}
}

func TestChipList_Exit_Tab(t *testing.T) {
	cl := NewChipList()
	cl.AddChip("chip")
	cl.EnterNavigation()

	var cmd tea.Cmd
	cl, cmd = cl.Update(tea.KeyMsg{Type: tea.KeyTab})

	if cl.state != ChipListInput {
		t.Errorf("expected ChipListInput state, got %v", cl.state)
	}

	if cmd == nil {
		t.Fatal("expected command")
	}
	msg := cmd()
	exitMsg, ok := msg.(ChipNavExitMsg)
	if !ok {
		t.Fatalf("expected ChipNavExitMsg, got %T", msg)
	}
	if exitMsg.Reason != ChipNavExitTab {
		t.Errorf("expected ChipNavExitTab, got %v", exitMsg.Reason)
	}
}

func TestChipList_Exit_Letter(t *testing.T) {
	cl := NewChipList()
	cl.AddChip("chip")
	cl.EnterNavigation()

	var cmd tea.Cmd
	cl, cmd = cl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	if cl.state != ChipListInput {
		t.Errorf("expected ChipListInput state, got %v", cl.state)
	}

	if cmd == nil {
		t.Fatal("expected command")
	}
	msg := cmd()
	exitMsg, ok := msg.(ChipNavExitMsg)
	if !ok {
		t.Fatalf("expected ChipNavExitMsg, got %T", msg)
	}
	if exitMsg.Reason != ChipNavExitTyping {
		t.Errorf("expected ChipNavExitTyping, got %v", exitMsg.Reason)
	}
	if exitMsg.Character != 'x' {
		t.Errorf("expected Character 'x', got '%c'", exitMsg.Character)
	}
}

func TestChipList_View_Empty(t *testing.T) {
	cl := NewChipList()
	view := cl.View()
	if view != "" {
		t.Errorf("expected empty view, got '%s'", view)
	}
}

func TestChipList_View_Normal(t *testing.T) {
	cl := NewChipList()
	cl.AddChip("backend")
	cl.AddChip("frontend")

	view := cl.View()
	if !strings.Contains(view, "backend") {
		t.Error("expected view to contain 'backend'")
	}
	if !strings.Contains(view, "frontend") {
		t.Error("expected view to contain 'frontend'")
	}
	// Pills use powerline characters instead of brackets
	if !strings.Contains(view, pillLeft) && !strings.Contains(view, pillRight) {
		t.Error("expected view to contain pill characters")
	}
}

func TestChipList_View_Highlighted(t *testing.T) {
	cl := NewChipList()
	cl.AddChip("backend")
	cl.AddChip("frontend")
	cl.EnterNavigation()
	// Move to first
	cl, _ = cl.Update(tea.KeyMsg{Type: tea.KeyLeft})

	view := cl.View()
	// Highlighted chip rendered with different background color (pill style)
	// Just verify the label is present - color styling can't be easily tested
	if !strings.Contains(view, "backend") {
		t.Errorf("expected highlighted chip to contain 'backend', got: %s", view)
	}
}

func TestChipList_View_Flash(t *testing.T) {
	cl := NewChipList()
	cl.AddChip("backend")
	cl.AddChip("backend") // duplicate triggers flash

	if cl.flashIndex != 0 {
		t.Errorf("expected flashIndex 0, got %d", cl.flashIndex)
	}

	view := cl.View()
	// Flash chip should be rendered (we can't easily check style, but it should exist)
	if !strings.Contains(view, "backend") {
		t.Error("expected view to contain 'backend'")
	}
}

func TestChipList_View_WordWrap(t *testing.T) {
	t.Run("Width30", func(t *testing.T) {
		cl := NewChipList().WithWidth(30)
		cl.AddChip("backend")
		cl.AddChip("frontend")
		cl.AddChip("api")
		cl.AddChip("urgent")

		view := cl.View()
		lines := strings.Split(view, "\n")
		if len(lines) < 2 {
			t.Errorf("expected multiple lines at width 30, got %d lines: %s", len(lines), view)
		}
	})

	t.Run("Width80", func(t *testing.T) {
		cl := NewChipList().WithWidth(80)
		cl.AddChip("a")
		cl.AddChip("b")
		cl.AddChip("c")

		view := cl.View()
		lines := strings.Split(view, "\n")
		if len(lines) != 1 {
			t.Errorf("expected 1 line at width 80 with short chips, got %d", len(lines))
		}
	})
}

func TestChipList_FocusBlur(t *testing.T) {
	cl := NewChipList()

	if cl.Focused() {
		t.Error("expected not focused initially")
	}

	cl.Focus()
	if !cl.Focused() {
		t.Error("expected focused after Focus()")
	}

	cl.Blur()
	if cl.Focused() {
		t.Error("expected not focused after Blur()")
	}
}

func TestChipList_Blur_ExitsNavigation(t *testing.T) {
	cl := NewChipList()
	cl.AddChip("chip")
	cl.Focus()
	cl.EnterNavigation()

	if cl.state != ChipListNavigation {
		t.Fatalf("expected navigation state, got %v", cl.state)
	}

	cl.Blur()

	if cl.state != ChipListInput {
		t.Errorf("expected ChipListInput after blur, got %v", cl.state)
	}
	if cl.navIndex != -1 {
		t.Errorf("expected navIndex -1 after blur, got %d", cl.navIndex)
	}
}

func TestChipList_RemoveHighlighted(t *testing.T) {
	t.Run("InNavigationMode", func(t *testing.T) {
		cl := NewChipList()
		cl.AddChip("first")
		cl.AddChip("second")
		cl.EnterNavigation()

		removed := cl.RemoveHighlighted()
		if removed != "second" {
			t.Errorf("expected 'second', got '%s'", removed)
		}
		if len(cl.Chips) != 1 {
			t.Errorf("expected 1 chip, got %d", len(cl.Chips))
		}
	})

	t.Run("NotInNavigationMode", func(t *testing.T) {
		cl := NewChipList()
		cl.AddChip("chip")

		removed := cl.RemoveHighlighted()
		if removed != "" {
			t.Errorf("expected empty string, got '%s'", removed)
		}
		if len(cl.Chips) != 1 {
			t.Errorf("expected 1 chip, got %d", len(cl.Chips))
		}
	})
}

func TestChipList_HighlightedIndex(t *testing.T) {
	cl := NewChipList()
	cl.AddChip("chip")

	if cl.HighlightedIndex() != -1 {
		t.Errorf("expected -1 when not in navigation, got %d", cl.HighlightedIndex())
	}

	cl.EnterNavigation()
	if cl.HighlightedIndex() != 0 {
		t.Errorf("expected 0 in navigation, got %d", cl.HighlightedIndex())
	}
}

func TestChipList_ExitNavigation(t *testing.T) {
	cl := NewChipList()
	cl.AddChip("chip")
	cl.EnterNavigation()

	cl.ExitNavigation()

	if cl.state != ChipListInput {
		t.Errorf("expected ChipListInput, got %v", cl.state)
	}
	if cl.navIndex != -1 {
		t.Errorf("expected navIndex -1, got %d", cl.navIndex)
	}
}

func TestChipList_FlashClear(t *testing.T) {
	cl := NewChipList()
	cl.AddChip("chip")
	cl.AddChip("chip") // duplicate sets flash

	if cl.flashIndex != 0 {
		t.Fatalf("expected flashIndex 0, got %d", cl.flashIndex)
	}

	// Simulate flash clear message
	cl, _ = cl.Update(chipFlashClearMsg{})

	if cl.flashIndex != -1 {
		t.Errorf("expected flashIndex -1 after clear, got %d", cl.flashIndex)
	}
}
