# Pipeline Query Language

This Go library compiles a pipeline-based query language
(inspired by the [Kusto Query Language][])
into SQL.
It has been specifically tested to work with the [Clickhouse SQL dialect][],
but the generated SQL is intentionally database agnostic.

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

## Features

The following tabular operators are supported:

- [`as`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/as-operator)
- [`count`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/count-operator)
- [`join`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/join-operator)
- [`project`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/project-operator)
- [`sort`/`order`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/sort-operator)
- [`summarize`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/summarize-operator)
- [`take`/`limit`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/take-operator)
- [`top`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/top-operator)
- [`where`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/where-operator)

The following functions are specifically handled.
Functions not in this list will be passed through to the underlying SQL engine.

- [`not`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/not-function)
- [`isnull`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/isnull-function)
  and [`isnotnull`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/isnotnull-function)
- [`strcat`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/strcat-function)
- [`iff`/`iif`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/iff-function)
- [`count`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/count-aggregation-function)
- [`countif`](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/countif-aggregation-function)

## License

[Apache 2.0](LICENSE)
