# Close Reason in Bead Details Pane - Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Display the close reason in the bead details pane when present, positioned right before the Description section.

**Architecture:** Add `CloseReason` field to `FullIssue` struct to capture the `close_reason` JSON field from issues.jsonl. In the detail view, render a new "Close Reason:" section between the relationship sections and Description, but only when the field is non-empty.

**Tech Stack:** Go, Bubble Tea TUI framework, lipgloss styling

---

### Task 1: Add CloseReason field to FullIssue struct

**Files:**
- Modify: `internal/beads/types.go:30-49`

**Step 1: Add the CloseReason field**

Add the field after `ClosedAt` (line 42) to keep close-related fields together:

```go
ClosedAt           string       `json:"closed_at"`
CloseReason        string       `json:"close_reason"`
```

**Step 2: Verify the change compiles**

Run: `go build ./...`
Expected: Build succeeds with no errors

**Step 3: Commit**

```bash
git add internal/beads/types.go
git commit -m "$(cat <<'EOF'
feat: add CloseReason field to FullIssue struct

Support parsing close_reason from issues.jsonl for display in detail pane.
EOF
)"
```

---

### Task 2: Write failing test for CloseReason display

**Files:**
- Modify: `internal/ui/detail_test.go`

**Step 1: Write the failing test**

Add test at end of file, before final closing brace or after existing tests:

```go
func TestDetailViewShowsCloseReason(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:          "ab-closed",
			Title:       "Completed Feature",
			Status:      "closed",
			IssueType:   "task",
			Priority:    2,
			Description: "Some description here.",
			CloseReason: "Work completed in commit abc123. All tests passing.",
			CreatedAt:   now,
			UpdatedAt:   now,
			ClosedAt:    now,
		},
		CommentsLoaded: true,
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  []graph.TreeRow{{Node: node}},
		viewport:     viewport.New(90, 40),
		outputFormat: "plain",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())

	// Close Reason section should be present
	if !strings.Contains(content, "Close Reason:") {
		t.Fatalf("expected 'Close Reason' section for closed bead with reason:\n%s", content)
	}
	// The reason text should be visible
	if !strings.Contains(content, "Work completed in commit abc123") {
		t.Fatalf("expected close reason text in detail view:\n%s", content)
	}
}

func TestDetailViewHidesEmptyCloseReason(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:          "ab-closed2",
			Title:       "Closed Without Reason",
			Status:      "closed",
			IssueType:   "task",
			Priority:    2,
			Description: "Some description.",
			CloseReason: "", // Empty - should not show section
			CreatedAt:   now,
			UpdatedAt:   now,
			ClosedAt:    now,
		},
		CommentsLoaded: true,
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  []graph.TreeRow{{Node: node}},
		viewport:     viewport.New(90, 40),
		outputFormat: "plain",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())

	// Close Reason section should NOT be present when empty
	if strings.Contains(content, "Close Reason:") {
		t.Fatalf("did not expect 'Close Reason' section when CloseReason is empty:\n%s", content)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v ./internal/ui -run TestDetailViewShowsCloseReason`
Expected: FAIL - "expected 'Close Reason' section for closed bead with reason"

**Step 3: Commit failing test**

```bash
git add internal/ui/detail_test.go
git commit -m "$(cat <<'EOF'
test: add failing tests for close reason display in detail pane
EOF
)"
```

---

### Task 3: Implement CloseReason display in detail pane

**Files:**
- Modify: `internal/ui/detail.go:206-219`

**Step 1: Add CloseReason section before Description**

In `updateViewportContent()`, find the `descSections` initialization (around line 206-207):

```go
renderMarkdown := buildMarkdownRenderer(m.outputFormat, vpWidth-2)
descSections := []string{
	renderContentSection("Description:", renderMarkdown(iss.Description)),
}
```

Change it to:

```go
renderMarkdown := buildMarkdownRenderer(m.outputFormat, vpWidth-2)
descSections := make([]string, 0, 5)
if strings.TrimSpace(iss.CloseReason) != "" {
	descSections = append(descSections, renderContentSection("Close Reason:", renderMarkdown(iss.CloseReason)))
}
descSections = append(descSections, renderContentSection("Description:", renderMarkdown(iss.Description)))
```

**Step 2: Run tests to verify they pass**

Run: `go test -v ./internal/ui -run TestDetailViewShowsCloseReason`
Expected: PASS

Run: `go test -v ./internal/ui -run TestDetailViewHidesEmptyCloseReason`
Expected: PASS

**Step 3: Run all tests to ensure no regressions**

Run: `make check-test`
Expected: All checks and tests pass

**Step 4: Commit**

```bash
git add internal/ui/detail.go
git commit -m "$(cat <<'EOF'
feat: display close reason in bead details pane

Show Close Reason section right before Description when populated.
Hidden when empty to avoid clutter for beads closed without a reason.
EOF
)"
```

---

### Task 4: Visual verification with TUI

**Files:**
- None (manual testing)

**Step 1: Build the application**

Run: `make build`
Expected: Build succeeds

**Step 2: Find a bead with close_reason to test**

Run: `grep "close_reason" .beads/issues.jsonl | grep -v '"close_reason":""' | head -1 | jq -r '.id'`
Expected: Returns a bead ID like `ab-3msi`

**Step 3: Launch TUI and verify visually**

Run: `./scripts/tui-test.sh start`

Navigate to a closed bead with close reason and verify:
- "Close Reason:" section appears between relationship sections and Description
- The reason text is readable and properly wrapped
- Open beads do NOT show a Close Reason section

Run: `./scripts/tui-test.sh quit`

**Step 4: Update bead status**

```bash
bd comment add ab-9x1l "Implemented close reason display. Shows before Description, only when populated."
bd close ab-9x1l --reason="Implemented in this commit"
```

---

### Task 5: Final verification and push

**Step 1: Run full test suite**

Run: `make check-test`
Expected: All checks and tests pass

**Step 2: Push changes**

```bash
git pull --rebase
bd sync
git push
git status
```
Expected: "Your branch is up to date with 'origin/main'"
