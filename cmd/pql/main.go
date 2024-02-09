package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/runreveal/pql"
	"github.com/runreveal/pql/parser"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	sb := new(strings.Builder)
	for scanner.Scan() {
		sb.Write(scanner.Bytes())
		statements := parser.SplitStatements(sb.String())
		if len(statements) == 1 {
			continue
		}

		for _, stmt := range statements[:len(statements)-1] {
			sql, err := pql.Compile(stmt)
			if err != nil {
				fmt.Fprintf(os.Stderr, "pql: %v\n", err)
				continue
			}
			fmt.Println(sql)
			fmt.Println()
		}

		sb.Reset()
		sb.WriteString(statements[len(statements)-1])
	}
}
