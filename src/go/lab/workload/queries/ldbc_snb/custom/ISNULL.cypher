MATCH (p:Person {id: $personId})<-[:HAS_CREATOR]-(m:Message)
WHERE m.content IS NOT NULL
RETURN m.id, m.content
ORDER BY m.id ASC
LIMIT 20
