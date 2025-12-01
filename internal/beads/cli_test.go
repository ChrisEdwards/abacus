package beads

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIClientAppliesDatabasePath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "fakebd.sh")

	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"echo '[]'\n"
	if err := os.WriteFile(script, []byte(scriptBody), 0o755); err != nil {
		t.Fatalf("write fake bd: %v", err)
	}

	dbPath := "/tmp/custom.db"
	client := NewCLIClient(
		WithBinaryPath(script),
		WithDatabasePath(dbPath),
	)

	ctx := context.Background()
	if _, err := client.List(ctx); err != nil {
		t.Fatalf("List: %v", err)
	}
	if _, err := client.Show(ctx, []string{"ab-123"}); err != nil {
		t.Fatalf("Show: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected two invocations, got %d (%q)", len(lines), lines)
	}

	if !strings.HasPrefix(lines[0], "--db "+dbPath+" list") {
		t.Fatalf("expected list call to include db override, got %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "--db "+dbPath+" show ab-123") {
		t.Fatalf("expected show call to include db override, got %q", lines[1])
	}
}

func TestCLIClient_CreateFull_ReturnsFullIssue(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	script := filepath.Join(dir, "fakebd.sh")

	// Fake bd script that returns valid JSON
	jsonResponse := `{"id":"ab-test","title":"Test Issue","description":"Test desc","status":"open","priority":2,"issue_type":"task","created_at":"2025-12-01T00:00:00Z","updated_at":"2025-12-01T00:00:00Z","labels":["test"]}`
	scriptBody := "#!/bin/sh\n" +
		"echo '" + jsonResponse + "'\n"
	if err := os.WriteFile(script, []byte(scriptBody), 0o755); err != nil {
		t.Fatalf("write fake bd: %v", err)
	}

	client := NewCLIClient(WithBinaryPath(script))

	ctx := context.Background()
	issue, err := client.CreateFull(ctx, "Test Issue", "task", 2, []string{"test"}, "alice", "")
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
		"echo 'not valid json{{{'\n"
	if err := os.WriteFile(script, []byte(scriptBody), 0o755); err != nil {
		t.Fatalf("write fake bd: %v", err)
	}

	client := NewCLIClient(WithBinaryPath(script))

	ctx := context.Background()
	_, err := client.CreateFull(ctx, "Test Issue", "task", 2, nil, "", "")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}

	// Error should contain snippet of bad output
	if !strings.Contains(err.Error(), "decode bd create output") {
		t.Errorf("expected error to mention decode failure, got: %v", err)
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
		"echo '{\"id\":\"ab-123\",\"title\":\"Test\",\"status\":\"open\",\"priority\":2,\"issue_type\":\"task\"}'\n"
	if err := os.WriteFile(script, []byte(scriptBody), 0o755); err != nil {
		t.Fatalf("write fake bd: %v", err)
	}

	client := NewCLIClient(WithBinaryPath(script))

	ctx := context.Background()
	_, err := client.CreateFull(ctx, "Test Title", "feature", 3, nil, "", "")
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

	// Fake bd script that logs arguments and returns minimal JSON
	scriptBody := "#!/bin/sh\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		"echo '{\"id\":\"ab-child\",\"title\":\"Child\",\"status\":\"open\",\"priority\":2,\"issue_type\":\"task\"}'\n"
	if err := os.WriteFile(script, []byte(scriptBody), 0o755); err != nil {
		t.Fatalf("write fake bd: %v", err)
	}

	client := NewCLIClient(WithBinaryPath(script))

	ctx := context.Background()
	_, err := client.CreateFull(ctx, "Child Task", "task", 2, nil, "", "ab-parent")
	if err != nil {
		t.Fatalf("CreateFull: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}

	args := strings.TrimSpace(string(data))
	if !strings.Contains(args, "--parent ab-parent") {
		t.Errorf("expected args to include --parent ab-parent, got: %q", args)
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
		"echo '{\"id\":\"ab-full\",\"title\":\"Full\",\"status\":\"open\",\"priority\":1,\"issue_type\":\"bug\",\"labels\":[\"urgent\",\"backend\"],\"assignee\":\"bob\"}'\n"
	if err := os.WriteFile(script, []byte(scriptBody), 0o755); err != nil {
		t.Fatalf("write fake bd: %v", err)
	}

	client := NewCLIClient(WithBinaryPath(script))

	ctx := context.Background()
	issue, err := client.CreateFull(ctx, "Bug Fix", "bug", 1, []string{"urgent", "backend"}, "bob", "")
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
}

func TestCLIClient_CreateFull_HandlesOutputWithPrefix(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	script := filepath.Join(dir, "fakebd.sh")

	// Fake bd script that outputs warning before JSON (simulates production db warning)
	scriptBody := "#!/bin/sh\n" +
		"echo '⚠ Creating issue with Test prefix in production database.'\n" +
		"echo '{\"id\":\"ab-prefix\",\"title\":\"Test\",\"status\":\"open\",\"priority\":2,\"issue_type\":\"task\"}'\n"
	if err := os.WriteFile(script, []byte(scriptBody), 0o755); err != nil {
		t.Fatalf("write fake bd: %v", err)
	}

	client := NewCLIClient(WithBinaryPath(script))

	ctx := context.Background()
	issue, err := client.CreateFull(ctx, "Test Title", "task", 2, nil, "", "")
	if err != nil {
		t.Fatalf("CreateFull should handle output with prefix: %v", err)
	}

	if issue.ID != "ab-prefix" {
		t.Errorf("expected ID 'ab-prefix', got %q", issue.ID)
	}
}
