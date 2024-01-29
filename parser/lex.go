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
	// TokenIdentifier is a plain identifier
	// that might be a keyword, depending on position.
	// The Value will be the identifier itself.
	TokenIdentifier TokenKind = 1 + iota
	// TokenQuotedIdentifier is an identifier
	// surrounded by `['` and `']` or `["` and `"]`.
	// The Value will be the contents of the quoted string.
	TokenQuotedIdentifier
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

func errorToken(span Span, format string, args ...any) Token {
	return Token{
		Kind:  TokenError,
		Span:  span,
		Value: fmt.Sprintf(format, args...),
	}
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
		case c == '[':
			s.prev()
			tokens = append(tokens, s.quotedIdent())
		default:
			span := Span{Start: start, End: s.pos}
			tokens = append(tokens, errorToken(span, "unrecognized character %q", spanString(query, span)))
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

func (s *scanner) quotedIdent() Token {
	span := Span{Start: s.pos}
	if c, ok := s.next(); !ok || c != '[' {
		span.End = s.pos
		return errorToken(span, "parse quoted identifier: expected '[', found %q", c)
	}
	quoteChar, ok := s.next()
	if !ok {
		span.End = s.pos
		return errorToken(span, "parse quoted identifier: expected '[', found %q", quoteChar)
	}
	if quoteChar != '\'' && quoteChar != '"' {
		s.prev()
		span.End = s.pos
		return errorToken(span, "parse quoted identifier: expected ' or \", found %q", quoteChar)
	}

	for {
		// Check for terminator.
		tail := s.s[s.pos:]
		if len(tail) >= 2 && tail[0] == byte(quoteChar) && tail[1] == ']' {
			value := s.s[span.Start+len(`["`) : s.pos]
			s.pos += len(`"]`)
			span.End = s.pos
			return Token{
				Kind:  TokenQuotedIdentifier,
				Span:  span,
				Value: value,
			}
		}

		// Now check for end of line or end of query.
		c, ok := s.next()
		if !ok {
			span.End = s.pos
			return errorToken(span, "parse quoted identifier: unexpected EOF")
		}
		if c == '\n' {
			s.prev()
			span.End = s.pos
			return errorToken(span, "parse quoted identifier: unexpected end of line")
		}
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
