package ui

import (
	"errors"
	"strings"

	"github.com/charmbracelet/huh"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// errTitleRequired is returned when the title field is empty.
var errTitleRequired = errors.New("title is required")

// CreateOverlay is a full modal form for creating a new bead.
type CreateOverlay struct {
	form          *huh.Form
	defaultParent string
	parentOptions []ParentOption

	// Form field values (bound to form via pointers)
	title     string
	issueType string
	priority  int
	parentID  string
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
	m := &CreateOverlay{
		defaultParent: defaultParentID,
		parentOptions: availableParents,
		issueType:     "task",
		priority:      2,
		parentID:      defaultParentID,
	}

	// Build parent options for the select
	parentOpts := []huh.Option[string]{
		huh.NewOption("(None)", ""),
	}
	for _, p := range availableParents {
		opt := huh.NewOption(p.Display, p.ID)
		if p.ID == defaultParentID {
			opt = opt.Selected(true)
		}
		parentOpts = append(parentOpts, opt)
	}

	// Build priority options
	priorityOpts := []huh.Option[int]{
		huh.NewOption("Critical (0)", 0),
		huh.NewOption("High (1)", 1),
		huh.NewOption("Medium (2)", 2),
		huh.NewOption("Low (3)", 3),
		huh.NewOption("Backlog (4)", 4),
	}

	// Build type options
	typeOpts := []huh.Option[string]{
		huh.NewOption("Task", "task"),
		huh.NewOption("Feature", "feature"),
		huh.NewOption("Bug", "bug"),
		huh.NewOption("Epic", "epic"),
		huh.NewOption("Chore", "chore"),
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Placeholder("Enter bead title...").
				Value(&m.title).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errTitleRequired
					}
					return nil
				}),

			huh.NewSelect[string]().
				Title("Type").
				Options(typeOpts...).
				Value(&m.issueType),

			huh.NewSelect[int]().
				Title("Priority").
				Options(priorityOpts...).
				Value(&m.priority),

			huh.NewSelect[string]().
				Title("Parent").
				Options(parentOpts...).
				Value(&m.parentID),
		),
	).WithShowHelp(false).WithShowErrors(true)

	return m
}

// Init implements tea.Model.
func (m *CreateOverlay) Init() tea.Cmd {
	return m.form.Init()
}

// Update implements tea.Model.
func (m *CreateOverlay) Update(msg tea.Msg) (*CreateOverlay, tea.Cmd) {
	// Check for Esc to cancel
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.Type == tea.KeyEsc {
			return m, func() tea.Msg { return CreateCancelledMsg{} }
		}
	}

	// Update form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	// Check if form completed
	if m.form.State == huh.StateCompleted {
		return m, func() tea.Msg {
			return BeadCreatedMsg{
				Title:     m.title,
				IssueType: m.issueType,
				Priority:  m.priority,
				ParentID:  m.parentID,
			}
		}
	}

	return m, cmd
}

// View implements tea.Model.
func (m *CreateOverlay) View() string {
	// Build styled container
	title := styleHelpTitle.Render("CREATE NEW BEAD")
	divider := styleHelpDivider.Render(strings.Repeat("─", 45))
	footer := styleHelpFooter.Render("Tab: Next    Enter: Submit    Esc: Cancel")

	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		divider,
		"",
		m.form.View(),
		"",
		divider,
		footer,
	)

	return styleHelpOverlay.Render(content)
}

// Title returns the current title value.
func (m *CreateOverlay) Title() string {
	return m.title
}

// IssueType returns the current issue type value.
func (m *CreateOverlay) IssueType() string {
	return m.issueType
}

// Priority returns the current priority value.
func (m *CreateOverlay) Priority() int {
	return m.priority
}

// ParentID returns the current parent ID value.
func (m *CreateOverlay) ParentID() string {
	return m.parentID
}
