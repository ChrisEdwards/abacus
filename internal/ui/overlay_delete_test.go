package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func newTestDeleteOverlay(issueID, title string) *DeleteOverlay {
	return NewDeleteOverlay(issueID, title, nil, nil)
}

func TestNewDeleteOverlay(t *testing.T) {
	t.Run("StoresIssueIDAndTitle", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "Test Issue Title")
		if overlay.issueID != "ab-123" {
			t.Errorf("expected issueID 'ab-123', got %q", overlay.issueID)
		}
		if overlay.issueTitle != "Test Issue Title" {
			t.Errorf("expected issueTitle 'Test Issue Title', got %q", overlay.issueTitle)
		}
	})

	t.Run("DefaultsToNo", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "Test")
		if overlay.selected != 0 {
			t.Errorf("expected selected=0 (No), got %d", overlay.selected)
		}
	})
}

func TestDeleteOverlay_YKeyConfirms(t *testing.T) {
	overlay := newTestDeleteOverlay("ab-xyz", "Test Issue")
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
	if confirmMsg.Cascade {
		t.Errorf("expected Cascade false, got true")
	}
	if len(confirmMsg.Children) != 0 {
		t.Errorf("expected no children, got %d", len(confirmMsg.Children))
	}
}

func TestDeleteOverlay_DKeyConfirms(t *testing.T) {
	overlay := newTestDeleteOverlay("ab-xyz", "Test Issue")
	_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if cmd == nil {
		t.Fatal("expected command from 'd' key")
	}
	msg := cmd()
	confirmMsg, ok := msg.(DeleteConfirmedMsg)
	if !ok {
		t.Fatalf("expected DeleteConfirmedMsg, got %T", msg)
	}
	if confirmMsg.IssueID != "ab-xyz" {
		t.Errorf("expected IssueID 'ab-xyz', got %s", confirmMsg.IssueID)
	}
	if confirmMsg.Cascade {
		t.Errorf("expected Cascade false, got true")
	}
	if len(confirmMsg.Children) != 0 {
		t.Errorf("expected no children, got %d", len(confirmMsg.Children))
	}
}

func TestDeleteOverlay_CascadeWhenChildren(t *testing.T) {
	children := []ChildInfo{{ID: "ab-child", Title: "Child One", Depth: 0}}
	overlay := NewDeleteOverlay("ab-parent", "Parent", children, []string{"ab-child"})
	_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if cmd == nil {
		t.Fatal("expected command from 'd' key")
	}
	msg := cmd()
	confirmMsg, ok := msg.(DeleteConfirmedMsg)
	if !ok {
		t.Fatalf("expected DeleteConfirmedMsg, got %T", msg)
	}
	if !confirmMsg.Cascade {
		t.Fatalf("expected Cascade true for parent with children")
	}
	if len(confirmMsg.Children) != 1 || confirmMsg.Children[0] != "ab-child" {
		t.Fatalf("expected child IDs returned, got %v", confirmMsg.Children)
	}
}

func TestDeleteOverlay_EnterConfirms(t *testing.T) {
	t.Run("EnterOnYesConfirms", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "Test")
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
		overlay := newTestDeleteOverlay("ab-123", "Test")
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

func TestDeleteOverlay_CancelKeys(t *testing.T) {
	t.Run("NKeyCancels", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "Test")
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

	t.Run("CKeyCancels", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "Test")
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
		if cmd == nil {
			t.Fatal("expected command from 'c' key")
		}
		msg := cmd()
		_, ok := msg.(DeleteCancelledMsg)
		if !ok {
			t.Fatalf("expected DeleteCancelledMsg, got %T", msg)
		}
	})

	t.Run("EscCancels", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "Test")
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
		overlay := newTestDeleteOverlay("ab-123", "Test")
		overlay.selected = 0
		overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if overlay.selected != 1 {
			t.Errorf("expected selected=1 after j, got %d", overlay.selected)
		}
	})

	t.Run("KMovesToNo", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "Test")
		overlay.selected = 1
		overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if overlay.selected != 0 {
			t.Errorf("expected selected=0 after k, got %d", overlay.selected)
		}
	})

	t.Run("DownMovesToYes", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "Test")
		overlay.selected = 0
		overlay.Update(tea.KeyMsg{Type: tea.KeyDown})
		if overlay.selected != 1 {
			t.Errorf("expected selected=1 after down, got %d", overlay.selected)
		}
	})

	t.Run("UpMovesToNo", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "Test")
		overlay.selected = 1
		overlay.Update(tea.KeyMsg{Type: tea.KeyUp})
		if overlay.selected != 0 {
			t.Errorf("expected selected=0 after up, got %d", overlay.selected)
		}
	})

	t.Run("TabMovesToYes", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "Test")
		overlay.selected = 0
		overlay.Update(tea.KeyMsg{Type: tea.KeyTab})
		if overlay.selected != 1 {
			t.Errorf("expected selected=1 after tab, got %d", overlay.selected)
		}
	})
}

func TestDeleteOverlay_View(t *testing.T) {
	t.Run("ContainsIssueID", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-xyz", "Test Issue")
		view := overlay.View()
		if !strings.Contains(view, "ab-xyz") {
			t.Error("expected view to contain issue ID")
		}
	})

	t.Run("ContainsDeleteLabel", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "Test")
		view := overlay.View()
		if !strings.Contains(view, "Delete") {
			t.Error("expected view to contain 'Delete'")
		}
	})

	t.Run("ContainsWarning", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "Test")
		view := overlay.View()
		if !strings.Contains(view, "cannot be undone") {
			t.Error("expected view to contain warning message")
		}
	})

	t.Run("ContainsTitle", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "My Important Task Title")
		view := overlay.View()
		if !strings.Contains(view, "My Important Task Title") {
			t.Error("expected view to contain issue title")
		}
	})

	t.Run("TruncatesLongTitle", func(t *testing.T) {
		longTitle := "This is a very long title that should be truncated in the dialog"
		overlay := newTestDeleteOverlay("ab-123", longTitle)
		view := overlay.View()
		// Should contain start of the title (first few words)
		if !strings.Contains(view, "This is a") {
			t.Error("expected view to contain start of long title")
		}
		// Should be truncated with ellipsis
		if !strings.Contains(view, "...") {
			t.Error("expected view to contain ellipsis for truncated title")
		}
	})

	t.Run("ContainsButtons", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "Test")
		view := overlay.View()
		if !strings.Contains(view, "Cancel") || !strings.Contains(view, "Delete") {
			t.Error("expected view to contain 'Cancel' and 'Delete' buttons")
		}
	})

	t.Run("ShowsDeleteBeadTitle", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "Test")
		view := overlay.View()
		if !strings.Contains(view, "Delete Bead") {
			t.Error("expected view to show 'Delete Bead' title")
		}
	})

	t.Run("ContainsConfirmationPrompt", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "Test")
		view := overlay.View()
		if !strings.Contains(view, "Are you sure you want to delete") {
			t.Error("expected view to contain confirmation prompt")
		}
	})

	t.Run("HasBoxBorder", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "Test")
		view := overlay.View()
		if !strings.Contains(view, "╭") || !strings.Contains(view, "╯") {
			t.Error("expected view to have box border characters")
		}
	})
}

func TestDeleteOverlay_ViewShowsChildWarning(t *testing.T) {
	children := []ChildInfo{
		{ID: "ab-1", Title: "Child One", Depth: 0},
		{ID: "ab-2", Title: "Nested", Depth: 1},
	}
	overlay := NewDeleteOverlay("ab-parent", "Parent", children, []string{"ab-1", "ab-2"})
	view := overlay.View()
	if !strings.Contains(view, "Delete All (3)") {
		t.Fatalf("expected view to include Delete All label, got: %s", view)
	}
	if !strings.Contains(view, "This bead has 2 children") {
		t.Fatalf("expected warning to mention children, got: %s", view)
	}
	if !strings.Contains(view, "└─ ") {
		t.Fatalf("expected child list to show tree markers, got: %s", view)
	}
}
