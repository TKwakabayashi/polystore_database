MATCH (a:Person)-[:MEMBER_OF]->(fo:Forum)<-[:MODERATES]-(mod:Person)
WHERE a.gender = $gender AND a.birthday >= $bornAfter AND fo.creation_date >= $since
RETURN mod.person_id, count(DISTINCT a.person_id) AS members
ORDER BY members DESC, mod.person_id ASC
LIMIT 20
