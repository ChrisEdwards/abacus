package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *CreateOverlay) handleEscape() (*CreateOverlay, tea.Cmd) {
	if m.hasBackendError {
		m.hasBackendError = false
		return m, func() tea.Msg { return DismissErrorToastMsg{} }
	}

	if m.parentCombo.IsDropdownOpen() {
		m.parentCombo.Blur()
		m.parentCombo.Focus()
		return m, nil
	}
	if m.labelsCombo.IsDropdownOpen() {
		m.labelsCombo, _ = m.labelsCombo.Update(tea.KeyMsg{Type: tea.KeyEsc})
		return m, nil
	}
	if m.assigneeCombo.IsDropdownOpen() {
		m.assigneeCombo, _ = m.assigneeCombo.Update(tea.KeyMsg{Type: tea.KeyEsc})
		return m, nil
	}
	if m.focus == FocusAssignee && m.assigneeCombo.InputValue() != m.assigneeCombo.Value() {
		m.assigneeCombo, _ = m.assigneeCombo.Update(tea.KeyMsg{Type: tea.KeyEsc})
		return m, nil
	}

	if m.focus == FocusParent {
		m.parentCombo.SetValue(m.parentOriginal)
		m.isRootMode = m.parentOriginal == ""
		m.parentCombo.Blur()
		m.focus = FocusTitle
		return m, m.titleInput.Focus()
	}

	return m, func() tea.Msg { return CreateCancelledMsg{} }
}

func (m *CreateOverlay) handleSubmit() (*CreateOverlay, tea.Cmd) {
	if strings.TrimSpace(m.titleInput.Value()) == "" {
		m.titleValidationError = true
		return m, titleFlashCmd()
	}

	m.hasBackendError = false
	m.isCreating = true

	if m.isEditMode() {
		return m, m.submitEdit()
	}

	return m, m.submit()
}

func (m *CreateOverlay) handleTab() (*CreateOverlay, tea.Cmd) {
	var cmds []tea.Cmd

	m.parentCombo.Blur()

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
		if m.isEditMode() {
			m.focus = FocusPriority
		} else {
			m.focus = FocusType
		}
	case FocusType:
		m.focus = FocusPriority
	case FocusPriority:
		m.focus = FocusLabels
		cmds = append(cmds, m.labelsCombo.Focus())
	case FocusLabels:
		var cmd tea.Cmd
		m.labelsCombo, cmd = m.labelsCombo.Update(tea.KeyMsg{Type: tea.KeyTab})
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		m.labelsCombo.Blur()
		m.focus = FocusAssignee
		cmds = append(cmds, m.assigneeCombo.Focus())
	case FocusAssignee:
		var cmd tea.Cmd
		m.assigneeCombo, cmd = m.assigneeCombo.Update(tea.KeyMsg{Type: tea.KeyTab})

		if cmd != nil {
			msg := cmd()
			if vsm, ok := msg.(ComboBoxTabSelectedMsg); ok && vsm.IsNew {
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

	m.parentCombo.Blur()
	m.assigneeCombo.Blur()
	m.labelsCombo.Blur()

	switch m.focus {
	case FocusTitle:
		m.titleInput.Blur()
		m.focus = FocusParent
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
		if m.isEditMode() {
			m.focus = FocusDescription
			cmds = append(cmds, m.descriptionInput.Focus())
		} else {
			m.focus = FocusType
		}
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
		if msg.Type == tea.KeyDelete || msg.Type == tea.KeyBackspace {
			if !m.parentCombo.IsDropdownOpen() {
				m.parentCombo.SetValue("")
				m.isRootMode = true
				return m, nil
			}
		}
		m.parentCombo, cmd = m.parentCombo.Update(msg)
		return m, cmd

	case FocusTitle:
		if m.titleInput.Value() != "" {
			lineInfo := m.titleInput.LineInfo()
			currentHeight := lineInfo.Height
			if currentHeight > 0 && currentHeight < 3 {
				isOnLastRow := lineInfo.RowOffset == currentHeight-1
				contentWidth := m.calcDialogWidth() - 4
				nearLineEnd := lineInfo.ColumnOffset > contentWidth-12
				if isOnLastRow && nearLineEnd {
					m.titleInput.SetHeight(currentHeight + 1)
				}
			}
		}

		oldTitle := m.titleInput.Value()
		m.titleInput, cmd = m.titleInput.Update(msg)
		m.updateTitleHeight()

		newTitle := m.titleInput.Value()
		if newTitle != oldTitle && !m.typeManuallySet {
			if inferredIdx := inferTypeFromTitle(newTitle); inferredIdx != -1 {
				if inferredIdx != m.typeIndex {
					m.typeIndex = inferredIdx
					m.typeInferenceActive = true
					return m, tea.Batch(cmd, typeInferenceFlashCmd())
				}
			}
		}

		return m, cmd

	case FocusDescription:
		m.descriptionInput, cmd = m.descriptionInput.Update(msg)
		return m, cmd

	case FocusType:
		switch msg.Type {
		case tea.KeyLeft:
			if m.typeIndex > 0 {
				m.typeIndex--
				m.typeManuallySet = true
			}
		case tea.KeyRight:
			if m.typeIndex < len(typeOptions)-1 {
				m.typeIndex++
				m.typeManuallySet = true
			}
		case tea.KeyDown:
			m.focus = FocusPriority
		case tea.KeyRunes:
			if len(msg.Runes) > 0 {
				r := msg.Runes[0]
				switch r {
				case 'h':
					if m.typeIndex > 0 {
						m.typeIndex--
						m.typeManuallySet = true
					}
				case 'l':
					if m.typeIndex < len(typeOptions)-1 {
						m.typeIndex++
						m.typeManuallySet = true
					}
				case 'j':
					m.focus = FocusPriority
				case 'k':
				default:
					m.handleTypeHotkey(r)
				}
			}
		}
		return m, nil

	case FocusPriority:
		switch msg.Type {
		case tea.KeyLeft:
			if m.priorityIndex > 0 {
				m.priorityIndex--
			}
		case tea.KeyRight:
			if m.priorityIndex < len(priorityLabels)-1 {
				m.priorityIndex++
			}
		case tea.KeyUp:
			m.focus = FocusType
		case tea.KeyRunes:
			if len(msg.Runes) > 0 {
				r := msg.Runes[0]
				switch r {
				case 'j':
				case 'k':
					m.focus = FocusType
				default:
					m.handlePriorityHotkey(r)
				}
			}
		}
		return m, nil

	case FocusLabels:
		m.labelsCombo, cmd = m.labelsCombo.Update(msg)
		return m, cmd

	case FocusAssignee:
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
		m.typeIndex = 0
		m.typeManuallySet = true
	case 'f', 'F':
		m.typeIndex = 1
		m.typeManuallySet = true
	case 'b', 'B':
		m.typeIndex = 2
		m.typeManuallySet = true
	case 'e', 'E':
		m.typeIndex = 3
		m.typeManuallySet = true
	case 'c', 'C':
		m.typeIndex = 4
		m.typeManuallySet = true
	}
}

func (m *CreateOverlay) handlePriorityHotkey(r rune) {
	switch r {
	case 'c', 'C':
		m.priorityIndex = 0
	case 'h', 'H':
		m.priorityIndex = 1
	case 'm', 'M':
		m.priorityIndex = 2
	case 'l', 'L':
		m.priorityIndex = 3
	case 'b', 'B':
		m.priorityIndex = 4
	}
}

// submit creates BeadCreatedMsg with form data.
func (m *CreateOverlay) submit() tea.Cmd {
	return func() tea.Msg {
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

		return BeadCreatedMsg{
			Title:       strings.TrimSpace(m.titleInput.Value()),
			Description: strings.TrimSpace(m.descriptionInput.Value()),
			IssueType:   typeOptions[m.typeIndex],
			Priority:    m.priorityIndex,
			ParentID:    parentID,
			Labels:      m.labelsCombo.GetChips(),
			Assignee:    m.getAssigneeValue(),
		}
	}
}

// updateTitleHeight dynamically adjusts the title textarea height based on content.
// Shows 1-3 lines depending on how much text wraps at the content width.
func (m *CreateOverlay) updateTitleHeight() {
	text := m.titleInput.Value()
	if text == "" {
		m.titleInput.SetHeight(1)
		return
	}

	lineInfo := m.titleInput.LineInfo()
	lines := lineInfo.Height
	if lines < 1 {
		lines = 1
	}
	if lines > 3 {
		lines = 3
	}
	m.titleInput.SetHeight(lines)
}
