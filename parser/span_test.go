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
