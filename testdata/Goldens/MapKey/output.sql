WITH "__subquery0" AS (SELECT * FROM "MapTable" WHERE ("a"['key2']) > 10 ORDER BY "id" ASC NULLS FIRST)
SELECT "a"['key1'] AS "Key1", "a"['key2'] AS "Key2" FROM "__subquery0";
