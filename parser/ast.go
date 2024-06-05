// Copyright 2024 RunReveal Inc.
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// Node is the interface implemented by all AST node types.
type Node interface {
	Span() Span
}

func nodeSpan(n Node) Span {
	if n == nil {
		return nullSpan()
	}
	return n.Span()
}

func nodeSliceSpan[T Node](nodes []T) Span {
	spans := make([]Span, 0, len(nodes))
	for _, n := range nodes {
		if span := nodeSpan(n); span.IsValid() {
			spans = append(spans, span)
		}
	}
	return unionSpans(spans...)
}

// An Ident node represents an identifier.
//
// Ident does not implement [Expr], but [QualifiedIdent] does.
// You can use [*Ident.AsQualified] to convert an *Ident to a *QualifiedIdent.
type Ident struct {
	Name     string
	NameSpan Span

	// Quoted is true if the identifier is quoted.
	Quoted bool
}

func (id *Ident) Span() Span {
	if id == nil {
		return nullSpan()
	}
	return id.NameSpan
}

// AsQualified converts the identifier to a [QualifiedIdent] with a single part.
func (id *Ident) AsQualified() *QualifiedIdent {
	if id == nil {
		return nil
	}
	return &QualifiedIdent{Parts: []*Ident{id}}
}

// A QualifiedIdent is one or more dot-separated identifiers.
type QualifiedIdent struct {
	Parts []*Ident
}

func (id *QualifiedIdent) Span() Span {
	if id == nil {
		return nullSpan()
	}
	return nodeSliceSpan(id.Parts)
}

func (id *QualifiedIdent) expression() {}

// TabularExpr is a query expression that produces a table.
type TabularExpr struct {
	Source    TabularDataSource
	Operators []TabularOperator
}

func (x *TabularExpr) Span() Span {
	if x == nil {
		return nullSpan()
	}
	return unionSpans(x.Source.Span(), nodeSliceSpan(x.Operators))
}

// TabularDataSource is the interface implemented by all AST node types
// that can be used as the data source of a [TabularExpr].
// At the moment, this can only be a [TableRef].
type TabularDataSource interface {
	Node
	tabularDataSource()
}

// A TableRef node refers to a specific table.
// It implements [TabularDataSource].
type TableRef struct {
	Table *Ident
}

func (ref *TableRef) tabularDataSource() {}

func (ref *TableRef) Span() Span {
	if ref == nil {
		return nullSpan()
	}
	return ref.Table.Span()
}

// TabularOperator is the interface implemented by all AST node types
// that can be used as operators in a [TabularExpr].
type TabularOperator interface {
	Node
	tabularOperator()
}

// CountOperator represents a `| count` operator in a [TabularExpr].
// It implements [TabularOperator].
type CountOperator struct {
	Pipe    Span
	Keyword Span
}

func (op *CountOperator) tabularOperator() {}

func (op *CountOperator) Span() Span {
	if op == nil {
		return nullSpan()
	}
	return unionSpans(op.Pipe, op.Keyword)
}

// WhereOperator represents a `| where` operator in a [TabularExpr].
// It implements [TabularOperator].
type WhereOperator struct {
	Pipe      Span
	Keyword   Span
	Predicate Expr
}

func (op *WhereOperator) tabularOperator() {}

func (op *WhereOperator) Span() Span {
	if op == nil {
		return nullSpan()
	}
	return unionSpans(op.Pipe, op.Keyword, nodeSpan(op.Predicate))
}

// SortOperator represents a `| sort by` operator in a [TabularExpr].
// It implements [TabularOperator].
type SortOperator struct {
	Pipe    Span
	Keyword Span
	Terms   []*SortTerm
}

func (op *SortOperator) tabularOperator() {}

func (op *SortOperator) Span() Span {
	if op == nil {
		return nullSpan()
	}
	return unionSpans(op.Pipe, op.Keyword, nodeSliceSpan(op.Terms))
}

// SortTerm is a single sort constraint in the [SortOperator].
type SortTerm struct {
	X           Expr
	Asc         bool
	AscDescSpan Span
	NullsFirst  bool
	NullsSpan   Span
}

func (term *SortTerm) Span() Span {
	if term == nil {
		return nullSpan()
	}
	return unionSpans(nodeSpan(term.X), term.AscDescSpan, term.NullsSpan)
}

// TakeOperator represents a `| take` operator in a [TabularExpr].
// It implements [TabularOperator].
type TakeOperator struct {
	Pipe     Span
	Keyword  Span
	RowCount Expr
}

func (op *TakeOperator) tabularOperator() {}

func (op *TakeOperator) Span() Span {
	if op == nil {
		return nullSpan()
	}
	return unionSpans(op.Pipe, op.Keyword, nodeSpan(op.RowCount))
}

// TopOperator represents a `| top` operator in a [TabularExpr].
// It implements [TabularOperator].
type TopOperator struct {
	Pipe     Span
	Keyword  Span
	RowCount Expr
	By       Span
	Col      *SortTerm
}

func (op *TopOperator) tabularOperator() {}

func (op *TopOperator) Span() Span {
	if op == nil {
		return nullSpan()
	}
	return unionSpans(op.Pipe, op.Keyword, nodeSpan(op.RowCount), op.By, nodeSpan(op.Col))
}

// ProjectOperator represents a `| project` operator in a [TabularExpr].
// It implements [TabularOperator].
type ProjectOperator struct {
	Pipe    Span
	Keyword Span
	Cols    []*ProjectColumn
}

func (op *ProjectOperator) tabularOperator() {}

func (op *ProjectOperator) Span() Span {
	if op == nil {
		return nullSpan()
	}
	return unionSpans(op.Pipe, op.Keyword, nodeSliceSpan(op.Cols))
}

// A ProjectColumn is a single column term in a [ProjectOperator].
// It consists of a column name,
// optionally followed by an expression specifying how to compute the column.
// If the expression is omitted, it is equivalent to using the Name as the expression.
type ProjectColumn struct {
	Name   *Ident
	Assign Span
	X      Expr
}

func (op *ProjectColumn) Span() Span {
	if op == nil {
		return nullSpan()
	}
	return unionSpans(op.Name.Span(), op.Assign, nodeSpan(op.X))
}

// ExtendOperator represents a `| extend` operator in a [TabularExpr].
// It implements [TabularOperator].
type ExtendOperator struct {
	Pipe    Span
	Keyword Span
	Cols    []*ExtendColumn
}

func (op *ExtendOperator) tabularOperator() {}

func (op *ExtendOperator) Span() Span {
	if op == nil {
		return nullSpan()
	}
	return unionSpans(op.Pipe, op.Keyword, nodeSliceSpan(op.Cols))
}

// A ExtendColumn is a single column term in a [ExtendOperator].
// It consists of an expression, optionally preceded by a column name.
// If the column name is omitted, one is derived from the expression.
type ExtendColumn struct {
	Name   *Ident
	Assign Span
	X      Expr
}

func (op *ExtendColumn) Span() Span {
	if op == nil {
		return nullSpan()
	}
	return unionSpans(op.Name.Span(), op.Assign, nodeSpan(op.X))
}

// SummarizeOperator represents a `| summarize` operator in a [TabularExpr].
// It implements [TabularOperator].
type SummarizeOperator struct {
	Pipe    Span
	Keyword Span
	Cols    []*SummarizeColumn
	By      Span
	GroupBy []*SummarizeColumn
}

func (op *SummarizeOperator) tabularOperator() {}

func (op *SummarizeOperator) Span() Span {
	if op == nil {
		return nullSpan()
	}
	return unionSpans(
		op.Pipe,
		op.Keyword,
		nodeSliceSpan(op.Cols),
		op.By,
		nodeSliceSpan(op.GroupBy),
	)
}

// A SummarizeColumn is a single column term in a [SummarizeOperator].
// It consists of an expression, optionally preceded by a column name.
// If the column name is omitted, one is derived from the expression.
type SummarizeColumn struct {
	Name   *Ident
	Assign Span
	X      Expr
}

func (op *SummarizeColumn) Span() Span {
	if op == nil {
		return nullSpan()
	}
	return unionSpans(op.Name.Span(), op.Assign, nodeSpan(op.X))
}

// JoinOperator represents a `| join` operator in a [TabularExpr].
// It implements [TabularOperator].
type JoinOperator struct {
	Pipe    Span
	Keyword Span

	Kind       Span
	KindAssign Span
	// Flavor is the type of join to use.
	// If absent, innerunique is implied.
	Flavor *Ident

	Lparen Span
	Right  *TabularExpr
	Rparen Span

	On Span
	// Conditions is one or more AND-ed conditions.
	// If the expression is a single identifier x,
	// then it is treated as equivalent to "$left.x == $right.x".
	Conditions []Expr
}

func (op *JoinOperator) tabularOperator() {}

func (op *JoinOperator) Span() Span {
	return unionSpans(
		op.Pipe,
		op.Keyword,
		op.Kind,
		op.KindAssign,
		op.Flavor.Span(),
		op.Lparen,
		op.Right.Span(),
		op.Rparen,
		op.On,
		nodeSliceSpan(op.Conditions),
	)
}

// AsOperator represents a `| as` operator in a [TabularExpr].
// It implements [TabularOperator].
type AsOperator struct {
	Pipe    Span
	Keyword Span
	Name    *Ident
}

func (op *AsOperator) tabularOperator() {}

func (op *AsOperator) Span() Span {
	return unionSpans(op.Pipe, op.Keyword, op.Name.Span())
}

// Expr is the interface implemented by all expression AST node types.
type Expr interface {
	Node
	expression()
}

// A BinaryExpr represents a binary expression.
type BinaryExpr struct {
	X      Expr
	OpSpan Span
	Op     TokenKind
	Y      Expr
}

func (expr *BinaryExpr) Span() Span {
	if expr == nil {
		return nullSpan()
	}
	return unionSpans(nodeSpan(expr.X), expr.OpSpan, nodeSpan(expr.Y))
}

func (expr *BinaryExpr) expression() {}

// A UnaryExpr represents a unary expression.
type UnaryExpr struct {
	OpSpan Span
	Op     TokenKind
	X      Expr
}

func (expr *UnaryExpr) Span() Span {
	if expr == nil {
		return nullSpan()
	}
	return unionSpans(expr.OpSpan, nodeSpan(expr.X))
}

func (expr *UnaryExpr) expression() {}

// An InExpr represents an "in" operator expression.
type InExpr struct {
	X      Expr
	In     Span
	Lparen Span
	Vals   []Expr
	Rparen Span
}

func (expr *InExpr) Span() Span {
	if expr == nil {
		return nullSpan()
	}
	return unionSpans(
		nodeSpan(expr.X),
		expr.In,
		expr.Lparen,
		nodeSliceSpan(expr.Vals),
		expr.Rparen,
	)
}

func (expr *InExpr) expression() {}

// A ParenExpr represents a parenthized expression.
type ParenExpr struct {
	Lparen Span
	X      Expr
	Rparen Span
}

func (expr *ParenExpr) Span() Span {
	if expr == nil {
		return nullSpan()
	}
	return unionSpans(expr.Lparen, nodeSpan(expr.X), expr.Rparen)
}

func (expr *ParenExpr) expression() {}

// A BasicLit node represents a numeric or string literal.
type BasicLit struct {
	ValueSpan Span
	Kind      TokenKind // [TokenNumber] or [TokenString]
	Value     string
}

func (lit *BasicLit) Span() Span {
	if lit == nil {
		return nullSpan()
	}
	return lit.ValueSpan
}

// IsFloat reports whether the literal is a floating point literal.
func (lit *BasicLit) IsFloat() bool {
	return lit.Kind == TokenNumber && strings.ContainsAny(lit.Value, ".eE")
}

// IsInteger reports whether the literal is a integer literal.
func (lit *BasicLit) IsInteger() bool {
	return lit.Kind == TokenNumber && !lit.IsFloat()
}

// Uint64 returns the numeric value of the literal as an unsigned integer.
// It returns 0 if the literal's kind is not [TokenNumber].
func (lit *BasicLit) Uint64() uint64 {
	if lit.Kind != TokenNumber {
		return 0
	}
	if lit.IsFloat() {
		return uint64(lit.Float64())
	}
	x, err := strconv.ParseUint(lit.Value, 10, 64)
	if err != nil {
		return 0
	}
	return x
}

// Float64 returns the numeric value of the literal as an unsigned integer.
// It returns 0 if the literal's kind is not [TokenNumber].
func (lit *BasicLit) Float64() float64 {
	if lit.Kind != TokenNumber {
		return 0
	}
	x, err := strconv.ParseFloat(lit.Value, 64)
	if err != nil {
		return 0
	}
	return x
}

func (lit *BasicLit) expression() {}

// A CallExpr node represents an unquoted identifier followed by an argument list.
type CallExpr struct {
	Func   *Ident
	Lparen Span
	Args   []Expr
	Rparen Span
}

func (call *CallExpr) Span() Span {
	if call == nil {
		return nullSpan()
	}
	return unionSpans(call.Func.Span(), call.Lparen, nodeSliceSpan(call.Args), call.Rparen)
}

func (call *CallExpr) expression() {}

// An IndexExpr node represents an array or map index.
type IndexExpr struct {
	X      Expr
	Lbrack Span
	Index  Expr
	Rbrack Span
}

func (idx *IndexExpr) Span() Span {
	if idx == nil {
		return nullSpan()
	}
	return unionSpans(nodeSpan(idx.X), idx.Lbrack, nodeSpan(idx.Index), idx.Rbrack)
}

func (idx *IndexExpr) expression() {}

// Walk traverses an AST in depth-first order.
// If the visit function returns true for a node,
// the visit function will be called for its children.
func Walk(n Node, visit func(n Node) bool) {
	stack := []Node{n}
	for len(stack) > 0 {
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		switch n := curr.(type) {
		case *Ident:
			visit(n)
		case *QualifiedIdent:
			if visit(n) {
				for i := len(n.Parts) - 1; i >= 0; i-- {
					stack = append(stack, n.Parts[i])
				}
			}
		case *TabularExpr:
			if visit(n) {
				for i := len(n.Operators) - 1; i >= 0; i-- {
					stack = append(stack, n.Operators[i])
				}
				stack = append(stack, n.Source)
			}
		case *TableRef:
			if visit(n) {
				stack = append(stack, n.Table)
			}
		case *CountOperator:
			visit(n)
		case *WhereOperator:
			if visit(n) {
				stack = append(stack, n.Predicate)
			}
		case *SortOperator:
			if visit(n) {
				for i := len(n.Terms) - 1; i >= 0; i-- {
					stack = append(stack, n.Terms[i])
				}
			}
		case *SortTerm:
			if visit(n) {
				stack = append(stack, n.X)
			}
		case *TakeOperator:
			if visit(n) {
				stack = append(stack, n.RowCount)
			}
		case *TopOperator:
			if visit(n) {
				stack = append(stack, n.Col)
				stack = append(stack, n.RowCount)
			}
		case *ProjectOperator:
			if visit(n) {
				for i := len(n.Cols) - 1; i >= 0; i-- {
					stack = append(stack, n.Cols[i])
				}
			}
		case *ProjectColumn:
			if visit(n) {
				if n.X != nil {
					stack = append(stack, n.X)
				}
				stack = append(stack, n.Name)
			}
		case *ExtendOperator:
			if visit(n) {
				for i := len(n.Cols) - 1; i >= 0; i-- {
					stack = append(stack, n.Cols[i])
				}
			}
		case *ExtendColumn:
			if visit(n) {
				if n.X != nil {
					stack = append(stack, n.X)
				}
				stack = append(stack, n.Name)
			}
		case *SummarizeOperator:
			if visit(n) {
				for i := len(n.GroupBy) - 1; i >= 0; i-- {
					stack = append(stack, n.GroupBy[i])
				}
				for i := len(n.Cols) - 1; i >= 0; i-- {
					stack = append(stack, n.Cols[i])
				}
			}
		case *SummarizeColumn:
			if visit(n) {
				stack = append(stack, n.X)
				if n.Name != nil {
					stack = append(stack, n.Name)
				}
			}
		case *JoinOperator:
			if visit(n) {
				// Skipping Flavor because it's more of a keyword on the operator than anything else.
				for i := len(n.Conditions) - 1; i >= 0; i-- {
					stack = append(stack, n.Conditions[i])
				}
				stack = append(stack, n.Right)
			}
		case *AsOperator:
			if visit(n) {
				stack = append(stack, n.Name)
			}
		case *BinaryExpr:
			if visit(n) {
				stack = append(stack, n.Y)
				stack = append(stack, n.X)
			}
		case *UnaryExpr:
			if visit(n) {
				stack = append(stack, n.X)
			}
		case *InExpr:
			if visit(n) {
				for i := len(n.Vals) - 1; i >= 0; i-- {
					stack = append(stack, n.Vals[i])
				}
				stack = append(stack, n.X)
			}
		case *BasicLit:
			visit(n)
		case *CallExpr:
			if visit(n) {
				// Skipping Func because it's flat.
				for i := len(n.Args) - 1; i >= 0; i-- {
					stack = append(stack, n.Args[i])
				}
			}
		case *IndexExpr:
			if visit(n) {
				stack = append(stack, n.Index)
				stack = append(stack, n.X)
			}
		default:
			panic(fmt.Errorf("unknown Node type %T", n))
		}
	}
}
