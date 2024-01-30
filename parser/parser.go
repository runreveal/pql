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
			err = joinErrors(err, fmt.Errorf("%s: %s", linecol(query, trailingToken.Span.Start), trailingToken.Value))
		} else {
			err = joinErrors(err, fmt.Errorf("%s: unrecognized token", linecol(query, trailingToken.Span.Start)))
		}
	} else if expr == nil && err == nil {
		err = fmt.Errorf("%s: empty query", linecol(query, len(query)))
	}
	if err != nil {
		return expr, fmt.Errorf("parse pipeline query language: %w", err)
	}
	return expr, nil
}

func (p *parser) tabularExpr() (*TabularExpr, error) {
	tableName := p.ident()
	if tableName == nil {
		return nil, nil
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
			returnedError = joinErrors(returnedError, fmt.Errorf("%s: missing operator name after pipe", linecol(p.source, pipeToken.Span.End)))
			return expr, returnedError
		}
		if operatorName.Kind != TokenIdentifier {
			// TODO(soon): Skip ahead to next pipe.
			returnedError = joinErrors(returnedError, fmt.Errorf("%s: expected operator name, got '%s'", linecol(p.source, operatorName.Span.Start), spanString(p.source, operatorName.Span)))
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
			returnedError = joinErrors(returnedError, fmt.Errorf("%s: unknown operator name %q", linecol(p.source, operatorName.Span.Start), operatorName.Value))
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

func (p *parser) ident() *Ident {
	tok, ok := p.next()
	if !ok {
		return nil
	}
	if tok.Kind != TokenIdentifier && tok.Kind != TokenQuotedIdentifier {
		p.prev()
		return nil
	}
	return &Ident{
		Name:      tok.Value,
		TokenSpan: tok.Span,
	}
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

func linecol(source string, pos int) string {
	line, col := 1, 1
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
	return fmt.Sprintf("%d:%d", line, col)
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
