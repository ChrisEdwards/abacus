// Package beads provides client implementations for beads issue tracking.
//
// EVOLVING: brCLIClient is the active development backend for beads_rust (br).
// This client will evolve as br adds new features. Unlike bdCLIClient which is
// frozen at bd v0.38.0, changes and new features go here.
package beads

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	appErrors "abacus/internal/errors"
)

// brCLIClient implements Client for the beads_rust (br) CLI.
// Read methods are stubs (use brSQLiteClient for reads in production).
// Write methods delegate to the br CLI binary.
type brCLIClient struct {
	bin     string
	dbArgs  []string
	workDir string // working directory for br commands (br finds workspace by walking up from cwd)
}

// BrCLIOption configures the br CLI client implementation.
type BrCLIOption func(*brCLIClient)

// WithBrBinaryPath overrides the command used to invoke the br CLI.
func WithBrBinaryPath(path string) BrCLIOption {
	return func(c *brCLIClient) {
		if strings.TrimSpace(path) != "" {
			c.bin = path
		}
	}
}

// WithBrDatabasePath sets the beads database path for all br CLI invocations.
func WithBrDatabasePath(path string) BrCLIOption {
	return func(c *brCLIClient) {
		if trimmed := strings.TrimSpace(path); trimmed != "" {
			c.dbArgs = []string{"--db", trimmed}
		}
	}
}

// WithBrWorkDir sets the working directory for br CLI invocations.
// br finds its workspace by walking up from cwd looking for .beads/,
// so this must be set to a directory containing or under .beads/.
func WithBrWorkDir(dir string) BrCLIOption {
	return func(c *brCLIClient) {
		if trimmed := strings.TrimSpace(dir); trimmed != "" {
			c.workDir = trimmed
		}
	}
}

// NewBrCLIClient constructs a beads_rust CLI-backed Client implementation.
// Note: Read methods (List, Show, Export, Comments) return errors as they are
// not implemented. Use NewBrSQLiteClient for read operations via SQLite.
func NewBrCLIClient(opts ...BrCLIOption) Client {
	client := &brCLIClient{bin: "br"}
	for _, opt := range opts {
		opt(client)
	}
	return client
}

// List is not implemented for br CLI client - use SQLite client for reads.
func (c *brCLIClient) List(_ context.Context) ([]LiteIssue, error) {
	return nil, fmt.Errorf("List not implemented: br CLI client only supports write operations; use SQLite client for reads")
}

// Show is not implemented for br CLI client - use SQLite client for reads.
func (c *brCLIClient) Show(_ context.Context, _ []string) ([]FullIssue, error) {
	return nil, fmt.Errorf("Show not implemented: br CLI client only supports write operations; use SQLite client for reads")
}

// Export is not implemented for br CLI client - use SQLite client for reads.
func (c *brCLIClient) Export(_ context.Context) ([]FullIssue, error) {
	return nil, fmt.Errorf("Export not implemented: br CLI client only supports write operations; use SQLite client for reads")
}

// Comments is not implemented for br CLI client - use SQLite client for reads.
func (c *brCLIClient) Comments(_ context.Context, _ string) ([]Comment, error) {
	return nil, fmt.Errorf("Comments not implemented: br CLI client only supports write operations; use SQLite client for reads")
}

func (c *brCLIClient) UpdateStatus(ctx context.Context, issueID, newStatus string) error {
	if strings.TrimSpace(issueID) == "" {
		return fmt.Errorf("issue id is required for status update")
	}
	if strings.TrimSpace(newStatus) == "" {
		return fmt.Errorf("new status is required for status update")
	}
	_, err := c.run(ctx, "update", issueID, "--status="+newStatus)
	if err != nil {
		return fmt.Errorf("run br update: %w", err)
	}
	return nil
}

func (c *brCLIClient) Close(ctx context.Context, issueID string) error {
	if strings.TrimSpace(issueID) == "" {
		return fmt.Errorf("issue id is required for close")
	}
	_, err := c.run(ctx, "close", issueID)
	if err != nil {
		return fmt.Errorf("run br close: %w", err)
	}
	return nil
}

func (c *brCLIClient) Reopen(ctx context.Context, issueID string) error {
	if strings.TrimSpace(issueID) == "" {
		return fmt.Errorf("issue id is required for reopen")
	}
	_, err := c.run(ctx, "reopen", issueID)
	if err != nil {
		return fmt.Errorf("run br reopen: %w", err)
	}
	return nil
}

func (c *brCLIClient) AddLabel(ctx context.Context, issueID, label string) error {
	if strings.TrimSpace(issueID) == "" {
		return fmt.Errorf("issue id is required for add label")
	}
	if strings.TrimSpace(label) == "" {
		return fmt.Errorf("label is required for add label")
	}
	_, err := c.run(ctx, "label", "add", issueID, label)
	if err != nil {
		return fmt.Errorf("run br label add: %w", err)
	}
	return nil
}

func (c *brCLIClient) RemoveLabel(ctx context.Context, issueID, label string) error {
	if strings.TrimSpace(issueID) == "" {
		return fmt.Errorf("issue id is required for remove label")
	}
	if strings.TrimSpace(label) == "" {
		return fmt.Errorf("label is required for remove label")
	}
	_, err := c.run(ctx, "label", "remove", issueID, label)
	if err != nil {
		return fmt.Errorf("run br label remove: %w", err)
	}
	return nil
}

// Create creates a new issue and returns its ID.
// Uses positional title syntax: br create "Title" --type task
func (c *brCLIClient) Create(ctx context.Context, title, issueType string, priority int, labels []string, assignee string) (string, error) {
	if strings.TrimSpace(title) == "" {
		return "", fmt.Errorf("title is required for create")
	}
	if strings.TrimSpace(issueType) == "" {
		issueType = "task"
	}

	// Use positional title with JSON output for reliable parsing
	args := []string{
		"create",
		title, // Positional title
		"--type", issueType,
		"--priority", fmt.Sprintf("%d", priority),
		"--json",
	}

	if len(labels) > 0 {
		args = append(args, "--labels", strings.Join(labels, ","))
	}

	if strings.TrimSpace(assignee) != "" {
		args = append(args, "--assignee", assignee)
	}

	out, err := c.run(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("run br create: %w", err)
	}

	// Parse ID from JSON output (br may print warnings before the JSON)
	jsonBytes := extractJSON(out)
	if jsonBytes == nil {
		return "", fmt.Errorf("no JSON found in br create output: %s", strings.TrimSpace(string(out)))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return "", fmt.Errorf("decode br create output: %w", err)
	}
	if result.ID == "" {
		return "", fmt.Errorf("empty ID in br create output")
	}
	return result.ID, nil
}

// CreateFull creates a new issue with all fields and returns the full issue object.
// Uses positional title syntax: br create "Title" --type task --json
func (c *brCLIClient) CreateFull(ctx context.Context, title, issueType string, priority int, labels []string, assignee, description, parentID string) (FullIssue, error) {
	if strings.TrimSpace(title) == "" {
		return FullIssue{}, fmt.Errorf("title is required for create")
	}
	if strings.TrimSpace(issueType) == "" {
		issueType = "task"
	}

	// Use positional title (works for both bd and br)
	args := []string{
		"create",
		title, // Positional title
		"--type", issueType,
		"--priority", fmt.Sprintf("%d", priority),
		"--json",
	}

	if len(labels) > 0 {
		args = append(args, "--labels", strings.Join(labels, ","))
	}

	if strings.TrimSpace(assignee) != "" {
		args = append(args, "--assignee", assignee)
	}

	if strings.TrimSpace(description) != "" {
		args = append(args, "--description", description)
	}

	out, err := c.run(ctx, args...)
	if err != nil {
		return FullIssue{}, fmt.Errorf("run br create: %w", err)
	}

	// Extract JSON from output (br may print warnings before the JSON)
	jsonBytes := extractJSON(out)
	if jsonBytes == nil {
		snippet := string(out)
		if len(snippet) > maxErrorSnippetLen {
			snippet = snippet[:maxErrorSnippetLen] + "..."
		}
		return FullIssue{}, fmt.Errorf("no JSON found in br create output: %s", strings.TrimSpace(snippet))
	}

	var issue FullIssue
	if err := json.Unmarshal(jsonBytes, &issue); err != nil {
		snippet := string(out)
		if len(snippet) > maxErrorSnippetLen {
			snippet = snippet[:maxErrorSnippetLen] + "..."
		}
		return FullIssue{}, fmt.Errorf("decode br create output: %w (output: %s)", err, strings.TrimSpace(snippet))
	}

	// Add parent-child dependency if parent was specified
	if strings.TrimSpace(parentID) != "" {
		if err := c.AddDependency(ctx, issue.ID, parentID, "parent-child"); err != nil {
			return FullIssue{}, fmt.Errorf("add parent-child dependency: %w", err)
		}
	}

	return issue, nil
}

func (c *brCLIClient) UpdateFull(ctx context.Context, issueID, title, issueType string, priority int, labels []string, assignee, description string) error {
	if strings.TrimSpace(issueID) == "" {
		return fmt.Errorf("issue id is required for update")
	}
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("title is required for update")
	}

	args := []string{
		"update",
		issueID,
		"--title", title,
		"--description", description,
		"--priority", fmt.Sprintf("%d", priority),
		"--assignee", assignee, // Always pass to allow clearing (empty string clears)
	}

	// br only accepts a single --set-labels flag with comma-separated values
	// (unlike bd which accepts multiple flags). See upstream issue:
	// https://github.com/Dicklesworthstone/beads_rust/issues/17
	if len(labels) > 0 {
		args = append(args, "--set-labels", strings.Join(labels, ","))
	} else {
		args = append(args, "--set-labels", "")
	}

	if _, err := c.run(ctx, args...); err != nil {
		return fmt.Errorf("run br update: %w", err)
	}
	return nil
}

func (c *brCLIClient) AddDependency(ctx context.Context, fromID, toID, depType string) error {
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
		return fmt.Errorf("run br dep add: %w", err)
	}
	return nil
}

func (c *brCLIClient) RemoveDependency(ctx context.Context, fromID, toID, depType string) error {
	if strings.TrimSpace(fromID) == "" {
		return fmt.Errorf("from ID is required for remove dependency")
	}
	if strings.TrimSpace(toID) == "" {
		return fmt.Errorf("to ID is required for remove dependency")
	}
	args := []string{"dep", "remove", fromID, toID}
	if _, err := c.run(ctx, args...); err != nil {
		return fmt.Errorf("run br dep remove: %w", err)
	}
	return nil
}

func (c *brCLIClient) Delete(ctx context.Context, issueID string, cascade bool) error {
	if strings.TrimSpace(issueID) == "" {
		return fmt.Errorf("issue id is required for delete")
	}
	args := []string{"delete", issueID, "--force"}
	if cascade {
		args = append(args, "--cascade")
	}
	_, err := c.run(ctx, args...)
	if err != nil {
		return fmt.Errorf("run br delete: %w", err)
	}
	return nil
}

func (c *brCLIClient) AddComment(ctx context.Context, issueID, text string) error {
	if strings.TrimSpace(issueID) == "" {
		return fmt.Errorf("issue id is required for add comment")
	}
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("comment text is required")
	}
	_, err := c.run(ctx, "comments", "add", issueID, text)
	if err != nil {
		return fmt.Errorf("run br comments add: %w", err)
	}
	return nil
}

func (c *brCLIClient) run(ctx context.Context, args ...string) ([]byte, error) {
	finalArgs := make([]string, 0, len(c.dbArgs)+len(args))
	finalArgs = append(finalArgs, c.dbArgs...)
	finalArgs = append(finalArgs, args...)
	//nolint:gosec // G204: CLI wrapper intentionally shells out to br command
	cmd := exec.CommandContext(ctx, c.bin, finalArgs...)
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, formatBrCommandError(c.bin, finalArgs, err, out)
	}
	return out, nil
}

func formatBrCommandError(bin string, args []string, cmdErr error, out []byte) error {
	snippet := strings.TrimSpace(string(out))
	if len(snippet) > maxErrorSnippetLen {
		snippet = snippet[:maxErrorSnippetLen] + "..."
	}
	command := append([]string{bin}, args...)
	msg := fmt.Sprintf("%s failed", strings.Join(command, " "))
	err := classifyBrCLIError(command, appErrors.New(appErrors.CodeCLIFailed, msg, cmdErr), snippet)
	return err
}

// classifyBrCLIError classifies br CLI errors (reuses bd classification logic for now).
func classifyBrCLIError(command []string, err error, snippet string) error {
	// For now, reuse the same error classification as bd
	// Can be specialized for br-specific errors later
	return classifyCLIError(command, err, snippet)
}

// extractJSON finds and returns the first JSON object in the output.
// This is shared with bdCLIClient in cli.go - defined there.
// br commands may print warnings or other text before the actual JSON response.
var _ = extractJSON // Reference to ensure extractJSON from cli.go is used
