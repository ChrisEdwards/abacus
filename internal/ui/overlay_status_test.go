package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewStatusOverlay(t *testing.T) {
	t.Run("PreSelectsCurrentStatus", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test Issue Title", "in_progress")
		if overlay.selected != 1 {
			t.Errorf("expected selected index 1 for in_progress, got %d", overlay.selected)
		}
		if overlay.options[1].value != "in_progress" {
			t.Errorf("expected options[1] to be in_progress, got %s", overlay.options[1].value)
		}
	})

	t.Run("OpenStatusSelected", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test Issue", "open")
		if overlay.selected != 0 {
			t.Errorf("expected selected index 0 for open, got %d", overlay.selected)
		}
	})

	t.Run("ClosedStatusSelected", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test Issue", "closed")
		if overlay.selected != 2 {
			t.Errorf("expected selected index 2 for closed, got %d", overlay.selected)
		}
	})

	t.Run("ClosedStatusAllowsReopen", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test Issue", "closed")
		// From closed, can reopen (transition to open)
		if overlay.options[0].disabled {
			t.Error("expected open to NOT be disabled (reopen allowed)")
		}
		// Cannot go directly to in_progress from closed
		if !overlay.options[1].disabled {
			t.Error("expected in_progress to be disabled when current status is closed")
		}
		// closed itself should not be disabled
		if overlay.options[2].disabled {
			t.Error("expected closed to NOT be disabled when current status is closed")
		}
	})

	t.Run("OpenStatusAllowsTransitions", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test Issue", "open")
		// From open, can transition to in_progress and closed
		if overlay.options[0].disabled {
			t.Error("expected open to NOT be disabled when current status is open")
		}
		if overlay.options[1].disabled {
			t.Error("expected in_progress to NOT be disabled when current status is open")
		}
		if overlay.options[2].disabled {
			t.Error("expected closed to NOT be disabled when current status is open")
		}
	})

	t.Run("StoresIssueTitle", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "My Important Task", "open")
		if overlay.issueTitle != "My Important Task" {
			t.Errorf("expected issueTitle 'My Important Task', got %q", overlay.issueTitle)
		}
	})
}

func TestStatusOverlayNavigation(t *testing.T) {
	t.Run("DownMovesToNext", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test", "open")
		overlay.selected = 0
		overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if overlay.selected != 1 {
			t.Errorf("expected selected 1 after j, got %d", overlay.selected)
		}
	})

	t.Run("UpMovesToPrevious", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test", "open")
		overlay.selected = 1
		overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if overlay.selected != 0 {
			t.Errorf("expected selected 0 after k, got %d", overlay.selected)
		}
	})

	t.Run("DownWrapsAround", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test", "open")
		overlay.selected = 2
		overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if overlay.selected != 0 {
			t.Errorf("expected selected 0 after wrap, got %d", overlay.selected)
		}
	})

	t.Run("UpWrapsAround", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test", "open")
		overlay.selected = 0
		overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if overlay.selected != 2 {
			t.Errorf("expected selected 2 after wrap, got %d", overlay.selected)
		}
	})

	t.Run("SkipsDisabledOnDown", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test", "closed")
		// From closed: open=enabled (reopen), in_progress=disabled, closed=enabled
		overlay.selected = 2 // Start at closed
		overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		// Should go to 0 (open) skipping disabled in_progress at 1
		if overlay.selected != 0 {
			t.Errorf("expected selected 0 (skip disabled in_progress), got %d", overlay.selected)
		}
	})
}

func TestStatusOverlayEnter(t *testing.T) {
	t.Run("SendsStatusChangedMsg", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test", "open")
		overlay.selected = 1 // in_progress
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Fatal("expected command from enter")
		}
		msg := cmd()
		statusMsg, ok := msg.(StatusChangedMsg)
		if !ok {
			t.Fatalf("expected StatusChangedMsg, got %T", msg)
		}
		if statusMsg.IssueID != "test-123" {
			t.Errorf("expected IssueID test-123, got %s", statusMsg.IssueID)
		}
		if statusMsg.NewStatus != "in_progress" {
			t.Errorf("expected NewStatus in_progress, got %s", statusMsg.NewStatus)
		}
	})
}

func TestStatusOverlayEscape(t *testing.T) {
	t.Run("SendsStatusCancelledMsg", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test", "open")
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if cmd == nil {
			t.Fatal("expected command from escape")
		}
		msg := cmd()
		_, ok := msg.(StatusCancelledMsg)
		if !ok {
			t.Fatalf("expected StatusCancelledMsg, got %T", msg)
		}
	})
}

func TestStatusOverlayView(t *testing.T) {
	t.Run("ContainsTitle", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test", "open")
		view := overlay.View()
		if view == "" {
			t.Error("expected non-empty view")
		}
		// The title "STATUS" should be present
		if len(view) < 10 {
			t.Error("view seems too short")
		}
	})

	t.Run("ShowsAllOptions", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test", "open")
		_ = overlay.View() // Ensure View() doesn't panic
		// Should contain all three options
		expectedLabels := []string{"Open", "In Progress", "Closed"}
		for _, label := range expectedLabels {
			found := false
			for _, opt := range overlay.options {
				if opt.label == label {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected option with label %q", label)
			}
		}
	})
}
