package ui

import (
	"fmt"
	"strings"
	"time"

	"abacus/internal/config"
	"abacus/internal/graph"

	"github.com/charmbracelet/lipgloss"
)

const columnSeparator = " │ "

var columnSeparatorWidth = lipgloss.Width(columnSeparator)

type treeColumn struct {
	ConfigKey string
	Width     int
	Render    func(*graph.Node) string
}

var defaultTreeColumns = []treeColumn{
	{
		ConfigKey: config.KeyTreeColumnsLastUpdated,
		Width:     8,
		Render:    renderLastUpdatedColumn,
	},
	{
		ConfigKey: config.KeyTreeColumnsComments,
		Width:     5,
		Render:    renderCommentsColumn,
	},
}

type columnState struct {
	columns    []treeColumn
	totalWidth int
}

func (c columnState) enabled() bool {
	return len(c.columns) > 0
}

func (c columnState) render(node *graph.Node, mode columnRenderMode) string {
	if !c.enabled() {
		return ""
	}
	return c.renderWithProvider(mode, func(col treeColumn) string {
		return col.Render(node)
	})
}

func (c columnState) renderWithProvider(mode columnRenderMode, valueProvider func(treeColumn) string) string {
	if !c.enabled() {
		return ""
	}

	sepStyle, valueStyle := columnStyles(mode)
	var builder strings.Builder
	builder.WriteString(sepStyle.Render(columnSeparator))

	for _, col := range c.columns {
		cellValue := valueProvider(col)
		cell := valueStyle.
			Width(col.Width).
			Align(lipgloss.Right).
			Render(cellValue)
		builder.WriteString(cell)
	}
	return builder.String()
}

type columnRenderMode int

const (
	columnRenderNormal columnRenderMode = iota
	columnRenderSelected
	columnRenderCrossHighlight
)

func columnStyles(mode columnRenderMode) (lipgloss.Style, lipgloss.Style) {
	t := currentThemeWrapper()
	switch mode {
	case columnRenderSelected:
		base := lipgloss.NewStyle().Background(t.BackgroundSecondary())
		sep := base.Foreground(t.BorderNormal())
		val := base.Foreground(t.Text())
		return sep, val
	case columnRenderCrossHighlight:
		base := lipgloss.NewStyle().Background(t.BorderNormal())
		sep := base.Foreground(t.TextMuted())
		val := base.Foreground(t.Text())
		return sep, val
	default:
		return styleColumnSeparator(), styleColumnText()
	}
}

func prepareColumnState(totalWidth int) (columnState, int) {
	if !config.GetBool(config.KeyTreeShowColumns) {
		return columnState{}, totalWidth
	}

	// Gather all enabled columns
	enabledCols := make([]treeColumn, 0, len(defaultTreeColumns))
	for _, col := range defaultTreeColumns {
		if config.GetBool(col.ConfigKey) {
			enabledCols = append(enabledCols, col)
		}
	}
	if len(enabledCols) == 0 {
		return columnState{}, totalWidth
	}

	// Progressive hiding: remove columns from right to left until they fit
	// Columns are ordered left-to-right by priority (leftmost = highest priority)
	// so we remove from the end (rightmost = lowest priority = hides first)
	for len(enabledCols) > 0 {
		width := columnSeparatorWidth
		for _, col := range enabledCols {
			width += col.Width
		}

		treeWidth := totalWidth - width
		if treeWidth >= minTreeWidth {
			// Columns fit while respecting minimum tree width
			return columnState{
				columns:    enabledCols,
				totalWidth: width,
			}, treeWidth
		}

		// Remove rightmost column (lowest priority) and try again
		enabledCols = enabledCols[:len(enabledCols)-1]
	}

	// No columns fit - return empty state
	return columnState{}, totalWidth
}

func renderLastUpdatedColumn(node *graph.Node) string {
	if node == nil || node.Issue.UpdatedAt == "" {
		return ""
	}
	ts, err := time.Parse(time.RFC3339, node.Issue.UpdatedAt)
	if err != nil {
		return ""
	}
	return FormatRelativeTime(ts)
}

func renderCommentsColumn(node *graph.Node) string {
	if node == nil || !node.CommentsLoaded {
		return ""
	}
	count := len(node.Issue.Comments)
	if count <= 0 {
		return ""
	}
	switch {
	case count > 99:
		return "💬99+"
	case count > 9:
		return fmt.Sprintf("💬%d", count)
	default:
		return fmt.Sprintf("💬 %d", count)
	}
}
