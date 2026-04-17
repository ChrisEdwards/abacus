package beads

import (
	"context"
	"errors"
	"testing"
)

func TestMockClient_UpdatePriority_RecordsCall(t *testing.T) {
	t.Parallel()

	m := NewMockClient()
	ctx := context.Background()

	if err := m.UpdatePriority(ctx, "ab-prio", 1); err != nil {
		t.Fatalf("UpdatePriority returned error: %v", err)
	}

	if m.UpdatePriorityCallCount != 1 {
		t.Errorf("expected 1 call, got %d", m.UpdatePriorityCallCount)
	}
	if len(m.UpdatePriorityCallArgs) != 1 {
		t.Fatalf("expected 1 recorded call, got %d", len(m.UpdatePriorityCallArgs))
	}
	got := m.UpdatePriorityCallArgs[0]
	if got.IssueID != "ab-prio" || got.Priority != 1 {
		t.Errorf("unexpected call args: %+v", got)
	}
}

func TestMockClient_UpdatePriority_InvokesStub(t *testing.T) {
	t.Parallel()

	want := errors.New("stub error")
	m := NewMockClient()
	m.UpdatePriorityFn = func(_ context.Context, _ string, _ int) error {
		return want
	}

	err := m.UpdatePriority(context.Background(), "ab-x", 3)
	if !errors.Is(err, want) {
		t.Errorf("expected stub error, got %v", err)
	}
}
