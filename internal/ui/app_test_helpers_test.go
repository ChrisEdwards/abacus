package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"abacus/internal/beads"
	"abacus/internal/graph"

	tea "github.com/charmbracelet/bubbletea"
)

func nodesToRows(nodes ...*graph.Node) []graph.TreeRow {
	rows := make([]graph.TreeRow, len(nodes))
	for i, n := range nodes {
		rows[i] = graph.TreeRow{Node: n, Depth: 0}
	}
	return rows
}

func nodeToRow(node *graph.Node) graph.TreeRow {
	return graph.TreeRow{Node: node, Depth: 0}
}

func buildTreeTestApp(nodes ...*graph.Node) *App {
	m := &App{
		roots:  nodes,
		width:  120,
		height: 40,
	}
	m.recalcVisibleRows()
	return m
}

func buildWrappedTreeApp(count int) *App {
	nodes := make([]*graph.Node, count)
	for i := 0; i < count; i++ {
		nodes[i] = &graph.Node{
			Issue: beads.FullIssue{
				ID:     fmt.Sprintf("ab-%02d", i+1),
				Title:  "Ensure selection stays visible even when this title wraps onto multiple lines within the viewport.",
				Status: "open",
			},
		}
	}
	app := &App{
		roots:  nodes,
		width:  50,
		height: 12,
	}
	app.recalcVisibleRows()
	return app
}

func treeLineContaining(t *testing.T, view, id string) string {
	t.Helper()
	clean := stripANSI(view)
	for _, line := range strings.Split(clean, "\n") {
		if strings.Contains(line, id) {
			return strings.TrimRight(line, " ")
		}
	}
	t.Fatalf("tree output missing %s:\n%s", id, clean)
	return ""
}

func loadFixtureIssues(t *testing.T, file string) []beads.FullIssue {
	t.Helper()
	candidates := []string{
		filepath.Join("testdata", file),
		filepath.Join("..", "..", "testdata", file),
	}
	var data []byte
	var err error
	for _, path := range candidates {
		data, err = os.ReadFile(path) //nolint:gosec // fixtures defined under repo
		if err == nil {
			break
		}
	}
	if err != nil {
		t.Fatalf("read fixture %s: %v", file, err)
	}
	var issues []beads.FullIssue
	if err := json.Unmarshal(data, &issues); err != nil {
		t.Fatalf("unmarshal fixture %s: %v", file, err)
	}
	return issues
}

func filterIssuesByID(issues []beads.FullIssue, ids []string) []beads.FullIssue {
	set := make(map[string]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	var filtered []beads.FullIssue
	for _, iss := range issues {
		if set[iss.ID] {
			filtered = append(filtered, iss)
		}
	}
	return filtered
}

func liteIssuesFromFixture(issues []beads.FullIssue) []beads.LiteIssue {
	results := make([]beads.LiteIssue, len(issues))
	for i, iss := range issues {
		results[i] = beads.LiteIssue{ID: iss.ID}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].ID < results[j].ID
	})
	return results
}

func createTempDBFile(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "beads.db")
	if err := os.WriteFile(path, []byte("db"), 0o600); err != nil {
		t.Fatalf("write temp db: %v", err)
	}
	return path
}

func mustNewTestApp(t *testing.T, client beads.Client) *App {
	t.Helper()
	app, err := NewApp(Config{
		RefreshInterval: time.Second,
		AutoRefresh:     false,
		Client:          client,
	})
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}
	return app
}

func extractRefreshMsg(t *testing.T, cmd tea.Cmd) refreshCompleteMsg {
	t.Helper()
	if cmd == nil {
		t.Fatalf("expected refresh cmd, got nil")
	}
	msg := cmd()
	if refreshMsg, ok := msg.(refreshCompleteMsg); ok {
		return refreshMsg
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			if c != nil {
				result := c()
				if refreshMsg, ok := result.(refreshCompleteMsg); ok {
					return refreshMsg
				}
			}
		}
	}
	t.Fatalf("could not find refreshCompleteMsg in %T", msg)
	return refreshCompleteMsg{}
}

func fileModTime(t *testing.T, path string) time.Time {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	return info.ModTime()
}

func changeWorkingDir(t *testing.T, dir string) func() {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	return func() {
		if err := os.Chdir(orig); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}
}

func normalizePath(t *testing.T, path string) string {
	t.Helper()
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}
	return abs
}
