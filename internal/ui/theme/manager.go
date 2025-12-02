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

// Available returns a list of all registered theme names.
func Available() []string {
	globalManager.mu.RLock()
	defer globalManager.mu.RUnlock()
	names := make([]string, 0, len(globalManager.themes))
	for name := range globalManager.themes {
		names = append(names, name)
	}
	return names
}
