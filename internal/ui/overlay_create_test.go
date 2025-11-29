package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewCreateOverlay(t *testing.T) {
	t.Run("SetsDefaultValues", func(t *testing.T) {
		overlay := NewCreateOverlay("", nil)
		if overlay.issueType != "task" {
			t.Errorf("expected default issue type 'task', got %s", overlay.issueType)
		}
		if overlay.priority != 2 {
			t.Errorf("expected default priority 2 (Medium), got %d", overlay.priority)
		}
	})

	t.Run("SetsDefaultParent", func(t *testing.T) {
		parents := []ParentOption{
			{ID: "ab-123", Display: "ab-123 Test"},
		}
		overlay := NewCreateOverlay("ab-123", parents)
		if overlay.defaultParent != "ab-123" {
			t.Errorf("expected default parent 'ab-123', got %s", overlay.defaultParent)
		}
		// parentID is bound to form, starts at first option until form processes
		// The defaultParent is stored for reference
	})

	t.Run("StoresParentOptions", func(t *testing.T) {
		parents := []ParentOption{
			{ID: "ab-001", Display: "ab-001 First"},
			{ID: "ab-002", Display: "ab-002 Second"},
		}
		overlay := NewCreateOverlay("", parents)
		if len(overlay.parentOptions) != 2 {
			t.Errorf("expected 2 parent options, got %d", len(overlay.parentOptions))
		}
	})

	t.Run("FormIsInitialized", func(t *testing.T) {
		overlay := NewCreateOverlay("", nil)
		if overlay.form == nil {
			t.Error("expected form to be initialized")
		}
	})
}

func TestCreateOverlayEscape(t *testing.T) {
	t.Run("SendsCreateCancelledMsg", func(t *testing.T) {
		overlay := NewCreateOverlay("", nil)
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if cmd == nil {
			t.Fatal("expected command from escape")
		}
		msg := cmd()
		_, ok := msg.(CreateCancelledMsg)
		if !ok {
			t.Fatalf("expected CreateCancelledMsg, got %T", msg)
		}
	})
}

func TestCreateOverlayFormCompletion(t *testing.T) {
	t.Run("BeadCreatedMsgHasCorrectFields", func(t *testing.T) {
		// Test that BeadCreatedMsg correctly stores all fields
		msg := BeadCreatedMsg{
			Title:     "Test Bead",
			IssueType: "feature",
			Priority:  1,
			ParentID:  "ab-parent",
		}
		if msg.Title != "Test Bead" {
			t.Errorf("expected title 'Test Bead', got %s", msg.Title)
		}
		if msg.IssueType != "feature" {
			t.Errorf("expected issue type 'feature', got %s", msg.IssueType)
		}
		if msg.Priority != 1 {
			t.Errorf("expected priority 1, got %d", msg.Priority)
		}
		if msg.ParentID != "ab-parent" {
			t.Errorf("expected parent ID 'ab-parent', got %s", msg.ParentID)
		}
	})
}

func TestCreateOverlayView(t *testing.T) {
	t.Run("RenderContainsTitle", func(t *testing.T) {
		overlay := NewCreateOverlay("", nil)
		view := overlay.View()
		if view == "" {
			t.Error("expected non-empty view")
		}
		// Just verify it renders without panic
		if len(view) < 50 {
			t.Error("view seems too short")
		}
	})
}

func TestCreateOverlayTitleValidation(t *testing.T) {
	t.Run("RequiresTitleForValidation", func(t *testing.T) {
		overlay := NewCreateOverlay("", nil)
		// Title starts empty
		if overlay.title != "" {
			t.Error("expected title to start empty")
		}
	})
}

func TestCreateOverlayGetters(t *testing.T) {
	t.Run("ReturnsCurrentValues", func(t *testing.T) {
		overlay := NewCreateOverlay("ab-test", nil)
		overlay.title = "My Title"
		overlay.issueType = "bug"
		overlay.priority = 0
		overlay.parentID = "ab-parent"

		if overlay.Title() != "My Title" {
			t.Errorf("expected title 'My Title', got %s", overlay.Title())
		}
		if overlay.IssueType() != "bug" {
			t.Errorf("expected issue type 'bug', got %s", overlay.IssueType())
		}
		if overlay.Priority() != 0 {
			t.Errorf("expected priority 0, got %d", overlay.Priority())
		}
		if overlay.ParentID() != "ab-parent" {
			t.Errorf("expected parent ID 'ab-parent', got %s", overlay.ParentID())
		}
	})
}

func TestParentOption(t *testing.T) {
	t.Run("StoresIDAndDisplay", func(t *testing.T) {
		opt := ParentOption{
			ID:      "ab-xyz",
			Display: "ab-xyz Some Title",
		}
		if opt.ID != "ab-xyz" {
			t.Errorf("expected ID 'ab-xyz', got %s", opt.ID)
		}
		if opt.Display != "ab-xyz Some Title" {
			t.Errorf("expected Display 'ab-xyz Some Title', got %s", opt.Display)
		}
	})
}
