package ui

import (
	"time"

	"abacus/internal/config"
	"abacus/internal/ui/theme"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// handleKeyMsg processes keyboard input when no overlay has focus.
func (m *App) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

	// Allow error recall even when overlays are open, as long as we're
	// not focused on a text input (e.g., create modal fields).
	if key.Matches(msg, m.keys.Error) && m.errorHotkeyAvailable() {
		if m.lastError != "" && !m.showErrorToast {
			m.showErrorToast = true
			m.errorToastStart = time.Now()
			return m, scheduleErrorToastTick()
		}
	}

	if m.searching {
		return m.handleSearchKey(msg)
	}

	// Delegate to overlays BEFORE global keys (overlays get priority)
	if cmd, handled := m.delegateToOverlay(msg); handled {
		return m, cmd
	}

	if handled, detailCmd := m.handleDetailNavigationKey(msg); handled {
		return m, detailCmd
	}

	return m.handleGlobalKey(msg)
}

// handleSearchKey processes keys when in search mode.
func (m *App) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Enter):
		m.searching = false
		m.textInput.Blur()
		return m, nil
	case key.Matches(msg, m.keys.Escape):
		m.clearSearchFilter()
		return m, nil
	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		m.setFilterText(m.textInput.Value())
		m.recalcVisibleRows()
		return m, cmd
	}
}

// delegateToOverlay forwards key events to the active overlay.
// Returns (cmd, handled) where handled indicates if an overlay processed the key.
func (m *App) delegateToOverlay(msg tea.KeyMsg) (tea.Cmd, bool) {
	var cmd tea.Cmd

	if m.activeOverlay == OverlayStatus && m.statusOverlay != nil {
		m.statusOverlay, cmd = m.statusOverlay.Update(msg)
		return cmd, true
	}

	if m.activeOverlay == OverlayLabels && m.labelsOverlay != nil {
		m.labelsOverlay, cmd = m.labelsOverlay.Update(msg)
		return cmd, true
	}

	if m.activeOverlay == OverlayCreate && m.createOverlay != nil {
		m.createOverlay, cmd = m.createOverlay.Update(msg)
		return cmd, true
	}

	if m.activeOverlay == OverlayDelete && m.deleteOverlay != nil {
		m.deleteOverlay, cmd = m.deleteOverlay.Update(msg)
		return cmd, true
	}

	if m.activeOverlay == OverlayComment && m.commentOverlay != nil {
		m.commentOverlay, cmd = m.commentOverlay.Update(msg)
		return cmd, true
	}

	return nil, false
}

// handleGlobalKey processes global hotkeys when no overlay is active.
func (m *App) handleGlobalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Search):
		if !m.searching {
			m.searching = true
			m.textInput.Focus()
			m.textInput.SetValue(m.filterText)
			m.textInput.SetCursor(len(m.filterText))
		}
	case key.Matches(msg, m.keys.Escape):
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
		return m.handleTreeExpand()
	case key.Matches(msg, m.keys.Left):
		return m.handleTreeCollapse()
	case key.Matches(msg, m.keys.Delete):
		return m.handleDeleteKey()
	case key.Matches(msg, m.keys.Backspace):
		return m.handleBackspaceKey()
	case key.Matches(msg, m.keys.Copy):
		return m.handleCopyKey()
	case key.Matches(msg, m.keys.Theme):
		return m.handleThemeKey(true)
	case key.Matches(msg, m.keys.ThemePrev):
		return m.handleThemeKey(false)
	case key.Matches(msg, m.keys.CycleViewMode):
		m.viewMode = m.viewMode.Next()
		m.recalcVisibleRows()
		return m, nil
	case key.Matches(msg, m.keys.CycleViewModeBack):
		m.viewMode = m.viewMode.Prev()
		m.recalcVisibleRows()
		return m, nil
	case key.Matches(msg, m.keys.Error):
		if m.lastError != "" && !m.showErrorToast {
			m.showErrorToast = true
			m.errorToastStart = time.Now()
			return m, scheduleErrorToastTick()
		}
	case key.Matches(msg, m.keys.Help):
		m.showHelp = true
		return m, nil
	case key.Matches(msg, m.keys.Status):
		return m.handleStatusKey()
	case key.Matches(msg, m.keys.Labels):
		return m.handleLabelsKey()
	case key.Matches(msg, m.keys.Edit):
		return m.handleEditKey()
	case key.Matches(msg, m.keys.Comment):
		return m.handleCommentKey()
	case key.Matches(msg, m.keys.NewBead):
		return m.handleNewBeadKey(false)
	case key.Matches(msg, m.keys.NewRootBead):
		return m.handleNewBeadKey(true)
	}

	return m, nil
}

// handleTreeExpand toggles expansion of the current tree node.
func (m *App) handleTreeExpand() (tea.Model, tea.Cmd) {
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
	return m, nil
}

// handleTreeCollapse collapses the current tree node if expanded.
func (m *App) handleTreeCollapse() (tea.Model, tea.Cmd) {
	if len(m.visibleRows) == 0 {
		return m, nil
	}
	row := m.visibleRows[m.cursor]
	if len(row.Node.Children) > 0 && m.isNodeExpandedInView(row) {
		m.collapseNodeForView(row)
		m.recalcVisibleRows()
	}
	return m, nil
}

// handleDeleteKey opens the delete confirmation overlay.
func (m *App) handleDeleteKey() (tea.Model, tea.Cmd) {
	if m.activeOverlay == OverlayNone && !m.searching && len(m.visibleRows) > 0 {
		row := m.visibleRows[m.cursor]
		childInfo, descendantIDs := collectChildInfo(row.Node)
		m.deleteOverlay = NewDeleteOverlay(row.Node.Issue.ID, row.Node.Issue.Title, childInfo, descendantIDs)
		m.activeOverlay = OverlayDelete
	}
	return m, nil
}

// handleBackspaceKey deletes filter chars or opens delete confirmation.
func (m *App) handleBackspaceKey() (tea.Model, tea.Cmd) {
	if !m.ShowDetails && !m.searching && len(m.filterText) > 0 {
		m.setFilterText(m.filterText[:len(m.filterText)-1])
		m.recalcVisibleRows()
		m.updateViewportContent()
		return m, nil
	}
	if m.activeOverlay == OverlayNone && !m.searching && m.filterText == "" && len(m.visibleRows) > 0 {
		row := m.visibleRows[m.cursor]
		childInfo, descendantIDs := collectChildInfo(row.Node)
		m.deleteOverlay = NewDeleteOverlay(row.Node.Issue.ID, row.Node.Issue.Title, childInfo, descendantIDs)
		m.activeOverlay = OverlayDelete
	}
	return m, nil
}

// handleCopyKey copies the current bead ID to clipboard.
func (m *App) handleCopyKey() (tea.Model, tea.Cmd) {
	if len(m.visibleRows) > 0 {
		id := m.visibleRows[m.cursor].Node.Issue.ID
		if err := clipboard.WriteAll(id); err == nil {
			m.copiedBeadID = id
			m.showCopyToast = true
			m.copyToastStart = time.Now()
			return m, scheduleCopyToastTick()
		}
	}
	return m, nil
}

// handleThemeKey cycles the theme forward or backward.
func (m *App) handleThemeKey(forward bool) (tea.Model, tea.Cmd) {
	var newTheme string
	if forward {
		newTheme = theme.CycleTheme()
	} else {
		newTheme = theme.CyclePreviousTheme()
	}
	_ = config.SaveTheme(newTheme)
	m.applyViewportTheme()
	m.themeToastVisible = true
	m.themeToastStart = time.Now()
	m.themeToastName = newTheme
	m.detailIssueID = ""
	m.updateViewportContent()
	return m, scheduleThemeToastTick()
}

// handleStatusKey opens the status overlay.
func (m *App) handleStatusKey() (tea.Model, tea.Cmd) {
	if len(m.visibleRows) > 0 {
		row := m.visibleRows[m.cursor]
		m.statusOverlay = NewStatusOverlay(row.Node.Issue.ID, row.Node.Issue.Title, row.Node.Issue.Status)
		m.activeOverlay = OverlayStatus
	}
	return m, nil
}

// handleLabelsKey opens the labels overlay.
func (m *App) handleLabelsKey() (tea.Model, tea.Cmd) {
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
}

// handleEditKey opens the edit overlay for the current bead.
func (m *App) handleEditKey() (tea.Model, tea.Cmd) {
	if len(m.visibleRows) > 0 {
		row := m.visibleRows[m.cursor]
		parentID := ""
		if row.Parent != nil {
			parentID = row.Parent.Issue.ID
		}
		m.createOverlay = NewEditOverlay(&row.Node.Issue, CreateOverlayOptions{
			DefaultParentID:    parentID,
			AvailableParents:   m.getAvailableParents(),
			AvailableLabels:    m.getAllLabels(),
			AvailableAssignees: m.getAllAssignees(),
			IsRootMode:         parentID == "",
		})
		m.createOverlay.SetSize(m.width, m.height)
		m.activeOverlay = OverlayCreate
		return m, m.createOverlay.Init()
	}
	return m, nil
}

// handleCommentKey opens the comment overlay.
func (m *App) handleCommentKey() (tea.Model, tea.Cmd) {
	if len(m.visibleRows) > 0 {
		row := m.visibleRows[m.cursor]
		m.commentOverlay = NewCommentOverlay(row.Node.Issue.ID, row.Node.Issue.Title)
		m.commentOverlay.SetSize(m.width, m.height)
		m.activeOverlay = OverlayComment
		return m, m.commentOverlay.Init()
	}
	return m, nil
}

// handleNewBeadKey opens the create overlay for a new bead.
func (m *App) handleNewBeadKey(isRoot bool) (tea.Model, tea.Cmd) {
	defaultParent := ""
	// When no beads exist, 'n' should behave like 'N' (create root node)
	if len(m.visibleRows) == 0 {
		isRoot = true
	} else if !isRoot {
		defaultParent = m.visibleRows[m.cursor].Node.Issue.ID
	}
	m.createOverlay = NewCreateOverlay(CreateOverlayOptions{
		DefaultParentID:    defaultParent,
		AvailableParents:   m.getAvailableParents(),
		AvailableLabels:    m.getAllLabels(),
		AvailableAssignees: m.getAllAssignees(),
		IsRootMode:         isRoot,
	})
	m.createOverlay.SetSize(m.width, m.height)
	m.activeOverlay = OverlayCreate
	return m, m.createOverlay.Init()
}
