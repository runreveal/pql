package parser

import (
	"strconv"
	"strings"
)

// Node is the interface implemented by all AST node types.
type Node interface {
	Span() Span
}

// An Ident node represents an identifier.
type Ident struct {
	Name     string
	NameSpan Span
}

func (id *Ident) Span() Span {
	return id.NameSpan
}

func (id *Ident) expression() {}

// TabularExpr is a query expression that produces a table.
type TabularExpr struct {
	Source    TabularDataSource
	Operators []TabularOperator
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
	return newSpan(op.Pipe.Start, op.Keyword.End)
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
	return newSpan(op.Pipe.Start, op.Predicate.Span().End)
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
	if len(op.Terms) == 0 {
		// Not technically valid, but want to avoid a panic.
		return newSpan(op.Pipe.Start, op.Keyword.End)
	}
	return newSpan(op.Pipe.Start, op.Terms[len(op.Terms)-1].Span().End)
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
	span := term.X.Span()
	if term.NullsSpan.IsValid() {
		span.End = term.NullsSpan.End
	} else if term.AscDescSpan.IsValid() {
		span.End = term.AscDescSpan.End
	}
	return span
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
	return newSpan(op.Pipe.Start, op.RowCount.Span().End)
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
	return Span{Start: expr.X.Span().Start, End: expr.Y.Span().End}
}

func (expr *BinaryExpr) expression() {}

// A UnaryExpr represents a unary expression.
type UnaryExpr struct {
	OpSpan Span
	Op     TokenKind
	X      Expr
}

func (expr *UnaryExpr) Span() Span {
	return Span{Start: expr.OpSpan.Start, End: expr.X.Span().End}
}

func (expr *UnaryExpr) expression() {}

// A ParenExpr represents a parenthized expression.
type ParenExpr struct {
	Lparen Span
	X      Expr
	Rparen Span
}

func (expr *ParenExpr) Span() Span {
	return Span{Start: expr.Lparen.Start, End: expr.Rparen.End}
}

func (expr *ParenExpr) expression() {}

// A BasicLit node represents a numeric or string literal.
type BasicLit struct {
	ValueSpan Span
	Kind      TokenKind // [TokenNumber] or [TokenString]
	Value     string
}

func (lit *BasicLit) Span() Span {
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
	return Span{Start: call.Func.NameSpan.Start, End: call.Rparen.End}
}

func (call *CallExpr) expression() {}
