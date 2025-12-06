package ui

import "strings"

// extractShortError extracts a short, user-friendly error message.
func extractShortError(fullError string, maxLen int) string {
	msg := fullError

	// Look for "Error:" pattern and extract from there
	if idx := strings.Index(msg, "Error:"); idx >= 0 {
		msg = strings.TrimSpace(msg[idx+6:]) // Skip "Error:"
	} else if idx := strings.Index(msg, "error:"); idx >= 0 {
		msg = strings.TrimSpace(msg[idx+6:])
	}

	// Take only the first line/sentence
	if idx := strings.Index(msg, "\n"); idx >= 0 {
		msg = msg[:idx]
	}
	// Also truncate at period if it makes sense
	if idx := strings.Index(msg, ". "); idx >= 0 && idx < maxLen {
		msg = msg[:idx]
	}

	// Remove any "Run 'bd..." suggestions
	if idx := strings.Index(msg, " Run '"); idx >= 0 {
		msg = msg[:idx]
	}
	if idx := strings.Index(msg, " Run \""); idx >= 0 {
		msg = msg[:idx]
	}

	msg = strings.TrimSpace(msg)

	// Truncate if still too long
	if len(msg) > maxLen {
		msg = msg[:maxLen-3] + "..."
	}

	return msg
}
