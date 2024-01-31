package parser

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
