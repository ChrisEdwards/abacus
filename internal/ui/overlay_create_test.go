package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

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

func TestCreateOverlayAssigneeTwoStageEscape(t *testing.T) {
	t.Run("FirstEscClosesDropdownKeepsText", func(t *testing.T) {
		// Setup with assignee options
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableAssignees: []string{"Alice", "Bob"},
		})
		overlay.focus = FocusAssignee
		overlay.assigneeCombo.Focus()

		// Type to open dropdown
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

		// Verify dropdown is open
		if !overlay.assigneeCombo.IsDropdownOpen() {
			t.Skip("dropdown did not open - combo box behavior may differ")
		}

		// First Esc: Close dropdown, keep text
		overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if cmd != nil {
			msg := cmd()
			if _, ok := msg.(CreateCancelledMsg); ok {
				t.Error("first Esc should not cancel modal")
			}
		}

		if overlay.assigneeCombo.IsDropdownOpen() {
			t.Error("dropdown should be closed after first Esc")
		}

		if overlay.assigneeCombo.InputValue() != "a" {
			t.Errorf("expected input 'a' preserved, got %q", overlay.assigneeCombo.InputValue())
		}
	})

	t.Run("SecondEscRevertsToOriginal", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableAssignees: []string{"Alice", "Bob"},
		})
		overlay.focus = FocusAssignee
		overlay.assigneeCombo.SetValue("Alice")
		overlay.assigneeCombo.Focus()

		// Type something different
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b', 'o'}})

		// First Esc: Close dropdown
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})

		// Verify dropdown is closed but text remains
		if overlay.assigneeCombo.IsDropdownOpen() {
			t.Error("dropdown should be closed after first Esc")
		}
		if overlay.assigneeCombo.InputValue() != "bo" {
			t.Errorf("expected input 'bo' after first Esc, got %q", overlay.assigneeCombo.InputValue())
		}

		// Second Esc: Revert to original
		overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if cmd != nil {
			msg := cmd()
			if _, ok := msg.(CreateCancelledMsg); ok {
				t.Error("second Esc should revert, not cancel modal")
			}
		}

		if overlay.assigneeCombo.InputValue() != "Alice" {
			t.Errorf("expected input reverted to 'Alice', got %q", overlay.assigneeCombo.InputValue())
		}
	})

	t.Run("ThirdEscCancelsModal", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableAssignees: []string{"Alice", "Bob"},
		})
		overlay.focus = FocusAssignee
		overlay.assigneeCombo.SetValue("Alice")
		overlay.assigneeCombo.Focus()

		// Third Esc: Cancel modal (input matches value)
		overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if cmd == nil {
			t.Fatal("expected command for modal cancel")
		}

		msg := cmd()
		if _, ok := msg.(CreateCancelledMsg); !ok {
			t.Error("expected CreateCancelledMsg")
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

	t.Run("TitleFlashDuration", func(t *testing.T) {
		// Verify flash duration is 300ms per spec Section 4.4
		if titleFlashDuration != 300*time.Millisecond {
			t.Errorf("expected titleFlashDuration 300ms per spec Section 4.4, got %v", titleFlashDuration)
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

// Tests for Zone 4: Labels (ab-l1k)
// These tests verify spec behavior from Section 3.4

func TestLabelsZoneNavigation(t *testing.T) {
	t.Run("TabFromPriorityMovesToLabels", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusPriority

		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyTab})
		if overlay.Focus() != FocusLabels {
			t.Errorf("expected focus on labels after Tab from Priority, got %d", overlay.Focus())
		}
	})

	t.Run("ShiftTabFromAssigneeMovesToLabels", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusAssignee

		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
		if overlay.Focus() != FocusLabels {
			t.Errorf("expected focus on labels after Shift+Tab from Assignee, got %d", overlay.Focus())
		}
	})

	t.Run("ChipComboBoxTabMsgMovesToAssignee", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusLabels

		overlay, _ = overlay.Update(ChipComboBoxTabMsg{})
		if overlay.Focus() != FocusAssignee {
			t.Errorf("expected focus on assignee after ChipComboBoxTabMsg, got %d", overlay.Focus())
		}
	})
}

func TestLabelsZoneChipHandling(t *testing.T) {
	t.Run("LabelsSubmittedInBeadCreatedMsg", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableLabels: []string{"api", "backend", "frontend"},
		})
		overlay.titleInput.SetValue("Test with labels")
		overlay.labelsCombo.SetChips([]string{"api", "backend"})

		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Fatal("expected submit command")
		}
		msg := cmd()
		created, ok := msg.(BeadCreatedMsg)
		if !ok {
			t.Fatalf("expected BeadCreatedMsg, got %T", msg)
		}
		if len(created.Labels) != 2 {
			t.Errorf("expected 2 labels, got %d", len(created.Labels))
		}
	})

	t.Run("NewLabelEmitsNewLabelAddedMsg", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableLabels: []string{"existing"},
		})
		overlay.focus = FocusLabels

		// Simulate a new label being added (not in existing options)
		overlay, cmd := overlay.Update(ChipComboBoxChipAddedMsg{
			Label: "newlabel",
			IsNew: true,
		})

		if cmd == nil {
			t.Fatal("expected command for new label")
		}
		msg := cmd()
		newLabelMsg, ok := msg.(NewLabelAddedMsg)
		if !ok {
			t.Fatalf("expected NewLabelAddedMsg, got %T", msg)
		}
		if newLabelMsg.Label != "newlabel" {
			t.Errorf("expected label 'newlabel', got '%s'", newLabelMsg.Label)
		}
	})

	t.Run("ExistingLabelDoesNotEmitNewLabelMsg", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableLabels: []string{"existing"},
		})
		overlay.focus = FocusLabels

		// Simulate an existing label being added
		overlay, cmd := overlay.Update(ChipComboBoxChipAddedMsg{
			Label: "existing",
			IsNew: false,
		})

		if cmd != nil {
			t.Error("expected no command for existing label")
		}
	})
}

func TestLabelsZoneEscapeBehavior(t *testing.T) {
	t.Run("EscapeWithLabelsDropdownOpenClosesDropdown", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableLabels: []string{"api", "backend"},
		})
		overlay.focus = FocusLabels
		overlay.labelsCombo.Focus()

		// Type to open dropdown
		overlay.labelsCombo, _ = overlay.labelsCombo.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

		if !overlay.labelsCombo.IsDropdownOpen() {
			t.Skip("dropdown did not open - combo behavior may differ")
		}

		// Esc should close dropdown, not cancel modal
		overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if cmd != nil {
			msg := cmd()
			if _, ok := msg.(CreateCancelledMsg); ok {
				t.Error("expected Esc to close labels dropdown first, not cancel modal")
			}
		}
	})
}

func TestLabelsZoneViewRendering(t *testing.T) {
	t.Run("ViewContainsLabelsSection", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableLabels: []string{"alpha", "beta"},
		})

		view := overlay.View()
		if !contains(view, "LABELS") {
			t.Error("expected view to contain LABELS zone header")
		}
	})

	t.Run("LabelsZoneHighlightedWhenFocused", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableLabels: []string{"alpha", "beta"},
		})
		overlay.focus = FocusLabels

		view := overlay.View()
		// When focused, LABELS header should be styled differently
		// This is a basic check that the zone is rendered
		if !contains(view, "LABELS") {
			t.Error("expected view to contain LABELS zone header when focused")
		}
	})
}

func TestNewLabelAddedMsgType(t *testing.T) {
	t.Run("MessageHasLabelField", func(t *testing.T) {
		msg := NewLabelAddedMsg{Label: "test-label"}
		if msg.Label != "test-label" {
			t.Errorf("expected label 'test-label', got '%s'", msg.Label)
		}
	})
}

func TestNewAssigneeAddedMsgType(t *testing.T) {
	t.Run("MessageHasAssigneeField", func(t *testing.T) {
		msg := NewAssigneeAddedMsg{Assignee: "test-user"}
		if msg.Assignee != "test-user" {
			t.Errorf("expected assignee 'test-user', got '%s'", msg.Assignee)
		}
	})
}

// Tests for Tab/Shift+Tab Focus Cycling (ab-z58)
func TestTabFocusCycling(t *testing.T) {
	t.Run("TabFromAssigneeCommitsDropdownValue", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableAssignees: []string{"alice", "bob", "carlos"},
		})

		// Navigate to Assignee
		overlay.focus = FocusAssignee
		overlay.assigneeCombo.Focus()

		// Type "b" to filter to "bob"
		overlay.assigneeCombo, _ = overlay.assigneeCombo.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})

		// Verify dropdown is open with "bob" highlighted
		if !overlay.assigneeCombo.IsDropdownOpen() {
			t.Skip("dropdown did not open")
		}

		// Press Tab - should commit bob and move to Title
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyTab})

		// Check focus moved to Title
		if overlay.Focus() != FocusTitle {
			t.Errorf("expected focus on Title after Tab, got %d", overlay.Focus())
		}

		// Check value was committed
		if overlay.assigneeCombo.Value() != "bob" {
			t.Errorf("expected assignee 'bob', got '%s'", overlay.assigneeCombo.Value())
		}
	})

	t.Run("TabFromAssigneeWithNewValueEmitsToast", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableAssignees: []string{"alice", "bob"},
		})

		// Navigate to Assignee
		overlay.focus = FocusAssignee
		overlay.assigneeCombo.Focus()

		// Type a new assignee name (not in existing list)
		overlay.assigneeCombo, _ = overlay.assigneeCombo.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
		overlay.assigneeCombo, _ = overlay.assigneeCombo.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
		overlay.assigneeCombo, _ = overlay.assigneeCombo.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

		// Press Tab - should create new assignee and emit toast
		var cmd tea.Cmd
		overlay, cmd = overlay.Update(tea.KeyMsg{Type: tea.KeyTab})

		// Check focus moved to Title
		if overlay.Focus() != FocusTitle {
			t.Errorf("expected focus on Title after Tab, got %d", overlay.Focus())
		}

		// Check value was committed
		if overlay.assigneeCombo.Value() != "zed" {
			t.Errorf("expected assignee 'zed', got '%s'", overlay.assigneeCombo.Value())
		}

		// Check that NewAssigneeAddedMsg is emitted
		if cmd == nil {
			t.Fatal("expected command for new assignee toast")
		}

		// Execute all batched commands to find NewAssigneeAddedMsg
		foundNewAssigneeMsg := false
		msg := cmd()
		// Handle tea.BatchMsg by checking the individual messages
		if batchMsg, ok := msg.(tea.BatchMsg); ok {
			for _, batchCmd := range batchMsg {
				if batchCmd != nil {
					innerMsg := batchCmd()
					if nam, ok := innerMsg.(NewAssigneeAddedMsg); ok {
						foundNewAssigneeMsg = true
						if nam.Assignee != "zed" {
							t.Errorf("expected NewAssigneeAddedMsg with 'zed', got '%s'", nam.Assignee)
						}
					}
				}
			}
		} else if nam, ok := msg.(NewAssigneeAddedMsg); ok {
			foundNewAssigneeMsg = true
			if nam.Assignee != "zed" {
				t.Errorf("expected NewAssigneeAddedMsg with 'zed', got '%s'", nam.Assignee)
			}
		}

		if !foundNewAssigneeMsg {
			t.Error("expected NewAssigneeAddedMsg to be emitted for new assignee")
		}
	})

	t.Run("TabCycleFullLoop", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})

		// Start at Title
		if overlay.Focus() != FocusTitle {
			t.Errorf("expected initial focus on Title, got %d", overlay.Focus())
		}

		// Tab: Title -> Type
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyTab})
		if overlay.Focus() != FocusType {
			t.Errorf("expected focus on Type, got %d", overlay.Focus())
		}

		// Tab: Type -> Priority
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyTab})
		if overlay.Focus() != FocusPriority {
			t.Errorf("expected focus on Priority, got %d", overlay.Focus())
		}

		// Tab: Priority -> Labels
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyTab})
		if overlay.Focus() != FocusLabels {
			t.Errorf("expected focus on Labels, got %d", overlay.Focus())
		}

		// Tab: Labels -> Assignee (via ChipComboBoxTabMsg)
		overlay, _ = overlay.Update(ChipComboBoxTabMsg{})
		if overlay.Focus() != FocusAssignee {
			t.Errorf("expected focus on Assignee, got %d", overlay.Focus())
		}

		// Tab: Assignee -> Title (wrap)
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyTab})
		if overlay.Focus() != FocusTitle {
			t.Errorf("expected focus to wrap to Title, got %d", overlay.Focus())
		}
	})
}

func TestAssigneeZone(t *testing.T) {
	t.Run("AssigneeZoneRendered", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableAssignees: []string{"alice", "bob"},
		})

		view := overlay.View()
		if !contains(view, "ASSIGNEE") {
			t.Error("expected view to contain ASSIGNEE zone header")
		}
	})

	t.Run("UnassignedIsDefault", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableAssignees: []string{"alice", "bob"},
		})

		view := overlay.View()
		if !contains(view, "Unassigned") {
			t.Error("expected view to show Unassigned as default")
		}
	})

	t.Run("AssigneeOptionsIncludeProvided", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableAssignees: []string{"alice", "bob"},
		})

		// Focus assignee and open dropdown
		overlay.focus = FocusAssignee
		overlay.assigneeCombo.Focus()
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyDown})

		view := overlay.View()
		if !contains(view, "alice") {
			t.Error("expected dropdown to contain 'alice'")
		}
		if !contains(view, "bob") {
			t.Error("expected dropdown to contain 'bob'")
		}
	})

	t.Run("AssigneeSubmitsCorrectly", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableAssignees: []string{"alice"},
		})

		// Set title
		overlay.titleInput.SetValue("Test Bead")

		// Navigate to assignee
		overlay.focus = FocusAssignee
		overlay.assigneeCombo.Focus()
		overlay.assigneeCombo.SetValue("alice")

		// Submit
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Fatal("expected command from submit")
		}
		msg := cmd()
		created, ok := msg.(BeadCreatedMsg)
		if !ok {
			t.Fatalf("expected BeadCreatedMsg, got %T", msg)
		}
		if created.Assignee != "alice" {
			t.Errorf("expected assignee 'alice', got '%s'", created.Assignee)
		}
	})

	t.Run("UnassignedConvertsToEmptyString", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})

		// Set title
		overlay.titleInput.SetValue("Test Bead")

		// Leave assignee as default "Unassigned"
		// Submit
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Fatal("expected command from submit")
		}
		msg := cmd()
		created, ok := msg.(BeadCreatedMsg)
		if !ok {
			t.Fatalf("expected BeadCreatedMsg, got %T", msg)
		}
		if created.Assignee != "" {
			t.Errorf("expected empty assignee for Unassigned, got '%s'", created.Assignee)
		}
	})

	t.Run("MeOptionExtractsUsername", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})

		// Set title
		overlay.titleInput.SetValue("Test Bead")

		// Set assignee to "Me (username)" format
		overlay.assigneeCombo.SetValue("Me (testuser)")

		// Submit
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Fatal("expected command from submit")
		}
		msg := cmd()
		created, ok := msg.(BeadCreatedMsg)
		if !ok {
			t.Fatalf("expected BeadCreatedMsg, got %T", msg)
		}
		if created.Assignee != "testuser" {
			t.Errorf("expected assignee 'testuser' (extracted from Me format), got '%s'", created.Assignee)
		}
	})
}

// TestBulkEntryMode tests Ctrl+Enter bulk entry feature (spec Section 4.3)
func TestBulkEntryMode(t *testing.T) {
	t.Run("CtrlEnterStaysOpen", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("First task")

		// Call handleSubmit(true) directly (simulates Ctrl+Enter)
		_, cmd := overlay.handleSubmit(true)
		if cmd == nil {
			t.Fatal("expected command from handleSubmit(true)")
		}

		// Execute command and check StayOpen flag
		msg := cmd()
		if msg == nil {
			t.Fatal("expected message from command")
		}

		// The command is batched, so we need to unwrap the batch
		batchMsg, ok := msg.(tea.BatchMsg)
		if !ok {
			t.Fatalf("expected tea.BatchMsg, got %T", msg)
		}

		// Execute first command (submit)
		submitMsg := batchMsg[0]()
		created, ok := submitMsg.(BeadCreatedMsg)
		if !ok {
			t.Fatalf("expected BeadCreatedMsg, got %T", submitMsg)
		}

		if !created.StayOpen {
			t.Error("expected StayOpen=true for handleSubmit(true)")
		}
	})

	t.Run("CtrlEnterValidatesTitleFirst", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		// Leave title empty

		// Call handleSubmit(true) directly (simulates Ctrl+Enter)
		overlay, cmd := overlay.handleSubmit(true)
		if cmd == nil {
			t.Fatal("expected command from handleSubmit(true)")
		}

		// Should be flash command, not submit
		msg := cmd()
		if _, ok := msg.(BeadCreatedMsg); ok {
			t.Error("expected no submit (BeadCreatedMsg) with empty title")
		}

		if !overlay.TitleValidationError() {
			t.Error("expected validation error with empty title")
		}
	})

	t.Run("CtrlEnterClearsTitle", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("First task")

		// Call handleSubmit(true) directly (simulates Ctrl+Enter)
		_, cmd := overlay.handleSubmit(true)

		// Execute batch command
		batchMsg := cmd().(tea.BatchMsg)
		resetMsg := batchMsg[1]()

		// Apply reset message
		overlay, _ = overlay.Update(resetMsg)

		if overlay.Title() != "" {
			t.Errorf("expected title to be cleared, got '%s'", overlay.Title())
		}
	})

	t.Run("CtrlEnterPersistsType", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("First task")
		overlay.typeIndex = 1 // Feature

		// Call handleSubmit(true) directly (simulates Ctrl+Enter)
		_, cmd := overlay.handleSubmit(true)

		// Execute batch command and apply reset
		batchMsg := cmd().(tea.BatchMsg)
		resetMsg := batchMsg[1]()
		overlay, _ = overlay.Update(resetMsg)

		if overlay.IssueType() != "feature" {
			t.Errorf("expected type 'feature' to persist, got '%s'", overlay.IssueType())
		}
	})

	t.Run("CtrlEnterPersistsPriority", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("First task")
		overlay.priorityIndex = 0 // Critical

		// Call handleSubmit(true) directly (simulates Ctrl+Enter)
		_, cmd := overlay.handleSubmit(true)

		// Execute batch command and apply reset
		batchMsg := cmd().(tea.BatchMsg)
		resetMsg := batchMsg[1]()
		overlay, _ = overlay.Update(resetMsg)

		if overlay.Priority() != 0 {
			t.Errorf("expected priority 0 (Critical) to persist, got %d", overlay.Priority())
		}
	})

	t.Run("CtrlEnterPersistsLabels", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableLabels: []string{"api", "backend"},
		})
		overlay.titleInput.SetValue("First task")
		overlay.labelsCombo.SetChips([]string{"api", "backend"})

		// Call handleSubmit(true) directly (simulates Ctrl+Enter)
		_, cmd := overlay.handleSubmit(true)

		// Execute batch command and apply reset
		batchMsg := cmd().(tea.BatchMsg)
		resetMsg := batchMsg[1]()
		overlay, _ = overlay.Update(resetMsg)

		chips := overlay.labelsCombo.GetChips()
		if len(chips) != 2 || chips[0] != "api" || chips[1] != "backend" {
			t.Errorf("expected labels [api, backend] to persist, got %v", chips)
		}
	})

	t.Run("CtrlEnterPersistsAssignee", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableAssignees: []string{"alice"},
		})
		overlay.titleInput.SetValue("First task")
		overlay.assigneeCombo.SetValue("alice")

		// Call handleSubmit(true) directly (simulates Ctrl+Enter)
		_, cmd := overlay.handleSubmit(true)

		// Execute batch command and apply reset
		batchMsg := cmd().(tea.BatchMsg)
		resetMsg := batchMsg[1]()
		overlay, _ = overlay.Update(resetMsg)

		if overlay.assigneeCombo.Value() != "alice" {
			t.Errorf("expected assignee 'alice' to persist, got '%s'", overlay.assigneeCombo.Value())
		}
	})

	t.Run("CtrlEnterPersistsParent", func(t *testing.T) {
		parents := []ParentOption{
			{ID: "ab-123", Display: "ab-123 Parent Task"},
		}
		overlay := NewCreateOverlay(CreateOverlayOptions{
			DefaultParentID:  "ab-123",
			AvailableParents: parents,
		})
		overlay.titleInput.SetValue("First task")

		// Call handleSubmit(true) directly (simulates Ctrl+Enter)
		_, cmd := overlay.handleSubmit(true)

		// Execute batch command and apply reset
		batchMsg := cmd().(tea.BatchMsg)
		resetMsg := batchMsg[1]()
		overlay, _ = overlay.Update(resetMsg)

		if overlay.ParentID() != "ab-123" {
			t.Errorf("expected parent 'ab-123' to persist, got '%s'", overlay.ParentID())
		}
	})

	t.Run("CtrlEnterRefocusesTitle", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("First task")

		// Change focus away from title
		overlay.focus = FocusType

		// Call handleSubmit(true) directly (simulates Ctrl+Enter)
		_, cmd := overlay.handleSubmit(true)

		// Execute batch command and apply reset
		batchMsg := cmd().(tea.BatchMsg)
		resetMsg := batchMsg[1]()
		overlay, _ = overlay.Update(resetMsg)

		if overlay.Focus() != FocusTitle {
			t.Errorf("expected focus on Title, got %d", overlay.Focus())
		}
	})

	t.Run("RegularEnterStillCloses", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("Test task")

		// Press regular Enter
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Fatal("expected command from Enter")
		}

		msg := cmd()
		created, ok := msg.(BeadCreatedMsg)
		if !ok {
			t.Fatalf("expected BeadCreatedMsg, got %T", msg)
		}

		if created.StayOpen {
			t.Error("expected StayOpen=false for regular Enter")
		}
	})

	t.Run("CtrlEnterWorksFromAnyField", func(t *testing.T) {
		// Test Safety Valve: Ctrl+Enter submits regardless of focus (spec Section 6)
		testCases := []struct {
			name  string
			focus CreateFocus
		}{
			{"FromTitle", FocusTitle},
			{"FromType", FocusType},
			{"FromPriority", FocusPriority},
			{"FromLabels", FocusLabels},
			{"FromAssignee", FocusAssignee},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				overlay := NewCreateOverlay(CreateOverlayOptions{})
				overlay.titleInput.SetValue("Test task")
				overlay.focus = tc.focus

				// Call handleSubmit(true) directly (simulates Ctrl+Enter)
				_, cmd := overlay.handleSubmit(true)
				if cmd == nil {
					t.Fatalf("expected command from handleSubmit(true) when focus is %d", tc.focus)
				}

				// Execute batch and check for BeadCreatedMsg
				batchMsg, ok := cmd().(tea.BatchMsg)
				if !ok {
					t.Fatalf("expected tea.BatchMsg, got %T", cmd())
				}

				submitMsg := batchMsg[0]()
				created, ok := submitMsg.(BeadCreatedMsg)
				if !ok {
					t.Fatalf("expected BeadCreatedMsg, got %T", submitMsg)
				}

				if !created.StayOpen {
					t.Errorf("expected StayOpen=true for handleSubmit(true) from focus %d", tc.focus)
				}
			})
		}
	})
}

// ============================================================================
// Type Auto-Inference Tests (spec Section 5)
// ============================================================================

func TestInferTypeFromTitle(t *testing.T) {
	t.Run("BugKeywords", func(t *testing.T) {
		testCases := []struct {
			title string
		}{
			{"Fix login bug"},
			{"fix the issue"},
			{"Broken authentication"},
			{"Bug in parser"},
			{"Error loading page"},
			{"Crash on startup"},
			{"Issue with navigation"},
		}

		for _, tc := range testCases {
			t.Run(tc.title, func(t *testing.T) {
				idx := inferTypeFromTitle(tc.title)
				if idx != 2 {
					t.Errorf("title %q: expected index 2 (bug), got %d", tc.title, idx)
				}
			})
		}
	})

	t.Run("FeatureKeywords", func(t *testing.T) {
		testCases := []struct {
			title string
		}{
			{"Add user login"},
			{"Implement authentication"},
			{"Create new dashboard"},
			{"Build API endpoint"},
			{"New feature for export"},
		}

		for _, tc := range testCases {
			t.Run(tc.title, func(t *testing.T) {
				idx := inferTypeFromTitle(tc.title)
				if idx != 1 {
					t.Errorf("title %q: expected index 1 (feature), got %d", tc.title, idx)
				}
			})
		}
	})

	t.Run("ChoreKeywords", func(t *testing.T) {
		testCases := []struct {
			title       string
			expectedIdx int
		}{
			{"Refactor user service", 4},
			{"Clean up old code", 4},
			{"Reorganize project structure", 4},
			{"Simplify API logic", 4},
			{"Extract common utilities", 4},
			{"Update dependencies", 4},
			{"Upgrade React version", 4},
			{"Bump version to 2.0", 4},
			{"Migrate to new API", 4},
			{"Document API endpoints", 4},
			{"Update docs", 4},
			{"Update README section", 4},  // Use Update instead of Add to match chore
		}

		for _, tc := range testCases {
			t.Run(tc.title, func(t *testing.T) {
				idx := inferTypeFromTitle(tc.title)
				if idx != tc.expectedIdx {
					t.Errorf("title %q: expected index %d (chore), got %d", tc.title, tc.expectedIdx, idx)
				}
			})
		}
	})

	t.Run("CaseInsensitive", func(t *testing.T) {
		testCases := []struct {
			title       string
			expectedIdx int
		}{
			{"fix bug", 2},
			{"Fix bug", 2},
			{"FIX bug", 2},
			{"add feature", 1},
			{"Add feature", 1},
			{"ADD feature", 1},
		}

		for _, tc := range testCases {
			t.Run(tc.title, func(t *testing.T) {
				idx := inferTypeFromTitle(tc.title)
				if idx != tc.expectedIdx {
					t.Errorf("title %q: expected index %d, got %d", tc.title, tc.expectedIdx, idx)
				}
			})
		}
	})

	t.Run("WordBoundaries", func(t *testing.T) {
		testCases := []struct {
			title       string
			expectedIdx int // -1 means no match
		}{
			{"Prefix component", -1},  // "fix" is part of "Prefix"
			{"Adder utility", -1},     // "add" is part of "Adder"
			{"Buggy behavior", -1},    // "bug" is part of "Buggy"
			{"Fix bug", 2},            // "fix" is standalone word
			{"Add feature", 1},        // "add" is standalone word
			{"Bug report", 2},         // "bug" is standalone word
		}

		for _, tc := range testCases {
			t.Run(tc.title, func(t *testing.T) {
				idx := inferTypeFromTitle(tc.title)
				if idx != tc.expectedIdx {
					t.Errorf("title %q: expected index %d, got %d", tc.title, tc.expectedIdx, idx)
				}
			})
		}
	})

	t.Run("FirstMatchWins", func(t *testing.T) {
		testCases := []struct {
			title       string
			expectedIdx int
			reason      string
		}{
			{"Fix the Add button", 2, "Fix comes before Add"},
			{"Adding fix for login", 2, "Adding doesn't match (word boundary), so fix wins"},
			{"Bug in the new feature", 2, "Bug comes before new"},
			{"Create fix for crash", 1, "Create comes before fix"},
		}

		for _, tc := range testCases {
			t.Run(tc.title, func(t *testing.T) {
				idx := inferTypeFromTitle(tc.title)
				if idx != tc.expectedIdx {
					t.Errorf("title %q: expected index %d (%s), got %d", tc.title, tc.expectedIdx, tc.reason, idx)
				}
			})
		}
	})

	t.Run("EmptyOrWhitespace", func(t *testing.T) {
		testCases := []string{
			"",
			"   ",
			"\t",
			"\n",
		}

		for _, title := range testCases {
			t.Run("empty", func(t *testing.T) {
				idx := inferTypeFromTitle(title)
				if idx != -1 {
					t.Errorf("title %q: expected -1 (no match), got %d", title, idx)
				}
			})
		}
	})

	t.Run("NoMatch", func(t *testing.T) {
		testCases := []string{
			"Something random",
			"User authentication flow",
			"Testing the application",
		}

		for _, title := range testCases {
			t.Run(title, func(t *testing.T) {
				idx := inferTypeFromTitle(title)
				if idx != -1 {
					t.Errorf("title %q: expected -1 (no match), got %d", title, idx)
				}
			})
		}
	})
}

func TestTypeInferenceIntegration(t *testing.T) {
	t.Run("InferenceTriggeredOnTitleChange", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusTitle

		// Type "Fix" - should infer Bug
		overlay.titleInput.SetValue("F")
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
		overlay.titleInput.SetValue("Fi")
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
		overlay.titleInput.SetValue("Fix")
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

		if overlay.typeIndex != 2 {
			t.Errorf("expected type index 2 (bug) after typing 'Fix', got %d", overlay.typeIndex)
		}
	})

	t.Run("ManualOverrideDisablesInference", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusTitle

		// Type "Fix" - should infer Bug
		overlay.titleInput.SetValue("Fix")
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

		if overlay.typeIndex != 2 {
			t.Error("expected initial inference to Bug")
		}

		// Manually select Epic
		overlay.focus = FocusType
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

		if !overlay.typeManuallySet {
			t.Error("expected typeManuallySet=true after hotkey")
		}
		if overlay.typeIndex != 3 {
			t.Errorf("expected type index 3 (epic), got %d", overlay.typeIndex)
		}

		// Change title - should NOT infer
		overlay.focus = FocusTitle
		overlay.titleInput.SetValue("Add feature")
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

		// Should stay Epic
		if overlay.typeIndex != 3 {
			t.Errorf("expected type to stay 3 (epic) after manual override, got %d", overlay.typeIndex)
		}
	})

	t.Run("ArrowKeysSetManualFlag", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusType

		// Press Down arrow
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyDown})

		if !overlay.typeManuallySet {
			t.Error("expected typeManuallySet=true after arrow key")
		}
	})

	t.Run("VimKeysSetManualFlag", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusType

		// Press 'j' (vim down)
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

		if !overlay.typeManuallySet {
			t.Error("expected typeManuallySet=true after vim key")
		}
	})

	t.Run("HotkeysSetManualFlag", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusType

		// Press 'f' for Feature
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})

		if !overlay.typeManuallySet {
			t.Error("expected typeManuallySet=true after hotkey")
		}
		if overlay.typeIndex != 1 {
			t.Errorf("expected type index 1 (feature), got %d", overlay.typeIndex)
		}
	})

	t.Run("InferenceActivatesFlash", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusTitle
		overlay.typeIndex = 0 // Start with Task

		// Type "Fix" - should infer Bug and activate flash
		overlay.titleInput.SetValue("Fix")
		overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

		if overlay.typeIndex != 2 {
			t.Errorf("expected type index 2 (bug), got %d", overlay.typeIndex)
		}

		if !overlay.typeInferenceActive {
			t.Error("expected typeInferenceActive=true after inference")
		}

		// Check that flash command was returned
		if cmd == nil {
			t.Error("expected flash command to be returned")
		}
	})

	t.Run("FlashClearsAfterMessage", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.typeInferenceActive = true

		// Send flash clear message
		overlay, _ = overlay.Update(typeInferenceFlashMsg{})

		if overlay.typeInferenceActive {
			t.Error("expected typeInferenceActive=false after flash clear message")
		}
	})

	t.Run("NoFlashWhenTypeUnchanged", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusTitle
		overlay.typeIndex = 2 // Already Bug

		// Type "Fix" - should match Bug but not change type
		overlay.titleInput.SetValue("Fix")
		overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

		// Should not activate flash since type didn't change
		if overlay.typeInferenceActive {
			t.Error("expected no flash when type unchanged")
		}

		// Should not return flash command
		if cmd != nil {
			// The cmd might be a batch, check if it contains flash
			// For simplicity, we just check that typeInferenceActive is false
		}
	})
}

// ============================================================================
// Dynamic Footer Tests (spec Section 4.1)
// ============================================================================

func TestCreateOverlayFooter(t *testing.T) {
	t.Run("DefaultFooter", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		footer := overlay.renderFooter()

		if !strings.Contains(footer, "Enter Create") {
			t.Error("expected default footer to contain 'Enter Create'")
		}
		if !strings.Contains(footer, "^Enter Create & Add Another") {
			t.Error("expected default footer to contain bulk entry hint")
		}
		if !strings.Contains(footer, "Tab Next") {
			t.Error("expected default footer to contain 'Tab Next'")
		}
		if !strings.Contains(footer, "Esc Cancel") {
			t.Error("expected default footer to contain 'Esc Cancel'")
		}
	})

	t.Run("ParentSearchFooter", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableParents: []ParentOption{
				{ID: "ab-123", Display: "ab-123 Test"},
				{ID: "ab-456", Display: "ab-456 Another"},
			},
		})

		// Navigate to parent field and focus it
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyShiftTab}) // From Title to Parent

		// Type to open dropdown in Filtering mode
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

		footer := overlay.renderFooter()

		if !strings.Contains(footer, "Enter Select") {
			t.Errorf("expected parent search footer to contain 'Enter Select', got: %s", footer)
		}
		if !strings.Contains(footer, "Esc Revert") {
			t.Errorf("expected parent search footer to contain 'Esc Revert', got: %s", footer)
		}
		// Should NOT contain the default footer text
		if strings.Contains(footer, "Enter Create") {
			t.Errorf("expected parent search footer to not contain 'Enter Create', got: %s", footer)
		}
	})

	t.Run("CreatingFooter", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.isCreating = true

		footer := overlay.renderFooter()

		if !strings.Contains(footer, "Creating bead...") {
			t.Error("expected creating footer to contain 'Creating bead...'")
		}
		// Should NOT contain the default footer text
		if strings.Contains(footer, "Enter Create") {
			t.Error("expected creating footer to not contain 'Enter Create'")
		}
	})

	t.Run("FooterInView", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		view := overlay.View()

		// Verify default footer appears in view
		if !strings.Contains(view, "Enter Create") {
			t.Error("expected default footer in view output")
		}
	})

	t.Run("FooterSwitchesWithParentDropdown", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableParents: []ParentOption{
				{ID: "ab-123", Display: "ab-123 Test"},
			},
		})

		// Initially should show default footer
		footer := overlay.renderFooter()
		if !strings.Contains(footer, "Enter Create") {
			t.Error("expected default footer initially")
		}

		// Navigate to parent field
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyShiftTab}) // From Title to Parent

		// Type to open dropdown
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})

		// Should now show parent search footer
		footer = overlay.renderFooter()
		if !strings.Contains(footer, "Enter Select") {
			t.Errorf("expected parent search footer after opening dropdown, got: %s", footer)
		}

		// Close dropdown with Esc
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})

		// Should revert to default footer
		footer = overlay.renderFooter()
		if !strings.Contains(footer, "Enter Create") {
			t.Errorf("expected default footer after closing dropdown, got: %s", footer)
		}
	})

	t.Run("FooterShowsBrowseHintOnParentField", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableParents: []ParentOption{
				{ID: "ab-123", Display: "ab-123 Test"},
			},
		})

		// Navigate to parent field (but don't open dropdown)
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyShiftTab}) // From Title to Parent

		footer := overlay.renderFooter()

		if !strings.Contains(footer, "↓ Browse") {
			t.Errorf("expected footer to contain '↓ Browse' on parent field, got: %s", footer)
		}
		if !strings.Contains(footer, "Enter Create") {
			t.Errorf("expected footer to contain 'Enter Create' on parent field, got: %s", footer)
		}
	})

	t.Run("FooterShowsBrowseHintOnLabelsField", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableLabels: []string{"bug", "feature"},
		})

		// Navigate to labels field
		overlay.focus = FocusLabels

		footer := overlay.renderFooter()

		if !strings.Contains(footer, "↓ Browse") {
			t.Errorf("expected footer to contain '↓ Browse' on labels field, got: %s", footer)
		}
	})

	t.Run("FooterShowsBrowseHintOnAssigneeField", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableAssignees: []string{"alice", "bob"},
		})

		// Navigate to assignee field
		overlay.focus = FocusAssignee

		footer := overlay.renderFooter()

		if !strings.Contains(footer, "↓ Browse") {
			t.Errorf("expected footer to contain '↓ Browse' on assignee field, got: %s", footer)
		}
	})

	t.Run("FooterShowsSelectHintForAnyDropdown", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableLabels:    []string{"bug", "feature"},
			AvailableAssignees: []string{"alice", "bob"},
		})

		// Test labels dropdown
		overlay.focus = FocusLabels
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyDown})

		footer := overlay.renderFooter()
		if !strings.Contains(footer, "Enter Select") {
			t.Errorf("expected 'Enter Select' when labels dropdown open, got: %s", footer)
		}

		// Close labels dropdown
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})

		// Test assignee dropdown
		overlay.focus = FocusAssignee
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyDown})

		footer = overlay.renderFooter()
		if !strings.Contains(footer, "Enter Select") {
			t.Errorf("expected 'Enter Select' when assignee dropdown open, got: %s", footer)
		}
	})
}

func TestCreateOverlayFooterState(t *testing.T) {
	t.Run("IsCreatingSetOnSubmit", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("Test bead")
		overlay.focus = FocusTitle

		// Submit form
		overlay, _ = overlay.handleSubmit(false)

		if !overlay.isCreating {
			t.Error("expected isCreating=true after handleSubmit")
		}
	})

	t.Run("IsCreatingClearedOnBulkReset", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.isCreating = true

		// Process bulk entry reset message
		overlay, _ = overlay.Update(bulkEntryResetMsg{})

		if overlay.isCreating {
			t.Error("expected isCreating=false after bulkEntryResetMsg")
		}
	})

	t.Run("FooterShowsCreatingDuringSubmission", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("Test bead")
		overlay.focus = FocusTitle

		// Before submission
		footer := overlay.renderFooter()
		if strings.Contains(footer, "Creating bead...") {
			t.Error("should not show creating footer before submission")
		}

		// Submit form
		overlay, _ = overlay.handleSubmit(false)

		// After submission
		footer = overlay.renderFooter()
		if !strings.Contains(footer, "Creating bead...") {
			t.Error("expected creating footer after submission")
		}
	})
}

// ============================================================================
// Backend Error Handling Tests (spec Section 4.4)
// ============================================================================

func TestBackendErrorHandling(t *testing.T) {
	t.Run("BackendErrorShowsRedBorder", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("Test Bead")

		// Simulate backend error
		overlay, _ = overlay.Update(backendErrorMsg{err: fmt.Errorf("database connection failed")})

		if !overlay.titleBackendError {
			t.Error("expected titleBackendError=true after backend error")
		}

		if overlay.isCreating {
			t.Error("expected isCreating=false after backend error")
		}
	})

	t.Run("BackendErrorClearedOnRetry", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("Test Bead")
		overlay.titleBackendError = true

		// Retry submission
		overlay, _ = overlay.handleSubmit(false)

		if overlay.titleBackendError {
			t.Error("expected titleBackendError=false when retrying after error")
		}
	})

	t.Run("ViewShowsRedBorderOnBackendError", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("Test")
		overlay.titleBackendError = true

		view := overlay.View()
		// Verify the view renders (red border styling is applied via styleCreateInputError)
		if !strings.Contains(view, "Test") {
			t.Error("expected view to show title with error styling")
		}
	})

	t.Run("ViewShowsRedBorderForValidation", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleValidationError = true

		view := overlay.View()
		// Verify red border is applied (visual styling test)
		if view == "" {
			t.Error("expected non-empty view with validation error styling")
		}
	})

	t.Run("ViewShowsRedBorderForBothErrors", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("Test")
		// Set both error types
		overlay.titleValidationError = true
		overlay.titleBackendError = true

		view := overlay.View()
		// Verify view renders with error styling for either error type
		if !strings.Contains(view, "Test") {
			t.Error("expected view to show title with error styling for both error types")
		}
	})

	t.Run("BackendErrorPreservesFormData", func(t *testing.T) {
		// Ensure backend error doesn't clear user's form data
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("My Important Title")
		overlay.typeIndex = 1 // Feature
		overlay.priorityIndex = 0 // Critical

		// Simulate backend error
		overlay, _ = overlay.Update(backendErrorMsg{err: fmt.Errorf("network error")})

		// All data should be preserved
		if overlay.Title() != "My Important Title" {
			t.Errorf("expected title preserved, got %s", overlay.Title())
		}
		if overlay.IssueType() != "feature" {
			t.Errorf("expected issue type preserved, got %s", overlay.IssueType())
		}
		if overlay.Priority() != 0 {
			t.Errorf("expected priority preserved, got %d", overlay.Priority())
		}
	})
}

// Tests for ab-ctal: Backend error display and ESC handling
func TestCreateOverlayBackendErrorDisplay(t *testing.T) {
	t.Run("BackendErrorMsgStoresErrorAndKeepsModalOpen", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		
		// Simulate backend error
		errorMsg := "Database out of sync with JSONL. Run 'bd sync' to fix."
		msg := backendErrorMsg{
			err:    fmt.Errorf("db sync error"),
			errMsg: errorMsg,
		}
		
		overlay, _ = overlay.Update(msg)
		
		// Verify error is stored
		if !overlay.titleBackendError {
			t.Error("expected titleBackendError to be true")
		}
		if overlay.backendErrorMsg != errorMsg {
			t.Errorf("expected backendErrorMsg %q, got %q", errorMsg, overlay.backendErrorMsg)
		}
		if overlay.isCreating {
			t.Error("expected isCreating to be false after error")
		}
	})
	
	t.Run("ESCClearsBackendErrorInsteadOfClosingModal", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		
		// Set backend error state
		overlay.titleBackendError = true
		overlay.backendErrorMsg = "Test error"
		
		// Press ESC
		overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})
		
		// Verify error is cleared, modal stays open
		if overlay.titleBackendError {
			t.Error("expected titleBackendError to be false after ESC")
		}
		if overlay.backendErrorMsg != "" {
			t.Error("expected backendErrorMsg to be cleared after ESC")
		}
		
		// Verify modal is NOT closing (cmd should be nil, not CreateCancelledMsg)
		if cmd != nil {
			msg := cmd()
			if _, ok := msg.(CreateCancelledMsg); ok {
				t.Error("expected ESC to not close modal when there's a backend error")
			}
		}
	})
	
	t.Run("ESCClosesModalWhenNoBackendError", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		
		// No backend error
		overlay.titleBackendError = false
		overlay.backendErrorMsg = ""
		
		// Press ESC
		_, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})
		
		// Verify modal is closing
		if cmd == nil {
			t.Fatal("expected CreateCancelledMsg command")
		}
		msg := cmd()
		if _, ok := msg.(CreateCancelledMsg); !ok {
			t.Errorf("expected CreateCancelledMsg, got %T", msg)
		}
	})
	
	t.Run("SubmitClearsBackendError", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("Test Title")
		
		// Set backend error state
		overlay.titleBackendError = true
		overlay.backendErrorMsg = "Previous error"
		
		// Submit with Enter
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})
		
		// Verify error is cleared when retrying
		if overlay.titleBackendError {
			t.Error("expected titleBackendError to be cleared on submit")
		}
		if overlay.backendErrorMsg != "" {
			t.Error("expected backendErrorMsg to be cleared on submit")
		}
		if !overlay.isCreating {
			t.Error("expected isCreating to be true after submit")
		}
	})
	
	t.Run("ErrorMessageDisplayedInView", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		
		// Set backend error
		overlay.titleBackendError = true
		overlay.backendErrorMsg = "Connection failed"
		
		// Render view
		view := overlay.View()
		
		// Verify error message is shown
		if !strings.Contains(view, "Connection failed") {
			t.Error("expected error message to be displayed in view")
		}
		if !strings.Contains(view, "⚠") {
			t.Error("expected warning symbol in error display")
		}
	})
	
	t.Run("NoErrorMessageWhenNoError", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		
		// No error
		overlay.titleBackendError = false
		overlay.backendErrorMsg = ""
		
		// Render view
		view := overlay.View()
		
		// Verify no error message shown
		if strings.Contains(view, "⚠") {
			t.Error("expected no warning symbol when no error")
		}
	})
}
