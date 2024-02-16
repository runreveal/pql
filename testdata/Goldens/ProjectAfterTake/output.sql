WITH "__subquery0" AS (SELECT * FROM "StormEvents" ORDER BY "EventId" ASC NULLS FIRST LIMIT 3)
SELECT "State" AS "State", "EventType" AS "EventType", "DamageProperty" AS "DamageProperty" FROM "__subquery0";
