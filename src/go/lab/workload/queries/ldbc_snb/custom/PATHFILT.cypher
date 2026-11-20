MATCH p=(person:Person {id: $personId})-[:KNOWS*1..3]-(other:Person)
WHERE all(r IN relationships(p) WHERE r.creationDate > $minDate)
      AND length(p) <= 3
RETURN other.id, other.firstName
ORDER BY other.id ASC
LIMIT 20
