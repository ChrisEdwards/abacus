package ui

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
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

func (m *App) checkDBForChanges() tea.Cmd {
	if m.refreshInFlight || m.dbPath == "" {
		return nil
	}

	modTime, err := m.latestDBModTime()
	if err != nil {
		m.lastError = fmt.Sprintf("refresh check failed: %v", err)
		m.lastErrorSource = errorSourceRefresh
		m.lastRefreshStats = "refresh error"
		return nil // Try again next tick
	}

	if !modTime.After(m.lastDBModTime) {
		return nil
	}

	return m.startRefresh(modTime)
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
		if latest, err := m.latestDBModTime(); err == nil {
			modTime = latest
		}
	}
	return m.startRefresh(modTime)
}

func (m *App) latestDBModTime() (time.Time, error) {
	if strings.TrimSpace(m.dbPath) == "" {
		return time.Time{}, fmt.Errorf("database path is empty")
	}
	return latestModTimeForDB(m.dbPath)
}

func latestModTimeForDB(dbPath string) (time.Time, error) {
	info, err := os.Stat(dbPath)
	if err != nil {
		return time.Time{}, err
	}
	latest := info.ModTime()
	for _, path := range []string{dbPath + "-wal", dbPath + "-shm"} {
		if modTime, err := optionalModTime(path); err != nil {
			return time.Time{}, err
		} else if modTime.After(latest) {
			latest = modTime
		}
	}
	return latest, nil
}

func optionalModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	return info.ModTime(), nil
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
	m.viewMode = state.viewMode // Restore view mode across refresh
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

// loadCommentsInBackground loads comments for all issues without blocking the UI (ab-fkyz).
// This is called after the TUI is displayed to avoid startup delay.
func (m *App) loadCommentsInBackground() tea.Cmd {
	if m.client == nil || len(m.roots) == 0 {
		return func() tea.Msg { return backgroundCommentLoadCompleteMsg{} }
	}

	client := m.client
	roots := m.roots

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		preloadAllComments(ctx, client, roots, nil)
		return backgroundCommentLoadCompleteMsg{}
	}
}
