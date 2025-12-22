package ui

import (
	"abacus/internal/beads"
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
func TestNewEditOverlay(t *testing.T) {
	parentDisplay := "ab-parent Parent"
	bead := &beads.FullIssue{
		ID:          "ab-123",
		Title:       "Test Title",
		Description: "Test Description",
		IssueType:   "bug",
		Priority:    2,
		Labels:      []string{"urgent", "backend"},
		Assignee:    "alice",
	}

	opts := CreateOverlayOptions{
		DefaultParentID: "ab-parent",
		AvailableParents: []ParentOption{
			{ID: "ab-parent", Display: parentDisplay},
		},
		AvailableLabels:    []string{"urgent", "backend", "frontend"},
		AvailableAssignees: []string{"alice"},
	}

	m := NewEditOverlay(bead, opts)

	if !m.isEditMode() {
		t.Fatal("expected edit mode to be true")
	}
	if got := m.Title(); got != "Test Title" {
		t.Errorf("expected title %q, got %q", "Test Title", got)
	}
	if got := m.Description(); got != "Test Description" {
		t.Errorf("expected description %q, got %q", "Test Description", got)
	}
	if got := m.IssueType(); got != "bug" {
		t.Errorf("expected issue type bug, got %s", got)
	}
	if got := m.Priority(); got != 2 {
		t.Errorf("expected priority 2, got %d", got)
	}
	if got := m.ParentID(); got != "ab-parent" {
		t.Errorf("expected parent ID ab-parent, got %s", got)
	}
	if chips := m.labelsCombo.GetChips(); len(chips) != 2 || chips[0] != "urgent" || chips[1] != "backend" {
		t.Errorf("expected labels [urgent backend], got %v", chips)
	}
	if m.assigneeCombo.Value() != "alice" {
		t.Errorf("expected assignee value alice, got %s", m.assigneeCombo.Value())
	}
	if header := m.header(); header != "EDIT: ab-123" {
		t.Errorf("expected header to show EDIT with ID, got %q", header)
	}
	if action := m.submitFooterText(); action != "Save" {
		t.Errorf("expected submit footer text 'Save', got %q", action)
	}
	if m.editingBeadParentID != "ab-parent" {
		t.Errorf("expected original parent ID ab-parent, got %s", m.editingBeadParentID)
	}
}
func TestEditOverlayShowsParentComboForRoot(t *testing.T) {
	bead := &beads.FullIssue{ID: "ab-root", IssueType: "task"}
	opts := CreateOverlayOptions{
		AvailableParents: []ParentOption{
			{ID: "ab-parent", Display: "ab-parent Parent"},
		},
	}
	m := NewEditOverlay(bead, opts)

	if m.isRootMode {
		t.Fatal("expected isRootMode to be false for edit mode root bead")
	}

	view := m.View()
	if !strings.Contains(view, "PARENT") {
		t.Fatal("expected view to contain PARENT label")
	}
	if m.ParentID() != "" {
		t.Fatalf("expected ParentID to be empty for root bead, got %s", m.ParentID())
	}
	if !strings.Contains(view, "No Parent (Root Item)") {
		t.Fatal("expected placeholder text for root in parent combo")
	}
}
func TestSubmitEditBuildsMessage(t *testing.T) {
	bead := &beads.FullIssue{ID: "ab-42"}
	opts := CreateOverlayOptions{
		DefaultParentID: "ab-parent",
		AvailableParents: []ParentOption{
			{ID: "ab-parent", Display: "ab-parent Parent"},
		},
	}
	m := NewEditOverlay(bead, opts)
	m.titleInput.SetValue(" Updated Title ")
	m.descriptionInput.SetValue("desc")
	m.typeIndex = 1     // feature
	m.priorityIndex = 3 // low
	m.labelsCombo.SetChips([]string{"backend"})
	m.assigneeCombo.SetValue("bob")

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command on submit")
	}
	msg := cmd()
	updateMsg, ok := msg.(BeadUpdatedMsg)
	if !ok {
		t.Fatalf("expected BeadUpdatedMsg, got %T", msg)
	}
	if updateMsg.ID != "ab-42" {
		t.Errorf("expected ID ab-42, got %s", updateMsg.ID)
	}
	if updateMsg.Title != "Updated Title" {
		t.Errorf("expected trimmed title, got %q", updateMsg.Title)
	}
	if updateMsg.IssueType != "feature" {
		t.Errorf("expected issue type feature, got %s", updateMsg.IssueType)
	}
	if updateMsg.Priority != 3 {
		t.Errorf("expected priority 3, got %d", updateMsg.Priority)
	}
	if updateMsg.ParentID != "ab-parent" {
		t.Errorf("expected parentID ab-parent, got %s", updateMsg.ParentID)
	}
	if len(updateMsg.Labels) != 1 || updateMsg.Labels[0] != "backend" {
		t.Errorf("expected labels [backend], got %v", updateMsg.Labels)
	}
	if updateMsg.Assignee != "bob" {
		t.Errorf("expected assignee bob, got %s", updateMsg.Assignee)
	}
	if updateMsg.OriginalParentID != "ab-parent" {
		t.Errorf("expected original parent ab-parent, got %s", updateMsg.OriginalParentID)
	}
}
func TestSubmitEditWithUnassignedAssignee(t *testing.T) {
	bead := &beads.FullIssue{ID: "ab-42"}
	m := NewEditOverlay(bead, CreateOverlayOptions{})
	m.titleInput.SetValue("Title")
	m.assigneeCombo.SetValue("Unassigned")

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command")
	}
	msg := cmd()
	updateMsg := msg.(BeadUpdatedMsg)
	if updateMsg.Assignee != "" {
		t.Errorf("expected empty assignee for Unassigned, got %q", updateMsg.Assignee)
	}
}
func TestTypeIndexFromString(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"task", 0},
		{"feature", 1},
		{"bug", 2},
		{"epic", 3},
		{"chore", 4},
		{"unknown", 0},
		{"", 0},
	}

	for _, tt := range tests {
		if got := typeIndexFromString(tt.input); got != tt.want {
			t.Errorf("typeIndexFromString(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
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
		// Tab: Title -> Description
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyTab})
		if overlay.Focus() != FocusDescription {
			t.Errorf("expected focus on description, got %d", overlay.Focus())
		}
		// Tab: Description -> Type
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
		// Tab: Labels -> Assignee (direct, no longer async via ChipComboBoxTabMsg)
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyTab})
		if overlay.Focus() != FocusAssignee {
			t.Errorf("expected focus on assignee, got %d", overlay.Focus())
		}
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

	t.Run("LeftRightChangesType", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusType
		if overlay.typeIndex != 0 {
			t.Error("expected initial type index 0")
		}
		// Right arrow increases index (horizontal layout)
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRight})
		if overlay.typeIndex != 1 {
			t.Errorf("expected type index 1, got %d", overlay.typeIndex)
		}
		// Left arrow decreases index
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyLeft})
		if overlay.typeIndex != 0 {
			t.Errorf("expected type index 0, got %d", overlay.typeIndex)
		}
	})

	t.Run("LeftRightChangesPriority", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusPriority
		if overlay.priorityIndex != 2 {
			t.Errorf("expected initial priority index 2 (Med), got %d", overlay.priorityIndex)
		}
		// Right arrow increases index (horizontal layout)
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRight})
		if overlay.priorityIndex != 3 {
			t.Errorf("expected priority index 3, got %d", overlay.priorityIndex)
		}
		// Left arrow decreases index
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyLeft})
		if overlay.priorityIndex != 2 {
			t.Errorf("expected priority index 2, got %d", overlay.priorityIndex)
		}
	})

	t.Run("UpDownNavigatesBetweenTypeAndPriority", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusType
		// Down arrow moves to priority (vertical row navigation)
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyDown})
		if overlay.Focus() != FocusPriority {
			t.Errorf("expected focus on priority, got %d", overlay.Focus())
		}
		// Up arrow moves back to type
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyUp})
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
		if !contains(view, "TYPE") {
			t.Error("expected view to contain TYPE column")
		}
		if !contains(view, "PRIORITY") {
			t.Error("expected view to contain PRIORITY column")
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
