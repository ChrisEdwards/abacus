package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"abacus/internal/beads"
	"abacus/internal/config"
	"abacus/internal/graph"
	"abacus/internal/update"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	minViewportWidth       = 20
	minViewportHeight      = 5
	minTreeWidth           = 18
	minListHeight          = 5
	refreshDisplayDuration = 3 * time.Second // How long delta metrics stay visible in footer
)

// OverlayType represents which overlay is currently active.
type OverlayType int

const (
	OverlayNone OverlayType = iota
	OverlayStatus
	OverlayLabels
	OverlayCreate
	OverlayDelete
	OverlayComment
)

// ViewMode represents the current view filter mode.
type ViewMode int

const (
	ViewModeAll    ViewMode = iota // Show all issues (default)
	ViewModeActive                 // Show non-closed issues (open + in_progress)
	ViewModeReady                  // Show ready issues (open + not blocked)
	viewModeCount                  // internal: number of modes for cycling
)

// String returns the display name of the view mode.
func (v ViewMode) String() string {
	switch v {
	case ViewModeActive:
		return "Active"
	case ViewModeReady:
		return "Ready"
	default:
		return "All"
	}
}

// Next returns the next view mode in the cycle.
func (v ViewMode) Next() ViewMode {
	return ViewMode((int(v) + 1) % int(viewModeCount))
}

// Prev returns the previous view mode in the cycle.
func (v ViewMode) Prev() ViewMode {
	return ViewMode((int(v) + int(viewModeCount) - 1) % int(viewModeCount))
}

// Config configures the UI application.
type Config struct {
	RefreshInterval time.Duration
	AutoRefresh     bool
	DBPathOverride  string
	OutputFormat    string
	StartupReporter StartupReporter
	Client          beads.Client
	Version         string // Version string to display in header
	UpdateChan      <-chan *update.UpdateInfo
}

// errorSource tracks where the last error originated so refresh success can
// decide whether to clear it.
type errorSource int

const (
	errorSourceNone errorSource = iota
	errorSourceRefresh
	errorSourceOperation
)

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
	viewMode   ViewMode // Current view filter mode (All, Active, Ready)
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
	refreshInFlight  bool
	lastRefreshTime  time.Time
	spinner          spinner.Model
	outputFormat     string
	version          string

	client beads.Client

	// Error toast state
	lastError       string // Full error message (separate from stats)
	lastErrorSource errorSource
	errorShownOnce  bool      // True after first toast display
	showErrorToast  bool      // Currently showing toast
	errorToastStart time.Time // When toast was shown (for countdown)

	// Copy toast state
	showCopyToast  bool
	copyToastStart time.Time
	copiedBeadID   string

	// Status toast state
	statusToastVisible   bool
	statusToastStart     time.Time
	statusToastNewStatus string
	statusToastBeadID    string

	// Help overlay state
	showHelp bool
	keys     KeyMap

	// Overlay state
	activeOverlay  OverlayType
	statusOverlay  *StatusOverlay
	labelsOverlay  *LabelsOverlay
	createOverlay  *CreateOverlay
	deleteOverlay  *DeleteOverlay
	commentOverlay *CommentOverlay

	// Labels toast state
	labelsToastVisible bool
	labelsToastStart   time.Time
	labelsToastAdded   []string
	labelsToastRemoved []string
	labelsToastBeadID  string

	// Create toast state
	createToastVisible  bool
	createToastStart    time.Time
	createToastBeadID   string
	createToastTitle    string
	createToastIsUpdate bool

	// New label toast state (shown during create overlay when new label added)
	newLabelToastVisible bool
	newLabelToastStart   time.Time
	newLabelToastLabel   string

	// New assignee toast state (shown during create overlay when new assignee added)
	newAssigneeToastVisible  bool
	newAssigneeToastStart    time.Time
	newAssigneeToastAssignee string

	// Delete toast state
	deleteToastVisible    bool
	deleteToastStart      time.Time
	deleteToastBeadID     string
	deleteToastCascade    bool
	deleteToastChildCount int

	// Theme toast state
	themeToastVisible bool
	themeToastStart   time.Time
	themeToastName    string

	// Comment toast state
	commentToastVisible bool
	commentToastStart   time.Time
	commentToastBeadID  string

	// Columns toast state
	columnsToastVisible bool
	columnsToastStart   time.Time
	columnsToastEnabled bool

	// Update notification state
	updateToastVisible bool
	updateToastStart   time.Time
	updateInfo         *update.UpdateInfo
	updateInProgress   bool
	updateError        string
	updateChan         <-chan *update.UpdateInfo

	// Update success/failure toast state (ab-w1wp)
	updateSuccessToastVisible bool
	updateSuccessToastStart   time.Time
	updateSuccessVersion      string

	updateFailureToastVisible bool
	updateFailureToastStart   time.Time
	updateFailureError        string
	updateFailureCommand      string

	// Session tracking for exit summary
	sessionStart time.Time
	initialStats Stats
}

// errorHotkeyAvailable returns true when the global error hotkey should
// be allowed to toggle the toast (i.e., not while typing into a text input).
func (m *App) errorHotkeyAvailable() bool {
	if m.searching {
		return false
	}
	if m.activeOverlay == OverlayCreate && m.createOverlay != nil && m.createOverlay.IsTextInputActive() {
		return false
	}
	return true
}

// NewApp creates a new UI app instance based on configuration and current working directory.
func NewApp(cfg Config) (*App, error) {
	if cfg.RefreshInterval <= 0 {
		cfg.RefreshInterval = time.Duration(config.GetInt(config.KeyAutoRefreshSeconds)) * time.Second
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
			if latest, err := latestModTimeForDB(dbPath); err == nil {
				dbModTime = latest
			}
		}
	}
	if dbPath == "" && dbErr == nil {
		dbPath, dbModTime, dbErr = findBeadsDB()
	}
	if reporter != nil && dbPath != "" && dbErr == nil {
		reporter.Stage(StartupStageFindingDatabase, fmt.Sprintf("Using database at %s", dbPath))
	}

	// If no database found and no client was injected (e.g., for tests), fail early
	// with a helpful error message telling the user how to set up beads.
	if dbErr != nil && cfg.Client == nil {
		return nil, dbErr
	}

	client := cfg.Client
	if client == nil {
		client = beads.NewSQLiteClient(dbPath)
	}

	roots, err := loadData(context.Background(), client, reporter)
	if err != nil {
		return nil, err
	}
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Prompt = "/"

	s := spinner.New()
	s.Spinner = spinner.Spinner{
		Frames: []string{"◴", "◷", "◶", "◵"},
		FPS:    time.Second / 4,
	}

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
		spinner:         s,
		keys:            DefaultKeyMap(),
		sessionStart:    time.Now(),
		updateChan:      cfg.UpdateChan,
	}
	if dbErr != nil {
		app.lastRefreshStats = fmt.Sprintf("refresh unavailable: %v", dbErr)
	}
	app.recalcVisibleRows()
	// Capture initial stats for session summary
	app.initialStats = app.getStats()
	app.applyViewportTheme()
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
	// Start background comment loading after TUI is displayed (ab-fkyz)
	cmds = append(cmds, scheduleBackgroundCommentLoad())
	// Start waiting for update check result (ab-a4qc)
	if m.updateChan != nil {
		cmds = append(cmds, m.waitForUpdateCheck())
	}
	return tea.Batch(cmds...)
}

func (m *App) applyViewportTheme() {
	m.viewport.Style = baseStyle()
}
