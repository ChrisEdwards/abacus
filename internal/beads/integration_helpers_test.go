//go:build integration

package beads

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// =============================================================================
// Test Helpers for Integration Tests
// =============================================================================

// backendTestEnv holds the test environment for backend integration tests.
type backendTestEnv struct {
	Backend string // "bd" or "br"
	DBPath  string
	WorkDir string
	cleanup func()
}

// skipIfNoBackend skips the test if the specified backend binary is not available.
func skipIfNoBackend(t *testing.T, backend string) string {
	t.Helper()
	path, err := exec.LookPath(backend)
	if err != nil {
		t.Skipf("%s binary not found, skipping integration test", backend)
	}
	return path
}

// setupBackendTestDB creates a temp directory with an initialized database.
// Returns the test environment with dbPath, workDir, and a cleanup function.
func setupBackendTestDB(t *testing.T, backend string) backendTestEnv {
	t.Helper()

	dir := t.TempDir()
	beadsDir := filepath.Join(dir, ".beads")
	dbPath := filepath.Join(beadsDir, "beads.db")

	// Initialize database with the backend
	cmd := exec.Command(backend, "init", "--prefix", "test")
	cmd.Dir = dir // Run from temp dir to create .beads/ there
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s init failed: %v\nOutput: %s", backend, err, out)
	}

	// Verify the db was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatalf("%s init did not create expected database at %s", backend, dbPath)
	}

	return backendTestEnv{
		Backend: backend,
		DBPath:  dbPath,
		WorkDir: dir,
		cleanup: func() {
			// TempDir cleanup is automatic
		},
	}
}

// newClientForBackend creates a Client for the given backend using the test env.
func newClientForBackend(t *testing.T, env backendTestEnv) Client {
	t.Helper()
	switch env.Backend {
	case "br":
		// br needs WorkDir because it finds workspace by walking up from cwd
		return NewBrSQLiteClient(env.DBPath, WithBrWorkDir(env.WorkDir))
	case "bd":
		// bd uses --db flag directly and doesn't need WorkDir
		return NewBdSQLiteClient(env.DBPath)
	default:
		t.Fatalf("unknown backend: %s", env.Backend)
		return nil
	}
}

// extractCreatedID extracts the issue ID from create command output.
// Expected formats include:
// - br: "Created test-xxx: Title"
// - br: "✓ Created test-xxx: Title"
// - bd: "Created issue: test-xxx"
// - bd: "✓ Created issue: test-xxx"
func extractCreatedID(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		for _, field := range strings.Fields(line) {
			id := strings.Trim(field, ":,.;[](){}")
			if strings.HasPrefix(id, "test-") || strings.HasPrefix(id, "ab-") {
				return id
			}
		}
	}
	return ""
}

func TestExtractCreatedID(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "br plain output",
			output: "Created test-abc: Show Test Issue\n",
			want:   "test-abc",
		},
		{
			name:   "br output with checkmark prefix",
			output: "✓ Created test-wtd: Show Test Issue\n",
			want:   "test-wtd",
		},
		{
			name:   "bd plain output",
			output: "Created issue: test-xyz\n",
			want:   "test-xyz",
		},
		{
			name:   "bd output with checkmark prefix",
			output: "✓ Created issue: ab-123\n",
			want:   "ab-123",
		},
		{
			name:   "id on its own line",
			output: "\n  ab-standalone \n",
			want:   "ab-standalone",
		},
		{
			name:   "no id present",
			output: "nothing useful here\n",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractCreatedID(tt.output)
			if got != tt.want {
				t.Fatalf("extractCreatedID(%q) = %q, want %q", tt.output, got, tt.want)
			}
		})
	}
}
