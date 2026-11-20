MATCH (p:Person) WHERE p.id IN $ids
MATCH (p)-[:KNOWS*1..2]-(f:Person)
WHERE f.birthday >= $start AND f.birthday < $end
RETURN f.id, f.firstName, f.lastName
ORDER BY f.birthday DESC, f.id ASC
LIMIT 100
