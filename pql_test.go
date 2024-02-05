package pql

import (
	"strings"
	"testing"
)

func TestQuoteSQLString(t *testing.T) {
	tests := []struct {
		s    string
		want string
	}{
		{``, `''`},
		{`x`, `'x'`},
		{`x'y`, `'x''y'`},
	}
	for _, test := range tests {
		sb := new(strings.Builder)
		quoteSQLString(sb, test.s)
		if got := sb.String(); got != test.want {
			t.Errorf("quoteSQLString(..., %q) = %q; want %q", test.s, got, test.want)
		}
	}
}
