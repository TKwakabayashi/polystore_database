MATCH (p:Person) WHERE p.id IN $ids
MATCH (p)-[:KNOWS*1..2]-(f:Person)<-[:HAS_CREATOR]-(m:Message)
WHERE m.creationDate >= $start AND m.creationDate < $end
RETURN f.id, f.lastName, count(m) AS n, avg(m.length) AS avgLen
