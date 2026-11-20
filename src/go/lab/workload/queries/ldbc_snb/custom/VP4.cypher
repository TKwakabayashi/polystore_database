MATCH (p:Person) WHERE p.id IN $ids
MATCH (p)-[:KNOWS*1..2]-(f:Person)
WITH DISTINCT f
WHERE f.birthday >= $start AND f.birthday < $end
RETURN f.id, f.firstName, f.lastName
