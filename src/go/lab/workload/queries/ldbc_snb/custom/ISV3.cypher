MATCH (p:Person {firstName: $name})-[:KNOWS*1..2]-(f:Person)
WHERE f.birthday >= $start AND f.birthday < $end
RETURN p.id, f.id, f.firstName, f.lastName
