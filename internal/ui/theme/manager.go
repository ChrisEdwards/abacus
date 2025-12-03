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
