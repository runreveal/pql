//go:generate stringer -type=TokenKind

package parser

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

// TokenKind is an enumeration of types of [Token]
// that can be returned by [Scan].
type TokenKind int

// Token kinds.
const (
	// TokenIdentifier is an identifier.
	// The Value will be the identifier's equivalent symbol.
	TokenIdentifier TokenKind = 1 + iota
	TokenPipe

	// TokenError is a marker for a scan error.
	// The Value will contain the error message.
	TokenError TokenKind = -1
)

// Token is a syntactical element in a query.
type Token struct {
	// Kind is the token's type.
	Kind TokenKind
	// Span holds the location of the token.
	Span Span
	// Value contains kind-specific information about the token.
	// See the docs for [TokenKind] for what Value represents.
	Value string
}

type scanner struct {
	s    string
	pos  int
	last int
}

// Scan turns a Pipeline Query Language statement into a sequence of [Token] values.
// Errors will be indicated with the [TokenError] kind.
func Scan(query string) []Token {
	s := scanner{s: query}
	var tokens []Token
	for {
		start := s.pos
		c, ok := s.next()
		if !ok {
			break
		}
		switch {
		case unicode.IsSpace(c):
			// Skip insignificant whitespace.
		case isAlpha(c) || c == '_':
			s.prev()
			tokens = append(tokens, s.ident())
		case c == '|':
			tokens = append(tokens, Token{
				Kind: TokenPipe,
				Span: Span{Start: start, End: s.pos},
			})
		default:
			span := Span{Start: start, End: s.pos}
			tokens = append(tokens, Token{
				Kind:  TokenError,
				Span:  span,
				Value: fmt.Sprintf("unrecognized character %q", spanString(query, span)),
			})
		}
	}
	return tokens
}

func (s *scanner) ident() Token {
	span := Span{Start: s.pos}
	s.next() // assume that the caller validated first character
	for {
		c, ok := s.next()
		if !ok {
			break
		}
		if !(isAlpha(c) || isDigit(c) || c == '_') {
			s.prev()
			break
		}
	}
	span.End = s.pos
	return Token{
		Kind:  TokenIdentifier,
		Span:  span,
		Value: spanString(s.s, span),
	}
}

func (s *scanner) next() (rune, bool) {
	if s.pos >= len(s.s) {
		return 0, false
	}
	c, n := utf8.DecodeRuneInString(s.s[s.pos:])
	s.last = s.pos
	s.pos += n
	return c, true
}

func (s *scanner) prev() {
	s.pos = s.last
}

// A Span is a reference contiguous sequence of bytes in a query.
type Span struct {
	// Start is the index of the first byte of the span,
	// relative to the beginning of the query.
	Start int
	// End is the end index of the span (exclusive),
	// relative to the beginning of the query.
	End int
}

// IsValid reports whether the span has a non-negative length
// and non-negative indices.
func (span Span) IsValid() bool {
	return span.Start >= 0 && span.End >= 0 && span.Start <= span.End
}

// Len returns the length of the span
// or zero if the span is invalid.
func (span Span) Len() int {
	if !span.IsValid() {
		return 0
	}
	return span.End - span.Start
}

// String formats the span indices as a mathematical range like "[12,34)".
func (span Span) String() string {
	return fmt.Sprintf("[%d,%d)", span.Start, span.End)
}

func spanString(s string, span Span) string {
	if !span.IsValid() {
		return ""
	}
	return s[span.Start:span.End]
}

func isAlpha(c rune) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

func isDigit(c rune) bool {
	return '0' <= c && c <= '9'
}
