package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInitializeLoadsDefaults(t *testing.T) {
	reset()
	t.Cleanup(reset)

	tmp := t.TempDir()
	userCfg := filepath.Join(tmp, "user.yaml")

	if err := Initialize(WithWorkingDir(tmp), WithUserConfig(userCfg)); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	if got := GetDuration(KeyRefreshInterval); got != 3*time.Second {
		t.Fatalf("expected default %s to be 3s, got %v", KeyRefreshInterval, got)
	}
	if GetBool(KeyOutputJSON) {
		t.Fatalf("expected default %s to be false", KeyOutputJSON)
	}
	if got := GetString(KeyDatabasePath); got != "" {
		t.Fatalf("expected default %s to be empty, got %q", KeyDatabasePath, got)
	}
	if !GetBool(KeyAutoRefresh) {
		t.Fatalf("expected default %s to be true", KeyAutoRefresh)
	}
	if !GetBool(KeySyncAuto) {
		t.Fatalf("expected alias %s to remain true", KeySyncAuto)
	}
	if got := GetString(KeyOutputFormat); got != "rich" {
		t.Fatalf("expected default %s to be rich, got %q", KeyOutputFormat, got)
	}
}

func TestProjectConfigOverridesUser(t *testing.T) {
	reset()
	t.Cleanup(reset)

	tmp := t.TempDir()
	projectDir := filepath.Join(tmp, "repo")
	mustMkdir(t, filepath.Join(projectDir, ".abacus"))
	projectCfg := filepath.Join(projectDir, ".abacus", "config.yaml")
	writeFile(t, projectCfg, `
refresh-interval: 10s
auto-refresh: true
output:
  format: project
database:
  path: /project/beads.db
`)

	userCfg := filepath.Join(tmp, "user.yaml")
	writeFile(t, userCfg, `
refresh-interval: 1s
auto-refresh: false
output:
  format: user
database:
  path: /user/beads.db
`)

	if err := Initialize(
		WithWorkingDir(projectDir),
		WithUserConfig(userCfg),
	); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	if got := GetString(KeyOutputFormat); got != "project" {
		t.Fatalf("expected project config to win for %s, got %q", KeyOutputFormat, got)
	}
	if got := GetString(KeyDatabasePath); got != "/project/beads.db" {
		t.Fatalf("expected project database path, got %q", got)
	}
	if got := GetDuration(KeyRefreshInterval); got != 10*time.Second {
		t.Fatalf("expected project refresh interval of 10s, got %v", got)
	}
	if !GetBool(KeyAutoRefresh) {
		t.Fatalf("expected auto-refresh to be true after merging project config")
	}
}

func TestEnvironmentAndOverridesPrecedence(t *testing.T) {
	reset()
	t.Cleanup(reset)

	tmp := t.TempDir()
	projectDir := filepath.Join(tmp, "repo")
	mustMkdir(t, filepath.Join(projectDir, ".abacus"))
	projectCfg := filepath.Join(projectDir, ".abacus", "config.yaml")
	writeFile(t, projectCfg, `
output:
  json: false
  format: project
database:
  path: /project/beads.db
refresh-interval: 5s
`)

	t.Setenv("AB_OUTPUT_JSON", "true")
	t.Setenv("AB_DATABASE_PATH", "/env/beads.db")
	t.Setenv("AB_REFRESH_INTERVAL", "250ms")
	t.Setenv("AB_AUTO_REFRESH", "false")

	if err := Initialize(
		WithWorkingDir(projectDir),
		WithProjectConfig(projectCfg),
	); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	if !GetBool(KeyOutputJSON) {
		t.Fatalf("expected environment variable to override %s", KeyOutputJSON)
	}
	if got := GetString(KeyDatabasePath); got != "/env/beads.db" {
		t.Fatalf("expected env override for %s, got %q", KeyDatabasePath, got)
	}
	if got := GetDuration(KeyRefreshInterval); got != 250*time.Millisecond {
		t.Fatalf("expected env override for %s=250ms, got %v", KeyRefreshInterval, got)
	}
	if GetBool(KeyAutoRefresh) {
		t.Fatalf("expected env override to disable %s", KeyAutoRefresh)
	}

	overrides := map[string]any{
		KeyOutputJSON:        false,
		KeyAutoRefresh:       true,
		"output.json_indent": 4,
	}
	if err := ApplyOverrides(overrides); err != nil {
		t.Fatalf("ApplyOverrides returned error: %v", err)
	}

	if GetBool(KeyOutputJSON) {
		t.Fatalf("expected CLI override to set %s=false", KeyOutputJSON)
	}
	if got := GetInt("output.json_indent"); got != 4 {
		t.Fatalf("expected override for output.json_indent = 4, got %d", got)
	}
	if !GetBool(KeyAutoRefresh) {
		t.Fatalf("expected CLI override to re-enable %s", KeyAutoRefresh)
	}
}

func TestSetUpdatesValue(t *testing.T) {
	reset()
	t.Cleanup(reset)

	tmp := t.TempDir()
	if err := Initialize(WithWorkingDir(tmp)); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	want := 42 * time.Second
	if err := Set(KeyRefreshInterval, want); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	if got := GetDuration(KeyRefreshInterval); got != want {
		t.Fatalf("expected Set to update %s to %v, got %v", KeyRefreshInterval, want, got)
	}
}

func mustMkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	mustMkdir(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
