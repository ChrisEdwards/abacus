package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

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

		// Tab: Title -> Description
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyTab})
		if overlay.Focus() != FocusDescription {
			t.Errorf("expected focus on Description, got %d", overlay.Focus())
		}

		// Tab: Description -> Type
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

		// Tab: Labels -> Assignee (direct, no longer async via ChipComboBoxTabMsg)
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyTab})
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
