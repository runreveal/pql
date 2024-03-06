// Copyright 2024 RunReveal Inc.
// SPDX-License-Identifier: Apache-2.0

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
	// surrounded by backticks
	// The Value will be the content between the backticks
	// with any double backticks reduced.
	TokenQuotedIdentifier
	// TokenNumber is a numeric literal like "123", "3.14", "1e-9", or "0xdeadbeef".
	// The Value will be a decimal formatted string.
	TokenNumber
	// TokenString is a string literal enclosed by single or double quotes.
	// The Value will be the literal's value (i.e. any escape sequences are evaluated).
	TokenString

	// TokenAnd is the keyword "and".
	// The Value will be the empty string.
	TokenAnd
	// TokenOr is the keyword "or".
	// The Value will be the empty string.
	TokenOr

	// TokenPipe is a single pipe character ("|").
	// The Value will be the empty string.
	TokenPipe
	// TokenDot is a period character (".").
	// The Value will be the empty string.
	TokenDot
	// TokenComma is a comma character (",").
	// The Value will be the empty string.
	TokenComma
	// TokenPlus is a single plus character ("+").
	// The Value will be the empty string.
	TokenPlus
	// TokenMinus is a single hyphen character ("-").
	// The Value will be the empty string.
	TokenMinus
	// TokenStar is a single asterisk character ("*").
	// The Value will be the empty string.
	TokenStar
	// TokenSlash is a single forward slash character ("/").
	// The Value will be the empty string.
	TokenSlash
	// TokenMod is a single percent sign character ("%").
	// The Value will be the empty string.
	TokenMod
	// TokenAssign is a single equals sign character ("=").
	// The Value will be the empty string.
	TokenAssign
	// TokenEq is a sequence of two equals sign characters ("==").
	// The Value will be the empty string.
	TokenEq
	// TokenNE is the sequence "!=", representing an inequality test.
	// The Value will be the empty string.
	TokenNE
	// TokenLT is the less than symbol ("<").
	// The Value will be the empty string.
	TokenLT
	// TokenLE is the less than or equal sequence "<=".
	// The Value will be the empty string.
	TokenLE
	// TokenGT is the greater than symbol (">").
	// The Value will be the empty string.
	TokenGT
	// TokenGE is the greater than or equal sequence ">=".
	// The Value will be the empty string.
	TokenGE
	// TokenCaseInsensitiveEq is the sequence "=~".
	// The Value will be the empty string.
	TokenCaseInsensitiveEq
	// TokenCaseInsensitiveNE is the sequence "!~".
	// The Value will be the empty string.
	TokenCaseInsensitiveNE

	// TokenLParen is a left parenthesis.
	// The Value will be the empty string.
	TokenLParen
	// TokenRParen is a right parenthesis.
	// The Value will be the empty string.
	TokenRParen

	// TokenLBracket is a left bracket ("[").
	// The Value will be the empty string.
	TokenLBracket
	// TokenRBracket is a right bracket ("]").
	// The Value will be the empty string.
	TokenRBracket

	// TokenBy is the keyword "in".
	// The Value will be the empty string.
	TokenIn
	// TokenBy is the keyword "by".
	// The Value will be the empty string.
	TokenBy

	// TokenSemi is the semicolon character (";").
	// The Value will be the empty string.
	TokenSemi

	// TokenLet is the keyword "let".
	// The Value will be the empty string.
	TokenLet

	// TokenBackslash is the character "\"
	// The Value will be the empty string.
	TokenBackslash

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
		case isAlpha(c) || c == '_' || c == '$':
			s.prev()
			tokens = append(tokens, s.ident())
		case isDigit(c) || c == '.':
			s.prev()
			tokens = append(tokens, s.numberOrDot())
		case c == ',':
			tokens = append(tokens, Token{
				Kind: TokenComma,
				Span: newSpan(start, s.pos),
			})
		case c == '"' || c == '\'':
			s.prev()
			tokens = append(tokens, s.string())
		case c == '`':
			s.prev()
			tokens = append(tokens, s.quotedIdent())
		case c == '|':
			tokens = append(tokens, Token{
				Kind: TokenPipe,
				Span: newSpan(start, s.pos),
			})
		case c == '(':
			tokens = append(tokens, Token{
				Kind: TokenLParen,
				Span: newSpan(start, s.pos),
			})
		case c == ')':
			tokens = append(tokens, Token{
				Kind: TokenRParen,
				Span: newSpan(start, s.pos),
			})
		case c == '[':
			tokens = append(tokens, Token{
				Kind: TokenLBracket,
				Span: newSpan(start, s.pos),
			})
		case c == ']':
			tokens = append(tokens, Token{
				Kind: TokenRBracket,
				Span: newSpan(start, s.pos),
			})
		case c == '\\':
			tokens = append(tokens, Token{
				Kind: TokenBackslash,
				Span: newSpan(start, s.pos),
			})
		case c == '=':
			c, ok := s.next()
			switch {
			case ok && c == '=':
				tokens = append(tokens, Token{
					Kind: TokenEq,
					Span: newSpan(start, s.pos),
				})
			case ok && c == '~':
				tokens = append(tokens, Token{
					Kind: TokenCaseInsensitiveEq,
					Span: newSpan(start, s.pos),
				})
			default:
				if ok {
					s.prev()
				}
				tokens = append(tokens, Token{
					Kind: TokenAssign,
					Span: newSpan(start, s.pos),
				})
			}
		case c == '!':
			c, ok := s.next()
			switch {
			case ok && c == '=':
				tokens = append(tokens, Token{
					Kind: TokenNE,
					Span: newSpan(start, s.pos),
				})
			case ok && c == '~':
				tokens = append(tokens, Token{
					Kind: TokenCaseInsensitiveNE,
					Span: newSpan(start, s.pos),
				})
			default:
				// TODO(maybe): Turn this into logical inversion?
				// KQL seems to use the not() function.
				tokens = append(tokens,
					errorToken(newSpan(start, s.pos), "unrecognized token '!'"),
				)
			}
		case c == '+':
			tokens = append(tokens, Token{
				Kind: TokenPlus,
				Span: newSpan(start, s.pos),
			})
		case c == '-':
			tokens = append(tokens, Token{
				Kind: TokenMinus,
				Span: newSpan(start, s.pos),
			})
		case c == '*':
			tokens = append(tokens, Token{
				Kind: TokenStar,
				Span: newSpan(start, s.pos),
			})
		case c == '/':
			// Check for double-slash comment.
			c, ok = s.next()
			if !ok {
				tokens = append(tokens, Token{
					Kind: TokenSlash,
					Span: newSpan(start, s.pos),
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
				Span: newSpan(start, s.pos),
			})
		case c == '%':
			tokens = append(tokens, Token{
				Kind: TokenMod,
				Span: newSpan(start, s.pos),
			})
		case c == '<':
			if c, ok := s.next(); ok && c == '=' {
				tokens = append(tokens, Token{
					Kind: TokenLE,
					Span: newSpan(start, s.pos),
				})
			} else {
				if ok {
					s.prev()
				}
				tokens = append(tokens, Token{
					Kind: TokenLT,
					Span: newSpan(start, s.pos),
				})
			}
		case c == '>':
			if c, ok := s.next(); ok && c == '=' {
				tokens = append(tokens, Token{
					Kind: TokenGE,
					Span: newSpan(start, s.pos),
				})
			} else {
				if ok {
					s.prev()
				}
				tokens = append(tokens, Token{
					Kind: TokenGT,
					Span: newSpan(start, s.pos),
				})
			}
		case c == ';':
			tokens = append(tokens, Token{
				Kind: TokenSemi,
				Span: newSpan(start, s.pos),
			})
		default:
			span := newSpan(start, s.pos)
			tokens = append(tokens, errorToken(span, "unrecognized character %q", spanString(query, span)))
		}
	}
	return tokens
}

// SplitStatements splits the given string by semicolons.
func SplitStatements(source string) []string {
	tokens := Scan(source)
	var parts []string
	start := 0
	for _, tok := range tokens {
		if tok.Kind == TokenBackslash {
			parts = append(parts, source[start:tok.Span.Start])
			start = tok.Span.End
		}
	}
	parts = append(parts, source[start:])
	return parts
}

var keywords = map[string]TokenKind{
	"and": TokenAnd,
	"by":  TokenBy,
	"in":  TokenIn,
	"or":  TokenOr,
}

func (s *scanner) ident() Token {
	start := s.pos
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
	tok := Token{
		Kind: TokenIdentifier,
		Span: newSpan(start, s.pos),
	}
	tok.Value = spanString(s.s, tok.Span)
	if kind, ok := keywords[tok.Value]; ok {
		tok.Kind = kind
		tok.Value = ""
	}
	return tok
}

func (s *scanner) quotedIdent() Token {
	start := s.pos
	if c, ok := s.next(); !ok || c != '`' {
		return errorToken(newSpan(start, s.pos), "parse quoted identifier: expected '`', found %q", c)
	}

	for {
		c, ok := s.next()
		if !ok {
			return errorToken(newSpan(start, s.pos), "parse quoted identifier: unexpected EOF")
		}
		switch c {
		case '`':
			// Check if double backtick.
			c, ok = s.next()
			if !ok || c != '`' {
				if ok {
					s.prev()
				}
				return Token{
					Kind:  TokenQuotedIdentifier,
					Span:  newSpan(start, s.pos),
					Value: strings.ReplaceAll(s.s[start+len("`"):s.pos-len("`")], "``", "`"),
				}
			}
		case '\n':
			s.prev()
			return errorToken(newSpan(start, s.pos), "parse quoted identifier: unexpected end of line")
		}
	}
}

func (s *scanner) numberOrDot() Token {
	start := s.pos
	c, ok := s.next()
	if !ok {
		return errorToken(indexSpan(start), "parse numeric literal: unexpected EOF")
	}

	// First character.
	hasDecimalPoint := false
	switch {
	case c == '0':
		c, ok := s.next()
		if !ok {
			return Token{
				Kind:  TokenNumber,
				Span:  newSpan(start, s.pos),
				Value: "0",
			}
		}
		switch {
		case c == '.':
			hasDecimalPoint = true
		case c == 'e' || c == 'E':
			s.prev()
			s.numberExponent()
			span := newSpan(start, s.pos)
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
					Span:  newSpan(start, s.pos),
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
			span := newSpan(start, s.pos)
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
				Span: newSpan(start, s.pos),
			}
		}
		if !isDigit(c) {
			s.prev()
			return Token{
				Kind: TokenDot,
				Span: newSpan(start, s.pos),
			}
		}
	case !isDigit(c):
		end := s.pos
		s.prev()
		return errorToken(newSpan(start, end), "parse numeric literal: unexpected character %q", c)
	}

	// Subsequent decimal digits.
	for {
		c, ok := s.next()
		switch {
		case !ok:
			span := newSpan(start, s.pos)
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
			span := newSpan(start, s.pos)
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

func (s *scanner) string() Token {
	start := s.pos
	quoteChar, ok := s.next()
	if !ok {
		return errorToken(indexSpan(start), "unexpected EOF (expected string)")
	}
	if quoteChar != '\'' && quoteChar != '"' {
		s.prev()
		return errorToken(indexSpan(start), "unexpected %q (expected string)", quoteChar)
	}

	valueStart := s.pos
	var valueBuilder *strings.Builder // nil if no escapes encountered
	for {
		c, ok := s.next()
		if !ok {
			return errorToken(newSpan(start, s.pos), "unterminated string")
		}
		switch c {
		case quoteChar:
			var value string
			if valueBuilder == nil {
				value = s.s[valueStart:s.last]
			} else {
				value = valueBuilder.String()
			}
			return Token{
				Kind:  TokenString,
				Span:  newSpan(start, s.pos),
				Value: value,
			}
		case '\n':
			s.prev()
			return errorToken(newSpan(start, s.pos), "unterminated string")
		case '\\':
			if valueBuilder == nil {
				valueBuilder = new(strings.Builder)
				valueBuilder.WriteString(s.s[valueStart:s.last])
			}
			c, ok := s.next()
			if !ok {
				return errorToken(newSpan(start, s.pos), "unterminated string")
			}
			switch c {
			case '\n':
				s.prev()
				return errorToken(newSpan(start, s.pos), "unterminated string")
			case 'n':
				valueBuilder.WriteRune('\n')
			case 't':
				valueBuilder.WriteRune('\t')
			default:
				valueBuilder.WriteRune(c)
			}
		default:
			if valueBuilder != nil {
				valueBuilder.WriteRune(c)
			}
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

func (s *scanner) setPos(pos int) {
	s.pos = pos
	s.last = pos
}

func isLet(ident *Ident) bool {
	return ident.Name == "let" && !ident.Quoted
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
