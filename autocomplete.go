// Copyright 2024 RunReveal Inc.
// SPDX-License-Identifier: Apache-2.0

package pql

import (
	"cmp"
	"slices"
	"strings"
	"unicode/utf8"

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
	// It is often the same as Text,
	// but Text may include extra characters for convenience.
	Label string
	// Text is the full string that is being placed.
	Text string
	// Span is the position where Text should be placed.
	// If the span's length is zero,
	// then the text should be inserted for a successful completion.
	// Otherwise, the span indicates text that should be replaced with the text.
	Span parser.Span
}

// SuggestCompletions suggests possible snippets to insert
// given a partial pql statement and a selected range.
func (ctx *AnalysisContext) SuggestCompletions(source string, cursor parser.Span) []*Completion {
	pos := cursor.End

	tokens := parser.Scan(source)
	stmts, _ := parser.Parse(source)
	prefix := completionPrefix(tokens, pos)
	i := spanBefore(stmts, pos, parser.Statement.Span)
	if i < 0 {
		return ctx.completeTableNames(source, prefix)
	}
	letNames := make(map[string]struct{})
	for _, stmt := range stmts[:i] {
		if stmt, ok := stmt.(*parser.LetStatement); ok && stmt.Name != nil {
			letNames[stmt.Name.Name] = struct{}{}
		}
	}

	// Try to figure out whether we're between a semicolon and another statement.
	nextTokenIndex, _ := slices.BinarySearchFunc(tokens, stmts[i].Span().End, func(tok parser.Token, i int) int {
		return cmp.Compare(tok.Span.Start, i)
	})
	if nextTokenIndex < len(tokens) && tokens[nextTokenIndex].Kind == parser.TokenSemi && pos >= tokens[nextTokenIndex].Span.End {
		return ctx.completeTableNames(source, prefix)
	}

	switch stmt := stmts[i].(type) {
	case *parser.LetStatement:
		return ctx.suggestLetStatement(source, stmt, letNames, prefix)
	case *parser.TabularExpr:
		return ctx.suggestTabularExpr(source, stmt, letNames, prefix)
	default:
		return nil
	}
}

func (ctx *AnalysisContext) suggestLetStatement(source string, stmt *parser.LetStatement, letNames map[string]struct{}, prefix parser.Span) []*Completion {
	if !stmt.Assign.IsValid() || prefix.End < stmt.Assign.End {
		return nil
	}
	return completeScope(source, prefix, letNames)
}

func (ctx *AnalysisContext) suggestTabularExpr(source string, expr *parser.TabularExpr, letNames map[string]struct{}, prefix parser.Span) []*Completion {
	if expr == nil {
		return ctx.completeTableNames(source, prefix)
	}

	if sourceSpan := expr.Source.Span(); prefix.Overlaps(sourceSpan) || prefix.End < sourceSpan.Start {
		// Assume that this is a table name.
		return ctx.completeTableNames(source, prefix)
	}

	// Find the operator that this cursor is associated with.
	i := spanBefore(expr.Operators, prefix.End, parser.TabularOperator.Span)
	if i < 0 {
		// Before the first operator.
		return completeOperators(source, prefix, true)
	}

	columns := ctx.determineColumnsInScope(expr.Source, expr.Operators[:i])

	switch op := expr.Operators[i].(type) {
	case *parser.UnknownTabularOperator:
		if prefix.End == op.Pipe.End {
			return completeOperators(source, prefix, false)
		}
		if name := op.Name(); name != nil && name.NameSpan.Overlaps(prefix) {
			return completeOperators(source, prefix, false)
		}
		if len(op.Tokens) == 0 || prefix.End < op.Tokens[0].Span.Start {
			return completeOperators(source, prefix, false)
		}
		return nil
	case *parser.WhereOperator:
		if prefix.End <= op.Keyword.Start {
			return completeOperators(source, prefix, false)
		}
		if prefix.End <= op.Keyword.End {
			return nil
		}
		return completeExpression(source, prefix, letNames, columns)
	case *parser.SortOperator:
		if prefix.End <= op.Keyword.Start {
			return completeOperators(source, prefix, false)
		}
		if prefix.End <= op.Keyword.End {
			return nil
		}
		return completeExpression(source, prefix, letNames, columns)
	case *parser.TakeOperator:
		if prefix.End <= op.Keyword.Start {
			return completeOperators(source, prefix, false)
		}
		if prefix.End <= op.Keyword.End {
			return nil
		}
		return completeScope(source, prefix, letNames)
	case *parser.TopOperator:
		if prefix.End <= op.Keyword.Start {
			return completeOperators(source, prefix, false)
		}
		if !op.By.IsValid() || prefix.End <= op.By.Start {
			return completeScope(source, prefix, letNames)
		}
		return completeExpression(source, prefix, letNames, columns)
	case *parser.ProjectOperator:
		if prefix.End <= op.Keyword.Start {
			return completeOperators(source, prefix, false)
		}
		if prefix.End <= op.Keyword.End {
			return nil
		}
		return completeExpression(source, prefix, letNames, columns)
	case *parser.ExtendOperator:
		if prefix.End <= op.Keyword.Start {
			return completeOperators(source, prefix, false)
		}
		if prefix.End <= op.Keyword.End {
			return nil
		}
		return completeExpression(source, prefix, letNames, columns)
	case *parser.SummarizeOperator:
		if prefix.End <= op.Keyword.Start {
			return completeOperators(source, prefix, false)
		}
		if prefix.End <= op.Keyword.End {
			return nil
		}
		return completeExpression(source, prefix, letNames, columns)
	case *parser.JoinOperator:
		if prefix.End <= op.Keyword.Start {
			return completeOperators(source, prefix, false)
		}
		if prefix.End <= op.Keyword.End {
			return nil
		}
		if op.Lparen.IsValid() && prefix.End >= op.Lparen.End && (!op.Rparen.IsValid() || prefix.End <= op.Rparen.Start) {
			return ctx.suggestTabularExpr(source, op.Right, letNames, prefix)
		}
		if op.On.IsValid() && prefix.End > op.On.End {
			return completeExpression(source, prefix, letNames, columns)
		}
		return nil
	case *parser.AsOperator:
		if prefix.End <= op.Keyword.Start {
			return completeOperators(source, prefix, false)
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

func completeExpression(source string, prefixSpan parser.Span, scope map[string]struct{}, columns []*AnalysisColumn) []*Completion {
	return append(completeColumnNames(source, prefixSpan, columns), completeScope(source, prefixSpan, scope)...)
}

func completeColumnNames(source string, prefixSpan parser.Span, columns []*AnalysisColumn) []*Completion {
	prefix := source[prefixSpan.Start:prefixSpan.End]

	result := make([]*Completion, 0, len(columns))
	for _, col := range columns {
		if hasFoldPrefix(col.Name, prefix) {
			result = append(result, &Completion{
				Label: col.Name,
				Text:  col.Name,
				Span:  prefixSpan,
			})
		}
	}
	return result
}

func completeScope(source string, prefixSpan parser.Span, scope map[string]struct{}) []*Completion {
	prefix := source[prefixSpan.Start:prefixSpan.End]

	result := make([]*Completion, 0, len(scope))
	for name := range scope {
		if hasFoldPrefix(name, prefix) {
			result = append(result, &Completion{
				Label: name,
				Text:  name,
				Span:  prefixSpan,
			})
		}
	}
	return result
}

func (ctx *AnalysisContext) completeTableNames(source string, prefixSpan parser.Span) []*Completion {
	prefix := source[prefixSpan.Start:prefixSpan.End]

	result := make([]*Completion, 0, len(ctx.Tables))
	for tableName := range ctx.Tables {
		if hasFoldPrefix(tableName, prefix) {
			result = append(result, &Completion{
				Label: tableName,
				Text:  tableName,
				Span:  prefixSpan,
			})
		}
	}
	slices.SortFunc(result, func(a, b *Completion) int {
		return cmp.Compare(a.Text, b.Text)
	})
	return result
}

func completeOperators(source string, prefixSpan parser.Span, includePipe bool) []*Completion {
	if includePipe {
		// Should always be an insert, not a replacement.
		prefixSpan = parser.Span{
			Start: prefixSpan.End,
			End:   prefixSpan.End,
		}
	}
	prefix := source[prefixSpan.Start:prefixSpan.End]
	leading := ""
	if includePipe {
		leading = "| "
	} else if prefixSpan.Len() == 0 && prefixSpan.Start > 0 && source[prefixSpan.Start-1] == '|' {
		// If directly adjacent to pipe, automatically add a space.
		leading = " "
	}

	result := make([]*Completion, 0, len(sortedOperatorNames))
	for _, name := range sortedOperatorNames {
		if !hasFoldPrefix(name, prefix) {
			continue
		}
		c := &Completion{
			Label: name,
			Text:  leading + name,
			Span:  prefixSpan,
		}
		if name == "order" || name == "sort" {
			c.Text += " by"
		}
		result = append(result, c)
	}
	return result
}

// completionPrefix returns a span of characters that should be considered for
// filtering completion results and replacement during the completion.
func completionPrefix(tokens []parser.Token, pos int) parser.Span {
	result := parser.Span{
		Start: pos,
		End:   pos,
	}
	if len(tokens) == 0 {
		return result
	}
	i := spanBefore(tokens, pos, func(tok parser.Token) parser.Span { return tok.Span })
	if i < 0 {
		return result
	}
	if !tokens[i].Span.Overlaps(parser.Span{Start: pos, End: pos}) || !isCompletableToken(tokens[i].Kind) {
		// Cursor is not adjacent to token. Assume there's whitespace.
		return result
	}
	result.Start = tokens[i].Span.Start
	if tokens[i].Kind == parser.TokenQuotedIdentifier {
		// Skip past initial backtick.
		result.Start += len("`")
	}
	return result
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

// hasFoldPrefix reports whether s starts with the given prefix,
// ignoring differences in case.
func hasFoldPrefix(s, prefix string) bool {
	n := utf8.RuneCountInString(prefix)

	// Find the end of the first n runes in s.
	// If s does not have that many runes,
	// then it can't have the prefix.
	if len(s) < n {
		return false
	}
	var prefixLen int
	for i := 0; i < n; i++ {
		_, sz := utf8.DecodeRuneInString(s[prefixLen:])
		if sz == 0 {
			return false
		}
		prefixLen += sz
	}

	return strings.EqualFold(s[:prefixLen], prefix)
}
