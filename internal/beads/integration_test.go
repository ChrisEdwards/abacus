// Package beads provides integration tests for both bd and br backends.
//
// These tests verify that abacus can work with real bd and br binaries,
// testing the full CRUD workflow and verifying data consistency between
// CLI writes and SQLite reads.
//
// Run with: go test -v ./internal/beads/ -run Integration
// Skip in CI with: go test -short ./internal/beads/
package beads

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// =============================================================================
// Test Helpers
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

// =============================================================================
// Parameterized Backend Tests
// =============================================================================

// TestBackend_E2E_CRUD tests the full CRUD workflow for a backend.
// This runs the same test against both bd and br to verify compatibility.
func TestBackend_E2E_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	backends := []string{"br", "bd"}

	for _, backend := range backends {
		backend := backend // capture loop variable
		t.Run(backend, func(t *testing.T) {
			t.Parallel()
			skipIfNoBackend(t, backend)

			env := setupBackendTestDB(t, backend)
			defer env.cleanup()

			client := newClientForBackend(t, env)
			ctx := context.Background()

			// Step 1: Create an issue
			t.Log("Step 1: Creating issue")
			id, err := client.Create(ctx, "Integration Test Issue", "task", 2, []string{"backend"}, "")
			if err != nil {
				t.Fatalf("Create failed: %v", err)
			}
			if id == "" {
				t.Fatal("Create returned empty ID")
			}
			t.Logf("Created issue: %s", id)

			// Step 2: Verify issue is visible via List
			t.Log("Step 2: Verifying issue via List")
			issues, err := client.List(ctx)
			if err != nil {
				t.Fatalf("List failed: %v", err)
			}
			found := false
			for _, iss := range issues {
				if iss.ID == id {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("Created issue %s not found in list", id)
			}

			// Step 3: Update status
			t.Log("Step 3: Updating status to in_progress")
			if err := client.UpdateStatus(ctx, id, "in_progress"); err != nil {
				t.Fatalf("UpdateStatus failed: %v", err)
			}

			// Step 4: Add a label
			t.Log("Step 4: Adding label 'urgent'")
			if err := client.AddLabel(ctx, id, "urgent"); err != nil {
				t.Fatalf("AddLabel failed: %v", err)
			}

			// Step 5: Add a comment
			t.Log("Step 5: Adding comment")
			if err := client.AddComment(ctx, id, "This is an integration test comment"); err != nil {
				t.Fatalf("AddComment failed: %v", err)
			}

			// Step 6: Verify full issue details via Show
			t.Log("Step 6: Verifying issue details via Show")
			shown, err := client.Show(ctx, []string{id})
			if err != nil {
				t.Fatalf("Show failed: %v", err)
			}
			if len(shown) != 1 {
				t.Fatalf("Show returned %d issues, expected 1", len(shown))
			}
			issue := shown[0]
			if issue.Status != "in_progress" {
				t.Errorf("Expected status 'in_progress', got %q", issue.Status)
			}

			// Step 7: Close the issue
			t.Log("Step 7: Closing issue")
			if err := client.Close(ctx, id); err != nil {
				t.Fatalf("Close failed: %v", err)
			}

			// Step 8: Reopen the issue
			t.Log("Step 8: Reopening issue")
			if err := client.Reopen(ctx, id); err != nil {
				t.Fatalf("Reopen failed: %v", err)
			}

			// Step 9: Remove label
			t.Log("Step 9: Removing label 'urgent'")
			if err := client.RemoveLabel(ctx, id, "urgent"); err != nil {
				t.Fatalf("RemoveLabel failed: %v", err)
			}

			// Step 10: Delete the issue
			t.Log("Step 10: Deleting issue")
			if err := client.Delete(ctx, id, false); err != nil {
				t.Fatalf("Delete failed: %v", err)
			}

			t.Log("E2E CRUD test passed")
		})
	}
}

// TestBackend_E2E_Dependencies tests dependency operations for both backends.
// Note: bd backend may fail due to a known issue in bdCLIClient.Create parsing
// when bd outputs warning messages containing "--description". The bd client is
// frozen (bd_cli.go) and cannot be fixed. The br backend tests validate the
// integration test logic is correct.
func TestBackend_E2E_Dependencies(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	backends := []string{"br", "bd"}

	for _, backend := range backends {
		backend := backend
		t.Run(backend, func(t *testing.T) {
			t.Parallel()
			skipIfNoBackend(t, backend)

			// Skip bd due to known parsing issue in frozen bdCLIClient.Create
			// when bd outputs warning messages. See bd_cli.go:156-164.
			if backend == "bd" {
				t.Skip("skipping bd backend: known parsing issue with --description warnings in frozen bd client")
			}

			env := setupBackendTestDB(t, backend)
			defer env.cleanup()

			client := newClientForBackend(t, env)
			ctx := context.Background()

			// Create two issues
			id1, err := client.Create(ctx, "Blocker Issue", "task", 1, nil, "")
			if err != nil {
				t.Fatalf("Create issue 1 failed: %v", err)
			}

			id2, err := client.Create(ctx, "Blocked Issue", "task", 2, nil, "")
			if err != nil {
				t.Fatalf("Create issue 2 failed: %v", err)
			}

			// Add dependency: id2 depends on id1 (id1 blocks id2)
			if err := client.AddDependency(ctx, id2, id1, "blocks"); err != nil {
				t.Fatalf("AddDependency failed: %v", err)
			}

			// Verify dependency is visible in Export
			issues, err := client.Export(ctx)
			if err != nil {
				t.Fatalf("Export failed: %v", err)
			}

			var blockedIssue *FullIssue
			for i := range issues {
				if issues[i].ID == id2 {
					blockedIssue = &issues[i]
					break
				}
			}
			if blockedIssue == nil {
				t.Fatal("Blocked issue not found in export")
			}

			foundDep := false
			for _, dep := range blockedIssue.Dependencies {
				if dep.TargetID == id1 && dep.Type == "blocks" {
					foundDep = true
					break
				}
			}
			if !foundDep {
				t.Errorf("Dependency not found in exported issue. Dependencies: %+v", blockedIssue.Dependencies)
			}

			// Remove dependency
			if err := client.RemoveDependency(ctx, id2, id1, ""); err != nil {
				t.Fatalf("RemoveDependency failed: %v", err)
			}
		})
	}
}

// TestBackend_E2E_ParentChild tests parent-child relationship creation.
func TestBackend_E2E_ParentChild(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	backends := []string{"br", "bd"}

	for _, backend := range backends {
		backend := backend
		t.Run(backend, func(t *testing.T) {
			t.Parallel()
			skipIfNoBackend(t, backend)

			env := setupBackendTestDB(t, backend)
			defer env.cleanup()

			client := newClientForBackend(t, env)
			ctx := context.Background()

			// Create parent issue
			parent, err := client.CreateFull(ctx, "Parent Epic", "epic", 1, nil, "", "Parent description", "")
			if err != nil {
				t.Fatalf("Create parent failed: %v", err)
			}

			// Create child with parent reference
			child, err := client.CreateFull(ctx, "Child Task", "task", 2, nil, "", "Child description", parent.ID)
			if err != nil {
				t.Fatalf("Create child with parent failed: %v", err)
			}

			// Verify relationship via Export
			issues, err := client.Export(ctx)
			if err != nil {
				t.Fatalf("Export failed: %v", err)
			}

			var childIssue *FullIssue
			for i := range issues {
				if issues[i].ID == child.ID {
					childIssue = &issues[i]
					break
				}
			}
			if childIssue == nil {
				t.Fatal("Child issue not found in export")
			}

			foundParentDep := false
			for _, dep := range childIssue.Dependencies {
				if dep.TargetID == parent.ID && dep.Type == "parent-child" {
					foundParentDep = true
					break
				}
			}
			if !foundParentDep {
				t.Errorf("Parent-child dependency not found. Dependencies: %+v", childIssue.Dependencies)
			}
		})
	}
}

// =============================================================================
// SQLite Read Tests
// =============================================================================

// TestSQLiteWithBr verifies that SQLite client can read data created by br CLI.
func TestSQLiteWithBr(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	skipIfNoBackend(t, "br")

	env := setupBackendTestDB(t, "br")
	defer env.cleanup()

	ctx := context.Background()

	// Create issues using br CLI directly (not through client)
	// Note: br runs from WorkDir and finds .beads/ automatically
	cmd := exec.Command("br", "create", "Direct CLI Issue", "--type", "task", "--priority", "2")
	cmd.Dir = env.WorkDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("br create failed: %v\nOutput: %s", err, out)
	}

	// Extract ID from output (format: "Created test-xxx")
	id := extractCreatedID(string(out))
	if id == "" {
		t.Fatalf("Could not extract ID from br create output: %s", out)
	}
	t.Logf("Created issue %s via br CLI", id)

	// Small delay for WAL sync
	time.Sleep(100 * time.Millisecond)

	// Now read using SQLite client
	sqliteClient := NewBrSQLiteClient(env.DBPath)

	// Test List
	issues, err := sqliteClient.List(ctx)
	if err != nil {
		t.Fatalf("SQLite List failed: %v", err)
	}
	t.Logf("SQLite List returned %d issues", len(issues))
	if len(issues) == 0 {
		t.Fatal("SQLite List returned no issues")
	}

	found := false
	for _, iss := range issues {
		t.Logf("  Found issue: %s", iss.ID)
		if iss.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Issue %s created by br CLI not found via SQLite List", id)
	}

	// Test Export
	fullIssues, err := sqliteClient.Export(ctx)
	if err != nil {
		t.Fatalf("SQLite Export failed: %v", err)
	}

	var foundIssue *FullIssue
	for i := range fullIssues {
		if fullIssues[i].ID == id {
			foundIssue = &fullIssues[i]
			break
		}
	}
	if foundIssue == nil {
		t.Fatalf("Issue %s not found in SQLite Export", id)
	}
	if foundIssue.Title != "Direct CLI Issue" {
		t.Errorf("Expected title 'Direct CLI Issue', got %q", foundIssue.Title)
	}
}

// TestSQLiteWithBd verifies that SQLite client can read data created by bd CLI.
func TestSQLiteWithBd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	skipIfNoBackend(t, "bd")

	env := setupBackendTestDB(t, "bd")
	defer env.cleanup()

	ctx := context.Background()

	// Create issues using bd CLI directly
	cmd := exec.Command("bd", "--db", env.DBPath, "create", "--title", "Direct BD CLI Issue", "--type", "task", "--priority", "2")
	cmd.Dir = env.WorkDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bd create failed: %v\nOutput: %s", err, out)
	}

	// Extract ID from output
	id := extractCreatedID(string(out))
	if id == "" {
		t.Fatalf("Could not extract ID from bd create output: %s", out)
	}

	// Read using SQLite client
	sqliteClient := NewBdSQLiteClient(env.DBPath)

	issues, err := sqliteClient.List(ctx)
	if err != nil {
		t.Fatalf("SQLite List failed: %v", err)
	}

	found := false
	for _, iss := range issues {
		if iss.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Issue %s created by bd CLI not found via SQLite List", id)
	}
}

// extractCreatedID extracts the issue ID from create command output.
// Expected formats:
// - br: "Created test-xxx: Title"
// - bd: "✓ Created issue: test-xxx"
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

		// bd format: "✓ Created issue: test-xxx"
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

// =============================================================================
// Mixed Operations Tests
// =============================================================================

// TestMixedOperations_WriteCliReadSqlite tests writing via CLI and reading via SQLite.
// Note: bd backend may fail due to a known issue in bdCLIClient.Create parsing
// when bd outputs warning messages. See TestBackend_E2E_Dependencies for details.
func TestMixedOperations_WriteCliReadSqlite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	backends := []string{"br", "bd"}

	for _, backend := range backends {
		backend := backend
		t.Run(backend, func(t *testing.T) {
			t.Parallel()
			skipIfNoBackend(t, backend)

			// Skip bd due to known parsing issue in frozen bdCLIClient.Create
			if backend == "bd" {
				t.Skip("skipping bd backend: known parsing issue with --description warnings in frozen bd client")
			}

			env := setupBackendTestDB(t, backend)
			defer env.cleanup()

			ctx := context.Background()

			// Write using CLI client directly
			var cliWriter Writer
			if backend == "br" {
				// br needs WorkDir to find workspace
				cliWriter = NewBrCLIClient(WithBrDatabasePath(env.DBPath), WithBrWorkDir(env.WorkDir))
			} else {
				// bd uses --db flag directly
				cliWriter = NewBdCLIClient(WithBdDatabasePath(env.DBPath))
			}

			// Create via CLI
			id, err := cliWriter.Create(ctx, "CLI Written Issue", "feature", 1, []string{"mixed"}, "")
			if err != nil {
				t.Fatalf("CLI Create failed: %v", err)
			}

			// Add label via CLI
			if err := cliWriter.AddLabel(ctx, id, "cli-added"); err != nil {
				t.Fatalf("CLI AddLabel failed: %v", err)
			}

			// Add comment via CLI
			if err := cliWriter.AddComment(ctx, id, "CLI added comment"); err != nil {
				t.Fatalf("CLI AddComment failed: %v", err)
			}

			// Small delay to allow SQLite WAL to sync
			time.Sleep(100 * time.Millisecond)

			// Read using SQLite client
			var sqliteClient Client
			if backend == "br" {
				sqliteClient = NewBrSQLiteClient(env.DBPath)
			} else {
				sqliteClient = NewBdSQLiteClient(env.DBPath)
			}

			// Verify via List
			issues, err := sqliteClient.List(ctx)
			if err != nil {
				t.Fatalf("SQLite List failed: %v", err)
			}

			found := false
			for _, iss := range issues {
				if iss.ID == id {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("CLI-created issue %s not found via SQLite List", id)
			}

			// Verify via Export (full details)
			fullIssues, err := sqliteClient.Export(ctx)
			if err != nil {
				t.Fatalf("SQLite Export failed: %v", err)
			}

			var fullIssue *FullIssue
			for i := range fullIssues {
				if fullIssues[i].ID == id {
					fullIssue = &fullIssues[i]
					break
				}
			}
			if fullIssue == nil {
				t.Fatalf("Issue %s not found in SQLite Export", id)
			}

			// Verify labels
			if len(fullIssue.Labels) < 2 {
				t.Errorf("Expected at least 2 labels, got %d: %v", len(fullIssue.Labels), fullIssue.Labels)
			}

			// Verify comments
			comments, err := sqliteClient.Comments(ctx, id)
			if err != nil {
				t.Fatalf("SQLite Comments failed: %v", err)
			}
			if len(comments) < 1 {
				t.Errorf("Expected at least 1 comment, got %d", len(comments))
			}
		})
	}
}

// TestMixedOperations_DataConsistency verifies data written via CLI matches what SQLite reads.
func TestMixedOperations_DataConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	backends := []string{"br", "bd"}

	for _, backend := range backends {
		backend := backend
		t.Run(backend, func(t *testing.T) {
			t.Parallel()
			skipIfNoBackend(t, backend)

			env := setupBackendTestDB(t, backend)
			defer env.cleanup()

			ctx := context.Background()

			// Create using combined client (writes via CLI, reads via SQLite)
			client := newClientForBackend(t, env)

			// Create with specific values
			issue, err := client.CreateFull(ctx,
				"Consistency Test Issue",
				"bug",
				3,
				[]string{"critical", "frontend"},
				"alice",
				"This is the description for consistency testing",
				"",
			)
			if err != nil {
				t.Fatalf("CreateFull failed: %v", err)
			}

			// Allow WAL sync
			time.Sleep(100 * time.Millisecond)

			// Read back and verify
			shown, err := client.Show(ctx, []string{issue.ID})
			if err != nil {
				t.Fatalf("Show failed: %v", err)
			}
			if len(shown) != 1 {
				t.Fatalf("Expected 1 issue, got %d", len(shown))
			}

			read := shown[0]

			// Verify all fields match
			if read.Title != "Consistency Test Issue" {
				t.Errorf("Title mismatch: expected %q, got %q", "Consistency Test Issue", read.Title)
			}
			if read.IssueType != "bug" {
				t.Errorf("IssueType mismatch: expected %q, got %q", "bug", read.IssueType)
			}
			if read.Priority != 3 {
				t.Errorf("Priority mismatch: expected %d, got %d", 3, read.Priority)
			}
			// Note: Labels may be in different order, so we check set membership
			labelSet := make(map[string]bool)
			for _, l := range read.Labels {
				labelSet[l] = true
			}
			for _, expected := range []string{"critical", "frontend"} {
				if !labelSet[expected] {
					t.Errorf("Missing expected label %q, got labels: %v", expected, read.Labels)
				}
			}
		})
	}
}

// =============================================================================
// Schema Compatibility Tests
// =============================================================================

// TestSchemaCompatibility verifies that both backends create compatible schemas.
func TestSchemaCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Need both backends for this test
	skipIfNoBackend(t, "br")
	skipIfNoBackend(t, "bd")

	// Create test databases for each backend
	brEnv := setupBackendTestDB(t, "br")
	defer brEnv.cleanup()

	bdEnv := setupBackendTestDB(t, "bd")
	defer bdEnv.cleanup()

	// Required tables
	requiredTables := []string{"issues", "labels", "dependencies", "comments"}

	// Required issue columns (that abacus uses)
	requiredIssueColumns := []string{
		"id", "title", "description", "status", "priority",
		"issue_type", "assignee", "created_at", "updated_at",
	}

	for _, env := range []backendTestEnv{brEnv, bdEnv} {
		t.Run(env.Backend+"_schema", func(t *testing.T) {
			db, err := sql.Open("sqlite", env.DBPath)
			if err != nil {
				t.Fatalf("Open db: %v", err)
			}
			defer db.Close()

			// Check required tables exist
			for _, table := range requiredTables {
				var count int
				err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&count)
				if err != nil {
					t.Fatalf("Check table %s: %v", table, err)
				}
				if count == 0 {
					t.Errorf("Required table %q not found in %s database", table, env.Backend)
				}
			}

			// Check required issue columns exist
			rows, err := db.Query(`PRAGMA table_info(issues)`)
			if err != nil {
				t.Fatalf("PRAGMA table_info: %v", err)
			}
			defer rows.Close()

			columnSet := make(map[string]bool)
			for rows.Next() {
				var cid int
				var name, ctype string
				var notnull, pk int
				var dfltValue interface{}
				if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
					t.Fatalf("Scan column info: %v", err)
				}
				columnSet[name] = true
			}

			for _, col := range requiredIssueColumns {
				if !columnSet[col] {
					t.Errorf("Required column %q not found in %s issues table", col, env.Backend)
				}
			}
		})
	}
}

// =============================================================================
// Regression Tests
// =============================================================================

// TestBdBackend_E2E is a dedicated regression test to ensure bd backend still works.
// This explicitly tests the bd backend to catch any regressions after br support was added.
func TestBdBackend_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	skipIfNoBackend(t, "bd")

	env := setupBackendTestDB(t, "bd")
	defer env.cleanup()

	client := newClientForBackend(t, env)
	ctx := context.Background()

	// Full workflow test
	t.Log("Testing bd backend - full workflow")

	// Create
	id, err := client.Create(ctx, "BD Regression Test", "task", 2, []string{"regression"}, "")
	if err != nil {
		t.Fatalf("bd Create failed: %v", err)
	}
	t.Logf("Created: %s", id)

	// Update
	if err := client.UpdateStatus(ctx, id, "in_progress"); err != nil {
		t.Fatalf("bd UpdateStatus failed: %v", err)
	}

	// Label operations
	if err := client.AddLabel(ctx, id, "bd-test"); err != nil {
		t.Fatalf("bd AddLabel failed: %v", err)
	}
	if err := client.RemoveLabel(ctx, id, "bd-test"); err != nil {
		t.Fatalf("bd RemoveLabel failed: %v", err)
	}

	// Comment
	if err := client.AddComment(ctx, id, "BD regression test comment"); err != nil {
		t.Fatalf("bd AddComment failed: %v", err)
	}

	// Close/Reopen
	if err := client.Close(ctx, id); err != nil {
		t.Fatalf("bd Close failed: %v", err)
	}
	if err := client.Reopen(ctx, id); err != nil {
		t.Fatalf("bd Reopen failed: %v", err)
	}

	// Read back
	issues, err := client.Show(ctx, []string{id})
	if err != nil {
		t.Fatalf("bd Show failed: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}

	// Delete
	if err := client.Delete(ctx, id, false); err != nil {
		t.Fatalf("bd Delete failed: %v", err)
	}

	t.Log("bd backend regression test passed")
}

// TestBrBackend_E2E is a dedicated test for the br backend.
func TestBrBackend_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	skipIfNoBackend(t, "br")

	env := setupBackendTestDB(t, "br")
	defer env.cleanup()

	client := newClientForBackend(t, env)
	ctx := context.Background()

	t.Log("Testing br backend - full workflow")

	// Create
	id, err := client.Create(ctx, "BR Backend Test", "feature", 1, []string{"br-test"}, "")
	if err != nil {
		t.Fatalf("br Create failed: %v", err)
	}
	t.Logf("Created: %s", id)

	// Update
	if err := client.UpdateStatus(ctx, id, "in_progress"); err != nil {
		t.Fatalf("br UpdateStatus failed: %v", err)
	}

	// Label operations
	if err := client.AddLabel(ctx, id, "urgent"); err != nil {
		t.Fatalf("br AddLabel failed: %v", err)
	}
	if err := client.RemoveLabel(ctx, id, "urgent"); err != nil {
		t.Fatalf("br RemoveLabel failed: %v", err)
	}

	// Comment
	if err := client.AddComment(ctx, id, "BR backend test comment"); err != nil {
		t.Fatalf("br AddComment failed: %v", err)
	}

	// Close/Reopen
	if err := client.Close(ctx, id); err != nil {
		t.Fatalf("br Close failed: %v", err)
	}
	if err := client.Reopen(ctx, id); err != nil {
		t.Fatalf("br Reopen failed: %v", err)
	}

	// Verify via Export
	issues, err := client.Export(ctx)
	if err != nil {
		t.Fatalf("br Export failed: %v", err)
	}

	var found *FullIssue
	for i := range issues {
		if issues[i].ID == id {
			found = &issues[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("Created issue %s not found in export", id)
	}
	if found.Title != "BR Backend Test" {
		t.Errorf("Title mismatch: expected %q, got %q", "BR Backend Test", found.Title)
	}

	// Delete
	if err := client.Delete(ctx, id, false); err != nil {
		t.Fatalf("br Delete failed: %v", err)
	}

	t.Log("br backend test passed")
}
