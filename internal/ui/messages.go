package ui

import (
	"time"

	"abacus/internal/beads"
	"abacus/internal/config"
	"abacus/internal/graph"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg struct{}

type refreshCompleteMsg struct {
	roots     []*graph.Node
	digest    map[string]string
	dbModTime time.Time
	err       error
}

func scheduleTick(interval time.Duration) tea.Cmd {
	if interval <= 0 {
		interval = time.Duration(config.GetInt(config.KeyAutoRefreshSeconds)) * time.Second
	}
	return tea.Tick(interval, func(time.Time) tea.Msg { return tickMsg{} })
}

type errorToastTickMsg struct{}

func scheduleErrorToastTick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return errorToastTickMsg{}
	})
}

type copyToastTickMsg struct{}

func scheduleCopyToastTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return copyToastTickMsg{}
	})
}

type newLabelToastTickMsg struct{}

func scheduleNewLabelToastTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return newLabelToastTickMsg{}
	})
}

type newAssigneeToastTickMsg struct{}

func scheduleNewAssigneeToastTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return newAssigneeToastTickMsg{}
	})
}

type themeToastTickMsg struct{}

func scheduleThemeToastTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return themeToastTickMsg{}
	})
}

type columnsToastTickMsg struct{}

func scheduleColumnsToastTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return columnsToastTickMsg{}
	})
}

// Background comment loading messages (ab-fkyz)
type startBackgroundCommentLoadMsg struct{}

type commentLoadedMsg struct {
	issueID  string
	comments []beads.Comment
	err      error
}

type commentBatchLoadedMsg struct {
	results []commentLoadedMsg
}

type backgroundCommentLoadCompleteMsg struct{}

func scheduleBackgroundCommentLoad() tea.Cmd {
	// Small delay to ensure TUI is fully rendered before starting background work
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return startBackgroundCommentLoadMsg{}
	})
}
