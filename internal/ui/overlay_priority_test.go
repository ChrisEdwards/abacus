package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewPriorityOverlay(t *testing.T) {
	t.Run("PreSelectsCurrentPriority", func(t *testing.T) {
		overlay := NewPriorityOverlay("test-123", "Test Issue", 1)
		if overlay.selected != 1 {
			t.Errorf("expected selected index 1 for P1, got %d", overlay.selected)
		}
	})

	t.Run("P0Selected", func(t *testing.T) {
		overlay := NewPriorityOverlay("test-123", "Test", 0)
		if overlay.selected != 0 {
			t.Errorf("expected selected 0 for P0, got %d", overlay.selected)
		}
	})

	t.Run("P2Selected", func(t *testing.T) {
		overlay := NewPriorityOverlay("test-123", "Test", 2)
		if overlay.selected != 2 {
			t.Errorf("expected selected 2 for P2, got %d", overlay.selected)
		}
	})

	t.Run("P3Selected", func(t *testing.T) {
		overlay := NewPriorityOverlay("test-123", "Test", 3)
		if overlay.selected != 3 {
			t.Errorf("expected selected 3 for P3, got %d", overlay.selected)
		}
	})

	t.Run("P4Selected", func(t *testing.T) {
		overlay := NewPriorityOverlay("test-123", "Test", 4)
		if overlay.selected != 4 {
			t.Errorf("expected selected 4 for P4, got %d", overlay.selected)
		}
	})

	t.Run("OutOfRangeNegativeDefaultsToP2", func(t *testing.T) {
		overlay := NewPriorityOverlay("test-123", "Test", -1)
		if overlay.selected != 2 {
			t.Errorf("expected selected 2 (P2) for invalid priority -1, got %d", overlay.selected)
		}
	})

	t.Run("OutOfRangeHighDefaultsToP2", func(t *testing.T) {
		overlay := NewPriorityOverlay("test-123", "Test", 99)
		if overlay.selected != 2 {
			t.Errorf("expected selected 2 (P2) for invalid priority 99, got %d", overlay.selected)
		}
	})

	t.Run("StoresIssueTitle", func(t *testing.T) {
		overlay := NewPriorityOverlay("test-123", "My Task", 2)
		if overlay.issueTitle != "My Task" {
			t.Errorf("expected issueTitle 'My Task', got %q", overlay.issueTitle)
		}
	})
}

func TestPriorityOverlayNavigation(t *testing.T) {
	t.Run("DownMovesToNext", func(t *testing.T) {
		overlay := NewPriorityOverlay("test-123", "Test", 0)
		overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if overlay.selected != 1 {
			t.Errorf("expected selected 1 after j, got %d", overlay.selected)
		}
	})

	t.Run("UpMovesToPrevious", func(t *testing.T) {
		overlay := NewPriorityOverlay("test-123", "Test", 1)
		overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if overlay.selected != 0 {
			t.Errorf("expected selected 0 after k, got %d", overlay.selected)
		}
	})

	t.Run("DownWrapsAround", func(t *testing.T) {
		overlay := NewPriorityOverlay("test-123", "Test", 4)
		overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if overlay.selected != 0 {
			t.Errorf("expected selected 0 after wrap, got %d", overlay.selected)
		}
	})

	t.Run("UpWrapsAround", func(t *testing.T) {
		overlay := NewPriorityOverlay("test-123", "Test", 0)
		overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if overlay.selected != 4 {
			t.Errorf("expected selected 4 after wrap, got %d", overlay.selected)
		}
	})

	t.Run("DownArrowAlsoWorks", func(t *testing.T) {
		overlay := NewPriorityOverlay("test-123", "Test", 0)
		overlay.Update(tea.KeyMsg{Type: tea.KeyDown})
		if overlay.selected != 1 {
			t.Errorf("expected selected 1 after down arrow, got %d", overlay.selected)
		}
	})

	t.Run("UpArrowAlsoWorks", func(t *testing.T) {
		overlay := NewPriorityOverlay("test-123", "Test", 1)
		overlay.Update(tea.KeyMsg{Type: tea.KeyUp})
		if overlay.selected != 0 {
			t.Errorf("expected selected 0 after up arrow, got %d", overlay.selected)
		}
	})
}

func TestPriorityOverlayEnter(t *testing.T) {
	t.Run("SendsPriorityChangedMsg", func(t *testing.T) {
		overlay := NewPriorityOverlay("test-123", "Test", 2)
		overlay.selected = 1 // P1 High
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Fatal("expected command from enter")
		}
		msg := cmd()
		pMsg, ok := msg.(PriorityChangedMsg)
		if !ok {
			t.Fatalf("expected PriorityChangedMsg, got %T", msg)
		}
		if pMsg.IssueID != "test-123" {
			t.Errorf("expected IssueID test-123, got %s", pMsg.IssueID)
		}
		if pMsg.NewPriority != 1 {
			t.Errorf("expected NewPriority 1, got %d", pMsg.NewPriority)
		}
	})

	t.Run("EnterOnDifferentSelection", func(t *testing.T) {
		overlay := NewPriorityOverlay("test-456", "Test", 0)
		overlay.selected = 4 // P4 Backlog
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Fatal("expected command from enter")
		}
		pMsg, ok := cmd().(PriorityChangedMsg)
		if !ok {
			t.Fatalf("expected PriorityChangedMsg, got %T", cmd())
		}
		if pMsg.NewPriority != 4 {
			t.Errorf("expected NewPriority 4, got %d", pMsg.NewPriority)
		}
	})
}

func TestPriorityOverlayEscape(t *testing.T) {
	t.Run("SendsPriorityCancelledMsg", func(t *testing.T) {
		overlay := NewPriorityOverlay("test-123", "Test", 2)
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if cmd == nil {
			t.Fatal("expected command from escape")
		}
		if _, ok := cmd().(PriorityCancelledMsg); !ok {
			t.Fatalf("expected PriorityCancelledMsg, got %T", cmd())
		}
	})
}

func TestPriorityOverlayView(t *testing.T) {
	t.Run("ContainsAllFiveLabels", func(t *testing.T) {
		overlay := NewPriorityOverlay("test-123", "Test", 2)
		view := overlay.View()
		for _, label := range []string{"P0", "P1", "P2", "P3", "P4"} {
			if !strings.Contains(view, label) {
				t.Errorf("expected view to contain %q", label)
			}
		}
		for _, name := range []string{"Critical", "High", "Medium", "Low", "Backlog"} {
			if !strings.Contains(view, name) {
				t.Errorf("expected view to contain %q", name)
			}
		}
	})

	t.Run("CurrentPriorityMarkedWithFilledCircle", func(t *testing.T) {
		overlay := NewPriorityOverlay("test-123", "Test", 1)
		view := overlay.View()
		if !strings.Contains(view, "●") {
			t.Error("expected view to contain ● for current priority")
		}
		if !strings.Contains(view, "○") {
			t.Error("expected view to contain ○ for non-current priorities")
		}
	})

	t.Run("IssueIDInHeader", func(t *testing.T) {
		overlay := NewPriorityOverlay("ab-xyz", "Test", 2)
		view := overlay.View()
		if !strings.Contains(view, "ab-xyz") {
			t.Errorf("expected view to contain issue ID 'ab-xyz'")
		}
	})
}

func TestPriorityOverlayNumericHotkeys(t *testing.T) {
	cases := []struct {
		key      rune
		priority int
	}{
		{'0', 0},
		{'1', 1},
		{'2', 2},
		{'3', 3},
		{'4', 4},
	}
	for _, c := range cases {
		t.Run(string(c.key), func(t *testing.T) {
			overlay := NewPriorityOverlay("test-123", "Test", 2)
			_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{c.key}})
			if cmd == nil {
				t.Fatalf("expected command from %q hotkey", c.key)
			}
			pMsg, ok := cmd().(PriorityChangedMsg)
			if !ok {
				t.Fatalf("expected PriorityChangedMsg, got %T", cmd())
			}
			if pMsg.NewPriority != c.priority {
				t.Errorf("expected NewPriority %d, got %d", c.priority, pMsg.NewPriority)
			}
			if pMsg.IssueID != "test-123" {
				t.Errorf("expected IssueID test-123, got %s", pMsg.IssueID)
			}
		})
	}
}
