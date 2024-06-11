// Copyright 2024 RunReveal Inc.
// SPDX-License-Identifier: Apache-2.0

package pql

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/runreveal/pql/parser"
)

func TestSuggestCompletions(t *testing.T) {
	tests := []struct {
		name string

		context      *AnalysisContext
		sourceBefore string
		sourceAfter  string

		want []*Completion
	}{
		{
			name: "Empty",
			context: &AnalysisContext{
				Tables: map[string]*AnalysisTable{
					"foo": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
							{Name: "n"},
						},
					},
					"bar": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
						},
					},
				},
			},
			sourceBefore: "",
			sourceAfter:  "",
			want: []*Completion{
				{
					Label: "foo",
					Text:  "foo",
					Span: parser.Span{
						Start: 0,
						End:   0,
					},
				},
				{
					Label: "bar",
					Text:  "bar",
					Span: parser.Span{
						Start: 0,
						End:   0,
					},
				},
			},
		},
		{
			name: "InitialSourceRef",
			context: &AnalysisContext{
				Tables: map[string]*AnalysisTable{
					"foo": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
							{Name: "n"},
						},
					},
					"bar": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
						},
					},
				},
			},
			sourceBefore: "f",
			sourceAfter:  "",
			want: []*Completion{
				{
					Label: "foo",
					Text:  "foo",
					Span: parser.Span{
						Start: 0,
						End:   1,
					},
				},
			},
		},
		{
			name: "SourceRefWithPipe",
			context: &AnalysisContext{
				Tables: map[string]*AnalysisTable{
					"foo": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
							{Name: "n"},
						},
					},
					"bar": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
						},
					},
				},
			},
			sourceBefore: "",
			sourceAfter:  " | count",
			want: []*Completion{
				{
					Label: "foo",
					Text:  "foo",
					Span: parser.Span{
						Start: 0,
						End:   0,
					},
				},
				{
					Label: "bar",
					Text:  "bar",
					Span: parser.Span{
						Start: 0,
						End:   0,
					},
				},
			},
		},
		{
			name: "BeforeCompleteExpr",
			context: &AnalysisContext{
				Tables: map[string]*AnalysisTable{
					"foo": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
							{Name: "n"},
						},
					},
					"bar": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
						},
					},
				},
			},
			sourceBefore: "",
			sourceAfter:  "o | count",
			want: []*Completion{
				{
					Label: "foo",
					Text:  "foo",
					Span: parser.Span{
						Start: 0,
						End:   0,
					},
				},
				{
					Label: "bar",
					Text:  "bar",
					Span: parser.Span{
						Start: 0,
						End:   0,
					},
				},
			},
		},
		{
			name: "BeforeSpaceThenCompleteExpr",
			context: &AnalysisContext{
				Tables: map[string]*AnalysisTable{
					"foo": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
							{Name: "n"},
						},
					},
					"bar": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
						},
					},
				},
			},
			sourceBefore: "",
			sourceAfter:  " x | count",
			want: []*Completion{
				{
					Label: "foo",
					Text:  "foo",
					Span: parser.Span{
						Start: 0,
						End:   0,
					}},
				{
					Label: "bar",
					Text:  "bar",
					Span: parser.Span{
						Start: 0,
						End:   0,
					}},
			},
		},
		{
			name: "FirstOperator",
			context: &AnalysisContext{
				Tables: map[string]*AnalysisTable{
					"foo": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
							{Name: "n"},
						},
					},
					"bar": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
						},
					},
				},
			},
			sourceBefore: "foo ",
			sourceAfter:  "",
			want: []*Completion{
				{
					Label: "as",
					Text:  "| as",
					Span: parser.Span{
						Start: 4,
						End:   4,
					},
				},
				{
					Label: "count",
					Text:  "| count",
					Span: parser.Span{
						Start: 4,
						End:   4,
					},
				},
				{
					Label: "extend",
					Text:  "| extend",
					Span: parser.Span{
						Start: 4,
						End:   4,
					},
				},
				{
					Label: "join",
					Text:  "| join",
					Span: parser.Span{
						Start: 4,
						End:   4,
					},
				},
				{
					Label: "limit",
					Text:  "| limit",
					Span: parser.Span{
						Start: 4,
						End:   4,
					},
				},
				{
					Label: "order",
					Text:  "| order by",
					Span: parser.Span{
						Start: 4,
						End:   4,
					},
				},
				{
					Label: "project",
					Text:  "| project",
					Span: parser.Span{
						Start: 4,
						End:   4,
					},
				},
				{
					Label: "sort",
					Text:  "| sort by",
					Span: parser.Span{
						Start: 4,
						End:   4,
					},
				},
				{
					Label: "summarize",
					Text:  "| summarize",
					Span: parser.Span{
						Start: 4,
						End:   4,
					},
				},
				{
					Label: "take",
					Text:  "| take",
					Span: parser.Span{
						Start: 4,
						End:   4,
					},
				},
				{
					Label: "top",
					Text:  "| top",
					Span: parser.Span{
						Start: 4,
						End:   4,
					},
				},
				{
					Label: "where",
					Text:  "| where",
					Span: parser.Span{
						Start: 4,
						End:   4,
					},
				},
			},
		},
		{
			name: "FirstOperatorAfterPipe",
			context: &AnalysisContext{
				Tables: map[string]*AnalysisTable{
					"foo": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
							{Name: "n"},
						},
					},
					"bar": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
						},
					},
				},
			},
			sourceBefore: "foo |",
			sourceAfter:  "",
			want: []*Completion{
				{
					Label: "as",
					Text:  " as",
					Span: parser.Span{
						Start: 5,
						End:   5,
					},
				},
				{
					Label: "count",
					Text:  " count",
					Span: parser.Span{
						Start: 5,
						End:   5,
					},
				},
				{
					Label: "extend",
					Text:  " extend",
					Span: parser.Span{
						Start: 5,
						End:   5,
					},
				},
				{
					Label: "join",
					Text:  " join",
					Span: parser.Span{
						Start: 5,
						End:   5,
					},
				},
				{
					Label: "limit",
					Text:  " limit",
					Span: parser.Span{
						Start: 5,
						End:   5,
					},
				},
				{
					Label: "order",
					Text:  " order by",
					Span: parser.Span{
						Start: 5,
						End:   5,
					},
				},
				{
					Label: "project",
					Text:  " project",
					Span: parser.Span{
						Start: 5,
						End:   5,
					},
				},
				{
					Label: "sort",
					Text:  " sort by",
					Span: parser.Span{
						Start: 5,
						End:   5,
					},
				},
				{
					Label: "summarize",
					Text:  " summarize",
					Span: parser.Span{
						Start: 5,
						End:   5,
					},
				},
				{
					Label: "take",
					Text:  " take",
					Span: parser.Span{
						Start: 5,
						End:   5,
					},
				},
				{
					Label: "top",
					Text:  " top",
					Span: parser.Span{
						Start: 5,
						End:   5,
					},
				},
				{
					Label: "where",
					Text:  " where",
					Span: parser.Span{
						Start: 5,
						End:   5,
					},
				},
			},
		},
		{
			name: "FirstOperatorAfterPipeSpace",
			context: &AnalysisContext{
				Tables: map[string]*AnalysisTable{
					"foo": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
							{Name: "n"},
						},
					},
					"bar": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
						},
					},
				},
			},
			sourceBefore: "foo |  ",
			sourceAfter:  "",
			want: []*Completion{
				{
					Label: "as",
					Text:  "as",
					Span: parser.Span{
						Start: 7,
						End:   7,
					},
				},
				{
					Label: "count",
					Text:  "count",
					Span: parser.Span{
						Start: 7,
						End:   7,
					},
				},
				{
					Label: "extend",
					Text:  "extend",
					Span: parser.Span{
						Start: 7,
						End:   7,
					},
				},
				{
					Label: "join",
					Text:  "join",
					Span: parser.Span{
						Start: 7,
						End:   7,
					},
				},
				{
					Label: "limit",
					Text:  "limit",
					Span: parser.Span{
						Start: 7,
						End:   7,
					},
				},
				{
					Label: "order",
					Text:  "order by",
					Span: parser.Span{
						Start: 7,
						End:   7,
					},
				},
				{
					Label: "project",
					Text:  "project",
					Span: parser.Span{
						Start: 7,
						End:   7,
					},
				},
				{
					Label: "sort",
					Text:  "sort by",
					Span: parser.Span{
						Start: 7,
						End:   7,
					},
				},
				{
					Label: "summarize",
					Text:  "summarize",
					Span: parser.Span{
						Start: 7,
						End:   7,
					},
				},
				{
					Label: "take",
					Text:  "take",
					Span: parser.Span{
						Start: 7,
						End:   7,
					},
				},
				{
					Label: "top",
					Text:  "top",
					Span: parser.Span{
						Start: 7,
						End:   7,
					},
				},
				{
					Label: "where",
					Text:  "where",
					Span: parser.Span{
						Start: 7,
						End:   7,
					},
				},
			},
		},
		{
			name: "FirstOperatorPartial",
			context: &AnalysisContext{
				Tables: map[string]*AnalysisTable{
					"foo": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
							{Name: "n"},
						},
					},
					"bar": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
						},
					},
				},
			},
			sourceBefore: "foo | whe",
			want: []*Completion{
				{
					Label: "where",
					Text:  "where",
					Span: parser.Span{
						Start: 6,
						End:   9,
					},
				},
			},
		},
		{
			name: "WhereExpression",
			context: &AnalysisContext{
				Tables: map[string]*AnalysisTable{
					"foo": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
							{Name: "name"},
						},
					},
					"bar": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
						},
					},
				},
			},
			sourceBefore: "foo | where n",
			want: []*Completion{
				{
					Label: "name",
					Text:  "name",
					Span: parser.Span{
						Start: 12,
						End:   13,
					},
				},
			},
		},
		{
			name: "JoinExpression",
			context: &AnalysisContext{
				Tables: map[string]*AnalysisTable{
					"foo": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
							{Name: "name"},
						},
					},
					"bar": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
						},
					},
				},
			},
			sourceBefore: "foo | join (b",
			want: []*Completion{
				{
					Label: "bar",
					Text:  "bar",
					Span: parser.Span{
						Start: 12,
						End:   13,
					},
				},
			},
		},
		{
			name: "JoinExpressionOn",
			context: &AnalysisContext{
				Tables: map[string]*AnalysisTable{
					"foo": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
							{Name: "name"},
						},
					},
					"bar": {
						Columns: []*AnalysisColumn{
							{Name: "id"},
						},
					},
				},
			},
			sourceBefore: "foo | join (bar) on i",
			want: []*Completion{
				{
					Label: "id",
					Text:  "id",
					Span: parser.Span{
						Start: 20,
						End:   21,
					},
				},
			},
		},
		{
			name: "Project",
			context: &AnalysisContext{
				Tables: map[string]*AnalysisTable{
					"People": {
						Columns: []*AnalysisColumn{
							{Name: "FirstName"},
							{Name: "LastName"},
						},
					},
				},
			},
			sourceBefore: "People\n| project F",
			sourceAfter:  ", LastName",
			want: []*Completion{
				{
					Label: "FirstName",
					Text:  "FirstName",
					Span: parser.Span{
						Start: 17,
						End:   18,
					},
				},
			},
		},
		{
			name: "LetTake",
			context: &AnalysisContext{
				Tables: map[string]*AnalysisTable{
					"People": {
						Columns: []*AnalysisColumn{
							{Name: "FirstName"},
							{Name: "LastName"},
						},
					},
				},
			},
			sourceBefore: "let foo = 5;\nPeople\n| take ",
			want: []*Completion{
				{
					Label: "foo",
					Text:  "foo",
					Span: parser.Span{
						Start: 27,
						End:   27,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.context.SuggestCompletions(test.sourceBefore+test.sourceAfter, parser.Span{
				Start: len(test.sourceBefore),
				End:   len(test.sourceBefore),
			})
			completionLess := func(a, b *Completion) bool {
				if a.Span.Start != b.Span.Start {
					return a.Span.Start < b.Span.Start
				}
				if a.Span.End != b.Span.End {
					return a.Span.End < b.Span.End
				}
				if a.Label != b.Label {
					return a.Label < b.Label
				}
				return a.Text < b.Text
			}
			if diff := cmp.Diff(test.want, got, cmpopts.SortSlices(completionLess)); diff != "" {
				t.Errorf("SuggestCompletions(...) (-want +got):\n%s", diff)
			}
		})
	}
}
