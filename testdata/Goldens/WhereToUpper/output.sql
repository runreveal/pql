SELECT * FROM "StormEvents" WHERE coalesce((UPPER("EventType")) = 'TORNADO', FALSE);
