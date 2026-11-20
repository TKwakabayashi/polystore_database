MATCH (u:User)-[:MADE_EDIT]->(r:Rev)<-[:HAS_EDIT]-(p:Post)
WHERE u.reputation >= $minReputation AND p.score >= $minScore
RETURN p.post_id, r.rev_id, r.text
ORDER BY r.creation_date DESC, r.rev_id ASC
LIMIT 10
