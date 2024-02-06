WITH "subquery0" AS (SELECT * FROM "SourceFiles" ORDER BY "LineCount" DESC NULLS LAST, "FileName" ASC NULLS FIRST)
SELECT "FileName" AS "FileName", CASE WHEN coalesce(("LineCount") >= (1000), FALSE) THEN 'Large' ELSE 'Smol' END AS "Size" FROM "subquery0";
