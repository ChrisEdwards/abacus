// Package beads provides client implementations for beads issue tracking.
//
// FROZEN: bdCLIClient is frozen at bd v0.38.0 compatibility.
// Do NOT add new features or modify behavior - only critical bug fixes allowed.
// All new development should go to brCLIClient in br_cli.go instead.
package beads

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	appErrors "abacus/internal/errors"
)

const maxErrorSnippetLen = 200

// bdCLIClient implements Writer for the beads Go (bd) CLI.
// FROZEN: Only critical bug fixes allowed. New features go to brCLIClient.
type bdCLIClient struct {
	bin    string
	dbArgs []string
}

// BdCLIOption configures the bd CLI client implementation.
type BdCLIOption func(*bdCLIClient)

// WithBdBinaryPath overrides the command used to invoke the bd CLI.
func WithBdBinaryPath(path string) BdCLIOption {
	return func(c *bdCLIClient) {
		if strings.TrimSpace(path) != "" {
			c.bin = path
		}
	}
}

// WithBdDatabasePath sets the Beads database path for all bd CLI invocations.
func WithBdDatabasePath(path string) BdCLIOption {
	return func(c *bdCLIClient) {
		if trimmed := strings.TrimSpace(path); trimmed != "" {
			c.dbArgs = []string{"--db", trimmed}
		}
	}
}

// NewBdCLIClient constructs a beads (bd) CLI-backed Writer implementation.
// FROZEN: Only implements Writer. Use NewBdSQLiteClient for full Client.
func NewBdCLIClient(opts ...BdCLIOption) Writer {
	client := &bdCLIClient{bin: "bd"}
	for _, opt := range opts {
		opt(client)
	}
	return client
}

func (c *bdCLIClient) UpdateStatus(ctx context.Context, issueID, newStatus string) error {
	if strings.TrimSpace(issueID) == "" {
		return fmt.Errorf("issue id is required for status update")
	}
	if strings.TrimSpace(newStatus) == "" {
		return fmt.Errorf("new status is required for status update")
	}
	_, err := c.run(ctx, "update", issueID, "--status="+newStatus)
	if err != nil {
		return fmt.Errorf("run bd update: %w", err)
	}
	return nil
}

func (c *bdCLIClient) Close(ctx context.Context, issueID string) error {
	if strings.TrimSpace(issueID) == "" {
		return fmt.Errorf("issue id is required for close")
	}
	_, err := c.run(ctx, "close", issueID)
	if err != nil {
		return fmt.Errorf("run bd close: %w", err)
	}
	return nil
}

func (c *bdCLIClient) Reopen(ctx context.Context, issueID string) error {
	if strings.TrimSpace(issueID) == "" {
		return fmt.Errorf("issue id is required for reopen")
	}
	_, err := c.run(ctx, "reopen", issueID)
	if err != nil {
		return fmt.Errorf("run bd reopen: %w", err)
	}
	return nil
}

func (c *bdCLIClient) AddLabel(ctx context.Context, issueID, label string) error {
	if strings.TrimSpace(issueID) == "" {
		return fmt.Errorf("issue id is required for add label")
	}
	if strings.TrimSpace(label) == "" {
		return fmt.Errorf("label is required for add label")
	}
	_, err := c.run(ctx, "label", "add", issueID, label)
	if err != nil {
		return fmt.Errorf("run bd label add: %w", err)
	}
	return nil
}

func (c *bdCLIClient) RemoveLabel(ctx context.Context, issueID, label string) error {
	if strings.TrimSpace(issueID) == "" {
		return fmt.Errorf("issue id is required for remove label")
	}
	if strings.TrimSpace(label) == "" {
		return fmt.Errorf("label is required for remove label")
	}
	_, err := c.run(ctx, "label", "remove", issueID, label)
	if err != nil {
		return fmt.Errorf("run bd label remove: %w", err)
	}
	return nil
}

func (c *bdCLIClient) Create(ctx context.Context, title, issueType string, priority int, labels []string, assignee string) (string, error) {
	if strings.TrimSpace(title) == "" {
		return "", fmt.Errorf("title is required for create")
	}
	if strings.TrimSpace(issueType) == "" {
		issueType = "task"
	}

	args := []string{
		"create",
		"--title", title,
		"--type", issueType,
		"--priority", fmt.Sprintf("%d", priority),
	}

	// Add labels if provided
	if len(labels) > 0 {
		args = append(args, "--labels", strings.Join(labels, ","))
	}

	// Add assignee if provided
	if strings.TrimSpace(assignee) != "" {
		args = append(args, "--assignee", assignee)
	}

	out, err := c.run(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("run bd create: %w", err)
	}
	// Parse the new bead ID from output
	// bd outputs: "✓ Created issue: ab-xyz" or "✓ Created issue: test-xyz"
	output := string(out)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		// Look for "Created issue: <id>" pattern
		if idx := strings.Index(line, "Created issue:"); idx != -1 {
			idPart := strings.TrimSpace(line[idx+len("Created issue:"):])
			// Take just the ID (first word, no trailing punctuation)
			if spaceIdx := strings.IndexAny(idPart, " \t"); spaceIdx != -1 {
				idPart = idPart[:spaceIdx]
			}
			return strings.TrimRight(idPart, ".,;:!"), nil
		}
	}
	// Fallback: look for issue ID pattern (ab-xxx or test-xxx prefix)
	for _, line := range strings.Split(output, "\n") {
		for _, part := range strings.Fields(line) {
			// Only match actual issue ID prefixes, not flags like --description
			if strings.HasPrefix(part, "ab-") || strings.HasPrefix(part, "test-") {
				return strings.TrimRight(part, ".,;:!"), nil
			}
		}
	}
	return "", fmt.Errorf("could not parse bead ID from output: %s", output)
}

func (c *bdCLIClient) CreateFull(ctx context.Context, title, issueType string, priority int, labels []string, assignee, description, parentID string) (FullIssue, error) {
	if strings.TrimSpace(title) == "" {
		return FullIssue{}, fmt.Errorf("title is required for create")
	}
	if strings.TrimSpace(issueType) == "" {
		issueType = "task"
	}

	args := []string{
		"create",
		"--title", title,
		"--type", issueType,
		"--priority", fmt.Sprintf("%d", priority),
		"--json",
	}

	// Add labels if provided
	if len(labels) > 0 {
		args = append(args, "--labels", strings.Join(labels, ","))
	}

	// Add assignee if provided
	if strings.TrimSpace(assignee) != "" {
		args = append(args, "--assignee", assignee)
	}

	// Add description if provided
	if strings.TrimSpace(description) != "" {
		args = append(args, "--description", description)
	}

	// Note: We don't pass --parent to bd create because that generates dotted IDs
	// (e.g., ab-kr7.1). Instead, we create the bead first with a random ID,
	// then add the parent-child dependency separately.

	out, err := c.run(ctx, args...)
	if err != nil {
		return FullIssue{}, fmt.Errorf("run bd create: %w", err)
	}

	// Extract JSON from output (bd may print warnings before the JSON)
	jsonBytes := extractJSON(out)
	if jsonBytes == nil {
		snippet := string(out)
		if len(snippet) > maxErrorSnippetLen {
			snippet = snippet[:maxErrorSnippetLen] + "..."
		}
		return FullIssue{}, fmt.Errorf("no JSON found in bd create output: %s", strings.TrimSpace(snippet))
	}

	// Parse JSON response
	var issue FullIssue
	if err := json.Unmarshal(jsonBytes, &issue); err != nil {
		snippet := string(out)
		if len(snippet) > maxErrorSnippetLen {
			snippet = snippet[:maxErrorSnippetLen] + "..."
		}
		return FullIssue{}, fmt.Errorf("decode bd create output: %w (output: %s)", err, strings.TrimSpace(snippet))
	}

	// Add parent-child dependency if parent was specified
	// This creates the relationship without using dotted IDs
	if strings.TrimSpace(parentID) != "" {
		if err := c.AddDependency(ctx, issue.ID, parentID, "parent-child"); err != nil {
			return FullIssue{}, fmt.Errorf("add parent-child dependency: %w", err)
		}
	}

	return issue, nil
}

func (c *bdCLIClient) UpdateFull(ctx context.Context, issueID, title, issueType string, priority int, labels []string, assignee, description string) error {
	if strings.TrimSpace(issueID) == "" {
		return fmt.Errorf("issue id is required for update")
	}
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("title is required for update")
	}
	// issueType is currently not configurable via `bd update`; keep parameter for future compatibility.

	args := []string{
		"update",
		issueID,
		"--title", title,
		"--description", description,
		"--priority", fmt.Sprintf("%d", priority),
		"--assignee", assignee, // Always pass to allow clearing (empty string clears)
	}

	if len(labels) > 0 {
		for _, l := range labels {
			args = append(args, "--set-labels", l)
		}
	} else {
		// Explicitly clear labels when none are provided
		args = append(args, "--set-labels", "")
	}

	if _, err := c.run(ctx, args...); err != nil {
		return fmt.Errorf("run bd update: %w", err)
	}
	return nil
}

func (c *bdCLIClient) AddDependency(ctx context.Context, fromID, toID, depType string) error {
	if strings.TrimSpace(fromID) == "" {
		return fmt.Errorf("from ID is required for add dependency")
	}
	if strings.TrimSpace(toID) == "" {
		return fmt.Errorf("to ID is required for add dependency")
	}
	if strings.TrimSpace(depType) == "" {
		depType = "blocks"
	}
	_, err := c.run(ctx, "dep", "add", fromID, toID, "--type", depType)
	if err != nil {
		return fmt.Errorf("run bd dep add: %w", err)
	}
	return nil
}

func (c *bdCLIClient) RemoveDependency(ctx context.Context, fromID, toID, depType string) error {
	if strings.TrimSpace(fromID) == "" {
		return fmt.Errorf("from ID is required for remove dependency")
	}
	if strings.TrimSpace(toID) == "" {
		return fmt.Errorf("to ID is required for remove dependency")
	}
	args := []string{"dep", "remove", fromID, toID}
	if _, err := c.run(ctx, args...); err != nil {
		return fmt.Errorf("run bd dep remove: %w", err)
	}
	return nil
}

func (c *bdCLIClient) Delete(ctx context.Context, issueID string, cascade bool) error {
	if strings.TrimSpace(issueID) == "" {
		return fmt.Errorf("issue id is required for delete")
	}
	args := []string{"delete", issueID, "--force"}
	if cascade {
		args = append(args, "--cascade")
	}
	_, err := c.run(ctx, args...)
	if err != nil {
		return fmt.Errorf("run bd delete: %w", err)
	}
	return nil
}

func (c *bdCLIClient) AddComment(ctx context.Context, issueID, text string) error {
	if strings.TrimSpace(issueID) == "" {
		return fmt.Errorf("issue id is required for add comment")
	}
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("comment text is required")
	}
	_, err := c.run(ctx, "comments", "add", issueID, text)
	if err != nil {
		return fmt.Errorf("run bd comments add: %w", err)
	}
	return nil
}

func (c *bdCLIClient) run(ctx context.Context, args ...string) ([]byte, error) {
	finalArgs := make([]string, 0, len(c.dbArgs)+len(args))
	finalArgs = append(finalArgs, c.dbArgs...)
	finalArgs = append(finalArgs, args...)
	//nolint:gosec // G204: CLI wrapper intentionally shells out to bd command
	cmd := exec.CommandContext(ctx, c.bin, finalArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, formatCommandError(c.bin, finalArgs, err, out)
	}
	return out, nil
}

func formatCommandError(bin string, args []string, cmdErr error, out []byte) error {
	snippet := strings.TrimSpace(string(out))
	if len(snippet) > maxErrorSnippetLen {
		snippet = snippet[:maxErrorSnippetLen] + "..."
	}
	command := append([]string{bin}, args...)
	msg := fmt.Sprintf("%s failed", strings.Join(command, " "))
	err := classifyCLIError(command, appErrors.New(appErrors.CodeCLIFailed, msg, cmdErr), snippet)
	return err
}

// extractJSON finds and returns the first JSON object in the output.
// bd/br commands may print warnings or other text before the actual JSON response.
// This function is shared between bdCLIClient and brCLIClient.
func extractJSON(out []byte) []byte {
	// Find the first '{' which starts the JSON object
	start := bytes.IndexByte(out, '{')
	if start == -1 {
		return nil
	}
	// Find the matching closing brace by counting braces
	depth := 0
	for i := start; i < len(out); i++ {
		switch out[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return out[start : i+1]
			}
		}
	}
	// No matching closing brace found, return from start to end
	return out[start:]
}
