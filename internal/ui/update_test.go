package ui

import (
	"context"
	"testing"

	"abacus/internal/beads"
)

func TestExecuteUpdateCmdUpdatesAndChangesParent(t *testing.T) {
	mockClient := beads.NewMockClient()

	updateCalled := false
	removed := false
	added := false

	mockClient.UpdateFullFn = func(_ context.Context, id, title, issueType string, priority int, labels []string, assignee, description string) error {
		updateCalled = id == "ab-1" && title == "New Title" && issueType == "task" && priority == 2 && description == "desc"
		return nil
	}
	mockClient.RemoveDependencyFn = func(_ context.Context, fromID, toID, depType string) error {
		removed = fromID == "ab-1" && toID == "ab-old" && depType == "parent-child"
		return nil
	}
	mockClient.AddDependencyFn = func(_ context.Context, fromID, toID, depType string) error {
		added = fromID == "ab-1" && toID == "ab-new" && depType == "parent-child"
		return nil
	}

	app := &App{client: mockClient}

	msg := BeadUpdatedMsg{
		ID:               "ab-1",
		Title:            "New Title",
		IssueType:        "task",
		Priority:         2,
		Description:      "desc",
		Labels:           []string{"ui"},
		ParentID:         "ab-new",
		OriginalParentID: "ab-old",
	}

	res := app.executeUpdateCmd(msg)()
	updateMsg, ok := res.(updateCompleteMsg)
	if !ok {
		t.Fatalf("expected updateCompleteMsg, got %T", res)
	}
	if updateMsg.Err != nil {
		t.Fatalf("expected nil error, got %v", updateMsg.Err)
	}
	if !updateCalled {
		t.Error("expected UpdateFull to be called with correct arguments")
	}
	if !removed {
		t.Error("expected RemoveDependency to be called for old parent")
	}
	if !added {
		t.Error("expected AddDependency to be called for new parent")
	}
}

func TestExecuteUpdateCmdNoParentChange(t *testing.T) {
	mockClient := beads.NewMockClient()
	mockClient.UpdateFullFn = func(_ context.Context, _id, _title, _issueType string, _priority int, _labels []string, _assignee, _description string) error {
		return nil
	}

	app := &App{client: mockClient}

	msg := BeadUpdatedMsg{
		ID:               "ab-1",
		Title:            "Same",
		IssueType:        "task",
		Priority:         1,
		Labels:           nil,
		ParentID:         "ab-parent",
		OriginalParentID: "ab-parent",
	}

	res := app.executeUpdateCmd(msg)()
	updateMsg := res.(updateCompleteMsg)
	if updateMsg.Err != nil {
		t.Fatalf("expected nil error, got %v", updateMsg.Err)
	}
	if mockClient.RemoveDependencyCallCount != 0 {
		t.Errorf("expected no RemoveDependency calls, got %d", mockClient.RemoveDependencyCallCount)
	}
	if mockClient.AddDependencyCallCount != 0 {
		t.Errorf("expected no AddDependency calls, got %d", mockClient.AddDependencyCallCount)
	}
}
