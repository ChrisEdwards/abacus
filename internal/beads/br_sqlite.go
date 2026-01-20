// Package beads provides client implementations for beads issue tracking.
//
// EVOLVING: brSQLiteClient is the active development backend for beads_rust (br).
// This client will evolve as br adds new features. Unlike bdSQLiteClient which is
// frozen at bd v0.38.0, changes and new features go here.
package beads

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite" // Pure Go SQLite driver, WAL-friendly
)

// brSQLiteClient reads issues/comments directly from the br SQLite database in
// read-only WAL mode. Mutating operations delegate to brCLIClient.
//
// Schema compatibility: br schema has 43 columns (superset of bd). All columns
// abacus reads exist in both schemas. Extra columns in br are ignored.
type brSQLiteClient struct {
	dbPath string
	dsn    string
	writer Writer // brCLIClient for write operations
}

// NewBrSQLiteClient constructs a client that reads via SQLite and writes via br CLI.
// A valid dbPath is required; use NewBrCLIClient for CLI-only operations.
func NewBrSQLiteClient(dbPath string, opts ...BrCLIOption) Client {
	trimmed := strings.TrimSpace(dbPath)
	if trimmed == "" {
		// brCLIClient only implements Writer, not Client, so we can't fall back.
		panic("NewBrSQLiteClient requires a non-empty dbPath; use NewBrCLIClient for CLI-only operations")
	}
	dsn := buildBrSQLiteDSN(trimmed)
	// Ensure writes go to the same DB when the CLI is used for mutations.
	opts = append(opts, WithBrDatabasePath(trimmed))
	return &brSQLiteClient{
		dbPath: trimmed,
		dsn:    dsn,
		writer: NewBrCLIClient(opts...),
	}
}

// buildBrSQLiteDSN creates a read-only WAL DSN for the given path.
func buildBrSQLiteDSN(dbPath string) string {
	u := url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(dbPath),
	}
	q := url.Values{}
	q.Set("mode", "ro")
	q.Set("_journal_mode", "WAL")
	q.Set("_busy_timeout", "3000")
	q.Set("_foreign_keys", "on")
	q.Set("cache", "shared")
	u.RawQuery = q.Encode()
	return u.String()
}

func (c *brSQLiteClient) openDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", c.dsn)
	if err != nil {
		return nil, fmt.Errorf("open br sqlite db: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping br sqlite db: %w", err)
	}
	return db, nil
}

// Reader interface implementation - direct SQLite queries

func (c *brSQLiteClient) List(ctx context.Context) ([]LiteIssue, error) {
	db, err := c.openDB(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	rows, err := db.QueryContext(ctx, `
		SELECT id
		FROM issues
		WHERE status != 'tombstone' AND (deleted_at IS NULL)
		ORDER BY created_at, id
	`)
	if err != nil {
		return nil, fmt.Errorf("query issues: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var issues []LiteIssue
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan issue id: %w", err)
		}
		issues = append(issues, LiteIssue{ID: id})
	}
	return issues, rows.Err()
}

func (c *brSQLiteClient) Show(ctx context.Context, ids []string) ([]FullIssue, error) {
	if len(ids) == 0 {
		return []FullIssue{}, nil
	}
	all, err := c.Export(ctx)
	if err != nil {
		return nil, err
	}
	set := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	var filtered []FullIssue
	for _, iss := range all {
		if _, ok := set[iss.ID]; ok {
			filtered = append(filtered, iss)
		}
	}
	return filtered, nil
}

func (c *brSQLiteClient) Export(ctx context.Context) ([]FullIssue, error) {
	db, err := c.openDB(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	issueMap, ordered, err := brLoadIssues(ctx, db)
	if err != nil {
		return nil, err
	}

	if err := brLoadLabels(ctx, db, issueMap); err != nil {
		return nil, err
	}
	if err := brLoadDependencies(ctx, db, issueMap); err != nil {
		return nil, err
	}
	if err := brLoadComments(ctx, db, issueMap); err != nil {
		return nil, err
	}

	out := make([]FullIssue, 0, len(ordered))
	for _, iss := range ordered {
		out = append(out, *iss)
	}
	return out, nil
}

func brLoadIssues(ctx context.Context, db *sql.DB) (map[string]*FullIssue, []*FullIssue, error) {
	const query = `SELECT id, title, description, design, acceptance_criteria, notes,
		       status, priority, issue_type, COALESCE(assignee, ''),
		       created_at, updated_at, COALESCE(closed_at, ''), COALESCE(external_ref, '')
		FROM issues WHERE status != 'tombstone' AND (deleted_at IS NULL) ORDER BY created_at, id`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, nil, fmt.Errorf("query issues: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	issues := make(map[string]*FullIssue)
	var ordered []*FullIssue
	for rows.Next() {
		var iss FullIssue
		scanErr := rows.Scan(
			&iss.ID,
			&iss.Title,
			&iss.Description,
			&iss.Design,
			&iss.AcceptanceCriteria,
			&iss.Notes,
			&iss.Status,
			&iss.Priority,
			&iss.IssueType,
			&iss.Assignee,
			&iss.CreatedAt,
			&iss.UpdatedAt,
			&iss.ClosedAt,
			&iss.ExternalRef,
		)
		if scanErr != nil {
			return nil, nil, fmt.Errorf("scan issue: %w", scanErr)
		}
		iss.Labels = []string{}
		iss.Dependencies = []Dependency{}
		iss.Dependents = []Dependent{}
		iss.Comments = []Comment{}
		issues[iss.ID] = &iss
		ordered = append(ordered, &iss)
	}
	return issues, ordered, rows.Err()
}

func brLoadLabels(ctx context.Context, db *sql.DB, issues map[string]*FullIssue) error {
	rows, err := db.QueryContext(ctx, `
		SELECT issue_id, label
		FROM labels
		ORDER BY issue_id, label
	`)
	if err != nil {
		return fmt.Errorf("query labels: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var issueID, label string
		if err := rows.Scan(&issueID, &label); err != nil {
			return fmt.Errorf("scan label: %w", err)
		}
		if iss, ok := issues[issueID]; ok {
			iss.Labels = append(iss.Labels, label)
		}
	}
	return rows.Err()
}

func brLoadDependencies(ctx context.Context, db *sql.DB, issues map[string]*FullIssue) error {
	rows, err := db.QueryContext(ctx, `
		SELECT issue_id, depends_on_id, type
		FROM dependencies
	`)
	if err != nil {
		return fmt.Errorf("query dependencies: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var issueID, dependsOnID, depType string
		if err := rows.Scan(&issueID, &dependsOnID, &depType); err != nil {
			return fmt.Errorf("scan dependency: %w", err)
		}
		if iss, ok := issues[issueID]; ok {
			iss.Dependencies = append(iss.Dependencies, Dependency{TargetID: dependsOnID, Type: depType})
		}
		if rev, ok := issues[dependsOnID]; ok {
			rev.Dependents = append(rev.Dependents, Dependent{ID: issueID, Type: depType})
		}
	}
	return rows.Err()
}

func brLoadComments(ctx context.Context, db *sql.DB, issues map[string]*FullIssue) error {
	rows, err := db.QueryContext(ctx, `
		SELECT id, issue_id, author, text, created_at
		FROM comments
		ORDER BY created_at, id
	`)
	if err != nil {
		return fmt.Errorf("query comments: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var cmt Comment
		if err := rows.Scan(&cmt.ID, &cmt.IssueID, &cmt.Author, &cmt.Text, &cmt.CreatedAt); err != nil {
			return fmt.Errorf("scan comment: %w", err)
		}
		if iss, ok := issues[cmt.IssueID]; ok {
			iss.Comments = append(iss.Comments, cmt)
		}
	}
	return rows.Err()
}

func (c *brSQLiteClient) Comments(ctx context.Context, issueID string) ([]Comment, error) {
	db, err := c.openDB(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	rows, err := db.QueryContext(ctx, `
		SELECT id, issue_id, author, text, created_at
		FROM comments
		WHERE issue_id = ?
		ORDER BY created_at, id
	`, issueID)
	if err != nil {
		return nil, fmt.Errorf("query comments: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var comments []Comment
	for rows.Next() {
		var cmt Comment
		if err := rows.Scan(&cmt.ID, &cmt.IssueID, &cmt.Author, &cmt.Text, &cmt.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		comments = append(comments, cmt)
	}
	if comments == nil {
		comments = []Comment{}
	}
	return comments, rows.Err()
}

// Writer interface - delegate to brCLIClient

func (c *brSQLiteClient) UpdateStatus(ctx context.Context, issueID, newStatus string) error {
	return c.writer.UpdateStatus(ctx, issueID, newStatus)
}

func (c *brSQLiteClient) Close(ctx context.Context, issueID string) error {
	return c.writer.Close(ctx, issueID)
}

func (c *brSQLiteClient) Reopen(ctx context.Context, issueID string) error {
	return c.writer.Reopen(ctx, issueID)
}

func (c *brSQLiteClient) AddLabel(ctx context.Context, issueID, label string) error {
	return c.writer.AddLabel(ctx, issueID, label)
}

func (c *brSQLiteClient) RemoveLabel(ctx context.Context, issueID, label string) error {
	return c.writer.RemoveLabel(ctx, issueID, label)
}

func (c *brSQLiteClient) UpdateFull(ctx context.Context, issueID, title, issueType string, priority int, labels []string, assignee, description string) error {
	return c.writer.UpdateFull(ctx, issueID, title, issueType, priority, labels, assignee, description)
}

func (c *brSQLiteClient) Create(ctx context.Context, title, issueType string, priority int, labels []string, assignee string) (string, error) {
	return c.writer.Create(ctx, title, issueType, priority, labels, assignee)
}

func (c *brSQLiteClient) CreateFull(ctx context.Context, title, issueType string, priority int, labels []string, assignee, description, parentID string) (FullIssue, error) {
	return c.writer.CreateFull(ctx, title, issueType, priority, labels, assignee, description, parentID)
}

func (c *brSQLiteClient) AddDependency(ctx context.Context, fromID, toID, depType string) error {
	return c.writer.AddDependency(ctx, fromID, toID, depType)
}

func (c *brSQLiteClient) RemoveDependency(ctx context.Context, fromID, toID, depType string) error {
	return c.writer.RemoveDependency(ctx, fromID, toID, depType)
}

func (c *brSQLiteClient) Delete(ctx context.Context, issueID string, cascade bool) error {
	return c.writer.Delete(ctx, issueID, cascade)
}

func (c *brSQLiteClient) AddComment(ctx context.Context, issueID, text string) error {
	return c.writer.AddComment(ctx, issueID, text)
}
