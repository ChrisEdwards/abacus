package beads

import (
	"context"
	"errors"
	"sync"
)

// ErrMockNotImplemented is returned when a MockClient method lacks an override.
var ErrMockNotImplemented = errors.New("beads.MockClient: method not implemented")

// MockClient is a test double for the Beads client interface.
type MockClient struct {
	ListFn         func(context.Context) ([]LiteIssue, error)
	ShowFn         func(context.Context, []string) ([]FullIssue, error)
	CommentsFn     func(context.Context, string) ([]Comment, error)
	UpdateStatusFn func(context.Context, string, string) error
	CloseFn        func(context.Context, string) error
	ReopenFn       func(context.Context, string) error

	mu                      sync.Mutex
	ListCallCount           int
	ShowCallCount           int
	CommentsCallCount       int
	UpdateStatusCallCount   int
	CloseCallCount          int
	ReopenCallCount         int
	ShowCallArgs            [][]string
	CommentIDs              []string
	UpdateStatusCallArgs    [][]string // [issueID, newStatus]
	CloseCallArgs           []string
	ReopenCallArgs          []string
}

// NewMockClient returns a MockClient with zeroed handlers.
func NewMockClient() *MockClient {
	return &MockClient{}
}

// List invokes the configured stub or returns ErrMockNotImplemented.
func (m *MockClient) List(ctx context.Context) ([]LiteIssue, error) {
	m.mu.Lock()
	m.ListCallCount++
	m.mu.Unlock()
	if m.ListFn == nil {
		return nil, ErrMockNotImplemented
	}
	return m.ListFn(ctx)
}

// Show invokes the configured stub or returns ErrMockNotImplemented.
func (m *MockClient) Show(ctx context.Context, ids []string) ([]FullIssue, error) {
	m.mu.Lock()
	m.ShowCallCount++
	copied := append([]string{}, ids...)
	m.ShowCallArgs = append(m.ShowCallArgs, copied)
	m.mu.Unlock()

	if m.ShowFn == nil {
		return nil, ErrMockNotImplemented
	}
	return m.ShowFn(ctx, ids)
}

// Comments invokes the configured stub or returns ErrMockNotImplemented.
func (m *MockClient) Comments(ctx context.Context, issueID string) ([]Comment, error) {
	m.mu.Lock()
	m.CommentsCallCount++
	m.CommentIDs = append(m.CommentIDs, issueID)
	m.mu.Unlock()

	if m.CommentsFn == nil {
		return nil, ErrMockNotImplemented
	}
	return m.CommentsFn(ctx, issueID)
}

// UpdateStatus invokes the configured stub or returns nil (no-op by default).
func (m *MockClient) UpdateStatus(ctx context.Context, issueID, newStatus string) error {
	m.mu.Lock()
	m.UpdateStatusCallCount++
	m.UpdateStatusCallArgs = append(m.UpdateStatusCallArgs, []string{issueID, newStatus})
	m.mu.Unlock()

	if m.UpdateStatusFn == nil {
		return nil // Default to no-op for tests
	}
	return m.UpdateStatusFn(ctx, issueID, newStatus)
}

// Close invokes the configured stub or returns nil (no-op by default).
func (m *MockClient) Close(ctx context.Context, issueID string) error {
	m.mu.Lock()
	m.CloseCallCount++
	m.CloseCallArgs = append(m.CloseCallArgs, issueID)
	m.mu.Unlock()

	if m.CloseFn == nil {
		return nil // Default to no-op for tests
	}
	return m.CloseFn(ctx, issueID)
}

// Reopen invokes the configured stub or returns nil (no-op by default).
func (m *MockClient) Reopen(ctx context.Context, issueID string) error {
	m.mu.Lock()
	m.ReopenCallCount++
	m.ReopenCallArgs = append(m.ReopenCallArgs, issueID)
	m.mu.Unlock()

	if m.ReopenFn == nil {
		return nil // Default to no-op for tests
	}
	return m.ReopenFn(ctx, issueID)
}
