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
				{Label: "foo", Insert: "foo"},
				{Label: "bar", Insert: "bar"},
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
				{Label: "foo", Insert: "oo"},
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
				{Label: "foo", Insert: "foo"},
				{Label: "bar", Insert: "bar"},
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
				{Label: "foo", Insert: "foo"},
				{Label: "bar", Insert: "bar"},
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
				{Label: "foo", Insert: "foo"},
				{Label: "bar", Insert: "bar"},
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
					Label:  "as",
					Insert: "| as",
				},
				{
					Label:  "count",
					Insert: "| count",
				},
				{
					Label:  "extend",
					Insert: "| extend",
				},
				{
					Label:  "join",
					Insert: "| join",
				},
				{
					Label:  "limit",
					Insert: "| limit",
				},
				{
					Label:  "order",
					Insert: "| order by",
				},
				{
					Label:  "project",
					Insert: "| project",
				},
				{
					Label:  "sort",
					Insert: "| sort by",
				},
				{
					Label:  "summarize",
					Insert: "| summarize",
				},
				{
					Label:  "take",
					Insert: "| take",
				},
				{
					Label:  "top",
					Insert: "| top",
				},
				{
					Label:  "where",
					Insert: "| where",
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
					Label:  "as",
					Insert: " as",
				},
				{
					Label:  "count",
					Insert: " count",
				},
				{
					Label:  "extend",
					Insert: " extend",
				},
				{
					Label:  "join",
					Insert: " join",
				},
				{
					Label:  "limit",
					Insert: " limit",
				},
				{
					Label:  "order",
					Insert: " order by",
				},
				{
					Label:  "project",
					Insert: " project",
				},
				{
					Label:  "sort",
					Insert: " sort by",
				},
				{
					Label:  "summarize",
					Insert: " summarize",
				},
				{
					Label:  "take",
					Insert: " take",
				},
				{
					Label:  "top",
					Insert: " top",
				},
				{
					Label:  "where",
					Insert: " where",
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
					Label:  "as",
					Insert: "as",
				},
				{
					Label:  "count",
					Insert: "count",
				},
				{
					Label:  "extend",
					Insert: "extend",
				},
				{
					Label:  "join",
					Insert: "join",
				},
				{
					Label:  "limit",
					Insert: "limit",
				},
				{
					Label:  "order",
					Insert: "order by",
				},
				{
					Label:  "project",
					Insert: "project",
				},
				{
					Label:  "sort",
					Insert: "sort by",
				},
				{
					Label:  "summarize",
					Insert: "summarize",
				},
				{
					Label:  "take",
					Insert: "take",
				},
				{
					Label:  "top",
					Insert: "top",
				},
				{
					Label:  "where",
					Insert: "where",
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
					Label:  "where",
					Insert: "re",
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
					Label:  "name",
					Insert: "ame",
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
				if a.Label != b.Label {
					return a.Label < b.Label
				}
				return a.Insert < b.Insert
			}
			if diff := cmp.Diff(test.want, got, cmpopts.SortSlices(completionLess)); diff != "" {
				t.Errorf("SuggestCompletions(...) (-want +got):\n%s", diff)
			}
		})
	}
}
