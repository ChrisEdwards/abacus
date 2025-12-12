# feat: Edit Existing Beads

## Overview

Add the ability to edit existing beads in the abacus TUI. Users can modify bead fields like title, description, type, priority, and assignee through a comprehensive edit modal, triggered by pressing `e` on any selected bead.

## Problem Statement

Currently, users can only edit status (via `s`) and labels (via `L`). Other fields like title, description, type, priority, and assignee require using the `bd` CLI directly. This breaks the TUI workflow and forces context switching.

## Proposed Solution

Add a comprehensive **Edit Overlay** (triggered by `e` key) that allows editing all common bead fields in a single modal, following the patterns established by CreateOverlay.

### Approach: Comprehensive Edit Modal

**Why this approach:**
- Matches existing CreateOverlay UX pattern
- Allows editing multiple fields in one operation
- Keeps field-specific overlays (status/labels) for quick single-field edits
- Follows established codebase patterns

### Fields to Edit

| Field | Input Type | Validation | Notes |
|-------|------------|------------|-------|
| Title | Single-line textarea | Required, max 100 chars | Flash red if empty |
| Description | Multi-line textarea | Optional | 6-line height |
| Type | Pill selector | task/feature/bug/epic/chore | Horizontal pills with hotkeys |
| Priority | Pill selector | 0-4 (P0-P4) | Horizontal pills with hotkeys |
| Assignee | ComboBox | Optional | Autocomplete from existing assignees |

### Hotkey

- `e` - Open edit overlay on selected bead
- Existing overlays remain: `s` (status), `L` (labels), `n/N` (create), `d` (delete)

## Technical Approach

### 1. Extend Client Interface

**File:** `internal/beads/client.go`

```go
// Add to Client interface (around line 20)
UpdateTitle(ctx context.Context, issueID, newTitle string) error
UpdateDescription(ctx context.Context, issueID, newDescription string) error
UpdateType(ctx context.Context, issueID, newType string) error
UpdatePriority(ctx context.Context, issueID string, newPriority int) error
UpdateAssignee(ctx context.Context, issueID, newAssignee string) error
```

### 2. Implement Client Methods

**File:** `internal/beads/cli.go`

```go
// After UpdateStatus (line 121), add new methods:
func (c *cliClient) UpdateTitle(ctx context.Context, issueID, newTitle string) error {
    _, err := c.run(ctx, "update", issueID, "--title="+newTitle)
    return err
}

func (c *cliClient) UpdateDescription(ctx context.Context, issueID, newDescription string) error {
    _, err := c.run(ctx, "update", issueID, "--description="+newDescription)
    return err
}

// ... similar for Type, Priority, Assignee
```

### 3. Create Edit Overlay

**New File:** `internal/ui/overlay_edit.go`

Structure (following CreateOverlay pattern):
- `EditOverlay` struct with fields for each editable property
- Focus management using `EditFocus` enum
- Pre-populate all fields from current bead values
- Tab/Shift+Tab navigation between fields
- Enter saves, Esc cancels (discard without confirmation)

### 4. Add Keybinding

**File:** `internal/ui/keys.go`

```go
// Add to KeyMap struct (around line 29)
Edit key.Binding

// Add to DefaultKeyMap() (around line 127)
Edit: key.NewBinding(
    key.WithKeys("e"),
    key.WithHelp("e", "edit"),
),
```

### 5. Wire Up Messages

**File:** `internal/ui/update.go`

```go
// Add message handlers (around line 22-50)
case BeadUpdatedMsg:
    // Close overlay, show toast, refresh tree
case EditCancelledMsg:
    // Close overlay only
```

## Acceptance Criteria

- [ ] `e` key opens edit overlay on selected bead
- [ ] All fields pre-populated with current values
- [ ] Tab/Shift+Tab navigation works correctly
- [ ] Escape closes without saving
- [ ] Enter saves all changed fields
- [ ] Toast shows "Updated ab-xxx" on success
- [ ] Tree refreshes after successful update
- [ ] Validation: empty title shows flash red border
- [ ] Type/Priority use pill selector UI

## MVP Scope

**In Scope:**
- Edit title, description, type, priority, assignee
- Single comprehensive modal
- Basic validation (required title)
- Success/error toasts

**Out of Scope (future):**
- $EDITOR integration for multi-line fields
- Inline title editing in tree view
- Conflict detection for concurrent edits
- Edit design/acceptance_criteria/notes fields

## References

### Internal Files
- `internal/ui/overlay_create.go` - Main pattern to follow
- `internal/ui/overlay_status.go` - Simple overlay pattern
- `internal/ui/overlay_labels.go` - ComboBox pattern
- `docs/CREATE_BEAD_SPEC.md` - Detailed spec reference

### CLI Commands
```bash
# bd update supports these flags
bd update <id> --title "New Title"
bd update <id> --description "New description"
bd update <id> --priority 2
bd update <id> --assignee "username"
```

## Test Plan

- [ ] Unit tests for `EditOverlay` component
- [ ] Test field validation (empty title)
- [ ] Test Tab navigation order
- [ ] Test Escape cancellation
- [ ] Test Enter save flow
- [ ] Integration test: verify `bd update` called with correct flags
- [ ] Run `./scripts/tui-test.sh` for visual testing
