// Package beads provides client implementations for beads issue tracking.
//
// DEPRECATED: This file provides backward-compatible wrappers for the bd backend.
// New code should use NewBdSQLiteClient directly, or use the backend detection
// in backend.go to get the appropriate client.
package beads

import "strings"

// NewSQLiteClient constructs a client that reads via SQLite and writes via the CLI.
// Deprecated: Use NewBdSQLiteClient for bd, or the backend factory for auto-detection.
// If dbPath is empty, it falls back to a pure CLI client (for backward compat).
func NewSQLiteClient(dbPath string, opts ...CLIOption) Client {
	trimmed := strings.TrimSpace(dbPath)
	if trimmed == "" {
		// Fallback for backward compatibility - return CLI wrapper
		return NewCLIClient(opts...)
	}
	return NewBdSQLiteClient(dbPath, opts...)
}
