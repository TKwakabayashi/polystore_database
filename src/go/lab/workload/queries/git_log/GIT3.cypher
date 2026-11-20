MATCH (au:Author)-[:AUTHORED]->(c:Commit)-[t:TOUCHED]->(f:File)
WHERE c.committed_date >= $since AND t.insertions >= $minInsertions
RETURN au.name, count(DISTINCT f.path) AS files
ORDER BY files DESC, au.name ASC
LIMIT 20
