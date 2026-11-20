MATCH (p:Person)-[:KNOWS]-(f:Person)-[:POSTED]->(m:Post)
WHERE p.first_name = $firstName AND f.browser = $browser AND m.creation_date >= $since
RETURN f.person_id, m.post_id, m.content
ORDER BY m.creation_date DESC, m.post_id ASC
LIMIT 20
