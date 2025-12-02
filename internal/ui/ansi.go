package ui

import (
	"regexp"
	"strings"

	"abacus/internal/ui/theme"
)

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiRegexp.ReplaceAllString(s, "")
}

// fillBackground replaces ANSI reset codes with sequences that preserve the theme background.
// This ensures all whitespace between styled segments has the correct background color.
func fillBackground(s string) string {
	// Get the background color escape sequence from the theme
	bgSeq := theme.Current().BackgroundANSI()

	// Replace full reset (\x1b[0m) with reset + background
	// This preserves the background after any style reset
	s = strings.ReplaceAll(s, "\x1b[0m", "\x1b[0m"+bgSeq)

	// Replace background-only reset (\x1b[49m) with the theme background
	s = strings.ReplaceAll(s, "\x1b[49m", bgSeq)

	return s
}
