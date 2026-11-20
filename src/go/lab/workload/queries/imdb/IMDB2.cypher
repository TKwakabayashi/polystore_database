MATCH (p:Person {nconst: $person})-[:WORKED_ON]->(t:Title)<-[:WORKED_ON]-(q:Person)-[:WORKED_ON]->(u:Title)
WHERE t.start_year >= $fromYear AND q.birth_year >= $bornAfter AND u.num_votes >= $minVotes
RETURN u.primary_title, count(*) AS c
ORDER BY c DESC, u.primary_title ASC
LIMIT 20
