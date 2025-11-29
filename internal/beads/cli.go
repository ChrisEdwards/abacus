package beads

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	appErrors "abacus/internal/errors"
)

const maxErrorSnippetLen = 200

type cliClient struct {
	bin    string
	dbArgs []string
}

// CLIOption configures the CLI client implementation.
type CLIOption func(*cliClient)

// WithBinaryPath overrides the command used to invoke the Beads CLI.
func WithBinaryPath(path string) CLIOption {
	return func(c *cliClient) {
		if strings.TrimSpace(path) != "" {
			c.bin = path
		}
	}
}

// WithDatabasePath sets the Beads database path for all CLI invocations.
func WithDatabasePath(path string) CLIOption {
	return func(c *cliClient) {
		if trimmed := strings.TrimSpace(path); trimmed != "" {
			c.dbArgs = []string{"--db", trimmed}
		}
	}
}

// NewCLIClient constructs a Beads CLI-backed client implementation.
func NewCLIClient(opts ...CLIOption) Client {
	client := &cliClient{bin: "bd"}
	for _, opt := range opts {
		opt(client)
	}
	return client
}

func (c *cliClient) List(ctx context.Context) ([]LiteIssue, error) {
	out, err := c.run(ctx, "list", "--json")
	if err != nil {
		return nil, fmt.Errorf("run bd list: %w", err)
	}
	var issues []LiteIssue
	if err := json.Unmarshal(out, &issues); err != nil {
		return nil, fmt.Errorf("decode bd list output: %w", err)
	}
	if issues == nil {
		issues = []LiteIssue{}
	}
	return issues, nil
}

func (c *cliClient) Show(ctx context.Context, ids []string) ([]FullIssue, error) {
	if len(ids) == 0 {
		return []FullIssue{}, nil
	}
	args := append([]string{"show"}, ids...)
	args = append(args, "--json")
	out, err := c.run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("run bd show: %w", err)
	}
	var issues []FullIssue
	if err := json.Unmarshal(out, &issues); err != nil {
		snippet := string(out)
		if len(snippet) > maxErrorSnippetLen {
			snippet = snippet[:maxErrorSnippetLen] + "..."
		}
		return nil, fmt.Errorf("decode bd show output: %w (output: %s)", err, strings.TrimSpace(snippet))
	}
	if issues == nil {
		issues = []FullIssue{}
	}
	return issues, nil
}

func (c *cliClient) Comments(ctx context.Context, issueID string) ([]Comment, error) {
	if strings.TrimSpace(issueID) == "" {
		return nil, fmt.Errorf("issue id is required for comments")
	}
	out, err := c.run(ctx, "comments", issueID, "--json")
	if err != nil {
		return nil, fmt.Errorf("run bd comments: %w", err)
	}
	var comments []Comment
	if err := json.Unmarshal(out, &comments); err != nil {
		return nil, fmt.Errorf("decode bd comments output: %w", err)
	}
	if comments == nil {
		comments = []Comment{}
	}
	return comments, nil
}

func (c *cliClient) UpdateStatus(ctx context.Context, issueID, newStatus string) error {
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

func (c *cliClient) Close(ctx context.Context, issueID string) error {
	if strings.TrimSpace(issueID) == "" {
		return fmt.Errorf("issue id is required for close")
	}
	_, err := c.run(ctx, "close", issueID)
	if err != nil {
		return fmt.Errorf("run bd close: %w", err)
	}
	return nil
}

func (c *cliClient) Reopen(ctx context.Context, issueID string) error {
	if strings.TrimSpace(issueID) == "" {
		return fmt.Errorf("issue id is required for reopen")
	}
	_, err := c.run(ctx, "reopen", issueID)
	if err != nil {
		return fmt.Errorf("run bd reopen: %w", err)
	}
	return nil
}

func (c *cliClient) AddLabel(ctx context.Context, issueID, label string) error {
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

func (c *cliClient) RemoveLabel(ctx context.Context, issueID, label string) error {
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

func (c *cliClient) Create(ctx context.Context, title, issueType string, priority int) (string, error) {
	if strings.TrimSpace(title) == "" {
		return "", fmt.Errorf("title is required for create")
	}
	if strings.TrimSpace(issueType) == "" {
		issueType = "task"
	}
	out, err := c.run(ctx, "create",
		"--title", title,
		"--type", issueType,
		"--priority", fmt.Sprintf("%d", priority),
	)
	if err != nil {
		return "", fmt.Errorf("run bd create: %w", err)
	}
	// Parse the new bead ID from output (e.g., "Created ab-xyz")
	output := strings.TrimSpace(string(out))
	// Look for pattern like "Created ab-xxx" or just "ab-xxx"
	parts := strings.Fields(output)
	for _, part := range parts {
		if strings.HasPrefix(part, "ab-") || strings.Contains(part, "-") {
			// Clean up any trailing punctuation
			id := strings.TrimRight(part, ".,;:!")
			if len(id) > 0 {
				return id, nil
			}
		}
	}
	// If we can't parse an ID, return the raw output (caller can handle)
	if len(parts) > 0 {
		return parts[len(parts)-1], nil
	}
	return "", fmt.Errorf("could not parse bead ID from output: %s", output)
}

func (c *cliClient) AddDependency(ctx context.Context, fromID, toID, depType string) error {
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

func (c *cliClient) run(ctx context.Context, args ...string) ([]byte, error) {
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
