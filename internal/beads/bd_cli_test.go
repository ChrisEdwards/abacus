package beads

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeTestScript writes a shell script to the given path with proper file sync.
// This avoids flaky tests caused by filesystem race conditions where the script
// file isn't fully written/synced before exec.CommandContext tries to execute it.
func writeTestScript(t *testing.T, path, content string) {
	t.Helper()

	// Create file with executable permissions from the start
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		t.Fatalf("create script file: %v", err)
	}

	if _, err := f.WriteString(content); err != nil {
		f.Close()
		t.Fatalf("write script content: %v", err)
	}

	if err := f.Sync(); err != nil {
		f.Close()
		t.Fatalf("sync script file: %v", err)
	}

	if err := f.Close(); err != nil {
		t.Fatalf("close script file: %v", err)
	}

	// Sync parent directory to ensure file metadata is persisted
	dir, err := os.Open(filepath.Dir(path))
	if err == nil {
		dir.Sync()
		dir.Close()
	}

	// Verify file is executable
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat script file: %v", err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("script file not executable: %v", info.Mode())
	}

	// Small delay for CI filesystem propagation.
	// Even with file and directory sync, some CI environments (especially
	// containerized Ubuntu runners) need a brief pause before exec can
	// reliably see newly created files.
	time.Sleep(10 * time.Millisecond)
}

func TestCLIClientAppliesDatabasePath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebd.sh")

	// Script logs args for write operations (close, reopen)
	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	dbPath := "/tmp/custom.db"
	client := NewBdCLIClient(
		WithBdBinaryPath(script),
		WithBdDatabasePath(dbPath),
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

func TestCLIClient_CreateFull_ReturnsFullIssue(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	script := filepath.Join(dir, "fakebd.sh")

	// Fake bd script that returns valid JSON
	jsonResponse := `{"id":"ab-test","title":"Test Issue","description":"Test desc","status":"open","priority":2,"issue_type":"task","created_at":"2025-12-01T00:00:00Z","updated_at":"2025-12-01T00:00:00Z","labels":["test"]}`
	scriptBody := "#!/bin/sh\n" +
		"echo '" + jsonResponse + "'\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBdCLIClient(WithBdBinaryPath(script))

	ctx := context.Background()
	issue, err := client.CreateFull(ctx, "Test Issue", "task", 2, []string{"test"}, "alice", "", "")
	if err != nil {
		t.Fatalf("CreateFull: %v", err)
	}

	if issue.ID != "ab-test" {
		t.Errorf("expected ID %q, got %q", "ab-test", issue.ID)
	}
	if issue.Title != "Test Issue" {
		t.Errorf("expected Title %q, got %q", "Test Issue", issue.Title)
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
	if len(issue.Labels) != 1 || issue.Labels[0] != "test" {
		t.Errorf("expected Labels [%q], got %v", "test", issue.Labels)
	}
}

func TestCLIClient_CreateFull_HandlesInvalidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	script := filepath.Join(dir, "fakebd.sh")

	// Fake bd script that returns malformed JSON
	scriptBody := "#!/bin/sh\n" +
		"echo 'not valid json{{{'\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBdCLIClient(WithBdBinaryPath(script))

	ctx := context.Background()
	_, err := client.CreateFull(ctx, "Test Issue", "task", 2, nil, "", "", "")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}

	// Error should contain snippet of bad output
	// extractJSON now rejects invalid JSON (returns nil), so error is "no JSON found"
	if !strings.Contains(err.Error(), "no JSON found") {
		t.Errorf("expected error to mention no JSON found, got: %v", err)
	}
	if !strings.Contains(err.Error(), "not valid json") {
		t.Errorf("expected error to include output snippet, got: %v", err)
	}
}

func TestCLIClient_CreateFull_PassesJSONFlag(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebd.sh")

	// Fake bd script that logs arguments and returns minimal JSON
	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"echo '{\"id\":\"ab-123\",\"title\":\"Test\",\"status\":\"open\",\"priority\":2,\"issue_type\":\"task\"}'\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBdCLIClient(WithBdBinaryPath(script))

	ctx := context.Background()
	_, err := client.CreateFull(ctx, "Test Title", "feature", 3, nil, "", "", "")
	if err != nil {
		t.Fatalf("CreateFull: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	args := strings.TrimSpace(string(data))
	if !strings.Contains(args, "--json") {
		t.Errorf("expected args to include --json, got: %q", args)
	}
	if !strings.Contains(args, "--title Test Title") {
		t.Errorf("expected args to include title, got: %q", args)
	}
	if !strings.Contains(args, "--type feature") {
		t.Errorf("expected args to include type, got: %q", args)
	}
	if !strings.Contains(args, "--priority 3") {
		t.Errorf("expected args to include priority, got: %q", args)
	}
}

func TestCLIClient_CreateFull_WithParent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebd.sh")

	// Fake bd script that logs arguments and handles both create and dep commands
	// For create: return JSON (without dotted ID)
	// For dep: just succeed
	scriptBody := `#!/bin/sh
echo "$@" >> ` + logFile + `
if echo "$@" | grep -q "^create"; then
  echo '{"id":"ab-xyz123","title":"Child","status":"open","priority":2,"issue_type":"task"}'
fi
exit 0
`
	writeTestScript(t, script, scriptBody)

	client := NewBdCLIClient(WithBdBinaryPath(script))

	ctx := context.Background()
	issue, err := client.CreateFull(ctx, "Child Task", "task", 2, nil, "", "", "ab-parent")
	if err != nil {
		t.Fatalf("CreateFull: %v", err)
	}

	// Verify the new ID is NOT dotted (not ab-parent.1)
	if strings.Contains(issue.ID, ".") {
		t.Errorf("expected non-dotted ID, got: %s", issue.ID)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	args := string(data)

	// Verify --parent is NOT passed to create command
	lines := strings.Split(strings.TrimSpace(args), "\n")
	if len(lines) < 1 {
		t.Fatalf("expected at least 1 command logged")
	}
	createLine := lines[0]
	if strings.Contains(createLine, "--parent") {
		t.Errorf("expected create args to NOT include --parent, got: %q", createLine)
	}

	// Verify dep add was called with parent-child type
	if len(lines) < 2 {
		t.Fatalf("expected dep add command to be called, only got: %v", lines)
	}
	depLine := lines[1]
	if !strings.Contains(depLine, "dep add ab-xyz123 ab-parent --type parent-child") {
		t.Errorf("expected dep add with parent-child, got: %q", depLine)
	}
}

func TestCLIClient_CreateFull_OptionalParameters(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebd.sh")

	// Fake bd script that logs arguments and returns JSON with labels and assignee
	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"echo '{\"id\":\"ab-full\",\"title\":\"Full\",\"status\":\"open\",\"priority\":1,\"issue_type\":\"bug\",\"labels\":[\"urgent\",\"backend\"],\"assignee\":\"bob\"}'\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBdCLIClient(WithBdBinaryPath(script))

	ctx := context.Background()
	issue, err := client.CreateFull(ctx, "Bug Fix", "bug", 1, []string{"urgent", "backend"}, "bob", "", "")
	if err != nil {
		t.Fatalf("CreateFull: %v", err)
	}

	// Verify returned issue has the optional fields
	if len(issue.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(issue.Labels))
	}
	if issue.Assignee != "bob" {
		t.Errorf("expected Assignee %q, got %q", "bob", issue.Assignee)
	}

	// Verify arguments were passed correctly
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	args := strings.TrimSpace(string(data))
	if !strings.Contains(args, "--labels urgent,backend") {
		t.Errorf("expected args to include --labels, got: %q", args)
	}
	if !strings.Contains(args, "--assignee bob") {
		t.Errorf("expected args to include --assignee, got: %q", args)
	}
}

func TestExtractJSON(t *testing.T) {
	t.Run("ExtractsJSONFromCleanOutput", func(t *testing.T) {
		input := []byte(`{"id":"ab-123","title":"Test"}`)
		result := extractJSON(input)
		if string(result) != `{"id":"ab-123","title":"Test"}` {
			t.Errorf("expected clean JSON, got: %s", result)
		}
	})

	t.Run("ExtractsJSONWithPrefix", func(t *testing.T) {
		input := []byte("⚠ Warning: creating in production\n{\"id\":\"ab-123\",\"title\":\"Test\"}")
		result := extractJSON(input)
		if string(result) != `{"id":"ab-123","title":"Test"}` {
			t.Errorf("expected JSON without prefix, got: %s", result)
		}
	})

	t.Run("ExtractsJSONWithSuffix", func(t *testing.T) {
		input := []byte(`{"id":"ab-123","title":"Test"}` + "\nsome trailing text")
		result := extractJSON(input)
		if string(result) != `{"id":"ab-123","title":"Test"}` {
			t.Errorf("expected JSON without suffix, got: %s", result)
		}
	})

	t.Run("ExtractsNestedJSON", func(t *testing.T) {
		input := []byte(`prefix{"id":"ab-123","nested":{"key":"value"}}suffix`)
		result := extractJSON(input)
		if string(result) != `{"id":"ab-123","nested":{"key":"value"}}` {
			t.Errorf("expected nested JSON, got: %s", result)
		}
	})

	t.Run("ReturnsNilForNoJSON", func(t *testing.T) {
		input := []byte("no json here")
		result := extractJSON(input)
		if result != nil {
			t.Errorf("expected nil for no JSON, got: %s", result)
		}
	})

	t.Run("HandlesEscapedQuotesInStrings", func(t *testing.T) {
		input := []byte(`{"msg":"say \"hello\" world"}`)
		result := extractJSON(input)
		if string(result) != `{"msg":"say \"hello\" world"}` {
			t.Errorf("expected JSON with escaped quotes, got: %s", result)
		}
	})

	t.Run("HandlesBracesInsideStrings", func(t *testing.T) {
		input := []byte(`{"msg":"use { and } carefully","id":"ab-123"}`)
		result := extractJSON(input)
		if string(result) != `{"msg":"use { and } carefully","id":"ab-123"}` {
			t.Errorf("expected JSON with braces in string, got: %s", result)
		}
	})

	t.Run("HandlesBracesInWarningBeforeJSON", func(t *testing.T) {
		input := []byte("Warning: {config} file missing\n" + `{"id":"ab-123","status":"open"}`)
		result := extractJSON(input)
		if string(result) != `{"id":"ab-123","status":"open"}` {
			t.Errorf("expected JSON after warning with braces, got: %s", result)
		}
	})

	t.Run("HandlesNestedBracesInStrings", func(t *testing.T) {
		input := []byte(`{"data":{"nested":"{\"inner\":\"value\"}"}}`)
		result := extractJSON(input)
		if string(result) != `{"data":{"nested":"{\"inner\":\"value\"}"}}` {
			t.Errorf("expected nested JSON string, got: %s", result)
		}
	})

	t.Run("HandlesBackslashEscapeSequences", func(t *testing.T) {
		input := []byte(`{"path":"C:\\Users\\test\\file.txt"}`)
		result := extractJSON(input)
		if string(result) != `{"path":"C:\\Users\\test\\file.txt"}` {
			t.Errorf("expected JSON with backslash escapes, got: %s", result)
		}
	})
}

func TestCLIClient_CreateFull_HandlesOutputWithPrefix(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	script := filepath.Join(dir, "fakebd.sh")

	// Fake bd script that outputs warning before JSON (simulates production db warning)
	scriptBody := "#!/bin/sh\n" +
		"echo '⚠ Creating issue with Test prefix in production database.'\n" +
		"echo '{\"id\":\"ab-prefix\",\"title\":\"Test\",\"status\":\"open\",\"priority\":2,\"issue_type\":\"task\"}'\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBdCLIClient(WithBdBinaryPath(script))

	ctx := context.Background()
	issue, err := client.CreateFull(ctx, "Test Title", "task", 2, nil, "", "", "")
	if err != nil {
		t.Fatalf("CreateFull should handle output with prefix: %v", err)
	}

	if issue.ID != "ab-prefix" {
		t.Errorf("expected ID 'ab-prefix', got %q", issue.ID)
	}
}

// NOTE: CLI Export tests removed - Export is not implemented on CLI client.
// In production, Export goes through SQLite client directly.
// The CLI read methods (List, Show, Export, Comments) are stub implementations
// that return errors, as all read operations should use SQLite client.

// TestBdCLIClient_UpdateFull_ClearAssignee verifies that an empty assignee is
// passed to the CLI to allow clearing it. Previously, empty assignee was omitted
// entirely, which meant selecting "Unassigned" in the UI had no effect.
func TestBdCLIClient_UpdateFull_ClearAssignee(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebd.sh")

	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"exit 0\n"
	writeTestScript(t, script, scriptBody)

	client := NewBdCLIClient(WithBdBinaryPath(script))

	ctx := context.Background()
	// Pass empty assignee - this should still include --assignee flag to clear it
	err := client.UpdateFull(ctx, "ab-test", "Title", "task", 2, nil, "", "desc")
	if err != nil {
		t.Fatalf("UpdateFull: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	args := string(data)

	// Verify --assignee is present even when empty (to clear the assignee)
	if !strings.Contains(args, "--assignee") {
		t.Errorf("expected --assignee flag even with empty value to clear assignee, got: %q", args)
	}
}
