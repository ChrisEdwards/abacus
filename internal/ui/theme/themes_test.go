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

// TestCyclePreviousTheme verifies that backward theme cycling works correctly.
func TestCyclePreviousTheme(t *testing.T) {
	// Set to a known starting point
	SetTheme("dracula")

	// Cycle backwards through themes and ensure we get different names
	seen := make(map[string]bool)
	seen[CurrentName()] = true

	for i := 0; i < 30; i++ { // Cycle more than total themes to test wraparound
		name := CyclePreviousTheme()
		seen[name] = true
	}

	// We should have seen multiple themes
	if len(seen) < 23 {
		t.Errorf("expected to cycle through at least 23 themes, only saw %d", len(seen))
	}
}

// TestCycleThemeRoundTrip verifies that cycling forward then backward returns to the same theme.
func TestCycleThemeRoundTrip(t *testing.T) {
	// Set to a known starting point
	SetTheme("github")
	start := CurrentName()

	// Cycle forward 5 times
	for i := 0; i < 5; i++ {
		CycleTheme()
	}

	// Cycle backward 5 times
	for i := 0; i < 5; i++ {
		CyclePreviousTheme()
	}

	// Should be back at start
	if CurrentName() != start {
		t.Errorf("expected to return to %q after round trip, got %q", start, CurrentName())
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

// TestDimmedTheme verifies that Dimmed() returns a theme with blended colors.
func TestDimmedTheme(t *testing.T) {
	SetTheme("dracula")
	normal := Current()
	dimmed := normal.Dimmed()

	// Dimmed colors should be different from normal (blended toward background)
	if normal.Text().Dark == dimmed.Text().Dark {
		t.Error("Dimmed().Text() should be different from normal Text()")
	}

	if normal.Accent().Dark == dimmed.Accent().Dark {
		t.Error("Dimmed().Accent() should be different from normal Accent()")
	}

	// Background should remain the same
	if normal.Background().Dark != dimmed.Background().Dark {
		t.Error("Dimmed().Background() should be the same as normal Background()")
	}
}

// TestDimmedThemeColorsValid verifies that dimmed colors are valid hex codes.
func TestDimmedThemeColorsValid(t *testing.T) {
	for _, name := range Available() {
		SetTheme(name)
		dimmed := Current().Dimmed()

		checkValidHex := func(colorName string, color lipgloss.AdaptiveColor) {
			for _, hex := range []string{color.Dark, color.Light} {
				if hex == "" {
					continue
				}
				if len(hex) != 7 || hex[0] != '#' {
					t.Errorf("theme %q dimmed: %s has invalid hex %q", name, colorName, hex)
				}
			}
		}

		checkValidHex("Primary", dimmed.Primary())
		checkValidHex("Secondary", dimmed.Secondary())
		checkValidHex("Accent", dimmed.Accent())
		checkValidHex("Text", dimmed.Text())
		checkValidHex("TextMuted", dimmed.TextMuted())
	}
}

// TestBlendHex verifies the color blending function.
func TestBlendHex(t *testing.T) {
	tests := []struct {
		hex1, hex2 string
		factor     float64
		expected   string
	}{
		{"#ffffff", "#000000", 0.0, "#ffffff"}, // 0% blend = original
		{"#ffffff", "#000000", 1.0, "#000000"}, // 100% blend = target
		{"#ffffff", "#000000", 0.5, "#7f7f7f"}, // 50% blend = midpoint
		{"#ff0000", "#0000ff", 0.5, "#7f007f"}, // Red + Blue = Purple-ish
	}

	for _, tc := range tests {
		result := blendHex(tc.hex1, tc.hex2, tc.factor)
		if result != tc.expected {
			t.Errorf("blendHex(%q, %q, %.1f) = %q, expected %q",
				tc.hex1, tc.hex2, tc.factor, result, tc.expected)
		}
	}
}
