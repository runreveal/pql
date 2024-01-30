//go:generate stringer -type=TokenKind

package parser

import (
	"fmt"
	"strconv"
	"strings"
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
	// TokenNumber is a numeric literal like "123", "3.14", "1e-9", or "0xdeadbeef".
	// The Value will be a decimal formatted string.
	TokenNumber

	// TokenPipe is a single pipe character ("|").
	// The Value will be the empty string.
	TokenPipe
	// TokenDot is a period character (".").
	// The Value will be the empty string.
	TokenDot
	// TokenSlash is a single slash character ("/").
	// The Value will be the empty string.
	TokenSlash

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
		case isDigit(c) || c == '.':
			s.prev()
			tokens = append(tokens, s.numberOrDot())
		case c == '|':
			tokens = append(tokens, Token{
				Kind: TokenPipe,
				Span: Span{Start: start, End: s.pos},
			})
		case c == '[':
			s.prev()
			tokens = append(tokens, s.quotedIdent())
		case c == '/':
			// Check for double-slash comment.
			c, ok = s.next()
			if !ok {
				tokens = append(tokens, Token{
					Kind: TokenSlash,
					Span: Span{Start: start, End: s.pos},
				})
				continue
			}
			if c == '/' {
				// It's a comment, consume to end of line.
				for {
					c, ok = s.next()
					if !ok || c == '\n' {
						break
					}
				}
				continue
			}
			s.prev()
			tokens = append(tokens, Token{
				Kind: TokenSlash,
				Span: Span{Start: start, End: s.pos},
			})
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

func (s *scanner) numberOrDot() Token {
	start := s.pos
	c, ok := s.next()
	if !ok {
		return errorToken(Span{Start: start, End: start}, "parse numeric literal: unexpected EOF")
	}

	// First character.
	hasDecimalPoint := false
	switch {
	case c == '0':
		c, ok := s.next()
		if !ok {
			return Token{
				Kind:  TokenNumber,
				Span:  Span{Start: start, End: s.pos},
				Value: "0",
			}
		}
		switch {
		case c == '.':
			hasDecimalPoint = true
		case c == 'e' || c == 'E':
			s.prev()
			s.numberExponent()
			span := Span{Start: start, End: s.pos}
			return Token{
				Kind:  TokenNumber,
				Span:  span,
				Value: normalizeNumberValue(spanString(s.s, span)),
			}
		case c == 'x' || c == 'X':
			// Hexadecimal constant.
			hexDigitStart := s.pos
			c, ok := s.next()
			if !ok || !isHexDigit(c) {
				s.setPos(start + 2)
				return Token{
					Kind:  TokenError,
					Span:  Span{Start: start, End: s.pos},
					Value: "invalid hex literal",
				}
			}

			for {
				c, ok := s.next()
				if !ok {
					break
				}
				if !isHexDigit(c) {
					s.prev()
					break
				}
			}
			span := Span{Start: start, End: s.pos}
			n, err := strconv.ParseUint(s.s[hexDigitStart:s.pos], 16, 64)
			if err != nil {
				return errorToken(span, "parse hex literal: %v", err)
			}
			return Token{
				Kind:  TokenNumber,
				Span:  span,
				Value: strconv.FormatUint(n, 10),
			}
		case !isDigit(c):
			s.prev()
		}
	case c == '.':
		// Must have at least one subsequent digit to be considered a numeric literal.
		hasDecimalPoint = true
		c, ok := s.next()
		if !ok {
			return Token{
				Kind: TokenDot,
				Span: Span{Start: start, End: s.pos},
			}
		}
		if !isDigit(c) {
			s.prev()
			return Token{
				Kind: TokenDot,
				Span: Span{Start: start, End: s.pos},
			}
		}
	case !isDigit(c):
		end := s.pos
		s.prev()
		return errorToken(Span{Start: start, End: end}, "parse numeric literal: unexpected character %q", c)
	}

	// Subsequent decimal digits.
	for {
		c, ok := s.next()
		switch {
		case !ok:
			span := Span{Start: start, End: s.pos}
			return Token{
				Kind:  TokenNumber,
				Span:  span,
				Value: normalizeNumberValue(spanString(s.s, span)),
			}
		case c == '.' && !hasDecimalPoint:
			hasDecimalPoint = true
		case !isDigit(c):
			s.prev()
			s.numberExponent()
			span := Span{Start: start, End: s.pos}
			return Token{
				Kind:  TokenNumber,
				Span:  span,
				Value: normalizeNumberValue(spanString(s.s, span)),
			}
		}
	}
}

func (s *scanner) numberExponent() (found bool) {
	start := s.pos
	defer func() {
		if !found {
			s.setPos(start)
		}
	}()

	c, ok := s.next()
	if !ok {
		return false
	}
	if c != 'e' && c != 'E' {
		return false
	}

	// Must have at least one digit.
	c, ok = s.next()
	if !ok {
		return false
	}
	if c == '+' || c == '-' {
		c, ok = s.next()
		if !ok {
			return false
		}
	}
	if !isDigit(c) {
		return false
	}

	for {
		c, ok = s.next()
		if !ok {
			return true
		}
		if !isDigit(c) {
			s.prev()
			return true
		}
	}
}

func normalizeNumberValue(s string) string {
	s = strings.TrimLeft(s, "0")
	switch {
	case s == "":
		return "0"
	case s[0] == '.' || s[0] == 'e' || s[0] == 'E':
		return "0" + s
	default:
		return s
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

func (s *scanner) setPos(pos int) {
	s.pos = pos
	s.last = pos
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

func isHexDigit(c rune) bool {
	return isDigit(c) || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F'
}
