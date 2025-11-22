package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitializeLoadsDefaults(t *testing.T) {
	reset()
	t.Cleanup(reset)

	tmp := t.TempDir()
	userCfg := filepath.Join(tmp, "user.yaml")

	if err := Initialize(WithWorkingDir(tmp), WithUserConfig(userCfg)); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	if GetBool(KeyOutputJSON) {
		t.Fatalf("expected default %s to be false", KeyOutputJSON)
	}
	if got := GetString(KeyDatabasePath); got != "" {
		t.Fatalf("expected default %s to be empty, got %q", KeyDatabasePath, got)
	}
	if !GetBool(KeySyncAuto) {
		t.Fatalf("expected default %s to be true", KeySyncAuto)
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
output:
  format: project
database:
  path: /project/beads.db
sync:
  auto: true
`)

	userCfg := filepath.Join(tmp, "user.yaml")
	writeFile(t, userCfg, `
output:
  format: user
database:
  path: /user/beads.db
sync:
  auto: false
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
	if !GetBool(KeySyncAuto) {
		t.Fatalf("expected sync.auto to be true after merging project config")
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
`)

	t.Setenv("AB_OUTPUT_JSON", "true")
	t.Setenv("AB_DATABASE_PATH", "/env/beads.db")

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

	overrides := map[string]any{
		KeyOutputJSON:        false,
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
