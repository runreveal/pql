package parser

type Node interface {
	Span() Span
}

// An Ident node represents an identifier.
type Ident struct {
	Name      string
	TokenSpan Span
}

func (id *Ident) Span() Span {
	return id.TokenSpan
}

type TabularExpr struct {
	Source    TabularDataSource
	Operators []TabularOperator
}

type TabularDataSource interface {
	Node
	tabularDataSource()
}

type TableRef struct {
	Table *Ident
}

func (ref *TableRef) tabularDataSource() {}

func (ref *TableRef) Span() Span {
	return ref.Table.Span()
}

type TabularOperator interface {
	Node
	tabularOperator()
}

type CountOperator struct {
	Pipe    Span
	Keyword Span
}

func (op *CountOperator) tabularOperator() {}

func (op *CountOperator) Span() Span {
	return Span{
		Start: op.Pipe.Start,
		End:   op.Keyword.End,
	}
}
