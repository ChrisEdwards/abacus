package ui

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	t.Run("HelpBinding", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
		if !key.Matches(msg, km.Help) {
			t.Error("expected ? to match Help binding")
		}
	})

	t.Run("NavigationBindings", func(t *testing.T) {
		// Up key
		if !key.Matches(tea.KeyMsg{Type: tea.KeyUp}, km.Up) {
			t.Error("expected up arrow to match Up binding")
		}
		if !key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}, km.Up) {
			t.Error("expected k to match Up binding")
		}

		// Down key
		if !key.Matches(tea.KeyMsg{Type: tea.KeyDown}, km.Down) {
			t.Error("expected down arrow to match Down binding")
		}
		if !key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, km.Down) {
			t.Error("expected j to match Down binding")
		}

		// Left key
		if !key.Matches(tea.KeyMsg{Type: tea.KeyLeft}, km.Left) {
			t.Error("expected left arrow to match Left binding")
		}
		if !key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}, km.Left) {
			t.Error("expected h to match Left binding")
		}

		// Right key
		if !key.Matches(tea.KeyMsg{Type: tea.KeyRight}, km.Right) {
			t.Error("expected right arrow to match Right binding")
		}
		if !key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}, km.Right) {
			t.Error("expected l to match Right binding")
		}
	})

	t.Run("EscapeBinding", func(t *testing.T) {
		if !key.Matches(tea.KeyMsg{Type: tea.KeyEscape}, km.Escape) {
			t.Error("expected escape to match Escape binding")
		}
	})

	t.Run("EnterBinding", func(t *testing.T) {
		if !key.Matches(tea.KeyMsg{Type: tea.KeyEnter}, km.Enter) {
			t.Error("expected enter to match Enter binding")
		}
	})

	t.Run("SearchBinding", func(t *testing.T) {
		if !key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}, km.Search) {
			t.Error("expected / to match Search binding")
		}
	})

	t.Run("ShiftTabBinding", func(t *testing.T) {
		if !key.Matches(tea.KeyMsg{Type: tea.KeyShiftTab}, km.ShiftTab) {
			t.Error("expected shift+tab to match ShiftTab binding")
		}
	})

	t.Run("BackspaceBinding", func(t *testing.T) {
		if !key.Matches(tea.KeyMsg{Type: tea.KeyBackspace}, km.Backspace) {
			t.Error("expected backspace to match Backspace binding")
		}
	})

	t.Run("NewBeadBinding", func(t *testing.T) {
		if !key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}}, km.NewBead) {
			t.Error("expected N (Shift+n) to match NewBead binding")
		}
	})

	t.Run("NewRootBeadBinding", func(t *testing.T) {
		if !key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}, km.NewRootBead) {
			t.Error("expected n to match NewRootBead binding")
		}
	})

	t.Run("EditBinding", func(t *testing.T) {
		if !key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}, km.Edit) {
			t.Error("expected e to match Edit binding")
		}
	})
}

func TestKeyBindingsHaveHelpText(t *testing.T) {
	km := DefaultKeyMap()

	bindings := []struct {
		name    string
		binding key.Binding
	}{
		{"Help", km.Help},
		{"Up", km.Up},
		{"Down", km.Down},
		{"Left", km.Left},
		{"Right", km.Right},
		{"Space", km.Space},
		{"Home", km.Home},
		{"End", km.End},
		{"PageUp", km.PageUp},
		{"PageDown", km.PageDown},
		{"Enter", km.Enter},
		{"Tab", km.Tab},
		{"Refresh", km.Refresh},
		{"Error", km.Error},
		{"Quit", km.Quit},
		{"Copy", km.Copy},
		{"Search", km.Search},
		{"Escape", km.Escape},
		{"ShiftTab", km.ShiftTab},
		{"Backspace", km.Backspace},
		{"NewBead", km.NewBead},
		{"NewRootBead", km.NewRootBead},
		{"Edit", km.Edit},
	}

	for _, b := range bindings {
		t.Run(b.name, func(t *testing.T) {
			help := b.binding.Help()
			if help.Key == "" {
				t.Errorf("%s binding has empty Key help text", b.name)
			}
			if help.Desc == "" {
				t.Errorf("%s binding has empty Desc help text", b.name)
			}
		})
	}
}

func TestRelatedBindingsShareHelpText(t *testing.T) {
	km := DefaultKeyMap()

	t.Run("UpDownShareHelpText", func(t *testing.T) {
		if km.Up.Help().Key != km.Down.Help().Key {
			t.Errorf("Up and Down should share Key help text: %q vs %q",
				km.Up.Help().Key, km.Down.Help().Key)
		}
		if km.Up.Help().Desc != km.Down.Help().Desc {
			t.Errorf("Up and Down should share Desc help text: %q vs %q",
				km.Up.Help().Desc, km.Down.Help().Desc)
		}
	})

	t.Run("LeftRightShareHelpText", func(t *testing.T) {
		if km.Left.Help().Key != km.Right.Help().Key {
			t.Errorf("Left and Right should share Key help text: %q vs %q",
				km.Left.Help().Key, km.Right.Help().Key)
		}
		if km.Left.Help().Desc != km.Right.Help().Desc {
			t.Errorf("Left and Right should share Desc help text: %q vs %q",
				km.Left.Help().Desc, km.Right.Help().Desc)
		}
	})
}
