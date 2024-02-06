SELECT * FROM "StormEvents" WHERE ("DamageProperty" > 5000) AND (coalesce("EventType" = 'Thunderstorm Wind', FALSE));
