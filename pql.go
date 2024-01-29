// Package pql provides a Pipeline Query Language that can be translated into SQL.
package pql

import "strings"

// Compile converts the given Pipeline Query Language statement
// into the equivalent SQL.
func Compile(source string) (string, error) {
	return "SELECT * FROM " + strings.TrimSpace(source) + ";", nil
}
