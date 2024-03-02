# pipelined query language

[![Website](https://img.shields.io/badge/INTRO-WEB-blue?style=for-the-badge)](https://pql.dev)
[![Playground](https://img.shields.io/badge/INTRO-PLAYGROUND-blue?style=for-the-badge)](https://pql.dev)
[![Discord](https://img.shields.io/discord/1120882187785470113?label=discord%20chat&style=for-the-badge)](https://discord.gg/PbeXzrWP)


This Go library compiles a pipelined-based query language
(inspired by the [Kusto Query Language][])
into SQL.
It has been specifically tested to work with the [Clickhouse SQL dialect][],
but the generated SQL is intentionally database agnostic. This repository
contains a the Go library, and a CLI to invoke the library.

For example, the following expression:

```plain
StormEvents
| where DamageProperty > 5000 and EventType == "Thunderstorm Wind"
| top 3 by DamageProperty
```

will be compiled to SQL that is similar to:

```sql
SELECT *
FROM StormEvents
WHERE DamageProperty > 5000 AND EventType = 'Thunderstorm Wind'
ORDER BY DamageProperty DESC
LIMIT 3;
```

[Kusto Query Language]: https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/
[Clickhouse SQL dialect]: https://clickhouse.com/docs/en/sql-reference

## Getting Started
If you'd like to see a demo along with some examples, check out https://pql.dev.

To use pql in your go code, a minimal example might look like this
```
package main

import (
	"github.com/runreveal/pql"
)

func main() {
	sql, err := pql.Compile("users | project id, email | limit 5")
	if err != nil {
		panic(err)
	}
	println(sql)
}
```

Running this program should give you the following output
```
$ go run test.go

WITH "__subquery0" AS (SELECT "id" AS "id", "email" AS "email" FROM "users")
SELECT * FROM "__subquery0" LIMIT 5;
```

## Documentation

The following tabular operators are supported and the Microsoft KQL
documentation is representative of the current pql api.

- [`as`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/as-operator)
- [`count`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/count-operator)
- [`join`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/join-operator)
- [`project`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/project-operator)
- [`extend`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/extend-operator)
- [`sort`/`order`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/sort-operator)
- [`summarize`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/summarize-operator)
- [`take`/`limit`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/take-operator)
- [`top`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/top-operator)
- [`where`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/where-operator)

The following scalar functions implemented within pql. Functions not in this
list will be passed through to the underlying SQL engine. This allows the usage
of the full APIs implemented by the underlying

- [`not`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/not-function)
- [`now`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/now-function)
- [`isnull`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/isnull-function)
  and [`isnotnull`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/isnotnull-function)
- [`strcat`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/strcat-function)
- [`iff`/`iif`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/iff-function)
- [`count`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/count-aggregation-function)
- [`countif`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/countif-aggregation-function)

Column names with special characters can be escaped with backticks.

## Get involved
- Join our [discord](https://discord.gg/XWKF5s5g)
- Contribute a [scalar function](./CONTRIBUTING.md)

## License
[Apache 2.0](LICENSE)
