package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keyboard shortcuts for the application.
// Each binding includes the actual keys and help text for display.
// Note: Related bindings (Up/Down, Left/Right) share identical help text
// since they appear as a single row in the help overlay.
type KeyMap struct {
	// Navigation
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Space    key.Binding
	Home     key.Binding
	End      key.Binding
	PageUp   key.Binding
	PageDown key.Binding

	// Actions
	Enter     key.Binding
	Tab       key.Binding
	Refresh   key.Binding
	Error     key.Binding
	Help      key.Binding
	Quit      key.Binding
	Copy      key.Binding
	Status    key.Binding
	StartWork key.Binding
	CloseBead key.Binding
	Labels    key.Binding
	NewBead   key.Binding

	// Search
	Search    key.Binding
	Escape    key.Binding
	ShiftTab  key.Binding
	Backspace key.Binding
}

// DefaultKeyMap returns the default keybindings for Abacus.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Navigation - Up/Down share help text (displayed as single row)
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/↓  j/k", "Move up/down"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↑/↓  j/k", "Move up/down"),
		),
		// Left/Right share help text (displayed as single row)
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/→  h/l", "Collapse/Expand"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("←/→  h/l", "Collapse/Expand"),
		),
		Space: key.NewBinding(
			key.WithKeys(" ", "space"),
			key.WithHelp("Space", "Toggle expand"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("Home  g", "Jump to top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("End   G", "Jump to bottom"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+b"),
			key.WithHelp("PgUp  Ctrl+B", "Page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+f"),
			key.WithHelp("PgDn  Ctrl+F", "Page down"),
		),

		// Actions
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("⏎ (Enter)", "Toggle detail"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("⇥ (Tab)", "Switch focus"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "Refresh"),
		),
		Error: key.NewBinding(
			key.WithKeys("!"),
			key.WithHelp("!", "Error details"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "Help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "Quit"),
		),
		Copy: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "Copy ID"),
		),
		Status: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "Change status"),
		),
		StartWork: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "Start work"),
		),
		CloseBead: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "Close bead"),
		),
		Labels: key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("L", "Manage labels"),
		),
		NewBead: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "New bead"),
		),

		// Search
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "Start search"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("Esc", "Clear/cancel"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("⇧⇥", "Previous focus"),
		),
		Backspace: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("⌫", "Delete char"),
		),
	}
}
