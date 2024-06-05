// Copyright 2024 RunReveal Inc.
// SPDX-License-Identifier: Apache-2.0

package pql

import (
	"cmp"
	"slices"
	"strings"

	"github.com/runreveal/pql/parser"
)

// AnalysisContext is information about the eventual execution environment
// passed in to assist in analysis tasks.
type AnalysisContext struct {
	Tables map[string]*AnalysisTable
}

// AnalysisTable is a table known to an [AnalysisContext].
type AnalysisTable struct {
	Columns []*AnalysisColumn
}

// AnalysisColumn is a column known to an [AnalysisTable].
type AnalysisColumn struct {
	Name string
}

// Completion is a single completion suggestion
// returned by [AnalysisContext.SuggestCompletions].
type Completion struct {
	// Label is the label that should be displayed for the completion.
	// It represents the full string that is being completed.
	Label string
	// Insert is the text that should be inserted after the cursor
	// to perform the completion.
	Insert string
}

// SuggestCompletions suggests possible snippets to insert
// given a partial pql statement and a selected range.
func (ctx *AnalysisContext) SuggestCompletions(source string, cursor parser.Span) []*Completion {
	pos := cursor.End

	tokens := parser.Scan(source)
	expr, _ := parser.Parse(source)
	prefix := completionPrefix(source, tokens, pos)
	return ctx.suggestCompletions(expr, prefix, cursor.End)
}

func (ctx *AnalysisContext) suggestCompletions(expr *parser.TabularExpr, prefix string, pos int) []*Completion {
	posSpan := parser.Span{Start: pos, End: pos}
	if expr == nil {
		return ctx.completeTableNames(prefix)
	}

	if sourceSpan := expr.Source.Span(); posSpan.Overlaps(sourceSpan) || pos < sourceSpan.Start {
		// Assume that this is a table name.
		return ctx.completeTableNames(prefix)
	}

	// Find the operator that this cursor is associated with.
	i := spanBefore(expr.Operators, pos, parser.TabularOperator.Span)
	if i < 0 {
		// Before the first operator.
		return completeOperators("")
	}

	columns := ctx.determineColumnsInScope(expr.Source, expr.Operators[:i])

	switch op := expr.Operators[i].(type) {
	case *parser.UnknownTabularOperator:
		if pos == op.Pipe.End {
			return completeOperators("|")
		}
		if name := op.Name(); name != nil && name.NameSpan.Overlaps(posSpan) {
			return completeOperators("| " + prefix)
		}
		if len(op.Tokens) == 0 || pos < op.Tokens[0].Span.Start {
			return completeOperators("| ")
		}
		return nil
	case *parser.WhereOperator:
		if pos <= op.Keyword.Start {
			return completeOperators("|")
		}
		if pos <= op.Keyword.End {
			return nil
		}
		return completeColumnNames(prefix, columns)
	case *parser.SortOperator:
		if pos <= op.Keyword.Start {
			return completeOperators("|")
		}
		if pos <= op.Keyword.End {
			return nil
		}
		return completeColumnNames(prefix, columns)
	case *parser.TopOperator:
		if pos <= op.Keyword.Start {
			return completeOperators("|")
		}
		if !op.By.IsValid() || pos <= op.By.End {
			return nil
		}
		return completeColumnNames(prefix, columns)
	case *parser.ProjectOperator:
		if pos <= op.Keyword.Start {
			return completeOperators("|")
		}
		if pos <= op.Keyword.End {
			return nil
		}
		return completeColumnNames(prefix, columns)
	case *parser.ExtendOperator:
		if pos <= op.Keyword.Start {
			return completeOperators("|")
		}
		if pos <= op.Keyword.End {
			return nil
		}
		return completeColumnNames(prefix, columns)
	case *parser.SummarizeOperator:
		if pos <= op.Keyword.Start {
			return completeOperators("|")
		}
		if pos <= op.Keyword.End {
			return nil
		}
		return completeColumnNames(prefix, columns)
	case *parser.JoinOperator:
		if pos <= op.Keyword.Start {
			return completeOperators("|")
		}
		if pos <= op.Keyword.End {
			return nil
		}
		if op.Lparen.IsValid() && pos >= op.Lparen.End && (!op.Rparen.IsValid() || pos <= op.Rparen.Start) {
			return ctx.suggestCompletions(op.Right, prefix, pos)
		}
		if op.On.IsValid() && pos > op.On.End {
			return completeColumnNames(prefix, columns)
		}
		return nil
	default:
		return nil
	}
}

var sortedOperatorNames = []string{
	"as",
	"count",
	"extend",
	"join",
	"limit",
	"order",
	"project",
	"sort",
	"summarize",
	"take",
	"top",
	"where",
}

func (ctx *AnalysisContext) determineColumnsInScope(source parser.TabularDataSource, ops []parser.TabularOperator) []*AnalysisColumn {
	var columns []*AnalysisColumn
	if source, ok := source.(*parser.TableRef); ok {
		if tab := ctx.Tables[source.Table.Name]; tab != nil {
			columns = tab.Columns
		}
	}
	for _, op := range ops {
		switch op := op.(type) {
		case *parser.CountOperator:
			columns = []*AnalysisColumn{{Name: "count()"}}
		case *parser.ProjectOperator:
			columns = make([]*AnalysisColumn, 0, len(op.Cols))
			for _, col := range op.Cols {
				columns = append(columns, &AnalysisColumn{
					Name: col.Name.Name,
				})
			}
		case *parser.ExtendOperator:
			columns = slices.Clip(columns)
			for _, col := range op.Cols {
				columns = append(columns, &AnalysisColumn{
					Name: col.Name.Name,
				})
			}
		case *parser.SummarizeOperator:
			columns = make([]*AnalysisColumn, 0, len(op.Cols)+len(op.GroupBy))
			for _, col := range op.Cols {
				columns = append(columns, &AnalysisColumn{
					Name: col.Name.Name,
				})
			}
			for _, col := range op.GroupBy {
				columns = append(columns, &AnalysisColumn{
					Name: col.Name.Name,
				})
			}
		case *parser.JoinOperator:
			columns = slices.Clip(columns)
			columns = append(columns, ctx.determineColumnsInScope(op.Right.Source, op.Right.Operators)...)
		}
	}
	return columns
}

func completeColumnNames(prefix string, columns []*AnalysisColumn) []*Completion {
	result := make([]*Completion, 0, len(columns))
	for _, col := range columns {
		if strings.HasPrefix(col.Name, prefix) {
			result = append(result, &Completion{
				Label:  col.Name,
				Insert: col.Name[len(prefix):],
			})
		}
	}
	return result
}

func (ctx *AnalysisContext) completeTableNames(prefix string) []*Completion {
	result := make([]*Completion, 0, len(ctx.Tables))
	for tableName := range ctx.Tables {
		if strings.HasPrefix(tableName, prefix) {
			result = append(result, &Completion{
				Label:  tableName,
				Insert: tableName[len(prefix):],
			})
		}
	}
	slices.SortFunc(result, func(a, b *Completion) int {
		return cmp.Compare(a.Label, b.Label)
	})
	return result
}

func completeOperators(prefix string) []*Completion {
	result := make([]*Completion, 0, len(sortedOperatorNames))
	var namePrefix string
	if rest, ok := strings.CutPrefix(prefix, "|"); ok {
		if rest, ok = strings.CutPrefix(rest, " "); ok {
			namePrefix = rest
		}
	}

	for _, name := range sortedOperatorNames {
		if !strings.HasPrefix(name, namePrefix) {
			continue
		}
		c := &Completion{
			Label:  name,
			Insert: ("| " + name)[len(prefix):],
		}
		if name == "order" || name == "sort" {
			c.Insert += " by"
		}
		result = append(result, c)
	}
	return result
}

func completionPrefix(source string, tokens []parser.Token, pos int) string {
	if len(tokens) == 0 {
		return ""
	}
	i := spanBefore(tokens, pos, func(tok parser.Token) parser.Span {
		return tok.Span
	})
	if i < 0 {
		return ""
	}
	if !tokens[i].Span.Overlaps(parser.Span{Start: pos, End: pos}) || !isCompletableToken(tokens[i].Kind) {
		// Cursor is not adjacent to token. Assume there's whitespace.
		return ""
	}
	start := tokens[i].Span.Start
	if tokens[i].Kind == parser.TokenQuotedIdentifier {
		// Skip past initial backtick.
		start += len("`")
	}
	return source[start:pos]
}

// spanBefore finds the first span in a sorted slice
// that starts before the given position.
// The span function is used to obtain the span of each element in the slice.
// If the position occurs before any spans,
// spanBefore returns -1.
func spanBefore[S ~[]E, E any](x S, pos int, span func(E) parser.Span) int {
	i, _ := slices.BinarySearchFunc(x, pos, func(elem E, pos int) int {
		return cmp.Compare(span(elem).Start, pos)
	})
	// Binary search will find the span that follows the position.
	return i - 1
}

func isCompletableToken(kind parser.TokenKind) bool {
	return kind == parser.TokenIdentifier ||
		kind == parser.TokenQuotedIdentifier ||
		kind == parser.TokenAnd ||
		kind == parser.TokenOr ||
		kind == parser.TokenIn ||
		kind == parser.TokenBy
}
