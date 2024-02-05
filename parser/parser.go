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
	for i := 0; ; i++ {
		pipeToken, ok := p.next()
		if !ok {
			break
		}
		if pipeToken.Kind != TokenPipe {
			p.prev()
			if i == 0 {
				returnedError = joinErrors(returnedError, &parseError{
					source: p.source,
					span:   pipeToken.Span,
					err:    fmt.Errorf("expected '|' after table data source, got %s", formatToken(p.source, pipeToken)),
				})
			} else {
				returnedError = joinErrors(returnedError, &parseError{
					source: p.source,
					span:   pipeToken.Span,
					err:    fmt.Errorf("expected '|', got %s", formatToken(p.source, pipeToken)),
				})
			}
			p.skipTo(TokenPipe)
			continue
		}

		operatorName, ok := p.next()
		if !ok {
			returnedError = joinErrors(returnedError, &parseError{
				source: p.source,
				span:   pipeToken.Span,
				err:    errors.New("missing operator name after pipe"),
			})
			p.skipTo(TokenPipe)
			continue
		}
		if operatorName.Kind != TokenIdentifier {
			returnedError = joinErrors(returnedError, &parseError{
				source: p.source,
				span:   operatorName.Span,
				err:    fmt.Errorf("expected operator name, got %s", formatToken(p.source, operatorName)),
			})
			p.skipTo(TokenPipe)
			continue
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
		case "sort", "order":
			op, err := p.sortOperator(pipeToken, operatorName)
			if op != nil {
				expr.Operators = append(expr.Operators, op)
			}
			returnedError = joinErrors(returnedError, err)
		case "take", "limit":
			op, err := p.takeOperator(pipeToken, operatorName)
			if op != nil {
				expr.Operators = append(expr.Operators, op)
			}
			returnedError = joinErrors(returnedError, err)
		case "project":
			op, err := p.projectOperator(pipeToken, operatorName)
			if op != nil {
				expr.Operators = append(expr.Operators, op)
			}
			returnedError = joinErrors(returnedError, err)
		case "summarize":
			op, err := p.summarizeOperator(pipeToken, operatorName)
			if op != nil {
				expr.Operators = append(expr.Operators, op)
			}
			returnedError = joinErrors(returnedError, err)
		default:
			returnedError = joinErrors(returnedError, &parseError{
				source: p.source,
				span:   operatorName.Span,
				err:    fmt.Errorf("unknown operator name %q", operatorName.Value),
			})
			p.skipTo(TokenPipe)
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

func (p *parser) sortOperator(pipe, keyword Token) (*SortOperator, error) {
	by, _ := p.next()
	if by.Kind != TokenBy {
		op := &SortOperator{
			Pipe:    pipe.Span,
			Keyword: keyword.Span,
		}
		err := &parseError{
			source: p.source,
			span:   by.Span,
			err:    fmt.Errorf("expected 'by', got %s", formatToken(p.source, by)),
		}
		return op, err
	}

	op := &SortOperator{
		Pipe:    pipe.Span,
		Keyword: newSpan(keyword.Span.Start, by.Span.End),
	}
	for {
		x, err := p.expr()
		if err != nil {
			return op, makeErrorOpaque(err)
		}
		term := &SortTerm{
			X:           x,
			AscDescSpan: nullSpan(),
			NullsSpan:   nullSpan(),
		}
		op.Terms = append(op.Terms, term)

		// asc/desc
		tok, ok := p.next()
		if !ok {
			return op, nil
		}
		switch tok.Kind {
		case TokenComma:
			continue
		case TokenIdentifier:
			switch tok.Value {
			case "asc":
				term.Asc = true
				term.AscDescSpan = tok.Span
				term.NullsFirst = true
			case "desc":
				term.Asc = false
				term.AscDescSpan = tok.Span
				term.NullsFirst = false
			case "nulls":
				// Good, but wait until next switch statement.
				p.prev()
			default:
				p.prev()
				return op, nil
			}
		default:
			p.prev()
			return op, nil
		}

		// nulls first/last
		tok, ok = p.next()
		if !ok {
			return op, nil
		}
		switch {
		case tok.Kind == TokenComma:
			continue
		case tok.Kind == TokenIdentifier && tok.Value == "nulls":
			switch tok2, _ := p.next(); {
			case tok2.Kind == TokenIdentifier && tok2.Value == "first":
				term.NullsFirst = true
				term.NullsSpan = newSpan(tok.Span.Start, tok2.Span.End)
			case tok2.Kind == TokenIdentifier && tok2.Value == "last":
				term.NullsFirst = false
				term.NullsSpan = newSpan(tok.Span.Start, tok2.Span.End)
			default:
				p.prev()
				return op, &parseError{
					source: p.source,
					span:   tok2.Span,
					err:    fmt.Errorf("expected 'first' or 'last', got %s", formatToken(p.source, tok2)),
				}
			}
		default:
			p.prev()
			return op, nil
		}

		// Check for a comma to see if we should proceed.
		tok, ok = p.next()
		if !ok {
			return op, nil
		}
		if tok.Kind != TokenComma {
			p.prev()
			return op, nil
		}
	}
}

func (p *parser) takeOperator(pipe, keyword Token) (*TakeOperator, error) {
	op := &TakeOperator{
		Pipe:    pipe.Span,
		Keyword: keyword.Span,
	}

	tok, _ := p.next()
	if tok.Kind != TokenNumber {
		return op, &parseError{
			source: p.source,
			span:   tok.Span,
			err:    fmt.Errorf("expected integer, got %s", formatToken(p.source, tok)),
		}
	}
	rowCount := &BasicLit{
		Kind:      tok.Kind,
		Value:     tok.Value,
		ValueSpan: tok.Span,
	}
	op.RowCount = rowCount
	if !rowCount.IsInteger() {
		return op, &parseError{
			source: p.source,
			span:   tok.Span,
			err:    fmt.Errorf("expected integer, got %s", formatToken(p.source, tok)),
		}
	}
	return op, nil
}

func (p *parser) projectOperator(pipe, keyword Token) (*ProjectOperator, error) {
	op := &ProjectOperator{
		Pipe:    pipe.Span,
		Keyword: keyword.Span,
	}

	for {
		colName, err := p.ident()
		if err != nil {
			return op, makeErrorOpaque(err)
		}
		col := &ProjectColumn{
			Name:   colName,
			Assign: nullSpan(),
		}
		op.Cols = append(op.Cols, col)

		sep, ok := p.next()
		if !ok {
			return op, nil
		}
		switch sep.Kind {
		case TokenComma:
			continue
		case TokenAssign:
			col.Assign = sep.Span
			col.X, err = p.expr()
			if err != nil {
				return op, makeErrorOpaque(err)
			}
			sep, ok = p.next()
			if !ok || sep.Kind != TokenComma {
				return op, nil
			}
		default:
			p.prev()
			return op, nil
		}
	}
}

func (p *parser) summarizeOperator(pipe, keyword Token) (*SummarizeOperator, error) {
	op := &SummarizeOperator{
		Pipe:    pipe.Span,
		Keyword: keyword.Span,
		By:      nullSpan(),
	}

	for {
		col, err := p.summarizeColumn()
		if isNotFound(err) {
			break
		}
		if col != nil {
			op.Cols = append(op.Cols, col)
		}
		if err != nil {
			return op, makeErrorOpaque(err)
		}

		sep, ok := p.next()
		if !ok {
			return op, nil
		}
		if sep.Kind != TokenComma {
			p.prev()
			break
		}
	}

	sep, ok := p.next()
	if !ok {
		if len(op.Cols) == 0 {
			return op, &parseError{
				source: p.source,
				span:   sep.Span,
				err:    fmt.Errorf("expected expression or 'by', got EOF"),
			}
		}
		return op, nil
	}
	if sep.Kind != TokenBy {
		p.prev()
		if len(op.Cols) == 0 {
			return op, &parseError{
				source: p.source,
				span:   sep.Span,
				err:    fmt.Errorf("expected expression or 'by', got %s", formatToken(p.source, sep)),
			}
		}
		return op, nil
	}
	op.By = sep.Span
	for {
		col, err := p.summarizeColumn()
		if isNotFound(err) {
			return op, makeErrorOpaque(err)
		}
		if col != nil {
			op.GroupBy = append(op.GroupBy, col)
		}
		if err != nil {
			return op, makeErrorOpaque(err)
		}

		sep, ok := p.next()
		if !ok {
			return op, nil
		}
		if sep.Kind != TokenComma {
			p.prev()
			return op, nil
		}
	}
}

func (p *parser) summarizeColumn() (*SummarizeColumn, error) {
	restorePos := p.pos

	col := &SummarizeColumn{
		Assign: nullSpan(),
	}

	var err error
	col.Name, err = p.ident()
	if err == nil {
		if assign, _ := p.next(); assign.Kind == TokenAssign {
			col.Assign = assign.Span
		} else {
			col.Name = nil
			p.pos = restorePos
		}
	} else if !isNotFound(err) {
		col.X = col.Name
		col.Name = nil
		return col, makeErrorOpaque(err)
	}

	col.X, err = p.expr()
	if col.Name != nil {
		err = makeErrorOpaque(err)
	}
	return col, err
}

// exprList parses one or more comma-separated expressions.
func (p *parser) exprList() ([]Expr, error) {
	first, err := p.expr()
	if err != nil {
		return nil, err
	}
	result := []Expr{first}
	for {
		restorePos := p.pos
		tok, ok := p.next()
		if !ok {
			return result, nil
		}
		if tok.Kind != TokenComma {
			p.prev()
			return result, nil
		}
		x, err := p.expr()
		if isNotFound(err) {
			p.pos = restorePos
			return result, nil
		}
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
	x, err1 := p.unaryExpr()
	if isNotFound(err1) {
		return x, err1
	}
	x, err2 := p.exprBinaryTrail(x, 0)
	return x, joinErrors(err1, err2)
}

// exprBinaryTrail parses zero or more (binaryOp, unaryExpr) sequences.
func (p *parser) exprBinaryTrail(x Expr, minPrecedence int) (Expr, error) {
	var finalError error
	for {
		op1, ok := p.next()
		if !ok {
			return x, finalError
		}
		precedence1 := operatorPrecedence(op1.Kind)
		if precedence1 < 0 || precedence1 < minPrecedence {
			// Not a binary operator or below precedence threshold.
			p.prev()
			return x, finalError
		}
		y, err := p.unaryExpr()
		if err != nil {
			finalError = joinErrors(finalError, makeErrorOpaque(err))
		}

		// Resolve any higher precedence operators first.
		for {
			op2, ok := p.next()
			if !ok {
				break
			}
			p.prev()

			precedence2 := operatorPrecedence(op2.Kind)
			if precedence2 < 0 || precedence2 <= precedence1 {
				// Not a binary operator or below the precedence of the original operator.
				break
			}
			y, err = p.exprBinaryTrail(y, precedence1+1)
			if err != nil {
				finalError = joinErrors(finalError, makeErrorOpaque(err))
			}
		}

		x = &BinaryExpr{
			X:      x,
			OpSpan: op1.Span,
			Op:     op1.Kind,
			Y:      y,
		}
	}
}

func operatorPrecedence(op TokenKind) int {
	switch op {
	case TokenStar, TokenSlash, TokenMod:
		return 4
	case TokenPlus, TokenMinus:
		return 3
	case TokenEq, TokenNE, TokenLT, TokenLE, TokenGT, TokenGE,
		TokenCaseInsensitiveEq, TokenCaseInsensitiveNE:
		return 2
	case TokenAnd:
		return 1
	case TokenOr:
		return 0
	default:
		return -1
	}
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
				p.skipTo(TokenRParen)
				finalTok, _ = p.next()
				if finalTok.Kind == TokenRParen {
					rparen = finalTok.Span
				} else {
					p.prev()
				}
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
		err = makeErrorOpaque(err) // already consumed a parenthesis
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
			err = joinErrors(err, &parseError{
				source: p.source,
				span:   endTok.Span,
				err:    fmt.Errorf("expected ')', got %s", formatToken(p.source, endTok)),
			})
			p.skipTo(TokenRParen)
			endTok, _ = p.next()
			if endTok.Kind != TokenRParen {
				return &ParenExpr{
					Lparen: tok.Span,
					X:      x,
					Rparen: nullSpan(),
				}, err
			}
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

// skipTo advances the parser to right before the next token of the given kind.
// It ignores tokens that are in parenthetical groups after the initial parse position.
// If no such token is found, skipTo advances to EOF.
func (p *parser) skipTo(search TokenKind) {
	parenLevel := 0
	for {
		tok, ok := p.next()
		if !ok {
			return
		}

		switch tok.Kind {
		case TokenLParen:
			if search == TokenLParen {
				p.prev()
				return
			}
			parenLevel++
		case TokenRParen:
			if parenLevel > 0 {
				parenLevel--
			} else if search == TokenRParen {
				p.prev()
				return
			}
		case search:
			if parenLevel <= 0 {
				p.prev()
				return
			}
		}
	}
}

func (p *parser) next() (Token, bool) {
	if p.pos >= len(p.tokens) {
		p.pos = len(p.tokens) + 1 // Once we produce EOF, don't permit rewinding.
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
	// Only allow rewinding to a previous token if it's in valid range,
	// and once we produce EOF, we will always produce EOF.
	if p.pos > 0 && p.pos <= len(p.tokens) {
		p.pos--
	}
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
