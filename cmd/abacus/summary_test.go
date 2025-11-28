package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"abacus/internal/ui"
)

func TestPrintExitSummary(t *testing.T) {
	tests := []struct {
		name     string
		summary  ExitSummary
		wantApp  string
		wantVer  string
		wantStat string
	}{
		{
			name: "full stats with version",
			summary: ExitSummary{
				Version: "0.4.0",
				EndStats: ui.Stats{
					Total:      12,
					InProgress: 2,
					Ready:      5,
					Blocked:    1,
					Closed:     4,
				},
				SessionInfo: ui.SessionInfo{
					StartTime: time.Now().Add(-5 * time.Minute),
					InitialStats: ui.Stats{
						Total:      12,
						InProgress: 2,
						Ready:      5,
						Blocked:    1,
						Closed:     4,
					},
				},
			},
			wantApp:  "Abacus",
			wantVer:  "v0.4.0",
			wantStat: "12 Beads: 2 In Progress, 5 Ready, 1 Blocked, 4 Closed",
		},
		{
			name: "no version",
			summary: ExitSummary{
				Version: "",
				EndStats: ui.Stats{
					Total: 5,
					Ready: 5,
				},
				SessionInfo: ui.SessionInfo{
					StartTime:    time.Now().Add(-1 * time.Minute),
					InitialStats: ui.Stats{Total: 5, Ready: 5},
				},
			},
			wantApp:  "Abacus",
			wantVer:  "",
			wantStat: "5 Beads: 5 Ready",
		},
		{
			name: "zero beads",
			summary: ExitSummary{
				Version:  "1.0.0",
				EndStats: ui.Stats{},
				SessionInfo: ui.SessionInfo{
					StartTime:    time.Now().Add(-30 * time.Second),
					InitialStats: ui.Stats{},
				},
			},
			wantApp:  "Abacus",
			wantVer:  "v1.0.0",
			wantStat: "0 Beads",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printExitSummary(&buf, tt.summary)
			output := buf.String()

			if !strings.Contains(output, tt.wantApp) {
				t.Errorf("output missing app name %q:\n%s", tt.wantApp, output)
			}

			if tt.wantVer != "" && !strings.Contains(output, tt.wantVer) {
				t.Errorf("output missing version %q:\n%s", tt.wantVer, output)
			}

			if tt.wantVer == "" && strings.Contains(output, " v0") {
				t.Errorf("output should not contain version marker:\n%s", output)
			}

			if !strings.Contains(output, tt.wantStat) {
				t.Errorf("output missing stats %q:\n%s", tt.wantStat, output)
			}

			// Verify session duration is shown
			if !strings.Contains(output, "session") {
				t.Errorf("output should contain session duration:\n%s", output)
			}
		})
	}
}

func TestClearLoadingScreen(t *testing.T) {
	var buf bytes.Buffer
	clearLoadingScreen(&buf)
	output := buf.String()

	// Verify ANSI escape codes for clearing loading screen
	if !strings.Contains(output, "\033[A") {
		t.Errorf("output should contain cursor-up ANSI codes:\n%q", output)
	}
	if !strings.Contains(output, "\033[J") {
		t.Errorf("output should contain clear-to-end ANSI code:\n%q", output)
	}
}

func TestPrintExitSummary_ShowsChanges(t *testing.T) {
	summary := ExitSummary{
		Version: "1.0.0",
		EndStats: ui.Stats{
			Total:      12,
			InProgress: 2,
			Ready:      4,
			Blocked:    1,
			Closed:     5,
		},
		SessionInfo: ui.SessionInfo{
			StartTime: time.Now().Add(-10 * time.Minute),
			InitialStats: ui.Stats{
				Total:      10,
				InProgress: 1,
				Ready:      5,
				Blocked:    1,
				Closed:     3,
			},
		},
	}

	var buf bytes.Buffer
	printExitSummary(&buf, summary)
	output := buf.String()

	// Should show +2 for total
	if !strings.Contains(output, "(+2)") {
		t.Errorf("output should show +2 total change:\n%s", output)
	}

	// Should show +1 for in progress
	if !strings.Contains(output, "(+1)") {
		t.Errorf("output should show +1 in progress change:\n%s", output)
	}

	// Should show -1 for ready
	if !strings.Contains(output, "(-1)") {
		t.Errorf("output should show -1 ready change:\n%s", output)
	}
}

func TestPrintExitSummary_NoChangesShown(t *testing.T) {
	// When stats don't change, no deltas should be shown
	stats := ui.Stats{
		Total:      10,
		InProgress: 1,
		Ready:      5,
		Blocked:    1,
		Closed:     3,
	}
	summary := ExitSummary{
		Version:  "1.0.0",
		EndStats: stats,
		SessionInfo: ui.SessionInfo{
			StartTime:    time.Now().Add(-5 * time.Minute),
			InitialStats: stats, // Same as end stats
		},
	}

	var buf bytes.Buffer
	printExitSummary(&buf, summary)
	output := buf.String()

	// Should NOT show any delta markers
	if strings.Contains(output, "(+") || strings.Contains(output, "(-") {
		t.Errorf("output should not show deltas when nothing changed:\n%s", output)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		want     string
	}{
		{30 * time.Second, "30s"},
		{59 * time.Second, "59s"},
		{60 * time.Second, "1m"},
		{90 * time.Second, "1m 30s"},
		{5 * time.Minute, "5m"},
		{5*time.Minute + 30*time.Second, "5m 30s"},
		{60 * time.Minute, "1h"},
		{90 * time.Minute, "1h 30m"},
		{2 * time.Hour, "2h"},
		{2*time.Hour + 15*time.Minute, "2h 15m"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestFormatDelta(t *testing.T) {
	tests := []struct {
		delta int
		want  string
	}{
		{1, "(+1)"},
		{5, "(+5)"},
		{-1, "(-1)"},
		{-5, "(-5)"},
		{0, "(0)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatDelta(tt.delta)
			if got != tt.want {
				t.Errorf("formatDelta(%d) = %q, want %q", tt.delta, got, tt.want)
			}
		})
	}
}
