package ui

import (
	"testing"
	"time"
)

func TestFormatRelativeTime(t *testing.T) {
	fixedNow := time.Date(2025, time.December, 25, 12, 0, 0, 0, time.UTC)
	orig := timeNow
	timeNow = func() time.Time { return fixedNow }
	defer func() { timeNow = orig }()

	tests := []struct {
		name string
		ts   time.Time
		want string
	}{
		{name: "zero time", ts: time.Time{}, want: ""},
		{name: "future timestamp", ts: fixedNow.Add(2 * time.Hour), want: "Dec 25"},
		{name: "seconds ago", ts: fixedNow.Add(-30 * time.Second), want: "now"},
		{name: "just over minute", ts: fixedNow.Add(-61 * time.Second), want: "1m ago"},
		{name: "fifty nine minutes", ts: fixedNow.Add(-59 * time.Minute), want: "59m ago"},
		{name: "hours", ts: fixedNow.Add(-23 * time.Hour), want: "23h ago"},
		{name: "over one hour", ts: fixedNow.Add(-61 * time.Minute), want: "1h ago"},
		{name: "days", ts: fixedNow.Add(-48 * time.Hour), want: "2d ago"},
		{name: "six days", ts: fixedNow.Add(-6 * 24 * time.Hour), want: "6d ago"},
		{name: "seven days relative", ts: fixedNow.Add(-7 * 24 * time.Hour), want: "7d ago"},
		{name: "fifteen days relative", ts: fixedNow.Add(-15 * 24 * time.Hour), want: "15d ago"},
		{name: "thirty days relative", ts: fixedNow.Add(-30 * 24 * time.Hour), want: "30d ago"},
		{name: "ninety nine days relative", ts: fixedNow.Add(-99 * 24 * time.Hour), want: "99d ago"},
		{name: "hundred days absolute", ts: fixedNow.Add(-100 * 24 * time.Hour), want: "Sep 16"},
		{
			name: "previous year absolute",
			ts:   time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
			want: "Jan '24",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatRelativeTime(tt.ts); got != tt.want {
				t.Fatalf("FormatRelativeTime(%v) = %q, want %q", tt.ts, got, tt.want)
			}
		})
	}
}
