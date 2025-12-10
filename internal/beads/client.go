package beads

import "context"

// Client defines the operations required to interact with the Beads CLI.
type Client interface {
	List(ctx context.Context) ([]LiteIssue, error)
	Show(ctx context.Context, ids []string) ([]FullIssue, error)
	Export(ctx context.Context) ([]FullIssue, error)
	Comments(ctx context.Context, issueID string) ([]Comment, error)
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
}
