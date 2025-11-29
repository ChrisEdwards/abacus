package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Field focus constants
const (
	focusTitle = iota
	focusType
	focusPriority
	focusParent
)

// Type options with descriptions
var typeOptions = []string{"task", "feature", "bug", "epic", "chore"}
var typeLabels = []string{"Task", "Feature", "Bug", "Epic", "Chore"}
var typeDescriptions = []string{
	"A small unit of work",
	"New functionality for users",
	"Something that's broken",
	"A large initiative with subtasks",
	"Maintenance or housekeeping",
}

// Priority options
var priorityLabels = []string{"Crit", "High", "Med", "Low", "Back"}

// CreateOverlay is a compact modal form for creating a new bead.
type CreateOverlay struct {
	// Focus management
	focus int

	// Title field
	titleInput textinput.Model

	// Type selection (horizontal pills)
	typeIndex int

	// Priority selection (horizontal pills)
	priorityIndex int

	// Parent selection with filter
	parentInput     textinput.Model
	parentOptions   []ParentOption
	filteredParents []ParentOption
	parentIndex     int    // Selected index in filtered list
	showDropdown    bool
	defaultParent   string
	selectedParent  *ParentOption // Currently selected parent (nil if none)
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
}

// CreateCancelledMsg is sent when the overlay is dismissed without action.
type CreateCancelledMsg struct{}

// NewCreateOverlay creates a new create overlay with smart defaults.
func NewCreateOverlay(defaultParentID string, availableParents []ParentOption) *CreateOverlay {
	// Title input
	ti := textinput.New()
	ti.Placeholder = "Enter bead title..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 40

	// Parent filter input - always starts empty for searching
	pi := textinput.New()
	pi.Placeholder = "type to search..."
	pi.CharLimit = 80
	pi.Width = 40

	m := &CreateOverlay{
		focus:           focusTitle,
		titleInput:      ti,
		typeIndex:       0, // Task
		priorityIndex:   2, // Medium
		parentInput:     pi,
		parentOptions:   availableParents,
		filteredParents: availableParents,
		defaultParent:   defaultParentID,
	}

	// Pre-select parent if default exists
	if defaultParentID != "" {
		for i := range availableParents {
			if availableParents[i].ID == defaultParentID {
				m.selectedParent = &availableParents[i]
				break
			}
		}
	}

	return m
}

// Init implements tea.Model.
func (m *CreateOverlay) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (m *CreateOverlay) Update(msg tea.Msg) (*CreateOverlay, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			// If dropdown is open, close it first
			if m.showDropdown {
				m.showDropdown = false
				m.parentInput.SetValue("")
				return m, nil
			}
			return m, func() tea.Msg { return CreateCancelledMsg{} }

		case tea.KeyEnter:
			// If in dropdown, select the item
			if m.showDropdown && len(m.filteredParents) > 0 {
				selected := m.filteredParents[m.parentIndex]
				m.selectedParent = &selected
				m.parentInput.SetValue("")
				m.showDropdown = false
				return m, nil
			}
			// Otherwise submit if title is valid
			if strings.TrimSpace(m.titleInput.Value()) != "" {
				return m, m.submit
			}
			return m, nil

		case tea.KeyTab, tea.KeyDown:
			// If in dropdown, navigate down
			if m.showDropdown {
				if m.parentIndex < len(m.filteredParents)-1 {
					m.parentIndex++
				}
				return m, nil
			}
			// Move to next field
			m.nextField()
			return m, nil

		case tea.KeyShiftTab, tea.KeyUp:
			// If in dropdown, navigate up
			if m.showDropdown {
				if m.parentIndex > 0 {
					m.parentIndex--
				}
				return m, nil
			}
			// Move to previous field
			m.prevField()
			return m, nil

		case tea.KeyLeft:
			if m.focus == focusType {
				if m.typeIndex > 0 {
					m.typeIndex--
				}
				return m, nil
			}
			if m.focus == focusPriority {
				if m.priorityIndex > 0 {
					m.priorityIndex--
				}
				return m, nil
			}

		case tea.KeyRight:
			if m.focus == focusType {
				if m.typeIndex < len(typeOptions)-1 {
					m.typeIndex++
				}
				return m, nil
			}
			if m.focus == focusPriority {
				if m.priorityIndex < len(priorityLabels)-1 {
					m.priorityIndex++
				}
				return m, nil
			}

		case tea.KeyBackspace, tea.KeyDelete:
			// In parent field with no input text but has selection, clear selection
			if m.focus == focusParent && m.parentInput.Value() == "" && m.selectedParent != nil {
				m.selectedParent = nil
				return m, nil
			}

		case tea.KeyCtrlN:
			// Ctrl+N to open/navigate dropdown
			if m.focus == focusParent {
				if !m.showDropdown {
					m.showDropdown = true
					m.parentIndex = 0
					m.filterParents()
				} else if m.parentIndex < len(m.filteredParents)-1 {
					m.parentIndex++
				}
			}
			return m, nil

		case tea.KeyCtrlP:
			// Ctrl+P to navigate up in dropdown
			if m.focus == focusParent && m.showDropdown {
				if m.parentIndex > 0 {
					m.parentIndex--
				}
			}
			return m, nil
		}
	}

	// Handle text input for focused field
	var cmd tea.Cmd
	switch m.focus {
	case focusTitle:
		m.titleInput, cmd = m.titleInput.Update(msg)
	case focusParent:
		oldValue := m.parentInput.Value()
		m.parentInput, cmd = m.parentInput.Update(msg)
		// If value changed, filter and show dropdown
		if m.parentInput.Value() != oldValue {
			m.filterParents()
			m.showDropdown = len(m.parentInput.Value()) > 0 && len(m.filteredParents) > 0
			m.parentIndex = 0
		}
	}

	return m, cmd
}

func (m *CreateOverlay) nextField() {
	m.showDropdown = false
	m.focus++
	if m.focus > focusParent {
		m.focus = focusTitle
	}
	m.updateFocus()
}

func (m *CreateOverlay) prevField() {
	m.showDropdown = false
	m.focus--
	if m.focus < focusTitle {
		m.focus = focusParent
	}
	m.updateFocus()
}

func (m *CreateOverlay) updateFocus() {
	m.titleInput.Blur()
	m.parentInput.Blur()
	switch m.focus {
	case focusTitle:
		m.titleInput.Focus()
	case focusParent:
		m.parentInput.Focus()
	}
}

func (m *CreateOverlay) filterParents() {
	filter := strings.ToLower(m.parentInput.Value())
	if filter == "" {
		m.filteredParents = m.parentOptions
		return
	}

	m.filteredParents = nil
	for _, p := range m.parentOptions {
		if strings.Contains(strings.ToLower(p.ID), filter) ||
			strings.Contains(strings.ToLower(p.Display), filter) {
			m.filteredParents = append(m.filteredParents, p)
		}
	}
	// Limit to 5 visible
	if len(m.filteredParents) > 5 {
		m.filteredParents = m.filteredParents[:5]
	}
}

func (m *CreateOverlay) submit() tea.Msg {
	parentID := ""
	if m.selectedParent != nil {
		parentID = m.selectedParent.ID
	}

	return BeadCreatedMsg{
		Title:     strings.TrimSpace(m.titleInput.Value()),
		IssueType: typeOptions[m.typeIndex],
		Priority:  m.priorityIndex,
		ParentID:  parentID,
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

	styleDropdownItem = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				PaddingLeft(2)

	styleDropdownSelected = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true).
				PaddingLeft(1)

	styleMatchHighlight = lipgloss.NewStyle().
				Foreground(lipgloss.Color("212")).
				Bold(true)

	styleCreateError = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Italic(true)

	styleTypeDesc = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)

	styleSelectedParent = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86"))

	styleParentClear = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))
)

// highlightMatch highlights the matching portion of text
func highlightMatch(text, filter string) string {
	if filter == "" {
		return text
	}
	lowerText := strings.ToLower(text)
	lowerFilter := strings.ToLower(filter)
	idx := strings.Index(lowerText, lowerFilter)
	if idx == -1 {
		return text
	}
	// Build highlighted string
	before := text[:idx]
	match := text[idx : idx+len(filter)]
	after := text[idx+len(filter):]
	return before + styleMatchHighlight.Render(match) + after
}

// View implements tea.Model.
func (m *CreateOverlay) View() string {
	var b strings.Builder

	// Title
	title := styleHelpTitle.Render("CREATE NEW BEAD")
	divider := styleHelpDivider.Render(strings.Repeat("─", 48))

	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(divider)
	b.WriteString("\n\n")

	// Title field
	titleLabel := styleCreateLabel.Render("Title")
	b.WriteString(titleLabel)
	b.WriteString("\n")

	titleStyle := styleCreateInput
	if m.focus == focusTitle {
		titleStyle = styleCreateInputFocused
	}
	b.WriteString(titleStyle.Render(m.titleInput.View()))

	// Validation hint
	if m.focus == focusTitle && strings.TrimSpace(m.titleInput.Value()) == "" {
		b.WriteString("\n")
		b.WriteString(styleCreateError.Render("  required"))
	}
	b.WriteString("\n\n")

	// Type pills with description
	typeLabel := styleCreateLabel.Render("Type")
	b.WriteString(typeLabel)
	b.WriteString("    ")
	for i, label := range typeLabels {
		style := styleCreatePill
		if i == m.typeIndex {
			if m.focus == focusType {
				style = styleCreatePillFocused
			} else {
				style = styleCreatePillSelected
			}
		}
		b.WriteString(style.Render(label))
		if i < len(typeLabels)-1 {
			b.WriteString(" ")
		}
	}
	b.WriteString("\n")
	// Show description for selected type
	b.WriteString("          ")
	b.WriteString(styleTypeDesc.Render(typeDescriptions[m.typeIndex]))
	b.WriteString("\n\n")

	// Priority pills
	prioLabel := styleCreateLabel.Render("Priority")
	b.WriteString(prioLabel)
	b.WriteString(" ")
	for i, label := range priorityLabels {
		style := styleCreatePill
		if i == m.priorityIndex {
			if m.focus == focusPriority {
				style = styleCreatePillFocused
			} else {
				style = styleCreatePillSelected
			}
		}
		b.WriteString(style.Render(label))
		if i < len(priorityLabels)-1 {
			b.WriteString(" ")
		}
	}
	b.WriteString("\n\n")

	// Parent field - show selected parent or search input
	parentLabel := styleCreateLabel.Render("Parent")
	b.WriteString(parentLabel)

	if m.selectedParent != nil && m.focus != focusParent {
		// Show selected parent with clear hint
		b.WriteString("  ")
		display := truncateTitle(m.selectedParent.Display, 38)
		b.WriteString(styleSelectedParent.Render(display))
		b.WriteString("\n")
	} else if m.selectedParent != nil && m.focus == focusParent && m.parentInput.Value() == "" {
		// Focused with selection but no search - show selected + hint
		b.WriteString("  ")
		display := truncateTitle(m.selectedParent.Display, 38)
		b.WriteString(styleSelectedParent.Render(display))
		b.WriteString(" ")
		b.WriteString(styleParentClear.Render("(⌫ clear)"))
		b.WriteString("\n")
		// Also show search input below
		parentStyle := styleCreateInputFocused
		b.WriteString(parentStyle.Render(m.parentInput.View()))
		b.WriteString("\n")
	} else {
		// No selection or searching
		b.WriteString("\n")
		parentStyle := styleCreateInput
		if m.focus == focusParent {
			parentStyle = styleCreateInputFocused
		}
		b.WriteString(parentStyle.Render(m.parentInput.View()))
		b.WriteString("\n")
	}

	// Dropdown results with highlighted matches
	if m.showDropdown && len(m.filteredParents) > 0 {
		filter := m.parentInput.Value()
		for i, p := range m.filteredParents {
			display := truncateTitle(p.Display, 42)
			highlighted := highlightMatch(display, filter)
			if i == m.parentIndex {
				b.WriteString(styleDropdownSelected.Render("› "))
				b.WriteString(highlighted)
			} else {
				b.WriteString(styleDropdownItem.Render("  "))
				b.WriteString(highlighted)
			}
			b.WriteString("\n")
		}
	} else if m.focus == focusParent && m.selectedParent == nil && m.parentInput.Value() == "" {
		b.WriteString(styleDropdownItem.Render("  (none) - type to search"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(divider)

	return styleHelpOverlay.Render(b.String())
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
	if m.selectedParent != nil {
		return m.selectedParent.ID
	}
	return ""
}
