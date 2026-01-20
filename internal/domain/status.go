package domain

import "strings"

// Status represents the lifecycle state of an issue.
type Status string

const (
	StatusUnknown    Status = ""
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusBlocked    Status = "blocked"
	StatusDeferred   Status = "deferred"
	StatusClosed     Status = "closed"
	// StatusTombstone is internal only - not in validStatuses.
	StatusTombstone Status = "tombstone"
)

var validStatuses = map[Status]struct{}{
	StatusOpen:       {},
	StatusInProgress: {},
	StatusBlocked:    {},
	StatusDeferred:   {},
	StatusClosed:     {},
}

var allowedTransitions = map[Status]map[Status]struct{}{
	StatusOpen: {
		StatusInProgress: {},
		StatusBlocked:    {},
		StatusDeferred:   {},
		StatusClosed:     {},
	},
	StatusInProgress: {
		StatusOpen:     {},
		StatusBlocked:  {},
		StatusDeferred: {},
		StatusClosed:   {},
	},
	StatusBlocked: {
		StatusOpen:       {},
		StatusInProgress: {},
		StatusDeferred:   {},
		StatusClosed:     {},
	},
	StatusDeferred: {
		StatusOpen:       {},
		StatusInProgress: {},
		StatusBlocked:    {},
		StatusClosed:     {},
	},
	StatusClosed: {
		StatusOpen: {},
	},
}

// ParseStatus normalises an incoming status string.
// It accepts any non-empty status, including unknown values from newer backends.
// Tombstone is rejected as it is an internal-only status.
// Use IsKnown() to check if the status is one we recognize.
func ParseStatus(raw string) (Status, error) {
	status := Status(strings.ToLower(strings.TrimSpace(raw)))
	if status == StatusUnknown {
		return StatusUnknown, invalidStatusError("blank")
	}
	if status == StatusTombstone {
		return StatusUnknown, invalidStatusError(raw)
	}
	return status, nil
}

// IsKnown returns true if the status is one of the known status values.
// Unknown statuses (e.g., "pinned" from br) are accepted but return false.
func (s Status) IsKnown() bool {
	_, ok := validStatuses[s]
	return ok
}

// Validate ensures the status is part of the supported workflow.
func (s Status) Validate() error {
	if _, ok := validStatuses[s]; !ok {
		return invalidStatusError(string(s))
	}
	return nil
}

// IsTerminal reports whether the status represents a finished issue.
func (s Status) IsTerminal() bool {
	return s == StatusClosed
}

// CanTransitionTo verifies whether a transition to the target status is allowed.
func (s Status) CanTransitionTo(target Status) error {
	if err := s.Validate(); err != nil {
		return err
	}
	if err := target.Validate(); err != nil {
		return err
	}
	if s == target {
		return nil
	}
	if transitions, ok := allowedTransitions[s]; ok {
		if _, allowed := transitions[target]; allowed {
			return nil
		}
	}
	return invalidTransitionError(s, target)
}
