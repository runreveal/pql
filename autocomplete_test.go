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
				{Identifier: "foo"},
				{Identifier: "bar"},
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
				{Identifier: "foo"},
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
				{Identifier: "foo"},
				{Identifier: "bar"},
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
				{Identifier: "foo"},
				{Identifier: "bar"},
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
				{Identifier: "foo"},
				{Identifier: "bar"},
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
				return a.Identifier < b.Identifier
			}
			if diff := cmp.Diff(test.want, got, cmpopts.SortSlices(completionLess)); diff != "" {
				t.Errorf("SuggestCompletions(...) (-want +got):\n%s", diff)
			}
		})
	}
}
