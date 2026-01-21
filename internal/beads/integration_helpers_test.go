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
// Expected formats:
// - br: "Created test-xxx: Title"
// - bd: "Created issue: test-xxx"
func extractCreatedID(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// br format: "Created test-xxx: Title" - extract just the ID
		if strings.HasPrefix(line, "Created ") {
			rest := strings.TrimPrefix(line, "Created ")
			// ID ends at ": " or end of line
			if colonIdx := strings.Index(rest, ": "); colonIdx != -1 {
				return rest[:colonIdx]
			}
			// No colon, return first word
			parts := strings.Fields(rest)
			if len(parts) > 0 {
				return parts[0]
			}
			return rest
		}

		// bd format: "Created issue: test-xxx"
		if strings.Contains(line, "Created issue:") {
			parts := strings.Split(line, "Created issue:")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[1])
			}
		}

		// Fallback: just the ID on a line
		if strings.HasPrefix(line, "test-") || strings.HasPrefix(line, "ab-") {
			// Make sure it's just the ID, not part of a longer line
			parts := strings.Fields(line)
			if len(parts) > 0 && (strings.HasPrefix(parts[0], "test-") || strings.HasPrefix(parts[0], "ab-")) {
				// Remove any trailing punctuation
				id := parts[0]
				id = strings.TrimSuffix(id, ":")
				return id
			}
		}
	}
	return ""
}
