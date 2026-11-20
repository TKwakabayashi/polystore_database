MATCH (p:Person {nconst: $person})-[:WORKED_ON]->(t:Title)<-[:WORKED_ON]-(q:Person)
WHERE t.title_type = $type AND t.start_year >= $fromYear AND t.average_rating >= $minRating
RETURN q.primary_name, count(*) AS together
ORDER BY together DESC, q.primary_name ASC
LIMIT 20
