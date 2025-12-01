package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewDeleteOverlay(t *testing.T) {
	t.Run("StoresIssueIDAndTitle", func(t *testing.T) {
		overlay := NewDeleteOverlay("ab-123", "Test Issue Title")
		if overlay.issueID != "ab-123" {
			t.Errorf("expected issueID 'ab-123', got %q", overlay.issueID)
		}
		if overlay.issueTitle != "Test Issue Title" {
			t.Errorf("expected issueTitle 'Test Issue Title', got %q", overlay.issueTitle)
		}
	})

	t.Run("DefaultsToNo", func(t *testing.T) {
		overlay := NewDeleteOverlay("ab-123", "Test")
		if overlay.selected != 0 {
			t.Errorf("expected selected=0 (No), got %d", overlay.selected)
		}
	})
}

func TestDeleteOverlay_YKeyConfirms(t *testing.T) {
	overlay := NewDeleteOverlay("ab-xyz", "Test Issue")
	_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Fatal("expected command from 'y' key")
	}
	msg := cmd()
	confirmMsg, ok := msg.(DeleteConfirmedMsg)
	if !ok {
		t.Fatalf("expected DeleteConfirmedMsg, got %T", msg)
	}
	if confirmMsg.IssueID != "ab-xyz" {
		t.Errorf("expected IssueID 'ab-xyz', got %s", confirmMsg.IssueID)
	}
}

func TestDeleteOverlay_EnterConfirms(t *testing.T) {
	t.Run("EnterOnYesConfirms", func(t *testing.T) {
		overlay := NewDeleteOverlay("ab-123", "Test")
		overlay.selected = 1 // Move to Yes
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Fatal("expected command from enter on Yes")
		}
		msg := cmd()
		_, ok := msg.(DeleteConfirmedMsg)
		if !ok {
			t.Fatalf("expected DeleteConfirmedMsg, got %T", msg)
		}
	})

	t.Run("EnterOnNoCancels", func(t *testing.T) {
		overlay := NewDeleteOverlay("ab-123", "Test")
		overlay.selected = 0 // Stay on No (default)
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Fatal("expected command from enter on No")
		}
		msg := cmd()
		_, ok := msg.(DeleteCancelledMsg)
		if !ok {
			t.Fatalf("expected DeleteCancelledMsg, got %T", msg)
		}
	})
}

func TestDeleteOverlay_NKeyOrEscCancels(t *testing.T) {
	t.Run("NKeyCancels", func(t *testing.T) {
		overlay := NewDeleteOverlay("ab-123", "Test")
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
		if cmd == nil {
			t.Fatal("expected command from 'n' key")
		}
		msg := cmd()
		_, ok := msg.(DeleteCancelledMsg)
		if !ok {
			t.Fatalf("expected DeleteCancelledMsg, got %T", msg)
		}
	})

	t.Run("EscCancels", func(t *testing.T) {
		overlay := NewDeleteOverlay("ab-123", "Test")
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if cmd == nil {
			t.Fatal("expected command from escape")
		}
		msg := cmd()
		_, ok := msg.(DeleteCancelledMsg)
		if !ok {
			t.Fatalf("expected DeleteCancelledMsg, got %T", msg)
		}
	})
}

func TestDeleteOverlay_Navigation(t *testing.T) {
	t.Run("JMovesToYes", func(t *testing.T) {
		overlay := NewDeleteOverlay("ab-123", "Test")
		overlay.selected = 0
		overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if overlay.selected != 1 {
			t.Errorf("expected selected=1 after j, got %d", overlay.selected)
		}
	})

	t.Run("KMovesToNo", func(t *testing.T) {
		overlay := NewDeleteOverlay("ab-123", "Test")
		overlay.selected = 1
		overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if overlay.selected != 0 {
			t.Errorf("expected selected=0 after k, got %d", overlay.selected)
		}
	})

	t.Run("DownMovesToYes", func(t *testing.T) {
		overlay := NewDeleteOverlay("ab-123", "Test")
		overlay.selected = 0
		overlay.Update(tea.KeyMsg{Type: tea.KeyDown})
		if overlay.selected != 1 {
			t.Errorf("expected selected=1 after down, got %d", overlay.selected)
		}
	})

	t.Run("UpMovesToNo", func(t *testing.T) {
		overlay := NewDeleteOverlay("ab-123", "Test")
		overlay.selected = 1
		overlay.Update(tea.KeyMsg{Type: tea.KeyUp})
		if overlay.selected != 0 {
			t.Errorf("expected selected=0 after up, got %d", overlay.selected)
		}
	})

	t.Run("TabMovesToYes", func(t *testing.T) {
		overlay := NewDeleteOverlay("ab-123", "Test")
		overlay.selected = 0
		overlay.Update(tea.KeyMsg{Type: tea.KeyTab})
		if overlay.selected != 1 {
			t.Errorf("expected selected=1 after tab, got %d", overlay.selected)
		}
	})
}

func TestDeleteOverlay_View(t *testing.T) {
	t.Run("ContainsIssueID", func(t *testing.T) {
		overlay := NewDeleteOverlay("ab-xyz", "Test Issue")
		view := overlay.View()
		if !strings.Contains(view, "ab-xyz") {
			t.Error("expected view to contain issue ID")
		}
	})

	t.Run("ContainsDeleteLabel", func(t *testing.T) {
		overlay := NewDeleteOverlay("ab-123", "Test")
		view := overlay.View()
		if !strings.Contains(view, "Delete") {
			t.Error("expected view to contain 'Delete'")
		}
	})

	t.Run("ContainsWarning", func(t *testing.T) {
		overlay := NewDeleteOverlay("ab-123", "Test")
		view := overlay.View()
		if !strings.Contains(view, "cannot be undone") {
			t.Error("expected view to contain warning message")
		}
	})

	t.Run("ContainsOptions", func(t *testing.T) {
		overlay := NewDeleteOverlay("ab-123", "Test")
		view := overlay.View()
		if !strings.Contains(view, "[n]o") {
			t.Error("expected view to contain '[n]o' option")
		}
		if !strings.Contains(view, "[y]es") {
			t.Error("expected view to contain '[y]es' option")
		}
	})

	t.Run("ShowsSelectionIndicator", func(t *testing.T) {
		overlay := NewDeleteOverlay("ab-123", "Test")
		overlay.selected = 0 // No selected
		view := overlay.View()
		// The selected option should have an arrow indicator
		if !strings.Contains(view, "←") {
			t.Error("expected view to show selection indicator")
		}
	})
}
