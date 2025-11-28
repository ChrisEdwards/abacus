package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// helpSection represents a group of keybindings for display.
type helpSection struct {
	title string
	rows  [][]string // Each row: [keys, description]
}

// getHelpSections returns the help content organized into sections.
// Layout is explicit - each section lists which bindings appear in which order.
// Text is derived from binding.Help() to maintain single source of truth.
func getHelpSections(keys KeyMap) []helpSection {
	return []helpSection{
		{
			title: "NAVIGATION",
			rows: [][]string{
				{keys.Up.Help().Key, keys.Up.Help().Desc},
				{keys.Left.Help().Key, keys.Left.Help().Desc},
				{keys.Space.Help().Key, keys.Space.Help().Desc},
				{keys.Home.Help().Key, keys.Home.Help().Desc},
				{keys.End.Help().Key, keys.End.Help().Desc},
				{keys.PageUp.Help().Key, keys.PageUp.Help().Desc},
				{keys.PageDown.Help().Key, keys.PageDown.Help().Desc},
			},
		},
		{
			title: "ACTIONS",
			rows: [][]string{
				{keys.Enter.Help().Key, keys.Enter.Help().Desc},
				{keys.Tab.Help().Key, keys.Tab.Help().Desc},
				{keys.Refresh.Help().Key, keys.Refresh.Help().Desc},
				{keys.Copy.Help().Key, keys.Copy.Help().Desc},
				{keys.Error.Help().Key, keys.Error.Help().Desc},
			},
		},
		{
			title: "SEARCH",
			rows: [][]string{
				{keys.Search.Help().Key, keys.Search.Help().Desc},
				{keys.Enter.Help().Key, "Confirm"},
				{keys.Escape.Help().Key, keys.Escape.Help().Desc},
			},
		},
	}
}

// renderHelpOverlay creates the centered help modal.
func renderHelpOverlay(keys KeyMap, width, height int) string {
	sections := getHelpSections(keys)

	// Build left column (Navigation)
	leftCol := renderHelpSectionTable(sections[0])

	// Build right column (Actions + Search)
	rightCol := lipgloss.JoinVertical(lipgloss.Left,
		renderHelpSectionTable(sections[1]),
		"",
		renderHelpSectionTable(sections[2]),
	)

	// Join columns horizontally with spacing
	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, "    ", rightCol)

	// Build complete overlay content
	title := styleHelpTitle.Render("✦ ABACUS HELP ✦")
	dividerWidth := lipgloss.Width(columns)
	if dividerWidth < 40 {
		dividerWidth = 40
	}
	divider := styleHelpDivider.Render(strings.Repeat("─", dividerWidth))
	footer := styleHelpFooter.Render("Press ? or Esc to close")

	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		divider,
		"",
		columns,
		"",
		footer,
	)

	// Apply overlay styling with border
	styled := styleHelpOverlay.Render(content)

	// Center on screen using lipgloss.Place()
	return lipgloss.Place(width, height,
		lipgloss.Center, lipgloss.Center,
		styled,
		lipgloss.WithWhitespaceChars(" "),
	)
}

// renderHelpSectionTable renders a single help section using lipgloss/table.
func renderHelpSectionTable(section helpSection) string {
	// Create table with hidden borders for clean look
	t := table.New().
		Border(lipgloss.HiddenBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			if col == 0 {
				return styleHelpKey.Width(14)
			}
			return styleHelpDesc
		}).
		Rows(section.rows...)

	// Build section with header and underline
	header := styleHelpSectionHeader.Render(section.title)
	underline := styleHelpUnderline.Render(strings.Repeat("─", len(section.title)))

	// Trim leading newline from table output (hidden border adds empty top row)
	tableStr := strings.TrimPrefix(t.String(), "\n")

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		underline,
		tableStr,
	)
}
