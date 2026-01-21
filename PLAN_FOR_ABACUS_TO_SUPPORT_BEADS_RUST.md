# Plan: Abacus Support for beads_rust (br)

> **Project:** abacus
> **Goal:** Add support for beads_rust (`br`) CLI alongside existing beads (`bd`) support
> **Author:** Chris Edwards
> **Status:** Planning Phase
> **Created:** 2026-01-18

---

## Executive Summary

### The Problem

Abacus is a TUI for visualizing beads issues, currently tightly coupled to the Go-based `bd` CLI. The beads ecosystem is evolving:

- **beads (Go)** is moving toward Dolt and Gastown features, increasing complexity
- **beads_rust (br)** preserves the classic SQLite + JSONL architecture in a simpler, non-invasive design
- Users who adopt `br` cannot use abacus without modifications
- Maintaining two separate forks of abacus would duplicate 95%+ of the codebase

### The Solution

Add beads_rust support to abacus via a new CLI client implementation, with automatic backend detection. Both `bd` and `br` backends will be supported indefinitely (removed only if maintenance burden justifies).

| Component | Change Required |
|-----------|-----------------|
| **bd Backend** | Rename/refactor existing clients to `bdCLIClient` + `bdSQLiteClient` (frozen) |
| **br Backend** | New `brCLIClient` + `brSQLiteClient` (evolving, new features here) |
| **Detection** | Binary check + user prompt when ambiguous; stored preference |
| **Status Bar** | Show backend indicator (`bd` or `br`) |
| **Version Check** | Update for `br` version validation |

### Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| Single codebase, not fork | 95% of code is shared (TUI, graph, themes, overlays) |
| Separate clients per backend | bd has CLI+SQLite (frozen); br has CLI+SQLite (evolving). **Note:** SQLite schemas are currently compatible, so duplication is intentional preparation for future divergence—avoids conditionals when br evolves. |
| Binary check + user prompt | Check which binaries work; ask user when ambiguous |
| Store preference per-repo | Create `.abacus/config.yaml` in project if needed |
| Always show backend indicator | Transparency about which tool is active |
| Keep both backends indefinitely | Remove bd only when maintenance burden justifies |

### Code-Level Verification Summary (2025-01-19)

All claimed differences between `bd` and `br` have been verified at the source code level:

| Claimed Issue | Verification Result | Impact |
|---------------|---------------------|--------|
| `br` lacks `export` command | ✅ **Non-issue** - Abacus uses SQLite direct reads in production, CLI Export is dead code | None |
| `dep_type` vs `dependency_type` JSON | ✅ **Non-issue** - CLI Show() parsing is dead code in production; SQLite reads bypass this | None |
| `duplicate_of`/`superseded_by` columns | ✅ **Fixable** - Both bd and br use dependency types (`duplicates`, `supersedes`). **Decision: Migrate from column-based to dependency-based approach** | Minor refactor |
| CLI mutation commands differ | ⚠️ **One difference** - The `create` command uses `--title-flag` in br vs `--title` in bd (abacus currently uses `--title`). GitHub issue filed: [beads_rust#7](https://github.com/Dicklesworthstone/beads_rust/issues/7). **Decision: Use positional title syntax** which works for both. | Minor refactor in abacus |
| Schema incompatibility | ✅ **False** - All columns abacus reads exist in both schemas | None |
| Unknown statuses/types/dep-types | ⚠️ **Fixable** - br has `pinned` status, `conditional-blocks`/`waits-for` dep types that abacus doesn't handle. **Decision: Permissive reading (accept unknowns), restrictive creation (known options only).** See [Handling Unknown Values](#handling-unknown-values-forward-compatibility-strategy). | Minor refactor |

**Bottom line:** br support requires infrastructure changes (backend detection, version checking, binary name configuration) plus two minor refactors: (1) change the `create` command to use positional title syntax instead of `--title` flag, and (2) make status/type/dependency-type handling permissive for reading while restrictive for creation. GitHub issue [beads_rust#7](https://github.com/Dicklesworthstone/beads_rust/issues/7) was filed for the `--title` flag, but abacus will use positional syntax regardless.

---

## Table of Contents

1. [Background](#background)
2. [Goals and Non-Goals](#goals-and-non-goals)
3. [Research: Command Compatibility](#research-command-compatibility)
4. [Research: Schema Compatibility](#research-schema-compatibility)
5. [Research: JSON Output Compatibility](#research-json-output-compatibility)
6. [Configuration System (Research Summary)](#configuration-system-research-summary)
7. [Architecture](#architecture)
8. [Implementation Phases](#implementation-phases)
9. [Testing Strategy](#testing-strategy)
10. [Migration Path](#migration-path)
11. [Open Questions](#open-questions)
12. [Decision Log](#decision-log)
13. [References](#references)

---

## Background

### What is Abacus?

Abacus is a Go-based terminal UI (TUI) for visualizing and navigating beads issues. Key features:

- **Hierarchical tree view** of issues with parent-child relationships
- **Dependency visualization** showing blocking relationships
- **CRUD operations** via overlays (create, edit, delete, status change)
- **Search and filtering** with live updates
- **20+ themes** for visual customization

**Tech Stack:**
- Go 1.25+
- Bubble Tea (TUI framework)
- Lipgloss (styling)
- SQLite (direct database access)

### What is beads (bd)?

The original Go-based issue tracker by Steve Yegge:

- **Architecture:** SQLite + JSONL hybrid with daemon RPC
- **Features:** Full-featured including Gastown (agents, molecules, gates)
- **Git Integration:** Auto-commit, hooks, sync operations
- **Binary:** `bd`
- **Current Version:** 0.38.0 (frozen for abacus compatibility)

### What is beads_rust (br)?

A Rust port preserving the "classic" architecture:

- **Architecture:** SQLite + JSONL only, no daemon
- **Features:** Core issue tracking without Gastown complexity
- **Git Integration:** Never auto-commits, non-invasive by design
- **Binary:** `br`
- **JSONL Compatibility:** 100% compatible with `bd`

### Why Support beads_rust?

1. **Simpler architecture** - No daemon, no auto-git operations
2. **Performance** - Rust binary is faster and smaller
3. **Stability** - Frozen at the "classic" feature set
4. **Future direction** - `bd` is evolving away from the architecture abacus depends on

---

## Goals and Non-Goals

### Goals

1. **Support `br` as a first-class backend** - Full feature parity with `bd` support
2. **Smart backend detection** - Check binaries, prompt user when ambiguous, store preference
3. **Smooth migration** - Existing `bd` users continue working; easy switch to `br`
4. **Always show backend indicator** - User knows which tool is active
5. **Keep both backends** - No fixed deprecation timeline; remove when burden justifies

### Non-Goals

1. **Supporting beads versions > 0.38.0** - We freeze `bd` compatibility at 0.38.0
2. **Feature parity with future `br` features** - Only core issue tracking operations
3. **Backward compatibility with pre-Decision-004 bd databases** - Only current schemas supported
4. **Creating a separate fork** - Single codebase supports both
5. **Fixed deprecation timeline** - Remove bd only when maintenance justifies

### Design Principles

| Principle | Implementation |
|-----------|----------------|
| **Minimal changes** | New client implementation, minimal core changes |
| **Backend agnostic** | Core code doesn't care which CLI it's using |
| **Graceful degradation** | Clear error messages if neither backend available |
| **User choice when ambiguous** | Prompt user, default to br, store preference |
| **Transparency** | Always show which backend is active in status bar |
| **Project-level config** | Backend preference stored in `.abacus/config.yaml` per project |

---

## Research: Command Compatibility

This section documents the mapping between `bd` and `br` CLI commands.

### Core Commands Used by Abacus

| Operation | bd Command | br Command | Status |
|-----------|------------|------------|--------|
| List issues | `bd list --json` | `br list --json` | ✅ **IDENTICAL** (abacus only uses ID field) |
| Show issue | `bd show ID --json` | `br show ID --json` | ✅ **COMPATIBLE** |
| Create issue | `bd create "Title" --type Y --priority N --json` | `br create "Title" --type Y --priority N --json` | ⚠️ **REQUIRES CHANGE** - Positional title works for both. Abacus currently uses `--title` flag (bd-only; br uses `--title-flag`). **Decision: Change abacus to use positional syntax.** See [beads_rust#7](https://github.com/Dicklesworthstone/beads_rust/issues/7) |
| Update issue | `bd update ID --status X` | `br update ID --status X` | ✅ **IDENTICAL** |
| Close issue | `bd close ID` | `br close ID` | ✅ **IDENTICAL** |
| Reopen issue | `bd reopen ID` | `br reopen ID` | ✅ **IDENTICAL** |
| Add label | `bd label add ID LABEL` | `br label add ID LABEL` | ✅ **IDENTICAL** |
| Remove label | `bd label remove ID LABEL` | `br label remove ID LABEL` | ✅ **IDENTICAL** |
| Add dependency | `bd dep add FROM TO --type TYPE` | `br dep add FROM TO --type TYPE` | ✅ **IDENTICAL** |
| Remove dependency | `bd dep remove FROM TO` | `br dep remove FROM TO` | ✅ **IDENTICAL** |
| Delete issue | `bd delete ID --force` | `br delete ID --force` | ✅ **IDENTICAL** |
| Delete cascade | `bd delete ID --force --cascade` | `br delete ID --force --cascade` | ✅ **IDENTICAL** |
| Add comment | `bd comments add ID TEXT` | `br comments add ID "TEXT"` | ✅ **COMPATIBLE** (also supports `--message`) |
| List comments | `bd comments ID --json` | `br comments ID --json` | ✅ **IDENTICAL** |

**Note on `--json` flag:** Abacus has two create methods: `Create()` (returns just the ID by parsing text output) and `CreateFull()` (uses `--json` to get the full issue object back). The table above shows `--json` for completeness, but basic `Create()` does not use it.

### Specific Flag Compatibility

| bd Flag | Abacus Usage | br Equivalent | Status |
|---------|--------------|---------------|--------|
| `--db PATH` | Override database path | `--db PATH` | ✅ **IDENTICAL** (global flag) |
| `--json` | JSON output | `--json` | ✅ **IDENTICAL** (global flag) |
| `--force` | Force delete | `--force` | ✅ **EXISTS** |
| `--cascade` | Cascade delete | `--cascade` | ✅ **EXISTS** |
| `--set-labels` | Replace all labels | `--set-labels` | ✅ **EXISTS** (on update command) |
| `--labels` | Create with labels | `--labels` or `-l` | ✅ **EXISTS** (comma-separated) |
| `--description` | Description text | `--description` or `-d` | ✅ **EXISTS** |
| `--assignee` | Assign to user | `--assignee` or `-a` | ✅ **EXISTS** |
| `--parent` | Parent issue | `--parent` | ✅ **EXISTS** |

### Export Command

**Finding:** `br` does NOT have a separate `export` command. **This is a non-issue.**

**Why it doesn't matter:** Abacus **never calls `bd export` in production**. The architecture is:

| Operation Type | Implementation | What's Called |
|----------------|----------------|---------------|
| **Reads** (Export, List, Show, Comments) | `sqliteClient` methods | Direct SQLite queries |
| **Writes** (Create, Update, Delete, etc.) | Delegated to embedded `cliClient` | CLI commands (`bd`/`br`) |

The `sqliteClient.Export()` reads directly from SQLite:
```go
func (c *sqliteClient) Export(ctx context.Context) ([]FullIssue, error) {
    db, err := c.openDB(ctx)
    // ... queries issues, labels, dependencies, comments from SQLite
}
```

The CLI-based `Export()` in `cli_export.go` is **dead code** - it only exists for:
1. Unit tests (`cli_test.go`)
2. Fallback if `dbPath` is empty (never happens - `main.go` always provides a path)

**Implication:** Since the SQLite client handles all read operations and the schema is compatible (verified in Schema Compatibility section), `br` support requires **zero changes** to read operations. The CLI client only needs to support mutation commands.

### Key Differences Requiring Code Changes

**CLI Syntax:** One difference - the `create` command's `--title` flag is named `--title-flag` in br. Abacus currently uses `--title`. A fix was requested on br side: [beads_rust#7](https://github.com/Dicklesworthstone/beads_rust/issues/7). **Decision: Change abacus to use positional title syntax** which works for both bd and br, avoiding the need to wait for the br fix.

**JSON Output:** One minor difference in dependency field naming (see JSON section below).

---

## Research: Schema Compatibility

Abacus reads directly from the SQLite database for performance. Research confirms high compatibility.

### Tables Used by Abacus

| Table | Columns Used by Abacus | br Schema | Status |
|-------|------------------------|-----------|--------|
| `issues` | id, title, description, design, acceptance_criteria, notes, status, priority, issue_type, assignee, created_at, updated_at, closed_at, external_ref | ✅ **ALL PRESENT** + extras | **COMPATIBLE** |
| `labels` | issue_id, label | ✅ **IDENTICAL** | **COMPATIBLE** |
| `dependencies` | issue_id, depends_on_id, type | ✅ **IDENTICAL** | **COMPATIBLE** |
| `comments` | id, issue_id, author, text, created_at | ✅ **IDENTICAL** | **COMPATIBLE** |

### br Issues Table Schema (Full)

The `br` issues table has **43 columns** (superset of bd). Key columns for abacus:

```sql
-- Core fields (all present, compatible)
id TEXT PRIMARY KEY,
title TEXT NOT NULL,
description TEXT NOT NULL DEFAULT '',
design TEXT NOT NULL DEFAULT '',
acceptance_criteria TEXT NOT NULL DEFAULT '',
notes TEXT NOT NULL DEFAULT '',
status TEXT NOT NULL,
priority INTEGER NOT NULL,
issue_type TEXT NOT NULL,
assignee TEXT,
owner TEXT NOT NULL DEFAULT '',
created_at TEXT NOT NULL,
updated_at TEXT NOT NULL,
closed_at TEXT,
external_ref TEXT,

-- Soft delete fields (NEW in br, optional for abacus)
deleted_at TEXT,
deleted_by TEXT NOT NULL DEFAULT '',
delete_reason TEXT NOT NULL DEFAULT '',

-- Additional fields (abacus can ignore)
due_at TEXT,
defer_until TEXT,
estimated_minutes INTEGER,
-- ... and more
```

### Schema Compatibility Summary

| Question | Answer |
|----------|--------|
| Does `br` have `deleted_at` for soft deletes? | ✅ **YES** - `deleted_at`, `deleted_by`, `delete_reason` |
| Does `br` have `duplicate_of` / `superseded_by`? | ❌ **NO** - Handled via dependency types instead |
| Is status enum compatible? | ✅ **YES** - open, in_progress, blocked, deferred, closed, tombstone, pinned |
| Is issue_type enum compatible? | ✅ **YES** - task, bug, feature, epic, chore, docs, question |
| Is dependency type enum compatible? | ✅ **YES** - blocks, parent-child, related, etc. + extras |

### Status Values

| Status | bd | br | Notes |
|--------|----|----|-------|
| `open` | ✅ | ✅ | Default |
| `in_progress` | ✅ | ✅ | |
| `blocked` | ✅ | ✅ | |
| `deferred` | ✅ | ✅ | |
| `closed` | ✅ | ✅ | |
| `tombstone` | ✅ | ✅ | Soft deleted |
| `pinned` | ? | ✅ | May be br-only |

### Dependency Types

| Type | bd | br | Notes |
|------|----|----|-------|
| `blocks` | ✅ | ✅ | Default blocking |
| `parent-child` | ✅ | ✅ | Hierarchical |
| `related` | ✅ | ✅ | Informational |
| `relates-to` | ✅ | ✅ | Alias for `related` |
| `discovered-from` | ✅ | ✅ | |
| `conditional-blocks` | ? | ✅ | May be br-only |
| `waits-for` | ? | ✅ | May be br-only |
| `duplicates` | ✅ | ✅ | |
| `supersedes` | ✅ | ✅ | |

### Handling Unknown Values (Forward Compatibility Strategy)

**Problem:** br is evolving and may introduce new statuses, issue types, or dependency types that abacus doesn't know about. Currently, abacus has hardcoded enums that would reject or mishandle unknown values:

| Type | Current Abacus Behavior | Problem |
|------|------------------------|---------|
| Status | `ParseStatus()` returns error for unknown | Would reject valid "pinned" issues |
| Issue Types | `typeIndexFromString()` returns 0 (task) | "docs" issues silently become "task" |
| Dependency Types | Switch statement ignores unknown | `conditional-blocks` blockers not detected |

**Core Principle: Abacus is not the source of truth.**

The beads CLI (bd/br) owns the data model. Abacus is a viewer/editor that should:
1. **Never reject valid data** from the backend
2. **Never silently corrupt data** (e.g., converting "docs" type to "task")
3. **Preserve what it doesn't understand** (be a "good citizen")

**Decision: Separate read/display logic from create/modify logic.**

Abacus has two distinct roles with different requirements:

| Role | Requirement | Approach |
|------|-------------|----------|
| **Viewing data** | Maximally permissive | Accept any string, display as-is |
| **Creating data** | Can be restrictive | Only offer known options in UI |
| **Editing data** | Preserve unknown values | Show current value, don't silently convert |

**Concrete Strategy by Category:**

#### Status
- **Reading**: Accept any string from SQLite. No validation errors.
- **Display**: Show the actual status value, even if unknown (e.g., "pinned").
- **Filtering UI**: Show known statuses as options. Issues with unknown statuses still appear in unfiltered list.
- **Changing status**: Only offer known statuses as targets. Allow transitions FROM unknown statuses (user might want to close a "pinned" issue).

#### Issue Types
- **Reading**: Accept any string. Store as-is.
- **Display**: Show the actual type (not "task" for everything unknown).
- **Create UI**: Only offer known types in dropdown.
- **Edit UI**: If current type is unknown, show it (grayed out/disabled) and preserve unless user explicitly changes to a known type.

#### Dependency Types
- **Known blocking types**: `blocks`, `conditional-blocks`, `waits-for` - add explicit cases.
- **Known hierarchy types**: `parent-child` - used for tree structure, not blocking.
- **Known non-blocking types**: `related`, `relates-to`, `discovered-from`, `duplicates`, `supersedes`.
- **Unknown types**: Default to **non-blocking** with debug log.

**Why default unknown dependency types to non-blocking?**
- Most dependency types are informational (non-blocking is the common case)
- False "ready" is less disruptive than false "blocked"
- User can see the dependency in detail view and understand the relationship
- If br adds a new blocking type, we update abacus to handle it explicitly

**Implementation Changes Required:**

1. **`internal/domain/status.go`**:
   - `ParseStatus()` accepts unknown strings (returns the string, not error)
   - Add `IsKnown() bool` method for UI logic
   - Transition validation only for known target statuses

2. **`internal/ui/overlay_create.go`**:
   - When editing, if current type not in `typeOptions`, display it (disabled)
   - Don't default unknown types to index 0

3. **`internal/graph/builder.go`**:
   - Add cases for `conditional-blocks`, `waits-for` (blocking)
   - Add `default` case: log warning, treat as non-blocking

**Why This Approach?**

| Benefit | Explanation |
|---------|-------------|
| **Forward compatible** | br can add new types without breaking abacus |
| **Data preserving** | Unknown values pass through unchanged |
| **Good UX** | Users see actual data, can only create known types |
| **Minimal maintenance** | Only update abacus for types with special semantics |
| **Fail-safe** | Unknown blocking types might be missed, but data isn't corrupted |

**Trade-off acknowledged:** If br adds a new blocking dependency type and abacus isn't updated, issues might appear "ready" when they're actually blocked. This is acceptable because:
1. The dependency is still visible in the detail view
2. Data integrity is preserved
3. The fix is simple (add a case to the switch)
4. False "ready" is recoverable; data corruption is not

### SQLite Client Compatibility

**Finding:** The existing `sqliteClient` in abacus works with `br` databases with **NO changes needed**:

1. **All required columns present** - br has all columns abacus queries (verified against actual schemas)
2. **Extra columns are ignored** - SELECT queries only fetch needed columns
3. **No schema migration needed** - Both current schemas provide what abacus needs
4. **detectSchema() will be removed** - Legacy schema detection for pre-Decision-004 databases is being dropped

**Verified columns abacus reads (all present in both bd and br):**
- issues: id, title, description, design, acceptance_criteria, notes, status, priority, issue_type, assignee, created_at, updated_at, closed_at, external_ref, deleted_at
- labels: issue_id, label
- dependencies: issue_id, depends_on_id, type
- comments: id, issue_id, author, text, created_at

**Schema clarification (verified at code level):**

Both **current bd** (after Decision 004) and **br** use the same approach:
- `duplicate_of`/`superseded_by` are stored as **dependency types** (`duplicates`, `supersedes`) in the dependencies table
- Neither current schema has these as issue-level columns

**Source verification:**
- bd schema: `beads/internal/storage/sqlite/schema.go:37` - "NOTE: replies_to, relates_to, duplicate_of, superseded_by removed per Decision 004"
- br schema: `beads_rust/src/storage/schema.rs:12-50` - Never had these columns

**Decision (2025-01-19): Replace column-based with dependency-based graph links**

The `detectSchema()` function and related code (ab-qdh0) was built for bd v0.0.31 which had `duplicate_of`/`superseded_by` as issue columns. Both current bd (after Decision 004) and br store these as **dependency types** instead.

**Note:** Pre-Decision-004 bd databases (with `duplicate_of`/`superseded_by` columns) are **not supported**. Users with very old bd databases must upgrade to current bd before using this version of abacus.

**Solution:** Migrate from column-based to dependency-based approach:

| File | Change |
|------|--------|
| `internal/beads/types.go` | **Remove** `DuplicateOf`, `SupersededBy` fields (no longer issue columns) |
| `internal/beads/sqlite_client.go` | **Remove** `detectSchema()`, `hasGraphLinkCols`, `schemaOnce`, conditional queries |
| `internal/graph/node.go` | **Keep** `DuplicateOf`, `SupersededBy` *Node pointers (still needed for UI) |
| `internal/graph/builder.go` | **Change** resolution to use dependency types instead of issue fields |
| `internal/ui/detail.go` | **Keep** UI rendering (will work once builder is fixed) |
| Test files | **Update** to test dependency-based approach |

**Builder change (builder.go):**

Resolve `DuplicateOf` and `SupersededBy` node pointers using dependency types instead of issue columns:

- **`duplicates` dependency:** If issue A has a `duplicates` dependency pointing to B, then A is a duplicate of B. Set `A.DuplicateOf = B`.
- **`supersedes` dependency:** If issue A has a `supersedes` dependency pointing to B, then A supersedes (replaces) B. Set `B.SupersededBy = A` (note: the *target* gets the pointer, not the source).

Implementation will need to handle the `supersedes` case carefully since it sets a pointer on the target node rather than the current node being processed. A two-pass approach or reverse lookup may be needed.

This makes the feature work with both bd and br, using the dependency data that's already loaded.

---

## Research: JSON Output Compatibility

Abacus parses JSON output from CLI commands. Research confirms high compatibility with one key difference.

### br JSON Output Structure

**Field naming:** `snake_case` (same as bd)
**Date format:** ISO 8601 UTC (`YYYY-MM-DDTHH:MM:SSZ`) (same as bd)
**Null handling:** Optional fields are **omitted** when null/empty (same as bd - both use omitempty)

### br show --json Output

```json
[
  {
    "id": "bd-xyz",
    "title": "Issue title",
    "status": "open",
    "issue_type": "task",
    "priority": 2,
    "description": "...",
    "design": "...",
    "acceptance_criteria": "...",
    "notes": "...",
    "created_at": "2026-01-15T12:00:00Z",
    "updated_at": "2026-01-15T12:00:00Z",
    "assignee": "alice",
    "labels": ["backend", "high"],
    "dependencies": [
      {
        "id": "bd-abc",
        "title": "Blocking task",
        "status": "open",
        "priority": 0,
        "dep_type": "blocks"
      }
    ],
    "dependents": [...],
    "comments": [
      {
        "id": 1,
        "issue_id": "bd-xyz",
        "author": "username",
        "text": "comment body",
        "created_at": "2026-01-15T12:00:00Z"
      }
    ],
    "events": [...],
    "parent": "bd-parent-id"
  }
]
```

### JSON Compatibility Summary

| Question | Answer |
|----------|--------|
| Is `br show --json` structure similar? | ✅ **YES** - Same core fields |
| Are field names identical? | ⚠️ **MOSTLY** - One key difference below |
| Are date formats identical (ISO 8601)? | ✅ **YES** |
| How are null/empty values handled? | ✅ **IDENTICAL** - both use omitempty/skip_serializing_if |
| Does `br create --json` return created issue? | ✅ **YES** - Returns full issue with labels/deps |

### Key Difference: Dependency Field Names

| Context | bd Field | br Field | Impact |
|---------|----------|----------|--------|
| In `show` output dependencies array | `dependency_type` | `dep_type` | ⚠️ **Non-issue** (see below) |
| In full Dependency object | `type` | `type` | ✅ Compatible |

**Note:** An upcoming br release will rename `dep_type` to `dependency_type` for consistency with bd. This makes the field name difference a temporary non-issue that will be fully resolved.

**Verified at code level:**

**Go (bd)** - `beads/internal/types/types.go:374-377`:
```go
type IssueWithDependencyMetadata struct {
    Issue
    DependencyType DependencyType `json:"dependency_type"`
}
```

**Rust (br)** - `beads_rust/src/format/output.rs:131-138`:
```rust
pub struct IssueWithDependencyMetadata {
    pub id: String,
    pub title: String,
    pub status: Status,
    pub priority: Priority,
    pub dep_type: String,  // Different field name!
}
```

**Why this is a non-issue in production:**

Abacus's `types.go` expects `dependency_type` in the `Dependency` struct:
```go
type Dependency struct {
    TargetID string `json:"id"`
    Type     string `json:"dependency_type"`
}
```

However, **this JSON parsing is only used when CLI `Show()` is called**, which happens:
- In tests (`cli_test.go`)
- As fallback if SQLite client can't be initialized

**In production**, abacus uses `sqliteClient` which reads dependencies directly from the SQLite `dependencies` table (columns: `issue_id`, `depends_on_id`, `type`). The CLI `Show()` method is **dead code in production**.

**Conclusion:** No code change needed for `dep_type` vs `dependency_type` since production reads from SQLite, not CLI JSON.

### Comment JSON Structure

**br format:**
```json
{
  "id": 1,
  "issue_id": "bd-xyz",
  "author": "username",
  "text": "Comment text",
  "created_at": "2026-01-15T12:00:00Z"
}
```

✅ **COMPATIBLE** - Same structure as bd. Note: br uses `text` field (Comment.body is renamed to `text` via serde).

### br create --json Output

Returns full issue with relations:
```json
{
  "id": "bd-xyz",
  "title": "...",
  ... (all issue fields),
  "labels": ["label1"],
  "dependencies": [...]
}
```

✅ **COMPATIBLE** - Returns created issue, can parse ID and full details.

### Code Changes Required for JSON

| Change | Reason | Effort |
|--------|--------|--------|
| ~~Handle `dep_type` vs `dependency_type`~~ | ~~br uses `dep_type`, bd uses `dependency_type`~~ | **0 lines - Non-issue** |

**Verified:** No JSON parsing changes needed. Production uses SQLite direct reads, not CLI JSON output. The CLI `Show()` method that parses `dependency_type` is dead code in production.

---

## Configuration System (Research Summary)

### Existing Configuration Architecture

Abacus has a robust, layered configuration system (`internal/config/config.go`) that will be leveraged for backend selection:

**Precedence (lowest to highest):**
```
Defaults < User Config < Project Config < Environment Variables < CLI Flags
```

**Configuration Files:**
| Location | Purpose |
|----------|---------|
| `~/.abacus/config.yaml` | User-level defaults |
| `.abacus/config.yaml` | Project-level overrides (discovered by walking up directory tree) |

**Environment Variables:**
- Prefix: `AB_`
- Key transformation: `.` and `-` become `_`
- Example: `auto-refresh-seconds` → `AB_AUTO_REFRESH_SECONDS`
- **Exception:** `beads.backend` does NOT use env var (project config only)

**Existing Keys:**
| Key | Type | Default | Purpose |
|-----|------|---------|---------|
| `auto-refresh-seconds` | int | 10 | Auto-refresh interval |
| `skip-version-check` | bool | false | Skip beads CLI version validation |
| `database.path` | string | "" | Override path to beads database |
| `output.format` | string | "rich" | Detail pane markdown style |
| `theme` | string | "tokyonight" | UI theme name |
| `tree.showPriority` | bool | true | Show priority column |

**Key Features:**
- Thread-safe singleton with lazy initialization
- Automatic environment variable binding
- `SaveTheme()` pattern for persisting changes (will use same pattern for `SaveBackend()`)

### New Keys for Backend Support

```go
KeyBeadsBackend                  = "beads.backend"                        // "bd" or "br", empty means auto-detect
KeyBdUnsupportedVersionWarnShown = "beads.bd_unsupported_version_warned"  // true if user has seen the bd > 0.38.0 warning
```

**Important:** `KeyBeadsBackend` does NOT use environment variable binding. Backend selection reads from project config only (via `GetProjectString()`) to ensure each repository uses its correct backend. The `KeyBdUnsupportedVersionWarnShown` is stored in user config as a one-time flag.

**Config file example:**
```yaml
# .abacus/config.yaml (project-level) - backend selection
beads:
  backend: br  # or "bd"

# ~/.abacus/config.yaml (user-level) - one-time warning flag
beads:
  bd_unsupported_version_warned: true  # user has been notified about bd > 0.38.0
  # NOTE: backend is NOT stored at user level - it's always per-project
```

---

## Architecture

### Current Architecture (bd only)

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Abacus TUI                                  │
│  (Bubble Tea, Overlays, Tree View, Detail Pane, Themes)            │
└───────────────────────────────┬─────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      internal/beads/                                │
│                                                                     │
│   ┌───────────────┐     ┌─────────────────────────────────────┐    │
│   │    Client     │     │                                     │    │
│   │  (interface)  │◀────│  sqliteClient (reads SQLite,        │    │
│   │               │     │               delegates writes to   │    │
│   │  - List()     │     │               CLI)                  │    │
│   │  - Show()     │     │                                     │    │
│   │  - Create()   │     └─────────────────────────────────────┘    │
│   │  - Update()   │                    │                           │
│   │  - Close()    │                    │ writes                    │
│   │  - ...        │                    ▼                           │
│   └───────────────┘     ┌─────────────────────────────────────┐    │
│          ▲              │  cliClient (wraps bd CLI)           │    │
│          │              │  - bin: "bd"                        │    │
│          │              │  - run("list", "--json")            │    │
│          └──────────────│  - run("create", "--title", ...)    │    │
│                         └─────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
                    ┌───────────────────────┐
                    │   bd CLI / SQLite DB  │
                    └───────────────────────┘
```

### Proposed Architecture (bd + br)

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Abacus TUI                                  │
│  (Bubble Tea, Overlays, Tree View, Detail Pane, Themes)            │
└───────────────────────────────┬─────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      internal/beads/                                │
│                                                                     │
│   ┌───────────────┐                                                │
│   │    Client     │◀───────────────────────────────────────────┐   │
│   │  (interface)  │                                            │   │
│   └───────────────┘                                            │   │
│          ▲                                                     │   │
│          │                                                     │   │
│   ┌──────┴──────────────────────────────┐                     │   │
│   │                                      │                     │   │
│   ▼                                      ▼                     │   │
│ ┌────────────────────────────┐  ┌────────────────────────────┐│   │
│ │     bd Backend (FROZEN)    │  │    br Backend (EVOLVING)   ││   │
│ │                            │  │                            ││   │
│ │  ┌──────────────────────┐  │  │  ┌──────────────────────┐  ││   │
│ │  │    bdSQLiteClient    │  │  │  │    brSQLiteClient    │  ││   │
│ │  │  (reads from SQLite) │  │  │  │  (reads from SQLite) │──┼┘   │
│ │  └──────────┬───────────┘  │  │  └──────────┬───────────┘  │    │
│ │             │ writes       │  │             │ writes       │    │
│ │             ▼              │  │             ▼              │    │
│ │  ┌──────────────────────┐  │  │  ┌──────────────────────┐  │    │
│ │  │     bdCLIClient      │  │  │  │     brCLIClient      │  │    │
│ │  │     (bin: "bd")      │  │  │  │     (bin: "br")      │  │    │
│ │  └──────────┬───────────┘  │  │  └──────────┬───────────┘  │    │
│ └─────────────┼──────────────┘  └─────────────┼──────────────┘    │
│               │                               │                    │
└───────────────┼───────────────────────────────┼────────────────────┘
                │                               │
                ▼                               ▼
          ┌──────────┐                   ┌──────────┐
          │  bd CLI  │                   │  br CLI  │
          └──────────┘                   └──────────┘
```

**Key Architectural Principle:**

Each backend is a self-contained unit with its own CLI and SQLite clients:

| Backend | CLI Client | SQLite Client | Status |
|---------|------------|---------------|--------|
| **bd** | `bdCLIClient` | `bdSQLiteClient` | **FROZEN** at 0.38.0 - no new features |
| **br** | `brCLIClient` | `brSQLiteClient` | **EVOLVING** - new features added here |

This separation ensures:
- bd code can be locked in place and never touched
- New br features don't require conditionals or schema checks
- When bd support is eventually dropped, just delete the bd files
- Each backend can evolve independently
- **No if-statement complexity** - br code can be modified freely without tiptoeing around bd compatibility concerns

### Key Changes

1. **Interface Refactor** - Split `Client` into `Reader` + `Writer` interfaces (eliminates dead code)
2. **bd Backend (FROZEN)** - Rename existing clients to `bdCLIClient` + `bdSQLiteClient`
3. **br Backend (NEW)** - New `brCLIClient` + `brSQLiteClient` implementations
4. **Backend Detection** - Binary check + user prompt when ambiguous + stored preference
5. **Version Checking** - Separate version validation for `br`
6. **Status Bar Indicator** - Show `bd` or `br` to indicate active backend
7. **Config Management** - Load/save backend preference (project-level only)

### File Changes

| File | Change |
|------|--------|
| `internal/config/config.go` | Update - Add `KeyBeadsBackend` + `SaveBackend()` |
| `internal/beads/client.go` | **Refactor** - Split into `Reader`, `Writer`, `Client` interfaces |
| `internal/beads/bd_cli.go` | **RENAME** from `cli.go` - bdCLIClient (frozen) |
| `internal/beads/bd_sqlite.go` | **RENAME** from `sqlite_client.go` - bdSQLiteClient (frozen) |
| `internal/beads/br_cli.go` | **NEW** - brCLIClient implementation |
| `internal/beads/br_sqlite.go` | **NEW** - brSQLiteClient (see note below) |
| `internal/beads/backend.go` | **NEW** - Backend detection, binary check, user prompt, factory |
| `internal/beads/version_check.go` | Update - Add `br` version validation |
| `internal/tui/statusline.go` | Update - Show backend indicator (`bd` or `br`) |
| `cmd/abacus/main.go` | Update - Initialize with detected backend |

**Note on SQLite duplication:** `bd_sqlite.go` and `br_sqlite.go` will initially be very similar since schemas are compatible. This intentional duplication avoids runtime conditionals and allows br's SQLite client to evolve independently when br adds new features or schema changes. The slight code duplication is preferred over coupling the two backends.

---

## Implementation Phases

### Phase 1: Research and Validation ✅ COMPLETE

**Goal:** Confirm compatibility between `bd` and `br` command/output formats.

- [x] ~~Run `bd list --json` and `br list --json` on same database, compare output~~ - Compatible
- [x] ~~Run `bd show ID --json` and `br show ID --json`, compare output~~ - Compatible (dep_type diff)
- [x] ~~Test `br create` with all flags abacus uses, verify JSON response~~ - Identical to bd
- [x] ~~Test `br update`, `br close`, `br reopen` commands~~ - Identical
- [x] ~~Test `br label add/remove` commands~~ - Identical to bd
- [x] ~~Test `br dep add/remove` commands~~ - Identical
- [x] ~~Test `br delete` with `--force` flag~~ - `--cascade` also exists
- [x] ~~Test `br comments add` and `br comments ID --json`~~ - Compatible
- [x] ~~Compare SQLite schemas between `bd` and `br` databases~~ - br is superset
- [x] ~~Document all incompatibilities found~~ - See research sections above

**Deliverables:**
- ✅ Completed compatibility matrix (this document)
- ✅ List of code changes needed (see Implementation Findings Summary)

### Phase 2: Interface Refactor

**Goal:** Split the `Client` interface into `Reader` + `Writer` to eliminate dead code and clarify responsibilities.

**Current problem:** The `Client` interface includes both read and write methods, but:
- CLI clients only do writes (read methods are dead code)
- SQLite clients do reads and delegate writes to CLI

**New structure:**

```go
// internal/beads/client.go

// Reader handles all read operations (SQLite-only in production)
type Reader interface {
    List(ctx context.Context) ([]LiteIssue, error)
    Show(ctx context.Context, ids []string) ([]FullIssue, error)
    Export(ctx context.Context) ([]FullIssue, error)
    Comments(ctx context.Context, issueID string) ([]Comment, error)
}

// Writer handles all mutation operations (CLI-only in production)
type Writer interface {
    UpdateStatus(ctx context.Context, issueID, newStatus string) error
    Close(ctx context.Context, issueID string) error
    Reopen(ctx context.Context, issueID string) error
    AddLabel(ctx context.Context, issueID, label string) error
    RemoveLabel(ctx context.Context, issueID, label string) error
    Create(ctx context.Context, title, issueType string, priority int, labels []string, assignee string) (string, error)
    CreateFull(ctx context.Context, title, issueType string, priority int, labels []string, assignee, description, parentID string) (FullIssue, error)
    UpdateFull(ctx context.Context, issueID, title, issueType string, priority int, labels []string, assignee, description string) error
    AddDependency(ctx context.Context, fromID, toID, depType string) error
    RemoveDependency(ctx context.Context, fromID, toID, depType string) error
    Delete(ctx context.Context, issueID string, cascade bool) error
    AddComment(ctx context.Context, issueID, text string) error
}

// Client combines Reader and Writer for full functionality
type Client interface {
    Reader
    Writer
}
```

**Implementation mapping:**

| Component | Implements | Notes |
|-----------|------------|-------|
| `bdSQLiteClient` | `Client` (Reader + Writer) | Reads from SQLite, delegates writes to embedded Writer |
| `brSQLiteClient` | `Client` (Reader + Writer) | Reads from SQLite, delegates writes to embedded Writer |
| `bdCLIClient` | `Writer` only | No more dead read methods |
| `brCLIClient` | `Writer` only | Clean implementation, no Export() needed |

- [ ] Create `Reader` interface with read methods
- [ ] Create `Writer` interface with write methods
- [ ] Keep `Client` as composition of `Reader` + `Writer`
- [ ] Update `sqliteClient` to embed `Writer` instead of full `Client`
- [ ] Remove read methods from `cliClient` (dead code elimination)
- [ ] Update all call sites that use the interface
- [ ] Ensure all existing tests still pass

**Forward Compatibility Tasks** (see [Handling Unknown Values](#handling-unknown-values-forward-compatibility-strategy)):

- [ ] `internal/domain/status.go`: Make `ParseStatus()` accept unknown strings; add `IsKnown()` method
- [ ] `internal/ui/overlay_create.go`: Preserve unknown issue types when editing; don't default to index 0
- [ ] `internal/graph/builder.go`: Add cases for `conditional-blocks`, `waits-for`; add `default` case (non-blocking with log)
- [ ] Update status transition validation to only validate known target statuses
- [ ] Add tests for unknown value handling

**Deliverables:**
- Clean interface separation
- CLI clients only implement `Writer`
- SQLite clients implement full `Client`
- Dead code removed
- Forward compatibility for unknown statuses, types, and dependency types

### Phase 3: Refactor bd Backend (FROZEN)

**Goal:** Rename and isolate existing bd clients so they can be frozen.

- [ ] Rename `cli.go` → `bd_cli.go` (rename struct to `bdCLIClient`)
- [ ] Rename `sqlite_client.go` → `bd_sqlite.go` (rename struct to `bdSQLiteClient`)
- [ ] Update `bdCLIClient` to implement `Writer` only (remove read methods)
- [ ] Update `bdSQLiteClient` to embed `Writer` instead of full `Client`
- [ ] Update all references to use new names
- [ ] Add "FROZEN" comment header to both files
- [ ] Ensure all existing tests still pass

**Deliverables:**
- bd backend isolated in `bd_cli.go` + `bd_sqlite.go`
- Existing functionality preserved
- Code marked as frozen

### Phase 4: Backend Detection Infrastructure

**Goal:** Create backend detection and preference management.

- [ ] Add `KeyBeadsBackend = "beads.backend"` to `internal/config/config.go`
- [ ] Create `internal/beads/backend.go` with detection logic
- [ ] Implement `commandExists()` using `exec.LookPath`
- [ ] Implement `loadStoredPreference()` using `config.GetProjectString(config.KeyBeadsBackend)`
- [ ] Implement `SaveBackend()` following the existing `SaveTheme()` pattern
- [ ] Implement user prompt when both binaries on PATH (br pre-selected)
- [ ] Create `NewClientForBackend(backend string)` factory function
- [ ] Add backend indicator to status bar
- [ ] Update version checking for `br` minimum version

**Deliverables:**
- Backend detection working via PATH check + user prompt
- Stored preference in project config only (`.abacus/config.yaml`)
- Backend indicator visible in status bar

### Phase 5: br CLI Client

**Goal:** Implement brCLIClient (Writer interface only).

- [ ] Create `internal/beads/br_cli.go`
- [ ] Implement `Writer` interface methods only (no read methods needed)
- [ ] Use positional title syntax for `create` command (works for both bd and br)
- [ ] Add unit tests for brCLIClient

**Note on `--title` flag:** br uses `--title-flag` while bd uses `--title`. GitHub issue [beads_rust#7](https://github.com/Dicklesworthstone/beads_rust/issues/7) was filed, but abacus will use positional title syntax regardless, which works for both backends.

**Note:** With the interface refactor, `brCLIClient` only implements `Writer`. No Export(), List(), Show(), or Comments() methods needed since those are handled by the SQLite client.

**Deliverables:**
- `brCLIClient` passing all tests
- Integration tests with real `br` binary

### Phase 6: br SQLite Client

**Goal:** Implement brSQLiteClient (full `Client` interface) that can evolve with br features.

- [ ] Create `internal/beads/br_sqlite.go`
- [ ] Implement `Reader` interface (SQLite queries - copy from bd_sqlite.go)
- [ ] Embed `brCLIClient` as `Writer` for mutation operations
- [ ] Verify compatibility with br schema
- [ ] Add unit tests for brSQLiteClient

**Deliverables:**
- `brSQLiteClient` passing all tests
- Ready for future br-specific features

### Phase 7: Integration and Testing

**Goal:** Full end-to-end testing with both backends using tmux-based TUI testing.

- [ ] TUI testing via tmux: all overlays work with `br` backend
- [ ] TUI testing via tmux: tree view, detail pane, search work with `br`
- [ ] TUI testing via tmux: backend indicator displays correctly
- [ ] TUI testing via tmux: backend switching works when config changes
- [ ] Add integration tests that run with both backends
- [ ] Add CI matrix testing both backends
- [ ] Update documentation

**TUI Testing Approach:**
Use `scripts/tui-test.sh` for comprehensive UI verification:
```bash
make build                         # Build after code changes
./scripts/tui-test.sh start        # Launch abacus in tmux
./scripts/tui-test.sh keys 'jjjl'  # Navigate tree (j=down, l=expand)
./scripts/tui-test.sh enter        # Open detail pane
./scripts/tui-test.sh view         # Capture current state for verification
./scripts/tui-test.sh quit         # Clean up
```

**Deliverables:**
- All TUI tests passing (verified via tmux automation)
- CI pipeline testing both backends
- Updated README with backend selection docs

### Phase 8: Deprecation Notices (Future)

**Goal:** Begin `bd` deprecation.

- [ ] Add deprecation warning when using `bd` backend
- [ ] Update docs to recommend `br`
- [ ] Set timeline for `bd` removal
- [ ] Remove `bd` support (future release)

---

## Testing Strategy

### TUI Testing (via tmux)

AI agents use `scripts/tui-test.sh` for comprehensive UI verification instead of manual testing:

```bash
# Basic workflow
make build                         # Build after code changes
./scripts/tui-test.sh start        # Launch abacus in tmux
./scripts/tui-test.sh keys 'jjjl'  # Navigate (j=down, l=expand)
./scripts/tui-test.sh enter        # Open detail pane
./scripts/tui-test.sh view         # Capture and verify state
./scripts/tui-test.sh quit         # Clean up
```

| Test Scenario | tmux Commands | Verification |
|---------------|---------------|--------------|
| Tree navigation | `keys 'jjjkk'` | View shows cursor movement |
| Expand/collapse | `keys 'lh'` | Children appear/disappear |
| Detail pane | `enter` | Pane opens with issue details |
| Create overlay | `keys 'c'` | Create form appears |
| Status change | `keys 's'` | Status picker appears |
| Backend indicator | `view` | Status bar shows `bd` or `br` |
| Search | `keys '/'`, type query | Filtered results appear |

**Backend-specific tests:**
| Test | Purpose |
|------|---------|
| Start with `br` only on PATH | Verify auto-detection selects `br` |
| Start with `bd` only on PATH | Verify auto-detection selects `bd` |
| Start with both, config set | Verify config preference is respected |
| Create issue via overlay | Verify `br` CLI is invoked for mutations |
| Backend indicator visible | Verify status bar shows correct backend |

### Unit Tests

| Test | Purpose |
|------|---------|
| `TestBrSQLiteClient_List` | Verify SQLite list query (Reader interface) |
| `TestBrSQLiteClient_Export` | Verify SQLite export query (Reader interface) |
| `TestBrCLIClient_Create` | Verify create command (Writer interface) |
| `TestBrCLIClient_Update` | Verify update command (Writer interface) |
| `TestBrCLIClient_Delete` | Verify delete command (Writer interface) |
| `TestBackendDetection` | Verify auto-detection logic |
| `TestBackendConfig` | Verify config-based selection |
| `TestStalePreferenceHandling` | Verify re-detection when binary missing |

### Integration Tests

| Test | Purpose |
|------|---------|
| `TestBrBackend_E2E` | Full workflow with `br` binary |
| `TestBdBackend_E2E` | Full workflow with `bd` binary (regression) |
| `TestSQLiteWithBr` | SQLite reads from `br`-created DB |
| `TestMixedOperations` | Create with `br`, read with SQLite |

### Conformance Tests

| Test | Purpose |
|------|---------|
| `TestOutputConformance_List` | `bd list` and `br list` produce equivalent output |
| `TestOutputConformance_Show` | `bd show` and `br show` produce equivalent output |
| `TestSchemaConformance` | Both backends produce compatible SQLite schemas |

**Note:** `TestOutputConformance_Show` will pass once br deploys the `dep_type` → `dependency_type` rename (see [Key Difference: Dependency Field Names](#key-difference-dependency-field-names)). Until then, the test should normalize or skip the dependency type field comparison.

---

## Migration Path

### For Existing bd Users

1. **Continue using `bd`** - Abacus detects `bd` on PATH and uses it
2. **One-time prompt if both installed** - If both `bd` and `br` are on PATH, abacus prompts once and saves preference
3. **When ready to migrate** - Use `br` in the repo; detection will update accordingly
4. **Mixed usage** - Different repos can use different backends

### For New Users

1. **Works with either** - Abacus works with whichever tool creates the repo
2. **Documentation** - Will explain both options are supported

### For Users with Mixed Repos

1. **Automatic per-repo detection** - Each repo detected independently
2. **Backend indicator** - Status bar shows which backend is active
3. **No configuration** - Just open abacus in the repo and it works

### Deprecation Strategy

**Decision:** Keep both backends indefinitely until maintenance burden justifies removal.

There is no fixed deprecation timeline. Instead:
- Monitor maintenance cost of supporting both backends
- Remove `bd` support only when it becomes burdensome
- This could be triggered by: breaking changes in bd > 0.38.0, test maintenance burden, or user base shifting to br

---

## Open Questions

### Resolved Questions (From Research)

| Question | Answer | Impact |
|----------|--------|--------|
| **Q1: CLI syntax identical?** | ⚠️ **Almost** - One difference: `create` uses `--title-flag` in br vs `--title` in bd. Fix requested: [beads_rust#7](https://github.com/Dicklesworthstone/beads_rust/issues/7). **Decision: Use positional title syntax** which works for both. | Minor refactor in abacus |
| **Q2: JSON format identical?** | ✅ **Non-issue** - `dep_type` vs `dependency_type` exists but CLI JSON parsing is dead code in production | No change needed |
| **Q3: SQLite schema compatible?** | ✅ Yes - all columns abacus needs exist in both; detectSchema() being removed (pre-Decision-004 DBs not supported) | Minor refactor |
| **Q4: Export command exists?** | ❌ No - **non-issue** (CLI export is dead code; SQLite reads used in production) | No change needed |
| **Q5: Minimum br version?** | **0.1.7** (currently the only version) | Infrastructure ready for future versions |
| **Q6: --cascade flag exists?** | ✅ Yes | No change needed |
| **Q7: Labels on create?** | ✅ `--labels` flag exists | No change needed |

---

### Implementation Findings Summary

Based on research, implementation is **straightforward** with these specific changes:

| Area | Change Required | Effort |
|------|-----------------|--------|
| **CLI commands** | ✅ Identical except `create` title flag (use positional syntax) | ~5 lines |
| **CLI Export** | ⚠️ Dead code - only used in tests | 0 lines (can ignore) |
| **Dependency JSON** | ✅ **Non-issue** - CLI JSON parsing is dead code in production | 0 lines |
| **Version check** | Add br version detection (min 0.1.7) | ~30 lines |
| **Backend selection** | PATH check + user prompt + config storage | ~120 lines |
| **SQLite** | No changes needed (schema compatible) | 0 lines |

**Key insight (verified at code level):** The CLI client is only used for **mutation operations** in production. All read operations (Export, List, Show, Comments) go through SQLite direct queries. This means:
- `br`'s lack of an `export` command is irrelevant
- The `dep_type` vs `dependency_type` JSON difference is irrelevant (dead code path)
- The `brCLIClient` only needs to implement write methods (12 mutation commands)
- Read methods use `sqliteClient` which already works with br's schema

**Total estimated changes:** See [File Structure](#file-structure) section for detailed breakdown (~760 new lines + ~550 refactored)

---

### Version Checking (Verified)

**Current implementation** (`internal/beads/version_check.go`):
- Minimum version: `0.30.0` (constant `MinBeadsVersion`)
- Runs `bd --version` and parses semver
- **Enforced** - hard exit if too old or missing (not just a warning)
- Bypass: `--skip-version-check` flag or `AB_SKIP_VERSION_CHECK=true`

**For br support:**
- `MinBrVersion = "0.1.7"` (currently the only br version; infrastructure ready for future versions)
- `br --version` command format is identical to `bd --version`

**Startup flow (order matters):**
```
1. Backend detection runs FIRST
   - Check config for stored preference
   - If not set, check PATH for available binaries
   - If both on PATH, prompt user
   - Result: backend is now known ("bd" or "br")

2. Version check runs SECOND (only for selected backend)
   - If backend == "bd": check bd --version >= 0.30.0
   - If backend == "br": check br --version >= 0.1.7
   - Only the selected backend is checked (not both)
```

**Implementation location:**
- `internal/beads/version_check.go` - CheckVersion(backend string), parseSemver()
- `cmd/abacus/version_gate.go` - handleVersionCheckResult()

---

### CLI Commands Actually Used in Production (Verified)

**Read operations** (SQLite direct - CLI is dead code):
- `Export()` - reads from SQLite, not `bd export`
- `List()` - reads from SQLite, not `bd list`
- `Show()` - reads from SQLite, not `bd show`
- `Comments()` - reads from SQLite, not `bd comments`

**Write operations** (CLI invoked - verified in `update_commands.go`):
| Method | Command | Verified |
|--------|---------|----------|
| UpdateStatus | `bd update ID --status=X` | ✅ |
| Reopen | `bd reopen ID` | ✅ |
| AddLabel | `bd label add ID LABEL` | ✅ |
| RemoveLabel | `bd label remove ID LABEL` | ✅ |
| Create | `bd create "Title" --type Y --priority N` | ✅ (no --json, parses ID from text) |
| CreateFull | `bd create "Title" --type Y --priority N --json` then `bd dep add` | ✅ (uses --json for full issue) |
| UpdateFull | `bd update ID --title ... --set-labels ...` | ✅ |
| AddDependency | `bd dep add FROM TO --type TYPE` | ✅ |
| RemoveDependency | `bd dep remove FROM TO` | ✅ |
| Delete | `bd delete ID --force [--cascade]` | ✅ |
| AddComment | `bd comments add ID TEXT` | ✅ |

**Almost all mutation commands are identical between bd and br** - verified against br CLI source code. The one exception is the `create` command where the `--title` flag differs (`--title` in bd, `--title-flag` in br). A fix was requested: [beads_rust#7](https://github.com/Dicklesworthstone/beads_rust/issues/7). **Decision: Abacus will use positional title syntax** which works for both backends, avoiding the need to wait for the br fix.

---

### Resolved User Questions

**Q8.1:** Should abacus display which backend is active in the status bar?

**Decision: A - Always show indicator**

User said: "Always show indicator" - this provides transparency about which tool is in use.

---

**Q8.2:** What should happen if both `bd` and `br` are available?

**Decision: Per-repo auto-detection (neither is globally preferred)**

User said: "Abacus must use whichever tool the user is using in that repo. It's not like we can prefer one over another. We must use what the user is using. So auto-detection is ideal."

**Key insight:** Users may have some repos on `bd` and others on `br`. There is no "preference" - abacus must detect and use the correct backend for each repo.

---

**Q8.3:** Should we support per-project backend selection?

**Decision: Yes, per-project config is always used**

- If only one backend (`bd` or `br`) is on PATH → use it automatically
- If both are on PATH → prompt user to choose (default: `br`)
- User's choice is **always stored per-project** in `.abacus/config.yaml` (created if needed)
- Different projects can use different backends
- User can also manually edit `beads.backend` in `.abacus/config.yaml`

---

**Q8.4:** What's the deprecation timeline for bd support?

**Decision: C - Keep both indefinitely (until maintenance burden justifies removal)**

User said: "I will keep support for bd until I feel that it's not necessary anymore, or it becomes a pain to keep supporting." This is a practical, maintenance-driven approach rather than a fixed timeline.

---

### Backend Detection Strategy

**Problem:** How does abacus determine which backend (`bd` or `br`) to use?

**Solution:** Check which binaries exist on PATH. If both exist, ask user.

**Integration with existing config system:** Abacus already has a robust configuration system (`internal/config/config.go`) that handles:
- User config: `~/.abacus/config.yaml`
- Project config: `.abacus/config.yaml` (discovered by walking up directory tree)
- Environment variables: `AB_` prefix (e.g., `AB_SKIP_VERSION_CHECK=true`)
- CLI flag overrides

The backend preference uses the **project config** portion of this system (not env vars) to ensure each repository uses its correct backend.

---

**Detection Flow:**

```
0. Check for CLI flag override (highest priority)
   - `--backend bd` or `--backend br` skips all detection
   - Used for CI/scripts or explicit override
   - Does NOT save to config (one-time override)
   - If specified binary not on PATH → error immediately

1. Check for stored preference
   - Read from project config (.abacus/config.yaml) only
   - NOTE: Env var (AB_BEADS_BACKEND) is NOT supported for backend selection
   → If stored pref exists AND binary exists on PATH → use it
   → If stored pref exists BUT binary NOT on PATH → prompt user before clearing (see step 1b)

1b. Handle stale preference (stored binary not on PATH)
   - Show: "This project is configured for 'br' but br is not found in PATH"
   - If other binary available → prompt: "Switch to bd? (Your preference will be updated)"
   - If neither available → error: "Neither bd nor br found in PATH"
   - Only update stored preference AFTER user confirms the switch

2. If no stored preference (or stale preference cleared):
   - Check which binaries exist on PATH (no testing, no probing)

   → Only br on PATH → use br
   → Only bd on PATH → use bd
   → Both on PATH AND interactive TTY → ask user (br pre-selected)
   → Both on PATH AND non-TTY (CI) → error: "Both bd and br found; use --backend flag or set config"
   → Neither on PATH → error

3. Version check (runs BEFORE saving)
   - Check selected backend meets minimum version (bd >= 0.30.0, br >= 0.1.7)
   - If version check fails AND other backend available → offer to switch (if interactive)
   - If version check fails AND no other backend → error with install instructions
   - This allows user to choose the other backend if their first choice is too old

4. Save validated backend (skip if --backend flag was used)
   - Only save AFTER version check passes
   - Use SaveBackend() - creates .abacus/config.yaml if needed
   - Saves even for auto-detected backends (avoids re-detection on next launch)
   → Show: "Saved to .abacus/config.yaml - edit beads.backend to change later"
```

**CI/Non-Interactive Usage:**

For CI pipelines or scripts where prompts would hang:
```bash
# Explicit backend selection (recommended for CI)
abacus --backend br

# Or pre-configure the project
echo "beads:\n  backend: br" > .abacus/config.yaml
```

If both binaries exist and no config/flag is provided in a non-TTY environment, abacus exits with an error rather than hanging on a prompt.

---

**Implementation:**

```go
// internal/beads/backend.go

import (
    "errors"
    "fmt"
    "log"
    "os"
    "os/exec"

    "github.com/yourusername/abacus/internal/config"
    "golang.org/x/term"
)

// DetectBackend determines which backend to use.
// cliFlag is the value of --backend flag (empty if not provided).
func detectBackend(cliFlag string) (string, error) {
    // 0. CLI flag override (highest priority, one-time, no save)
    if cliFlag != "" {
        if cliFlag != "bd" && cliFlag != "br" {
            return "", fmt.Errorf("invalid --backend value: %q (must be 'bd' or 'br')", cliFlag)
        }
        if !commandExists(cliFlag) {
            return "", fmt.Errorf("--backend %s specified but %s not found in PATH", cliFlag, cliFlag)
        }
        if err := checkVersion(cliFlag); err != nil {
            return "", fmt.Errorf("--backend %s version check failed: %w", cliFlag, err)
        }
        // Don't save - CLI flag is a one-time override
        return cliFlag, nil
    }

    // 1. Check stored preference (project config ONLY - no env var support)
    storedPref := config.GetProjectString(config.KeyBeadsBackend)
    if storedPref != "" {
        // Verify the stored preference is still valid (binary exists)
        if commandExists(storedPref) {
            // Version check for stored preference (already saved, just validate)
            if err := checkVersion(storedPref); err != nil {
                return "", fmt.Errorf("stored backend '%s' version check failed: %w", storedPref, err)
            }
            return storedPref, nil
        }
        // 1b. Stale preference - prompt user before clearing
        return handleStalePreference(storedPref)
    }

    // 2. Check binary availability (PATH only, no probing)
    brExists := commandExists("br")
    bdExists := commandExists("bd")

    var choice string
    switch {
    case !brExists && !bdExists:
        return "", errors.New("neither bd nor br found in PATH")
    case brExists && !bdExists:
        choice = "br"
    case bdExists && !brExists:
        choice = "bd"
    case brExists && bdExists:
        // Both exist - need user input
        if !isInteractiveTTY() {
            return "", errors.New("both bd and br found in PATH; use --backend flag or set beads.backend in .abacus/config.yaml")
        }
        choice = promptUserForBackend() // returns "br" or "bd"
    }

    // 3. Version check BEFORE saving - allows user to switch if version fails
    choice, err := validateWithFallback(choice, brExists, bdExists)
    if err != nil {
        return "", err
    }

    // 4. Save validated choice
    // Note: SaveBackend may fail if no .beads/ directory exists, but main.go
    // validates .beads/ presence before calling detectBackend(), so this is
    // defense-in-depth. Log warning but continue since detection succeeded.
    if err := SaveBackend(choice); err != nil {
        log.Printf("warning: could not save backend preference: %v", err)
    }

    return choice, nil
}

func validateWithFallback(choice string, brExists, bdExists bool) (string, error) {
    if err := checkVersion(choice); err == nil {
        return choice, nil // Version check passed
    }

    // Version check failed - is there an alternative?
    other := "bd"
    if choice == "bd" {
        other = "br"
    }
    otherExists := (other == "br" && brExists) || (other == "bd" && bdExists)

    if !otherExists {
        return "", fmt.Errorf("%s version is too old (see requirements) and no alternative backend available", choice)
    }

    // Offer to switch to the other backend
    fmt.Printf("%s version is too old. Would you like to use %s instead?\n", choice, other)
    if confirmed := promptSwitchBackend(other); confirmed {
        // Check the alternative's version too
        if err := checkVersion(other); err != nil {
            return "", fmt.Errorf("both backends have version issues: %s and %s", choice, other)
        }
        return other, nil
    }

    return "", fmt.Errorf("cannot continue: %s version too old and user declined switch to %s", choice, other)
}

func handleStalePreference(storedPref string) (string, error) {
    // Determine which binary (if any) is available as alternative
    other := "bd"
    if storedPref == "bd" {
        other = "br"
    }
    otherExists := commandExists(other)

    if !otherExists {
        return "", fmt.Errorf("this project is configured for '%s' but neither bd nor br found in PATH", storedPref)
    }

    // Prompt user: their configured backend is missing, offer to switch
    fmt.Printf("This project is configured for '%s' but %s is not found in PATH.\n", storedPref, storedPref)
    if confirmed := promptSwitchBackend(other); confirmed {
        // Version check BEFORE saving
        if err := checkVersion(other); err != nil {
            return "", fmt.Errorf("cannot switch to %s: version too old (%w)", other, err)
        }
        if err := SaveBackend(other); err != nil {
            log.Printf("warning: could not save backend preference: %v", err)
        }
        return other, nil
    }

    // User declined to switch - exit with helpful message
    return "", fmt.Errorf("cannot continue: '%s' not found in PATH (add it to PATH or accept switch to '%s')", storedPref, other)
}

func commandExists(name string) bool {
    _, err := exec.LookPath(name)
    return err == nil
}

func isInteractiveTTY() bool {
    return term.IsTerminal(int(os.Stdin.Fd()))
}
```

```go
// internal/config/config.go - add new key

const (
    // ... existing keys ...
    KeyBeadsBackend = "beads.backend"  // "bd" or "br"
)

func setDefaults(v *viper.Viper) {
    // ... existing defaults ...
    v.SetDefault(KeyBeadsBackend, "")  // Empty means auto-detect
}

// SaveBackend persists the backend name to the project config file.
// Unlike SaveTheme(), this ALWAYS saves to project config (.abacus/config.yaml)
// because backend selection is inherently per-project.
func SaveBackend(backend string) error {
    // Always save to project config - backend is per-project
    wd, err := os.Getwd()
    if err != nil {
        return fmt.Errorf("get working directory: %w", err)
    }

    // Find .beads/ directory to locate project root
    beadsDir := findBeadsDir(wd)
    if beadsDir == "" {
        return fmt.Errorf("no .beads directory found")
    }
    projectRoot := filepath.Dir(beadsDir)
    targetPath := filepath.Join(projectRoot, ".abacus", "config.yaml")

    v := viper.New()
    v.SetConfigType("yaml")
    v.SetConfigFile(targetPath)
    _ = v.ReadInConfig()  // ignore error if file doesn't exist

    v.Set(KeyBeadsBackend, backend)

    // Create .abacus/ directory if needed
    dir := filepath.Dir(targetPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("create config directory: %w", err)
    }

    if err := v.WriteConfigAs(targetPath); err != nil {
        return fmt.Errorf("write config: %w", err)
    }

    return nil
}
```

---

**Config file format:**

```yaml
# .abacus/config.yaml (project-level only)
beads:
  backend: br  # or "bd"

# NOTE: Environment variable AB_BEADS_BACKEND is NOT supported.
# Backend selection is always per-project to ensure the correct
# backend is used for each repository.
```

---

**User prompt (when both binaries on PATH):**

```
Both bd and br are available. Which backend does this project use?

  > br
    bd

Saved to .abacus/config.yaml - edit beads.backend to change later.
```

**Prompt timing:** The backend selection prompt appears **before** the main Bubble Tea TUI starts:

1. Abacus launches (no TUI yet)
2. Backend detection runs in `main.go`
3. If prompt needed → use `huh` library (Charm's form library) for interactive selection
4. Save preference, print confirmation
5. Initialize and start the Bubble Tea TUI

**Implementation:** Use [huh](https://github.com/charmbracelet/huh) for all startup prompts. `huh` is a Charm library specifically designed for CLI forms and prompts that runs independently of Bubble Tea. This avoids the complexity of a Bubble Tea model just for simple yes/no or selection prompts during startup.

```go
// Example using huh for backend selection
import "github.com/charmbracelet/huh"

func promptUserForBackend() string {
    var choice string
    huh.NewSelect[string]().
        Title("Both bd and br are available. Which backend does this project use?").
        Options(
            huh.NewOption("br (recommended)", "br"),
            huh.NewOption("bd", "bd"),
        ).
        Value(&choice).
        Run()
    return choice
}
```

---

**Edge cases:**

| Scenario | Behavior |
|----------|----------|
| `--backend bd` or `--backend br` flag | Use specified backend (one-time, no save to config) |
| No `.beads/` directory | Error - no beads project |
| Neither binary on PATH | Error - "install bd or br" |
| Both on PATH (interactive TTY) | Ask user (br pre-selected) |
| Both on PATH (non-TTY / CI) | Error - "use --backend flag or configure .abacus/config.yaml" |
| Project has `.abacus/config.yaml` | Read preference from there |
| No `.abacus/config.yaml` in project | Create it after user prompt |
| Stored preference but binary missing | Prompt user to confirm switch; only update preference after confirmation |
| `AB_BEADS_BACKEND` env var set | **Ignored** - env var not supported for backend |

---

**Resolved: bd version > 0.38.0 handling**

**Decision:** Show one-time warning, never prompt again.

When user selects `bd` backend and version is > 0.38.0, check user config for `beads.bd_unsupported_version_warned`:
- If `true` → skip warning (already notified)
- If not set → show warning once, set flag, continue

**Warning message (non-blocking):**

```
Note: abacus officially supports beads (bd) up to version 0.38.0.
You have version X.Y.Z installed. Newer versions may work but are
not officially supported.

Future development is focused on beads_rust (br):
https://github.com/Dicklesworthstone/beads_rust

This message will not be shown again.
```

**Behavior:**
- **One-time notification** → display message, save `beads.bd_unsupported_version_warned: true` to user config, proceed automatically
- **No prompt/acceptance required** → just informational
- **Stored in user config** (`~/.abacus/config.yaml`) since this is user-level
- **Never shown again** → regardless of future bd version upgrades

---

## Decision Log

| Date | Decision | Rationale | Alternatives Considered |
|------|----------|-----------|------------------------|
| 2025-01-18 | Single codebase, not fork | 95% code shared, maintenance burden | Fork abacus, maintain separately |
| 2025-01-18 | Support both during transition | Smooth migration for users | Hard cutover, `br` only |
| 2025-01-18 | Freeze `bd` support at 0.38.0 | Avoid tracking moving target | Track latest `bd` |
| 2025-01-18 | Always show backend indicator | User wants transparency about which tool is active | Hide when default, debug only |
| 2025-01-18 | Keep both indefinitely | Remove bd only when maintenance burden justifies | Fixed deprecation timeline |
| 2025-01-18 | PATH check + user prompt | Simple, no probing, no daemon risk | Binary testing, daemon detection, fingerprinting |
| 2025-01-18 | Default selection br when both on PATH | br is future direction; user still must confirm | Auto-select br, no default |
| 2025-01-18 | Store pref in existing config format | Use `.abacus/config.yaml` format, not a new file type | Create new file type like `.beads/abacus.json` |
| 2025-01-18 | Config system supports user-level defaults | General config uses `~/.abacus/config.yaml` for user defaults (theme, etc.). **Exception:** `beads.backend` is always per-project only. | Global only, repo only |
| 2025-01-19 | Separate CLI+SQLite per backend | bd frozen, br evolving - no conditionals needed | Shared SQLite client with conditionals |
| 2025-01-19 | **Code-level verification complete** | All claimed bd/br differences verified at source level - most are non-issues due to SQLite reads in production | Accept plan assumptions without verification |
| 2025-01-19 | **Migrate graph links to dependency-based** | ab-qdh0 used columns that don't exist anymore. Both bd and br use `duplicates`/`supersedes` dependency types instead. Refactor to use dependency data. | Remove feature entirely |
| 2025-01-19 | **Use existing config system** | Abacus has robust config infrastructure at `internal/config/` with Viper. Add `beads.backend` key and `SaveBackend()` function following `SaveTheme()` pattern. No need for separate config file in `.beads/` | Create separate config file at `.beads/config.yaml` |
| 2025-01-19 | **bd > 0.38.0: one-time warning** | Show informational warning once when user first uses bd > 0.38.0. Set `bd_unsupported_version_warned` flag in user config. Never prompt again regardless of future bd upgrades. | Hard block, per-version acceptance, warn every session |
| 2025-01-19 | **Validate stale preferences** | If stored backend preference points to binary not on PATH, prompt user before clearing | Trust stored preference, error on missing binary, silent clear |
| 2025-01-19 | **Version check before save** | Run version check BEFORE saving backend preference; if version fails and alternative exists, offer to switch | Save first then check (locks user into bad choice) |
| 2025-01-19 | **--backend CLI flag for CI** | Add `--backend bd\|br` flag for non-interactive override; in non-TTY when both binaries exist, error instead of hanging on prompt | Env var override, always prompt, silent default |
| 2025-01-19 | **Project-level config for backend** | Backend preference stored in `.abacus/config.yaml` per project; create directory if needed | User-level only, no new repo files |
| 2025-01-19 | **No pre-Decision-004 support** | Don't support old bd databases with `duplicate_of`/`superseded_by` columns; users must upgrade bd | Backward compatibility shim |
| 2025-01-19 | **Use huh for startup prompts** | Backend selection and other startup prompts use `huh` library (Charm's form library) which runs before Bubble Tea starts | Bubble Tea model for prompt, custom stdin handling |
| 2025-01-19 | **Read-only SQLite, CLI writes** | Abacus only reads from SQLite; all writes go through bd/br CLI which handles concurrency | Custom locking logic |
| 2025-01-19 | **Separate clients over parameterization** | Duplicate code preferred over shared client with conditionals; allows br to evolve without affecting frozen bd code | Single parameterized client |
| 2025-01-19 | **User-level bd warning flag** | `bd_unsupported_version_warned` flag stored in user config (`~/.abacus/config.yaml`) since it's a one-time user notification | Project-level flag |
| 2025-01-19 | **No env var for backend** | `AB_BEADS_BACKEND` env var not supported; backend read from project config only to ensure correct backend per repo | Env var overrides project config |
| 2025-01-19 | **Split Client into Reader + Writer** | CLI clients only do writes (read methods were dead code); SQLite clients do reads + delegate writes. Clean separation eliminates dead code and means brCLIClient doesn't need Export() | Keep single interface, stub Export() on br |
| 2025-01-19 | **Use positional title for create command** | br uses `--title-flag` while bd uses `--title`. Positional title (`bd create "Title"`) works for both. GitHub issue filed: [beads_rust#7](https://github.com/Dicklesworthstone/beads_rust/issues/7) but we won't wait for the fix. | Wait for br fix, conditional flag names |
| 2025-01-19 | **Permissive reading, restrictive creation** | Abacus is a viewer/editor, not the source of truth. Accept unknown statuses/types/dep-types when reading (never reject valid data). Only offer known options when creating. Preserve unknown values when editing. Unknown dep types default to non-blocking. | Strict validation (reject unknowns), auto-convert unknowns to defaults |

---

## File Structure

### File Layout

```
internal/config/
└── config.go           # Update - Add KeyBeadsBackend + SaveBackend()

internal/beads/
├── client.go           # **Refactor** - Split into Reader, Writer, Client interfaces
├── types.go            # Data types (shared, may split later if needed)
│
│   # bd Backend (FROZEN at 0.38.0)
├── bd_cli.go           # RENAME from cli.go - bdCLIClient
├── bd_sqlite.go        # RENAME from sqlite_client.go - bdSQLiteClient
│
│   # br Backend (EVOLVING - new features go here)
├── br_cli.go           # NEW - brCLIClient
├── br_sqlite.go        # NEW - brSQLiteClient
│
│   # Shared Infrastructure
├── backend.go          # NEW - Detection, binary check, user prompt, factory
└── version_check.go    # Update - Add br version validation

internal/tui/
└── statusline.go       # Update - Add backend indicator
```

**Note:** No separate `config.go` in beads package needed - leverages existing `internal/config/` system.

### Estimated Lines of Code

| File | Estimated Lines | Notes |
|------|-----------------|-------|
| `client.go` refactor | ~60 | Split into Reader + Writer interfaces |
| `bd_cli.go` | ~250 | Existing code, renamed, **read methods removed** (frozen) |
| `bd_sqlite.go` | ~300 | Existing code, renamed, embeds Writer (frozen) |
| `br_cli.go` | ~200 | Writer only - no read methods needed |
| `br_sqlite.go` | ~300 | Based on bd_sqlite.go, can evolve independently |
| `backend.go` | ~120 | Binary check, user prompt, factory |
| `internal/config/config.go` | ~40 | Add KeyBeadsBackend + SaveBackend() |
| Version check | ~30 | br version validation |
| Status bar indicator | ~10 | Show `bd` or `br` |
| **New code** | **~760** | Plus tests |
| **Refactored code** | **~550** | Existing, frozen (smaller due to removed dead code) |

---

## Dependencies

**New dependency:**
- `github.com/charmbracelet/huh` - CLI forms/prompts for backend selection (runs before Bubble Tea)

**Existing dependencies (no changes):**
- `os/exec` - CLI execution
- `encoding/json` - JSON parsing
- `database/sql` + `modernc.org/sqlite` - SQLite access

---

## References

- **Abacus source:** `/Users/chrisedwards/projects/oss/abacus`
- **beads source:** `/Users/chrisedwards/projects/oss/beads`
- **beads_rust source:** `/Users/chrisedwards/projects/oss/beads_rust`
- **beads_rust planning doc:** `/Users/chrisedwards/projects/oss/beads_rust/PLAN_TO_PORT_BEADS_WITH_SQLITE_AND_ISSUES_JSONL_TO_RUST.md`
- **beads_rust CLI docs:** `/Users/chrisedwards/projects/oss/beads_rust/docs/CLI_REFERENCE.md`

---

## Next Steps

1. ✅ ~~**Research Phase** - Answer all [RESEARCH NEEDED] questions~~ - COMPLETE
2. ✅ ~~**Interview** - Discuss remaining user questions (Q8.1-Q8.4) with project owner~~ - COMPLETE
3. ✅ ~~**Update Plan** - Fill in placeholders with findings~~ - COMPLETE
4. ⏳ **Begin Implementation** - Phase 2 onwards (ready to start)

### Implementation Ready

All research and user decisions are complete. The plan is ready for implementation:

- **Backend detection:** PATH check + user prompt when both binaries exist
- **Separate clients:** bd backend (frozen) + br backend (evolving)
- **UI:** Always show backend indicator (`bd` or `br`)
- **Config:** Backend preference stored per-repo (`.abacus/config.yaml`); bd unsupported version warning flag stored per-user (`~/.abacus/config.yaml`)
- **Deprecation:** Keep both backends until maintenance burden justifies removal

**Recommended first step:** Begin Phase 2 - refactor the Client interface into Reader + Writer to eliminate dead code and prepare for clean br implementation.

---

## Appendix A: Refactored Abacus Client Interface

**Current interface (before refactor):** Single `Client` interface with both read and write methods, leading to dead code in CLI clients.

**New interface (after Phase 2 refactor):**

```go
// internal/beads/client.go

// Reader handles all read operations (SQLite-only in production)
type Reader interface {
    List(ctx context.Context) ([]LiteIssue, error)
    Show(ctx context.Context, ids []string) ([]FullIssue, error)
    Export(ctx context.Context) ([]FullIssue, error)
    Comments(ctx context.Context, issueID string) ([]Comment, error)
}

// Writer handles all mutation operations (CLI-only in production)
type Writer interface {
    UpdateStatus(ctx context.Context, issueID, newStatus string) error
    Close(ctx context.Context, issueID string) error
    Reopen(ctx context.Context, issueID string) error
    AddLabel(ctx context.Context, issueID, label string) error
    RemoveLabel(ctx context.Context, issueID, label string) error
    UpdateFull(ctx context.Context, issueID, title, issueType string, priority int, labels []string, assignee, description string) error
    Create(ctx context.Context, title, issueType string, priority int, labels []string, assignee string) (string, error)
    CreateFull(ctx context.Context, title, issueType string, priority int, labels []string, assignee, description, parentID string) (FullIssue, error)
    AddDependency(ctx context.Context, fromID, toID, depType string) error
    RemoveDependency(ctx context.Context, fromID, toID, depType string) error
    Delete(ctx context.Context, issueID string, cascade bool) error
    AddComment(ctx context.Context, issueID, text string) error
}

// Client combines Reader and Writer for full functionality
type Client interface {
    Reader
    Writer
}
```

**Implementation mapping:**

| Component | Implements | Read Methods | Write Methods |
|-----------|------------|--------------|---------------|
| `bdSQLiteClient` | `Client` | SQLite queries | Delegates to embedded `bdCLIClient` |
| `brSQLiteClient` | `Client` | SQLite queries | Delegates to embedded `brCLIClient` |
| `bdCLIClient` | `Writer` | N/A (removed) | CLI commands (`bd`) |
| `brCLIClient` | `Writer` | N/A (not needed) | CLI commands (`br`) |

**Benefits:**
- CLI clients don't need to implement read methods (eliminates dead code)
- `brCLIClient` doesn't need Export() (which `br` doesn't support)
- Clear separation of concerns
- SQLite clients embed a `Writer` instead of full `Client`

**Concurrency note:** SQLite concurrent access is not a concern because abacus only performs **read-only** SQLite operations. All writes go through the `bd`/`br` CLI, which handles its own locking and transaction management.

---

## Appendix B: Current bd Command Mappings

From `internal/beads/cli.go`. **Note:** In production, `sqliteClient` is used, which reads from SQLite directly and only delegates **mutations** to the CLI.

### Read Operations (SQLite only - removed from CLI clients after Phase 2)

| Method | CLI Command | Status After Refactor |
|--------|-------------|----------------------|
| Export | `bd export` | ❌ **Removed** from CLI clients |
| List | `bd list --json` | ❌ **Removed** from CLI clients |
| Show | `bd show ID... --json` | ❌ **Removed** from CLI clients |
| Comments | `bd comments ID --json` | ❌ **Removed** from CLI clients |

**Note:** After the Phase 2 interface refactor, CLI clients only implement `Writer`. Read operations are handled exclusively by SQLite clients (which implement `Reader`).

### Write Operations (CLI in production)

| Method | bd Command | Production Usage |
|--------|------------|------------------|
| UpdateStatus | `bd update ID --status=X` | ✅ CLI |
| Close | `bd close ID` | ✅ CLI |
| Reopen | `bd reopen ID` | ✅ CLI |
| AddLabel | `bd label add ID LABEL` | ✅ CLI |
| RemoveLabel | `bd label remove ID LABEL` | ✅ CLI |
| Create | `bd create --title X --type Y --priority N [--labels L] [--assignee A]` | ✅ CLI |
| CreateFull | `bd create --title X --type Y --priority N --json [--description D]` + `bd dep add` | ✅ CLI |
| UpdateFull | `bd update ID --title X --description D --priority N [--assignee A] [--set-labels L]` | ✅ CLI |
| AddDependency | `bd dep add FROM TO --type TYPE` | ✅ CLI |
| RemoveDependency | `bd dep remove FROM TO` | ✅ CLI |
| Delete | `bd delete ID --force [--cascade]` | ✅ CLI |
| AddComment | `bd comments add ID TEXT` | ✅ CLI |

**Implication for br support:** Only the write operations need CLI compatibility. Almost all mutation commands are identical between `bd` and `br`. The `create` command has one difference: abacus currently uses `--title` flag (bd-only), but positional title works for both. **Decision: Change abacus to use positional title syntax** - this avoids waiting for br fix and works with both backends.
