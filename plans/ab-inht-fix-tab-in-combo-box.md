# fix: Tab in ChipComboBox should select highlighted item before moving focus

**Issue:** ab-inht
**Type:** Bug Fix
**Priority:** P2
**Created:** 2025-12-07
**Updated:** 2025-12-07 (v4-final - separate message types, explicit methods)

## Problem Statement

When selecting an item in the labels combobox on the "add new bead" screen:
1. User types "ui" in the labels field
2. Dropdown opens with "UI" highlighted
3. User presses Tab
4. **Bug:** The typed text "ui" stays in the input, no chip is added
5. Focus moves to the next field

**Expected behavior:** Tab should select the highlighted item, add it as a chip, then move focus.

## Root Cause

**File:** `internal/ui/chipcombobox.go:180-195` (`handleTab()` function)

```go
if c.combo.IsDropdownOpen() {
    c.combo, cmd = c.combo.Update(tea.KeyMsg{Type: tea.KeyTab})
    if cmd != nil {
        cmds = append(cmds, cmd)  // ← Async: ComboBoxValueSelectedMsg
    }
    cmds = append(cmds, func() tea.Msg { return ChipComboBoxTabMsg{} })  // ← Races ahead
    return c, tea.Batch(cmds...)
}
```

**The timing issue:** `ChipComboBoxTabMsg` is batched alongside the selection command. The parent processes Tab before the selection completes.

## Solution: Separate Message Types (Idiomatic TEA)

Use **distinct message types** for different selection triggers. Let Go's type system do the work instead of inspecting parameters.

**Principle:** "Messages are events, not generic DTOs. Explicitness beats cleverness."

```go
// Before: One parameterized message
type ComboBoxValueSelectedMsg struct {
    Value string
    IsNew bool
}

// After: Explicit message types for each trigger
type ComboBoxEnterSelectedMsg struct {
    Value string
    IsNew bool
}

type ComboBoxTabSelectedMsg struct {
    Value string
    IsNew bool
}
```

Consumers use Go's type switch - no conditionals on parameters:

```go
case ComboBoxTabSelectedMsg:
    c = c.addChip(msg.Value)
    return c, func() tea.Msg { return ChipComboBoxTabMsg{} }

case ComboBoxEnterSelectedMsg:
    c = c.addChip(msg.Value)
    return c, nil  // Stay focused
```

## Acceptance Criteria

- [ ] Tab with dropdown open adds highlighted item as chip
- [ ] Tab with dropdown open moves focus to next field after chip is added
- [ ] Tab with ghost text visible completes the suggestion and adds chip
- [ ] Tab with duplicate selection flashes the existing chip but does NOT advance (user stays to fix)
- [ ] Tab when no valid selection (no matches, AllowNew=false) just moves focus
- [ ] Enter key behavior remains unchanged (adds chip, stays in field)
- [ ] Existing tests pass
- [ ] New tests cover the bug scenario

## Implementation Plan

### Phase 1: Add new message types in combobox.go

**File:** `internal/ui/combobox.go:25-29`

Replace the single message with two explicit types:

```go
// ComboBoxEnterSelectedMsg is sent when Enter confirms a selection.
// The component stays focused for additional input.
type ComboBoxEnterSelectedMsg struct {
    Value string
    IsNew bool
}

// ComboBoxTabSelectedMsg is sent when Tab confirms a selection.
// Signals that the component should advance to next field after processing.
type ComboBoxTabSelectedMsg struct {
    Value string
    IsNew bool
}
```

### Phase 2: Create explicit selection methods for each trigger

**File:** `internal/ui/combobox.go:279-313`

Replace the generic `selectHighlighted()` with explicit methods. "A little copying is better than a little dependency."

```go
// selectHighlightedWithEnter selects the highlighted option via Enter key.
func (c ComboBox) selectHighlightedWithEnter() (ComboBox, tea.Cmd) {
    if len(c.filteredOptions) > 0 && c.highlightIndex >= 0 && c.highlightIndex < len(c.filteredOptions) {
        selected := c.filteredOptions[c.highlightIndex]
        c.value = selected
        c.textInput.SetValue(selected)
        c.originalValue = selected
        c.state = ComboBoxIdle
        return c, func() tea.Msg {
            return ComboBoxEnterSelectedMsg{Value: selected, IsNew: false}
        }
    }
    c.state = ComboBoxIdle
    return c, nil
}

// selectHighlightedWithTab selects the highlighted option via Tab key.
func (c ComboBox) selectHighlightedWithTab() (ComboBox, tea.Cmd) {
    if len(c.filteredOptions) > 0 && c.highlightIndex >= 0 && c.highlightIndex < len(c.filteredOptions) {
        selected := c.filteredOptions[c.highlightIndex]
        c.value = selected
        c.textInput.SetValue(selected)
        c.originalValue = selected
        c.state = ComboBoxIdle
        return c, func() tea.Msg {
            return ComboBoxTabSelectedMsg{Value: selected, IsNew: false}
        }
    }
    c.state = ComboBoxIdle
    return c, nil
}

// selectHighlightedOrNewWithEnter handles Enter: select highlighted or create new.
func (c ComboBox) selectHighlightedOrNewWithEnter() (ComboBox, tea.Cmd) {
    if len(c.filteredOptions) > 0 && c.highlightIndex >= 0 && c.highlightIndex < len(c.filteredOptions) {
        return c.selectHighlightedWithEnter()
    }

    if c.AllowNew && strings.TrimSpace(c.textInput.Value()) != "" {
        newValue := strings.TrimSpace(c.textInput.Value())
        c.value = newValue
        c.originalValue = newValue
        c.state = ComboBoxIdle
        return c, func() tea.Msg {
            return ComboBoxEnterSelectedMsg{Value: newValue, IsNew: true}
        }
    }

    c.state = ComboBoxIdle
    return c, nil
}

// selectHighlightedOrNewWithTab handles Tab: select highlighted or create new.
func (c ComboBox) selectHighlightedOrNewWithTab() (ComboBox, tea.Cmd) {
    if len(c.filteredOptions) > 0 && c.highlightIndex >= 0 && c.highlightIndex < len(c.filteredOptions) {
        return c.selectHighlightedWithTab()
    }

    if c.AllowNew && strings.TrimSpace(c.textInput.Value()) != "" {
        newValue := strings.TrimSpace(c.textInput.Value())
        c.value = newValue
        c.originalValue = newValue
        c.state = ComboBoxIdle
        return c, func() tea.Msg {
            return ComboBoxTabSelectedMsg{Value: newValue, IsNew: true}
        }
    }

    c.state = ComboBoxIdle
    return c, nil
}
```

### Phase 3: Update key handlers to call explicit methods

**File:** `internal/ui/combobox.go:186-226` (handleBrowsingKey)

```go
case tea.KeyEnter:
    return c.selectHighlightedWithEnter()

case tea.KeyTab:
    return c.selectHighlightedWithTab()
```

**File:** `internal/ui/combobox.go:228-277` (handleFilteringKey)

```go
case tea.KeyEnter:
    return c.selectHighlightedOrNewWithEnter()

case tea.KeyTab:
    return c.selectHighlightedOrNewWithTab()
```

### Phase 4: Update ChipComboBox to handle both message types

**File:** `internal/ui/chipcombobox.go:91-117` (Update function)

Replace the single `ComboBoxValueSelectedMsg` case with two cases:

```go
case ComboBoxEnterSelectedMsg:
    // Enter: add chip, stay in field
    return c.handleEnterSelection(msg.Value, msg.IsNew)

case ComboBoxTabSelectedMsg:
    // Tab: add chip, then advance field
    return c.handleTabSelection(msg.Value, msg.IsNew)
```

### Phase 5: Implement selection handlers in ChipComboBox

**File:** `internal/ui/chipcombobox.go` (new functions)

```go
// handleEnterSelection adds a chip and stays in the field.
func (c ChipComboBox) handleEnterSelection(value string, isNew bool) (ChipComboBox, tea.Cmd) {
    if value == "" {
        return c, nil
    }

    // Duplicate - flash, stay in field
    if c.chips.Contains(value) {
        c.chips.AddChip(value) // Sets flashIndex
        c.combo.SetValue("")
        return c, FlashCmd()
    }

    // Add chip, stay in field
    c.chips.AddChip(value)
    c.combo.SetValue("")
    c.updateAvailableOptions()
    return c, func() tea.Msg {
        return ChipComboBoxChipAddedMsg{Label: value, IsNew: isNew}
    }
}

// handleTabSelection adds a chip and advances to next field.
// Exception: duplicates stay in field (error state).
func (c ChipComboBox) handleTabSelection(value string, isNew bool) (ChipComboBox, tea.Cmd) {
    if value == "" {
        // No selection - still advance field
        return c, func() tea.Msg { return ChipComboBoxTabMsg{} }
    }

    // Duplicate - flash, DON'T advance (user needs to fix)
    if c.chips.Contains(value) {
        c.chips.AddChip(value) // Sets flashIndex
        c.combo.SetValue("")
        return c, FlashCmd()  // No TabMsg - stay in field
    }

    // Add chip AND advance field
    c.chips.AddChip(value)
    c.combo.SetValue("")
    c.updateAvailableOptions()
    return c, tea.Batch(
        func() tea.Msg { return ChipComboBoxChipAddedMsg{Label: value, IsNew: isNew} },
        func() tea.Msg { return ChipComboBoxTabMsg{} },
    )
}
```

### Phase 6: Simplify `handleTab()` in ChipComboBox

**File:** `internal/ui/chipcombobox.go:180-226`

```go
func (c ChipComboBox) handleTab() (ChipComboBox, tea.Cmd) {
    // If dropdown is open, forward Tab to ComboBox
    // ComboBox will send ComboBoxTabSelectedMsg
    // which is handled by handleTabSelection()
    if c.combo.IsDropdownOpen() {
        var cmd tea.Cmd
        c.combo, cmd = c.combo.Update(tea.KeyMsg{Type: tea.KeyTab})
        if cmd != nil {
            return c, cmd  // Let ComboBoxTabSelectedMsg flow through
        }
        // No selection possible - signal Tab immediately
        return c, func() tea.Msg { return ChipComboBoxTabMsg{} }
    }

    // Dropdown closed - handle raw input text
    inputVal := strings.TrimSpace(c.combo.InputValue())
    if inputVal != "" {
        // Duplicate check
        if c.chips.Contains(inputVal) {
            c.chips.AddChip(inputVal) // Flash
            c.combo.SetValue("")
            return c, FlashCmd()  // Stay in field
        }
        // Add chip and advance
        isNew := true
        for _, opt := range c.allOptions {
            if strings.EqualFold(opt, inputVal) {
                isNew = false
                break
            }
        }
        c.chips.AddChip(inputVal)
        c.combo.SetValue("")
        c.updateAvailableOptions()
        return c, tea.Batch(
            func() tea.Msg { return ChipComboBoxChipAddedMsg{Label: inputVal, IsNew: isNew} },
            func() tea.Msg { return ChipComboBoxTabMsg{} },
        )
    }

    // No input - just advance
    return c, func() tea.Msg { return ChipComboBoxTabMsg{} }
}
```

### Phase 7: Update existing tests

**File:** `internal/ui/combobox_test.go`

Update tests that check for `ComboBoxValueSelectedMsg` to check for the appropriate new type:

```go
// Before
case ComboBoxValueSelectedMsg:
    // ...

// After - check for specific message type based on test
case ComboBoxEnterSelectedMsg:
    // Test Enter behavior
case ComboBoxTabSelectedMsg:
    // Test Tab behavior
```

### Phase 8: Add new tests

**File:** `internal/ui/chipcombobox_test.go`

```go
func TestChipComboBox_TabSelectedMsg_AddsChipAndAdvances(t *testing.T) {
    cc := NewChipComboBox([]string{"UI", "Bug", "Feature"})

    // Simulate receiving TabSelectedMsg
    cc, cmd := cc.Update(ComboBoxTabSelectedMsg{Value: "UI", IsNew: false})

    // Chip added
    assert.Equal(t, 1, len(cc.chips.Chips))
    assert.Equal(t, "UI", cc.chips.Chips[0])

    // Should return batch with ChipAddedMsg AND TabMsg
    assertBatchContains(t, cmd, ChipComboBoxChipAddedMsg{})
    assertBatchContains(t, cmd, ChipComboBoxTabMsg{})
}

func TestChipComboBox_EnterSelectedMsg_AddsChipStaysInField(t *testing.T) {
    cc := NewChipComboBox([]string{"UI", "Bug", "Feature"})

    // Simulate receiving EnterSelectedMsg
    cc, cmd := cc.Update(ComboBoxEnterSelectedMsg{Value: "UI", IsNew: false})

    // Chip added
    assert.Equal(t, 1, len(cc.chips.Chips))

    // Should return ChipAddedMsg but NOT TabMsg
    assertCmdProduces(t, cmd, ChipComboBoxChipAddedMsg{})
    assertCmdDoesNotProduce(t, cmd, ChipComboBoxTabMsg{})
}

func TestChipComboBox_TabSelectedMsg_DuplicateStaysInField(t *testing.T) {
    cc := NewChipComboBox([]string{"UI", "Bug", "Feature"})
    cc.chips.AddChip("UI")  // Pre-add chip

    // Tab on duplicate
    cc, cmd := cc.Update(ComboBoxTabSelectedMsg{Value: "UI", IsNew: false})

    // Still only 1 chip
    assert.Equal(t, 1, len(cc.chips.Chips))

    // Should return FlashCmd, NOT TabMsg (stay in field)
    assertCmdProduces(t, cmd, chipFlashClearMsg{})
    assertCmdDoesNotProduce(t, cmd, ChipComboBoxTabMsg{})
}

func TestChipComboBox_EnterSelectedMsg_DuplicateFlashes(t *testing.T) {
    cc := NewChipComboBox([]string{"UI", "Bug", "Feature"})
    cc.chips.AddChip("UI")

    // Enter on duplicate
    cc, cmd := cc.Update(ComboBoxEnterSelectedMsg{Value: "UI", IsNew: false})

    // Still only 1 chip, flash triggered
    assert.Equal(t, 1, len(cc.chips.Chips))
    assertCmdProduces(t, cmd, chipFlashClearMsg{})
}

func TestChipComboBox_TabSelectedMsg_EmptyValueStillAdvances(t *testing.T) {
    cc := NewChipComboBox([]string{"UI", "Bug"})

    // Tab with no selection
    cc, cmd := cc.Update(ComboBoxTabSelectedMsg{Value: "", IsNew: false})

    // No chips added
    assert.Equal(t, 0, len(cc.chips.Chips))

    // But Tab signal sent (advance anyway)
    assertCmdProduces(t, cmd, ChipComboBoxTabMsg{})
}

func TestChipComboBox_TabSelectedMsg_NewValueAddedAndAdvances(t *testing.T) {
    cc := NewChipComboBox([]string{"UI", "Bug"})

    // Tab with new value (not in options)
    cc, cmd := cc.Update(ComboBoxTabSelectedMsg{Value: "NewLabel", IsNew: true})

    // Chip added
    assert.Equal(t, 1, len(cc.chips.Chips))
    assert.Equal(t, "NewLabel", cc.chips.Chips[0])

    // Should include IsNew: true in ChipAddedMsg
    assertBatchContains(t, cmd, ChipComboBoxChipAddedMsg{Label: "NewLabel", IsNew: true})
    assertBatchContains(t, cmd, ChipComboBoxTabMsg{})
}

func TestChipComboBox_Tab_RawInputWhenDropdownClosed(t *testing.T) {
    cc := NewChipComboBox([]string{"UI", "Bug"})
    // Simulate raw input with dropdown closed
    cc.combo.SetValue("Backend")

    // Press Tab directly (dropdown closed path)
    cc, cmd := cc.Update(tea.KeyMsg{Type: tea.KeyTab})

    // Chip added from raw input
    assert.Equal(t, 1, len(cc.chips.Chips))
    assert.Equal(t, "Backend", cc.chips.Chips[0])

    // Tab signal sent
    assertBatchContains(t, cmd, ChipComboBoxTabMsg{})
}

func TestChipComboBox_Tab_NoSelectionAllowNewFalse(t *testing.T) {
    cc := NewChipComboBox([]string{"UI", "Bug"}).WithAllowNew(false, "")

    // Type something that doesn't match, dropdown closed
    cc.combo.SetValue("xyz")

    // Press Tab - no valid selection possible
    cc, cmd := cc.Update(tea.KeyMsg{Type: tea.KeyTab})

    // No chips added (AllowNew=false, no match)
    assert.Equal(t, 0, len(cc.chips.Chips))

    // Tab signal still sent (advance anyway)
    assertCmdProduces(t, cmd, ChipComboBoxTabMsg{})
}
```

## Edge Cases

| Scenario | Message Type | Tab Signal? | Rationale |
|----------|--------------|-------------|-----------|
| Enter + valid selection | `EnterSelectedMsg` | No | Multi-select: stay for more input |
| Enter + duplicate | `EnterSelectedMsg` | No | Flash error, stay to fix |
| Tab + valid selection | `TabSelectedMsg` | Yes | Success: advance to next field |
| Tab + duplicate | `TabSelectedMsg` | **No** | Error: stay to fix |
| Tab + empty (no selection possible) | `TabSelectedMsg` | Yes | No error: just advance |
| Tab + dropdown closed + text | N/A (direct handling) | Yes | Raw input processed |

## Why This Approach?

### Messages Are Events, Not DTOs
`ComboBoxTabSelectedMsg` means "Tab was pressed and selection completed" - it's self-describing. No need to inspect parameters.

### Go's Type System Does the Work
Type switches are cleaner than parameter conditionals:

```go
// Clean - type switch
switch msg := msg.(type) {
case ComboBoxTabSelectedMsg:
    // Tab behavior
case ComboBoxEnterSelectedMsg:
    // Enter behavior
}

// Cluttered - parameter inspection
if msg.TriggeredBy == tea.KeyTab {
    // Tab behavior
} else if msg.TriggeredBy == tea.KeyEnter {
    // Enter behavior
}
```

### Consumers Choose Their Concern
If a consumer only cares about Enter, it handles `ComboBoxEnterSelectedMsg`. If it needs both, it handles both. No forced coupling.

### Duplicate Handling (Kieran's Bug Fix)
Duplicates are error states. Regardless of Tab/Enter, duplicates should NOT advance the field - user needs to stay and fix the error.

## Files Changed

| File | Changes |
|------|---------|
| `internal/ui/combobox.go` | New message types, update selection functions |
| `internal/ui/chipcombobox.go` | Handle both message types, new selection handlers |
| `internal/ui/combobox_test.go` | Update for new message types |
| `internal/ui/chipcombobox_test.go` | Add tests for Tab/Enter behavior |

## References

### Internal References
- ComboBox messages: `internal/ui/combobox.go:25-29`
- Selection functions: `internal/ui/combobox.go:279-313`
- ChipComboBox Update: `internal/ui/chipcombobox.go:80-154`
- handleTab: `internal/ui/chipcombobox.go:180-226`

### Best Practices
- Elm Architecture: Messages should be explicit events
- Go idiom: Use type switches over parameter inspection
- UX: Error states (duplicates) should not advance fields

### Related Issues
- Parent issue: ab-1hxl (Abacus Release v0.5.0)
