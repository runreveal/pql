WITH "subquery0" AS (SELECT "State" AS "State" FROM "StormEvents"),
"subquery1" AS (SELECT upper("State") AS "State", "StateCapital" AS "StateCapital" FROM "StateCapitals")
SELECT * FROM "subquery0" AS "$left" JOIN "subquery1" AS "$right" ON "$left"."State" = "$right"."State" ORDER BY "State" ASC NULLS FIRST;
