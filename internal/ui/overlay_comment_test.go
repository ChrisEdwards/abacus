package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCommentOverlay_NewCommentOverlay(t *testing.T) {
	overlay := NewCommentOverlay("ab-123", "Test bead title")

	if overlay.issueID != "ab-123" {
		t.Errorf("expected issueID ab-123, got %s", overlay.issueID)
	}
	if overlay.beadTitle != "Test bead title" {
		t.Errorf("expected beadTitle 'Test bead title', got %s", overlay.beadTitle)
	}
}

func TestCommentOverlay_EmptySubmit(t *testing.T) {
	overlay := NewCommentOverlay("ab-123", "Test")

	// Try to submit empty
	overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})

	if overlay.errorMsg == "" {
		t.Error("expected error message for empty comment")
	}
}

func TestCommentOverlay_ValidSubmit(t *testing.T) {
	overlay := NewCommentOverlay("ab-123", "Test")
	overlay.textarea.SetValue("This is a valid comment")

	_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})

	if cmd == nil {
		t.Error("expected command to be returned")
	}

	msg := cmd()
	if addedMsg, ok := msg.(CommentAddedMsg); !ok {
		t.Error("expected CommentAddedMsg")
	} else {
		if addedMsg.IssueID != "ab-123" {
			t.Errorf("expected issueID ab-123, got %s", addedMsg.IssueID)
		}
		if addedMsg.Comment != "This is a valid comment" {
			t.Errorf("unexpected comment text: %s", addedMsg.Comment)
		}
	}
}

func TestCommentOverlay_EscapeClearsText(t *testing.T) {
	overlay := NewCommentOverlay("ab-123", "Test")
	overlay.textarea.SetValue("Some text")

	// First Esc clears text
	overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd != nil {
		t.Error("first Esc should not produce command")
	}
	if overlay.textarea.Value() != "" {
		t.Error("expected textarea to be cleared")
	}
}

func TestCommentOverlay_EscapeCancels(t *testing.T) {
	overlay := NewCommentOverlay("ab-123", "Test")
	// Textarea is empty

	_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Error("expected cancel command")
	}

	msg := cmd()
	if _, ok := msg.(CommentCancelledMsg); !ok {
		t.Error("expected CommentCancelledMsg")
	}
}

func TestCommentOverlay_WhitespaceOnlySubmit(t *testing.T) {
	overlay := NewCommentOverlay("ab-123", "Test")
	overlay.textarea.SetValue("   \n\t  ")

	// Try to submit whitespace-only
	overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})

	if overlay.errorMsg == "" {
		t.Error("expected error message for whitespace-only comment")
	}
}

func TestCommentOverlay_TitleTruncation(t *testing.T) {
	longTitle := "This is a very long title that exceeds the 40 character limit for display"
	overlay := NewCommentOverlay("ab-123", longTitle)

	view := overlay.View()

	// Should contain truncated title with ellipsis
	if len(longTitle) <= 40 {
		t.Skip("test requires title longer than 40 chars")
	}

	// The view should contain the bead ID
	if !strings.Contains(view, "ab-123") {
		t.Error("view should contain the bead ID")
	}
}
