MATCH (c:Commit)-[:PARENT]->(p:Commit)
WHERE c.insertions >= $minInsertions
RETURN c.sha, c.insertions, p.message
ORDER BY c.insertions DESC, c.sha ASC
LIMIT 10
