package ui

import (
	"strings"
	"testing"

	"abacus/internal/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

func TestNewCenteredOverlayLayer(t *testing.T) {
	content := "OVER\nOK"
	layer := newCenteredOverlayLayer(content, 20, 10, 1, 1)
	if layer == nil {
		t.Fatal("expected layer to be created")
	}
	canvas := layer.Render()
	if canvas == nil {
		t.Fatal("expected canvas from layer")
	}
	assertCanvasBackground(t, canvas, theme.Current().BackgroundSecondary())
	lines := strings.Split(stripANSI(canvas.Render()), "\n")
	if len(lines) == 0 || !strings.Contains(strings.Join(lines, "\n"), "OVER") {
		t.Fatalf("expected overlay content in canvas:\n%s", strings.Join(lines, "\n"))
	}
	expectedWidth := lipgloss.Width("OVER")
	if canvas.Width() != expectedWidth {
		t.Fatalf("expected overlay width %d, got %d", expectedWidth, canvas.Width())
	}
	expectedHeight := lipgloss.Height(content)
	if canvas.Height() != expectedHeight {
		t.Fatalf("expected overlay height %d, got %d", expectedHeight, canvas.Height())
	}
	ox, oy := canvas.Offset()
	if ox <= 0 || oy <= 0 {
		t.Fatalf("expected overlay offsets to be positive, got (%d,%d)", ox, oy)
	}
}

func TestNewToastLayerPositionsBottomRight(t *testing.T) {
	const width, height = 30, 10
	mainBodyStart := 2
	mainBodyHeight := 6
	content := "toast\nok"
	layer := newToastLayer(content, width, height, mainBodyStart, mainBodyHeight)
	if layer == nil {
		t.Fatal("expected toast layer")
	}
	canvas := layer.Render()
	if canvas == nil {
		t.Fatal("expected toast canvas")
	}
	assertCanvasBackground(t, canvas, theme.Current().Background())
	if !strings.Contains(canvas.Render(), "toast") {
		t.Fatalf("expected toast content, got:\n%s", canvas.Render())
	}
	expectedWidth := lipgloss.Width("toast")
	if canvas.Width() != expectedWidth {
		t.Fatalf("expected toast width %d, got %d", expectedWidth, canvas.Width())
	}
	expectedHeight := lipgloss.Height(content)
	if canvas.Height() != expectedHeight {
		t.Fatalf("expected toast height %d, got %d", expectedHeight, canvas.Height())
	}
	ox, oy := canvas.Offset()
	if oy < mainBodyStart {
		t.Fatalf("expected toast offset >= %d, got %d", mainBodyStart, oy)
	}
	if ox < width-lipgloss.Width("toast")-4 {
		t.Fatalf("expected toast near right edge, got offset %d", ox)
	}
}

func assertCanvasBackground(t *testing.T, canvas *Canvas, target lipgloss.TerminalColor) {
	t.Helper()
	if canvas == nil {
		t.Fatal("canvas is nil")
	}
	tr, tg, tb, _ := target.RGBA()
	for y := 0; y < canvas.Height(); y++ {
		for x := 0; x < canvas.Width(); x++ {
			cell := canvas.Cell(x, y)
			if cell == nil {
				t.Fatalf("missing cell at %d,%d", x, y)
			}
			bg := cell.Style.Bg
			if bg == nil {
				t.Fatalf("cell (%d,%d) missing background color", x, y)
			}
			br, bgc, bb, _ := bg.RGBA()
			if br != tr || bgc != tg || bb != tb {
				t.Fatalf("cell (%d,%d) background mismatch got RGB(%d,%d,%d)", x, y, br, bgc, bb)
			}
		}
	}
}
