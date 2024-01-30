package parser

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestScan(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  []Token
	}{
		{
			name:  "Empty",
			query: "",
			want:  []Token{},
		},
		{
			name:  "SingleIdent",
			query: "StormEvents\n",
			want: []Token{
				{Kind: TokenIdentifier, Span: Span{Start: 0, End: 11}, Value: "StormEvents"},
			},
		},
		{
			name:  "Pipeline",
			query: "foo | bar",
			want: []Token{
				{Kind: TokenIdentifier, Span: Span{Start: 0, End: 3}, Value: "foo"},
				{Kind: TokenPipe, Span: Span{Start: 4, End: 5}},
				{Kind: TokenIdentifier, Span: Span{Start: 6, End: 9}, Value: "bar"},
			},
		},
		{
			name:  "SingleQuotedIdent",
			query: "['foo']\n",
			want: []Token{
				{Kind: TokenQuotedIdentifier, Span: Span{Start: 0, End: 7}, Value: "foo"},
			},
		},
		{
			name:  "DoubleQuotedIdent",
			query: `["foo"]`,
			want: []Token{
				{Kind: TokenQuotedIdentifier, Span: Span{Start: 0, End: 7}, Value: "foo"},
			},
		},
		{
			name:  "UnterminatedQuotedIdent",
			query: `["foo"`,
			want: []Token{
				{Kind: TokenError, Span: Span{Start: 0, End: 6}},
			},
		},
		{
			name:  "LineSplitQuotedIdent",
			query: "['foo\nbar']",
			want: []Token{
				{Kind: TokenError, Span: Span{Start: 0, End: 5}},
				{Kind: TokenIdentifier, Span: Span{Start: 6, End: 9}, Value: "bar"},
				{Kind: TokenError, Span: Span{Start: 9, End: 11}},
			},
		},
		{
			name:  "Comments",
			query: "StormEvents // the table name\n// Another comment\n| count",
			want: []Token{
				{Kind: TokenIdentifier, Span: Span{Start: 0, End: 11}, Value: "StormEvents"},
				{Kind: TokenPipe, Span: Span{Start: 49, End: 50}},
				{Kind: TokenIdentifier, Span: Span{Start: 51, End: 56}, Value: "count"},
			},
		},
		{
			name:  "Slash",
			query: "foo / bar",
			want: []Token{
				{Kind: TokenIdentifier, Span: Span{Start: 0, End: 3}, Value: "foo"},
				{Kind: TokenSlash, Span: Span{Start: 4, End: 5}},
				{Kind: TokenIdentifier, Span: Span{Start: 6, End: 9}, Value: "bar"},
			},
		},
		{
			name:  "Zero",
			query: "0",
			want: []Token{
				{Kind: TokenNumber, Span: Span{Start: 0, End: 1}, Value: "0"},
			},
		},
		{
			name:  "LeadingZeroes",
			query: "007",
			want: []Token{
				{Kind: TokenNumber, Span: Span{Start: 0, End: 3}, Value: "7"},
			},
		},
		{
			name:  "Integer",
			query: "123",
			want: []Token{
				{Kind: TokenNumber, Span: Span{Start: 0, End: 3}, Value: "123"},
			},
		},
		{
			name:  "Float",
			query: "3.14",
			want: []Token{
				{Kind: TokenNumber, Span: Span{Start: 0, End: 4}, Value: "3.14"},
			},
		},
		{
			name:  "Exponent",
			query: "1e-9",
			want: []Token{
				{Kind: TokenNumber, Span: Span{Start: 0, End: 4}, Value: "1e-9"},
			},
		},
		{
			name:  "ZeroExponent",
			query: "0e9",
			want: []Token{
				{Kind: TokenNumber, Span: Span{Start: 0, End: 3}, Value: "0e9"},
			},
		},
		{
			name:  "LeadingDot",
			query: ".001",
			want: []Token{
				{Kind: TokenNumber, Span: Span{Start: 0, End: 4}, Value: "0.001"},
			},
		},
		{
			name:  "ZeroDotDecimal",
			query: "0.001",
			want: []Token{
				{Kind: TokenNumber, Span: Span{Start: 0, End: 5}, Value: "0.001"},
			},
		},
		{
			name:  "LeadingDotIdentifier",
			query: ".foo",
			want: []Token{
				{Kind: TokenDot, Span: Span{Start: 0, End: 1}},
				{Kind: TokenIdentifier, Span: Span{Start: 1, End: 4}, Value: "foo"},
			},
		},
		{
			name:  "Hexadecimal",
			query: "0xdeadbeef",
			want: []Token{
				{Kind: TokenNumber, Span: Span{Start: 0, End: 10}, Value: "3735928559"},
			},
		},
		{
			name:  "UnterminatedHex",
			query: "0x",
			want: []Token{
				{Kind: TokenError, Span: Span{Start: 0, End: 2}},
			},
		},
		{
			name:  "BrokenHex",
			query: "0xy",
			want: []Token{
				{Kind: TokenError, Span: Span{Start: 0, End: 2}},
				{Kind: TokenIdentifier, Span: Span{Start: 2, End: 3}, Value: "y"},
			},
		},
		{
			name:  "JustDot",
			query: ".",
			want: []Token{
				{Kind: TokenDot, Span: Span{Start: 0, End: 1}},
			},
		},
		{
			name:  "SingleQuotedLiteral",
			query: `'abc'`,
			want: []Token{
				{Kind: TokenString, Span: Span{Start: 0, End: 5}, Value: "abc"},
			},
		},
		{
			name:  "DoubleQuotedLiteral",
			query: `"abc"`,
			want: []Token{
				{Kind: TokenString, Span: Span{Start: 0, End: 5}, Value: "abc"},
			},
		},
		{
			name:  "UnterminatedString",
			query: `"abc`,
			want: []Token{
				{Kind: TokenError, Span: Span{Start: 0, End: 4}},
			},
		},
		{
			name:  "StringWithNewline",
			query: "\"abc\ndef\"",
			want: []Token{
				{Kind: TokenError, Span: Span{Start: 0, End: 4}},
				{Kind: TokenIdentifier, Span: Span{Start: 5, End: 8}, Value: "def"},
				{Kind: TokenError, Span: Span{Start: 8, End: 9}},
			},
		},
		{
			name:  "StringWithEscapes",
			query: `"abc\"\n\t\\def"`,
			want: []Token{
				{Kind: TokenString, Span: Span{Start: 0, End: 16}, Value: "abc\"\n\t\\def"},
			},
		},
		{
			name:  "Parentheses",
			query: "(x)",
			want: []Token{
				{Kind: TokenLParen, Span: Span{Start: 0, End: 1}},
				{Kind: TokenIdentifier, Span: Span{Start: 1, End: 2}, Value: "x"},
				{Kind: TokenRParen, Span: Span{Start: 2, End: 3}},
			},
		},
		{
			name:  "And",
			query: "this and that",
			want: []Token{
				{Kind: TokenIdentifier, Span: Span{Start: 0, End: 4}, Value: "this"},
				{Kind: TokenAnd, Span: Span{Start: 5, End: 8}},
				{Kind: TokenIdentifier, Span: Span{Start: 9, End: 13}, Value: "that"},
			},
		},
		{
			name:  "Or",
			query: "this or that",
			want: []Token{
				{Kind: TokenIdentifier, Span: Span{Start: 0, End: 4}, Value: "this"},
				{Kind: TokenOr, Span: Span{Start: 5, End: 7}},
				{Kind: TokenIdentifier, Span: Span{Start: 8, End: 12}, Value: "that"},
			},
		},
		{
			name:  "Equals",
			query: "x == 5",
			want: []Token{
				{Kind: TokenIdentifier, Span: Span{Start: 0, End: 1}, Value: "x"},
				{Kind: TokenEq, Span: Span{Start: 2, End: 4}},
				{Kind: TokenNumber, Span: Span{Start: 5, End: 6}, Value: "5"},
			},
		},
		{
			name:  "NotEquals",
			query: "x != 5",
			want: []Token{
				{Kind: TokenIdentifier, Span: Span{Start: 0, End: 1}, Value: "x"},
				{Kind: TokenNE, Span: Span{Start: 2, End: 4}},
				{Kind: TokenNumber, Span: Span{Start: 5, End: 6}, Value: "5"},
			},
		},
		{
			name:  "WhereClause",
			query: `where EventType == "Tornado" or EventType != "Thunderstorm Wind"`,
			want: []Token{
				{Kind: TokenIdentifier, Span: Span{Start: 0, End: 5}, Value: "where"},
				{Kind: TokenIdentifier, Span: Span{Start: 6, End: 15}, Value: "EventType"},
				{Kind: TokenEq, Span: Span{Start: 16, End: 18}},
				{Kind: TokenString, Span: Span{Start: 19, End: 28}, Value: "Tornado"},
				{Kind: TokenOr, Span: Span{Start: 29, End: 31}},
				{Kind: TokenIdentifier, Span: Span{Start: 32, End: 41}, Value: "EventType"},
				{Kind: TokenNE, Span: Span{Start: 42, End: 44}},
				{Kind: TokenString, Span: Span{Start: 45, End: 64}, Value: "Thunderstorm Wind"},
			},
		},
		{
			name:  "Add",
			query: "3.14 + 3.14",
			want: []Token{
				{Kind: TokenNumber, Span: Span{Start: 0, End: 4}, Value: "3.14"},
				{Kind: TokenPlus, Span: Span{Start: 5, End: 6}},
				{Kind: TokenNumber, Span: Span{Start: 7, End: 11}, Value: "3.14"},
			},
		},
		{
			name:  "Subtract",
			query: "0.23 - 0.22",
			want: []Token{
				{Kind: TokenNumber, Span: Span{Start: 0, End: 4}, Value: "0.23"},
				{Kind: TokenMinus, Span: Span{Start: 5, End: 6}},
				{Kind: TokenNumber, Span: Span{Start: 7, End: 11}, Value: "0.22"},
			},
		},
		{
			name:  "Multiply",
			query: "2 * 2",
			want: []Token{
				{Kind: TokenNumber, Span: Span{Start: 0, End: 1}, Value: "2"},
				{Kind: TokenStar, Span: Span{Start: 2, End: 3}},
				{Kind: TokenNumber, Span: Span{Start: 4, End: 5}, Value: "2"},
			},
		},
		{
			name:  "Divide",
			query: "4 / 2",
			want: []Token{
				{Kind: TokenNumber, Span: Span{Start: 0, End: 1}, Value: "4"},
				{Kind: TokenSlash, Span: Span{Start: 2, End: 3}},
				{Kind: TokenNumber, Span: Span{Start: 4, End: 5}, Value: "2"},
			},
		},
		{
			name:  "Modulo",
			query: "4 % 2",
			want: []Token{
				{Kind: TokenNumber, Span: Span{Start: 0, End: 1}, Value: "4"},
				{Kind: TokenMod, Span: Span{Start: 2, End: 3}},
				{Kind: TokenNumber, Span: Span{Start: 4, End: 5}, Value: "2"},
			},
		},
		{
			name:  "Less",
			query: "1 < 10",
			want: []Token{
				{Kind: TokenNumber, Span: Span{Start: 0, End: 1}, Value: "1"},
				{Kind: TokenLT, Span: Span{Start: 2, End: 3}},
				{Kind: TokenNumber, Span: Span{Start: 4, End: 6}, Value: "10"},
			},
		},
		{
			name:  "Greater",
			query: "0.23 > 0.22",
			want: []Token{
				{Kind: TokenNumber, Span: Span{Start: 0, End: 4}, Value: "0.23"},
				{Kind: TokenGT, Span: Span{Start: 5, End: 6}},
				{Kind: TokenNumber, Span: Span{Start: 7, End: 11}, Value: "0.22"},
			},
		},
		{
			name:  "LessOrEqual",
			query: "4 <= 5",
			want: []Token{
				{Kind: TokenNumber, Span: Span{Start: 0, End: 1}, Value: "4"},
				{Kind: TokenLE, Span: Span{Start: 2, End: 4}},
				{Kind: TokenNumber, Span: Span{Start: 5, End: 6}, Value: "5"},
			},
		},
		{
			name:  "GreaterOrEqual",
			query: "5 >= 4",
			want: []Token{
				{Kind: TokenNumber, Span: Span{Start: 0, End: 1}, Value: "5"},
				{Kind: TokenGE, Span: Span{Start: 2, End: 4}},
				{Kind: TokenNumber, Span: Span{Start: 5, End: 6}, Value: "4"},
			},
		},
		{
			name:  "CaseInsensitiveEquals",
			query: `"abc" =~ "ABC"`,
			want: []Token{
				{Kind: TokenString, Span: Span{Start: 0, End: 5}, Value: "abc"},
				{Kind: TokenCaseInsensitiveEq, Span: Span{Start: 6, End: 8}},
				{Kind: TokenString, Span: Span{Start: 9, End: 14}, Value: "ABC"},
			},
		},
		{
			name:  "CaseInsensitiveNotEquals",
			query: `"aBc" !~ "xyz"`,
			want: []Token{
				{Kind: TokenString, Span: Span{Start: 0, End: 5}, Value: "aBc"},
				{Kind: TokenCaseInsensitiveNE, Span: Span{Start: 6, End: 8}},
				{Kind: TokenString, Span: Span{Start: 9, End: 14}, Value: "xyz"},
			},
		},
	}

	ignoreErrorValues := cmp.FilterValues(func(tok1, tok2 Token) bool {
		return tok1.Kind == TokenError || tok2.Kind == TokenError
	}, cmp.Transformer("ignoreErrorValues", func(tok Token) Token {
		if tok.Kind == TokenError {
			tok.Value = ""
		}
		return tok
	}))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Scan(test.query)

			if diff := cmp.Diff(test.want, got, cmpopts.EquateEmpty(), ignoreErrorValues); diff != "" {
				t.Errorf("Scan(%q) (-want +got):\n%s", test.query, diff)
			}
		})
	}
}

var tokenType = reflect.TypeOf((*Token)(nil)).Elem()

func TestSpan(t *testing.T) {
	tests := []struct {
		span   Span
		valid  bool
		len    int
		string string
	}{
		{
			span:   Span{},
			valid:  true,
			len:    0,
			string: "[0,0)",
		},
		{
			span:   Span{-1, 0},
			valid:  false,
			len:    0,
			string: "[-1,0)",
		},
		{
			span:   Span{0, 1},
			valid:  true,
			len:    1,
			string: "[0,1)",
		},
		{
			span:   Span{1, 0},
			valid:  false,
			len:    0,
			string: "[1,0)",
		},
		{
			span:   Span{5, 7},
			valid:  true,
			len:    2,
			string: "[5,7)",
		},
	}

	t.Run("IsValid", func(t *testing.T) {
		for _, test := range tests {
			if got := test.span.IsValid(); got != test.valid {
				t.Errorf("(%#v).IsValid() = %t; want %t", test.span, got, test.valid)
			}
		}
	})

	t.Run("Len", func(t *testing.T) {
		for _, test := range tests {
			if got := test.span.Len(); got != test.len {
				t.Errorf("(%#v).Len() = %d; want %d", test.span, got, test.len)
			}
		}
	})

	t.Run("String", func(t *testing.T) {
		for _, test := range tests {
			if got := test.span.String(); got != test.string {
				t.Errorf("(%#v).String() = %q; want %q", test.span, got, test.string)
			}
		}
	})
}
