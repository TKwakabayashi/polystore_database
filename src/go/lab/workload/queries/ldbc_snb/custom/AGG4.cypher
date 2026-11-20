MATCH (p:Person {id: $personId})<-[:HAS_CREATOR]-(m:Message)
WHERE m.length >= $minLen AND m.length <= $maxLen
RETURN m.id, m.length
ORDER BY m.length DESC, m.id ASC
LIMIT 20
