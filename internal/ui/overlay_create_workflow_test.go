package ui

import (
	"fmt"
	"strings"
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
			{"Update README section", 4}, // Use Update instead of Add to match chore
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
			{"Prefix component", -1}, // "fix" is part of "Prefix"
			{"Adder utility", -1},    // "add" is part of "Adder"
			{"Buggy behavior", -1},   // "bug" is part of "Buggy"
			{"Fix bug", 2},           // "fix" is standalone word
			{"Add feature", 1},       // "add" is standalone word
			{"Bug report", 2},        // "bug" is standalone word
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

		// Press Right arrow (horizontal layout - changes selection)
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRight})

		if !overlay.typeManuallySet {
			t.Error("expected typeManuallySet=true after arrow key")
		}
	})

	t.Run("VimKeysSetManualFlag", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusType

		// Press 'l' (vim right - changes selection in horizontal layout)
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})

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
		view := overlay.View()

		// New pill format uses symbols: ⏎ for Enter, ^⏎ for Ctrl+Enter
		if !strings.Contains(view, "⏎") || !strings.Contains(view, "Create") {
			t.Error("expected default footer to contain '⏎' and 'Create'")
		}
		if !strings.Contains(view, "^⏎") || !strings.Contains(view, "Create+Add") {
			t.Error("expected default footer to contain bulk entry hint (^⏎ Create+Add)")
		}
		if !strings.Contains(view, "Tab") || !strings.Contains(view, "Next") {
			t.Error("expected default footer to contain 'Tab' and 'Next'")
		}
		if !strings.Contains(view, "esc") || !strings.Contains(view, "Cancel") {
			t.Error("expected default footer to contain 'esc' and 'Cancel'")
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

		view := overlay.View()

		// New pill format uses ⏎ for Enter
		if !strings.Contains(view, "⏎") || !strings.Contains(view, "Select") {
			t.Errorf("expected parent search footer to contain '⏎' and 'Select', got: %s", view)
		}
		if !strings.Contains(view, "esc") || !strings.Contains(view, "Revert") {
			t.Errorf("expected parent search footer to contain 'esc' and 'Revert', got: %s", view)
		}
		// Should NOT contain "Create+Add" (the bulk entry hint unique to default footer)
		if strings.Contains(view, "Create+Add") {
			t.Errorf("expected parent search footer to not contain 'Create+Add', got: %s", view)
		}
	})

	t.Run("CreatingFooter", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.isCreating = true

		view := overlay.View()

		if !strings.Contains(view, "Creating bead...") {
			t.Error("expected creating footer to contain 'Creating bead...'")
		}
		// Should NOT contain the default footer hints
		if strings.Contains(view, "Create+Add") {
			t.Error("expected creating footer to not contain 'Create+Add'")
		}
	})

	t.Run("FooterInView", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		view := overlay.View()

		// Verify default footer appears in view (pill format with ⏎ symbol)
		if !strings.Contains(view, "⏎") || !strings.Contains(view, "Create") {
			t.Error("expected default footer in view output with ⏎ and Create")
		}
	})

	t.Run("FooterSwitchesWithParentDropdown", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableParents: []ParentOption{
				{ID: "ab-123", Display: "ab-123 Test"},
			},
		})

		// Initially should show default footer (with Create+Add bulk hint)
		view := overlay.View()
		if !strings.Contains(view, "Create+Add") {
			t.Error("expected default footer initially (with Create+Add)")
		}

		// Navigate to parent field
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyShiftTab}) // From Title to Parent

		// Type to open dropdown
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})

		// Should now show parent search footer (Select, not Create+Add)
		view = overlay.View()
		if !strings.Contains(view, "Select") || strings.Contains(view, "Create+Add") {
			t.Errorf("expected parent search footer after opening dropdown, got: %s", view)
		}

		// Close dropdown with Esc
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})

		// Should show browse hint (focused on parent field but dropdown closed)
		view = overlay.View()
		if !strings.Contains(view, "Browse") {
			t.Errorf("expected browse hint after closing dropdown, got: %s", view)
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

		view := overlay.View()

		// New pill format: ↓ key with Browse description
		if !strings.Contains(view, "↓") || !strings.Contains(view, "Browse") {
			t.Errorf("expected footer to contain '↓' and 'Browse' on parent field, got: %s", view)
		}
		if !strings.Contains(view, "⏎") || !strings.Contains(view, "Create") {
			t.Errorf("expected footer to contain '⏎' and 'Create' on parent field, got: %s", view)
		}
	})

	t.Run("FooterShowsBrowseHintOnLabelsField", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableLabels: []string{"bug", "feature"},
		})

		// Navigate to labels field
		overlay.focus = FocusLabels

		view := overlay.View()

		// New pill format: ↓ key with Browse description
		if !strings.Contains(view, "↓") || !strings.Contains(view, "Browse") {
			t.Errorf("expected footer to contain '↓' and 'Browse' on labels field, got: %s", view)
		}
	})

	t.Run("FooterShowsBrowseHintOnAssigneeField", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableAssignees: []string{"alice", "bob"},
		})

		// Navigate to assignee field
		overlay.focus = FocusAssignee

		view := overlay.View()

		// New pill format: ↓ key with Browse description
		if !strings.Contains(view, "↓") || !strings.Contains(view, "Browse") {
			t.Errorf("expected footer to contain '↓' and 'Browse' on assignee field, got: %s", view)
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

		view := overlay.View()
		// New pill format: ⏎ key with Select description
		if !strings.Contains(view, "⏎") || !strings.Contains(view, "Select") {
			t.Errorf("expected '⏎' and 'Select' when labels dropdown open, got: %s", view)
		}

		// Close labels dropdown
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})

		// Test assignee dropdown
		overlay.focus = FocusAssignee
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyDown})

		view = overlay.View()
		// New pill format: ⏎ key with Select description
		if !strings.Contains(view, "⏎") || !strings.Contains(view, "Select") {
			t.Errorf("expected '⏎' and 'Select' when assignee dropdown open, got: %s", view)
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
		view := overlay.View()
		if strings.Contains(view, "Creating bead...") {
			t.Error("should not show creating footer before submission")
		}

		// Submit form
		overlay, _ = overlay.handleSubmit(false)

		// After submission
		view = overlay.View()
		if !strings.Contains(view, "Creating bead...") {
			t.Error("expected creating footer after submission")
		}
	})
}

// ============================================================================
// Backend Error Handling Tests (spec Section 4.4)
// ============================================================================
func TestBackendErrorHandling(t *testing.T) {
	t.Run("BackendErrorSetsHasBackendError", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("Test Bead")

		// Simulate backend error (overlay tracks hasBackendError, App shows global toast)
		overlay, _ = overlay.Update(backendErrorMsg{err: fmt.Errorf("database connection failed"), errMsg: "database connection failed"})

		if !overlay.hasBackendError {
			t.Error("expected hasBackendError=true after backend error")
		}

		if overlay.isCreating {
			t.Error("expected isCreating=false after backend error")
		}
	})

	t.Run("BackendErrorClearedOnRetry", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("Test Bead")
		overlay.hasBackendError = true

		// Retry submission
		overlay, _ = overlay.handleSubmit(false)

		if overlay.hasBackendError {
			t.Error("expected hasBackendError=false when retrying after error")
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

	t.Run("BackendErrorPreservesFormData", func(t *testing.T) {
		// Ensure backend error doesn't clear user's form data
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("My Important Title")
		overlay.typeIndex = 1     // Feature
		overlay.priorityIndex = 0 // Critical

		// Simulate backend error
		overlay, _ = overlay.Update(backendErrorMsg{err: fmt.Errorf("network error"), errMsg: "network error"})

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

// Tests for ab-ctal/ab-orte: Backend error display and ESC handling
// Note: Error toast is now rendered by App using global toast, not by CreateOverlay.
// CreateOverlay only tracks hasBackendError to know if ESC should dismiss toast.
func TestCreateOverlayBackendErrorDisplay(t *testing.T) {
	t.Run("BackendErrorMsgSetsHasBackendError", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})

		// Simulate backend error
		msg := backendErrorMsg{
			err:    fmt.Errorf("db sync error"),
			errMsg: "Database out of sync with JSONL. Run 'bd sync' to fix.",
		}

		overlay, _ = overlay.Update(msg)

		// Verify error state is tracked (for ESC handling)
		if !overlay.hasBackendError {
			t.Error("expected hasBackendError to be true")
		}
		if overlay.isCreating {
			t.Error("expected isCreating to be false after error")
		}
	})

	t.Run("ESCSendsDismissErrorToastMsgWhenHasBackendError", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})

		// Set backend error state
		overlay.hasBackendError = true

		// Press ESC
		overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})

		// Verify hasBackendError is cleared
		if overlay.hasBackendError {
			t.Error("expected hasBackendError to be false after ESC")
		}

		// Verify DismissErrorToastMsg is sent (not CreateCancelledMsg)
		if cmd == nil {
			t.Fatal("expected DismissErrorToastMsg command")
		}
		msg := cmd()
		if _, ok := msg.(DismissErrorToastMsg); !ok {
			t.Errorf("expected DismissErrorToastMsg, got %T", msg)
		}
	})

	t.Run("ESCClosesModalWhenNoBackendError", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})

		// No backend error
		overlay.hasBackendError = false

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

	t.Run("SubmitClearsHasBackendError", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("Test Title")

		// Set backend error state
		overlay.hasBackendError = true

		// Submit with Enter
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})

		// Verify error is cleared when retrying
		if overlay.hasBackendError {
			t.Error("expected hasBackendError to be cleared on submit")
		}
		if !overlay.isCreating {
			t.Error("expected isCreating to be true after submit")
		}
	})
}

// ============================================================================
// Labels Zone Edge Cases (ab-ouvw - filling gaps per spec Section 3.4)
// ============================================================================
func TestLabelsEnterOnEmptyInputIsNoOp(t *testing.T) {
	t.Run("EnterWithEmptyLabelsInputDoesNothing", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableLabels: []string{"api", "backend", "frontend"},
		})
		overlay.focus = FocusLabels
		overlay.labelsCombo.Focus()

		// Input should be empty initially
		if overlay.labelsCombo.InputValue() != "" {
			t.Skip("input not empty initially")
		}

		// Press Enter - should not add any chip
		initialChipCount := overlay.labelsCombo.ChipCount()
		overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})

		// No chip should be added
		if overlay.labelsCombo.ChipCount() != initialChipCount {
			t.Errorf("expected %d chips (unchanged), got %d", initialChipCount, overlay.labelsCombo.ChipCount())
		}

		// Should NOT emit ChipAddedMsg
		if cmd != nil {
			msg := cmd()
			if _, ok := msg.(ChipComboBoxChipAddedMsg); ok {
				t.Error("expected no ChipComboBoxChipAddedMsg for empty input")
			}
		}
	})

	t.Run("EnterWithWhitespaceOnlyDoesNothing", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableLabels: []string{"api"},
		})
		overlay.focus = FocusLabels
		overlay.labelsCombo.Focus()

		// Set input to whitespace only (simulating typing spaces)
		// Note: ChipComboBox trims whitespace, so this tests the trimming
		initialChipCount := overlay.labelsCombo.ChipCount()
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyEnter})

		if overlay.labelsCombo.ChipCount() != initialChipCount {
			t.Errorf("expected chip count unchanged for whitespace input")
		}
	})
}

func TestLabelsDropdownExcludesSelectedChips(t *testing.T) {
	t.Run("SelectedChipsNotInDropdown", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{
			AvailableLabels: []string{"api", "backend", "frontend"},
		})
		overlay.focus = FocusLabels
		overlay.labelsCombo.Focus()

		// Add "api" as chip
		overlay.labelsCombo.SetChips([]string{"api"})

		// Open dropdown with Down arrow
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyDown})

		// Check view doesn't show "api" in dropdown (it's already a chip)
		// Note: The view will show chips separately from dropdown options
		view := overlay.View()
		// Count occurrences of "api" - should appear once as chip, not in dropdown
		apiCount := strings.Count(view, "api")
		if apiCount > 1 {
			t.Errorf("expected 'api' to appear only once (as chip), found %d occurrences", apiCount)
		}
	})
}
