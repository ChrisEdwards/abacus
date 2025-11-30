// Demo program to visually test the ComboBox component
package main

import (
	"fmt"
	"os"

	"abacus/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	combo    ui.ComboBox
	selected string
	isNew    bool
	quit     bool
}

func initialModel() model {
	options := []string{
		"Alice",
		"Bob",
		"Carlos",
		"Diana",
		"Edward",
		"Fiona",
		"George",
	}

	cb := ui.NewComboBox(options).
		WithPlaceholder("Select or type name...").
		WithWidth(40).
		WithMaxVisible(5).
		WithAllowNew(true, "New assignee: %s")

	cb.Focus()

	return model{
		combo: cb,
	}
}

func (m model) Init() tea.Cmd {
	return m.combo.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if !m.combo.IsDropdownOpen() {
				m.quit = true
				return m, tea.Quit
			}
		}

	case ui.ComboBoxValueSelectedMsg:
		m.selected = msg.Value
		m.isNew = msg.IsNew
	}

	var cmd tea.Cmd
	m.combo, cmd = m.combo.Update(msg)
	return m, cmd
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)

	newBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)
)

func (m model) View() string {
	if m.quit {
		return ""
	}

	s := titleStyle.Render("ComboBox Demo")
	s += "\n\n"
	s += "Assignee:\n"
	s += m.combo.View()
	s += "\n\n"

	if m.selected != "" {
		s += "Selected: " + selectedStyle.Render(m.selected)
		if m.isNew {
			s += " " + newBadge.Render("(NEW)")
		}
		s += "\n"
	}

	s += helpStyle.Render("\n↓ open • type to filter • Enter select • Esc close • q quit")

	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
