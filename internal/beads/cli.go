package beads

import (
	"bufio"
	"bytes"
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

func (c *cliClient) Create(ctx context.Context, title, issueType string, priority int, labels []string, assignee string) (string, error) {
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

func (c *cliClient) CreateFull(ctx context.Context, title, issueType string, priority int, labels []string, assignee, description, parentID string) (FullIssue, error) {
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

func (c *cliClient) UpdateFull(ctx context.Context, issueID, title, issueType string, priority int, labels []string, assignee, description string) error {
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
	}

	if strings.TrimSpace(assignee) != "" {
		args = append(args, "--assignee", assignee)
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

func (c *cliClient) RemoveDependency(ctx context.Context, fromID, toID, depType string) error {
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

func (c *cliClient) Delete(ctx context.Context, issueID string, cascade bool) error {
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

// exportIssue handles the nested dependency format difference from bd export.
// bd export uses "depends_on_id" and "type" while bd show uses "id" and "dependency_type".
type exportIssue struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	Status             string   `json:"status"`
	Priority           int      `json:"priority"`
	IssueType          string   `json:"issue_type"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
	ClosedAt           string   `json:"closed_at"`
	Assignee           string   `json:"assignee"`
	Labels             []string `json:"labels"`
	ExternalRef        string   `json:"external_ref"`
	Design             string   `json:"design"`
	AcceptanceCriteria string   `json:"acceptance_criteria"`
	Notes              string   `json:"notes"`
	Dependencies       []struct {
		DependsOnID string `json:"depends_on_id"`
		Type        string `json:"type"`
	} `json:"dependencies"`
}

func (c *cliClient) Export(ctx context.Context) ([]FullIssue, error) {
	output, err := c.run(ctx, "export")
	if err != nil {
		return nil, fmt.Errorf("bd export: %w", err)
	}

	var issues []FullIssue
	scanner := bufio.NewScanner(bytes.NewReader(output))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		var raw exportIssue
		if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
			return nil, fmt.Errorf("parse export line %d: %w", lineNum, err)
		}
		issue, err := convertExportIssue(raw, lineNum)
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read export stream: %w", err)
	}
	if len(issues) == 0 {
		return nil, fmt.Errorf("bd export returned no issues")
	}
	return issues, nil
}

func convertExportIssue(raw exportIssue, lineNum int) (FullIssue, error) {
	if raw.ID == "" {
		return FullIssue{}, fmt.Errorf("export line %d: missing issue ID", lineNum)
	}

	// Use append to avoid zero-value entries when skipping invalid dependencies
	deps := make([]Dependency, 0, len(raw.Dependencies))
	for _, d := range raw.Dependencies {
		if d.DependsOnID == "" {
			continue // Skip invalid dependencies
		}
		deps = append(deps, Dependency{
			TargetID: d.DependsOnID,
			Type:     d.Type,
		})
	}

	return FullIssue{
		ID:                 raw.ID,
		Title:              raw.Title,
		Description:        raw.Description,
		Status:             raw.Status,
		Priority:           raw.Priority,
		IssueType:          raw.IssueType,
		CreatedAt:          raw.CreatedAt,
		UpdatedAt:          raw.UpdatedAt,
		ClosedAt:           raw.ClosedAt,
		Assignee:           raw.Assignee,
		Labels:             raw.Labels,
		ExternalRef:        raw.ExternalRef,
		Design:             raw.Design,
		AcceptanceCriteria: raw.AcceptanceCriteria,
		Notes:              raw.Notes,
		Dependencies:       deps,
		// Dependents not populated - graph builder derives Children from Dependencies
	}, nil
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

// extractJSON finds and returns the first JSON object in the output.
// bd commands may print warnings or other text before the actual JSON response.
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
