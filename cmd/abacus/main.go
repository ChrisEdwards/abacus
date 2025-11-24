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
	"abacus/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

const versionCheckTimeout = 5 * time.Second

func main() {
	if err := config.Initialize(); err != nil {
		fmt.Printf("Error initializing config: %v\n", err)
		os.Exit(1)
	}

	autoRefreshSecondsDefault := config.GetInt(config.KeyAutoRefreshSeconds)
	if autoRefreshSecondsDefault < 0 {
		autoRefreshSecondsDefault = 0
	}
	dbPathDefault := config.GetString(config.KeyDatabasePath)
	outputFormatDefault := config.GetString(config.KeyOutputFormat)
	skipVersionCheckDefault := config.GetBool(config.KeySkipVersionCheck)

	versionFlag := flag.Bool("version", false, "Print version information and exit")
	autoRefreshSecondsFlag := flag.Int("auto-refresh-seconds", autoRefreshSecondsDefault, "Auto-refresh interval in seconds (0 disables auto refresh)")
	dbPathFlag := flag.String("db-path", dbPathDefault, "Path to the Beads database file")
	outputFormatFlag := flag.String("output-format", outputFormatDefault, "Detail panel markdown style (rich, light, plain)")
	skipVersionCheckFlag := flag.Bool("skip-version-check", skipVersionCheckDefault, "Skip Beads CLI version validation (or set AB_SKIP_VERSION_CHECK=true)")
	jsonOutputFlag := flag.Bool("json-output", config.GetBool(config.KeyOutputJSON), "Print all issues as JSON and exit")
	flag.Parse()

	if *versionFlag {
		printVersion()
		os.Exit(0)
	}

	visited := map[string]struct{}{}
	flag.CommandLine.Visit(func(f *flag.Flag) {
		visited[f.Name] = struct{}{}
	})

	runtime := computeRuntimeOptions(runtimeFlags{
		autoRefreshSeconds: autoRefreshSecondsFlag,
		dbPath:             dbPathFlag,
		outputFormat:       outputFormatFlag,
		skipVersionCheck:   skipVersionCheckFlag,
		jsonOutput:         jsonOutputFlag,
	}, visited)

	skipVersionCheck := runtime.skipVersionCheck

	if !skipVersionCheck {
		ctx, cancel := context.WithTimeout(context.Background(), versionCheckTimeout)
		info, err := beads.CheckVersion(ctx, beads.VersionCheckOptions{})
		cancel()
		if handleVersionCheckResult(os.Stderr, info, err) {
			os.Exit(1)
		}
	}

	if err := runWithRuntime(runtime, ui.NewApp, func(app *ui.App) programRunner {
		return tea.NewProgram(app, tea.WithAltScreen())
	}, func() startupAnimator {
		return newStartupSpinner(os.Stderr, 500*time.Millisecond)
	}, ui.OutputIssuesJSON, func(path string) beads.Client {
		return beads.NewCLIClient(beads.WithDatabasePath(path))
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
	jsonOutput         *bool
}

type runtimeOptions struct {
	refreshInterval  time.Duration
	autoRefresh      bool
	dbPath           string
	outputFormat     string
	skipVersionCheck bool
	jsonOutput       bool
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

	jsonOutput := config.GetBool(config.KeyOutputJSON)
	if flagWasExplicitlySet("json-output", visited) {
		jsonOutput = *flags.jsonOutput
	}

	return runtimeOptions{
		refreshInterval:  refreshInterval,
		autoRefresh:      autoRefresh,
		dbPath:           dbPath,
		outputFormat:     outputFormat,
		skipVersionCheck: skipVersionCheck,
		jsonOutput:       jsonOutput,
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
	jsonPrinter func(context.Context, beads.Client) error,
	clientFactory func(string) beads.Client,
) error {
	if runtime.jsonOutput {
		if clientFactory == nil {
			clientFactory = func(path string) beads.Client {
				return beads.NewCLIClient(beads.WithDatabasePath(path))
			}
		}
		client := clientFactory(runtime.dbPath)
		if jsonPrinter == nil {
			jsonPrinter = ui.OutputIssuesJSON
		}
		return jsonPrinter(context.Background(), client)
	}

	var spinner startupAnimator
	if spinnerFactory != nil {
		spinner = spinnerFactory()
	}

	cfg := ui.Config{
		RefreshInterval: runtime.refreshInterval,
		AutoRefresh:     runtime.autoRefresh,
		DBPathOverride:  runtime.dbPath,
		OutputFormat:    runtime.outputFormat,
		Version:         Version,
	}
	if spinner != nil {
		cfg.StartupReporter = spinner
	}

	spinnerStopped := false
	wrappedFactory := func(app *ui.App) programRunner {
		if spinner != nil && !spinnerStopped {
			spinner.Stop()
			spinnerStopped = true
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
	return err
}
