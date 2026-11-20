MATCH (u:User)-[r:RATED]->(m:Movie)
WHERE r.timestamp >= $since AND r.rating = $rating
RETURN m.movie_id, count(*) AS c
ORDER BY c DESC, m.movie_id ASC
LIMIT 20
