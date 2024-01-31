// Package parser provides a parser and an Abstract Syntax Tree (AST) for the Pipeline Query Language.
package parser

import (
	"errors"
	"fmt"
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
	} else if expr == nil && err == nil {
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
				err:    fmt.Errorf("expected operator name, got '%s'", spanString(p.source, operatorName.Span)),
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

func (p *parser) ident() (*Ident, error) {
	tok, ok := p.next()
	if !ok {
		return nil, &parseError{
			source: p.source,
			span:   indexSpan(len(p.source)),
			err:    notFoundError{errors.New("expected identifier, got EOF")},
		}
	}
	if tok.Kind != TokenIdentifier && tok.Kind != TokenQuotedIdentifier {
		p.prev()
		return nil, &parseError{
			source: p.source,
			span:   indexSpan(len(p.source)),
			err:    notFoundError{fmt.Errorf("expected identifier, got %q", spanString(p.source, tok.Span))},
		}
	}
	return &Ident{
		Name:      tok.Value,
		TokenSpan: tok.Span,
	}, nil
}

func (p *parser) next() (Token, bool) {
	if p.pos >= len(p.tokens) {
		return Token{}, false
	}
	tok := p.tokens[p.pos]
	p.pos++
	return tok, true
}

func (p *parser) prev() {
	p.pos--
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
		unwrapper, ok := err.(interface {
			Unwrap() []error
		})
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
