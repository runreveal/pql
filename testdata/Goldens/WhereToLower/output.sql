SELECT * FROM "StormEvents" WHERE coalesce((LOWER("EventType")) = 'tornado', FALSE);
