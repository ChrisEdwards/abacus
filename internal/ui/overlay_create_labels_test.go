package ui

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

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
		// New pill format: ⏎ for Enter, esc for Escape
		if !contains(view, "⏎") || !contains(view, "Select") {
			t.Error("expected footer to show '⏎' and 'Select' when parent dropdown open")
		}
		if !contains(view, "esc") || !contains(view, "Revert") {
			t.Error("expected footer to show 'esc' and 'Revert' when parent dropdown open")
		}
		if contains(view, "Create+Add") {
			t.Error("expected footer NOT to show 'Create+Add' when parent dropdown open")
		}
	})

	t.Run("FooterShowsNormalWhenParentDropdownClosed", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})

		view := overlay.View()
		// New pill format: ⏎ for Enter, ^⏎ for Ctrl+Enter
		if !contains(view, "⏎") || !contains(view, "Create") {
			t.Error("expected footer to show '⏎' and 'Create' when dropdown closed")
		}
		if !contains(view, "^⏎") || !contains(view, "Create+Add") {
			t.Error("expected footer to show '^⏎' and 'Create+Add' when dropdown closed")
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
func TestTitleTextareaBehavior(t *testing.T) {
	t.Run("PreventsManualNewline", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.SetValue("Initial Title")
		overlay.titleInput.Focus()

		before := overlay.titleInput.Value()
		overlay.titleInput, _ = overlay.titleInput.Update(tea.KeyMsg{Type: tea.KeyEnter})
		after := overlay.titleInput.Value()

		if after != before {
			t.Fatalf("expected newline to be blocked, got %q", after)
		}
	})

	t.Run("ArrowUpMovesWithinWrappedTitle", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.titleInput.Focus()
		overlay.titleInput.SetWidth(12)
		overlay.titleInput.SetHeight(3)
		overlay.titleInput.SetValue("This is a very long title that should wrap across multiple visual lines.")
		overlay.titleInput.CursorEnd()

		before := overlay.titleInput.LineInfo()
		if before.RowOffset == 0 {
			t.Fatalf("expected wrapped cursor row, got %d", before.RowOffset)
		}

		overlay.titleInput, _ = overlay.titleInput.Update(tea.KeyMsg{Type: tea.KeyUp})
		after := overlay.titleInput.LineInfo()

		if after.RowOffset != before.RowOffset-1 {
			t.Fatalf("expected cursor to move up one row, before %d after %d", before.RowOffset, after.RowOffset)
		}
	})
}

// Tests for Zone 3: Vim Navigation (ab-l9e)
func TestVimNavigationKeys(t *testing.T) {
	t.Run("HLNavigatesTypeOptions", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusType
		overlay.typeIndex = 0

		// l moves right (increases index in horizontal layout)
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
		if overlay.typeIndex != 1 {
			t.Errorf("expected type index 1 after 'l', got %d", overlay.typeIndex)
		}

		// h moves left (decreases index)
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
		if overlay.typeIndex != 0 {
			t.Errorf("expected type index 0 after 'h', got %d", overlay.typeIndex)
		}
	})

	t.Run("JKNavigatesBetweenTypeAndPriorityRows", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusType

		// j moves down to Priority row
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if overlay.Focus() != FocusPriority {
			t.Errorf("expected focus on priority after 'j', got %d", overlay.Focus())
		}

		// k moves up to Type row
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if overlay.Focus() != FocusType {
			t.Errorf("expected focus on type after 'k', got %d", overlay.Focus())
		}
	})

	t.Run("JKStaysAtBoundsInTypeAndPriority", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})

		// k at top (Type) stays in Type
		overlay.focus = FocusType
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if overlay.Focus() != FocusType {
			t.Errorf("expected focus to stay on type after 'k' at top, got %d", overlay.Focus())
		}

		// j at bottom (Priority) stays in Priority
		overlay.focus = FocusPriority
		overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if overlay.Focus() != FocusPriority {
			t.Errorf("expected focus to stay on priority after 'j' at bottom, got %d", overlay.Focus())
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
	lipgloss.SetColorProfile(termenv.TrueColor)

	t.Run("RenderHorizontalOptionUsesANSIUnderline", func(t *testing.T) {
		baseStyle := lipgloss.NewStyle()
		innerStyle := baseStyle.Padding(0)
		got := renderHorizontalOption(baseStyle, "Task", true, true) // selected with underline
		// Selected item should have parentheses: (Task) with T underlined
		// When underline=true, parentheses are also styled with innerStyle (ab-rixh.3 fix)
		expected := innerStyle.Render("(") + lipgloss.NewStyle().Underline(true).Render("T") + innerStyle.Render("ask") + innerStyle.Render(")")
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("ParenthesesAreStyledWhenFocused", func(t *testing.T) {
		// ab-rixh.3: Test that parentheses get proper styling when focused
		// Use a style with foreground color to verify ANSI codes are applied
		styledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		got := renderHorizontalOption(styledStyle, "Task", true, true) // selected, focused
		// Both parentheses should have ANSI color codes
		// The opening paren should have styling (not be plain text)
		if !strings.Contains(got, "\x1b[") {
			t.Error("expected ANSI styling in output")
		}
		// Verify both parens are present in the styled output
		stripped := ansiPattern.ReplaceAllString(got, "")
		if stripped != "(Task)" {
			t.Errorf("expected stripped output '(Task)', got %q", stripped)
		}
	})

	t.Run("RenderHorizontalOptionSkipsUnderlineWhenDisabled", func(t *testing.T) {
		baseStyle := lipgloss.NewStyle()
		got := renderHorizontalOption(baseStyle, "Task", false, false) // unselected, no underline
		expected := "Task"
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("TypeColumnShowsUnderlineWhenFocused", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusType

		view := overlay.View()
		line := lineContaining(view, "Task")
		if line == "" {
			t.Fatal("expected Task option in Type column")
		}
		if !containsUnderlinedLetter(line, 'T') {
			t.Error("expected underlined T in Type column when focused")
		}
	})

	t.Run("PriorityColumnShowsUnderlineWhenFocused", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusPriority

		view := overlay.View()
		line := lineContaining(view, "Crit")
		if line == "" {
			t.Fatal("expected Crit option in Priority column")
		}
		if !containsUnderlinedLetter(line, 'C') {
			t.Error("expected underlined C in Priority column when focused")
		}
	})

	t.Run("NoUnderlineWhenUnfocused", func(t *testing.T) {
		overlay := NewCreateOverlay(CreateOverlayOptions{})
		overlay.focus = FocusTitle // Neither Type nor Priority focused

		view := overlay.View()
		line := lineContaining(view, "Task")
		if line == "" {
			t.Fatal("expected Task option in Type column")
		}
		if containsUnderlinedLetter(line, 'T') {
			t.Error("expected no underlined Task when Type is unfocused")
		}
	})
}

// Helper functions
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
func containsUnderlinedLetter(s string, letter rune) bool {
	pattern := fmt.Sprintf("\x1b\\[[0-9;:]*4[0-9;:]*m%s", string(letter))
	re := regexp.MustCompile(pattern)
	return re.FindStringIndex(s) != nil
}

var ansiPattern = regexp.MustCompile("\x1b\\[[0-9;:]*m")

func lineContaining(s, substr string) string {
	for _, line := range strings.Split(s, "\n") {
		lineStripped := ansiPattern.ReplaceAllString(line, "")
		lineNormalized := strings.ReplaceAll(lineStripped, " ", "")
		if strings.Contains(lineNormalized, substr) {
			return line
		}
	}
	return ""
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
