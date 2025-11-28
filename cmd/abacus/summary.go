package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"abacus/internal/ui"

	"github.com/charmbracelet/lipgloss"
)

// ExitSummary holds data for the exit summary display shown when the TUI exits.
type ExitSummary struct {
	Version     string
	EndStats    ui.Stats
	SessionInfo ui.SessionInfo
}

// printExitSummary prints a formatted exit summary to the writer.
// This is displayed after the TUI exits alt screen mode.
func printExitSummary(w io.Writer, summary ExitSummary) {
	appStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor)

	versionStyle := lipgloss.NewStyle().
		Foreground(dimColor)

	statsStyle := lipgloss.NewStyle().
		Foreground(textColor)

	changePositiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50FA7B")) // Green for positive changes (closed)

	changeNeutralStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8BE9FD")) // Cyan for neutral changes

	// Calculate session duration
	duration := time.Since(summary.SessionInfo.StartTime)
	durationStr := formatDuration(duration)

	// Build version string with session duration
	versionStr := ""
	if summary.Version != "" {
		versionStr = versionStyle.Render(fmt.Sprintf(" v%s", summary.Version))
	}
	sessionStr := versionStyle.Render(fmt.Sprintf(" • %s session", durationStr))

	// Calculate changes
	startStats := summary.SessionInfo.InitialStats
	endStats := summary.EndStats
	totalDelta := endStats.Total - startStats.Total
	inProgressDelta := endStats.InProgress - startStats.InProgress
	readyDelta := endStats.Ready - startStats.Ready
	blockedDelta := endStats.Blocked - startStats.Blocked
	closedDelta := endStats.Closed - startStats.Closed

	// Build stats breakdown with inline changes
	var parts []string
	if endStats.InProgress > 0 || inProgressDelta != 0 {
		part := fmt.Sprintf("%d In Progress", endStats.InProgress)
		if inProgressDelta != 0 {
			part += " " + changeNeutralStyle.Render(formatDelta(inProgressDelta))
		}
		parts = append(parts, part)
	}
	if endStats.Ready > 0 || readyDelta != 0 {
		part := fmt.Sprintf("%d Ready", endStats.Ready)
		if readyDelta != 0 {
			part += " " + changeNeutralStyle.Render(formatDelta(readyDelta))
		}
		parts = append(parts, part)
	}
	if endStats.Blocked > 0 || blockedDelta != 0 {
		part := fmt.Sprintf("%d Blocked", endStats.Blocked)
		if blockedDelta != 0 {
			part += " " + changeNeutralStyle.Render(formatDelta(blockedDelta))
		}
		parts = append(parts, part)
	}
	if endStats.Closed > 0 || closedDelta != 0 {
		part := fmt.Sprintf("%d Closed", endStats.Closed)
		if closedDelta > 0 {
			// Closed increases are positive progress - use green
			part += " " + changePositiveStyle.Render(formatDelta(closedDelta))
		} else if closedDelta < 0 {
			part += " " + changeNeutralStyle.Render(formatDelta(closedDelta))
		}
		parts = append(parts, part)
	}

	statsStr := fmt.Sprintf("%d Beads", endStats.Total)
	if totalDelta != 0 {
		statsStr += " " + changeNeutralStyle.Render(formatDelta(totalDelta))
	}
	if len(parts) > 0 {
		statsStr += ": " + strings.Join(parts, ", ")
	}

	// Print exit summary
	_, _ = fmt.Fprintln(w, appStyle.Render("Abacus")+versionStr+sessionStr)
	_, _ = fmt.Fprintln(w, statsStyle.Render(statsStr))
}

// clearLoadingScreen clears the loading screen area before entering alt screen.
// This ensures a clean terminal state when the TUI exits.
func clearLoadingScreen(w io.Writer) {
	// The loading screen uses approximately 8 lines (logo + padding + progress)
	const loadingScreenLines = 8
	for i := 0; i < loadingScreenLines; i++ {
		_, _ = fmt.Fprint(w, "\033[A") // Move cursor up one line
	}
	_, _ = fmt.Fprint(w, "\033[J") // Clear from cursor to end of screen
}

// formatDuration formats a duration into a human-readable string.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		secs := int(d.Seconds()) % 60
		if secs == 0 {
			return fmt.Sprintf("%dm", mins)
		}
		return fmt.Sprintf("%dm %ds", mins, secs)
	}
	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	if mins == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh %dm", hours, mins)
}

// formatDelta formats a numeric delta with +/- prefix.
func formatDelta(delta int) string {
	if delta > 0 {
		return fmt.Sprintf("(+%d)", delta)
	}
	return fmt.Sprintf("(%d)", delta)
}
