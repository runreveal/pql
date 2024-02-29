WITH "__subquery0" AS (SELECT "State" AS "State", "EventType" AS "EventType", "DamageProperty" AS "DamageProperty" FROM "StormEvents")
SELECT *, 1 AS "foo" FROM "__subquery0";
