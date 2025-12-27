package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
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
				{keys.CycleViewMode.Help().Key, keys.CycleViewMode.Help().Desc},
				{keys.Refresh.Help().Key, keys.Refresh.Help().Desc},
				{keys.Error.Help().Key, keys.Error.Help().Desc},
				{keys.Theme.Help().Key, keys.Theme.Help().Desc},
				{keys.Update.Help().Key, keys.Update.Help().Desc},
			},
		},
		{
			title: "BEAD ACTIONS",
			rows: [][]string{
				{keys.Copy.Help().Key, keys.Copy.Help().Desc},
				{keys.Status.Help().Key, keys.Status.Help().Desc},
				{keys.Labels.Help().Key, keys.Labels.Help().Desc},
				{keys.NewBead.Help().Key, keys.NewBead.Help().Desc},
				{keys.NewRootBead.Help().Key, keys.NewRootBead.Help().Desc},
				{keys.Edit.Help().Key, keys.Edit.Help().Desc},
				{keys.Comment.Help().Key, keys.Comment.Help().Desc},
				{keys.Delete.Help().Key, keys.Delete.Help().Desc},
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

// renderHelpOverlay builds the help modal content.
func renderHelpOverlay(keys KeyMap) string {
	sections := getHelpSections(keys)

	// Build left column (Navigation + Actions)
	leftCol := lipgloss.JoinVertical(lipgloss.Left,
		renderHelpSectionTable(sections[0]),
		"",
		renderHelpSectionTable(sections[1]),
	)

	// Build right column (Bead Actions + Search)
	rightCol := lipgloss.JoinVertical(lipgloss.Left,
		renderHelpSectionTable(sections[2]),
		"",
		renderHelpSectionTable(sections[3]),
	)

	// Join columns horizontally with spacing
	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, "    ", rightCol)

	// Build complete overlay content
	title := styleHelpTitle().Render("✦ ABACUS HELP ✦")
	dividerWidth := lipgloss.Width(columns)
	if dividerWidth < 40 {
		dividerWidth = 40
	}
	divider := styleHelpDivider().Render(strings.Repeat("─", dividerWidth))
	footer := styleHelpFooter().Render("Press ? or Esc to close")

	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		divider,
		"",
		columns,
		"",
		footer,
	)

	// Apply overlay styling with border
	return styleHelpOverlay().Render(content)
}

func newHelpOverlayLayer(keys KeyMap, width, height, topMargin, bottomMargin int) Layer {
	// Return a LayerFunc that renders lazily, so the theme is correct when Render() is called
	return LayerFunc(func() *Canvas {
		content := renderHelpOverlay(keys)
		if content == "" {
			return nil
		}
		overlayWidth, overlayHeight := blockDimensions(content)
		if overlayWidth <= 0 || overlayHeight <= 0 {
			return nil
		}

		surface := NewSecondarySurface(overlayWidth, overlayHeight)
		surface.Draw(0, 0, content)
		x, y := centeredOffsets(width, height, overlayWidth, overlayHeight, topMargin, bottomMargin)

		surface.Canvas.SetOffset(x, y)
		return surface.Canvas
	})
}

// renderHelpSectionTable renders a single help section with styled rows.
func renderHelpSectionTable(section helpSection) string {
	// Build section header and underline
	header := styleHelpSectionHeader().Render(section.title)
	underline := styleHelpUnderline().Render(strings.Repeat("─", len(section.title)))

	// Build rows manually to avoid table border spacing issues
	var rowStrings []string
	for _, row := range section.rows {
		key := styleHelpKey().Width(14).Render(row[0])
		desc := styleHelpDesc().Render(row[1])
		rowStrings = append(rowStrings, key+desc)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		underline,
		strings.Join(rowStrings, "\n"),
	)
}
