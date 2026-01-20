// Package beads provides client implementations for beads issue tracking.
//
// FROZEN: bdSQLiteClient is frozen at bd v0.38.0 compatibility.
// Do NOT add new features or modify behavior - only critical bug fixes allowed.
// All new development should go to brSQLiteClient in br_sqlite.go instead.
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

// bdSQLiteClient reads issues/comments directly from the SQLite database in
// read-only WAL mode to avoid bd export churn. Mutating operations delegate to
// the CLI writer to keep behavior consistent with the daemon.
// FROZEN: Only critical bug fixes allowed. New features go to brSQLiteClient.
type bdSQLiteClient struct {
	dbPath string
	dsn    string
	writer Writer // Only write operations needed; Reader implemented directly via SQLite
}

// NewBdSQLiteClient constructs a client that reads via SQLite and writes via the bd CLI.
// FROZEN: Use NewBrSQLiteClient for new development.
func NewBdSQLiteClient(dbPath string, opts ...BdCLIOption) Client {
	trimmed := strings.TrimSpace(dbPath)
	if trimmed == "" {
		// Unlike the old NewSQLiteClient, we don't fall back to pure CLI.
		// A valid dbPath is required for SQLite client.
		panic("NewBdSQLiteClient requires a non-empty dbPath; use NewBdCLIClient for CLI-only operations")
	}
	dsn := buildBdSQLiteDSN(trimmed)
	// Ensure writes go to the same DB when the CLI is used for mutations.
	opts = append(opts, WithBdDatabasePath(trimmed))
	return &bdSQLiteClient{
		dbPath: trimmed,
		dsn:    dsn,
		writer: NewBdCLIClient(opts...),
	}
}

// buildBdSQLiteDSN creates a read-only WAL DSN for the given path.
func buildBdSQLiteDSN(dbPath string) string {
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

func (c *bdSQLiteClient) openDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", c.dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite db: %w", err)
	}
	return db, nil
}

func (c *bdSQLiteClient) List(ctx context.Context) ([]LiteIssue, error) {
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

func (c *bdSQLiteClient) Show(ctx context.Context, ids []string) ([]FullIssue, error) {
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

func (c *bdSQLiteClient) Export(ctx context.Context) ([]FullIssue, error) {
	db, err := c.openDB(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	issueMap, ordered, err := loadBdIssues(ctx, db)
	if err != nil {
		return nil, err
	}

	if err := loadBdLabels(ctx, db, issueMap); err != nil {
		return nil, err
	}
	if err := loadBdDependencies(ctx, db, issueMap); err != nil {
		return nil, err
	}
	if err := loadBdComments(ctx, db, issueMap); err != nil {
		return nil, err
	}

	out := make([]FullIssue, 0, len(ordered))
	for _, iss := range ordered {
		out = append(out, *iss)
	}
	return out, nil
}

func loadBdIssues(ctx context.Context, db *sql.DB) (map[string]*FullIssue, []*FullIssue, error) {
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

func loadBdLabels(ctx context.Context, db *sql.DB, issues map[string]*FullIssue) error {
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

func loadBdDependencies(ctx context.Context, db *sql.DB, issues map[string]*FullIssue) error {
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

func loadBdComments(ctx context.Context, db *sql.DB, issues map[string]*FullIssue) error {
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
		var c Comment
		if err := rows.Scan(&c.ID, &c.IssueID, &c.Author, &c.Text, &c.CreatedAt); err != nil {
			return fmt.Errorf("scan comment: %w", err)
		}
		if iss, ok := issues[c.IssueID]; ok {
			iss.Comments = append(iss.Comments, c)
		}
	}
	return rows.Err()
}

func (c *bdSQLiteClient) Comments(ctx context.Context, issueID string) ([]Comment, error) {
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

// Mutating operations delegate to the Writer (CLI commands).
func (c *bdSQLiteClient) UpdateStatus(ctx context.Context, issueID, newStatus string) error {
	return c.writer.UpdateStatus(ctx, issueID, newStatus)
}

func (c *bdSQLiteClient) Close(ctx context.Context, issueID string) error {
	return c.writer.Close(ctx, issueID)
}

func (c *bdSQLiteClient) Reopen(ctx context.Context, issueID string) error {
	return c.writer.Reopen(ctx, issueID)
}

func (c *bdSQLiteClient) AddLabel(ctx context.Context, issueID, label string) error {
	return c.writer.AddLabel(ctx, issueID, label)
}

func (c *bdSQLiteClient) RemoveLabel(ctx context.Context, issueID, label string) error {
	return c.writer.RemoveLabel(ctx, issueID, label)
}

func (c *bdSQLiteClient) UpdateFull(ctx context.Context, issueID, title, issueType string, priority int, labels []string, assignee, description string) error {
	return c.writer.UpdateFull(ctx, issueID, title, issueType, priority, labels, assignee, description)
}

func (c *bdSQLiteClient) Create(ctx context.Context, title, issueType string, priority int, labels []string, assignee string) (string, error) {
	return c.writer.Create(ctx, title, issueType, priority, labels, assignee)
}

func (c *bdSQLiteClient) CreateFull(ctx context.Context, title, issueType string, priority int, labels []string, assignee, description, parentID string) (FullIssue, error) {
	return c.writer.CreateFull(ctx, title, issueType, priority, labels, assignee, description, parentID)
}

func (c *bdSQLiteClient) AddDependency(ctx context.Context, fromID, toID, depType string) error {
	return c.writer.AddDependency(ctx, fromID, toID, depType)
}

func (c *bdSQLiteClient) RemoveDependency(ctx context.Context, fromID, toID, depType string) error {
	return c.writer.RemoveDependency(ctx, fromID, toID, depType)
}

func (c *bdSQLiteClient) Delete(ctx context.Context, issueID string, cascade bool) error {
	return c.writer.Delete(ctx, issueID, cascade)
}

func (c *bdSQLiteClient) AddComment(ctx context.Context, issueID, text string) error {
	return c.writer.AddComment(ctx, issueID, text)
}
