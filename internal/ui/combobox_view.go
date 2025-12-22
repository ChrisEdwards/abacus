package ui

import (
	"strings"

	"abacus/internal/ui/theme"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// View implements tea.Model.
func (c ComboBox) View() string {
	var b strings.Builder

	// Build input view - may include inline ghost text
	var inputView string
	ghostText := c.GhostText()
	if ghostText != "" && c.focused {
		// Build custom view with ghost text in block cursor style
		// First ghost char is inside inverted block cursor (grey on bright bg)
		// Rest of ghost text in grey
		typed := c.textInput.Value()
		prompt := "> "
		firstGhostChar := string([]rune(ghostText)[0])
		restGhostText := ""
		if len([]rune(ghostText)) > 1 {
			restGhostText = string([]rune(ghostText)[1:])
		}
		// Inverted cursor: grey text on bright background
		cursorWithChar := styleGhostCursor().Render(firstGhostChar)
		inputView = prompt + typed + cursorWithChar + styleGhostText().Render(restGhostText)
	} else {
		inputView = c.textInput.View()
	}

	// c.Width is the desired VISUAL width including border
	// Border adds 2 chars outside Width, so use Width - 2 for lipgloss
	inputStyle := styleComboBoxInput().Width(c.Width - 2)
	if c.focused {
		inputStyle = styleComboBoxInputFocused().Width(c.Width - 2)
	}
	b.WriteString(inputStyle.Render(inputView))

	// Render dropdown if open
	if c.state != ComboBoxIdle {
		b.WriteString("\n")
		if len(c.filteredOptions) == 0 {
			b.WriteString(c.renderEmptyDropdown())
		} else {
			b.WriteString(c.renderDropdownItems())
		}
	}

	return b.String()
}

// renderEmptyDropdown renders the dropdown when no options match.
func (c ComboBox) renderEmptyDropdown() string {
	var b strings.Builder
	if c.AllowNew && strings.TrimSpace(c.textInput.Value()) != "" {
		// No matches state with AllowNew
		b.WriteString(styleComboBoxNoMatch().Render("  No matches"))
		b.WriteString("\n")
		b.WriteString(styleComboBoxHint().Render("  \u23ce to add new"))
	} else {
		b.WriteString(styleComboBoxNoMatch().Render("  No matches"))
	}
	return b.String()
}

// renderDropdownItems renders the visible dropdown items with scrolling.
func (c ComboBox) renderDropdownItems() string {
	var b strings.Builder

	// Show scroll-up indicator if there are items above
	if c.scrollOffset > 0 {
		b.WriteString(styleComboBoxHint().Render("  ▲ more above"))
		b.WriteString("\n")
	}

	// Calculate visible window
	endIndex := c.scrollOffset + c.MaxVisible
	if endIndex > len(c.filteredOptions) {
		endIndex = len(c.filteredOptions)
	}

	// Calculate content width for dropdown (Width minus border/padding of 4)
	dropdownContentWidth := c.Width - 4
	if dropdownContentWidth < 10 {
		dropdownContentWidth = 10 // Minimum reasonable width
	}

	for i := c.scrollOffset; i < endIndex; i++ {
		opt := c.filteredOptions[i]
		// Style "Unassigned" as muted per spec Section 8
		isMuted := opt == "Unassigned"

		// Truncate option text to fit width (account for 2-char prefix "▸ " or "  ")
		displayOpt := opt
		maxOptLen := dropdownContentWidth - 2 // 2 chars for prefix
		if len(displayOpt) > maxOptLen && maxOptLen > 3 {
			displayOpt = displayOpt[:maxOptLen-3] + "..."
		}

		b.WriteString(c.renderDropdownItem(displayOpt, i, isMuted, dropdownContentWidth))
		if i < endIndex-1 {
			b.WriteString("\n")
		}
	}

	// Show scroll-down indicator if there are items below
	if endIndex < len(c.filteredOptions) {
		b.WriteString("\n")
		b.WriteString(styleComboBoxHint().Render("  ▼ more below"))
	}

	return b.String()
}

// renderDropdownItem renders a single dropdown item.
func (c ComboBox) renderDropdownItem(displayOpt string, index int, isMuted bool, width int) string {
	if index == c.highlightIndex {
		if isMuted {
			return styleComboBoxHighlight().Width(width).Foreground(theme.Current().BorderNormal()).Render("\u25b8 " + displayOpt)
		}
		return styleComboBoxHighlight().Width(width).Render("\u25b8 " + displayOpt)
	}
	if isMuted {
		return styleComboBoxOption().Width(width).Foreground(theme.Current().BorderNormal()).Render("  " + displayOpt)
	}
	return styleComboBoxOption().Width(width).Render("  " + displayOpt)
}

// ComboBox styles

func styleComboBoxInput() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().BorderDim()).
		Padding(0, 1)
}

func styleComboBoxInputFocused() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().Secondary()).
		Padding(0, 1)
}

func styleComboBoxOption() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Text()).
		PaddingLeft(2)
}

func styleComboBoxHighlight() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Secondary()).
		Bold(true).
		PaddingLeft(2)
}

func styleComboBoxNoMatch() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().BorderNormal()).
		Italic(true)
}

func styleComboBoxHint() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().TextMuted())
}

func styleGhostText() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().TextMuted())
}

// styleGhostCursor: grey text on bright background (inverted block cursor)
func styleGhostCursor() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().TextMuted()).
		Background(theme.Current().TextMuted())
}

// Value returns the current selected value.
func (c ComboBox) Value() string {
	return c.value
}

// SetValue sets the current value.
func (c *ComboBox) SetValue(v string) {
	c.value = v
	c.originalValue = v
	c.textInput.SetValue(v)
}

// SetOptions updates the available options.
func (c *ComboBox) SetOptions(opts []string) {
	c.Options = opts
	c.filteredOptions = opts
}

// Focus focuses the combo box and returns a blink command.
func (c *ComboBox) Focus() tea.Cmd {
	c.focused = true
	c.originalValue = c.value
	return c.textInput.Focus()
}

// Blur removes focus from the combo box.
func (c *ComboBox) Blur() {
	c.focused = false
	c.state = ComboBoxIdle
	c.textInput.Blur()
}

// Focused returns whether the combo box is focused.
func (c ComboBox) Focused() bool {
	return c.focused
}

// IsDropdownOpen returns whether the dropdown is currently visible.
func (c ComboBox) IsDropdownOpen() bool {
	return c.state != ComboBoxIdle
}

// State returns the current state for testing.
func (c ComboBox) State() ComboBoxState {
	return c.state
}

// FilteredOptions returns the current filtered options for testing.
func (c ComboBox) FilteredOptions() []string {
	return c.filteredOptions
}

// HighlightIndex returns the current highlight index for testing.
func (c ComboBox) HighlightIndex() int {
	return c.highlightIndex
}

// InputValue returns the current text input value for testing.
func (c ComboBox) InputValue() string {
	return c.textInput.Value()
}

// GhostText returns the autocomplete ghost text if applicable.
// Returns empty string if no ghost text should be shown.
func (c ComboBox) GhostText() string {
	// Only show ghost text when filtering and have a highlighted match
	if c.state != ComboBoxFiltering {
		return ""
	}
	if c.highlightIndex < 0 || c.highlightIndex >= len(c.filteredOptions) {
		return ""
	}

	typed := c.textInput.Value()
	// Don't show ghost text when input is empty
	if typed == "" {
		return ""
	}

	highlighted := c.filteredOptions[c.highlightIndex]

	// Check if highlighted option starts with typed text (case-insensitive)
	typedLower := strings.ToLower(typed)
	highlightedLower := strings.ToLower(highlighted)
	if !strings.HasPrefix(highlightedLower, typedLower) {
		return ""
	}

	// Return the completion portion
	return highlighted[len(typed):]
}

// HasGhostText returns whether ghost text is currently visible.
func (c ComboBox) HasGhostText() bool {
	return c.GhostText() != ""
}

// ClearHighlight deselects the current highlight (reject autocomplete).
func (c *ComboBox) ClearHighlight() {
	c.highlightIndex = -1
}
