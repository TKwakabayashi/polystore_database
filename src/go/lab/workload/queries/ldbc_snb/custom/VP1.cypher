MATCH (p:Person) WHERE p.id IN $ids
MATCH (p)-[:KNOWS*1..2]-(f:Person)
WHERE f.birthday >= $start AND f.birthday < $end
RETURN p.id, f.id, f.firstName, f.lastName
