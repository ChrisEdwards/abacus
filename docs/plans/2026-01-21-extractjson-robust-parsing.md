# extractJSON Robust Parsing Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Bead:** ab-o909

**Goal:** Make `extractJSON` robust to braces inside JSON string values and quoted text in warnings.

**Architecture:** Replace naive brace-counting with a string-aware scanner that tracks whether we're inside a JSON string literal (handling escape sequences). Use `json.Valid()` as a final validation step.

**Tech Stack:** Go standard library (`encoding/json`, `bytes`)

---

## Problem Analysis

The current `extractJSON` function at `internal/beads/bd_cli.go:373-394` counts `{` and `}` naively:

```go
for i := start; i < len(out); i++ {
    switch out[i] {
    case '{':
        depth++
    case '}':
        depth--
        if depth == 0 {
            return out[start : i+1]
        }
    }
}
```

This fails when:
1. **Braces in string values:** `{"msg": "use { and } carefully"}` - inner braces counted incorrectly
2. **Braces in warnings:** `Warning: {config} not found\n{"id":"ab-123"}` - warning braces confuse the scanner

## Solution

Implement a string-aware scanner that:
1. Tracks when inside a JSON string (between unescaped `"`)
2. Handles escape sequences (`\"`, `\\`)
3. Only counts braces when NOT inside a string
4. Validates result with `json.Valid()`

---

## Task 1: Add Failing Tests for Braces in Strings

**Files:**
- Modify: `internal/beads/bd_cli_test.go:322-362`

**Step 1: Write the failing tests**

Add these test cases to `TestExtractJSON`:

```go
t.Run("HandlesEscapedQuotesInStrings", func(t *testing.T) {
    input := []byte(`{"msg":"say \"hello\" world"}`)
    result := extractJSON(input)
    if string(result) != `{"msg":"say \"hello\" world"}` {
        t.Errorf("expected JSON with escaped quotes, got: %s", result)
    }
})

t.Run("HandlesBracesInsideStrings", func(t *testing.T) {
    input := []byte(`{"msg":"use { and } carefully","id":"ab-123"}`)
    result := extractJSON(input)
    if string(result) != `{"msg":"use { and } carefully","id":"ab-123"}` {
        t.Errorf("expected JSON with braces in string, got: %s", result)
    }
})

t.Run("HandlesBracesInWarningBeforeJSON", func(t *testing.T) {
    input := []byte("Warning: {config} file missing\n" + `{"id":"ab-123","status":"open"}`)
    result := extractJSON(input)
    if string(result) != `{"id":"ab-123","status":"open"}` {
        t.Errorf("expected JSON after warning with braces, got: %s", result)
    }
})

t.Run("HandlesNestedBracesInStrings", func(t *testing.T) {
    input := []byte(`{"data":{"nested":"{\"inner\":\"value\"}"}}`)
    result := extractJSON(input)
    if string(result) != `{"data":{"nested":"{\"inner\":\"value\"}"}}` {
        t.Errorf("expected nested JSON string, got: %s", result)
    }
})

t.Run("HandlesBackslashEscapeSequences", func(t *testing.T) {
    input := []byte(`{"path":"C:\\Users\\test\\file.txt"}`)
    result := extractJSON(input)
    if string(result) != `{"path":"C:\\Users\\test\\file.txt"}` {
        t.Errorf("expected JSON with backslash escapes, got: %s", result)
    }
})
```

**Step 2: Run tests to verify they fail**

Run: `make test VERBOSE=1 2>&1 | grep -A5 "HandlesBracesInsideStrings\|HandlesEscapedQuotes\|HandlesBracesInWarning\|HandlesNestedBraces\|HandlesBackslash" | head -30`

Expected: Multiple FAIL messages showing the bug

**Step 3: Commit failing tests**

```bash
git add internal/beads/bd_cli_test.go
git commit -m "test: add failing tests for extractJSON brace handling in strings

These tests demonstrate the bug where extractJSON counts braces
inside JSON string values, causing incorrect parsing.

Ref: ab-o909"
```

---

## Task 2: Implement String-Aware extractJSON

**Files:**
- Modify: `internal/beads/bd_cli.go:370-394`

**Step 1: Replace the extractJSON implementation**

Replace the current `extractJSON` function with:

```go
// extractJSON finds and returns the first valid JSON object in the output.
// bd/br commands may print warnings or other text before the actual JSON response.
// This function is shared between bdCLIClient and brCLIClient.
//
// The scanner is string-aware: braces inside JSON string values are not counted.
// It handles escape sequences like \" and \\ correctly.
func extractJSON(out []byte) []byte {
	// Try each '{' as a potential JSON start
	for start := 0; start < len(out); start++ {
		idx := bytes.IndexByte(out[start:], '{')
		if idx == -1 {
			return nil
		}
		start += idx

		// Scan for matching closing brace, tracking string state
		depth := 0
		inString := false
		for i := start; i < len(out); i++ {
			b := out[i]

			if inString {
				if b == '\\' && i+1 < len(out) {
					// Skip escaped character
					i++
					continue
				}
				if b == '"' {
					inString = false
				}
				continue
			}

			// Not in string
			switch b {
			case '"':
				inString = true
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					candidate := out[start : i+1]
					if json.Valid(candidate) {
						return candidate
					}
					// Invalid JSON, try next '{'
					break
				}
			}
		}
	}
	return nil
}
```

**Step 2: Add json import if needed**

Verify `encoding/json` is imported at top of file. If not, add it to the import block.

**Step 3: Run tests to verify they pass**

Run: `make test VERBOSE=1 2>&1 | grep -E "TestExtractJSON|PASS|FAIL"`

Expected: All TestExtractJSON subtests PASS

**Step 4: Run full test suite**

Run: `make check-test`

Expected: All checks and tests pass

**Step 5: Commit implementation**

```bash
git add internal/beads/bd_cli.go
git commit -m "fix: make extractJSON robust to braces in JSON strings

The scanner now tracks whether it's inside a JSON string literal,
correctly handling escape sequences (\\, \"). Braces inside strings
are not counted toward depth. Uses json.Valid() to verify candidates.

If the first candidate fails validation, subsequent '{' characters
are tried as potential JSON starts.

Fixes: ab-o909"
```

---

## Task 3: Add Edge Case Tests

**Files:**
- Modify: `internal/beads/bd_cli_test.go`

**Step 1: Add edge case tests**

Add these additional test cases for robustness:

```go
t.Run("HandlesEmptyStringValues", func(t *testing.T) {
    input := []byte(`{"empty":"","id":"ab-123"}`)
    result := extractJSON(input)
    if string(result) != `{"empty":"","id":"ab-123"}` {
        t.Errorf("expected JSON with empty string, got: %s", result)
    }
})

t.Run("HandlesUnicodeEscapes", func(t *testing.T) {
    input := []byte(`{"emoji":"\u263A","id":"ab-123"}`)
    result := extractJSON(input)
    if string(result) != `{"emoji":"\u263A","id":"ab-123"}` {
        t.Errorf("expected JSON with unicode escape, got: %s", result)
    }
})

t.Run("SkipsInvalidJSONToFindValid", func(t *testing.T) {
    input := []byte("{invalid json here}{\"id\":\"ab-123\"}")
    result := extractJSON(input)
    if string(result) != `{"id":"ab-123"}` {
        t.Errorf("expected to skip invalid and find valid JSON, got: %s", result)
    }
})

t.Run("HandlesArraysInValues", func(t *testing.T) {
    input := []byte(`{"tags":["one","two"],"id":"ab-123"}`)
    result := extractJSON(input)
    if string(result) != `{"tags":["one","two"],"id":"ab-123"}` {
        t.Errorf("expected JSON with array, got: %s", result)
    }
})
```

**Step 2: Run tests**

Run: `make test VERBOSE=1 2>&1 | grep -E "TestExtractJSON|PASS|FAIL"`

Expected: All tests pass

**Step 3: Commit edge case tests**

```bash
git add internal/beads/bd_cli_test.go
git commit -m "test: add edge case tests for extractJSON

Cover empty strings, unicode escapes, invalid JSON skipping,
and arrays in values.

Ref: ab-o909"
```

---

## Task 4: Final Verification and Bead Closure

**Step 1: Run full quality checks**

Run: `make check-test`

Expected: All checks and tests pass

**Step 2: Run integration tests (if bd/br available)**

Run: `make test-integration 2>&1 | tail -20`

Expected: Integration tests pass (or skip gracefully if binaries unavailable)

**Step 3: Add closing comment to bead**

```bash
bd comments add ab-o909 "Fixed extractJSON to be string-aware. The scanner now tracks JSON string state and handles escape sequences correctly. Uses json.Valid() for final validation. Added comprehensive tests including edge cases for escaped quotes, braces in strings, unicode escapes, and invalid JSON skipping."
```

**Step 4: Close the bead**

```bash
bd status ab-o909 closed
```

**Step 5: Sync and push**

```bash
bd sync
git push
```

---

## Summary

| Task | Description | Files |
|------|-------------|-------|
| 1 | Add failing tests for braces in strings | `bd_cli_test.go` |
| 2 | Implement string-aware extractJSON | `bd_cli.go` |
| 3 | Add edge case tests | `bd_cli_test.go` |
| 4 | Final verification and bead closure | - |

**Total commits:** 3 code commits + bead closure
