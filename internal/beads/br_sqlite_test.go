package beads

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

// testBrDB creates a test database with the br schema and returns the path.
// The caller should clean up with t.TempDir() or defer cleanup.
func testBrDB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Create br schema (minimal - only tables/columns used by brSQLiteClient)
	schema := `
		CREATE TABLE issues (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			design TEXT NOT NULL DEFAULT '',
			acceptance_criteria TEXT NOT NULL DEFAULT '',
			notes TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL,
			priority INTEGER NOT NULL,
			issue_type TEXT NOT NULL,
			assignee TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			closed_at TEXT,
			external_ref TEXT,
			deleted_at TEXT
		);

		CREATE TABLE labels (
			issue_id TEXT NOT NULL,
			label TEXT NOT NULL,
			PRIMARY KEY (issue_id, label)
		);

		CREATE TABLE dependencies (
			issue_id TEXT NOT NULL,
			depends_on_id TEXT NOT NULL,
			type TEXT NOT NULL,
			PRIMARY KEY (issue_id, depends_on_id)
		);

		CREATE TABLE comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			issue_id TEXT NOT NULL,
			author TEXT NOT NULL,
			text TEXT NOT NULL,
			created_at TEXT NOT NULL
		);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return dbPath
}

// seedTestData populates the test database with sample data.
func seedTestData(t *testing.T, dbPath string) {
	t.Helper()

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Insert test issues
	issues := []struct {
		id, title, description, status, issueType, createdAt, updatedAt string
		priority                                                        int
	}{
		{"ab-001", "First Issue", "Description 1", "open", "task", "2025-01-01T00:00:00Z", "2025-01-01T00:00:00Z", 2},
		{"ab-002", "Second Issue", "Description 2", "in_progress", "feature", "2025-01-02T00:00:00Z", "2025-01-02T00:00:00Z", 1},
		{"ab-003", "Closed Issue", "Description 3", "closed", "bug", "2025-01-03T00:00:00Z", "2025-01-03T00:00:00Z", 3},
	}

	for _, iss := range issues {
		_, err := db.Exec(`
			INSERT INTO issues (id, title, description, status, priority, issue_type, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, iss.id, iss.title, iss.description, iss.status, iss.priority, iss.issueType, iss.createdAt, iss.updatedAt)
		if err != nil {
			t.Fatalf("insert issue %s: %v", iss.id, err)
		}
	}

	// Insert labels
	labels := []struct{ issueID, label string }{
		{"ab-001", "backend"},
		{"ab-001", "urgent"},
		{"ab-002", "frontend"},
	}
	for _, l := range labels {
		if _, err := db.Exec(`INSERT INTO labels (issue_id, label) VALUES (?, ?)`, l.issueID, l.label); err != nil {
			t.Fatalf("insert label: %v", err)
		}
	}

	// Insert dependencies
	deps := []struct{ issueID, dependsOnID, depType string }{
		{"ab-002", "ab-001", "blocks"},
		{"ab-003", "ab-001", "parent-child"},
	}
	for _, d := range deps {
		if _, err := db.Exec(`INSERT INTO dependencies (issue_id, depends_on_id, type) VALUES (?, ?, ?)`, d.issueID, d.dependsOnID, d.depType); err != nil {
			t.Fatalf("insert dependency: %v", err)
		}
	}

	// Insert comments
	comments := []struct{ issueID, author, text, createdAt string }{
		{"ab-001", "Alice", "First comment", "2025-01-01T10:00:00Z"},
		{"ab-001", "Bob", "Second comment", "2025-01-01T11:00:00Z"},
		{"ab-002", "Charlie", "Another comment", "2025-01-02T10:00:00Z"},
	}
	for _, c := range comments {
		if _, err := db.Exec(`INSERT INTO comments (issue_id, author, text, created_at) VALUES (?, ?, ?, ?)`, c.issueID, c.author, c.text, c.createdAt); err != nil {
			t.Fatalf("insert comment: %v", err)
		}
	}
}

func TestBrSQLiteClient_List(t *testing.T) {
	t.Parallel()

	dbPath := testBrDB(t)
	seedTestData(t, dbPath)

	client := NewBrSQLiteClient(dbPath)
	ctx := context.Background()

	issues, err := client.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(issues) != 3 {
		t.Errorf("expected 3 issues, got %d", len(issues))
	}

	// Verify order by created_at
	expected := []string{"ab-001", "ab-002", "ab-003"}
	for i, iss := range issues {
		if iss.ID != expected[i] {
			t.Errorf("expected issue[%d].ID = %q, got %q", i, expected[i], iss.ID)
		}
	}
}

func TestBrSQLiteClient_List_ExcludesDeleted(t *testing.T) {
	t.Parallel()

	dbPath := testBrDB(t)
	seedTestData(t, dbPath)

	// Mark one issue as deleted
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if _, err := db.Exec(`UPDATE issues SET deleted_at = '2025-01-05T00:00:00Z' WHERE id = 'ab-002'`); err != nil {
		t.Fatalf("mark deleted: %v", err)
	}
	_ = db.Close()

	client := NewBrSQLiteClient(dbPath)
	ctx := context.Background()

	issues, err := client.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(issues) != 2 {
		t.Errorf("expected 2 issues (excluding deleted), got %d", len(issues))
	}

	for _, iss := range issues {
		if iss.ID == "ab-002" {
			t.Error("deleted issue ab-002 should not appear in list")
		}
	}
}

func TestBrSQLiteClient_List_ExcludesTombstone(t *testing.T) {
	t.Parallel()

	dbPath := testBrDB(t)
	seedTestData(t, dbPath)

	// Mark one issue as tombstone
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if _, err := db.Exec(`UPDATE issues SET status = 'tombstone' WHERE id = 'ab-001'`); err != nil {
		t.Fatalf("mark tombstone: %v", err)
	}
	_ = db.Close()

	client := NewBrSQLiteClient(dbPath)
	ctx := context.Background()

	issues, err := client.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(issues) != 2 {
		t.Errorf("expected 2 issues (excluding tombstone), got %d", len(issues))
	}

	for _, iss := range issues {
		if iss.ID == "ab-001" {
			t.Error("tombstone issue ab-001 should not appear in list")
		}
	}
}

func TestBrSQLiteClient_Export(t *testing.T) {
	t.Parallel()

	dbPath := testBrDB(t)
	seedTestData(t, dbPath)

	client := NewBrSQLiteClient(dbPath)
	ctx := context.Background()

	issues, err := client.Export(ctx)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}

	if len(issues) != 3 {
		t.Fatalf("expected 3 issues, got %d", len(issues))
	}

	// Check first issue has all fields populated
	iss := issues[0]
	if iss.ID != "ab-001" {
		t.Errorf("expected first issue ID = ab-001, got %s", iss.ID)
	}
	if iss.Title != "First Issue" {
		t.Errorf("expected title = 'First Issue', got %q", iss.Title)
	}
	if iss.Description != "Description 1" {
		t.Errorf("expected description = 'Description 1', got %q", iss.Description)
	}
	if iss.Status != "open" {
		t.Errorf("expected status = 'open', got %q", iss.Status)
	}
	if iss.Priority != 2 {
		t.Errorf("expected priority = 2, got %d", iss.Priority)
	}

	// Check labels loaded
	if len(iss.Labels) != 2 {
		t.Errorf("expected 2 labels for ab-001, got %d", len(iss.Labels))
	}

	// Check dependents loaded (ab-002 and ab-003 depend on ab-001)
	if len(iss.Dependents) != 2 {
		t.Errorf("expected 2 dependents for ab-001, got %d", len(iss.Dependents))
	}

	// Check comments loaded
	if len(iss.Comments) != 2 {
		t.Errorf("expected 2 comments for ab-001, got %d", len(iss.Comments))
	}
}

func TestBrSQLiteClient_Export_LoadsDependencies(t *testing.T) {
	t.Parallel()

	dbPath := testBrDB(t)
	seedTestData(t, dbPath)

	client := NewBrSQLiteClient(dbPath)
	ctx := context.Background()

	issues, err := client.Export(ctx)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}

	// Find ab-002 which has a dependency on ab-001
	var ab002 *FullIssue
	for i := range issues {
		if issues[i].ID == "ab-002" {
			ab002 = &issues[i]
			break
		}
	}
	if ab002 == nil {
		t.Fatal("ab-002 not found in export")
	}

	if len(ab002.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency for ab-002, got %d", len(ab002.Dependencies))
	}

	dep := ab002.Dependencies[0]
	if dep.TargetID != "ab-001" {
		t.Errorf("expected dependency target = ab-001, got %s", dep.TargetID)
	}
	if dep.Type != "blocks" {
		t.Errorf("expected dependency type = blocks, got %s", dep.Type)
	}
}

func TestBrSQLiteClient_Show(t *testing.T) {
	t.Parallel()

	dbPath := testBrDB(t)
	seedTestData(t, dbPath)

	client := NewBrSQLiteClient(dbPath)
	ctx := context.Background()

	// Show single issue
	issues, err := client.Show(ctx, []string{"ab-001"})
	if err != nil {
		t.Fatalf("Show: %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}

	if issues[0].ID != "ab-001" {
		t.Errorf("expected ID = ab-001, got %s", issues[0].ID)
	}
}

func TestBrSQLiteClient_Show_MultipleIDs(t *testing.T) {
	t.Parallel()

	dbPath := testBrDB(t)
	seedTestData(t, dbPath)

	client := NewBrSQLiteClient(dbPath)
	ctx := context.Background()

	issues, err := client.Show(ctx, []string{"ab-001", "ab-003"})
	if err != nil {
		t.Fatalf("Show: %v", err)
	}

	if len(issues) != 2 {
		t.Errorf("expected 2 issues, got %d", len(issues))
	}
}

func TestBrSQLiteClient_Show_EmptyIDs(t *testing.T) {
	t.Parallel()

	dbPath := testBrDB(t)
	seedTestData(t, dbPath)

	client := NewBrSQLiteClient(dbPath)
	ctx := context.Background()

	issues, err := client.Show(ctx, []string{})
	if err != nil {
		t.Fatalf("Show: %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("expected 0 issues for empty IDs, got %d", len(issues))
	}
}

func TestBrSQLiteClient_Show_NonexistentID(t *testing.T) {
	t.Parallel()

	dbPath := testBrDB(t)
	seedTestData(t, dbPath)

	client := NewBrSQLiteClient(dbPath)
	ctx := context.Background()

	issues, err := client.Show(ctx, []string{"ab-nonexistent"})
	if err != nil {
		t.Fatalf("Show: %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("expected 0 issues for nonexistent ID, got %d", len(issues))
	}
}

func TestBrSQLiteClient_Comments(t *testing.T) {
	t.Parallel()

	dbPath := testBrDB(t)
	seedTestData(t, dbPath)

	client := NewBrSQLiteClient(dbPath)
	ctx := context.Background()

	comments, err := client.Comments(ctx, "ab-001")
	if err != nil {
		t.Fatalf("Comments: %v", err)
	}

	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}

	// Verify order by created_at
	if comments[0].Author != "Alice" {
		t.Errorf("expected first comment author = Alice, got %s", comments[0].Author)
	}
	if comments[1].Author != "Bob" {
		t.Errorf("expected second comment author = Bob, got %s", comments[1].Author)
	}
}

func TestBrSQLiteClient_Comments_NoComments(t *testing.T) {
	t.Parallel()

	dbPath := testBrDB(t)
	seedTestData(t, dbPath)

	client := NewBrSQLiteClient(dbPath)
	ctx := context.Background()

	comments, err := client.Comments(ctx, "ab-003")
	if err != nil {
		t.Fatalf("Comments: %v", err)
	}

	if comments == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(comments) != 0 {
		t.Errorf("expected 0 comments, got %d", len(comments))
	}
}

func TestBrSQLiteClient_List_EmptyDB(t *testing.T) {
	t.Parallel()

	dbPath := testBrDB(t)
	// Don't seed data - empty database

	client := NewBrSQLiteClient(dbPath)
	ctx := context.Background()

	issues, err := client.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("expected empty slice for empty db, got %v", issues)
	}
}

func TestBrSQLiteClient_Export_EmptyDB(t *testing.T) {
	t.Parallel()

	dbPath := testBrDB(t)
	// Don't seed data - empty database

	client := NewBrSQLiteClient(dbPath)
	ctx := context.Background()

	issues, err := client.Export(ctx)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("expected 0 issues for empty db, got %d", len(issues))
	}
}

// Test Writer delegation - verify brSQLiteClient delegates to brCLIClient

func TestBrSQLiteClient_WriterDelegation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebr.sh")

	// Script logs args for write operations
	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	dbPath := testBrDB(t)
	client := NewBrSQLiteClient(dbPath, WithBrBinaryPath(script))

	ctx := context.Background()

	// Test UpdateStatus delegation
	if err := client.UpdateStatus(ctx, "ab-test", "in_progress"); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	// Verify the CLI was called
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected CLI to be called for UpdateStatus")
	}
}

func TestBrSQLiteClient_Close_Delegation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebr.sh")

	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	dbPath := testBrDB(t)
	client := NewBrSQLiteClient(dbPath, WithBrBinaryPath(script))

	ctx := context.Background()
	if err := client.Close(ctx, "ab-close"); err != nil {
		t.Fatalf("Close: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	args := string(data)
	if !brContainsLine(args, "close ab-close") {
		t.Errorf("expected 'close ab-close', got: %q", args)
	}
}

func TestBrSQLiteClient_AddLabel_Delegation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebr.sh")

	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	dbPath := testBrDB(t)
	client := NewBrSQLiteClient(dbPath, WithBrBinaryPath(script))

	ctx := context.Background()
	if err := client.AddLabel(ctx, "ab-label", "urgent"); err != nil {
		t.Fatalf("AddLabel: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	args := string(data)
	if !brContainsLine(args, "label add ab-label urgent") {
		t.Errorf("expected 'label add ab-label urgent', got: %q", args)
	}
}

// brContainsLine checks if any line in output contains the expected substring.
func brContainsLine(output, expected string) bool {
	for _, line := range brSplitLines(output) {
		if line == expected || brContains(line, expected) {
			return true
		}
	}
	return false
}

func brSplitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func brContains(s, substr string) bool {
	return len(s) >= len(substr) && brFindSubstring(s, substr) >= 0
}

func brFindSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestNewBrSQLiteClient_PanicsOnEmptyPath(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty dbPath")
		}
	}()

	_ = NewBrSQLiteClient("")
}

func TestBrSQLiteClient_InvalidDB(t *testing.T) {
	t.Parallel()

	client := NewBrSQLiteClient("/nonexistent/path/test.db")
	ctx := context.Background()

	_, err := client.List(ctx)
	if err == nil {
		t.Error("expected error for invalid db path")
	}
}
