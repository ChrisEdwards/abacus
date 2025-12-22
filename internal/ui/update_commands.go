package ui

import (
	"context"
	"sort"
	"time"

	"abacus/internal/graph"

	tea "github.com/charmbracelet/bubbletea"
)

func scheduleStatusToastTick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return statusToastTickMsg{}
	})
}

// executeStatusChangeCmd runs the bd update command asynchronously without toast.
func (m *App) executeStatusChangeCmd(issueID, newStatus string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.UpdateStatus(context.Background(), issueID, newStatus)
		return statusUpdateCompleteMsg{err: err}
	}
}

// executeReopenCmd runs the bd reopen command asynchronously.
func (m *App) executeReopenCmd(issueID string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.Reopen(context.Background(), issueID)
		return statusUpdateCompleteMsg{err: err}
	}
}

// displayStatusToast displays a success toast for status changes.
func (m *App) displayStatusToast(issueID, newStatus string) {
	m.statusToastNewStatus = newStatus
	m.statusToastBeadID = issueID
	m.statusToastVisible = true
	m.statusToastStart = time.Now()
}

// formatStatusLabel converts a status value to a display label.
func formatStatusLabel(status string) string {
	switch status {
	case "open":
		return "Open"
	case "in_progress":
		return "In Progress"
	case "closed":
		return "Closed"
	default:
		return status
	}
}

func scheduleLabelsToastTick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return labelsToastTickMsg{}
	})
}

// getAllLabels collects all unique labels from all issues in the tree.
func (m *App) getAllLabels() []string {
	labelSet := make(map[string]bool)
	var collectLabels func([]*graph.Node)
	collectLabels = func(nodes []*graph.Node) {
		for _, n := range nodes {
			for _, l := range n.Issue.Labels {
				labelSet[l] = true
			}
			collectLabels(n.Children)
		}
	}
	collectLabels(m.roots)

	labels := make([]string, 0, len(labelSet))
	for l := range labelSet {
		labels = append(labels, l)
	}
	sort.Strings(labels)
	return labels
}

// getAllAssignees collects all unique assignees from all issues in the tree.
func (m *App) getAllAssignees() []string {
	assigneeSet := make(map[string]bool)
	var collectAssignees func([]*graph.Node)
	collectAssignees = func(nodes []*graph.Node) {
		for _, n := range nodes {
			if a := n.Issue.Assignee; a != "" {
				assigneeSet[a] = true
			}
			collectAssignees(n.Children)
		}
	}
	collectAssignees(m.roots)

	assignees := make([]string, 0, len(assigneeSet))
	for a := range assigneeSet {
		assignees = append(assignees, a)
	}
	sort.Strings(assignees)
	return assignees
}

// executeLabelsUpdate runs the bd label add/remove commands asynchronously.
func (m *App) executeLabelsUpdate(msg LabelsUpdatedMsg) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		for _, label := range msg.Added {
			if err := m.client.AddLabel(ctx, msg.IssueID, label); err != nil {
				return labelUpdateCompleteMsg{err: err}
			}
		}
		for _, label := range msg.Removed {
			if err := m.client.RemoveLabel(ctx, msg.IssueID, label); err != nil {
				return labelUpdateCompleteMsg{err: err}
			}
		}
		return labelUpdateCompleteMsg{err: nil}
	}
}

// displayLabelsToast displays a success toast for label changes.
func (m *App) displayLabelsToast(issueID string, added, removed []string) {
	m.labelsToastBeadID = issueID
	m.labelsToastAdded = added
	m.labelsToastRemoved = removed
	m.labelsToastVisible = true
	m.labelsToastStart = time.Now()
}

func scheduleCreateToastTick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return createToastTickMsg{}
	})
}

// getAvailableParents collects all beads that can be used as parents.
func (m *App) getAvailableParents() []ParentOption {
	var parents []ParentOption
	var collectParents func([]*graph.Node)
	collectParents = func(nodes []*graph.Node) {
		for _, n := range nodes {
			display := n.Issue.ID + " " + truncateTitle(n.Issue.Title, 30)
			parents = append(parents, ParentOption{
				ID:      n.Issue.ID,
				Display: display,
			})
			collectParents(n.Children)
		}
	}
	collectParents(m.roots)
	return parents
}

// executeCreateBead runs the bd create command asynchronously.
func (m *App) executeCreateBead(msg BeadCreatedMsg) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		issue, err := m.client.CreateFull(ctx, msg.Title, msg.IssueType, msg.Priority, msg.Labels, msg.Assignee, msg.Description, msg.ParentID)
		if err != nil {
			return createCompleteMsg{err: err, stayOpen: msg.StayOpen}
		}
		return createCompleteMsg{
			id:        issue.ID,
			stayOpen:  msg.StayOpen,
			fullIssue: &issue,
			parentID:  msg.ParentID,
		}
	}
}

func (m *App) validateUpdate(msg BeadUpdatedMsg) error {
	if msg.IssueType != "epic" {
		return nil
	}
	if msg.ParentID == "" || msg.ParentID == msg.OriginalParentID {
		return nil
	}
	parent := m.findNodeByID(msg.ParentID)
	if parent == nil {
		return nil
	}
	if parent.Issue.IssueType != "epic" {
		return errInvalidEpicParent
	}
	return nil
}

// executeUpdateCmd runs the bd update command asynchronously.
func (m *App) executeUpdateCmd(msg BeadUpdatedMsg) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := m.client.UpdateFull(ctx, msg.ID, msg.Title, msg.IssueType, msg.Priority, msg.Labels, msg.Assignee, msg.Description); err != nil {
			return updateCompleteMsg{ID: msg.ID, Title: msg.Title, Err: err}
		}

		if msg.ParentID != msg.OriginalParentID {
			if msg.OriginalParentID != "" {
				if err := m.client.RemoveDependency(ctx, msg.ID, msg.OriginalParentID, "parent-child"); err != nil {
					return updateCompleteMsg{ID: msg.ID, Title: msg.Title, Err: err}
				}
			}
			if msg.ParentID != "" {
				if err := m.client.AddDependency(ctx, msg.ID, msg.ParentID, "parent-child"); err != nil {
					return updateCompleteMsg{ID: msg.ID, Title: msg.Title, Err: err}
				}
			}
		}

		return updateCompleteMsg{ID: msg.ID, Title: msg.Title, Err: nil}
	}
}

// displayCreateToast displays a success toast for bead creation.
func (m *App) displayCreateToast(title string, isUpdate bool) {
	m.createToastTitle = title
	m.createToastVisible = true
	m.createToastStart = time.Now()
	m.createToastIsUpdate = isUpdate
}

// displayNewLabelToast displays a toast for a newly created label (not in existing options).
func (m *App) displayNewLabelToast(label string) {
	m.newLabelToastLabel = label
	m.newLabelToastVisible = true
	m.newLabelToastStart = time.Now()
}

// displayNewAssigneeToast displays a toast for a newly created assignee (not in existing options).
func (m *App) displayNewAssigneeToast(assignee string) {
	m.newAssigneeToastAssignee = assignee
	m.newAssigneeToastVisible = true
	m.newAssigneeToastStart = time.Now()
}

func scheduleDeleteToastTick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return deleteToastTickMsg{}
	})
}

// executeDelete runs the bd delete command asynchronously and shows toast.
func (m *App) executeDelete(issueID string, cascade bool, childIDs []string) tea.Cmd {
	m.displayDeleteToast(issueID, cascade, len(childIDs))
	return func() tea.Msg {
		err := m.client.Delete(context.Background(), issueID, cascade)
		return deleteCompleteMsg{issueID: issueID, children: childIDs, cascade: cascade, err: err}
	}
}

// displayDeleteToast displays a success toast for deletion.
func (m *App) displayDeleteToast(issueID string, cascade bool, childCount int) {
	m.deleteToastBeadID = issueID
	m.deleteToastCascade = cascade
	m.deleteToastChildCount = childCount
	m.deleteToastVisible = true
	m.deleteToastStart = time.Now()
}

func scheduleCommentToastTick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(_ time.Time) tea.Msg {
		return commentToastTickMsg{}
	})
}

// executeAddComment runs the bd comments add command asynchronously.
func (m *App) executeAddComment(msg CommentAddedMsg) tea.Cmd {
	return func() tea.Msg {
		err := m.client.AddComment(context.Background(), msg.IssueID, msg.Comment)
		return commentCompleteMsg{issueID: msg.IssueID, err: err}
	}
}

// displayCommentToast displays a success toast for comment addition.
func (m *App) displayCommentToast(issueID string) {
	m.commentToastBeadID = issueID
	m.commentToastVisible = true
	m.commentToastStart = time.Now()
}
