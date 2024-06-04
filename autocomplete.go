// Copyright 2024 RunReveal Inc.
// SPDX-License-Identifier: Apache-2.0

package pql

import (
	"cmp"
	"slices"
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
	Identifier string
}

func (ctx *AnalysisContext) SuggestCompletions(source string, cursor parser.Span) []*Completion {
	pos := cursor.End
	posSpan := parser.Span{Start: pos, End: pos}

	tokens := parser.Scan(source)
	expr, _ := parser.Parse(source)
	if expr == nil {
		prefix := completionPrefix(source, tokens, pos)
		result := make([]*Completion, 0, len(ctx.Tables))
		for tableName := range ctx.Tables {
			if strings.HasPrefix(tableName, prefix) {
				result = append(result, &Completion{
					Identifier: tableName,
				})
			}
		}
		return result
	}

	if sourceSpan := expr.Source.Span(); posSpan.Overlaps(sourceSpan) || pos < sourceSpan.Start {
		// Assume that this is a table name.
		prefix := completionPrefix(source, tokens, pos)
		result := make([]*Completion, 0, len(ctx.Tables))
		for tableName := range ctx.Tables {
			if strings.HasPrefix(tableName, prefix) {
				result = append(result, &Completion{
					Identifier: tableName,
				})
			}
		}
		return result
	}

	// TODO(now): More.

	return nil
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
