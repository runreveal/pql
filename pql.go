// Package pql provides a Pipeline Query Language that can be translated into SQL.
package pql

import (
	"fmt"

	"github.com/runreveal/pql/parser"
)

// Compile converts the given Pipeline Query Language statement
// into the equivalent SQL.
func Compile(source string) (string, error) {
	tokens := parser.Scan(source)
	if len(tokens) == 1 && tokens[0].Kind == parser.TokenIdentifier {
		return "SELECT * FROM " + tokens[0].Value + ";", nil
	}
	return "", fmt.Errorf("compile pipeline %q: not handled", source)
}
