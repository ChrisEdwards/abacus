package ui

import (
	"abacus/internal/beads"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

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
		if overlay.Focus() != FocusDescription {
			t.Errorf("expected Type -> Description, got %d", overlay.Focus())
		}

		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
		if overlay.Focus() != FocusTitle {
			t.Errorf("expected Description -> Title, got %d", overlay.Focus())
		}
	})
}

func TestEditOverlaySkipsTypeField(t *testing.T) {
	overlay := NewEditOverlay(&beads.FullIssue{ID: "ab-1", IssueType: "task"}, CreateOverlayOptions{})

	// Tab sequence should skip type and go Title -> Description -> Priority
	if overlay.focus != FocusTitle {
		t.Fatalf("expected initial focus on title, got %d", overlay.focus)
	}
	overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyTab}) // to Description
	overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyTab}) // should go to Priority (skip Type)
	if overlay.focus != FocusPriority {
		t.Fatalf("expected focus on Priority after skipping Type in edit mode, got %d", overlay.focus)
	}

	// View should not contain TYPE label
	if contains(overlay.View(), "TYPE") {
		t.Fatal("expected edit overlay view to hide TYPE row")
	}
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
		overlay.typeIndex = 1     // feature
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
