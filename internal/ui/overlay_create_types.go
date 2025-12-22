package ui

import (
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

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
