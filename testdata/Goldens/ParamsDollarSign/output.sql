SELECT * FROM "Tokens" WHERE coalesce("Kind" = {$desiredKind: Int32}, FALSE);
