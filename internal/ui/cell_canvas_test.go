package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/cellbuf"
)

func TestCanvasNormalizesNewlines(t *testing.T) {
	canvas := NewCanvas(8, 4)
	canvas.DrawStringAt(0, 0, "A\nB")

	output := canvas.Render()
	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got %d", len(lines))
	}
	if got := strings.TrimSpace(stripANSI(lines[0])); got != "A" {
		t.Fatalf("line 0 mismatch, expected A got %q", got)
	}
	if got := strings.TrimSpace(stripANSI(lines[1])); got != "B" {
		t.Fatalf("line 1 mismatch, expected B got %q", got)
	}
}

func TestCanvasOverlayAt(t *testing.T) {
	base := NewCanvas(10, 5)
	base.DrawStringAt(0, 0, baseStyle().Width(10).Height(5).Render(""))

	top := NewCanvas(4, 2)
	top.DrawStringAt(0, 0, "AB\nCD")

	base.OverlayAt(3, 2, top)
	lines := strings.Split(stripANSI(base.Render()), "\n")
	if !strings.Contains(lines[2], "AB") {
		t.Fatalf("expected overlay row to include AB, got %q", lines[2])
	}
	if !strings.Contains(lines[3], "CD") {
		t.Fatalf("expected overlay row to include CD, got %q", lines[3])
	}
}

func TestCanvasApplyDimmer(t *testing.T) {
	canvas := NewCanvas(4, 1)
	canvas.DrawStringAt(0, 0, "Text")
	canvas.ApplyDimmer()

	for x := 0; x < 4; x++ {
		cell := canvas.screen.Cell(x, 0)
		if cell == nil || cell.Empty() {
			t.Fatalf("expected cell at %d to exist", x)
		}
		if cell.Style.Attrs&cellbuf.FaintAttr == 0 {
			t.Fatalf("expected cell at %d to be faint", x)
		}
	}
}
