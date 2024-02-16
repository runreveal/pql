WITH "subquery0" AS (SELECT * FROM "Tokens"),
"subquery1" AS (SELECT * FROM (SELECT DISTINCT * FROM "LexResults") AS "$left" JOIN "subquery0" AS "$right" ON ("$left"."Kind" = "$right"."Kind") AND (coalesce("Value" <> 'bar', FALSE)) ORDER BY "SpanStart" ASC NULLS FIRST)
SELECT "TokenConstant" AS "TokenConstant", "Value" AS "Value" FROM "subquery1";
