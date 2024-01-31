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
		{
			query: "StormEvents | where true",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "StormEvents",
						NameSpan: newSpan(0, 11),
					},
				},
				Operators: []TabularOperator{
					&WhereOperator{
						Pipe:    newSpan(12, 13),
						Keyword: newSpan(14, 19),
						Predicate: &Ident{
							Name:     "true",
							NameSpan: newSpan(20, 24),
						},
					},
				},
			},
		},
		{
			query: "StormEvents | where -42",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "StormEvents",
						NameSpan: newSpan(0, 11),
					},
				},
				Operators: []TabularOperator{
					&WhereOperator{
						Pipe:    newSpan(12, 13),
						Keyword: newSpan(14, 19),
						Predicate: &UnaryExpr{
							OpSpan: newSpan(20, 21),
							Op:     TokenMinus,
							X: &BasicLit{
								ValueSpan: newSpan(21, 23),
								Kind:      TokenNumber,
								Value:     "42",
							},
						},
					},
				},
			},
		},
		{
			query: `StormEvents | where rand()`,
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "StormEvents",
						NameSpan: newSpan(0, 11),
					},
				},
				Operators: []TabularOperator{
					&WhereOperator{
						Pipe:    newSpan(12, 13),
						Keyword: newSpan(14, 19),
						Predicate: &CallExpr{
							Func: &Ident{
								Name:     "rand",
								NameSpan: newSpan(20, 24),
							},
							Lparen: newSpan(24, 25),
							Rparen: newSpan(25, 26),
						},
					},
				},
			},
		},
		{
			query: "StormEvents | where not(false)",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "StormEvents",
						NameSpan: newSpan(0, 11),
					},
				},
				Operators: []TabularOperator{
					&WhereOperator{
						Pipe:    newSpan(12, 13),
						Keyword: newSpan(14, 19),
						Predicate: &CallExpr{
							Func: &Ident{
								Name:     "not",
								NameSpan: newSpan(20, 23),
							},
							Lparen: newSpan(23, 24),
							Args: []Expr{
								&Ident{
									Name:     "false",
									NameSpan: newSpan(24, 29),
								},
							},
							Rparen: newSpan(29, 30),
						},
					},
				},
			},
		},
		{
			query: `StormEvents | where strcat("abc", "def")`,
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "StormEvents",
						NameSpan: newSpan(0, 11),
					},
				},
				Operators: []TabularOperator{
					&WhereOperator{
						Pipe:    newSpan(12, 13),
						Keyword: newSpan(14, 19),
						Predicate: &CallExpr{
							Func: &Ident{
								Name:     "strcat",
								NameSpan: newSpan(20, 26),
							},
							Lparen: newSpan(26, 27),
							Args: []Expr{
								&BasicLit{
									Kind:      TokenString,
									Value:     "abc",
									ValueSpan: newSpan(27, 32),
								},
								&BasicLit{
									Kind:      TokenString,
									Value:     "def",
									ValueSpan: newSpan(34, 39),
								},
							},
							Rparen: newSpan(39, 40),
						},
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
