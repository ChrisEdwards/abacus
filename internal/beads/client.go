package beads

import "context"

// Client defines the operations required to interact with the Beads CLI.
type Client interface {
	List(ctx context.Context) ([]LiteIssue, error)
	Show(ctx context.Context, ids []string) ([]FullIssue, error)
	Comments(ctx context.Context, issueID string) ([]Comment, error)
	UpdateStatus(ctx context.Context, issueID, newStatus string) error
	Close(ctx context.Context, issueID string) error
	Reopen(ctx context.Context, issueID string) error
}
