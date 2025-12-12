package theme

import (
	"fmt"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

var globalManager = &manager{
	themes: make(map[string]Theme),
}

type manager struct {
	mu           sync.RWMutex
	themes       map[string]Theme
	currentName  string
	currentTheme Theme
}

// RegisterTheme adds a theme to the registry.
// The first registered theme becomes the default.
func RegisterTheme(name string, t Theme) {
	globalManager.mu.Lock()
	defer globalManager.mu.Unlock()
	globalManager.themes[name] = t
	if globalManager.currentTheme == nil {
		globalManager.currentName = name
		globalManager.currentTheme = t
	}
}

// SetTheme switches to a registered theme by name.
// Returns true if the theme was found and set.
func SetTheme(name string) bool {
	globalManager.mu.Lock()
	defer globalManager.mu.Unlock()
	if t, ok := globalManager.themes[name]; ok {
		globalManager.currentName = name
		globalManager.currentTheme = t
		return true
	}
	return false
}

// CurrentName returns the name of the active theme.
func CurrentName() string {
	globalManager.mu.RLock()
	defer globalManager.mu.RUnlock()
	return globalManager.currentName
}

// Available returns a list of all registered theme names in sorted order.
func Available() []string {
	globalManager.mu.RLock()
	defer globalManager.mu.RUnlock()
	names := make([]string, 0, len(globalManager.themes))
	for name := range globalManager.themes {
		names = append(names, name)
	}
	// Sort for consistent ordering
	sortStrings(names)
	return names
}

// CycleTheme switches to the next theme in the sorted list.
// Returns the name of the new active theme.
func CycleTheme() string {
	return cycleThemeByOffset(1)
}

// CyclePreviousTheme switches to the previous theme in the sorted list.
// Returns the name of the new active theme.
func CyclePreviousTheme() string {
	return cycleThemeByOffset(-1)
}

// cycleThemeByOffset cycles the theme by the given offset (positive = forward, negative = backward).
func cycleThemeByOffset(offset int) string {
	globalManager.mu.Lock()
	defer globalManager.mu.Unlock()

	names := make([]string, 0, len(globalManager.themes))
	for name := range globalManager.themes {
		names = append(names, name)
	}
	sortStrings(names)

	if len(names) == 0 {
		return ""
	}

	// Find current index
	currentIdx := 0
	for i, name := range names {
		if name == globalManager.currentName {
			currentIdx = i
			break
		}
	}

	// Cycle by offset (handle negative wraparound)
	nextIdx := (currentIdx + offset + len(names)) % len(names)
	nextName := names[nextIdx]
	globalManager.currentName = nextName
	globalManager.currentTheme = globalManager.themes[nextName]

	return nextName
}

// sortStrings sorts a slice of strings in place.
func sortStrings(s []string) {
	for i := 0; i < len(s)-1; i++ {
		for j := i + 1; j < len(s); j++ {
			if s[i] > s[j] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

// ThemeWrapper wraps a Theme to provide additional utility methods.
type ThemeWrapper struct {
	Theme
}

// BackgroundANSI returns the ANSI escape sequence for the theme's background color.
// This is used for post-processing rendered content to fill background gaps.
func (w ThemeWrapper) BackgroundANSI() string {
	return colorToANSIBackground(w.Background())
}

// BackgroundSecondaryANSI returns the ANSI escape sequence for the theme's secondary background color.
func (w ThemeWrapper) BackgroundSecondaryANSI() string {
	return colorToANSIBackground(w.BackgroundSecondary())
}

// hexToRGB converts a hex color string to RGB values.
func hexToRGB(hex string) (r, g, b uint8) {
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return 0, 0, 0
	}
	var ri, gi, bi int
	if n, err := fmt.Sscanf(hex, "%02x%02x%02x", &ri, &gi, &bi); err != nil || n != 3 {
		return 0, 0, 0
	}
	return clampToUint8(ri), clampToUint8(gi), clampToUint8(bi)
}

func clampToUint8(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func colorToANSIBackground(color lipgloss.AdaptiveColor) string {
	hex := color.Dark
	if hex == "" {
		hex = color.Light
	}
	r, g, b := hexToRGB(hex)
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b)
}

// Current returns the active theme wrapped with utility methods.
func Current() ThemeWrapper {
	globalManager.mu.RLock()
	defer globalManager.mu.RUnlock()
	return ThemeWrapper{globalManager.currentTheme}
}

// Dimmed returns a new ThemeWrapper with all colors blended toward the background.
// This is used when an overlay is active to visually de-emphasize the background content.
func (w ThemeWrapper) Dimmed() ThemeWrapper {
	return ThemeWrapper{&dimmedTheme{base: w.Theme, blendFactor: 0.45}}
}

// dimmedTheme wraps a base theme and blends all colors toward the background.
type dimmedTheme struct {
	base        Theme
	blendFactor float64 // 0.0 = original color, 1.0 = fully background
}

// blendColor blends a color toward the background by the given factor.
func blendColor(color, background lipgloss.AdaptiveColor, factor float64) lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{
		Light: blendHex(color.Light, background.Light, factor),
		Dark:  blendHex(color.Dark, background.Dark, factor),
	}
}

// blendHex blends two hex colors. Factor 0.0 = color1, 1.0 = color2.
func blendHex(hex1, hex2 string, factor float64) string {
	r1, g1, b1 := hexToRGB(hex1)
	r2, g2, b2 := hexToRGB(hex2)

	r := uint8(float64(r1)*(1-factor) + float64(r2)*factor)
	g := uint8(float64(g1)*(1-factor) + float64(g2)*factor)
	b := uint8(float64(b1)*(1-factor) + float64(b2)*factor)

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// Theme interface implementation for dimmedTheme

func (d *dimmedTheme) Primary() lipgloss.AdaptiveColor {
	return blendColor(d.base.Primary(), d.base.Background(), d.blendFactor)
}

func (d *dimmedTheme) Secondary() lipgloss.AdaptiveColor {
	return blendColor(d.base.Secondary(), d.base.Background(), d.blendFactor)
}

func (d *dimmedTheme) Accent() lipgloss.AdaptiveColor {
	return blendColor(d.base.Accent(), d.base.Background(), d.blendFactor)
}

func (d *dimmedTheme) Error() lipgloss.AdaptiveColor {
	return blendColor(d.base.Error(), d.base.Background(), d.blendFactor)
}

func (d *dimmedTheme) Warning() lipgloss.AdaptiveColor {
	return blendColor(d.base.Warning(), d.base.Background(), d.blendFactor)
}

func (d *dimmedTheme) Success() lipgloss.AdaptiveColor {
	return blendColor(d.base.Success(), d.base.Background(), d.blendFactor)
}

func (d *dimmedTheme) Info() lipgloss.AdaptiveColor {
	return blendColor(d.base.Info(), d.base.Background(), d.blendFactor)
}

func (d *dimmedTheme) Text() lipgloss.AdaptiveColor {
	return blendColor(d.base.Text(), d.base.Background(), d.blendFactor)
}

func (d *dimmedTheme) TextMuted() lipgloss.AdaptiveColor {
	return blendColor(d.base.TextMuted(), d.base.Background(), d.blendFactor)
}

func (d *dimmedTheme) TextEmphasized() lipgloss.AdaptiveColor {
	return blendColor(d.base.TextEmphasized(), d.base.Background(), d.blendFactor)
}

func (d *dimmedTheme) Background() lipgloss.AdaptiveColor {
	// Background stays the same - we blend other colors toward it
	return d.base.Background()
}

func (d *dimmedTheme) BackgroundSecondary() lipgloss.AdaptiveColor {
	// Secondary background also stays the same
	return d.base.BackgroundSecondary()
}

func (d *dimmedTheme) BackgroundDarker() lipgloss.AdaptiveColor {
	return d.base.BackgroundDarker()
}

func (d *dimmedTheme) BorderNormal() lipgloss.AdaptiveColor {
	return blendColor(d.base.BorderNormal(), d.base.Background(), d.blendFactor)
}

func (d *dimmedTheme) BorderFocused() lipgloss.AdaptiveColor {
	return blendColor(d.base.BorderFocused(), d.base.Background(), d.blendFactor)
}

func (d *dimmedTheme) BorderDim() lipgloss.AdaptiveColor {
	return blendColor(d.base.BorderDim(), d.base.Background(), d.blendFactor)
}
