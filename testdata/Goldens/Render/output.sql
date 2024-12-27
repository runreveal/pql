WITH "__subquery0" AS (SELECT "State" AS "State", "EventType" AS "EventType", "DamageProperty" AS "DamageProperty" FROM "StormEvents"),
     "__subquery1" AS (SELECT *, 1 AS "foo" FROM "__subquery0"),
     "__subquery2" AS (SELECT "State" AS "State", sum("DamageProperty") AS "TotalDamage" FROM "__subquery1" GROUP BY "State"),
     "__subquery3" AS (SELECT * FROM "__subquery2" ORDER BY "TotalDamage" DESC NULLS LAST LIMIT 10)
SELECT *,
    'barchart' as "render_type",
    'Property Damage by State' as "render_prop_title",
    'State' as "render_prop_xtitle",
    'Total Damage ($)' as "render_prop_ytitle"
FROM "__subquery3";
