package ui

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"abacus/internal/beads"
	"abacus/internal/graph"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/wordwrap"
)

func TestFindBeadsDBPrefersEnv(t *testing.T) {
	t.Setenv("BEADS_DB", "")
	tmp := t.TempDir()
	dbFile := filepath.Join(tmp, "custom.db")
	if err := os.WriteFile(dbFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("write db file: %v", err)
	}
	t.Setenv("BEADS_DB", dbFile)

	cleanup := changeWorkingDir(t, tmp)
	defer cleanup()

	path, modTime, err := findBeadsDB()
	if err != nil {
		t.Fatalf("findBeadsDB: %v", err)
	}
	if normalizePath(t, path) != normalizePath(t, dbFile) {
		t.Fatalf("expected path %s, got %s", dbFile, path)
	}
	info, err := os.Stat(dbFile)
	if err != nil {
		t.Fatalf("stat db file: %v", err)
	}
	if !modTime.Equal(info.ModTime()) {
		t.Fatalf("expected mod time %v, got %v", info.ModTime(), modTime)
	}
}

func TestFindBeadsDBWalksUpDirectories(t *testing.T) {
	t.Setenv("BEADS_DB", "")
	root := t.TempDir()
	beadsDir := filepath.Join(root, ".beads")
	if err := os.MkdirAll(beadsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	dbFile := filepath.Join(beadsDir, "beads.db")
	if err := os.WriteFile(dbFile, []byte("db"), 0o644); err != nil {
		t.Fatalf("write db: %v", err)
	}
	nested := filepath.Join(root, "nested", "deep")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	cleanup := changeWorkingDir(t, nested)
	defer cleanup()

	path, _, err := findBeadsDB()
	if err != nil {
		t.Fatalf("findBeadsDB: %v", err)
	}
	if normalizePath(t, path) != normalizePath(t, dbFile) {
		t.Fatalf("expected %s, got %s", dbFile, path)
	}
}

func TestFindBeadsDBFallsBackToDefault(t *testing.T) {
	t.Setenv("BEADS_DB", "")
	projectDir := t.TempDir()
	cleanup := changeWorkingDir(t, projectDir)
	defer cleanup()

	home := t.TempDir()
	t.Setenv("HOME", home)
	defaultDir := filepath.Join(home, ".beads")
	if err := os.MkdirAll(defaultDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	defaultDB := filepath.Join(defaultDir, "default.db")
	if err := os.WriteFile(defaultDB, []byte("db"), 0o644); err != nil {
		t.Fatalf("write db: %v", err)
	}

	path, _, err := findBeadsDB()
	if err != nil {
		t.Fatalf("findBeadsDB: %v", err)
	}
	if normalizePath(t, path) != normalizePath(t, defaultDB) {
		t.Fatalf("expected fallback %s, got %s", defaultDB, path)
	}
}

func TestBuildMarkdownRendererPlainStyle(t *testing.T) {
	text := "alpha beta gamma delta"
	width := 6
	want := wordwrap.String(text, width)

	render := buildMarkdownRenderer("plain", width)
	if got := render(text); got != want {
		t.Fatalf("expected plain renderer to match fallback %q, got %q", want, got)
	}
}

func TestRecalcVisibleRowsMatchesIDs(t *testing.T) {
	nodes := []*graph.Node{
		{Issue: beads.FullIssue{ID: "ab-123", Title: "Alpha"}},
		{Issue: beads.FullIssue{ID: "ab-456", Title: "Beta"}},
	}
	m := App{
		roots:      nodes,
		filterText: "ab-123",
	}
	m.recalcVisibleRows()

	if len(m.visibleRows) != 1 {
		t.Fatalf("expected 1 match, got %d", len(m.visibleRows))
	}
	if got := m.visibleRows[0].Node.Issue.ID; got != "ab-123" {
		t.Fatalf("expected ID ab-123, got %s", got)
	}
}

func TestRecalcVisibleRowsMatchesPartialIDs(t *testing.T) {
	nodes := []*graph.Node{
		{Issue: beads.FullIssue{ID: "ab-123", Title: "Alpha"}},
		{Issue: beads.FullIssue{ID: "ab-456", Title: "Beta"}},
	}
	m := App{
		roots:      nodes,
		filterText: "456",
	}
	m.recalcVisibleRows()

	if len(m.visibleRows) != 1 {
		t.Fatalf("expected 1 match, got %d", len(m.visibleRows))
	}
	if got := m.visibleRows[0].Node.Issue.ID; got != "ab-456" {
		t.Fatalf("expected ID ab-456, got %s", got)
	}
}

func TestNewAppWithMockClientLoadsIssues(t *testing.T) {
	t.Parallel()
	fixture := loadFixtureIssues(t, "issues_basic.json")
	mock := beads.NewMockClient()
	mock.ExportFn = func(ctx context.Context) ([]beads.FullIssue, error) {
		return fixture, nil
	}
	mock.CommentsFn = func(ctx context.Context, issueID string) ([]beads.Comment, error) {
		return []beads.Comment{
			{ID: 1, IssueID: issueID, Author: "tester", Text: "hello", CreatedAt: time.Now().UTC().Format(time.RFC3339)},
		}, nil
	}

	dbFile := createTempDBFile(t)
	app, err := NewApp(Config{
		RefreshInterval: time.Second,
		AutoRefresh:     false,
		DBPathOverride:  dbFile,
		Client:          mock,
	})
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	if len(app.roots) != 1 {
		t.Fatalf("expected a single root, got %d", len(app.roots))
	}
	// Comments are now loaded in background after TUI starts (ab-fkyz)
	// so we don't check CommentsCallCount here - see TestPreloadAllComments
}

func TestAppRefreshWithMockClient(t *testing.T) {
	fixtureInitial := loadFixtureIssues(t, "issues_basic.json")
	fixtureUpdated := loadFixtureIssues(t, "issues_refresh.json")
	mock := beads.NewMockClient()
	var exportCalls int
	mock.ExportFn = func(ctx context.Context) ([]beads.FullIssue, error) {
		exportCalls++
		if exportCalls == 1 {
			return fixtureInitial, nil
		}
		return fixtureUpdated, nil
	}
	mock.CommentsFn = func(ctx context.Context, issueID string) ([]beads.Comment, error) {
		return nil, nil
	}

	app := mustNewTestApp(t, mock)
	cmd := app.forceRefresh()
	refreshMsg := extractRefreshMsg(t, cmd)
	app.Update(refreshMsg)

	if got := app.roots[0].Issue.Title; got != "Root Epic Updated" {
		t.Fatalf("expected updated root title, got %s", got)
	}
}

func TestNewAppCapturesClientError(t *testing.T) {
	mock := beads.NewMockClient()
	mock.ExportFn = func(ctx context.Context) ([]beads.FullIssue, error) {
		return nil, errors.New("boom")
	}
	dbFile := createTempDBFile(t)
	_, err := NewApp(Config{
		RefreshInterval: time.Second,
		DBPathOverride:  dbFile,
		Client:          mock,
	})
	if err == nil {
		t.Fatalf("expected error when client export fails")
	}
}

func TestNewAppSucceedsWithEmptyDatabase(t *testing.T) {
	mock := beads.NewMockClient()
	mock.ExportFn = func(ctx context.Context) ([]beads.FullIssue, error) {
		return []beads.FullIssue{}, nil
	}
	dbFile := createTempDBFile(t)
	app, err := NewApp(Config{
		DBPathOverride: dbFile,
		Client:         mock,
	})
	if err != nil {
		t.Fatalf("expected no error for empty database, got %v", err)
	}
	if len(app.roots) != 0 {
		t.Fatalf("expected empty roots, got %d", len(app.roots))
	}
}

func TestCheckDBForChangesDetectsModification(t *testing.T) {
	dbFile := createTempDBFile(t)
	app := &App{
		client:        beads.NewMockClient(),
		dbPath:        dbFile,
		lastDBModTime: fileModTime(t, dbFile),
	}

	if cmd := app.checkDBForChanges(); cmd != nil {
		t.Fatalf("expected no refresh when mod time unchanged")
	}
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(dbFile, []byte("update"), 0o644); err != nil {
		t.Fatalf("write db: %v", err)
	}
	if cmd := app.checkDBForChanges(); cmd == nil {
		t.Fatalf("expected refresh command after db modification")
	}
}

func TestCheckDBForChangesDetectsWalModification(t *testing.T) {
	dbFile := createTempDBFile(t)
	app := &App{
		client:        beads.NewMockClient(),
		dbPath:        dbFile,
		lastDBModTime: fileModTime(t, dbFile),
	}

	// Update WAL file without touching the main database file.
	time.Sleep(10 * time.Millisecond)
	walPath := dbFile + "-wal"
	if err := os.WriteFile(walPath, []byte("wal update"), 0o644); err != nil {
		t.Fatalf("write wal: %v", err)
	}

	if cmd := app.checkDBForChanges(); cmd == nil {
		t.Fatalf("expected refresh command after wal modification")
	}
}

func TestRefreshHandlesClientError(t *testing.T) {
	fixtureInitial := loadFixtureIssues(t, "issues_basic.json")
	mock := beads.NewMockClient()
	var exportCalls int
	mock.ExportFn = func(ctx context.Context) ([]beads.FullIssue, error) {
		exportCalls++
		if exportCalls == 1 {
			return fixtureInitial, nil
		}
		return nil, errors.New("export failed")
	}
	mock.CommentsFn = func(ctx context.Context, issueID string) ([]beads.Comment, error) { return nil, nil }

	app := mustNewTestApp(t, mock)
	cmd := app.forceRefresh()
	refreshMsg := extractRefreshMsg(t, cmd)
	app.Update(refreshMsg)
	// Errors are now stored in lastError, not lastRefreshStats
	if app.lastError == "" {
		t.Fatalf("expected error to be stored in lastError")
	}
	if !strings.Contains(app.lastError, "export failed") {
		t.Fatalf("expected error message to contain 'export failed', got %s", app.lastError)
	}
}

func TestErrorToastShowsOnFirstError(t *testing.T) {
	fixtureInitial := loadFixtureIssues(t, "issues_basic.json")
	mock := beads.NewMockClient()
	var exportCalls int
	mock.ExportFn = func(ctx context.Context) ([]beads.FullIssue, error) {
		exportCalls++
		if exportCalls == 1 {
			return fixtureInitial, nil
		}
		return nil, errors.New("connection failed")
	}
	mock.CommentsFn = func(ctx context.Context, issueID string) ([]beads.Comment, error) { return nil, nil }

	app := mustNewTestApp(t, mock)

	// Trigger refresh error
	cmd := app.forceRefresh()
	refreshMsg := extractRefreshMsg(t, cmd)
	_, nextCmd := app.Update(refreshMsg)

	// Toast should be shown on first error
	if !app.showErrorToast {
		t.Error("expected showErrorToast to be true on first error")
	}
	if !app.errorShownOnce {
		t.Error("expected errorShownOnce to be true")
	}
	if nextCmd == nil {
		t.Error("expected tick command to be returned for toast countdown")
	}
}

func TestErrorToastEscDismisses(t *testing.T) {
	app := &App{
		lastError:      "test error",
		showErrorToast: true,
		errorShownOnce: true,
		ready:          true,
		keys:           DefaultKeyMap(),
	}

	// Press ESC
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if app.showErrorToast {
		t.Error("expected showErrorToast to be false after ESC")
	}
	// Error should still be stored
	if app.lastError == "" {
		t.Error("expected lastError to remain after dismissing toast")
	}
}

func TestErrorToastEKeyRecalls(t *testing.T) {
	app := &App{
		lastError:      "test error",
		showErrorToast: false,
		errorShownOnce: true,
		ready:          true,
		keys:           DefaultKeyMap(),
	}

	// Press '!' (error key)
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}})

	if !app.showErrorToast {
		t.Error("expected showErrorToast to be true after pressing '!'")
	}
	if cmd == nil {
		t.Error("expected tick command to be returned")
	}
}

func TestErrorToastDoesNotReappearOnSubsequentErrors(t *testing.T) {
	app := &App{
		lastError:      "first error",
		showErrorToast: false,
		errorShownOnce: true, // Already shown once
		ready:          true,
	}

	// Simulate another refresh error
	msg := refreshCompleteMsg{err: errors.New("second error")}
	_, cmd := app.Update(msg)

	// Toast should NOT show again
	if app.showErrorToast {
		t.Error("expected showErrorToast to remain false for subsequent errors")
	}
	if cmd != nil {
		t.Error("expected no tick command since toast is not shown")
	}
	// But error should be updated
	if !strings.Contains(app.lastError, "second error") {
		t.Errorf("expected lastError to be updated, got %s", app.lastError)
	}
}

func TestRefreshSuccessClearsOnlyRefreshErrors(t *testing.T) {
	t.Run("clearsRefreshErrors", func(t *testing.T) {
		app := &App{
			lastError:       "previous error",
			lastErrorSource: errorSourceRefresh,
			showErrorToast:  true,
			errorShownOnce:  true,
			ready:           true,
		}

		// Simulate successful refresh
		msg := refreshCompleteMsg{
			roots:  []*graph.Node{},
			digest: map[string]string{},
		}
		app.Update(msg)

		if app.lastError != "" {
			t.Errorf("expected lastError to be cleared, got %s", app.lastError)
		}
		if app.lastErrorSource != errorSourceNone {
			t.Errorf("expected lastErrorSource to be reset, got %v", app.lastErrorSource)
		}
		if app.showErrorToast {
			t.Error("expected showErrorToast to be false")
		}
		if app.errorShownOnce {
			t.Error("expected errorShownOnce to be reset")
		}
	})

	t.Run("preservesOperationErrors", func(t *testing.T) {
		app := &App{
			lastError:       "dependency add failed",
			lastErrorSource: errorSourceOperation,
			showErrorToast:  true,
			errorShownOnce:  true,
			ready:           true,
		}

		msg := refreshCompleteMsg{
			roots:  []*graph.Node{},
			digest: map[string]string{},
		}
		app.Update(msg)

		if app.lastError != "dependency add failed" {
			t.Errorf("expected lastError to be preserved, got %s", app.lastError)
		}
		if app.lastErrorSource != errorSourceOperation {
			t.Errorf("expected lastErrorSource to remain operation, got %v", app.lastErrorSource)
		}
		if !app.showErrorToast {
			t.Error("expected showErrorToast to remain visible")
		}
	})
}

func TestErrorToastRecallWhileOverlayActiveWhenNotTyping(t *testing.T) {
	overlay := NewCreateOverlay(CreateOverlayOptions{})
	overlay.focus = FocusType // not a text input

	app := &App{
		lastError:      "backend failed",
		showErrorToast: false,
		errorShownOnce: true,
		ready:          true,
		keys:           DefaultKeyMap(),
		createOverlay:  overlay,
		activeOverlay:  OverlayCreate,
	}

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}})

	if !app.showErrorToast {
		t.Fatal("expected showErrorToast to be true when overlay not in text input")
	}
	if cmd == nil {
		t.Fatal("expected tick command to be returned")
	}
}

func TestErrorToastRecallBlockedWhileTypingInOverlay(t *testing.T) {
	overlay := NewCreateOverlay(CreateOverlayOptions{})
	overlay.focus = FocusTitle

	app := &App{
		lastError:      "backend failed",
		showErrorToast: false,
		errorShownOnce: true,
		ready:          true,
		keys:           DefaultKeyMap(),
		createOverlay:  overlay,
		activeOverlay:  OverlayCreate,
	}

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}})

	if app.showErrorToast {
		t.Fatal("expected showErrorToast to remain false while typing")
	}
	_ = cmd // overlay may return its own command; the toast should remain hidden
}

func TestErrorToastCountdown(t *testing.T) {
	app := &App{
		lastError:       "test error",
		showErrorToast:  true,
		errorToastStart: time.Now().Add(-5 * time.Second), // Started 5 seconds ago
		ready:           true,
	}

	// Process tick - should continue countdown
	_, cmd := app.Update(errorToastTickMsg{})
	if !app.showErrorToast {
		t.Error("toast should still be visible before 10 seconds")
	}
	if cmd == nil {
		t.Error("expected another tick to be scheduled")
	}

	// Simulate 10+ seconds elapsed
	app.errorToastStart = time.Now().Add(-11 * time.Second)
	_, cmd = app.Update(errorToastTickMsg{})
	if app.showErrorToast {
		t.Error("toast should auto-dismiss after 10 seconds")
	}
}

func TestExtractShortError(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "extracts from Error: prefix",
			input:    "list issues: run bd list: Error: Database out of sync with JSONL. Run 'bd sync' to fix.",
			maxLen:   80,
			expected: "Database out of sync with JSONL",
		},
		{
			name:     "removes Run suggestion",
			input:    "Error: Something failed. Run 'bd fix' to resolve.",
			maxLen:   80,
			expected: "Something failed",
		},
		{
			name:     "truncates long message",
			input:    "Error: This is a very long error message that exceeds the maximum length allowed",
			maxLen:   30,
			expected: "This is a very long error m...",
		},
		{
			name:     "handles multiline error",
			input:    "Error: First line of error\nSecond line with more details",
			maxLen:   80,
			expected: "First line of error",
		},
		{
			name:     "handles simple error without prefix",
			input:    "connection refused",
			maxLen:   80,
			expected: "connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractShortError(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("extractShortError(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestRefreshUpdatesLastDBModTimeToLatest(t *testing.T) {
	tmp := t.TempDir()
	db := filepath.Join(tmp, "beads.db")
	f, err := os.Create(db)
	if err != nil {
		t.Fatalf("create db: %v", err)
	}
	_ = f.Close()

	// Set an initial mod time
	initial := time.Now().Add(-2 * time.Minute)
	if err := os.Chtimes(db, initial, initial); err != nil {
		t.Fatalf("set initial mod time: %v", err)
	}

	root := &graph.Node{Issue: beads.FullIssue{ID: "ab-123", Title: "Root"}}
	app := &App{
		roots:           []*graph.Node{root},
		visibleRows:     nodesToRows(root),
		dbPath:          db,
		lastDBModTime:   initial,
		autoRefresh:     true,
		refreshInFlight: true,
	}

	// Simulate bd export touching the DB by bumping mod time after the refresh started.
	latest := initial.Add(90 * time.Second)
	if err := os.Chtimes(db, latest, latest); err != nil {
		t.Fatalf("bump mod time: %v", err)
	}

	msg := refreshCompleteMsg{
		roots:     []*graph.Node{root},
		digest:    map[string]string{"ab-123": "x"},
		dbModTime: initial, // pre-refresh timestamp
	}

	model, _ := app.Update(msg)
	updated := model.(*App)

	if !updated.lastDBModTime.Equal(latest) {
		t.Fatalf("expected lastDBModTime to update to latest mod time; got %v want %v", updated.lastDBModTime, latest)
	}
	if updated.refreshInFlight {
		t.Fatal("expected refreshInFlight to be false after refreshComplete")
	}
}
