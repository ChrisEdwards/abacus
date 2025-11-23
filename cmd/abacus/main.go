package main

import (
	"context"
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

	refreshDefault := config.GetDuration(config.KeyRefreshInterval)
	if refreshDefault <= 0 {
		refreshDefault = 3 * time.Second
	}
	autoRefreshDefault := config.GetBool(config.KeyAutoRefresh)
	noAutoRefreshDefault := config.GetBool(config.KeyNoAutoRefresh)
	jsonOutputDefault := config.GetBool(config.KeyOutputJSON)
	dbPathDefault := config.GetString(config.KeyDatabasePath)
	outputFormatDefault := config.GetString(config.KeyOutputFormat)
	skipVersionCheckDefault := config.GetBool(config.KeySkipVersionCheck)

	versionFlag := flag.Bool("version", false, "Print version information and exit")
	refreshIntervalFlag := flag.Duration("refresh-interval", refreshDefault, "Interval for automatic refresh polling (e.g. 2s, 500ms)")
	autoRefreshFlag := flag.Bool("auto-refresh", autoRefreshDefault, "Enable automatic background refresh")
	noAutoRefreshFlag := flag.Bool("no-auto-refresh", noAutoRefreshDefault, "Disable automatic background refresh (overrides --auto-refresh)")
	jsonOutputFlag := flag.Bool("json-output", jsonOutputDefault, "Print issue data as JSON and exit")
	dbPathFlag := flag.String("db-path", dbPathDefault, "Path to the Beads database file")
	outputFormatFlag := flag.String("output-format", outputFormatDefault, "Detail panel markdown style (rich, light, plain)")
	skipVersionCheckFlag := flag.Bool("skip-version-check", skipVersionCheckDefault, "Skip Beads CLI version validation (or set AB_SKIP_VERSION_CHECK=true)")
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
		refreshInterval:  refreshIntervalFlag,
		autoRefresh:      autoRefreshFlag,
		noAutoRefresh:    noAutoRefreshFlag,
		dbPath:           dbPathFlag,
		outputFormat:     outputFormatFlag,
		jsonOutput:       jsonOutputFlag,
		skipVersionCheck: skipVersionCheckFlag,
		refreshDefault:   refreshDefault,
	}, visited)

	refreshInterval := runtime.refreshInterval
	autoRefresh := runtime.autoRefresh
	dbPath := runtime.dbPath
	outputFormat := runtime.outputFormat
	jsonOutput := runtime.jsonOutput
	skipVersionCheck := runtime.skipVersionCheck

	if !skipVersionCheck {
		ctx, cancel := context.WithTimeout(context.Background(), versionCheckTimeout)
		info, err := beads.CheckVersion(ctx, beads.VersionCheckOptions{})
		cancel()
		if handleVersionCheckResult(os.Stderr, info, err) {
			os.Exit(1)
		}
	}

	client := beads.NewCLIClient()
	appCfg := ui.Config{
		RefreshInterval: refreshInterval,
		AutoRefresh:     autoRefresh,
		DBPathOverride:  dbPath,
		OutputFormat:    outputFormat,
		Client:          client,
		Version:         Version,
	}

	if jsonOutput {
		if err := ui.OutputIssuesJSON(context.Background(), client); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	app, err := ui.NewApp(appCfg)
	if err != nil {
		fmt.Printf("Error initializing UI: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

type runtimeFlags struct {
	refreshInterval  *time.Duration
	autoRefresh      *bool
	noAutoRefresh    *bool
	dbPath           *string
	outputFormat     *string
	jsonOutput       *bool
	skipVersionCheck *bool
	refreshDefault   time.Duration
}

type runtimeOptions struct {
	refreshInterval  time.Duration
	autoRefresh      bool
	dbPath           string
	outputFormat     string
	jsonOutput       bool
	skipVersionCheck bool
}

func computeRuntimeOptions(flags runtimeFlags, visited map[string]struct{}) runtimeOptions {
	refreshInterval := flags.refreshDefault
	if flagWasExplicitlySet("refresh-interval", visited) {
		refreshInterval = *flags.refreshInterval
	} else {
		if cfgInterval := config.GetDuration(config.KeyRefreshInterval); cfgInterval > 0 {
			refreshInterval = cfgInterval
		}
	}

	autoRefresh := config.GetBool(config.KeyAutoRefresh)
	if config.GetBool(config.KeyNoAutoRefresh) {
		autoRefresh = false
	}
	if flagWasExplicitlySet("auto-refresh", visited) {
		autoRefresh = *flags.autoRefresh
	}
	if flagWasExplicitlySet("no-auto-refresh", visited) {
		autoRefresh = !*flags.noAutoRefresh
	}

	dbPath := strings.TrimSpace(config.GetString(config.KeyDatabasePath))
	if flagWasExplicitlySet("db-path", visited) {
		dbPath = strings.TrimSpace(*flags.dbPath)
	}

	outputFormat := strings.TrimSpace(config.GetString(config.KeyOutputFormat))
	if flagWasExplicitlySet("output-format", visited) {
		outputFormat = strings.TrimSpace(*flags.outputFormat)
	}

	jsonOutput := config.GetBool(config.KeyOutputJSON)
	if flagWasExplicitlySet("json-output", visited) {
		jsonOutput = *flags.jsonOutput
	}

	skipVersionCheck := config.GetBool(config.KeySkipVersionCheck)
	if flagWasExplicitlySet("skip-version-check", visited) {
		skipVersionCheck = *flags.skipVersionCheck
	}

	return runtimeOptions{
		refreshInterval:  refreshInterval,
		autoRefresh:      autoRefresh,
		dbPath:           dbPath,
		outputFormat:     outputFormat,
		jsonOutput:       jsonOutput,
		skipVersionCheck: skipVersionCheck,
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
