package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

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

	var selectedNodeID, selectedParentID string
	if len(m.visibleRows) > 0 && m.cursor >= 0 && m.cursor < len(m.visibleRows) {
		row := m.visibleRows[m.cursor]
		selectedNodeID = row.Node.Issue.ID
		if row.Parent != nil {
			selectedParentID = row.Parent.Issue.ID
		}
	}

	m.transferFilterExpansionState()

	if selectedNodeID != "" {
		m.expandAncestorsForRow(selectedNodeID, selectedParentID)
	}

	m.setFilterText("")
	m.recalcVisibleRows()

	if selectedNodeID != "" {
		if !m.restoreCursorToRow(selectedNodeID, selectedParentID) {
			m.restoreCursorToID(selectedNodeID)
		}
	}

	m.updateViewportContent()
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
	if key.Matches(msg, m.keys.Up) ||
		key.Matches(msg, m.keys.Down) ||
		key.Matches(msg, m.keys.Left) ||
		key.Matches(msg, m.keys.Right) ||
		key.Matches(msg, m.keys.PageUp) ||
		key.Matches(msg, m.keys.PageDown) ||
		key.Matches(msg, m.keys.Space) {
		return true
	}
	switch msg.String() {
	case "f", "b", "d", "u", "ctrl+d", "ctrl+u":
		return true
	}
	return msg.Type == tea.KeySpace
}

// applyLoadedComment updates a single node's comment state and returns true if the
// currently focused detail pane should be refreshed.
func (m *App) applyLoadedComment(msg commentLoadedMsg) bool {
	node := m.findNodeByID(msg.issueID)
	if node == nil {
		return false
	}
	if msg.err != nil {
		node.CommentError = fmt.Sprintf("failed: %v", msg.err)
		node.CommentsLoaded = false
	} else {
		node.CommentError = ""
		node.Issue.Comments = msg.comments
		node.CommentsLoaded = true
	}

	if len(m.visibleRows) == 0 || m.cursor < 0 || m.cursor >= len(m.visibleRows) {
		return false
	}
	return m.visibleRows[m.cursor].Node.Issue.ID == msg.issueID
}
