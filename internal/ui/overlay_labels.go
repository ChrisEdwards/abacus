package ui

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// LabelsOverlay is a compact popup for managing labels on a bead.
type LabelsOverlay struct {
	issueID     string
	allLabels   []string        // All labels used in project (sorted)
	selected    map[string]bool // Currently selected labels
	original    map[string]bool // Original state (for diff)
	cursor      int
	filterInput textinput.Model
}

// LabelsUpdatedMsg is sent when label changes are confirmed.
type LabelsUpdatedMsg struct {
	IssueID string
	Added   []string
	Removed []string
}

// LabelsCancelledMsg is sent when the overlay is dismissed without changes.
type LabelsCancelledMsg struct{}

// NewLabelsOverlay creates a new labels overlay for the given issue.
func NewLabelsOverlay(issueID string, currentLabels, allProjectLabels []string) *LabelsOverlay {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Prompt = ""
	ti.CharLimit = 50
	ti.Focus()

	selected := make(map[string]bool)
	original := make(map[string]bool)
	for _, l := range currentLabels {
		selected[l] = true
		original[l] = true
	}

	// Ensure all labels are sorted
	sortedLabels := make([]string, len(allProjectLabels))
	copy(sortedLabels, allProjectLabels)
	sort.Strings(sortedLabels)

	return &LabelsOverlay{
		issueID:     issueID,
		allLabels:   sortedLabels,
		selected:    selected,
		original:    original,
		filterInput: ti,
	}
}

// Init implements tea.Model.
func (m *LabelsOverlay) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (m *LabelsOverlay) Update(msg tea.Msg) (*LabelsOverlay, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle special keys by type first (more reliable when textinput is focused)
		switch msg.Type {
		case tea.KeyEsc:
			// If filter has text, clear it first; otherwise close overlay
			if m.filterInput.Value() != "" {
				m.filterInput.SetValue("")
				m.cursor = 0
				return m, nil
			}
			return m, func() tea.Msg { return LabelsCancelledMsg{} }
		case tea.KeyEnter:
			return m, m.confirm()
		case tea.KeyUp:
			m.moveUp()
			return m, nil
		case tea.KeyDown:
			m.moveDown()
			return m, nil
		case tea.KeySpace:
			m.toggleCurrent()
			return m, nil
		default:
			// Handle vim-style navigation by string
			switch msg.String() {
			case "j":
				m.moveDown()
				return m, nil
			case "k":
				m.moveUp()
				return m, nil
			default:
				// Pass other keys to the text input for filtering
				var cmd tea.Cmd
				m.filterInput, cmd = m.filterInput.Update(msg)
				// Reset cursor when filter changes
				m.cursor = 0
				return m, cmd
			}
		}
	}
	return m, nil
}

// filteredLabels returns the labels matching the current filter.
func (m *LabelsOverlay) filteredLabels() []string {
	filter := strings.ToLower(m.filterInput.Value())
	if filter == "" {
		return m.allLabels
	}
	var filtered []string
	for _, l := range m.allLabels {
		if strings.Contains(strings.ToLower(l), filter) {
			filtered = append(filtered, l)
		}
	}
	return filtered
}

// hasExactMatch checks if the filter text exactly matches any existing label.
func (m *LabelsOverlay) hasExactMatch() bool {
	filter := strings.ToLower(strings.TrimSpace(m.filterInput.Value()))
	if filter == "" {
		return true // No filter means no new label option
	}
	for _, l := range m.allLabels {
		if strings.ToLower(l) == filter {
			return true
		}
	}
	return false
}

// canAddNew returns true if we should show the "add new" option.
func (m *LabelsOverlay) canAddNew() bool {
	filter := strings.TrimSpace(m.filterInput.Value())
	return filter != "" && !m.hasExactMatch()
}

// visibleItemCount returns the number of items in the list (labels + add new option).
func (m *LabelsOverlay) visibleItemCount() int {
	count := len(m.filteredLabels())
	if m.canAddNew() {
		count++
	}
	return count
}

// moveDown moves the cursor down, wrapping around.
func (m *LabelsOverlay) moveDown() {
	count := m.visibleItemCount()
	if count == 0 {
		return
	}
	m.cursor = (m.cursor + 1) % count
}

// moveUp moves the cursor up, wrapping around.
func (m *LabelsOverlay) moveUp() {
	count := m.visibleItemCount()
	if count == 0 {
		return
	}
	m.cursor = (m.cursor - 1 + count) % count
}

// toggleCurrent toggles the label at the current cursor position.
func (m *LabelsOverlay) toggleCurrent() {
	filtered := m.filteredLabels()

	// Check if cursor is on "add new" option
	if m.canAddNew() && m.cursor == len(filtered) {
		// Add the new label
		newLabel := strings.TrimSpace(m.filterInput.Value())
		m.allLabels = append(m.allLabels, newLabel)
		sort.Strings(m.allLabels)
		m.selected[newLabel] = true
		m.filterInput.SetValue("")
		m.cursor = 0
		return
	}

	if m.cursor < len(filtered) {
		label := filtered[m.cursor]
		m.selected[label] = !m.selected[label]
	}
}

// computeDiff calculates which labels were added and removed.
func (m *LabelsOverlay) computeDiff() (added, removed []string) {
	for label, isSelected := range m.selected {
		wasSelected := m.original[label]
		if isSelected && !wasSelected {
			added = append(added, label)
		}
	}
	for label, wasSelected := range m.original {
		isSelected := m.selected[label]
		if wasSelected && !isSelected {
			removed = append(removed, label)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	return
}

// confirm sends the LabelsUpdatedMsg with the diff.
func (m *LabelsOverlay) confirm() tea.Cmd {
	added, removed := m.computeDiff()
	return func() tea.Msg {
		return LabelsUpdatedMsg{
			IssueID: m.issueID,
			Added:   added,
			Removed: removed,
		}
	}
}

// View implements tea.Model.
func (m *LabelsOverlay) View() string {
	var b strings.Builder

	// Breadcrumb header: ab-xxx › Labels
	header := styleID.Render(m.issueID) + styleStatsDim.Render(" › ") + styleStatsDim.Render("Labels")
	b.WriteString(header)
	b.WriteString("\n")

	// Divider
	divider := styleStatusDivider.Render(strings.Repeat("─", 28))
	b.WriteString(divider)
	b.WriteString("\n")

	// Filter input - show static prompt when empty to avoid textinput display bugs
	filterValue := m.filterInput.Value()
	if filterValue == "" {
		b.WriteString(styleStatsDim.Render("/ type to filter..."))
	} else {
		b.WriteString("/ " + filterValue)
	}
	b.WriteString("\n")

	// Another divider
	b.WriteString(divider)
	b.WriteString("\n")

	// Label list
	filtered := m.filteredLabels()
	if len(filtered) == 0 && !m.canAddNew() {
		b.WriteString(styleStatsDim.Render("  (no labels)"))
		b.WriteString("\n")
	} else {
		for i, label := range filtered {
			checkbox := "[ ]"
			style := styleLabelUnchecked
			if m.selected[label] {
				checkbox = "[x]"
				style = styleLabelChecked
			}

			line := "  " + checkbox + " " + label
			if i == m.cursor {
				// Cursor gets highlight background, overrides other styling
				style = styleLabelCursor
			}

			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}

		// "Add new" option
		if m.canAddNew() {
			newLabel := strings.TrimSpace(m.filterInput.Value())
			line := "  + Add \"" + newLabel + "\""
			style := styleLabelNewOption
			if m.cursor == len(filtered) {
				// Cursor gets highlight background
				style = styleLabelCursor
			}
			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}
	}

	content := b.String()
	return styleStatusOverlay.Render(content)
}
