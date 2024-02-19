WITH "__subquery0" AS (SELECT * FROM "Tokens"),
     "__subquery1" AS (SELECT * FROM (SELECT DISTINCT * FROM "LexResults") AS "$left" JOIN "__subquery0" AS "$right" ON ("$left"."Kind" = "$right"."Kind") AND (coalesce("Value" <> 'bar', FALSE)) ORDER BY "SpanStart" ASC NULLS FIRST)
SELECT "TokenConstant" AS "TokenConstant", "Value" AS "Value" FROM "__subquery1";
