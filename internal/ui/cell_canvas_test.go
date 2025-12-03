package ui

import (
	"strings"
	"testing"
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

func TestCanvasCenterOverlayPositionsContent(t *testing.T) {
	const width, height = 20, 10
	canvas := NewCanvas(width, height)
	canvas.DrawStringAt(0, 0, baseStyle().Width(width).Height(height).Render(""))

	canvas.centerOverlay("AA\nBB", 1, 1)
	lines := strings.Split(canvas.Render(), "\n")

	expectedRow := 4 // computed from topMargin=1, bottomMargin=1, overlay height=2
	if len(lines) <= expectedRow+1 {
		t.Fatalf("not enough lines rendered, got %d", len(lines))
	}
	first := stripANSI(lines[expectedRow])
	if idx := strings.Index(first, "AA"); idx != 9 {
		t.Fatalf("expected overlay 'AA' centered at column 9, got column %d", idx)
	}

	second := stripANSI(lines[expectedRow+1])
	if idx := strings.Index(second, "BB"); idx != 9 {
		t.Fatalf("expected overlay 'BB' centered at column 9, got column %d", idx)
	}
}

func TestCanvasBottomRightOverlayAnchorsToast(t *testing.T) {
	const width, height = 30, 6
	canvas := NewCanvas(width, height)
	canvas.DrawStringAt(0, 0, baseStyle().Width(width).Height(height).Render(""))

	canvas.bottomRightOverlay("ERR", 1)
	lines := strings.Split(canvas.Render(), "\n")
	targetRow := height - 1 - 1 // padding of 1
	line := stripANSI(lines[targetRow])

	idx := strings.Index(line, "ERR")
	if idx == -1 {
		t.Fatalf("expected toast text in row %d, got %q", targetRow, line)
	}
	if idx < width-len("ERR")-2 {
		t.Fatalf("expected toast near right edge, got column %d", idx)
	}
}
