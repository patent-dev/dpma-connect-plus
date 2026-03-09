package query

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

// Query represents a parsed DPMAregister expert search query.
type Query struct {
	// Raw is the original query string.
	Raw string

	// Tokens is the parsed token stream.
	Tokens []Token

	// Valid indicates whether the query passed validation.
	Valid bool

	// Errors contains any validation errors.
	Errors []string

	// Service is the service context used for field validation.
	Service Service
}

// Token represents a token in a DPMAregister query.
type Token struct {
	Type  TokenType
	Value string
	Pos   int
}

// TokenType represents the type of a query token.
type TokenType int

// TokenType constants for DPMAregister query tokens.
const (
	TokenField      TokenType = iota // Field name (before =)
	TokenOperator                    // Boolean operator (AND, OR, NOT, UND, ODER, NICHT)
	TokenValue                       // Search value
	TokenEquals                      // Comparison operator (=, >=, <=, >, <)
	TokenLParen                      // (
	TokenRParen                      // )
	TokenLBrace                      // { (procedure data)
	TokenRBrace                      // } (procedure data)
	TokenQuote                       // "
	TokenWhitespace                  // Whitespace (stripped during tokenization)
	TokenUnknown                     // Unrecognized
)

// String returns a human-readable name for a token type.
func (t TokenType) String() string {
	switch t {
	case TokenField:
		return "FIELD"
	case TokenOperator:
		return "OPERATOR"
	case TokenValue:
		return "VALUE"
	case TokenEquals:
		return "EQUALS"
	case TokenLParen:
		return "LPAREN"
	case TokenRParen:
		return "RPAREN"
	case TokenLBrace:
		return "LBRACE"
	case TokenRBrace:
		return "RBRACE"
	case TokenQuote:
		return "QUOTE"
	case TokenWhitespace:
		return "WHITESPACE"
	default:
		return "UNKNOWN"
	}
}

// ParseQuery parses a DPMAregister expert search query string and returns
// a Query with tokens and validation results.
//
// If service is not ServiceAny, field names are validated against that service's
// known field codes.
//
// Example queries:
//   - "TI=Elektrofahrzeug"
//   - "TI=Elektrofahrzeug AND INH=Siemens"
//   - "(TI=Motor OR TI=Antrieb) AND IC=H02K?"
//   - "INH=\"München\" AND AT>=01.01.2024"
//   - "{VST=pub-offenlegungschrift UND VSTT=05.01.2011}"
//   - "MARKE=?brain?"
//   - "exists INH"
func ParseQuery(queryStr string, service Service) (*Query, error) {
	if strings.TrimSpace(queryStr) == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	q := &Query{
		Raw:     queryStr,
		Tokens:  tokenize(queryStr),
		Valid:   true,
		Errors:  []string{},
		Service: service,
	}

	q.validate()

	return q, nil
}

// tokenize splits a query string into tokens.
func tokenize(query string) []Token {
	var tokens []Token
	var current strings.Builder
	var inQuotes bool
	var pos int

	runes := []rune(query)

	flushCurrent := func(forceType TokenType) {
		if current.Len() == 0 {
			return
		}
		val := current.String()
		t := forceType
		if t == TokenUnknown {
			t = classifyToken(val)
		}
		tokens = append(tokens, Token{Type: t, Value: val, Pos: pos})
		current.Reset()
	}

	for i := 0; i < len(runes); i++ {
		ch := runes[i]

		switch {
		case ch == '"':
			if inQuotes {
				// End of quoted value - flush as value
				tokens = append(tokens, Token{Type: TokenValue, Value: current.String(), Pos: pos})
				current.Reset()
				tokens = append(tokens, Token{Type: TokenQuote, Value: "\"", Pos: i})
				inQuotes = false
			} else {
				// Start of quoted value
				flushCurrent(TokenUnknown)
				tokens = append(tokens, Token{Type: TokenQuote, Value: "\"", Pos: i})
				inQuotes = true
			}
			pos = i + 1

		case inQuotes:
			if current.Len() == 0 {
				pos = i
			}
			current.WriteRune(ch)

		case (ch == '=' || ch == '<' || ch == '>') && !inQuotes:
			flushCurrent(TokenField)

			// Check for compound operators (>=, <=)
			if (ch == '>' || ch == '<') && i+1 < len(runes) && runes[i+1] == '=' {
				tokens = append(tokens, Token{Type: TokenEquals, Value: string([]rune{ch, '='}), Pos: i})
				i++
			} else {
				tokens = append(tokens, Token{Type: TokenEquals, Value: string(ch), Pos: i})
			}
			pos = i + 1

		case ch == '(':
			flushCurrent(TokenUnknown)
			tokens = append(tokens, Token{Type: TokenLParen, Value: "(", Pos: i})
			pos = i + 1

		case ch == ')':
			flushCurrent(TokenUnknown)
			tokens = append(tokens, Token{Type: TokenRParen, Value: ")", Pos: i})
			pos = i + 1

		case ch == '{':
			flushCurrent(TokenUnknown)
			tokens = append(tokens, Token{Type: TokenLBrace, Value: "{", Pos: i})
			pos = i + 1

		case ch == '}':
			flushCurrent(TokenUnknown)
			tokens = append(tokens, Token{Type: TokenRBrace, Value: "}", Pos: i})
			pos = i + 1

		case unicode.IsSpace(ch):
			flushCurrent(TokenUnknown)
			pos = i + 1

		default:
			if current.Len() == 0 {
				pos = i
			}
			current.WriteRune(ch)
		}
	}

	// Flush remaining token
	flushCurrent(TokenUnknown)

	return tokens
}

// classifyToken determines the type of a bare token.
func classifyToken(value string) TokenType {
	if IsValidOperator(value) {
		return TokenOperator
	}
	// "exists" is a special keyword
	if strings.EqualFold(value, "exists") {
		return TokenOperator
	}
	return TokenValue
}

// validate performs all validation checks on the parsed query.
func (q *Query) validate() {
	q.checkBracketMatching()
	q.checkBraceMatching()
	q.checkFieldNames()
	q.checkQueryStructure()

	if len(q.Errors) > 0 {
		q.Valid = false
	}
}

// checkBracketMatching verifies that parentheses are properly matched.
func (q *Query) checkBracketMatching() {
	depth := 0
	for _, token := range q.Tokens {
		switch token.Type {
		case TokenLParen:
			depth++
		case TokenRParen:
			depth--
			if depth < 0 {
				q.Errors = append(q.Errors, fmt.Sprintf("unmatched closing parenthesis at position %d", token.Pos))
				return
			}
		}
	}
	if depth > 0 {
		q.Errors = append(q.Errors, fmt.Sprintf("unclosed parentheses: %d opening without matching closing", depth))
	}
}

// checkBraceMatching verifies that curly braces are properly matched.
func (q *Query) checkBraceMatching() {
	depth := 0
	for _, token := range q.Tokens {
		switch token.Type {
		case TokenLBrace:
			depth++
		case TokenRBrace:
			depth--
			if depth < 0 {
				q.Errors = append(q.Errors, fmt.Sprintf("unmatched closing brace at position %d", token.Pos))
				return
			}
		}
	}
	if depth > 0 {
		q.Errors = append(q.Errors, fmt.Sprintf("unclosed braces: %d opening without matching closing", depth))
	}
}

// checkFieldNames validates that all field names are recognized for the configured service.
func (q *Query) checkFieldNames() {
	for i, token := range q.Tokens {
		if token.Type != TokenField {
			continue
		}
		// Token is before an equals sign, so it's a field name
		if i+1 < len(q.Tokens) && q.Tokens[i+1].Type == TokenEquals {
			if !IsValidField(token.Value, q.Service) {
				q.Errors = append(q.Errors, fmt.Sprintf(
					"unknown field %q at position %d", token.Value, token.Pos,
				))
			}
		}
	}
}

// checkQueryStructure validates the overall structure of the query.
func (q *Query) checkQueryStructure() {
	if len(q.Tokens) == 0 {
		q.Errors = append(q.Errors, "query has no tokens")
		return
	}

	// Must contain at least one field=value pattern or "exists FIELD"
	hasPattern := false
	for i := 0; i < len(q.Tokens)-1; i++ {
		if q.Tokens[i].Type == TokenField && q.Tokens[i+1].Type == TokenEquals {
			hasPattern = true
			break
		}
		// "exists FIELD" pattern
		if q.Tokens[i].Type == TokenOperator && strings.EqualFold(q.Tokens[i].Value, "exists") {
			hasPattern = true
			break
		}
	}

	if !hasPattern && len(q.Errors) == 0 {
		q.Errors = append(q.Errors, "query must contain at least one FIELD=value pattern")
	}
}

// Validate returns an error if the query is invalid, or nil if valid.
func (q *Query) Validate() error {
	if q.Valid {
		return nil
	}
	if len(q.Errors) == 1 {
		return fmt.Errorf("query validation error: %s", q.Errors[0])
	}
	return fmt.Errorf("query validation errors: %s", strings.Join(q.Errors, "; "))
}

// URLEncode returns the query URL-encoded for use in API requests.
func (q *Query) URLEncode() string {
	return url.QueryEscape(q.Raw)
}

// String returns the raw query string.
func (q *Query) String() string {
	return q.Raw
}

// TokenCount returns the number of tokens in the parsed query.
func (q *Query) TokenCount() int {
	return len(q.Tokens)
}

// HasField checks if the query references a specific field code.
func (q *Query) HasField(field string) bool {
	for i, token := range q.Tokens {
		if token.Type == TokenField && token.Value == field &&
			i+1 < len(q.Tokens) && q.Tokens[i+1].Type == TokenEquals {
			return true
		}
	}
	return false
}

// GetFields returns all unique field codes used in the query.
func (q *Query) GetFields() []string {
	seen := make(map[string]bool)
	var result []string
	for i, token := range q.Tokens {
		if token.Type == TokenField && i+1 < len(q.Tokens) && q.Tokens[i+1].Type == TokenEquals {
			if !seen[token.Value] {
				seen[token.Value] = true
				result = append(result, token.Value)
			}
		}
	}
	return result
}
