package ui

import (
	"context"
	"fmt"
	"os"
	"time"

	"abacus/internal/beads"
	"abacus/internal/graph"

	tea "github.com/charmbracelet/bubbletea"
)

const refreshTimeout = 10 * time.Second

func refreshDataCmd(client beads.Client, targetModTime time.Time) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), refreshTimeout)
		defer cancel()

		issues, err := client.Export(ctx)
		if err != nil {
			return refreshCompleteMsg{err: err}
		}

		roots, err := graph.NewBuilder().Build(issues)
		if err != nil {
			return refreshCompleteMsg{err: err}
		}

		return refreshCompleteMsg{
			roots:     roots,
			digest:    buildIssueDigest(roots),
			dbModTime: targetModTime,
		}
	}
}

const maxRefreshFailures = 10 // Disable after 10 consecutive failures (~30 seconds)

func (m *App) checkDBForChanges() tea.Cmd {
	if m.refreshInFlight || m.dbPath == "" {
		return nil
	}
	info, err := os.Stat(m.dbPath)
	if err != nil {
		m.refreshFailures++
		// Only disable after many consecutive failures
		if m.refreshFailures >= maxRefreshFailures {
			m.autoRefresh = false
			m.lastRefreshStats = fmt.Sprintf("refresh disabled: %v (after %d failures)", err, m.refreshFailures)
		} else {
			m.lastRefreshStats = fmt.Sprintf("refresh error (%d/%d): %v", m.refreshFailures, maxRefreshFailures, err)
		}
		m.lastRefreshTime = time.Now()
		return nil
	}
	// Reset failure count on success
	m.refreshFailures = 0
	if !info.ModTime().After(m.lastDBModTime) {
		return nil
	}
	return m.startRefresh(info.ModTime())
}

func (m *App) startRefresh(targetModTime time.Time) tea.Cmd {
	if m.refreshInFlight {
		return nil
	}
	m.refreshInFlight = true
	return tea.Batch(m.spinner.Tick, refreshDataCmd(m.client, targetModTime))
}

func (m *App) forceRefresh() tea.Cmd {
	var modTime time.Time
	if m.dbPath != "" {
		if info, err := os.Stat(m.dbPath); err == nil {
			modTime = info.ModTime()
		}
	}
	return m.startRefresh(modTime)
}

func (m *App) applyRefresh(newRoots []*graph.Node, newDigest map[string]string, newModTime time.Time) {
	state := m.captureState()
	oldDigest := buildIssueDigest(m.roots)

	m.roots = newRoots
	if !newModTime.IsZero() {
		m.lastDBModTime = newModTime
	}

	m.restoreExpandedState(state.expandedIDs)
	m.expandedInstances = copyBoolMapAll(state.expandedInstances)
	m.setFilterText(state.filterText)
	m.filterCollapsed = copyBoolMap(state.filterCollapsed)
	m.filterForcedExpanded = copyBoolMap(state.filterForcedExpanded)
	m.textInput.SetValue(state.filterText)
	m.recalcVisibleRows()

	if state.currentID != "" {
		m.restoreCursorToID(state.currentID)
	} else {
		m.cursor = state.cursorIndex
		m.clampCursor()
	}

	if state.currentID != "" {
		m.detailIssueID = state.currentID
	} else {
		m.detailIssueID = ""
	}

	if m.ShowDetails {
		m.focus = state.focus
	} else {
		m.focus = FocusTree
	}

	if m.ShowDetails {
		m.viewport.YOffset = state.viewportYOffset
	}
	m.updateViewportContent()

	m.lastRefreshStats = computeDiffStats(oldDigest, newDigest)
	m.lastRefreshTime = time.Now()
}

// eventualRefreshMsg is sent after a delay to trigger a background consistency refresh.
type eventualRefreshMsg struct{}

// scheduleEventualRefresh schedules a delayed consistency refresh after fast injection.
// This ensures the tree stays consistent with the database without blocking the UI.
func (m *App) scheduleEventualRefresh() tea.Cmd {
	// Wait 2 seconds before triggering consistency refresh
	// This gives user time to interact with the new node
	return tea.Tick(2*time.Second, func(_ time.Time) tea.Msg {
		return eventualRefreshMsg{}
	})
}
