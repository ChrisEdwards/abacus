package beads

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestBdSQLiteClient_Comments_NullCreatedAt(t *testing.T) {
	t.Parallel()

	dbPath := testBrDB(t)
	seedTestData(t, dbPath)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO comments (issue_id, author, text, created_at) VALUES ('ab-001', 'Dave', 'Null timestamp comment', NULL)`); err != nil {
		t.Fatalf("insert comment with null created_at: %v", err)
	}
	_ = db.Close()

	client := NewBdSQLiteClient(dbPath)
	ctx := context.Background()

	comments, err := client.Comments(ctx, "ab-001")
	if err != nil {
		t.Fatalf("Comments: %v", err)
	}

	var nullCmt *Comment
	for i := range comments {
		if comments[i].Author == "Dave" {
			nullCmt = &comments[i]
			break
		}
	}
	if nullCmt == nil {
		t.Fatal("comment with null created_at not found")
	}
	if nullCmt.CreatedAt != "" {
		t.Errorf("expected empty string for null created_at, got %q", nullCmt.CreatedAt)
	}
}

func TestBdSQLiteClient_Export_NullCommentCreatedAt(t *testing.T) {
	t.Parallel()

	dbPath := testBrDB(t)
	seedTestData(t, dbPath)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO comments (issue_id, author, text, created_at) VALUES ('ab-001', 'Dave', 'Null timestamp comment', NULL)`); err != nil {
		t.Fatalf("insert comment with null created_at: %v", err)
	}
	_ = db.Close()

	client := NewBdSQLiteClient(dbPath)
	ctx := context.Background()

	issues, err := client.Export(ctx)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}

	var ab001 *FullIssue
	for i := range issues {
		if issues[i].ID == "ab-001" {
			ab001 = &issues[i]
			break
		}
	}
	if ab001 == nil {
		t.Fatal("ab-001 not found in export")
	}

	var nullCmt *Comment
	for i := range ab001.Comments {
		if ab001.Comments[i].Author == "Dave" {
			nullCmt = &ab001.Comments[i]
			break
		}
	}
	if nullCmt == nil {
		t.Fatal("comment with null created_at not found in export")
	}
	if nullCmt.CreatedAt != "" {
		t.Errorf("expected empty string for null created_at, got %q", nullCmt.CreatedAt)
	}
}
