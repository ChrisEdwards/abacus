package ui

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"abacus/internal/ui/theme"
)

var updateGolden = flag.Bool("update-golden", false, "update UI golden snapshot files")

var goldenThemes = []string{"dracula", "solarized", "nord"}

func TestOverlayAndToastGoldenSnapshots(t *testing.T) {
	for _, name := range goldenThemes {
		name := name
		t.Run(name, func(t *testing.T) {
			if !theme.SetTheme(name) {
				t.Fatalf("theme %q not registered", name)
			}

			overlay := NewStatusOverlay("ab-smg0", "Golden Snapshot Surface Render", "open")
			layer := overlay.Layer(80, 24, 1, 1)
			if layer == nil {
				t.Fatalf("expected status overlay layer for theme %s", name)
			}
			canvas := layer.Render()
			if canvas == nil {
				t.Fatalf("expected overlay canvas for theme %s", name)
			}
			assertGoldenSnapshot(t, fmt.Sprintf("%s_status_overlay.golden", name), canvas.Render())

			app := &App{
				statusToastVisible:   true,
				statusToastNewStatus: "in_progress",
				statusToastBeadID:    "ab-smg0",
				statusToastStart:     time.Now(),
			}
			toast := app.statusToastLayer(80, 24, 2, 12)
			if toast == nil {
				t.Fatalf("expected status toast layer for theme %s", name)
			}
			canvas = toast.Render()
			if canvas == nil {
				t.Fatalf("expected toast canvas for theme %s", name)
			}
			assertGoldenSnapshot(t, fmt.Sprintf("%s_status_toast.golden", name), canvas.Render())
		})
	}
}

func assertGoldenSnapshot(t *testing.T, filename, got string) {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "ui", "golden", filename)
	if *updateGolden {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("writing golden %s: %v", path, err)
		}
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading golden %s: %v (use -update-golden to refresh)", path, err)
	}
	if string(want) != got {
		t.Fatalf("golden mismatch for %s\nwant:\n%s\n\ngot:\n%s\n", filename, want, got)
	}
}
