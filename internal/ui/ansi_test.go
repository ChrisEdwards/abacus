package ui

import (
	"strings"
	"testing"

	"abacus/internal/ui/theme"
)

func TestFillBackgroundPrefixesEveryLine(t *testing.T) {
	input := "top line\n\nsecond line"
	got := fillBackground(input)
	bgSeq := theme.Current().BackgroundANSI()
	if bgSeq == "" {
		t.Fatalf("theme did not provide background ANSI sequence")
	}

	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	for i, line := range lines {
		if !strings.HasPrefix(line, bgSeq) {
			t.Fatalf("line %d missing background prefix: %q", i, line)
		}
	}

	if lines[1] != bgSeq {
		t.Fatalf("blank line should be background-only, got %q", lines[1])
	}
}

func TestFillBackgroundReAppliesAfterResets(t *testing.T) {
	input := "foo\x1b[0mbar\nalpha\x1b[49mbeta\ngamma\x1b[momega\ndelta\x1b[39;49mtheta\nmargin\x1b[0K"
	got := fillBackground(input)
	bgSeq := theme.Current().BackgroundANSI()
	if bgSeq == "" {
		t.Fatalf("theme did not provide background ANSI sequence")
	}

	if !strings.Contains(got, "\x1b[0m"+bgSeq) {
		t.Fatalf("reset \x1b[0m not followed by background: %q", got)
	}
	if strings.Contains(got, "\x1b[49m") {
		t.Fatalf("background reset \x1b[49m should be replaced: %q", got)
	}
	if !strings.Contains(got, "\x1b[m"+bgSeq) {
		t.Fatalf("reset \x1b[m not followed by background: %q", got)
	}
	if strings.Contains(got, "\x1b[39;49m") {
		t.Fatalf("combined reset \x1b[39;49m should be rewritten: %q", got)
	}
	if !strings.Contains(got, "\x1b[39m"+bgSeq) {
		t.Fatalf("\x1b[39;49m should reapply background: %q", got)
	}
	if strings.Contains(got, "\x1b[0K") {
		t.Fatalf("line reset \x1b[0K should be replaced: %q", got)
	}
}

func TestApplyDimmerEnsuresDimSequenceEachLine(t *testing.T) {
	input := "line one\nline two\nline three"
	got := applyDimmer(input)
	const dimSeq = "\x1b[2m"

	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	for i, line := range lines {
		if !strings.HasPrefix(line, dimSeq) {
			t.Fatalf("line %d missing dim prefix: %q", i, line)
		}
	}
}

func TestApplyDimmerReappliesAfterResets(t *testing.T) {
	input := "foo\x1b[0mbar\nbaz\x1b[22mqux"
	got := applyDimmer(input)
	if !strings.Contains(got, "\x1b[0m\x1b[2m") {
		t.Fatalf("expected dimmer to reapply after \\x1b[0m: %q", got)
	}
	if !strings.Contains(got, "\x1b[22m\x1b[2m") {
		t.Fatalf("expected dimmer to reapply after \\x1b[22m: %q", got)
	}
}
