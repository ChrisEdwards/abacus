package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// TestAllThemesRegistered verifies that all expected themes are registered.
func TestAllThemesRegistered(t *testing.T) {
	expected := []string{
		"aura",
		"ayu",
		"catppuccin",
		"cobalt2",
		"dracula",
		"everforest",
		"flexoki",
		"github",
		"gruvbox",
		"kanagawa",
		"material",
		"matrix",
		"monokai",
		"nightowl",
		"nord",
		"onedark",
		"palenight",
		"rosepine",
		"solarized",
		"synthwave84",
		"tokyonight",
		"vesper",
		"zenburn",
	}

	available := Available()
	availableMap := make(map[string]bool)
	for _, name := range available {
		availableMap[name] = true
	}

	for _, name := range expected {
		if !availableMap[name] {
			t.Errorf("expected theme %q to be registered, but it was not found", name)
		}
	}
}

// TestThemeCount verifies that we have at least the expected number of themes.
func TestThemeCount(t *testing.T) {
	available := Available()
	// We should have at least 23 themes (6 original + 17 new)
	if len(available) < 23 {
		t.Errorf("expected at least 23 themes, got %d: %v", len(available), available)
	}
}

// TestSetTheme verifies that theme switching works.
func TestSetTheme(t *testing.T) {
	themes := []string{"dracula", "nord", "solarized", "aura", "matrix"}

	for _, name := range themes {
		if !SetTheme(name) {
			t.Errorf("SetTheme(%q) returned false, expected true", name)
			continue
		}
		if CurrentName() != name {
			t.Errorf("CurrentName() = %q, expected %q", CurrentName(), name)
		}
	}
}

// TestSetInvalidTheme verifies that setting an invalid theme returns false.
func TestSetInvalidTheme(t *testing.T) {
	if SetTheme("nonexistent-theme") {
		t.Error("SetTheme(\"nonexistent-theme\") returned true, expected false")
	}
}

// TestCycleTheme verifies that theme cycling works correctly.
func TestCycleTheme(t *testing.T) {
	// Set to a known starting point
	SetTheme("dracula")

	// Cycle through themes and ensure we get different names
	seen := make(map[string]bool)
	seen[CurrentName()] = true

	for i := 0; i < 30; i++ { // Cycle more than total themes to test wraparound
		name := CycleTheme()
		seen[name] = true
	}

	// We should have seen multiple themes
	if len(seen) < 23 {
		t.Errorf("expected to cycle through at least 23 themes, only saw %d", len(seen))
	}
}

// TestThemeColorsNotEmpty verifies that all theme methods return non-empty colors.
func TestThemeColorsNotEmpty(t *testing.T) {
	for _, name := range Available() {
		SetTheme(name)
		theme := Current()

		checkColor := func(colorName string, color lipgloss.AdaptiveColor) {
			if color.Dark == "" && color.Light == "" {
				t.Errorf("theme %q: %s has empty Dark and Light values", name, colorName)
			}
		}

		checkColor("Primary", theme.Primary())
		checkColor("Secondary", theme.Secondary())
		checkColor("Accent", theme.Accent())
		checkColor("Error", theme.Error())
		checkColor("Warning", theme.Warning())
		checkColor("Success", theme.Success())
		checkColor("Info", theme.Info())
		checkColor("Text", theme.Text())
		checkColor("TextMuted", theme.TextMuted())
		checkColor("TextEmphasized", theme.TextEmphasized())
		checkColor("Background", theme.Background())
		checkColor("BackgroundSecondary", theme.BackgroundSecondary())
		checkColor("BackgroundDarker", theme.BackgroundDarker())
		checkColor("BorderNormal", theme.BorderNormal())
		checkColor("BorderFocused", theme.BorderFocused())
		checkColor("BorderDim", theme.BorderDim())
	}
}

// TestThemeWrapper verifies that the ThemeWrapper provides utility methods.
func TestThemeWrapper(t *testing.T) {
	SetTheme("dracula")
	wrapper := Current()

	// Test BackgroundANSI
	ansi := wrapper.BackgroundANSI()
	if ansi == "" {
		t.Error("BackgroundANSI() returned empty string")
	}
	// Should be an ANSI escape sequence
	if len(ansi) < 10 || ansi[0] != '\x1b' {
		t.Errorf("BackgroundANSI() = %q, expected ANSI escape sequence", ansi)
	}

	// Test BackgroundSecondaryANSI
	ansi2 := wrapper.BackgroundSecondaryANSI()
	if ansi2 == "" {
		t.Error("BackgroundSecondaryANSI() returned empty string")
	}
}

// TestAvailableSorted verifies that Available returns sorted theme names.
func TestAvailableSorted(t *testing.T) {
	available := Available()

	for i := 1; i < len(available); i++ {
		if available[i-1] > available[i] {
			t.Errorf("Available() not sorted: %q > %q at index %d", available[i-1], available[i], i-1)
		}
	}
}
