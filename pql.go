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

	subqueries, err := splitQueries(expr)
	if err != nil {
		return "", err
	}

	sb := new(strings.Builder)
	ctes := subqueries[:len(subqueries)-1]
	query := subqueries[len(subqueries)-1]
	if len(ctes) > 0 {
		sb.WriteString("WITH ")
		for i, sub := range ctes {
			quoteIdentifier(sb, sub.name)
			sb.WriteString(" AS (")
			sub.write(sb)
			sb.WriteString(")")
			if i < len(ctes)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
	}
	query.write(sb)
	sb.WriteString(";")
	return sb.String(), nil
}

type subquery struct {
	name      string
	sourceSQL string

	op   parser.TabularOperator
	sort *parser.SortOperator
	take *parser.TakeOperator
}

func splitQueries(expr *parser.TabularExpr) ([]*subquery, error) {
	var subqueries []*subquery
	var lastSubquery *subquery
	for i := 0; i < len(expr.Operators); i++ {
		switch op := expr.Operators[i].(type) {
		case *parser.SortOperator:
			if lastSubquery != nil && lastSubquery.sort == nil && lastSubquery.take == nil {
				lastSubquery.sort = op
			} else {
				lastSubquery = &subquery{
					sort: op,
				}
				subqueries = append(subqueries, lastSubquery)
			}
		case *parser.TakeOperator:
			if lastSubquery != nil && lastSubquery.take == nil {
				lastSubquery.take = op
			} else {
				lastSubquery = &subquery{
					take: op,
				}
				subqueries = append(subqueries, lastSubquery)
			}
		default:
			lastSubquery = &subquery{
				op: op,
			}
			subqueries = append(subqueries, lastSubquery)
		}
	}

	if len(subqueries) == 0 {
		subqueries = append(subqueries, new(subquery))
	}
	buf := new(strings.Builder)
	for i, sub := range subqueries {
		if i == 0 {
			var err error
			sub.sourceSQL, err = dataSourceSQL(expr.Source)
			if err != nil {
				return nil, err
			}
		} else {
			buf.Reset()
			quoteIdentifier(buf, subqueries[i-1].name)
			sub.sourceSQL = buf.String()
		}

		if i < len(subqueries)-1 {
			sub.name = fmt.Sprintf("subquery%d", i)
		}
	}

	return subqueries, nil
}

func (sub *subquery) write(sb *strings.Builder) {
	switch op := sub.op.(type) {
	case nil:
		sb.WriteString("SELECT * FROM ")
		sb.WriteString(sub.sourceSQL)
	case *parser.CountOperator:
		sb.WriteString("SELECT COUNT(*) FROM ")
		sb.WriteString(sub.sourceSQL)
	default:
		fmt.Fprintf(sb, "/* unsupported operator %T */", op)
	}
}

func dataSourceSQL(src parser.TabularDataSource) (string, error) {
	switch src := src.(type) {
	case *parser.TableRef:
		sb := new(strings.Builder)
		quoteIdentifier(sb, src.Table.Name)
		return sb.String(), nil
	default:
		return "", fmt.Errorf("unhandled data source %T", src)
	}
}

func quoteIdentifier(sb *strings.Builder, name string) {
	const quoteEscape = `""`
	sb.Grow(len(name) + strings.Count(name, `"`)*(len(quoteEscape)-1) + len(`""`))

	sb.WriteString(`"`)
	for _, b := range []byte(name) {
		if b == '"' {
			sb.WriteString(quoteEscape)
		} else {
			sb.WriteByte(b)
		}
	}
	sb.WriteString(`"`)
}
