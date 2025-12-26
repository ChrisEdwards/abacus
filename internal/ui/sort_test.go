package ui

import (
	"testing"

	"abacus/internal/beads"
	"abacus/internal/graph"
)

func TestSortSubtasks(t *testing.T) {
	t.Run("ordersStatusCategories", func(t *testing.T) {
		inProgress := &graph.Node{Issue: beads.FullIssue{ID: "ab-ip", Status: "in_progress"}}
		ready := &graph.Node{Issue: beads.FullIssue{ID: "ab-ready", Status: "open"}}
		blocked := &graph.Node{Issue: beads.FullIssue{ID: "ab-blocked", Status: "open"}, IsBlocked: true}
		closed := &graph.Node{Issue: beads.FullIssue{ID: "ab-closed", Status: "closed"}}

		// Provide in reverse order
		input := []*graph.Node{closed, blocked, ready, inProgress}
		result := sortSubtasks(input)

		expected := []string{"ab-ip", "ab-ready", "ab-blocked", "ab-closed"}
		for i, id := range expected {
			if result[i].Issue.ID != id {
				t.Fatalf("position %d: expected %s, got %s", i, id, result[i].Issue.ID)
			}
		}
	})

	t.Run("ordersExplicitBlockedAndDeferredStatuses", func(t *testing.T) {
		inProgress := &graph.Node{Issue: beads.FullIssue{ID: "ab-ip", Status: "in_progress"}}
		ready := &graph.Node{Issue: beads.FullIssue{ID: "ab-ready", Status: "open"}}
		explicitBlocked := &graph.Node{Issue: beads.FullIssue{ID: "ab-explicit-blocked", Status: "blocked"}}
		deferred := &graph.Node{Issue: beads.FullIssue{ID: "ab-deferred", Status: "deferred"}}
		closed := &graph.Node{Issue: beads.FullIssue{ID: "ab-closed", Status: "closed"}}

		// Provide in reverse order
		input := []*graph.Node{closed, deferred, explicitBlocked, ready, inProgress}
		result := sortSubtasks(input)

		// Expected: in_progress → ready → blocked → deferred → closed
		expected := []string{"ab-ip", "ab-ready", "ab-explicit-blocked", "ab-deferred", "ab-closed"}
		for i, id := range expected {
			if result[i].Issue.ID != id {
				t.Fatalf("position %d: expected %s, got %s", i, id, result[i].Issue.ID)
			}
		}
	})

	t.Run("explicitBlockedBeforeDependencyBlocked", func(t *testing.T) {
		// Both are in "blocked" category but explicit status first (same priority)
		explicitBlocked := &graph.Node{Issue: beads.FullIssue{ID: "ab-explicit", Status: "blocked", Priority: 2}}
		depBlocked := &graph.Node{Issue: beads.FullIssue{ID: "ab-dep", Status: "open", Priority: 2}, IsBlocked: true}

		input := []*graph.Node{depBlocked, explicitBlocked}
		result := sortSubtasks(input)

		// With same priority and category, sorted by ID
		if result[0].Issue.ID != "ab-dep" || result[1].Issue.ID != "ab-explicit" {
			t.Fatalf("expected alphabetical order within same category, got %s, %s",
				result[0].Issue.ID, result[1].Issue.ID)
		}
	})

	t.Run("readyItemsThatUnblockOthersFirst", func(t *testing.T) {
		// Ready item that blocks nothing
		readyNoImpact := &graph.Node{Issue: beads.FullIssue{ID: "ab-no-impact", Status: "open", Priority: 2}}
		// Ready item that blocks 2 things
		readyHighImpact := &graph.Node{
			Issue:  beads.FullIssue{ID: "ab-high-impact", Status: "open", Priority: 2},
			Blocks: []*graph.Node{{}, {}}, // 2 blocked items
		}
		// Ready item that blocks 1 thing
		readyMedImpact := &graph.Node{
			Issue:  beads.FullIssue{ID: "ab-med-impact", Status: "open", Priority: 2},
			Blocks: []*graph.Node{{}}, // 1 blocked item
		}

		input := []*graph.Node{readyNoImpact, readyMedImpact, readyHighImpact}
		result := sortSubtasks(input)

		// High impact (2 blocks) → Med impact (1 block) → No impact (0 blocks)
		expected := []string{"ab-high-impact", "ab-med-impact", "ab-no-impact"}
		for i, id := range expected {
			if result[i].Issue.ID != id {
				t.Fatalf("position %d: expected %s, got %s", i, id, result[i].Issue.ID)
			}
		}
	})

	t.Run("blockedItemsClosestToReadyFirst", func(t *testing.T) {
		blocker1 := &graph.Node{Issue: beads.FullIssue{ID: "blocker1", Status: "open"}}
		blocker2 := &graph.Node{Issue: beads.FullIssue{ID: "blocker2", Status: "open"}}
		blocker3 := &graph.Node{Issue: beads.FullIssue{ID: "blocker3", Status: "open"}}

		// Blocked by 1 open issue
		blocked1 := &graph.Node{
			Issue:     beads.FullIssue{ID: "ab-b1", Status: "open", Priority: 2},
			IsBlocked: true,
			BlockedBy: []*graph.Node{blocker1},
		}
		// Blocked by 3 open issues
		blocked3 := &graph.Node{
			Issue:     beads.FullIssue{ID: "ab-b3", Status: "open", Priority: 2},
			IsBlocked: true,
			BlockedBy: []*graph.Node{blocker1, blocker2, blocker3},
		}
		// Blocked by 2 open issues
		blocked2 := &graph.Node{
			Issue:     beads.FullIssue{ID: "ab-b2", Status: "open", Priority: 2},
			IsBlocked: true,
			BlockedBy: []*graph.Node{blocker1, blocker2},
		}

		input := []*graph.Node{blocked3, blocked1, blocked2}
		result := sortSubtasks(input)

		// Fewer blockers = closer to ready = comes first
		expected := []string{"ab-b1", "ab-b2", "ab-b3"}
		for i, id := range expected {
			if result[i].Issue.ID != id {
				t.Fatalf("position %d: expected %s, got %s", i, id, result[i].Issue.ID)
			}
		}
	})

	t.Run("priorityTiebreaker", func(t *testing.T) {
		p1 := &graph.Node{Issue: beads.FullIssue{ID: "ab-p1", Status: "open", Priority: 1}}
		p2 := &graph.Node{Issue: beads.FullIssue{ID: "ab-p2", Status: "open", Priority: 2}}
		p3 := &graph.Node{Issue: beads.FullIssue{ID: "ab-p3", Status: "open", Priority: 3}}

		input := []*graph.Node{p3, p1, p2}
		result := sortSubtasks(input)

		expected := []string{"ab-p1", "ab-p2", "ab-p3"}
		for i, id := range expected {
			if result[i].Issue.ID != id {
				t.Fatalf("position %d: expected %s, got %s", i, id, result[i].Issue.ID)
			}
		}
	})
}

func TestSortBlockers(t *testing.T) {
	t.Run("closedBlockersLast", func(t *testing.T) {
		open := &graph.Node{Issue: beads.FullIssue{ID: "ab-open", Status: "open"}}
		closed := &graph.Node{Issue: beads.FullIssue{ID: "ab-closed", Status: "closed"}}

		input := []*graph.Node{closed, open}
		result := sortBlockers(input)

		if result[0].Issue.ID != "ab-open" {
			t.Fatalf("expected open blocker first, got %s", result[0].Issue.ID)
		}
		if result[1].Issue.ID != "ab-closed" {
			t.Fatalf("expected closed blocker last, got %s", result[1].Issue.ID)
		}
	})

	t.Run("fewerBlockersFirst", func(t *testing.T) {
		blocker := &graph.Node{Issue: beads.FullIssue{ID: "blocker", Status: "open"}}

		// Has no blockers - can be worked on immediately
		noBlockers := &graph.Node{Issue: beads.FullIssue{ID: "ab-nb", Status: "open", Priority: 2}}
		// Has 1 blocker
		oneBlocker := &graph.Node{
			Issue:     beads.FullIssue{ID: "ab-1b", Status: "open", Priority: 2},
			BlockedBy: []*graph.Node{blocker},
		}

		input := []*graph.Node{oneBlocker, noBlockers}
		result := sortBlockers(input)

		expected := []string{"ab-nb", "ab-1b"}
		for i, id := range expected {
			if result[i].Issue.ID != id {
				t.Fatalf("position %d: expected %s, got %s", i, id, result[i].Issue.ID)
			}
		}
	})

	t.Run("inProgressFirst", func(t *testing.T) {
		inProgress := &graph.Node{Issue: beads.FullIssue{ID: "ab-ip", Status: "in_progress", Priority: 2}}
		open := &graph.Node{Issue: beads.FullIssue{ID: "ab-open", Status: "open", Priority: 2}}

		input := []*graph.Node{open, inProgress}
		result := sortBlockers(input)

		if result[0].Issue.ID != "ab-ip" {
			t.Fatalf("expected in_progress blocker first, got %s", result[0].Issue.ID)
		}
	})
}

func TestSortBlocked(t *testing.T) {
	t.Run("fewerBlockersFirst", func(t *testing.T) {
		blocker1 := &graph.Node{Issue: beads.FullIssue{ID: "blocker1", Status: "open"}}
		blocker2 := &graph.Node{Issue: beads.FullIssue{ID: "blocker2", Status: "open"}}

		// Will become ready (only 1 blocker - us)
		oneBlocker := &graph.Node{
			Issue:     beads.FullIssue{ID: "ab-1b", Status: "open", Priority: 2},
			BlockedBy: []*graph.Node{blocker1},
		}
		// Still blocked after us (2 blockers)
		twoBlockers := &graph.Node{
			Issue:     beads.FullIssue{ID: "ab-2b", Status: "open", Priority: 2},
			BlockedBy: []*graph.Node{blocker1, blocker2},
		}

		input := []*graph.Node{twoBlockers, oneBlocker}
		result := sortBlocked(input)

		expected := []string{"ab-1b", "ab-2b"}
		for i, id := range expected {
			if result[i].Issue.ID != id {
				t.Fatalf("position %d: expected %s, got %s", i, id, result[i].Issue.ID)
			}
		}
	})

	t.Run("priorityTiebreaker", func(t *testing.T) {
		p1 := &graph.Node{Issue: beads.FullIssue{ID: "ab-p1", Status: "open", Priority: 1}}
		p3 := &graph.Node{Issue: beads.FullIssue{ID: "ab-p3", Status: "open", Priority: 3}}

		input := []*graph.Node{p3, p1}
		result := sortBlocked(input)

		expected := []string{"ab-p1", "ab-p3"}
		for i, id := range expected {
			if result[i].Issue.ID != id {
				t.Fatalf("position %d: expected %s, got %s", i, id, result[i].Issue.ID)
			}
		}
	})
}
