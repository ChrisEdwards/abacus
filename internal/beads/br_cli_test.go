package beads

import (
	"context"
	"os"
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

	// Fake br script that returns "Created ab-test123"
	scriptBody := "#!/bin/sh\n" +
		"echo 'Created ab-test123'\n" +
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

	// Script logs args and returns ID
	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"echo 'Created ab-xyz'\n" +
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
