# Plan: Add Comment Modal to TUI (ab-d03)

## Overview

Add the ability to add comments to beads directly from the TUI using a modal with a textarea, triggered by the `m` key.

**Bead**: ab-d03
**Type**: Feature
**Priority**: P1
**Dependencies**: ab-gtm (closed), ab-3m12 (parent epic)

## Problem Statement

Users currently need to exit the TUI and use the CLI (`bd comments add`) to add comments to beads. This breaks workflow and reduces productivity. Comments are essential for collaboration and context tracking.

## Proposed Solution

Implement a minimal comment modal overlay that:
- Opens with `m` key on a selected bead
- Shows a multi-line textarea for comment input
- Submits with `Ctrl+Enter`, allows newlines with `Enter`
- Cancels with `Esc`
- Shows success toast after submission
- Refreshes detail pane to display the new comment

## Technical Approach

### Architecture

Follow existing overlay patterns established in:
- `internal/ui/overlay_status.go` - Simple selection overlay pattern
- `internal/ui/overlay_create.go:209-227` - Textarea usage pattern
- `internal/ui/overlay_labels.go:119-148` - Multi-stage Escape handling

Use `bubbles/textarea` (already in dependencies) rather than huh forms, as the codebase consistently uses this approach.

### Implementation Phases

#### Phase 1: Backend Integration

**Files to modify:**

1. **`internal/beads/client.go`** (after line 18)
   - Add `AddComment(ctx context.Context, issueID, text string) error` to `Client` interface

2. **`internal/beads/cli.go`** (after line 100)
   - Implement `AddComment` method:
   ```go
   func (c *cliClient) AddComment(ctx context.Context, issueID, text string) error {
       _, err := c.run(ctx, "comments", "add", issueID, text)
       return err
   }
   ```

3. **`internal/beads/mock.go`** (if exists)
   - Add mock implementation for testing

#### Phase 2: Key Binding & Overlay Type

**Files to modify:**

1. **`internal/ui/keys.go`**
   - Line 34: Add `Comment key.Binding` to KeyMap struct
   - After line 143: Add binding definition:
   ```go
   Comment: key.NewBinding(
       key.WithKeys("m"),
       key.WithHelp("m", "Add comment"),
   ),
   ```

2. **`internal/ui/app.go`**
   - Line 38: Add `OverlayComment` to OverlayType enum
   - After line 162: Add `commentOverlay *CommentOverlay` to App struct
   - After line 199: Add toast state fields:
   ```go
   // Comment toast state
   commentToastVisible bool
   commentToastStart   time.Time
   commentToastBeadID  string
   ```

#### Phase 3: Comment Overlay Component

**File to create: `internal/ui/overlay_comment.go`**

```go
package ui

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/bubbles/textarea"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

const (
    commentModalWidth    = 54  // Content width
    commentTextareaLines = 6   // Visible lines
    commentCharLimit     = 2000
)

// Message types
type CommentAddedMsg struct {
    IssueID string
    Comment string
}

type CommentCancelledMsg struct{}

type CommentOverlay struct {
    issueID    string
    beadTitle  string
    textarea   textarea.Model
    errorMsg   string
    termWidth  int
}

func NewCommentOverlay(issueID, beadTitle string) *CommentOverlay {
    ta := textarea.New()
    ta.Placeholder = "Type your comment here..."
    ta.CharLimit = commentCharLimit
    ta.SetWidth(commentModalWidth)
    ta.SetHeight(commentTextareaLines)
    ta.ShowLineNumbers = false
    ta.Focus()

    return &CommentOverlay{
        issueID:   issueID,
        beadTitle: beadTitle,
        textarea:  ta,
    }
}

func (m *CommentOverlay) Init() tea.Cmd {
    return textarea.Blink
}

func (m *CommentOverlay) Update(msg tea.Msg) (*CommentOverlay, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case msg.Type == tea.KeyEsc:
            // Multi-stage escape: clear text first, then cancel
            if strings.TrimSpace(m.textarea.Value()) != "" {
                m.textarea.SetValue("")
                m.errorMsg = ""
                return m, nil
            }
            return m, func() tea.Msg { return CommentCancelledMsg{} }

        case msg.String() == "ctrl+enter":
            return m.submit()
        }
    }

    // Pass to textarea
    var cmd tea.Cmd
    m.textarea, cmd = m.textarea.Update(msg)
    return m, cmd
}

func (m *CommentOverlay) submit() (*CommentOverlay, tea.Cmd) {
    text := strings.TrimSpace(m.textarea.Value())
    if text == "" {
        m.errorMsg = "Comment cannot be empty"
        return m, nil
    }
    m.errorMsg = ""
    return m, func() tea.Msg {
        return CommentAddedMsg{
            IssueID: m.issueID,
            Comment: text,
        }
    }
}

func (m *CommentOverlay) View() string {
    var b strings.Builder

    // Header
    header := styleHelpTitle().Render("ADD COMMENT")
    divider := styleHelpDivider().Render(strings.Repeat("─", commentModalWidth+4))

    b.WriteString(header)
    b.WriteString("\n")
    b.WriteString(divider)
    b.WriteString("\n\n")

    // Bead context line using common ID + title renderer
    // styleID() renders the bead ID in gold/bold
    // Truncate title if too long
    title := m.beadTitle
    if len(title) > 40 {
        title = title[:37] + "..."
    }
    contextLine := styleID().Render(m.issueID) + " " + title
    b.WriteString(contextLine)
    b.WriteString("\n\n")

    // Textarea with border
    taStyle := styleCreateInput(commentModalWidth + 4)
    b.WriteString(taStyle.Render(m.textarea.View()))
    b.WriteString("\n")

    // Character count
    count := len(m.textarea.Value())
    countStyle := lipgloss.NewStyle().Foreground(theme.Current().TextMuted())
    if count > commentCharLimit-100 {
        countStyle = lipgloss.NewStyle().Foreground(theme.Current().Warning())
    }
    b.WriteString(countStyle.Render(fmt.Sprintf("  %d/%d", count, commentCharLimit)))
    b.WriteString("\n")

    // Error message
    if m.errorMsg != "" {
        errorStyle := lipgloss.NewStyle().Foreground(theme.Current().Error())
        b.WriteString(errorStyle.Render("  ⚠ " + m.errorMsg))
        b.WriteString("\n")
    }

    b.WriteString("\n")
    b.WriteString(divider)
    b.WriteString("\n")

    // Footer
    hints := []footerHint{
        {"^⏎", "Submit"},
        {"esc", "Cancel"},
    }
    b.WriteString(overlayFooterLine(hints, commentModalWidth+4))

    return styleHelpOverlay().Render(b.String())
}

func (m *CommentOverlay) Layer(width, height, topMargin, bottomMargin int) Layer {
    return LayerFunc(func() *Canvas {
        content := m.View()
        if strings.TrimSpace(content) == "" {
            return nil
        }

        overlayWidth := lipgloss.Width(content)
        overlayHeight := lipgloss.Height(content)

        surface := NewSecondarySurface(overlayWidth, overlayHeight)
        surface.Draw(0, 0, content)

        x, y := centeredOffsets(width, height, overlayWidth, overlayHeight, topMargin, bottomMargin)
        surface.Canvas.SetOffset(x, y)
        return surface.Canvas
    })
}

func (m *CommentOverlay) SetSize(width, height int) {
    m.termWidth = width
}
```

**File to create: `internal/ui/overlay_comment_test.go`**

```go
package ui

import (
    "testing"

    tea "github.com/charmbracelet/bubbletea"
)

func TestCommentOverlay_NewCommentOverlay(t *testing.T) {
    overlay := NewCommentOverlay("ab-123", "Test bead title")

    if overlay.issueID != "ab-123" {
        t.Errorf("expected issueID ab-123, got %s", overlay.issueID)
    }
    if overlay.beadTitle != "Test bead title" {
        t.Errorf("expected beadTitle 'Test bead title', got %s", overlay.beadTitle)
    }
}

func TestCommentOverlay_EmptySubmit(t *testing.T) {
    overlay := NewCommentOverlay("ab-123", "Test")

    // Try to submit empty
    overlay, _ = overlay.Update(tea.KeyMsg{Type: tea.KeyCtrlEnter})

    if overlay.errorMsg == "" {
        t.Error("expected error message for empty comment")
    }
}

func TestCommentOverlay_ValidSubmit(t *testing.T) {
    overlay := NewCommentOverlay("ab-123", "Test")
    overlay.textarea.SetValue("This is a valid comment")

    _, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyCtrlEnter})

    if cmd == nil {
        t.Error("expected command to be returned")
    }

    msg := cmd()
    if addedMsg, ok := msg.(CommentAddedMsg); !ok {
        t.Error("expected CommentAddedMsg")
    } else {
        if addedMsg.IssueID != "ab-123" {
            t.Errorf("expected issueID ab-123, got %s", addedMsg.IssueID)
        }
        if addedMsg.Comment != "This is a valid comment" {
            t.Errorf("unexpected comment text: %s", addedMsg.Comment)
        }
    }
}

func TestCommentOverlay_EscapeClearsText(t *testing.T) {
    overlay := NewCommentOverlay("ab-123", "Test")
    overlay.textarea.SetValue("Some text")

    // First Esc clears text
    overlay, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})

    if cmd != nil {
        t.Error("first Esc should not produce command")
    }
    if overlay.textarea.Value() != "" {
        t.Error("expected textarea to be cleared")
    }
}

func TestCommentOverlay_EscapeCancels(t *testing.T) {
    overlay := NewCommentOverlay("ab-123", "Test")
    // Textarea is empty

    _, cmd := overlay.Update(tea.KeyMsg{Type: tea.KeyEsc})

    if cmd == nil {
        t.Error("expected cancel command")
    }

    msg := cmd()
    if _, ok := msg.(CommentCancelledMsg); !ok {
        t.Error("expected CommentCancelledMsg")
    }
}
```

#### Phase 4: Integration with App

**Files to modify:**

1. **`internal/ui/update.go`**

   After line 448 (overlay delegation section), add:
   ```go
   if m.activeOverlay == OverlayComment && m.commentOverlay != nil {
       m.commentOverlay, cmd = m.commentOverlay.Update(msg)
       return m, cmd
   }
   ```

   After line 299 (message handling section), add:
   ```go
   case CommentAddedMsg:
       m.activeOverlay = OverlayNone
       m.commentOverlay = nil
       return m, tea.Batch(
           m.executeAddComment(msg),
           scheduleCommentToastTick(),
       )
   case CommentCancelledMsg:
       m.activeOverlay = OverlayNone
       m.commentOverlay = nil
       return m, nil
   case commentCompleteMsg:
       if msg.err != nil {
           m.lastError = msg.err.Error()
           m.lastErrorSource = errorSourceOperation
           m.showErrorToast = true
           m.errorToastStart = time.Now()
           return m, scheduleErrorToastTick()
       }
       m.displayCommentToast(msg.issueID)
       return m, tea.Batch(m.forceRefresh(), scheduleCommentToastTick())
   case commentToastTickMsg:
       if !m.commentToastVisible {
           return m, nil
       }
       if time.Since(m.commentToastStart) >= 7*time.Second {
           m.commentToastVisible = false
           return m, nil
       }
       return m, scheduleCommentToastTick()
   ```

   After line 628 (key handling section), add:
   ```go
   case key.Matches(msg, m.keys.Comment):
       if len(m.visibleRows) > 0 {
           row := m.visibleRows[m.cursor]
           m.commentOverlay = NewCommentOverlay(row.Node.Issue.ID, row.Node.Issue.Title)
           m.activeOverlay = OverlayComment
           return m, m.commentOverlay.Init()
       }
       return m, nil
   ```

   Add helper functions:
   ```go
   type commentCompleteMsg struct {
       issueID string
       err     error
   }

   func (m *App) executeAddComment(msg CommentAddedMsg) tea.Cmd {
       return func() tea.Msg {
           ctx := context.Background()
           err := m.client.AddComment(ctx, msg.IssueID, msg.Comment)
           return commentCompleteMsg{issueID: msg.IssueID, err: err}
       }
   }

   func (m *App) displayCommentToast(issueID string) {
       m.commentToastBeadID = issueID
       m.commentToastVisible = true
       m.commentToastStart = time.Now()
   }

   type commentToastTickMsg struct{}

   func scheduleCommentToastTick() tea.Cmd {
       return tea.Tick(200*time.Millisecond, func(_ time.Time) tea.Msg {
           return commentToastTickMsg{}
       })
   }
   ```

2. **`internal/ui/view.go`**

   After line 157 (overlay rendering), add:
   ```go
   } else if m.activeOverlay == OverlayComment && m.commentOverlay != nil {
       if layer := m.commentOverlay.Layer(m.width, m.height, headerHeight, bottomMargin); layer != nil {
           overlayLayers = append(overlayLayers, layer)
       }
   ```

   Line 176 (toast factories), add:
   ```go
   m.commentToastLayer,
   ```

   Add toast layer method (after line 504):
   ```go
   func (m *App) commentToastLayer(width, height, mainBodyStart, mainBodyHeight int) Layer {
       if !m.commentToastVisible || m.commentToastBeadID == "" {
           return nil
       }

       elapsed := time.Since(m.commentToastStart)
       remaining := 7 - int(elapsed.Seconds())
       if remaining < 0 {
           remaining = 0
       }

       content := fmt.Sprintf("  ✓ Comment added\n  %s               [%ds]",
           m.commentToastBeadID, remaining)

       return newToastLayer(
           styleSuccessToast().Render(content),
           width, height, mainBodyStart, mainBodyHeight,
       )
   }
   ```

3. **`internal/ui/footer.go`**

   Line 27 (global footer hints), add after Labels:
   ```go
   {"m", "Comment"},
   ```

4. **`internal/ui/help.go`**

   Add Comment to the help display (actions section)

## Acceptance Criteria

- [ ] `m` key opens comment modal on selected bead
- [ ] Issue ID shown in modal header
- [ ] Textarea accepts multi-line input
- [ ] `Enter` adds newlines (not submit)
- [ ] `Ctrl+Enter` submits comment
- [ ] Empty comment shows validation error
- [ ] `Esc` clears text first, then cancels
- [ ] `bd comments add` command executes on submit
- [ ] Toast shows "✓ Comment added" on success
- [ ] Error toast shows on failure
- [ ] Detail pane refreshes with new comment
- [ ] Author/timestamp auto-populated by bd CLI
- [ ] Footer shows `m Comment` hint in tree view
- [ ] Help overlay includes Comment action

## Testing Plan

### Unit Tests
- [ ] `TestCommentOverlay_NewCommentOverlay` - Constructor initializes correctly
- [ ] `TestCommentOverlay_EmptySubmit` - Empty comment shows error
- [ ] `TestCommentOverlay_ValidSubmit` - Valid comment produces message
- [ ] `TestCommentOverlay_EscapeClearsText` - First Esc clears textarea
- [ ] `TestCommentOverlay_EscapeCancels` - Second Esc cancels modal

### Visual Testing
```bash
make build
./scripts/tui-test.sh start
./scripts/tui-test.sh keys 'jjj'     # Navigate to a bead
./scripts/tui-test.sh keys 'm'        # Open comment modal
./scripts/tui-test.sh view            # Verify modal appearance
./scripts/tui-test.sh quit
```

### Integration Testing
1. Open TUI, navigate to a bead
2. Press `m` to open modal
3. Type multi-line comment with Enter
4. Submit with Ctrl+Enter
5. Verify toast appears
6. Press Enter to show detail pane
7. Verify comment appears with author/timestamp

## Files Summary

| File | Action |
|------|--------|
| `internal/beads/client.go` | Add `AddComment` to interface |
| `internal/beads/cli.go` | Implement `AddComment` method |
| `internal/ui/keys.go` | Add `Comment` key binding |
| `internal/ui/app.go` | Add `OverlayComment`, toast state |
| `internal/ui/overlay_comment.go` | **CREATE** - Comment modal |
| `internal/ui/overlay_comment_test.go` | **CREATE** - Unit tests |
| `internal/ui/update.go` | Handle Comment key, messages |
| `internal/ui/view.go` | Render overlay and toast |
| `internal/ui/footer.go` | Add `m` to global hints |
| `internal/ui/help.go` | Add Comment to help display |

## References

### Internal References
- `internal/ui/overlay_create.go:209-227` - Textarea pattern
- `internal/ui/overlay_status.go:13-35` - Simple overlay structure
- `internal/ui/overlay_labels.go:119-148` - Multi-stage Escape
- `docs/UI_PRINCIPLES.md` - Design system guidelines

### External References
- charmbracelet/bubbles textarea: https://pkg.go.dev/github.com/charmbracelet/bubbles/textarea
- bd comments CLI: `bd comments add <issue-id> "text"`

## Design Decisions (From SpecFlow Analysis)

### Resolved Questions

| Question | Decision | Rationale |
|----------|----------|-----------|
| **Library choice** | `bubbles/textarea` | Matches existing CreateOverlay pattern |
| **Overlay conflicts** | `m` disabled when `activeOverlay != OverlayNone` | Prevents UI corruption |
| **Command escaping** | Go `exec.Command` with args slice | Handles escaping automatically |
| **Backend timeout** | 10 seconds via `context.Context` | Prevents UI freeze |
| **Character limit** | 2000 characters | Reasonable for comments, enforced client-side |
| **Textarea height** | 6 visible lines, scrolls for more | Good balance |
| **Modal width** | 54 chars content + borders | Comfortable reading width |
| **Empty validation** | Inline error message | Matches existing patterns |
| **State on error** | Preserve text in modal | Better UX for retry |
| **Closed bead comments** | Allowed | Comments are collaboration tool |

### Edge Cases Handled

- **Press `m` with no bead selected**: No-op (guard condition)
- **Press `m` while another overlay open**: Ignored (overlay check)
- **Special characters in text**: Handled by `exec.Command`
- **Terminal resize during modal**: Modal re-centers (Layer system)
- **Concurrent comment**: Detail pane refreshes on next view

## Notes

- Author is auto-detected by bd from git config
- Timestamp is auto-added by bd
- Markdown in comments will be rendered in detail pane (uses existing glamour renderer)
- This is a simpler overlay than CreateOverlay - single textarea, no complex focus zones
