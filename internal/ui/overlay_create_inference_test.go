package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

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
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

		// Should not activate flash since type didn't change
		if overlay.typeInferenceActive {
			t.Error("expected no flash when type unchanged")
		}
	})
}
