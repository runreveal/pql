// Copyright 2024 RunReveal Inc.
// SPDX-License-Identifier: Apache-2.0

package pql

import (
	"cmp"
	"slices"
	"sort"
	"strings"

	"github.com/runreveal/pql/parser"
)

type AnalysisContext struct {
	Tables map[string]*AnalysisTable
}

type AnalysisTable struct {
	Columns []*AnalysisColumn
}

type AnalysisColumn struct {
	Name string
}

type Completion struct {
	Label  string
	Insert string
}

func (ctx *AnalysisContext) SuggestCompletions(source string, cursor parser.Span) []*Completion {
	pos := cursor.End
	posSpan := parser.Span{Start: pos, End: pos}

	tokens := parser.Scan(source)
	expr, _ := parser.Parse(source)
	if expr == nil {
		prefix := completionPrefix(source, tokens, pos)
		return ctx.completeTableNames(prefix)
	}

	if sourceSpan := expr.Source.Span(); posSpan.Overlaps(sourceSpan) || pos < sourceSpan.Start {
		// Assume that this is a table name.
		prefix := completionPrefix(source, tokens, pos)
		return ctx.completeTableNames(prefix)
	}

	// Find the operator that this cursor is associated with.
	i := sort.Search(len(expr.Operators), func(i int) bool {
		return expr.Operators[i].Span().Start >= pos
	})
	// Binary search will find the operator that follows the position.
	// Since the first character is a pipe,
	// we want to associate an exact match with the previous operator.
	i--
	if i < 0 {
		// Before the first operator.
		return completeOperators("")
	}

	switch op := expr.Operators[i].(type) {
	case *parser.UnknownTabularOperator:
		if pos <= op.Pipe.Start {
			return completeOperators("")
		}
		if pos == op.Pipe.End {
			return completeOperators("|")
		}
		if name := op.Name(); name != nil && name.NameSpan.Overlaps(posSpan) {
			return completeOperators("| " + completionPrefix(source, tokens, pos))
		}
		if len(op.Tokens) == 0 || pos < op.Tokens[0].Span.Start {
			return completeOperators("| ")
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
	i, _ := slices.BinarySearchFunc(tokens, pos, func(tok parser.Token, pos int) int {
		return cmp.Compare(tok.Span.Start, pos)
	})
	i = min(i, len(tokens)-1)
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

func isCompletableToken(kind parser.TokenKind) bool {
	return kind == parser.TokenIdentifier ||
		kind == parser.TokenQuotedIdentifier ||
		kind == parser.TokenAnd ||
		kind == parser.TokenOr ||
		kind == parser.TokenIn ||
		kind == parser.TokenBy
}
