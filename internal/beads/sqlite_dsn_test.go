package beads

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildSQLiteDSN_EscapesReservedPathCharacters(t *testing.T) {
	t.Parallel()

	dsn := buildSQLiteDSN("/tmp/project#1/.beads/beads?.db")

	if !strings.HasPrefix(dsn, "file:/tmp/project%231/.beads/beads%3F.db?") {
		t.Fatalf("reserved path characters must be escaped in DSN: %q", dsn)
	}
	if strings.Contains(dsn, "/tmp/project#1/.beads/beads?.db?") {
		t.Fatalf("DSN must not include raw reserved characters in the path: %q", dsn)
	}
}

func TestSQLiteClients_ListWithReservedCharactersInDBPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		newClient func(string) Client
	}{
		{
			name: "br",
			newClient: func(dbPath string) Client {
				return NewBrSQLiteClient(dbPath)
			},
		},
		{
			name: "bd",
			newClient: func(dbPath string) Client {
				return NewBdSQLiteClient(dbPath)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dbPath := filepath.Join(t.TempDir(), "project#1", ".beads", "beads#1.db")
			createTestBrDB(t, dbPath)
			seedTestData(t, dbPath)

			issues, err := tt.newClient(dbPath).List(context.Background())
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if len(issues) != 3 {
				t.Fatalf("got %d issues, want 3", len(issues))
			}
		})
	}
}
