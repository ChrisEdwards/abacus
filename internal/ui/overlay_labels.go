package ui

import (
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LabelsOverlay is a chip-based popup for managing labels on a bead.
// Uses ChipComboBox for a modern, intuitive label management experience.
type LabelsOverlay struct {
	issueID       string
	beadTitle     string       // For header display
	chipCombo     ChipComboBox // Main component
	originalChips []string     // For cancel/revert
	allLabels     []string     // All project labels (sorted)
}

// LabelsUpdatedMsg is sent when label changes are confirmed.
type LabelsUpdatedMsg struct {
	IssueID string
	Added   []string
	Removed []string
}

// LabelsCancelledMsg is sent when the overlay is dismissed without changes.
type LabelsCancelledMsg struct{}

// NewLabelsOverlay creates a new chip-based labels overlay for the given issue.
func NewLabelsOverlay(issueID, beadTitle string, currentLabels, allProjectLabels []string) *LabelsOverlay {
	// Sort all labels for consistent display
	sortedLabels := make([]string, len(allProjectLabels))
	copy(sortedLabels, allProjectLabels)
	sort.Strings(sortedLabels)

	// Store original chips for cancel/revert
	originalChips := make([]string, len(currentLabels))
	copy(originalChips, currentLabels)

	// Create ChipComboBox with all labels as options
	chipCombo := NewChipComboBox(sortedLabels).
		WithWidth(40).
		WithMaxVisible(5).
		WithPlaceholder("type to filter...").
		WithAllowNew(true, "New label: %s")

	// Pre-populate with current labels
	chipCombo.SetChips(currentLabels)

	return &LabelsOverlay{
		issueID:       issueID,
		beadTitle:     beadTitle,
		chipCombo:     chipCombo,
		originalChips: originalChips,
		allLabels:     sortedLabels,
	}
}

// Init implements tea.Model.
func (m *LabelsOverlay) Init() tea.Cmd {
	return m.chipCombo.Focus()
}

// Update implements tea.Model.
func (m *LabelsOverlay) Update(msg tea.Msg) (*LabelsOverlay, tea.Cmd) {
	switch msg := msg.(type) {
	case ChipComboBoxTabMsg:
		// Tab in labels overlay does nothing (unlike create modal)
		// Just pass through - no confirm
		return m, nil

	case ChipComboBoxChipAddedMsg:
		// Chip was added - visual feedback already happened, no action needed
		return m, nil

	case ChipRemovedMsg:
		// Chip was removed - visual feedback already happened, no action needed
		return m, nil

	case chipFlashClearMsg:
		// Pass through to chipCombo
		m.chipCombo, _ = m.chipCombo.Update(msg)
		return m, nil

	case ComboBoxValueSelectedMsg:
		// Forward to chipCombo to add as chip
		var cmd tea.Cmd
		m.chipCombo, cmd = m.chipCombo.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		// Handle global keys first
		switch msg.Type {
		case tea.KeyEsc:
			return m.handleEscape()

		case tea.KeyEnter:
			// Enter confirms if idle (dropdown closed, input empty, not in chip nav)
			if m.isIdle() {
				return m, m.confirm()
			}
			// Otherwise pass to ChipComboBox
		}

		// Pass to ChipComboBox
		var cmd tea.Cmd
		m.chipCombo, cmd = m.chipCombo.Update(msg)
		return m, cmd
	}

	// Pass other messages to ChipComboBox
	var cmd tea.Cmd
	m.chipCombo, cmd = m.chipCombo.Update(msg)
	return m, cmd
}

// handleEscape implements multi-stage escape:
// 1. If dropdown open: close dropdown
// 2. If input has text: clear input
// 3. If in chip nav mode: exit chip nav
// 4. Otherwise: cancel and close
func (m *LabelsOverlay) handleEscape() (*LabelsOverlay, tea.Cmd) {
	// Check if dropdown is open
	if m.chipCombo.IsDropdownOpen() {
		// Close dropdown
		m.chipCombo, _ = m.chipCombo.Update(tea.KeyMsg{Type: tea.KeyEsc})
		return m, nil
	}

	// Check if input has text
	if m.chipCombo.InputValue() != "" {
		// Clear input
		m.chipCombo, _ = m.chipCombo.Update(tea.KeyMsg{Type: tea.KeyEsc})
		return m, nil
	}

	// Check if in chip nav mode
	if m.chipCombo.InChipNavMode() {
		// Exit chip nav
		m.chipCombo, _ = m.chipCombo.Update(tea.KeyMsg{Type: tea.KeyEsc})
		return m, nil
	}

	// Cancel and close - discard all changes
	return m, func() tea.Msg { return LabelsCancelledMsg{} }
}

// isIdle returns true if the ChipComboBox is in idle state
// (dropdown closed, input empty, not in chip nav mode).
func (m *LabelsOverlay) isIdle() bool {
	return !m.chipCombo.IsDropdownOpen() &&
		m.chipCombo.InputValue() == "" &&
		!m.chipCombo.InChipNavMode()
}

// computeDiff calculates which labels were added and removed
// compared to the original state.
func (m *LabelsOverlay) computeDiff() (added, removed []string) {
	currentChips := m.chipCombo.GetChips()

	// Build sets for comparison
	originalSet := make(map[string]bool)
	for _, l := range m.originalChips {
		originalSet[l] = true
	}

	currentSet := make(map[string]bool)
	for _, l := range currentChips {
		currentSet[l] = true
	}

	// Find added (in current but not original)
	for _, l := range currentChips {
		if !originalSet[l] {
			added = append(added, l)
		}
	}

	// Find removed (in original but not current)
	for _, l := range m.originalChips {
		if !currentSet[l] {
			removed = append(removed, l)
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

	// Header: Line 1 = "Edit Labels", Line 2 = "ab-xxx: Bead Title"
	b.WriteString(styleHelpSectionHeader.Render("Edit Labels"))
	b.WriteString("\n")

	// Truncate title if too long
	title := m.beadTitle
	maxTitleLen := 30
	if len(title) > maxTitleLen {
		title = title[:maxTitleLen-3] + "..."
	}
	contextLine := styleID.Render(m.issueID) + styleStatsDim.Render(": ") + styleStatsDim.Render(title)
	b.WriteString(contextLine)
	b.WriteString("\n")

	// Divider
	divider := styleStatusDivider.Render(strings.Repeat("─", 44))
	b.WriteString(divider)
	b.WriteString("\n\n")

	// ChipComboBox
	b.WriteString(m.chipCombo.View())
	b.WriteString("\n\n")

	// Footer
	b.WriteString(divider)
	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	return styleStatusOverlay.Render(b.String())
}

// renderFooter returns the dynamic footer based on current state.
func (m *LabelsOverlay) renderFooter() string {
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))

	switch {
	case m.chipCombo.IsDropdownOpen():
		// Dropdown open: show selection hints
		return footerStyle.Render("Enter Select • ↑↓ Navigate • Esc Clear")
	case m.chipCombo.InChipNavMode():
		// Chip navigation mode: show chip nav hints
		return footerStyle.Render("Delete Remove • ←→ Navigate • Esc Exit")
	default:
		// Idle state: show confirm/cancel hints
		return footerStyle.Render("Enter Save • Esc Cancel")
	}
}

// IssueID returns the issue ID (for testing).
func (m *LabelsOverlay) IssueID() string {
	return m.issueID
}

// BeadTitle returns the bead title (for testing).
func (m *LabelsOverlay) BeadTitle() string {
	return m.beadTitle
}

// GetChips returns the current chips (for testing).
func (m *LabelsOverlay) GetChips() []string {
	return m.chipCombo.GetChips()
}

// OriginalChips returns the original chips (for testing).
func (m *LabelsOverlay) OriginalChips() []string {
	return m.originalChips
}

// IsIdle returns whether the overlay is in idle state (for testing).
func (m *LabelsOverlay) IsIdle() bool {
	return m.isIdle()
}
