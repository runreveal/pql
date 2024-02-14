package parser

import (
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

func (id *Ident) expression() {}

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
