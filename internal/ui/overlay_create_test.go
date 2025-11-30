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
		// Title is empty - should flash red and return flash command (not submit)
		overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Error("expected flash command with empty title")
		}
		// Should NOT be a BeadCreatedMsg (i.e., no submit)
		msg := cmd()
		if _, ok := msg.(BeadCreatedMsg); ok {
			t.Error("expected no submit (BeadCreatedMsg) with empty title")
		}
		// Should show validation error
		if !overlay.TitleValidationError() {
			t.Error("expected validation error with empty title")
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

func TestCreateOverlayBoundsChecking(t *testing.T) {
	t.Run("TypeIndexStaysAtZeroOnUpArrow", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusType
		overlay.typeIndex = 0

		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyUp})
		if overlay.typeIndex != 0 {
			t.Errorf("expected type index to stay at 0, got %d", overlay.typeIndex)
		}
	})

	t.Run("TypeIndexStaysAtMaxOnDownArrow", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusType
		overlay.typeIndex = 4 // chore (last)

		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyDown})
		if overlay.typeIndex != 4 {
			t.Errorf("expected type index to stay at 4, got %d", overlay.typeIndex)
		}
	})

	t.Run("PriorityIndexStaysAtZeroOnUpArrow", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusPriority
		overlay.priorityIndex = 0 // critical

		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyUp})
		if overlay.priorityIndex != 0 {
			t.Errorf("expected priority index to stay at 0, got %d", overlay.priorityIndex)
		}
	})

	t.Run("PriorityIndexStaysAtMaxOnDownArrow", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusPriority
		overlay.priorityIndex = 4 // backlog (last)

		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyDown})
		if overlay.priorityIndex != 4 {
			t.Errorf("expected priority index to stay at 4, got %d", overlay.priorityIndex)
		}
	})
}

func TestCreateOverlayFullNavigationCycle(t *testing.T) {
	t.Run("AssigneeTabWrapsToTitle", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusAssignee

		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyTab})
		if overlay.Focus() != FocusTitle {
			t.Errorf("expected focus to wrap to title, got %d", overlay.Focus())
		}
	})

	t.Run("ShiftTabFullCycle", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})

		// Start at Title, go backwards
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
		if overlay.Focus() != FocusParent {
			t.Errorf("expected Title -> Parent, got %d", overlay.Focus())
		}

		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
		if overlay.Focus() != FocusAssignee {
			t.Errorf("expected Parent -> Assignee, got %d", overlay.Focus())
		}

		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
		if overlay.Focus() != FocusLabels {
			t.Errorf("expected Assignee -> Labels, got %d", overlay.Focus())
		}

		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
		if overlay.Focus() != FocusPriority {
			t.Errorf("expected Labels -> Priority, got %d", overlay.Focus())
		}

		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
		if overlay.Focus() != FocusType {
			t.Errorf("expected Priority -> Type, got %d", overlay.Focus())
		}

		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
		if overlay.Focus() != FocusTitle {
			t.Errorf("expected Type -> Title, got %d", overlay.Focus())
		}
	})
}

func TestCreateOverlayEscapeWithDropdown(t *testing.T) {
	t.Run("EscapeClosesParentDropdownFirst", func(t *testing.T) {
		parents := []ParentOption{{ID: "ab-1", Display: "ab-1 Test"}}
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableParents: parents,
		})
		overlay.focus = FocusParent

		// Focus the combo and type to open dropdown
		overlay.parentCombo.Focus()
		overlay.parentCombo, _ = overlay.parentCombo.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

		// Verify dropdown is open
		if !overlay.parentCombo.IsDropdownOpen() {
			t.Skip("dropdown did not open - combo box behavior may differ")
		}

		// Now Esc should close dropdown, not cancel modal
		overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if cmd != nil {
			msg := cmd()
			if _, ok := msg.(CreateCancelledMsg); ok {
				t.Error("expected Esc to close dropdown first, not cancel modal")
			}
		}
	})

	t.Run("EscapeClosesAssigneeDropdownFirst", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableAssignees: []string{"alice", "bob"},
		})
		overlay.focus = FocusAssignee

		// Focus the combo and type to open dropdown
		overlay.assigneeCombo.Focus()
		overlay.assigneeCombo, _ = overlay.assigneeCombo.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

		// Verify dropdown is open
		if !overlay.assigneeCombo.IsDropdownOpen() {
			t.Skip("dropdown did not open - combo box behavior may differ")
		}

		// Now Esc should close dropdown, not cancel modal
		overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if cmd != nil {
			msg := cmd()
			if _, ok := msg.(CreateCancelledMsg); ok {
				t.Error("expected Esc to close dropdown first, not cancel modal")
			}
		}
	})
}

func TestCreateOverlaySubmitPopulatesAllFields(t *testing.T) {
	t.Run("SubmitIncludesTypeAndPriority", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("Test Bead")
		overlay.typeIndex = 1  // feature
		overlay.priorityIndex = 0 // critical

		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Fatal("expected submit command")
		}
		msg := cmd()
		created, ok := msg.(BeadCreatedMsg)
		if !ok {
			t.Fatalf("expected BeadCreatedMsg, got %T", msg)
		}
		if created.IssueType != "feature" {
			t.Errorf("expected issue type 'feature', got %s", created.IssueType)
		}
		if created.Priority != 0 {
			t.Errorf("expected priority 0 (critical), got %d", created.Priority)
		}
	})

	t.Run("SubmitIncludesParentID", func(t *testing.T) {
		parents := []ParentOption{
			{ID: "ab-123", Display: "ab-123 Parent Bead"},
		}
		overlay := NewCreateOverlay(CreateOverlayOptions{
			DefaultParentID:  "ab-123",
			AvailableParents: parents,
		})
		overlay.titleInput.SetValue("Child Bead")

		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Fatal("expected submit command")
		}
		msg := cmd()
		created := msg.(BeadCreatedMsg)
		if created.ParentID != "ab-123" {
			t.Errorf("expected parent ID 'ab-123', got %s", created.ParentID)
		}
	})

	t.Run("SubmitIncludesLabels", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableLabels: []string{"backend", "frontend", "urgent"},
		})
		overlay.titleInput.SetValue("Labeled Bead")

		// Add chips to labels combo
		overlay.labelsCombo.SetChips([]string{"backend", "urgent"})

		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Fatal("expected submit command")
		}
		msg := cmd()
		created := msg.(BeadCreatedMsg)
		if len(created.Labels) != 2 {
			t.Errorf("expected 2 labels, got %d", len(created.Labels))
		}
		if created.Labels[0] != "backend" || created.Labels[1] != "urgent" {
			t.Errorf("expected labels [backend, urgent], got %v", created.Labels)
		}
	})

	t.Run("UnassignedMapsToEmptyString", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("Unassigned Bead")
		// Default assignee is "Unassigned"

		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Fatal("expected submit command")
		}
		msg := cmd()
		created := msg.(BeadCreatedMsg)
		if created.Assignee != "" {
			t.Errorf("expected empty assignee for 'Unassigned', got '%s'", created.Assignee)
		}
	})

	t.Run("SubmitIncludesAssignee", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableAssignees: []string{"alice", "bob"},
		})
		overlay.titleInput.SetValue("Assigned Bead")
		overlay.assigneeCombo.SetValue("alice")

		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Fatal("expected submit command")
		}
		msg := cmd()
		created := msg.(BeadCreatedMsg)
		if created.Assignee != "alice" {
			t.Errorf("expected assignee 'alice', got '%s'", created.Assignee)
		}
	})
}

func TestCreateOverlayInit(t *testing.T) {
	t.Run("InitReturnsBlinkCommand", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		cmd := overlay.Init()
		if cmd == nil {
			t.Error("expected Init to return a command for cursor blink")
		}
	})
}

func TestCreateOverlayOptionsPassThrough(t *testing.T) {
	t.Run("AvailableLabelsPopulatesCombo", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableLabels: []string{"alpha", "beta", "gamma"},
		})
		if len(overlay.labelsOptions) != 3 {
			t.Errorf("expected 3 labels options, got %d", len(overlay.labelsOptions))
		}
	})

	t.Run("AvailableAssigneesPopulatesCombo", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableAssignees: []string{"alice", "bob"},
		})
		if len(overlay.assigneeOptions) != 2 {
			t.Errorf("expected 2 assignee options, got %d", len(overlay.assigneeOptions))
		}
	})
}

// Tests for Zone 1: Parent Field (ab-6yx)
// These tests verify the spec behavior from Section 3.1 and 4.2

func TestParentFieldDeleteClearsToRoot(t *testing.T) {
	t.Run("DeleteKeyFromParentClearsValue", func(t *testing.T) {
		parents := []ParentOption{
			{ID: "ab-123", Display: "ab-123 Test Parent"},
		}
		overlay := NewCreateOverlay(CreateOverlayOptions{
			DefaultParentID:  "ab-123",
			AvailableParents: parents,
		})
		// Focus parent field
		overlay.focus = FocusParent
		overlay.parentCombo.Focus()

		// Verify parent has value
		if overlay.ParentID() != "ab-123" {
			t.Errorf("expected initial parent ID 'ab-123', got %s", overlay.ParentID())
		}

		// Press Delete
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyDelete})

		// Verify parent cleared
		if overlay.ParentID() != "" {
			t.Errorf("expected empty parent ID after Delete, got %s", overlay.ParentID())
		}
		if !overlay.isRootMode {
			t.Error("expected isRootMode to be true after Delete")
		}
	})

	t.Run("BackspaceKeyFromParentClearsValue", func(t *testing.T) {
		parents := []ParentOption{
			{ID: "ab-456", Display: "ab-456 Another Parent"},
		}
		overlay := NewCreateOverlay(CreateOverlayOptions{
			DefaultParentID:  "ab-456",
			AvailableParents: parents,
		})
		// Focus parent field
		overlay.focus = FocusParent
		overlay.parentCombo.Focus()

		// Press Backspace
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyBackspace})

		// Verify parent cleared
		if overlay.ParentID() != "" {
			t.Errorf("expected empty parent ID after Backspace, got %s", overlay.ParentID())
		}
		if !overlay.isRootMode {
			t.Error("expected isRootMode to be true after Backspace")
		}
	})

	t.Run("DeleteShowsRootIndicatorInView", func(t *testing.T) {
		parents := []ParentOption{
			{ID: "ab-789", Display: "ab-789 Test"},
		}
		overlay := NewCreateOverlay(CreateOverlayOptions{
			DefaultParentID:  "ab-789",
			AvailableParents: parents,
		})
		overlay.focus = FocusParent
		overlay.parentCombo.Focus()

		// Press Delete to clear
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyDelete})

		// Check view shows root indicator
		view := overlay.View()
		if !contains(view, "No Parent") {
			t.Error("expected view to show 'No Parent' after Delete")
		}
	})
}

func TestParentFieldEscRevertsChanges(t *testing.T) {
	t.Run("EscFromParentRevertsToOriginalValue", func(t *testing.T) {
		parents := []ParentOption{
			{ID: "ab-111", Display: "ab-111 Original"},
			{ID: "ab-222", Display: "ab-222 Other"},
		}
		overlay := NewCreateOverlay(CreateOverlayOptions{
			DefaultParentID:  "ab-111",
			AvailableParents: parents,
		})

		// Move to parent field (Shift+Tab from Title sets parentOriginal)
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
		if overlay.Focus() != FocusParent {
			t.Fatalf("expected focus on parent, got %d", overlay.Focus())
		}

		// Verify parentOriginal was set
		if overlay.parentOriginal != "ab-111 Original" {
			t.Errorf("expected parentOriginal 'ab-111 Original', got %s", overlay.parentOriginal)
		}

		// Clear the parent with Delete
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyDelete})
		if overlay.ParentID() != "" {
			t.Errorf("expected parent cleared after Delete, got %s", overlay.ParentID())
		}

		// Press Esc to revert
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})

		// Verify value reverted
		if overlay.ParentID() != "ab-111" {
			t.Errorf("expected parent reverted to 'ab-111', got %s", overlay.ParentID())
		}
		// Verify focus moved to Title
		if overlay.Focus() != FocusTitle {
			t.Errorf("expected focus on Title after Esc, got %d", overlay.Focus())
		}
	})

	t.Run("EscFromParentRestoresIsRootMode", func(t *testing.T) {
		parents := []ParentOption{
			{ID: "ab-333", Display: "ab-333 Test"},
		}
		overlay := NewCreateOverlay(CreateOverlayOptions{
			DefaultParentID:  "ab-333",
			AvailableParents: parents,
		})

		// Move to parent field
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyShiftTab})

		// Clear the parent (sets isRootMode = true)
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyDelete})
		if !overlay.isRootMode {
			t.Error("expected isRootMode true after Delete")
		}

		// Press Esc to revert
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})

		// isRootMode should be false since original had a parent
		if overlay.isRootMode {
			t.Error("expected isRootMode false after Esc revert")
		}
	})
}

func TestParentFieldNavigation(t *testing.T) {
	t.Run("TabFromParentMovesToTitle", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusParent
		overlay.parentCombo.Focus()

		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyTab})
		if overlay.Focus() != FocusTitle {
			t.Errorf("expected focus on Title after Tab from Parent, got %d", overlay.Focus())
		}
	})

	t.Run("ShiftTabFromTitleSetsParentOriginal", func(t *testing.T) {
		parents := []ParentOption{
			{ID: "ab-nav", Display: "ab-nav Navigation Test"},
		}
		overlay := NewCreateOverlay(CreateOverlayOptions{
			DefaultParentID:  "ab-nav",
			AvailableParents: parents,
		})

		// Shift+Tab from Title to Parent
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyShiftTab})

		if overlay.Focus() != FocusParent {
			t.Errorf("expected focus on Parent, got %d", overlay.Focus())
		}
		if overlay.parentOriginal != "ab-nav Navigation Test" {
			t.Errorf("expected parentOriginal set, got %s", overlay.parentOriginal)
		}
	})
}

func TestParentFieldFooterUpdates(t *testing.T) {
	t.Run("FooterShowsSelectEscWhenParentDropdownOpen", func(t *testing.T) {
		parents := []ParentOption{
			{ID: "ab-foot", Display: "ab-foot Footer Test"},
		}
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableParents: parents,
		})
		overlay.focus = FocusParent
		overlay.parentCombo.Focus()

		// Open dropdown with Down arrow
		overlay.parentCombo, _ = overlay.parentCombo.Update(tea.KeyMsg{Type: tea.KeyDown})

		if !overlay.parentCombo.IsDropdownOpen() {
			t.Skip("dropdown did not open")
		}

		view := overlay.View()
		if !contains(view, "Enter Select") {
			t.Error("expected footer to show 'Enter Select' when parent dropdown open")
		}
		if !contains(view, "Esc Revert") {
			t.Error("expected footer to show 'Esc Revert' when parent dropdown open")
		}
		if contains(view, "Create & Add Another") {
			t.Error("expected footer NOT to show 'Create & Add Another' when parent dropdown open")
		}
	})

	t.Run("FooterShowsNormalWhenParentDropdownClosed", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})

		view := overlay.View()
		if !contains(view, "Enter Create") {
			t.Error("expected footer to show 'Enter Create' when dropdown closed")
		}
		if !contains(view, "^Enter Create & Add Another") {
			t.Error("expected footer to show '^Enter Create & Add Another' when dropdown closed")
		}
	})
}

func TestParentFieldRootModeInitialization(t *testing.T) {
	t.Run("RootModeShowsNoParentIndicator", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			IsRootMode: true,
		})

		view := overlay.View()
		if !contains(view, "No Parent") {
			t.Error("expected view to show 'No Parent' in root mode")
		}
	})

	t.Run("NormalModeShowsParentCombo", func(t *testing.T) {
		parents := []ParentOption{
			{ID: "ab-norm", Display: "ab-norm Normal Mode"},
		}
		overlay := NewCreateOverlay(CreateOverlayOptions{
			DefaultParentID:  "ab-norm",
			AvailableParents: parents,
		})

		view := overlay.View()
		if contains(view, "No Parent") {
			t.Error("expected view NOT to show 'No Parent' when parent is selected")
		}
	})
}

func TestParentFieldStateTracking(t *testing.T) {
	t.Run("ParentOriginalGetter", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.parentOriginal = "test-original"
		if overlay.parentOriginal != "test-original" {
			t.Errorf("expected parentOriginal 'test-original', got %s", overlay.parentOriginal)
		}
	})
}

// Tests for Zone 2: Title Field Validation Flash (ab-4rv)
// These tests verify spec behavior from Section 4.4

func TestTitleValidationFlash(t *testing.T) {
	t.Run("EmptySubmitSetsValidationError", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		// Title is empty by default

		if overlay.TitleValidationError() {
			t.Error("expected no validation error initially")
		}

		// Try to submit with empty title
		overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})

		if !overlay.TitleValidationError() {
			t.Error("expected validation error after empty submit")
		}

		// Should return a flash clear command
		if cmd == nil {
			t.Error("expected flash clear command")
		}
	})

	t.Run("FlashClearMsgClearsError", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})

		// Set error state
		overlay.titleValidationError = true

		// Send flash clear message
		overlay, _ = overlay.Update(titleFlashClearMsg{})

		if overlay.TitleValidationError() {
			t.Error("expected validation error to be cleared after titleFlashClearMsg")
		}
	})

	t.Run("ValidTitleDoesNotFlash", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("Valid Title")

		// Submit with valid title
		overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})

		if overlay.TitleValidationError() {
			t.Error("expected no validation error with valid title")
		}

		// cmd should be submit command, not flash command
		if cmd == nil {
			t.Fatal("expected submit command")
		}
		msg := cmd()
		_, ok := msg.(BeadCreatedMsg)
		if !ok {
			t.Errorf("expected BeadCreatedMsg, got %T", msg)
		}
	})

	t.Run("WhitespaceOnlyTitleFlashes", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("   \t\n  ")

		// Submit with whitespace-only title
		overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})

		if !overlay.TitleValidationError() {
			t.Error("expected validation error for whitespace-only title")
		}

		if cmd == nil {
			t.Error("expected flash clear command")
		}
	})
}

func TestTitleValidationFlashGetter(t *testing.T) {
	t.Run("GetterReturnsCorrectState", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})

		if overlay.TitleValidationError() {
			t.Error("expected false initially")
		}

		overlay.titleValidationError = true
		if !overlay.TitleValidationError() {
			t.Error("expected true after setting")
		}
	})
}

// Tests for Zone 3: Vim Navigation (ab-l9e)
func TestVimNavigationKeys(t *testing.T) {
	t.Run("JKNavigatesTypeOptions", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusType
		overlay.typeIndex = 0

		// j moves down
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if overlay.typeIndex != 1 {
			t.Errorf("expected type index 1 after 'j', got %d", overlay.typeIndex)
		}

		// k moves up
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if overlay.typeIndex != 0 {
			t.Errorf("expected type index 0 after 'k', got %d", overlay.typeIndex)
		}
	})

	t.Run("JKNavigatesPriorityOptions", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusPriority
		overlay.priorityIndex = 2 // Medium

		// j moves down
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if overlay.priorityIndex != 3 {
			t.Errorf("expected priority index 3 after 'j', got %d", overlay.priorityIndex)
		}

		// k moves up
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if overlay.priorityIndex != 2 {
			t.Errorf("expected priority index 2 after 'k', got %d", overlay.priorityIndex)
		}
	})

	t.Run("HLNavigatesBetweenColumnsFromType", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusType

		// l moves to priority
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
		if overlay.Focus() != FocusPriority {
			t.Errorf("expected focus on priority after 'l', got %d", overlay.Focus())
		}

		// Note: In Priority column, 'h' is a hotkey for High, not navigation
		// So we test that h doesn't navigate from Type (stays put)
		overlay.focus = FocusType
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
		if overlay.Focus() != FocusType {
			t.Errorf("expected focus to stay on type after 'h', got %d", overlay.Focus())
		}
	})

	t.Run("HLArePriorityHotkeysNotNavigation", func(t *testing.T) {
		// In Priority column, h and l are hotkeys for High and Low
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusPriority
		overlay.priorityIndex = 2 // Medium

		// 'h' selects High (index 1), not navigation
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
		if overlay.priorityIndex != 1 {
			t.Errorf("expected priority index 1 (High) after 'h', got %d", overlay.priorityIndex)
		}

		// 'l' selects Low (index 3), not navigation
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
		if overlay.priorityIndex != 3 {
			t.Errorf("expected priority index 3 (Low) after 'l', got %d", overlay.priorityIndex)
		}
	})

	t.Run("VimKeysBoundsChecking", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusType
		overlay.typeIndex = 0

		// k at top stays at 0
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if overlay.typeIndex != 0 {
			t.Errorf("expected type index to stay at 0, got %d", overlay.typeIndex)
		}

		// j at bottom stays at max
		overlay.typeIndex = 4
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if overlay.typeIndex != 4 {
			t.Errorf("expected type index to stay at 4, got %d", overlay.typeIndex)
		}
	})
}

// Tests for Zone 3: Hotkey Underlines (ab-l9e)
func TestHotkeyUnderlines(t *testing.T) {
	t.Run("UnderlineFirstCharHelper", func(t *testing.T) {
		result := underlineFirstChar("Task")
		// Should have combining underline after 'T'
		if result != "T\u0332ask" {
			t.Errorf("expected T̲ask, got %s", result)
		}

		result = underlineFirstChar("Feature")
		if result != "F\u0332eature" {
			t.Errorf("expected F̲eature, got %s", result)
		}

		// Empty string should return empty
		result = underlineFirstChar("")
		if result != "" {
			t.Errorf("expected empty string, got %s", result)
		}
	})

	t.Run("TypeColumnShowsUnderlineWhenFocused", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusType

		view := overlay.View()
		// Should contain underlined T in Task (T̲ = T + combining underline)
		if !contains(view, "T\u0332") {
			t.Error("expected underlined T in Type column when focused")
		}
	})

	t.Run("PriorityColumnShowsUnderlineWhenFocused", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusPriority

		view := overlay.View()
		// Should contain underlined C in Crit (C̲ = C + combining underline)
		if !contains(view, "C\u0332") {
			t.Error("expected underlined C in Priority column when focused")
		}
	})

	t.Run("NoUnderlineWhenUnfocused", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusTitle // Neither Type nor Priority focused

		view := overlay.View()
		// Should NOT contain combining underline in labels
		// Note: This is a soft check since the view might have other combining chars
		if contains(view, "T\u0332ask") {
			t.Error("expected no underlined Task when Type is unfocused")
		}
	})
}

// Helper function
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
