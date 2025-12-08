package ui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"abacus/internal/ui/theme"
)

// ChipListState represents the current mode of the chip list.
type ChipListState int

const (
	// ChipListInput - normal mode, cursor after chips.
	ChipListInput ChipListState = iota
	// ChipListNavigation - navigating chips with arrows.
	ChipListNavigation
)

// ChipNavExitReason indicates why chip navigation mode was exited.
type ChipNavExitReason int

const (
	// ChipNavExitRight - → pressed past last chip.
	ChipNavExitRight ChipNavExitReason = iota
	// ChipNavExitEscape - Esc pressed.
	ChipNavExitEscape
	// ChipNavExitTab - Tab pressed.
	ChipNavExitTab
	// ChipNavExitTyping - letter key pressed (Character field has the key).
	ChipNavExitTyping
)

// ChipRemovedMsg is sent when a chip is deleted via navigation.
type ChipRemovedMsg struct {
	Label string
	Index int
}

// ChipNavExitMsg signals the chip list wants to exit navigation mode.
type ChipNavExitMsg struct {
	Reason    ChipNavExitReason
	Character rune // For ExitTyping: the key that was pressed
}

// chipFlashClearMsg is sent to clear the flash state.
type chipFlashClearMsg struct{}

// ChipList manages a collection of label chips with navigation support.
type ChipList struct {
	// Configuration
	Width int // Available width for word wrapping (default 40)

	// State
	Chips      []string // Currently selected labels
	state      ChipListState
	navIndex   int  // Highlighted chip index (-1 = none)
	focused    bool
	flashIndex int // Index of chip to flash for duplicate (-1 = none)
}

// NewChipList creates a new ChipList.
func NewChipList() ChipList {
	return ChipList{
		Width:      40,
		Chips:      nil,
		state:      ChipListInput,
		navIndex:   -1,
		focused:    false,
		flashIndex: -1,
	}
}

// WithWidth sets the available width for word wrapping.
func (c ChipList) WithWidth(w int) ChipList {
	c.Width = w
	return c
}

// Init implements tea.Model-like interface.
func (c ChipList) Init() tea.Cmd {
	return nil
}

// Update handles messages and returns updated state.
func (c ChipList) Update(msg tea.Msg) (ChipList, tea.Cmd) {
	switch msg := msg.(type) {
	case chipFlashClearMsg:
		c.flashIndex = -1
		return c, nil

	case tea.KeyMsg:
		if c.state == ChipListNavigation {
			return c.handleNavigationKey(msg)
		}
	}

	return c, nil
}

func (c ChipList) handleNavigationKey(msg tea.KeyMsg) (ChipList, tea.Cmd) {
	switch msg.Type {
	case tea.KeyLeft:
		// Move to previous chip, stop at first
		if c.navIndex > 0 {
			c.navIndex--
		}
		return c, nil

	case tea.KeyRight:
		// Move to next chip, or exit if past last
		if c.navIndex < len(c.Chips)-1 {
			c.navIndex++
			return c, nil
		}
		// Past last chip - exit navigation
		c.state = ChipListInput
		c.navIndex = -1
		return c, func() tea.Msg {
			return ChipNavExitMsg{Reason: ChipNavExitRight}
		}

	case tea.KeyDown:
		// Down arrow exits chip nav back to text box (chips are above input)
		c.state = ChipListInput
		c.navIndex = -1
		return c, func() tea.Msg {
			return ChipNavExitMsg{Reason: ChipNavExitRight} // Reuse Right exit reason
		}

	case tea.KeyBackspace, tea.KeyDelete:
		if len(c.Chips) == 0 || c.navIndex < 0 || c.navIndex >= len(c.Chips) {
			return c, nil
		}
		removed := c.Chips[c.navIndex]
		removedIndex := c.navIndex

		// Remove the chip
		c.Chips = append(c.Chips[:c.navIndex], c.Chips[c.navIndex+1:]...)

		// Adjust navIndex based on position
		if len(c.Chips) == 0 {
			// No chips left - exit navigation
			c.state = ChipListInput
			c.navIndex = -1
			return c, func() tea.Msg {
				return ChipRemovedMsg{Label: removed, Index: removedIndex}
			}
		} else if c.navIndex >= len(c.Chips) {
			// Was at last position, highlight previous
			c.navIndex = len(c.Chips) - 1
		}
		// Otherwise stay at same index (next item slid into place)

		return c, func() tea.Msg {
			return ChipRemovedMsg{Label: removed, Index: removedIndex}
		}

	case tea.KeyEsc:
		c.state = ChipListInput
		c.navIndex = -1
		return c, func() tea.Msg {
			return ChipNavExitMsg{Reason: ChipNavExitEscape}
		}

	case tea.KeyTab:
		c.state = ChipListInput
		c.navIndex = -1
		return c, func() tea.Msg {
			return ChipNavExitMsg{Reason: ChipNavExitTab}
		}

	case tea.KeyRunes:
		// Letter key - exit and pass character
		if len(msg.Runes) > 0 {
			c.state = ChipListInput
			c.navIndex = -1
			char := msg.Runes[0]
			return c, func() tea.Msg {
				return ChipNavExitMsg{Reason: ChipNavExitTyping, Character: char}
			}
		}
	}

	return c, nil
}

// View renders the chip list.
func (c ChipList) View() string {
	if len(c.Chips) == 0 {
		return ""
	}

	var renderedChips []string
	for i, chip := range c.Chips {
		var chipStr string
		if c.flashIndex == i {
			chipStr = renderPillChip(chip, chipStateFlash)
		} else if c.state == ChipListNavigation && i == c.navIndex {
			chipStr = renderPillChip(chip, chipStateHighlight)
		} else {
			chipStr = renderPillChip(chip, chipStateNormal)
		}
		renderedChips = append(renderedChips, chipStr)
	}

	return c.wrapChips(renderedChips)
}

func (c ChipList) wrapChips(renderedChips []string) string {
	if c.Width <= 0 {
		return strings.Join(renderedChips, " ")
	}

	var lines []string
	var currentLine []string
	currentWidth := 0

	for _, chip := range renderedChips {
		chipWidth := lipgloss.Width(chip)
		spaceNeeded := chipWidth
		if len(currentLine) > 0 {
			spaceNeeded++ // +1 for space separator
		}

		if currentWidth+spaceNeeded > c.Width && len(currentLine) > 0 {
			// Start new line
			lines = append(lines, strings.Join(currentLine, " "))
			currentLine = []string{chip}
			currentWidth = chipWidth
		} else {
			currentLine = append(currentLine, chip)
			currentWidth += spaceNeeded
		}
	}

	if len(currentLine) > 0 {
		lines = append(lines, strings.Join(currentLine, " "))
	}

	return strings.Join(lines, "\n")
}

// RenderChips returns styled chip strings without word wrapping.
// Used by ChipComboBox to combine chips with input before wrapping.
func (c ChipList) RenderChips() []string {
	var result []string
	for i, chip := range c.Chips {
		var chipStr string
		if c.flashIndex == i {
			chipStr = renderPillChip(chip, chipStateFlash)
		} else if c.state == ChipListNavigation && i == c.navIndex {
			chipStr = renderPillChip(chip, chipStateHighlight)
		} else {
			chipStr = renderPillChip(chip, chipStateNormal)
		}
		result = append(result, chipStr)
	}
	return result
}

// AddChip adds a label to the chip list. Returns false if duplicate.
func (c *ChipList) AddChip(label string) bool {
	label = strings.TrimSpace(label)
	if label == "" {
		return false
	}

	// Check for duplicate (case-insensitive)
	for i, existing := range c.Chips {
		if strings.EqualFold(existing, label) {
			c.flashIndex = i
			return false
		}
	}

	c.Chips = append(c.Chips, label)
	return true
}

// RemoveChip removes a label from the chip list.
func (c *ChipList) RemoveChip(label string) {
	for i, chip := range c.Chips {
		if strings.EqualFold(chip, label) {
			c.Chips = append(c.Chips[:i], c.Chips[i+1:]...)
			return
		}
	}
}

// RemoveHighlighted removes the currently highlighted chip.
// Returns the removed label, or empty string if not in navigation mode.
func (c *ChipList) RemoveHighlighted() string {
	if c.state != ChipListNavigation || c.navIndex < 0 || c.navIndex >= len(c.Chips) {
		return ""
	}

	removed := c.Chips[c.navIndex]
	c.Chips = append(c.Chips[:c.navIndex], c.Chips[c.navIndex+1:]...)

	// Adjust navIndex
	if len(c.Chips) == 0 {
		c.state = ChipListInput
		c.navIndex = -1
	} else if c.navIndex >= len(c.Chips) {
		c.navIndex = len(c.Chips) - 1
	}

	return removed
}

// Contains checks if a label exists in the chip list (case-insensitive).
func (c ChipList) Contains(label string) bool {
	for _, chip := range c.Chips {
		if strings.EqualFold(chip, label) {
			return true
		}
	}
	return false
}

// GetChips returns a copy of the chips slice.
func (c ChipList) GetChips() []string {
	result := make([]string, len(c.Chips))
	copy(result, c.Chips)
	return result
}

// EnterNavigation enters chip navigation mode, highlighting the last chip.
// Returns false if there are no chips to navigate.
func (c *ChipList) EnterNavigation() bool {
	if len(c.Chips) == 0 {
		return false
	}
	c.state = ChipListNavigation
	c.navIndex = len(c.Chips) - 1 // Highlight LAST chip
	return true
}

// ExitNavigation exits chip navigation mode.
func (c *ChipList) ExitNavigation() {
	c.state = ChipListInput
	c.navIndex = -1
}

// InNavigationMode returns true if in chip navigation mode.
func (c ChipList) InNavigationMode() bool {
	return c.state == ChipListNavigation
}

// HighlightedChip returns the currently highlighted chip label.
// Returns empty string if not in navigation mode.
func (c ChipList) HighlightedChip() string {
	if c.state != ChipListNavigation || c.navIndex < 0 || c.navIndex >= len(c.Chips) {
		return ""
	}
	return c.Chips[c.navIndex]
}

// HighlightedIndex returns the index of the highlighted chip.
// Returns -1 if not in navigation mode.
func (c ChipList) HighlightedIndex() int {
	if c.state != ChipListNavigation {
		return -1
	}
	return c.navIndex
}

// Focus focuses the chip list.
func (c *ChipList) Focus() {
	c.focused = true
}

// Blur removes focus and exits navigation mode.
func (c *ChipList) Blur() {
	c.focused = false
	c.state = ChipListInput
	c.navIndex = -1
}

// Focused returns whether the chip list is focused.
func (c ChipList) Focused() bool {
	return c.focused
}

// NavIndex returns the current navigation index (for testing).
func (c ChipList) NavIndex() int {
	return c.navIndex
}

// State returns the current state (for testing).
func (c ChipList) State() ChipListState {
	return c.state
}

// FlashIndex returns the current flash index (for testing).
func (c ChipList) FlashIndex() int {
	return c.flashIndex
}

// ClearFlash clears the flash state and returns a command to clear after delay.
func (c *ChipList) ClearFlash() tea.Cmd {
	c.flashIndex = -1
	return nil
}

// FlashCmd returns a command that will clear the flash after a delay.
func FlashCmd() tea.Cmd {
	return tea.Tick(flashDuration, func(_ time.Time) tea.Msg {
		return chipFlashClearMsg{}
	})
}

const flashDuration = 150 * time.Millisecond

// Chip visual states for pill rendering
type chipState int

const (
	chipStateNormal chipState = iota
	chipStateHighlight
	chipStateFlash
)

// Powerline characters for pill-shaped chips
const (
	pillLeft  = "\ue0b6" // Left half-circle (rounded left edge)
	pillRight = "\ue0b4" // Right half-circle (rounded right edge)
)

// renderPillChip renders a label as a pill-shaped chip using powerline glyphs.
// The pill has curved edges and a solid background color for the label text.
func renderPillChip(label string, state chipState) string {
	var bgColor, fgColor lipgloss.TerminalColor

	t := theme.Current()
	switch state {
	case chipStateHighlight:
		bgColor = t.BackgroundSecondary() // Purple for selection
		fgColor = t.Text()
	case chipStateFlash:
		bgColor = t.Warning() // Orange flash for duplicate
		fgColor = t.Text()
	default:
		bgColor = t.Info() // Try Info (cyan/teal) for visible labels
		fgColor = t.Background() // Use background color for contrast
	}

	// Left cap: foreground is the chip color (creates the curved colored edge)
	leftCap := lipgloss.NewStyle().Foreground(bgColor).Render(pillLeft)

	// Label text: colored background with contrasting text
	labelStyle := lipgloss.NewStyle().
		Foreground(fgColor).
		Background(bgColor)
	if state == chipStateHighlight || state == chipStateFlash {
		labelStyle = labelStyle.Bold(true)
	}
	labelText := labelStyle.Render(label)

	// Right cap: foreground is the chip color
	rightCap := lipgloss.NewStyle().Foreground(bgColor).Render(pillRight)

	return leftCap + labelText + rightCap
}
