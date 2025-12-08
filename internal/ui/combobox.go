package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"abacus/internal/ui/theme"
)

// ComboBoxState represents the current state of the combo box.
type ComboBoxState int

const (
	// ComboBoxIdle - focused, dropdown closed.
	ComboBoxIdle ComboBoxState = iota
	// ComboBoxBrowsing - dropdown open with full list.
	ComboBoxBrowsing
	// ComboBoxFiltering - dropdown open with filtered list.
	ComboBoxFiltering
)

// ComboBoxEnterSelectedMsg is sent when Enter confirms a selection.
// The component stays focused for additional input.
type ComboBoxEnterSelectedMsg struct {
	Value string
	IsNew bool // True if value was created (not in original options)
}

// ComboBoxTabSelectedMsg is sent when Tab confirms a selection.
// Signals that the component should advance to next field after processing.
type ComboBoxTabSelectedMsg struct {
	Value string
	IsNew bool // True if value was created (not in original options)
}

// ComboBox implements a single-select autocomplete field following
// the "Selection Follows Visual Focus" contract from the spec.
type ComboBox struct {
	// Configuration (set at creation)
	Options      []string // All available options
	Placeholder  string   // Placeholder when empty
	Width        int      // Display width
	MaxVisible   int      // Max items in dropdown (default 5)
	AllowNew     bool     // Allow creating new values not in Options
	NewItemLabel string   // e.g., "New Assignee Added: %s"

	// Current state
	state           ComboBoxState
	textInput       textinput.Model // Embedded bubbles textinput
	value           string          // Committed/selected value
	originalValue   string          // For Esc revert
	filteredOptions []string        // Current filtered list
	highlightIndex  int             // Currently highlighted item in filteredOptions
	scrollOffset    int             // First visible item index for scrolling
	focused         bool            // Is this component focused
}

// NewComboBox creates a new ComboBox with the given options.
func NewComboBox(options []string) ComboBox {
	ti := textinput.New()
	ti.CharLimit = 100

	c := ComboBox{
		Options:         options,
		Placeholder:     "",
		Width:           40,
		MaxVisible:      5,
		AllowNew:        false,
		NewItemLabel:    "New item added: %s",
		state:           ComboBoxIdle,
		textInput:       ti,
		value:           "",
		originalValue:   "",
		filteredOptions: options,
		highlightIndex:  0,
		focused:         false,
	}
	c.textInput.Width = c.Width - 4 // Account for border padding
	return c
}

// WithPlaceholder sets the placeholder text.
func (c ComboBox) WithPlaceholder(s string) ComboBox {
	c.Placeholder = s
	c.textInput.Placeholder = s
	return c
}

// WithWidth sets the display width.
func (c ComboBox) WithWidth(w int) ComboBox {
	c.Width = w
	c.textInput.Width = w - 4
	return c
}

// WithMaxVisible sets the maximum visible items in dropdown.
func (c ComboBox) WithMaxVisible(n int) ComboBox {
	c.MaxVisible = n
	return c
}

// WithAllowNew enables creating new values not in Options.
func (c ComboBox) WithAllowNew(allow bool, label string) ComboBox {
	c.AllowNew = allow
	if label != "" {
		c.NewItemLabel = label
	}
	return c
}

// Init implements tea.Model.
func (c ComboBox) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (c ComboBox) Update(msg tea.Msg) (ComboBox, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return c.handleKeyMsg(msg)
	}

	// Pass through other messages to textinput
	var cmd tea.Cmd
	c.textInput, cmd = c.textInput.Update(msg)
	return c, cmd
}

func (c ComboBox) handleKeyMsg(msg tea.KeyMsg) (ComboBox, tea.Cmd) {
	switch c.state {
	case ComboBoxIdle:
		return c.handleIdleKey(msg)
	case ComboBoxBrowsing:
		return c.handleBrowsingKey(msg)
	case ComboBoxFiltering:
		return c.handleFilteringKey(msg)
	}
	return c, nil
}

func (c ComboBox) handleIdleKey(msg tea.KeyMsg) (ComboBox, tea.Cmd) {
	switch msg.Type {
	case tea.KeyDown:
		// Open dropdown with full list
		c.state = ComboBoxBrowsing
		c.filteredOptions = c.Options
		c.highlightCurrentValue()
		return c, nil

	case tea.KeyEnter:
		// Keep current value, confirm field
		return c, nil

	case tea.KeyTab:
		// Keep current value, signal move to next field
		return c, nil

	case tea.KeyEsc:
		// Two-stage escape: if we have typed text different from value, revert
		if c.textInput.Value() != c.value {
			c.textInput.SetValue(c.value)
			return c, nil
		}
		// Otherwise passthrough (let parent handle)
		return c, nil

	default:
		// Typing opens filtered dropdown
		if msg.Type == tea.KeyRunes || msg.Type == tea.KeyBackspace {
			// When typing starts with an existing value, clear input first
			// This follows VS Code IntelliSense behavior where typing replaces selection
			if c.value != "" && c.textInput.Value() == c.value {
				c.textInput.SetValue("")
			}
			oldValue := c.textInput.Value()
			var cmd tea.Cmd
			c.textInput, cmd = c.textInput.Update(msg)
			if c.textInput.Value() != oldValue {
				c.state = ComboBoxFiltering
				c.filterOptions() // filterOptions now handles highlightIndex
			}
			return c, cmd
		}
	}

	var cmd tea.Cmd
	c.textInput, cmd = c.textInput.Update(msg)
	return c, cmd
}

func (c ComboBox) handleBrowsingKey(msg tea.KeyMsg) (ComboBox, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		if c.highlightIndex > 0 {
			c.highlightIndex--
			c.adjustScrollOffset()
		}
		return c, nil

	case tea.KeyDown:
		if c.highlightIndex < len(c.filteredOptions)-1 {
			c.highlightIndex++
			c.adjustScrollOffset()
		}
		return c, nil

	case tea.KeyEnter:
		return c.selectHighlightedWithEnter()

	case tea.KeyTab:
		return c.selectHighlightedWithTab()

	case tea.KeyEsc:
		// Close dropdown
		c.state = ComboBoxIdle
		return c, nil

	default:
		// Typing switches to filtering mode
		if msg.Type == tea.KeyRunes || msg.Type == tea.KeyBackspace {
			c.state = ComboBoxFiltering
			var cmd tea.Cmd
			c.textInput, cmd = c.textInput.Update(msg)
			c.filterOptions() // filterOptions now handles highlightIndex
			return c, cmd
		}
	}

	return c, nil
}

func (c ComboBox) handleFilteringKey(msg tea.KeyMsg) (ComboBox, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		if c.highlightIndex > 0 {
			c.highlightIndex--
			c.adjustScrollOffset()
		}
		return c, nil

	case tea.KeyDown:
		if c.highlightIndex < len(c.filteredOptions)-1 {
			c.highlightIndex++
			c.adjustScrollOffset()
		}
		return c, nil

	case tea.KeyEnter:
		return c.selectHighlightedOrNewWithEnter()

	case tea.KeyTab:
		return c.selectHighlightedOrNewWithTab()

	case tea.KeyEsc:
		// First Esc: close dropdown, keep typed text
		c.state = ComboBoxIdle
		return c, nil

	case tea.KeyDelete:
		// Delete with ghost text visible: reject autocomplete (clear highlight)
		// This lets user keep their typed text without accepting the suggestion
		if c.HasGhostText() {
			c.highlightIndex = -1
			return c, nil
		}
		// No ghost text - ignore delete (nothing to reject)
		return c, nil

	default:
		// Continue filtering
		if msg.Type == tea.KeyRunes || msg.Type == tea.KeyBackspace {
			var cmd tea.Cmd
			c.textInput, cmd = c.textInput.Update(msg)
			c.filterOptions() // filterOptions now handles highlightIndex
			return c, cmd
		}
	}

	return c, nil
}

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

func (c *ComboBox) filterOptions() {
	input := strings.ToLower(c.textInput.Value())
	if input == "" {
		c.filteredOptions = c.Options
		c.scrollOffset = 0
		return
	}
	c.filteredOptions = nil
	exactMatchIdx := -1
	for _, opt := range c.Options {
		lower := strings.ToLower(opt)
		if strings.Contains(lower, input) {
			if lower == input && exactMatchIdx == -1 {
				exactMatchIdx = len(c.filteredOptions)
			}
			c.filteredOptions = append(c.filteredOptions, opt)
		}
	}
	// Reset scroll when filter changes
	c.scrollOffset = 0
	// Highlight exact match if found, otherwise first match
	if exactMatchIdx >= 0 {
		c.highlightIndex = exactMatchIdx
	} else {
		c.highlightIndex = 0
	}
}

func (c *ComboBox) highlightCurrentValue() {
	if c.value == "" {
		c.highlightIndex = 0
		c.scrollOffset = 0
		return
	}
	for i, opt := range c.filteredOptions {
		if opt == c.value {
			c.highlightIndex = i
			c.adjustScrollOffset()
			return
		}
	}
	c.highlightIndex = 0 // Value not in list
	c.scrollOffset = 0
}

// adjustScrollOffset ensures the highlighted item is visible in the dropdown window.
func (c *ComboBox) adjustScrollOffset() {
	// If highlight is above visible window, scroll up
	if c.highlightIndex < c.scrollOffset {
		c.scrollOffset = c.highlightIndex
	}
	// If highlight is below visible window, scroll down
	if c.highlightIndex >= c.scrollOffset+c.MaxVisible {
		c.scrollOffset = c.highlightIndex - c.MaxVisible + 1
	}
	// Clamp scroll offset
	if c.scrollOffset < 0 {
		c.scrollOffset = 0
	}
	maxOffset := len(c.filteredOptions) - c.MaxVisible
	if maxOffset < 0 {
		maxOffset = 0
	}
	if c.scrollOffset > maxOffset {
		c.scrollOffset = maxOffset
	}
}

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

	inputStyle := styleComboBoxInput().Width(c.Width)
	if c.focused {
		inputStyle = styleComboBoxInputFocused().Width(c.Width)
	}
	b.WriteString(inputStyle.Render(inputView))

	// Render dropdown if open
	if c.state != ComboBoxIdle {
		b.WriteString("\n")
		if len(c.filteredOptions) == 0 {
			if c.AllowNew && strings.TrimSpace(c.textInput.Value()) != "" {
				// No matches state with AllowNew
				b.WriteString(styleComboBoxNoMatch().Render("  No matches"))
				b.WriteString("\n")
				b.WriteString(styleComboBoxHint().Render("  \u23ce to add new"))
			} else {
				b.WriteString(styleComboBoxNoMatch().Render("  No matches"))
			}
		} else {
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

			// Render only visible items
			for i := c.scrollOffset; i < endIndex; i++ {
				opt := c.filteredOptions[i]
				// Style "Unassigned" as muted per spec Section 8
				isMuted := opt == "Unassigned"
				if i == c.highlightIndex {
					if isMuted {
						b.WriteString(styleComboBoxHighlight().Foreground(theme.Current().BorderNormal()).Render("\u25b8 " + opt))
					} else {
						b.WriteString(styleComboBoxHighlight().Render("\u25b8 " + opt))
					}
				} else {
					if isMuted {
						b.WriteString(styleComboBoxOption().Foreground(theme.Current().BorderNormal()).Render("  " + opt))
					} else {
						b.WriteString(styleComboBoxOption().Render("  " + opt))
					}
				}
				if i < endIndex-1 {
					b.WriteString("\n")
				}
			}

			// Show scroll-down indicator if there are items below
			if endIndex < len(c.filteredOptions) {
				b.WriteString("\n")
				b.WriteString(styleComboBoxHint().Render("  ▼ more below"))
			}
		}
	}

	return b.String()
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
		PaddingLeft(1)
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
