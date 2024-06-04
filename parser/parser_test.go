// Copyright 2024 RunReveal Inc.
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var parserTests = []struct {
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
		name:  "BadToken",
		query: "!",
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
		name:  "OnlyQuotedTableName",
		query: "`StormEvents`",
		want: &TabularExpr{
			Source: &TableRef{
				Table: &Ident{
					Name:     "StormEvents",
					NameSpan: newSpan(0, 13),
					Quoted:   true,
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
					Predicate: (&Ident{
						Name:     "true",
						NameSpan: newSpan(20, 24),
					}).AsQualified(),
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
		name:  "ZeroArgFunctionWithTrailingComma",
		query: `StormEvents | where rand(,)`,
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
							Name:     "rand",
							NameSpan: newSpan(20, 24),
						},
						Lparen: newSpan(24, 25),
						Rparen: newSpan(26, 27),
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
							(&Ident{
								Name:     "false",
								NameSpan: newSpan(24, 29),
							}).AsQualified(),
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
					Predicate: (&Ident{
						Name:     "true",
						NameSpan: newSpan(30, 34),
					}).AsQualified(),
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
						X: (&Ident{
							Name:     "DamageProperty",
							NameSpan: newSpan(20, 34),
						}).AsQualified(),
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
								X: (&Ident{
									Name:     "x",
									NameSpan: newSpan(12, 13),
								}).AsQualified(),
								OpSpan: newSpan(14, 15),
								Op:     TokenSlash,
								Y: (&Ident{
									Name:     "y",
									NameSpan: newSpan(16, 17),
								}).AsQualified(),
							},
							OpSpan: newSpan(18, 19),
							Op:     TokenStar,
							Y: (&Ident{
								Name:     "z",
								NameSpan: newSpan(20, 21),
							}).AsQualified(),
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
							X: (&Ident{
								Name:     "x",
								NameSpan: newSpan(12, 13),
							}).AsQualified(),
							OpSpan: newSpan(14, 15),
							Op:     TokenSlash,
							Y: &ParenExpr{
								Lparen: newSpan(16, 17),
								X: &BinaryExpr{
									X: (&Ident{
										Name:     "y",
										NameSpan: newSpan(17, 18),
									}).AsQualified(),
									OpSpan: newSpan(19, 20),
									Op:     TokenStar,
									Y: (&Ident{
										Name:     "z",
										NameSpan: newSpan(21, 22),
									}).AsQualified(),
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
		name:  "In",
		query: `StormEvents | where State in ("GEORGIA", "MISSISSIPPI")`,
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
					Predicate: &InExpr{
						X: (&Ident{
							Name:     "State",
							NameSpan: newSpan(20, 25),
						}).AsQualified(),
						In:     newSpan(26, 28),
						Lparen: newSpan(29, 30),
						Vals: []Expr{
							&BasicLit{
								Kind:      TokenString,
								ValueSpan: newSpan(30, 39),
								Value:     "GEORGIA",
							},
							&BasicLit{
								Kind:      TokenString,
								ValueSpan: newSpan(41, 54),
								Value:     "MISSISSIPPI",
							},
						},
						Rparen: newSpan(54, 55),
					},
				},
			},
		},
	},
	{
		name:  "InAnd",
		query: `StormEvents | where State in ("GEORGIA", "MISSISSIPPI") and DamageProperty > 10000`,
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
						X: &InExpr{
							X: (&Ident{
								Name:     "State",
								NameSpan: newSpan(20, 25),
							}).AsQualified(),
							In:     newSpan(26, 28),
							Lparen: newSpan(29, 30),
							Vals: []Expr{
								&BasicLit{
									Kind:      TokenString,
									ValueSpan: newSpan(30, 39),
									Value:     "GEORGIA",
								},
								&BasicLit{
									Kind:      TokenString,
									ValueSpan: newSpan(41, 54),
									Value:     "MISSISSIPPI",
								},
							},
							Rparen: newSpan(54, 55),
						},
						Op:     TokenAnd,
						OpSpan: newSpan(56, 59),
						Y: &BinaryExpr{
							X: (&Ident{
								Name:     "DamageProperty",
								NameSpan: newSpan(60, 74),
							}).AsQualified(),
							Op:     TokenGT,
							OpSpan: newSpan(75, 76),
							Y: &BasicLit{
								Kind:      TokenNumber,
								Value:     "10000",
								ValueSpan: newSpan(77, 82),
							},
						},
					},
				},
			},
		},
	},
	{
		name:  "InAndFlipped",
		query: `StormEvents | where DamageProperty > 10000 and State in ("GEORGIA", "MISSISSIPPI")`,
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
						X: &BinaryExpr{
							X: (&Ident{
								Name:     "DamageProperty",
								NameSpan: newSpan(20, 34),
							}).AsQualified(),
							Op:     TokenGT,
							OpSpan: newSpan(35, 36),
							Y: &BasicLit{
								Kind:      TokenNumber,
								Value:     "10000",
								ValueSpan: newSpan(37, 42),
							},
						},
						Op:     TokenAnd,
						OpSpan: newSpan(43, 46),
						Y: &InExpr{
							X: (&Ident{
								Name:     "State",
								NameSpan: newSpan(47, 52),
							}).AsQualified(),
							In:     newSpan(53, 55),
							Lparen: newSpan(56, 57),
							Vals: []Expr{
								&BasicLit{
									Kind:      TokenString,
									ValueSpan: newSpan(57, 66),
									Value:     "GEORGIA",
								},
								&BasicLit{
									Kind:      TokenString,
									ValueSpan: newSpan(68, 81),
									Value:     "MISSISSIPPI",
								},
							},
							Rparen: newSpan(81, 82),
						},
					},
				},
			},
		},
	},
	{
		name:  "MapKey",
		query: `tab | where mapcol['strkey'] == 42`,
		want: &TabularExpr{
			Source: &TableRef{
				Table: &Ident{
					Name:     "tab",
					NameSpan: newSpan(0, 3),
				},
			},
			Operators: []TabularOperator{
				&WhereOperator{
					Pipe:    newSpan(4, 5),
					Keyword: newSpan(6, 11),
					Predicate: &BinaryExpr{
						X: &IndexExpr{
							X: (&Ident{
								Name:     "mapcol",
								NameSpan: newSpan(12, 18),
							}).AsQualified(),
							Lbrack: newSpan(18, 19),
							Index: &BasicLit{
								Kind:      TokenString,
								Value:     "strkey",
								ValueSpan: newSpan(19, 27),
							},
							Rbrack: newSpan(27, 28),
						},
						Op:     TokenEq,
						OpSpan: newSpan(29, 31),
						Y: &BasicLit{
							Kind:      TokenNumber,
							Value:     "42",
							ValueSpan: newSpan(32, 34),
						},
					},
				},
			},
		},
	},
	{
		name:  "MapKeyTrailingExpression",
		query: `tab | where mapcol['strkey' x] == 42`,
		err:   true,
		want: &TabularExpr{
			Source: &TableRef{
				Table: &Ident{
					Name:     "tab",
					NameSpan: newSpan(0, 3),
				},
			},
			Operators: []TabularOperator{
				&WhereOperator{
					Pipe:    newSpan(4, 5),
					Keyword: newSpan(6, 11),
					Predicate: &BinaryExpr{
						X: &IndexExpr{
							X: (&Ident{
								Name:     "mapcol",
								NameSpan: newSpan(12, 18),
							}).AsQualified(),
							Lbrack: newSpan(18, 19),
							Index: &BasicLit{
								Kind:      TokenString,
								Value:     "strkey",
								ValueSpan: newSpan(19, 27),
							},
							Rbrack: newSpan(29, 30),
						},
						Op:     TokenEq,
						OpSpan: newSpan(31, 33),
						Y: &BasicLit{
							Kind:      TokenNumber,
							Value:     "42",
							ValueSpan: newSpan(34, 36),
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
							X: (&Ident{
								Name:     "bar",
								NameSpan: newSpan(14, 17),
							}).AsQualified(),
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
							X: (&Ident{
								Name:     "bar",
								NameSpan: newSpan(15, 18),
							}).AsQualified(),
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
							X: (&Ident{
								Name:     "bar",
								NameSpan: newSpan(14, 17),
							}).AsQualified(),
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
							X: (&Ident{
								Name:     "State",
								NameSpan: newSpan(22, 27),
							}).AsQualified(),
							Asc:         true,
							AscDescSpan: newSpan(28, 31),
							NullsFirst:  true,
							NullsSpan:   nullSpan(),
						},
						{
							X: (&Ident{
								Name:     "StartTime",
								NameSpan: newSpan(33, 42),
							}).AsQualified(),
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
							X: (&Ident{
								Name:     "bar",
								NameSpan: newSpan(14, 17),
							}).AsQualified(),
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
		name:  "ProjectError",
		query: "StormEvents | project EventId=1 State",
		err:   true,
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
							Assign: Span{
								Start: 29,
								End:   30,
							},
							X: &BasicLit{
								Kind:      TokenNumber,
								Value:     "1",
								ValueSpan: newSpan(30, 31),
							},
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
								X: (&Ident{
									Name:     "InjuriesDirect",
									NameSpan: newSpan(38, 52),
								}).AsQualified(),
								OpSpan: newSpan(53, 54),
								Op:     TokenPlus,
								Y: (&Ident{
									Name:     "InjuriesIndirect",
									NameSpan: newSpan(55, 71),
								}).AsQualified(),
							},
						},
					},
				},
			},
		},
	},
	{
		name:  "ExtendExpr",
		query: "StormEvents | extend TotalInjuries = InjuriesDirect + InjuriesIndirect",
		want: &TabularExpr{
			Source: &TableRef{
				Table: &Ident{
					Name:     "StormEvents",
					NameSpan: newSpan(0, 11),
				},
			},
			Operators: []TabularOperator{
				&ExtendOperator{
					Pipe:    newSpan(12, 13),
					Keyword: newSpan(14, 20),
					Cols: []*ExtendColumn{
						{
							Name: &Ident{
								Name:     "TotalInjuries",
								NameSpan: newSpan(21, 34),
							},
							Assign: newSpan(35, 36),
							X: &BinaryExpr{
								X: (&Ident{
									Name:     "InjuriesDirect",
									NameSpan: newSpan(37, 51),
								}).AsQualified(),
								OpSpan: newSpan(52, 53),
								Op:     TokenPlus,
								Y: (&Ident{
									Name:     "InjuriesIndirect",
									NameSpan: newSpan(54, 70),
								}).AsQualified(),
							},
						},
					},
				},
			},
		},
	},
	{
		name:  "ExtendError",
		query: "StormEvents | extend FooFooF=1 State",
		err:   true,
		want: &TabularExpr{
			Source: &TableRef{
				Table: &Ident{
					Name:     "StormEvents",
					NameSpan: newSpan(0, 11),
				},
			},
			Operators: []TabularOperator{
				&ExtendOperator{
					Pipe:    newSpan(12, 13),
					Keyword: newSpan(14, 20),
					Cols: []*ExtendColumn{
						{
							Name: &Ident{
								Name:     "FooFooF",
								NameSpan: newSpan(21, 28),
							},
							Assign: Span{
								Start: 28,
								End:   29,
							},
							X: &BasicLit{
								Kind:      TokenNumber,
								Value:     "1",
								ValueSpan: newSpan(29, 30),
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
							X: (&Ident{
								Name:     "State",
								NameSpan: newSpan(27, 32),
							}).AsQualified(),
						},
						{
							Assign: nullSpan(),
							X: (&Ident{
								Name:     "EventType",
								NameSpan: newSpan(34, 43),
							}).AsQualified(),
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
								Args: []Expr{(&Ident{
									Name:     "Duration",
									NameSpan: newSpan(34, 42),
								}).AsQualified()},
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
								Args: []Expr{(&Ident{
									Name:     "Duration",
									NameSpan: newSpan(55, 63),
								}).AsQualified()},
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
								Args: []Expr{(&Ident{
									Name:     "EventType",
									NameSpan: newSpan(45, 54),
								}).AsQualified()},
								Rparen: newSpan(54, 55),
							},
						},
					},
					By: newSpan(56, 58),
					GroupBy: []*SummarizeColumn{
						{
							Assign: nullSpan(),
							X: (&Ident{
								Name:     "State",
								NameSpan: newSpan(59, 64),
							}).AsQualified(),
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
	{
		name:  "Top",
		query: "StormEvents | top 3 by InjuriesDirect",
		want: &TabularExpr{
			Source: &TableRef{
				Table: &Ident{
					Name:     "StormEvents",
					NameSpan: newSpan(0, 11),
				},
			},
			Operators: []TabularOperator{
				&TopOperator{
					Pipe:    newSpan(12, 13),
					Keyword: newSpan(14, 17),
					RowCount: &BasicLit{
						Kind:      TokenNumber,
						Value:     "3",
						ValueSpan: newSpan(18, 19),
					},
					By: newSpan(20, 22),
					Col: &SortTerm{
						X: (&Ident{
							Name:     "InjuriesDirect",
							NameSpan: newSpan(23, 37),
						}).AsQualified(),
						AscDescSpan: nullSpan(),
						NullsSpan:   nullSpan(),
					},
				},
			},
		},
	},
	{
		name:  "Join",
		query: "X | join (Y) on Key",
		want: &TabularExpr{
			Source: &TableRef{
				Table: &Ident{
					Name:     "X",
					NameSpan: newSpan(0, 1),
				},
			},
			Operators: []TabularOperator{
				&JoinOperator{
					Pipe:    newSpan(2, 3),
					Keyword: newSpan(4, 8),

					Kind:       nullSpan(),
					KindAssign: nullSpan(),

					Lparen: newSpan(9, 10),
					Right: &TabularExpr{
						Source: &TableRef{
							Table: &Ident{
								Name:     "Y",
								NameSpan: newSpan(10, 11),
							},
						},
					},
					Rparen: newSpan(11, 12),
					On:     newSpan(13, 15),
					Conditions: []Expr{
						(&Ident{
							Name:     "Key",
							NameSpan: newSpan(16, 19),
						}).AsQualified(),
					},
				},
			},
		},
	},
	{
		name:  "JoinLeft",
		query: "X | join kind=leftouter (Y) on Key",
		want: &TabularExpr{
			Source: &TableRef{
				Table: &Ident{
					Name:     "X",
					NameSpan: newSpan(0, 1),
				},
			},
			Operators: []TabularOperator{
				&JoinOperator{
					Pipe:    newSpan(2, 3),
					Keyword: newSpan(4, 8),

					Kind:       newSpan(9, 13),
					KindAssign: newSpan(13, 14),
					Flavor: &Ident{
						Name:     "leftouter",
						NameSpan: newSpan(14, 23),
					},

					Lparen: newSpan(24, 25),
					Right: &TabularExpr{
						Source: &TableRef{
							Table: &Ident{
								Name:     "Y",
								NameSpan: newSpan(25, 26),
							},
						},
					},
					Rparen: newSpan(26, 27),
					On:     newSpan(28, 30),
					Conditions: []Expr{
						(&Ident{
							Name:     "Key",
							NameSpan: newSpan(31, 34),
						}).AsQualified(),
					},
				},
			},
		},
	},
	{
		name:  "JoinBadFlavor",
		query: "X | join kind=salt (Y) on Key",
		err:   true,
		want: &TabularExpr{
			Source: &TableRef{
				Table: &Ident{
					Name:     "X",
					NameSpan: newSpan(0, 1),
				},
			},
			Operators: []TabularOperator{
				&JoinOperator{
					Pipe:    newSpan(2, 3),
					Keyword: newSpan(4, 8),

					Kind:       newSpan(9, 13),
					KindAssign: newSpan(13, 14),
					Flavor: &Ident{
						Name:     "salt",
						NameSpan: newSpan(14, 18),
					},

					Lparen: newSpan(19, 20),
					Right: &TabularExpr{
						Source: &TableRef{
							Table: &Ident{
								Name:     "Y",
								NameSpan: newSpan(20, 21),
							},
						},
					},
					Rparen: newSpan(21, 22),
					On:     newSpan(23, 25),
					Conditions: []Expr{
						(&Ident{
							Name:     "Key",
							NameSpan: newSpan(26, 29),
						}).AsQualified(),
					},
				},
			},
		},
	},
	{
		name:  "JoinComplexRight",
		query: "X | join (Y | where z == 5) on Key",
		want: &TabularExpr{
			Source: &TableRef{
				Table: &Ident{
					Name:     "X",
					NameSpan: newSpan(0, 1),
				},
			},
			Operators: []TabularOperator{
				&JoinOperator{
					Pipe:    newSpan(2, 3),
					Keyword: newSpan(4, 8),

					Kind:       nullSpan(),
					KindAssign: nullSpan(),

					Lparen: newSpan(9, 10),
					Right: &TabularExpr{
						Source: &TableRef{
							Table: &Ident{
								Name:     "Y",
								NameSpan: newSpan(10, 11),
							},
						},
						Operators: []TabularOperator{
							&WhereOperator{
								Pipe:    newSpan(12, 13),
								Keyword: newSpan(14, 19),
								Predicate: &BinaryExpr{
									X: (&Ident{
										Name:     "z",
										NameSpan: newSpan(20, 21),
									}).AsQualified(),
									Op:     TokenEq,
									OpSpan: newSpan(22, 24),
									Y: &BasicLit{
										Kind:      TokenNumber,
										Value:     "5",
										ValueSpan: newSpan(25, 26),
									},
								},
							},
						},
					},
					Rparen: newSpan(26, 27),
					On:     newSpan(28, 30),
					Conditions: []Expr{
						(&Ident{
							Name:     "Key",
							NameSpan: newSpan(31, 34),
						}).AsQualified(),
					},
				},
			},
		},
	},
	{
		name:  "JoinExplicitCondition",
		query: "X | join (Y) on $left.Key == $right.Key",
		want: &TabularExpr{
			Source: &TableRef{
				Table: &Ident{
					Name:     "X",
					NameSpan: newSpan(0, 1),
				},
			},
			Operators: []TabularOperator{
				&JoinOperator{
					Pipe:    newSpan(2, 3),
					Keyword: newSpan(4, 8),

					Kind:       nullSpan(),
					KindAssign: nullSpan(),

					Lparen: newSpan(9, 10),
					Right: &TabularExpr{
						Source: &TableRef{
							Table: &Ident{
								Name:     "Y",
								NameSpan: newSpan(10, 11),
							},
						},
					},
					Rparen: newSpan(11, 12),
					On:     newSpan(13, 15),
					Conditions: []Expr{
						&BinaryExpr{
							X: &QualifiedIdent{
								Parts: []*Ident{
									{
										Name:     "$left",
										NameSpan: newSpan(16, 21),
									},
									{
										Name:     "Key",
										NameSpan: newSpan(22, 25),
									},
								},
							},
							Op:     TokenEq,
							OpSpan: newSpan(26, 28),
							Y: &QualifiedIdent{
								Parts: []*Ident{
									{
										Name:     "$right",
										NameSpan: newSpan(29, 35),
									},
									{
										Name:     "Key",
										NameSpan: newSpan(36, 39),
									},
								},
							},
						},
					},
				},
			},
		},
	},
	{
		name:  "JoinAndCount",
		query: "X | join (Y) on Key | count",
		want: &TabularExpr{
			Source: &TableRef{
				Table: &Ident{
					Name:     "X",
					NameSpan: newSpan(0, 1),
				},
			},
			Operators: []TabularOperator{
				&JoinOperator{
					Pipe:    newSpan(2, 3),
					Keyword: newSpan(4, 8),

					Kind:       nullSpan(),
					KindAssign: nullSpan(),

					Lparen: newSpan(9, 10),
					Right: &TabularExpr{
						Source: &TableRef{
							Table: &Ident{
								Name:     "Y",
								NameSpan: newSpan(10, 11),
							},
						},
					},
					Rparen: newSpan(11, 12),
					On:     newSpan(13, 15),
					Conditions: []Expr{
						(&Ident{
							Name:     "Key",
							NameSpan: newSpan(16, 19),
						}).AsQualified(),
					},
				},
				&CountOperator{
					Pipe:    newSpan(20, 21),
					Keyword: newSpan(22, 27),
				},
			},
		},
	},
	{
		name:  "As",
		query: "X | as Y",
		want: &TabularExpr{
			Source: &TableRef{
				Table: &Ident{
					Name:     "X",
					NameSpan: newSpan(0, 1),
				},
			},
			Operators: []TabularOperator{
				&AsOperator{
					Pipe:    newSpan(2, 3),
					Keyword: newSpan(4, 6),
					Name: &Ident{
						Name:     "Y",
						NameSpan: newSpan(7, 8),
					},
				},
			},
		},
	},
	{
		name:  "TrailingPipe",
		query: "X |",
		want: &TabularExpr{
			Source: &TableRef{
				Table: &Ident{
					Name:     "X",
					NameSpan: newSpan(0, 1),
				},
			},
			Operators: []TabularOperator{
				&UnknownTabularOperator{
					Pipe: newSpan(2, 3),
				},
			},
		},
		err: true,
	},
	{
		name:  "UnknownOperator",
		query: "X | xyzzy",
		want: &TabularExpr{
			Source: &TableRef{
				Table: &Ident{
					Name:     "X",
					NameSpan: newSpan(0, 1),
				},
			},
			Operators: []TabularOperator{
				&UnknownTabularOperator{
					Pipe: newSpan(2, 3),
					Tokens: []Token{
						{
							Kind:  TokenIdentifier,
							Span:  newSpan(4, 9),
							Value: "xyzzy",
						},
					},
				},
			},
		},
		err: true,
	},
	{
		name:  "UnknownOperatorInMiddle",
		query: "X | xyzzy (Y | Z) | count",
		want: &TabularExpr{
			Source: &TableRef{
				Table: &Ident{
					Name:     "X",
					NameSpan: newSpan(0, 1),
				},
			},
			Operators: []TabularOperator{
				&UnknownTabularOperator{
					Pipe: newSpan(2, 3),
					Tokens: []Token{
						{
							Kind:  TokenIdentifier,
							Span:  newSpan(4, 9),
							Value: "xyzzy",
						},
						{
							Kind: TokenLParen,
							Span: newSpan(10, 11),
						},
						{
							Kind:  TokenIdentifier,
							Span:  newSpan(11, 12),
							Value: "Y",
						},
						{
							Kind: TokenPipe,
							Span: newSpan(13, 14),
						},
						{
							Kind:  TokenIdentifier,
							Span:  newSpan(15, 16),
							Value: "Z",
						},
						{
							Kind: TokenRParen,
							Span: newSpan(16, 17),
						},
					},
				},
				&CountOperator{
					Pipe:    newSpan(18, 19),
					Keyword: newSpan(20, 25),
				},
			},
		},
		err: true,
	},
}

func TestParse(t *testing.T) {
	equateInvalidSpans := cmp.FilterValues(func(span1, span2 Span) bool {
		return !span1.IsValid() && !span2.IsValid()
	}, cmp.Comparer(func(span1, span2 Span) bool {
		return true
	}))

	for _, test := range parserTests {
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

func FuzzParse(f *testing.F) {
	for _, test := range parserTests {
		f.Add(test.query)
	}

	f.Fuzz(func(t *testing.T, query string) {
		// At the moment, just check for not crashing.
		Parse(query)
	})
}

func BenchmarkParse(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if _, err := Parse(`StormEvents | where EventType == "Tornado" or EventType != "Thunderstorm Wind"`); err != nil {
			b.Fatal(err)
		}
	}
}
