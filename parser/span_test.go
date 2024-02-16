// Copyright 2024 RunReveal Inc.
// SPDX-License-Identifier: Apache-2.0

package parser

import "testing"

func TestSpan(t *testing.T) {
	tests := []struct {
		span   Span
		valid  bool
		len    int
		string string
	}{
		{
			span:   newSpan(0, 0),
			valid:  true,
			len:    0,
			string: "[0,0)",
		},
		{
			span:   newSpan(-1, 0),
			valid:  false,
			len:    0,
			string: "[-1,0)",
		},
		{
			span:   newSpan(0, 1),
			valid:  true,
			len:    1,
			string: "[0,1)",
		},
		{
			span:   newSpan(1, 0),
			valid:  false,
			len:    0,
			string: "[1,0)",
		},
		{
			span:   newSpan(5, 7),
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

func TestUnionSpans(t *testing.T) {
	tests := []struct {
		spans []Span
		want  Span
	}{
		{
			spans: []Span{},
			want:  nullSpan(),
		},
		{
			spans: []Span{nullSpan(), nullSpan()},
			want:  nullSpan(),
		},
		{
			spans: []Span{newSpan(1, 5)},
			want:  newSpan(1, 5),
		},
		{
			spans: []Span{nullSpan(), newSpan(1, 5), nullSpan()},
			want:  newSpan(1, 5),
		},
		{
			spans: []Span{nullSpan(), newSpan(4, 5), newSpan(1, 2), nullSpan()},
			want:  newSpan(1, 5),
		},
		{
			spans: []Span{nullSpan(), newSpan(4, 5), newSpan(1, 1), nullSpan()},
			want:  newSpan(1, 5),
		},
	}
	for _, test := range tests {
		got := unionSpans(test.spans...)
		if got != test.want {
			t.Errorf("unionSpans(%v...) = %v; want %v", test.spans, got, test.want)
		}
	}
}
