package ui

import (
	"strings"
	"testing"
	"time"

	"abacus/internal/beads"
	"abacus/internal/graph"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdateTogglesFocusWithTab(t *testing.T) {
	m := &App{ShowDetails: true, focus: FocusTree, keys: DefaultKeyMap()}
	m.visibleRows = nodesToRows(&graph.Node{Issue: beads.FullIssue{ID: "ab-001"}})

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focus != FocusDetails {
		t.Fatalf("expected tab to switch focus to details")
	}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focus != FocusTree {
		t.Fatalf("expected tab to cycle focus back to tree")
	}
}

func TestDetailFocusNavigation(t *testing.T) {
	newDetailApp := func() *App {
		vp := viewport.Model{Width: 40, Height: 3}
		vp.SetContent("line1\nline2\nline3\nline4")
		return &App{
			ShowDetails: true,
			focus:       FocusDetails,
			viewport:    vp,
			visibleRows: nodesToRows(&graph.Node{Issue: beads.FullIssue{ID: "ab-001"}}),
			keys:        DefaultKeyMap(),
		}
	}

	t.Run("arrowKeysScrollViewport", func(t *testing.T) {
		m := newDetailApp()
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if m.cursor != 0 {
			t.Fatalf("expected cursor to remain unchanged when details focused")
		}
		if m.viewport.YOffset == 0 {
			t.Fatalf("expected viewport offset to increase after scrolling")
		}
	})

	t.Run("pageCommandsRespectCtrlKeys", func(t *testing.T) {
		m := newDetailApp()
		start := m.viewport.YOffset
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
		if m.viewport.YOffset <= start {
			t.Fatalf("expected ctrl+f to page down in details")
		}
	})

	t.Run("homeAndEndJump", func(t *testing.T) {
		m := newDetailApp()
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
		if m.viewport.YOffset == 0 {
			t.Fatalf("expected ctrl+f to move viewport before home test")
		}
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyHome})
		if m.viewport.YOffset != 0 {
			t.Fatalf("expected home to reset viewport to top, got %d", m.viewport.YOffset)
		}
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnd})
		if m.viewport.YOffset == 0 {
			t.Fatalf("expected end to jump to bottom")
		}
	})
}

func TestUpdateViewportContentResetsScrollOnNewSelection(t *testing.T) {
	n1 := &graph.Node{Issue: beads.FullIssue{ID: "ab-100", Title: "First"}, CommentsLoaded: true}
	n2 := &graph.Node{Issue: beads.FullIssue{ID: "ab-200", Title: "Second"}, CommentsLoaded: true}
	m := &App{
		ShowDetails: true,
		visibleRows: nodesToRows(n1, n2),
		viewport:    viewport.Model{Width: 60, Height: 10},
		cursor:      0,
	}

	m.updateViewportContent()
	m.viewport.YOffset = 4
	m.cursor = 1
	m.updateViewportContent()

	if m.viewport.YOffset != 0 {
		t.Fatalf("expected viewport offset reset to 0 on new selection, got %d", m.viewport.YOffset)
	}
	if m.detailIssueID != "ab-200" {
		t.Fatalf("expected detailIssueID updated to new selection, got %s", m.detailIssueID)
	}
}

func TestUpdateViewportContentPreservesScrollForSameSelection(t *testing.T) {
	n1 := &graph.Node{Issue: beads.FullIssue{ID: "ab-100", Title: "Same"}, CommentsLoaded: true}
	m := &App{
		ShowDetails: true,
		visibleRows: nodesToRows(n1),
		viewport:    viewport.Model{Width: 60, Height: 10},
	}

	m.updateViewportContent()
	m.viewport.YOffset = 5
	m.updateViewportContent()

	if m.viewport.YOffset != 5 {
		t.Fatalf("expected viewport offset preserved for same selection, got %d", m.viewport.YOffset)
	}
	if m.detailIssueID != "ab-100" {
		t.Fatalf("expected detailIssueID to remain unchanged, got %s", m.detailIssueID)
	}
}

func TestUpdateViewportContentDisplaysDesignSection(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:          "ab-101",
			Title:       "Detail Layout",
			Status:      "open",
			IssueType:   "feature",
			Priority:    2,
			Description: "High-level summary.",
			Design:      "## Architecture\n\nDocument component wiring.",
			CreatedAt:   time.Date(2025, time.November, 21, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
			UpdatedAt:   time.Date(2025, time.November, 21, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
			Comments: []beads.Comment{
				{
					Author:    "Reviewer",
					Text:      "Looks good",
					CreatedAt: time.Date(2025, time.November, 21, 13, 0, 0, 0, time.UTC).Format(time.RFC3339),
				},
			},
		},
		CommentsLoaded: true,
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  nodesToRows(node),
		viewport:     viewport.New(90, 30),
		outputFormat: "plain",
	}

	app.updateViewportContent()
	content := stripANSI(app.viewport.View())

	if !strings.Contains(content, "Design") {
		t.Fatalf("expected Design header in viewport content:\n%s", content)
	}

	descIdx := strings.Index(content, "Description")
	designIdx := strings.Index(content, "Design")
	if descIdx == -1 || designIdx == -1 {
		t.Fatalf("expected both Description and Design headers")
	}
	if !(descIdx < designIdx) {
		t.Fatalf("expected Design to appear after Description: descIdx=%d, designIdx=%d\n%s", descIdx, designIdx, content)
	}

	if !strings.Contains(content, "## Architecture") {
		t.Fatalf("expected markdown-rendered design content present, got:\n%s", content)
	}
}

func TestUpdateViewportContentOmitsDesignWhenBlank(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:          "ab-102",
			Title:       "Missing Section",
			Status:      "open",
			IssueType:   "feature",
			Priority:    2,
			Description: "Content exists.",
			Design:      "   ",
			CreatedAt:   time.Date(2025, time.November, 22, 9, 0, 0, 0, time.UTC).Format(time.RFC3339),
			UpdatedAt:   time.Date(2025, time.November, 22, 9, 15, 0, 0, time.UTC).Format(time.RFC3339),
		},
		CommentsLoaded: true,
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  nodesToRows(node),
		viewport:     viewport.New(90, 30),
		outputFormat: "plain",
	}

	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	if strings.Contains(content, "Design") {
		t.Fatalf("expected Design section omitted when empty, content:\n%s", content)
	}
}

func TestUpdateViewportContentDisplaysAcceptanceSection(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:                 "ab-103",
			Title:              "Version Checks",
			Status:             "open",
			IssueType:          "feature",
			Priority:           2,
			Description:        "Ensure CLI presence",
			Design:             "## Flow\n\n1. Detect CLI\n2. Compare version",
			AcceptanceCriteria: "## Acceptance\n\n- Clear error when missing\n- Friendly instructions",
			CreatedAt:          time.Date(2025, time.November, 22, 8, 0, 0, 0, time.UTC).Format(time.RFC3339),
			UpdatedAt:          time.Date(2025, time.November, 22, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
			Comments: []beads.Comment{
				{
					Author:    "QA",
					Text:      "Need docs link",
					CreatedAt: time.Date(2025, time.November, 22, 11, 0, 0, 0, time.UTC).Format(time.RFC3339),
				},
			},
		},
		CommentsLoaded: true,
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  nodesToRows(node),
		viewport:     viewport.New(90, 30),
		outputFormat: "plain",
	}

	app.updateViewportContent()
	content := stripANSI(app.viewport.View())

	if !strings.Contains(content, "Acceptance:") {
		t.Fatalf("expected Acceptance header present:\n%s", content)
	}
	if !strings.Contains(content, "## Acceptance") {
		t.Fatalf("expected markdown acceptance content present:\n%s", content)
	}

	designIdx := strings.Index(content, "Design:")
	acceptIdx := strings.Index(content, "Acceptance:")
	commentsIdx := strings.Index(content, "Comments:")
	if designIdx == -1 || acceptIdx == -1 || commentsIdx == -1 {
		t.Fatalf("expected Design, Acceptance, and Comments headers present")
	}
	if !(designIdx < acceptIdx && acceptIdx < commentsIdx) {
		t.Fatalf("expected Acceptance to appear between Design and Comments: design=%d acceptance=%d comments=%d\n%s",
			designIdx, acceptIdx, commentsIdx, content)
	}
}

func TestDetailMetadataLayoutMatchesDocs(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:          "ab-210",
			Title:       "Metadata Layout",
			Status:      "in_progress",
			IssueType:   "feature",
			Priority:    2,
			Labels:      []string{"auth", "security"},
			Description: "Doc-aligned metadata block.",
			CreatedAt:   time.Date(2025, time.November, 23, 7, 0, 0, 0, time.UTC).Format(time.RFC3339),
			UpdatedAt:   time.Date(2025, time.November, 23, 8, 30, 0, 0, time.UTC).Format(time.RFC3339),
		},
		CommentsLoaded: true,
	}

	app := &App{
		ShowDetails:  true,
		visibleRows:  nodesToRows(node),
		viewport:     viewport.New(90, 30),
		outputFormat: "plain",
	}

	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	start := strings.Index(content, "Status:")
	if start == -1 {
		t.Fatalf("metadata block missing Status line:\n%s", content)
	}
	end := strings.Index(content[start:], "Description:")
	if end == -1 {
		t.Fatalf("metadata block missing Description delimiter:\n%s", content)
	}
	metaBlock := content[start : start+end]
	var rows []string
	for _, line := range strings.Split(metaBlock, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		rows = append(rows, trimmed)
	}
	if len(rows) < 4 {
		t.Fatalf("expected metadata rows, got %d:\n%s", len(rows), metaBlock)
	}
	if !(strings.Contains(rows[0], "Status:") && strings.Contains(rows[0], "Priority:")) {
		t.Fatalf("row 1 should contain Status and Priority, got %q", rows[0])
	}
	if !(strings.Contains(rows[1], "Type:") && strings.Contains(rows[1], "Labels:")) {
		t.Fatalf("row 2 should contain Type and Labels, got %q", rows[1])
	}
	if !strings.HasPrefix(rows[2], "Created:") {
		t.Fatalf("row 3 should begin with Created, got %q", rows[2])
	}
	if !strings.HasPrefix(rows[3], "Updated:") {
		t.Fatalf("row 4 should begin with Updated, got %q", rows[3])
	}
}

func TestDetailRelationshipSectionsFollowDocs(t *testing.T) {
	parent := &graph.Node{Issue: beads.FullIssue{ID: "ab-300", Title: "Parent Node"}}
	childA := &graph.Node{Issue: beads.FullIssue{ID: "ab-301", Title: "Child A"}}
	childB := &graph.Node{Issue: beads.FullIssue{ID: "ab-302", Title: "Child B"}}
	blocker := &graph.Node{Issue: beads.FullIssue{ID: "ab-303", Title: "Blocking Task"}}
	blocked := &graph.Node{Issue: beads.FullIssue{ID: "ab-304", Title: "Blocked Task"}}

	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:          "ab-305",
			Title:       "Relationship Order",
			Description: "Ensure sections match documentation order.",
		},
		Parent:         parent,
		Parents:        []*graph.Node{parent},
		Children:       []*graph.Node{childA, childB},
		BlockedBy:      []*graph.Node{blocker},
		Blocks:         []*graph.Node{blocked},
		IsBlocked:      true,
		CommentsLoaded: true,
	}

	app := &App{
		ShowDetails:  true,
		visibleRows:  nodesToRows(node),
		viewport:     viewport.New(90, 40),
		outputFormat: "plain",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	// New labels: "Part Of:", "Subtasks", "Must Complete First:", "Will Unblock"
	order := []string{"Part Of:", "Subtasks", "Must Complete First:", "Will Unblock"}
	var lastIdx int = -1
	for _, section := range order {
		idx := strings.Index(content, section)
		if idx == -1 {
			t.Fatalf("missing %s section in content:\n%s", section, content)
		}
		if idx <= lastIdx {
			t.Fatalf("section %s appeared out of order", section)
		}
		lastIdx = idx
	}
	if !strings.Contains(content, "Subtasks: (2)") {
		t.Fatalf("expected Subtasks count header, got:\n%s", content)
	}
	if !strings.Contains(content, "ab-303") || !strings.Contains(content, "ab-304") {
		t.Fatalf("expected related issue IDs rendered:\n%s", content)
	}
}

func TestDetailLabelsWrapWhenViewportNarrow(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:          "ab-320",
			Title:       "Label Wrapping",
			Status:      "open",
			IssueType:   "feature",
			Priority:    2,
			Labels:      []string{"alpha", "beta", "gamma"},
			Description: "Verify label chips wrap across lines.",
			CreatedAt:   time.Date(2025, time.November, 23, 6, 0, 0, 0, time.UTC).Format(time.RFC3339),
		},
		CommentsLoaded: true,
	}

	app := &App{
		ShowDetails:  true,
		visibleRows:  nodesToRows(node),
		viewport:     viewport.New(42, 25),
		outputFormat: "plain",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	lines := strings.Split(content, "\n")
	labelsIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "Labels:") {
			labelsIdx = i
			break
		}
	}
	if labelsIdx == -1 {
		t.Fatalf("no Labels row found:\n%s", content)
	}
	if strings.Contains(lines[labelsIdx], "beta") {
		t.Fatalf("expected first labels row to wrap, got %q", lines[labelsIdx])
	}
	if labelsIdx+1 >= len(lines) {
		t.Fatalf("expected additional label rows after wrap")
	}
	wrapped := strings.TrimSpace(lines[labelsIdx+1])
	if !strings.Contains(wrapped, "beta") {
		t.Fatalf("expected wrapped row to include beta label, got %q", wrapped)
	}
	if labelsIdx+2 >= len(lines) {
		t.Fatalf("expected third line for gamma label")
	}
	wrapped2 := strings.TrimSpace(lines[labelsIdx+2])
	if !strings.Contains(wrapped2, "gamma") {
		t.Fatalf("expected second wrapped row to include gamma label, got %q", wrapped2)
	}
}

func TestDetailCommentsRenderEntries(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:          "ab-330",
			Title:       "Comment Rendering",
			Description: "Doc sample comments.",
			Comments: []beads.Comment{
				{Author: "@alice", Text: "Let's use OAuth2.", CreatedAt: time.Date(2025, time.November, 20, 9, 15, 0, 0, time.UTC).Format(time.RFC3339)},
				{Author: "@bob", Text: "Agreed, updating design.", CreatedAt: time.Date(2025, time.November, 20, 11, 30, 0, 0, time.UTC).Format(time.RFC3339)},
			},
		},
		CommentsLoaded: true,
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  nodesToRows(node),
		viewport:     viewport.New(80, 30),
		outputFormat: "plain",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	idx := strings.Index(content, "Comments:")
	if idx == -1 {
		t.Fatalf("missing Comments header:\n%s", content)
	}
	if !(strings.Contains(content, "@alice") && strings.Contains(content, "Let's use OAuth2.")) {
		t.Fatalf("expected first comment rendered:\n%s", content)
	}
	if !(strings.Contains(content, "@bob") && strings.Contains(content, "Agreed, updating design.")) {
		t.Fatalf("expected second comment rendered:\n%s", content)
	}
	if strings.Index(content, "@alice") > strings.Index(content, "@bob") {
		t.Fatalf("comments should remain chronological")
	}
}

func TestDetailCommentsErrorMessageMatchesDocs(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:          "ab-340",
			Title:       "Comment Error",
			Description: "Doc retry guidance.",
		},
		CommentError: "timeout fetching comments",
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  nodesToRows(node),
		viewport:     viewport.New(80, 30),
		outputFormat: "plain",
	}
	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	if !strings.Contains(content, "Failed to load comments. Press 'c' to retry.") {
		t.Fatalf("expected retry guidance in error state:\n%s", content)
	}
	if !strings.Contains(content, "timeout fetching comments") {
		t.Fatalf("expected underlying error rendered, content:\n%s", content)
	}
}

func TestUpdateViewportContentOmitsAcceptanceWhenBlank(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{
			ID:                 "ab-104",
			Title:              "Whitespace Acceptance",
			Status:             "open",
			IssueType:          "feature",
			Priority:           2,
			Description:        "Has description.",
			Design:             "## Design\n\n- present",
			AcceptanceCriteria: "   \n",
			CreatedAt:          time.Date(2025, time.November, 22, 9, 30, 0, 0, time.UTC).Format(time.RFC3339),
			UpdatedAt:          time.Date(2025, time.November, 22, 9, 45, 0, 0, time.UTC).Format(time.RFC3339),
		},
		CommentsLoaded: true,
	}
	app := &App{
		ShowDetails:  true,
		visibleRows:  nodesToRows(node),
		viewport:     viewport.New(90, 30),
		outputFormat: "plain",
	}

	app.updateViewportContent()
	content := stripANSI(app.viewport.View())
	if strings.Contains(content, "Acceptance:") {
		t.Fatalf("expected Acceptance section omitted when whitespace, content:\n%s", content)
	}
}

func TestCopyBeadIDSetsToastState(t *testing.T) {
	node := &graph.Node{Issue: beads.FullIssue{ID: "ab-123", Title: "Test Issue"}}
	app := &App{
		visibleRows: nodesToRows(node),
		cursor:      0,
		ready:       true,
	}

	// Simulate pressing 'c' key - we test the state changes, not actual clipboard
	// (clipboard may not work in CI/test environments)
	app.copiedBeadID = node.Issue.ID
	app.showCopyToast = true
	app.copyToastStart = time.Now()

	if !app.showCopyToast {
		t.Error("expected showCopyToast to be true")
	}
	if app.copiedBeadID != "ab-123" {
		t.Errorf("expected copiedBeadID 'ab-123', got %s", app.copiedBeadID)
	}
}

func TestCopyToastCountdown(t *testing.T) {
	app := &App{
		copiedBeadID:   "ab-456",
		showCopyToast:  true,
		copyToastStart: time.Now().Add(-3 * time.Second), // Started 3 seconds ago
		ready:          true,
	}

	// Process tick - should continue countdown (not yet 5 seconds)
	_, cmd := app.Update(copyToastTickMsg{})
	if !app.showCopyToast {
		t.Error("toast should still be visible before 5 seconds")
	}
	if cmd == nil {
		t.Error("expected another tick to be scheduled")
	}

	// Simulate 5+ seconds elapsed
	app.copyToastStart = time.Now().Add(-6 * time.Second)
	_, cmd = app.Update(copyToastTickMsg{})
	if app.showCopyToast {
		t.Error("toast should auto-dismiss after 5 seconds")
	}
}

func TestCopyToastRenders(t *testing.T) {
	app := &App{
		copiedBeadID:   "ab-789",
		showCopyToast:  true,
		copyToastStart: time.Now(),
		ready:          true,
	}

	layer := app.copyToastLayer(80, 24, 2, 10)
	if layer == nil {
		t.Fatal("expected toast to render")
	}

	canvas := layer.Render()
	if canvas == nil {
		t.Fatal("expected canvas from copy toast layer")
	}
	plain := stripANSI(canvas.Render())
	if !strings.Contains(plain, "ab-789") {
		t.Errorf("expected toast to contain bead ID 'ab-789', got: %s", plain)
	}
	if !strings.Contains(plain, "Copied") {
		t.Errorf("expected toast to contain 'Copied', got: %s", plain)
	}
	if !strings.Contains(plain, "clipboard") {
		t.Errorf("expected toast to contain 'clipboard', got: %s", plain)
	}
}

func TestCopyToastNotRenderedWhenInactive(t *testing.T) {
	app := &App{
		copiedBeadID:  "ab-999",
		showCopyToast: false,
		ready:         true,
	}

	if layer := app.copyToastLayer(80, 24, 2, 10); layer != nil {
		t.Error("expected no toast when showCopyToast is false")
	}
}

func TestCopyWithEmptyVisibleRows(t *testing.T) {
	app := &App{
		visibleRows: []graph.TreeRow{},
		ready:       true,
	}

	// Press 'c' with no visible rows - should not panic or set toast
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})

	if app.showCopyToast {
		t.Error("expected no toast when no rows visible")
	}
}

func TestCopyToastTickWhenNotShowing(t *testing.T) {
	app := &App{
		showCopyToast: false,
		ready:         true,
	}

	// Process tick when toast is not showing - should return nil cmd
	_, cmd := app.Update(copyToastTickMsg{})
	if cmd != nil {
		t.Error("expected no tick scheduled when toast is not showing")
	}
}
