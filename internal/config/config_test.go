package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitialize(t *testing.T) {
	reset()
	t.Cleanup(reset)

	tmp := t.TempDir()

	if err := Initialize(WithWorkingDir(tmp)); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}
	// Second call should no-op and still return nil.
	if err := Initialize(); err != nil {
		t.Fatalf("Initialize should be idempotent: %v", err)
	}
}

func TestDefaults(t *testing.T) {
	reset()
	t.Cleanup(reset)

	tmp := t.TempDir()
	userCfg := filepath.Join(tmp, "user.yaml")

	if err := Initialize(WithWorkingDir(tmp), WithUserConfig(userCfg)); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	if got := GetInt(KeyAutoRefreshSeconds); got != defaultAutoRefreshSeconds {
		t.Fatalf("expected default %s to be %ds, got %d", KeyAutoRefreshSeconds, defaultAutoRefreshSeconds, got)
	}
	if GetBool(KeyOutputJSON) {
		t.Fatalf("expected default %s to be false", KeyOutputJSON)
	}
	if got := GetString(KeyDatabasePath); got != "" {
		t.Fatalf("expected default %s to be empty, got %q", KeyDatabasePath, got)
	}
	if got := GetString(KeyOutputFormat); got != "rich" {
		t.Fatalf("expected default %s to be rich, got %q", KeyOutputFormat, got)
	}
}

func TestConfigFile(t *testing.T) {
	reset()
	t.Cleanup(reset)

	tmp := t.TempDir()
	projectDir := filepath.Join(tmp, "repo")
	mustMkdir(t, filepath.Join(projectDir, ".abacus"))
	projectCfg := filepath.Join(projectDir, ".abacus", "config.yaml")
	writeFile(t, projectCfg, `
auto-refresh-seconds: 10
output:
  format: project
database:
  path: /project/beads.db
`)

	userCfg := filepath.Join(tmp, "user.yaml")
	writeFile(t, userCfg, `
auto-refresh-seconds: 1
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
	if got := GetInt(KeyAutoRefreshSeconds); got != 10 {
		t.Fatalf("expected project auto-refresh seconds of 10, got %d", got)
	}
}

func TestEnvironmentBinding(t *testing.T) {
	reset()
	t.Cleanup(reset)

	tmp := t.TempDir()
	t.Setenv("AB_OUTPUT_FORMAT", "plain")
	t.Setenv("AB_AUTO_REFRESH_SECONDS", "12")

	if err := Initialize(WithWorkingDir(tmp)); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	if got := GetString(KeyOutputFormat); got != "plain" {
		t.Fatalf("expected env override for %s, got %q", KeyOutputFormat, got)
	}
	if got := GetInt(KeyAutoRefreshSeconds); got != 12 {
		t.Fatalf("expected env override for %s, got %d", KeyAutoRefreshSeconds, got)
	}
}

func TestLegacyAutoRefreshKeys(t *testing.T) {
	reset()
	t.Cleanup(reset)

	tmp := t.TempDir()
	projectDir := filepath.Join(tmp, "repo")
	mustMkdir(t, filepath.Join(projectDir, ".abacus"))
	projectCfg := filepath.Join(projectDir, ".abacus", "config.yaml")
	writeFile(t, projectCfg, `
refresh-interval: 750ms
auto-refresh: true
`)

	t.Setenv("AB_NO_AUTO_REFRESH", "true")

	if err := Initialize(
		WithWorkingDir(projectDir),
		WithProjectConfig(projectCfg),
	); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	if got := GetInt(KeyAutoRefreshSeconds); got != 0 {
		t.Fatalf("expected legacy no-auto-refresh env to disable auto refresh, got %d", got)
	}

	reset()

	t.Setenv("AB_REFRESH_INTERVAL", "1900ms")
	t.Setenv("AB_NO_AUTO_REFRESH", "")
	if err := Initialize(WithWorkingDir(tmp)); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}
	if got := GetInt(KeyAutoRefreshSeconds); got != 2 {
		t.Fatalf("expected legacy refresh-interval env to map to 2 seconds, got %d", got)
	}
}

func TestConfigPrecedence(t *testing.T) {
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
auto-refresh-seconds: 5
`)

	t.Setenv("AB_OUTPUT_JSON", "true")
	t.Setenv("AB_DATABASE_PATH", "/env/beads.db")
	t.Setenv("AB_AUTO_REFRESH_SECONDS", "7")

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
	if got := GetInt(KeyAutoRefreshSeconds); got != 7 {
		t.Fatalf("expected env override for %s=7, got %d", KeyAutoRefreshSeconds, got)
	}

	overrides := map[string]any{
		KeyOutputJSON:         false,
		KeyAutoRefreshSeconds: 11,
		"output.json_indent":  4,
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
	if got := GetInt(KeyAutoRefreshSeconds); got != 11 {
		t.Fatalf("expected CLI override to update %s to 11, got %d", KeyAutoRefreshSeconds, got)
	}
}

func TestSetUpdatesValue(t *testing.T) {
	reset()
	t.Cleanup(reset)

	tmp := t.TempDir()
	if err := Initialize(WithWorkingDir(tmp)); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	want := 42
	if err := Set(KeyAutoRefreshSeconds, want); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	if got := GetInt(KeyAutoRefreshSeconds); got != want {
		t.Fatalf("expected Set to update %s to %d, got %d", KeyAutoRefreshSeconds, want, got)
	}
}

func TestFindsAncestorProjectConfig(t *testing.T) {
	reset()
	t.Cleanup(reset)

	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	deep := filepath.Join(repo, "a", "b", "c")
	mustMkdir(t, filepath.Join(repo, ".abacus"))
	mustMkdir(t, deep)

	projectCfg := filepath.Join(repo, ".abacus", "config.yaml")
	writeFile(t, projectCfg, `
output:
  format: ancestor
`)

	if err := Initialize(
		WithWorkingDir(deep),
		WithUserConfig(filepath.Join(tmp, "user.yaml")),
	); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	if got := GetString(KeyOutputFormat); got != "ancestor" {
		t.Fatalf("expected ancestor config discovery, got %q", got)
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
