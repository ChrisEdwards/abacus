package ui

import (
	"context"
	"fmt"
	"sort"
	"time"

	"abacus/internal/beads"
	"abacus/internal/config"
	"abacus/internal/graph"
	"abacus/internal/ui/theme"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle overlay messages regardless of overlay state
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
			// Use Reopen command when transitioning from closed to open
			if oldStatus == "closed" && msg.NewStatus == "open" {
				return m, tea.Batch(m.executeReopenCmd(msg.IssueID), scheduleStatusToastTick())
			}
			return m, tea.Batch(m.executeStatusChangeCmd(msg.IssueID, msg.NewStatus), scheduleStatusToastTick())
		}
		return m, nil
	case StatusCancelledMsg:
		m.activeOverlay = OverlayNone
		m.statusOverlay = nil
		return m, nil
	case statusUpdateCompleteMsg:
		if msg.err != nil {
			m.lastError = msg.err.Error()
			m.showErrorToast = true
			m.errorToastStart = time.Now()
			return m, scheduleErrorToastTick()
		}
		return m, m.forceRefresh()
	case statusToastTickMsg:
		if !m.statusToastVisible {
			return m, nil
		}
		if time.Since(m.statusToastStart) >= 7*time.Second {
			m.statusToastVisible = false
			return m, nil
		}
		return m, scheduleStatusToastTick()
	case LabelsUpdatedMsg:
		m.activeOverlay = OverlayNone
		m.labelsOverlay = nil
		if len(msg.Added) > 0 || len(msg.Removed) > 0 {
			m.displayLabelsToast(msg.IssueID, msg.Added, msg.Removed)
			return m, tea.Batch(m.executeLabelsUpdate(msg), scheduleLabelsToastTick())
		}
		return m, nil
	case LabelsCancelledMsg:
		m.activeOverlay = OverlayNone
		m.labelsOverlay = nil
		return m, nil
	case ComboBoxValueSelectedMsg:
		// Route to labelsOverlay if active (for adding chips)
		if m.activeOverlay == OverlayLabels && m.labelsOverlay != nil {
			var labelCmd tea.Cmd
			m.labelsOverlay, labelCmd = m.labelsOverlay.Update(msg)
			return m, labelCmd
		}
		// Otherwise fall through to default handling
		return m, nil
	case labelUpdateCompleteMsg:
		if msg.err != nil {
			m.lastError = msg.err.Error()
			m.showErrorToast = true
			m.errorToastStart = time.Now()
			return m, scheduleErrorToastTick()
		}
		return m, m.forceRefresh()
	case labelsToastTickMsg:
		if !m.labelsToastVisible {
			return m, nil
		}
		if time.Since(m.labelsToastStart) >= 7*time.Second {
			m.labelsToastVisible = false
			return m, nil
		}
		return m, scheduleLabelsToastTick()
	case BeadCreatedMsg:
		// Don't close overlay yet - wait for backend confirmation (spec Section 4.4)
		// Modal will close in createCompleteMsg if successful
		return m, m.executeCreateBead(msg)
	case CreateCancelledMsg:
		m.activeOverlay = OverlayNone
		m.createOverlay = nil
		return m, nil
	case DismissErrorToastMsg:
		// Dismiss the global error toast (sent from overlay when ESC pressed with error)
		m.showErrorToast = false
		return m, nil
	case createCompleteMsg:
		if msg.err != nil {
			// Backend error: show toast and notify overlay (spec Section 4.4)
			errMsg := msg.err.Error()
			m.lastError = errMsg
			m.showErrorToast = true
			m.errorToastStart = time.Now()

			if m.activeOverlay == OverlayCreate && m.createOverlay != nil {
				// Also notify overlay so it knows ESC should dismiss toast first
				cmd := func() tea.Msg {
					return backendErrorMsg{
						err:    msg.err,
						errMsg: errMsg,
					}
				}
				return m, tea.Batch(cmd, scheduleErrorToastTick())
			}
			return m, scheduleErrorToastTick()
		}

		// NEW: Fast injection path (if fullIssue available)
		if msg.fullIssue != nil {
			if err := m.fastInjectBead(*msg.fullIssue, msg.parentID); err != nil {
				// Fall back to full refresh on error
				m.lastError = fmt.Sprintf("Fast injection failed: %v, refreshing...", err)
				// Continue to full refresh below
			} else {
				// Success! Show toast and return
				m.createToastBeadID = msg.id
				m.displayCreateToast("")

				// Close modal if not in bulk mode (spec Section 4.3)
				if m.activeOverlay == OverlayCreate && m.createOverlay != nil {
					if !msg.stayOpen {
						m.activeOverlay = OverlayNone
						m.createOverlay = nil
					}
					// else: bulk mode, keep overlay open for next entry
				}

				// Schedule eventual consistency refresh (2 seconds delay)
				return m, tea.Batch(
					scheduleCreateToastTick(),
					m.scheduleEventualRefresh(),
				)
			}
		}

		// Fallback: Success with full refresh (old path or injection failed)
		if m.activeOverlay == OverlayCreate && m.createOverlay != nil {
			if !msg.stayOpen {
				m.activeOverlay = OverlayNone
				m.createOverlay = nil
			}
			// else: bulk mode, keep overlay open for next entry
		}

		// Show success toast and refresh
		m.createToastBeadID = msg.id
		m.displayCreateToast("")
		return m, tea.Batch(m.forceRefresh(), scheduleCreateToastTick())
	case createToastTickMsg:
		if !m.createToastVisible {
			return m, nil
		}
		if time.Since(m.createToastStart) >= 7*time.Second {
			m.createToastVisible = false
			return m, nil
		}
		return m, scheduleCreateToastTick()
	case NewLabelAddedMsg:
		// New label was created during bead creation - show toast
		m.displayNewLabelToast(msg.Label)
		return m, scheduleNewLabelToastTick()
	case newLabelToastTickMsg:
		if !m.newLabelToastVisible {
			return m, nil
		}
		if time.Since(m.newLabelToastStart) >= 3*time.Second {
			m.newLabelToastVisible = false
			return m, nil
		}
		return m, scheduleNewLabelToastTick()
	case NewAssigneeAddedMsg:
		// New assignee was created during bead creation - show toast
		m.displayNewAssigneeToast(msg.Assignee)
		return m, scheduleNewAssigneeToastTick()
	case typeInferenceFlashMsg:
		// Forward to CreateOverlay to clear the type inference flash (ab-i0ye)
		if m.activeOverlay == OverlayCreate && m.createOverlay != nil {
			var cmd tea.Cmd
			m.createOverlay, cmd = m.createOverlay.Update(msg)
			return m, cmd
		}
		return m, nil
	case newAssigneeToastTickMsg:
		if !m.newAssigneeToastVisible {
			return m, nil
		}
		if time.Since(m.newAssigneeToastStart) >= 3*time.Second {
			m.newAssigneeToastVisible = false
			return m, nil
		}
		return m, scheduleNewAssigneeToastTick()
	case themeToastTickMsg:
		if !m.themeToastVisible {
			return m, nil
		}
		if time.Since(m.themeToastStart) >= 3*time.Second {
			m.themeToastVisible = false
			return m, nil
		}
		return m, scheduleThemeToastTick()
	case DeleteConfirmedMsg:
		m.activeOverlay = OverlayNone
		m.deleteOverlay = nil
		return m, tea.Batch(m.executeDelete(msg.IssueID, msg.Cascade, msg.Children), scheduleDeleteToastTick())
	case DeleteCancelledMsg:
		m.activeOverlay = OverlayNone
		m.deleteOverlay = nil
		return m, nil
	case deleteCompleteMsg:
		if msg.err != nil {
			m.lastError = msg.err.Error()
			m.showErrorToast = true
			m.errorToastStart = time.Now()
			return m, scheduleErrorToastTick()
		}
		// Immediately remove from tree for instant visual feedback
		m.removeNodeFromTree(msg.issueID)
		for _, childID := range msg.children {
			m.removeNodeFromTree(childID)
		}
		m.recalcVisibleRows()
		return m, m.forceRefresh()
	case deleteToastTickMsg:
		if !m.deleteToastVisible {
			return m, nil
		}
		if time.Since(m.deleteToastStart) >= 5*time.Second {
			m.deleteToastVisible = false
			return m, nil
		}
		return m, scheduleDeleteToastTick()
	}

	// Handle background messages before delegating to overlays
	// This ensures auto-refresh continues even when overlays are open
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.refreshInFlight {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	case tickMsg:
		if m.refreshInterval <= 0 {
			return m, nil // Only stop if interval is invalid
		}
		cmds := []tea.Cmd{}
		if m.autoRefresh {
			if refreshCmd := m.checkDBForChanges(); refreshCmd != nil {
				cmds = append(cmds, refreshCmd)
			}
		}
		// Always reschedule tick so loop can recover if autoRefresh is re-enabled
		cmds = append(cmds, scheduleTick(m.refreshInterval))
		return m, tea.Batch(cmds...)
	case refreshCompleteMsg:
		m.refreshInFlight = false
		if msg.err != nil {
			m.lastError = msg.err.Error()
			m.lastRefreshStats = "" // Clear stats when we have an error
			if !m.errorShownOnce {
				m.showErrorToast = true
				m.errorToastStart = time.Now()
				m.errorShownOnce = true
				return m, scheduleErrorToastTick()
			}
			return m, nil
		}
		// On success, clear error state
		m.lastError = ""
		m.errorShownOnce = false
		m.showErrorToast = false
		m.applyRefresh(msg.roots, msg.digest, msg.dbModTime)
		return m, nil
	case eventualRefreshMsg:
		// Only refresh if not actively creating beads
		if m.activeOverlay != OverlayCreate {
			return m, m.forceRefresh()
		}
		// If still in create overlay, skip refresh (user is bulk creating)
		return m, nil
	case errorToastTickMsg:
		if !m.showErrorToast {
			return m, nil
		}
		elapsed := time.Since(m.errorToastStart)
		if elapsed >= 10*time.Second {
			m.showErrorToast = false
			return m, nil
		}
		return m, scheduleErrorToastTick()
	case copyToastTickMsg:
		if !m.showCopyToast {
			return m, nil
		}
		if time.Since(m.copyToastStart) >= 5*time.Second {
			m.showCopyToast = false
			return m, nil
		}
		return m, scheduleCopyToastTick()
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		rawViewportWidth := int(float64(msg.Width)*0.45) - 2
		maxViewportWidth := msg.Width - minTreeWidth - 4
		m.viewport.Width = clampDimension(rawViewportWidth, minViewportWidth, maxViewportWidth)

		rawViewportHeight := msg.Height - 5
		maxViewportHeight := msg.Height - 2
		m.viewport.Height = clampDimension(rawViewportHeight, minViewportHeight, maxViewportHeight)
		m.applyViewportTheme()
		m.updateViewportContent()

	case tea.KeyMsg:
		// Help overlay takes precedence - blocks all other keys
		if m.showHelp {
			switch {
			case key.Matches(msg, m.keys.Help),
				key.Matches(msg, m.keys.Escape),
				key.Matches(msg, m.keys.Quit):
				m.showHelp = false
			}
			return m, nil
		}

		if m.searching {
			switch {
			case key.Matches(msg, m.keys.Enter):
				m.searching = false
				m.textInput.Blur()
				return m, nil
			case key.Matches(msg, m.keys.Escape):
				m.clearSearchFilter()
				return m, nil
			default:
				m.textInput, cmd = m.textInput.Update(msg)
				m.setFilterText(m.textInput.Value())
				m.recalcVisibleRows()
				return m, cmd
			}
		}

		// Delegate to overlays BEFORE global keys (overlays get priority)
		// This prevents global hotkeys from interfering with text input
		if m.activeOverlay == OverlayStatus && m.statusOverlay != nil {
			m.statusOverlay, cmd = m.statusOverlay.Update(msg)
			return m, cmd
		}

		if m.activeOverlay == OverlayLabels && m.labelsOverlay != nil {
			m.labelsOverlay, cmd = m.labelsOverlay.Update(msg)
			return m, cmd
		}

		if m.activeOverlay == OverlayCreate && m.createOverlay != nil {
			m.createOverlay, cmd = m.createOverlay.Update(msg)
			return m, cmd
		}

		if m.activeOverlay == OverlayDelete && m.deleteOverlay != nil {
			m.deleteOverlay, cmd = m.deleteOverlay.Update(msg)
			return m, cmd
		}

		if handled, detailCmd := m.handleDetailNavigationKey(msg); handled {
			return m, detailCmd
		}

		switch {
		case key.Matches(msg, m.keys.Search):
			if !m.searching {
				m.searching = true
				m.textInput.Focus()
				m.textInput.SetValue(m.filterText)
				m.textInput.SetCursor(len(m.filterText))
			}
		case key.Matches(msg, m.keys.Escape):
			// ESC dismisses error toast first, then clears search filter
			if m.showErrorToast {
				m.showErrorToast = false
				return m, nil
			}
			if m.filterText != "" {
				m.clearSearchFilter()
				return m, nil
			}
		case key.Matches(msg, m.keys.Tab):
			if m.ShowDetails {
				if m.focus == FocusTree {
					m.focus = FocusDetails
				} else {
					m.focus = FocusTree
				}
			}
		case key.Matches(msg, m.keys.ShiftTab):
			if m.ShowDetails {
				if m.focus == FocusDetails {
					m.focus = FocusTree
				} else {
					m.focus = FocusDetails
				}
			}
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Enter):
			m.ShowDetails = !m.ShowDetails
			m.focus = FocusTree
			m.updateViewportContent()
		case key.Matches(msg, m.keys.Refresh):
			if refreshCmd := m.forceRefresh(); refreshCmd != nil {
				return m, refreshCmd
			}
		case key.Matches(msg, m.keys.Down):
			m.cursor++
			m.clampCursor()
			m.updateViewportContent()
		case key.Matches(msg, m.keys.Up):
			m.cursor--
			m.clampCursor()
			m.updateViewportContent()
		case key.Matches(msg, m.keys.Home):
			m.cursor = 0
			m.clampCursor()
			m.updateViewportContent()
		case key.Matches(msg, m.keys.End):
			m.cursor = len(m.visibleRows) - 1
			m.clampCursor()
			m.updateViewportContent()
		case key.Matches(msg, m.keys.PageDown):
			m.cursor += clampDimension(m.viewport.Height, 1, len(m.visibleRows))
			m.clampCursor()
			m.updateViewportContent()
		case key.Matches(msg, m.keys.PageUp):
			m.cursor -= clampDimension(m.viewport.Height, 1, len(m.visibleRows))
			m.clampCursor()
			m.updateViewportContent()
		case key.Matches(msg, m.keys.Space), key.Matches(msg, m.keys.Right):
			if len(m.visibleRows) == 0 {
				return m, nil
			}
			row := m.visibleRows[m.cursor]
			if len(row.Node.Children) > 0 {
				if m.isNodeExpandedInView(row) {
					m.collapseNodeForView(row)
				} else {
					m.expandNodeForView(row)
				}
				m.recalcVisibleRows()
			}
		case key.Matches(msg, m.keys.Left):
			if len(m.visibleRows) == 0 {
				return m, nil
			}
			row := m.visibleRows[m.cursor]
			if len(row.Node.Children) > 0 && m.isNodeExpandedInView(row) {
				m.collapseNodeForView(row)
				m.recalcVisibleRows()
			}
		case key.Matches(msg, m.keys.Delete):
			// Delete key opens delete confirmation (only when tree focused)
			if m.activeOverlay == OverlayNone && !m.searching && len(m.visibleRows) > 0 {
				row := m.visibleRows[m.cursor]
				childInfo, descendantIDs := collectChildInfo(row.Node)
				m.deleteOverlay = NewDeleteOverlay(row.Node.Issue.ID, row.Node.Issue.Title, childInfo, descendantIDs)
				m.activeOverlay = OverlayDelete
			}
			return m, nil
		case key.Matches(msg, m.keys.Backspace):
			// Backspace deletes filter chars if filter is active
			if !m.ShowDetails && !m.searching && len(m.filterText) > 0 {
				m.setFilterText(m.filterText[:len(m.filterText)-1])
				m.recalcVisibleRows()
				m.updateViewportContent()
				return m, nil
			}
			// Backspace also opens delete confirmation when no filter is active
			if m.activeOverlay == OverlayNone && !m.searching && m.filterText == "" && len(m.visibleRows) > 0 {
				row := m.visibleRows[m.cursor]
				childInfo, descendantIDs := collectChildInfo(row.Node)
				m.deleteOverlay = NewDeleteOverlay(row.Node.Issue.ID, row.Node.Issue.Title, childInfo, descendantIDs)
				m.activeOverlay = OverlayDelete
			}
		case key.Matches(msg, m.keys.Copy):
			if len(m.visibleRows) > 0 {
				id := m.visibleRows[m.cursor].Node.Issue.ID
				if err := clipboard.WriteAll(id); err == nil {
					m.copiedBeadID = id
					m.showCopyToast = true
					m.copyToastStart = time.Now()
					return m, scheduleCopyToastTick()
				}
			}
		case key.Matches(msg, m.keys.Theme):
			// Cycle to next theme and show toast
			newTheme := theme.CycleTheme()
			_ = config.SaveTheme(newTheme) // Persist theme (ignore errors to avoid disrupting UX)
			m.applyViewportTheme()
			m.themeToastVisible = true
			m.themeToastStart = time.Now()
			m.themeToastName = newTheme
			// Force viewport refresh to apply new theme colors
			m.detailIssueID = ""
			m.updateViewportContent()
			return m, scheduleThemeToastTick()
		case key.Matches(msg, m.keys.ThemePrev):
			// Cycle to previous theme and show toast
			newTheme := theme.CyclePreviousTheme()
			_ = config.SaveTheme(newTheme) // Persist theme (ignore errors to avoid disrupting UX)
			m.applyViewportTheme()
			m.themeToastVisible = true
			m.themeToastStart = time.Now()
			m.themeToastName = newTheme
			// Force viewport refresh to apply new theme colors
			m.detailIssueID = ""
			m.updateViewportContent()
			return m, scheduleThemeToastTick()
		case key.Matches(msg, m.keys.Error):
			// Show error toast if there's an error and toast isn't already visible
			if m.lastError != "" && !m.showErrorToast {
				m.showErrorToast = true
				m.errorToastStart = time.Now()
				return m, scheduleErrorToastTick()
			}
		case key.Matches(msg, m.keys.Help):
			m.showHelp = true
			return m, nil
		case key.Matches(msg, m.keys.Status):
			if len(m.visibleRows) > 0 {
				row := m.visibleRows[m.cursor]
				m.statusOverlay = NewStatusOverlay(row.Node.Issue.ID, row.Node.Issue.Title, row.Node.Issue.Status)
				m.activeOverlay = OverlayStatus
			}
			return m, nil
		case key.Matches(msg, m.keys.StartWork):
			if len(m.visibleRows) > 0 {
				row := m.visibleRows[m.cursor]
				return m, m.executeStatusChange(row.Node.Issue.ID, "in_progress")
			}
		case key.Matches(msg, m.keys.CloseBead):
			if len(m.visibleRows) > 0 {
				row := m.visibleRows[m.cursor]
				return m, m.executeClose(row.Node.Issue.ID)
			}
		case key.Matches(msg, m.keys.Labels):
			if len(m.visibleRows) > 0 {
				row := m.visibleRows[m.cursor]
				allLabels := m.getAllLabels()
				m.labelsOverlay = NewLabelsOverlay(
					row.Node.Issue.ID,
					row.Node.Issue.Title,
					row.Node.Issue.Labels,
					allLabels,
				)
				m.activeOverlay = OverlayLabels
				return m, m.labelsOverlay.Init()
			}
			return m, nil
		case key.Matches(msg, m.keys.NewBead):
			defaultParent := ""
			if len(m.visibleRows) > 0 {
				defaultParent = m.visibleRows[m.cursor].Node.Issue.ID
			}
			m.createOverlay = NewCreateOverlay(CreateOverlayOptions{
				DefaultParentID:    defaultParent,
				AvailableParents:   m.getAvailableParents(),
				AvailableLabels:    m.getAllLabels(),
				AvailableAssignees: m.getAllAssignees(),
				IsRootMode:         false,
			})
			m.activeOverlay = OverlayCreate
			return m, m.createOverlay.Init()
		case key.Matches(msg, m.keys.NewRootBead):
			m.createOverlay = NewCreateOverlay(CreateOverlayOptions{
				DefaultParentID:    "",
				AvailableParents:   m.getAvailableParents(),
				AvailableLabels:    m.getAllLabels(),
				AvailableAssignees: m.getAllAssignees(),
				IsRootMode:         true,
			})
			m.activeOverlay = OverlayCreate
			return m, m.createOverlay.Init()
		}
	default:
		if m.ShowDetails && m.focus == FocusDetails {
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
	}

	return m, cmd
}

// clearSearchFilter exits search mode and removes any applied filter.
// It preserves the current selection by capturing the selected node/parent
// before clearing, expanding ancestors, and restoring the cursor position.
func (m *App) clearSearchFilter() {
	prevFilter := m.filterText
	m.searching = false
	m.textInput.Blur()
	m.textInput.Reset()
	if prevFilter == "" {
		return
	}

	// 1. Capture current selection (node + parent for multi-parent support)
	var selectedNodeID, selectedParentID string
	if len(m.visibleRows) > 0 && m.cursor >= 0 && m.cursor < len(m.visibleRows) {
		row := m.visibleRows[m.cursor]
		selectedNodeID = row.Node.Issue.ID
		if row.Parent != nil {
			selectedParentID = row.Parent.Issue.ID
		}
	}

	// 2. Transfer manually expanded nodes to permanent state
	m.transferFilterExpansionState()

	// 3. Expand ancestors so selected node will be visible
	if selectedNodeID != "" {
		m.expandAncestorsForRow(selectedNodeID, selectedParentID)
	}

	m.setFilterText("")
	m.recalcVisibleRows()

	// 4. Restore cursor to exact row (handles multi-parent)
	if selectedNodeID != "" {
		if !m.restoreCursorToRow(selectedNodeID, selectedParentID) {
			m.restoreCursorToID(selectedNodeID) // Fallback
		}
	}

	m.updateViewportContent()
	// Note: Scrolling is handled automatically by renderTreeView()
}

func (m *App) setFilterText(value string) {
	if m.filterText == value {
		return
	}
	prevEmpty := m.filterText == ""
	newEmpty := value == ""
	m.filterText = value
	m.filterEval = nil
	if newEmpty {
		m.filterCollapsed = nil
		m.filterForcedExpanded = nil
		return
	}
	if prevEmpty {
		m.filterCollapsed = nil
		m.filterForcedExpanded = nil
	}
}

func (m *App) detailFocusActive() bool {
	return m.ShowDetails && m.focus == FocusDetails
}

func (m *App) handleDetailNavigationKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	if !m.detailFocusActive() {
		return false, nil
	}

	switch {
	case key.Matches(msg, m.keys.Home):
		m.viewport.GotoTop()
		return true, nil
	case key.Matches(msg, m.keys.End):
		m.viewport.GotoBottom()
		return true, nil
	case key.Matches(msg, m.keys.PageDown):
		_ = m.viewport.PageDown()
		return true, nil
	case key.Matches(msg, m.keys.PageUp):
		_ = m.viewport.PageUp()
		return true, nil
	}

	if m.isDetailScrollKey(msg) {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return true, cmd
	}

	return false, nil
}

func (m *App) isDetailScrollKey(msg tea.KeyMsg) bool {
	// Standard navigation keys via KeyMap
	if key.Matches(msg, m.keys.Up) ||
		key.Matches(msg, m.keys.Down) ||
		key.Matches(msg, m.keys.Left) ||
		key.Matches(msg, m.keys.Right) ||
		key.Matches(msg, m.keys.PageUp) ||
		key.Matches(msg, m.keys.PageDown) ||
		key.Matches(msg, m.keys.Space) {
		return true
	}
	// Viewport-specific keys (half-page, etc.) not in KeyMap
	switch msg.String() {
	case "f", "b", "d", "u", "ctrl+d", "ctrl+u":
		return true
	}
	return msg.Type == tea.KeySpace
}

// Message types for status operations
type statusUpdateCompleteMsg struct {
	err error
}

type statusToastTickMsg struct{}

func scheduleStatusToastTick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return statusToastTickMsg{}
	})
}

// executeStatusChange runs the bd update command asynchronously and shows toast.
func (m *App) executeStatusChange(issueID, newStatus string) tea.Cmd {
	m.displayStatusToast(issueID, newStatus)
	return tea.Batch(m.executeStatusChangeCmd(issueID, newStatus), scheduleStatusToastTick())
}

// executeStatusChangeCmd runs the bd update command asynchronously without toast.
func (m *App) executeStatusChangeCmd(issueID, newStatus string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.UpdateStatus(context.Background(), issueID, newStatus)
		return statusUpdateCompleteMsg{err: err}
	}
}

// executeClose runs the bd close command asynchronously.
func (m *App) executeClose(issueID string) tea.Cmd {
	m.displayStatusToast(issueID, "closed")
	closeCmd := func() tea.Msg {
		err := m.client.Close(context.Background(), issueID)
		return statusUpdateCompleteMsg{err: err}
	}
	return tea.Batch(closeCmd, scheduleStatusToastTick())
}

// executeReopenCmd runs the bd reopen command asynchronously.
func (m *App) executeReopenCmd(issueID string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.Reopen(context.Background(), issueID)
		return statusUpdateCompleteMsg{err: err}
	}
}

// displayStatusToast displays a success toast for status changes.
func (m *App) displayStatusToast(issueID, newStatus string) {
	m.statusToastNewStatus = newStatus
	m.statusToastBeadID = issueID
	m.statusToastVisible = true
	m.statusToastStart = time.Now()
}

// formatStatusLabel converts a status value to a display label.
func formatStatusLabel(status string) string {
	switch status {
	case "open":
		return "Open"
	case "in_progress":
		return "In Progress"
	case "closed":
		return "Closed"
	default:
		return status
	}
}

// Message types for label operations
type labelUpdateCompleteMsg struct {
	err error
}

type labelsToastTickMsg struct{}

func scheduleLabelsToastTick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return labelsToastTickMsg{}
	})
}

// getAllLabels collects all unique labels from all issues in the tree.
func (m *App) getAllLabels() []string {
	labelSet := make(map[string]bool)
	var collectLabels func([]*graph.Node)
	collectLabels = func(nodes []*graph.Node) {
		for _, n := range nodes {
			for _, l := range n.Issue.Labels {
				labelSet[l] = true
			}
			collectLabels(n.Children)
		}
	}
	collectLabels(m.roots)

	labels := make([]string, 0, len(labelSet))
	for l := range labelSet {
		labels = append(labels, l)
	}
	sort.Strings(labels)
	return labels
}

// getAllAssignees collects all unique assignees from all issues in the tree.
func (m *App) getAllAssignees() []string {
	assigneeSet := make(map[string]bool)
	var collectAssignees func([]*graph.Node)
	collectAssignees = func(nodes []*graph.Node) {
		for _, n := range nodes {
			if a := n.Issue.Assignee; a != "" {
				assigneeSet[a] = true
			}
			collectAssignees(n.Children)
		}
	}
	collectAssignees(m.roots)

	assignees := make([]string, 0, len(assigneeSet))
	for a := range assigneeSet {
		assignees = append(assignees, a)
	}
	sort.Strings(assignees)
	return assignees
}

// executeLabelsUpdate runs the bd label add/remove commands asynchronously.
func (m *App) executeLabelsUpdate(msg LabelsUpdatedMsg) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		for _, label := range msg.Added {
			if err := m.client.AddLabel(ctx, msg.IssueID, label); err != nil {
				return labelUpdateCompleteMsg{err: err}
			}
		}
		for _, label := range msg.Removed {
			if err := m.client.RemoveLabel(ctx, msg.IssueID, label); err != nil {
				return labelUpdateCompleteMsg{err: err}
			}
		}
		return labelUpdateCompleteMsg{err: nil}
	}
}

// displayLabelsToast displays a success toast for label changes.
func (m *App) displayLabelsToast(issueID string, added, removed []string) {
	m.labelsToastBeadID = issueID
	m.labelsToastAdded = added
	m.labelsToastRemoved = removed
	m.labelsToastVisible = true
	m.labelsToastStart = time.Now()
}

// Message types for create operations
type createCompleteMsg struct {
	id        string
	err       error
	stayOpen  bool             // from BeadCreatedMsg (Ctrl+Enter bulk mode)
	fullIssue *beads.FullIssue // NEW: full issue data for fast injection
	parentID  string           // Explicit parent context for fast injection
}

type createToastTickMsg struct{}

func scheduleCreateToastTick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return createToastTickMsg{}
	})
}

// getAvailableParents collects all beads that can be used as parents.
func (m *App) getAvailableParents() []ParentOption {
	var parents []ParentOption
	var collectParents func([]*graph.Node)
	collectParents = func(nodes []*graph.Node) {
		for _, n := range nodes {
			// Create display string: "ab-xxx Title..." (truncated)
			display := n.Issue.ID + " " + truncateTitle(n.Issue.Title, 30)
			parents = append(parents, ParentOption{
				ID:      n.Issue.ID,
				Display: display,
			})
			collectParents(n.Children)
		}
	}
	collectParents(m.roots)
	return parents
}

// executeCreateBead runs the bd create command asynchronously.
func (m *App) executeCreateBead(msg BeadCreatedMsg) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		issue, err := m.client.CreateFull(ctx, msg.Title, msg.IssueType, msg.Priority, msg.Labels, msg.Assignee, msg.Description, msg.ParentID)
		if err != nil {
			return createCompleteMsg{err: err, stayOpen: msg.StayOpen}
		}
		// Note: parent-child dependency is handled by CreateFull via --parent flag
		// NEW: Return full issue for fast injection
		return createCompleteMsg{
			id:        issue.ID,
			stayOpen:  msg.StayOpen,
			fullIssue: &issue, // Pass actual data from database
			parentID:  msg.ParentID,
		}
	}
}

// displayCreateToast displays a success toast for bead creation.
func (m *App) displayCreateToast(title string) {
	m.createToastTitle = title
	m.createToastVisible = true
	m.createToastStart = time.Now()
}

// displayNewLabelToast displays a toast for a newly created label (not in existing options).
func (m *App) displayNewLabelToast(label string) {
	m.newLabelToastLabel = label
	m.newLabelToastVisible = true
	m.newLabelToastStart = time.Now()
}

// displayNewAssigneeToast displays a toast for a newly created assignee (not in existing options).
func (m *App) displayNewAssigneeToast(assignee string) {
	m.newAssigneeToastAssignee = assignee
	m.newAssigneeToastVisible = true
	m.newAssigneeToastStart = time.Now()
}

// Message types for delete operations
type deleteCompleteMsg struct {
	issueID  string
	children []string
	cascade  bool
	err      error
}

type deleteToastTickMsg struct{}

func scheduleDeleteToastTick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return deleteToastTickMsg{}
	})
}

// executeDelete runs the bd delete command asynchronously and shows toast.
func (m *App) executeDelete(issueID string, cascade bool, childIDs []string) tea.Cmd {
	m.displayDeleteToast(issueID, cascade, len(childIDs))
	return func() tea.Msg {
		err := m.client.Delete(context.Background(), issueID, cascade)
		return deleteCompleteMsg{issueID: issueID, children: childIDs, cascade: cascade, err: err}
	}
}

// displayDeleteToast displays a success toast for deletion.
func (m *App) displayDeleteToast(issueID string, cascade bool, childCount int) {
	m.deleteToastBeadID = issueID
	m.deleteToastCascade = cascade
	m.deleteToastChildCount = childCount
	m.deleteToastVisible = true
	m.deleteToastStart = time.Now()
}
