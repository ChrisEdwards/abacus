package theme

import "sync"

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

// Current returns the active theme.
func Current() Theme {
	globalManager.mu.RLock()
	defer globalManager.mu.RUnlock()
	return globalManager.currentTheme
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

	// Cycle to next
	nextIdx := (currentIdx + 1) % len(names)
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
