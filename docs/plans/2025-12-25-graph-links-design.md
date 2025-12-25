# Graph Links Support Design

**Bead:** ab-qdh0
**Date:** 2025-12-25
**Status:** Approved

## Summary

Add support for beads graph link types: `relates-to`, `duplicates`, and `supersedes`.

## Scope

| Type | Storage | Abacus Support |
|------|---------|----------------|
| `relates-to` | Dependency table (bidirectional) | Treat same as `related` |
| `duplicates` | Issue field `duplicate_of` | New node field + detail display |
| `supersedes` | Issue field `superseded_by` | New node field + detail display |

**Out of scope:** `replies-to`, `conditional-blocks`, `waits-for`

## Implementation

### 1. Types: Add Issue Fields

**File:** `internal/beads/types.go`

```go
type FullIssue struct {
    // ... existing fields ...

    // Graph link fields (issue-level, not dependencies)
    DuplicateOf  string `json:"duplicate_of,omitempty"`
    SupersededBy string `json:"superseded_by,omitempty"`
}
```

### 2. Node: Add Relationship Pointers

**File:** `internal/graph/node.go`

```go
type Node struct {
    // ... existing fields ...

    // Graph link metadata (closed issue context)
    DuplicateOf  *Node  // This issue is a duplicate of another
    SupersededBy *Node  // This issue was replaced by a newer version
}
```

### 3. Builder: Handle `relates-to` Dependency Type

**File:** `internal/graph/builder.go`

Update the dependency switch to handle both old and new related types:

```go
case "related", "relates-to":
    if related, ok := nodeMap[dep.TargetID]; ok {
        // Check if already linked (relates-to stores both directions)
        alreadyLinked := false
        for _, r := range node.Related {
            if r.Issue.ID == related.Issue.ID {
                alreadyLinked = true
                break
            }
        }
        if !alreadyLinked {
            node.Related = append(node.Related, related)
            related.Related = append(related.Related, node)
        }
    }
```

### 4. Builder: Resolve Issue Field References

**File:** `internal/graph/builder.go`

After building nodeMap, resolve `duplicate_of` and `superseded_by`:

```go
// Resolve duplicate_of and superseded_by references
for _, node := range nodeMap {
    if node.Issue.DuplicateOf != "" {
        if canonical, ok := nodeMap[node.Issue.DuplicateOf]; ok {
            node.DuplicateOf = canonical
        }
    }
    if node.Issue.SupersededBy != "" {
        if replacement, ok := nodeMap[node.Issue.SupersededBy]; ok {
            node.SupersededBy = replacement
        }
    }
}
```

### 5. Detail View: Display New Relationships

**File:** `internal/ui/detail.go`

Add after "Discovered While Working On" section:

```go
// Duplicate Of - this issue is a duplicate of another (canonical)
if node.DuplicateOf != nil {
    if section := renderRelSection("Duplicate Of: (1)", []*graph.Node{node.DuplicateOf}); section != "" {
        relSections = append(relSections, section)
    }
}
// Superseded By - this issue was replaced by a newer version
if node.SupersededBy != nil {
    if section := renderRelSection("Superseded By: (1)", []*graph.Node{node.SupersededBy}); section != "" {
        relSections = append(relSections, section)
    }
}
```

## Testing

1. **relates-to**: Create two issues, run `bd relate`, verify both appear in each other's "See Also" section
2. **duplicates**: Run `bd duplicate X --of Y`, verify X shows "Duplicate Of: Y" in detail view
3. **supersedes**: Run `bd supersede X --with Y`, verify X shows "Superseded By: Y" in detail view

## Backward Compatibility (beads v0.0.30)

Abacus must work with both beads v0.0.30 (no graph link columns) and v0.0.31+ (has columns).

### Problem

The SQLite client queries the `issues` table directly. If we add `duplicate_of, superseded_by` to the SELECT and those columns don't exist, the query fails.

### Solution: Lazy Schema Detection

**File:** `internal/beads/sqlite_client.go`

```go
type sqliteClient struct {
    // ... existing fields ...

    // Schema detection (lazy, cached)
    schemaOnce       sync.Once
    hasGraphLinkCols bool
}

func (c *sqliteClient) detectSchema(ctx context.Context, db *sql.DB) {
    c.schemaOnce.Do(func() {
        rows, _ := db.QueryContext(ctx, `PRAGMA table_info(issues)`)
        if rows == nil {
            return
        }
        defer rows.Close()

        for rows.Next() {
            var cid int
            var name, colType string
            var notNull, pk int
            var dfltValue sql.NullString
            rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk)
            if name == "duplicate_of" || name == "superseded_by" {
                c.hasGraphLinkCols = true
                return
            }
        }
    })
}
```

In `loadIssues()`, conditionally include columns:

```go
c.detectSchema(ctx, db)

cols := `id, title, ...`
if c.hasGraphLinkCols {
    cols += `, COALESCE(duplicate_of, ''), COALESCE(superseded_by, '')`
}
```

### Why This Approach?

- **Lazy**: No cost if Export() never called
- **Cached**: sync.Once ensures detection runs exactly once
- **Safe**: Missing columns → empty strings, no errors
- **No version checks**: Works regardless of beads version string

## Notes

- `related` vs `relates-to`: Both map to `node.Related`. The old `related` is one-way (manual), the new `relates-to` is bidirectional (auto). Dedup check prevents double-linking.
- `duplicate_of` and `superseded_by` are issue fields, not dependency table entries. The dependency type constants (`DepDuplicates`, `DepSupersedes`) exist in beads but aren't used by CLI commands.
