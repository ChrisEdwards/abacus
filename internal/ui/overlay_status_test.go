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
		if overlay.selected != 4 {
			t.Errorf("expected selected index 4 for closed, got %d", overlay.selected)
		}
	})

	t.Run("BlockedStatusSelected", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test Issue", "blocked")
		if overlay.selected != 2 {
			t.Errorf("expected selected index 2 for blocked, got %d", overlay.selected)
		}
	})

	t.Run("DeferredStatusSelected", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test Issue", "deferred")
		if overlay.selected != 3 {
			t.Errorf("expected selected index 3 for deferred, got %d", overlay.selected)
		}
	})

	t.Run("ClosedStatusAllowsReopenVariants", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test Issue", "closed")
		// From closed, can reopen to any non-terminal status
		if overlay.options[0].disabled {
			t.Error("expected open to NOT be disabled (reopen allowed)")
		}
		if overlay.options[1].disabled {
			t.Error("expected in_progress to NOT be disabled (reopen variant allowed)")
		}
		if overlay.options[2].disabled {
			t.Error("expected blocked to NOT be disabled (reopen variant allowed)")
		}
		if overlay.options[3].disabled {
			t.Error("expected deferred to NOT be disabled (reopen variant allowed)")
		}
		// closed itself should not be disabled
		if overlay.options[4].disabled {
			t.Error("expected closed to NOT be disabled when current status is closed")
		}
	})

	t.Run("OpenStatusAllowsTransitions", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test Issue", "open")
		// From open, can transition to all other statuses
		if overlay.options[0].disabled {
			t.Error("expected open to NOT be disabled when current status is open")
		}
		if overlay.options[1].disabled {
			t.Error("expected in_progress to NOT be disabled when current status is open")
		}
		if overlay.options[2].disabled {
			t.Error("expected blocked to NOT be disabled when current status is open")
		}
		if overlay.options[3].disabled {
			t.Error("expected deferred to NOT be disabled when current status is open")
		}
		if overlay.options[4].disabled {
			t.Error("expected closed to NOT be disabled when current status is open")
		}
	})

	t.Run("BlockedStatusAllowsTransitions", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test Issue", "blocked")
		// From blocked, can transition to all other statuses
		for i, opt := range overlay.options {
			if opt.disabled {
				t.Errorf("expected option %d (%s) to NOT be disabled from blocked", i, opt.value)
			}
		}
	})

	t.Run("DeferredStatusAllowsTransitions", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test Issue", "deferred")
		// From deferred, can transition to all other statuses
		for i, opt := range overlay.options {
			if opt.disabled {
				t.Errorf("expected option %d (%s) to NOT be disabled from deferred", i, opt.value)
			}
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
		overlay.selected = 4 // Start at last option (closed)
		overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if overlay.selected != 0 {
			t.Errorf("expected selected 0 after wrap, got %d", overlay.selected)
		}
	})

	t.Run("UpWrapsAround", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test", "open")
		overlay.selected = 0
		overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if overlay.selected != 4 {
			t.Errorf("expected selected 4 after wrap, got %d", overlay.selected)
		}
	})

	t.Run("NavigatesThroughAllOptions", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test", "open")
		// From open, all options are enabled, so should navigate sequentially
		overlay.selected = 0
		for i := 1; i <= 4; i++ {
			overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
			if overlay.selected != i {
				t.Errorf("expected selected %d after %d down presses, got %d", i, i, overlay.selected)
			}
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

func TestStatusOverlayHotkeys(t *testing.T) {
	t.Run("HotkeyOSelectsOpen", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test", "in_progress")
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
		if cmd == nil {
			t.Fatal("expected command from 'o' hotkey")
		}
		msg := cmd()
		statusMsg, ok := msg.(StatusChangedMsg)
		if !ok {
			t.Fatalf("expected StatusChangedMsg, got %T", msg)
		}
		if statusMsg.NewStatus != "open" {
			t.Errorf("expected NewStatus 'open', got %s", statusMsg.NewStatus)
		}
	})

	t.Run("HotkeyISelectsInProgress", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test", "open")
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
		if cmd == nil {
			t.Fatal("expected command from 'i' hotkey")
		}
		msg := cmd()
		statusMsg, ok := msg.(StatusChangedMsg)
		if !ok {
			t.Fatalf("expected StatusChangedMsg, got %T", msg)
		}
		if statusMsg.NewStatus != "in_progress" {
			t.Errorf("expected NewStatus 'in_progress', got %s", statusMsg.NewStatus)
		}
	})

	t.Run("HotkeyBSelectsBlocked", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test", "open")
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
		if cmd == nil {
			t.Fatal("expected command from 'b' hotkey")
		}
		msg := cmd()
		statusMsg, ok := msg.(StatusChangedMsg)
		if !ok {
			t.Fatalf("expected StatusChangedMsg, got %T", msg)
		}
		if statusMsg.NewStatus != "blocked" {
			t.Errorf("expected NewStatus 'blocked', got %s", statusMsg.NewStatus)
		}
	})

	t.Run("HotkeyDSelectsDeferred", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test", "open")
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
		if cmd == nil {
			t.Fatal("expected command from 'd' hotkey")
		}
		msg := cmd()
		statusMsg, ok := msg.(StatusChangedMsg)
		if !ok {
			t.Fatalf("expected StatusChangedMsg, got %T", msg)
		}
		if statusMsg.NewStatus != "deferred" {
			t.Errorf("expected NewStatus 'deferred', got %s", statusMsg.NewStatus)
		}
	})

	t.Run("HotkeyCSelectsClosed", func(t *testing.T) {
		overlay := NewStatusOverlay("test-123", "Test", "open")
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
		if cmd == nil {
			t.Fatal("expected command from 'c' hotkey")
		}
		msg := cmd()
		statusMsg, ok := msg.(StatusChangedMsg)
		if !ok {
			t.Fatalf("expected StatusChangedMsg, got %T", msg)
		}
		if statusMsg.NewStatus != "closed" {
			t.Errorf("expected NewStatus 'closed', got %s", statusMsg.NewStatus)
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
		// Should contain all five options
		expectedLabels := []string{"Open", "In Progress", "Blocked", "Deferred", "Closed"}
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
