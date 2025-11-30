# Create Bead Modal: Design Specification

> **"Neural Node Creator"** — A high-velocity modal that treats bead creation as instantiating a node in a neural network, not filling out a form.

---

## 1. Executive Summary

### The Problem

The current "Add New Bead" modal suffers from **Web Form Syndrome**: it feels like a linear HTML form stuffed into a terminal. It lacks the density, context, and speed that TUI power users expect.

### The Solution

Replace the form with a **Heads-Up Display (HUD)** architecture that prioritizes:

- **Context** — Where does this bead live? (Parent visible at top)
- **Flow** — Enter data without friction (keyboard-first, instant feedback)
- **Composability** — Create fast, enrich later (labels, description via separate commands)

### Key Capabilities

| Capability | Description |
|------------|-------------|
| Editable Parent | Change parent or create root items without leaving modal |
| Quick Root Creation | `N` (Shift+n) opens modal with no parent pre-selected |
| Quick Child Creation | `n` pre-fills parent from current selection |
| Bulk Entry | `Ctrl+Enter` creates bead and keeps modal open for next |
| Instant Feedback | New bead appears in tree in <50ms (no reload wait) |
| Footer Contract | Dynamic footer eliminates Enter key ambiguity |

---

## 2. Design Principles

These principles are derived from `UI_PRINCIPLES.md` and applied specifically to this modal.

### 2.1 Scannability Over Readability

Users should verify the bead's state (Parent, Type, Priority) in a **single glance (<1 second)**.

**Application:**
- Grid layout with distinct visual zones
- No verbose labels or instructions
- Icons and colors communicate faster than words

### 2.2 Visual Hierarchy

One thing draws the eye. Everything else supports it.

**Application:**
- **Title** is the hero (brightest border, largest input)
- **Parent** is the anchor (top position, muted when inactive)
- **Properties** are secondary (smaller, grid layout)

### 2.3 Interaction Density

Remove "air" — every pixel should earn its place.

**Application:**
- No "Post-Action" checkboxes (use composable commands instead)
- No inline help text (hints in footer only)
- Compact property grid instead of vertical list

### 2.4 Modal Depth

Create temporary interaction layers within the single view.

**Application:**
- When searching parents, the rest of the form **dims**
- Footer **flips** to show context-appropriate actions
- User always knows what Enter will do

### 2.5 Composability Over Wizards

Actions should chain. `n` → type → `Enter` → `L` should flow naturally.

**Application:**
- No labels UI in this modal (press `L` after creation)
- No description field (use detail view)
- Creation is fast; enrichment is separate

### 2.6 Perceived Instantaneity

The UI must **never** wait for database reload.

**Application:**
- Manual tree injection on create
- Background sync happens later
- User can press `L` immediately after `Enter`

---

## 3. Visual Layout

The modal is divided into five distinct **zones**.

```
╭──────────────────────────────────────────────────────────────────────────────╮
│  NEW BEAD                                                                    │
│                                                                              │
│  PARENT                                                            Shift+Tab │
│ ╭──────────────────────────────────────────────────────────────────────────╮ │
│ │ ↳ ab-83s Create and Edit Beads in TUI                                    │ │
│ ╰──────────────────────────────────────────────────────────────────────────╯ │
│                                                                              │
│  TITLE                                                                       │
│ ╭──────────────────────────────────────────────────────────────────────────╮ │
│ │ ┃                                                                        │ │
│ ╰──────────────────────────────────────────────────────────────────────────╯ │
│                                                                              │
│  PROPERTIES                                                                  │
│ ╭────────────────────╮  ╭────────────────────╮  ╭────────────────────╮       │
│ │ TYPE               │  │ PRIORITY           │  │ EFFORT             │       │
│ │ ► Task             │  │ ► Med              │  │ [          ]       │       │
│ │   Feature          │  │   High             │  │                    │       │
│ │   Bug              │  │   Critical         │  │                    │       │
│ │   Epic             │  │   Low              │  │                    │       │
│ │   Chore            │  │   Backlog          │  │                    │       │
│ ╰────────────────────╯  ╰────────────────────╯  ╰────────────────────╯       │
│                                                                              │
│  LABELS                                                                      │
│ ╭──────────────────────────────────────────────────────────────────────────╮ │
│ │ [backend] [urgent] [api] [|                                            ] │ │
│ │  ↓ to browse, type to filter                                             │ │
│ ╰──────────────────────────────────────────────────────────────────────────╯ │
│                                                                              │
│  ASSIGNEE                                                                    │
│ ╭──────────────────────────────────────────────────────────────────────────╮ │
│ │ [Unassigned                                                            ] │ │
│ │  ↓ to browse, type to filter                                             │ │
│ ╰──────────────────────────────────────────────────────────────────────────╯ │
│                                                                              │
├──────────────────────────────────────────────────────────────────────────────┤
│ Enter Create   ^Enter Create & Add Another   Tab Next   Esc Cancel           │
╰──────────────────────────────────────────────────────────────────────────────╯
```

### Zone 1: Parent (The Anchor)

| Aspect | Specification |
|--------|---------------|
| **Position** | Top of modal |
| **Default Value** | Pre-filled with currently selected bead (context-aware) |
| **Inactive Style** | `styleStatsDim` (gray) — recedes into background |
| **Active Style** | `styleFocus` (cyan border) — draws attention |
| **Purpose** | Anchors user; prevents "orphaned thought" syndrome |

**Behavior:**
- Shows bead ID + title (truncated if needed)
- Editable via `Shift+Tab` from Title
- Clearing creates a **root** bead

### Zone 2: Title (The Hero)

| Aspect | Specification |
|--------|---------------|
| **Position** | Center, prominent |
| **Style** | Bright border color (cyan/green) |
| **Focus** | Auto-focused when modal opens |
| **Validation** | Required — red border flash on empty submit |

**Behavior:**
- Large text input area
- No placeholder text needed (cursor implies "type here")
- No "required" label (submission validates)

### Zone 3: Properties (Type, Priority, Effort)

| Aspect | Specification |
|--------|---------------|
| **Layout** | Three-column horizontal grid (all fixed width) |
| **Columns** | Type, Priority, Effort |
| **Navigation** | `Tab` / `Shift+Tab` cycles columns; `↑/↓` selects within Type/Priority |
| **Purpose** | "What kind of bead, how important, how long?" — core metadata |

**Selection Visual:**
- Current option marked with `►` prefix
- Use **inverse colors** (background highlight) for focused column
- Unfocused columns show selection but with muted styling

**Single-Key Selection (Type & Priority):**

When Type or Priority is focused, single keys select directly:

| Type Focused | Key | Selects |
|--------------|-----|---------|
| | `t` | **T**ask |
| | `f` | **F**eature |
| | `b` | **B**ug |
| | `e` | **E**pic |
| | `c` | **C**hore |

| Priority Focused | Key | Selects |
|------------------|-----|---------|
| | `c` | **C**ritical |
| | `h` | **H**igh |
| | `m` | **M**edium |
| | `l` | **L**ow |
| | `b` | **B**acklog |

**Hotkey Underlines:**
- When field is **focused**: Underline the hotkey letter (e.g., "T̲ask", "F̲eature")
- When field is **unfocused**: No underlines (cleaner appearance)
- Keys are **case-insensitive** (`t` and `T` both select Task)

**Effort Field:**

| Aspect | Specification |
|--------|---------------|
| **Type** | Free-text input |
| **Formats** | `30m`, `2h`, `1.5h`, `1d` (converts to minutes internally) |
| **Optional** | Can leave blank |

### Zone 4: Labels

| Aspect | Specification |
|--------|---------------|
| **Type** | Multi-select combo box with tokenizing (chips) |
| **Data Source** | Extracted from existing beads at modal open |
| **Display** | Selected labels shown as chips: `[backend] [urgent] [|          ]` |
| **Empty State** | Shows input cursor only (no placeholder needed) |
| **Word Wrap** | Chips wrap to next line; no chip spans multiple lines |
| **New Label Toast** | When new label created: `New Label Added: [name]` |

#### Labels Combo Box Contract

The Labels field follows the same **"Selection Follows Visual Focus"** contract as Assignee, but with **multi-select tokenizing** behavior.

**Key Difference from Assignee:** After selection, the input clears and focus **stays** in the Labels field (ready for next label). Assignee is single-select; Labels is multi-select.

**Initial State:** Cursor in input area after any existing chips.

**Opening the Dropdown:**

| Action | Result |
|--------|--------|
| **`↓`** | Opens dropdown with full list (excluding already-selected labels) |
| **Start typing** | Opens dropdown with filtered list, auto-highlights first match |

**When dropdown is OPEN with a MATCH highlighted:**

| Key | Behavior | Result |
|-----|----------|--------|
| **Enter** | Select match | Creates chip, clears input, **stays in field** |
| **Tab** | Commit & move | Creates chip (if input has text), moves to Assignee |
| **↓/↑** | Navigate | Highlight different match |

**When dropdown is OPEN with NO MATCHES:**

| Key | Behavior | Result |
|-----|----------|--------|
| **Enter** | Create new | Creates chip with literal text, toast appears, clears input |
| **Tab** | Commit & move | Creates chip with literal text, moves to Assignee |

**When input is EMPTY:**

| Key | Behavior | Result |
|-----|----------|--------|
| **Enter** | Nothing | No-op (can't create empty label) |
| **Tab** | Move | Just moves to Assignee (no chip created) |
| **←** | Chip nav | Enter chip navigation mode (if chips exist) |

**The Escape Hatch (same as Assignee):**
- 1st `Esc`: Close dropdown, keep typed text (enables literal entry)
- 2nd `Esc`: Clear input, stay in field (or close modal if nothing to clear)

**Duplicate Handling:**
- Already-selected labels are **removed from dropdown** (not shown)
- If user types exact match to existing chip: flash that chip, clear input, don't add duplicate

#### Chip Navigation Mode

When input is empty, pressing `←` enters **chip navigation mode** to delete specific chips.

**Visual:**
```
Normal (input mode):
│ [backend] [urgent] [bug] [|              ] │
                            ↑ cursor here

Chip nav mode (after pressing ←):
│ [backend] [urgent] [►bug◄]                 │
                      ↑ this chip highlighted
```

**Chip Navigation Keys:**

| Key | Behavior |
|-----|----------|
| **←** | Highlight previous chip (stop at first) |
| **→** | Highlight next chip, or exit to input if past last |
| **Delete/⌫** | Remove highlighted chip, highlight next (or exit if none left) |
| **Any letter** | Exit to input mode, type that character |
| **Esc** | Exit to input mode (no deletion) |
| **Tab** | Exit to input, then move to Assignee |

**Entry condition:** `←` only enters chip mode when input is empty AND dropdown is closed.

**After deletion:** Stay in chip mode, highlight adjacent chip. If no chips remain, exit to input mode.

#### Labels Visual States

*Empty field, focused:*
```
╭──────────────────────────────────────╮
│ LABELS                               │
│ [|                                 ] │ ← Cursor ready
╰──────────────────────────────────────╯
```

*With chips, typing to filter:*
```
╭──────────────────────────────────────╮
│ LABELS                               │
│ [backend] [urgent] [sec|           ] │
│ ────────────────────────────────────── │
│ ► security                           │ ← Filtered match
│   section-a                          │
╰──────────────────────────────────────╯
```

*Chip navigation mode:*
```
╭──────────────────────────────────────╮
│ LABELS                               │
│ [backend] [►urgent◄] [bug]           │ ← "urgent" highlighted
╰──────────────────────────────────────╯
```
→ Press `Delete` to remove "urgent"

*Word wrap (many chips):*
```
╭──────────────────────────────────────╮
│ LABELS                               │
│ [backend] [frontend] [api] [urgent]  │
│ [needs-review] [v2.0] [|           ] │ ← Wraps, cursor at end
╰──────────────────────────────────────╯
```

#### Labels State Machine

```
State: INPUT (cursor in text field, after chips)
  ├─ ↓         → Open dropdown (full list minus selected)
  ├─ typing    → Filter dropdown, auto-highlight first match
  ├─ Enter     → If match: create chip, clear input, stay
  │              If no match: create new chip, toast, stay
  │              If empty: no-op
  ├─ Tab       → Create chip if input has text, move to Assignee
  ├─ ←         → If input empty: CHIP_NAV mode. Else: move cursor
  └─ Esc       → Close dropdown (1st), clear input (2nd)

State: CHIP_NAV (navigating chips)
  ├─ ←         → Highlight previous chip (stop at first)
  ├─ →         → Highlight next, or exit to INPUT if past last
  ├─ Delete/⌫  → Remove chip, highlight next (or exit if none)
  ├─ typing    → Exit to INPUT, type that character
  ├─ Tab       → Exit to INPUT, move to Assignee
  └─ Esc       → Exit to INPUT (no deletion)
```

### Zone 5: Assignee

| Aspect | Specification |
|--------|---------------|
| **Layout** | Full-width row |
| **Navigation** | `Tab` / `Shift+Tab` cycles to/from Labels |
| **Purpose** | "Who does it?" — ownership |

| Aspect | Specification |
|--------|---------------|
| **Type** | Combo box with autocomplete (allows new entries) |
| **Data Source** | Extracted from existing beads at modal open |
| **Default Options** | "Unassigned", "Me" (`$USER`), + any assignees from existing beads |
| **Default Value** | Unassigned |
| **New Assignee Toast** | When new assignee is created: `New Assignee Added: [name]` |

#### The Combo Box Contract: "Selection Follows Visual Focus"

This is a high-velocity tool. Users assign to existing team members 95% of the time. The autocomplete behavior prioritizes **speed over flexibility**.

**Initial State:** When field is focused, the dropdown is **NOT visible yet**. The cursor is ready in the input.

**Opening the Dropdown:**

| Action | Result |
|--------|--------|
| **`↓`** | Opens dropdown with **full unfiltered list**, highlights current value (or first item if no match) |
| **Start typing** | Opens dropdown with **filtered list**, auto-highlights first match |

This "on-demand" dropdown keeps the UI clean until the user needs to see options.

**Preserving Current Selection:**

When the field already has a value (e.g., "Carlos") and user presses `↓`:
- Dropdown shows full list
- "Carlos" is highlighted (scrolled into view if needed)
- User can navigate from there with `↑/↓`
- Enter/Tab confirms the highlighted item (no accidental change)

**Core Rule:** Once dropdown is open, the first match is **automatically highlighted**. Enter/Tab accept the highlighted match.

**When dropdown is OPEN with a MATCH highlighted:**

| Key | Behavior | Result |
|-----|----------|--------|
| **Enter** | Select match | Value becomes highlighted match, field confirmed |
| **Tab** | Select & wrap | Value becomes highlighted match, focus wraps to Title |
| **↓/↑** | Navigate | Highlight different match |

**When dropdown is OPEN with NO MATCHES:**

| Key | Behavior | Result |
|-----|----------|--------|
| **Enter** | Create new | Value becomes typed text, toast: "New Assignee Added" |
| **Tab** | Create & wrap | Value becomes typed text, focus wraps to Title |

**The Escape Hatch (Two-Stage Escape):**

To create "Carl" when "Carlos" exists and keeps auto-highlighting:

| Stage | Key | Result |
|-------|-----|--------|
| 1st | **Esc** | Closes dropdown, clears highlight. Input still shows "Carl" |
| 2nd | **Esc** | Reverts field to original value (e.g., "Unassigned") |

After 1st Esc, Enter/Tab accepts the literal typed value.

**Workflow Examples:**

*Focused, idle (dropdown not yet open):*
```
╭────────────────────────────────╮
│ ASSIGNEE                       │
│ [Unassigned|                 ] │ ← Cursor ready, no dropdown
╰────────────────────────────────╯
```
→ Press `↓` to browse full list, OR start typing to filter

*Browsing full list (pressed `↓` from idle, no current value):*
```
╭────────────────────────────────╮
│ ASSIGNEE                       │
│ [Unassigned                  ] │
│ ────────────────────────────── │
│ ► Unassigned                   │ ← Highlighted (first item)
│   Me (chris)                   │
│   alice                        │
│   bob                          │
│   carlos                       │
╰────────────────────────────────╯
```
→ Use `↓/↑` to navigate, `Enter/Tab` to select

*Browsing full list (pressed `↓` from idle, has existing value "carlos"):*
```
╭────────────────────────────────╮
│ ASSIGNEE                       │
│ [carlos                      ] │ ← Current value
│ ────────────────────────────── │
│   Unassigned                   │
│   Me (chris)                   │
│   alice                        │
│   bob                          │
│ ► carlos                       │ ← Highlighted (matches current)
│   carlotta                     │
╰────────────────────────────────╯
```
→ Current value is pre-highlighted; navigate from there

*Selecting existing assignee "Carlos" (Happy Path):*
```
╭────────────────────────────────╮
│ ASSIGNEE                       │
│ [Carl|                       ] │ ← User typed "Carl"
│ ────────────────────────────── │
│ ► Carlos                       │ ← Auto-highlighted (filtered)
│   Carlotta                     │
╰────────────────────────────────╯
```
→ Press `Tab` → Field becomes "Carlos", focus wraps to Title ✓

*Creating new assignee "Carl" (Escape Hatch):*
```
╭────────────────────────────────╮
│ ASSIGNEE                       │
│ [Carl|                       ] │ ← User typed "Carl"
│ ────────────────────────────── │
│ ► Carlos                       │ ← Auto-highlighted
╰────────────────────────────────╯
```
→ Press `Esc` → Dropdown closes, input still shows "Carl"
```
╭────────────────────────────────╮
│ ASSIGNEE                       │
│ [Carl|                       ] │ ← Dropdown closed
╰────────────────────────────────╯
```
→ Press `Tab` → Field becomes "Carl" (new), focus wraps to Title, toast appears ✓

*No match found:*
```
╭────────────────────────────────╮
│ ASSIGNEE                       │
│ [newperson|                  ] │
│ ────────────────────────────── │
│   No matches                   │
│   ⏎ to add new assignee        │
╰────────────────────────────────╯
```
→ Press `Enter` → Field becomes "newperson", toast: "New Assignee Added: newperson" ✓

**Design Rationale:**
This follows the VS Code IntelliSense model: **"Suggestion wins unless explicitly dismissed."** Speed is prioritized because selecting existing assignees is the 95% case.

---

## 4. Interaction Model

### 4.0 Opening the Modal

The modal can be opened from the tree view with two shortcuts:

| Key | Action | Parent Field Shows |
|-----|--------|-------------------|
| `n` | New bead as child | Pre-filled with currently selected bead |
| `N` (Shift+n) | New root bead | "◇ No Parent (Root Item)" |

This allows quick creation of both child beads (common case) and root beads (epics, top-level features) without extra steps.

### 4.1 The Footer Contract (Footer Flipping)

The footer is **not static**. It reacts to user focus to eliminate **Ambiguity of Intent**.

| Context | Footer Shows | Enter Does |
|---------|--------------|------------|
| **Title / Properties** | `Enter Create  ^Enter Create & Add Another` | Submit form, close modal |
| **Parent Search** | `Enter Select  Esc Revert` | Select parent, return to Title |
| **Creating...** | `Creating bead...` | Nothing (debounced) |

**Why This Matters:**

Users must never hesitate before pressing Enter. The footer is a **contract** that tells them exactly what will happen.

### 4.2 Parent Workflow (Re-Parenting)

This workflow allows changing parent or creating root items **without leaving the modal**.

#### State Machine

```
┌─────────────────┐
│  Default State  │ ← Modal opens here
│  (Title focus)  │
└────────┬────────┘
         │ Shift+Tab
         ▼
┌─────────────────┐
│  Parent Edit  │ ← Can type to search
│  (Parent focus)│
└────────┬────────┘
         │ Type characters
         ▼
┌─────────────────┐
│  Search Active  │ ← Dropdown visible, form dimmed
│  (Dropdown open)│
└────────┬────────┘
         │ Enter (select) or Esc (revert)
         ▼
┌─────────────────┐
│  Default State  │ ← Back to Title
│  (Title focus)  │
└─────────────────┘
```

#### Interaction Model: Combo Box with Search

The Parent field behaves like a **combo box**, not a text input. When focused, the entire current value is "selected" — your next action replaces or clears it.

**From Focused State:**

| Action | Result |
|--------|--------|
| Type any character | Opens search, character becomes first char of query |
| Delete / Backspace | Clears to "◇ No Parent (Root Item)" |
| Enter | Confirm current value, move to Title |
| Tab | Confirm current value, move to Title |
| Esc | Discard any pending changes, move to Title |
| Down Arrow | Opens search dropdown (empty query) |

**From Search Active State:**

| Action | Result |
|--------|--------|
| Continue typing | Filters search results |
| Up / Down | Navigate results |
| Enter | Select highlighted result as new parent, move to Title |
| Esc | Close search, revert to value when field was focused |

**From Cleared State ("No Parent"):**

| Action | Result |
|--------|--------|
| Type any character | Opens search from cleared state |
| Enter | Confirm "No Parent", move to Title |
| Esc | Revert to original parent, move to Title |

**Key behaviors:**
- **No manual text deletion** — you're selecting a value, not editing text
- **Typing immediately searches** — no need to clear first
- **Delete/Backspace = "I want no parent"** — clear intent
- **All changes are pending** until you leave via Enter or Tab
- **Esc discards pending changes** — reverts to state when you focused the field

#### Creating Root Items

1. Focus Parent field (Shift+Tab or Up from Title)
2. Press Delete or Backspace → Shows "◇ No Parent (Root Item)" in magenta
3. Press Enter → Confirms "No Parent", moves to Title
4. Complete the form and submit → Bead is created as a root node

#### Visual Feedback

- **Focused**: Current parent appears "selected" (highlighted background)
- **Search Active**: Dropdown visible, rest of form dims to 50% opacity
- **Cleared**: Shows "◇ No Parent (Root Item)" in magenta

#### Escape Hierarchy

| State | Esc Does |
|-------|----------|
| Search dropdown open | Close search, revert to pre-search value |
| Parent focused (cleared or unchanged) | Revert any pending change, move to Title |
| Title / Properties focused | Close entire modal |

### 4.3 Create & Add Another (Bulk Entry)

Power users need to dump a brain-cache of tasks rapidly.

#### Flow

1. User fills Title
2. User presses `Ctrl+Enter`
3. Bead created, appears in tree **immediately**
4. Modal **stays open**
5. Title **clears**
6. Properties **persist** (Type, Priority remain selected)
7. User types next task
8. Repeat until done, then press `Esc`

#### Field Persistence on Bulk Create

| Field | On Ctrl+Enter | Rationale |
|-------|---------------|-----------|
| Parent | Persists | Adding siblings to same parent |
| Title | **Clears** | Must be unique per bead |
| Type | Persists | Batch often same type |
| Priority | Persists | Batch often same priority |
| Labels | Persists | Batch context (e.g., all "backend") |
| Assignee | Persists | Batch often same assignee |
| Effort | **Clears** | Each task has unique effort |

**Why most persist:** If you're adding 5 backend tasks to an epic, you don't want to set Type, Priority, Labels, and Assignee five times. These are "batch context."

**Why Effort clears:** Each task likely has different time estimates.

### 4.4 Validation & Error Handling

#### Empty Title

1. User presses Enter with empty Title
2. Title border flashes **red**
3. Focus stays on Title
4. No toast needed (obvious from visual)

#### Backend Error

1. User presses Enter
2. Backend returns error
3. Modal **stays open** (no data lost)
4. Toast appears: `Error: [details]`
5. Title border turns red
6. User can retry or cancel

---

## 5. Type Auto-Inference

The modal automatically suggests a bead type based on title keywords. This reduces friction for common patterns while remaining non-intrusive.

### Behavior

1. User types in the Title field
2. As they type, the system scans for keyword patterns
3. If a pattern matches **and** Type has not been manually changed, the Type selector updates automatically
4. Visual feedback: Type selection animates/highlights briefly to show it changed
5. User can override at any time by manually selecting a different Type

### Keyword Patterns

| Title Contains | Inferred Type |
|----------------|---------------|
| "Fix", "Broken", "Bug", "Error", "Crash", "Issue with" | **Bug** |
| "Add", "Implement", "Create", "Build", "New" | **Feature** |
| "Refactor", "Clean up", "Reorganize", "Simplify", "Extract" | **Chore** |
| "Update", "Upgrade", "Bump", "Migrate" | **Chore** |
| "Document", "Docs", "README" | **Chore** |

### Rules

- **Case-insensitive**: "fix" and "Fix" both trigger Bug
- **Word boundary aware**: "Prefix" should not trigger (contains "fix" but not as a word)
- **First match wins**: If title contains both "Fix" and "Add", use the first pattern found
- **Manual override respected**: Once user manually changes Type, auto-inference stops for that session
- **Non-blocking**: Inference happens on keystroke, not on submit
- **New Items Only**: Inference is only used when creating new Beads, not editing existing ones.

### Visual Feedback

When type auto-changes:
- The Type selector briefly highlights (subtle pulse or border flash)
- This signals "I understood your intent" without being disruptive
- No toast or modal interruption

### Edge Cases

| Scenario | Behavior |
|----------|----------|
| Title: "Fix the Add button" | Infers **Bug** (first match: "Fix") |
| Title: "Adding fix for login" | Infers **Feature** (first match: "Adding") |
| User manually selects Epic, then types "Fix..." | Stays **Epic** (manual override) |
| User clears title and retypes | Auto-inference re-enabled |

---

## 6. Keyboard Reference

### Navigation

| Key | Action |
|-----|--------|
| `Tab` | Next field (Title → Type → Priority → Effort → Labels → Assignee) |
| `Shift+Tab` | Previous field (from Title goes to Parent) |
| `↑` / `k` | Previous option in current selector |
| `↓` / `j` | Next option in current selector |
| `←` / `h` | Previous column (in Properties zone) |
| `→` / `l` | Next column (in Properties zone) |

### Actions

| Key | Context | Action |
|-----|---------|--------|
| `Enter` | Title / Selectors / Effort | Create bead, close modal |
| `Enter` | Parent focused | Confirm parent value, move to Title |
| `Enter` | Parent dropdown | Select highlighted result, move to Title |
| `Enter` | Labels (match highlighted) | Create chip, clear input, stay in Labels |
| `Enter` | Labels (no match) | Create new label chip, toast, stay in Labels |
| `Enter` | Labels (empty input) | No-op |
| `Ctrl+Enter` | Any | Create bead, stay open (Add Another) |
| `Delete` / `Backspace` | Parent focused | Clear to "No Parent (Root)" |
| `Delete` / `Backspace` | Labels chip nav | Delete highlighted chip |
| `Esc` | Parent (any state) | Discard pending changes, move to Title |
| `Esc` | Labels (1st) | Close dropdown, keep typed text |
| `Esc` | Labels (2nd) | Clear input |
| `←` | Labels (input empty) | Enter chip navigation mode |
| `→` | Labels (chip nav) | Next chip, or exit to input if past last |
| `↓` | Labels (idle, no dropdown) | Open dropdown with full list (minus selected) |
| `↑/↓` | Labels (dropdown open) | Navigate autocomplete results |
| `Tab` | Labels (match/text) | Create chip if input has text, move to Assignee |
| `Tab` | Labels (empty) | Just move to Assignee |
| `Esc` | Assignee (1st) | Close dropdown, keep typed text (enables literal entry) |
| `Esc` | Assignee (2nd) | Revert to original value |
| `Esc` | Title / Other fields | Close modal (cancel) |
| `↓` | Assignee (idle, no dropdown) | Open dropdown with full unfiltered list |
| `↑/↓` | Assignee (dropdown open) | Navigate autocomplete results |
| `Enter` | Assignee (match highlighted) | Select highlighted match |
| `Enter` | Assignee (no match) | Create new assignee with typed text |
| `Tab` | Assignee (match highlighted) | Select match, wrap to Title |
| `Tab` | Assignee (no match/dismissed) | Use typed text, wrap to Title |

### The Safety Valve

`Ctrl+Enter` **always** submits, regardless of focus. This trains power users to rely on it when they want certainty.

---

## 7. Use Cases

### 7.1 Quick Single Task

**Scenario:** Add a task under the current epic

1. User hovers "Login Feature" in tree
2. Presses `n`
3. Modal opens, Parent shows "Login Feature"
4. Types "Implement password reset"
5. Presses `Enter`
6. Modal closes
7. New task highlighted in tree
8. Presses `L` to add labels

**Time:** ~3 seconds

### 7.2 Bulk Subtask Entry

**Scenario:** Break down an epic into tasks

1. User hovers "API Refactor" epic
2. Presses `n`
3. Types "Move auth to middleware"
4. Presses `Ctrl+Enter` (create, stay open)
5. Types "Update route handlers"
6. Presses `Ctrl+Enter`
7. Types "Add integration tests"
8. Presses `Enter` (create, close)

**Result:** 3 tasks created in ~10 seconds

### 7.3 Create Root Epic (Quick Method)

**Scenario:** Add a new top-level epic using the `N` shortcut

1. User presses `N` (Shift+n) from anywhere in tree
2. Modal opens with Parent already showing "◇ No Parent (Root Item)"
3. Types "Q4 Infrastructure Overhaul"
4. Tabs to Type, selects "Epic"
5. Presses `Enter`

**Result:** New root epic created in ~3 seconds

### 7.3b Create Root Epic (Alternative Method)

**Scenario:** Convert to root during creation (if you started with `n`)

1. User presses `n` (modal opens with parent pre-filled)
2. Presses `Shift+Tab` to focus Parent
3. Presses `Delete` to clear
4. Parent shows "◇ No Parent (Root Item)"
5. Presses `Enter` to confirm, moves to Title
6. Types title, sets Type to Epic
7. Presses `Enter`

**Result:** New root epic created (slightly longer path)

### 7.4 Re-Parent During Creation

**Scenario:** Started creating under wrong parent

1. User presses `n` on "Frontend Tasks" (wrong parent)
2. Types "Fix API rate limiting" (realizes mistake)
3. Presses `Shift+Tab` to focus Parent
4. Types "backend" to search
5. Selects "Backend Infrastructure" from dropdown
6. Presses `Enter` (selects parent, returns to Title)
7. Title still contains "Fix API rate limiting" (preserved!)
8. Presses `Enter` to create

**Result:** Task created under correct parent, no retyping

### 7.5 Quick Bug Report (with Auto-Inference)

**Scenario:** Log a bug fast — auto-inference handles the type

1. User presses `n`
2. Types "Fix login crash on Safari"
3. Type selector **automatically changes to Bug** (inferred from "Fix" + "crash")
4. User notices the Type highlight briefly
5. Presses `Enter`

**Result:** Bug created in ~2 seconds (no manual type selection needed)

### 7.6 Override Auto-Inference

**Scenario:** Auto-inference guessed wrong

1. User presses `n`
2. Types "Fix the deployment pipeline"
3. Type auto-infers to **Bug**
4. User actually wants **Chore** (this is maintenance, not a bug)
5. Presses `t`, selects "Chore"
6. Presses `Enter`

**Result:** Chore created; auto-inference disabled for rest of this modal session

### 7.7 Assign to New Team Member

**Scenario:** Assigning to someone not yet in the system

1. User presses `n`
2. Types "Review security audit report"
3. Tabs to Assignee field
4. Types "sarah" (not in existing assignees)
5. Dropdown shows "No matches" with hint to add
6. Presses `Enter` to confirm
7. Assignee set to "sarah"
8. Tabs to complete other fields, presses `Enter`
9. Bead created
10. Toast appears: `New Assignee Added: sarah`

**Result:** Task assigned to new team member; toast confirms the addition

### 7.8 Select Existing Assignee with Autocomplete

**Scenario:** Quick assignment to known team member

1. User presses `n`
2. Types "Update documentation"
3. Tabs to Assignee field
4. Types "al" to filter
5. Dropdown shows "alice" as highlighted match
6. Presses `Enter` (selects "alice", creates bead)

**Result:** Task assigned to existing team member in ~3 keystrokes

**Note:** Both `Enter` and `Tab` select the highlighted match. Since Assignee is the last field, `Enter` is typically used to confirm and create.

### 7.9 Create Assignee That's Substring of Existing (Escape Hatch)

**Scenario:** Create "Carl" when "Carlos" already exists

1. User presses `n`
2. Types "Onboard new hire"
3. Tabs to Assignee field
4. Types "Carl"
5. Dropdown shows "Carlos" auto-highlighted (partial match)
6. User wants literal "Carl", presses `Esc` (dropdown closes, input still "Carl")
7. Presses `Enter` (accepts literal "Carl", creates bead)
8. Toast appears: `New Assignee Added: Carl`

**Result:** New assignee created despite partial match with existing name

### 7.10 Add Multiple Labels Quickly

**Scenario:** Tag a task with several labels

1. User presses `n`
2. Types "Implement OAuth2 login"
3. Tabs to Labels field
4. Types "auth" → "auth" highlighted → presses `Enter`
5. Chip `[auth]` appears, input clears, focus stays
6. Types "back" → "backend" highlighted → presses `Enter`
7. Chip `[backend]` appears
8. Types "sec" → "security" highlighted → presses `Tab`
9. Chip `[security]` appears, focus moves to Assignee

**Result:** Three labels added in rapid succession with minimal keystrokes

### 7.11 Create New Label

**Scenario:** Add a label that doesn't exist yet

1. User is in Labels field
2. Types "needs-review" (doesn't exist in system)
3. Dropdown shows "No matches"
4. Presses `Enter`
5. Chip `[needs-review]` created
6. Toast appears: `New Label Added: needs-review`

**Result:** New label created on-the-fly

### 7.12 Remove a Label (Chip Navigation)

**Scenario:** Remove a mistakenly added label

1. User has chips: `[auth] [backend] [frontend]`
2. Input is empty
3. Presses `←` to enter chip nav mode
4. `[frontend]` is highlighted (last chip)
5. Presses `←` → `[backend]` highlighted
6. Presses `Delete` → `[backend]` removed
7. `[frontend]` now highlighted (next chip)
8. Presses `→` → exits to input mode
9. Remaining chips: `[auth] [frontend]`

**Result:** Specific label removed without affecting others

---

## 8. Visual Styling

### Color Assignments

| Element | Style | Purpose |
|---------|-------|---------|
| Modal border | Default | Neutral container |
| Parent (inactive) | `styleStatsDim` | Recedes, not distracting |
| Parent (active) | `styleFocus` | Shows it's editable |
| Title border | Bright cyan | Hero element, draws eye |
| Title border (error) | Red flash | Validation failed |
| Property column (focused) | Inverse bg | Shows current column |
| Property column (unfocused) | Normal | Muted but readable |
| Selected option | `►` prefix | Clear selection indicator |
| Root indicator | Magenta | Distinct from normal parents |
| Ghost parent | `styleStatsDim` | Hint, not primary |
| Search results | Normal | List items |
| Search highlight | Inverse bg | Current selection |
| Label chips | `[name]` brackets | Show selected labels |
| Label chip (highlighted) | `[►name◄]` inverse bg | Chip nav mode selection |
| Labels input | Cursor after chips | Ready for typing |
| Labels dropdown | Normal | Filtered results (minus selected) |
| Labels no-match hint | `styleStatsDim` | "⏎ to add new label" |
| Assignee (unassigned) | Muted text | Default state |
| Assignee (assigned) | Normal text | Has assignee |
| Assignee input | Inverse bg cursor | Typing mode |
| Assignee match | Normal | Filtered results |
| Assignee no-match hint | `styleStatsDim` | "⏎ to add new assignee" |
| New Assignee toast | Green accent | Confirms new addition |

### Dimming for Modal Depth

When Parent search is active:
- Title zone: 50% opacity
- Properties zone: 50% opacity
- Parent zone: 100% opacity (focused)
- Footer: Updated text (flipped)

This creates visual **layering** within the single modal.

---

## 9. Explicit Exclusions

These features are **intentionally omitted** to keep the modal fast and focused.

| Feature | Reason | Alternative |
|---------|--------|-------------|
| Description field | Takes space, rarely needed immediately | Use detail view |
| Design notes | Advanced field | Use detail view |
| Acceptance criteria | Advanced field | Use detail view |
| "Edit after create" checkbox | Unnecessary | Just press the key after |
| Blockers/Dependencies | Complex relationships | Use `bd dep` after |

**Philosophy:** This modal creates the **node** with essential metadata (Type, Priority, Labels, Assignee). Advanced enrichment (description, design notes, dependencies) happens after via composable commands.

---

## 10. Success Criteria

| Metric | Target | Rationale |
|--------|--------|-----------|
| Time to create simple task | < 3 seconds | Faster than opening a web form |
| Time to create 5 subtasks | < 15 seconds | Bulk entry must be efficient |
| Perceived latency | < 50ms | Enables action chaining |
| Enter key hesitation | Zero | Footer contract eliminates ambiguity |
| Escape to safe state | Always possible | User never feels trapped |
| Learning curve | < 2 minutes | Discoverable via footer hints |

---

## 11. Optional Enhancements (Future)

These are explicitly **out of scope** for v1 but noted for consideration.

### Smart Title Parsing

Parse hashtags in title for power users:

| Input | Result |
|-------|--------|
| `Fix login #bug` | Title: "Fix login", Type: Bug |
| `Add dark mode #feature` | Title: "Add dark mode", Type: Feature |

### Recent Parents

Quick access to recently-used parent beads (last 5).

### Templates

Pre-fill properties based on parent type or user-defined templates.

---

## 12. Implementation Notes

> These are guidance for implementation, not strict requirements.

### Performance: Manual Tree Injection

To achieve <50ms perceived latency:

1. User presses Enter
2. Run `bd create --json ...` (returns full JSON of created bead)
3. Parse JSON response to get the **actual stored data**
4. **Do NOT** call full tree refresh
5. Construct node from the returned JSON (not from input data)
6. Insert into parent's children, re-sort
7. Set cursor to new node
8. Return immediately
9. Background: trigger tree refresh (eventually consistent)

**Why use `--json` flag:**
- Returns the exact data stored in the database
- Handles any server-side defaults or transformations
- Guarantees the displayed node matches the actual bead
- Avoids sync issues if input values were normalized/rejected
- Single command — no need for separate `bd show` call

**Example response from `bd create --json`:**
More fields are returned if they are populated.
```json
{"id":"ab-7c9","title":"Fix login bug","description":"","status":"open","priority":2,"issue_type":"bug","created_at":"2025-11-29T17:00:13Z","updated_at":"2025-11-29T17:00:13Z"}
```

### State Management

```
Focus: Parent | Title | Type | Priority | Effort | Labels | Assignee
ParentID: string (empty = Root, committed value)
ParentPending: string (pending value while in Parent field)
ParentOriginal: string (value when Parent field was focused, for Esc revert)
SearchQuery: string
SearchResults: []Bead
SelectedType: enum
SelectedPriority: enum
TypeManuallySet: bool (disables auto-inference when true)

// Labels (multi-select combo box)
SelectedLabels: []string (chips already added)
LabelsInput: string (current typed value)
LabelsOptions: []string (all available labels from existing beads)
LabelsFilteredMatches: []string (matches minus already-selected)
LabelsHighlightedIndex: int (index of highlighted match)
IsLabelsDropdownOpen: bool
LabelsChipNavMode: bool (true when navigating chips with arrows)
LabelsChipNavIndex: int (which chip is highlighted, -1 = none)

// Assignee (single-select combo box)
SelectedAssignee: string (empty = Unassigned)
AssigneeInput: string (current typed value)
AssigneeOriginal: string (value when field was focused, for 2nd Esc revert)
AssigneeOptions: []string (extracted from existing beads at modal open)
AssigneeFilteredMatches: []string (matches for current input)
AssigneeHighlightedIndex: int (index of highlighted match, 0 = first)
IsAssigneeDropdownOpen: bool (false after 1st Esc)

EffortInput: string
IsSearching: bool (Parent search active)
IsCreating: bool
```

**Note:** `ParentOriginal` stores the value when the Parent field was focused. Esc reverts to this. Enter/Tab commits `ParentPending` to `ParentID`.

**Assignee Combo Box State Machine:**

```
State: IDLE (focused, dropdown closed)
  ├─ ↓        → BROWSING (open dropdown with full list)
  ├─ typing   → FILTERING (open dropdown with filtered list)
  ├─ Tab      → Keep current value, wrap to Title
  ├─ Enter    → Keep current value, confirm field
  └─ Esc      → (depends on context, may close modal)

State: BROWSING (dropdown open, full list)
  ├─ ↓/↑     → Navigate list, move highlight
  ├─ typing  → FILTERING (switch to filtered mode)
  ├─ Enter   → Select highlighted item
  ├─ Tab     → Select highlighted item, wrap to Title
  └─ Esc     → IDLE (close dropdown)

State: FILTERING (dropdown open, filtered list)
  ├─ ↓/↑     → Navigate filtered results
  ├─ typing  → Update filter, re-highlight first match
  ├─ Enter   → Select highlighted match (or literal if no matches)
  ├─ Tab     → Select highlighted match (or literal), wrap to Title
  └─ Esc     → IDLE (close dropdown, keep typed text) [1st Esc]
              → Revert to original [2nd Esc from IDLE]
```

If final value not in `AssigneeOptions`, it's a new assignee → show toast.

### Message Types

```
BeadCreatedMsg { ID, ParentID }
BeadCreateErrorMsg { Error }
ParentSearchResultsMsg { Results }
NewLabelAddedMsg { Name }    // triggers toast: "New Label Added: [Name]"
NewAssigneeAddedMsg { Name } // triggers toast: "New Assignee Added: [Name]"
CreateModalCancelledMsg {}
```

---

## Appendix A: Comparison to Current Implementation

| Aspect | Current | Neural Node Creator |
|--------|---------|---------------------|
| Layout | Linear vertical form | HUD with zones |
| Parent display | Static text at bottom | Editable field at top |
| Root items | Unclear how to create | `N` shortcut or Backspace in Parent |
| Labels | Set after creation | Inline in modal, persists for bulk |
| Assignee | Set after creation | Inline in modal, persists for bulk |
| Feedback | Reload-based | Instant injection |
| Bulk entry | Not supported | Ctrl+Enter workflow |
| Enter key | Always submit | Context-dependent |
| Footer | Static help text | Dynamic (flips) |
| Properties | Vertical list | Three-row layout (Type/Priority/Effort, Labels, Assignee) |

---

## Appendix B: Design References

- `UI_PRINCIPLES.md` — Core design guidelines
- `BEAD_MODEL.md` — Data model specification
- Gemini Design Session — Original collaborative design exploration
