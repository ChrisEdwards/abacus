package domain

import (
	"fmt"

	appErrors "abacus/internal/errors"
)

func invalidStatusError(status string) error {
	return appErrors.New(appErrors.CodeInvalidStatus, fmt.Sprintf("invalid status: %s", status), nil)
}

func invalidPriorityError(priority int) error {
	return appErrors.New(appErrors.CodeInvalidPriority, fmt.Sprintf("invalid priority: %d", priority), nil)
}

func invalidTransitionError(from, to Status) error {
	return appErrors.New(appErrors.CodeInvalidTransition, fmt.Sprintf("cannot transition from %s to %s", from, to), nil)
}

func invalidIssueError(reason string, err error) error {
	return appErrors.New(appErrors.CodeInvalidIssueData, reason, err)
}
