MATCH (m:Message)
WHERE m.creationDate >= $start AND m.creationDate < $end
MATCH (m)-[:REPLY_OF*0..]->(:Post)<-[:CONTAINER_OF]-(f:Forum)-[:HAS_MODERATOR]->(mod:Person)
RETURN f.id, f.title, mod.id, mod.firstName, mod.lastName
