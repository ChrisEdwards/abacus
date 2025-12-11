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

func TestDeleteOverlay_CancelKeys(t *testing.T) {
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

func TestDeleteOverlay_View(t *testing.T) {
	t.Run("ContainsDeletePrompt", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "Test")
		view := overlay.View()
		plain := stripANSI(view)
		if !strings.Contains(plain, "Delete this bead?") {
			t.Error("expected view to contain delete prompt")
		}
	})

	t.Run("ShowsBeadUnderPrompt", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-xyz", "Test Issue")
		view := overlay.View()
		plain := stripANSI(view)
		if !strings.Contains(plain, "Delete this bead?") {
			t.Error("expected delete prompt to be present")
		}
		if !strings.Contains(plain, "  ● ab-xyz") {
			t.Error("expected bead to appear with indentation")
		}
	})

	t.Run("ContainsWarning", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "Test")
		view := overlay.View()
		plain := stripANSI(view)
		if !strings.Contains(plain, "cannot be undone") {
			t.Error("expected view to contain warning message")
		}
	})

	t.Run("ContainsTitleOnce", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "My Important Task Title")
		view := overlay.View()
		plain := stripANSI(view)
		if strings.Count(plain, "My Important Task Title") != 1 {
			t.Error("expected view to contain issue title once")
		}
	})

	t.Run("TruncatesLongTitle", func(t *testing.T) {
		longTitle := "This is a very long title that should be truncated in the dialog"
		overlay := newTestDeleteOverlay("ab-123", longTitle)
		view := overlay.View()
		plain := stripANSI(view)
		// Should contain start of the title (first few words)
		if !strings.Contains(plain, "This is a") {
			t.Error("expected view to contain start of long title")
		}
		// Should be truncated with ellipsis
		if !strings.Contains(plain, "...") {
			t.Error("expected view to contain ellipsis for truncated title")
		}
	})

	t.Run("ContainsConfirmationPrompt", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "Test")
		view := overlay.View()
		plain := stripANSI(view)
		if !strings.Contains(plain, "Delete this bead?") {
			t.Error("expected view to contain confirmation prompt")
		}
	})

	t.Run("RendersFooterHints", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "Test")
		view := overlay.View()
		plain := stripANSI(view)
		if !strings.Contains(plain, "d") || !strings.Contains(plain, "c/esc") {
			t.Error("expected view to include keyboard hint footer")
		}
	})

	t.Run("HasBoxBorder", func(t *testing.T) {
		overlay := newTestDeleteOverlay("ab-123", "Test")
		view := overlay.View()
		plain := stripANSI(view)
		if !strings.Contains(plain, "╭") || !strings.Contains(plain, "╯") {
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
	if !strings.Contains(view, "This will also delete 2 children") {
		t.Fatalf("expected warning to mention children, got: %s", view)
	}
	if !strings.Contains(view, "└─ ") {
		t.Fatalf("expected child list to show tree markers, got: %s", view)
	}
}
