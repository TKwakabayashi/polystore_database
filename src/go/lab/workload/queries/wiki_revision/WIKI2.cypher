MATCH (p:Page)-[:HAS_REVISION]->(r:Revision)<-[:EDITED]-(c:Contributor)
WHERE p.namespace = $namespace AND r.timestamp >= $since
RETURN p.title, c.name, r.text
ORDER BY r.text_bytes DESC, r.rev_id ASC
LIMIT 10
