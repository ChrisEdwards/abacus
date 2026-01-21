# Closed Items Reverse Chronological Order Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Display closed items in reverse chronological order (most recently closed first) instead of oldest first.

**Architecture:** Modify the sort comparator in `sortNodes()` to sort closed items by timestamp in descending order (After) instead of ascending (Before), while preserving ascending order for all other status categories.

**Tech Stack:** Go, standard library `sort` package

---

## Analysis

**Current behavior:** In `internal/graph/builder.go:307-308`, `sortNodes()` sorts all items by `SortTimestamp.Before()`, meaning earliest timestamp first. For closed items, this shows oldest closed first.

**Desired behavior:** Closed items should sort by `SortTimestamp.After()` (most recent first), while open/in_progress/blocked/deferred items retain `Before()` ordering.

**Key insight:** The sort priority already groups closed items together (priority 5). We only need to reverse the timestamp comparison when BOTH items being compared are closed.

---

## Task 1: Add test for closed items reverse chronological order

**Files:**
- Modify: `internal/graph/builder_test.go`

**Step 1: Write the failing test**

Add this test at the end of `builder_test.go`:

```go
func TestSortNodesClosedItemsReverseChronological(t *testing.T) {
	// Closed items should appear in reverse chronological order (most recently closed first)
	closedOld := &Node{
		Issue: beads.FullIssue{
			ID:       "ab-old",
			Status:   "closed",
			ClosedAt: "2024-01-01T00:00:00Z",
		},
	}
	closedMid := &Node{
		Issue: beads.FullIssue{
			ID:       "ab-mid",
			Status:   "closed",
			ClosedAt: "2024-02-01T00:00:00Z",
		},
	}
	closedNew := &Node{
		Issue: beads.FullIssue{
			ID:       "ab-new",
			Status:   "closed",
			ClosedAt: "2024-03-01T00:00:00Z",
		},
	}

	parent := &Node{
		Issue:    beads.FullIssue{ID: "ab-parent", Status: "open", CreatedAt: "2024-01-01T00:00:00Z"},
		Children: []*Node{closedOld, closedNew, closedMid},
	}

	computeSortMetrics(parent)

	// Expected: most recently closed first (reverse chronological)
	wantOrder := []string{"ab-new", "ab-mid", "ab-old"}
	for i, want := range wantOrder {
		if parent.Children[i].Issue.ID != want {
			got := make([]string, len(parent.Children))
			for j, c := range parent.Children {
				got[j] = c.Issue.ID
			}
			t.Fatalf("closed items order mismatch: got %v, want %v", got, wantOrder)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v ./internal/graph -run TestSortNodesClosedItemsReverseChronological`
Expected: FAIL - order will be `[ab-old, ab-mid, ab-new]` (oldest first)

**Step 3: Commit the failing test**

```bash
git add internal/graph/builder_test.go
git commit -m "test: add failing test for closed items reverse chronological order"
```

---

## Task 2: Add test for mixed status items with closed

**Files:**
- Modify: `internal/graph/builder_test.go`

**Step 1: Write the test**

Add this test after the previous one:

```go
func TestSortNodesMixedStatusWithClosedReverseChronological(t *testing.T) {
	// Verify that:
	// 1. Open items still come before closed items
	// 2. Open items are sorted oldest-first (ascending)
	// 3. Closed items are sorted newest-first (descending)
	openOld := &Node{
		Issue: beads.FullIssue{
			ID:        "ab-open-old",
			Status:    "open",
			CreatedAt: "2024-01-01T00:00:00Z",
		},
	}
	openNew := &Node{
		Issue: beads.FullIssue{
			ID:        "ab-open-new",
			Status:    "open",
			CreatedAt: "2024-02-01T00:00:00Z",
		},
	}
	closedOld := &Node{
		Issue: beads.FullIssue{
			ID:       "ab-closed-old",
			Status:   "closed",
			ClosedAt: "2024-01-15T00:00:00Z",
		},
	}
	closedNew := &Node{
		Issue: beads.FullIssue{
			ID:       "ab-closed-new",
			Status:   "closed",
			ClosedAt: "2024-02-15T00:00:00Z",
		},
	}

	parent := &Node{
		Issue:    beads.FullIssue{ID: "ab-parent", Status: "open", CreatedAt: "2024-01-01T00:00:00Z"},
		Children: []*Node{closedOld, openNew, closedNew, openOld},
	}

	computeSortMetrics(parent)

	// Expected: open items first (oldest first), then closed items (newest first)
	wantOrder := []string{"ab-open-old", "ab-open-new", "ab-closed-new", "ab-closed-old"}
	for i, want := range wantOrder {
		if parent.Children[i].Issue.ID != want {
			got := make([]string, len(parent.Children))
			for j, c := range parent.Children {
				got[j] = c.Issue.ID
			}
			t.Fatalf("mixed status order mismatch: got %v, want %v", got, wantOrder)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v ./internal/graph -run TestSortNodesMixedStatusWithClosedReverseChronological`
Expected: FAIL - closed items will be in wrong order

**Step 3: Commit the failing test**

```bash
git add internal/graph/builder_test.go
git commit -m "test: add failing test for mixed status items with closed reverse chronological"
```

---

## Task 3: Implement reverse chronological sorting for closed items

**Files:**
- Modify: `internal/graph/builder.go:301-312`

**Step 1: Modify sortNodes function**

Replace the `sortNodes` function at lines 301-312:

```go
func sortNodes(nodes []*Node) {
	sort.SliceStable(nodes, func(i, j int) bool {
		a, b := nodes[i], nodes[j]
		if a.SortPriority != b.SortPriority {
			return a.SortPriority < b.SortPriority
		}
		if !a.SortTimestamp.Equal(b.SortTimestamp) {
			// Closed items: reverse chronological (most recent first)
			// All other statuses: chronological (oldest first)
			if a.SortPriority == sortPriorityClosed {
				return a.SortTimestamp.After(b.SortTimestamp)
			}
			return a.SortTimestamp.Before(b.SortTimestamp)
		}
		return a.Issue.ID < b.Issue.ID
	})
}
```

**Step 2: Run all sorting tests to verify they pass**

Run: `go test -v ./internal/graph -run 'TestSort|TestClosed'`
Expected: All tests PASS including the new ones

**Step 3: Run full test suite**

Run: `make test`
Expected: All tests pass

**Step 4: Commit the implementation**

```bash
git add internal/graph/builder.go
git commit -m "feat: display closed items in reverse chronological order

Closed items now appear with most recently closed first, making it
easier to see what was completed recently. Open/in_progress/blocked/
deferred items retain their existing oldest-first ordering.

Closes: ab-s4bq"
```

---

## Task 4: Add test for detail pane subtasks closed ordering

**Files:**
- Modify: `internal/ui/sort_test.go`

**Step 1: Write the failing test**

Add this test to `sort_test.go`:

```go
func TestSortSubtasksClosedReverseChronological(t *testing.T) {
	t.Run("closedItemsNewestFirst", func(t *testing.T) {
		closedOld := &graph.Node{
			Issue: beads.FullIssue{
				ID:       "ab-old",
				Status:   "closed",
				ClosedAt: "2024-01-01T00:00:00Z",
			},
		}
		closedMid := &graph.Node{
			Issue: beads.FullIssue{
				ID:       "ab-mid",
				Status:   "closed",
				ClosedAt: "2024-02-01T00:00:00Z",
			},
		}
		closedNew := &graph.Node{
			Issue: beads.FullIssue{
				ID:       "ab-new",
				Status:   "closed",
				ClosedAt: "2024-03-01T00:00:00Z",
			},
		}

		input := []*graph.Node{closedOld, closedNew, closedMid}
		result := sortSubtasks(input)

		// Expected: most recently closed first
		expected := []string{"ab-new", "ab-mid", "ab-old"}
		for i, id := range expected {
			if result[i].Issue.ID != id {
				t.Fatalf("position %d: expected %s, got %s", i, id, result[i].Issue.ID)
			}
		}
	})
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v ./internal/ui -run TestSortSubtasksClosedReverseChronological`
Expected: FAIL - order will be alphabetical by ID since no timestamp comparison exists

**Step 3: Commit the failing test**

```bash
git add internal/ui/sort_test.go
git commit -m "test: add failing test for detail pane closed items reverse chronological"
```

---

## Task 5: Implement reverse chronological sorting in sortSubtasks

**Files:**
- Modify: `internal/ui/sort.go:65-110`

**Step 1: Add time import and helper function**

Add to imports at top of file:

```go
import (
	"sort"
	"time"

	"abacus/internal/graph"
)
```

Add helper function after `countOpenBlockers`:

```go
// parseClosedAt parses the ClosedAt timestamp, returning zero time if unparseable
func parseClosedAt(n *graph.Node) time.Time {
	if n.Issue.ClosedAt == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, n.Issue.ClosedAt)
	if err != nil {
		return time.Time{}
	}
	return t
}
```

**Step 2: Modify sortSubtasks to handle closed items**

Update the sort function to add closed item handling after the blocked section check (around line 98) and before the priority check:

```go
// sortSubtasks orders children: in_progress → ready (unblocking others first) → blocked → closed
func sortSubtasks(nodes []*graph.Node) []*graph.Node {
	if len(nodes) <= 1 {
		return nodes
	}

	result := make([]*graph.Node, len(nodes))
	copy(result, nodes)

	sort.SliceStable(result, func(i, j int) bool {
		catI := nodeStatusCategory(result[i])
		catJ := nodeStatusCategory(result[j])

		// Primary: status category
		if catI != catJ {
			return catI < catJ
		}

		// Secondary within ready: items that unblock others come first
		if catI == statusReady {
			blocksI := len(result[i].Blocks)
			blocksJ := len(result[j].Blocks)
			if blocksI != blocksJ {
				return blocksI > blocksJ // More blocks = higher priority
			}
		}

		// Secondary within blocked: closest to ready (fewest blockers) first
		if catI == statusBlocked {
			openBlockersI := countOpenBlockers(result[i])
			openBlockersJ := countOpenBlockers(result[j])
			if openBlockersI != openBlockersJ {
				return openBlockersI < openBlockersJ // Fewer blockers = higher priority
			}
		}

		// Secondary within closed: reverse chronological (most recent first)
		if catI == statusClosed {
			closedI := parseClosedAt(result[i])
			closedJ := parseClosedAt(result[j])
			if !closedI.IsZero() && !closedJ.IsZero() && !closedI.Equal(closedJ) {
				return closedI.After(closedJ) // More recent = higher priority
			}
		}

		// Tertiary: priority (lower number = higher priority)
		if result[i].Issue.Priority != result[j].Issue.Priority {
			return result[i].Issue.Priority < result[j].Issue.Priority
		}

		// Quaternary: by ID for stability
		return result[i].Issue.ID < result[j].Issue.ID
	})

	return result
}
```

**Step 3: Run tests to verify they pass**

Run: `go test -v ./internal/ui -run TestSortSubtasks`
Expected: All tests PASS

**Step 4: Run full test suite**

Run: `make test`
Expected: All tests pass

**Step 5: Commit the implementation**

```bash
git add internal/ui/sort.go
git commit -m "feat: display closed subtasks in reverse chronological order in detail pane"
```

---

## Task 6: Add test for sortBlockers closed items ordering

**Files:**
- Modify: `internal/ui/sort_test.go`

**Step 1: Write the failing test**

Add to `sort_test.go`:

```go
func TestSortBlockersClosedReverseChronological(t *testing.T) {
	t.Run("closedBlockersNewestFirst", func(t *testing.T) {
		closedOld := &graph.Node{
			Issue: beads.FullIssue{
				ID:       "ab-old",
				Status:   "closed",
				ClosedAt: "2024-01-01T00:00:00Z",
			},
		}
		closedNew := &graph.Node{
			Issue: beads.FullIssue{
				ID:       "ab-new",
				Status:   "closed",
				ClosedAt: "2024-02-01T00:00:00Z",
			},
		}

		input := []*graph.Node{closedOld, closedNew}
		result := sortBlockers(input)

		// Expected: most recently closed first (among closed items)
		expected := []string{"ab-new", "ab-old"}
		for i, id := range expected {
			if result[i].Issue.ID != id {
				t.Fatalf("position %d: expected %s, got %s", i, id, result[i].Issue.ID)
			}
		}
	})
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v ./internal/ui -run TestSortBlockersClosedReverseChronological`
Expected: FAIL

**Step 3: Commit the failing test**

```bash
git add internal/ui/sort_test.go
git commit -m "test: add failing test for blockers closed items reverse chronological"
```

---

## Task 7: Implement reverse chronological in sortBlockers

**Files:**
- Modify: `internal/ui/sort.go:114-153`

**Step 1: Update sortBlockers to handle closed items**

Replace `sortBlockers` function:

```go
// sortBlockers orders blockers topologically - things to do first appear first
// Items with no blockers of their own come first, then by priority
func sortBlockers(nodes []*graph.Node) []*graph.Node {
	if len(nodes) <= 1 {
		return nodes
	}

	result := make([]*graph.Node, len(nodes))
	copy(result, nodes)

	sort.SliceStable(result, func(i, j int) bool {
		// Closed items last
		iClosed := result[i].Issue.Status == "closed"
		jClosed := result[j].Issue.Status == "closed"
		if iClosed != jClosed {
			return jClosed // closed items come last
		}

		// Within closed items: reverse chronological (most recent first)
		if iClosed && jClosed {
			closedI := parseClosedAt(result[i])
			closedJ := parseClosedAt(result[j])
			if !closedI.IsZero() && !closedJ.IsZero() && !closedI.Equal(closedJ) {
				return closedI.After(closedJ)
			}
		}

		// Items with fewer open blockers come first (can be worked on sooner)
		openBlockersI := countOpenBlockers(result[i])
		openBlockersJ := countOpenBlockers(result[j])
		if openBlockersI != openBlockersJ {
			return openBlockersI < openBlockersJ
		}

		// In-progress items come first among same-blocker-count items
		iInProgress := result[i].Issue.Status == "in_progress"
		jInProgress := result[j].Issue.Status == "in_progress"
		if iInProgress != jInProgress {
			return iInProgress
		}

		// By priority
		if result[i].Issue.Priority != result[j].Issue.Priority {
			return result[i].Issue.Priority < result[j].Issue.Priority
		}

		return result[i].Issue.ID < result[j].Issue.ID
	})

	return result
}
```

**Step 2: Run tests to verify they pass**

Run: `go test -v ./internal/ui -run TestSortBlockers`
Expected: All tests PASS

**Step 3: Commit the implementation**

```bash
git add internal/ui/sort.go
git commit -m "feat: display closed blockers in reverse chronological order"
```

---

## Task 8: Run full verification

**Step 1: Run linting**

Run: `make check`
Expected: No errors

**Step 2: Run all tests**

Run: `make test`
Expected: All tests pass

**Step 3: Visual verification**

Run: `make build && ./scripts/tui-test.sh start`
Navigate to items with closed children to verify order.
Run: `./scripts/tui-test.sh quit`

**Step 4: Update bead and commit**

```bash
bd update ab-s4bq --status closed
bd comments add ab-s4bq "Implemented reverse chronological ordering for closed items in both main tree and detail pane. Commit: $(git rev-parse --short HEAD)"
bd sync
git add .beads/
git commit -m "chore: close bead ab-s4bq"
git push
```

---

## Summary

| Task | Description | Files |
|------|-------------|-------|
| 1 | Test: closed items reverse chronological | `builder_test.go` |
| 2 | Test: mixed status with closed | `builder_test.go` |
| 3 | Implement: main tree sorting | `builder.go` |
| 4 | Test: detail pane subtasks | `sort_test.go` |
| 5 | Implement: sortSubtasks | `sort.go` |
| 6 | Test: detail pane blockers | `sort_test.go` |
| 7 | Implement: sortBlockers | `sort.go` |
| 8 | Final verification | - |

Total: 4 test additions, 3 implementation changes, ~50 lines of code.
