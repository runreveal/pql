WITH "__subquery0" AS (SELECT "State" AS "State", "EventType" AS "EventType", "DamageProperty" AS "DamageProperty" FROM "StormEvents")
SELECT *, 42 AS "42" FROM "__subquery0";
