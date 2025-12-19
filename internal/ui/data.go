package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"

	"abacus/internal/beads"
	"abacus/internal/graph"
)

var ErrNoIssues = errors.New("no issues found in beads database")

const maxConcurrentCommentFetches = 8

func fetchCommentsForNode(ctx context.Context, client beads.Client, n *graph.Node) error {
	if n.CommentsLoaded {
		return nil
	}
	if client == nil {
		return fmt.Errorf("comments client unavailable")
	}

	comments, err := client.Comments(ctx, n.Issue.ID)
	if err != nil {
		return fmt.Errorf("fetch comments for %s: %w", n.Issue.ID, err)
	}
	n.Issue.Comments = comments
	n.CommentsLoaded = true
	n.CommentError = ""
	return nil
}

func preloadAllComments(ctx context.Context, client beads.Client, roots []*graph.Node, reporter StartupReporter) {
	if client == nil {
		return
	}
	nodeMap := make(map[string]*graph.Node)
	var collectNodes func([]*graph.Node)
	collectNodes = func(nodes []*graph.Node) {
		for _, n := range nodes {
			nodeMap[n.Issue.ID] = n
			collectNodes(n.Children)
		}
	}
	collectNodes(roots)

	total := len(nodeMap)
	if total == 0 {
		return
	}

	workerLimit := maxConcurrentCommentFetches
	if total < workerLimit {
		workerLimit = total
	}
	if workerLimit <= 0 {
		workerLimit = 1
	}

	sem := make(chan struct{}, workerLimit)

	var wg sync.WaitGroup
	var mu sync.Mutex
	completed := 0

	for issueID, node := range nodeMap {
		wg.Add(1)
		go func(id string, n *graph.Node) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			comments, err := client.Comments(ctx, id)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				n.CommentError = fmt.Sprintf("failed: %v", err)
			} else {
				if comments == nil {
					comments = []beads.Comment{}
				}
				n.Issue.Comments = comments
				n.CommentsLoaded = true
			}

			completed++
			if reporter != nil {
				reporter.Stage(StartupStageOrganizingTree, fmt.Sprintf("Loading comments... %d/%d", completed, total))
			}
		}(issueID, node)
	}

	wg.Wait()
}

func loadData(ctx context.Context, client beads.Client, reporter StartupReporter) ([]*graph.Node, error) {
	if reporter != nil {
		reporter.Stage(StartupStageLoadingIssues, "Loading issues...")
	}

	issues, err := client.Export(ctx)
	if err != nil {
		return nil, fmt.Errorf("export: %w", err)
	}
	if len(issues) == 0 {
		return nil, ErrNoIssues
	}

	if reporter != nil {
		reporter.Stage(StartupStageLoadingIssues, fmt.Sprintf("Loaded %d issues", len(issues)))
		reporter.Stage(StartupStageBuildingGraph, "Building dependency graph...")
	}

	roots, err := graph.NewBuilder().Build(issues)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(roots, func(i, j int) bool {
		a, b := roots[i], roots[j]
		rankA, rankB := 2, 2
		if a.HasInProgress {
			rankA = 0
		} else if a.HasReady {
			rankA = 1
		}
		if b.HasInProgress {
			rankB = 0
		} else if b.HasReady {
			rankB = 1
		}
		if rankA != rankB {
			return rankA < rankB
		}
		return a.Issue.CreatedAt < b.Issue.CreatedAt
	})
	// Comments are now loaded in background after TUI starts (ab-fkyz)
	// Lazy loading via fetchCommentsForNode() handles the detail view case
	return roots, nil
}

// OutputIssuesJSON writes all issues to stdout as formatted JSON.
func OutputIssuesJSON(ctx context.Context, client beads.Client) error {
	issues, err := client.Export(ctx)
	if err != nil {
		return err
	}
	if issues == nil {
		issues = []beads.FullIssue{}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(issues)
}

func buildIssueDigest(nodes []*graph.Node) map[string]string {
	digest := make(map[string]string)
	var walk func(nodes []*graph.Node)
	walk = func(nodes []*graph.Node) {
		for _, n := range nodes {
			key := fmt.Sprintf("%s|%s|%d|%s", n.Issue.Title, n.Issue.Status, n.Issue.Priority, n.Issue.UpdatedAt)
			digest[n.Issue.ID] = key
			walk(n.Children)
		}
	}
	walk(nodes)
	return digest
}

func computeDiffStats(oldIssues, newIssues map[string]string) string {
	if oldIssues == nil {
		oldIssues = map[string]string{}
	}
	if newIssues == nil {
		newIssues = map[string]string{}
	}

	added := 0
	removed := 0
	changed := 0

	for id, oldDigest := range oldIssues {
		newDigest, exists := newIssues[id]
		if !exists {
			removed++
			continue
		}
		if newDigest != oldDigest {
			changed++
		}
	}

	for id := range newIssues {
		if _, exists := oldIssues[id]; !exists {
			added++
		}
	}

	return fmt.Sprintf("+%d / Δ%d / -%d", added, changed, removed)
}
