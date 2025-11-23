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
