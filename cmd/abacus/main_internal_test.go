package main

import (
	"flag"
	"sync"
	"testing"
	"time"

	"abacus/internal/config"
)

var configInitOnce sync.Once

func ensureTestConfig(t *testing.T) {
	t.Helper()
	configInitOnce.Do(func() {
		dir := t.TempDir()
		if err := config.Initialize(
			config.WithProjectConfig(""),
			config.WithUserConfig(""),
			config.WithWorkingDir(dir),
		); err != nil {
			t.Fatalf("init config: %v", err)
		}
	})
	overrides := map[string]any{
		config.KeyRefreshInterval:  3 * time.Second,
		config.KeyAutoRefresh:      true,
		config.KeyNoAutoRefresh:    false,
		config.KeyOutputJSON:       false,
		config.KeyDatabasePath:     "",
		config.KeyOutputFormat:     "",
		config.KeySkipVersionCheck: false,
	}
	if err := config.ApplyOverrides(overrides); err != nil {
		t.Fatalf("apply overrides: %v", err)
	}
}

func buildRuntimeOptionsForArgs(t *testing.T, args []string, overrides ...map[string]any) runtimeOptions {
	t.Helper()
	ensureTestConfig(t)
	if len(overrides) > 0 && len(overrides[0]) > 0 {
		if err := config.ApplyOverrides(overrides[0]); err != nil {
			t.Fatalf("apply custom overrides: %v", err)
		}
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

	fs := flag.NewFlagSet("abacus-test", flag.ContinueOnError)
	refreshIntervalFlag := fs.Duration("refresh-interval", refreshDefault, "test interval")
	autoRefreshFlag := fs.Bool("auto-refresh", autoRefreshDefault, "test auto refresh")
	noAutoRefreshFlag := fs.Bool("no-auto-refresh", noAutoRefreshDefault, "test no auto refresh")
	jsonOutputFlag := fs.Bool("json-output", jsonOutputDefault, "test json output")
	dbPathFlag := fs.String("db-path", dbPathDefault, "db path")
	outputFormatFlag := fs.String("output-format", outputFormatDefault, "output format")
	skipVersionCheckFlag := fs.Bool("skip-version-check", skipVersionCheckDefault, "skip version")

	if err := fs.Parse(args); err != nil {
		t.Fatalf("parse args: %v", err)
	}
	visited := map[string]struct{}{}
	fs.Visit(func(f *flag.Flag) {
		visited[f.Name] = struct{}{}
	})

	flags := runtimeFlags{
		refreshInterval:  refreshIntervalFlag,
		autoRefresh:      autoRefreshFlag,
		noAutoRefresh:    noAutoRefreshFlag,
		dbPath:           dbPathFlag,
		outputFormat:     outputFormatFlag,
		jsonOutput:       jsonOutputFlag,
		skipVersionCheck: skipVersionCheckFlag,
		refreshDefault:   refreshDefault,
	}
	return computeRuntimeOptions(flags, visited)
}

func TestComputeRuntimeOptions_NoAutoRefreshOverrides(t *testing.T) {
	opts := buildRuntimeOptionsForArgs(t, []string{"--no-auto-refresh"})
	if opts.autoRefresh {
		t.Fatalf("expected auto refresh disabled when --no-auto-refresh supplied")
	}
}

func TestComputeRuntimeOptions_AutoRefreshOverridesConfig(t *testing.T) {
	opts := buildRuntimeOptionsForArgs(t, []string{"--auto-refresh"}, map[string]any{config.KeyAutoRefresh: false})
	if !opts.autoRefresh {
		t.Fatalf("expected --auto-refresh to enable auto refresh even if config disabled")
	}
}

func TestComputeRuntimeOptions_RefreshIntervalFromFlag(t *testing.T) {
	opts := buildRuntimeOptionsForArgs(t, []string{"--refresh-interval=5s"})
	if opts.refreshInterval != 5*time.Second {
		t.Fatalf("expected refresh interval 5s, got %v", opts.refreshInterval)
	}
}

func TestComputeRuntimeOptions_RefreshIntervalFromConfig(t *testing.T) {
	opts := buildRuntimeOptionsForArgs(t, []string{}, map[string]any{config.KeyRefreshInterval: 9 * time.Second})
	if opts.refreshInterval != 9*time.Second {
		t.Fatalf("expected refresh interval from config 9s, got %v", opts.refreshInterval)
	}
}

func TestComputeRuntimeOptions_DBPathOverride(t *testing.T) {
	opts := buildRuntimeOptionsForArgs(t, []string{"--db-path", " /tmp/custom.db "})
	if opts.dbPath != "/tmp/custom.db" {
		t.Fatalf("expected db path trimmed, got %q", opts.dbPath)
	}
}

func TestComputeRuntimeOptions_JSONOutputFlag(t *testing.T) {
	opts := buildRuntimeOptionsForArgs(t, []string{"--json-output"})
	if !opts.jsonOutput {
		t.Fatalf("expected json output flag to be respected")
	}
}

func TestComputeRuntimeOptions_SkipVersionFlag(t *testing.T) {
	opts := buildRuntimeOptionsForArgs(t, []string{"--skip-version-check"})
	if !opts.skipVersionCheck {
		t.Fatalf("expected skip version flag to be true")
	}
}
