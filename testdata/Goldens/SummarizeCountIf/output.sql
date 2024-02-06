WITH "subquery0" AS (SELECT "Directory" AS "Directory", sum(CASE WHEN coalesce(NOT endsWith("FileName", '_test.go'), FALSE) THEN 1 ELSE 0 END) AS "NonTestFiles" FROM "SourceFiles" GROUP BY "Directory")
SELECT * FROM "subquery0" ORDER BY "Directory" ASC NULLS FIRST;
