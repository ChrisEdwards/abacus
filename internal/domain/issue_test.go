package domain

import (
	"testing"

	"abacus/internal/beads"
)

func TestNewIssueValidation(t *testing.T) {
	issue, err := NewIssue("ab-1", "Alpha", "desc", StatusOpen, PriorityHigh, []string{"ready"}, false)
	if err != nil {
		t.Fatalf("expected issue to be created, got error: %v", err)
	}
	if issue.Status() != StatusOpen {
		t.Fatalf("expected status open, got %s", issue.Status())
	}
	if issue.PriorityValue() != PriorityHigh {
		t.Fatalf("expected priority high, got %d", issue.PriorityValue())
	}

	if _, err := NewIssue("", "missing id", "", StatusOpen, PriorityHigh, nil, false); err == nil {
		t.Fatalf("expected error for missing id")
	}
	if _, err := NewIssue("ab-2", "bad status", "", Status("weird"), PriorityHigh, nil, false); err == nil {
		t.Fatalf("expected error for invalid status")
	}
	if _, err := NewIssue("ab-3", "bad priority", "", StatusOpen, Priority(-1), nil, false); err == nil {
		t.Fatalf("expected error for invalid priority")
	}
}

func TestIssueBusinessLogic(t *testing.T) {
	issue, err := NewIssue("ab-1", "Blocked", "", StatusOpen, PriorityCritical, nil, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issue.IsReady() {
		t.Fatalf("blocked issue should not be ready")
	}
	if !issue.IsBlocked() {
		t.Fatalf("expected issue to be blocked")
	}

	if err := issue.CanTransitionTo(StatusClosed); err != nil {
		t.Fatalf("expected transition to closed, got %v", err)
	}
	if err := issue.CanTransitionTo(Status("unknown")); err == nil {
		t.Fatalf("expected error for invalid target status")
	}
}

func TestNewIssueFromFull(t *testing.T) {
	full := beads.FullIssue{
		ID:       "ab-42",
		Title:    "Meaning",
		Status:   "open",
		Priority: 2,
		Labels:   []string{"deep"},
	}
	issue, err := NewIssueFromFull(full, false)
	if err != nil {
		t.Fatalf("NewIssueFromFull returned error: %v", err)
	}
	if !issue.IsReady() {
		t.Fatalf("expected issue to be ready")
	}

	full.Status = "invalid"
	if _, err := NewIssueFromFull(full, false); err == nil {
		t.Fatalf("expected error for invalid status")
	}
}
