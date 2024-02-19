WITH "__subquery0" AS (SELECT * FROM "MyLogTable" WHERE coalesce("TargetType" = 'X', FALSE)),
"T" AS (SELECT * FROM "__subquery0"),
"__subquery2" AS (SELECT * FROM "T" WHERE coalesce("EventType" = 'Start', FALSE)),
"__subquery3" AS (SELECT * FROM "T" WHERE coalesce("EventType" = 'Stop', FALSE)),
"__subquery4" AS (SELECT "TargetId" AS "TargetId", "EventId" AS "StopEventId" FROM "__subquery3"),
"__subquery5" AS (SELECT * FROM "__subquery2" AS "$left" LEFT JOIN "__subquery4" AS "$right" ON "$left"."TargetId" = "$right"."TargetId"),
"__subquery6" AS (SELECT "TargetId" AS "TargetId", "EventId" AS "StartEventId", coalesce("StopEventId", -1) AS "StopEventId" FROM "__subquery5")
SELECT * FROM "__subquery6" ORDER BY "StartEventId" ASC NULLS FIRST;
