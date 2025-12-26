package ui

import (
	"strings"
	"testing"
	"time"

	"abacus/internal/update"
)

func TestRenderThemeToastSpacerKeepsBackground(t *testing.T) {
	app := &App{
		themeToastVisible: true,
		themeToastName:    "dracula",
		themeToastStart:   time.Now(),
	}

	layer := app.themeToastLayer(80, 24, 2, 10)
	if layer == nil {
		t.Fatal("expected theme toast to render")
	}

	canvas := layer.Render()
	if canvas == nil {
		t.Fatal("expected canvas from theme toast layer")
	}
	out := canvas.Render()
	if strings.Contains(out, "Theme:\x1b[0m ") {
		t.Fatalf("found raw space with default background: %q", out)
	}
}

func TestHeaderVersionGapUsesBackground(t *testing.T) {
	app := &App{
		ready:    true,
		width:    80,
		height:   20,
		version:  "dev",
		repoName: "abacus",
	}

	view := app.View()
	titleSegment := styleAppHeader().Render("ABACUS vdev")
	gap := baseStyle().Render(" ")
	if !strings.Contains(view, titleSegment+gap) {
		t.Fatalf("expected themed gap after header title, got: %q", view)
	}
}

func TestViewOmitsDefaultResetGaps(t *testing.T) {
	app := &App{
		ready:                true,
		width:                100,
		height:               30,
		repoName:             "abacus",
		activeOverlay:        OverlayStatus,
		statusOverlay:        NewStatusOverlay("ab-smg0", "Snapshot", "in_progress"),
		statusToastVisible:   true,
		statusToastStart:     time.Now(),
		statusToastBeadID:    "ab-smg0",
		statusToastNewStatus: "in_progress",
	}

	view := app.View()
	if strings.Contains(view, "\x1b[0m ") {
		t.Fatalf("view contains default reset gap: %q", view)
	}
}

func TestUpdateToastRenders(t *testing.T) {
	app := &App{
		updateToastVisible: true,
		updateToastStart:   time.Now(),
		updateInfo: &update.UpdateInfo{
			UpdateAvailable: true,
			LatestVersion:   update.Version{Major: 1, Minor: 2, Patch: 3},
			InstallMethod:   update.InstallDirect,
			UpdateCommand:   "download from github",
		},
	}

	layer := app.updateToastLayer(80, 24, 2, 10)
	if layer == nil {
		t.Fatal("expected update toast to render when visible")
	}

	canvas := layer.Render()
	if canvas == nil {
		t.Fatal("expected canvas from update toast layer")
	}
	out := canvas.Render()
	if !strings.Contains(out, "Update available") {
		t.Errorf("expected toast to contain 'Update available', got: %q", out)
	}
	if !strings.Contains(out, "1.2.3") {
		t.Errorf("expected toast to contain version '1.2.3', got: %q", out)
	}
}

func TestUpdateToastNotRenderedWhenInvisible(t *testing.T) {
	app := &App{
		updateToastVisible: false,
		updateInfo: &update.UpdateInfo{
			UpdateAvailable: true,
			LatestVersion:   update.Version{Major: 1, Minor: 0, Patch: 0},
		},
	}

	layer := app.updateToastLayer(80, 24, 2, 10)
	if layer != nil {
		t.Error("expected update toast not to render when invisible")
	}
}

func TestUpdateToastNotRenderedWithoutUpdateInfo(t *testing.T) {
	app := &App{
		updateToastVisible: true,
		updateToastStart:   time.Now(),
		updateInfo:         nil,
	}

	layer := app.updateToastLayer(80, 24, 2, 10)
	if layer != nil {
		t.Error("expected update toast not to render without updateInfo")
	}
}

func TestUpdateToastShowsHomebrewCommand(t *testing.T) {
	app := &App{
		updateToastVisible: true,
		updateToastStart:   time.Now(),
		updateInfo: &update.UpdateInfo{
			UpdateAvailable: true,
			LatestVersion:   update.Version{Major: 2, Minor: 0, Patch: 0},
			InstallMethod:   update.InstallHomebrew,
			UpdateCommand:   "brew upgrade bv",
		},
	}

	layer := app.updateToastLayer(80, 24, 2, 10)
	if layer == nil {
		t.Fatal("expected update toast to render")
	}

	canvas := layer.Render()
	out := canvas.Render()
	if !strings.Contains(out, "brew upgrade") {
		t.Errorf("expected homebrew toast to contain 'brew upgrade', got: %q", out)
	}
}

func TestUpdateToastShowsHotkeyForDirect(t *testing.T) {
	app := &App{
		updateToastVisible: true,
		updateToastStart:   time.Now(),
		updateInfo: &update.UpdateInfo{
			UpdateAvailable: true,
			LatestVersion:   update.Version{Major: 2, Minor: 0, Patch: 0},
			InstallMethod:   update.InstallDirect,
		},
	}

	layer := app.updateToastLayer(80, 24, 2, 10)
	if layer == nil {
		t.Fatal("expected update toast to render")
	}

	canvas := layer.Render()
	out := canvas.Render()
	if !strings.Contains(out, "[U]") {
		t.Errorf("expected direct install toast to contain '[U]' hotkey, got: %q", out)
	}
}
