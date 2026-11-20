MATCH (a:Person {id: $personId})-[:KNOWS]-(b:Person)
WHERE a.id < b.id
RETURN a.id, b.id, b.firstName
ORDER BY b.id ASC
LIMIT 20
