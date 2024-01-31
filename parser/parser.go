// Package parser provides a parser and an Abstract Syntax Tree (AST) for the Pipeline Query Language.
package parser

import (
	"errors"
	"fmt"
	"slices"
)

type parser struct {
	source string
	tokens []Token
	pos    int
}

// Parse converts a Pipeline Query Language tabular expression
// into an Abstract Syntax Tree (AST).
func Parse(query string) (*TabularExpr, error) {
	p := &parser{
		source: query,
		tokens: Scan(query),
	}
	expr, err := p.tabularExpr()
	if p.pos < len(p.tokens) {
		trailingToken := p.tokens[p.pos]
		if trailingToken.Kind == TokenError {
			err = joinErrors(err, &parseError{
				source: p.source,
				span:   trailingToken.Span,
				err:    errors.New(trailingToken.Value),
			})
		} else {
			err = joinErrors(err, &parseError{
				source: p.source,
				span:   trailingToken.Span,
				err:    errors.New("unrecognized token"),
			})
		}
	} else if isNotFound(err) {
		err = &parseError{
			source: p.source,
			span:   indexSpan(len(query)),
			err:    errors.New("empty query"),
		}
	}
	if err != nil {
		return expr, fmt.Errorf("parse pipeline query language: %w", err)
	}
	return expr, nil
}

func (p *parser) tabularExpr() (*TabularExpr, error) {
	tableName, err := p.ident()
	if err != nil {
		return nil, err
	}
	expr := &TabularExpr{
		Source: &TableRef{Table: tableName},
	}

	var returnedError error
	for {
		pipeToken, ok := p.next()
		if !ok {
			break
		}
		if pipeToken.Kind != TokenPipe {
			p.prev()
			break
		}

		operatorName, ok := p.next()
		if !ok {
			returnedError = joinErrors(returnedError, &parseError{
				source: p.source,
				span:   pipeToken.Span,
				err:    errors.New("missing operator name after pipe"),
			})
			return expr, returnedError
		}
		if operatorName.Kind != TokenIdentifier {
			// TODO(soon): Skip ahead to next pipe.
			returnedError = joinErrors(returnedError, &parseError{
				source: p.source,
				span:   operatorName.Span,
				err:    fmt.Errorf("expected operator name, got %s", formatToken(p.source, operatorName)),
			})
			return expr, returnedError
		}
		switch operatorName.Value {
		case "count":
			op, err := p.countOperator(pipeToken, operatorName)
			if op != nil {
				expr.Operators = append(expr.Operators, op)
			}
			returnedError = joinErrors(returnedError, err)
		case "where", "filter":
			op, err := p.whereOperator(pipeToken, operatorName)
			if op != nil {
				expr.Operators = append(expr.Operators, op)
			}
			returnedError = joinErrors(returnedError, err)
		default:
			// TODO(soon): Skip ahead to next pipe.
			returnedError = joinErrors(returnedError, &parseError{
				source: p.source,
				span:   operatorName.Span,
				err:    fmt.Errorf("unknown operator name %q", operatorName.Value),
			})
			return expr, returnedError
		}
	}
	return expr, returnedError
}

func (p *parser) countOperator(pipe, keyword Token) (*CountOperator, error) {
	return &CountOperator{
		Pipe:    pipe.Span,
		Keyword: keyword.Span,
	}, nil
}

func (p *parser) whereOperator(pipe, keyword Token) (*WhereOperator, error) {
	x, err := p.expr()
	err = makeErrorOpaque(err)
	return &WhereOperator{
		Pipe:      pipe.Span,
		Keyword:   keyword.Span,
		Predicate: x,
	}, err
}

// exprList parses one or more comma-separated expressions.
func (p *parser) exprList() ([]Expr, error) {
	first, err := p.expr()
	if err != nil {
		return nil, err
	}
	result := []Expr{first}
	for {
		tok, ok := p.next()
		if !ok {
			return result, nil
		}
		if tok.Kind != TokenComma {
			p.prev()
			return result, nil
		}
		x, err := p.expr()
		if x != nil {
			result = append(result, x)
		}
		if err != nil {
			// If there's a notFoundError, we want to mask it from the caller.
			return result, makeErrorOpaque(err)
		}
	}
}

func (p *parser) expr() (Expr, error) {
	// TODO(now)
	return p.unaryExpr()
}

func (p *parser) unaryExpr() (Expr, error) {
	tok, ok := p.next()
	if !ok {
		return nil, &parseError{
			source: p.source,
			span:   indexSpan(len(p.source)),
			err:    notFoundError{errors.New("expected expression, got EOF")},
		}
	}
	switch tok.Kind {
	case TokenPlus, TokenMinus:
		x, err := p.primaryExpr()
		err = makeErrorOpaque(err) // already parsed a symbol
		return &UnaryExpr{
			OpSpan: tok.Span,
			Op:     tok.Kind,
			X:      x,
		}, err
	default:
		p.prev()
		return p.primaryExpr()
	}
}

func (p *parser) primaryExpr() (Expr, error) {
	tok, ok := p.next()
	if !ok {
		return nil, &parseError{
			source: p.source,
			span:   indexSpan(len(p.source)),
			err:    notFoundError{errors.New("expected expression, got EOF")},
		}
	}
	switch tok.Kind {
	case TokenNumber, TokenString:
		return &BasicLit{
			ValueSpan: tok.Span,
			Kind:      tok.Kind,
			Value:     tok.Value,
		}, nil
	case TokenIdentifier:
		// Look ahead for opening parenthesis for a function call.
		nextTok, ok := p.next()
		if !ok {
			return &Ident{
				NameSpan: tok.Span,
				Name:     tok.Value,
			}, nil
		}

		if nextTok.Kind == TokenLParen {
			args, err := p.exprList()
			if isNotFound(err) {
				err = nil
			}
			finalTok, _ := p.next()
			if finalTok.Kind == TokenComma {
				finalTok, _ = p.next()
			}
			rparen := nullSpan()
			if finalTok.Kind == TokenRParen {
				rparen = finalTok.Span
			} else {
				err = joinErrors(err, &parseError{
					source: p.source,
					span:   finalTok.Span,
					err:    fmt.Errorf("expected ')', got %s", formatToken(p.source, finalTok)),
				})
			}
			return &CallExpr{
				Func: &Ident{
					Name:     tok.Value,
					NameSpan: tok.Span,
				},
				Lparen: nextTok.Span,
				Args:   args,
				Rparen: rparen,
			}, err
		}

		p.prev()
		return &Ident{
			NameSpan: tok.Span,
			Name:     tok.Value,
		}, nil
	case TokenQuotedIdentifier:
		p.prev()
		return p.ident()
	case TokenLParen:
		x, err := p.expr()
		endTok, ok := p.next()
		if !ok {
			err2 := &parseError{
				source: p.source,
				span:   indexSpan(len(p.source)),
				err:    errors.New("expected ')', got EOF"),
			}
			return &ParenExpr{
				Lparen: tok.Span,
				X:      x,
				Rparen: nullSpan(),
			}, joinErrors(err, err2)
		}
		if endTok.Kind != TokenRParen {
			err2 := &parseError{
				source: p.source,
				span:   endTok.Span,
				err:    fmt.Errorf("expected ')', got %s", formatToken(p.source, endTok)),
			}
			return &ParenExpr{
				Lparen: tok.Span,
				X:      x,
				Rparen: nullSpan(),
			}, joinErrors(err, err2)
		}
		return &ParenExpr{
			Lparen: tok.Span,
			X:      x,
			Rparen: endTok.Span,
		}, err
	default:
		p.prev()
		return nil, &parseError{
			source: p.source,
			span:   tok.Span,
			err:    notFoundError{fmt.Errorf("expected expression, got %s", formatToken(p.source, tok))},
		}
	}
}

func (p *parser) ident() (*Ident, error) {
	tok, _ := p.next()
	if tok.Kind != TokenIdentifier && tok.Kind != TokenQuotedIdentifier {
		p.prev()
		return nil, &parseError{
			source: p.source,
			span:   indexSpan(len(p.source)),
			err:    notFoundError{fmt.Errorf("expected identifier, got %s", formatToken(p.source, tok))},
		}
	}
	return &Ident{
		Name:     tok.Value,
		NameSpan: tok.Span,
	}, nil
}

func (p *parser) next() (Token, bool) {
	if p.pos >= len(p.tokens) {
		return Token{
			Kind:  TokenError,
			Span:  indexSpan(len(p.source)),
			Value: "EOF",
		}, false
	}
	tok := p.tokens[p.pos]
	p.pos++
	return tok, true
}

func (p *parser) prev() {
	p.pos--
}

func formatToken(source string, tok Token) string {
	if tok.Span.Start == len(source) && tok.Span.End == len(source) {
		return "EOF"
	}
	if tok.Span.Len() == 0 {
		if tok.Kind == TokenError {
			return "<scan error>"
		}
		return "''"
	}
	return "'" + spanString(source, tok.Span) + "'"
}

type parseError struct {
	source string
	span   Span
	err    error
}

func (e *parseError) Error() string {
	line, col := linecol(e.source, e.span.Start)
	return fmt.Sprintf("%d:%d: %s", line, col, e.err.Error())
}

func (e *parseError) Unwrap() error {
	return e.err
}

func linecol(source string, pos int) (line, col int) {
	line, col = 1, 1
	for _, c := range source[:pos] {
		switch c {
		case '\n':
			line++
			col = 1
		case '\t':
			const tabWidth = 8
			tabLoc := (col - 1) % tabWidth
			col += tabWidth - tabLoc
		default:
			col++
		}
	}
	return
}

func joinErrors(args ...error) error {
	var errorList []error
	for _, err := range args {
		if err == nil {
			continue
		}
		unwrapper, ok := err.(multiUnwrapper)
		if ok {
			errorList = append(errorList, unwrapper.Unwrap()...)
		} else {
			errorList = append(errorList, err)
		}
	}
	if len(errorList) == 0 {
		return nil
	}
	return errors.Join(errorList...)
}

// opaqueError is an error that does not unwrap its underlying error.
type opaqueError struct {
	error
}

func makeErrorOpaque(err error) error {
	switch e := err.(type) {
	case nil:
		return nil
	case *parseError:
		err2 := new(parseError)
		*err2 = *e
		err2.err = opaqueError{e.err}
		return err2
	case multiUnwrapper:
		errorList := slices.Clone(e.Unwrap())
		for i, err := range errorList {
			errorList[i] = opaqueError{err}
		}
		return errors.Join(errorList...)
	default:
		return opaqueError{err}
	}
}

// notFoundError is a sentinel for a production that did not parse anything.
type notFoundError struct {
	err error
}

func isNotFound(err error) bool {
	return errors.As(err, new(notFoundError))
}

func (e notFoundError) Error() string {
	return e.err.Error()
}

func (e notFoundError) Unwrap() error {
	return e.err
}

type multiUnwrapper interface {
	Unwrap() []error
}
