package beads

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	_ "modernc.org/sqlite" // Pure Go SQLite driver, WAL-friendly
)

// sqliteClient reads issues/comments directly from the SQLite database in
// read-only WAL mode to avoid bd export churn. Mutating operations delegate to
// the CLI client to keep behavior consistent with the daemon.
type sqliteClient struct {
	dbPath string
	dsn    string
	cli    Client

	// Schema detection (lazy, cached)
	schemaOnce       sync.Once
	hasGraphLinkCols bool // true if duplicate_of, superseded_by columns exist
}

// NewSQLiteClient constructs a client that reads via SQLite and writes via the CLI.
// If dbPath is empty, it falls back to a pure CLI client.
func NewSQLiteClient(dbPath string, opts ...CLIOption) Client {
	trimmed := strings.TrimSpace(dbPath)
	if trimmed == "" {
		return NewCLIClient(opts...)
	}
	dsn := buildSQLiteDSN(trimmed)
	// Ensure writes go to the same DB when the CLI is used for mutations.
	opts = append(opts, WithDatabasePath(trimmed))
	return &sqliteClient{
		dbPath: trimmed,
		dsn:    dsn,
		cli:    NewCLIClient(opts...),
	}
}

// buildSQLiteDSN creates a read-only WAL DSN for the given path.
func buildSQLiteDSN(dbPath string) string {
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

func (c *sqliteClient) openDB(ctx context.Context) (*sql.DB, error) {
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

// detectSchema checks if the database has the new graph link columns
// (duplicate_of, superseded_by) added in beads v0.0.31+.
func (c *sqliteClient) detectSchema(ctx context.Context, db *sql.DB) {
	c.schemaOnce.Do(func() {
		rows, err := db.QueryContext(ctx, `PRAGMA table_info(issues)`)
		if err != nil {
			return // Assume no new columns on error
		}
		defer func() {
			_ = rows.Close()
		}()

		for rows.Next() {
			var cid int
			var name, colType string
			var notNull, pk int
			var dfltValue sql.NullString
			if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
				continue
			}
			if name == "duplicate_of" || name == "superseded_by" {
				c.hasGraphLinkCols = true
				return // Found at least one, that's enough
			}
		}
	})
}

func (c *sqliteClient) List(ctx context.Context) ([]LiteIssue, error) {
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

func (c *sqliteClient) Show(ctx context.Context, ids []string) ([]FullIssue, error) {
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

func (c *sqliteClient) Export(ctx context.Context) ([]FullIssue, error) {
	db, err := c.openDB(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	// Detect schema to check for new graph link columns (beads v0.0.31+)
	c.detectSchema(ctx, db)

	issueMap, ordered, err := loadIssues(ctx, db, c.hasGraphLinkCols)
	if err != nil {
		return nil, err
	}

	if err := loadLabels(ctx, db, issueMap); err != nil {
		return nil, err
	}
	if err := loadDependencies(ctx, db, issueMap); err != nil {
		return nil, err
	}
	if err := loadComments(ctx, db, issueMap); err != nil {
		return nil, err
	}

	out := make([]FullIssue, 0, len(ordered))
	for _, iss := range ordered {
		out = append(out, *iss)
	}
	return out, nil
}

func loadIssues(ctx context.Context, db *sql.DB, hasGraphLinkCols bool) (map[string]*FullIssue, []*FullIssue, error) {
	// Build query conditionally based on schema detection
	baseCols := `id, title, description, design, acceptance_criteria, notes,
		       status, priority, issue_type, COALESCE(assignee, ''),
		       created_at, updated_at, COALESCE(closed_at, ''), COALESCE(external_ref, '')`
	if hasGraphLinkCols {
		baseCols += `, COALESCE(duplicate_of, ''), COALESCE(superseded_by, '')`
	}
	//nolint:gosec // G201: baseCols is hardcoded column names, not user input
	query := fmt.Sprintf(`SELECT %s FROM issues WHERE status != 'tombstone' AND (deleted_at IS NULL) ORDER BY created_at, id`, baseCols)

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
		var scanErr error
		if hasGraphLinkCols {
			scanErr = rows.Scan(
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
				&iss.DuplicateOf,
				&iss.SupersededBy,
			)
		} else {
			scanErr = rows.Scan(
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
		}
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

func loadLabels(ctx context.Context, db *sql.DB, issues map[string]*FullIssue) error {
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

func loadDependencies(ctx context.Context, db *sql.DB, issues map[string]*FullIssue) error {
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

func loadComments(ctx context.Context, db *sql.DB, issues map[string]*FullIssue) error {
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

func (c *sqliteClient) Comments(ctx context.Context, issueID string) ([]Comment, error) {
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

// Mutating operations delegate to the CLI (keeps daemon-aware behavior).
func (c *sqliteClient) UpdateStatus(ctx context.Context, issueID, newStatus string) error {
	return c.cli.UpdateStatus(ctx, issueID, newStatus)
}

func (c *sqliteClient) Close(ctx context.Context, issueID string) error {
	return c.cli.Close(ctx, issueID)
}

func (c *sqliteClient) Reopen(ctx context.Context, issueID string) error {
	return c.cli.Reopen(ctx, issueID)
}

func (c *sqliteClient) AddLabel(ctx context.Context, issueID, label string) error {
	return c.cli.AddLabel(ctx, issueID, label)
}

func (c *sqliteClient) RemoveLabel(ctx context.Context, issueID, label string) error {
	return c.cli.RemoveLabel(ctx, issueID, label)
}

func (c *sqliteClient) UpdateFull(ctx context.Context, issueID, title, issueType string, priority int, labels []string, assignee, description string) error {
	return c.cli.UpdateFull(ctx, issueID, title, issueType, priority, labels, assignee, description)
}

func (c *sqliteClient) Create(ctx context.Context, title, issueType string, priority int, labels []string, assignee string) (string, error) {
	return c.cli.Create(ctx, title, issueType, priority, labels, assignee)
}

func (c *sqliteClient) CreateFull(ctx context.Context, title, issueType string, priority int, labels []string, assignee, description, parentID string) (FullIssue, error) {
	return c.cli.CreateFull(ctx, title, issueType, priority, labels, assignee, description, parentID)
}

func (c *sqliteClient) AddDependency(ctx context.Context, fromID, toID, depType string) error {
	return c.cli.AddDependency(ctx, fromID, toID, depType)
}

func (c *sqliteClient) RemoveDependency(ctx context.Context, fromID, toID, depType string) error {
	return c.cli.RemoveDependency(ctx, fromID, toID, depType)
}

func (c *sqliteClient) Delete(ctx context.Context, issueID string, cascade bool) error {
	return c.cli.Delete(ctx, issueID, cascade)
}

func (c *sqliteClient) AddComment(ctx context.Context, issueID, text string) error {
	return c.cli.AddComment(ctx, issueID, text)
}
