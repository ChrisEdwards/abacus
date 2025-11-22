package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderRefRow_HangingIndentForWrappedTitles(t *testing.T) {
	id := "ab-ltb"
	title := "Research Beads repository structure and CI/CD setup"
	targetWidth := 34 // force wrapping with a modest viewport width

	row := renderRefRow(
		id,
		title,
		targetWidth,
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.Color(""),
	)

	lines := strings.Split(row, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected wrapped output, got %q", row)
	}

	prefix := id + "  "
	if !strings.HasPrefix(lines[0], prefix) {
		t.Fatalf("expected id prefix on first line: %q", lines[0])
	}

	expectedIndent := strings.Repeat(" ", len(prefix))
	if !strings.HasPrefix(lines[1], expectedIndent) {
		t.Fatalf("wrapped line should start with indent %q:\n%s", expectedIndent, row)
	}

	if strings.Contains(lines[1], id) {
		t.Fatalf("ID should not appear on wrapped lines: %q", lines[1])
	}

	if got := strings.TrimSpace(lines[1]); !strings.HasPrefix(got, "structure") {
		t.Fatalf("wrapped line missing continuation text: %q", lines[1])
	}
}

func TestRenderRefRow_DetailHeaderDoesNotWrapID(t *testing.T) {
	id := "ab-ltb"
	title := "Research Beads repository structure and CI/CD setup"

	for _, width := range []int{12, 16, 20, 24, 34, 50} {
		headerContent := renderRefRow(
			id,
			title,
			width,
			styleDetailHeaderCombined.Foreground(cGold),
			styleDetailHeaderCombined.Foreground(cWhite),
			cHighlight,
		)
		headerBlock := styleDetailHeaderBlock.Width(width).Render(headerContent)

		lines := strings.Split(headerBlock, "\n")
		if len(lines) < 2 {
			continue
		}

		if strings.HasPrefix(strings.TrimSpace(lines[1]), "ltb") {
			t.Fatalf("width %d caused ID to wrap into next line:\n%s", width, headerBlock)
		}
	}
}
