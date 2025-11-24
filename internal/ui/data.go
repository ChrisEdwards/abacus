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

func preloadAllComments(ctx context.Context, client beads.Client, roots []*graph.Node) {
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

	var wg sync.WaitGroup
	var mu sync.Mutex

	for issueID, node := range nodeMap {
		wg.Add(1)
		go func(id string, n *graph.Node) {
			defer wg.Done()

			comments, err := client.Comments(ctx, id)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				n.CommentError = fmt.Sprintf("failed: %v", err)
				return
			}

			if comments == nil {
				comments = []beads.Comment{}
			}
			n.Issue.Comments = comments
			n.CommentsLoaded = true
		}(issueID, node)
	}

	wg.Wait()
}

func loadData(ctx context.Context, client beads.Client, reporter StartupReporter) ([]*graph.Node, error) {
	if reporter != nil {
		reporter.Stage(StartupStageLoadingIssues, "Loading issues...")
	}
	fullIssues, totalIssues, err := fetchFullIssues(ctx, client)
	if err != nil {
		return nil, err
	}
	if len(fullIssues) == 0 {
		return nil, ErrNoIssues
	}
	if reporter != nil {
		reporter.Stage(StartupStageLoadingIssues, fmt.Sprintf("Loaded %d issues", totalIssues))
		reporter.Stage(StartupStageBuildingGraph, "Building dependency graph...")
	}

	roots, err := graph.NewBuilder().Build(fullIssues)
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
	if reporter != nil {
		reporter.Stage(StartupStageOrganizingTree, "Organizing tree...")
	}

	preloadAllComments(ctx, client, roots)
	return roots, nil
}

func fetchFullIssues(ctx context.Context, client beads.Client) ([]beads.FullIssue, int, error) {
	liteIssues, err := client.List(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list issues: %w", err)
	}
	if len(liteIssues) == 0 {
		return []beads.FullIssue{}, 0, nil
	}

	ids := make([]string, 0, len(liteIssues))
	for _, l := range liteIssues {
		ids = append(ids, l.ID)
	}

	full, err := batchFetchIssues(ctx, client, ids)
	return full, len(liteIssues), err
}

func batchFetchIssues(ctx context.Context, client beads.Client, ids []string) ([]beads.FullIssue, error) {
	var results []beads.FullIssue
	chunkSize := 20
	for i := 0; i < len(ids); i += chunkSize {
		end := i + chunkSize
		if end > len(ids) {
			end = len(ids)
		}

		batch, err := client.Show(ctx, ids[i:end])
		if err != nil {
			return nil, fmt.Errorf("show issues %v: %w", ids[i:end], err)
		}
		results = append(results, batch...)
	}
	return results, nil
}

// OutputIssuesJSON writes all issues to stdout as formatted JSON.
func OutputIssuesJSON(ctx context.Context, client beads.Client) error {
	issues, _, err := fetchFullIssues(ctx, client)
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
