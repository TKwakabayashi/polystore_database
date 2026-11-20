MATCH (u:User)-[r:REVIEWED]->(i:Item)
WHERE i.average_rating <= $maxAvgRating AND r.helpful_vote >= $minHelpful
RETURN i.title, r.title, r.text
ORDER BY r.helpful_vote DESC, i.title ASC
LIMIT 10
