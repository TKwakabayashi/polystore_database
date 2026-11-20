MATCH (u:User {user_id: '1'})-[r1:RATED]->(m:Movie)<-[r2:RATED]-(v:User)
WHERE r1.rating >= $minRating AND r2.rating >= $minRating
RETURN v.user_id, count(*) AS common
ORDER BY common DESC, v.user_id ASC
LIMIT 20
