# Feature Specification: The Neural Node Creator

## 1\. Overview

The **Neural Node Creator** is a high-velocity modal designed to replace the standard "Create Bead" form. It abandons linear web-form conventions in favor of a **Heads-Up Display (HUD)** architecture.

The design prioritizes **Context** (where the bead lives) and **Flow** (entering data without friction). It is built to support the "Thought Speed" of power users, enabling rapid entry of single tasks or bulk creation of sub-trees without visual disruption.

## 2\. Design Philosophy

This feature implements the core tenets of the Abacus `UI_PRINCIPLES.md`:

  * **Scannability over Readability:** The layout uses a strict grid with distinct zones. Users can verify the state of the bead (Parent, Type, Priority) in a single glance (\< 1s).
  * **Visual Hierarchy:** The **Title** is the hero element. The **Parent Context** is the anchor. Metadata is secondary.
  * **Interaction Density:** We remove "air" and unnecessary controls (like "Post-Action" checkboxes) to keep the modal compact and the workflow linear.
  * **Modal Depth:** We use "Footer Flipping" and "Dimming" to create temporary interaction layers within the single modal view.

-----

## 3\. Visual Layout

The modal is divided into three distinct zones: **Location**, **Definition**, and **Properties**.

```text
╭──────────────────────────────────────────────────────────────────────────────╮
│  NEW BEAD (n)                                                                │
│                                                                              │
│  LOCATION (Shift+Tab to edit)                                                │
│ ╭──────────────────────────────────────────────────────────────────────────╮ │
│ │ ↳ ab-83s Create and Edit Beads in TUI                                    │ │
│ ╰──────────────────────────────────────────────────────────────────────────╯ │
│    <span style="color:gray"><i>(Dimmed Hint: Type to search parent. Backspace to clear/make Root.)</i></span>      │
│                                                                              │
│  TITLE (Hero Focus)                                                          │
│ ╭──────────────────────────────────────────────────────────────────────────╮ │
│ │ ┃                                                                        │ │
│ ╰──────────────────────────────────────────────────────────────────────────╯ │
│                                                                              │
│  PROPERTIES                                                                  │
│ ╭──────────────╮  ╭──────────────╮  ╭──────────────╮                         │
│ │ TYPE (t)     │  │ PRIORITY (p) │  │ EFFORT (e)   │                         │
│ │ [ Task    ]  │  │ [ Med     ]  │  │ [ 30m     ]  │                         │
│ │   Feature    │  │   High       │  │              │                         │
│ │   Bug        │  │   Crit       │  │              │                         │
│ │   Epic       │  │   Low        │  │              │                         │
│ │   Chore      │  │   Backlog    │  │              │                         │
│ ╰──────────────╯  ╰──────────────╯  ╰──────────────╯                         │
│                                                                              │
├──────────────────────────────────────────────────────────────────────────────┤
│ ⏎ Create   ^⏎ Create & Add Another   Tab Next Field   Esc Cancel             │
╰──────────────────────────────────────────────────────────────────────────────╯
```

### Zone 1: Location (The Anchor)

  * **Placement:** Top of the modal.
  * **Default State:** Pre-filled with the ID/Title of the currently selected node in the tree (or its parent, depending on context).
  * **Visuals:** Uses `styleStatsDim` (gray) when inactive to recede into the background. Highlights with `styleFocus` when active.
  * **Purpose:** Anchors the user. Prevents "orphaned thought" syndrome where users forget where they are adding a task.

### Zone 2: Definition (The Hero)

  * **Placement:** Center.
  * **Input:** Large text field.
  * **Visuals:** Brighter border color.
  * **Behavior:** Auto-focused on open.
  * **Smart Parsing:** (Optional) If user types `#bug`, the Type selector below automatically updates to "Bug" using green feedback text.

### Zone 3: Properties (The Deck)

  * **Layout:** Three-column grid (Type, Priority, Effort).
  * **Navigation:** Accessible via `Tab` or specific hotkeys (`t`, `p`, `e`).
  * **Visuals:** Selected items use **inverse color blocks** (background highlight) rather than just colored text, ensuring the selection state is unambiguous.

-----

## 4\. Interaction Model

### 4.1 The "Footer Contract" (Footer Flipping)

The footer is not static. It reacts to the user's focus to resolve the "Ambiguity of Intent" regarding the `Enter` key.

| Context | Footer Display | Enter Key Behavior |
| :--- | :--- | :--- |
| **Default** (Title/Props) | `⏎ Create Bead` `^⏎ Create & Add Another` | Submits form & closes modal. |
| **Location Search** | `⏎ Select Parent` `Esc Revert` | Selects highlighted parent & **Focuses Title**. |
| **Creating...** | `Generating Bead...` | Ignored. |

### 4.2 The Location Workflow (Re-Parenting)

This workflow allows users to move the bead to a different parent or make it a Root item without leaving the modal.

1.  **Activation:** User presses `Shift+Tab` from Title.
2.  **Search:** Typing initiates a fuzzy search dropdown below the input.
      * *Visual Detail:* The rest of the form (Title/Props) dims to 50% opacity to signal a "modal-within-a-modal" state.
3.  **Root Creation:** Pressing `Backspace` on an empty field sets the state to **ROOT**.
      * *Display:* `⚬ No Parent (Root Item)` in Magenta.
4.  **Revert Safety:** Below the input, a "Ghost Label" displays the original parent: `Original: ab-83s (Esc to revert)`.
5.  **Escape Hierarchy:**
      * 1st `Esc`: Reverts Location to the "Ghost/Original" value and returns focus to **Title**.
      * 2nd `Esc`: Closes the entire modal.

### 4.3 The "Create & Add Another" Flow

Power users need to dump a brain-cache of tasks rapidly.

1.  User fills Title.
2.  User presses `Ctrl+Enter`.
3.  **Instant Update:** The new bead appears in the background tree *immediately* (via manual injection).
4.  **Persistence:** The modal **does not close**.
5.  **Reset:** The Title field clears. Properties persist (e.g., if "Bug" was selected, it stays "Bug").
6.  User types the next task.

-----

## 5 Performance & Feedback Principles

### 5.1 Perceived Instantaneity

The UI must not wait for the database (`bd list`) to reload.

  * **Design Requirement:** When `Enter` is pressed, the modal must disappear (or reset) and the new node must be visible and selected in the tree within **\< 50ms**.
  * **Why:** This allows the user to immediately press `L` (Label) or `s` (Status) on the new item. Any lag breaks the "Action Chaining" workflow.

### 5.2 Error Handling

If the creation fails (backend error):

  * The modal remains open.
  * A "Toast" notification appears overlaying the top right: `Error creating bead: [Details]`.
  * The Title border turns Red.
  * The user loses no data.

## 6\. Default Keymap

  * **Navigation:** `Tab` / `Shift+Tab` (Cycle fields), `Arrows` (Select options).
  * **Focus Hotkeys:** `(t)` Type, `(p)` Priority, `(e)` Effort.
  * **Actions:**
      * `Enter`: Context-sensitive Submit/Select.
      * `Ctrl+Enter`: Force Submit & Reset (Add Another).
      * `Esc`: Revert Field or Cancel Modal.