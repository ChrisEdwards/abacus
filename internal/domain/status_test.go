package domain

import "testing"

func TestStatusValidate(t *testing.T) {
	valid := []Status{StatusOpen, StatusInProgress, StatusBlocked, StatusDeferred, StatusClosed}
	for _, status := range valid {
		if err := status.Validate(); err != nil {
			t.Errorf("expected %q to be valid, got error: %v", status, err)
		}
	}

	invalid := []Status{StatusUnknown, Status("invalid"), StatusTombstone}
	for _, status := range invalid {
		if err := status.Validate(); err == nil {
			t.Errorf("expected %q to be invalid", status)
		}
	}
}

func TestParseStatus(t *testing.T) {
	cases := map[string]Status{
		"open":        StatusOpen,
		"in_progress": StatusInProgress,
		" blocked ":   StatusBlocked,
		"Deferred":    StatusDeferred,
		"CLOSED":      StatusClosed,
	}

	for raw, expected := range cases {
		got, err := ParseStatus(raw)
		if err != nil {
			t.Fatalf("ParseStatus(%q) returned error: %v", raw, err)
		}
		if got != expected {
			t.Fatalf("ParseStatus(%q) = %q, want %q", raw, got, expected)
		}
	}

	// Only blank and tombstone should return errors
	for _, raw := range []string{"", "tombstone"} {
		if _, err := ParseStatus(raw); err == nil {
			t.Fatalf("expected ParseStatus(%q) to return error", raw)
		}
	}

	// Unknown statuses should be accepted (forward compatibility)
	// Use IsKnown() to distinguish known from unknown
	unknownStatus, err := ParseStatus("pinned")
	if err != nil {
		t.Fatalf("ParseStatus(\"pinned\") should accept unknown status, got error: %v", err)
	}
	if unknownStatus != Status("pinned") {
		t.Fatalf("ParseStatus(\"pinned\") = %q, want \"pinned\"", unknownStatus)
	}
}

func TestStatusIsKnown(t *testing.T) {
	known := []Status{StatusOpen, StatusInProgress, StatusBlocked, StatusDeferred, StatusClosed}
	for _, status := range known {
		if !status.IsKnown() {
			t.Errorf("expected %q.IsKnown() to be true", status)
		}
	}

	unknown := []Status{StatusUnknown, Status("pinned"), Status("weird"), StatusTombstone}
	for _, status := range unknown {
		if status.IsKnown() {
			t.Errorf("expected %q.IsKnown() to be false", status)
		}
	}
}

func TestAllowedTransitions(t *testing.T) {
	allowed := []struct {
		from Status
		to   Status
	}{
		{StatusOpen, StatusInProgress},
		{StatusOpen, StatusBlocked},
		{StatusOpen, StatusDeferred},
		{StatusOpen, StatusClosed},
		{StatusInProgress, StatusOpen},
		{StatusInProgress, StatusBlocked},
		{StatusInProgress, StatusDeferred},
		{StatusInProgress, StatusClosed},
		{StatusBlocked, StatusOpen},
		{StatusBlocked, StatusInProgress},
		{StatusBlocked, StatusDeferred},
		{StatusBlocked, StatusClosed},
		{StatusDeferred, StatusOpen},
		{StatusDeferred, StatusInProgress},
		{StatusDeferred, StatusBlocked},
		{StatusDeferred, StatusClosed},
		{StatusClosed, StatusOpen},
	}

	for _, tc := range allowed {
		if err := tc.from.CanTransitionTo(tc.to); err != nil {
			t.Fatalf("expected transition from %q to %q to be allowed: %v", tc.from, tc.to, err)
		}
	}

	disallowed := []struct {
		from Status
		to   Status
	}{
		{StatusClosed, StatusInProgress},
		{StatusClosed, StatusBlocked},
		{StatusClosed, StatusDeferred},
		{StatusBlocked, StatusTombstone},
		{StatusDeferred, Status("invalid")},
	}

	for _, tc := range disallowed {
		if err := tc.from.CanTransitionTo(tc.to); err == nil {
			t.Fatalf("expected transition from %q to %q to be rejected", tc.from, tc.to)
		}
	}
}
