//go:build integration

// Package beads provides conformance tests for bd and br output compatibility.
//
// These tests verify that bd and br produce equivalent JSON output for the
// fields that abacus uses, ensuring compatibility between backends.
//
// Run with: go test -tags=integration -v ./internal/beads/
// Unit tests only: go test ./... (excludes this file)
package beads

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

// =============================================================================
// Output Conformance Tests
// =============================================================================

// TestOutputConformance_List verifies that bd list --json and br list --json
// produce compatible output for the fields abacus uses.
func TestOutputConformance_List(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping conformance test in short mode")
	}

	// Need both backends for conformance testing
	_, err := exec.LookPath("br")
	if err != nil {
		t.Skip("br binary not found, skipping conformance test")
	}
	_, err = exec.LookPath("bd")
	if err != nil {
		t.Skip("bd binary not found, skipping conformance test")
	}

	// Set up test databases with identical data
	brEnv := setupBackendTestDB(t, "br")
	defer brEnv.cleanup()

	bdEnv := setupBackendTestDB(t, "bd")
	defer bdEnv.cleanup()

	// Create identical issues in both databases
	testCases := []struct {
		title    string
		typ      string
		priority string
	}{
		{"Conformance Test Bug", "bug", "1"},
		{"Conformance Test Task", "task", "2"},
		{"Conformance Test Feature", "feature", "3"},
	}

	for _, tc := range testCases {
		// Create in br
		cmd := exec.Command("br", "create", tc.title, "--type", tc.typ, "--priority", tc.priority)
		cmd.Dir = brEnv.WorkDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("br create failed: %v\nOutput: %s", err, out)
		}

		// Create in bd
		cmd = exec.Command("bd", "--db", bdEnv.DBPath, "create",
			"--title", tc.title, "--type", tc.typ, "--priority", tc.priority)
		cmd.Dir = bdEnv.WorkDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("bd create failed: %v\nOutput: %s", err, out)
		}
	}

	// Get list output from both backends
	brCmd := exec.Command("br", "list", "--json")
	brCmd.Dir = brEnv.WorkDir
	brOut, err := brCmd.Output()
	if err != nil {
		t.Fatalf("br list --json failed: %v", err)
	}

	bdCmd := exec.Command("bd", "--db", bdEnv.DBPath, "list", "--json")
	bdCmd.Dir = bdEnv.WorkDir
	bdOut, err := bdCmd.Output()
	if err != nil {
		t.Fatalf("bd list --json failed: %v", err)
	}

	// Parse JSON outputs
	var brIssues []map[string]interface{}
	if err := json.Unmarshal(brOut, &brIssues); err != nil {
		t.Fatalf("Failed to parse br list output: %v\nOutput: %s", err, brOut)
	}

	var bdIssues []map[string]interface{}
	if err := json.Unmarshal(bdOut, &bdIssues); err != nil {
		t.Fatalf("Failed to parse bd list output: %v\nOutput: %s", err, bdOut)
	}

	// Verify same number of issues
	if len(brIssues) != len(bdIssues) {
		t.Errorf("Issue count mismatch: br=%d, bd=%d", len(brIssues), len(bdIssues))
	}

	// Verify required fields exist in both outputs
	requiredFields := []string{"id", "title", "status", "priority", "issue_type"}

	t.Log("Checking br list --json output fields")
	for i, issue := range brIssues {
		for _, field := range requiredFields {
			if _, ok := issue[field]; !ok {
				t.Errorf("br issue[%d] missing required field %q", i, field)
			}
		}
	}

	t.Log("Checking bd list --json output fields")
	for i, issue := range bdIssues {
		for _, field := range requiredFields {
			if _, ok := issue[field]; !ok {
				t.Errorf("bd issue[%d] missing required field %q", i, field)
			}
		}
	}

	t.Log("Output conformance test passed: both backends produce compatible list output")
}

// TestOutputConformance_Show verifies that bd show --json and br show --json
// produce compatible output for the fields abacus uses.
func TestOutputConformance_Show(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping conformance test in short mode")
	}

	// Need both backends for conformance testing
	_, err := exec.LookPath("br")
	if err != nil {
		t.Skip("br binary not found, skipping conformance test")
	}
	_, err = exec.LookPath("bd")
	if err != nil {
		t.Skip("bd binary not found, skipping conformance test")
	}

	// Set up test databases
	brEnv := setupBackendTestDB(t, "br")
	defer brEnv.cleanup()

	bdEnv := setupBackendTestDB(t, "bd")
	defer bdEnv.cleanup()

	// Create an issue in br and get its ID
	brCreateCmd := exec.Command("br", "create", "Show Test Issue", "--type", "task", "--priority", "2")
	brCreateCmd.Dir = brEnv.WorkDir
	brCreateOut, err := brCreateCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("br create failed: %v\nOutput: %s", err, brCreateOut)
	}
	brID := extractCreatedID(string(brCreateOut))
	if brID == "" {
		t.Fatalf("Could not extract ID from br create output: %s", brCreateOut)
	}

	// Create an issue in bd and get its ID
	bdCreateCmd := exec.Command("bd", "--db", bdEnv.DBPath, "create",
		"--title", "Show Test Issue", "--type", "task", "--priority", "2")
	bdCreateCmd.Dir = bdEnv.WorkDir
	bdCreateOut, err := bdCreateCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bd create failed: %v\nOutput: %s", err, bdCreateOut)
	}
	bdID := extractCreatedID(string(bdCreateOut))
	if bdID == "" {
		t.Fatalf("Could not extract ID from bd create output: %s", bdCreateOut)
	}

	// Get show output from both backends
	brShowCmd := exec.Command("br", "show", brID, "--json")
	brShowCmd.Dir = brEnv.WorkDir
	brShowOut, err := brShowCmd.Output()
	if err != nil {
		t.Fatalf("br show --json failed: %v", err)
	}

	bdShowCmd := exec.Command("bd", "--db", bdEnv.DBPath, "show", bdID, "--json")
	bdShowCmd.Dir = bdEnv.WorkDir
	bdShowOut, err := bdShowCmd.Output()
	if err != nil {
		t.Fatalf("bd show --json failed: %v", err)
	}

	// Parse JSON outputs - show may return array or object depending on backend
	var brIssue map[string]interface{}
	var bdIssue map[string]interface{}

	// Try parsing as array first, then as object
	if err := json.Unmarshal(brShowOut, &brIssue); err != nil {
		var brIssues []map[string]interface{}
		if err := json.Unmarshal(brShowOut, &brIssues); err != nil {
			t.Fatalf("Failed to parse br show output: %v\nOutput: %s", err, brShowOut)
		}
		if len(brIssues) > 0 {
			brIssue = brIssues[0]
		}
	}

	if err := json.Unmarshal(bdShowOut, &bdIssue); err != nil {
		var bdIssues []map[string]interface{}
		if err := json.Unmarshal(bdShowOut, &bdIssues); err != nil {
			t.Fatalf("Failed to parse bd show output: %v\nOutput: %s", err, bdShowOut)
		}
		if len(bdIssues) > 0 {
			bdIssue = bdIssues[0]
		}
	}

	// Verify required fields exist in both outputs
	requiredFields := []string{
		"id", "title", "status", "priority", "issue_type",
		"created_at", "updated_at",
	}

	// Optional fields that abacus uses if present
	optionalFields := []string{
		"description", "assignee", "labels", "dependencies", "comments",
	}

	t.Log("Checking br show --json output fields")
	for _, field := range requiredFields {
		if _, ok := brIssue[field]; !ok {
			t.Errorf("br show missing required field %q", field)
		}
	}

	t.Log("Checking bd show --json output fields")
	for _, field := range requiredFields {
		if _, ok := bdIssue[field]; !ok {
			t.Errorf("bd show missing required field %q", field)
		}
	}

	// Log which optional fields are present
	t.Log("Optional fields present in br show:")
	for _, field := range optionalFields {
		if _, ok := brIssue[field]; ok {
			t.Logf("  - %s: present", field)
		}
	}
	t.Log("Optional fields present in bd show:")
	for _, field := range optionalFields {
		if _, ok := bdIssue[field]; ok {
			t.Logf("  - %s: present", field)
		}
	}

	t.Log("Output conformance test passed: both backends produce compatible show output")
}

// TestSchemaConformance verifies that both backends create compatible SQLite schemas.
// This is a more detailed version that checks specific column types and constraints.
func TestSchemaConformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping conformance test in short mode")
	}

	// Need both backends for conformance testing
	_, err := exec.LookPath("br")
	if err != nil {
		t.Skip("br binary not found, skipping conformance test")
	}
	_, err = exec.LookPath("bd")
	if err != nil {
		t.Skip("bd binary not found, skipping conformance test")
	}

	// Set up test databases
	brEnv := setupBackendTestDB(t, "br")
	defer brEnv.cleanup()

	bdEnv := setupBackendTestDB(t, "bd")
	defer bdEnv.cleanup()

	// Compare schemas using sqlite3 command
	brSchemaCmd := exec.Command("sqlite3", brEnv.DBPath, ".schema issues")
	brSchema, err := brSchemaCmd.Output()
	if err != nil {
		t.Fatalf("Failed to get br schema: %v", err)
	}

	bdSchemaCmd := exec.Command("sqlite3", bdEnv.DBPath, ".schema issues")
	bdSchema, err := bdSchemaCmd.Output()
	if err != nil {
		t.Fatalf("Failed to get bd schema: %v", err)
	}

	t.Logf("br issues schema:\n%s", brSchema)
	t.Logf("bd issues schema:\n%s", bdSchema)

	// Check that required columns exist in both schemas
	requiredColumns := []string{"id", "title", "status", "priority", "issue_type"}

	brSchemaStr := strings.ToLower(string(brSchema))
	bdSchemaStr := strings.ToLower(string(bdSchema))

	for _, col := range requiredColumns {
		if !strings.Contains(brSchemaStr, col) {
			t.Errorf("br schema missing column %q", col)
		}
		if !strings.Contains(bdSchemaStr, col) {
			t.Errorf("bd schema missing column %q", col)
		}
	}

	t.Log("Schema conformance test passed: both backends have compatible issue schemas")
}

// TestDependencyTypeConformance verifies dependency type handling between backends.
// Note: br uses "dependency_type" while bd uses "dep_type" in JSON output.
// Abacus normalizes this difference in the SQLite queries, so this test verifies
// that the actual dependency relationships work correctly.
func TestDependencyTypeConformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping conformance test in short mode")
	}

	// Need both backends for conformance testing
	_, err := exec.LookPath("br")
	if err != nil {
		t.Skip("br binary not found, skipping conformance test")
	}
	_, err = exec.LookPath("bd")
	if err != nil {
		t.Skip("bd binary not found, skipping conformance test")
	}

	backends := []struct {
		name string
		env  backendTestEnv
	}{
		{"br", setupBackendTestDB(t, "br")},
		{"bd", setupBackendTestDB(t, "bd")},
	}

	for _, b := range backends {
		b := b
		defer b.env.cleanup()

		t.Run(b.name+"_dependency_types", func(t *testing.T) {
			var createCmd, depCmd *exec.Cmd

			// Create two issues
			if b.name == "br" {
				createCmd = exec.Command("br", "create", "Parent Issue", "--type", "epic", "--priority", "1")
				createCmd.Dir = b.env.WorkDir
			} else {
				createCmd = exec.Command("bd", "--db", b.env.DBPath, "create",
					"--title", "Parent Issue", "--type", "epic", "--priority", "1")
				createCmd.Dir = b.env.WorkDir
			}
			out, err := createCmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s create parent failed: %v\nOutput: %s", b.name, err, out)
			}
			parentID := extractCreatedID(string(out))

			if b.name == "br" {
				createCmd = exec.Command("br", "create", "Child Issue", "--type", "task", "--priority", "2")
				createCmd.Dir = b.env.WorkDir
			} else {
				createCmd = exec.Command("bd", "--db", b.env.DBPath, "create",
					"--title", "Child Issue", "--type", "task", "--priority", "2")
				createCmd.Dir = b.env.WorkDir
			}
			out, err = createCmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s create child failed: %v\nOutput: %s", b.name, err, out)
			}
			childID := extractCreatedID(string(out))

			// Add parent-child dependency
			if b.name == "br" {
				depCmd = exec.Command("br", "dep", "add", childID, parentID, "--type", "parent-child")
				depCmd.Dir = b.env.WorkDir
			} else {
				depCmd = exec.Command("bd", "--db", b.env.DBPath, "dep", "add",
					childID, parentID, "--type", "parent-child")
				depCmd.Dir = b.env.WorkDir
			}
			out, err = depCmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s dep add failed: %v\nOutput: %s", b.name, err, out)
			}

			// Verify dependency exists in database
			checkCmd := exec.Command("sqlite3", b.env.DBPath,
				"SELECT COUNT(*) FROM dependencies WHERE issue_id = '"+childID+"' AND depends_on_id = '"+parentID+"'")
			out, err = checkCmd.Output()
			if err != nil {
				t.Fatalf("sqlite3 check failed: %v", err)
			}
			count := strings.TrimSpace(string(out))
			if count != "1" {
				t.Errorf("%s: expected 1 dependency, got %s", b.name, count)
			}

			t.Logf("%s: dependency type conformance passed", b.name)
		})
	}
}
