package beads

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrCLIClient_AppliesDatabasePath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebr.sh")

	// Script logs args for write operations (close, reopen)
	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	dbPath := "/tmp/custom.db"
	client := NewBrCLIClient(
		WithBrBinaryPath(script),
		WithBrDatabasePath(dbPath),
	)

	ctx := context.Background()
	// Use write operations to test --db flag is applied
	if err := client.Close(ctx, "ab-123"); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := client.Reopen(ctx, "ab-456"); err != nil {
		t.Fatalf("Reopen: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected two invocations, got %d (%q)", len(lines), lines)
	}

	if !strings.HasPrefix(lines[0], "--db "+dbPath+" close ab-123") {
		t.Fatalf("expected close call to include db override, got %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "--db "+dbPath+" reopen ab-456") {
		t.Fatalf("expected reopen call to include db override, got %q", lines[1])
	}
}

func TestBrCLIClient_Create_ReturnsID(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	script := filepath.Join(dir, "fakebr.sh")

	// Fake br script that returns JSON output (--json flag is now used)
	scriptBody := "#!/bin/sh\n" +
		"echo '{\"id\":\"ab-test123\"}'\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBrCLIClient(WithBrBinaryPath(script))

	ctx := context.Background()
	id, err := client.Create(ctx, "Test Issue", "task", 2, nil, "")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if id != "ab-test123" {
		t.Errorf("expected ID %q, got %q", "ab-test123", id)
	}
}

func TestBrCLIClient_Create_UsesPositionalTitle(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebr.sh")

	// Script logs args and returns JSON (code uses --json flag)
	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"echo '{\"id\":\"ab-xyz\"}'\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBrCLIClient(WithBrBinaryPath(script))

	ctx := context.Background()
	_, err := client.Create(ctx, "My Test Title", "feature", 3, []string{"urgent"}, "alice")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	args := strings.TrimSpace(string(data))
	// Verify positional title (title comes right after "create")
	if !strings.Contains(args, "create My Test Title --type feature") {
		t.Errorf("expected positional title syntax, got: %q", args)
	}
	if !strings.Contains(args, "--labels urgent") {
		t.Errorf("expected --labels flag, got: %q", args)
	}
	if !strings.Contains(args, "--assignee alice") {
		t.Errorf("expected --assignee flag, got: %q", args)
	}
}

func TestBrCLIClient_CreateFull_ReturnsFullIssue(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	script := filepath.Join(dir, "fakebr.sh")

	// Fake br script that returns valid JSON
	jsonResponse := `{"id":"ab-full","title":"Full Issue","description":"Test desc","status":"open","priority":2,"issue_type":"task","created_at":"2025-12-01T00:00:00Z","updated_at":"2025-12-01T00:00:00Z","labels":["test"]}`
	scriptBody := "#!/bin/sh\n" +
		"echo '" + jsonResponse + "'\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBrCLIClient(WithBrBinaryPath(script))

	ctx := context.Background()
	issue, err := client.CreateFull(ctx, "Full Issue", "task", 2, []string{"test"}, "alice", "Test desc", "")
	if err != nil {
		t.Fatalf("CreateFull: %v", err)
	}

	if issue.ID != "ab-full" {
		t.Errorf("expected ID %q, got %q", "ab-full", issue.ID)
	}
	if issue.Title != "Full Issue" {
		t.Errorf("expected Title %q, got %q", "Full Issue", issue.Title)
	}
	if issue.Status != "open" {
		t.Errorf("expected Status %q, got %q", "open", issue.Status)
	}
	if issue.Priority != 2 {
		t.Errorf("expected Priority %d, got %d", 2, issue.Priority)
	}
	if issue.IssueType != "task" {
		t.Errorf("expected IssueType %q, got %q", "task", issue.IssueType)
	}
}

func TestBrCLIClient_CreateFull_HandlesInvalidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	script := filepath.Join(dir, "fakebr.sh")

	// Fake br script that returns malformed JSON
	scriptBody := "#!/bin/sh\n" +
		"echo 'not valid json{{{'\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBrCLIClient(WithBrBinaryPath(script))

	ctx := context.Background()
	_, err := client.CreateFull(ctx, "Test Issue", "task", 2, nil, "", "", "")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}

	// Error should indicate decode failure (extractJSON finds partial JSON starting with '{')
	if !strings.Contains(err.Error(), "decode br create output") {
		t.Errorf("expected error to mention decode failure, got: %v", err)
	}
}

func TestBrCLIClient_CreateFull_HandlesOutputWithPrefix(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	script := filepath.Join(dir, "fakebr.sh")

	// Fake br script that outputs warning before JSON
	scriptBody := "#!/bin/sh\n" +
		"echo 'Warning: creating issue in production database.'\n" +
		"echo '{\"id\":\"ab-prefix\",\"title\":\"Test\",\"status\":\"open\",\"priority\":2,\"issue_type\":\"task\"}'\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBrCLIClient(WithBrBinaryPath(script))

	ctx := context.Background()
	issue, err := client.CreateFull(ctx, "Test Title", "task", 2, nil, "", "", "")
	if err != nil {
		t.Fatalf("CreateFull should handle output with prefix: %v", err)
	}

	if issue.ID != "ab-prefix" {
		t.Errorf("expected ID 'ab-prefix', got %q", issue.ID)
	}
}

func TestBrCLIClient_CreateFull_WithParent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebr.sh")

	// Fake br script that logs arguments and handles both create and dep commands
	scriptBody := `#!/bin/sh
echo "$@" >> ` + logFile + `
if echo "$@" | grep -q "^create"; then
  echo '{"id":"ab-child","title":"Child","status":"open","priority":2,"issue_type":"task"}'
fi
exit 0
`
	writeTestScript(t, script, scriptBody)

	client := NewBrCLIClient(WithBrBinaryPath(script))

	ctx := context.Background()
	issue, err := client.CreateFull(ctx, "Child Task", "task", 2, nil, "", "", "ab-parent")
	if err != nil {
		t.Fatalf("CreateFull: %v", err)
	}

	if issue.ID != "ab-child" {
		t.Errorf("expected ID 'ab-child', got %q", issue.ID)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 commands logged, got: %v", lines)
	}

	// Verify dep add was called with parent-child type
	depLine := lines[1]
	if !strings.Contains(depLine, "dep add ab-child ab-parent --type parent-child") {
		t.Errorf("expected dep add with parent-child, got: %q", depLine)
	}
}

func TestBrCLIClient_Close(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebr.sh")

	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBrCLIClient(WithBrBinaryPath(script))

	ctx := context.Background()
	if err := client.Close(ctx, "ab-close"); err != nil {
		t.Fatalf("Close: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	args := strings.TrimSpace(string(data))
	if args != "close ab-close" {
		t.Errorf("expected 'close ab-close', got: %q", args)
	}
}

func TestBrCLIClient_Reopen(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebr.sh")

	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBrCLIClient(WithBrBinaryPath(script))

	ctx := context.Background()
	if err := client.Reopen(ctx, "ab-reopen"); err != nil {
		t.Fatalf("Reopen: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	args := strings.TrimSpace(string(data))
	if args != "reopen ab-reopen" {
		t.Errorf("expected 'reopen ab-reopen', got: %q", args)
	}
}

func TestBrCLIClient_UpdateStatus(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebr.sh")

	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBrCLIClient(WithBrBinaryPath(script))

	ctx := context.Background()
	if err := client.UpdateStatus(ctx, "ab-status", "in_progress"); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	args := strings.TrimSpace(string(data))
	if !strings.Contains(args, "update ab-status --status=in_progress") {
		t.Errorf("expected update with status flag, got: %q", args)
	}
}

func TestBrCLIClient_AddLabel(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebr.sh")

	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBrCLIClient(WithBrBinaryPath(script))

	ctx := context.Background()
	if err := client.AddLabel(ctx, "ab-label", "urgent"); err != nil {
		t.Fatalf("AddLabel: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	args := strings.TrimSpace(string(data))
	if args != "label add ab-label urgent" {
		t.Errorf("expected 'label add ab-label urgent', got: %q", args)
	}
}

func TestBrCLIClient_RemoveLabel(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebr.sh")

	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBrCLIClient(WithBrBinaryPath(script))

	ctx := context.Background()
	if err := client.RemoveLabel(ctx, "ab-label", "urgent"); err != nil {
		t.Fatalf("RemoveLabel: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	args := strings.TrimSpace(string(data))
	if args != "label remove ab-label urgent" {
		t.Errorf("expected 'label remove ab-label urgent', got: %q", args)
	}
}

func TestBrCLIClient_AddDependency(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebr.sh")

	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBrCLIClient(WithBrBinaryPath(script))

	ctx := context.Background()
	if err := client.AddDependency(ctx, "ab-from", "ab-to", "blocks"); err != nil {
		t.Fatalf("AddDependency: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	args := strings.TrimSpace(string(data))
	if args != "dep add ab-from ab-to --type blocks" {
		t.Errorf("expected 'dep add ab-from ab-to --type blocks', got: %q", args)
	}
}

func TestBrCLIClient_AddDependency_DefaultsToBlocks(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebr.sh")

	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBrCLIClient(WithBrBinaryPath(script))

	ctx := context.Background()
	// Pass empty depType - should default to "blocks"
	if err := client.AddDependency(ctx, "ab-from", "ab-to", ""); err != nil {
		t.Fatalf("AddDependency: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	args := strings.TrimSpace(string(data))
	if !strings.Contains(args, "--type blocks") {
		t.Errorf("expected default type 'blocks', got: %q", args)
	}
}

func TestBrCLIClient_RemoveDependency(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebr.sh")

	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBrCLIClient(WithBrBinaryPath(script))

	ctx := context.Background()
	if err := client.RemoveDependency(ctx, "ab-from", "ab-to", ""); err != nil {
		t.Fatalf("RemoveDependency: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	args := strings.TrimSpace(string(data))
	if args != "dep remove ab-from ab-to" {
		t.Errorf("expected 'dep remove ab-from ab-to', got: %q", args)
	}
}

func TestBrCLIClient_Delete(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebr.sh")

	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBrCLIClient(WithBrBinaryPath(script))

	ctx := context.Background()
	if err := client.Delete(ctx, "ab-delete", false); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	args := strings.TrimSpace(string(data))
	if args != "delete ab-delete --force" {
		t.Errorf("expected 'delete ab-delete --force', got: %q", args)
	}
}

func TestBrCLIClient_Delete_WithCascade(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebr.sh")

	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBrCLIClient(WithBrBinaryPath(script))

	ctx := context.Background()
	if err := client.Delete(ctx, "ab-delete", true); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	args := strings.TrimSpace(string(data))
	if args != "delete ab-delete --force --cascade" {
		t.Errorf("expected 'delete ab-delete --force --cascade', got: %q", args)
	}
}

func TestBrCLIClient_AddComment(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebr.sh")

	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBrCLIClient(WithBrBinaryPath(script))

	ctx := context.Background()
	if err := client.AddComment(ctx, "ab-comment", "This is a test comment"); err != nil {
		t.Fatalf("AddComment: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	args := strings.TrimSpace(string(data))
	if !strings.Contains(args, "comments add ab-comment") {
		t.Errorf("expected 'comments add ab-comment', got: %q", args)
	}
	if !strings.Contains(args, "This is a test comment") {
		t.Errorf("expected comment text in args, got: %q", args)
	}
}

func TestBrCLIClient_UpdateFull(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebr.sh")

	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBrCLIClient(WithBrBinaryPath(script))

	ctx := context.Background()
	err := client.UpdateFull(ctx, "ab-update", "New Title", "feature", 3, []string{"backend", "urgent"}, "bob", "New description")
	if err != nil {
		t.Fatalf("UpdateFull: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	args := strings.TrimSpace(string(data))
	if !strings.Contains(args, "update ab-update") {
		t.Errorf("expected 'update ab-update', got: %q", args)
	}
	if !strings.Contains(args, "--title New Title") {
		t.Errorf("expected --title flag, got: %q", args)
	}
	if !strings.Contains(args, "--description New description") {
		t.Errorf("expected --description flag, got: %q", args)
	}
	if !strings.Contains(args, "--priority 3") {
		t.Errorf("expected --priority flag, got: %q", args)
	}
	if !strings.Contains(args, "--assignee bob") {
		t.Errorf("expected --assignee flag, got: %q", args)
	}
}

// TestBrCLIClient_UpdateFull_CommaSeparatedLabels verifies that multiple labels
// are passed as a single comma-separated --set-labels flag, NOT multiple flags.
// br only accepts one --set-labels flag (unlike bd which accepts multiple).
// See: https://github.com/Dicklesworthstone/beads_rust/issues/17
func TestBrCLIClient_UpdateFull_CommaSeparatedLabels(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebr.sh")

	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBrCLIClient(WithBrBinaryPath(script))

	ctx := context.Background()
	err := client.UpdateFull(ctx, "ab-test", "Title", "task", 2, []string{"backend", "urgent", "api"}, "", "desc")
	if err != nil {
		t.Fatalf("UpdateFull: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	args := string(data)

	// Verify labels are comma-separated in a single flag
	if !strings.Contains(args, "--set-labels backend,urgent,api") {
		t.Errorf("expected comma-separated labels '--set-labels backend,urgent,api', got: %q", args)
	}

	// Verify we DON'T have multiple --set-labels flags (the old broken behavior)
	count := strings.Count(args, "--set-labels")
	if count != 1 {
		t.Errorf("expected exactly 1 --set-labels flag, got %d in: %q", count, args)
	}
}

// Test validation errors
func TestBrCLIClient_ValidationErrors(t *testing.T) {
	t.Parallel()

	client := NewBrCLIClient(WithBrBinaryPath("/nonexistent"))
	ctx := context.Background()

	tests := []struct {
		name string
		fn   func() error
	}{
		{"UpdateStatus empty issueID", func() error { return client.UpdateStatus(ctx, "", "open") }},
		{"UpdateStatus empty status", func() error { return client.UpdateStatus(ctx, "ab-1", "") }},
		{"Close empty issueID", func() error { return client.Close(ctx, "") }},
		{"Reopen empty issueID", func() error { return client.Reopen(ctx, "") }},
		{"AddLabel empty issueID", func() error { return client.AddLabel(ctx, "", "label") }},
		{"AddLabel empty label", func() error { return client.AddLabel(ctx, "ab-1", "") }},
		{"RemoveLabel empty issueID", func() error { return client.RemoveLabel(ctx, "", "label") }},
		{"RemoveLabel empty label", func() error { return client.RemoveLabel(ctx, "ab-1", "") }},
		{"Create empty title", func() error { _, err := client.Create(ctx, "", "task", 2, nil, ""); return err }},
		{"CreateFull empty title", func() error { _, err := client.CreateFull(ctx, "", "task", 2, nil, "", "", ""); return err }},
		{"UpdateFull empty issueID", func() error { return client.UpdateFull(ctx, "", "title", "task", 2, nil, "", "") }},
		{"UpdateFull empty title", func() error { return client.UpdateFull(ctx, "ab-1", "", "task", 2, nil, "", "") }},
		{"AddDependency empty fromID", func() error { return client.AddDependency(ctx, "", "ab-to", "blocks") }},
		{"AddDependency empty toID", func() error { return client.AddDependency(ctx, "ab-from", "", "blocks") }},
		{"RemoveDependency empty fromID", func() error { return client.RemoveDependency(ctx, "", "ab-to", "") }},
		{"RemoveDependency empty toID", func() error { return client.RemoveDependency(ctx, "ab-from", "", "") }},
		{"Delete empty issueID", func() error { return client.Delete(ctx, "", false) }},
		{"AddComment empty issueID", func() error { return client.AddComment(ctx, "", "text") }},
		{"AddComment empty text", func() error { return client.AddComment(ctx, "ab-1", "") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err == nil {
				t.Errorf("expected validation error, got nil")
			}
		})
	}
}

// =============================================================================
// Integration Tests - require real br binary
// =============================================================================

// skipIfNoBr skips the test if the br binary is not available.
func skipIfNoBr(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("br")
	if err != nil {
		t.Skip("br binary not found, skipping integration test")
	}
	return path
}

// brTestEnv holds the test environment for br integration tests.
type brTestEnv struct {
	DBPath  string
	WorkDir string
	cleanup func()
}

// setupBrTestDB creates a temp directory with an initialized br database.
// Returns the test environment with dbPath, workDir, and a cleanup function.
func setupBrTestDB(t *testing.T) brTestEnv {
	t.Helper()

	dir := t.TempDir()
	// br init ignores --db flag and always creates .beads/beads.db in cwd
	beadsDir := filepath.Join(dir, ".beads")
	dbPath := filepath.Join(beadsDir, "beads.db")

	// Initialize br database with a test prefix.
	// We run init from the temp dir to avoid finding the repo's existing db.
	cmd := exec.Command("br", "init", "--prefix", "test")
	cmd.Dir = dir // Run from temp dir to create .beads/ there
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("br init failed: %v\nOutput: %s", err, out)
	}

	// Verify the db was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatalf("br init did not create expected database at %s", dbPath)
	}

	return brTestEnv{
		DBPath:  dbPath,
		WorkDir: dir, // br needs to run from a dir containing .beads/
		cleanup: func() {
			// TempDir cleanup is automatic
		},
	}
}

func TestBrCLIClient_Integration_CreateAndClose(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	skipIfNoBr(t)

	env := setupBrTestDB(t)
	defer env.cleanup()

	client := NewBrCLIClient(WithBrDatabasePath(env.DBPath), WithBrWorkDir(env.WorkDir))
	ctx := context.Background()

	// Create an issue
	id, err := client.Create(ctx, "Integration Test Issue", "task", 2, []string{"test-label"}, "")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if id == "" {
		t.Fatal("Create returned empty ID")
	}
	// Note: ID prefix depends on br's init prefix setting; we just verify an ID was returned

	// Close the issue
	if err := client.Close(ctx, id); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Reopen the issue
	if err := client.Reopen(ctx, id); err != nil {
		t.Fatalf("Reopen failed: %v", err)
	}
}

func TestBrCLIClient_Integration_CreateFull(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	skipIfNoBr(t)

	env := setupBrTestDB(t)
	defer env.cleanup()

	client := NewBrCLIClient(WithBrDatabasePath(env.DBPath), WithBrWorkDir(env.WorkDir))
	ctx := context.Background()

	// Create with full options
	issue, err := client.CreateFull(ctx, "Full Integration Issue", "feature", 1,
		[]string{"urgent", "backend"}, "alice", "This is a test description", "")
	if err != nil {
		t.Fatalf("CreateFull failed: %v", err)
	}

	if issue.ID == "" {
		t.Fatal("CreateFull returned empty ID")
	}
	if issue.Title != "Full Integration Issue" {
		t.Errorf("expected title 'Full Integration Issue', got: %s", issue.Title)
	}
	if issue.IssueType != "feature" {
		t.Errorf("expected type 'feature', got: %s", issue.IssueType)
	}
	if issue.Priority != 1 {
		t.Errorf("expected priority 1, got: %d", issue.Priority)
	}
}

func TestBrCLIClient_Integration_Labels(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	skipIfNoBr(t)

	env := setupBrTestDB(t)
	defer env.cleanup()

	client := NewBrCLIClient(WithBrDatabasePath(env.DBPath), WithBrWorkDir(env.WorkDir))
	ctx := context.Background()

	// Create an issue first
	id, err := client.Create(ctx, "Label Test Issue", "task", 2, nil, "")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Add a label
	if err := client.AddLabel(ctx, id, "new-label"); err != nil {
		t.Fatalf("AddLabel failed: %v", err)
	}

	// Remove the label
	if err := client.RemoveLabel(ctx, id, "new-label"); err != nil {
		t.Fatalf("RemoveLabel failed: %v", err)
	}
}

func TestBrCLIClient_Integration_Dependencies(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	skipIfNoBr(t)

	env := setupBrTestDB(t)
	defer env.cleanup()

	client := NewBrCLIClient(WithBrDatabasePath(env.DBPath), WithBrWorkDir(env.WorkDir))
	ctx := context.Background()

	// Create two issues
	id1, err := client.Create(ctx, "Dependency Test Issue 1", "task", 2, nil, "")
	if err != nil {
		t.Fatalf("Create issue 1 failed: %v", err)
	}

	id2, err := client.Create(ctx, "Dependency Test Issue 2", "task", 2, nil, "")
	if err != nil {
		t.Fatalf("Create issue 2 failed: %v", err)
	}

	// Add dependency: id1 blocks id2
	if err := client.AddDependency(ctx, id2, id1, "blocks"); err != nil {
		t.Fatalf("AddDependency failed: %v", err)
	}

	// Remove the dependency
	if err := client.RemoveDependency(ctx, id2, id1, ""); err != nil {
		t.Fatalf("RemoveDependency failed: %v", err)
	}
}

func TestBrCLIClient_Integration_Comments(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	skipIfNoBr(t)

	env := setupBrTestDB(t)
	defer env.cleanup()

	client := NewBrCLIClient(WithBrDatabasePath(env.DBPath), WithBrWorkDir(env.WorkDir))
	ctx := context.Background()

	// Create an issue
	id, err := client.Create(ctx, "Comment Test Issue", "task", 2, nil, "")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Add a comment
	if err := client.AddComment(ctx, id, "This is an integration test comment"); err != nil {
		t.Fatalf("AddComment failed: %v", err)
	}
}

func TestBrCLIClient_Integration_UpdateStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	skipIfNoBr(t)

	env := setupBrTestDB(t)
	defer env.cleanup()

	client := NewBrCLIClient(WithBrDatabasePath(env.DBPath), WithBrWorkDir(env.WorkDir))
	ctx := context.Background()

	// Create an issue
	id, err := client.Create(ctx, "Status Update Test", "task", 2, nil, "")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update status to in_progress
	if err := client.UpdateStatus(ctx, id, "in_progress"); err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}
}

func TestBrCLIClient_Integration_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	skipIfNoBr(t)

	env := setupBrTestDB(t)
	defer env.cleanup()

	client := NewBrCLIClient(WithBrDatabasePath(env.DBPath), WithBrWorkDir(env.WorkDir))
	ctx := context.Background()

	// Create an issue
	id, err := client.Create(ctx, "Delete Test Issue", "task", 2, nil, "")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Delete the issue
	if err := client.Delete(ctx, id, false); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestBrCLIClient_Integration_ParentChild(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	skipIfNoBr(t)

	env := setupBrTestDB(t)
	defer env.cleanup()

	client := NewBrCLIClient(WithBrDatabasePath(env.DBPath), WithBrWorkDir(env.WorkDir))
	ctx := context.Background()

	// Create parent issue
	parent, err := client.CreateFull(ctx, "Parent Issue", "epic", 1, nil, "", "Parent description", "")
	if err != nil {
		t.Fatalf("Create parent failed: %v", err)
	}

	// Create child with parent reference
	child, err := client.CreateFull(ctx, "Child Issue", "task", 2, nil, "", "Child description", parent.ID)
	if err != nil {
		t.Fatalf("Create child with parent failed: %v", err)
	}

	if child.ID == "" {
		t.Fatal("Child issue has empty ID")
	}
}
