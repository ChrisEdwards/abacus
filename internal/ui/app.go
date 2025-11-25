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
	"github.com/charmbracelet/x/ansi"
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
	visibleRows []graph.TreeRow
	cursor      int
	treeTopLine int
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
	// expandedInstances tracks expanded state per TreeRow instance for multi-parent nodes.
	// Key format: "parentID:nodeID" where parentID is empty for root nodes.
	expandedInstances map[string]bool

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

	// Error toast state
	lastError       string    // Full error message (separate from stats)
	errorShownOnce  bool      // True after first toast display
	showErrorToast  bool      // Currently showing toast
	errorToastStart time.Time // When toast was shown (for countdown)
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
		textInput:       ti,
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
			// ESC dismisses error toast first, then clears search filter
			if m.showErrorToast {
				m.showErrorToast = false
				return m, nil
			}
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
			row := m.visibleRows[m.cursor]
			if len(row.Node.Children) > 0 {
				if m.isNodeExpandedInView(row) {
					m.collapseNodeForView(row)
				} else {
					m.expandNodeForView(row)
				}
				m.recalcVisibleRows()
			}
		case "left", "h":
			if len(m.visibleRows) == 0 {
				return m, nil
			}
			row := m.visibleRows[m.cursor]
			if len(row.Node.Children) > 0 && m.isNodeExpandedInView(row) {
				m.collapseNodeForView(row)
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
		case "e":
			// Show error toast if there's an error and toast isn't already visible
			if m.lastError != "" && !m.showErrorToast {
				m.showErrorToast = true
				m.errorToastStart = time.Now()
				return m, scheduleErrorToastTick()
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

	// Build header with right-aligned error indicator if present
	leftContent := styleAppHeader.Render(title) + " " + status
	var header string
	if m.lastError != "" {
		rightContent := styleErrorIndicator.Render("⚠ Refresh error (e)")
		availableWidth := m.width - lipgloss.Width(leftContent) - lipgloss.Width(rightContent) - 2
		if availableWidth > 0 {
			header = leftContent + strings.Repeat(" ", availableWidth) + rightContent
		} else {
			header = leftContent + " " + rightContent
		}
	} else {
		header = leftContent
	}
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

	// Overlay error toast on mainBody if visible
	if toast := m.renderErrorToast(); toast != "" {
		// Measure actual rendered width for proper right-alignment
		containerWidth := lipgloss.Width(mainBody)
		mainBody = overlayBottomRight(mainBody, toast, containerWidth, 1)
	}

	return fmt.Sprintf("%s\n%s\n%s", header, mainBody, bottomBar)
}

// renderErrorToast renders the error toast content if visible.
func (m *App) renderErrorToast() string {
	if !m.showErrorToast || m.lastError == "" {
		return ""
	}
	elapsed := time.Since(m.errorToastStart)
	remaining := 10 - int(elapsed.Seconds())
	if remaining < 0 {
		remaining = 0
	}

	// Extract a short, user-friendly error message
	errMsg := extractShortError(m.lastError, 80)

	// Build content: title + bd error message + countdown right-aligned
	titleLine := "⚠ Refresh Error"
	bdErrLine := fmt.Sprintf("bd: %s", errMsg)
	countdownStr := fmt.Sprintf("[%ds]", remaining)

	// Calculate toast width based on longest line
	toastWidth := 50
	if w := lipgloss.Width(titleLine); w > toastWidth {
		toastWidth = w
	}
	if w := lipgloss.Width(bdErrLine); w > toastWidth {
		toastWidth = w
	}

	padding := toastWidth - len(countdownStr)
	if padding < 0 {
		padding = 0
	}
	content := fmt.Sprintf("%s\n%s\n%s%s", titleLine, bdErrLine, strings.Repeat(" ", padding), countdownStr)

	return styleErrorToast.Render(content)
}

// extractShortError extracts a short, user-friendly error message.
func extractShortError(fullError string, maxLen int) string {
	msg := fullError

	// Look for "Error:" pattern and extract from there
	if idx := strings.Index(msg, "Error:"); idx >= 0 {
		msg = strings.TrimSpace(msg[idx+6:]) // Skip "Error:"
	} else if idx := strings.Index(msg, "error:"); idx >= 0 {
		msg = strings.TrimSpace(msg[idx+6:])
	}

	// Take only the first line/sentence
	if idx := strings.Index(msg, "\n"); idx >= 0 {
		msg = msg[:idx]
	}
	// Also truncate at period if it makes sense
	if idx := strings.Index(msg, ". "); idx >= 0 && idx < maxLen {
		msg = msg[:idx]
	}

	// Remove any "Run 'bd..." suggestions
	if idx := strings.Index(msg, " Run '"); idx >= 0 {
		msg = msg[:idx]
	}
	if idx := strings.Index(msg, " Run \""); idx >= 0 {
		msg = msg[:idx]
	}

	msg = strings.TrimSpace(msg)

	// Truncate if still too long
	if len(msg) > maxLen {
		msg = msg[:maxLen-3] + "..."
	}

	return msg
}

// overlayBottomRight positions the overlay at bottom-right of the background.
// containerWidth specifies the known width of the container for proper right-alignment.
func overlayBottomRight(background, overlay string, containerWidth, padding int) string {
	if overlay == "" {
		return background
	}

	bgLines := strings.Split(background, "\n")
	overlayLines := strings.Split(overlay, "\n")

	bgHeight := len(bgLines)
	overlayHeight := len(overlayLines)
	overlayWidth := lipgloss.Width(overlay)

	// Calculate position: bottom-right with padding
	startRow := bgHeight - overlayHeight - padding
	if startRow < 0 {
		startRow = 0
	}

	// Insert column: account for border (1 char) plus padding inside
	borderWidth := 1
	insertCol := containerWidth - overlayWidth - padding - borderWidth
	if insertCol < 0 {
		insertCol = 0
	}

	// For each overlay line, merge it with background
	overlayLineWidth := 0
	if len(overlayLines) > 0 {
		overlayLineWidth = lipgloss.Width(overlayLines[0])
	}

	for i, overlayLine := range overlayLines {
		bgRow := startRow + i
		if bgRow >= bgHeight {
			break
		}

		bgLine := bgLines[bgRow]
		bgLineWidth := lipgloss.Width(bgLine)

		// Pad background line to reach insert position
		for lipgloss.Width(bgLine) < insertCol {
			bgLine += " "
		}

		// Build: left part + overlay + right part (preserves border)
		leftPart := truncateVisualWidth(bgLine, insertCol)
		rightStart := insertCol + overlayLineWidth
		rightPart := ""
		if rightStart < bgLineWidth {
			// Extract the right portion after the overlay ends
			rightPart = ansi.Cut(bgLines[bgRow], rightStart, bgLineWidth)
		}
		bgLines[bgRow] = leftPart + overlayLine + rightPart
	}

	return strings.Join(bgLines, "\n")
}

// truncateVisualWidth truncates a string to the specified visual width,
// properly handling ANSI escape sequences.
func truncateVisualWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	// Use proper ANSI-aware truncation
	return ansi.Truncate(s, width, "")
}
