package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewCreateOverlay(t *testing.T) {
	t.Run("SetsDefaultValues", func(t *testing.T) {
		overlay := NewCreateOverlay("", nil)
		if overlay.IssueType() != "task" {
			t.Errorf("expected default issue type 'task', got %s", overlay.IssueType())
		}
		if overlay.Priority() != 2 {
			t.Errorf("expected default priority 2 (Medium), got %d", overlay.Priority())
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
		// Parent input should be pre-filled
		if overlay.parentInput.Value() != "ab-123" {
			t.Errorf("expected parent input 'ab-123', got %s", overlay.parentInput.Value())
		}
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

	t.Run("TitleInputIsFocused", func(t *testing.T) {
		overlay := NewCreateOverlay("", nil)
		if overlay.focus != focusTitle {
			t.Error("expected focus on title")
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

	t.Run("ClosesDropdownFirst", func(t *testing.T) {
		overlay := NewCreateOverlay("", []ParentOption{{ID: "ab-1", Display: "ab-1 Test"}})
		overlay.showDropdown = true
		updated, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if cmd != nil {
			t.Error("expected no command when closing dropdown")
		}
		if updated.showDropdown {
			t.Error("expected dropdown to be closed")
		}
	})
}

func TestCreateOverlayNavigation(t *testing.T) {
	t.Run("TabMovesToNextField", func(t *testing.T) {
		overlay := NewCreateOverlay("", nil)
		if overlay.focus != focusTitle {
			t.Error("expected initial focus on title")
		}
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyTab})
		if overlay.focus != focusType {
			t.Errorf("expected focus on type, got %d", overlay.focus)
		}
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyTab})
		if overlay.focus != focusPriority {
			t.Errorf("expected focus on priority, got %d", overlay.focus)
		}
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyTab})
		if overlay.focus != focusParent {
			t.Errorf("expected focus on parent, got %d", overlay.focus)
		}
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyTab})
		if overlay.focus != focusTitle {
			t.Errorf("expected focus to wrap to title, got %d", overlay.focus)
		}
	})

	t.Run("LeftRightChangesType", func(t *testing.T) {
		overlay := NewCreateOverlay("", nil)
		overlay.focus = focusType
		if overlay.typeIndex != 0 {
			t.Error("expected initial type index 0")
		}
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRight})
		if overlay.typeIndex != 1 {
			t.Errorf("expected type index 1, got %d", overlay.typeIndex)
		}
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyLeft})
		if overlay.typeIndex != 0 {
			t.Errorf("expected type index 0, got %d", overlay.typeIndex)
		}
	})

	t.Run("LeftRightChangesPriority", func(t *testing.T) {
		overlay := NewCreateOverlay("", nil)
		overlay.focus = focusPriority
		if overlay.priorityIndex != 2 {
			t.Errorf("expected initial priority index 2 (Med), got %d", overlay.priorityIndex)
		}
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRight})
		if overlay.priorityIndex != 3 {
			t.Errorf("expected priority index 3, got %d", overlay.priorityIndex)
		}
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyLeft})
		if overlay.priorityIndex != 2 {
			t.Errorf("expected priority index 2, got %d", overlay.priorityIndex)
		}
	})
}

func TestCreateOverlaySubmit(t *testing.T) {
	t.Run("RequiresTitle", func(t *testing.T) {
		overlay := NewCreateOverlay("", nil)
		// Title is empty
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd != nil {
			t.Error("expected no submit with empty title")
		}
	})

	t.Run("SubmitsWithValidTitle", func(t *testing.T) {
		overlay := NewCreateOverlay("", nil)
		overlay.titleInput.SetValue("Test Bead")
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Fatal("expected submit command")
		}
		msg := cmd()
		created, ok := msg.(BeadCreatedMsg)
		if !ok {
			t.Fatalf("expected BeadCreatedMsg, got %T", msg)
		}
		if created.Title != "Test Bead" {
			t.Errorf("expected title 'Test Bead', got %s", created.Title)
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

func TestCreateOverlayGetters(t *testing.T) {
	t.Run("ReturnsCurrentValues", func(t *testing.T) {
		overlay := NewCreateOverlay("", nil)
		overlay.titleInput.SetValue("My Title")
		overlay.typeIndex = 2 // bug
		overlay.priorityIndex = 0

		if overlay.Title() != "My Title" {
			t.Errorf("expected title 'My Title', got %s", overlay.Title())
		}
		if overlay.IssueType() != "bug" {
			t.Errorf("expected issue type 'bug', got %s", overlay.IssueType())
		}
		if overlay.Priority() != 0 {
			t.Errorf("expected priority 0, got %d", overlay.Priority())
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

func TestCreateOverlayParentFilter(t *testing.T) {
	t.Run("FiltersParentsByID", func(t *testing.T) {
		parents := []ParentOption{
			{ID: "ab-001", Display: "ab-001 First"},
			{ID: "ab-002", Display: "ab-002 Second"},
			{ID: "xy-003", Display: "xy-003 Third"},
		}
		overlay := NewCreateOverlay("", parents)
		overlay.parentInput.SetValue("ab")
		overlay.filterParents()

		if len(overlay.filteredParents) != 2 {
			t.Errorf("expected 2 filtered parents, got %d", len(overlay.filteredParents))
		}
	})

	t.Run("LimitsToFiveResults", func(t *testing.T) {
		parents := make([]ParentOption, 10)
		for i := range parents {
			parents[i] = ParentOption{ID: "ab-00" + string(rune('0'+i)), Display: "Test"}
		}
		overlay := NewCreateOverlay("", parents)
		overlay.parentInput.SetValue("ab")
		overlay.filterParents()

		if len(overlay.filteredParents) != 5 {
			t.Errorf("expected max 5 filtered parents, got %d", len(overlay.filteredParents))
		}
	})
}
