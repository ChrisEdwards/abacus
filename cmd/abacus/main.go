package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"abacus/internal/beads"
	"abacus/internal/config"
	"abacus/internal/debug"
	"abacus/internal/ui"
	"abacus/internal/ui/theme"
	"abacus/internal/update"

	tea "github.com/charmbracelet/bubbletea"
)

const versionCheckTimeout = 10 * time.Second

func main() {
	if err := config.Initialize(); err != nil {
		fmt.Printf("Error initializing config: %v\n", err)
		os.Exit(1)
	}

	// Load theme from config (silently ignore if theme doesn't exist)
	if themeName := config.GetString(config.KeyTheme); themeName != "" {
		theme.SetTheme(themeName)
	}

	autoRefreshSecondsDefault := config.GetInt(config.KeyAutoRefreshSeconds)
	if autoRefreshSecondsDefault < 0 {
		autoRefreshSecondsDefault = 0
	}
	dbPathDefault := config.GetString(config.KeyDatabasePath)
	outputFormatDefault := config.GetString(config.KeyOutputFormat)
	skipVersionCheckDefault := config.GetBool(config.KeySkipVersionCheck)
	skipUpdateCheckDefault := config.GetBool(config.KeySkipUpdateCheck)

	versionFlag := flag.Bool("version", false, "Print version information and exit")
	autoRefreshSecondsFlag := flag.Int("auto-refresh-seconds", autoRefreshSecondsDefault, "Auto-refresh interval in seconds (0 disables auto refresh)")
	dbPathFlag := flag.String("db-path", dbPathDefault, "Path to the Beads database file")
	outputFormatFlag := flag.String("output-format", outputFormatDefault, "Detail panel markdown style (rich, light, plain)")
	skipVersionCheckFlag := flag.Bool("skip-version-check", skipVersionCheckDefault, "Skip Beads CLI version validation (or set AB_SKIP_VERSION_CHECK=true)")
	skipUpdateCheckFlag := flag.Bool("skip-update-check", skipUpdateCheckDefault, "Skip checking for updates at startup (or set AB_SKIP_UPDATE_CHECK=true)")
	debugFlag := flag.Bool("debug", config.GetBool(config.KeyDebug), "Enable debug logging to ~/.abacus/debug.log")
	backendFlag := flag.String("backend", "", "Force backend (bd or br) - overrides auto-detection, one-time only")
	flag.Parse()

	if *versionFlag {
		printVersion()
		os.Exit(0)
	}

	// Initialize debug logging (must be after flag parsing)
	if err := debug.Init(*debugFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize debug logging: %v\n", err)
	}
	defer debug.Close()

	visited := map[string]struct{}{}
	flag.CommandLine.Visit(func(f *flag.Flag) {
		visited[f.Name] = struct{}{}
	})

	runtime := computeRuntimeOptions(runtimeFlags{
		autoRefreshSeconds: autoRefreshSecondsFlag,
		dbPath:             dbPathFlag,
		outputFormat:       outputFormatFlag,
		skipVersionCheck:   skipVersionCheckFlag,
		skipUpdateCheck:    skipUpdateCheckFlag,
		backend:            backendFlag,
	}, visited)

	skipVersionCheck := runtime.skipVersionCheck

	// Start the startup display immediately - don't let users stare at nothing
	startup := NewStartupDisplay(os.Stderr)
	startup.Stage(ui.StartupStageInit, "Starting up...")

	// Backend detection (includes version check internally)
	// This determines which backend (bd or br) to use for this project.
	// DetectBackend validates versions, so we skip the old version check if detection succeeds.
	var detectedBackend string
	if !skipVersionCheck {
		if startup != nil {
			startup.Stage(ui.StartupStageVersionCheck, "Detecting backend...")
			// Set hook to stop spinner before any user prompts
			beads.BeforePromptHook = func() {
				startup.Stop()
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), versionCheckTimeout)
		var err error
		detectedBackend, err = beads.DetectBackend(ctx, runtime.backend)
		cancel()
		// Clear the hook after detection
		beads.BeforePromptHook = nil
		if err != nil {
			if startup != nil {
				startup.Stop()
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Show bd version warning if using bd (one-time warning for versions > MaxSupportedBdVersion)
		if detectedBackend == beads.BackendBd {
			ctx, cancel := context.WithTimeout(context.Background(), versionCheckTimeout)
			beads.CheckBdVersionWarning(ctx)
			cancel()
		}
	} else {
		// Version check disabled - use backend from flag or default to bd
		detectedBackend = runtime.backend
		if detectedBackend == "" {
			detectedBackend = beads.BackendBd
		}
	}
	runtime.backend = detectedBackend

	// Pass the existing startup display to runWithRuntime
	if err := runWithRuntime(runtime, ui.NewApp, func(app *ui.App) programRunner {
		return tea.NewProgram(app, tea.WithAltScreen())
	}, func() startupAnimator {
		return startup
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

type programRunner interface {
	Run() (tea.Model, error)
}

type programFactory func(*ui.App) programRunner

type startupAnimator interface {
	ui.StartupReporter
	Stop()
}

func runProgram(cfg ui.Config, builder func(ui.Config) (*ui.App, error), factory programFactory) error {
	app, err := builder(cfg)
	if err != nil {
		if errors.Is(err, ui.ErrNoIssues) {
			return err
		}
		return fmt.Errorf("initialize UI: %w", err)
	}
	if factory == nil {
		return fmt.Errorf("program factory is nil")
	}
	prog := factory(app)
	if prog == nil {
		return fmt.Errorf("program is nil")
	}
	if _, err := prog.Run(); err != nil {
		return fmt.Errorf("run UI: %w", err)
	}
	return nil
}

type runtimeFlags struct {
	autoRefreshSeconds *int
	dbPath             *string
	outputFormat       *string
	skipVersionCheck   *bool
	skipUpdateCheck    *bool
	backend            *string
}

type runtimeOptions struct {
	refreshInterval  time.Duration
	autoRefresh      bool
	dbPath           string
	outputFormat     string
	skipVersionCheck bool
	skipUpdateCheck  bool
	backend          string
}

func computeRuntimeOptions(flags runtimeFlags, visited map[string]struct{}) runtimeOptions {
	seconds := sanitizeAutoRefreshSeconds(config.GetInt(config.KeyAutoRefreshSeconds))
	if flagWasExplicitlySet("auto-refresh-seconds", visited) {
		seconds = sanitizeAutoRefreshSeconds(*flags.autoRefreshSeconds)
	}
	refreshInterval := time.Duration(seconds) * time.Second
	autoRefresh := seconds > 0

	dbPath := strings.TrimSpace(config.GetString(config.KeyDatabasePath))
	if flagWasExplicitlySet("db-path", visited) {
		dbPath = strings.TrimSpace(*flags.dbPath)
	}

	outputFormat := strings.TrimSpace(config.GetString(config.KeyOutputFormat))
	if flagWasExplicitlySet("output-format", visited) {
		outputFormat = strings.TrimSpace(*flags.outputFormat)
	}

	skipVersionCheck := config.GetBool(config.KeySkipVersionCheck)
	if flagWasExplicitlySet("skip-version-check", visited) {
		skipVersionCheck = *flags.skipVersionCheck
	}

	skipUpdateCheck := config.GetBool(config.KeySkipUpdateCheck)
	if flagWasExplicitlySet("skip-update-check", visited) {
		skipUpdateCheck = *flags.skipUpdateCheck
	}

	// Backend flag is a one-time override - only use if explicitly set
	// Empty string means auto-detect (will happen in ab-pccw.3.14)
	backend := ""
	if flagWasExplicitlySet("backend", visited) {
		backend = strings.TrimSpace(*flags.backend)
	}

	return runtimeOptions{
		refreshInterval:  refreshInterval,
		autoRefresh:      autoRefresh,
		dbPath:           dbPath,
		outputFormat:     outputFormat,
		skipVersionCheck: skipVersionCheck,
		skipUpdateCheck:  skipUpdateCheck,
		backend:          backend,
	}
}

func flagWasExplicitlySet(name string, visited map[string]struct{}) bool {
	if _, ok := visited[name]; ok {
		return true
	}
	f := flag.CommandLine.Lookup(name)
	if f == nil {
		return false
	}
	return f.Value.String() != f.DefValue
}

func sanitizeAutoRefreshSeconds(seconds int) int {
	if seconds < 0 {
		return 0
	}
	return seconds
}

func runWithRuntime(
	runtime runtimeOptions,
	builder func(ui.Config) (*ui.App, error),
	factory programFactory,
	spinnerFactory func() startupAnimator,
) error {
	var spinner startupAnimator
	if spinnerFactory != nil {
		spinner = spinnerFactory()
	}

	// Start async update check (ab-a4qc)
	var updateChan chan *update.UpdateInfo
	if !runtime.skipUpdateCheck && Version != "" && Version != "dev" && Version != "development" {
		updateChan = make(chan *update.UpdateInfo, 1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					updateChan <- nil
				}
			}()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			checker := update.NewChecker(update.DefaultRepoOwner, update.DefaultRepoName)
			info, _ := checker.Check(ctx, Version)
			updateChan <- info // nil on error, which is fine
		}()
	}

	cfg := ui.Config{
		RefreshInterval: runtime.refreshInterval,
		AutoRefresh:     runtime.autoRefresh,
		DBPathOverride:  runtime.dbPath,
		OutputFormat:    runtime.outputFormat,
		Version:         Version,
		UpdateChan:      updateChan,
		Backend:         runtime.backend,
	}
	if spinner != nil {
		cfg.StartupReporter = spinner
	}

	spinnerStopped := false
	var appRef *ui.App // Keep reference to app for exit summary
	wrappedFactory := func(app *ui.App) programRunner {
		appRef = app // Store reference for exit summary
		if spinner != nil && !spinnerStopped {
			spinner.Stop()
			spinnerStopped = true
			// Clear the loading screen area before entering alt screen
			clearLoadingScreen(os.Stderr)
		}
		if factory == nil {
			return nil
		}
		return factory(app)
	}

	err := runProgram(cfg, builder, wrappedFactory)
	if spinner != nil && !spinnerStopped {
		spinner.Stop()
		spinnerStopped = true
	}

	// Print exit summary AFTER TUI exits (with final stats and session duration)
	if appRef != nil && err == nil {
		printExitSummary(os.Stderr, ExitSummary{
			Version:     cfg.Version,
			EndStats:    appRef.GetStats(),
			SessionInfo: appRef.GetSessionInfo(),
		})
	}

	return err
}
