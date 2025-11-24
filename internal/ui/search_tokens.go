package ui

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"abacus/internal/graph"
)

// SearchToken represents a parsed field/operator/value triple.
type SearchToken struct {
	Key      string
	Operator string
	Value    string
}

// SuggestionMode indicates whether the overlay should suggest fields or values.
type SuggestionMode int

const (
	SuggestionModeField SuggestionMode = iota
	SuggestionModeValue
)

func (m SuggestionMode) String() string {
	switch m {
	case SuggestionModeValue:
		return "value"
	default:
		return "field"
	}
}

type searchParseResult struct {
	tokens       []SearchToken
	freeText     []string
	mode         SuggestionMode
	pendingField string
	pendingText  string
	err          string
}

func parseSearchInput(input string) searchParseResult {
	parser := tokenParser{input: input}
	result := searchParseResult{mode: SuggestionModeField}

	for {
		parser.skipWhitespace()
		if parser.eof() {
			return result
		}

		key := parser.readKey()
		if key == "" {
			parser.advance()
			continue
		}

		parser.skipWhitespace()
		if parser.eof() {
			result.pendingText = key
			result.freeText = appendWord(result.freeText, key)
			result.mode = SuggestionModeField
			return result
		}

		op := parser.peek()
		if op != ':' && op != '=' {
			result.freeText = appendWord(result.freeText, key)
			continue
		}
		operator := string(op)
		parser.advance()

		parser.skipWhitespace()
		if parser.eof() {
			result.pendingField = key
			result.pendingText = ""
			result.mode = SuggestionModeValue
			return result
		}

		valueStart := parser.pos
		value, complete := parser.readValue()
		if !complete {
			result.pendingField = key
			result.pendingText = parser.input[valueStart:]
			result.mode = SuggestionModeValue
			if value != "" {
				result.err = "unterminated quote"
			}
			return result
		}

		result.tokens = append(result.tokens, SearchToken{
			Key:      key,
			Operator: operator,
			Value:    value,
		})
	}
}

type tokenParser struct {
	input string
	pos   int
}

func (p *tokenParser) eof() bool {
	return p.pos >= len(p.input)
}

func (p *tokenParser) peek() rune {
	if p.eof() {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(p.input[p.pos:])
	return r
}

func (p *tokenParser) advance() {
	if p.eof() {
		return
	}
	_, size := utf8.DecodeRuneInString(p.input[p.pos:])
	p.pos += size
}

func (p *tokenParser) skipWhitespace() {
	for !p.eof() {
		r := p.peek()
		if !unicode.IsSpace(r) {
			return
		}
		p.advance()
	}
}

func (p *tokenParser) readKey() string {
	var b strings.Builder
	for !p.eof() {
		r := p.peek()
		if unicode.IsSpace(r) || r == ':' || r == '=' {
			break
		}
		b.WriteRune(r)
		p.advance()
	}
	if b.Len() == 0 {
		return ""
	}
	return strings.TrimSpace(b.String())
}

func (p *tokenParser) readValue() (string, bool) {
	if p.eof() {
		return "", false
	}
	if p.peek() == '"' {
		return p.readQuotedValue()
	}
	return p.readUnquotedValue()
}

func (p *tokenParser) readQuotedValue() (string, bool) {
	// consume opening quote
	p.advance()
	var b strings.Builder
	for !p.eof() {
		r := p.peek()
		p.advance()
		if r == '\\' {
			if p.eof() {
				return b.String(), false
			}
			escaped := p.peek()
			p.advance()
			b.WriteRune(escaped)
			continue
		}
		if r == '"' {
			return b.String(), true
		}
		b.WriteRune(r)
	}
	return b.String(), false
}

func (p *tokenParser) readUnquotedValue() (string, bool) {
	var b strings.Builder
	for !p.eof() {
		r := p.peek()
		if unicode.IsSpace(r) {
			break
		}
		b.WriteRune(r)
		p.advance()
	}
	if b.Len() == 0 {
		return "", false
	}
	return b.String(), true
}

func trimLeadingWhitespace(s string) string {
	return strings.TrimLeftFunc(s, unicode.IsSpace)
}

func appendWord(words []string, word string) []string {
	word = strings.TrimSpace(word)
	if word == "" {
		return words
	}
	return append(words, strings.ToLower(word))
}

func matchesTokens(tokens []SearchToken, node *graph.Node) bool {
	if len(tokens) == 0 {
		return true
	}
	for _, token := range tokens {
		if !matchesToken(token, node) {
			return false
		}
	}
	return true
}

func matchesToken(token SearchToken, node *graph.Node) bool {
	issue := node.Issue
	value := strings.ToLower(token.Value)
	switch strings.ToLower(token.Key) {
	case "status":
		return strings.EqualFold(issue.Status, value)
	case "id":
		return strings.Contains(strings.ToLower(issue.ID), value)
	case "title":
		return strings.Contains(strings.ToLower(issue.Title), value)
	case "label", "labels":
		for _, label := range issue.Labels {
			if strings.EqualFold(label, token.Value) {
				return true
			}
		}
		return false
	case "priority", "prio":
		return fmt.Sprintf("%v", issue.Priority) == token.Value
	case "type", "issue_type":
		return strings.EqualFold(issue.IssueType, token.Value)
	case "blocked":
		boolVal, _ := strconv.ParseBool(value)
		return node.IsBlocked == boolVal
	default:
		return strings.Contains(strings.ToLower(issue.Title), value) || strings.Contains(strings.ToLower(issue.ID), value)
	}
}
