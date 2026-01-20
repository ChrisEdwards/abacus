// Package beads provides client implementations for beads issue tracking.
//
// DEPRECATED: This file provides backward-compatible wrappers for the bd backend.
// New code should use NewBdCLIClient and NewBdSQLiteClient directly, or use
// the backend detection in backend.go to get the appropriate client.
package beads

import "context"

// CLIOption is an alias for BdCLIOption for backward compatibility.
// Deprecated: Use BdCLIOption directly.
type CLIOption = BdCLIOption

// WithBinaryPath is an alias for WithBdBinaryPath for backward compatibility.
// Deprecated: Use WithBdBinaryPath directly.
var WithBinaryPath = WithBdBinaryPath

// WithDatabasePath is an alias for WithBdDatabasePath for backward compatibility.
// Deprecated: Use WithBdDatabasePath directly.
var WithDatabasePath = WithBdDatabasePath

// NewCLIClient constructs a Beads CLI-backed Writer implementation.
// Deprecated: Use NewBdCLIClient directly for bd, or NewBrCLIClient for br.
// Note: Returns Client for backward compatibility, but only Writer methods work.
func NewCLIClient(opts ...CLIOption) Client {
	writer := NewBdCLIClient(opts...)
	return &cliClientWrapper{writer: writer}
}

// cliClientWrapper wraps a Writer to implement Client interface for backward compat.
// Read methods return errors as they're not implemented.
type cliClientWrapper struct {
	writer Writer
}

// List is not implemented for CLI client - use SQLite client for reads.
func (c *cliClientWrapper) List(_ context.Context) ([]LiteIssue, error) {
	return nil, ErrReadNotImplemented
}

// Show is not implemented for CLI client - use SQLite client for reads.
func (c *cliClientWrapper) Show(_ context.Context, _ []string) ([]FullIssue, error) {
	return nil, ErrReadNotImplemented
}

// Export is not implemented for CLI client - use SQLite client for reads.
func (c *cliClientWrapper) Export(_ context.Context) ([]FullIssue, error) {
	return nil, ErrReadNotImplemented
}

// Comments is not implemented for CLI client - use SQLite client for reads.
func (c *cliClientWrapper) Comments(_ context.Context, _ string) ([]Comment, error) {
	return nil, ErrReadNotImplemented
}

// Writer method delegations
func (c *cliClientWrapper) UpdateStatus(ctx context.Context, issueID, newStatus string) error {
	return c.writer.UpdateStatus(ctx, issueID, newStatus)
}

func (c *cliClientWrapper) Close(ctx context.Context, issueID string) error {
	return c.writer.Close(ctx, issueID)
}

func (c *cliClientWrapper) Reopen(ctx context.Context, issueID string) error {
	return c.writer.Reopen(ctx, issueID)
}

func (c *cliClientWrapper) AddLabel(ctx context.Context, issueID, label string) error {
	return c.writer.AddLabel(ctx, issueID, label)
}

func (c *cliClientWrapper) RemoveLabel(ctx context.Context, issueID, label string) error {
	return c.writer.RemoveLabel(ctx, issueID, label)
}

func (c *cliClientWrapper) Create(ctx context.Context, title, issueType string, priority int, labels []string, assignee string) (string, error) {
	return c.writer.Create(ctx, title, issueType, priority, labels, assignee)
}

func (c *cliClientWrapper) CreateFull(ctx context.Context, title, issueType string, priority int, labels []string, assignee, description, parentID string) (FullIssue, error) {
	return c.writer.CreateFull(ctx, title, issueType, priority, labels, assignee, description, parentID)
}

func (c *cliClientWrapper) UpdateFull(ctx context.Context, issueID, title, issueType string, priority int, labels []string, assignee, description string) error {
	return c.writer.UpdateFull(ctx, issueID, title, issueType, priority, labels, assignee, description)
}

func (c *cliClientWrapper) AddDependency(ctx context.Context, fromID, toID, depType string) error {
	return c.writer.AddDependency(ctx, fromID, toID, depType)
}

func (c *cliClientWrapper) RemoveDependency(ctx context.Context, fromID, toID, depType string) error {
	return c.writer.RemoveDependency(ctx, fromID, toID, depType)
}

func (c *cliClientWrapper) Delete(ctx context.Context, issueID string, cascade bool) error {
	return c.writer.Delete(ctx, issueID, cascade)
}

func (c *cliClientWrapper) AddComment(ctx context.Context, issueID, text string) error {
	return c.writer.AddComment(ctx, issueID, text)
}
