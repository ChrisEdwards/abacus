package domain

import "abacus/internal/beads"

// Priority expresses scheduling urgency with lower numbers being higher priority.
type Priority int

const (
	PriorityCritical Priority = 0
	PriorityHigh     Priority = 1
	PriorityMedium   Priority = 2
	PriorityLow      Priority = 3
	PriorityBacklog  Priority = 4
)

// Validate ensures the priority falls within the supported range.
func (p Priority) Validate() error {
	if p < PriorityCritical || p > PriorityBacklog {
		return invalidPriorityError(int(p))
	}
	return nil
}

// Issue encapsulates the business behaviour around Beads issues.
//
// Business rules enforced:
//   - Status values must be part of the supported workflow (open -> in_progress -> closed).
//   - Only open, unblocked issues are considered "ready".
//   - Status transitions must respect the allowed workflow.
type Issue struct {
	id          string
	title       string
	description string
	status      Status
	priority    Priority
	labels      []string
	blocked     bool
}

// NewIssue constructs an Issue from primitives, enforcing validation.
func NewIssue(id, title, description string, status Status, priority Priority, labels []string, blocked bool) (Issue, error) {
	if id == "" {
		return Issue{}, invalidIssueError("issue id is required", nil)
	}
	if err := status.Validate(); err != nil {
		return Issue{}, err
	}
	if err := priority.Validate(); err != nil {
		return Issue{}, err
	}
	copiedLabels := append([]string(nil), labels...)
	return Issue{
		id:          id,
		title:       title,
		description: description,
		status:      status,
		priority:    priority,
		labels:      copiedLabels,
		blocked:     blocked,
	}, nil
}

// NewIssueFromFull converts a beads.FullIssue into a domain Issue.
func NewIssueFromFull(full beads.FullIssue, blocked bool) (Issue, error) {
	status, err := ParseStatus(full.Status)
	if err != nil {
		return Issue{}, err
	}
	priority := Priority(full.Priority)
	if err := priority.Validate(); err != nil {
		return Issue{}, err
	}
	return NewIssue(
		full.ID,
		full.Title,
		full.Description,
		status,
		priority,
		full.Labels,
		blocked,
	)
}

// ID returns the immutable identifier.
func (i Issue) ID() string {
	return i.id
}

// Title returns the issue title.
func (i Issue) Title() string {
	return i.title
}

// Status returns the current lifecycle status.
func (i Issue) Status() Status {
	return i.status
}

// PriorityValue returns the scheduled priority.
func (i Issue) PriorityValue() Priority {
	return i.priority
}

// Labels returns a copy of the associated labels.
func (i Issue) Labels() []string {
	return append([]string(nil), i.labels...)
}

// IsBlocked reports whether the issue is currently blocked by another item.
func (i Issue) IsBlocked() bool {
	return i.blocked
}

// IsReady reports whether the issue can be started.
func (i Issue) IsReady() bool {
	return i.status == StatusOpen && !i.blocked
}

// CanTransitionTo validates the requested status change.
func (i Issue) CanTransitionTo(target Status) error {
	return i.status.CanTransitionTo(target)
}
