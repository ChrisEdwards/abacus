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

// Type options
var typeOptions = []string{"task", "feature", "bug", "epic", "chore"}
var typeLabels = []string{"Task", "Feature", "Bug", "Epic", "Chore"}

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
	parentIndex     int // Selected index in filtered list
	showDropdown    bool
	defaultParent   string
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

	// Parent filter input
	pi := textinput.New()
	pi.Placeholder = "(none) type to filter..."
	pi.CharLimit = 50
	pi.Width = 35

	// Pre-fill parent if default exists
	if defaultParentID != "" {
		for _, p := range availableParents {
			if p.ID == defaultParentID {
				pi.SetValue(p.ID)
				break
			}
		}
	}

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

	m.filterParents()
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
				return m, nil
			}
			return m, func() tea.Msg { return CreateCancelledMsg{} }

		case tea.KeyEnter:
			// If in dropdown, select the item
			if m.showDropdown && len(m.filteredParents) > 0 {
				selected := m.filteredParents[m.parentIndex]
				m.parentInput.SetValue(selected.ID)
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

		case tea.KeyCtrlN:
			// Ctrl+N to open/navigate dropdown
			if m.focus == focusParent {
				if !m.showDropdown {
					m.showDropdown = true
					m.parentIndex = 0
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
	// Check if input matches a valid parent ID
	inputVal := strings.TrimSpace(m.parentInput.Value())
	for _, p := range m.parentOptions {
		if p.ID == inputVal {
			parentID = p.ID
			break
		}
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

	styleCreateError = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Italic(true)
)

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

	// Type pills
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

	// Parent field
	parentLabel := styleCreateLabel.Render("Parent")
	b.WriteString(parentLabel)
	b.WriteString("\n")

	parentStyle := styleCreateInput
	if m.focus == focusParent {
		parentStyle = styleCreateInputFocused
	}
	b.WriteString(parentStyle.Render(m.parentInput.View()))
	b.WriteString("\n")

	// Dropdown results
	if m.showDropdown && len(m.filteredParents) > 0 {
		for i, p := range m.filteredParents {
			display := truncateTitle(p.Display, 42)
			if i == m.parentIndex {
				b.WriteString(styleDropdownSelected.Render("› " + display))
			} else {
				b.WriteString(styleDropdownItem.Render(display))
			}
			b.WriteString("\n")
		}
	} else if m.focus == focusParent && m.parentInput.Value() == "" {
		b.WriteString(styleDropdownItem.Render("type to search parents..."))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(divider)
	b.WriteString("\n")
	footer := styleHelpFooter.Render("Tab: Next  ←→: Select  Enter: Submit  Esc: Cancel")
	b.WriteString(footer)

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
	inputVal := strings.TrimSpace(m.parentInput.Value())
	for _, p := range m.parentOptions {
		if p.ID == inputVal {
			return p.ID
		}
	}
	return ""
}
