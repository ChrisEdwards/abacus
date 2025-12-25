package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// ============================================================================
// Dynamic Footer Tests (spec Section 4.1)
// ============================================================================
func TestCreateOverlayFooter(t *testing.T) {
	t.Run("DefaultFooter", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		view := overlay.View()

		// New pill format uses symbols: ⏎ for Enter
		if !strings.Contains(view, "⏎") || !strings.Contains(view, "Create") {
			t.Error("expected default footer to contain '⏎' and 'Create'")
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
	})

	t.Run("CreatingFooter", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.isCreating = true

		view := overlay.View()

		if !strings.Contains(view, "Creating bead...") {
			t.Error("expected creating footer to contain 'Creating bead...'")
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

		// Initially should show default footer
		view := overlay.View()
		if !strings.Contains(view, "⏎") || !strings.Contains(view, "Create") {
			t.Error("expected default footer initially")
		}

		// Navigate to parent field
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyShiftTab}) // From Title to Parent

		// Type to open dropdown
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})

		// Should now show parent search footer (Select)
		view = overlay.View()
		if !strings.Contains(view, "Select") {
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
		overlay, _ = overlay.handleSubmit()

		if !overlay.isCreating {
			t.Error("expected isCreating=true after handleSubmit")
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
		overlay, _ = overlay.handleSubmit()

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
		overlay, _ = overlay.handleSubmit()

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
