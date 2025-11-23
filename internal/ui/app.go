package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"abacus/internal/beads"
	"abacus/internal/graph"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	minViewportWidth       = 20
	minViewportHeight      = 5
	minTreeWidth           = 18
	minListHeight          = 5
	defaultRefreshInterval = 3 * time.Second
	refreshFlashDuration   = 2 * time.Second
)

// Config configures the UI application.
type Config struct {
	RefreshInterval time.Duration
	AutoRefresh     bool
	DBPathOverride  string
	OutputFormat    string
	Client          beads.Client
	Version         string // Version string to display in header
}

// App implements the Bubble Tea model for Abacus.
type App struct {
	roots       []*graph.Node
	visibleRows []*graph.Node
	cursor      int
	repoName    string

	viewport      viewport.Model
	ShowDetails   bool
	focus         FocusArea
	ready         bool
	detailIssueID string

	textInput  textinput.Model
	searching  bool
	filterText string
	// filterCollapsed tracks nodes explicitly collapsed while a search filter is active.
	filterCollapsed map[string]bool
	// filterForcedExpanded tracks nodes temporarily expanded to surface filter matches.
	filterForcedExpanded map[string]bool
	filterEval           map[string]filterEvaluation

	width            int
	height           int
	err              error
	refreshInterval  time.Duration
	autoRefresh      bool
	dbPath           string
	lastDBModTime    time.Time
	lastRefreshStats string
	showRefreshFlash bool
	refreshInFlight  bool
	lastRefreshTime  time.Time
	outputFormat     string
	version          string

	client beads.Client
}

// NewApp creates a new UI app instance based on configuration and current working directory.
func NewApp(cfg Config) (*App, error) {
	if cfg.RefreshInterval <= 0 {
		cfg.RefreshInterval = defaultRefreshInterval
	}
	client := cfg.Client
	if client == nil {
		client = beads.NewCLIClient()
	}

	var (
		dbPath    string
		dbModTime time.Time
		dbErr     error
	)
	if trimmed := strings.TrimSpace(cfg.DBPathOverride); trimmed != "" {
		info, err := os.Stat(trimmed)
		if err != nil {
			dbErr = fmt.Errorf("db override: %w", err)
		} else if info.IsDir() {
			dbErr = fmt.Errorf("db override must point to a file: %s", trimmed)
		} else {
			dbPath = trimmed
			dbModTime = info.ModTime()
		}
	}
	if dbPath == "" && dbErr == nil {
		dbPath, dbModTime, dbErr = findBeadsDB()
	}
	roots, err := loadData(context.Background(), client)
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Prompt = "/"

	repo := "abacus"
	if wd, err := os.Getwd(); err == nil && wd != "" {
		repo = filepath.Base(wd)
	}

	autoRefresh := cfg.AutoRefresh
	if dbErr != nil {
		autoRefresh = false
	}

	app := &App{
		roots:           roots,
		err:             err,
		textInput:       ti,
		repoName:        repo,
		focus:           FocusTree,
		refreshInterval: cfg.RefreshInterval,
		autoRefresh:     autoRefresh,
		outputFormat:    cfg.OutputFormat,
		version:         cfg.Version,
		client:          client,
	}
	if dbErr == nil {
		app.dbPath = dbPath
		app.lastDBModTime = dbModTime
	} else if autoRefresh {
		app.autoRefresh = false
	}
	if dbErr != nil {
		app.lastRefreshStats = fmt.Sprintf("refresh unavailable: %v", dbErr)
	}
	app.recalcVisibleRows()
	return app, nil
}

func (m *App) Init() tea.Cmd {
	cmds := []tea.Cmd{textinput.Blink}
	if m.autoRefresh && m.refreshInterval > 0 {
		cmds = append(cmds, scheduleTick(m.refreshInterval))
	}
	return tea.Batch(cmds...)
}

func (m *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
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
			m.lastRefreshStats = fmt.Sprintf("refresh failed: %v", msg.err)
			m.showRefreshFlash = true
			m.lastRefreshTime = time.Now()
			return m, nil
		}
		m.applyRefresh(msg.roots, msg.digest, msg.dbModTime)
		return m, nil
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
		if m.searching {
			switch msg.String() {
			case "enter":
				m.searching = false
				m.textInput.Blur()
				return m, nil
			case "esc":
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

		switch msg.String() {
		case "/":
			if !m.searching {
				m.searching = true
				m.textInput.Focus()
				m.textInput.SetValue(m.filterText)
				m.textInput.SetCursor(len(m.filterText))
			}
		case "esc":
			if m.filterText != "" {
				m.clearSearchFilter()
				return m, nil
			}
		case "tab":
			if m.ShowDetails {
				if m.focus == FocusTree {
					m.focus = FocusDetails
				} else {
					m.focus = FocusTree
				}
			}
		case "shift+tab":
			if m.ShowDetails {
				if m.focus == FocusDetails {
					m.focus = FocusTree
				} else {
					m.focus = FocusDetails
				}
			}
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			m.ShowDetails = !m.ShowDetails
			m.focus = FocusTree
			m.updateViewportContent()
		case "r":
			if refreshCmd := m.forceRefresh(); refreshCmd != nil {
				return m, refreshCmd
			}
		case "j", "down":
			m.cursor++
			m.clampCursor()
			m.updateViewportContent()
		case "k", "up":
			m.cursor--
			m.clampCursor()
			m.updateViewportContent()
		case "home", "g":
			m.cursor = 0
			m.updateViewportContent()
		case "end", "G":
			m.cursor = len(m.visibleRows) - 1
			m.updateViewportContent()
		case "pgdown", "ctrl+f":
			m.cursor += clampDimension(m.viewport.Height, 1, len(m.visibleRows))
			m.clampCursor()
			m.updateViewportContent()
		case "pgup", "ctrl+b":
			m.cursor -= clampDimension(m.viewport.Height, 1, len(m.visibleRows))
			m.clampCursor()
			m.updateViewportContent()
		case "space", "right", "l":
			if len(m.visibleRows) == 0 {
				return m, nil
			}
			node := m.visibleRows[m.cursor]
			if len(node.Children) > 0 {
				if m.isNodeExpandedInView(node) {
					m.collapseNodeForView(node)
				} else {
					m.expandNodeForView(node)
				}
				m.recalcVisibleRows()
			}
		case "left", "h":
			if len(m.visibleRows) == 0 {
				return m, nil
			}
			node := m.visibleRows[m.cursor]
			if len(node.Children) > 0 && m.isNodeExpandedInView(node) {
				m.collapseNodeForView(node)
				m.recalcVisibleRows()
			}
		case "backspace":
			if !m.ShowDetails && !m.searching && len(m.filterText) > 0 {
				m.setFilterText(m.filterText[:len(m.filterText)-1])
				m.recalcVisibleRows()
				m.updateViewportContent()
			}
		case "c":
			if m.ShowDetails && m.focus == FocusDetails {
				m.retryCommentsForCurrentNode()
			}
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
func (m *App) clearSearchFilter() {
	prevFilter := m.filterText
	m.searching = false
	m.textInput.Blur()
	m.textInput.Reset()
	if prevFilter == "" {
		return
	}
	m.setFilterText("")
	m.recalcVisibleRows()
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

	switch msg.String() {
	case "home", "g":
		m.viewport.GotoTop()
		return true, nil
	case "end", "G":
		m.viewport.GotoBottom()
		return true, nil
	case "ctrl+f":
		return true, m.viewportPageDownCmd()
	case "ctrl+b":
		return true, m.viewportPageUpCmd()
	}

	if isDetailScrollKey(msg) {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return true, cmd
	}

	return false, nil
}

func isDetailScrollKey(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "j", "k", "down", "up", "pgdown", "pgup", "f", "b", "d", "u", "ctrl+d", "ctrl+u", "left", "right", "h", "l", "space", " ":
		return true
	}
	if msg.Type == tea.KeySpace {
		return true
	}
	return false
}

func (m *App) viewportPageDownCmd() tea.Cmd {
	lines := m.viewport.PageDown()
	if m.viewport.HighPerformanceRendering {
		return viewport.ViewDown(m.viewport, lines)
	}
	return nil
}

func (m *App) viewportPageUpCmd() tea.Cmd {
	lines := m.viewport.PageUp()
	if m.viewport.HighPerformanceRendering {
		return viewport.ViewUp(m.viewport, lines)
	}
	return nil
}

func (m *App) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}
	if !m.ready {
		return "Initializing..."
	}

	stats := m.getStats()
	status := fmt.Sprintf("Beads: %d", stats.Total)

	breakdown := []string{}
	if stats.InProgress > 0 {
		breakdown = append(breakdown, fmt.Sprintf("%d In Progress", stats.InProgress))
	}
	if stats.Ready > 0 {
		breakdown = append(breakdown, fmt.Sprintf("%d Ready", stats.Ready))
	}
	if stats.Blocked > 0 {
		breakdown = append(breakdown, fmt.Sprintf("%d Blocked", stats.Blocked))
	}
	if stats.Closed > 0 {
		breakdown = append(breakdown, fmt.Sprintf("%d Closed", stats.Closed))
	}

	if len(breakdown) > 0 {
		status += " • " + strings.Join(breakdown, " • ")
	}

	if m.filterText != "" {
		filterLabel := fmt.Sprintf("Filter: %s", m.filterText)
		status += " " + styleFilterInfo.Render(filterLabel)
	}

	if m.lastRefreshStats != "" {
		refreshStr := fmt.Sprintf(" Δ %s", m.lastRefreshStats)
		if m.showRefreshFlash && time.Since(m.lastRefreshTime) < refreshFlashDuration {
			refreshStr = styleSelected.Render(refreshStr)
		} else {
			refreshStr = styleStatsDim.Render(refreshStr)
			m.showRefreshFlash = false
		}
		status += " " + refreshStr
	}

	title := "ABACUS"
	if m.version != "" {
		title = fmt.Sprintf("ABACUS v%s", m.version)
	}
	header := styleAppHeader.Render(title) + " " + status
	treeViewStr := m.renderTreeView()

	var mainBody string
	listHeight := clampDimension(m.height-4, minListHeight, m.height-2)
	if m.ShowDetails {
		leftStyle := stylePane
		rightStyle := stylePane
		if m.focus == FocusTree {
			leftStyle = stylePaneFocused
		} else {
			rightStyle = stylePaneFocused
		}

		leftWidth := m.width - m.viewport.Width - 4
		if leftWidth < 1 {
			leftWidth = 1
		}
		rightWidth := m.viewport.Width
		if rightWidth < 1 {
			rightWidth = 1
		}

		left := leftStyle.Width(leftWidth).Height(listHeight).Render(treeViewStr)
		right := rightStyle.Width(rightWidth).Height(listHeight).Render(m.viewport.View())
		mainBody = lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	} else {
		singleWidth := m.width - 2
		if singleWidth < 1 {
			singleWidth = 1
		}
		mainBody = stylePane.Width(singleWidth).Height(listHeight).Render(treeViewStr)
	}

	var bottomBar string
	if m.searching {
		bottomBar = m.textInput.View()
	} else {
		footerStr := " [ / ] Search  [ enter ] Detail  [ tab ] Switch Focus  [ q ] Quit"
		if m.ShowDetails && m.focus == FocusDetails {
			footerStr += "  [ j/k ] Scroll Details"
		} else {
			footerStr += "  [ space ] Expand"
		}
		footerStr += "  [ r ] Refresh"
		bottomBar = lipgloss.NewStyle().Foreground(cLightGray).Render(
			fmt.Sprintf("%s   %s", footerStr,
				lipgloss.PlaceHorizontal(m.width-len(footerStr)-5, lipgloss.Right, "Repo: "+m.repoName)))
	}

	return fmt.Sprintf("%s\n%s\n%s", header, mainBody, bottomBar)
}
