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
	defaultRefreshInterval = 3 * time.Second
	refreshDisplayDuration = 3 * time.Second // How long delta metrics stay visible in footer
)

// OverlayType represents which overlay is currently active.
type OverlayType int

const (
	OverlayNone OverlayType = iota
	OverlayStatus
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
	refreshInFlight  bool
	lastRefreshTime  time.Time
	spinner          spinner.Model
	outputFormat     string
	version          string

	client beads.Client

	// Error toast state
	lastError       string    // Full error message (separate from stats)
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

	// Status overlay state
	activeOverlay OverlayType
	statusOverlay *StatusOverlay

	// Session tracking for exit summary
	sessionStart time.Time
	initialStats Stats
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
	}
	if dbErr != nil {
		app.lastRefreshStats = fmt.Sprintf("refresh unavailable: %v", dbErr)
	}
	app.recalcVisibleRows()
	// Capture initial stats for session summary
	app.initialStats = app.getStats()
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

