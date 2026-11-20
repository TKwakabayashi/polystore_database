MATCH (m:Movie {movie_id: '1'})<-[r:RATED]-(u:User)-[r2:RATED]->(o:Movie)
WHERE r.rating >= $minRating AND r.timestamp >= $since
      AND r2.rating >= $minRating AND o.year >= $minYear
RETURN o.title, count(*) AS c
ORDER BY c DESC, o.title ASC
LIMIT 20
