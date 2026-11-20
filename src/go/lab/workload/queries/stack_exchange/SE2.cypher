MATCH (q:Post)-[:PARENT_OF]->(a:Post)
WHERE q.score >= $minScore
RETURN q.post_id, a.post_id, a.body
ORDER BY a.score DESC, a.post_id ASC
LIMIT 10
