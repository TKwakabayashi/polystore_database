MATCH (u:User)-[r:RATED]->(m:Movie)
WHERE m.year = $year
RETURN m.title, count(r) AS ratings, avg(r.rating) AS avgRating
ORDER BY ratings DESC, m.title ASC
LIMIT 20
