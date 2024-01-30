package parser

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestParse(t *testing.T) {
	tests := []struct {
		query string
		want  *TabularExpr
	}{
		{
			query: "StormEvents",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:      "StormEvents",
						TokenSpan: Span{Start: 0, End: 11},
					},
				},
			},
		},
		{
			query: "StormEvents | count",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:      "StormEvents",
						TokenSpan: Span{Start: 0, End: 11},
					},
				},
				Operators: []TabularOperator{
					&CountOperator{
						Pipe:    Span{Start: 12, End: 13},
						Keyword: Span{Start: 14, End: 19},
					},
				},
			},
		},
		{
			query: "StormEvents | count | count",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:      "StormEvents",
						TokenSpan: Span{Start: 0, End: 11},
					},
				},
				Operators: []TabularOperator{
					&CountOperator{
						Pipe:    Span{Start: 12, End: 13},
						Keyword: Span{Start: 14, End: 19},
					},
					&CountOperator{
						Pipe:    Span{Start: 20, End: 21},
						Keyword: Span{Start: 22, End: 27},
					},
				},
			},
		},
	}
	for _, test := range tests {
		got, err := Parse(test.query)
		if err != nil {
			t.Errorf("Parse(%q): %v", test.query, err)
			continue
		}
		if diff := cmp.Diff(test.want, got, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("Parse(%q) (-want +got):\n%s", test.query, diff)
		}
	}
}
