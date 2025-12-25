package ui

import "abacus/internal/beads"

type labelUpdateCompleteMsg struct {
	err error
}

type labelsToastTickMsg struct{}

// Message types for create operations
type createCompleteMsg struct {
	id        string
	err       error
	fullIssue *beads.FullIssue // full issue data for fast injection
	parentID  string           // Explicit parent context for fast injection
}

// Message types for update operations
type updateCompleteMsg struct {
	ID    string
	Title string
	Err   error
}

type createToastTickMsg struct{}

// Message types for delete operations
type deleteCompleteMsg struct {
	issueID  string
	children []string
	cascade  bool
	err      error
}

type deleteToastTickMsg struct{}

// Message types for comment operations
type commentCompleteMsg struct {
	issueID string
	err     error
}

type commentToastTickMsg struct{}
