package ui

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"sync"
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

	// Preserve loaded comments from old nodes to avoid flicker during refresh
	oldCommentState := collectCommentState(m.roots)
	m.roots = newRoots
	transferCommentState(m.roots, oldCommentState)
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
		return func() tea.Msg { return commentBatchLoadedMsg{} }
	}

	client := m.client
	// Prioritize the currently focused issue so the detail pane updates quickly.
	var priorityIDs []string
	if len(m.visibleRows) > 0 && m.cursor >= 0 && m.cursor < len(m.visibleRows) {
		priorityIDs = append(priorityIDs, m.visibleRows[m.cursor].Node.Issue.ID)
	}
	nodes := collectCommentNodes(m.roots, priorityIDs)
	if len(nodes) == 0 {
		return func() tea.Msg { return commentBatchLoadedMsg{} }
	}

	workerLimit := maxConcurrentCommentFetches
	if len(nodes) < workerLimit {
		workerLimit = len(nodes)
	}
	if workerLimit <= 0 {
		workerLimit = 1
	}

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		type result struct {
			id       string
			comments []beads.Comment
			err      error
		}

		results := make([]result, 0, len(nodes))
		var mu sync.Mutex

		jobs := make(chan *graph.Node, workerLimit)
		var wg sync.WaitGroup

		for i := 0; i < workerLimit; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for node := range jobs {
					comments, err := client.Comments(ctx, node.Issue.ID)
					if err == nil && comments == nil {
						comments = []beads.Comment{}
					}
					mu.Lock()
					results = append(results, result{
						id:       node.Issue.ID,
						comments: comments,
						err:      err,
					})
					mu.Unlock()
				}
			}()
		}

		for _, n := range nodes {
			jobs <- n
		}
		close(jobs)
		wg.Wait()

		batch := commentBatchLoadedMsg{
			results: make([]commentLoadedMsg, 0, len(results)),
		}
		for _, r := range results {
			batch.results = append(batch.results, commentLoadedMsg{
				issueID:  r.id,
				comments: r.comments,
				err:      r.err,
			})
		}

		return batch
	}
}

// commentState holds the comment data for a single issue.
type commentState struct {
	comments       []beads.Comment
	commentsLoaded bool
	commentError   string
}

// collectCommentState builds a map of issue ID -> comment state from the tree.
func collectCommentState(roots []*graph.Node) map[string]commentState {
	state := make(map[string]commentState)
	var walk func([]*graph.Node)
	walk = func(nodes []*graph.Node) {
		for _, n := range nodes {
			if n.CommentsLoaded || n.CommentError != "" {
				state[n.Issue.ID] = commentState{
					comments:       n.Issue.Comments,
					commentsLoaded: n.CommentsLoaded,
					commentError:   n.CommentError,
				}
			}
			walk(n.Children)
		}
	}
	walk(roots)
	return state
}

// transferCommentState applies previously collected comment state to new nodes.
func transferCommentState(roots []*graph.Node, state map[string]commentState) {
	var walk func([]*graph.Node)
	walk = func(nodes []*graph.Node) {
		for _, n := range nodes {
			if cs, ok := state[n.Issue.ID]; ok {
				n.Issue.Comments = cs.comments
				n.CommentsLoaded = cs.commentsLoaded
				n.CommentError = cs.commentError
			}
			walk(n.Children)
		}
	}
	walk(roots)
}
