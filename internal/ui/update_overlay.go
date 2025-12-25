package ui

import (
	"errors"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const epicParentConstraintMessage = "Epics may only be children of other epics. Note: beads does not currently support changing bead types (see GitHub issue #522)."

var errInvalidEpicParent = errors.New("epics may only be children of other epics")

// handleOverlayMsg processes overlay-related messages (status, labels, create, delete, comment).
// Returns (model, cmd, handled). If handled is false, the message was not processed.
func (m *App) handleOverlayMsg(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case StatusChangedMsg:
		m.activeOverlay = OverlayNone
		oldStatus := ""
		if m.statusOverlay != nil {
			oldStatus = m.statusOverlay.currentStatus
		}
		m.statusOverlay = nil
		if msg.NewStatus != "" {
			m.displayStatusToast(msg.IssueID, msg.NewStatus)
			if oldStatus == "closed" && msg.NewStatus == "open" {
				return m, tea.Batch(m.executeReopenCmd(msg.IssueID), scheduleStatusToastTick()), true
			}
			return m, tea.Batch(m.executeStatusChangeCmd(msg.IssueID, msg.NewStatus), scheduleStatusToastTick()), true
		}
		return m, nil, true

	case StatusCancelledMsg:
		m.activeOverlay = OverlayNone
		m.statusOverlay = nil
		return m, nil, true

	case statusUpdateCompleteMsg:
		if msg.err != nil {
			m.lastError = msg.err.Error()
			m.lastErrorSource = errorSourceOperation
			m.showErrorToast = true
			m.errorToastStart = time.Now()
			return m, scheduleErrorToastTick(), true
		}
		return m, m.forceRefresh(), true

	case statusToastTickMsg:
		if !m.statusToastVisible {
			return m, nil, true
		}
		if time.Since(m.statusToastStart) >= 7*time.Second {
			m.statusToastVisible = false
			return m, nil, true
		}
		return m, scheduleStatusToastTick(), true

	case LabelsUpdatedMsg:
		m.activeOverlay = OverlayNone
		m.labelsOverlay = nil
		if len(msg.Added) > 0 || len(msg.Removed) > 0 {
			m.displayLabelsToast(msg.IssueID, msg.Added, msg.Removed)
			return m, tea.Batch(m.executeLabelsUpdate(msg), scheduleLabelsToastTick()), true
		}
		return m, nil, true

	case LabelsCancelledMsg:
		m.activeOverlay = OverlayNone
		m.labelsOverlay = nil
		return m, nil, true

	case ComboBoxEnterSelectedMsg, ComboBoxTabSelectedMsg:
		if m.activeOverlay == OverlayLabels && m.labelsOverlay != nil {
			var labelCmd tea.Cmd
			m.labelsOverlay, labelCmd = m.labelsOverlay.Update(msg)
			return m, labelCmd, true
		}
		if m.activeOverlay == OverlayCreate && m.createOverlay != nil {
			var createCmd tea.Cmd
			m.createOverlay, createCmd = m.createOverlay.Update(msg)
			return m, createCmd, true
		}
		return m, nil, true

	case labelUpdateCompleteMsg:
		if msg.err != nil {
			m.lastError = msg.err.Error()
			m.lastErrorSource = errorSourceOperation
			m.showErrorToast = true
			m.errorToastStart = time.Now()
			return m, scheduleErrorToastTick(), true
		}
		return m, m.forceRefresh(), true

	case labelsToastTickMsg:
		if !m.labelsToastVisible {
			return m, nil, true
		}
		if time.Since(m.labelsToastStart) >= 7*time.Second {
			m.labelsToastVisible = false
			return m, nil, true
		}
		return m, scheduleLabelsToastTick(), true

	case BeadCreatedMsg:
		return m, m.executeCreateBead(msg), true

	case CreateCancelledMsg:
		m.activeOverlay = OverlayNone
		m.createOverlay = nil
		return m, nil, true

	case DismissErrorToastMsg:
		m.showErrorToast = false
		return m, nil, true

	case createCompleteMsg:
		return m.handleCreateComplete(msg)

	case createToastTickMsg:
		if !m.createToastVisible {
			return m, nil, true
		}
		if time.Since(m.createToastStart) >= 7*time.Second {
			m.createToastVisible = false
			return m, nil, true
		}
		return m, scheduleCreateToastTick(), true

	case NewLabelAddedMsg:
		m.displayNewLabelToast(msg.Label)
		return m, scheduleNewLabelToastTick(), true

	case newLabelToastTickMsg:
		if !m.newLabelToastVisible {
			return m, nil, true
		}
		if time.Since(m.newLabelToastStart) >= 3*time.Second {
			m.newLabelToastVisible = false
			return m, nil, true
		}
		return m, scheduleNewLabelToastTick(), true

	case NewAssigneeAddedMsg:
		m.displayNewAssigneeToast(msg.Assignee)
		return m, scheduleNewAssigneeToastTick(), true

	case BeadUpdatedMsg:
		if err := m.validateUpdate(msg); err != nil {
			if errors.Is(err, errInvalidEpicParent) {
				m.lastError = epicParentConstraintMessage
			} else {
				m.lastError = err.Error()
			}
			m.lastErrorSource = errorSourceOperation
			m.showErrorToast = true
			m.errorToastStart = time.Now()
			if m.createOverlay != nil {
				m.createOverlay.isCreating = false
			}
			return m, scheduleErrorToastTick(), true
		}
		return m, m.executeUpdateCmd(msg), true

	case updateCompleteMsg:
		m.activeOverlay = OverlayNone
		m.createOverlay = nil
		if msg.Err != nil {
			m.showErrorToast = true
			m.errorToastStart = time.Now()
			m.lastError = msg.Err.Error()
			m.lastErrorSource = errorSourceOperation
			return m, scheduleErrorToastTick(), true
		}
		m.createToastBeadID = msg.ID
		m.createToastIsUpdate = true
		m.displayCreateToast(msg.Title, true)
		return m, tea.Batch(m.forceRefresh(), scheduleCreateToastTick()), true

	case typeInferenceFlashMsg:
		if m.activeOverlay == OverlayCreate && m.createOverlay != nil {
			var cmd tea.Cmd
			m.createOverlay, cmd = m.createOverlay.Update(msg)
			return m, cmd, true
		}
		return m, nil, true

	case newAssigneeToastTickMsg:
		if !m.newAssigneeToastVisible {
			return m, nil, true
		}
		if time.Since(m.newAssigneeToastStart) >= 3*time.Second {
			m.newAssigneeToastVisible = false
			return m, nil, true
		}
		return m, scheduleNewAssigneeToastTick(), true

	case themeToastTickMsg:
		if !m.themeToastVisible {
			return m, nil, true
		}
		if time.Since(m.themeToastStart) >= 3*time.Second {
			m.themeToastVisible = false
			return m, nil, true
		}
		return m, scheduleThemeToastTick(), true

	case DeleteConfirmedMsg:
		m.activeOverlay = OverlayNone
		m.deleteOverlay = nil
		return m, tea.Batch(m.executeDelete(msg.IssueID, msg.Cascade, msg.Children), scheduleDeleteToastTick()), true

	case DeleteCancelledMsg:
		m.activeOverlay = OverlayNone
		m.deleteOverlay = nil
		return m, nil, true

	case deleteCompleteMsg:
		if msg.err != nil {
			m.lastError = msg.err.Error()
			m.lastErrorSource = errorSourceOperation
			m.showErrorToast = true
			m.errorToastStart = time.Now()
			return m, scheduleErrorToastTick(), true
		}
		m.removeNodeFromTree(msg.issueID)
		for _, childID := range msg.children {
			m.removeNodeFromTree(childID)
		}
		m.recalcVisibleRows()
		return m, m.forceRefresh(), true

	case deleteToastTickMsg:
		if !m.deleteToastVisible {
			return m, nil, true
		}
		if time.Since(m.deleteToastStart) >= 5*time.Second {
			m.deleteToastVisible = false
			return m, nil, true
		}
		return m, scheduleDeleteToastTick(), true

	case CommentAddedMsg:
		m.activeOverlay = OverlayNone
		m.commentOverlay = nil
		return m, tea.Batch(m.executeAddComment(msg), scheduleCommentToastTick()), true

	case CommentCancelledMsg:
		m.activeOverlay = OverlayNone
		m.commentOverlay = nil
		return m, nil, true

	case commentCompleteMsg:
		if msg.err != nil {
			m.lastError = msg.err.Error()
			m.lastErrorSource = errorSourceOperation
			m.showErrorToast = true
			m.errorToastStart = time.Now()
			return m, scheduleErrorToastTick(), true
		}
		m.displayCommentToast(msg.issueID)
		return m, tea.Batch(m.forceRefresh(), scheduleCommentToastTick()), true

	case commentToastTickMsg:
		if !m.commentToastVisible {
			return m, nil, true
		}
		if time.Since(m.commentToastStart) >= 7*time.Second {
			m.commentToastVisible = false
			return m, nil, true
		}
		return m, scheduleCommentToastTick(), true
	}

	return nil, nil, false
}

// handleCreateComplete processes the createCompleteMsg with fast injection support.
func (m *App) handleCreateComplete(msg createCompleteMsg) (tea.Model, tea.Cmd, bool) {
	if msg.err != nil {
		errMsg := msg.err.Error()
		m.lastError = errMsg
		m.lastErrorSource = errorSourceOperation
		m.showErrorToast = true
		m.errorToastStart = time.Now()

		if m.activeOverlay == OverlayCreate && m.createOverlay != nil {
			cmd := func() tea.Msg {
				return backendErrorMsg{
					err:    msg.err,
					errMsg: errMsg,
				}
			}
			return m, tea.Batch(cmd, scheduleErrorToastTick()), true
		}
		return m, scheduleErrorToastTick(), true
	}

	// Fast injection path (if fullIssue available)
	if msg.fullIssue != nil {
		if err := m.fastInjectBead(*msg.fullIssue, msg.parentID); err != nil {
			m.lastError = fmt.Sprintf("Fast injection failed: %v, refreshing...", err)
			m.lastErrorSource = errorSourceOperation
		} else {
			m.createToastBeadID = msg.id
			m.displayCreateToast("", false)

			if m.activeOverlay == OverlayCreate && m.createOverlay != nil {
				m.activeOverlay = OverlayNone
				m.createOverlay = nil
			}

			return m, tea.Batch(scheduleCreateToastTick(), m.scheduleEventualRefresh()), true
		}
	}

	// Fallback: Success with full refresh
	if m.activeOverlay == OverlayCreate && m.createOverlay != nil {
		m.activeOverlay = OverlayNone
		m.createOverlay = nil
	}

	m.createToastBeadID = msg.id
	m.displayCreateToast("", false)
	return m, tea.Batch(m.forceRefresh(), scheduleCreateToastTick()), true
}
