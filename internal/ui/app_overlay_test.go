package ui

import (
	"strings"
	"testing"
	"time"

	"abacus/internal/beads"
	"abacus/internal/graph"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHelpToggle(t *testing.T) {
	app := &App{
		visibleRows: nodesToRows(&graph.Node{Issue: beads.FullIssue{ID: "ab-001"}}),
		keys:        DefaultKeyMap(),
	}

	// Initially help is not shown
	if app.showHelp {
		t.Error("expected showHelp to be false initially")
	}

	// Press ? to show help
	result, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	app = result.(*App)
	if !app.showHelp {
		t.Error("expected showHelp to be true after pressing ?")
	}

	// Press ? again to hide help
	result, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	app = result.(*App)
	if app.showHelp {
		t.Error("expected showHelp to be false after pressing ? again")
	}
}

func TestHelpDismissWithEsc(t *testing.T) {
	app := &App{
		showHelp: true,
		keys:     DefaultKeyMap(),
	}

	result, _ := app.Update(tea.KeyMsg{Type: tea.KeyEscape})
	app = result.(*App)

	if app.showHelp {
		t.Error("expected showHelp to be false after pressing Esc")
	}
}

func TestHelpDismissWithQ(t *testing.T) {
	app := &App{
		showHelp:    true,
		keys:        DefaultKeyMap(),
		visibleRows: nodesToRows(&graph.Node{Issue: beads.FullIssue{ID: "ab-001"}}),
	}

	// q should dismiss help, NOT quit the app
	result, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	app = result.(*App)

	if app.showHelp {
		t.Error("expected showHelp to be false after pressing q")
	}
	if cmd != nil {
		t.Error("expected no quit command when dismissing help with q")
	}
}

func TestHelpBlocksOtherKeys(t *testing.T) {
	app := &App{
		showHelp:    true,
		keys:        DefaultKeyMap(),
		visibleRows: nodesToRows(&graph.Node{Issue: beads.FullIssue{ID: "ab-001"}}),
		cursor:      0,
	}

	initialCursor := app.cursor

	// Navigation keys should be blocked
	result, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	app = result.(*App)
	if app.cursor != initialCursor {
		t.Error("expected cursor to remain unchanged when help is shown")
	}
	if !app.showHelp {
		t.Error("expected help to still be shown after pressing j")
	}

	// Search key should be blocked
	result, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	app = result.(*App)
	if app.searching {
		t.Error("expected searching to remain false when help is shown")
	}
	if !app.showHelp {
		t.Error("expected help to still be shown after pressing /")
	}
}

func TestHelpOverlayInView(t *testing.T) {
	app := &App{
		showHelp:    true,
		keys:        DefaultKeyMap(),
		visibleRows: nodesToRows(&graph.Node{Issue: beads.FullIssue{ID: "ab-001"}}),
		width:       80,
		height:      24,
		ready:       true,
	}

	view := app.View()

	// View should contain help overlay content
	if !strings.Contains(view, "ABACUS HELP") {
		t.Error("expected view to contain 'ABACUS HELP' when help is shown")
	}
	if !strings.Contains(view, "NAVIGATION") {
		t.Error("expected view to contain 'NAVIGATION' section when help is shown")
	}
}

func TestStatusOverlayKeepsBaseContentVisible(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{ID: "ab-ov1", Title: "Overlay Test", Status: "open"},
	}
	app := &App{
		ready:           true,
		width:           80,
		height:          24,
		visibleRows:     []graph.TreeRow{{Node: node}},
		cursor:          0,
		activeOverlay:   OverlayStatus,
		statusOverlay:   NewStatusOverlay(node.Issue.ID, node.Issue.Title, node.Issue.Status),
		repoName:        "abacus",
		lastRefreshTime: time.Now(),
	}

	view := app.View()
	plain := stripANSI(view)

	if !strings.Contains(plain, node.Issue.ID) {
		t.Fatalf("expected base content (tree) to remain visible, got:\n%s", plain)
	}
	if !strings.Contains(plain, "Status") {
		t.Fatalf("expected status overlay content to be present, got:\n%s", plain)
	}
	// Dimming is now handled via theme-level color blending (useStyleTheme),
	// not ANSI dim sequences. Background coverage is verified via layer helper tests.
}

func TestCreateOverlayShowsErrorToast(t *testing.T) {
	node := &graph.Node{
		Issue: beads.FullIssue{ID: "ab-ov2", Title: "Overlay Toast", Status: "open"},
	}
	app := &App{
		ready:           true,
		width:           80,
		height:          24,
		visibleRows:     []graph.TreeRow{{Node: node}},
		cursor:          0,
		activeOverlay:   OverlayCreate,
		createOverlay:   NewCreateOverlay(CreateOverlayOptions{}),
		showErrorToast:  true,
		lastError:       "Error: backend unavailable",
		errorToastStart: time.Now(),
		repoName:        "abacus",
		lastRefreshTime: time.Now(),
	}

	view := app.View()
	plain := stripANSI(view)
	if !strings.Contains(plain, "⚠ Error") {
		t.Fatalf("expected error toast to appear on create overlay, got:\n%s", plain)
	}
	if !strings.Contains(plain, "ABACUS") {
		t.Fatalf("expected header content to remain visible beneath overlay, got:\n%s", plain)
	}
	// Dimming is now handled via theme-level color blending (useStyleTheme),
	// not ANSI dim sequences. Background coverage is validated in layer helper tests.
}

func TestNewLabelToastDisplay(t *testing.T) {
	t.Run("NewLabelAddedMsgTriggersToast", func(t *testing.T) {
		app := &App{
			ready: true,
			roots: []*graph.Node{},
		}

		// Send NewLabelAddedMsg
		msg := NewLabelAddedMsg{Label: "test-label"}
		model, cmd := app.Update(msg)

		app = model.(*App)

		// Verify toast is visible
		if !app.newLabelToastVisible {
			t.Error("expected newLabelToastVisible to be true")
		}

		// Verify label name is set
		if app.newLabelToastLabel != "test-label" {
			t.Errorf("expected label 'test-label', got '%s'", app.newLabelToastLabel)
		}

		// Verify toast start time is recent (within last second)
		if time.Since(app.newLabelToastStart) > time.Second {
			t.Error("expected newLabelToastStart to be recent")
		}

		// Verify tick scheduler was returned
		if cmd == nil {
			t.Error("expected command to schedule toast tick")
		}
	})
}

func TestNewLabelToastTimeout(t *testing.T) {
	t.Run("ToastClearsAfter3Seconds", func(t *testing.T) {
		app := &App{
			ready:                true,
			roots:                []*graph.Node{},
			newLabelToastVisible: true,
			newLabelToastLabel:   "test-label",
			newLabelToastStart:   time.Now().Add(-4 * time.Second), // 4 seconds ago
		}

		// Send tick message
		msg := newLabelToastTickMsg{}
		model, cmd := app.Update(msg)

		app = model.(*App)

		// Verify toast is cleared
		if app.newLabelToastVisible {
			t.Error("expected newLabelToastVisible to be false after timeout")
		}

		// Verify no further ticks scheduled
		if cmd != nil {
			t.Error("expected no command after toast cleared")
		}
	})

	t.Run("ToastRemainsVisibleBefore3Seconds", func(t *testing.T) {
		app := &App{
			ready:                true,
			roots:                []*graph.Node{},
			newLabelToastVisible: true,
			newLabelToastLabel:   "test-label",
			newLabelToastStart:   time.Now().Add(-1 * time.Second), // 1 second ago
		}

		// Send tick message
		msg := newLabelToastTickMsg{}
		model, cmd := app.Update(msg)

		app = model.(*App)

		// Verify toast is still visible
		if !app.newLabelToastVisible {
			t.Error("expected newLabelToastVisible to remain true before timeout")
		}

		// Verify tick is rescheduled
		if cmd == nil {
			t.Error("expected command to reschedule toast tick")
		}
	})

	t.Run("TickHandlerReturnsEarlyWhenNotVisible", func(t *testing.T) {
		app := &App{
			ready:                true,
			roots:                []*graph.Node{},
			newLabelToastVisible: false,
			newLabelToastLabel:   "",
		}

		// Send tick message when toast not visible
		msg := newLabelToastTickMsg{}
		model, cmd := app.Update(msg)

		app = model.(*App)

		// Verify no command returned
		if cmd != nil {
			t.Error("expected no command when toast not visible")
		}
	})
}

func TestNewAssigneeToastDisplay(t *testing.T) {
	t.Run("NewAssigneeAddedMsgTriggersToast", func(t *testing.T) {
		app := &App{
			ready: true,
			roots: []*graph.Node{},
		}

		// Send NewAssigneeAddedMsg
		msg := NewAssigneeAddedMsg{Assignee: "test-user"}
		model, cmd := app.Update(msg)

		app = model.(*App)

		// Verify toast is visible
		if !app.newAssigneeToastVisible {
			t.Error("expected newAssigneeToastVisible to be true")
		}

		// Verify assignee name is set
		if app.newAssigneeToastAssignee != "test-user" {
			t.Errorf("expected assignee 'test-user', got '%s'", app.newAssigneeToastAssignee)
		}

		// Verify toast start time is recent
		if time.Since(app.newAssigneeToastStart) > time.Second {
			t.Error("expected newAssigneeToastStart to be recent")
		}

		// Verify tick scheduler was returned
		if cmd == nil {
			t.Error("expected command to schedule toast tick")
		}
	})
}

func TestNewAssigneeToastTimeout(t *testing.T) {
	t.Run("ToastClearsAfter3Seconds", func(t *testing.T) {
		app := &App{
			ready:                    true,
			roots:                    []*graph.Node{},
			newAssigneeToastVisible:  true,
			newAssigneeToastAssignee: "test-user",
			newAssigneeToastStart:    time.Now().Add(-4 * time.Second), // 4 seconds ago
		}

		// Send tick message
		msg := newAssigneeToastTickMsg{}
		model, cmd := app.Update(msg)

		app = model.(*App)

		// Verify toast is cleared
		if app.newAssigneeToastVisible {
			t.Error("expected newAssigneeToastVisible to be false after timeout")
		}

		// Verify no further ticks scheduled
		if cmd != nil {
			t.Error("expected no command after toast cleared")
		}
	})

	t.Run("ToastRemainsVisibleBefore3Seconds", func(t *testing.T) {
		app := &App{
			ready:                    true,
			roots:                    []*graph.Node{},
			newAssigneeToastVisible:  true,
			newAssigneeToastAssignee: "test-user",
			newAssigneeToastStart:    time.Now().Add(-1 * time.Second), // 1 second ago
		}

		// Send tick message
		msg := newAssigneeToastTickMsg{}
		model, cmd := app.Update(msg)

		app = model.(*App)

		// Verify toast is still visible
		if !app.newAssigneeToastVisible {
			t.Error("expected newAssigneeToastVisible to remain true before timeout")
		}

		// Verify tick is rescheduled
		if cmd == nil {
			t.Error("expected command to reschedule toast tick")
		}
	})

	t.Run("TickHandlerReturnsEarlyWhenNotVisible", func(t *testing.T) {
		app := &App{
			ready:                    true,
			roots:                    []*graph.Node{},
			newAssigneeToastVisible:  false,
			newAssigneeToastAssignee: "",
		}

		// Send tick message when toast not visible
		msg := newAssigneeToastTickMsg{}
		model, cmd := app.Update(msg)

		app = model.(*App)

		// Verify no command returned
		if cmd != nil {
			t.Error("expected no command when toast not visible")
		}
	})
}

func TestToastRendering(t *testing.T) {
	t.Run("RenderNewLabelToastReturnsEmptyWhenNotVisible", func(t *testing.T) {
		app := &App{
			newLabelToastVisible: false,
		}

		if layer := app.newLabelToastLayer(80, 24, 2, 10); layer != nil {
			t.Error("expected nil layer when toast not visible")
		}
	})

	t.Run("RenderNewLabelToastReturnsEmptyWhenLabelEmpty", func(t *testing.T) {
		app := &App{
			newLabelToastVisible: true,
			newLabelToastLabel:   "",
		}

		if layer := app.newLabelToastLayer(80, 24, 2, 10); layer != nil {
			t.Error("expected nil layer when label is empty")
		}
	})

	t.Run("RenderNewLabelToastReturnsFormattedToast", func(t *testing.T) {
		app := &App{
			newLabelToastVisible: true,
			newLabelToastLabel:   "test-label",
			newLabelToastStart:   time.Now(),
		}

		layer := app.newLabelToastLayer(80, 24, 2, 10)
		if layer == nil {
			t.Error("expected non-empty toast when visible")
			return
		}

		canvas := layer.Render()
		if canvas == nil {
			t.Fatal("expected canvas from label toast layer")
		}
		result := canvas.Render()

		// Check for label name in output
		if !strings.Contains(result, "test-label") {
			t.Error("expected toast to contain label name")
		}

		// Check for checkmark symbol
		if !strings.Contains(result, "✓") {
			t.Error("expected toast to contain checkmark")
		}

		// Check for "New Label Added" text
		if !strings.Contains(result, "New Label Added") {
			t.Error("expected toast to contain 'New Label Added' text")
		}
	})

	t.Run("RenderNewLabelToastReturnsEmptyAfterTimeout", func(t *testing.T) {
		app := &App{
			newLabelToastVisible: true,
			newLabelToastLabel:   "test-label",
			newLabelToastStart:   time.Now().Add(-4 * time.Second), // 4 seconds ago
		}

		if layer := app.newLabelToastLayer(80, 24, 2, 10); layer != nil {
			t.Error("expected nil layer when toast has timed out")
		}
	})

	t.Run("RenderNewAssigneeToastReturnsEmptyWhenNotVisible", func(t *testing.T) {
		app := &App{
			newAssigneeToastVisible: false,
		}

		if layer := app.newAssigneeToastLayer(80, 24, 2, 10); layer != nil {
			t.Error("expected nil layer when toast not visible")
		}
	})

	t.Run("RenderNewAssigneeToastReturnsEmptyWhenAssigneeEmpty", func(t *testing.T) {
		app := &App{
			newAssigneeToastVisible:  true,
			newAssigneeToastAssignee: "",
		}

		if layer := app.newAssigneeToastLayer(80, 24, 2, 10); layer != nil {
			t.Error("expected nil layer when assignee is empty")
		}
	})

	t.Run("RenderNewAssigneeToastReturnsFormattedToast", func(t *testing.T) {
		app := &App{
			newAssigneeToastVisible:  true,
			newAssigneeToastAssignee: "test-user",
			newAssigneeToastStart:    time.Now(),
		}

		layer := app.newAssigneeToastLayer(80, 24, 2, 10)
		if layer == nil {
			t.Error("expected non-empty toast when visible")
			return
		}

		canvas := layer.Render()
		if canvas == nil {
			t.Fatal("expected canvas from assignee toast layer")
		}
		result := canvas.Render()

		// Check for assignee name in output
		if !strings.Contains(result, "test-user") {
			t.Error("expected toast to contain assignee name")
		}

		// Check for checkmark symbol
		if !strings.Contains(result, "✓") {
			t.Error("expected toast to contain checkmark")
		}

		// Check for "New Assignee Added" text
		if !strings.Contains(result, "New Assignee Added") {
			t.Error("expected toast to contain 'New Assignee Added' text")
		}
	})

	t.Run("RenderNewAssigneeToastReturnsEmptyAfterTimeout", func(t *testing.T) {
		app := &App{
			newAssigneeToastVisible:  true,
			newAssigneeToastAssignee: "test-user",
			newAssigneeToastStart:    time.Now().Add(-4 * time.Second), // 4 seconds ago
		}

		if layer := app.newAssigneeToastLayer(80, 24, 2, 10); layer != nil {
			t.Error("expected nil layer when toast has timed out")
		}
	})

	t.Run("RenderUpdateToastShowsUpdatedLabel", func(t *testing.T) {
		app := &App{
			createToastVisible:  true,
			createToastBeadID:   "ab-123",
			createToastTitle:    "",
			createToastStart:    time.Now(),
			createToastIsUpdate: true,
		}

		layer := app.createToastLayer(80, 24, 2, 10)
		if layer == nil {
			t.Fatal("expected update toast layer")
		}
		canvas := layer.Render()
		if canvas == nil {
			t.Fatal("expected canvas from update toast layer")
		}
		output := canvas.Render()
		if !strings.Contains(output, "Updated") {
			t.Errorf("expected toast to contain 'Updated', got: %s", output)
		}
		if !strings.Contains(output, "ab-123") {
			t.Errorf("expected toast to contain bead ID, got: %s", output)
		}
		if !strings.Contains(output, "[") {
			t.Errorf("expected toast to show countdown, got: %s", output)
		}
	})

	t.Run("RenderUpdateToastTimeoutHidesLayer", func(t *testing.T) {
		app := &App{
			createToastVisible:  true,
			createToastBeadID:   "ab-123",
			createToastStart:    time.Now().Add(-8 * time.Second),
			createToastIsUpdate: true,
		}

		if layer := app.createToastLayer(80, 24, 2, 10); layer != nil {
			t.Error("expected nil update toast layer after timeout")
		}
	})
}

func TestGlobalHotkeysDisabledDuringTextInput(t *testing.T) {
	t.Run("nKeyIgnoredWhenCreateOverlayActive", func(t *testing.T) {
		// Setup app with create overlay active
		node := &graph.Node{Issue: beads.FullIssue{ID: "ab-001", Title: "Test"}}
		app := &App{
			visibleRows: nodesToRows(node),
			keys:        DefaultKeyMap(),
			ready:       true,
		}

		// Create initial overlay
		createOverlay := NewCreateOverlay(CreateOverlayOptions{
			DefaultParentID:  "ab-001",
			AvailableParents: []ParentOption{{ID: "ab-001", Display: "ab-001 Test"}},
		})
		app.createOverlay = createOverlay
		app.activeOverlay = OverlayCreate

		// Press 'n' key - should be passed to overlay, not create new overlay
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
		result, _ := app.Update(msg)
		app = result.(*App)

		// Verify: Still only one overlay, not a new one created
		if app.activeOverlay != OverlayCreate {
			t.Error("expected CreateOverlay to remain active")
		}
		if app.createOverlay == nil {
			t.Error("expected createOverlay to still exist")
		}
	})

	t.Run("nKeyWorksNormallyWhenNoOverlayActive", func(t *testing.T) {
		// Setup app without any overlay
		node := &graph.Node{Issue: beads.FullIssue{ID: "ab-002", Title: "Test"}}
		app := &App{
			visibleRows:   nodesToRows(node),
			keys:          DefaultKeyMap(),
			ready:         true,
			activeOverlay: OverlayNone,
		}

		// Press 'n' key - should create new overlay
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
		result, _ := app.Update(msg)
		app = result.(*App)

		// Verify: CreateOverlay was created
		if app.activeOverlay != OverlayCreate {
			t.Error("expected CreateOverlay to be active")
		}
		if app.createOverlay == nil {
			t.Error("expected createOverlay to be created")
		}
	})

	t.Run("sKeyIgnoredWhenCreateOverlayActive", func(t *testing.T) {
		// Setup app with create overlay active
		node := &graph.Node{Issue: beads.FullIssue{ID: "ab-003", Title: "Test", Status: "open"}}
		app := &App{
			visibleRows: nodesToRows(node),
			keys:        DefaultKeyMap(),
			ready:       true,
		}

		// Create overlay
		createOverlay := NewCreateOverlay(CreateOverlayOptions{})
		app.createOverlay = createOverlay
		app.activeOverlay = OverlayCreate

		// Press 's' key - should be passed to overlay, not open status overlay
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
		result, _ := app.Update(msg)
		app = result.(*App)

		// Verify: Still in CreateOverlay, not StatusOverlay
		if app.activeOverlay != OverlayCreate {
			t.Error("expected CreateOverlay to remain active")
		}
		if app.statusOverlay != nil {
			t.Error("expected statusOverlay to not be created")
		}
	})

	t.Run("CapitalLKeyIgnoredWhenCreateOverlayActive", func(t *testing.T) {
		// Setup app with create overlay active
		node := &graph.Node{Issue: beads.FullIssue{ID: "ab-004", Title: "Test", Labels: []string{"test"}}}
		app := &App{
			visibleRows: nodesToRows(node),
			keys:        DefaultKeyMap(),
			ready:       true,
		}

		// Create overlay
		createOverlay := NewCreateOverlay(CreateOverlayOptions{})
		app.createOverlay = createOverlay
		app.activeOverlay = OverlayCreate

		// Press 'L' key - should be passed to overlay, not open labels overlay
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'L'}}
		result, _ := app.Update(msg)
		app = result.(*App)

		// Verify: Still in CreateOverlay, not LabelsOverlay
		if app.activeOverlay != OverlayCreate {
			t.Error("expected CreateOverlay to remain active")
		}
		if app.labelsOverlay != nil {
			t.Error("expected labelsOverlay to not be created")
		}
	})

	t.Run("xKeyIgnoredWhenCreateOverlayActive", func(t *testing.T) {
		// Setup app with create overlay active and mock client
		node := &graph.Node{Issue: beads.FullIssue{ID: "ab-005", Title: "Test", Status: "open"}}
		mock := beads.NewMockClient()
		app := &App{
			visibleRows: nodesToRows(node),
			keys:        DefaultKeyMap(),
			ready:       true,
			client:      mock,
		}

		// Create overlay
		createOverlay := NewCreateOverlay(CreateOverlayOptions{})
		app.createOverlay = createOverlay
		app.activeOverlay = OverlayCreate

		// Press 'x' key - should be passed to overlay, not trigger close
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
		result, _ := app.Update(msg)
		app = result.(*App)

		// Verify: Still in CreateOverlay, close was not triggered
		if app.activeOverlay != OverlayCreate {
			t.Error("expected CreateOverlay to remain active")
		}
		// No way to directly check if close wasn't called, but overlay remaining active is sufficient
	})

	t.Run("StatusOverlayBlocksGlobalHotkeys", func(t *testing.T) {
		// Setup app with status overlay active
		node := &graph.Node{Issue: beads.FullIssue{ID: "ab-006", Title: "Test", Status: "open"}}
		app := &App{
			visibleRows: nodesToRows(node),
			keys:        DefaultKeyMap(),
			ready:       true,
		}

		// Create status overlay
		statusOverlay := NewStatusOverlay("ab-006", "Test", "open")
		app.statusOverlay = statusOverlay
		app.activeOverlay = OverlayStatus

		// Press 'n' key - should be passed to overlay, not create new bead overlay
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
		result, _ := app.Update(msg)
		app = result.(*App)

		// Verify: Still in StatusOverlay, CreateOverlay not created
		if app.activeOverlay != OverlayStatus {
			t.Error("expected StatusOverlay to remain active")
		}
		if app.createOverlay != nil {
			t.Error("expected createOverlay to not be created")
		}
	})

	t.Run("LabelsOverlayBlocksGlobalHotkeys", func(t *testing.T) {
		// Setup app with labels overlay active
		node := &graph.Node{Issue: beads.FullIssue{ID: "ab-007", Title: "Test", Labels: []string{"existing"}}}
		app := &App{
			visibleRows: nodesToRows(node),
			keys:        DefaultKeyMap(),
			ready:       true,
			roots:       []*graph.Node{node},
		}

		// Create labels overlay
		labelsOverlay := NewLabelsOverlay("ab-007", "Test", []string{"existing"}, []string{"existing", "other"})
		app.labelsOverlay = labelsOverlay
		app.activeOverlay = OverlayLabels

		// Press 'n' key - should be passed to overlay, not create new bead overlay
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
		result, _ := app.Update(msg)
		app = result.(*App)

		// Verify: Still in LabelsOverlay, CreateOverlay not created
		if app.activeOverlay != OverlayLabels {
			t.Error("expected LabelsOverlay to remain active")
		}
		if app.createOverlay != nil {
			t.Error("expected createOverlay to not be created")
		}
	})
}
