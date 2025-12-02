package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ChipComboBoxChipAddedMsg is sent when a chip is added.
type ChipComboBoxChipAddedMsg struct {
	Label string
	IsNew bool // True if not in original options (for toast)
}

// ChipComboBoxTabMsg signals Tab was pressed (for field navigation).
type ChipComboBoxTabMsg struct{}

// ChipComboBox composes ChipList + ComboBox for multi-select autocomplete
// fields with tokenizing chips.
type ChipComboBox struct {
	// Composed components
	chips ChipList
	combo ComboBox

	// Configuration
	allOptions []string // Full option list (for restoring after chip removal)
	Width      int

	// State
	focused bool
}

// NewChipComboBox creates a new ChipComboBox with the given options.
func NewChipComboBox(options []string) ChipComboBox {
	// Make a copy of options
	opts := make([]string, len(options))
	copy(opts, options)

	return ChipComboBox{
		chips:      NewChipList(),
		combo:      NewComboBox(opts).WithAllowNew(true, "New label: %s"),
		allOptions: opts,
		Width:      50,
		focused:    false,
	}
}

// WithWidth sets the display width.
func (c ChipComboBox) WithWidth(w int) ChipComboBox {
	c.Width = w
	c.chips = c.chips.WithWidth(w)
	c.combo = c.combo.WithWidth(w)
	return c
}

// WithMaxVisible sets the maximum visible items in dropdown.
func (c ChipComboBox) WithMaxVisible(n int) ChipComboBox {
	c.combo = c.combo.WithMaxVisible(n)
	return c
}

// WithPlaceholder sets the placeholder text.
func (c ChipComboBox) WithPlaceholder(s string) ChipComboBox {
	c.combo = c.combo.WithPlaceholder(s)
	return c
}

// WithAllowNew enables/disables creating new values not in Options.
func (c ChipComboBox) WithAllowNew(allow bool, label string) ChipComboBox {
	c.combo = c.combo.WithAllowNew(allow, label)
	return c
}

// Init implements tea.Model.
func (c ChipComboBox) Init() tea.Cmd {
	return nil
}

// Update handles messages and returns updated state.
func (c ChipComboBox) Update(msg tea.Msg) (ChipComboBox, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle flash clear message
	if _, ok := msg.(chipFlashClearMsg); ok {
		c.chips, _ = c.chips.Update(msg)
		return c, nil
	}

	// Handle messages from ChipList when in nav mode
	switch msg := msg.(type) {
	case ChipNavExitMsg:
		// Chip navigation exited
		switch msg.Reason {
		case ChipNavExitTyping:
			// Forward character to combo
			var cmd tea.Cmd
			c.combo, cmd = c.combo.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{msg.Character}})
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		case ChipNavExitTab:
			// Signal parent to move to next field
			return c, func() tea.Msg { return ChipComboBoxTabMsg{} }
		}
		return c, tea.Batch(cmds...)

	case ChipRemovedMsg:
		// Chip was removed, restore option to dropdown
		c.updateAvailableOptions()
		// Pass through the message
		return c, func() tea.Msg { return msg }

	case ComboBoxValueSelectedMsg:
		// Value selected from dropdown
		return c.handleSelection(msg)
	}

	// If in chip nav mode, route to ChipList
	if c.chips.InNavigationMode() {
		var cmd tea.Cmd
		c.chips, cmd = c.chips.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return c, tea.Batch(cmds...)
	}

	// Handle key messages for chip nav entry
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// Intercept ↑ for chip nav entry (chips are above input)
		if keyMsg.Type == tea.KeyUp &&
			c.combo.InputValue() == "" &&
			!c.combo.IsDropdownOpen() &&
			len(c.chips.Chips) > 0 {
			c.chips.EnterNavigation()
			return c, nil
		}

		// Intercept Tab to potentially add chip and signal move
		if keyMsg.Type == tea.KeyTab {
			return c.handleTab()
		}
	}

	// Route to ComboBox
	var cmd tea.Cmd
	c.combo, cmd = c.combo.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return c, tea.Batch(cmds...)
}

func (c ChipComboBox) handleSelection(msg ComboBoxValueSelectedMsg) (ChipComboBox, tea.Cmd) {
	label := msg.Value
	if label == "" {
		return c, nil
	}

	// Check for duplicate
	if c.chips.Contains(label) {
		// Trigger flash, don't add
		c.chips.AddChip(label) // Returns false, sets flashIndex
		c.combo.SetValue("")
		return c, FlashCmd()
	}

	// Add chip
	c.chips.AddChip(label)
	c.combo.SetValue("")
	c.updateAvailableOptions()

	return c, func() tea.Msg {
		return ChipComboBoxChipAddedMsg{Label: label, IsNew: msg.IsNew}
	}
}

func (c ChipComboBox) handleTab() (ChipComboBox, tea.Cmd) {
	var cmds []tea.Cmd

	// If dropdown is open, forward Tab to ComboBox to select highlighted item
	// This handles ghost text completion correctly
	if c.combo.IsDropdownOpen() {
		// Forward Tab to ComboBox which will call selectHighlightedOrNew()
		var cmd tea.Cmd
		c.combo, cmd = c.combo.Update(tea.KeyMsg{Type: tea.KeyTab})
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		// Signal move to next field
		cmds = append(cmds, func() tea.Msg { return ChipComboBoxTabMsg{} })
		return c, tea.Batch(cmds...)
	}

	// Dropdown closed - if there's text in the input, try to add it as a chip
	inputVal := strings.TrimSpace(c.combo.InputValue())
	if inputVal != "" {
		// Check for duplicate
		if c.chips.Contains(inputVal) {
			c.chips.AddChip(inputVal) // Flash
			c.combo.SetValue("")
			cmds = append(cmds, FlashCmd())
		} else {
			// Determine if it's a new value
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
			cmds = append(cmds, func() tea.Msg {
				return ChipComboBoxChipAddedMsg{Label: inputVal, IsNew: isNew}
			})
		}
	}

	// Signal move to next field
	cmds = append(cmds, func() tea.Msg { return ChipComboBoxTabMsg{} })
	return c, tea.Batch(cmds...)
}

func (c *ChipComboBox) updateAvailableOptions() {
	available := []string{}
	for _, opt := range c.allOptions {
		if !c.chips.Contains(opt) {
			available = append(available, opt)
		}
	}
	c.combo.SetOptions(available)
}

// View renders the chip combo box.
func (c ChipComboBox) View() string {
	var elements []string

	// Get styled chips from ChipList
	chips := c.chips.RenderChips()
	if len(chips) == 0 {
		// Empty state indicator when no chips selected
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Italic(true)
		elements = append(elements, emptyStyle.Render("No labels"))
	} else {
		elements = append(elements, chips...)
	}

	// Always show input box (even during chip nav for visual continuity)
	elements = append(elements, c.combo.View())

	// Word wrap all elements together
	return c.wrapElements(elements)
}

func (c ChipComboBox) wrapElements(elements []string) string {
	if c.Width <= 0 || len(elements) == 0 {
		return strings.Join(elements, " ")
	}

	var lines []string
	var currentLine []string
	currentWidth := 0

	for _, elem := range elements {
		elemWidth := lipgloss.Width(elem)
		spaceNeeded := elemWidth
		if len(currentLine) > 0 {
			spaceNeeded++ // +1 for space separator
		}

		if currentWidth+spaceNeeded > c.Width && len(currentLine) > 0 {
			// Start new line
			lines = append(lines, strings.Join(currentLine, " "))
			currentLine = []string{elem}
			currentWidth = elemWidth
		} else {
			currentLine = append(currentLine, elem)
			currentWidth += spaceNeeded
		}
	}

	if len(currentLine) > 0 {
		lines = append(lines, strings.Join(currentLine, " "))
	}

	return strings.Join(lines, "\n")
}

// GetChips returns the selected chip labels.
func (c ChipComboBox) GetChips() []string {
	return c.chips.GetChips()
}

// SetChips sets the selected chips.
func (c *ChipComboBox) SetChips(chips []string) {
	c.chips.Chips = nil
	for _, chip := range chips {
		c.chips.AddChip(chip)
	}
	c.updateAvailableOptions()
}

// SetOptions updates the available options.
func (c *ChipComboBox) SetOptions(options []string) {
	c.allOptions = make([]string, len(options))
	copy(c.allOptions, options)
	c.updateAvailableOptions()
}

// Focus focuses the chip combo box.
func (c *ChipComboBox) Focus() tea.Cmd {
	c.focused = true
	return c.combo.Focus()
}

// Blur removes focus from the chip combo box.
func (c *ChipComboBox) Blur() {
	c.focused = false
	c.chips.Blur()
	c.combo.Blur()
}

// Focused returns whether the chip combo box is focused.
func (c ChipComboBox) Focused() bool {
	return c.focused
}

// InChipNavMode returns whether chip navigation is active.
func (c ChipComboBox) InChipNavMode() bool {
	return c.chips.InNavigationMode()
}

// IsDropdownOpen returns whether the dropdown is visible.
func (c ChipComboBox) IsDropdownOpen() bool {
	return c.combo.IsDropdownOpen()
}

// ChipCount returns the number of selected chips.
func (c ChipComboBox) ChipCount() int {
	return len(c.chips.Chips)
}

// InputValue returns the current input text (for testing).
func (c ChipComboBox) InputValue() string {
	return c.combo.InputValue()
}

// FlashIndex returns the flash index (for testing).
func (c ChipComboBox) FlashIndex() int {
	return c.chips.FlashIndex()
}
