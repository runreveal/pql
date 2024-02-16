WITH "__subquery0" AS (SELECT "Directory" AS "Directory", count() FILTER (WHERE NOT endsWith("FileName", '_test.go')) AS "NonTestFiles" FROM "SourceFiles" GROUP BY "Directory")
SELECT * FROM "__subquery0" ORDER BY "Directory" ASC NULLS FIRST;
