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
						Name:     "StormEvents",
						NameSpan: newSpan(0, 11),
					},
				},
			},
		},
		{
			query: "StormEvents | count",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "StormEvents",
						NameSpan: newSpan(0, 11),
					},
				},
				Operators: []TabularOperator{
					&CountOperator{
						Pipe:    newSpan(12, 13),
						Keyword: newSpan(14, 19),
					},
				},
			},
		},
		{
			query: "StormEvents | count | count",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "StormEvents",
						NameSpan: newSpan(0, 11),
					},
				},
				Operators: []TabularOperator{
					&CountOperator{
						Pipe:    newSpan(12, 13),
						Keyword: newSpan(14, 19),
					},
					&CountOperator{
						Pipe:    newSpan(20, 21),
						Keyword: newSpan(22, 27),
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
