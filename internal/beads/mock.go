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
	ListFn     func(context.Context) ([]LiteIssue, error)
	ShowFn     func(context.Context, []string) ([]FullIssue, error)
	CommentsFn func(context.Context, string) ([]Comment, error)

	mu                sync.Mutex
	ListCallCount     int
	ShowCallCount     int
	CommentsCallCount int
	ShowCallArgs      [][]string
	CommentIDs        []string
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
