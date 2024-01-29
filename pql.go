// Package pql provides a Pipeline Query Language that can be translated into SQL.
package pql

import (
	"fmt"
	"strings"

	"github.com/runreveal/pql/parser"
)

// Compile converts the given Pipeline Query Language statement
// into the equivalent SQL.
func Compile(source string) (string, error) {
	tokens := parser.Scan(source)
	if len(tokens) == 1 && (tokens[0].Kind == parser.TokenIdentifier || tokens[0].Kind == parser.TokenQuotedIdentifier) {
		return "SELECT * FROM " + quoteIdentifier(tokens[0].Value) + ";", nil
	}
	return "", fmt.Errorf("compile pipeline %q: not handled", source)
}

func quoteIdentifier(name string) string {
	if sqlIdentifierNeedsQuote(name) {
		return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
	}
	return name
}

func sqlIdentifierNeedsQuote(name string) bool {
	if name == "" || !isAlpha(rune(name[0])) && name[0] != '_' {
		return true
	}
	for i := 1; i < len(name); i++ {
		if !isAlpha(rune(name[i])) && !isDigit(rune(name[i])) && name[i] != '_' {
			return true
		}
	}
	return false
}

func isAlpha(c rune) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

func isDigit(c rune) bool {
	return '0' <= c && c <= '9'
}
