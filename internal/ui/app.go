package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
	StartupReporter StartupReporter
	Client          beads.Client
	Version         string // Version string to display in header
}

// App implements the Bubble Tea model for Abacus.
type App struct {
	roots       []*graph.Node
	visibleRows []*graph.Node
	cursor      int
	treeTopLine int
	repoName    string

	viewport      viewport.Model
	ShowDetails   bool
	focus         FocusArea
	ready         bool
	detailIssueID string

	textInput           textinput.Model
	searching           bool
	filterText          string
	filterTokens        []SearchToken
	filterFreeText      []string
	valueCache          map[string][]string
	suggestionFormatter map[string]func(string) string
	overlay             SearchOverlay
	// filterCollapsed tracks nodes explicitly collapsed while a search filter is active.
	filterCollapsed map[string]bool
	// filterForcedExpanded tracks nodes temporarily expanded to surface filter matches.
	filterForcedExpanded map[string]bool
	filterEval           map[string]filterEvaluation

	width            int
	height           int
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

	reporter := cfg.StartupReporter
	if reporter != nil {
		reporter.Stage(StartupStageFindingDatabase, "Finding Beads database...")
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
	if reporter != nil && dbPath != "" && dbErr == nil {
		reporter.Stage(StartupStageFindingDatabase, fmt.Sprintf("Using database at %s", dbPath))
	}

	client := cfg.Client
	if client == nil {
		client = beads.NewCLIClient(beads.WithDatabasePath(dbPath))
	}

	roots, err := loadData(context.Background(), client, reporter)
	if err != nil {
		return nil, err
	}
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Prompt = "/"
	overlay := NewSearchOverlay()

	repo := "abacus"
	if wd, err := os.Getwd(); err == nil && wd != "" {
		repo = filepath.Base(wd)
	}

	autoRefresh := cfg.AutoRefresh
	if dbErr != nil {
		autoRefresh = false
	}

	app := &App{
		roots:      roots,
		textInput:  ti,
		overlay:    overlay,
		valueCache: make(map[string][]string),
		suggestionFormatter: map[string]func(string) string{
			"status":   formatStatusSuggestion,
			"priority": formatPrioritySuggestion,
			"labels":   func(v string) string { return v },
			"label":    func(v string) string { return v },
			"type":     strings.Title,
		},
		repoName:        repo,
		focus:           FocusTree,
		refreshInterval: cfg.RefreshInterval,
		autoRefresh:     autoRefresh,
		outputFormat:    cfg.OutputFormat,
		version:         cfg.Version,
		client:          client,
		dbPath:          dbPath,
		lastDBModTime:   dbModTime,
	}
	if dbErr != nil {
		app.lastRefreshStats = fmt.Sprintf("refresh unavailable: %v", dbErr)
	}
	app.recalcVisibleRows()
	app.buildValueCache()
	if reporter != nil {
		reporter.Stage(StartupStageReady, "Ready!")
	}
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
		m.buildValueCache()
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
				if m.overlay.HasSuggestions() {
					if suggestion := m.overlay.SelectedSuggestion(); suggestion != "" {
						m.applySuggestion(suggestion)
						return m, nil
					}
				}
				m.searching = false
				m.textInput.Blur()
				return m, nil
			case "esc":
				m.clearSearchFilter()
				return m, nil
			case "up", "k":
				if m.overlay.HasSuggestions() {
					m.overlay.CursorUp()
					return m, nil
				}
			case "down", "j", "tab":
				if m.overlay.HasSuggestions() {
					m.overlay.CursorDown()
					return m, nil
				}
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
			m.clampCursor()
			m.updateViewportContent()
		case "end", "G":
			m.cursor = len(m.visibleRows) - 1
			m.clampCursor()
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
	m.filterTokens = nil
	m.filterFreeText = nil
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
	m.overlay.UpdateInput(value)
	m.filterTokens = m.overlay.Tokens()
	m.filterFreeText = m.overlay.FreeTextTerms()
	if m.overlay.SuggestionMode() == SuggestionModeValue {
		field := m.overlay.PendingField()
		if field != "" {
			query := strings.TrimSpace(m.overlay.PendingText())
			m.overlay.SetSuggestions(m.suggestionsForField(field, query))
		}
	} else {
		m.overlay.SetSuggestions(nil)
	}
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

func (m *App) applySuggestion(suggestion string) {
	field := m.overlay.PendingField()
	if field == "" {
		return
	}
	value := suggestion
	if entry, ok := m.overlay.SelectedEntry(); ok && entry.Value != "" {
		value = entry.Value
	}
	raw := strings.TrimSpace(m.textInput.Value())
	base := raw
	if pending := strings.TrimSpace(m.overlay.PendingText()); pending != "" {
		base = strings.TrimSuffix(raw, pending)
	}
	base = strings.TrimSpace(base)
	replacement := fmt.Sprintf("%s:%s", field, value)
	if base != "" {
		base += " "
	}
	m.textInput.SetValue(strings.TrimSpace(base+replacement) + " ")
	m.textInput.SetCursor(len(m.textInput.Value()))
	m.setFilterText(m.textInput.Value())
}

func (m *App) buildValueCache() {
	if m == nil {
		return
	}
	sets := map[string]map[string]struct{}{
		"status":   {},
		"priority": {},
		"type":     {},
		"labels":   {},
	}
	var walk func(nodes []*graph.Node)
	walk = func(nodes []*graph.Node) {
		for _, node := range nodes {
			issue := node.Issue
			addValueToSet(sets["status"], strings.ToLower(strings.TrimSpace(issue.Status)))
			addValueToSet(sets["priority"], fmt.Sprintf("%d", issue.Priority))
			addValueToSet(sets["type"], strings.ToLower(strings.TrimSpace(issue.IssueType)))
			for _, label := range issue.Labels {
				addValueToSet(sets["labels"], strings.ToLower(strings.TrimSpace(label)))
			}
			walk(node.Children)
		}
	}
	walk(m.roots)
	m.valueCache = make(map[string][]string, len(sets)+1)
	for field, set := range sets {
		m.valueCache[field] = sortedValues(set)
	}
	if labels := m.valueCache["labels"]; len(labels) > 0 {
		m.valueCache["label"] = labels
	}
}

func addValueToSet(set map[string]struct{}, value string) {
	if set == nil {
		return
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	set[value] = struct{}{}
}

func sortedValues(set map[string]struct{}) []string {
	if len(set) == 0 {
		return nil
	}
	values := make([]string, 0, len(set))
	for v := range set {
		values = append(values, v)
	}
	sort.Strings(values)
	return values
}

func (m *App) suggestionsForField(field, query string) []suggestionEntry {
	field = strings.ToLower(strings.TrimSpace(field))
	if field == "" {
		return nil
	}
	values := m.valueCache[field]
	if len(values) == 0 {
		return nil
	}
	formatter := m.suggestionFormatter[field]
	entries := make([]suggestionEntry, 0, len(values))
	for _, v := range values {
		display := v
		if formatter != nil {
			display = formatter(v)
		}
		entries = append(entries, suggestionEntry{Display: display, Value: v})
	}
	return rankSuggestions(entries, query)
}

func formatStatusSuggestion(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "open":
		return "🟢 Open"
	case "in_progress", "in-progress":
		return "🟡 In Progress"
	case "blocked":
		return "⛔ Blocked"
	case "closed":
		return "✔ Closed"
	default:
		return titleCase(value)
	}
}

func formatPrioritySuggestion(value string) string {
	value = strings.TrimSpace(value)
	if p, err := strconv.Atoi(value); err == nil {
		switch p {
		case 0:
			return "P0 (Critical)"
		case 1:
			return "P1 (High)"
		case 2:
			return "P2 (Medium)"
		case 3:
			return "P3 (Low)"
		default:
			return fmt.Sprintf("P%d", p)
		}
	}
	return strings.ToUpper(value)
}

func titleCase(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	runes := []rune(value)
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
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
		_ = m.viewport.PageDown()
		return true, nil
	case "ctrl+b":
		_ = m.viewport.PageUp()
		return true, nil
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
	return msg.Type == tea.KeySpace
}

func (m *App) View() string {
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

	footerStr := " [ / ] Search  [ enter ] Detail  [ tab ] Switch Focus  [ q ] Quit"
	if m.ShowDetails && m.focus == FocusDetails {
		footerStr += "  [ j/k ] Scroll Details"
	} else {
		footerStr += "  [ space ] Expand"
	}
	footerStr += "  [ r ] Refresh"
	bottomBar := lipgloss.NewStyle().Foreground(cLightGray).Render(
		fmt.Sprintf("%s   %s", footerStr,
			lipgloss.PlaceHorizontal(m.width-len(footerStr)-5, lipgloss.Right, "Repo: "+m.repoName)))

	view := fmt.Sprintf("%s\n%s\n%s", header, mainBody, bottomBar)
	if m.searching {
		inputWidth := m.overlay.InputWidth(m.width)
		m.textInput.Width = inputWidth
		overlayInput := m.textInput.View()
		containerWidth := m.width
		if containerWidth < overlayMinWidth {
			containerWidth = overlayMinWidth
		}
		containerHeight := m.height
		if containerHeight < listHeight+4 {
			containerHeight = listHeight + 4
		}
		return m.overlay.View(overlayInput, containerWidth, containerHeight)
	}
	return view
}
