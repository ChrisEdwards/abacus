package main

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"abacus/internal/ui"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Colors - a nice purple/magenta theme
var (
	primaryColor   = lipgloss.Color("#7D56F4")
	secondaryColor = lipgloss.Color("#FF79C6")
	dimColor       = lipgloss.Color("#6272A4")
	textColor      = lipgloss.Color("#F8F8F2")
	successColor   = lipgloss.Color("#50FA7B")
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(dimColor).
			Italic(true)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	statusStyle = lipgloss.NewStyle().
			Foreground(textColor)

	progressStyle = lipgloss.NewStyle().
			Foreground(primaryColor)

	countStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	containerStyle = lipgloss.NewStyle().
			Padding(1, 2)
)

// ASCII art logo - compact but distinctive
const logo = `
   ▄▀▄ █▄▄ ▄▀▄ █▀▀ █ █ █▀
   █▀█ █▄█ █▀█ █▄▄ █▄█ ▄█
`

// startupModel is the bubbletea model for the startup screen
type startupModel struct {
	spinner  spinner.Model
	progress progress.Model

	stage      string
	detail     string
	percent    float64
	current    int
	total      int
	isProgress bool // true when showing progress bar instead of spinner

	width  int
	height int
	ready  bool
	done   bool

	// Channel to receive updates from main goroutine
	updates chan startupUpdate
}

type startupUpdate struct {
	stage      string
	detail     string
	percent    float64
	current    int
	total      int
	isProgress bool
	done       bool
}

type updateMsg startupUpdate
type tickMsg struct{}

func newStartupModel() *startupModel {
	// Create spinner with a nice style
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = spinnerStyle

	// Create progress bar with gradient
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	return &startupModel{
		spinner:  s,
		progress: p,
		stage:    "Starting",
		updates:  make(chan startupUpdate, 16),
	}
}

func (m *startupModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.waitForUpdate(),
	)
}

func (m *startupModel) waitForUpdate() tea.Cmd {
	return func() tea.Msg {
		update := <-m.updates
		return updateMsg(update)
	}
}

func (m *startupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case updateMsg:
		m.stage = msg.stage
		m.detail = msg.detail
		m.percent = msg.percent
		m.current = msg.current
		m.total = msg.total
		m.isProgress = msg.isProgress

		if msg.done {
			m.done = true
			return m, tea.Quit
		}

		var cmds []tea.Cmd
		if m.isProgress {
			cmds = append(cmds, m.progress.SetPercent(m.percent))
		}
		cmds = append(cmds, m.waitForUpdate())
		return m, tea.Batch(cmds...)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case tea.KeyMsg:
		// Allow quitting with ctrl+c
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *startupModel) View() string {
	if !m.ready {
		return ""
	}

	var b strings.Builder

	// Logo
	b.WriteString(titleStyle.Render(logo))
	b.WriteString("\n")

	// Spinner or Progress bar
	if m.isProgress && m.total > 0 {
		// Progress bar mode
		b.WriteString(m.progress.View())
		b.WriteString("\n")

		// Count with label (e.g., "Loading issues 45 / 150" or "Loading comments 12 / 50")
		label := m.detail
		if label == "" {
			label = "Loading"
		}
		countText := fmt.Sprintf("%s %d / %d", label, m.current, m.total)
		b.WriteString(countStyle.Render(countText))
	} else {
		// Spinner mode
		b.WriteString(m.spinner.View())
		b.WriteString(" ")
		b.WriteString(statusStyle.Render(m.detail))
	}

	return containerStyle.Render(b.String())
}

// Send updates to the model
func (m *startupModel) sendUpdate(update startupUpdate) {
	select {
	case m.updates <- update:
	default:
		// Drop if channel is full
	}
}

// StartupDisplay wraps the bubbletea program for the startup animation
type StartupDisplay struct {
	program *tea.Program
	model   *startupModel
	done    chan struct{}
	mu      sync.Mutex
	stopped bool
}

// NewStartupDisplay creates a new startup display
func NewStartupDisplay(w io.Writer) *StartupDisplay {
	model := newStartupModel()

	// Use inline mode so it doesn't take over the whole screen
	program := tea.NewProgram(
		model,
		tea.WithOutput(w),
		tea.WithoutSignalHandler(), // We handle signals ourselves
	)

	d := &StartupDisplay{
		program: program,
		model:   model,
		done:    make(chan struct{}),
	}

	// Start the program in a goroutine
	go func() {
		_, _ = program.Run()
		close(d.done)
	}()

	// Give the program a moment to start
	time.Sleep(10 * time.Millisecond)

	return d
}

// Stage implements ui.StartupReporter
func (d *StartupDisplay) Stage(stage ui.StartupStage, detail string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.stopped {
		return
	}

	update := startupUpdate{
		stage:  stageToString(stage),
		detail: detail,
	}

	// Parse progress info if present
	if strings.Contains(detail, "/") {
		var current, total int
		// Try parsing issues progress
		if _, err := fmt.Sscanf(detail, "Loading issues... %d/%d", &current, &total); err == nil && total > 0 {
			update.isProgress = true
			update.current = current
			update.total = total
			update.percent = float64(current) / float64(total)
			update.detail = "Loading issues"
		}
		// Try parsing comments progress
		if _, err := fmt.Sscanf(detail, "Loading comments... %d/%d", &current, &total); err == nil && total > 0 {
			update.isProgress = true
			update.current = current
			update.total = total
			update.percent = float64(current) / float64(total)
			update.detail = "Loading comments"
		}
	}

	d.model.sendUpdate(update)
}

// Stop stops the startup display
func (d *StartupDisplay) Stop() {
	d.mu.Lock()
	if d.stopped {
		d.mu.Unlock()
		return
	}
	d.stopped = true
	d.mu.Unlock()

	// Send done signal
	d.model.sendUpdate(startupUpdate{done: true})

	// Wait for program to finish
	select {
	case <-d.done:
	case <-time.After(500 * time.Millisecond):
		d.program.Kill()
	}

	// Clear the line
	fmt.Print("\r\033[K")
}

func stageToString(stage ui.StartupStage) string {
	switch stage {
	case ui.StartupStageInit:
		return "Initializing"
	case ui.StartupStageVersionCheck:
		return "Checking CLI"
	case ui.StartupStageFindingDatabase:
		return "Finding database"
	case ui.StartupStageLoadingIssues:
		return "Loading issues"
	case ui.StartupStageBuildingGraph:
		return "Building graph"
	case ui.StartupStageOrganizingTree:
		return "Organizing"
	case ui.StartupStageReady:
		return "Ready"
	default:
		return "Loading"
	}
}
