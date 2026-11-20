MATCH (u:User)-[r:REVIEWED]->(i:Item)
WHERE i.main_category = $category AND r.rating >= $minRating
RETURN i.parent_asin, count(*) AS reviews, avg(r.rating) AS avgRating
ORDER BY reviews DESC, i.parent_asin ASC
LIMIT 20
