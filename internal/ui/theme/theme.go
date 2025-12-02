// Package theme provides a semantic color system for the Abacus UI.
package theme

import "github.com/charmbracelet/lipgloss"

// Theme defines the 16 semantic colors for Abacus UI.
// All methods return AdaptiveColor for automatic light/dark terminal support.
type Theme interface {
	// Base colors
	Primary() lipgloss.AdaptiveColor   // Main accent (focused borders, header bg)
	Secondary() lipgloss.AdaptiveColor // Secondary accent (field labels, links)
	Accent() lipgloss.AdaptiveColor    // Highlights (IDs, titles)

	// Status colors
	Error() lipgloss.AdaptiveColor   // Errors, blocked, destructive
	Warning() lipgloss.AdaptiveColor // Warnings, priority badges
	Success() lipgloss.AdaptiveColor // Success, checked, in-progress
	Info() lipgloss.AdaptiveColor    // Informational highlights

	// Text colors
	Text() lipgloss.AdaptiveColor          // Primary text
	TextMuted() lipgloss.AdaptiveColor     // De-emphasized text
	TextEmphasized() lipgloss.AdaptiveColor // Bold/important text

	// Background colors
	Background() lipgloss.AdaptiveColor          // Main background
	BackgroundSecondary() lipgloss.AdaptiveColor // Selected rows, elevated surfaces
	BackgroundDarker() lipgloss.AdaptiveColor    // Chips, badges

	// Border colors
	BorderNormal() lipgloss.AdaptiveColor  // Default borders
	BorderFocused() lipgloss.AdaptiveColor // Active/focused borders
	BorderDim() lipgloss.AdaptiveColor     // Subtle borders
}
