package ui

import (
	"strings"
	"unicode"
	"unicode/utf8"
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
	mode         SuggestionMode
	pendingField string
	pendingText  string
}

func parseSearchInput(input string) searchParseResult {
	parser := tokenParser{input: input}
	result := searchParseResult{mode: SuggestionModeField}

	for {
		parser.skipWhitespace()
		if parser.eof() {
			result.pendingText = ""
			result.mode = SuggestionModeField
			return result
		}

		keyStart := parser.pos
		key := parser.readKey()
		if key == "" {
			result.pendingText = trimLeadingWhitespace(parser.input[keyStart:])
			result.mode = SuggestionModeField
			return result
		}

		parser.skipWhitespace()
		if parser.eof() {
			result.pendingText = key
			result.mode = SuggestionModeField
			return result
		}

		op := parser.peek()
		if op != ':' && op != '=' {
			parser.pos = keyStart
			result.pendingText = trimLeadingWhitespace(parser.input[keyStart:])
			result.mode = SuggestionModeField
			return result
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
