package ui

import (
	"testing"

	"abacus/internal/beads"
	"abacus/internal/graph"
)

func TestParseSearchInputSimpleTokens(t *testing.T) {
	result := parseSearchInput("status:open priority:1")
	if len(result.tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(result.tokens))
	}
	if result.tokens[0].Key != "status" || result.tokens[0].Value != "open" {
		t.Fatalf("unexpected first token: %+v", result.tokens[0])
	}
	if result.tokens[1].Key != "priority" || result.tokens[1].Value != "1" {
		t.Fatalf("unexpected second token: %+v", result.tokens[1])
	}
	if result.mode != SuggestionModeField {
		t.Fatalf("expected mode field after complete tokens, got %v", result.mode)
	}
}

func TestParseSearchInputQuotedValue(t *testing.T) {
	result := parseSearchInput(`status:"in progress"`)
	if len(result.tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(result.tokens))
	}
	if result.tokens[0].Value != "in progress" {
		t.Fatalf("expected quoted value preserved, got %q", result.tokens[0].Value)
	}
}

func TestParseSearchInputIncompleteValue(t *testing.T) {
	result := parseSearchInput(`status:"in progress`)
	if len(result.tokens) != 0 {
		t.Fatalf("expected zero finalized tokens, got %d", len(result.tokens))
	}
	if result.mode != SuggestionModeValue {
		t.Fatalf("expected value suggestion mode, got %v", result.mode)
	}
	if result.pendingField != "status" {
		t.Fatalf("expected pending field 'status', got %q", result.pendingField)
	}
	if result.pendingText != `"in progress` {
		t.Fatalf("unexpected pending text: %q", result.pendingText)
	}
}

func TestParseSearchInputPartialField(t *testing.T) {
	result := parseSearchInput("stat")
	if len(result.tokens) != 0 {
		t.Fatalf("expected no tokens while typing field, got %d", len(result.tokens))
	}
	if result.mode != SuggestionModeField {
		t.Fatalf("expected field mode, got %v", result.mode)
	}
	if result.pendingText != "stat" {
		t.Fatalf("expected pending text 'stat', got %q", result.pendingText)
	}
}

func TestSearchOverlayUpdateInputStoresTokens(t *testing.T) {
	overlay := NewSearchOverlay()
	overlay.UpdateInput("status:open owner:\"me\"")
	tokens := overlay.Tokens()
	if len(tokens) != 2 {
		t.Fatalf("expected two tokens, got %d", len(tokens))
	}
	if tokens[1].Key != "owner" || tokens[1].Value != "me" {
		t.Fatalf("unexpected second token: %+v", tokens[1])
	}
	if overlay.SuggestionMode() != SuggestionModeField {
		t.Fatalf("expected overlay back in field mode after full tokens")
	}
}

func TestParseSearchInputCollectsFreeText(t *testing.T) {
	result := parseSearchInput("auth status:open login")
	if len(result.tokens) != 1 {
		t.Fatalf("expected one token, got %d", len(result.tokens))
	}
	if len(result.freeText) != 2 {
		t.Fatalf("expected two free-text terms, got %v", result.freeText)
	}
	if result.freeText[0] != "auth" || result.freeText[1] != "login" {
		t.Fatalf("unexpected free-text terms: %v", result.freeText)
	}
}

func TestParseSearchInputSetsErrorForUnterminatedQuote(t *testing.T) {
	result := parseSearchInput(`status:"open`)
	if result.err == "" {
		t.Fatalf("expected parse error for unterminated quote")
	}
}

func TestMatchesTokensEvaluatesNodes(t *testing.T) {
	node := &graph.Node{Issue: beads.FullIssue{ID: "ab-1", Title: "Auth Login", Status: "open", Labels: []string{"beta"}}}
	if !nodeMatchesTokenFilter(nil, []SearchToken{{Key: "status", Operator: ":", Value: "open"}}, node) {
		t.Fatalf("expected status token to match node")
	}
	if nodeMatchesTokenFilter(nil, []SearchToken{{Key: "status", Operator: ":", Value: "closed"}}, node) {
		t.Fatalf("expected status mismatch to fail")
	}
	if !nodeMatchesTokenFilter([]string{"auth"}, nil, node) {
		t.Fatalf("expected free-text term to match title")
	}
}
