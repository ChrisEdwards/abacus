# Slack-Style Submit for Multiline Text Fields

**Bead:** ab-b5lw
**Status:** Design approved
**Date:** 2025-12-25

## Problem

Ctrl+Enter is broken in Ghostty terminal. Ghostty uses the Kitty keyboard protocol which sends `CSI 13;5u` for Ctrl+Enter, but bubbletea v1.x doesn't recognize this sequence. The current code checks `msg.String() == "ctrl+enter"` which never matches in Ghostty.

This particularly affects the **CommentOverlay** which only has a multiline textarea - users cannot submit comments because:
- Enter inserts newlines (expected textarea behavior)
- Ctrl+S doesn't work reliably in Ghostty
- There's no other field to Tab to

## Solution

Adopt Slack/Discord-style keybindings for multiline text fields:

| Key | Action |
|-----|--------|
| Enter | Submit the form |
| Shift+Enter | Insert newline |

This pattern is widely understood and works reliably across all terminals.

## Scope

**Affected components:**
1. `overlay_comment.go` - CommentOverlay textarea
2. `overlay_create.go` - Description field only

**Not affected:**
- Single-line fields (title, parent, labels, assignee) - Enter already submits
- Type/Priority selectors - not text input

## Implementation

### 1. CommentOverlay (`internal/ui/overlay_comment.go`)

Change the Update handler:

```go
case tea.KeyMsg:
    switch msg.Type {
    case tea.KeyEsc:
        // ... existing escape handling ...

    case tea.KeyEnter:
        // Shift+Enter inserts newline, plain Enter submits
        if msg.String() == "shift+enter" {
            // Let textarea handle it (insert newline)
            break
        }
        return m.submit()
    }
```

Update footer hints: `⏎ Save` instead of `^S Save` / `⌘S Save`

### 2. CreateOverlay (`internal/ui/overlay_create.go`)

In the KeyEnter handler, change description field behavior:

```go
case tea.KeyEnter:
    if m.isCreating {
        return m, nil
    }

    // In description field: Shift+Enter for newline, Enter for submit
    if m.focus == FocusDescription {
        if msg.String() == "shift+enter" {
            // Let textarea handle newline
            return m.handleZoneInput(msg)
        }
        // Plain Enter submits
        return m.handleSubmit(false)
    }

    // Remove ctrl+enter bulk mode check (broken in Ghostty, rarely used)

    // Regular Enter submits if not in dropdown
    if !m.isAnyDropdownOpen() && m.labelsCombo.combo.Value() == "" {
        return m.handleSubmit(false)
    }
```

### 3. Remove Bulk Mode

Remove the `ctrl+enter` check and `stayOpen` logic:
- Delete `msg.String() == "ctrl+enter"` check
- Remove `handleSubmit(true)` calls (bulk mode)
- Update footer hints to remove `^⏎ Create+Add`

### 4. Update Footer Hints

**CommentOverlay:**
- Before: `^S Save` / `⌘S Save`
- After: `⏎ Save`

**CreateOverlay:**
- Before: `⏎ Create`, `^⏎ Create+Add`
- After: `⏎ Create` (remove bulk hint)

## Testing

1. Verify Enter submits from CommentOverlay
2. Verify Enter submits from CreateOverlay description field
3. Verify Shift+Enter inserts newlines in both
4. Verify existing Enter behavior unchanged for other fields
5. Test in both Ghostty and iTerm2

## Files to Modify

- `internal/ui/overlay_comment.go`
- `internal/ui/overlay_comment_test.go`
- `internal/ui/overlay_create.go`
- `internal/ui/overlay_create_view.go` (footer hints)
- `internal/ui/overlay_create_workflow_test.go`
- `internal/ui/overlay_create_bulk_test.go` (remove or update)
- `internal/ui/update_messages.go` (remove stayOpen if unused)
