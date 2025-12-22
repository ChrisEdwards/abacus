package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
