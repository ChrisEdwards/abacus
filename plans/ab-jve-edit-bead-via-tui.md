# Edit Bead via TUI (ab-jve)

## Overview

Add the ability to edit existing beads directly from the TUI using a full modal form. Press `e` on a selected bead to open the edit form with pre-populated values.

**Prerequisites**: ab-w0r (Create bead via TUI), ab-1hxl, ab-4fun
**Blocks**: ab-gtm (Add/remove/change parent of a bead)

---

## Design Decisions (Post-Review)

Based on reviewer feedback and CLI testing:

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Change detection | **Removed** | `bd update` is idempotent - handles unchanged values |
| Label handling | **`--set-labels`** | Flag exists, atomic, replaces manual diffing |
| Factory pattern | **Two factories** | Clearer intent at call sites |
| "No changes" toast | **Removed** | Can't detect without diffing; update is fast anyway |

**Key insight:** `bd update ab-jve --title "same title"` succeeds without error. No need for client-side diffing.

---

## Implementation Plan (3 Phases)

### Phase 1: Add Edit Mode to CreateOverlay

#### 1.1 Add `editingBead` field

**File**: `internal/ui/overlay_create.go`

```go
type CreateOverlay struct {
    editingBead *beads.FullIssue  // nil = create mode, non-nil = edit mode
    // ... existing fields unchanged
}

func (m *CreateOverlay) isEditMode() bool {
    return m.editingBead != nil
}
```

#### 1.2 Add `NewEditOverlay` factory

**File**: `internal/ui/overlay_create.go`

```go
// NewEditOverlay creates a CreateOverlay pre-populated with existing bead data.
func NewEditOverlay(bead *beads.FullIssue, opts CreateOverlayOptions) *CreateOverlay {
    m := NewCreateOverlay(opts)
    m.editingBead = bead

    // Pre-populate form fields
    m.titleInput.SetValue(bead.Title)
    m.descriptionInput.SetValue(bead.Description)
    m.typeIndex = typeIndexFromString(bead.IssueType)
    m.priorityIndex = bead.Priority
    m.typeManuallySet = true  // Disable auto-inference for edits

    // Pre-select parent if exists
    if bead.ParentID != "" {
        for _, p := range opts.AvailableParents {
            if p.ID == bead.ParentID {
                m.parentCombo.SetValue(p.Display)
                m.parentOriginal = p.Display
                break
            }
        }
    }

    // Pre-select labels
    for _, label := range bead.Labels {
        m.labelsCombo.AddChip(label)
    }

    // Pre-select assignee
    if bead.Assignee != "" {
        m.assigneeCombo.SetValue(bead.Assignee)
    }

    return m
}

// typeIndexFromString converts issue type string to index.
// Returns 0 (Task) as safe default for unknown types.
func typeIndexFromString(issueType string) int {
    for i, t := range typeOptions {
        if t == issueType {
            return i
        }
    }
    return 0
}
```

#### 1.3 Add mode-aware helpers

**File**: `internal/ui/overlay_create.go`

```go
func (m *CreateOverlay) header() string {
    if m.isEditMode() {
        return fmt.Sprintf("EDIT: %s", m.editingBead.ID)
    }
    return "NEW BEAD"
}

func (m *CreateOverlay) submitFooterText() string {
    if m.isEditMode() {
        return "Save"
    }
    return "Create"
}
```

#### 1.4 Add edit submission (no change detection)

**File**: `internal/ui/overlay_create.go`

```go
// BeadUpdatedMsg is sent when edit form is submitted.
type BeadUpdatedMsg struct {
    ID          string
    Title       string
    Description string
    IssueType   string
    Priority    int
    ParentID    string
    Labels      []string
    Assignee    string
}

func (m *CreateOverlay) submitEdit() tea.Cmd {
    return func() tea.Msg {
        return BeadUpdatedMsg{
            ID:          m.editingBead.ID,
            Title:       strings.TrimSpace(m.titleInput.Value()),
            Description: strings.TrimSpace(m.descriptionInput.Value()),
            IssueType:   typeOptions[m.typeIndex],
            Priority:    m.priorityIndex,
            ParentID:    m.ParentID(),
            Labels:      m.labelsCombo.GetChips(),
            Assignee:    m.getAssigneeValue(),
        }
    }
}

// Update existing submit handling to route based on mode
func (m *CreateOverlay) handleSubmit(stayOpen bool) (*CreateOverlay, tea.Cmd) {
    // Existing title validation...
    if strings.TrimSpace(m.titleInput.Value()) == "" {
        m.titleValidationError = true
        return m, titleFlashCmd()
    }

    m.isCreating = true

    if m.isEditMode() {
        return m, m.submitEdit()  // Edit ignores stayOpen
    }
    return m, m.submitWithMode(stayOpen)
}
```

#### 1.5 Update View() for dynamic header/footer

Update `View()` method to use `header()` and `submitFooterText()` helpers instead of hardcoded strings.

---

### Phase 2: Wire Up Edit Key and Backend

#### 2.1 Add Edit key binding

**File**: `internal/ui/keys.go`

```go
type KeyMap struct {
    // ... existing bindings
    Edit key.Binding
}

// In DefaultKeyMap():
Edit: key.NewBinding(
    key.WithKeys("e"),
    key.WithHelp("e", "Edit bead"),
),
```

#### 2.2 Handle Edit key press

**File**: `internal/ui/update.go`

```go
case key.Matches(msg, m.keys.Edit):
    if len(m.visibleRows) > 0 {
        row := m.visibleRows[m.cursor]
        m.activeOverlay = OverlayCreate
        m.createOverlay = NewEditOverlay(&row.Node.FullIssue, CreateOverlayOptions{
            AvailableParents:   m.getAvailableParents(),
            AvailableLabels:    m.getAvailableLabels(),
            AvailableAssignees: m.getAvailableAssignees(),
        })
        return m, m.createOverlay.Init()
    }
```

#### 2.3 Handle BeadUpdatedMsg

**File**: `internal/ui/update.go`

```go
type updateCompleteMsg struct {
    ID  string
    Err error
}

case BeadUpdatedMsg:
    return m, m.executeUpdateCmd(msg)

case updateCompleteMsg:
    m.activeOverlay = OverlayNone
    m.createOverlay = nil
    if msg.Err != nil {
        m.toast = NewToast(msg.Err.Error(), ToastError)
        return m, nil
    }
    m.toast = NewToast(fmt.Sprintf("Updated %s", msg.ID), ToastSuccess)
    return m, m.forceRefresh()
```

#### 2.4 Backend execution (simplified)

**File**: `internal/ui/update.go`

```go
func (m *App) executeUpdateCmd(msg BeadUpdatedMsg) tea.Cmd {
    return func() tea.Msg {
        args := []string{"update", msg.ID,
            "--title", msg.Title,
            "--description", msg.Description,
            "--type", msg.IssueType,
            "--priority", fmt.Sprintf("%d", msg.Priority),
        }

        // Parent (empty string clears it)
        if msg.ParentID != "" {
            args = append(args, "--parent", msg.ParentID)
        }

        // Assignee
        if msg.Assignee != "" {
            args = append(args, "--assignee", msg.Assignee)
        }

        // Labels - use --set-labels for atomic replacement
        if len(msg.Labels) > 0 {
            for _, label := range msg.Labels {
                args = append(args, "--set-labels", label)
            }
        } else {
            // Clear all labels
            args = append(args, "--set-labels", "")
        }

        cmd := exec.Command("bd", args...)
        var stderr bytes.Buffer
        cmd.Stderr = &stderr

        err := cmd.Run()
        if err != nil && stderr.Len() > 0 {
            err = fmt.Errorf("%s", strings.TrimSpace(stderr.String()))
        }

        return updateCompleteMsg{ID: msg.ID, Err: err}
    }
}
```

---

### Phase 3: Polish

#### 3.1 Update help overlay

**File**: `internal/ui/help.go`

Add `Edit` key to help display, grouped with other bead actions.

#### 3.2 Add tests

**File**: `internal/ui/overlay_create_test.go`

```go
func TestNewEditOverlay(t *testing.T) {
    bead := &beads.FullIssue{
        ID:          "ab-123",
        Title:       "Test Title",
        Description: "Test Description",
        IssueType:   "bug",
        Priority:    2,
        Labels:      []string{"urgent", "backend"},
        Assignee:    "alice",
    }

    m := NewEditOverlay(bead, CreateOverlayOptions{})

    assert.True(t, m.isEditMode())
    assert.Equal(t, "Test Title", m.titleInput.Value())
    assert.Equal(t, "Test Description", m.descriptionInput.Value())
    assert.Equal(t, 1, m.typeIndex)  // bug index
    assert.Equal(t, 2, m.priorityIndex)
    assert.Equal(t, []string{"urgent", "backend"}, m.labelsCombo.GetChips())
}

func TestTypeIndexFromString(t *testing.T) {
    tests := []struct {
        input string
        want  int
    }{
        {"task", 0},
        {"bug", 1},
        {"feature", 2},
        {"epic", 3},
        {"unknown", 0},  // defaults to task
        {"", 0},
    }
    for _, tt := range tests {
        assert.Equal(t, tt.want, typeIndexFromString(tt.input))
    }
}

func TestIsEditMode(t *testing.T) {
    create := NewCreateOverlay(CreateOverlayOptions{})
    assert.False(t, create.isEditMode())

    edit := NewEditOverlay(&beads.FullIssue{ID: "ab-1"}, CreateOverlayOptions{})
    assert.True(t, edit.isEditMode())
}
```

---

## Files to Modify

| File | Changes |
|------|---------|
| `internal/ui/overlay_create.go` | Add `editingBead` field, `NewEditOverlay()`, `isEditMode()`, `header()`, `submitFooterText()`, `submitEdit()`, `typeIndexFromString()`, `BeadUpdatedMsg`; update `handleSubmit()` and `View()` |
| `internal/ui/keys.go` | Add `Edit` key binding |
| `internal/ui/update.go` | Handle `Edit` key, `BeadUpdatedMsg`, add `executeUpdateCmd()`, `updateCompleteMsg` |
| `internal/ui/help.go` | Add Edit to help display |
| `internal/ui/overlay_create_test.go` | Add tests for edit mode |

---

## Testing Checklist

### Unit Tests
- [ ] `NewEditOverlay` pre-populates all fields correctly
- [ ] `isEditMode()` returns true when `editingBead != nil`
- [ ] `typeIndexFromString()` handles all types + unknown
- [ ] `header()` returns "EDIT: ab-xxx" in edit mode
- [ ] `submitFooterText()` returns "Save" in edit mode

### Integration Tests (manual via TUI)
- [ ] `e` on selected bead opens form with pre-populated values
- [ ] Tab navigation works identically to Create
- [ ] Enter saves, Esc cancels
- [ ] Success toast shows "Updated ab-xxx"
- [ ] Tree refreshes after edit
- [ ] Detail pane shows updated values
- [ ] Edit bead with empty description works
- [ ] Edit bead, clear all labels → labels cleared
- [ ] Backend error → error toast displayed

---

## Acceptance Criteria

- [ ] `e` key opens edit form on selected bead
- [ ] All current values pre-populated in form
- [ ] Title, Description, Type, Priority, Parent, Labels, Assignee are editable
- [ ] Ctrl+Enter saves, Esc cancels
- [ ] Toast shows on success/error
- [ ] Detail pane refreshes after edit
- [ ] Help overlay shows Edit key

---

## What Was Removed (vs Original Plan)

| Removed | Rationale |
|---------|-----------|
| `BeadUpdate` struct with pointer semantics | `bd update` is idempotent |
| `buildUpdate()` change detection | Not needed |
| `labelsEqual()` set comparison | Using `--set-labels` instead |
| `syncLabels()` add/remove logic | Using `--set-labels` instead |
| `IsEmpty()` / "No changes" toast | Can't detect without diffing |
| 6 implementation phases | Collapsed to 3 |

**LOC estimate:** ~100 lines (vs ~200+ original)
