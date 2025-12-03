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

### 5. ASCII over Unicode for Selection States

For checkboxes in lists, prefer ASCII:
- `[ ]` / `[x]` - Clear at any size, any terminal font
- Avoid `☐` / `☑` - Too small to distinguish in many fonts

### 6. Cursor vs Selection State

When lists have both a cursor AND checked/selected items:
- **Cursor**: Background highlight (shows where you ARE)
- **Selection state**: Foreground color change (shows what's CHECKED)

Never use the same color for both - it creates confusion.

### 7. Color-Coded Feedback

Use universal color associations for instant comprehension:
- **Green** (`styleLabelChecked`): Additions, positive changes (`+label`)
- **Red** (`styleBlockedText`): Removals, negative changes (`-label`)
- **Gray** (`styleStatsDim`): Neutral/supporting info

---

## Surface + Layering System

All UI rendering now flows through the Canvas/Surface pipeline described in `docs/SURFACE_RENDERING_REDESIGN.md`.

1. **Pick a surface** – `NewPrimarySurface(width, height)` for base/toasts, `NewSecondarySurface` for overlays. Each surface fills the backing canvas with the correct theme background instantly.
2. **Draw once** – call `surface.Draw(x, y, block)` with Lip Gloss output. Every style already includes the baked background, so legacy `fillBackground*` helpers should never show up in new code.
3. **Return a Layer** – set offsets on the canvas and return it via `LayerFunc` so the view compositor can stack it along with dimming, overlays, and toasts.

Example overlay:
```go
surf := NewSecondarySurface(popupWidth, popupHeight)
surf.Draw(0, 0, popupContents)
return LayerFunc(func() *Canvas {
    surf.Canvas.SetOffset(centerX, centerY)
    return surf.Canvas
})
```

Toast helpers follow the same pattern via `newToastLayer`: draw onto a primary surface, then let the compositor place it near the bottom-right. No manual `"   "` padding or `lipgloss.Place` gap fillers are required.

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
| Tree view | `↑↓` Navigate, `←→` Expand, `s` ✎ Status, `L` Labels |
| Details view | `↑↓` Scroll |
| Status overlay | `o` Open, `i` In Progress, `c` Close, `esc` Cancel |
| Labels overlay | `↑↓` Navigate, `Space` Toggle, `⏎` Apply, `esc` Cancel |

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

### Escape Key Hierarchy

When overlays have filter/input fields, Escape should work in stages:
1. **First Escape**: Clear the filter/input if populated
2. **Second Escape**: Close the overlay

This is intuitive - users expect Escape to "undo" the most recent action first.

### Hotkey Conflict Avoidance

- **Preserve vim `hjkl` navigation** - It's sacred in TUI apps
- **Use Shift variants** when lowercase conflicts (e.g., `L` for Labels since `l` is vim-right)
- **Avoid "second consonant" mnemonics** (e.g., `b` for la**B**els) - clever but not guessable

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

## Testing Guardrails

- **Golden snapshots**: `TestOverlayAndToastGoldenSnapshots` captures Dracula/Solarized/Nord overlays and toasts straight from the cell buffer. Intentionally refresh files with `go test ./internal/ui -run TestOverlayAndToastGoldenSnapshots -update-golden`.
- **Integration guard**: `TestViewOmitsDefaultResetGaps` fails anytime `App.View()` emits the old `\x1b[0m ` pattern, preventing regressions to the string post-processing era.

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
