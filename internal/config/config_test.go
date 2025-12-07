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

	if got := GetString(KeyDatabasePath); got != "/env/beads.db" {
		t.Fatalf("expected env override for %s, got %q", KeyDatabasePath, got)
	}
	if got := GetInt(KeyAutoRefreshSeconds); got != 7 {
		t.Fatalf("expected env override for %s=7, got %d", KeyAutoRefreshSeconds, got)
	}
	if got := GetBool(KeyOutputJSON); !got {
		t.Fatalf("expected env override to set %s true", KeyOutputJSON)
	}

	overrides := map[string]any{
		KeyAutoRefreshSeconds: 11,
	}
	if err := ApplyOverrides(overrides); err != nil {
		t.Fatalf("ApplyOverrides returned error: %v", err)
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

func TestThemeDefault(t *testing.T) {
	reset()
	t.Cleanup(reset)

	tmp := t.TempDir()
	userCfg := filepath.Join(tmp, "user.yaml")

	if err := Initialize(WithWorkingDir(tmp), WithUserConfig(userCfg)); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	if got := GetString(KeyTheme); got != "tokyonight" {
		t.Fatalf("expected default theme to be tokyonight, got %q", got)
	}
}

func TestSaveThemeToUserConfig(t *testing.T) {
	reset()
	t.Cleanup(reset)

	tmp := t.TempDir()
	userDir := filepath.Join(tmp, ".abacus")
	userCfg := filepath.Join(userDir, "config.yaml")

	// Initialize with no project config - creates empty working dir
	workDir := filepath.Join(tmp, "work")
	mustMkdir(t, workDir)

	// Change to work dir so SaveTheme finds no project config
	oldWd, _ := os.Getwd()
	_ = os.Chdir(workDir)
	defer func() { _ = os.Chdir(oldWd) }()

	if err := Initialize(WithWorkingDir(workDir), WithUserConfig(userCfg)); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	// Set override so SaveTheme writes to our test path instead of real home
	setUserConfigPathOverride(userCfg)

	// Save theme - should create user config
	if err := SaveTheme("nord"); err != nil {
		t.Fatalf("SaveTheme returned error: %v", err)
	}

	// Verify file was created and contains theme
	data, err := os.ReadFile(userCfg)
	if err != nil {
		t.Fatalf("failed to read user config: %v", err)
	}
	if !contains(string(data), "theme: nord") {
		t.Fatalf("expected user config to contain 'theme: nord', got:\n%s", data)
	}
}

func TestSaveThemeToProjectConfig(t *testing.T) {
	reset()
	t.Cleanup(reset)

	tmp := t.TempDir()
	projectDir := filepath.Join(tmp, "repo")
	mustMkdir(t, filepath.Join(projectDir, ".abacus"))
	projectCfg := filepath.Join(projectDir, ".abacus", "config.yaml")

	// Create existing project config with other settings
	writeFile(t, projectCfg, `
output:
  format: rich
`)

	// Change to project dir so SaveTheme finds project config
	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectDir)
	defer func() { _ = os.Chdir(oldWd) }()

	userCfg := filepath.Join(tmp, "user.yaml")
	if err := Initialize(WithWorkingDir(projectDir), WithUserConfig(userCfg)); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	// Save theme - should update project config
	if err := SaveTheme("catppuccin"); err != nil {
		t.Fatalf("SaveTheme returned error: %v", err)
	}

	// Verify project config contains theme
	data, err := os.ReadFile(projectCfg)
	if err != nil {
		t.Fatalf("failed to read project config: %v", err)
	}
	if !contains(string(data), "theme: catppuccin") {
		t.Fatalf("expected project config to contain 'theme: catppuccin', got:\n%s", data)
	}
}

func TestSaveThemePreservesOtherSettings(t *testing.T) {
	reset()
	t.Cleanup(reset)

	tmp := t.TempDir()
	projectDir := filepath.Join(tmp, "repo")
	mustMkdir(t, filepath.Join(projectDir, ".abacus"))
	projectCfg := filepath.Join(projectDir, ".abacus", "config.yaml")

	// Create existing project config with various settings
	writeFile(t, projectCfg, `
output:
  format: plain
database:
  path: /custom/beads.db
auto-refresh-seconds: 15
`)

	// Change to project dir
	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectDir)
	defer func() { _ = os.Chdir(oldWd) }()

	userCfg := filepath.Join(tmp, "user.yaml")
	if err := Initialize(WithWorkingDir(projectDir), WithUserConfig(userCfg)); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	// Save theme
	if err := SaveTheme("tokyonight"); err != nil {
		t.Fatalf("SaveTheme returned error: %v", err)
	}

	// Verify other settings are preserved
	data, err := os.ReadFile(projectCfg)
	if err != nil {
		t.Fatalf("failed to read project config: %v", err)
	}
	content := string(data)
	if !contains(content, "theme: tokyonight") {
		t.Fatalf("expected theme to be saved, got:\n%s", content)
	}
	if !contains(content, "format: plain") {
		t.Fatalf("expected output.format to be preserved, got:\n%s", content)
	}
	if !contains(content, "/custom/beads.db") {
		t.Fatalf("expected database.path to be preserved, got:\n%s", content)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
