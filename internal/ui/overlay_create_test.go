package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewCreateOverlay(t *testing.T) {
	t.Run("SetsDefaultValues", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
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
		overlay := NewCreateOverlay(CreateOverlayOptions{
			DefaultParentID:  "ab-123",
			AvailableParents: parents,
		})
		if overlay.DefaultParentID() != "ab-123" {
			t.Errorf("expected default parent 'ab-123', got %s", overlay.DefaultParentID())
		}
		// ParentID() should return the ID via combo box
		if overlay.ParentID() != "ab-123" {
			t.Errorf("expected ParentID() 'ab-123', got %s", overlay.ParentID())
		}
	})

	t.Run("StoresParentOptions", func(t *testing.T) {
		parents := []ParentOption{
			{ID: "ab-001", Display: "ab-001 First"},
			{ID: "ab-002", Display: "ab-002 Second"},
		}
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableParents: parents,
		})
		if len(overlay.parentOptions) != 2 {
			t.Errorf("expected 2 parent options, got %d", len(overlay.parentOptions))
		}
	})

	t.Run("TitleInputIsFocused", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		if overlay.Focus() != FocusTitle {
			t.Error("expected focus on title")
		}
	})

	t.Run("RootModeIsSet", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			IsRootMode: true,
		})
		if !overlay.IsRootMode() {
			t.Error("expected root mode to be true")
		}
	})
}

func TestCreateOverlayEscape(t *testing.T) {
	t.Run("SendsCreateCancelledMsg", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
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

func TestCreateOverlayNavigation(t *testing.T) {
	t.Run("TabMovesToNextField", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		if overlay.Focus() != FocusTitle {
			t.Error("expected initial focus on title")
		}
		// Tab: Title -> Type
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyTab})
		if overlay.Focus() != FocusType {
			t.Errorf("expected focus on type, got %d", overlay.Focus())
		}
		// Tab: Type -> Priority
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyTab})
		if overlay.Focus() != FocusPriority {
			t.Errorf("expected focus on priority, got %d", overlay.Focus())
		}
		// Tab: Priority -> Labels
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyTab})
		if overlay.Focus() != FocusLabels {
			t.Errorf("expected focus on labels, got %d", overlay.Focus())
		}
		// Note: Labels to Assignee transition requires ChipComboBoxTabMsg
	})

	t.Run("ShiftTabFromTitleGoesToParent", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		if overlay.Focus() != FocusTitle {
			t.Error("expected initial focus on title")
		}
		// Shift+Tab: Title -> Parent
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
		if overlay.Focus() != FocusParent {
			t.Errorf("expected focus on parent, got %d", overlay.Focus())
		}
	})

	t.Run("UpDownChangesType", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusType
		if overlay.typeIndex != 0 {
			t.Error("expected initial type index 0")
		}
		// Down arrow increases index
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyDown})
		if overlay.typeIndex != 1 {
			t.Errorf("expected type index 1, got %d", overlay.typeIndex)
		}
		// Up arrow decreases index
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyUp})
		if overlay.typeIndex != 0 {
			t.Errorf("expected type index 0, got %d", overlay.typeIndex)
		}
	})

	t.Run("UpDownChangesPriority", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusPriority
		if overlay.priorityIndex != 2 {
			t.Errorf("expected initial priority index 2 (Med), got %d", overlay.priorityIndex)
		}
		// Down arrow increases index
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyDown})
		if overlay.priorityIndex != 3 {
			t.Errorf("expected priority index 3, got %d", overlay.priorityIndex)
		}
		// Up arrow decreases index
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyUp})
		if overlay.priorityIndex != 2 {
			t.Errorf("expected priority index 2, got %d", overlay.priorityIndex)
		}
	})

	t.Run("LeftRightNavigatesBetweenTypeAndPriority", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusType
		// Right arrow moves to priority
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRight})
		if overlay.Focus() != FocusPriority {
			t.Errorf("expected focus on priority, got %d", overlay.Focus())
		}
		// Left arrow moves back to type
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyLeft})
		if overlay.Focus() != FocusType {
			t.Errorf("expected focus on type, got %d", overlay.Focus())
		}
	})
}

func TestCreateOverlaySubmit(t *testing.T) {
	t.Run("RequiresTitle", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		// Title is empty
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd != nil {
			t.Error("expected no submit with empty title")
		}
	})

	t.Run("SubmitsWithValidTitle", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
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
		msg := BeadCreatedMsg{
			Title:     "Test Bead",
			IssueType: "feature",
			Priority:  1,
			ParentID:  "ab-parent",
			Labels:    []string{"backend", "urgent"},
			Assignee:  "alice",
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
		if len(msg.Labels) != 2 || msg.Labels[0] != "backend" {
			t.Errorf("expected labels [backend, urgent], got %v", msg.Labels)
		}
		if msg.Assignee != "alice" {
			t.Errorf("expected assignee 'alice', got %s", msg.Assignee)
		}
	})
}

func TestCreateOverlayView(t *testing.T) {
	t.Run("RenderContainsTitle", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		view := overlay.View()
		if view == "" {
			t.Error("expected non-empty view")
		}
		if len(view) < 50 {
			t.Error("view seems too short")
		}
	})

	t.Run("ContainsZoneLabels", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		view := overlay.View()
		// Should contain zone headers
		if !contains(view, "PARENT") {
			t.Error("expected view to contain PARENT zone")
		}
		if !contains(view, "TITLE") {
			t.Error("expected view to contain TITLE zone")
		}
		if !contains(view, "PROPERTIES") {
			t.Error("expected view to contain PROPERTIES zone")
		}
		if !contains(view, "LABELS") {
			t.Error("expected view to contain LABELS zone")
		}
		if !contains(view, "ASSIGNEE") {
			t.Error("expected view to contain ASSIGNEE zone")
		}
	})

	t.Run("ShowsRootModeIndicator", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			IsRootMode: true,
		})
		view := overlay.View()
		if !contains(view, "No Parent") {
			t.Error("expected view to show 'No Parent' for root mode")
		}
	})
}

func TestCreateOverlayGetters(t *testing.T) {
	t.Run("ReturnsCurrentValues", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
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

func TestCreateOverlayTypeHotkeys(t *testing.T) {
	t.Run("TypeHotkeysWork", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusType

		// Press 'f' for feature
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
		if overlay.typeIndex != 1 {
			t.Errorf("expected type index 1 (feature), got %d", overlay.typeIndex)
		}

		// Press 'b' for bug
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
		if overlay.typeIndex != 2 {
			t.Errorf("expected type index 2 (bug), got %d", overlay.typeIndex)
		}

		// Press 'e' for epic
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
		if overlay.typeIndex != 3 {
			t.Errorf("expected type index 3 (epic), got %d", overlay.typeIndex)
		}

		// Press 'c' for chore
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
		if overlay.typeIndex != 4 {
			t.Errorf("expected type index 4 (chore), got %d", overlay.typeIndex)
		}

		// Press 't' for task
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
		if overlay.typeIndex != 0 {
			t.Errorf("expected type index 0 (task), got %d", overlay.typeIndex)
		}
	})
}

func TestCreateOverlayPriorityHotkeys(t *testing.T) {
	t.Run("PriorityHotkeysWork", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusPriority

		// Press 'c' for critical
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
		if overlay.priorityIndex != 0 {
			t.Errorf("expected priority index 0 (critical), got %d", overlay.priorityIndex)
		}

		// Press 'h' for high
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
		if overlay.priorityIndex != 1 {
			t.Errorf("expected priority index 1 (high), got %d", overlay.priorityIndex)
		}

		// Press 'l' for low
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
		if overlay.priorityIndex != 3 {
			t.Errorf("expected priority index 3 (low), got %d", overlay.priorityIndex)
		}

		// Press 'b' for backlog
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
		if overlay.priorityIndex != 4 {
			t.Errorf("expected priority index 4 (backlog), got %d", overlay.priorityIndex)
		}

		// Press 'm' for medium
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
		if overlay.priorityIndex != 2 {
			t.Errorf("expected priority index 2 (medium), got %d", overlay.priorityIndex)
		}
	})
}

func TestChipComboBoxTabMsg(t *testing.T) {
	t.Run("LabelsTabMovesToAssignee", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusLabels

		// Simulate ChipComboBoxTabMsg from labels combo
		overlay, _ = overlay.Update(ChipComboBoxTabMsg{})
		if overlay.Focus() != FocusAssignee {
			t.Errorf("expected focus on assignee after labels Tab, got %d", overlay.Focus())
		}
	})
}

// Helper function
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
