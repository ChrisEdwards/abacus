package ui

import (
	"fmt"
	"time"
)

var timeNow = time.Now

// FormatRelativeTime returns a compact, human-friendly description of how long
// ago t occurred using the Stripe-inspired rules described in ab-xcyg. Results
// never exceed ~8 characters so they fit inside tight tree columns.
func FormatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	now := timeNow()

	// Defensive: future timestamps fall back to absolute dates.
	if t.After(now) {
		return formatAbsoluteTime(t, now)
	}

	diff := now.Sub(t)
	switch {
	case diff < time.Minute:
		return "now"
	case diff < time.Hour:
		minutes := int(diff / time.Minute)
		if minutes < 1 {
			minutes = 1
		}
		return fmt.Sprintf("%dm ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff / time.Hour)
		return fmt.Sprintf("%dh ago", hours)
	case diff < 100*24*time.Hour:
		days := int(diff / (24 * time.Hour))
		return fmt.Sprintf("%dd ago", days)
	default:
		return formatAbsoluteTime(t, now)
	}
}

func formatAbsoluteTime(t, now time.Time) string {
	local := t.In(now.Location())
	if local.Year() == now.Year() {
		return local.Format("Jan 2")
	}
	return local.Format("Jan '06")
}
