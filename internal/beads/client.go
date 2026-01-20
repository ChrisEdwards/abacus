package beads

import "context"

// Reader handles all read operations.
// In production, only SQLite clients implement this interface.
// CLI clients do NOT implement Reader since read operations go through SQLite.
type Reader interface {
	List(ctx context.Context) ([]LiteIssue, error)
	Show(ctx context.Context, ids []string) ([]FullIssue, error)
	Export(ctx context.Context) ([]FullIssue, error)
	Comments(ctx context.Context, issueID string) ([]Comment, error)
}

// Writer handles all mutation operations.
// CLI clients implement Writer only (bdCLIClient, brCLIClient).
// SQLite clients embed a Writer for delegation to CLI.
type Writer interface {
	UpdateStatus(ctx context.Context, issueID, newStatus string) error
	Close(ctx context.Context, issueID string) error
	Reopen(ctx context.Context, issueID string) error
	AddLabel(ctx context.Context, issueID, label string) error
	RemoveLabel(ctx context.Context, issueID, label string) error
	UpdateFull(ctx context.Context, issueID, title, issueType string, priority int, labels []string, assignee, description string) error
	Create(ctx context.Context, title, issueType string, priority int, labels []string, assignee string) (string, error)
	CreateFull(ctx context.Context, title, issueType string, priority int, labels []string, assignee, description, parentID string) (FullIssue, error)
	AddDependency(ctx context.Context, fromID, toID, depType string) error
	RemoveDependency(ctx context.Context, fromID, toID, depType string) error
	Delete(ctx context.Context, issueID string, cascade bool) error
	AddComment(ctx context.Context, issueID, text string) error
}

// Client combines Reader and Writer for full functionality.
// SQLite clients implement this by providing Reader methods directly
// and embedding a Writer (CLI client) for mutations.
type Client interface {
	Reader
	Writer
}
