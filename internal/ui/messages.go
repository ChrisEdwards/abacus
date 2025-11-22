package ui

import (
	"time"

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
		interval = defaultRefreshInterval
	}
	return tea.Tick(interval, func(time.Time) tea.Msg { return tickMsg{} })
}
