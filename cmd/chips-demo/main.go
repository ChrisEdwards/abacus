// Demo program to visually test the ChipList component
package main

import (
	"fmt"
	"os"
	"strings"

	"abacus/internal/ui"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	chips ui.ChipList
	input textinput.Model
	log   []string
	width int
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type label and press Enter..."
	ti.Focus()
	ti.Width = 30

	return model{
		chips: ui.NewChipList().WithWidth(50),
		input: ti,
		log:   []string{},
		width: 50,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle global keys first
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if !m.chips.InNavigationMode() && m.input.Value() == "" {
				return m, tea.Quit
			}
		case "+", "=":
			m.width += 10
			m.chips = m.chips.WithWidth(m.width)
			m.addLog(fmt.Sprintf("Width: %d", m.width))
			return m, nil
		case "-", "_":
			if m.width > 20 {
				m.width -= 10
				m.chips = m.chips.WithWidth(m.width)
				m.addLog(fmt.Sprintf("Width: %d", m.width))
			}
			return m, nil
		}

		// Route based on mode
		if m.chips.InNavigationMode() {
			var cmd tea.Cmd
			m.chips, cmd = m.chips.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		// Input mode handling
		switch msg.Type {
		case tea.KeyEnter:
			if m.input.Value() != "" {
				label := strings.TrimSpace(m.input.Value())
				if m.chips.AddChip(label) {
					m.addLog(fmt.Sprintf("Added: [%s]", label))
				} else {
					m.addLog(fmt.Sprintf("Duplicate: [%s] (flash!)", label))
					// Trigger flash clear after delay
					cmds = append(cmds, ui.FlashCmd())
				}
				m.input.SetValue("")
			}
			return m, tea.Batch(cmds...)

		case tea.KeyLeft:
			if m.input.Value() == "" && len(m.chips.Chips) > 0 {
				m.chips.EnterNavigation()
				m.addLog("Entered chip navigation (last chip)")
				return m, nil
			}
		}

	case ui.ChipNavExitMsg:
		m.chips.ExitNavigation()
		switch msg.Reason {
		case ui.ChipNavExitRight:
			m.addLog("Exit: → past last chip")
		case ui.ChipNavExitEscape:
			m.addLog("Exit: Esc pressed")
		case ui.ChipNavExitTab:
			m.addLog("Exit: Tab pressed")
		case ui.ChipNavExitTyping:
			m.addLog(fmt.Sprintf("Exit: typed '%c'", msg.Character))
			// Forward the character to input
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{msg.Character}})
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case ui.ChipRemovedMsg:
		m.addLog(fmt.Sprintf("Removed: [%s] at index %d", msg.Label, msg.Index))
		return m, nil
	}

	// Update chips (for flash clear, etc.)
	var chipCmd tea.Cmd
	m.chips, chipCmd = m.chips.Update(msg)
	if chipCmd != nil {
		cmds = append(cmds, chipCmd)
	}

	// Update text input
	if !m.chips.InNavigationMode() {
		var inputCmd tea.Cmd
		m.input, inputCmd = m.input.Update(msg)
		if inputCmd != nil {
			cmds = append(cmds, inputCmd)
		}
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
)

func (m model) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("ChipList Demo"))
	s.WriteString("\n\n")

	// Mode indicator
	mode := "INPUT"
	if m.chips.InNavigationMode() {
		mode = "CHIP NAV"
	}
	s.WriteString(fmt.Sprintf("Mode: %s  Width: %d\n\n", modeStyle.Render(mode), m.width))

	// Labels section
	s.WriteString(labelStyle.Render("LABELS"))
	s.WriteString("\n")

	// Chips display
	chipsView := m.chips.View()
	if chipsView == "" {
		chipsView = "(no chips)"
	}

	if !m.chips.InNavigationMode() {
		// Show input cursor after chips
		if chipsView != "(no chips)" {
			chipsView += " "
		} else {
			chipsView = ""
		}
		chipsView += m.input.View()
	}

	s.WriteString(boxStyle.Width(m.width + 4).Render(chipsView))
	s.WriteString("\n")

	// Help
	if m.chips.InNavigationMode() {
		s.WriteString(helpStyle.Render("←/→ navigate • Backspace delete • letter/Esc/Tab exit"))
	} else {
		s.WriteString(helpStyle.Render("Type + Enter add • ← chip nav • +/- width • q quit"))
	}
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
