package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CreateFocus represents which zone has focus in the create overlay.
type CreateFocus int

// Focus zones in tab order (spec Section 6)
const (
	FocusParent CreateFocus = iota
	FocusTitle
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

	// Zone 2: Title (hero element)
	titleInput textinput.Model

	// Zone 3: Properties (2-column grid)
	typeIndex       int
	priorityIndex   int
	typeManuallySet bool // Disables auto-inference when true

	// Zone 4: Labels (multi-select chips)
	labelsCombo   ChipComboBox
	labelsOptions []string

	// Zone 5: Assignee (single-select combo)
	assigneeCombo   ComboBox
	assigneeOptions []string
}

// ParentOption represents a bead that can be selected as a parent.
type ParentOption struct {
	ID      string
	Display string // e.g., "ab-83s Create and Edit..."
}

// BeadCreatedMsg is sent when form submission is confirmed.
type BeadCreatedMsg struct {
	Title     string
	IssueType string
	Priority  int
	ParentID  string
	Labels    []string // Selected labels (backend integration in ab-l1k)
	Assignee  string   // Selected assignee (backend integration in ab-39r)
}

// CreateCancelledMsg is sent when the overlay is dismissed without action.
type CreateCancelledMsg struct{}

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
	// Zone 2: Title input (hero element)
	ti := textinput.New()
	ti.Placeholder = ""
	ti.CharLimit = 100
	ti.Width = 44

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
	if !opts.IsRootMode && opts.DefaultParentID != "" {
		for _, p := range opts.AvailableParents {
			if p.ID == opts.DefaultParentID {
				parentCombo.SetValue(p.Display)
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
	// Prepend "Unassigned" and "Me" options
	assigneeOpts := []string{"Unassigned"}
	assigneeOpts = append(assigneeOpts, opts.AvailableAssignees...)
	assigneeCombo := NewComboBox(assigneeOpts).
		WithWidth(44).
		WithMaxVisible(5).
		WithPlaceholder("type to filter...").
		WithAllowNew(true, "New assignee: %s")
	assigneeCombo.SetValue("Unassigned")

	m := &CreateOverlay{
		focus:           FocusTitle, // Title is auto-focused (spec Section 3.2)
		titleInput:      ti,
		typeIndex:       0, // Task
		priorityIndex:   2, // Medium
		parentCombo:     parentCombo,
		parentOptions:   opts.AvailableParents,
		isRootMode:      opts.IsRootMode,
		defaultParentID: opts.DefaultParentID,
		labelsCombo:     labelsCombo,
		labelsOptions:   opts.AvailableLabels,
		assigneeCombo:   assigneeCombo,
		assigneeOptions: opts.AvailableAssignees,
	}

	// Focus title input
	ti.Focus()

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
	case ChipComboBoxTabMsg:
		// Labels combo requested Tab - move to Assignee
		m.focus = FocusAssignee
		m.labelsCombo.Blur()
		cmds = append(cmds, m.assigneeCombo.Focus())
		return m, tea.Batch(cmds...)

	case ComboBoxValueSelectedMsg:
		// Parent or Assignee combo selected a value
		// No special action needed, value is already set in the combo
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
			// Regular Enter submits if not in a dropdown
			if !m.isAnyDropdownOpen() {
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

func (m *CreateOverlay) handleEscape() (*CreateOverlay, tea.Cmd) {
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
		m.assigneeCombo.Blur()
		m.assigneeCombo.Focus()
		return m, nil
	}

	// No dropdown open - cancel the modal
	return m, func() tea.Msg { return CreateCancelledMsg{} }
}

func (m *CreateOverlay) handleSubmit(_ bool) (*CreateOverlay, tea.Cmd) {
	// Validate title
	if strings.TrimSpace(m.titleInput.Value()) == "" {
		return m, nil
	}

	// TODO: implement "Create & Add Another" (Ctrl+Enter) using the stayOpen parameter
	// For now, both submit and close
	return m, m.submit
}

func (m *CreateOverlay) handleTab() (*CreateOverlay, tea.Cmd) {
	var cmds []tea.Cmd

	// Close any open dropdowns
	m.parentCombo.Blur()
	m.assigneeCombo.Blur()

	// Tab order: Title -> Type -> Priority -> Labels -> Assignee -> (wrap to Title)
	switch m.focus {
	case FocusParent:
		m.focus = FocusTitle
		cmds = append(cmds, m.titleInput.Focus())
	case FocusTitle:
		m.titleInput.Blur()
		m.focus = FocusType
	case FocusType:
		m.focus = FocusPriority
	case FocusPriority:
		m.focus = FocusLabels
		cmds = append(cmds, m.labelsCombo.Focus())
	case FocusLabels:
		// Labels combo handles its own Tab via ChipComboBoxTabMsg
		m.labelsCombo, _ = m.labelsCombo.Update(tea.KeyMsg{Type: tea.KeyTab})
		return m, tea.Batch(cmds...)
	case FocusAssignee:
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
		cmds = append(cmds, m.parentCombo.Focus())
	case FocusType:
		m.focus = FocusTitle
		cmds = append(cmds, m.titleInput.Focus())
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
		// Parent uses ComboBox
		m.parentCombo, cmd = m.parentCombo.Update(msg)
		return m, cmd

	case FocusTitle:
		m.titleInput, cmd = m.titleInput.Update(msg)
		return m, cmd

	case FocusType:
		// Type uses arrow keys and single-key selection
		switch msg.Type {
		case tea.KeyUp, tea.KeyDown:
			if msg.Type == tea.KeyUp && m.typeIndex > 0 {
				m.typeIndex--
			} else if msg.Type == tea.KeyDown && m.typeIndex < len(typeOptions)-1 {
				m.typeIndex++
			}
		case tea.KeyLeft:
			m.focus = FocusType // Stay in type (leftmost column)
		case tea.KeyRight:
			m.focus = FocusPriority // Move to priority column
		case tea.KeyRunes:
			// Single-key selection: t=task, f=feature, b=bug, e=epic, c=chore
			if len(msg.Runes) > 0 {
				m.handleTypeHotkey(msg.Runes[0])
			}
		}
		return m, nil

	case FocusPriority:
		// Priority uses arrow keys and single-key selection
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
			// Single-key selection: c=crit, h=high, m=med, l=low, b=backlog
			if len(msg.Runes) > 0 {
				m.handlePriorityHotkey(msg.Runes[0])
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

func (m *CreateOverlay) submit() tea.Msg {
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

	// Get assignee (empty string if "Unassigned")
	assignee := m.assigneeCombo.Value()
	if assignee == "Unassigned" {
		assignee = ""
	}

	return BeadCreatedMsg{
		Title:     strings.TrimSpace(m.titleInput.Value()),
		IssueType: typeOptions[m.typeIndex],
		Priority:  m.priorityIndex,
		ParentID:  parentID,
		Labels:    m.labelsCombo.GetChips(),
		Assignee:  assignee,
	}
}

// Styles for the create overlay
var (
	styleCreateLabel = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				MarginRight(1)

	styleCreatePill = lipgloss.NewStyle().
				Padding(0, 1).
				Foreground(lipgloss.Color("241"))

	styleCreatePillSelected = lipgloss.NewStyle().
				Padding(0, 1).
				Bold(true).
				Foreground(lipgloss.Color("212"))

	styleCreatePillFocused = lipgloss.NewStyle().
				Padding(0, 1).
				Bold(true).
				Foreground(lipgloss.Color("86")).
				Background(lipgloss.Color("236"))

	styleCreateInput = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("236")).
				Padding(0, 1).
				Width(44)

	styleCreateInputFocused = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("86")).
				Padding(0, 1).
				Width(44)

	styleCreateError = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Italic(true)
)

// View implements tea.Model - 5-zone HUD layout per spec Section 3.
func (m *CreateOverlay) View() string {
	var b strings.Builder

	// Header
	title := styleHelpTitle.Render("NEW BEAD")
	divider := styleHelpDivider.Render(strings.Repeat("─", 52))

	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(divider)
	b.WriteString("\n\n")

	// Zone 1: Parent (anchor at top)
	parentLabel := styleCreateLabel.Render("PARENT")
	if m.focus == FocusParent {
		parentLabel = styleHelpSectionHeader.Render("PARENT")
	}
	b.WriteString(parentLabel)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	b.WriteString(hintStyle.Render("                                    Shift+Tab"))
	b.WriteString("\n")

	// Parent combo box or root indicator
	if m.isRootMode && m.parentCombo.Value() == "" {
		rootStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("201")) // Magenta
		b.WriteString(styleCreateInput.Render(rootStyle.Render("◇ No Parent (Root Item)")))
	} else {
		b.WriteString(m.parentCombo.View())
	}
	b.WriteString("\n\n")

	// Zone 2: Title (hero element)
	titleLabel := styleCreateLabel.Render("TITLE")
	if m.focus == FocusTitle {
		titleLabel = styleHelpSectionHeader.Render("TITLE")
	}
	b.WriteString(titleLabel)
	b.WriteString("\n")

	titleStyle := styleCreateInput
	if m.focus == FocusTitle {
		titleStyle = styleCreateInputFocused
	}
	b.WriteString(titleStyle.Render(m.titleInput.View()))

	// Validation hint
	if m.focus == FocusTitle && strings.TrimSpace(m.titleInput.Value()) == "" {
		b.WriteString("\n")
		b.WriteString(styleCreateError.Render("  required"))
	}
	b.WriteString("\n\n")

	// Zone 3: Properties (2-column grid)
	propsLabel := styleCreateLabel.Render("PROPERTIES")
	if m.focus == FocusType || m.focus == FocusPriority {
		propsLabel = styleHelpSectionHeader.Render("PROPERTIES")
	}
	b.WriteString(propsLabel)
	b.WriteString("\n")

	// Type and Priority columns side-by-side
	propsGrid := lipgloss.JoinHorizontal(lipgloss.Top,
		m.renderTypeColumn(),
		"    ", // Spacer between columns
		m.renderPriorityColumn(),
	)
	b.WriteString(propsGrid)
	b.WriteString("\n\n")

	// Zone 4: Labels (inline chips)
	labelsLabel := styleCreateLabel.Render("LABELS")
	if m.focus == FocusLabels {
		labelsLabel = styleHelpSectionHeader.Render("LABELS")
	}
	b.WriteString(labelsLabel)
	b.WriteString("\n")
	b.WriteString(m.labelsCombo.View())
	b.WriteString("\n\n")

	// Zone 5: Assignee
	assigneeLabel := styleCreateLabel.Render("ASSIGNEE")
	if m.focus == FocusAssignee {
		assigneeLabel = styleHelpSectionHeader.Render("ASSIGNEE")
	}
	b.WriteString(assigneeLabel)
	b.WriteString("\n")
	b.WriteString(m.assigneeCombo.View())
	b.WriteString("\n\n")

	// Footer with keyboard hints
	b.WriteString(divider)
	b.WriteString("\n")
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	b.WriteString(footerStyle.Render("Enter Create   ^Enter Create & Add Another   Tab Next   Esc Cancel"))

	return styleHelpOverlay.Render(b.String())
}

func (m *CreateOverlay) renderTypeColumn() string {
	var b strings.Builder

	// Column header
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	if m.focus == FocusType {
		headerStyle = lipgloss.NewStyle().Foreground(cCyan).Bold(true)
	}
	b.WriteString(headerStyle.Render("TYPE"))
	b.WriteString("\n")

	// Options
	for i, label := range typeLabels {
		prefix := "  "
		style := styleCreatePill
		if i == m.typeIndex {
			prefix = "► "
			if m.focus == FocusType {
				style = styleCreatePillFocused
			} else {
				style = styleCreatePillSelected
			}
		}
		b.WriteString(style.Render(prefix + label))
		if i < len(typeLabels)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m *CreateOverlay) renderPriorityColumn() string {
	var b strings.Builder

	// Column header
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	if m.focus == FocusPriority {
		headerStyle = lipgloss.NewStyle().Foreground(cCyan).Bold(true)
	}
	b.WriteString(headerStyle.Render("PRIORITY"))
	b.WriteString("\n")

	// Options
	for i, label := range priorityLabels {
		prefix := "  "
		style := styleCreatePill
		if i == m.priorityIndex {
			prefix = "► "
			if m.focus == FocusPriority {
				style = styleCreatePillFocused
			} else {
				style = styleCreatePillSelected
			}
		}
		b.WriteString(style.Render(prefix + label))
		if i < len(priorityLabels)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// Title returns the current title value.
func (m *CreateOverlay) Title() string {
	return m.titleInput.Value()
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
