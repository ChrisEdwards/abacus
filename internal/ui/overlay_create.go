package ui

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"abacus/internal/ui/theme"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Flash duration for validation errors (spec Section 4.4: 300ms)
const (
	titleFlashDuration = 300 * time.Millisecond
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
	// Focus management
	focus CreateFocus

	// Zone 1: Parent (anchor at top)
	parentCombo     ComboBox
	parentOptions   []ParentOption
	isRootMode      bool   // True if opened with 'N' (no parent)
	defaultParentID string // Pre-selected parent ID
	parentOriginal  string // Value when Parent field focused, for Esc revert (spec Section 12)

	// Zone 2: Title (hero element)
	titleInput           textinput.Model
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
	// Zone 2: Title input (hero element) - uses textinput with custom wrapping in View
	ti := textinput.New()
	ti.Placeholder = ""
	ti.CharLimit = 100
	ti.Width = 42 // Slightly narrower to account for cursor

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
		WithPlaceholder("type to search...")

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

	// Focus title input BEFORE assigning to struct (textinput.Model is a value type)
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

// Init implements tea.Model.
func (m *CreateOverlay) Init() tea.Cmd {
	return textinput.Blink
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

	case ComboBoxValueSelectedMsg:
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

	case tea.KeyMsg:
		// Handle global keys first
		switch msg.Type {
		case tea.KeyEsc:
			return m.handleEscape()

		case tea.KeyEnter:
			// Ctrl+Enter always submits
			if msg.String() == "ctrl+enter" {
				return m.handleSubmit(true)
			}
			// Regular Enter submits if not in a dropdown and not in description
			// (Description field uses Enter for newlines, not submit)
			if !m.isAnyDropdownOpen() && m.focus != FocusDescription {
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

func (m *CreateOverlay) handleEscape() (*CreateOverlay, tea.Cmd) {
	// If there's a backend error, ESC dismisses the toast first
	if m.hasBackendError {
		m.hasBackendError = false
		return m, func() tea.Msg { return DismissErrorToastMsg{} }
	}

	// Check if any dropdown is open
	if m.parentCombo.IsDropdownOpen() {
		m.parentCombo.Blur()
		m.parentCombo.Focus()
		return m, nil
	}
	if m.labelsCombo.IsDropdownOpen() {
		// Labels handles its own multi-stage escape
		m.labelsCombo, _ = m.labelsCombo.Update(tea.KeyMsg{Type: tea.KeyEsc})
		return m, nil
	}
	if m.assigneeCombo.IsDropdownOpen() {
		// First stage: Close dropdown, keep typed text
		m.assigneeCombo, _ = m.assigneeCombo.Update(tea.KeyMsg{Type: tea.KeyEsc})
		return m, nil
	}
	if m.focus == FocusAssignee && m.assigneeCombo.InputValue() != m.assigneeCombo.Value() {
		// Second stage: Revert to original value
		m.assigneeCombo, _ = m.assigneeCombo.Update(tea.KeyMsg{Type: tea.KeyEsc})
		return m, nil
	}

	// If Parent field is focused (no dropdown open), Esc reverts and moves to Title (spec Section 4.2)
	if m.focus == FocusParent {
		m.parentCombo.SetValue(m.parentOriginal)
		// Restore isRootMode based on whether original was empty
		m.isRootMode = m.parentOriginal == ""
		m.parentCombo.Blur()
		m.focus = FocusTitle
		return m, m.titleInput.Focus()
	}

	// No dropdown open - cancel the modal
	return m, func() tea.Msg { return CreateCancelledMsg{} }
}

func (m *CreateOverlay) handleSubmit(stayOpen bool) (*CreateOverlay, tea.Cmd) {
	// Validate title - flash red if empty (spec Section 4.4)
	if strings.TrimSpace(m.titleInput.Value()) == "" {
		m.titleValidationError = true
		return m, titleFlashCmd()
	}

	// Clear any previous backend error (user is retrying)
	m.hasBackendError = false

	// Set creating state for footer (spec Section 4.1)
	m.isCreating = true

	if stayOpen {
		// Bulk entry mode: submit and prepare for next entry (spec Section 4.3)
		return m, tea.Batch(
			m.submitWithMode(true),
			m.prepareForNextEntry(),
		)
	}

	// Normal mode: submit and close
	return m, m.submitWithMode(false)
}

func (m *CreateOverlay) handleTab() (*CreateOverlay, tea.Cmd) {
	var cmds []tea.Cmd

	// Close parent dropdown (assignee is handled in its case to allow Tab commit)
	m.parentCombo.Blur()

	// Tab order: Title -> Description -> Type -> Priority -> Labels -> Assignee -> (wrap to Title)
	switch m.focus {
	case FocusParent:
		m.focus = FocusTitle
		cmds = append(cmds, m.titleInput.Focus())
	case FocusTitle:
		m.titleInput.Blur()
		m.focus = FocusDescription
		cmds = append(cmds, m.descriptionInput.Focus())
	case FocusDescription:
		m.descriptionInput.Blur()
		m.focus = FocusType
	case FocusType:
		m.focus = FocusPriority
	case FocusPriority:
		m.focus = FocusLabels
		cmds = append(cmds, m.labelsCombo.Focus())
	case FocusLabels:
		// Labels combo handles its own Tab via ChipComboBoxTabMsg
		var cmd tea.Cmd
		m.labelsCombo, cmd = m.labelsCombo.Update(tea.KeyMsg{Type: tea.KeyTab})
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	case FocusAssignee:
		// Let ComboBox process Tab to commit any pending value (spec Section 6)
		var cmd tea.Cmd
		m.assigneeCombo, cmd = m.assigneeCombo.Update(tea.KeyMsg{Type: tea.KeyTab})

		// Process value selection inline to capture new assignee toast
		// (focus will have moved by the time the msg would normally be processed)
		if cmd != nil {
			msg := cmd()
			if vsm, ok := msg.(ComboBoxValueSelectedMsg); ok && vsm.IsNew {
				cmds = append(cmds, func() tea.Msg {
					return NewAssigneeAddedMsg{Assignee: vsm.Value}
				})
			}
		}

		m.assigneeCombo.Blur()
		m.focus = FocusTitle
		cmds = append(cmds, m.titleInput.Focus())
	}

	return m, tea.Batch(cmds...)
}

func (m *CreateOverlay) handleShiftTab() (*CreateOverlay, tea.Cmd) {
	var cmds []tea.Cmd

	// Close any open dropdowns
	m.parentCombo.Blur()
	m.assigneeCombo.Blur()
	m.labelsCombo.Blur()

	// Reverse tab order, with Title -> Parent special case
	switch m.focus {
	case FocusTitle:
		m.titleInput.Blur()
		m.focus = FocusParent
		// Store original value for Esc revert (spec Section 4.2)
		m.parentOriginal = m.parentCombo.Value()
		cmds = append(cmds, m.parentCombo.Focus())
	case FocusDescription:
		m.descriptionInput.Blur()
		m.focus = FocusTitle
		cmds = append(cmds, m.titleInput.Focus())
	case FocusType:
		m.focus = FocusDescription
		cmds = append(cmds, m.descriptionInput.Focus())
	case FocusPriority:
		m.focus = FocusType
	case FocusLabels:
		m.focus = FocusPriority
	case FocusAssignee:
		m.assigneeCombo.Blur()
		m.focus = FocusLabels
		cmds = append(cmds, m.labelsCombo.Focus())
	case FocusParent:
		m.parentCombo.Blur()
		m.focus = FocusAssignee
		cmds = append(cmds, m.assigneeCombo.Focus())
	}

	return m, tea.Batch(cmds...)
}

func (m *CreateOverlay) handleZoneInput(msg tea.KeyMsg) (*CreateOverlay, tea.Cmd) {
	var cmd tea.Cmd

	switch m.focus {
	case FocusParent:
		// Handle Delete/Backspace to clear to root (spec Section 4.2)
		if msg.Type == tea.KeyDelete || msg.Type == tea.KeyBackspace {
			// Only clear when dropdown is not open (otherwise let ComboBox handle it)
			if !m.parentCombo.IsDropdownOpen() {
				m.parentCombo.SetValue("")
				m.isRootMode = true
				return m, nil
			}
		}
		// Parent uses ComboBox
		m.parentCombo, cmd = m.parentCombo.Update(msg)
		return m, cmd

	case FocusTitle:
		// Capture old title for comparison (spec Section 5: type auto-inference)
		oldTitle := m.titleInput.Value()

		// Update title input
		m.titleInput, cmd = m.titleInput.Update(msg)

		// Auto-infer type if title changed and not manually set (spec Section 5)
		newTitle := m.titleInput.Value()
		if newTitle != oldTitle && !m.typeManuallySet {
			if inferredIdx := inferTypeFromTitle(newTitle); inferredIdx != -1 {
				// Only update if inference actually changed the type
				if inferredIdx != m.typeIndex {
					m.typeIndex = inferredIdx
					m.typeInferenceActive = true
					// Return command to trigger visual feedback
					return m, tea.Batch(cmd, typeInferenceFlashCmd())
				}
			}
		}

		return m, cmd

	case FocusDescription:
		// Description uses textarea - forward all keys to it
		// Tab/Shift+Tab/Esc are handled at global level, Enter falls through here
		m.descriptionInput, cmd = m.descriptionInput.Update(msg)
		return m, cmd

	case FocusType:
		// Type uses arrow keys, vim keys (j/k/h/l), and single-key selection
		switch msg.Type {
		case tea.KeyUp, tea.KeyDown:
			if msg.Type == tea.KeyUp && m.typeIndex > 0 {
				m.typeIndex--
				m.typeManuallySet = true // Disable auto-inference (spec Section 5)
			} else if msg.Type == tea.KeyDown && m.typeIndex < len(typeOptions)-1 {
				m.typeIndex++
				m.typeManuallySet = true // Disable auto-inference (spec Section 5)
			}
		case tea.KeyLeft:
			m.focus = FocusType // Stay in type (leftmost column)
		case tea.KeyRight:
			m.focus = FocusPriority // Move to priority column
		case tea.KeyRunes:
			if len(msg.Runes) > 0 {
				r := msg.Runes[0]
				switch r {
				case 'j': // Vim down
					if m.typeIndex < len(typeOptions)-1 {
						m.typeIndex++
						m.typeManuallySet = true // Disable auto-inference (spec Section 5)
					}
				case 'k': // Vim up
					if m.typeIndex > 0 {
						m.typeIndex--
						m.typeManuallySet = true // Disable auto-inference (spec Section 5)
					}
				case 'h': // Vim left - stay in type (leftmost)
					// Already in leftmost column
				case 'l': // Vim right - move to priority
					m.focus = FocusPriority
				default:
					// Single-key selection: t=task, f=feature, b=bug, e=epic, c=chore
					m.handleTypeHotkey(r)
				}
			}
		}
		return m, nil

	case FocusPriority:
		// Priority uses arrow keys, vim keys (j/k), and single-key selection
		// Note: h/l are priority hotkeys (High, Low) not vim navigation here
		switch msg.Type {
		case tea.KeyUp, tea.KeyDown:
			if msg.Type == tea.KeyUp && m.priorityIndex > 0 {
				m.priorityIndex--
			} else if msg.Type == tea.KeyDown && m.priorityIndex < len(priorityLabels)-1 {
				m.priorityIndex++
			}
		case tea.KeyLeft:
			m.focus = FocusType // Move to type column
		case tea.KeyRight:
			m.focus = FocusPriority // Stay in priority (rightmost column)
		case tea.KeyRunes:
			if len(msg.Runes) > 0 {
				r := msg.Runes[0]
				switch r {
				case 'j': // Vim down
					if m.priorityIndex < len(priorityLabels)-1 {
						m.priorityIndex++
					}
				case 'k': // Vim up
					if m.priorityIndex > 0 {
						m.priorityIndex--
					}
				default:
					// Single-key selection: c=crit, h=high, m=med, l=low, b=backlog
					// (h/l are hotkeys here, not vim navigation - use arrow keys for columns)
					m.handlePriorityHotkey(r)
				}
			}
		}
		return m, nil

	case FocusLabels:
		// Labels uses ChipComboBox
		m.labelsCombo, cmd = m.labelsCombo.Update(msg)
		return m, cmd

	case FocusAssignee:
		// Assignee uses ComboBox
		m.assigneeCombo, cmd = m.assigneeCombo.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *CreateOverlay) passToFocusedZone(msg tea.Msg) (*CreateOverlay, tea.Cmd) {
	var cmd tea.Cmd

	switch m.focus {
	case FocusParent:
		m.parentCombo, cmd = m.parentCombo.Update(msg)
	case FocusTitle:
		m.titleInput, cmd = m.titleInput.Update(msg)
	case FocusDescription:
		m.descriptionInput, cmd = m.descriptionInput.Update(msg)
	case FocusLabels:
		m.labelsCombo, cmd = m.labelsCombo.Update(msg)
	case FocusAssignee:
		m.assigneeCombo, cmd = m.assigneeCombo.Update(msg)
	}

	return m, cmd
}

func (m *CreateOverlay) isAnyDropdownOpen() bool {
	return m.parentCombo.IsDropdownOpen() ||
		m.labelsCombo.IsDropdownOpen() ||
		m.assigneeCombo.IsDropdownOpen()
}

func (m *CreateOverlay) handleTypeHotkey(r rune) {
	switch r {
	case 't', 'T':
		m.typeIndex = 0 // Task
		m.typeManuallySet = true
	case 'f', 'F':
		m.typeIndex = 1 // Feature
		m.typeManuallySet = true
	case 'b', 'B':
		m.typeIndex = 2 // Bug
		m.typeManuallySet = true
	case 'e', 'E':
		m.typeIndex = 3 // Epic
		m.typeManuallySet = true
	case 'c', 'C':
		m.typeIndex = 4 // Chore
		m.typeManuallySet = true
	}
}

func (m *CreateOverlay) handlePriorityHotkey(r rune) {
	switch r {
	case 'c', 'C':
		m.priorityIndex = 0 // Critical
	case 'h', 'H':
		m.priorityIndex = 1 // High
	case 'm', 'M':
		m.priorityIndex = 2 // Medium
	case 'l', 'L':
		m.priorityIndex = 3 // Low
	case 'b', 'B':
		m.priorityIndex = 4 // Backlog
	}
}

// submitWithMode creates BeadCreatedMsg with the specified StayOpen mode (spec Section 4.3).
func (m *CreateOverlay) submitWithMode(stayOpen bool) tea.Cmd {
	return func() tea.Msg {
		// Get parent ID from selected parent display
		parentID := ""
		selectedParentDisplay := m.parentCombo.Value()
		if selectedParentDisplay != "" {
			for _, p := range m.parentOptions {
				if p.Display == selectedParentDisplay {
					parentID = p.ID
					break
				}
			}
		}

		// Get assignee (empty string if "Unassigned", extract username from "Me (username)")
		assignee := m.assigneeCombo.Value()
		if assignee == "Unassigned" {
			assignee = ""
		} else if strings.HasPrefix(assignee, "Me (") && strings.HasSuffix(assignee, ")") {
			// Extract username from "Me (username)" format
			assignee = strings.TrimSuffix(strings.TrimPrefix(assignee, "Me ("), ")")
		}

		return BeadCreatedMsg{
			Title:       strings.TrimSpace(m.titleInput.Value()),
			Description: strings.TrimSpace(m.descriptionInput.Value()),
			IssueType:   typeOptions[m.typeIndex],
			Priority:    m.priorityIndex,
			ParentID:    parentID,
			Labels:      m.labelsCombo.GetChips(),
			Assignee:    assignee,
			StayOpen:    stayOpen,
		}
	}
}

// prepareForNextEntry resets Title only, keeps everything else persistent (spec Section 4.3).
func (m *CreateOverlay) prepareForNextEntry() tea.Cmd {
	return func() tea.Msg {
		return bulkEntryResetMsg{}
	}
}

// Styles for the create overlay

func styleCreateLabel() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().TextMuted()).
		MarginRight(1)
}

func styleCreatePill() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(theme.Current().TextMuted())
}

func styleCreatePillSelected() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 1).
		Bold(true).
		Foreground(theme.Current().Primary())
}

func styleCreatePillFocused() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 1).
		Bold(true).
		Foreground(theme.Current().Success()).
		Background(theme.Current().BorderDim())
}

func styleCreateInput() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().BorderDim()).
		Padding(0, 1).
		Width(44)
}

func styleCreateInputFocused() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().Success()).
		Padding(0, 1).
		Width(44)
}

// Error state for title validation (spec Section 4.4 - red border flash)
func styleCreateInputError() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().Error()).
		Padding(0, 1).
		Width(44)
}

// Dimmed style for modal depth effect (spec Section 2.4)
func styleCreateDimmed() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().BorderNormal())
}

// wrapText wraps text at the given width, breaking on word boundaries.
func wrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}

	var result strings.Builder
	words := strings.Fields(text)
	lineLen := 0

	for i, word := range words {
		wordLen := len(word)
		if lineLen+wordLen > width && lineLen > 0 {
			result.WriteString("\n")
			lineLen = 0
		}
		if lineLen > 0 {
			result.WriteString(" ")
			lineLen++
		}
		result.WriteString(word)
		lineLen += wordLen

		// Handle very long words that exceed width
		if lineLen > width && i < len(words)-1 {
			result.WriteString("\n")
			lineLen = 0
		}
	}
	return result.String()
}

// View implements tea.Model - 5-zone HUD layout per spec Section 3.
func (m *CreateOverlay) View() string {
	var b strings.Builder

	// Check if parent search is active for dimming effect (spec Section 2.4)
	parentSearchActive := m.parentCombo.IsDropdownOpen()

	// Header
	title := styleHelpTitle().Render("NEW BEAD")
	divider := styleHelpDivider().Render(strings.Repeat("─", 52))

	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(divider)
	b.WriteString("\n\n")

	// Zone 1: Parent (anchor at top) - never dimmed
	parentLabel := styleCreateLabel().Render("PARENT")
	if m.focus == FocusParent {
		parentLabel = styleHelpSectionHeader().Render("PARENT")
	}
	b.WriteString(parentLabel)
	hintStyle := lipgloss.NewStyle().Foreground(theme.Current().TextMuted())
	b.WriteString(hintStyle.Render("                                    Shift+Tab"))
	b.WriteString("\n")

	// Parent combo box or root indicator
	if m.isRootMode && m.parentCombo.Value() == "" {
		rootStyle := lipgloss.NewStyle().Foreground(theme.Current().Primary())
		b.WriteString(styleCreateInput().Render(rootStyle.Render("◇ No Parent (Root Item)")))
	} else {
		b.WriteString(m.parentCombo.View())
	}
	b.WriteString("\n\n")

	// Zone 2: Title (hero element) - dimmed when parent search active
	titleLabel := styleCreateLabel().Render("TITLE")
	if m.focus == FocusTitle {
		titleLabel = styleHelpSectionHeader().Render("TITLE")
	}
	if parentSearchActive {
		titleLabel = styleCreateDimmed().Render("TITLE")
	}
	b.WriteString(titleLabel)
	b.WriteString("\n")

	titleStyle := styleCreateInput()
	if m.focus == FocusTitle {
		titleStyle = styleCreateInputFocused()
	}
	// Show red border for both validation and backend errors
	if m.titleValidationError {
		titleStyle = styleCreateInputError()
	}

	// Render title with word wrapping (max width ~40 to fit in border)
	titleText := m.titleInput.Value()
	wrappedTitle := wrapText(titleText, 40)
	if m.focus == FocusTitle {
		// Add cursor when focused
		if titleText == "" {
			wrappedTitle = "│" // Cursor only
		} else {
			wrappedTitle += "│"
		}
	}
	titleView := titleStyle.Render(wrappedTitle)
	if parentSearchActive {
		titleView = styleCreateDimmed().Render(wrappedTitle)
	}
	b.WriteString(titleView)

	b.WriteString("\n\n")

	// Zone 2b: Description (multi-line textarea) - dimmed when parent search active
	descLabel := styleCreateLabel().Render("DESCRIPTION")
	if m.focus == FocusDescription {
		descLabel = styleHelpSectionHeader().Render("DESCRIPTION")
	}
	if parentSearchActive {
		descLabel = styleCreateDimmed().Render("DESCRIPTION")
	}
	b.WriteString(descLabel)
	b.WriteString("\n")

	descStyle := styleCreateInput()
	if m.focus == FocusDescription {
		descStyle = styleCreateInputFocused()
	}
	descView := descStyle.Render(m.descriptionInput.View())
	if parentSearchActive {
		descView = styleCreateDimmed().Render(m.descriptionInput.View())
	}
	b.WriteString(descView)
	b.WriteString("\n\n")

	// Zone 3: Type and Priority (2-column grid) - dimmed when parent search active
	// Type and Priority columns side-by-side
	propsGrid := lipgloss.JoinHorizontal(lipgloss.Top,
		m.renderTypeColumn(),
		"    ", // Spacer between columns
		m.renderPriorityColumn(),
	)
	if parentSearchActive {
		propsGrid = styleCreateDimmed().Render(propsGrid)
	}
	b.WriteString(propsGrid)
	b.WriteString("\n\n")

	// Zone 4: Labels (inline chips) - dimmed when parent search active
	labelsLabel := styleCreateLabel().Render("LABELS")
	if m.focus == FocusLabels {
		labelsLabel = styleHelpSectionHeader().Render("LABELS")
	}
	if parentSearchActive {
		labelsLabel = styleCreateDimmed().Render("LABELS")
	}
	b.WriteString(labelsLabel)
	b.WriteString("\n")
	labelsView := m.labelsCombo.View()
	if parentSearchActive {
		labelsView = styleCreateDimmed().Render(labelsView)
	}
	b.WriteString(labelsView)
	b.WriteString("\n\n")

	// Zone 5: Assignee - dimmed when parent search active
	assigneeLabel := styleCreateLabel().Render("ASSIGNEE")
	if m.focus == FocusAssignee {
		assigneeLabel = styleHelpSectionHeader().Render("ASSIGNEE")
	}
	if parentSearchActive {
		assigneeLabel = styleCreateDimmed().Render("ASSIGNEE")
	}
	b.WriteString(assigneeLabel)
	b.WriteString("\n")
	assigneeView := m.assigneeCombo.View()
	if parentSearchActive {
		assigneeView = styleCreateDimmed().Render(assigneeView)
	}
	b.WriteString(assigneeView)
	b.WriteString("\n\n")

	// Footer with keyboard hints (spec Section 4.1 - footer flipping)
	b.WriteString(divider)
	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	return styleHelpOverlay().Render(b.String())
}

// Layer returns a centered layer for the create overlay.
func (m *CreateOverlay) Layer(width, height, topMargin, bottomMargin int) Layer {
	return LayerFunc(func() *Canvas {
		content := m.View()
		if strings.TrimSpace(content) == "" {
			return nil
		}

		overlayWidth := lipgloss.Width(content)
		if overlayWidth <= 0 {
			return nil
		}
		overlayHeight := lipgloss.Height(content)
		if overlayHeight <= 0 {
			return nil
		}

		surface := NewSecondarySurface(overlayWidth, overlayHeight)
		surface.Draw(0, 0, content)

		x, y := centeredOffsets(width, height, overlayWidth, overlayHeight, topMargin, bottomMargin)
		surface.Canvas.SetOffset(x, y)
		return surface.Canvas
	})
}

func (m *CreateOverlay) renderTypeColumn() string {
	var b strings.Builder

	// Column header
	headerStyle := lipgloss.NewStyle().Foreground(theme.Current().TextMuted())
	if m.focus == FocusType {
		headerStyle = lipgloss.NewStyle().Foreground(theme.Current().Secondary()).Bold(true)
	}
	// Add flash animation when type was auto-inferred (spec Section 5)
	if m.typeInferenceActive {
		headerStyle = lipgloss.NewStyle().Foreground(theme.Current().Warning()).Bold(true)
	}
	b.WriteString(headerStyle.Render("TYPE"))
	b.WriteString("\n")

	// Options
	for i, label := range typeLabels {
		prefix := "  "
		style := styleCreatePill()
		underlineHotkey := false
		if i == m.typeIndex {
			prefix = "► "
			if m.focus == FocusType {
				style = styleCreatePillFocused()
				// Underline hotkey letter when focused (spec Section 3.3)
				underlineHotkey = true
			} else {
				style = styleCreatePillSelected()
			}
		} else if m.focus == FocusType {
			// Underline hotkey letter for all options when column is focused
			underlineHotkey = true
		}
		b.WriteString(renderHotkeyPill(style, prefix, label, underlineHotkey))
		if i < len(typeLabels)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m *CreateOverlay) renderPriorityColumn() string {
	var b strings.Builder

	// Column header
	headerStyle := lipgloss.NewStyle().Foreground(theme.Current().TextMuted())
	if m.focus == FocusPriority {
		headerStyle = lipgloss.NewStyle().Foreground(theme.Current().Secondary()).Bold(true)
	}
	b.WriteString(headerStyle.Render("PRIORITY"))
	b.WriteString("\n")

	// Options
	for i, label := range priorityLabels {
		prefix := "  "
		style := styleCreatePill()
		underlineHotkey := false
		if i == m.priorityIndex {
			prefix = "► "
			if m.focus == FocusPriority {
				style = styleCreatePillFocused()
				// Underline hotkey letter when focused (spec Section 3.3)
				underlineHotkey = true
			} else {
				style = styleCreatePillSelected()
			}
		} else if m.focus == FocusPriority {
			// Underline hotkey letter for all options when column is focused
			underlineHotkey = true
		}
		b.WriteString(renderHotkeyPill(style, prefix, label, underlineHotkey))
		if i < len(priorityLabels)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func renderHotkeyPill(style lipgloss.Style, prefix, label string, underline bool) string {
	if !underline || label == "" {
		return style.Render(prefix + label)
	}

	runes := []rune(label)
	var b strings.Builder
	b.WriteString(style.Render(prefix))

	underlineStyle := style.Underline(true)
	b.WriteString(underlineStyle.Render(string(runes[0])))
	if len(runes) > 1 {
		b.WriteString(style.Render(string(runes[1:])))
	}

	return b.String()
}

// renderFooter returns the dynamic footer based on current context (spec Section 4.1).
// Footer "flips" between states to eliminate ambiguity of intent.
// Uses keyPill() for consistency with the global footer styling.
func (m *CreateOverlay) renderFooter() string {
	var hints []footerHint

	switch {
	case m.isCreating:
		// Creating state: user must wait (no hints, just message)
		return styleFooterMuted().Render("Creating bead...")
	case m.parentCombo.IsDropdownOpen() || m.labelsCombo.IsDropdownOpen() || m.assigneeCombo.IsDropdownOpen():
		// Dropdown search active: Enter selects, Esc reverts
		hints = []footerHint{
			{"⏎", "Select"},
			{"esc", "Revert"},
		}
	case m.focus == FocusParent || m.focus == FocusLabels || m.focus == FocusAssignee:
		// Combo box field focused (but dropdown closed): show browse hint
		hints = []footerHint{
			{"↓", "Browse"},
			{"⏎", "Create"},
			{"Tab", "Next"},
			{"esc", "Cancel"},
		}
	default:
		// Default state: Title, Type, Priority fields
		hints = []footerHint{
			{"⏎", "Create"},
			{"^⏎", "Create+Add"},
			{"Tab", "Next"},
			{"esc", "Cancel"},
		}
	}

	// Render hints as pills (same as global footer)
	var parts []string
	for _, h := range hints {
		parts = append(parts, keyPill(h.key, h.desc))
	}
	return strings.Join(parts, "  ")
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
