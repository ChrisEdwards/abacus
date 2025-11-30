// Demo program to visually test the ChipComboBox component
package main

import (
	"fmt"
	"os"
	"strings"

	"abacus/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var sampleOptions = []string{
	"backend",
	"frontend",
	"api",
	"urgent",
	"bug",
	"feature",
	"security",
	"performance",
	"documentation",
	"testing",
}

type model struct {
	chipCombo ui.ChipComboBox
	log       []string
	width     int
}

func initialModel() model {
	cc := ui.NewChipComboBox(sampleOptions).
		WithWidth(50).
		WithPlaceholder("Type to filter or ↓ to browse...").
		WithAllowNew(true, "New label: %s")
	cc.Focus()

	return model{
		chipCombo: cc,
		log:       []string{},
		width:     50,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if !m.chipCombo.InChipNavMode() && !m.chipCombo.IsDropdownOpen() && m.chipCombo.InputValue() == "" {
				return m, tea.Quit
			}
		case "=", "+":
			m.width += 10
			m.chipCombo = m.chipCombo.WithWidth(m.width)
			m.addLog(fmt.Sprintf("Width: %d", m.width))
			return m, nil
		case "-", "_":
			if m.width > 30 {
				m.width -= 10
				m.chipCombo = m.chipCombo.WithWidth(m.width)
				m.addLog(fmt.Sprintf("Width: %d", m.width))
			}
			return m, nil
		}

	case ui.ChipComboBoxChipAddedMsg:
		if msg.IsNew {
			m.addLog(fmt.Sprintf("NEW: [%s] added", msg.Label))
		} else {
			m.addLog(fmt.Sprintf("Added: [%s]", msg.Label))
		}
		return m, nil

	case ui.ChipComboBoxTabMsg:
		m.addLog("Tab pressed - would move to next field")
		return m, nil

	case ui.ChipRemovedMsg:
		m.addLog(fmt.Sprintf("Removed: [%s]", msg.Label))
		return m, nil

	case ui.ChipNavExitMsg:
		switch msg.Reason {
		case ui.ChipNavExitRight:
			m.addLog("Chip nav: exited right")
		case ui.ChipNavExitEscape:
			m.addLog("Chip nav: exited via Esc")
		case ui.ChipNavExitTyping:
			m.addLog(fmt.Sprintf("Chip nav: exited typing '%c'", msg.Character))
		}
	}

	// Update the chip combo
	var cmd tea.Cmd
	m.chipCombo, cmd = m.chipCombo.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) addLog(entry string) {
	m.log = append(m.log, entry)
	if len(m.log) > 8 {
		m.log = m.log[1:]
	}
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	logStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			MarginTop(1)

	modeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1)

	chipsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39"))
)

func (m model) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("ChipComboBox Demo"))
	s.WriteString("\n\n")

	// Mode indicator
	mode := "INPUT"
	if m.chipCombo.InChipNavMode() {
		mode = "CHIP NAV"
	} else if m.chipCombo.IsDropdownOpen() {
		mode = "DROPDOWN"
	}
	s.WriteString(fmt.Sprintf("Mode: %s  Width: %d  Chips: %d\n\n",
		modeStyle.Render(mode), m.width, m.chipCombo.ChipCount()))

	// Labels section
	s.WriteString(labelStyle.Render("LABELS"))
	s.WriteString("\n")

	// ChipComboBox view
	comboView := m.chipCombo.View()
	s.WriteString(boxStyle.Width(m.width + 4).Render(comboView))
	s.WriteString("\n")

	// Show selected chips summary
	chips := m.chipCombo.GetChips()
	if len(chips) > 0 {
		s.WriteString(chipsStyle.Render(fmt.Sprintf("Selected: %s", strings.Join(chips, ", "))))
		s.WriteString("\n")
	}

	// Help
	if m.chipCombo.InChipNavMode() {
		s.WriteString(helpStyle.Render("←/→ navigate • Backspace delete • letter/Esc exit"))
	} else if m.chipCombo.IsDropdownOpen() {
		s.WriteString(helpStyle.Render("↑/↓ navigate • Enter select • Esc close • type to filter"))
	} else {
		s.WriteString(helpStyle.Render("↓ open dropdown • type to filter • ← chip nav • +/- width • q quit"))
	}
	s.WriteString("\n")

	// Available options
	s.WriteString("\n")
	s.WriteString(labelStyle.Render("Available: " + strings.Join(sampleOptions, ", ")))
	s.WriteString("\n")

	// Log
	if len(m.log) > 0 {
		s.WriteString(logStyle.Render("\nLog:"))
		s.WriteString("\n")
		for _, entry := range m.log {
			s.WriteString(logStyle.Render("  " + entry))
			s.WriteString("\n")
		}
	}

	return s.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
