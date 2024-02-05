package parser

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  *TabularExpr
		err   bool
	}{
		{
			name:  "Empty",
			query: "",
			want:  nil,
			err:   true,
		},
		{
			name:  "OnlyTableName",
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
			name:  "PipeCount",
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
			name:  "DoublePipeCount",
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
			name:  "WhereTrue",
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
			name:  "NegativeNumber",
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
			name:  "ZeroArgFunction",
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
			name:  "OneArgFunction",
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
			name:  "TwoArgFunction",
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
		{
			name:  "TwoArgFunctionWithTrailingComma",
			query: `StormEvents | where strcat("abc", "def",)`,
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
							Rparen: newSpan(40, 41),
						},
					},
				},
			},
		},
		{
			name:  "TwoArgFunctionWithTwoTrailingCommas",
			query: `StormEvents | where strcat("abc", "def",,)`,
			err:   true,
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
							Rparen: newSpan(41, 42),
						},
					},
				},
			},
		},
		{
			name:  "ExtraContentInCount",
			query: `StormEvents | count x | where true`,
			err:   true,
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
					&WhereOperator{
						Pipe:    newSpan(22, 23),
						Keyword: newSpan(24, 29),
						Predicate: &Ident{
							Name:     "true",
							NameSpan: newSpan(30, 34),
						},
					},
				},
			},
		},
		{
			name:  "BinaryOp",
			query: "StormEvents | where DamageProperty > 0",
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
						Predicate: &BinaryExpr{
							X: &Ident{
								Name:     "DamageProperty",
								NameSpan: newSpan(20, 34),
							},
							OpSpan: newSpan(35, 36),
							Op:     TokenGT,
							Y: &BasicLit{
								Kind:      TokenNumber,
								ValueSpan: newSpan(37, 38),
								Value:     "0",
							},
						},
					},
				},
			},
		},
		{
			name:  "ComparisonWithSamePrecedenceLHS",
			query: "foo | where x / y * z == 1",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "foo",
						NameSpan: newSpan(0, 3),
					},
				},
				Operators: []TabularOperator{
					&WhereOperator{
						Pipe:    newSpan(4, 5),
						Keyword: newSpan(6, 11),
						Predicate: &BinaryExpr{
							X: &BinaryExpr{
								X: &BinaryExpr{
									X: &Ident{
										Name:     "x",
										NameSpan: newSpan(12, 13),
									},
									OpSpan: newSpan(14, 15),
									Op:     TokenSlash,
									Y: &Ident{
										Name:     "y",
										NameSpan: newSpan(16, 17),
									},
								},
								OpSpan: newSpan(18, 19),
								Op:     TokenStar,
								Y: &Ident{
									Name:     "z",
									NameSpan: newSpan(20, 21),
								},
							},
							OpSpan: newSpan(22, 24),
							Op:     TokenEq,
							Y: &BasicLit{
								Kind:      TokenNumber,
								Value:     "1",
								ValueSpan: newSpan(25, 26),
							},
						},
					},
				},
			},
		},
		{
			name:  "ParenthesizedExpr",
			query: "foo | where x / (y * z) == 1",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "foo",
						NameSpan: newSpan(0, 3),
					},
				},
				Operators: []TabularOperator{
					&WhereOperator{
						Pipe:    newSpan(4, 5),
						Keyword: newSpan(6, 11),
						Predicate: &BinaryExpr{
							X: &BinaryExpr{
								X: &Ident{
									Name:     "x",
									NameSpan: newSpan(12, 13),
								},
								OpSpan: newSpan(14, 15),
								Op:     TokenSlash,
								Y: &ParenExpr{
									Lparen: newSpan(16, 17),
									X: &BinaryExpr{
										X: &Ident{
											Name:     "y",
											NameSpan: newSpan(17, 18),
										},
										OpSpan: newSpan(19, 20),
										Op:     TokenStar,
										Y: &Ident{
											Name:     "z",
											NameSpan: newSpan(21, 22),
										},
									},
									Rparen: newSpan(22, 23),
								},
							},
							OpSpan: newSpan(24, 26),
							Op:     TokenEq,
							Y: &BasicLit{
								Kind:      TokenNumber,
								Value:     "1",
								ValueSpan: newSpan(27, 28),
							},
						},
					},
				},
			},
		},
		{
			name:  "OperatorPrecedence",
			query: "foo | where 2 + 3 * 4 + 5 == 19",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "foo",
						NameSpan: newSpan(0, 3),
					},
				},
				Operators: []TabularOperator{
					&WhereOperator{
						Pipe:    newSpan(4, 5),
						Keyword: newSpan(6, 11),
						Predicate: &BinaryExpr{
							X: &BinaryExpr{
								X: &BinaryExpr{
									X: &BasicLit{
										Kind:      TokenNumber,
										Value:     "2",
										ValueSpan: newSpan(12, 13),
									},
									OpSpan: newSpan(14, 15),
									Op:     TokenPlus,
									Y: &BinaryExpr{
										X: &BasicLit{
											Kind:      TokenNumber,
											Value:     "3",
											ValueSpan: newSpan(16, 17),
										},
										OpSpan: newSpan(18, 19),
										Op:     TokenStar,
										Y: &BasicLit{
											Kind:      TokenNumber,
											Value:     "4",
											ValueSpan: newSpan(20, 21),
										},
									},
								},
								OpSpan: newSpan(22, 23),
								Op:     TokenPlus,
								Y: &BasicLit{
									Kind:      TokenNumber,
									Value:     "5",
									ValueSpan: newSpan(24, 25),
								},
							},
							OpSpan: newSpan(26, 28),
							Op:     TokenEq,
							Y: &BasicLit{
								Kind:      TokenNumber,
								Value:     "19",
								ValueSpan: newSpan(29, 31),
							},
						},
					},
				},
			},
		},
		{
			name:  "BadArgument",
			query: "foo | where strcat('a', .bork, 'x', 'y')",
			err:   true,
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "foo",
						NameSpan: newSpan(0, 3),
					},
				},
				Operators: []TabularOperator{
					&WhereOperator{
						Pipe:    newSpan(4, 5),
						Keyword: newSpan(6, 11),
						Predicate: &CallExpr{
							Func: &Ident{
								Name:     "strcat",
								NameSpan: newSpan(12, 18),
							},
							Lparen: newSpan(18, 19),
							Args: []Expr{
								&BasicLit{
									Kind:      TokenString,
									Value:     "a",
									ValueSpan: newSpan(19, 22),
								},
							},
							Rparen: newSpan(39, 40),
						},
					},
				},
			},
		},
		{
			name:  "BadParentheticalExpr",
			query: "foo | where (.bork) + 2",
			err:   true,
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "foo",
						NameSpan: newSpan(0, 3),
					},
				},
				Operators: []TabularOperator{
					&WhereOperator{
						Pipe:    newSpan(4, 5),
						Keyword: newSpan(6, 11),
						Predicate: &BinaryExpr{
							X: &ParenExpr{
								Lparen: newSpan(12, 13),
								Rparen: newSpan(18, 19),
							},
							OpSpan: newSpan(20, 21),
							Op:     TokenPlus,
							Y: &BasicLit{
								Kind:      TokenNumber,
								Value:     "2",
								ValueSpan: newSpan(22, 23),
							},
						},
					},
				},
			},
		},
		{
			name:  "SortBy",
			query: "foo | sort by bar",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "foo",
						NameSpan: newSpan(0, 3),
					},
				},
				Operators: []TabularOperator{
					&SortOperator{
						Pipe:    newSpan(4, 5),
						Keyword: newSpan(6, 13),
						Terms: []*SortTerm{
							{
								X: &Ident{
									Name:     "bar",
									NameSpan: newSpan(14, 17),
								},
								AscDescSpan: nullSpan(),
								NullsSpan:   nullSpan(),
							},
						},
					},
				},
			},
		},
		{
			name:  "OrderBy",
			query: "foo | order by bar",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "foo",
						NameSpan: newSpan(0, 3),
					},
				},
				Operators: []TabularOperator{
					&SortOperator{
						Pipe:    newSpan(4, 5),
						Keyword: newSpan(6, 14),
						Terms: []*SortTerm{
							{
								X: &Ident{
									Name:     "bar",
									NameSpan: newSpan(15, 18),
								},
								AscDescSpan: nullSpan(),
								NullsSpan:   nullSpan(),
							},
						},
					},
				},
			},
		},
		{
			name:  "SortByTake",
			query: "foo | sort by bar | take 1",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "foo",
						NameSpan: newSpan(0, 3),
					},
				},
				Operators: []TabularOperator{
					&SortOperator{
						Pipe:    newSpan(4, 5),
						Keyword: newSpan(6, 13),
						Terms: []*SortTerm{
							{
								X: &Ident{
									Name:     "bar",
									NameSpan: newSpan(14, 17),
								},
								AscDescSpan: nullSpan(),
								NullsSpan:   nullSpan(),
							},
						},
					},
					&TakeOperator{
						Pipe:    newSpan(18, 19),
						Keyword: newSpan(20, 24),
						RowCount: &BasicLit{
							Kind:      TokenNumber,
							Value:     "1",
							ValueSpan: newSpan(25, 26),
						},
					},
				},
			},
		},
		{
			name:  "SortByMultiple",
			query: "StormEvents | sort by State asc, StartTime desc",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "StormEvents",
						NameSpan: newSpan(0, 11),
					},
				},
				Operators: []TabularOperator{
					&SortOperator{
						Pipe:    newSpan(12, 13),
						Keyword: newSpan(14, 21),
						Terms: []*SortTerm{
							{
								X: &Ident{
									Name:     "State",
									NameSpan: newSpan(22, 27),
								},
								Asc:         true,
								AscDescSpan: newSpan(28, 31),
								NullsFirst:  true,
								NullsSpan:   nullSpan(),
							},
							{
								X: &Ident{
									Name:     "StartTime",
									NameSpan: newSpan(33, 42),
								},
								Asc:         false,
								AscDescSpan: newSpan(43, 47),
								NullsFirst:  false,
								NullsSpan:   nullSpan(),
							},
						},
					},
				},
			},
		},
		{
			name:  "SortByNullsFirst",
			query: "foo | sort by bar nulls first",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "foo",
						NameSpan: newSpan(0, 3),
					},
				},
				Operators: []TabularOperator{
					&SortOperator{
						Pipe:    newSpan(4, 5),
						Keyword: newSpan(6, 13),
						Terms: []*SortTerm{
							{
								X: &Ident{
									Name:     "bar",
									NameSpan: newSpan(14, 17),
								},
								AscDescSpan: nullSpan(),
								NullsFirst:  true,
								NullsSpan:   newSpan(18, 29),
							},
						},
					},
				},
			},
		},
		{
			name:  "Take",
			query: "StormEvents | take 5",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "StormEvents",
						NameSpan: newSpan(0, 11),
					},
				},
				Operators: []TabularOperator{
					&TakeOperator{
						Pipe:    newSpan(12, 13),
						Keyword: newSpan(14, 18),
						RowCount: &BasicLit{
							Kind:      TokenNumber,
							Value:     "5",
							ValueSpan: newSpan(19, 20),
						},
					},
				},
			},
		},
		{
			name:  "Limit",
			query: "StormEvents | limit 5",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "StormEvents",
						NameSpan: newSpan(0, 11),
					},
				},
				Operators: []TabularOperator{
					&TakeOperator{
						Pipe:    newSpan(12, 13),
						Keyword: newSpan(14, 19),
						RowCount: &BasicLit{
							Kind:      TokenNumber,
							Value:     "5",
							ValueSpan: newSpan(20, 21),
						},
					},
				},
			},
		},
		{
			name:  "Project",
			query: "StormEvents | project EventId, State, EventType",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "StormEvents",
						NameSpan: newSpan(0, 11),
					},
				},
				Operators: []TabularOperator{
					&ProjectOperator{
						Pipe:    newSpan(12, 13),
						Keyword: newSpan(14, 21),
						Cols: []*ProjectColumn{
							{
								Name: &Ident{
									Name:     "EventId",
									NameSpan: newSpan(22, 29),
								},
								Assign: nullSpan(),
							},
							{
								Name: &Ident{
									Name:     "State",
									NameSpan: newSpan(31, 36),
								},
								Assign: nullSpan(),
							},
							{
								Name: &Ident{
									Name:     "EventType",
									NameSpan: newSpan(38, 47),
								},
								Assign: nullSpan(),
							},
						},
					},
				},
			},
		},
		{
			name:  "ProjectExpr",
			query: "StormEvents | project TotalInjuries = InjuriesDirect + InjuriesIndirect",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "StormEvents",
						NameSpan: newSpan(0, 11),
					},
				},
				Operators: []TabularOperator{
					&ProjectOperator{
						Pipe:    newSpan(12, 13),
						Keyword: newSpan(14, 21),
						Cols: []*ProjectColumn{
							{
								Name: &Ident{
									Name:     "TotalInjuries",
									NameSpan: newSpan(22, 35),
								},
								Assign: newSpan(36, 37),
								X: &BinaryExpr{
									X: &Ident{
										Name:     "InjuriesDirect",
										NameSpan: newSpan(38, 52),
									},
									OpSpan: newSpan(53, 54),
									Op:     TokenPlus,
									Y: &Ident{
										Name:     "InjuriesIndirect",
										NameSpan: newSpan(55, 71),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "UniqueCombination",
			query: "StormEvents | summarize by State, EventType",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "StormEvents",
						NameSpan: newSpan(0, 11),
					},
				},
				Operators: []TabularOperator{
					&SummarizeOperator{
						Pipe:    newSpan(12, 13),
						Keyword: newSpan(14, 23),
						By:      newSpan(24, 26),
						GroupBy: []*SummarizeColumn{
							{
								Assign: nullSpan(),
								X: &Ident{
									Name:     "State",
									NameSpan: newSpan(27, 32),
								},
							},
							{
								Assign: nullSpan(),
								X: &Ident{
									Name:     "EventType",
									NameSpan: newSpan(34, 43),
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "MinAndMax",
			query: "StormEvents | summarize Min = min(Duration), Max = max(Duration)",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "StormEvents",
						NameSpan: newSpan(0, 11),
					},
				},
				Operators: []TabularOperator{
					&SummarizeOperator{
						Pipe:    newSpan(12, 13),
						Keyword: newSpan(14, 23),
						Cols: []*SummarizeColumn{
							{
								Name: &Ident{
									Name:     "Min",
									NameSpan: newSpan(24, 27),
								},
								Assign: newSpan(28, 29),
								X: &CallExpr{
									Func: &Ident{
										Name:     "min",
										NameSpan: newSpan(30, 33),
									},
									Lparen: newSpan(33, 34),
									Args: []Expr{&Ident{
										Name:     "Duration",
										NameSpan: newSpan(34, 42),
									}},
									Rparen: newSpan(42, 43),
								},
							},
							{
								Name: &Ident{
									Name:     "Max",
									NameSpan: newSpan(45, 48),
								},
								Assign: newSpan(49, 50),
								X: &CallExpr{
									Func: &Ident{
										Name:     "max",
										NameSpan: newSpan(51, 54),
									},
									Lparen: newSpan(54, 55),
									Args: []Expr{&Ident{
										Name:     "Duration",
										NameSpan: newSpan(55, 63),
									}},
									Rparen: newSpan(63, 64),
								},
							},
						},
						By: nullSpan(),
					},
				},
			},
		},
		{
			name:  "DistinctCount",
			query: "StormEvents | summarize TypesOfStorms=dcount(EventType) by State",
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "StormEvents",
						NameSpan: newSpan(0, 11),
					},
				},
				Operators: []TabularOperator{
					&SummarizeOperator{
						Pipe:    newSpan(12, 13),
						Keyword: newSpan(14, 23),
						Cols: []*SummarizeColumn{
							{
								Name: &Ident{
									Name:     "TypesOfStorms",
									NameSpan: newSpan(24, 37),
								},
								Assign: newSpan(37, 38),
								X: &CallExpr{
									Func: &Ident{
										Name:     "dcount",
										NameSpan: newSpan(38, 44),
									},
									Lparen: newSpan(44, 45),
									Args: []Expr{&Ident{
										Name:     "EventType",
										NameSpan: newSpan(45, 54),
									}},
									Rparen: newSpan(54, 55),
								},
							},
						},
						By: newSpan(56, 58),
						GroupBy: []*SummarizeColumn{
							{
								Assign: nullSpan(),
								X: &Ident{
									Name:     "State",
									NameSpan: newSpan(59, 64),
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "ShortSummarize",
			query: "StormEvents | summarize",
			err:   true,
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "StormEvents",
						NameSpan: newSpan(0, 11),
					},
				},
				Operators: []TabularOperator{
					&SummarizeOperator{
						Pipe:    newSpan(12, 13),
						Keyword: newSpan(14, 23),
						By:      nullSpan(),
					},
				},
			},
		},
		{
			name:  "SummarizeByTerminated",
			query: "StormEvents | summarize by",
			err:   true,
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "StormEvents",
						NameSpan: newSpan(0, 11),
					},
				},
				Operators: []TabularOperator{
					&SummarizeOperator{
						Pipe:    newSpan(12, 13),
						Keyword: newSpan(14, 23),
						By:      newSpan(24, 26),
					},
				},
			},
		},
		{
			name:  "SummarizeRandomToken",
			query: "StormEvents | summarize and",
			err:   true,
			want: &TabularExpr{
				Source: &TableRef{
					Table: &Ident{
						Name:     "StormEvents",
						NameSpan: newSpan(0, 11),
					},
				},
				Operators: []TabularOperator{
					&SummarizeOperator{
						Pipe:    newSpan(12, 13),
						Keyword: newSpan(14, 23),
						By:      nullSpan(),
					},
				},
			},
		},
	}

	equateInvalidSpans := cmp.FilterValues(func(span1, span2 Span) bool {
		return !span1.IsValid() && !span2.IsValid()
	}, cmp.Comparer(func(span1, span2 Span) bool {
		return true
	}))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Parse(test.query)
			if err != nil {
				if test.err {
					t.Logf("Parse(%q) error (as expected): %v", test.query, err)
				} else {
					t.Errorf("Parse(%q) returned unexpected error: %v", test.query, err)
				}
			}
			if err == nil && test.err {
				t.Errorf("Parse(%q) did not return an error", test.query)
			}
			if diff := cmp.Diff(test.want, got, cmpopts.EquateEmpty(), equateInvalidSpans); diff != "" {
				t.Errorf("Parse(%q) (-want +got):\n%s", test.query, diff)
			}
		})
	}
}
