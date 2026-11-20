MATCH (i:Item)<-[r1:REVIEWED]-(u:User)-[r2:REVIEWED]->(j:Item)
WHERE i.rating_number >= $minRatings AND r1.helpful_vote >= $minHelpful AND r2.rating >= $minRating
RETURN j.parent_asin, count(*) AS c
ORDER BY c DESC, j.parent_asin ASC
LIMIT 20
