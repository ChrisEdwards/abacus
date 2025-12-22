package ui

import (
	"strings"
	"time"

	"abacus/internal/domain"
	"abacus/internal/graph"

	"github.com/charmbracelet/lipgloss"
)

func commonLeadingSpaces(lines []string) int {
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(stripANSI(line)) == "" {
			continue
		}
		count := countLeadingSpacesANSI(line)
		if minIndent == -1 || count < minIndent {
			minIndent = count
		}
	}
	if minIndent < 0 {
		return 0
	}
	return minIndent
}

func countLeadingSpacesANSI(line string) int {
	count := 0
	i := 0
	for i < len(line) {
		if line[i] == '\x1b' {
			end := i + 1
			for end < len(line) && line[end-1] != 'm' {
				end++
			}
			if end >= len(line) {
				break
			}
			i = end
			continue
		}
		if line[i] == ' ' {
			count++
			i++
			continue
		}
		break
	}
	return count
}

func trimANSIIndent(line string, spaces int) string {
	if spaces <= 0 {
		return line
	}
	var b strings.Builder
	remaining := spaces
	i := 0
	for i < len(line) {
		if line[i] == '\x1b' {
			end := i + 1
			for end < len(line) && line[end-1] != 'm' {
				end++
			}
			if end > len(line) {
				end = len(line)
			}
			b.WriteString(line[i:end])
			i = end
			continue
		}
		if line[i] == ' ' && remaining > 0 {
			remaining--
			i++
			continue
		}
		break
	}
	b.WriteString(line[i:])
	return b.String()
}

func isVisualBlankLine(line string) bool {
	return strings.TrimSpace(stripANSI(line)) == ""
}

func renderRefRow(id, title string, targetWidth int, idStyle, titleStyle lipgloss.Style, bgColor lipgloss.TerminalColor) string {
	const gap = "  "

	idStyled := idStyle.Background(bgColor)
	gapStyled := lipgloss.NewStyle().Background(bgColor)
	titleStyled := titleStyle.Background(bgColor)

	idRendered := idStyled.Render(id)
	idWidth := lipgloss.Width(idRendered)
	gapRendered := gapStyled.Render(gap)
	gapWidth := lipgloss.Width(gapRendered)
	prefixWidth := idWidth + gapWidth

	titleWidth := targetWidth - prefixWidth
	if titleWidth < 1 {
		titleWidth = 1
	}

	titleLines := wrapTitleWithoutHyphenBreaks(title, titleWidth)
	if len(titleLines) == 0 {
		titleLines = []string{""}
	}

	idBlank := idStyled.Width(idWidth).Render("")
	gapBlank := gapStyled.Width(gapWidth).Render("")

	lines := make([]string, len(titleLines))
	for i, line := range titleLines {
		idSegment := idBlank
		gapSegment := gapBlank
		if i == 0 {
			idSegment = idStyled.Width(idWidth).Render(id)
			gapSegment = gapStyled.Width(gapWidth).Render(gap)
		}
		titleSegment := titleStyled.Width(titleWidth).Render(line)
		lines[i] = lipgloss.JoinHorizontal(lipgloss.Left, idSegment, gapSegment, titleSegment)
	}

	return strings.Join(lines, "\n")
}

func renderRefRowWithIcon(icon string, iconStyle lipgloss.Style, id, title string, targetWidth int, idStyle, titleStyle lipgloss.Style) string {
	const gap = "  "
	bgStyle := baseStyle()
	bg := currentThemeWrapper().Background()
	// Ensure icon and id styles have background
	iconRendered := iconStyle.Background(bg).Render(icon)
	sp := bgStyle.Render(" ")
	idRendered := idStyle.Background(bg).Render(id)
	gapRendered := bgStyle.Render(gap)

	// Calculate widths for alignment
	iconWidth := lipgloss.Width(iconRendered)
	idWidth := lipgloss.Width(idRendered)
	gapWidth := lipgloss.Width(gapRendered)
	prefixWidth := iconWidth + 1 + idWidth + gapWidth
	titleWidth := targetWidth - prefixWidth
	if titleWidth < 1 {
		titleWidth = 1
	}
	titleLines := wrapTitleWithoutHyphenBreaks(title, titleWidth)
	if len(titleLines) == 0 {
		titleLines = []string{""}
	}

	// Build prefix by direct concatenation (avoids JoinHorizontal resets)
	prefixFirst := iconRendered + sp + idRendered + gapRendered
	prefixBlank := bgStyle.Render(strings.Repeat(" ", lipgloss.Width(prefixFirst)))

	// Ensure title style has background
	titleStyleWithBg := titleStyle.Background(bg)

	lines := make([]string, len(titleLines))
	for i, line := range titleLines {
		prefix := prefixBlank
		if i == 0 {
			prefix = prefixFirst
		}
		// Direct concatenation instead of JoinHorizontal
		lines[i] = prefix + titleStyleWithBg.Render(line)
	}
	return strings.Join(lines, "\n")
}

func relatedStatusPresentation(node *graph.Node) (string, lipgloss.Style, lipgloss.Style) {
	domainIssue, err := domain.NewIssueFromFull(node.Issue, node.IsBlocked)
	status := node.Issue.Status
	if err == nil {
		status = string(domainIssue.Status())
	}
	switch status {
	case "in_progress":
		return "◐", styleIconInProgress(), styleInProgressText()
	case "closed":
		return "✔", styleIconDone(), styleDoneText()
	default:
		if node.IsBlocked {
			return "⛔", styleIconBlocked(), styleBlockedText()
		}
		return "○", styleIconOpen(), styleNormalText()
	}
}

// wrapTitleWithoutHyphenBreaks wraps the title text while treating the ID as a
// fixed-width prefix so the ID never wraps and continuation lines align with
// the start of the title. The wrapping only happens at whitespace boundaries
// to avoid splitting hyphenated words like "real-time".
func wrapTitleWithoutHyphenBreaks(title string, width int) []string {
	if width < 1 {
		width = 1
	}
	words := strings.Fields(title)
	if len(words) == 0 {
		return []string{""}
	}

	lines := make([]string, 0, len(words))
	var current strings.Builder
	currentWidth := 0

	flush := func() {
		if current.Len() == 0 {
			return
		}
		lines = append(lines, current.String())
		current.Reset()
		currentWidth = 0
	}

	for _, word := range words {
		wordWidth := lipgloss.Width(word)
		if currentWidth == 0 {
			current.WriteString(word)
			currentWidth = wordWidth
			continue
		}

		nextWidth := currentWidth + 1 + wordWidth
		if nextWidth > width {
			flush()
			current.WriteString(word)
			currentWidth = wordWidth
			continue
		}

		current.WriteByte(' ')
		current.WriteString(word)
		currentWidth = nextWidth
	}

	flush()
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func indentBlock(text string, spaces int) string {
	padding := baseStyle().Render(strings.Repeat(" ", spaces))
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = padding + line
		}
	}
	return strings.Join(lines, "\n")
}

func formatTime(isoStr string) string {
	if isoStr == "" {
		return "-"
	}
	t, err := time.Parse(time.RFC3339, isoStr)
	if err != nil {
		return isoStr
	}
	return t.Local().Format("Jan 02, 3:04 PM")
}
