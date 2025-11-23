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
		config.KeyAutoRefreshSeconds: 3,
		config.KeyDatabasePath:       "",
		config.KeyOutputFormat:       "",
		config.KeySkipVersionCheck:   false,
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

	autoRefreshSecondsDefault := config.GetInt(config.KeyAutoRefreshSeconds)
	dbPathDefault := config.GetString(config.KeyDatabasePath)
	outputFormatDefault := config.GetString(config.KeyOutputFormat)
	skipVersionCheckDefault := config.GetBool(config.KeySkipVersionCheck)

	fs := flag.NewFlagSet("abacus-test", flag.ContinueOnError)
	autoRefreshSecondsFlag := fs.Int("auto-refresh-seconds", autoRefreshSecondsDefault, "test auto refresh seconds")
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
		autoRefreshSeconds: autoRefreshSecondsFlag,
		dbPath:             dbPathFlag,
		outputFormat:       outputFormatFlag,
		skipVersionCheck:   skipVersionCheckFlag,
	}
	return computeRuntimeOptions(flags, visited)
}

func TestComputeRuntimeOptions_AutoRefreshSecondsFlagOverridesConfig(t *testing.T) {
	opts := buildRuntimeOptionsForArgs(t, []string{"--auto-refresh-seconds=5"}, map[string]any{config.KeyAutoRefreshSeconds: 9})
	if opts.refreshInterval != 5*time.Second {
		t.Fatalf("expected refresh interval 5s, got %v", opts.refreshInterval)
	}
	if !opts.autoRefresh {
		t.Fatalf("expected positive seconds to enable auto refresh")
	}
}

func TestComputeRuntimeOptions_AutoRefreshSecondsZeroDisables(t *testing.T) {
	opts := buildRuntimeOptionsForArgs(t, []string{"--auto-refresh-seconds=0"})
	if opts.autoRefresh {
		t.Fatalf("expected zero seconds to disable auto refresh")
	}
	if opts.refreshInterval != 0 {
		t.Fatalf("expected refresh interval 0 when disabled, got %v", opts.refreshInterval)
	}
}

func TestComputeRuntimeOptions_ConfigSecondsUsed(t *testing.T) {
	opts := buildRuntimeOptionsForArgs(t, []string{}, map[string]any{config.KeyAutoRefreshSeconds: 7})
	if opts.refreshInterval != 7*time.Second {
		t.Fatalf("expected config auto refresh seconds to drive interval, got %v", opts.refreshInterval)
	}
	if !opts.autoRefresh {
		t.Fatalf("expected positive config seconds to enable auto refresh")
	}
}

func TestComputeRuntimeOptions_NegativeSecondsDisable(t *testing.T) {
	opts := buildRuntimeOptionsForArgs(t, []string{"--auto-refresh-seconds=-5"})
	if opts.autoRefresh {
		t.Fatalf("expected negative seconds to disable auto refresh")
	}
	if opts.refreshInterval != 0 {
		t.Fatalf("expected refresh interval 0 for negative seconds, got %v", opts.refreshInterval)
	}
}

func TestComputeRuntimeOptions_DBPathOverride(t *testing.T) {
	opts := buildRuntimeOptionsForArgs(t, []string{"--db-path", " /tmp/custom.db "})
	if opts.dbPath != "/tmp/custom.db" {
		t.Fatalf("expected db path trimmed, got %q", opts.dbPath)
	}
}

func TestComputeRuntimeOptions_SkipVersionFlag(t *testing.T) {
	opts := buildRuntimeOptionsForArgs(t, []string{"--skip-version-check"})
	if !opts.skipVersionCheck {
		t.Fatalf("expected skip version flag to be true")
	}
}
