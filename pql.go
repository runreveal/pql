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
	expr, err := parser.Parse(source)
	if err != nil {
		return "", err
	}

	dataSource, err := dataSourceSQL(expr.Source)
	if err != nil {
		return "", err
	}

	switch {
	case len(expr.Operators) == 0:
		return "SELECT * FROM " + dataSource + ";", nil
	case len(expr.Operators) == 1:
		switch op := expr.Operators[0].(type) {
		case *parser.CountOperator:
			return "SELECT COUNT(*) FROM " + dataSource + ";", nil
		default:
			return "", fmt.Errorf("unsupported operator %T", op)
		}
	default:
		return "", fmt.Errorf("only one operator implemented")
	}
}

func dataSourceSQL(src parser.TabularDataSource) (string, error) {
	switch src := src.(type) {
	case *parser.TableRef:
		return quoteIdentifier(src.Table.Name), nil
	default:
		return "", fmt.Errorf("unhandled data source %T", src)
	}
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
