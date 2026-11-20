MATCH (t:Title {tconst: $title})<-[:WORKED_ON]-(p:Person)-[:WORKED_ON]->(o:Title)
WHERE o.title_type = $type
RETURN o.start_year, count(*) AS titles, avg(o.average_rating) AS avgRating
ORDER BY o.start_year ASC
