package ui

import (
	"context"
	"sort"
	"time"

	"abacus/internal/graph"

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
		m.activeOverlay = OverlayNone
		m.createOverlay = nil
		m.displayCreateToast(msg.Title)
		return m, tea.Batch(m.executeCreateBead(msg), scheduleCreateToastTick())
	case CreateCancelledMsg:
		m.activeOverlay = OverlayNone
		m.createOverlay = nil
		return m, nil
	case createCompleteMsg:
		if msg.err != nil {
			m.lastError = msg.err.Error()
			m.showErrorToast = true
			m.errorToastStart = time.Now()
			return m, scheduleErrorToastTick()
		}
		// Store the new bead ID for toast display
		m.createToastBeadID = msg.id
		return m, m.forceRefresh()
	case createToastTickMsg:
		if !m.createToastVisible {
			return m, nil
		}
		if time.Since(m.createToastStart) >= 7*time.Second {
			m.createToastVisible = false
			return m, nil
		}
		return m, scheduleCreateToastTick()
	}

	// Delegate to status overlay if active
	if m.activeOverlay == OverlayStatus && m.statusOverlay != nil {
		var cmd tea.Cmd
		m.statusOverlay, cmd = m.statusOverlay.Update(msg)
		return m, cmd
	}

	// Delegate to labels overlay if active
	if m.activeOverlay == OverlayLabels && m.labelsOverlay != nil {
		var cmd tea.Cmd
		m.labelsOverlay, cmd = m.labelsOverlay.Update(msg)
		return m, cmd
	}

	// Delegate to create overlay if active
	if m.activeOverlay == OverlayCreate && m.createOverlay != nil {
		var cmd tea.Cmd
		m.createOverlay, cmd = m.createOverlay.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.refreshInFlight {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	case tickMsg:
		if !m.autoRefresh || m.refreshInterval <= 0 {
			return m, nil
		}
		cmds := []tea.Cmd{}
		if refreshCmd := m.checkDBForChanges(); refreshCmd != nil {
			cmds = append(cmds, refreshCmd)
		}
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
		case key.Matches(msg, m.keys.Backspace):
			if !m.ShowDetails && !m.searching && len(m.filterText) > 0 {
				m.setFilterText(m.filterText[:len(m.filterText)-1])
				m.recalcVisibleRows()
				m.updateViewportContent()
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
					row.Node.Issue.Labels,
					allLabels,
				)
				m.activeOverlay = OverlayLabels
			}
			return m, nil
		case key.Matches(msg, m.keys.NewBead):
			defaultParent := ""
			if len(m.visibleRows) > 0 {
				defaultParent = m.visibleRows[m.cursor].Node.Issue.ID
			}
			parents := m.getAvailableParents()
			m.createOverlay = NewCreateOverlay(defaultParent, parents)
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
	id  string
	err error
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
		newID, err := m.client.Create(ctx, msg.Title, msg.IssueType, msg.Priority)
		if err != nil {
			return createCompleteMsg{err: err}
		}
		// If parent specified, add parent-child dependency
		if msg.ParentID != "" {
			if err := m.client.AddDependency(ctx, newID, msg.ParentID, "parent-child"); err != nil {
				return createCompleteMsg{id: newID, err: err}
			}
		}
		return createCompleteMsg{id: newID}
	}
}

// displayCreateToast displays a success toast for bead creation.
func (m *App) displayCreateToast(title string) {
	m.createToastTitle = title
	m.createToastVisible = true
	m.createToastStart = time.Now()
}
