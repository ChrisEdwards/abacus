package ui

import (
	"abacus/internal/beads"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// Flash duration for validation errors (spec Section 4.4: 300ms)
const (
	titleFlashDuration = 300 * time.Millisecond
	// titleContentWidth is the width of the title textarea content area.
	// Width 40 = 44 (styleCreateInput total) - 2 (border) - 2 (padding)
	titleContentWidth = 40
)

// titleFlashClearMsg clears the title validation error flash.
type titleFlashClearMsg struct{}

// titleFlashCmd returns a command that clears the title flash after a delay.
func titleFlashCmd() tea.Cmd {
	return tea.Tick(titleFlashDuration, func(_ time.Time) tea.Msg {
		return titleFlashClearMsg{}
	})
}

// backendErrorMsg signals backend error during create.
type backendErrorMsg struct {
	err    error
	errMsg string // Human-readable error message to display
}

// typeInferenceFlashMsg signals the type was auto-inferred (spec Section 5).
type typeInferenceFlashMsg struct{}

// typeInferenceFlashCmd returns a command that clears the type inference flash after a delay.
func typeInferenceFlashCmd() tea.Cmd {
	return tea.Tick(flashDuration, func(_ time.Time) tea.Msg {
		return typeInferenceFlashMsg{}
	})
}

// bulkEntryResetMsg signals to reset for next bulk entry (spec Section 4.3).
type bulkEntryResetMsg struct{}

// CreateFocus represents which zone has focus in the create overlay.
type CreateFocus int

// Focus zones in tab order (spec Section 6)
const (
	FocusParent CreateFocus = iota
	FocusTitle
	FocusDescription
	FocusType
	FocusPriority
	FocusLabels
	FocusAssignee
)

// Type options
var typeOptions = []string{"task", "feature", "bug", "epic", "chore"}
var typeLabels = []string{"Task", "Feature", "Bug", "Epic", "Chore"}

// Priority options
var priorityLabels = []string{"Crit", "High", "Med", "Low", "Back"}

// Type inference patterns for auto-detecting bead type from title (spec Section 5).
// Patterns are checked in order; first match wins.
var typeInferencePatterns = []struct {
	pattern *regexp.Regexp
	typeIdx int
}{
	// Bug patterns - check first (most specific)
	{regexp.MustCompile(`(?i)\b(fix|broken|bug|error|crash|issue with)\b`), 2},
	// Feature patterns
	{regexp.MustCompile(`(?i)\b(add|implement|create|build|new)\b`), 1},
	// Chore patterns
	{regexp.MustCompile(`(?i)\b(refactor|clean up|reorganize|simplify|extract)\b`), 4},
	{regexp.MustCompile(`(?i)\b(update|upgrade|bump|migrate)\b`), 4},
	{regexp.MustCompile(`(?i)\b(document|docs|readme)\b`), 4},
}

// inferTypeFromTitle analyzes the title and returns the inferred type index.
// Returns -1 if no pattern matches.
// Uses word boundaries (\b) to avoid false matches (e.g., "Prefix" won't match "fix").
// First match wins: returns the type for the keyword that appears earliest in the title.
func inferTypeFromTitle(title string) int {
	title = strings.TrimSpace(title)
	if title == "" {
		return -1
	}

	// Find the earliest match by position (spec Section 5: first match wins)
	earliestPos := -1
	earliestType := -1

	for _, p := range typeInferencePatterns {
		loc := p.pattern.FindStringIndex(title)
		if loc != nil {
			// loc[0] is the start position of the match
			if earliestPos == -1 || loc[0] < earliestPos {
				earliestPos = loc[0]
				earliestType = p.typeIdx
			}
		}
	}

	return earliestType
}

// CreateOverlay is a 5-zone HUD for creating a new bead.
// See docs/CREATE_BEAD_SPEC.md Section 3 for zone layout.
type CreateOverlay struct {
	editingBead *beads.FullIssue // nil = create mode, non-nil = edit mode
	// Track original parent for edits to manage dependencies
	editingBeadParentID string

	// Focus management
	focus CreateFocus

	// Zone 1: Parent (anchor at top)
	parentCombo     ComboBox
	parentOptions   []ParentOption
	isRootMode      bool   // True if opened with 'N' (no parent)
	defaultParentID string // Pre-selected parent ID
	parentOriginal  string // Value when Parent field focused, for Esc revert (spec Section 12)

	// Zone 2: Title (hero element)
	titleInput           textarea.Model
	titleValidationError bool // True when flashing red for validation
	hasBackendError      bool // True when backend error occurred (for ESC handling)

	// Zone 2b: Description (multi-line textarea)
	descriptionInput textarea.Model

	// Zone 3: Properties (2-column grid)
	typeIndex           int
	priorityIndex       int
	typeManuallySet     bool // Disables auto-inference when true
	typeInferenceActive bool // True during flash animation (150ms)

	// Zone 4: Labels (multi-select chips)
	labelsCombo   ChipComboBox
	labelsOptions []string

	// Zone 5: Assignee (single-select combo)
	assigneeCombo   ComboBox
	assigneeOptions []string

	// State management
	isCreating bool // True during form submission (spec Section 4.1)

	// Responsive sizing (ab-11wd)
	termWidth int // Terminal width for responsive dialog width calculation
}

// ParentOption represents a bead that can be selected as a parent.
type ParentOption struct {
	ID      string
	Display string // e.g., "ab-83s Create and Edit..."
}

// BeadCreatedMsg is sent when form submission is confirmed.
type BeadCreatedMsg struct {
	Title       string
	Description string
	IssueType   string
	Priority    int
	ParentID    string
	Labels      []string // Selected labels (backend integration in ab-l1k)
	Assignee    string   // Selected assignee (backend integration in ab-39r)
	StayOpen    bool     // true for Ctrl+Enter bulk entry (spec Section 4.3)
}

// CreateCancelledMsg is sent when the overlay is dismissed without action.
type CreateCancelledMsg struct{}

// NewLabelAddedMsg signals a new label was created (not in existing options).
// Used to trigger a toast notification.
type NewLabelAddedMsg struct {
	Label string
}

// NewAssigneeAddedMsg signals a new assignee was created (not in existing options).
// Used to trigger a toast notification.
type NewAssigneeAddedMsg struct {
	Assignee string
}

// CreateOverlayOptions configures the create overlay.
type CreateOverlayOptions struct {
	DefaultParentID    string         // Pre-selected parent (empty for root)
	AvailableParents   []ParentOption // All beads that can be parents
	AvailableLabels    []string       // All labels from existing beads
	AvailableAssignees []string       // All assignees from existing beads
	IsRootMode         bool           // True if opened with 'N' (no parent)
}

// NewCreateOverlay creates a new 5-zone create overlay.
func NewCreateOverlay(opts CreateOverlayOptions) *CreateOverlay {
	// Zone 2: Title input (hero element) - textarea for native wrapping behavior
	// Starts as single line, expands up to 3 lines based on content
	ti := textarea.New()
	ti.Placeholder = ""
	ti.Prompt = ""
	ti.ShowLineNumbers = false
	ti.CharLimit = 100
	ti.SetWidth(titleContentWidth)            // Content width inside border+padding
	ti.SetHeight(1)                           // Start as single line, expands dynamically
	ti.KeyMap.InsertNewline.SetEnabled(false) // Enter submits instead of inserting newlines

	// Zone 2b: Description textarea (multi-line, 5 lines visible)
	desc := textarea.New()
	desc.Placeholder = ""
	desc.SetWidth(44)
	desc.SetHeight(5)
	desc.CharLimit = 2000 // Reasonable limit for descriptions
	desc.ShowLineNumbers = false

	// Zone 1: Parent combo box
	parentDisplays := make([]string, len(opts.AvailableParents))
	for i, p := range opts.AvailableParents {
		parentDisplays[i] = p.Display
	}
	parentCombo := NewComboBox(parentDisplays).
		WithWidth(44).
		WithMaxVisible(5).
		WithPlaceholder("No Parent (Root Item)")

	// Pre-select parent if default exists and not root mode
	parentOriginal := "" // Track original value for Esc revert
	if !opts.IsRootMode && opts.DefaultParentID != "" {
		for _, p := range opts.AvailableParents {
			if p.ID == opts.DefaultParentID {
				parentCombo.SetValue(p.Display)
				parentOriginal = p.Display // Initialize original for Esc revert
				break
			}
		}
	}

	// Zone 4: Labels combo box (multi-select with chips)
	labelsCombo := NewChipComboBox(opts.AvailableLabels).
		WithWidth(44).
		WithMaxVisible(5).
		WithPlaceholder("type to filter...")

	// Zone 5: Assignee combo box (single-select)
	// Prepend "Unassigned" and "Me ($USER)" options per spec Section 3.5
	assigneeOpts := []string{"Unassigned"}
	if user := os.Getenv("USER"); user != "" {
		assigneeOpts = append(assigneeOpts, fmt.Sprintf("Me (%s)", user))
	}
	assigneeOpts = append(assigneeOpts, opts.AvailableAssignees...)
	assigneeCombo := NewComboBox(assigneeOpts).
		WithWidth(44).
		WithMaxVisible(5).
		WithPlaceholder("type to filter...").
		WithAllowNew(true, "New assignee: %s")
	assigneeCombo.SetValue("Unassigned")

	// Focus title input BEFORE assigning to struct (textarea.Model is a value type)
	ti.Focus()

	m := &CreateOverlay{
		focus:            FocusTitle, // Title is auto-focused (spec Section 3.2)
		titleInput:       ti,
		descriptionInput: desc,
		typeIndex:        0, // Task
		priorityIndex:    2, // Medium
		parentCombo:      parentCombo,
		parentOptions:    opts.AvailableParents,
		parentOriginal:   parentOriginal, // Set original for Esc revert
		isRootMode:       opts.IsRootMode,
		defaultParentID:  opts.DefaultParentID,
		labelsCombo:      labelsCombo,
		labelsOptions:    opts.AvailableLabels,
		assigneeCombo:    assigneeCombo,
		assigneeOptions:  opts.AvailableAssignees,
	}

	return m
}

// NewEditOverlay creates a CreateOverlay pre-populated with existing bead data.
func NewEditOverlay(bead *beads.FullIssue, opts CreateOverlayOptions) *CreateOverlay {
	m := NewCreateOverlay(opts)
	m.editingBead = bead
	m.editingBeadParentID = opts.DefaultParentID
	// Edit mode should always show parent picker, even for roots.
	m.isRootMode = false
	m.parentCombo.SetValue("")
	m.parentOriginal = ""

	m.titleInput.SetValue(bead.Title)
	m.descriptionInput.SetValue(bead.Description)
	m.typeIndex = typeIndexFromString(bead.IssueType)
	m.priorityIndex = bead.Priority
	m.typeManuallySet = true

	// Pre-select parent and track original for Esc revert
	if beadParent := opts.DefaultParentID; beadParent != "" && !opts.IsRootMode {
		for _, p := range opts.AvailableParents {
			if p.ID == beadParent {
				m.parentCombo.SetValue(p.Display)
				m.parentOriginal = p.Display
				break
			}
		}
	}

	// Pre-select labels
	m.labelsCombo.SetChips(append([]string{}, bead.Labels...))

	// Pre-select assignee (empty string maps to Unassigned placeholder)
	if bead.Assignee != "" {
		m.assigneeCombo.SetValue(bead.Assignee)
	} else {
		m.assigneeCombo.SetValue("Unassigned")
	}

	return m
}

// isEditMode returns true when the overlay is editing an existing bead.
func (m *CreateOverlay) isEditMode() bool {
	return m.editingBead != nil
}

// header returns the overlay header text based on mode.
// Init implements tea.Model.
func (m *CreateOverlay) Init() tea.Cmd {
	return textarea.Blink
}

// IsTextInputActive reports whether the overlay is currently focused on a text
// entry area (including dropdown search inputs). When true, global hotkeys
// like the error recall should be suppressed.
func (m *CreateOverlay) IsTextInputActive() bool {
	switch m.focus {
	case FocusTitle, FocusDescription, FocusParent, FocusLabels, FocusAssignee:
		return true
	}
	if m.parentCombo.IsDropdownOpen() || m.labelsCombo.IsDropdownOpen() || m.assigneeCombo.IsDropdownOpen() {
		return true
	}
	return false
}

// Update implements tea.Model.
func (m *CreateOverlay) Update(msg tea.Msg) (*CreateOverlay, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle messages from composed components
	switch msg := msg.(type) {
	case titleFlashClearMsg:
		m.titleValidationError = false
		return m, nil

	case backendErrorMsg:
		// Backend error: keep modal open (spec Section 4.4)
		// The App shows the global error toast; we just track state for ESC handling
		m.hasBackendError = true
		m.isCreating = false // Stop showing "Creating..." footer
		// Keep focus on title so user can fix/retry
		return m, nil

	case typeInferenceFlashMsg:
		m.typeInferenceActive = false
		return m, nil

	case bulkEntryResetMsg:
		// Clear title and description, keep other fields persistent (spec Section 4.3)
		m.titleInput.SetValue("")
		m.descriptionInput.SetValue("")
		m.titleValidationError = false
		m.isCreating = false // Clear creating state (spec Section 4.1)
		m.focus = FocusTitle
		return m, m.titleInput.Focus()

	case ChipComboBoxTabMsg:
		// Labels combo requested Tab - move to Assignee
		m.focus = FocusAssignee
		m.labelsCombo.Blur()
		cmds = append(cmds, m.assigneeCombo.Focus())
		return m, tea.Batch(cmds...)

	case ChipComboBoxChipAddedMsg:
		// Chip was added to labels - if new, signal for toast
		if msg.IsNew {
			return m, func() tea.Msg {
				return NewLabelAddedMsg{Label: msg.Label}
			}
		}
		return m, nil

	case ComboBoxEnterSelectedMsg:
		// Forward to labelsCombo if in FocusLabels mode (to add chip)
		if m.focus == FocusLabels {
			var cmd tea.Cmd
			m.labelsCombo, cmd = m.labelsCombo.Update(msg)
			return m, cmd
		}
		// Assignee combo - check if new assignee was created
		if m.focus == FocusAssignee && msg.IsNew {
			return m, func() tea.Msg {
				return NewAssigneeAddedMsg{Assignee: msg.Value}
			}
		}
		// Parent combo - no special action needed
		return m, nil

	case ComboBoxTabSelectedMsg:
		// Always forward to labelsCombo to add chip (it's the only ChipComboBox).
		// Focus may have already moved to Assignee by the time this message arrives,
		// so we can't rely on focus check. The message originated from labelsCombo's
		// internal ComboBox, so route it there unconditionally.
		var cmd tea.Cmd
		m.labelsCombo, cmd = m.labelsCombo.Update(msg)
		// Also check if assignee was new (for toast)
		if msg.IsNew {
			return m, tea.Batch(cmd, func() tea.Msg {
				return NewAssigneeAddedMsg{Assignee: msg.Value}
			})
		}
		return m, cmd

	case tea.KeyMsg:
		// Handle global keys first
		switch msg.Type {
		case tea.KeyEsc:
			return m.handleEscape()

		case tea.KeyEnter:
			// Guard: prevent duplicate submissions (ab-ip2p)
			if m.isCreating {
				return m, nil
			}

			// Ctrl+Enter always submits
			if msg.String() == "ctrl+enter" {
				return m.handleSubmit(true)
			}
			// Regular Enter submits if not in a dropdown and not in description
			// (Description field uses Enter for newlines, not submit)
			// Also check if labelsCombo has a pending value (ab-mod2: value selected but not yet added as chip)
			if !m.isAnyDropdownOpen() && m.focus != FocusDescription && m.labelsCombo.combo.Value() == "" {
				return m.handleSubmit(false)
			}

		case tea.KeyTab:
			return m.handleTab()

		case tea.KeyShiftTab:
			return m.handleShiftTab()
		}

		// Route to focused zone
		return m.handleZoneInput(msg)
	}

	// Pass through other messages to focused zone
	return m.passToFocusedZone(msg)
}

// DismissErrorToastMsg tells the App to dismiss the global error toast.
type DismissErrorToastMsg struct{}

// BeadUpdatedMsg is sent when edit form is submitted.
type BeadUpdatedMsg struct {
	ID               string
	Title            string
	Description      string
	IssueType        string
	Priority         int
	ParentID         string
	OriginalParentID string
	Labels           []string
	Assignee         string
}

// submitEdit packages the current form values for update.
func (m *CreateOverlay) submitEdit() tea.Cmd {
	return func() tea.Msg {
		return BeadUpdatedMsg{
			ID:          m.editingBead.ID,
			Title:       strings.TrimSpace(m.titleInput.Value()),
			Description: strings.TrimSpace(m.descriptionInput.Value()),
			IssueType:   typeOptions[m.typeIndex],
			Priority:    m.priorityIndex,
			ParentID:    m.ParentID(),
			OriginalParentID: func() string {
				if m.isEditMode() {
					return m.editingBeadParentID
				}
				return ""
			}(),
			Labels:   m.labelsCombo.GetChips(),
			Assignee: m.getAssigneeValue(),
		}
	}
}

// getAssigneeValue returns a normalized assignee string for submission.
func (m *CreateOverlay) getAssigneeValue() string {
	assignee := m.assigneeCombo.Value()
	if assignee == "Unassigned" {
		return ""
	}
	if strings.HasPrefix(assignee, "Me (") && strings.HasSuffix(assignee, ")") {
		return strings.TrimSuffix(strings.TrimPrefix(assignee, "Me ("), ")")
	}
	return assignee
}

// typeIndexFromString converts an issue type to the corresponding index in typeOptions.
// Returns 0 (task) as a safe default for unknown values.
func typeIndexFromString(issueType string) int {
	for i, t := range typeOptions {
		if t == issueType {
			return i
		}
	}
	return 0
}

// Title returns the current title value.
func (m *CreateOverlay) Title() string {
	return m.titleInput.Value()
}

// Description returns the current description value.
func (m *CreateOverlay) Description() string {
	return m.descriptionInput.Value()
}

// IssueType returns the current issue type value.
func (m *CreateOverlay) IssueType() string {
	return typeOptions[m.typeIndex]
}

// Priority returns the current priority value.
func (m *CreateOverlay) Priority() int {
	return m.priorityIndex
}

// ParentID returns the current parent ID value.
func (m *CreateOverlay) ParentID() string {
	selectedDisplay := m.parentCombo.Value()
	if selectedDisplay == "" {
		return ""
	}
	for _, p := range m.parentOptions {
		if p.Display == selectedDisplay {
			return p.ID
		}
	}
	return ""
}

// Focus returns the current focus zone (for testing).
func (m *CreateOverlay) Focus() CreateFocus {
	return m.focus
}

// DefaultParentID returns the default parent ID (for testing).
func (m *CreateOverlay) DefaultParentID() string {
	return m.defaultParentID
}

// IsRootMode returns whether the overlay was opened in root mode (for testing).
func (m *CreateOverlay) IsRootMode() bool {
	return m.isRootMode
}

// TitleValidationError returns whether the title is showing validation error (for testing).
func (m *CreateOverlay) TitleValidationError() bool {
	return m.titleValidationError
}
