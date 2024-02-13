# Hacking on PQL

## How to implement new scalar functions

Function calls are the easiest to work with,
since they require no parser changes.
A function call is represented as a [`*parser.CallExpr`][]
and specific translations are registered
in `initKnownFunctions` in [pql.go](pql.go).

1.  Add a new entry to the function table in `initKnownFunctions`.
2.  Create a callback function that writes the SQL you want.
    Callbacks can inspect the full AST of their function call
    and can use `writeExpressionMaybeParen` to translate their arguments.

[`*parser.CallExpr`]: https://pkg.go.dev/github.com/runreveal/pql/parser#CallExpr

## How to implement new tabular operators

Tabular operators are more involved than scalar functions
because each operator's syntax is distinct.
PQL uses a [recursive descent parser][],
so parsing rules are written as plain Go code that consumes tokens.

1.  Create a new struct for your operator in [ast.go](parser/ast.go)
    that implements `TabularOperator`.
    Look to other `TabularOperator` types in ast.go
    for inspiration.
2.  Add another case into the `switch` inside the `*parser.tabularExpr` method.
3.  Add a parsing method to `*parser`
    that converts tokens into your new `TabularOperator` type.
    You can look at `*parser.whereOperator` and `*parser.takeOperator`
    as basic examples.
4.  Add a [test](parser/parser_test.go)
    to ensure that your tabular operator is parsed as you expect.

Now for compilation:

5.  Add a new case to `*subquery.write` in [pql.go](pql.go)
    to transform your parsed tabular operator struct into a `SELECT` statement.
6.  If necessary, add logic into the `canAttachSort` function
    to signal the behavior of the operator to the subquery split algorithm.
7.  Add an [end-to-end test](testdata/Goldens/README.md)
    to verify that your operator compiles as expected.

[recursive descent parser]: https://en.wikipedia.org/wiki/Recursive_descent_parser

## How to add a new token

If you're adding a new syntactical element, you should first add it to the [lexer](parser/lex.go).

1.  Write a [test](parser/lex_test.go) that uses your new token.
    It will fail to start.
    Defects in lexing can lead to surprising problems during parsing,
    so it's important to always test lexing independently.
2.  Add the new token type to the list in [lex.go](parser/lex.go).
3.  Run `go generate` inside the `parser` directory.
4.  Modify the `Scan` function to detect your token.
    As a special case, if you are adding a new keyword,
    you can add it to the `keywords` map variable.
5.  Re-run your test to ensure it passes.

You will then need to modify the parser to handle this new type of token.
The exact set of changes varies depending on how the token is used,
but the process will be similar to adding a new tabular operator.
