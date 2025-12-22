package beads

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

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
