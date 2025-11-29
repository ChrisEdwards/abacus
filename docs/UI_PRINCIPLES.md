# UI Design Principles

This document captures the design principles established for the Abacus TUI. Follow these guidelines when building new UI features to ensure consistency.

## Core Philosophy

**Scannability over readability.** Users should understand UI elements in under 1 second. Use icons and colors to communicate faster than words.

---

## Visual Design

### 1. Visual Hierarchy

- **Hero message** prominent (e.g., new status in color)
- **Supporting info** muted (e.g., bead ID, countdown timer)
- Single focal point - one thing draws the eye

### 2. Consistent Visual Language

Reuse existing icons from the tree view:
- `○` Open (white)
- `◐` In Progress (neon green icon, cyan text)
- `✔` Closed (gray)
- `⛔` Blocked (red)

Use the `statusPresentation()` helper to get icon + styles for any status.

### 3. Consistent Styling

Reuse existing styles - don't create new ones unless necessary:
- `styleID` (gold, bold) - bead IDs everywhere
- `styleStatsDim` (gray) - secondary/muted info
- `styleHelpKey` (cyan, bold) - hotkeys
- `styleHelpDesc` (light gray) - hotkey descriptions

### 4. Icons for Clarity

Use icons to clarify actions:
- `✎` for edit/change actions
- `↑↓` for navigation
- `←→` for expand/collapse

---

## Component Patterns

### Toast Notifications

- **Duration**: 7 seconds with visible countdown `[5s]`
- **Format**: Hero line (action + result) + supporting line (bead ID)
- **Auto-refresh**: Trigger data refresh after successful action
- **Error handling**: Show error toast on failure

Example:
```
╭─────────────────────────────╮
│  Status → ◐ In Progress     │  ← hero: icon + label in color
│  ab-6s4               [5s]  │  ← supporting: bead ID + countdown
╰─────────────────────────────╯
```

### Overlay/Popup Design

- **Breadcrumb header**: `ab-6s4 › Status` (subject › field being edited)
- **No verbose titles**: Skip "CHANGE STATUS" - the options make it obvious
- **Hotkeys in footer**: Not inline in the popup
- **Disable invalid options**: Gray out, skip in navigation

Example:
```
╭──────────────────────╮
│  ab-6s4 › Status     │  ← breadcrumb header
│  ────────────────    │
│    ○ Open            │
│    ◐ In Progress  ←  │  ← current selection
│    ○ Closed          │
╰──────────────────────╯
```

### Context-Aware Footer

The footer changes based on active context:

| Context | Hints |
|---------|-------|
| Tree view | `↑↓` Navigate, `←→` Expand, `s` ✎ Status |
| Details view | `↑↓` Scroll |
| Status overlay | `o` Open, `i` In Progress, `c` Close, `esc` Cancel |

Implementation: Check `m.activeOverlay` or `m.focus` in `renderFooter()`.

---

## Interaction Patterns

### Hotkey Design

- **Mnemonic**: `o` = Open, `i` = In progress, `c` = Close
- **Discoverable**: Show in footer
- **Quick workflows**: Enable chaining like `si` (open popup + select in-progress)
- **Consistent**: Match popup hotkeys to direct action keys

### Navigation in Overlays

- `j/k` or `↑/↓` to navigate options
- **Skip disabled options** in navigation
- `Enter` to confirm
- `Esc` to cancel (always)

### Status Transitions

- Gray out invalid transitions
- Special case: `closed → open` uses `Reopen()` method, not `UpdateStatus()`
- Allow reopening from the status overlay

---

## Code Patterns

### File Organization

- Separate file per overlay: `overlay_status.go`, `overlay_labels.go`, etc.
- Message types for overlay communication: `StatusChangedMsg`, `StatusCancelledMsg`

### Helper Functions

Create helpers for repeated patterns:
```go
// statusPresentation returns icon, icon style, and text style for a status
func statusPresentation(status string) (string, lipgloss.Style, lipgloss.Style)

// formatStatusLabel converts "in_progress" to "In Progress"
func formatStatusLabel(status string) string
```

### Client Interface

- Use semantic methods: `Close()`, `Reopen()` not just `UpdateStatus()`
- Every interface method needs a mock implementation
- Return consistent message types (e.g., `statusUpdateCompleteMsg`)

---

## Design Process

1. **ULTRATHINK before implementing** - analyze options deeply
2. **Question every element** - "Do we need this? Does it add value?"
3. **Start with user need** - What do they actually need to know?
4. **Iterate on feedback** - Refine based on real usage
5. **Test behavior changes** - Update tests when logic changes

---

## Checklist for New UI Features

- [ ] Reuses existing styles (don't create new ones unnecessarily)
- [ ] Reuses existing icons where applicable
- [ ] Has context-aware footer hints
- [ ] Hotkeys are mnemonic and discoverable
- [ ] Invalid options are disabled and skipped in navigation
- [ ] Toast shows result with countdown
- [ ] Auto-refreshes data after successful action
- [ ] Has error handling with error toast
- [ ] Tests updated for new behavior
